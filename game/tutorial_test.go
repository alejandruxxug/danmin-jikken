package game

import (
	"fmt"
	"testing"
)

// The reason on the citation slip, for every visitor of every day.
func TestReasonOfEveryVisitor(t *testing.T) {
	special := map[string]string{ // day and name -> reason; everyone else has valid papers
		"1 Wren Calloway": reasonExpired,
		"2 Ozzie Pike":    reasonNoBarrow,
		"2 Marcus Trell":  reasonName,
		"3 Pearl Wick":    reasonDiagnosis,
		"3 Old Cecily":    reasonWristband,
		"3 YOU":           reasonWristband,
	}
	for day := 1; day <= LastDay(); day++ {
		rules, err := RulesForDay(day)
		if err != nil {
			t.Fatal(err)
		}
		for _, v := range rules.Visitors {
			want, ok := special[fmt.Sprintf("%d %s", day, v.Name)]
			if !ok {
				want = reasonPapersOK
			}
			got := v.Reason()
			if got != want {
				t.Errorf("day %d %s: reason %q, want %q", day, v.Name, got, want)
			}
			// The reason and the stamp that scores always agree.
			wantStamp := map[string]Stamp{
				reasonExpired: StampRefer, reasonName: StampRefer, reasonNoBarrow: StampRefer,
				reasonDiagnosis: StampRefer, reasonWristband: StampDischarge, reasonPapersOK: StampAdmit,
			}[got]
			if v.RightStamp() != wantStamp {
				t.Errorf("day %d %s: reason %q but the right stamp is %v", day, v.Name, got, v.RightStamp())
			}
		}
	}
}

// A DISCHARGE for somebody with valid papers and no wristband says so.
func TestCitationReasonNoWristband(t *testing.T) {
	rules, _ := RulesForDay(1)
	matty := rules.Visitors[1]
	if got := CitationReason(matty, StampDischarge); got != reasonNoWristband {
		t.Errorf("discharging Matty: %q, want %q", got, reasonNoWristband)
	}
	if got := CitationReason(matty, StampRefer); got != reasonPapersOK {
		t.Errorf("referring Matty: %q, want %q", got, reasonPapersOK)
	}
	wren := rules.Visitors[2]
	if got := CitationReason(wren, StampDischarge); got != reasonExpired {
		t.Errorf("discharging Wren: %q, want %q", got, reasonExpired)
	}
}

// Every hint line fits the box.
func TestHintLinesFit(t *testing.T) {
	font, err := LoadFont()
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, h := range hintList {
		if seen[h.ID] {
			t.Errorf("two hints are called %q", h.ID)
		}
		seen[h.ID] = true
		for _, l := range h.Lines {
			if w := font.Width(l); w > hintMaxLine {
				t.Errorf("%s: %q is %d px wide, the box takes %d", h.ID, l, w, hintMaxLine)
			}
		}
		if h.after != "" && !seen[h.after] {
			t.Errorf("%s comes after %q, which is not listed before it", h.ID, h.after)
		}
	}
}

