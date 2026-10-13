package game

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// The scores live in danmin_scores.txt next to the executable. It is plain
// text, one row per line, fields separated by spaces:
//
//	<name> <score> <HH:MM> [message...]
//
//	alejo 880 14:02
//	alejo 912 16:40
//	alejo --- 03:17 why did you let me sleep
//
// <name> is the OS username (spaces become '_'), <score> is the points of
// one shift, <HH:MM> is the time of day it was saved. A row whose score is
// "---" is the ghost row (Game Design -> Score + top-10): the game adds it
// by itself between runs, with the time 03:17 and a message. Blank lines and
// lines with fewer than three fields are ignored, so a hand-edited file
// can never stop the game.

// ScoresFile is the file name, next to the executable.
const ScoresFile = "danmin_scores.txt"

// Ghost row text (the DDLC twist).
const (
	ghostTime    = "03:17"
	ghostMessage = "why did you let me sleep"
	ghostSlot    = 2 // the ghost shows up as the third row of the top-10
	TopRows      = 10
)

// ScoreRow is one line of the file.
type ScoreRow struct {
	Name    string
	Score   int
	Time    string // HH:MM
	Message string
	Ghost   bool // the "---" row
}

// DefaultScoresPath is danmin_scores.txt in the folder of the executable.
func DefaultScoresPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("finding the executable: %w", err)
	}
	return filepath.Join(filepath.Dir(exe), ScoresFile), nil
}

// Username is the OS username, or "clerk" when it cannot be found.
func Username() string {
	if u, err := user.Current(); err == nil && u.Username != "" {
		return strings.ReplaceAll(u.Username, " ", "_")
	}
	return "clerk"
}

// LoadScores reads the file. A missing file is not an error: it is the
// first run, and there are no rows.
func LoadScores(path string) ([]ScoreRow, error) {
	f, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading scores: %w", err)
	}
	defer f.Close()

	var rows []ScoreRow
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if row, ok := parseScoreRow(scanner.Text()); ok {
			rows = append(rows, row)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading scores: %w", err)
	}
	return rows, nil
}

// parseScoreRow reads one line. ok is false for lines to ignore.
func parseScoreRow(line string) (ScoreRow, bool) {
	words := strings.Fields(line)
	if len(words) < 3 {
		return ScoreRow{}, false
	}
	row := ScoreRow{Name: words[0], Time: words[2], Message: strings.Join(words[3:], " ")}
	if words[1] == "---" {
		row.Ghost = true
		return row, true
	}
	score, err := strconv.Atoi(words[1])
	if err != nil {
		return ScoreRow{}, false
	}
	row.Score = score
	return row, true
}

// format writes a row the way the file stores it.
func (r ScoreRow) format() string {
	score := "---"
	if !r.Ghost {
		score = strconv.Itoa(r.Score)
	}
	line := fmt.Sprintf("%s %s %s", strings.ReplaceAll(r.Name, " ", "_"), score, r.Time)
	if r.Message != "" {
		line += " " + r.Message
	}
	return line
}

// AppendScore adds a row to the end of the file, creating it if needed.
func AppendScore(path string, row ScoreRow) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("saving score: %w", err)
	}
	if _, err := fmt.Fprintln(f, row.format()); err != nil {
		f.Close()
		return fmt.Errorf("saving score: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("saving score: %w", err)
	}
	return nil
}

// EnsureGhost adds the ghost row to a file that already has real rows and
// no ghost yet. Call it when the game starts: the row appears between runs,
// once, and a first run (no file) has nothing to haunt.
func EnsureGhost(path, name string) error {
	rows, err := LoadScores(path)
	if err != nil || len(rows) == 0 {
		return err
	}
	for _, r := range rows {
		if r.Ghost {
			return nil
		}
	}
	return AppendScore(path, ScoreRow{Name: name, Time: ghostTime, Message: ghostMessage, Ghost: true})
}

// RankScores orders the rows for the top-10: real rows by score, highest
// first (the earlier row wins a tie), and the ghost in the third place. It
// never returns more than TopRows rows.
func RankScores(rows []ScoreRow) []ScoreRow {
	var played, ghosts []ScoreRow
	for _, r := range rows {
		if r.Ghost {
			ghosts = append(ghosts, r)
		} else {
			played = append(played, r)
		}
	}
	sort.SliceStable(played, func(i, j int) bool { return played[i].Score > played[j].Score })

	at := min(ghostSlot, len(played))
	ranked := append([]ScoreRow(nil), played[:at]...)
	ranked = append(ranked, ghosts...)
	ranked = append(ranked, played[at:]...)
	if len(ranked) > TopRows {
		ranked = ranked[:TopRows]
	}
	return ranked
}

// BestScore is the highest real score in the rows, 0 if there is none.
func BestScore(rows []ScoreRow) int {
	best := 0
	for _, r := range rows {
		if !r.Ghost && r.Score > best {
			best = r.Score
		}
	}
	return best
}
