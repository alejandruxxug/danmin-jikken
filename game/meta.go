package game

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// The meta-horror (Game Design -> Meta-horror): the secret 3:17 ending leaves
// a real file behind, written by Hana. Two rules keep it harmless:
//   - it is one new text file with a fixed name, never an overwrite of
//     anything else, and
//   - it goes next to the scores file (the game's own folder), or if that
//     folder cannot be written, into the player's home folder, and nowhere
//     else.

// SleepLogFile is the file's name.
const SleepLogFile = "sleep_log_317.txt"

// sleepLogText is what Hana wrote (Night 3 -> Secret ending).
const sleepLogText = `SLEEP LOG - BED 4
Day 1   awake 0 h
Day 2   awake 0 h
Day 3   awake 0 h
Night   awake 3:17

If you're reading this, you're awake.
Stay that way. For me.
- H.
`

// WriteSleepLog writes the file next to scoresPath and returns where it
// went. With no scores path (test mode) it writes nothing and returns "":
// a test run must not leave files around.
func WriteSleepLog(scoresPath string) (string, error) {
	if scoresPath == "" {
		return "", nil
	}
	path := filepath.Join(filepath.Dir(scoresPath), SleepLogFile)
	err := os.WriteFile(path, []byte(sleepLogText), 0o644)
	if err == nil {
		return path, nil
	}
	// The game's folder is read-only (an installed copy): try the home folder.
	home, homeErr := os.UserHomeDir()
	if homeErr != nil {
		return "", fmt.Errorf("writing %s: %w", SleepLogFile, errors.Join(err, homeErr))
	}
	path = filepath.Join(home, SleepLogFile)
	if err := os.WriteFile(path, []byte(sleepLogText), 0o644); err != nil {
		return "", fmt.Errorf("writing %s: %w", SleepLogFile, err)
	}
	return path, nil
}