// Day 1's steps, one after the other.
func TestCecilyStepsInOrder(t *testing.T) {
	tut := newTutorial(true)
	f := tutFacts{day: 1, window: "Old Cecily", reason: reasonPapersOK, filesWaiting: true}
	if h := tut.Current(f, hintReception); h == nil || h.Step != 1 {
		t.Fatalf("at the window: %+v, want step 1", h)
	}
	if h := tut.Current(f, hintDesk); h != nil {
		t.Errorf("step 1 belongs to the reception, the desk showed %s", h.ID)
	}
	f.visited = true // she looked down at the desk: step 1 is done
	tut.Observe(f)
	if h := tut.Current(f, hintDesk); h != nil {
		t.Errorf("with the folder closed the desk showed %s", h.ID)
	}
	f.folderOpen = true
	for step := 2; step <= 6; step++ {
		h := tut.Current(f, hintDesk)
		if h == nil || h.Step != step {
			t.Fatalf("want step %d, got %+v", step, h)
		}
		if ticks := tut.Ticks(); ticks != max(step-3, 0) {
			t.Errorf("at step %d the sheet has %d ticks, want %d", step, ticks, max(step-3, 0))
		}
		tut.Dismiss(h.ID)
	}
	if tut.Ticks() != 4 {
		t.Errorf("ticks after step 6 = %d, want 4", tut.Ticks())
	}
	if h := tut.Current(f, hintDesk); h == nil || h.Step != 7 {
		t.Fatalf("want step 7, got %+v", h)
	}
	f.holding, f.held = true, StampAdmit // lifting ADMIT is the action step 7 asks for
	tut.Observe(f)
	if h := tut.Current(f, hintDesk); h == nil || h.Step != 8 {
		t.Fatalf("want step 8, got %+v", h)
	}
	f.served = 1 // pressing it ends the steps
	tut.Observe(f)
	f.window, f.folderOpen = "", false
	if h := tut.Current(f, hintDesk); h != nil {
		t.Errorf("after the stamp: %s", h.ID)
	}
}

// Stamping Cecily at once ends the whole chain in one tick, ticks and all.
func TestStampingEndsTheSteps(t *testing.T) {
	tut := newTutorial(true)
	tut.Observe(tutFacts{day: 1, window: "Old Cecily", served: 1})
	for _, h := range hintList {
		if h.Step >= 1 && h.Step <= 8 && !tut.seen[h.ID] {
			t.Errorf("%s still waits after Cecily was stamped", h.ID)
		}
	}
	if tut.Ticks() != 4 {
		t.Errorf("ticks = %d, want 4", tut.Ticks())
	}
}

// Esc and a sleeping guest hide every hint.
func TestHintsHide(t *testing.T) {
	f := tutFacts{day: 1, window: "Old Cecily", filesWaiting: true}
	tut := newTutorial(true)
	f.dozing = true
	if h := tut.Current(f, hintReception); h != nil {
		t.Errorf("a dozing guest: %s showed", h.ID)
	}
	f.dozing = false
	if tut.Current(f, hintReception) == nil {
		t.Fatal("the hint should show")
	}
	tut.Skip()
	if h := tut.Current(f, hintReception); h != nil {
		t.Errorf("after Esc: %s showed", h.ID)
	}
}

// Matty gets one line, Wren two, each in their own moment.
func TestMattyAndWren(t *testing.T) {
	tut := newTutorial(true)
	f := tutFacts{day: 1, served: 1, window: "Matty", reason: reasonPapersOK, filesWaiting: true, folderOpen: true}
	h := tut.Current(f, hintDesk)
	if h == nil || h.ID != "d1.matty" || h.Step != 0 {
		t.Fatalf("Matty: %+v", h)
	}
	tut.Dismiss(h.ID)
	if h := tut.Current(f, hintDesk); h != nil {
		t.Errorf("Matty has free play, but %s showed", h.ID)
	}
	f = tutFacts{day: 1, served: 2, window: "Wren Calloway", reason: reasonExpired, filesWaiting: true}
	if h := tut.Current(f, hintDesk); h != nil {
		t.Errorf("Wren's folder is closed, but %s showed", h.ID)
	}
	f.folderOpen = true
	if h := tut.Current(f, hintDesk); h == nil || h.Step != 9 {
		t.Fatalf("Wren: %+v, want step 9", h)
	}
	f.holding, f.held = true, StampRefer // reaching for REFER skips ahead
	tut.Dismiss("d1.wren1")
	tut.Observe(f)
	if h := tut.Current(f, hintDesk); h != nil {
		t.Errorf("after REFER was lifted: %s", h.ID)
	}
}

