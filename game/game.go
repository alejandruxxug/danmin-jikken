package game

import (
	"errors"
	"fmt"
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"danmingame/night"
	"danmingame/sound"
)

// Game is the Ebitengine game: it owns the shared Env, the scene manager
// and, in test mode, the agent backdoor.
type Game struct {
	env    *Env
	scenes *SceneManager
	agent  *Agent // nil unless -agent was passed

	canvas  *ebiten.Image // the 480x270 picture every scene draws on
	crt     *CRT          // the day filter, drawn to the screen
	shotCRT *CRT          // the same shader for the x4 screenshots (a buffer of its own)
	ticks   int           // frames updated so far: the shader's clock
}

// New builds the game from the command-line options and enters the first
// scene (the title, or whatever -scene asked for in test mode).
func New(opts Options) (*Game, error) {
	fsys, err := SpriteFS()
	if err != nil {
		return nil, err
	}
	env := &Env{
		Seed:    time.Now().UnixNano(),
		Day:     1,
		Speed:   1,
		CRT:     true,
		Clock:   NewClock(9, 0), // the day shift opens at nine
		Flags:   make(night.Flags),
		Scores:  make(map[string]int),
		Dark:    -1,
		Input:   NewInput(),
		Sprites: fsys,
		CameraX: DefaultCameraX,

		ClockRate: 1,
	}
	soundFS, err := SoundFS()
	if err != nil {
		return nil, err
	}
	env.Sound = sound.New(soundFS, opts.Enabled) // test mode only records the cues
	activeSound = env.Sound
	env.Apply(DefaultSettings())
	crt, err := NewCRT()
	if err != nil {
		return nil, err
	}
	g := &Game{env: env, scenes: NewSceneManager(env), crt: crt, shotCRT: &CRT{shader: crt.shader}}
	g.scenes.Register(SceneTitle, func() Scene { return &titleScene{} })
	g.scenes.Register(SceneSettings, func() Scene { return &settingsScene{} })
	g.scenes.Register(SceneDayReception, func() Scene { return &receptionScene{} })
	g.scenes.Register(SceneDayComputer, func() Scene { return &computerScene{} })
	g.scenes.Register(SceneDayDesk, func() Scene { return &deskScene{} })
	for id, open := range map[string]struct {
		open bool
		page Page
	}{
		SceneDeskClosed: {false, PageReferral}, SceneDeskReferral: {true, PageReferral},
		SceneDeskSleepLog: {true, PageSleepLog}, SceneDeskWristband: {true, PageWristband},
	} {
		g.scenes.Register(id, func() Scene { return &deskScene{setup: deskSetup(open.open, open.page)} })
	}
	for id, tab := range map[string]string{SceneRecords: TabRecords, SceneSleepLogs: TabSleepLogs, SceneWardFeed: TabWardFeed} {
		g.scenes.Register(id, func() Scene { return &computerScene{startTab: tab} })
	}
	g.scenes.Register(SceneDayTop10, func() Scene { return &top10Scene{} })
	g.scenes.Register(SceneNightPrologue, func() Scene { return &nightScene{script: night.Prologue} })
	g.scenes.Register(SceneStrobe1, func() Scene { return &strobeScene{night: 1, next: SceneNight1} })
	g.scenes.Register(SceneNight1, func() Scene { return &nightScene{script: night.Night1} })
	g.scenes.Register(SceneStrobe2, func() Scene { return &strobeScene{night: 2, next: SceneNight2} })
	g.scenes.Register(SceneNight2, func() Scene { return &nightScene{script: night.Night2} })
	g.scenes.Register(SceneStrobe3, func() Scene { return &strobeScene{night: 3, next: SceneNight3} })
	g.scenes.Register(SceneNight3, func() Scene { return &nightScene{script: night.Night3} })
	for _, ending := range night.Endings {
		g.scenes.Register(endingSceneID(ending), func() Scene { return &endingScene{name: ending} })
	}

	first := SceneTitle
	if opts.Enabled {
		agent, err := newAgent(opts, env)
		if err != nil {
			return nil, err
		}
		g.agent = agent
		first = opts.Scene
		env.SeenPrologue = opts.SeenPrologue
	} else {
		first = SceneNightPrologue // the cold open: the Prologue plays before the title
		// A real player: the scores live next to the executable, and the
		// ghost row may appear between runs. Test mode keeps its hands off
		// that file unless -scores names another one.
		path, err := DefaultScoresPath()
		if err != nil {
			return nil, err
		}
		env.ScoresPath = path
		// The saved settings come first; the command-line options (like
		// -fullscreen) only ever turn something on or off for this run. A
		// settings file that cannot be read gives the defaults: the game
		// still starts.
		env.SettingsPath = SettingsPathFor(path)
		saved, _ := LoadSettings(env.SettingsPath)
		env.Apply(saved)
		if opts.NoCRT {
			env.CRT = false
		}
		if opts.ReduceFlashing {
			env.ReduceFlashing = true
		}
		if err := EnsureGhost(path, Username()); err != nil {
			return nil, err
		}
	}
	if err := g.scenes.Switch(first); err != nil {
		return nil, err
	}
	return g, nil
}

