package game

import (
	"errors"
	"fmt"
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// Game is the Ebitengine game: it owns the shared Env, the scene manager
// and, in test mode, the agent backdoor.
type Game struct {
	env    *Env
	scenes *SceneManager
	agent  *Agent // nil unless -agent was passed

	canvas *ebiten.Image // the 480x270 picture every scene draws on
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
		Flags:   make(map[string]bool),
		Scores:  make(map[string]int),
		Input:   NewInput(),
		Sprites: fsys,
		CameraX: DefaultCameraX,

		ClockRate: 1,
	}
	g := &Game{env: env, scenes: NewSceneManager(env)}
	g.scenes.Register(SceneTitle, func() Scene { return &titleScene{} })
	g.scenes.Register(SceneDayReception, func() Scene { return &receptionScene{} })
	g.scenes.Register(SceneDayComputer, func() Scene { return &computerScene{} })
	g.scenes.Register(SceneDayTop10, func() Scene { return &top10Scene{} })

	first := SceneTitle
	if opts.Enabled {
		agent, err := newAgent(opts, env)
		if err != nil {
			return nil, err
		}
		g.agent = agent
		first = opts.Scene
	} else {
		// A real player: the scores live next to the executable, and the
		// ghost row may appear between runs. Test mode keeps its hands off
		// that file unless -scores names another one.
		path, err := DefaultScoresPath()
		if err != nil {
			return nil, err
		}
		env.ScoresPath = path
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
		if sh.Over() {
			g.env.SwitchTo(SceneDayTop10) // the shift is over: show the scores
		}
	}
	err := g.scenes.Update()
	g.env.Input.endFrame()
	return err
}

// Draw paints the 480x270 picture on the canvas, saves the screenshots the
// agent asked for (they are always of the canvas, at 1x or x4, whatever the
// window looks like), then puts the canvas on the screen.
func (g *Game) Draw(screen *ebiten.Image) {
	if g.canvas == nil {
		g.canvas = ebiten.NewImage(ScreenW, ScreenH)
	}
	g.scenes.Draw(g.canvas)
	if g.agent != nil {
		g.agent.AfterDraw(g, g.canvas)
	}
	scale, offX, offY := fitScreen(screen.Bounds().Dx(), screen.Bounds().Dy())
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(float64(scale), float64(scale))
	op.GeoM.Translate(float64(offX), float64(offY))
	op.Filter = ebiten.FilterNearest // whole-number scale + nearest = crisp pixels
	screen.Fill(color.Black)         // the bars around the picture
	screen.DrawImage(g.canvas, &op)
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
