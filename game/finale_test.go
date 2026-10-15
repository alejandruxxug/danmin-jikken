package game

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"danmingame/night"
)

// The stages follow the spec's seconds: the roll runs 33.4 s from 5.5 s, the
// thanks holds 3 s, the logo 5 s, the fade 1.5 s, then 1 s of black.
func TestFinaleStageTable(t *testing.T) {
	roll := finaleStages[1].Ticks // derived from the credits' layout (rollTicks)
	for _, c := range []struct {
		tick  int
		stage string
		in    int
	}{
		{0, "shade-in", 0},
		{29, "shade-in", 29},
		{30, "roll", 0},                   // 5.5 s after the title pop
		{30 + roll - 1, "roll", roll - 1}, // the last tick of the roll
		{30 + roll, "thanks", 0},
		{30 + roll + 179, "thanks", 179},
		{30 + roll + 180, "shade-out", 0}, // the thanks goes off
		{30 + roll + 210, "logo", 0},
		{30 + roll + 510, "fade", 0},
		{30 + roll + 600, "black", 0},
		{30 + roll + 659, "black", 60 - 1}, // the last tick before the turn-on
	} {
		st, ok := FinaleAt(c.tick)
		if !ok || st.Stage != c.stage || st.Tick != c.in || st.At != c.tick {
			t.Errorf("FinaleAt(%d) = %+v, %v; want %s %d", c.tick, st, ok, c.stage, c.in)
		}
	}
	if _, ok := FinaleAt(FinaleTicks()); ok {
		t.Error("the finale goes on past its end")
	}
	if _, ok := FinaleAt(-1); ok {
		t.Error("the finale starts before tick 0")
	}
	// Everything after the roll is fixed: 30 ticks of shade-in before it, 660 after.
	if FinaleTicks() != 30+roll+660 {
		t.Errorf("the finale is %d ticks, want %d", FinaleTicks(), 30+roll+660)
	}
}

// The shade steps in 0.25 s at a time up to the ending's strongest frame and
// out the same way; the fade steps 0.5 s at a time.
func TestShadeAndFadeFrames(t *testing.T) {
	at := func(tick int) FinaleState {
		st, _ := FinaleAt(tick)
		return st
	}
	out := FinaleStageStart("shade-out")
	for _, c := range []struct {
		name string
		tick int
		max  int
		want int
	}{
		{"half shade, first step", 0, 1, 0},
		{"half shade, second step", 15, 1, 1},
		{"half shade holds", 29, 1, 1},
		{"solid shade, third step", 29, 2, 1},
		{"solid shade in the roll", 100, 2, 2},
		{"half shade in the thanks", FinaleStageStart("thanks") + 5, 1, 1},
		{"half shade steps out", out, 1, 0},
		{"half shade gone", out + 15, 1, -1},
		{"solid shade steps out", out, 2, 1},
		{"solid shade, next step", out + 15, 2, 0},
		{"no shade under the logo", FinaleStageStart("logo"), 2, -1},
		{"no shade in the fade", FinaleStageStart("fade"), 1, -1},
	} {
		if got := shadeFrame(at(c.tick), c.max); got != c.want {
			t.Errorf("%s: shadeFrame = %d, want %d", c.name, got, c.want)
		}
	}
	fade := FinaleStageStart("fade")
	for _, c := range []struct{ tick, want int }{
		{fade - 1, -1}, {fade, 0}, {fade + 29, 0}, {fade + 30, 1}, {fade + 60, 2}, {fade + 89, 2},
		{FinaleStageStart("black"), 3}, {FinaleTicks() - 1, 3},
	} {
		if got := fadeFrame(at(c.tick)); got != c.want {
			t.Errorf("fadeFrame at %d = %d, want %d", c.tick, got, c.want)
		}
	}
}

