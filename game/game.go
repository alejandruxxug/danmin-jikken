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
	}
	g := &Game{env: env, scenes: NewSceneManager(env)}
	g.scenes.Register(SceneTitle, func() Scene { return titleScene{} })
	g.scenes.Register(SceneDayReception, func() Scene { return &receptionScene{} })

	first := SceneTitle
	if opts.Enabled {
		agent, err := newAgent(opts, env)
		if err != nil {
			return nil, err
		}
		g.agent = agent
		first = opts.Scene
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
	g.env.Clock.Advance(g.env.Speed / 60) // one in-game minute per second at x1
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
	if runErr != nil && !errors.Is(runErr, ebiten.Termination) {
		return fmt.Errorf("running the game: %w", runErr)
	}
	return nil
}
