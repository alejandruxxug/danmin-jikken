package night

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// TicksPerSecond is how often the game calls Tick.
const TicksPerSecond = 60

// paperSpeed is how fast the pen writes a Note: or Letter: line, in letters
// per second (Art Direction -> Cinematic pass).
const paperSpeed = 30

// DefaultSpeed is the typewriter speed, in letters per second.
const DefaultSpeed = 45

// maxSteps stops a script that jumps in a circle without ever waiting for
// the player (a mistake in a script file) from freezing the game.
const maxSteps = 10000

// autoBeat is how long (in ticks) a line that types itself stays on screen
// once it is complete.
const autoBeat = 60

// fxLength is how many ticks each timed effect lasts. The painter divides
// the ticks since the effect began by this to get a progress from 0 to 1.
var fxLength = map[string]int{
	"fade_in":    90,
	"eyes_close": 72,
	"faces":      40,
	"red_pixel":  30,
	"shake":      45,
	// The redscale comes in from the floor in 1 s and drains out in 1 s.
	"redscale":     60,
	"redscale_off": 60,
	"dim":          120,
	"eyes_slit":    150,
	"eyes_open":    72,
	"sting":        48, // 0.8 s
	"wash":         150,
}

// oneShotFX are the effects that leave nothing behind once they have played:
// the shake, the red pixel, the sting, the fade-in and the eyes opening.
var oneShotFX = map[string]int{"shake": 45, "red_pixel": 30, "sting": 48, "fade_in": 90, "eyes_open": 72}

// hopTable is the talk hop: how many pixels up (negative) the speaking girl
// is, tick by tick after she starts talking. 20 ticks = 0.33 s, whole pixels:
// a slow rise, a short hang at the top, the fall, then a small bounce.
var hopTable = [20]int{-1, -1, -2, -2, -3, -3, -3, -3, -2, -2, -1, -1, 0, 0, -1, -1, 0, 0, 0, 0}

// Cue is a sound the script asked for: Kind is "sfx" or "music".
type Cue struct{ Kind, Name string }

// TakeCues returns the sound cues since the last call. The scene plays them.
func (e *Engine) TakeCues() []Cue {
	c := e.cues
	e.cues = nil
	return c
}

// Engine runs one script. It knows nothing about drawing: the painter
// (draw.go) reads what the engine exposes, and the scene feeds it ticks and
// the player's clicks. That keeps the whole story logic testable.
type Engine struct {
	script *Script
	flags  Flags
	pc     int // index of the node being shown or waited on

	// What the screen shows.
	bg     string
	props  map[string]bool
	layers map[string]bool // a CG background: the layers switched on by "layer"
	// The paper (bg note / bg letter): whose hand writes on it, and every
	// Note:/Letter: line so far. Earlier lines stay; a new bg empties it.
	writer string
	paper  []string
	bgAt   int // the tick the current bg came up (an animated CG plays from there)
	// The last paper that was put away (the realization's sting shows it).
	lastWriter string
	lastPaper  []string
	// Cinematic mode: the bars, the camera zoom and the sideways pan.
	bars    barState
	zoom    zoomState
	pan     panState
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
	holding  bool // the wait is a hold: only a click (or AUTO) ends it
	holdTick int  // ticks the hold has been on screen
	locked   bool // the wait ignores clicks (the game "falls asleep")

	autoNext bool // the next line goes on by itself
	autoLine bool // ... this one does
	doneFor  int  // ticks the auto line has been fully typed

	// AutoPlay is the AUTO toggle on the text box: a finished line goes on by
	// itself after autoDelay. It never answers a menu.
	AutoPlay bool
	autoWait int // ticks the finished line has been on screen

	// Player is the name {player} stands for in the words (the OS username).
	Player string

	ended  bool
	endsTo string

	still string // see CutToStill

	cues []Cue // sound cues not yet taken by the scene

	// Cues for the sound (M7): the last ones the script asked for.
	LastSFX, LastMusic string
	ticks              int

	hopStart int    // the tick the current hop began; -1 for none
	hopFor   string // who spoke last ("" after a thought or a scene change): no hop while it stays her
}

// NewEngine starts a script. flags is the shared store: choices write into
// it and "if" reads from it. The engine runs up to the first line the
// player has to read.
func NewEngine(s *Script, flags Flags) (*Engine, error) {
	return NewEngineFor(s, flags, "")
}

