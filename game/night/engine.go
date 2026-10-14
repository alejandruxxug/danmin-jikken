package night

import (
	"errors"
	"fmt"
	"sort"
)

// TicksPerSecond is how often the game calls Tick.
const TicksPerSecond = 60

// DefaultSpeed is the typewriter speed, in letters per second.
const DefaultSpeed = 45

// maxSteps stops a script that jumps in a circle without ever waiting for
// the player (a mistake in a script file) from freezing the game.
const maxSteps = 10000

// fxLength is how many ticks each timed effect lasts. The painter divides
// the ticks since the effect began by this to get a progress from 0 to 1.
var fxLength = map[string]int{
	"fade_in":    90,
	"eyes_close": 72,
	"red_pixel":  30,
}

// Engine runs one script. It knows nothing about drawing: the painter
// (draw.go) reads what the engine exposes, and the scene feeds it ticks and
// the player's clicks. That keeps the whole story logic testable.
type Engine struct {
	script *Script
	flags  Flags
	pc     int // index of the node being shown or waited on

	// What the screen shows.
	bg      string
	props   map[string]bool
	visible string // the character on screen ("" for none)
	faces   bool   // the Prologue's two upside-down faces
	poses   map[string]Pose
	fx      map[string]int // effect name -> ticks since it started

	// The line being said.
	speaker  string // lower-case id; "" for a thought
	text     []rune
	shown    float64 // letters revealed so far
	speed    float64 // letters per second
	thinking bool

	// A menu, while the player has to pick.
	options []Option
	hover   int

	waitLeft int // ticks left of a wait command
	waiting  bool

	ended  bool
	endsTo string

	// Cues for the sound (M7): the last ones the script asked for.
	LastSFX, LastMusic string
	ticks              int
}

// NewEngine starts a script. flags is the shared store: choices write into
// it and "if" reads from it. The engine runs up to the first line the
// player has to read.
func NewEngine(s *Script, flags Flags) (*Engine, error) {
	e := &Engine{
		script: s, flags: flags,
		bg: "black", props: map[string]bool{},
		poses: map[string]Pose{"mayu": startPose("mayu"), "hana": startPose("hana")},
		fx:    map[string]int{}, speed: DefaultSpeed,
	}
	if err := e.run(); err != nil {
		return nil, err
	}
	return e, nil
}

// run executes nodes from pc until one needs the player: a line, a menu, a
// wait or the end.
func (e *Engine) run() error {
	for step := 0; step < maxSteps; step++ {
		if e.pc >= len(e.script.Nodes) {
			e.ended, e.endsTo = true, ""
			return nil
		}
		n := e.script.Nodes[e.pc]
		switch n.Op {
		case OpSay, OpThink:
			e.startLine(n)
			return nil
		case OpChoice:
			e.options, e.hover = n.Options, 0
			return nil
		case OpWait:
			e.waiting, e.waitLeft = true, int(n.Number*TicksPerSecond)
			return nil
		case OpEnd:
			e.ended, e.endsTo = true, n.Name
			return nil
		}
		e.exec(n)
	}
	return fmt.Errorf("%s: more than %d commands without a line: a jump loop? (near line %d)",
		e.script.Name, maxSteps, e.script.Nodes[e.pc].Line)
}

// exec does one command that does not stop the story and moves pc on.
func (e *Engine) exec(n Node) {
	e.pc++
	switch n.Op {
	case OpJump:
		e.pc = e.script.Labels[n.Text]
	case OpIf:
		if e.flags.Has(n.Name) != n.Not {
			e.pc = e.script.Labels[n.Text]
		}
	case OpSet:
		e.flags.Set(n.Name)
	case OpBG:
		e.bg = n.Name
	case OpProp:
		if n.Name == "clear" {
			e.props = map[string]bool{}
		} else {
			e.props[n.Name] = true
		}
	case OpShow:
		e.visible, e.faces = n.Name, false
	case OpHide:
		e.visible, e.faces = "", false
	case OpFaces:
		e.visible, e.faces = "", true
		e.fx["faces"] = 0
	case OpPose:
		e.setPose(n)
	case OpFx:
		switch n.Name {
		case "flat_line":
			delete(e.fx, "line_back")
		case "line_back":
			delete(e.fx, "flat_line")
		}
		e.fx[n.Name] = 0
	case OpSfx:
		e.LastSFX = n.Name
	case OpMusic:
		e.LastMusic = n.Name
	case OpSpeed:
		e.speed = n.Number
	}
}

