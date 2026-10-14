package game

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// Columns of the top-10 table (Figma px / 4).
const (
	colRank, colClerk, colScore, colTime, colNote = 40, 64, 160, 224, 272
	firstRow, rowPitch                            = 62, 13
	maxNameChars                                  = 12
)

// top10Scene is the end of a shift: the scores, read back from
// danmin_scores.txt, with today's score saved first. The file may hold a row
// the player never played (EnsureGhost); it is shown like any other.
type top10Scene struct {
	env       *Env
	art       *DeskArt
	rows      []ScoreRow // ranked, at most 10
	score     int        // today's points
	newRecord bool
	saved     bool
	ticks     int
	drawErr   error
}

func (s *top10Scene) Enter(env *Env) error {
	s.env = env
	env.SetMusic("") // the rock of an ending ends here
	var err error
	if s.art, err = env.Art(); err != nil {
		return err
	}
	// The shift is over: stop its goroutines. Reaching this scene directly
	// (the backdoor) has no shift to stop.
	ranShift := env.Shift() != nil
	if err := env.StopShift(); err != nil {
		return fmt.Errorf("the shift ended with an error: %w", err)
	}
	s.score = env.Scores[fmt.Sprintf("day%d", env.Day)]

	// With no scores file (test mode without -scores) the rows are kept in
	// memory for this run, as the backdoor docs say.
	before := env.MemScores
	if env.ScoresPath != "" {
		var err error
		if before, err = LoadScores(env.ScoresPath); err != nil {
			return err
		}
	}
	s.newRecord = s.score > 0 && s.score > BestScore(before)
	rows := before
	// A shift that was played (or a score the backdoor set) is saved.
	if ranShift || s.score > 0 {
		row, err := env.saveScore(s.score)
		if err != nil {
			return err
		}
		s.saved = true
		rows = append(append([]ScoreRow(nil), rows...), row)
	}
	s.rows = RankScores(rows)
	env.Ranked = s.rows
	return nil
}

func (s *top10Scene) Exit() { s.env.MusicCue = "" } // the ending's rock is over

// saveScore writes today's score to the scores file (or keeps it in memory
// in test mode) and returns the row.
func (e *Env) saveScore(score int) (ScoreRow, error) {
	if row, done := e.savedRows[e.Day]; done {
		return row, nil // the strobe saved this day silently already
	}
	row := ScoreRow{Name: Username(), Score: score, Time: time.Now().Format("15:04")}
	if e.ScoresPath != "" {
		if err := AppendScore(e.ScoresPath, row); err != nil {
			return row, err
		}
	} else {
		e.MemScores = append(e.MemScores, row)
	}
	if e.savedRows == nil {
		e.savedRows = make(map[int]ScoreRow)
	}
	e.savedRows[e.Day] = row
	return row, nil
}

// EndShiftSilently stops the shift's goroutines and saves its score without
// showing the top-10: that table comes once, at the end of the run. A scene
// entered with no shift and no score (the backdoor) saves nothing.
func (e *Env) EndShiftSilently() error {
	ran := e.Shift() != nil
	if err := e.StopShift(); err != nil {
		return fmt.Errorf("the shift ended with an error: %w", err)
	}
	if score := e.Scores[fmt.Sprintf("day%d", e.Day)]; ran || score > 0 {
		_, err := e.saveScore(score)
		return err
	}
	return nil
}

func (s *top10Scene) Update() error {
	if s.drawErr != nil {
		return s.drawErr
	}
	s.ticks++
	if x, y, ok := s.env.Input.Click(); ok && closeRect.has(x, y) {
		s.next()
		if s.env.QuitAfterTop10 {
			return ebiten.Termination // the secret ending: the game closes itself
		}
	}
	return nil
}

// next goes on after the top-10: the next day, or the title after the last.
func (s *top10Scene) next() {
	if s.env.Day >= LastDay() {
		s.env.Day = 1
		s.env.SwitchTo(SceneTitle)
		return
	}
	s.env.Day++
	s.env.SwitchTo(SceneDayReception)
}

func (s *top10Scene) Draw(screen *ebiten.Image) {
	p := &painter{art: s.art, screen: screen, day: s.env.Day, elapse: time.Duration(s.ticks) * time.Second / 60}
	p.terminal(nil, int(s.env.CameraX), "V CLOSE V")

	p.text("TOP 10  DRIFTWELL CLERKS", textLeft, 27, colPaper)
	p.textRight("SHIFT RECORDS", 430, 27, colPaper)
	p.fill(rect{textLeft, 38, textRight - textLeft + 4, 1}, colPaper)

	p.text("#", colRank, 47, colDim)
	p.text("CLERK", colClerk, 47, colDim)
	p.text("SCORE", colScore, 47, colDim)
	p.text("TIME", colTime, 47, colDim)

	for i := range TopRows {
		y := firstRow + rowPitch*i
		if i >= len(s.rows) {
			p.row(fmt.Sprint(i+1), "---", "---", "--:--", y)
			continue
		}
		r := s.rows[i]
		rank, score := fmt.Sprint(i+1), fmt.Sprint(r.Score)
		if r.Ghost {
			rank, score = "-", "---"
		}
		p.row(rank, shortName(r.Name), score, r.Time, y)
		p.text(r.Message, colNote, y, colPaper)
	}

	footer := fmt.Sprintf("YOUR SCORE TODAY: %d", s.score)
	if s.newRecord {
		footer += "  NEW RECORD!"
	}
	p.text(footer, textLeft, 195, colPaper)
	saved := "SAVED: " + ScoresFile
	switch {
	case !s.saved:
		saved = "NOT SAVED"
	case s.env.ScoresPath == "":
		saved = "KEPT IN MEMORY"
	}
	p.text(saved, 280, 195, colPaper)

	p.glass()
	if p.err != nil && s.drawErr == nil {
		s.drawErr = p.err
	}
}

// shortName cuts a name to maxNameChars letters. It counts runes, not bytes,
// so an OS username with accents is never cut in the middle of a letter.
func shortName(name string) string {
	if runes := []rune(name); len(runes) > maxNameChars {
		return string(runes[:maxNameChars])
	}
	return name
}

// row prints the four columns of one table row, its top at y.
func (p *painter) row(rank, name, score, when string, y int) {
	p.text(rank, colRank, y, colPaper)
	p.text(name, colClerk, y, colPaper)
	p.text(score, colScore, y, colPaper)
	p.text(when, colTime, y, colPaper)
}
