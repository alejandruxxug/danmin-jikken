package game

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScoresRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), ScoresFile)
	if rows, err := LoadScores(path); err != nil || rows != nil {
		t.Fatalf("missing file: rows %v, err %v; want none", rows, err)
	}
	if err := AppendScore(path, ScoreRow{Name: "alejo", Score: 880, Time: "14:02"}); err != nil {
		t.Fatal(err)
	}
	if err := AppendScore(path, ScoreRow{Name: "ana maria", Score: 912, Time: "16:40"}); err != nil {
		t.Fatal(err)
	}
	rows, err := LoadScores(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].Score != 880 || rows[1].Name != "ana_maria" || rows[1].Time != "16:40" {
		t.Errorf("rows = %+v", rows)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "alejo 880 14:02\nana_maria 912 16:40\n" {
		t.Errorf("file = %q", data)
	}
}

func TestLoadScoresSkipsBadLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), ScoresFile)
	text := "alejo 100 10:00\n\nshort line\nbob notanumber 11:00\ncarol 50 12:00\n"
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	rows, err := LoadScores(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].Name != "alejo" || rows[1].Name != "carol" {
		t.Errorf("rows = %+v", rows)
	}
}

// The ghost appears once, only after a real run, and reads like a real row.
func TestEnsureGhost(t *testing.T) {
	path := filepath.Join(t.TempDir(), ScoresFile)
	if err := EnsureGhost(path, "alejo"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err == nil {
		t.Fatal("a first run must not create the file")
	}
	if err := AppendScore(path, ScoreRow{Name: "alejo", Score: 880, Time: "14:02"}); err != nil {
		t.Fatal(err)
	}
	for range 2 { // the second call must add nothing
		if err := EnsureGhost(path, "alejo"); err != nil {
			t.Fatal(err)
		}
	}
	rows, _ := LoadScores(path)
	if len(rows) != 2 {
		t.Fatalf("rows = %+v, want the real row and one ghost", rows)
	}
	g := rows[1]
	if !g.Ghost || g.Name != "alejo" || g.Time != "03:17" || g.Message != "why did you let me sleep" {
		t.Errorf("ghost = %+v", g)
	}
	data, _ := os.ReadFile(path)
	if !strings.HasSuffix(string(data), "alejo --- 03:17 why did you let me sleep\n") {
		t.Errorf("file = %q", data)
	}
}

func TestRankScores(t *testing.T) {
	rows := []ScoreRow{
		{Name: "a", Score: 640}, {Name: "ghost", Ghost: true},
		{Name: "b", Score: 912}, {Name: "c", Score: 880},
	}
	got := RankScores(rows)
	var order []string
	for _, r := range got {
		order = append(order, r.Name)
	}
	if strings.Join(order, ",") != "b,c,ghost,a" {
		t.Errorf("order = %v, want b,c,ghost,a", order)
	}
	if got := RankScores(rows[:2]); len(got) != 2 || !got[1].Ghost {
		t.Errorf("with one real row the ghost goes last: %+v", got)
	}
	var many []ScoreRow
	for i := range 15 {
		many = append(many, ScoreRow{Name: "x", Score: i})
	}
	if got := RankScores(many); len(got) != TopRows || got[0].Score != 14 {
		t.Errorf("top-10 has %d rows, first score %d", len(got), got[0].Score)
	}
	if BestScore(rows) != 912 {
		t.Errorf("BestScore = %d", BestScore(rows))
	}
}

func TestUsernameIsNeverEmpty(t *testing.T) {
	if Username() == "" || strings.Contains(Username(), " ") {
		t.Errorf("Username = %q", Username())
	}
}