// setPose changes a character's pose from a pose command or a line.
func (e *Engine) setPose(n Node) {
	p := e.poses[n.Name]
	if n.Expr != "" {
		p = expressions[n.Name][n.Expr]
	}
	for _, part := range n.Parts {
		// The parser already checked every part, so an error cannot happen.
		_ = applyPart(&p, n.Name, part.Key, part.Value)
	}
	e.poses[n.Name] = p
}

// startLine begins a line: a character who speaks steps in (unless the two
// faces are on screen) and takes her expression.
func (e *Engine) startLine(n Node) {
	e.speaker, e.thinking = "", n.Op == OpThink
	e.text, e.shown = []rune(n.Text), 0
	if n.Op == OpThink {
		return
	}
	e.speaker = n.Name
	if IsCharacter(n.Name) {
		e.setPose(n)
		if !e.faces {
			e.visible = n.Name
		}
	}
}

// Tick runs one frame: the typewriter, the effect clocks and the waits.
func (e *Engine) Tick() error {
	e.ticks++
	for name := range e.fx {
		e.fx[name]++
	}
	if e.waiting {
		e.waitLeft--
		if e.waitLeft <= 0 {
			return e.leaveWait()
		}
		return nil
	}
	if e.speaker != "" || e.thinking {
		e.shown = min(e.shown+e.speed/TicksPerSecond, float64(len(e.text)))
	}
	return nil
}

func (e *Engine) leaveWait() error {
	e.waiting = false
	e.pc++
	return e.run()
}

// Advance is the player's click (or Space, or Enter). A line that is still
// typing shows in full; a finished line goes on to the next. It cuts a wait
// short and does nothing in a menu or at the end.
func (e *Engine) Advance() error {
	switch {
	case e.ended || len(e.options) > 0:
		return nil
	case e.waiting:
		return e.leaveWait()
	case !e.LineDone():
		e.shown = float64(len(e.text))
		return nil
	}
	e.pc++
	return e.run()
}

// Choose picks option i (0-based) of the open menu: its flags are set and
// the story goes to its label.
func (e *Engine) Choose(i int) error {
	if len(e.options) == 0 {
		return errors.New("night: no choice is open")
	}
	if i < 0 || i >= len(e.options) {
		return fmt.Errorf("night: option %d: the menu has %d options", i+1, len(e.options))
	}
	opt := e.options[i]
	for _, flag := range opt.Set {
		e.flags.Set(flag)
	}
	e.options = nil
	if opt.Jump != "" {
		e.pc = e.script.Labels[opt.Jump]
	} else {
		e.pc++
	}
	return e.run()
}

// MoveHover moves the highlighted option with the arrow keys, wrapping round.
func (e *Engine) MoveHover(delta int) {
	if n := len(e.options); n > 0 {
		e.hover = ((e.hover+delta)%n + n) % n
	}
}

// SetHover highlights option i (the mouse is over it).
func (e *Engine) SetHover(i int) {
	if i >= 0 && i < len(e.options) {
		e.hover = i
	}
}

// FastForward plays the script silently up to the first command on or after
// the given file line, for the agent backdoor (-line N). Lines are
// skipped as if clicked; a menu is answered with the first option whose
// flags are already on (-flags), or the first option.
func (e *Engine) FastForward(line int) error {
	for step := 0; step < maxSteps; step++ {
		if e.ended || e.script.Nodes[e.pc].Line >= line {
			return nil
		}
		var err error
		switch {
		case len(e.options) > 0:
			err = e.Choose(e.pickedOption())
		case e.waiting:
			err = e.leaveWait()
		default:
			e.shown = float64(len(e.text))
			err = e.Advance()
		}
		if err != nil {
			return err
		}
	}
	return fmt.Errorf("%s: -line %d did not stop", e.script.Name, line)
}

// pickedOption is the option -flags already chose, else the first.
func (e *Engine) pickedOption() int {
	for i, opt := range e.options {
		if len(opt.Set) > 0 && allOn(e.flags, opt.Set) {
			return i
		}
	}
	return 0
}

func allOn(f Flags, names []string) bool {
	for _, name := range names {
		if !f.Has(name) {
			return false
		}
	}
	return true
}

// ---- what the painter and the dump read ----

// LineDone says whether the whole line is on screen.
func (e *Engine) LineDone() bool { return e.shown >= float64(len(e.text)) }

// Shown is the part of the line typed so far.
func (e *Engine) Shown() string { return string(e.text[:int(e.shown)]) }

