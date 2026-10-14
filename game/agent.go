package game

// The agent backdoor (Architecture -> Agent backdoor). A set of command-line
// flags that only work when -agent is passed, so a tester can jump straight
// to a scene, play a script of inputs, and look at the result as a PNG and a
// json dump. Everything here uses things the game already has (the scene
// manager, the clock, the ward). It never touches hospital/.

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"danmingame/night"
)

// Options are the command-line flags. Without Enabled, all but Enabled are
// left at their zero value and the game ignores them.
type Options struct {
	Enabled bool
	Scene   string
	Day     int
	Clock   string
	Seed    int64
	Flags   string // story flags, comma separated
	Auto    bool   // the VN's AUTO play starts on
	Line    int    // night scenes: start at this line of the script file
	Speed   float64
	Input   string // path of the input script
	Shot    string // 1x PNG, written when the run ends
	Shot4   string // x4 PNG, written when the run ends
	Dump    string // json state, written when the run ends
	NoCRT   bool
	// SeenPrologue pretends the Prologue was seen before: Esc skips it.
	SeenPrologue bool
	PowerOn      bool // start the day with the CRT turn-on playing
	// ReduceFlashing is a player option (Settings -> Reduce flashing), so it
	// works without -agent.
	ReduceFlashing bool
	Quit           bool
	Frames         int
	Pan            int    // camera x at the start (used only when PanSet)
	PanSet         bool   // -pan was given; without it the camera starts at 120
	Scores         string // the scores file (without it scores stay in memory)
	Score          int    // today's score at the start (day.top10 saves it)
	Sleepy         bool   // patients fall asleep within a second and sleep long
	Dark           int    // -dark N: draw every guest N palette steps darker (0-3), -1 = as the game decides

	// Fullscreen is an ordinary option for players, so it also works
	// without -agent.
	Fullscreen bool
}

// ParseOptions reads the flags. Every flag except -agent, -fullscreen, -nocrt and -reduce-flashing is
// ignored (and reset) when -agent is missing, so a player can't trip the backdoor.
func ParseOptions(args []string) (Options, error) {
	var o Options
	fs := flag.NewFlagSet("danmin", flag.ContinueOnError)
	fs.BoolVar(&o.Enabled, "agent", false, "turn on test mode (the other flags need it)")
	fs.StringVar(&o.Scene, "scene", SceneTitle, "scene id to start in")
	fs.IntVar(&o.Day, "day", 1, "day number")
	fs.StringVar(&o.Clock, "clock", "", "in-game time, HH:MM")
	fs.Int64Var(&o.Seed, "seed", 0, "random seed (0 = random)")
	fs.StringVar(&o.Flags, "flags", "", "story flags, comma separated")
	fs.BoolVar(&o.Auto, "auto", false, "night scenes: start with AUTO play on")
	fs.IntVar(&o.Line, "line", 0, "night scenes: start at this line of the script file")
	fs.Float64Var(&o.Speed, "speed", 1, "game time multiplier")
	fs.StringVar(&o.Input, "input", "", "input script file")
	fs.StringVar(&o.Shot, "shot", "", "save a 1x PNG screenshot when the run ends")
	fs.StringVar(&o.Shot4, "shot4", "", "save a x4 PNG screenshot when the run ends")
	fs.StringVar(&o.Dump, "dump", "", "save the game state as json when the run ends")
	fs.BoolVar(&o.NoCRT, "nocrt", false, "turn the day CRT filter off")
	fs.BoolVar(&o.ReduceFlashing, "reduce-flashing", false, "replace the day-to-night strobe by one slow invert-and-fade")
	fs.BoolVar(&o.SeenPrologue, "seen-prologue", false, "the Prologue counts as seen before: Esc skips it (the cold open)")
	fs.BoolVar(&o.PowerOn, "poweron", false, "play the CRT turn-on at the start (see it with wait N and shot4)")
	fs.BoolVar(&o.Quit, "quit", false, "exit when the script (or -frames) ends")
	fs.IntVar(&o.Frames, "frames", 0, "end the run after N frames")
	fs.IntVar(&o.Pan, "pan", 0, "camera x at the start, 0-240 (default 120, the centre)")
	fs.StringVar(&o.Scores, "scores", "", "scores file to use (default: none, kept in memory)")
	fs.IntVar(&o.Score, "score", 0, "today's score at the start")
	fs.IntVar(&o.Dark, "dark", -1, "draw every guest N palette steps darker, 0-3 (to judge the Severe darkening)")
	fs.BoolVar(&o.Sleepy, "sleepy", false, "patients fall asleep fast and sleep long")
	fs.BoolVar(&o.Fullscreen, "fullscreen", false, "start in fullscreen (F11 or Alt+Enter toggles it)")
	if err := fs.Parse(args); err != nil {
		return Options{}, err
	}
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "pan" {
			o.PanSet = true
		}
	})
	if !o.Enabled {
		return Options{Fullscreen: o.Fullscreen, NoCRT: o.NoCRT, ReduceFlashing: o.ReduceFlashing}, nil
	}
	return o, nil
}

