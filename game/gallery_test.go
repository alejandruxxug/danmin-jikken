package game

import (
	"image"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"danmingame/night"
)

// playOut plays a replay engine to its end the way a player would (click
// through every line) and returns the file lines it showed and where it ended.
func playOut(t *testing.T, e *night.Engine) (lines []int, to string) {
	t.Helper()
	for step := 0; step < 20000; step++ {
		if to, ended := e.Ended(); ended {
			return lines, to
		}
		if len(e.Options()) > 0 {
			t.Fatalf("a replay stopped at a menu on line %d", e.Line())
		}
		if e.Line() > 0 && (len(lines) == 0 || lines[len(lines)-1] != e.Line()) {
			lines = append(lines, e.Line())
		}
		if err := e.Advance(); err != nil {
			t.Fatal(err)
		}
		if err := e.Tick(); err != nil {
			t.Fatal(err)
		}
	}
	t.Fatal("a replay never ended")
	return nil, ""
}

// Every card's range is real: the replay starts inside it, stays inside it,
// and ends. Run for every saved route, so every variant is covered.
func TestEveryCardReplaysItsRange(t *testing.T) {
	saves := map[string][]string{
		"none":   nil,
		"mayu":   {"liked_mayu", "stay_mayu", "route_mayu", "closed_eyes"},
		"hana":   {"liked_hana", "find_hana", "route_hana", "read_letter"},
		"nobody": {"confront_mayu", "route_nobody"},
	}
	for _, c := range sceneCards {
		for name, saved := range saves {
			e, err := newReplayEngine(c, saved)
			if err != nil {
				t.Fatalf("%s (%s): %v", c.ID, name, err)
			}
			lines, to := playOut(t, e)
			if len(lines) == 0 {
				t.Errorf("%s (%s): shows no line", c.ID, name)
				continue
			}
			if first := lines[0]; first < c.From || (c.To > 0 && first >= c.To) {
				t.Errorf("%s (%s): starts on line %d, outside %d-%d", c.ID, name, first, c.From, c.To)
			}
			if last := lines[len(lines)-1]; c.To > 0 && last >= c.To {
				t.Errorf("%s (%s): reached line %d, past its end %d", c.ID, name, last, c.To)
			}
			if c.To > 0 && to != "" {
				t.Errorf("%s (%s): ended with %q, want the range's stop", c.ID, name, to)
			}
		}
	}
}

// A replay works on a copy: the player's flags are never touched.
func TestReplayLeavesTheRunAlone(t *testing.T) {
	env := testEnv(t)
	env.Flags = night.Flags{"liked_mayu": true}
	env.UnlockFromOption("all")
	before := slices.Clone(env.Unlocked)
	i := cardIndex("note")
	s := &nightScene{script: sceneCards[i].Script, replay: &sceneCards[i]}
	if err := s.Enter(env); err != nil {
		t.Fatal(err)
	}
	if env.Replaying != "note" {
		t.Errorf("Replaying = %q, want note", env.Replaying)
	}
	for range 3000 {
		env.Input.InjectKey(ebiten.KeySpace)
		if err := s.Update(); err != nil {
			t.Fatal(err)
		}
		env.Input.endFrame()
		if env.nextScene != "" {
			break
		}
	}
	if env.nextScene != SceneScenes {
		t.Errorf("the replay ended in %q, want the gallery", env.nextScene)
	}
	if !reflect.DeepEqual(env.Flags, night.Flags{"liked_mayu": true}) || !slices.Equal(env.Unlocked, before) {
		t.Errorf("a replay changed the run: flags %v unlocked %v", env.Flags, env.Unlocked)
	}
	s.Exit()
	if env.Replaying != "" {
		t.Error("Replaying must be empty after the replay")
	}
}

