package game

import (
	"errors"
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// Game is the Ebitengine game: it owns the shared Env, the scene manager
// and, in test mode, the agent backdoor.
type Game struct {
	env    *Env
	scenes *SceneManager
	agent  *Agent // nil unless -agent was passed
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
	g.scenes.Register(SceneTitle, func() Scene { return titleScene{} })
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
	g.env.Clock.Advance(g.env.ClockRate * g.env.Speed / 60)
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

// Draw paints the 480x270 screen. In test mode it then saves the
// screenshots the agent asked for.
func (g *Game) Draw(screen *ebiten.Image) {
	g.scenes.Draw(screen)
	if g.agent != nil {
		g.agent.AfterDraw(g, screen)
	}
}

// Layout fixes the game at 480x270 whatever the window size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenW, ScreenH
}

// Run opens the window (480x270 at x4) and plays until the player or the
// agent quits. It stops the ward's goroutines on the way out.
func Run(opts Options) error {
	g, err := New(opts)
	if err != nil {
		return err
	}
	ebiten.SetWindowSize(ScreenW*4, ScreenH*4)
	ebiten.SetWindowTitle("断眠実験 DANMIN JIKKEN")
	ebiten.SetScreenFilterEnabled(false) // nearest-neighbour: crisp pixels, no blur

	runErr := ebiten.RunGame(g)
	g.scenes.Close()
	_ = g.env.StopShift() // the goroutines must not outlive the window
	if runErr != nil && !errors.Is(runErr, ebiten.Termination) {
		return fmt.Errorf("running the game: %w", runErr)
	}
	return nil
}
