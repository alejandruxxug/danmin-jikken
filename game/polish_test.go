package game

import (
	"danmingame/night"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// The blink is slow and sleepy (design log, 2026-10-13): 0.25 s shut, every
// 4 to 8 s.
func TestBlinkDesignValues(t *testing.T) {
	if blinkFor != 0.25 || blinkMin != 4 || blinkMin+blinkSpread != 8 {
		t.Errorf("blink = %v s every %v to %v s, want 0.25 s every 4 to 8 s", blinkFor, blinkMin, blinkMin+blinkSpread)
	}
}

// Every visitor has a window sheet with the layers the desk draws.
func TestWindowSheetsHaveTheirLayers(t *testing.T) {
	fsys, err := SpriteFS()
	if err != nil {
		t.Fatal(err)
	}
	for day := 1; day <= LastDay(); day++ {
		rules, _ := RulesForDay(day)
		for _, v := range rules.Visitors {
			sheet, err := LoadSheet(fsys, v.Sprite)
			if err != nil {
				t.Errorf("%s: %v", v.Name, err)
				continue
			}
			for _, layers := range [][]string{windowAwake(v), patientAsleep, eyesFor(windowAwake(v), true)} {
				for _, l := range layers {
					if _, err := sheet.Rect(l, 0); err != nil {
						t.Errorf("%s (%s): %v", v.Name, v.Sprite, err)
					}
				}
			}
		}
	}
	for day := 1; day <= LastDay(); day++ {
		sheet, err := LoadSheet(fsys, "desk")
		if err != nil {
			t.Fatal(err)
		}
		for _, l := range deskLayers(day) {
			if _, err := sheet.Rect(l, 0); err != nil {
				t.Errorf("day %d desk: %v", day, err)
			}
		}
	}
}

// Day 1: Old Cecily nods off at the window. The stamp is refused until the
// player clicks her awake; the hospital never hears about it; it happens
// once and not on other days.
func TestDay1NodOffTutorial(t *testing.T) {
	sh := newTestShift(t, 1)
	runTicks(sh, 60*10)
	cecily := sh.Guests[0]
	if sh.Dozing() != cecily || sh.LookOf(cecily) != LookDozing {
		t.Fatalf("dozing = %v, look = %v; want Old Cecily asleep at the window", sh.Dozing(), sh.LookOf(cecily))
	}
	if !strings.Contains(sh.Message(), "CLICK HER") {
		t.Errorf("message = %q, want the click hint", sh.Message())
	}
	if err := sh.Stamp(StampAdmit); !errors.Is(err, ErrDozing) {
		t.Errorf("stamp while she sleeps: %v, want ErrDozing", err)
	}
	if n := len(sh.Snapshot().Patients); n != 0 {
		t.Errorf("%d patients in the hospital, want 0 (she is not admitted)", n)
	}
	if sh.needsStaff(cecily) {
		t.Error("the window guest must not count as a sleeper who needs staff")
	}
	if err := sh.Wake(cecily); err != nil {
		t.Fatal(err)
	}
	runTicks(sh, 60*10)
	if sh.Dozing() != nil || sh.LookOf(cecily) != LookStanding {
		t.Error("she must stay awake after the click: the nod-off is once only")
	}
	if err := sh.Stamp(StampAdmit); err != nil {
		t.Errorf("stamp after waking her: %v", err)
	}
	if err := sh.Wake(cecily); !errors.Is(err, ErrNothingToAttend) {
		t.Errorf("waking an awake guest: %v, want ErrNothingToAttend", err)
	}

	other := newTestShift(t, 2)
	runTicks(other, 60*10)
	if other.Dozing() != nil {
		t.Error("only day 1 has the scripted nod-off")
	}
}

// The picture is scaled by a whole number and centred; the rest is black.
func TestFitScreen(t *testing.T) {
	cases := []struct{ w, h, scale, offX, offY int }{
		{480, 270, 1, 0, 0},
		{1920, 1080, 4, 0, 0},
		{3840, 2160, 8, 0, 0},
		{2560, 1440, 5, 80, 45}, // 2400 x 1350 picture, bars around
		{1920, 1200, 4, 0, 60},  // taller than 16:9: bars above and below
		{300, 200, 1, -90, -35}, // smaller than the picture: never below x1
	}
	for _, c := range cases {
		scale, offX, offY := fitScreen(c.w, c.h)
		if scale != c.scale || offX != c.offX || offY != c.offY {
			t.Errorf("fitScreen(%d, %d) = %d, %d, %d; want %d, %d, %d", c.w, c.h, scale, offX, offY, c.scale, c.offX, c.offY)
		}
	}
}

func TestFloorDiv(t *testing.T) {
	for _, c := range []struct{ a, b, want int }{{9, 4, 2}, {8, 4, 2}, {0, 4, 0}, {-1, 4, -1}, {-4, 4, -1}, {-5, 4, -2}} {
		if got := floorDiv(c.a, c.b); got != c.want {
			t.Errorf("floorDiv(%d, %d) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestFullscreenOption(t *testing.T) {
	o, err := ParseOptions([]string{"-fullscreen"})
	if err != nil {
		t.Fatal(err)
	}
	if !o.Fullscreen || o.Enabled {
		t.Errorf("-fullscreen alone gave %+v; it should work without -agent", o)
	}
	o, err = ParseOptions([]string{"-scene", "day.top10", "-fullscreen"})
	if err != nil {
		t.Fatal(err)
	}
	if o.Scene != "" || !o.Fullscreen {
		t.Errorf("without -agent only -fullscreen is kept, got %+v", o)
	}
}

// A shot or dump shows the state before the commands after it, so the
// script stops there for the frame.
func TestShotAndDumpComeBeforeLaterCommands(t *testing.T) {
	cmds, err := ParseScript(strings.NewReader("dump a.json\nclick 5 5\nshot b.png\npan 0"))
	if err != nil {
		t.Fatal(err)
	}
	env := &Env{Input: NewInput(), CameraX: 100}
	a := &Agent{script: cmds}
	a.runCommands(env)
	if len(a.pending) != 1 || a.next != 1 || env.Input.clicked {
		t.Fatalf("after frame 1: pending %d, next %d, clicked %v; want the dump only", len(a.pending), a.next, env.Input.clicked)
	}
	a.pending = nil
	a.runCommands(env)
	if !env.Input.clicked || len(a.pending) != 0 {
		t.Fatalf("frame 2 should run the click and stop: clicked %v, pending %d", env.Input.clicked, len(a.pending))
	}
	env.Input.endFrame()
	a.runCommands(env)
	if len(a.pending) != 1 || env.CameraX != 100 {
		t.Errorf("frame 3: pending %d, camera %v; want the shot, camera untouched until frame 4", len(a.pending), env.CameraX)
	}
}

func TestShortNameCutsRunes(t *testing.T) {
	if got := shortName("Åsa Jürgensen Ñandú"); got != "Åsa Jürgense" {
		t.Errorf("shortName = %q", got)
	}
	if got := shortName("alejo"); got != "alejo" {
		t.Errorf("shortName = %q", got)
	}
}

// Closing the computer forgets the tab.
func TestComputerExitResetsTab(t *testing.T) {
	env := &Env{Tab: TabWardFeed}
	s := &computerScene{env: env}
	s.Exit()
	if env.Tab != "" {
		t.Errorf("tab = %q after closing, want empty", env.Tab)
	}
}

// The clock stands still on the top-10.
func TestClockStopsOnTop10(t *testing.T) {
	g, err := New(Options{Enabled: true, Scene: SceneDayTop10, Seed: 7})
	if err != nil {
		t.Fatal(err)
	}
	g.env.Clock.minutes = 17 * 60
	before := g.env.Clock.String()
	for range 60 * 30 {
		if err := g.Update(); err != nil {
			t.Fatal(err)
		}
	}
	if g.env.Clock.String() != before {
		t.Errorf("clock moved from %s to %s on the top-10", before, g.env.Clock)
	}
}

// The title waits for input; START begins day 1 and the day turns on like a CRT.
func TestTitleStartsDayOne(t *testing.T) {
	fsys, err := SpriteFS()
	if err != nil {
		t.Fatal(err)
	}
	env := &Env{Input: NewInput(), Day: 3, Sprites: fsys}
	s := &titleScene{}
	if err := s.Enter(env); err != nil {
		t.Fatal(err)
	}
	env.Input.InjectKey(ebiten.KeySpace)
	s.Update() // too early: the logo is still falling
	if env.nextScene != "" {
		t.Fatal("the title must not answer while the logo falls")
	}
	for range titleMenuFrom {
		s.Update()
	}
	env.Input.InjectKey(ebiten.KeySpace)
	s.Update()
	if env.nextScene != SceneDayReception || env.Day != 1 || !env.PoweringOn {
		t.Errorf("next %q day %d poweringOn %v, want day.reception day 1 with the CRT turn-on", env.nextScene, env.Day, env.PoweringOn)
	}
}

func TestTitleDropFallsThenSettles(t *testing.T) {
	if titleDrop(0) >= 0 || titleDrop(titleFall+titleBounce) != 0 {
		t.Errorf("the logo starts above the screen (%d) and ends in place (%d)", titleDrop(0), titleDrop(titleFall+titleBounce))
	}
	for ticks := 1; ticks < titleFall; ticks++ {
		if titleDrop(ticks) <= titleDrop(ticks-1) {
			t.Fatalf("tick %d: the logo must keep falling", ticks)
		}
	}
	if dx, dy := titleShakeOffset(titleFall + titleShake); dx != 0 || dy != 0 {
		t.Error("the shake must be over by the end of its table")
	}
}

// The cold open: the Prologue ends in the title (with its white), a later
// launch may skip the Prologue with a key, and a fresh run never repeats it.
func TestColdOpen(t *testing.T) {
	env := &Env{Day: 1, Flags: night.Flags{}}
	s := &nightScene{env: env, script: night.Prologue}
	if err := s.finish("day1"); err != nil {
		t.Fatal(err)
	}
	if env.nextScene != SceneTitle || !env.TitleWhite || !env.SeenPrologue {
		t.Errorf("after the Prologue: next %q white %v seen %v, want the title, with the white, remembered", env.nextScene, env.TitleWhite, env.SeenPrologue)
	}
}

func TestSettingsFileRoundTrip(t *testing.T) {
	path := SettingsPathFor(filepath.Join(t.TempDir(), ScoresFile))
	got, err := LoadSettings(path)
	if err != nil || got != DefaultSettings() {
		t.Fatalf("a missing file must give the defaults: %+v, %v", got, err)
	}
	want := Settings{CRT: false, ReduceFlashing: true, Auto: true, Music: 0.3, SFX: 1, SeenPrologue: true}
	if err := SaveSettings(path, want); err != nil {
		t.Fatal(err)
	}
	if got, err = LoadSettings(path); err != nil || got != want {
		t.Errorf("read back %+v, %v; want %+v", got, err, want)
	}
	if SettingsPathFor("") != "" || SaveSettings("", want) != nil {
		t.Error("test mode (no scores file) must save nothing")
	}
}

func TestStepVolumeStaysInRange(t *testing.T) {
	if stepVolume(1, 1) != 1 || stepVolume(0, -1) != 0 || stepVolume(0.5, 1) != 0.6 {
		t.Errorf("volumes: %v %v %v", stepVolume(1, 1), stepVolume(0, -1), stepVolume(0.5, 1))
	}
}

func TestPickScale(t *testing.T) {
	cases := []struct{ w, h, want int }{
		{3840, 2160, 4}, // 4K: x4 is 1920x1080
		{1920, 1080, 3}, // x4 would not fit with the menu bar: x3 = 1440x810
		{1440, 900, 2},  // a small MacBook: x3 is 810 high, over 90% of 900
		{1512, 982, 2},  // x3 is 1440 wide, over 90% of 1512
		{1728, 1117, 3},
		{800, 600, 1},
		{0, 0, 1},
	}
	for _, c := range cases {
		if got := pickScale(c.w, c.h); got != c.want {
			t.Errorf("pickScale(%d, %d) = %d, want %d", c.w, c.h, got, c.want)
		}
	}
}

// A later launch skips the Prologue with any key; the first launch does not.
func TestPrologueSkipOnlyWhenSeen(t *testing.T) {
	for _, seen := range []bool{false, true} {
		g, err := New(Options{Enabled: true, Scene: SceneNightPrologue, SeenPrologue: seen})
		if err != nil {
			t.Fatal(err)
		}
		for range 5 {
			if err := g.Update(); err != nil {
				t.Fatal(err)
			}
		}
		if d := BuildDump(g.env, g.scenes.ID(), 1); !d.Dream {
			t.Error("the Prologue must be a dream (washed out)")
		}
		g.env.Input.InjectKey(ebiten.KeyB)
		if err := g.Update(); err != nil {
			t.Fatal(err)
		}
		if skipped := g.scenes.ID() == SceneTitle; skipped != seen {
			t.Errorf("seen=%v: scene after a key is %q", seen, g.scenes.ID())
		}
	}
}

// The arrow keys move the Settings selection even when the cursor rests on
// a row (as it does in test mode), and Enter changes the selected row.
func TestSettingsKeysWorkWithRestingCursor(t *testing.T) {
	env := testEnv(t)
	env.Input.IgnoreReal = true // the cursor rests mid-screen, over a row
	s := &settingsScene{}
	if err := s.Enter(env); err != nil {
		t.Fatal(err)
	}
	press := func(k ebiten.Key) {
		env.Input.InjectKey(k)
		if err := s.Update(); err != nil {
			t.Fatal(err)
		}
		env.Input.endFrame()
	}
	press(ebiten.KeyArrowDown)
	press(ebiten.KeyArrowDown)
	if s.sel != 2 {
		t.Fatalf("after two Down presses sel = %d, want 2", s.sel)
	}
	press(ebiten.KeyArrowUp)
	if s.sel != 1 {
		t.Errorf("after Up sel = %d, want 1", s.sel)
	}
	was := env.ReduceFlashing
	press(ebiten.KeyEnter)
	if env.ReduceFlashing == was {
		t.Error("Enter on REDUCE FLASHING should flip it")
	}
}

// Both content warnings fit across the title screen, below the menu.
func TestTitleWarningsFit(t *testing.T) {
	font, err := LoadFont()
	if err != nil {
		t.Fatal(err)
	}
	lines := 0
	for _, w := range titleWarnings {
		for _, line := range night.WrapText(w, warningWidth, font.Width) {
			if font.Width(line) > warningWidth {
				t.Errorf("warning line %q is %d px wide, more than %d", line, font.Width(line), warningWidth)
			}
			lines++
		}
	}
	if lines != 2 {
		t.Errorf("the warnings take %d lines, want 2 (one each)", lines)
	}
}
