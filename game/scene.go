package game

import (
	"fmt"
	"io/fs"

	"github.com/hajimehoshi/ebiten/v2"
)

// Internal resolution. The window shows it at x4 (1920x1080).
const (
	ScreenW = 480
	ScreenH = 270
)

// Scene ids. The table in Architecture -> Agent backdoor lists them all.
const (
	SceneTitle        = "title"
	SceneDayReception = "day.reception"
)

// Env is what every scene shares: the clock, the ward, the input, the
// day and the story flags. The scene manager owns one.
type Env struct {
	Seed    int64
	Day     int
	Speed   float64
	CRT     bool // the day CRT filter; no filter exists yet, so nothing reads it
	Clock   *Clock
	Flags   map[string]bool
	Scores  map[string]int
	Input   *Input
	Sprites fs.FS // the embedded assets/sprites folder

	ward      *Ward
	nextScene string // set by SwitchTo, applied when Update ends
}

// Ward returns the running ward, or nil when no scene has started one.
func (e *Env) Ward() *Ward { return e.ward }

// StartWard builds the ward for the current seed and starts its goroutines.
// A scene that needs patients calls it in Enter; StopWard in Exit.
func (e *Env) StartWard() error {
	if e.ward != nil {
		return nil
	}
	w, err := NewWard(e.Seed, e.Speed)
	if err != nil {
		return err
	}
	w.Start()
	e.ward = w
	return nil
}

// StopWard stops the patients' goroutines and forgets the ward.
func (e *Env) StopWard() error {
	if e.ward == nil {
		return nil
	}
	err := e.ward.Stop()
	e.ward = nil
	return err
}

// SwitchTo asks the manager to change scene after this frame's Update.
func (e *Env) SwitchTo(id string) { e.nextScene = id }

// Scene is one screen of the game: the title, the reception desk, a night.
type Scene interface {
	Enter(env *Env) error // the scene becomes active; start whatever it needs
	Update() error        // one game tick
	Draw(screen *ebiten.Image)
	Exit() // the scene is left; stop whatever Enter started
}

// SceneManager keeps the registered scenes and runs the active one.
type SceneManager struct {
	env       *Env
	factories map[string]func() Scene
	current   Scene
	currentID string
}

// NewSceneManager makes an empty manager. Register scenes, then Switch.
func NewSceneManager(env *Env) *SceneManager {
	return &SceneManager{env: env, factories: make(map[string]func() Scene)}
}

// Register adds a scene under an id. The factory builds a fresh scene each
// time the id is entered.
func (m *SceneManager) Register(id string, factory func() Scene) {
	m.factories[id] = factory
}

// ID is the id of the active scene.
func (m *SceneManager) ID() string { return m.currentID }

// Switch leaves the active scene and enters the one registered under id.
func (m *SceneManager) Switch(id string) error {
	factory, ok := m.factories[id]
	if !ok {
		return fmt.Errorf("unknown scene %q", id)
	}
	m.Close()
	next := factory()
	if err := next.Enter(m.env); err != nil {
		return fmt.Errorf("entering scene %q: %w", id, err)
	}
	m.current, m.currentID = next, id
	return nil
}

// Update runs the active scene, then applies a scene change it asked for.
func (m *SceneManager) Update() error {
	if m.current == nil {
		return nil
	}
	if err := m.current.Update(); err != nil {
		return err
	}
	if next := m.env.nextScene; next != "" {
		m.env.nextScene = ""
		return m.Switch(next)
	}
	return nil
}

// Draw draws the active scene.
func (m *SceneManager) Draw(screen *ebiten.Image) {
	if m.current != nil {
		m.current.Draw(screen)
	}
}

// Close leaves the active scene. Call it when the game quits.
func (m *SceneManager) Close() {
	if m.current != nil {
		m.current.Exit()
		m.current, m.currentID = nil, ""
	}
}