// A one-shot hint is over once it has been shown and its moment has passed;
// a hint that was never shown comes back.
func TestOneShotHints(t *testing.T) {
	tut := newTutorial(true)
	f := tutFacts{day: 2, window: "Marcus Trell", reason: reasonName, filesWaiting: true, folderOpen: true}
	h := tut.Current(f, hintDesk)
	if h == nil || h.ID != "d2.name" {
		t.Fatalf("Marcus: %+v", h)
	}
	f.window, f.folderOpen = "Tobias Reef", false // he left before the hint was drawn
	tut.Observe(f)
	f.window, f.reason, f.folderOpen = "Marcus Trell", reasonName, true
	if tut.Current(f, hintDesk) == nil {
		t.Error("a hint that was never drawn should still wait")
	}
	tut.markShown(h)
	f.folderOpen = false
	tut.Observe(f)
	f.folderOpen = true
	if h := tut.Current(f, hintDesk); h != nil {
		t.Errorf("a shown hint came back: %s", h.ID)
	}
}

// The new rule of each day, once.
func TestNewRulesOfDays2And3(t *testing.T) {
	cases := []struct {
		name  string
		f     tutFacts
		scene hintScene
		want  string
	}{
		{"Ozzie", tutFacts{day: 2, window: "Ozzie Pike", reason: reasonNoBarrow, folderOpen: true}, hintDesk, "d2.nostamp"},
		{"Dot", tutFacts{day: 2, window: "Dot Harlan", reason: reasonPapersOK, wardC: true, folderOpen: true}, hintDesk, "d2.wardc"},
		{"Cecily returns", tutFacts{day: 3, window: "Old Cecily", reason: reasonWristband, returning: true, folderOpen: true}, hintDesk, "d3.wristband"},
		{"day 3 sleeper", tutFacts{day: 3, sleeper: true}, hintReception, "d3.sleeper"},
		{"day 1 sleeper", tutFacts{day: 1, served: 1, sleeper: true}, hintReception, "d1.sleeper"},
		{"a citation", tutFacts{day: 2, slip: true}, hintDesk, "citation"},
		{"day 2 has no sleeper hint", tutFacts{day: 2, sleeper: true}, hintReception, ""},
		{"Tobias", tutFacts{day: 2, window: "Tobias Reef", reason: reasonPapersOK, folderOpen: true}, hintDesk, ""},
	}
	for _, c := range cases {
		tut := newTutorial(true)
		got := ""
		if h := tut.Current(c.f, c.scene); h != nil {
			got = h.ID
		}
		if got != c.want {
			t.Errorf("%s: hint %q, want %q", c.name, got, c.want)
		}
	}
}

// The sleeper hint ends when staff are sent (day 1).
func TestSleeperHintEndsWhenAttended(t *testing.T) {
	tut := newTutorial(true)
	f := tutFacts{day: 1, served: 1, sleeper: true}
	tut.markShown(tut.Current(f, hintReception))
	f.attended, f.sleeper = true, false
	tut.Observe(f)
	f.sleeper = true // a second sleeper: the lesson was given already
	if h := tut.Current(f, hintReception); h != nil {
		t.Errorf("second sleeper: %s", h.ID)
	}
}

// The guided steps hold the shift's clock; nothing else does.
func TestGuidedStepsHoldTheClock(t *testing.T) {
	sh := newTestShift(t, 1)
	sh.EnableTutorial(true)
	if err := deskSetup(true, PageOrientation)(sh); err != nil {
		t.Fatal(err)
	}
	if sh.Paused() {
		t.Fatal("paused before the player opened her files")
	}
	sh.tut.Dismiss(hintFiles)
	if !sh.Paused() {
		t.Fatal("not paused during the steps")
	}
	before := sh.Seconds()
	runTicks(sh, 600)
	if sh.Seconds() != before {
		t.Errorf("the shift ran %.1f s during the steps", sh.Seconds()-before)
	}
	sh.SkipTutorial()
	runTicks(sh, 60)
	if sh.Seconds() <= before {
		t.Error("the clock should run again after Esc")
	}
}