// FullText is the whole line (the painter wraps it before cutting).
func (e *Engine) FullText() string { return string(e.text) }

// ShownCount is how many letters are revealed.
func (e *Engine) ShownCount() int { return int(e.shown) }

// Speaker is the lower-case id of the speaker ("" for a thought).
func (e *Engine) Speaker() string { return e.speaker }

// Thinking says whether the line is the player's thought.
func (e *Engine) Thinking() bool { return e.thinking }

// HasLine says whether a line (not a menu or a wait) is on the screen.
func (e *Engine) HasLine() bool {
	return !e.ended && !e.waiting && len(e.options) == 0 && len(e.text) > 0
}

// Options returns the open menu's options (nil when there is none).
func (e *Engine) Options() []Option { return e.options }

// Hover is the highlighted option.
func (e *Engine) Hover() int { return e.hover }

// BG is the background name.
func (e *Engine) BG() string { return e.bg }

// Prop says whether a prop is out.
func (e *Engine) Prop(name string) bool { return e.props[name] }

// Visible is the character on screen ("" for none).
func (e *Engine) Visible() string { return e.visible }

// Faces says whether the Prologue's two faces are leaning in.
func (e *Engine) Faces() bool { return e.faces }

// PoseOf is a character's current pose.
func (e *Engine) PoseOf(character string) Pose { return e.poses[character] }

// FxTicks is how long an effect has been running; ok is false if it has not
// started. Timed effects (fxLength) stay at their last tick.
func (e *Engine) FxTicks(name string) (ticks int, ok bool) {
	ticks, ok = e.fx[name]
	return ticks, ok
}

// FxProgress is a timed effect's progress from 0 to 1 (1 once finished,
// 0 if it never started).
func (e *Engine) FxProgress(name string) float64 {
	ticks, ok := e.fx[name]
	if !ok {
		return 0
	}
	return min(1, float64(ticks)/float64(fxLength[name]))
}

// Ended says whether the script is over, and where it goes next (the name
// after "end").
func (e *Engine) Ended() (string, bool) { return e.endsTo, e.ended }

// Line is the file line of the command being shown (0 at the end).
func (e *Engine) Line() int {
	if e.ended || e.pc >= len(e.script.Nodes) {
		return 0
	}
	return e.script.Nodes[e.pc].Line
}

// Ticks is how many frames the engine has run.
func (e *Engine) Ticks() int { return e.ticks }

// Flags is the shared store.
func (e *Engine) Flags() Flags { return e.flags }

// Name is the script's name.
func (e *Engine) Name() string { return e.script.Name }

// State is a snapshot for the agent dump.
type State struct {
	Script    string   `json:"script"`
	Line      int      `json:"line"`
	Speaker   string   `json:"speaker"`
	Text      string   `json:"text"`
	Thinking  bool     `json:"thinking"`
	LineDone  bool     `json:"lineDone"`
	Options   []string `json:"options"`
	Hover     int      `json:"hover"`
	BG        string   `json:"bg"`
	Visible   string   `json:"visible"`
	Faces     bool     `json:"faces"`
	Layers    []string `json:"poseLayers"`
	Props     []string `json:"props"`
	Effects   []string `json:"effects"`
	Ended     bool     `json:"ended"`
	EndsTo    string   `json:"endsTo"`
	LastSFX   string   `json:"sfx"`
	LastMusic string   `json:"music"`
}

// Snapshot describes the engine for the dump.
func (e *Engine) Snapshot() State {
	st := State{
		Script: e.script.Name, Line: e.Line(), Speaker: e.speaker,
		Text: e.Shown(), Thinking: e.thinking, LineDone: e.LineDone(),
		Options: []string{}, Hover: e.hover, BG: e.bg, Visible: e.visible,
		Faces: e.faces, Layers: []string{}, Props: []string{}, Effects: []string{},
		Ended: e.ended, EndsTo: e.endsTo, LastSFX: e.LastSFX, LastMusic: e.LastMusic,
	}
	for _, o := range e.options {
		st.Options = append(st.Options, o.Text)
	}
	if e.visible != "" {
		st.Layers = e.poses[e.visible].Layers()
	}
	for name, on := range e.props {
		if on {
			st.Props = append(st.Props, name)
		}
	}
	for name := range e.fx {
		st.Effects = append(st.Effects, name)
	}
	sort.Strings(st.Props)
	sort.Strings(st.Effects)
	return st
}