// Command is one line of an input script.
type Command struct {
	Op    string // "next", "choose", "wait", "key", "click", "mouse", "down", "up", "pan", "dark", "advance", "type", "stamp", "card", "shot" or "dump"
	Text  string // type: the characters
	N     int    // wait: frames; pan: camera x
	Stamp Stamp  // stamp: which one
	Key   ebiten.Key
	X, Y  int    // click, mouse: 1x game coordinates
	Path  string // shot, dump: output file
}

// ParseScript reads an input script, one command per line:
//
//	wait 30         frames
//	key Space
//	click 212 140   1x game coordinates
//	mouse 5 135     hold the cursor there (edge of the screen pans the camera)
//	type marcus     type text (the RECORDS tab's name field)
//	dark 2          draw every guest 2 palette steps darker (0-3); -dark N does it from the start
//	advance 120     run the shift 120 game seconds at once (queue, clock, ward)
//	down 320 210    press the left button there and hold it (mouse moves it)
//	up              let the button go
//	drag 318 213 200 213   down, move, up: drag the page corner left
//	next            advance the night text (the Space key: first shows the whole line, then goes on)
//	auto [on|off]   night: flip AUTO play (the A key) or set it
//	skipstrobe      end the day-to-night strobe at once
//	choose 2        pick option 2 of the open night menu
//	pan 0           put the camera at x = 0 (0 to 240)
//	desk            look down to the desk (reception: look back up)
//	folder          open the folder on the desk
//	turn            turn the open folder's page
//	page sleeplog   show a page at once (referral, sleeplog, wristband)
//	lift admit      lift a stamp off the desk into the hand
//	apply           press the stamp in the hand on the open referral
//	putback         put the stamp in the hand back on the desk
//	stamp admit     use a stamp on the window patient (admit, refer, discharge)
//	shot out.png    (1x)
//	shot4 out.png   (x4, to compare with the Figma views)
//	dump out.json
//
// Everything after '#' is a comment. Errors name the line.
func ParseScript(r io.Reader) ([]Command, error) {
	var cmds []Command
	scanner := bufio.NewScanner(r)
	for lineNo := 1; scanner.Scan(); lineNo++ {
		line := scanner.Text()
		if i := strings.Index(line, "#"); i >= 0 {
			line = line[:i]
		}
		words := strings.Fields(line)
		if len(words) == 0 {
			continue
		}
		if words[0] == "drag" {
			steps, err := parseDrag(words[1:])
			if err != nil {
				return nil, fmt.Errorf("script line %d: %w", lineNo, err)
			}
			cmds = append(cmds, steps...)
			continue
		}
		cmd, err := parseCommand(words)
		if err != nil {
			return nil, fmt.Errorf("script line %d: %w", lineNo, err)
		}
		cmds = append(cmds, cmd)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading script: %w", err)
	}
	return cmds, nil
}

// parseDrag turns "drag X1 Y1 X2 Y2" into the presses it stands for: the
// button goes down at the first point, the cursor moves to the middle and
// then the end (one frame each), and the button comes up.
func parseDrag(args []string) ([]Command, error) {
	if len(args) != 4 {
		return nil, fmt.Errorf("drag needs x1 y1 x2 y2")
	}
	var n [4]int
	for i, a := range args {
		v, err := strconv.Atoi(a)
		if err != nil {
			return nil, fmt.Errorf("drag %q: want four numbers", a)
		}
		n[i] = v
	}
	return []Command{
		{Op: "down", X: n[0], Y: n[1]},
		{Op: "mouse", X: (n[0] + n[2]) / 2, Y: (n[1] + n[3]) / 2},
		{Op: "wait", N: 1},
		{Op: "mouse", X: n[2], Y: n[3]},
		{Op: "wait", N: 1},
		{Op: "up"},
	}, nil
}

// parseXY reads the two numbers of "down X Y".
func parseXY(cmd Command, args []string) (Command, error) {
	if len(args) != 2 {
		return cmd, fmt.Errorf("%s needs x and y", cmd.Op)
	}
	x, errX := strconv.Atoi(args[0])
	y, errY := strconv.Atoi(args[1])
	if errX != nil || errY != nil {
		return cmd, fmt.Errorf("%s %q %q: want two numbers", cmd.Op, args[0], args[1])
	}
	cmd.X, cmd.Y = x, y
	return cmd, nil
}

// scriptKeys are the keys the game listens to, by the name a script uses.
// Ebitengine accepts more names (and aliases like "Down"), but a name that
// no scene reads would do nothing, so the script parser rejects it.
var scriptKeys = map[string]ebiten.Key{
	"A": ebiten.KeyA, "B": ebiten.KeyB, "F": ebiten.KeyF,
	"Alt": ebiten.KeyAlt, "Meta": ebiten.KeyMeta,
	"ArrowUp": ebiten.KeyArrowUp, "ArrowDown": ebiten.KeyArrowDown,
	"ArrowLeft": ebiten.KeyArrowLeft, "ArrowRight": ebiten.KeyArrowRight,
	"Backspace": ebiten.KeyBackspace, "Enter": ebiten.KeyEnter,
	"Escape": ebiten.KeyEscape, "Space": ebiten.KeySpace, "Tab": ebiten.KeyTab,
	"F10": ebiten.KeyF10, "F11": ebiten.KeyF11,
}