// NewEngineFor is NewEngine for a named player: {player} in the words
// becomes that name (the OS username, see the meta-horror in Script).
func NewEngineFor(s *Script, flags Flags, player string) (*Engine, error) {
	e := &Engine{Player: player,
		script: s, flags: flags,
		bg: "black", props: map[string]bool{}, layers: map[string]bool{}, writer: "hana",
		poses: map[string]Pose{"mayu": startPose("mayu"), "hana": startPose("hana")},
		fx:    map[string]int{}, speed: DefaultSpeed, hopStart: -1,
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
		case OpWait, OpLock:
			e.waiting, e.waitLeft = true, int(n.Number*TicksPerSecond)
			e.locked = n.Op == OpLock
			return nil
		case OpHold:
			e.waiting, e.holding, e.holdTick = true, true, 0
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
	case OpRoute:
		e.flags.Set("route_" + RouteFor(e.flags))
	case OpAuto:
		e.autoNext = true
	case OpBG:
		e.hopFor = ""
		if e.OnPaperBG() { // the sting still needs the letter after a cut away from it
			e.lastWriter, e.lastPaper = e.writer, e.paper
		}
		e.bg, e.bgAt = n.Name, e.ticks
		e.layers = map[string]bool{} // a new picture starts with its layers off
		e.writer, e.paper = "hana", nil
	case OpWriter:
		e.writer = n.Name
	case OpPaper:
		for range int(n.Number) { // empty lines are blank rows, so the words land on their row
			e.paper = append(e.paper, "")
		}
		e.paper = append(e.paper, n.Text)
	case OpStill:
		e.still = n.Name
	case OpBars, OpPunch, OpPush, OpPan:
		e.execCinema(n)
	case OpLayer:
		for _, p := range n.Parts {
			if p.Key == "+" {
				e.layers[p.Value] = true
			} else {
				delete(e.layers, p.Value)
			}
		}
	case OpProp:
		if n.Name == "clear" {
			e.props = map[string]bool{}
		} else {
			e.props[n.Name] = true
		}
	case OpShow:
		e.hopFor = ""
		e.visible, e.faces = n.Name, false
	case OpHide:
		e.hopFor = ""
		e.visible, e.faces = "", false
	case OpFaces:
		e.hopFor = ""
		e.visible, e.faces = "", true
		e.fx["faces"] = 0
	case OpPose:
		e.setPose(n)
	case OpFx:
		switch n.Name {
		case "clear":
			e.fx = map[string]int{}
			return
		case "wash":
			if _, on := e.fx["wash"]; on {
				return // already washing: do not start over
			}
		case "dark":
			delete(e.fx, "dusk")
		case "dusk":
			delete(e.fx, "dark")
		case "light":
			delete(e.fx, "dark")
			delete(e.fx, "dusk")
			return
		case "figure_gone":
			delete(e.fx, "figure")
			return
		case "redscale":
			delete(e.fx, "redscale_off")
		case "redscale_off":
			if _, on := e.fx["redscale"]; !on {
				return // nothing red to drain
			}
		case "flat_line":
			delete(e.fx, "line_back")
		case "line_back":
			delete(e.fx, "flat_line")
		}
		e.fx[n.Name] = 0
	case OpSfx:
		e.LastSFX = n.Name
		e.cues = append(e.cues, Cue{"sfx", n.Name})
	case OpMusic:
		e.LastMusic = n.Name
		e.cues = append(e.cues, Cue{"music", n.Name})
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
	e.text, e.shown = []rune(strings.ReplaceAll(n.Text, "{player}", e.Player)), 0
	e.autoLine, e.autoNext, e.doneFor, e.autoWait = e.autoNext, false, 0, 0
	if n.Op == OpThink {
		e.hopFor = "" // never a hop on a thought; the next spoken line hops again
		return
	}
	e.speaker = n.Name
	if n.Name == "note" || n.Name == "letter" {
		e.paper = append(e.paper, string(e.text)) // written on the paper as it types
	}
	if IsCharacter(n.Name) {
		if e.hopFor != n.Name {
			e.hopStart = e.ticks
		}
		e.hopFor = n.Name
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
		// A one-shot effect that has played out is forgotten, so the dump and
		// the painter do not carry it on (the persistent ones stay: redscale,
		// dim, the closed eyelids).
		if length, shot := oneShotFX[name]; shot && e.fx[name] > length {
			delete(e.fx, name)
		}
	}
	if e.autoLine && e.LineDone() && len(e.options) == 0 {
		// A line that types itself, then goes on after a short beat.
		if e.doneFor++; e.doneFor > autoBeat {
			return e.Advance()
		}
		return nil
	}
	if e.AutoPlay && e.HasLine() && e.LineDone() && len(e.options) == 0 && !e.locked && !e.fxRunning() {
		if e.autoWait++; e.autoWait >= AutoDelay(len(e.text)) {
			return e.Advance()
		}
		return nil
	}
	if e.holding {
		// A hold has no timer: only AUTO ends it, after holdAutoTicks.
		if e.holdTick++; e.AutoPlay && e.holdTick >= holdAutoTicks {
			return e.leaveWait()
		}
		return nil
	}
	if e.waiting {
		e.waitLeft--
		if e.waitLeft <= 0 {
			return e.leaveWait()
		}
		return nil
	}
	if e.speaker != "" || e.thinking {
		speed := e.speed
		if e.OnPaper() {
			speed = paperSpeed
		}
		e.shown = min(e.shown+speed/TicksPerSecond, float64(len(e.text)))
	}
	return nil
}

// holdAutoTicks is how long AUTO leaves a hold (a line-less shot) up: 2 s.
const holdAutoTicks = 2 * TicksPerSecond

// Holding says whether a hold is up: a shot with no line that waits for a
// click. The painter shows the continue arrow for it.
func (e *Engine) Holding() bool { return e.holding }

// AutoDelay is how many ticks a finished line waits in AUTO mode: 1.2 s plus
// 0.05 s for every character, never more than 6 s.
func AutoDelay(chars int) int {
	seconds := min(1.2+0.05*float64(chars), 6)
	return int(seconds * TicksPerSecond)
}

// fxRunning says whether a timed effect is still playing: AUTO waits for
// the beat to finish instead of skipping it.
func (e *Engine) fxRunning() bool {
	for name, ticks := range e.fx {
		if length, timed := fxLength[name]; timed && ticks < length {
			return true
		}
	}
	return false
}

func (e *Engine) leaveWait() error {
	e.waiting, e.locked, e.holding = false, false, false
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
	case e.locked:
		return nil // the game is "asleep": clicks do nothing
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
			e.finishEffects()
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

// finishEffects lets every timed effect that is running play out at once, so
// a backdoor jump lands on the settled picture (open eyes, risen red) and
// not on the first frame of a fade.
func (e *Engine) finishEffects() {
	e.bars.dur, e.zoom.dur, e.pan.dur = 0, 0, 0
	for name := range e.fx {
		if length, timed := fxLength[name]; timed {
			e.fx[name] = length
		}
	}
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

// HasLine says whether the text box is on the screen: a line is being
// shown (or, in a menu, stays shown above the buttons).
func (e *Engine) HasLine() bool {
	return !e.ended && !e.waiting && len(e.text) > 0
}

// GirlShown says whether the standing girl is drawn: she is not in front of
// a paper held up to the camera, nor in a CG that is its own picture.
func (e *Engine) GirlShown() bool {
	def, ok := bgTable[e.bg]
	return e.visible != "" && !e.faces && !e.OnPaperBG() && !(ok && def.cg)
}

// OnPaperBG says whether the picture is a sheet of paper held up to the
// camera (the note, the letter or the report).
func (e *Engine) OnPaperBG() bool { return e.bg == "note" || e.bg == "letter" || e.bg == "report" }

// LastWriter and LastPaper are the hand and the lines of the paper that was
// on screen before the last cut away from it.
func (e *Engine) LastWriter() string  { return e.lastWriter }
func (e *Engine) LastPaper() []string { return e.lastPaper }

// CutToStill is what an ending does when its script is over and the title
// comes back: if the script named a "still", that background is up, and the
// picture is the same as at the end of the script otherwise. It reports
// whether there was one.
func (e *Engine) CutToStill() bool {
	if e.still == "" {
		return false
	}
	e.bg, e.layers = e.still, map[string]bool{}
	e.visible = ""
	return true
}

// DayLook says whether the picture is a day picture (the report on the
// desk) that is shown through the day's CRT.
func (e *Engine) DayLook() bool { return e.bg == "report" }

// Writer is whose hand writes on the paper: "hana" or "mayu".
func (e *Engine) Writer() string { return e.writer }

// Paper is every line written on the paper so far, the one being typed last.
func (e *Engine) Paper() []string { return e.paper }

// OnPaper says whether the line on screen is written on the paper (a Note:
// or Letter: line): the text box stays away and the pen writes it.
func (e *Engine) OnPaper() bool {
	return e.HasLine() && !e.thinking && (e.speaker == "note" || e.speaker == "letter")
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

// RedscaleAmounts are the numbers redscale.kage takes: how far the red has
// risen from the floor and how far it has drained from the top. Both are 0
// when the effect is off.
func (e *Engine) RedscaleAmounts() (rise, drain float64) {
	if _, on := e.fx["redscale"]; !on || e.OnPaperBG() {
		return 0, 0 // the paper in your hands is not part of the red room
	}
	return e.FxProgress("redscale"), e.FxProgress("redscale_off")
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

// BGTicks is how many ticks the current background has been up.
func (e *Engine) BGTicks() int { return e.ticks - e.bgAt }

// HasLayer says whether a CG layer is switched on by the script.
func (e *Engine) HasLayer(name string) bool { return e.layers[name] }

// Flags is the shared store.
func (e *Engine) Flags() Flags { return e.flags }

// Name is the script's name.
func (e *Engine) Name() string { return e.script.Name }

// State is a snapshot for the agent dump.
type State struct {
	Script    string     `json:"script"`
	Line      int        `json:"line"`
	Speaker   string     `json:"speaker"`
	Text      string     `json:"text"`
	Thinking  bool       `json:"thinking"`
	LineDone  bool       `json:"lineDone"`
	Options   []string   `json:"options"`
	Hover     int        `json:"hover"`
	BG        string     `json:"bg"`
	Visible   string     `json:"visible"`
	Faces     bool       `json:"faces"`
	Layers    []string   `json:"poseLayers"`
	Props     []string   `json:"props"`
	CGLayers  []string   `json:"cgLayers"`
	Effects   []string   `json:"effects"`
	Ended     bool       `json:"ended"`
	EndsTo    string     `json:"endsTo"`
	Auto      bool       `json:"auto"`
	Holding   bool       `json:"holding"`
	Bars      float64    `json:"bars"`      // letterbox bar height now, in pixels
	Cinematic bool       `json:"cinematic"` // the bars are up: subtitles instead of the text box
	Zoom      float64    `json:"zoom"`      // camera scale (1 = the whole picture)
	ZoomAt    [2]float64 `json:"zoomAt"`    // the centre of the zoom
	PanX      float64    `json:"panX"`      // camera x on a wide picture
	Writer    string     `json:"writer"`    // whose hand is on the paper
	Paper     []string   `json:"paper"`     // the lines written on it so far
	LastSFX   string     `json:"sfx"`
	LastMusic string     `json:"music"`
}

// Snapshot describes the engine for the dump.
func (e *Engine) Snapshot() State {
	st := State{
		Script: e.script.Name, Line: e.Line(), Speaker: e.speaker,
		Text: e.Shown(), Thinking: e.thinking, LineDone: e.LineDone(),
		Options: []string{}, Hover: e.hover, BG: e.bg, Visible: e.visible,
		Faces: e.faces, Layers: []string{}, Props: []string{}, CGLayers: []string{}, Effects: []string{},
		Bars: e.BarsHeight(), Cinematic: e.Cinematic(), PanX: e.PanX(), Writer: e.writer, Paper: append([]string{}, e.paper...), Ended: e.ended, EndsTo: e.endsTo, Auto: e.AutoPlay, Holding: e.holding, LastSFX: e.LastSFX, LastMusic: e.LastMusic,
	}
	st.Zoom, st.ZoomAt[0], st.ZoomAt[1] = e.Zoom()
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
	for name := range e.layers {
		st.CGLayers = append(st.CGLayers, name)
	}
	sort.Strings(st.CGLayers)
	for name := range e.fx {
		st.Effects = append(st.Effects, name)
	}
	sort.Strings(st.Props)
	sort.Strings(st.Effects)
	return st
}

// HopOffset is how many pixels the speaker is lifted right now (0 or
// negative; the Prologue's hanging faces hop the other way). A pure function
// of the ticks since the hop began.
func (e *Engine) HopOffset() int {
	i := e.ticks - e.hopStart
	if e.hopStart < 0 || i < 0 || i >= len(hopTable) {
		return 0
	}
	return hopTable[i]
}

// FaceHop is the hop of one of the Prologue's hanging faces: downward
// (positive), and only for the one who is talking.
func (e *Engine) FaceHop(character string) int {
	if e.hopFor != character {
		return 0
	}
	return -e.HopOffset()
}

// BGNames lists the backgrounds the running script uses (the dump compares
// them with the art that exists).
func (e *Engine) BGNames() []string { return e.script.BGNames() }