// Update runs once per tick (60 times a second).
func (g *Game) Update() error {
	if g.agent != nil {
		if err := g.agent.Update(g); err != nil {
			return err
		}
	}
	g.ticks++
	g.env.tickPowerOn()
	if g.env.Input.KeyPressed(ebiten.KeyF10) {
		g.env.CRT = !g.env.CRT // Settings -> CRT effect: some players find the curved, flickering text hard to read
	}
	if g.env.Input.FullscreenPressed() {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	// The clock stops on the top-10: the shift is over, so it must not run
	// on past midnight while the scores are on the screen.
	if g.scenes.ID() != SceneDayTop10 {
		g.env.Clock.Advance(g.env.ClockRate * g.env.Speed / 60)
	}
	if sh := g.env.Shift(); sh != nil {
		sh.Update()
		if g.env.UpdateLiveStrobe(sh) {
			g.env.SwitchTo(strobeSceneID(g.env.StrobeNight)) // the strobe goes on in its own scene
		} else if sh.Over() && g.env.StrobeNight == 0 {
			if err := g.endShift(); err != nil {
				return err
			}
		}
	}
	err := g.scenes.Update()
	g.env.Input.endFrame()
	return err
}

// endShift handles a shift that is over on a day with no strobe. Every day
// has one now, so this only runs for a day past the last (a test); the top-10
// comes once, after the last day, the other days save their score silently.
func (g *Game) endShift() error {
	if g.env.Day >= LastDay() {
		g.env.SwitchTo(SceneDayTop10)
		return nil
	}
	if err := g.env.EndShiftSilently(); err != nil {
		return err
	}
	g.env.Day++
	g.env.SwitchTo(SceneDayReception)
	return nil
}

// Draw paints the 480x270 picture on the canvas, saves the screenshots the
// agent asked for (they are always of the canvas, at 1x or x4, whatever the
// window looks like), then puts the canvas on the screen.
func (g *Game) Draw(screen *ebiten.Image) {
	if g.canvas == nil {
		g.canvas = ebiten.NewImage(ScreenW, ScreenH)
	}
	g.scenes.Draw(g.canvas)
	if usesCRT(g.scenes.ID(), g.scenes.Current()) {
		g.fadeFromBlack(g.canvas)
	}
	if g.agent != nil {
		g.agent.AfterDraw(g, g.canvas)
	}
	scale, offX, offY := fitScreen(screen.Bounds().Dx(), screen.Bounds().Dy())
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(float64(offX), float64(offY))
	screen.Fill(color.Black) // the bars around the picture
	if g.crtOn() {
		// The shader scales the picture itself, so it is drawn at 1:1.
		screen.DrawImage(g.crt.Apply(g.canvas, scale, g.crtParams()), &op)
		return
	}
	op.GeoM.Reset()
	op.GeoM.Scale(float64(scale), float64(scale))
	op.GeoM.Translate(float64(offX), float64(offY))
	op.Filter = ebiten.FilterNearest // whole-number scale + nearest = crisp pixels
	screen.DrawImage(g.canvas, &op)
}

// crtOn says whether this frame goes through the CRT: the setting is on and
// the scene is a day scene (the night never has it).
func (g *Game) crtOn() bool {
	return g.env.CRT && usesCRT(g.scenes.ID(), g.scenes.Current())
}

// crtParams are the shader's numbers for this frame.
func (g *Game) crtParams() CRTParams {
	p := DefaultCRT(g.env.Day, float64(g.ticks)/60, g.env.Invert)
	if st, ok := PowerOnAt(g.env.PowerOnTick); ok && g.env.PoweringOn {
		p.PowerOn = st.Progress
	}
	return p
}

// fadeFromBlack is the turn-on without the CRT (-nocrt): a plain fast fade.
func (g *Game) fadeFromBlack(canvas *ebiten.Image) {
	if st, ok := PowerOnAt(g.env.PowerOnTick); ok && g.env.PoweringOn && !g.env.CRT {
		vector.FillRect(canvas, 0, 0, ScreenW, ScreenH, color.NRGBA{0x0b, 0x0b, 0x0c, uint8((1 - st.Progress) * 255)}, false)
	}
}

// fitScreen says how to show the 480x270 picture on a screen of w x h real
// pixels: the biggest whole-number scale that fits (never below 1), and the
// offset that centres it. A whole number keeps every game pixel the same
// size, so the picture stays sharp at any monitor size; the rest of the
// screen is left black (letterboxing).
func fitScreen(w, h int) (scale, offX, offY int) {
	scale = max(1, min(w/ScreenW, h/ScreenH))
	return scale, (w - ScreenW*scale) / 2, (h - ScreenH*scale) / 2
}

// pickScale is the biggest whole-number scale from x1 to x4 whose window
// fits in 90% of the monitor (w x h in points), leaving room for the menu
// bar and the dock. It is never below x1.
func pickScale(monW, monH int) int {
	best := 1
	for k := 2; k <= 4; k++ {
		if ScreenW*k*10 <= monW*9 && ScreenH*k*10 <= monH*9 {
			best = k
		}
	}
	return best
}

// Layout is the whole-pixel version for tools that ask for it; LayoutF below
// is the one Ebitengine uses.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

// LayoutF gives the game as many pixels as the window really has (the size
// in points times the monitor's scale factor, 2 on a Retina screen), so the
// canvas is scaled by fitScreen and not by Ebitengine, which would stretch
// it by a fraction. It also tells the input where the picture sits, to turn
// mouse positions back into 480x270 game pixels.
func (g *Game) LayoutF(outsideWidth, outsideHeight float64) (float64, float64) {
	factor := ebiten.Monitor().DeviceScaleFactor()
	w, h := int(math.Ceil(outsideWidth*factor)), int(math.Ceil(outsideHeight*factor))
	scale, offX, offY := fitScreen(w, h)
	g.env.Input.SetView(scale, offX, offY)
	return float64(w), float64(h)
}

// Run opens the window (the biggest whole-number scale that fits the screen;
// F11, Alt+Enter or Cmd+F, or -fullscreen, makes it fullscreen) and plays until the player or the agent quits. It
// stops the ward's goroutines on the way out.
func Run(opts Options) error {
	g, err := New(opts)
	if err != nil {
		return err
	}
	// The window is the biggest whole-number size that fits the screen, and
	// can be resized (the picture stays whole-number scaled with black bars).
	monW, monH := ebiten.ScreenSizeInFullscreen()
	scale := pickScale(monW, monH)
	ebiten.SetWindowSize(ScreenW*scale, ScreenH*scale)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("断眠実験 DANMIN JIKKEN")
	ebiten.SetFullscreen(opts.Fullscreen)

	runErr := ebiten.RunGame(g)
	g.scenes.Close()
	_ = g.env.StopShift() // the goroutines must not outlive the window
	if runErr != nil && !errors.Is(runErr, ebiten.Termination) {
		return fmt.Errorf("running the game: %w", runErr)
	}
	return nil
}