// scriptKeyNames lists the valid key names, sorted, for the error message.
func scriptKeyNames() string {
	names := make([]string, 0, len(scriptKeys))
	for name := range scriptKeys {
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

func parseCommand(words []string) (Command, error) {
	cmd := Command{Op: words[0]}
	args := words[1:]
	switch cmd.Op {
	case "wait":
		if len(args) != 1 {
			return cmd, fmt.Errorf("wait needs one number of frames")
		}
		n, err := strconv.Atoi(args[0])
		if err != nil || n < 0 {
			return cmd, fmt.Errorf("wait %q: want a number of frames", args[0])
		}
		cmd.N = n
	case "key":
		if len(args) != 1 {
			return cmd, fmt.Errorf("key needs one key name, like Space")
		}
		k, ok := scriptKeys[args[0]]
		if !ok {
			return cmd, fmt.Errorf("key %q: unknown key name, want one of %s", args[0], scriptKeyNames())
		}
		cmd.Key = k
	case "down":
		return parseXY(cmd, args)
	case "up":
		if len(args) != 0 {
			return cmd, fmt.Errorf("up takes no argument")
		}
	case "click", "mouse":
		if len(args) != 2 {
			return cmd, fmt.Errorf("%s needs x and y", cmd.Op)
		}
		x, errX := strconv.Atoi(args[0])
		y, errY := strconv.Atoi(args[1])
		if errX != nil || errY != nil {
			return cmd, fmt.Errorf("%s %q %q: want two numbers", cmd.Op, args[0], args[1])
		}
		cmd.X, cmd.Y = x, y
	case "pan":
		if len(args) != 1 {
			return cmd, fmt.Errorf("pan needs one camera x")
		}
		n, err := strconv.Atoi(args[0])
		if err != nil || n < 0 || n > MaxCameraX {
			return cmd, fmt.Errorf("pan %q: want a number from 0 to %d", args[0], MaxCameraX)
		}
		cmd.N = n
	case "type":
		if len(args) == 0 {
			return cmd, fmt.Errorf("type needs some text")
		}
		cmd.Text = strings.Join(args, " ")
	case "dark":
		if len(args) != 1 {
			return cmd, fmt.Errorf("dark needs a number of steps")
		}
		n, err := strconv.Atoi(args[0])
		if err != nil || n < 0 || n > maxDarkness {
			return cmd, fmt.Errorf("dark %q: want a number from 0 to %d", args[0], maxDarkness)
		}
		cmd.N = n
	case "advance":
		if len(args) != 1 {
			return cmd, fmt.Errorf("advance needs one number of game seconds")
		}
		n, err := strconv.Atoi(args[0])
		if err != nil || n < 0 {
			return cmd, fmt.Errorf("advance %q: want a number of game seconds", args[0])
		}
		cmd.N = n
	case "stamp", "lift":
		if len(args) != 1 {
			return cmd, fmt.Errorf("%s needs admit, refer or discharge", cmd.Op)
		}
		switch strings.ToLower(args[0]) {
		case "admit":
			cmd.Stamp = StampAdmit
		case "refer":
			cmd.Stamp = StampRefer
		case "discharge":
			cmd.Stamp = StampDischarge
		default:
			return cmd, fmt.Errorf("%s %q: want admit, refer or discharge", cmd.Op, args[0])
		}
	case "page":
		if len(args) != 1 {
			return cmd, fmt.Errorf("page needs referral, sleeplog, wristband or report")
		}
		switch strings.ToLower(args[0]) {
		case "referral":
			cmd.N = int(PageReferral)
		case "sleeplog":
			cmd.N = int(PageSleepLog)
		case "wristband":
			cmd.N = int(PageWristband)
		case "report":
			cmd.N = int(PageReport)
		default:
			return cmd, fmt.Errorf("page %q: want referral, sleeplog, wristband or report", args[0])
		}
	case "auto": // auto, auto on, auto off
		switch {
		case len(args) == 0:
			cmd.N = -1 // toggle
		case len(args) == 1 && args[0] == "on":
			cmd.N = 1
		case len(args) == 1 && args[0] == "off":
			cmd.N = 0
		default:
			return cmd, fmt.Errorf("auto takes nothing, on or off")
		}
	case "next", "skipstrobe", "card":
		if len(args) != 0 {
			return cmd, fmt.Errorf("%s takes no argument", cmd.Op)
		}
	case "choose":
		if len(args) != 1 {
			return cmd, fmt.Errorf("choose needs the number of the option (1, 2 or 3)")
		}
		n, err := strconv.Atoi(args[0])
		if err != nil || n < 1 || n > 3 {
			return cmd, fmt.Errorf("choose %q: want 1, 2 or 3", args[0])
		}
		cmd.N = n
	case "folder", "turn", "apply", "putback", "desk", "reception":
		if len(args) != 0 {
			return cmd, fmt.Errorf("%s takes no argument", cmd.Op)
		}
	case "shot", "shot4", "dump":
		if len(args) != 1 {
			return cmd, fmt.Errorf("%s needs one file name", cmd.Op)
		}
		cmd.Path = args[0]
	default:
		return cmd, fmt.Errorf("unknown command %q", cmd.Op)
	}
	return cmd, nil
}

// output is a file the agent still has to write after this frame is drawn.
type output struct {
	kind string // "shot", "shot4" or "dump"
	path string
}

// Agent plays the script and writes the screenshots and dumps.
type Agent struct {
	opts    Options
	script  []Command
	hasFile bool // an -input script was given
	next    int  // index of the next command to run
	waiting int  // frames left to wait
	frame   int  // frames updated so far

	pending  []output // written in Draw, once the frame is on the screen
	ended    bool     // the run has reached its end (final outputs queued)
	quitNext bool
	err      error // a write failed; Update returns it on the next frame
}

// newAgent applies the options to the env (day, clock, flags, speed, seed)
// and reads the input script.
func newAgent(opts Options, env *Env) (*Agent, error) {
	if opts.Day > 0 {
		if _, err := RulesForDay(opts.Day); err != nil {
			return nil, fmt.Errorf("-day: %w", err)
		}
		env.Day = opts.Day
	}
	if opts.Clock != "" {
		clock, err := ParseClock(opts.Clock)
		if err != nil {
			return nil, err
		}
		env.Clock = clock
		env.KeepClock = true
	}
	env.ScoresPath = opts.Scores
	env.Sleepy = opts.Sleepy
	if opts.Dark > maxDarkness {
		return nil, fmt.Errorf("-dark %d: want 0 to %d", opts.Dark, maxDarkness)
	}
	env.Dark = opts.Dark
	if opts.Score > 0 {
		env.Scores[fmt.Sprintf("day%d", env.Day)] = opts.Score
	}
	if opts.Seed != 0 {
		env.Seed = opts.Seed
	}
	if opts.Speed > 0 {
		env.Speed = opts.Speed
	}
	for _, name := range strings.Split(opts.Flags, ",") {
		if name = strings.TrimSpace(name); name != "" {
			env.Flags[name] = true
		}
	}
	env.NightLine = opts.Line
	env.Auto = opts.Auto
	env.CRT = !opts.NoCRT
	if opts.PowerOn {
		env.StartPowerOn()
	}
	env.ReduceFlashing = opts.ReduceFlashing
	env.Input.IgnoreReal = true
	if opts.PanSet {
		if opts.Pan > MaxCameraX {
			return nil, fmt.Errorf("-pan %d: want a number from 0 to %d", opts.Pan, MaxCameraX)
		}
		env.CameraX = float64(opts.Pan)
	}

	a := &Agent{opts: opts}
	if opts.Input != "" {
		f, err := os.Open(opts.Input)
		if err != nil {
			return nil, fmt.Errorf("input script: %w", err)
		}
		defer f.Close()
		if a.script, err = ParseScript(f); err != nil {
			return nil, err
		}
		a.hasFile = true
	}
	return a, nil
}

// Update runs before the scene's Update. It feeds the script's inputs in
// and decides when the run is over.
func (a *Agent) Update(g *Game) error {
	if a.err != nil {
		return a.err
	}
	if a.quitNext && len(a.pending) == 0 {
		return ebiten.Termination
	}
	a.frame++
	if a.waiting > 0 {
		a.waiting--
	} else {
		a.runCommands(g.env)
	}
	if !a.ended && a.finished() {
		a.ended = true
		a.queueFinalOutputs()
		a.quitNext = a.opts.Quit
	}
	return nil
}

// runCommands runs script commands until one needs a frame to itself.
func (a *Agent) runCommands(env *Env) {
	for a.next < len(a.script) {
		cmd := a.script[a.next]
		a.next++
		switch cmd.Op {
		case "wait":
			if cmd.N > 0 {
				a.waiting = cmd.N - 1 // this frame is the first of the wait
				return
			}
		case "key":
			env.Input.InjectKey(cmd.Key)
			return
		case "click":
			env.Input.InjectClick(cmd.X, cmd.Y)
			return
		case "mouse":
			env.Input.InjectMouse(cmd.X, cmd.Y)
		case "down":
			env.Input.InjectDown(cmd.X, cmd.Y)
			return
		case "up":
			env.Input.InjectUp()
			return
		case "type":
			env.Input.InjectText(cmd.Text)
			return
		case "dark":
			if sh := env.Shift(); sh != nil {
				sh.ForceDark(cmd.N)
			}
		case "advance":
			advance(env, cmd.N)
		case "pan":
			env.CameraX = float64(cmd.N)
		case "next": // the same Space press a player would make
			env.Input.InjectKey(ebiten.KeySpace)
			return
		case "auto": // the AUTO toggle, as the A key does
			if cmd.N < 0 {
				env.Input.InjectKey(ebiten.KeyA)
				return
			}
			env.Auto = cmd.N == 1
			if env.Night != nil {
				env.Night.AutoPlay = env.Auto
			}
		case "skipstrobe": // the strobe scene ends on its next frame
			env.SkipStrobe = true
		case "card": // an ending scene jumps from its script to its card
			env.SkipToCard = true
		case "choose":
			if env.Night != nil {
				_ = env.Night.Choose(cmd.N - 1) // no menu open: nothing happens, as for a click
			}
			return
		case "desk":
			env.SwitchTo(SceneDayDesk)
			return
		case "reception":
			env.SwitchTo(SceneDayReception)
			return
		case "folder", "turn", "page", "lift", "apply", "putback":
			if sh := env.Shift(); sh != nil {
				runDeskCommand(sh, cmd) // a click that does nothing is no error, as for a person
			}
		case "stamp":
			if sh := env.Shift(); sh != nil {
				_ = sh.Stamp(cmd.Stamp) // nobody at the window: nothing happens, as for a click
			}
		case "shot", "shot4":
			a.pending = append(a.pending, output{cmd.Op, cmd.Path})
			return // the commands after it run next frame, so it shows the state before them
		case "dump":
			a.pending = append(a.pending, output{"dump", cmd.Path})
			return
		}
	}
}

// runDeskCommand does what the desk's clicks do: open the folder, turn the
// page, lift a stamp, press it on the referral. Errors are dropped, like the
// clicks that do nothing.
func runDeskCommand(sh *Shift, cmd Command) {
	switch cmd.Op {
	case "folder":
		_ = sh.OpenFolder()
	case "turn":
		_ = sh.TurnPage()
	case "page":
		_ = sh.GoToPage(Page(cmd.N))
	case "lift":
		sh.LiftStamp(cmd.Stamp)
	case "apply":
		_ = sh.ApplyStamp()
	case "putback":
		sh.PutBackStamp()
	}
}

// deskSetup puts a folder on the desk for the scene ids "day.desk.*": it runs
// the shift until the first guest stands at the window and their folder has
// slid in, then opens it on a page if asked.
func deskSetup(open bool, page Page) func(*Shift) error {
	return func(sh *Shift) error {
		for range 1200 {
			if sh.FolderReady() {
				break
			}
			sh.Update()
		}
		if !sh.FolderReady() {
			return errors.New("desk scene: no folder arrived on the desk")
		}
		if !open {
			return nil
		}
		return sh.GoToPage(page)
	}
}

// advance runs the shift (and the clock) seconds game seconds at once, so
// a test can reach the middle of a day without waiting for it. The patients'
// own timers run in real time and do not jump.
func advance(env *Env, seconds int) {
	sh := env.Shift()
	if sh == nil {
		return
	}
	for range seconds * 60 {
		env.Clock.Advance(env.ClockRate * env.Speed / 60)
		sh.Update()
	}
}

// finished says whether the run has reached its end: -frames frames have
// passed, or the script is done, or (with neither given) one frame has run.
func (a *Agent) finished() bool {
	scriptDone := a.next >= len(a.script) && a.waiting == 0
	switch {
	case a.opts.Frames > 0 && a.frame >= a.opts.Frames:
		return true
	case a.opts.Frames > 0:
		return false
	case a.hasFile:
		return scriptDone
	}
	return a.frame >= 1
}

func (a *Agent) queueFinalOutputs() {
	if a.opts.Shot != "" {
		a.pending = append(a.pending, output{"shot", a.opts.Shot})
	}
	if a.opts.Shot4 != "" {
		a.pending = append(a.pending, output{"shot4", a.opts.Shot4})
	}
	if a.opts.Dump != "" {
		a.pending = append(a.pending, output{"dump", a.opts.Dump})
	}
}

// AfterDraw writes the queued files. Screenshots need the finished frame,
// which is why this runs at the end of Draw and not in Update.
func (a *Agent) AfterDraw(g *Game, screen *ebiten.Image) {
	for _, out := range a.pending {
		var err error
		switch out.kind {
		case "shot":
			err = writePNG(out.path, readScreen(screen), 1)
		case "shot4":
			// With the CRT on the shot is the filtered picture, so the
			// shader can be checked; -nocrt gives the plain picture.
			if g.crtOn() {
				err = writePNG(out.path, readScreen(g.shotCRT.Apply(screen, 4, g.crtParams())), 1)
			} else {
				err = writePNG(out.path, readScreen(screen), 4)
			}
		case "dump":
			err = writeDump(out.path, BuildDump(g.env, g.scenes.ID(), a.frame))
		}
		if err != nil && a.err == nil {
			a.err = err
		}
	}
	a.pending = nil
}

// readScreen copies the screen's pixels into an ordinary image.
func readScreen(screen *ebiten.Image) *image.RGBA {
	b := screen.Bounds()
	img := image.NewRGBA(b)
	screen.ReadPixels(img.Pix)
	return img
}

// scaleUp repeats every pixel k times in each direction (nearest-neighbour),
// for the x4 screenshots that line up with the Figma views.
func scaleUp(src *image.RGBA, k int) *image.RGBA {
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx()*k, b.Dy()*k))
	for y := 0; y < dst.Bounds().Dy(); y++ {
		for x := 0; x < dst.Bounds().Dx(); x++ {
			dst.SetRGBA(x, y, src.RGBAAt(b.Min.X+x/k, b.Min.Y+y/k))
		}
	}
	return dst
}

