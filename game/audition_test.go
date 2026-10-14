package game

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

var auditionFolder = []string{
	"README.txt", "aggressor.mp3", "ending_scene.wav", "ironbound.mp3",
	"preview_aggressor.mp3", "preview_ending_scene.wav", "preview_ironbound.mp3",
}

func TestRockNamesSkipPreviewsAndOtherFiles(t *testing.T) {
	got := rockNames(auditionFolder)
	want := []string{"aggressor", "ending_scene", "ironbound"}
	if !slices.Equal(got, want) {
		t.Errorf("rockNames = %v, want %v", got, want)
	}
}

func TestLookupRock(t *testing.T) {
	for name, want := range map[string]string{
		"ironbound":    "ironbound.mp3",
		"ending_scene": "ending_scene.wav",
	} {
		got, err := LookupRock(name, auditionFolder)
		if err != nil || got != want {
			t.Errorf("LookupRock(%q) = %q, %v; want %q", name, got, err, want)
		}
	}
	for _, bad := range []string{"", "nope", "preview_ironbound", "README", "ironbound.mp3"} {
		_, err := LookupRock(bad, auditionFolder)
		if !errors.Is(err, ErrUnknownRock) {
			t.Errorf("LookupRock(%q) error = %v, want ErrUnknownRock", bad, err)
			continue
		}
		if !strings.Contains(err.Error(), "aggressor, ending_scene, ironbound") {
			t.Errorf("the error for %q should list the choices: %v", bad, err)
		}
	}
}

func TestEndingOfScene(t *testing.T) {
	for id, want := range map[string]string{
		"ending.forever": "forever", "ending.317": "317", "night.1": "", "ending.": "", "title": "",
	} {
		if got := endingOfScene(id); got != want {
			t.Errorf("endingOfScene(%q) = %q, want %q", id, got, want)
		}
	}
}

func TestParseAuditionFlags(t *testing.T) {
	o, err := ParseOptions([]string{"-play", "ending.forever", "-rock", "ironbound", "-line", "23", "-flags", "route_hana", "-seed", "9"})
	if err != nil {
		t.Fatal(err)
	}
	// The audition keeps its own flags and the line; every other test-mode flag stays off.
	want := Options{Play: "ending.forever", Rock: "ironbound", Line: 23, Flags: "route_hana"}
	if o != want {
		t.Errorf("options = %+v, want %+v", o, want)
	}
	if o.Enabled {
		t.Error("the audition is not test mode: sound must stay on")
	}
	if o, err := ParseOptions([]string{"-rock", "list"}); err != nil || o.Rock != "list" {
		t.Errorf("-rock list alone: %+v, %v", o, err)
	}
	for _, args := range [][]string{
		{"-rock", "ironbound"},                                       // no scene to play
		{"-agent", "-play", "ending.forever"},                        // the backdoor has -scene
		{"-agent", "-scene", "ending.forever", "-rock", "ironbound"}, // the backdoor is silent
	} {
		if _, err := ParseOptions(args); err == nil {
			t.Errorf("%v: want an error", args)
		}
	}
}

func TestRockOnlyForEndings(t *testing.T) {
	env := testEnv(t)
	if err := env.applyRock("night.1", "ironbound"); err == nil {
		t.Error("-rock with a scene that is not an ending must fail")
	}
}
