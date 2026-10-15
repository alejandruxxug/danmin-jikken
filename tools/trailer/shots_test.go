package main

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseShots(t *testing.T) {
	dir := t.TempDir()
	list := `
# a comment
stamp | 146 | -scene day.desk.referral -day 2 -seed 7 | input stamp.txt | from 21  # trailing comment
plain | 60 | -scene title
`
	shots, err := ParseShots(strings.NewReader(list), dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(shots) != 2 {
		t.Fatalf("got %d shots", len(shots))
	}
	want := Shot{Name: "stamp", Frames: 146, Args: []string{"-scene", "day.desk.referral", "-day", "2", "-seed", "7"},
		Input: filepath.Join(dir, "stamp.txt"), From: 21}
	if !reflect.DeepEqual(shots[0], want) {
		t.Errorf("shot 0 = %+v, want %+v", shots[0], want)
	}
	if got := shots[0].GameArgs(); got[0] != "-agent" || got[len(got)-2] != "-input" {
		t.Errorf("game args = %v", got)
	}
	if shots[1].From != 1 || shots[1].Input != "" {
		t.Errorf("shot 1 should save every frame and have no input: %+v", shots[1])
	}
}

func TestParseShotsErrors(t *testing.T) {
	for name, line := range map[string]string{
		"few fields":  "a | 30",
		"bad frames":  "a | x | -scene title",
		"zero frames": "a | 0 | -scene title",
		"no args":     "a | 30 |",
		"bad name":    "a b | 30 | -scene title",
		"bad clause":  "a | 30 | -scene title | jump 3",
		"bad from":    "a | 30 | -scene title | from soon",
		"from zero":   "a | 30 | -scene title | from 0",
		"from late":   "a | 30 | -scene title | from 31",
		"input args":  "a | 30 | -scene title | input a b",
	} {
		if _, err := ParseShots(strings.NewReader(line), "."); err == nil {
			t.Errorf("%s: %q should fail", name, line)
		}
	}
	_, err := ParseShots(strings.NewReader("a | 30 | -scene title\na | 30 | -scene title"), ".")
	if err == nil || !strings.Contains(err.Error(), "line 2") {
		t.Errorf("a repeated name should fail naming line 2, got %v", err)
	}
}