func writePNG(path string, img *image.RGBA, scale int) error {
	if scale > 1 {
		img = scaleUp(img, scale)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("screenshot: %w", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("screenshot %s: %w", path, err)
	}
	return nil
}

// Dump is the game state written by -dump, as json.
type Dump struct {
	Scene    string         `json:"scene"`
	Day      int            `json:"day"`
	Clock    string         `json:"clock"`
	Seed     int64          `json:"seed"`
	Frame    int            `json:"frame"`
	CameraX  int            `json:"cameraX"`
	Patients []DumpPatient  `json:"patients"`
	Rooms    []DumpRoom     `json:"rooms"`
	Scores   map[string]int `json:"scores"`
	Flags    []string       `json:"flags"`

	// The day at the desk (empty when no shift is running).
	Score     int           `json:"score"`
	Served    int           `json:"served"`
	Total     int           `json:"total"`
	ShiftSecs float64       `json:"shiftSeconds"`
	ForceDark int           `json:"forceDark"` // -dark / dark: steps every guest is drawn darker, -1 when off
	Queue     []DumpGuest   `json:"queue"`
	Stamps    []DumpStamp   `json:"stamps"`
	Sleepers  []DumpSleeper `json:"sleepers"`
	Hallway   []string      `json:"hallway"` // sleepers with no bed, by name
	Staff     []DumpStaff   `json:"staff"`
	Alerts    []string      `json:"alerts"`
	Message   string        `json:"message"`
	Dozing    string        `json:"dozing"` // the day 1 guest asleep at the window, "" otherwise
	Tab       string        `json:"computerTab"`
	Top10     []DumpScore   `json:"top10"` // as the top-10 scene showed it
	ShiftOver bool          `json:"shiftOver"`
	Desk      DumpDesk      `json:"desk"`
	CallLight bool          `json:"callLight"` // a Severe patient sleeps unattended
	Banner    string        `json:"banner"`    // the reception's top bar: the message or the nurse call

	// The route and the ending (from the flags), the ending's phase, the
	// secret ending's file, and the backgrounds that still have no art.
	Route        string   `json:"route"`
	Ending       string   `json:"ending"`
	EndingPhase  string   `json:"endingPhase"`
	View         View     `json:"view"` // where the picture sits on the screen (fullscreen fit)
	MusicCue     string   `json:"musicCue"`
	EndingGlitch float64  `json:"endingGlitch"` // the secret ending's glitch strength, 0 to 1
	Dream        bool     `json:"dream"`        // the Prologue's washed-out filter is on
	Settings     Settings `json:"settings"`     // the settings the game runs with
	SleepLogPath string   `json:"sleepLogPath"`
	SleepLogErr  string   `json:"sleepLogError"`
	Placeholders []string `json:"placeholders"`
	SoundLog     []string `json:"soundLog"`     // the last 30 sound requests (muted in test mode)
	SoundMissing []string `json:"soundMissing"` // cues asked for that have no file (real playback only)
	// SoundFiles maps each cue in SoundLog to the file the game would play
	// ("" when there is none), so a muted test run still shows what resolves.
	SoundFiles map[string]string `json:"soundFiles"`

	// The night script that is playing (nil in the day).
	Night *night.State `json:"night,omitempty"`

	// The day-to-night strobe (nil unless a transition scene is playing).
	Strobe *DumpStrobe `json:"strobe,omitempty"`
}

// DumpStrobe is where the strobe is: its tick, what it shows, and how inverted.
type DumpStrobe struct {
	Night  int     `json:"night"`
	Tick   int     `json:"tick"`
	Layer  string  `json:"layer"` // day, night, white or fall
	Invert float64 `json:"invert"`
	Length int     `json:"length"` // the whole strobe, in ticks
}

// DumpDesk is the look-down desk: the folder, the page, the stamp in the hand.
type DumpDesk struct {
	Folder       string `json:"folder"` // whose folder is on the desk, "" for none
	State        string `json:"state"`  // none, arriving, closed, opening, open, leaving
	Page         string `json:"page"`   // the page showing when the folder is open
	Turning      bool   `json:"turning"`
	Holding      string `json:"holding"`      // the stamp in the hand, "" for none
	FilesWaiting bool   `json:"filesWaiting"` // the FILES tab shows files
	Pulsing      bool   `json:"pulsing"`      // ... and pulses (day 1 tutorial)
	Visited      bool   `json:"visited"`
	Slip         string `json:"citation"` // the citation slip's text, "" for none
	Ink          bool   `json:"ink"`      // the ink mark shows on the referral
}

// DumpPatient is one patient, with the enums written as text.
type DumpPatient struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Level    string `json:"level"`
	State    string `json:"state"`
	Room     int    `json:"room"` // 0 = no room
	Location string `json:"location"`
}