// Without hints the first folder of day 1 has the sheet last, like all the
// others; with them it has it first.
func TestOrientationSheetPlace(t *testing.T) {
	off := newTestShift(t, 1)
	if got := pageOrder(off.Guests[0]); got[0] != PageReferral || got[len(got)-1] != PageOrientation {
		t.Errorf("hints off: %v", got)
	}
	on := newTestShift(t, 1)
	on.EnableTutorial(true)
	if got := pageOrder(on.Guests[0]); got[0] != PageOrientation {
		t.Errorf("hints on: %v", got)
	}
	if got := pageOrder(on.Guests[1]); got[len(got)-1] != PageOrientation {
		t.Errorf("Matty: %v, the sheet is the last page of every other folder", got)
	}
}

// Every guided step can be reached with -tutorial-step.
func TestJumpToEveryStep(t *testing.T) {
	for step := 1; step <= 10; step++ {
		sh := newTestShift(t, 1)
		sh.EnableTutorial(true)
		if err := sh.jumpToTutorialStep(step); err != nil {
			t.Fatalf("step %d: %v", step, err)
		}
		scene := hintDesk
		if step == 1 {
			scene = hintReception
		} else if err := sh.GoToPage(pageOrder(sh.desk.guest)[0]); err != nil {
			t.Fatalf("step %d: opening the folder: %v", step, err)
		}
		if step == 3 {
			_ = sh.GoToPage(PageReferral)
		}
		h := sh.CurrentHint(scene)
		if h == nil || h.Step != step {
			t.Errorf("step %d: shows %+v", step, h)
		}
	}
	if err := newTestShift(t, 2).jumpToTutorialStep(1); err == nil {
		t.Error("the steps belong to day 1")
	}
}

// Pressing the stamp on Cecily goes through the whole lesson with real clicks'
// worth of calls: the ticks come as the steps are read.
func TestClickingThroughTheSteps(t *testing.T) {
	sh := newTestShift(t, 1)
	sh.EnableTutorial(true)
	if err := deskSetup(true, PageOrientation)(sh); err != nil {
		t.Fatal(err)
	}
	sh.tut.Dismiss(hintFiles)
	for step := 2; step <= 6; step++ {
		h := sh.CurrentHint(hintDesk)
		if h == nil || h.Step != step {
			t.Fatalf("want step %d, got %+v", step, h)
		}
		sh.AdvanceHint(hintDesk)
		if step == 2 && sh.Desk().Page() != PageReferral {
			t.Errorf("after the sheet the folder shows %v, want the referral", sh.Desk().Page())
		}
	}
	if sh.TutorialTicks() != 4 {
		t.Errorf("ticks = %d, want 4", sh.TutorialTicks())
	}
	sh.LiftStamp(StampAdmit)
	runTicks(sh, 1)
	if h := sh.CurrentHint(hintDesk); h == nil || h.Step != 8 {
		t.Fatalf("holding ADMIT: %+v, want step 8", h)
	}
	if err := sh.ApplyStamp(); err != nil {
		t.Fatal(err)
	}
	runTicks(sh, 1)
	if h := sh.CurrentHint(hintDesk); h != nil && h.Step > 0 && h.Step < 9 {
		t.Errorf("after the stamp step %d still shows", h.Step)
	}
}

// The HUD clock stands still with the shift while the steps hold it.
func TestGuidedStepsHoldTheHUDClock(t *testing.T) {
	env := testEnv(t)
	env.TutorialHints = true
	if err := env.StartShift(); err != nil {
		t.Fatal(err)
	}
	defer env.StopShift()
	sh := env.Shift()
	if err := deskSetup(true, PageOrientation)(sh); err != nil {
		t.Fatal(err)
	}
	sh.tut.Dismiss(hintFiles)
	start := env.Clock.String()
	for range 600 {
		env.tickClock()
		sh.Update()
	}
	if env.Clock.String() != start {
		t.Errorf("HUD clock went %s -> %s while paused", start, env.Clock)
	}
	sh.SkipTutorial()
	for range 600 {
		env.tickClock()
		sh.Update()
	}
	if env.Clock.String() == start {
		t.Error("HUD clock should run again after Esc")
	}
}
