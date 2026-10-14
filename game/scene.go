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
	SceneDayDesk      = "day.desk" // the look-down desk, with the tilt
	// The same desk entered settled, with the folder on it (test mode).
	SceneDeskClosed    = "day.desk.closed"
	SceneDeskReferral  = "day.desk.referral"
	SceneDeskSleepLog  = "day.desk.sleeplog"
	SceneDeskWristband = "day.desk.wristband"
	SceneDayComputer   = "day.computer" // the BEDS tab
	SceneRecords       = "day.computer.records"
	SceneSleepLogs     = "day.computer.sleeplogs"
	SceneWardFeed      = "day.computer.wardfeed"
	SceneDayTop10      = "day.top10"
)

// Env is what every scene shares: the clock, the shift, the input, the
// day and the story flags. The scene manager owns one.
type Env struct {
	Seed    int64
	Day     int
	Speed   float64
	CRT     bool // the day CRT filter; no filter exists yet, so nothing reads it
	Clock   *Clock
	Flags   map[string]bool
	Scores  map[string]int // points of each shift, keyed "day1", "day2", ...
	Input   *Input
	Sprites fs.FS // the embedded assets/sprites folder

	// CameraX is the left edge of the 480-wide view over the 720-wide
	// reception room (0 to 240). It lives here so the backdoor can set it
	// and the camera stays where it was when the desk scenes switch.
	CameraX float64

	ClockRate  float64    // in-game minutes per second at x1; the shift sets it
	Sleepy     bool       // test mode: patients fall asleep fast and sleep long
	ScoresPath string     // the scores file; "" keeps scores in memory only
	Dark       int        // test mode: every guest this many steps darker, -1 = off
	MemScores  []ScoreRow // the rows kept in memory when there is no file
	Tab        string     // the open tab of the patient computer
	Query      string     // what is typed in the RECORDS name field
	LogPick    string     // the name picked on the SLEEP LOGS tab
	Ranked     []ScoreRow // the top-10 as the last top-10 scene showed it
	KeepClock  bool       // test mode: the next shift starts at Clock, not 09:00

	art   *DeskArt
	shift *Shift
	// nextScene is set by SwitchTo, applied when Update ends
	nextScene string
}

// Art loads the day sheets and the font the first time it is asked.
func (e *Env) Art() (*DeskArt, error) {
	if e.art == nil {
		art, err := LoadDeskArt(e.Sprites)
		if err != nil {
			return nil, err
		}
		e.art = art
	}
	return e.art, nil
}

// Shift returns the running shift, or nil when no day scene has started one.
func (e *Env) Shift() *Shift { return e.shift }

// StartShift builds today's queue and ward and starts the clock at nine.
// A day scene calls it in Enter; it does nothing if a shift is running, so
// the reception, the computer and the top-10 can all be entered directly.
func (e *Env) StartShift() error {
	if e.shift != nil {
		return nil
	}
	sh, err := NewShift(e.Day, e.Seed, e.Speed, e.Sleepy, e.Scores)
	if err != nil {
		return err
	}
	e.shift = sh
	sh.forceDark = e.Dark
	e.ClockRate = shiftMinutes / float64(sh.Rules.ShiftSeconds)
	if !e.KeepClock {
		e.Clock = NewClock(9, 0)
	}
	e.KeepClock = false
	return nil
}

// StopShift stops the patients' goroutines and forgets the shift. The
// points stay in Scores.
func (e *Env) StopShift() error {
	if e.shift == nil {
		return nil
	}
	err := e.shift.Stop()
	e.shift = nil
	return err
}

// shiftMinutes is how much clock a shift covers: 09:00 to 17:00.
const shiftMinutes = 8 * 60

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
