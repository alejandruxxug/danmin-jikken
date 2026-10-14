package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// SettingsFile is the settings file's name. It sits next to the scores file
// (Settings screen, Build Plan -> M7).
const SettingsFile = "danmin_settings.json"

// Settings is what the player chose, and what the game remembers between
// runs. The same fields live in Env while the game runs.
type Settings struct {
	CRT            bool    `json:"crt"`            // the day CRT filter
	ReduceFlashing bool    `json:"reduceFlashing"` // the strobe and the glitch lose their hard flicker
	Auto           bool    `json:"auto"`           // the VN's AUTO play
	Music          float64 `json:"music"`          // 0 to 1
	SFX            float64 `json:"sfx"`            // 0 to 1
	SeenPrologue   bool    `json:"seenPrologue"`   // later launches may skip the Prologue with a key
	// The SCENES gallery. An older settings file has neither: nothing unlocked.
	Scenes     []string `json:"scenes"`     // ids of the unlocked cards
	SceneFlags []string `json:"sceneFlags"` // flags of the run seen most recently
}

// DefaultSettings are the settings of a first launch.
func DefaultSettings() Settings {
	return Settings{CRT: true, Music: 0.8, SFX: 0.8}
}

// SettingsPathFor is where the settings file goes: next to the scores file.
// An empty scores path (test mode) gives an empty settings path: nothing is saved.
func SettingsPathFor(scoresPath string) string {
	if scoresPath == "" {
		return ""
	}
	return filepath.Join(filepath.Dir(scoresPath), SettingsFile)
}

// LoadSettings reads the file. A missing file is the first launch, not an
// error; a file that cannot be read gives the defaults and the error.
func LoadSettings(path string) (Settings, error) {
	s := DefaultSettings()
	if path == "" {
		return s, nil
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return DefaultSettings(), fmt.Errorf("reading the settings: %w", err)
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return DefaultSettings(), fmt.Errorf("reading the settings %s: %w", path, err)
	}
	s.Music, s.SFX = min(max(s.Music, 0), 1), min(max(s.SFX, 0), 1)
	s.Scenes = knownCards(s.Scenes)
	return s, nil
}

// SaveSettings writes the file. An empty path saves nothing.
func SaveSettings(path string, s Settings) error {
	if path == "" {
		return nil
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("writing the settings: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing the settings: %w", err)
	}
	return nil
}

// Settings gathers the settings the game is running with.
func (e *Env) Settings() Settings {
	return Settings{CRT: e.CRT, ReduceFlashing: e.ReduceFlashing, Auto: e.Auto,
		Music: e.MusicVolume, SFX: e.SFXVolume, SeenPrologue: e.SeenPrologue,
		Scenes: e.Unlocked, SceneFlags: e.SavedFlags}
}

// Apply puts settings into the running game: the flags, and the two volumes
// on the sound player.
func (e *Env) Apply(s Settings) {
	e.CRT, e.ReduceFlashing, e.Auto = s.CRT, s.ReduceFlashing, s.Auto
	e.MusicVolume, e.SFXVolume, e.SeenPrologue = s.Music, s.SFX, s.SeenPrologue
	e.Unlocked, e.SavedFlags = s.Scenes, s.SceneFlags
	if e.Sound != nil {
		e.Sound.SetMusicVolume(s.Music)
		e.Sound.SetSFXVolume(s.SFX)
	}
}

// SaveSettings writes the settings to the settings file (nothing in test mode).
func (e *Env) SaveSettings() error { return SaveSettings(e.SettingsPath, e.Settings()) }
