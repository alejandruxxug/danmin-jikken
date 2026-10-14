package game

// The rock audition (Alejo, 2026-10-14: "a preview with different rock in the
// forever ending, I would love to explore options"). It is for listening, not
// for players: `go run ./cmd/danmin -play ending.forever -rock <name>` opens
// that scene with sound on, and the ending plays a file from
// assets/sound/audition/ (on disk, never embedded) instead of its own rock.
// It has nothing to do with the -agent backdoor.

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// AuditionDir is where the candidate tracks are (run the game from game/).
const AuditionDir = "assets/sound/audition"

// previewPrefix marks the 30 s clips made for listening outside the game.
const previewPrefix = "preview_"

// ErrUnknownRock is returned for a -rock name that is not in the folder.
var ErrUnknownRock = errors.New("unknown rock track")

// rockName gives the -rock name of a file in the audition folder: the file
// name without its extension. Previews, the README and anything that is not
// an mp3, ogg or wav are not tracks.
func rockName(file string) (string, bool) {
	ext := filepath.Ext(file)
	switch ext {
	case ".mp3", ".ogg", ".wav":
	default:
		return "", false
	}
	name := strings.TrimSuffix(file, ext)
	if strings.HasPrefix(name, previewPrefix) {
		return "", false
	}
	return name, true
}

// rockNames lists the track names among a folder's file names, in order.
func rockNames(files []string) []string {
	var names []string
	for _, f := range files {
		if name, ok := rockName(f); ok {
			names = append(names, name)
		}
	}
	return names
}

// LookupRock finds the file of a -rock name among a folder's file names. An
// unknown name gives an error that lists the choices.
func LookupRock(name string, files []string) (string, error) {
	for _, f := range files {
		if n, ok := rockName(f); ok && n == name {
			return f, nil
		}
	}
	return "", fmt.Errorf("%w %q (choices: %s)", ErrUnknownRock, name, strings.Join(rockNames(files), ", "))
}

// auditionFiles reads the file names of the audition folder.
func auditionFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("audition folder %s (run the game from the game/ folder): %w", dir, err)
	}
	var files []string
	for _, e := range entries {
		files = append(files, e.Name())
	}
	return files, nil
}

// ListRocks prints the -rock choices, one per line (-rock list).
func ListRocks(w io.Writer) error {
	files, err := auditionFiles(AuditionDir)
	if err != nil {
		return err
	}
	names := rockNames(files)
	if len(names) == 0 {
		return fmt.Errorf("no tracks in %s", AuditionDir)
	}
	for _, n := range names {
		fmt.Fprintln(w, n)
	}
	return nil
}

// endingOfScene gives the ending a scene id stands for ("" if the scene is not an ending).
func endingOfScene(id string) string {
	name, ok := strings.CutPrefix(id, "ending.")
	if !ok || name == "" {
		return ""
	}
	return name
}

// applyRock makes the ending of the -play scene use the named audition
// track. It checks the name before the window opens.
func (e *Env) applyRock(scene, rock string) error {
	ending := endingOfScene(scene)
	if ending == "" {
		return fmt.Errorf("-rock only works with -play ending.<sleep|wake|forever|317>, not %q", scene)
	}
	files, err := auditionFiles(AuditionDir)
	if err != nil {
		return err
	}
	file, err := LookupRock(rock, files)
	if err != nil {
		return err
	}
	if e.Sound != nil {
		e.Sound.Override("ending_rock_"+ending, filepath.Join(AuditionDir, file))
	}
	return nil
}