// The roll's layout is the Figma one: heading to line 10 px, line to line 10,
// last line to the next heading 24, the thanks 50 below the last line, and
// the thanks reaches y 130 exactly when the roll stops.
func TestRollLayoutAndPositions(t *testing.T) {
	lines := rollLayout(CreditCards(night.EndingForever))
	at := map[string]int{}
	for _, l := range lines {
		at[l.Text] = l.Y
	}
	for text, y := range map[string]int{
		"DANMIN JIKKEN":               0,
		"A game for the course":       24,
		"Made by":                     58,
		"Art":                         126,
		"Music and sound":             204,
		"Halcyon Rest":                378,
		"Drift off well.":             388,
		"Thank you for staying awake": 472,
	} {
		if got, ok := at[text]; !ok || got != y {
			t.Errorf("%q is at y %d (found %v), want %d", text, got, ok, y)
		}
	}
	headings := 0
	for _, l := range lines {
		if l.Text == "Music and sound" {
			headings++
		}
	}
	if headings != 1 {
		t.Errorf("the music lines sit under %d headings, want 1", headings)
	}
	rollEnd := finaleStages[0].Ticks + finaleStages[1].Ticks
	if y := rollStartY + thanksY(lines) - rollDY(rollEnd); y != thanksStopY {
		t.Errorf("the thanks stops with its caps top at y %d, want %d", y, thanksStopY)
	}
	if rollDY(0) != 0 || rollDY(finaleStages[0].Ticks) != 0 || rollDY(finaleStages[0].Ticks+60) != 16 {
		t.Error("the roll does not climb 16 px per second from the end of the shade-in")
	}
	if rollDY(rollEnd+500) != rollDY(rollEnd) {
		t.Error("the roll goes on climbing under the thanks")
	}
}

// Each ending has its own look.
func TestFinaleStyles(t *testing.T) {
	forever, sleep, wake, secret := finaleStyleFor(night.EndingForever), finaleStyleFor(night.EndingSleep), finaleStyleFor(night.EndingWake), finaleStyleFor(night.Ending317)
	if forever.Head != colBrightRed || !forever.HeadOutline || forever.ShadeMax != 1 || forever.Logo != logoBlood {
		t.Errorf("forever %+v", forever)
	}
	if secret.Head != colBrightRed || !secret.HeadOutline || secret.ShadeMax != 1 || secret.Logo != logoDark {
		t.Errorf("3:17 %+v", secret)
	}
	if wake.ShadeLayer != "shade_dark" || wake.ShadeMax != 2 || wake.Head != colLightGrey || wake.Name != colOffWhite || wake.HeadOutline || wake.Logo != logoTitle {
		t.Errorf("wake %+v", wake)
	}
	if sleep.ShadeLayer != "shade_light" || sleep.ShadeMax != 2 || sleep.Name != colInk || !sleep.HeadShadow || sleep.Logo != logoTitle {
		t.Errorf("sleep %+v", sleep)
	}
}

// The slipped band is the same for the same tick, only lasts two frames, and
// is an 8 px row that moves 3 to 6 px.
func TestBandSlip(t *testing.T) {
	seen := 0
	for tick := 0; tick < 60*60; tick++ {
		y, dx, ok := bandSlip(tick, 7)
		if y2, dx2, ok2 := bandSlip(tick, 7); y != y2 || dx != dx2 || ok != ok2 {
			t.Fatalf("tick %d gives two answers", tick)
		}
		if !ok {
			continue
		}
		seen++
		if tick%slipWindow >= 2 {
			t.Errorf("tick %d slips outside the two frames", tick)
		}
		if y%slipBand != 0 || y < 0 || y+slipBand > ScreenH {
			t.Errorf("band at y %d is not an 8 px row on the screen", y)
		}
		if a := max(dx, -dx); a < 3 || a > 6 {
			t.Errorf("slip of %d px, want 3 to 6", dx)
		}
	}
	if seen == 0 || seen > 60*60/slipWindow*2 {
		t.Errorf("%d slipped frames in a minute: not 'now and then'", seen)
	}
}

// enterFinale opens an ending scene and jumps it to a tick of its finale.
func enterFinale(t *testing.T, ending string, tick int) (*endingScene, *Env) {
	t.Helper()
	env := testEnv(t)
	s := &endingScene{name: ending}
	if err := s.Enter(env); err != nil {
		t.Fatal(err)
	}
	env.JumpToFinale, env.FinaleJumpTick = true, tick
	if err := s.Update(); err != nil {
		t.Fatal(err)
	}
	return s, env
}

// After the black the TV turns on onto the top-10, for every ending.
func TestFinaleEndsWithTheTurnOnOntoTheTop10(t *testing.T) {
	for _, ending := range night.Endings {
		s, env := enterFinale(t, ending, FinaleTicks()-3)
		if env.EndingPhase != "finale" || env.EndingStage != "black" {
			t.Fatalf("%s: phase %q stage %q, want finale and black", ending, env.EndingPhase, env.EndingStage)
		}
		for i := 0; i < 5 && env.nextScene == ""; i++ {
			if err := s.Update(); err != nil {
				t.Fatal(err)
			}
		}
		if env.nextScene != SceneDayTop10 || !env.PoweringOn || env.PowerOnTick != 0 {
			t.Errorf("%s: next %q, powering on %v at %d; want the top-10 and the turn-on at its start", ending, env.nextScene, env.PoweringOn, env.PowerOnTick)
		}
	}
}