func TestEscLeavesAReplay(t *testing.T) {
	env := testEnv(t)
	c := sceneCards[cardIndex("coffee")]
	s := &nightScene{script: c.Script, replay: &c}
	if err := s.Enter(env); err != nil {
		t.Fatal(err)
	}
	env.Input.InjectKey(ebiten.KeyEscape)
	if err := s.Update(); err != nil {
		t.Fatal(err)
	}
	if env.nextScene != SceneScenes {
		t.Errorf("after Esc: %q, want the gallery", env.nextScene)
	}
}

// Playing the story unlocks the cards it reaches, and only those.
func TestReachingARangeUnlocks(t *testing.T) {
	tests := []struct {
		script string
		line   int
		flags  night.Flags
		want   []string
	}{
		{night.Prologue, 8, nil, []string{"gurney"}},
		{night.Night1, 7, nil, nil},
		{night.Night1, 17, nil, []string{"coffee"}},
		{night.Night1, 47, nil, nil}, // the menu is past the range
		{night.Night1, 63, nil, []string{"note"}},
		{night.Night2, 60, nil, []string{"confession"}},
		{night.Night2, 80, nil, []string{"supply"}},
		{night.Night2, 130, nil, []string{"kanashibari"}},
		{night.Night3, 30, nil, []string{"hana_room"}},
		{night.Night3, 81, night.Flags{"route_mayu": true}, nil}, // the letter, but not Hana's route
		{night.Night3, 96, night.Flags{"route_hana": true}, []string{"forged"}},
		{night.EndingScript("317"), 3, nil, []string{"317"}},
		{night.EndingScript("sleep"), 0, nil, nil}, // line 0: the script is over
	}
	for _, tt := range tests {
		if got := cardsReached(tt.script, tt.line, tt.flags); !slices.Equal(got, tt.want) {
			t.Errorf("%s line %d: %v, want %v", tt.script, tt.line, got, tt.want)
		}
	}
}

func TestNoteSceneSavesOnlyWhenSomethingChanged(t *testing.T) {
	path := filepath.Join(t.TempDir(), SettingsFile)
	env := testEnv(t)
	env.SettingsPath = path
	env.Flags = night.Flags{"stay_mayu": true, "route_mayu": true, "liked_mayu": true}
	if err := env.NoteScene(night.Night2, 130); err != nil {
		t.Fatal(err)
	}
	saved, err := LoadSettings(path)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(saved.Scenes, []string{"kanashibari"}) ||
		!slices.Equal(saved.SceneFlags, []string{"liked_mayu", "stay_mayu", "route_mayu"}) {
		t.Errorf("saved %v / %v", saved.Scenes, saved.SceneFlags)
	}
	os.Remove(path)
	if err := env.NoteScene(night.Night2, 131); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err == nil {
		t.Error("nothing new happened: the file must not be written again")
	}
}