// DumpRoom is one room.
type DumpRoom struct {
	Number    int      `json:"number"`
	Capacity  int      `json:"capacity"`
	State     string   `json:"state"`
	Occupants []string `json:"occupants"`
}

// DumpGuest is one visitor of the day's queue, with the data the desk
// decides from.
type DumpGuest struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	ReferralOK bool   `json:"referralOK"`
	Problem    string `json:"problem"`
	Returning  bool   `json:"returning"`
	RightStamp string `json:"rightStamp"`
	Look       string `json:"look"`
	Blinking   bool   `json:"blinking"`
	Pips       int    `json:"pips"`     // 1, 2 or 3 dots for Mild, Moderate, Severe
	Episodes   int    `json:"episodes"` // recorded by the hospital today
	Darkness   int    `json:"darkness"` // palette steps darker (Severe patients)
	LogRows    int    `json:"logRows"`  // rows on the sleep log
	X          int    `json:"x"`
	Y          int    `json:"y"`
}

// DumpStamp is one stamp the player used.
type DumpStamp struct {
	Patient string `json:"patient"`
	Stamp   string `json:"stamp"`
	Right   bool   `json:"right"`
}

// DumpSleeper is a patient who fell asleep and has not woken up yet.
type DumpSleeper struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Location string  `json:"location"`
	Attended bool    `json:"attended"`
	Seconds  float64 `json:"secondsAsleep"`
}

