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
stamp | 180 | -scene day.reception -day 1 -seed 7 | input stamp.txt | fadein 20  # trailing comment
title | 120 | | card trailer_cards:logo | music driftwell_jingle 4.5 0.8 | fadeout 30 | cut
plain | 60 | -scene title
`
	shots, err := ParseShots(strings.NewReader(list), dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(shots) != 3 {
		t.Fatalf("got %d shots", len(shots))
	}
	want := Shot{Name: "stamp", Frames: 180, Args: []string{"-scene", "day.reception", "-day", "1", "-seed", "7"},
		Input: filepath.Join(dir, "stamp.txt"), FadeIn: 20}
	if !reflect.DeepEqual(shots[0], want) {
		t.Errorf("shot 0 = %+v, want %+v", shots[0], want)
	}
	if s := shots[1]; s.Card != "trailer_cards:logo" || s.FadeOut != 30 || !s.Cut || *s.Bed != (Bed{"driftwell_jingle", 4.5, 0.8}) {
		t.Errorf("shot 1 = %+v", s)
	}
	if got := shots[1].GameArgs(); !reflect.DeepEqual(got, []string{"-agent", "-scene", "trailer.card", "-card", "trailer_cards:logo"}) {
		t.Errorf("card args = %v", got)
	}
	if got := shots[0].GameArgs(); got[0] != "-agent" || got[len(got)-2] != "-input" {
		t.Errorf("game args = %v", got)
	}
	if shots[2].Bed != nil || shots[2].Cut {
		t.Errorf("shot 2 should have no clauses: %+v", shots[2])
	}
}

func TestParseShotsErrors(t *testing.T) {
	for name, line := range map[string]string{
		"few fields":     "a | 30",
		"bad frames":     "a | x | -scene title",
		"zero frames":    "a | 0 | -scene title",
		"no args":        "a | 30 |",
		"bad name":       "a b | 30 | -scene title",
		"bad clause":     "a | 30 | -scene title | jump 3",
		"bad music":      "a | 30 | -scene title | music x",
		"bad offset":     "a | 30 | -scene title | music x -1",
		"bad fade":       "a | 30 | -scene title | fadein soon",
		"fades too long": "a | 30 | -scene title | fadein 20 | fadeout 20",
		"cut arg":        "a | 30 | -scene title | cut now",
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

func TestBedStarts(t *testing.T) {
	bed := &Bed{"theme", 10, 1}
	other := &Bed{"theme", 10, 0.5}
	shots := []Shot{
		{Frames: 60, Bed: bed},
		{Frames: 120, Bed: bed},  // the same bed goes on
		{Frames: 60},             // no bed
		{Frames: 60, Bed: bed},   // starts again at its offset
		{Frames: 60, Bed: other}, // another clause: a new bed
	}
	got := BedStarts(shots)
	want := []float64{10, 11, -1, 10, 10}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BedStarts = %v, want %v", got, want)
	}
}

func TestPlanAndFades(t *testing.T) {
	plan := Plan([]Shot{{Frames: 4, FadeIn: 2}, {Frames: 3, FadeOut: 2}})
	want := []FrameRef{
		{0, 1, 0}, {0, 2, 0.5}, {0, 3, 1}, {0, 4, 1},
		{1, 1, 1}, {1, 2, 1}, {1, 3, 0},
	}
	// fade out of 2 on 3 frames: left = 2, 1, 0 -> factors 1, 0.5, 0
	want[5].Factor = 0.5
	if !reflect.DeepEqual(plan, want) {
		t.Errorf("Plan = %v, want %v", plan, want)
	}
}