// An audition ends with the finale: no top-10, no turn-on.
func TestFinaleAuditionEndsTheGame(t *testing.T) {
	s, env := enterFinale(t, night.EndingWake, FinaleTicks()-1)
	env.Audition = true
	if err := s.Update(); !errors.Is(err, ebiten.Termination) {
		t.Errorf("audition update = %v, want ebiten.Termination", err)
	}
	if env.nextScene != "" || env.PoweringOn {
		t.Error("an audition must not go on to the top-10")
	}
}

// The ending plays by itself: a click or Space does not skip the finale.
func TestClicksDoNotSkipTheFinale(t *testing.T) {
	s, env := enterFinale(t, night.EndingForever, 100)
	for i := 0; i < 3; i++ {
		env.Input.InjectClick(100, 100)
		env.Input.InjectKey(ebiten.KeySpace)
		if err := s.Update(); err != nil {
			t.Fatal(err)
		}
		env.Input.endFrame()
	}
	if env.nextScene != "" || env.EndingStage != "roll" || env.EndingTick != 103 {
		t.Errorf("after three clicks: next %q, stage %q tick %d; want the roll at 103", env.nextScene, env.EndingStage, env.EndingTick)
	}
}

// Every stage of every ending draws without an error (the sheets and fonts are there).
func TestFinaleDrawsEveryStage(t *testing.T) {
	screen := ebiten.NewImage(ScreenW, ScreenH)
	for _, ending := range night.Endings {
		s, _ := enterFinale(t, ending, 0)
		for _, stage := range finaleStages {
			s.ticks = FinaleStageStart(stage.Name) + stage.Ticks/2
			s.Draw(screen)
		}
	}
}

// The Sleep ending is shown through the day CRT from the white reel on (the
// day desk), and its night shots are not. The other endings never are.
func TestSleepEndingUsesTheDayCRT(t *testing.T) {
	for _, ending := range night.Endings {
		env := testEnv(t)
		s := &endingScene{name: ending}
		if err := s.Enter(env); err != nil {
			t.Fatal(err)
		}
		day, nightShots := 0, 0
		for i := 0; i < 1400 && s.phase == phaseScript; i++ {
			if err := s.Update(); err != nil {
				t.Fatal(err)
			}
			bg := s.script.engine.BG()
			want := ending == night.EndingSleep && (bg == "white" || bg == "desk_day")
			if got := s.UsesCRT(); got != want {
				t.Fatalf("%s: UsesCRT = %v on bg %q, want %v", ending, got, bg, want)
			}
			if want {
				day++
			} else {
				nightShots++
			}
		}
		if nightShots == 0 || (ending == night.EndingSleep) != (day > 0) {
			t.Errorf("%s: %d day frames, %d night frames", ending, day, nightShots)
		}
	}
}

// The first run that reaches its top-10 turns the hints off for good, for
// every ending, and saves that (bug found by the tester: only endShift did).
func TestReachingTheTop10TurnsHintsOff(t *testing.T) {
	for _, ending := range night.Endings {
		path := filepath.Join(t.TempDir(), SettingsFile)
		s, env := enterFinale(t, ending, FinaleTicks()-3)
		env.SettingsPath = path
		env.TutorialHints = true
		for i := 0; i < 5 && env.nextScene == ""; i++ {
			if err := s.Update(); err != nil {
				t.Fatal(err)
			}
		}
		saved, err := LoadSettings(path)
		if err != nil {
			t.Fatal(err)
		}
		if env.TutorialHints || saved.TutorialHints {
			t.Errorf("%s: hints still on after the top-10 (env %v, saved %v)", ending, env.TutorialHints, saved.TutorialHints)
		}
	}
}

// An audition never touches the hints setting.
func TestAuditionKeepsTheHints(t *testing.T) {
	s, env := enterFinale(t, night.EndingWake, FinaleTicks()-1)
	env.Audition, env.TutorialHints = true, true
	_ = s.Update()
	if !env.TutorialHints {
		t.Error("an audition turned the hints off")
	}
}