// DumpStaff is one attender and whether they are busy.
type DumpStaff struct {
	Name string `json:"name"`
	Busy bool   `json:"busy"`
}

// DumpScore is one row of the top-10.
type DumpScore struct {
	Rank    string `json:"rank"`
	Name    string `json:"name"`
	Score   string `json:"score"`
	Time    string `json:"time"`
	Message string `json:"message"`
}

var lookNames = map[Look]string{
	LookHidden: "hidden", LookStanding: "standing", LookAsleep: "asleep", LookHallway: "hallway",
	LookDozing: "dozing",
}

// BuildDump collects the state. The hospital part comes from Snapshot, so
// it is safe to call while the patient goroutines run. With no shift the
// day's lists are empty.
func BuildDump(env *Env, sceneID string, frame int) Dump {
	d := Dump{
		Scene:    sceneID,
		Day:      env.Day,
		Clock:    env.Clock.String(),
		Seed:     env.Seed,
		Frame:    frame,
		CameraX:  int(env.CameraX),
		Patients: []DumpPatient{},
		Rooms:    []DumpRoom{},
		Scores:   env.Scores,
		Flags:    []string{},
		Queue:    []DumpGuest{},
		Stamps:   []DumpStamp{},
		Sleepers: []DumpSleeper{},
		Hallway:  []string{},
		Staff:    []DumpStaff{},
		Alerts:   []string{},
		Top10:    []DumpScore{},
		Tab:      env.Tab,
	}
	if sh := env.shift; sh != nil {
		snap := sh.ward.Snapshot()
		for _, p := range snap.Patients {
			d.Patients = append(d.Patients, DumpPatient{
				ID: p.ID, Name: p.Name, Level: p.Level.String(),
				State: p.State.String(), Room: p.Room, Location: p.Location,
			})
		}
		for _, r := range snap.Rooms {
			occupants := r.Occupants
			if occupants == nil {
				occupants = []string{}
			}
			d.Rooms = append(d.Rooms, DumpRoom{
				Number: r.Number, Capacity: r.Capacity,
				State: r.State.String(), Occupants: occupants,
			})
		}
		d.addShift(sh)
	}
	for i, r := range env.Ranked {
		row := DumpScore{Rank: fmt.Sprint(i + 1), Name: r.Name, Score: fmt.Sprint(r.Score), Time: r.Time, Message: r.Message}
		if r.Ghost {
			row.Rank, row.Score = "-", "---"
		}
		d.Top10 = append(d.Top10, row)
	}
	for name, on := range env.Flags {
		if on {
			d.Flags = append(d.Flags, name)
		}
	}
	d.View = env.View
	d.Route, d.Ending, d.EndingPhase, d.MusicCue = night.RouteFor(env.Flags), env.Ending, env.EndingPhase, env.MusicCue
	d.SleepLogPath, d.SleepLogErr = env.SleepLogPath, env.SleepLogError
	d.EndingGlitch = env.EndingGlitch
	d.Dream = env.nightArt != nil && env.nightArt.Dream
	d.Settings = env.Settings()
	d.Placeholders = []string{}
	d.SoundLog, d.SoundMissing, d.SoundFiles = []string{}, []string{}, map[string]string{}
	if env.Sound != nil {
		d.SoundLog = append(d.SoundLog, env.Sound.Log[max(0, len(env.Sound.Log)-30):]...)
		d.SoundMissing = append(d.SoundMissing, env.Sound.Missing()...)
		for _, entry := range d.SoundLog {
			if _, cue, ok := strings.Cut(entry, " "); ok && cue != "off" && cue != "fade" {
				d.SoundFiles[cue] = env.Sound.FileName(cue)
			}
		}
	}
	if env.Night != nil {
		st := env.Night.Snapshot()
		d.Night = &st
		if art, err := env.NightArt(); err == nil {
			for _, bg := range env.Night.BGNames() {
				if art.IsPlaceholder(bg) {
					d.Placeholders = append(d.Placeholders, bg)
				}
			}
		}
	}
	if env.StrobeNight > 0 {
		st := StrobeAt(env.StrobeNight, env.StrobeTick, env.ReduceFlashing)
		d.Strobe = &DumpStrobe{Night: env.StrobeNight, Tick: env.StrobeTick, Layer: st.Layer.String(), Invert: st.Invert, Length: StrobeLength(env.StrobeNight)}
	}
	sort.Strings(d.Flags) // map order is random; sorted keeps dumps comparable
	return d
}

