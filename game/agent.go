package game

// The agent backdoor (Architecture -> Agent backdoor). A set of command-line
// flags that only work when -agent is passed, so a tester can jump straight
// to a scene, play a script of inputs, and look at the result as a PNG and a
// json dump. Everything here uses things the game already has (the scene
// manager, the clock, the ward). It never touches hospital/.

import (
	"bufio"
	"encoding/json"
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
	Speed   float64
	Input   string // path of the input script
	Shot    string // 1x PNG, written when the run ends
	Shot4   string // x4 PNG, written when the run ends
	Dump    string // json state, written when the run ends
	NoCRT   bool
	Quit    bool
	Frames  int
	Pan     int  // camera x at the start (used only when PanSet)
	PanSet  bool // -pan was given; without it the camera starts at 120
}

// ParseOptions reads the flags. Every flag except -agent is ignored (and
// reset) when -agent is missing, so a player can't trip the backdoor.
func ParseOptions(args []string) (Options, error) {
	var o Options
	fs := flag.NewFlagSet("danmin", flag.ContinueOnError)
	fs.BoolVar(&o.Enabled, "agent", false, "turn on test mode (the other flags need it)")
	fs.StringVar(&o.Scene, "scene", SceneTitle, "scene id to start in")
	fs.IntVar(&o.Day, "day", 1, "day number")
	fs.StringVar(&o.Clock, "clock", "", "in-game time, HH:MM")
	fs.Int64Var(&o.Seed, "seed", 0, "random seed (0 = random)")
	fs.StringVar(&o.Flags, "flags", "", "story flags, comma separated")
	fs.Float64Var(&o.Speed, "speed", 1, "game time multiplier")
	fs.StringVar(&o.Input, "input", "", "input script file")
	fs.StringVar(&o.Shot, "shot", "", "save a 1x PNG screenshot when the run ends")
	fs.StringVar(&o.Shot4, "shot4", "", "save a x4 PNG screenshot when the run ends")
	fs.StringVar(&o.Dump, "dump", "", "save the game state as json when the run ends")
	fs.BoolVar(&o.NoCRT, "nocrt", false, "turn the day CRT filter off")
	fs.BoolVar(&o.Quit, "quit", false, "exit when the script (or -frames) ends")
	fs.IntVar(&o.Frames, "frames", 0, "end the run after N frames")
	fs.IntVar(&o.Pan, "pan", 0, "camera x at the start, 0-240 (default 120, the centre)")
	if err := fs.Parse(args); err != nil {
		return Options{}, err
	}
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "pan" {
			o.PanSet = true
		}
	})
	if !o.Enabled {
		return Options{}, nil
	}
	return o, nil
}

// Command is one line of an input script.
type Command struct {
	Op   string // "wait", "key", "click", "mouse", "pan", "shot" or "dump"
	N    int    // wait: frames; pan: camera x
	Key  ebiten.Key
	X, Y int    // click, mouse: 1x game coordinates
	Path string // shot, dump: output file
}

// ParseScript reads an input script, one command per line:
//
//	wait 30         frames
//	key Space
//	click 212 140   1x game coordinates
//	mouse 5 135     hold the cursor there (edge of the screen pans the camera)
//	pan 0           put the camera at x = 0 (0 to 240)
//	shot out.png
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
		if err := cmd.Key.UnmarshalText([]byte(args[0])); err != nil {
			return cmd, fmt.Errorf("key %q: %w", args[0], err)
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
	case "shot", "dump":
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
		env.Day = opts.Day
	}
	if opts.Clock != "" {
		clock, err := ParseClock(opts.Clock)
		if err != nil {
			return nil, err
		}
		env.Clock = clock
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
	env.CRT = !opts.NoCRT
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
		case "pan":
			env.CameraX = float64(cmd.N)
		case "shot":
			a.pending = append(a.pending, output{"shot", cmd.Path})
		case "dump":
			a.pending = append(a.pending, output{"dump", cmd.Path})
		}
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
			err = writePNG(out.path, readScreen(screen), 4)
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

// BuildDump collects the state. The hospital part comes from Snapshot, so
// it is safe to call while the patient goroutines run. With no ward the
// patient and room lists are empty.
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
	}
	if env.ward != nil {
		snap := env.ward.Snapshot()
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
	}
	for name, on := range env.Flags {
		if on {
			d.Flags = append(d.Flags, name)
		}
	}
	sort.Strings(d.Flags) // map order is random; sorted keeps dumps comparable
	return d
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
