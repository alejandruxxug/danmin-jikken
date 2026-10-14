package game

import (
	"danmingame/sound"
	"fmt"
	"io/fs"

	"github.com/hajimehoshi/ebiten/v2"

	"danmingame/night"
)

// Internal resolution. The window shows it at x4 (1920x1080).
const (
	ScreenW = 480
	ScreenH = 270
)

// Scene ids. The table in Architecture -> Agent backdoor lists them all.
const (
	SceneTitle        = "title"
	SceneSettings     = "settings"
	SceneScenes       = "scenes" // the SCENES gallery (replays: "replay.<card id>")
	SceneDayReception = "day.reception"
	SceneDayDesk      = "day.desk" // the look-down desk, with the tilt
	// The same desk entered settled, with the folder on it (test mode).
	SceneDeskClosed    = "day.desk.closed"
	SceneDeskReferral  = "day.desk.referral"
	SceneDeskSleepLog  = "day.desk.sleeplog"
	SceneDeskWristband = "day.desk.wristband"
	SceneDeskReport    = "day.desk.report" // day 3: Hana's night report, the last page of the first folder
	SceneDayComputer   = "day.computer"    // the BEDS tab
	SceneRecords       = "day.computer.records"
	SceneSleepLogs     = "day.computer.sleeplogs"
	SceneWardFeed      = "day.computer.wardfeed"
	SceneDayTop10      = "day.top10"
	SceneNightPrologue = "night.prologue"    // the gurney, before day 1
	SceneNight1        = "night.1"           // after day 1
	SceneStrobe1       = "transition.night1" // the day-to-night strobe into night 1
	SceneNight2        = "night.2"
	SceneStrobe2       = "transition.night2"
	SceneNight3        = "night.3"
	SceneStrobe3       = "transition.night3"
)

// Env is what every scene shares: the clock, the shift, the input, the
// day and the story flags. The scene manager owns one.
type Env struct {
	Seed           int64
	Day            int
	Speed          float64
	CRT            bool      // the day CRT filter (Settings -> CRT effect); the night never has it
	ReduceFlashing bool      // Settings -> Reduce flashing: the strobe becomes one slow invert-and-fade
	SkipStrobe     bool      // test mode: the strobe scene ends at once
	StrobeNight    int       // the night the strobe is leading into (0 when none plays), for the dump
	StrobeTick     int       // ticks into it
	fade           musicFade // the music fading out into a day (music.go)
	PoweringOn     bool      // the day is turning on like a CRT (poweron.go)
	PowerOnTick    int
	Invert         float64 // 1 while the strobe shows its inverted frames (the CRT shader swaps black and white)
	Clock          *Clock
	Flags          night.Flags    // story flags the nights set and the day will read
	Scores         map[string]int // points of each shift, keyed "day1", "day2", ...
	Input          *Input
	Sprites        fs.FS // the embedded assets/sprites folder

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

	Sound *sound.Player // music and sound effects; muted in test mode

	Auto         bool    // the VN's AUTO toggle (Settings -> Auto-play)
	MusicVolume  float64 // Settings: 0 to 1
	SFXVolume    float64
	TitleWhite   bool   // the Prologue's white carries into the title drop
	TitleSettled bool   // the title is entered with its logo already down (coming back from the settings)
	SeenPrologue bool   // saved: later launches may skip the Prologue with a key
	SettingsPath string // the settings file ("" in test mode: nothing is saved)

	// The SCENES gallery (gallery_cards.go). Unlocked and SavedFlags are saved
	// in the settings file.
	Unlocked     []string // ids of the cards the player has reached, in grid order
	SavedFlags   []string // flags of the run seen most recently (the replays' variants)
	Replaying    string   // the card that is replaying ("" when none), for the dump
	GallerySel   int      // the card selected in the gallery, kept while a replay plays
	GalleryState *DumpGallery

	// The end of the run.
	Ending         string // the ending that is playing: sleep, wake, forever, 317 ("" before)
	EndingPhase    string // script, silence, title or credits
	View           View   // where the picture sits on the screen (the dump shows it)
	MusicCue       string // the music the story asked for last (the sound system plays it)
	SleepLogPath   string // where the secret ending wrote sleep_log_317.txt ("" if it did not)
	SleepLogError  string
	EndingGlitch   float64          // the secret ending's glitch strength now (0 to 1), for the dump
	SkipToCard     bool             // test mode: an ending scene jumps from its script to its card
	Audition       bool             // -play: a scene opened to listen to; it leaves no trace (no unlocks, no sleep log) and ends the game after the credits
	QuitAfterTop10 bool             // the 3:17 ending closes the game after the top-10
	savedRows      map[int]ScoreRow // the score row of each day already saved, so a day is saved once

	Night     *night.Engine // the night script that is playing, nil in the day (for the dump)
	NightLine int           // the backdoor's -line N: start the night script at this file line

	art      *DeskArt
	nightArt *night.Art
	shift    *Shift
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

// NightArt loads the night sheets and the font the first time it is asked.
func (e *Env) NightArt() (*night.Art, error) {
	if e.nightArt == nil {
		font, err := LoadFont()
		if err != nil {
			return nil, err
		}
		if e.nightArt, err = night.LoadArt(e.Sprites, font); err != nil {
			return nil, err
		}
	}
	return e.nightArt, nil
}

// NewRun forgets the last run: flags, scores and the ending (the title starts
// the Prologue with it).
func (e *Env) NewRun() {
	e.Day = 1
	e.Flags = make(night.Flags)
	e.Scores = make(map[string]int)
	e.savedRows = nil
	e.Ending, e.EndingPhase, e.MusicCue = "", "", ""
	e.QuitAfterTop10 = false
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
	e.SetMusic("day_hum") // the fluorescent hum under the desk
	sh.Flags = e.Flags
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

// Current is the active scene (nil before the first Switch).
func (m *SceneManager) Current() Scene { return m.current }

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

// SetMusic plays a looping track; asking for the one that is already playing
// changes nothing. "" stops the music.
func (e *Env) SetMusic(name string) {
	if e.Sound == nil {
		return
	}
	if e.fade.active { // the old track is still fading out: this one waits for it
		e.fade.next, e.fade.queued = name, true
		return
	}
	if name == "" {
		e.Sound.StopMusic()
		return
	}
	e.Sound.PlayMusic(name, true, 0)
}

// View is where the 480x270 picture sits on the real screen: its scale, the
// offset of its top-left corner, and whether it fills the width (the visual
// novel in fullscreen).
type View struct {
	W         int     `json:"screenW"`
	H         int     `json:"screenH"`
	Scale     float64 `json:"scale"`
	OffX      float64 `json:"offX"`
	OffY      float64 `json:"offY"`
	FillWidth bool    `json:"fillWidth"`
}