// addShift fills in the day-at-the-desk fields.
func (d *Dump) addShift(sh *Shift) {
	d.Score, d.Served, d.Total = sh.Score(), sh.Served(), sh.Total()
	d.ShiftSecs, d.ShiftOver = sh.Seconds(), sh.Over()
	d.Message = sh.Message()
	if g := sh.Dozing(); g != nil {
		d.Dozing = g.Name
	}
	for _, g := range sh.Guests {
		d.Queue = append(d.Queue, DumpGuest{
			ID: g.ID(), Name: g.Name, Status: g.status.String(),
			ReferralOK: g.ReferralOK(), Problem: g.Problem(), Returning: g.Returning,
			RightStamp: g.RightStamp().String(), Look: lookNames[sh.LookOf(g)], Blinking: sh.Blinking(g),
			X: int(g.x), Y: int(g.y),
			Pips: Pips(g.Level), Episodes: sh.EpisodesOf(g), Darkness: sh.Darkness(g), LogRows: len(g.Log),
		})
	}
	d.ForceDark = sh.forceDark
	d.CallLight, d.Banner = sh.SevereCalling(), sh.Banner()
	d.Desk = dumpDesk(sh)
	for _, st := range sh.Stamps() {
		d.Stamps = append(d.Stamps, DumpStamp{Patient: st.Patient, Stamp: st.Stamp.String(), Right: st.Right})
	}
	for _, g := range sh.Guests {
		if sl := sh.sleepers[g.ID()]; sl != nil {
			d.Sleepers = append(d.Sleepers, DumpSleeper{
				ID: g.ID(), Name: g.Name, Location: sl.Location,
				Attended: sl.Attended, Seconds: sh.SleptFor(g),
			})
		}
	}
	for _, g := range sh.Waiting() {
		d.Hallway = append(d.Hallway, g.Name)
	}
	for _, name := range sh.StaffNames() {
		d.Staff = append(d.Staff, DumpStaff{Name: name, Busy: sh.Busy(name)})
	}
	if a := sh.Alert(); a != nil {
		d.Alerts = append(d.Alerts, "no room: "+a.Guest.Name)
	}
}

// dumpDesk describes the desk the way the drawing sees it.
func dumpDesk(sh *Shift) DumpDesk {
	desk := sh.Desk()
	dd := DumpDesk{State: "none", FilesWaiting: sh.FilesWaiting(), Pulsing: sh.Pulsing(), Visited: desk.Visited()}
	if pose, ok := sh.FolderPose(); ok {
		dd.Folder, dd.Page, dd.Ink = pose.guest.Name, pose.page.String(), pose.ink
		switch {
		case desk.FolderGuest() == nil:
			dd.State = "leaving"
		case !sh.FolderReady():
			dd.State = "arriving"
		case pose.open >= 1:
			dd.State = "open"
		case pose.open > 0:
			dd.State = "opening"
		default:
			dd.State = "closed"
		}
	}
	if _, turning := sh.Turning(); turning {
		dd.Turning = true
	}
	if st, ok := desk.Holding(); ok {
		dd.Holding = st.String()
	}
	if stamped, right, ok := sh.Slip(); ok {
		l1, l2 := citationLines(stamped, right)
		dd.Slip = l1 + " / " + l2
	}
	return dd
}

func writeDump(path string, d Dump) error {
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return fmt.Errorf("dump: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("dump: %w", err)
	}
	return nil
}