// An old settings file (no gallery fields) still loads, and junk ids are dropped.
func TestOldSettingsFileLoads(t *testing.T) {
	path := filepath.Join(t.TempDir(), SettingsFile)
	old := `{"crt": false, "reduceFlashing": true, "auto": false, "music": 0.5, "sfx": 0.5, "seenPrologue": true}`
	if err := os.WriteFile(path, []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := LoadSettings(path)
	if err != nil || got.CRT || !got.ReduceFlashing || len(got.Scenes) != 0 {
		t.Fatalf("old file: %+v, %v", got, err)
	}
	junk := `{"music": 1, "sfx": 1, "scenes": ["wake", "bogus", "gurney", "wake"]}`
	os.WriteFile(path, []byte(junk), 0o644)
	if got, _ = LoadSettings(path); !slices.Equal(got.Scenes, []string{"gurney", "wake"}) {
		t.Errorf("scenes = %v, want gurney and wake in grid order", got.Scenes)
	}
}

func TestMergeSceneFlags(t *testing.T) {
	saved := []string{"route_hana", "find_hana"}
	if got := mergeSceneFlags(saved, night.Flags{"liked_mayu": true}); !slices.Equal(got, saved) {
		t.Errorf("a run with no route yet must keep the saved flags, got %v", got)
	}
	now := night.Flags{"stay_mayu": true, "route_mayu": true, "closed_eyes": true, "stamped_self": true}
	if got := mergeSceneFlags(saved, now); !slices.Equal(got, []string{"stay_mayu", "route_mayu", "closed_eyes"}) {
		t.Errorf("a decided run replaces them: %v", got)
	}
}

func TestReplayFlags(t *testing.T) {
	routes := func(f night.Flags) []string {
		var out []string
		for _, r := range []string{"mayu", "hana", "nobody"} {
			if f.Has("route_" + r) {
				out = append(out, r)
			}
		}
		return out
	}
	mayu := []string{"stay_mayu", "route_mayu"}
	hana := []string{"find_hana", "route_hana"}
	tests := []struct {
		card  string
		saved []string
		route string
	}{
		{"kanashibari", mayu, "mayu"},                 // any route: the saved one
		{"kanashibari", hana, "hana"},                 // ... including Hana's
		{"kanashibari", nil, "nobody"},                // nothing saved: nobody's
		{"confession", hana, "mayu"},                  // Mayu's own scene
		{"supply", mayu, "hana"},                      // Hana's own scene
		{"sleep", hana, "nobody"},                     // Hana's route never ends in Sleep
		{"wake", mayu, "mayu"},                        // an allowed route is kept
		{"forever", mayu, "hana"},                     // always Hana's
		{"hana_room", []string{"route_hana"}, "hana"}, // route flag without its choice
	}
	for _, tt := range tests {
		f := replayFlags(sceneCards[cardIndex(tt.card)], tt.saved)
		if got := routes(f); !slices.Equal(got, []string{tt.route}) {
			t.Errorf("%s with %v: routes %v, want %s", tt.card, tt.saved, got, tt.route)
		}
		if tt.route != "nobody" && night.RouteFor(f) != tt.route {
			t.Errorf("%s with %v: the choices say %s, want %s", tt.card, tt.saved, night.RouteFor(f), tt.route)
		}
	}
	note := replayFlags(sceneCards[cardIndex("note")], []string{"liked_mayu"})
	if !note.Has("liked_hana") || note.Has("liked_mayu") {
		t.Errorf("Hana's note must go to the note: %v", note.Names())
	}
	// The saved list is never changed by building flags from it.
	saved := []string{"stay_mayu", "route_mayu"}
	replayFlags(sceneCards[cardIndex("supply")], saved)
	if !slices.Equal(saved, []string{"stay_mayu", "route_mayu"}) {
		t.Error("replayFlags changed the saved flags")
	}
}

func TestCounter(t *testing.T) {
	env := testEnv(t)
	if seen, total := env.SceneCounts(); seen != 0 || total != 11 {
		t.Errorf("counter %d / %d, want 0 / 11", seen, total)
	}
	env.UnlockFromOption("gurney,coffee")
	if seen, total := env.SceneCounts(); seen != 2 || total != 11 {
		t.Errorf("counter %d / %d, want 2 / 11", seen, total)
	}
	env.UnlockFromOption("all")
	if seen, total := env.SceneCounts(); seen != 12 || total != 12 {
		t.Errorf("counter %d / %d, want 12 / 12", seen, total)
	}
	if err := env.UnlockFromOption("nope"); err == nil {
		t.Error("an unknown card id must be an error")
	}
}

func TestGridHitTesting(t *testing.T) {
	tests := []struct {
		p    image.Point
		want int
	}{
		{image.Pt(18, 25), 0},    // top-left corner of the first frame
		{image.Pt(119, 82), 0},   // bottom-right of it (origin + 102, + 58)
		{image.Pt(120, 82), -1},  // one past
		{image.Pt(17, 24), -1},   // the slot corner is outside the 1-px frame
		{image.Pt(132, 25), 1},   // second card
		{image.Pt(370, 190), 11}, // the 3:17 card
		{image.Pt(240, 135), -1}, // the gap the test-mode cursor rests in
		{image.Pt(50, 84), -1},   // the title strip under a card
	}
	for _, tt := range tests {
		if got := cardAt(tt.p); got != tt.want {
			t.Errorf("cardAt(%v) = %d, want %d", tt.p, got, tt.want)
		}
	}
	if o := cardOrigin(11); o != image.Pt(359, 172) {
		t.Errorf("origin of card 11 = %v, want (359, 172)", o)
	}
}

func TestMoveSel(t *testing.T) {
	tests := []struct{ sel, dx, dy, want int }{
		{0, 1, 0, 1}, {3, 1, 0, 3}, {0, -1, 0, 0}, {0, 0, -1, 0},
		{0, 0, 1, 4}, {5, 0, 1, 9}, {9, 0, 1, backSel},
		{backSel, 0, -1, 8}, {backSel, 1, 0, backSel}, {11, 0, 1, backSel},
	}
	for _, tt := range tests {
		if got := moveSel(tt.sel, tt.dx, tt.dy); got != tt.want {
			t.Errorf("moveSel(%d, %d, %d) = %d, want %d", tt.sel, tt.dx, tt.dy, got, tt.want)
		}
	}
}

func TestTearStutter(t *testing.T) {
	var frozen, jumped int
	for tick := 0; tick < stutterEvery; tick++ {
		f, j := stutterAt(tick, false)
		if f && j {
			t.Fatalf("tick %d is frozen and jumped", tick)
		}
		if f {
			frozen++
		}
		if j {
			jumped++
		}
	}
	if frozen != 6 || jumped != 2 {
		t.Errorf("per 1.2 s: %d frozen ticks and %d jumped, want 6 and 2", frozen, jumped)
	}
	if f, j := stutterAt(stutterEvery, false); f || j {
		t.Error("a new cycle starts calm")
	}
	for tick := 0; tick < stutterEvery; tick++ {
		if _, j := stutterAt(tick, true); j {
			t.Fatal("reduce flashing must drop the 2 px jump")
		}
	}
	if _, last := stutterAt(stutterEvery-1, false); !last {
		t.Error("the jump comes right after the freeze, at the end of the cycle")
	}
}

// The tear's timing follows the sheet: 140 / 90 / 60 / 180 ms.
func TestTearFramesFollowTheSheet(t *testing.T) {
	env := testEnv(t)
	sheet, err := LoadSheet(env.Sprites, "scenes_gallery")
	if err != nil {
		t.Fatal(err)
	}
	at := func(ms int) int { return sheet.FrameAt(tearTag, time.Duration(ms)*time.Millisecond) }
	for ms, want := range map[int]int{0: 0, 139: 0, 140: 1, 229: 1, 230: 2, 289: 2, 290: 3, 469: 3, 470: 0} {
		if got := at(ms); got != want {
			t.Errorf("frame at %d ms = %d, want %d", ms, got, want)
		}
	}
	if ripStartTicks != 14 { // 230 ms: the start of frame 3
		t.Errorf("the rip starts at %d ticks", ripStartTicks)
	}
	// The locked glint: frame 1 rests, then frames 2 to 4 play once.
	seen := map[int]bool{}
	for tick := 0; ; tick++ {
		f, over := glintFrame(sheet.Meta, tick, false)
		if over {
			if f != 0 || tick < 19 || tick > 21 {
				t.Errorf("the glint ended at tick %d on frame %d, want ~20 ticks (330 ms) on frame 0", tick, f)
			}
			break
		}
		seen[f] = true
		if tick > 100 {
			t.Fatal("the glint never ends")
		}
	}
	if seen[0] || !seen[1] || !seen[2] || !seen[3] {
		t.Errorf("the glint played frames %v, want 1, 2 and 3 (the sheet's frames 2 to 4)", seen)
	}
	slow, _ := glintFrame(sheet.Meta, 20, true)
	if _, over := glintFrame(sheet.Meta, 20, true); over || slow == 0 {
		t.Error("at half speed, tick 20 is still inside the glint")
	}
}

// Locked, the crack glints every 3 to 5 s; found, it loops.
func TestGalleryTearStates(t *testing.T) {
	env := testEnv(t)
	s := &galleryScene{}
	if err := s.Enter(env); err != nil {
		t.Fatal(err)
	}
	glints, rest := 0, 0
	for range 60 * 20 {
		s.updateTear()
		if s.tearFrame == 0 {
			rest++
		} else if s.glintTicks == 0 || s.glintTicks == 1 {
			glints++
		}
	}
	if glints < 3 || glints > 7 || rest < 60*15 {
		t.Errorf("in 20 s: %d glint starts, %d resting ticks; want 4 to 6 glints and mostly rest", glints, rest)
	}
	env.UnlockFromOption("317")
	seen := map[int]bool{}
	for range 60 {
		s.updateTear()
		seen[s.tearFrame] = true
	}
	if len(seen) != 4 {
		t.Errorf("the found tear loops %v, want all four frames in a second", seen)
	}
}

// The title now has START, SCENES and SETTINGS, and the menu clears the warnings.
func TestTitleMenuHasScenes(t *testing.T) {
	if !slices.Equal(titleItems, []string{"START", "SCENES", "SETTINGS"}) {
		t.Fatalf("titleItems = %v", titleItems)
	}
	if titleStartY != 207 || titleStepY != 16 || titleRuleDY != -16 {
		t.Errorf("menu at y %d step %d rule %d, want 207, 16, -16 (Figma 52:471)", titleStartY, titleStepY, titleRuleDY)
	}
	lastRowBottom := titleStartY + (len(titleItems)-1)*titleStepY + capTop + 5
	firstWarningTop := ScreenH - warningBase - warningStep + capTop
	if lastRowBottom >= firstWarningTop {
		t.Errorf("SETTINGS letters end at y %d, the first warning's start at %d", lastRowBottom, firstWarningTop)
	}
	env := testEnv(t)
	s := &titleScene{}
	if err := s.Enter(env); err != nil {
		t.Fatal(err)
	}
	for range titleMenuFrom {
		s.Update()
	}
	env.Input.InjectKey(ebiten.KeyArrowDown)
	s.Update()
	env.Input.endFrame()
	env.Input.InjectKey(ebiten.KeyEnter)
	s.Update()
	if env.nextScene != SceneScenes || !env.TitleSettled {
		t.Errorf("SCENES went to %q (settled %v), want the gallery", env.nextScene, env.TitleSettled)
	}
}

// Enter on a seen card starts its replay, on a locked card does nothing, and
// Esc goes back to the title.
func TestGalleryPicks(t *testing.T) {
	env := testEnv(t)
	env.UnlockFromOption("gurney")
	s := &galleryScene{}
	if err := s.Enter(env); err != nil {
		t.Fatal(err)
	}
	env.Input.InjectKey(ebiten.KeyArrowRight) // coffee: locked
	s.Update()
	env.Input.endFrame()
	env.Input.InjectKey(ebiten.KeyEnter)
	s.Update()
	env.Input.endFrame()
	if env.nextScene != "" {
		t.Fatalf("a locked card started %q", env.nextScene)
	}
	env.Input.InjectKey(ebiten.KeyArrowLeft)
	s.Update()
	env.Input.endFrame()
	env.Input.InjectKey(ebiten.KeySpace)
	s.Update()
	env.Input.endFrame()
	if env.nextScene != "replay.gurney" {
		t.Fatalf("a seen card went to %q, want replay.gurney", env.nextScene)
	}
	env.nextScene = ""
	env.Input.InjectKey(ebiten.KeyEscape)
	s.Update()
	if env.nextScene != SceneTitle {
		t.Errorf("Esc went to %q, want the title", env.nextScene)
	}
}
