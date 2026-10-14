package game

import (
	"errors"
	"strings"
	"testing"
)

// folderOnDesk runs a shift until the window guest's folder has slid in.
func folderOnDesk(t *testing.T, day int) *Shift {
	t.Helper()
	sh := newTestShift(t, day)
	if err := deskSetup(false, PageReferral)(sh); err != nil {
		t.Fatal(err)
	}
	return sh
}

func TestFolderArrivesWithTheGuestAtTheWindow(t *testing.T) {
	sh := newTestShift(t, 2)
	if sh.FilesWaiting() {
		t.Fatal("files on the desk before anyone came")
	}
	runTicks(sh, 5*60)
	if !sh.FilesWaiting() || sh.Desk().FolderGuest() != sh.Guests[0] {
		t.Fatal("the first guest's folder should be on the desk")
	}
	pose, ok := sh.FolderPose()
	if !ok || pose.open != 0 {
		t.Errorf("pose %+v, want a closed folder", pose)
	}
	if err := sh.OpenFolder(); !errors.Is(err, nil) && !errors.Is(err, ErrFolderBusy) {
		t.Errorf("OpenFolder = %v", err)
	}
}

func TestOpenAndTurnPages(t *testing.T) {
	sh := folderOnDesk(t, 3) // day 3 starts with Old Cecily, a returning patient
	if err := sh.TurnPage(); !errors.Is(err, ErrFolderClosed) {
		t.Errorf("turning a closed folder: %v, want ErrFolderClosed", err)
	}
	if err := sh.OpenFolder(); err != nil {
		t.Fatal(err)
	}
	runTicks(sh, openTicks)
	var seen []Page
	for range 3 {
		seen = append(seen, sh.Desk().Page())
		if err := sh.TurnPage(); err != nil {
			t.Fatal(err)
		}
		runTicks(sh, flipTestTicks)
	}
	want := []Page{PageReferral, PageSleepLog, PageWristband}
	for i := range want {
		if seen[i] != want[i] {
			t.Fatalf("pages seen %v, want %v", seen, want)
		}
	}
	if sh.Desk().Page() != PageReferral {
		t.Error("after the last page the stack turns back to the referral")
	}
	// A visitor without a wristband has two pages only.
	two := folderOnDesk(t, 2)
	if err := two.GoToPage(PageWristband); !errors.Is(err, ErrNoSuchPage) {
		t.Errorf("wristband page for a new patient: %v, want ErrNoSuchPage", err)
	}
}

func TestApplyStampRules(t *testing.T) {
	sh := folderOnDesk(t, 2)
	if err := sh.ApplyStamp(); !errors.Is(err, ErrNoStampHeld) {
		t.Errorf("empty hand: %v", err)
	}
	sh.LiftStamp(StampAdmit)
	if err := sh.ApplyStamp(); !errors.Is(err, ErrFolderClosed) {
		t.Errorf("closed folder: %v, want ErrFolderClosed", err)
	}
	if err := sh.GoToPage(PageSleepLog); err != nil {
		t.Fatal(err)
	}
	if err := sh.ApplyStamp(); !errors.Is(err, ErrWrongPage) {
		t.Errorf("sleep log page: %v, want ErrWrongPage", err)
	}
	if sh.Score() != 0 || len(sh.Stamps()) != 0 {
		t.Error("a refused stamp must not count")
	}
}

// Stamping on the desk is the same stamping as before: ADMIT admits, a right
// call scores, then the folder closes and slides back to the window.
func TestApplyStampAdmitsAndFolderLeaves(t *testing.T) {
	sh := folderOnDesk(t, 2)
	if err := sh.GoToPage(PageReferral); err != nil {
		t.Fatal(err)
	}
	runTicks(sh, openTicks)
	sh.LiftStamp(StampAdmit)
	if err := sh.ApplyStamp(); err != nil {
		t.Fatal(err)
	}
	if sh.Score() != pointsRightStamp || sh.Served() != 1 {
		t.Errorf("score %d served %d", sh.Score(), sh.Served())
	}
	if _, ok := sh.Desk().Holding(); ok {
		t.Error("the stamp should be out of the hand after it is used")
	}
	pose, ok := sh.FolderPose()
	if !ok || !pose.ink || pose.open != 1 || pose.stamp != StampAdmit {
		t.Errorf("right after the stamp: %+v, want the open folder with the ink mark", pose)
	}
	if _, _, slip := sh.Slip(); slip {
		t.Error("a right stamp must not bring a citation")
	}
	runTicks(sh, inkTicks+closeTicks/2)
	if pose, _ := sh.FolderPose(); pose.open <= 0 || pose.open >= 1 {
		t.Errorf("the folder should be closing, open = %v", pose.open)
	}
	runTicks(sh, closeTicks+slideOutTicks)
	if _, ok := sh.FolderPose(); ok {
		t.Error("the folder should be gone after it slid back")
	}
}

func TestWrongStampBringsACitation(t *testing.T) {
	sh := folderOnDesk(t, 2)
	_ = sh.GoToPage(PageReferral)
	runTicks(sh, openTicks)
	sh.LiftStamp(StampRefer) // Tobias Reef's papers are fine: ADMIT is right
	if err := sh.ApplyStamp(); err != nil {
		t.Fatal(err)
	}
	stamped, right, ok := sh.Slip()
	if !ok || stamped != StampRefer || right != StampAdmit {
		t.Errorf("slip = %v %v %v", stamped, right, ok)
	}
	if sh.Score() != 0 {
		t.Errorf("score = %d after a wrong stamp", sh.Score())
	}
	l1, l2 := citationLines(stamped, right)
	if l1 != "STAMPED: REFER" || l2 != "RIGHT CALL: ADMIT" {
		t.Errorf("lines %q, %q", l1, l2)
	}
	runTicks(sh, citationTicks+1)
	if _, _, ok := sh.Slip(); ok {
		t.Error("the slip should go away")
	}
}

// The day 1 tutorial pulse stops once the player has been to the desk, and
// only day 1 pulses.
func TestFilesTabPulse(t *testing.T) {
	sh := newTestShift(t, 1)
	if sh.Pulsing() {
		t.Error("pulsing with no files on the desk")
	}
	runTicks(sh, 4*60)
	if !sh.Pulsing() {
		t.Fatal("day 1 should pulse while the first folder waits")
	}
	sh.Desk().MarkVisited()
	if sh.Pulsing() || !sh.FilesWaiting() {
		t.Errorf("after opening the desk: pulsing %v waiting %v, want a static folder mark", sh.Pulsing(), sh.FilesWaiting())
	}
	day2 := newTestShift(t, 2)
	runTicks(day2, 4*60)
	if day2.Pulsing() {
		t.Error("only the day 1 tutorial pulses")
	}
}

func TestDozingGuestCannotBeStampedOnTheDesk(t *testing.T) {
	sh := folderOnDesk(t, 1)
	runTicks(sh, 4*60) // Old Cecily nods off mid-joke
	_ = sh.GoToPage(PageReferral)
	sh.LiftStamp(StampAdmit)
	if err := sh.ApplyStamp(); !errors.Is(err, ErrDozing) {
		t.Errorf("err = %v, want ErrDozing", err)
	}
	if !strings.Contains(sh.Message(), "CLICK HER") {
		t.Errorf("message %q should still tell the player to click her", sh.Message())
	}
}

func TestLookDownShapes(t *testing.T) {
	full := [4]point{{0, 0}, {ScreenW, 0}, {0, ScreenH}, {ScreenW, ScreenH}}
	_, _, _, _, dtl, dtr, dbl, dbr := lookDownShapes(1)
	if [4]point{dtl, dtr, dbl, dbr} != full {
		t.Errorf("at the end the desk is %v, want the full screen", [4]point{dtl, dtr, dbl, dbr})
	}
	rtl, rtr, rbl, rbr, dtl, _, _, _ := lookDownShapes(0)
	if [4]point{rtl, rtr, rbl, rbr} != full {
		t.Errorf("at the start the reception is %v, want the full screen", [4]point{rtl, rtr, rbl, rbr})
	}
	if dtl.y < ScreenH {
		t.Errorf("the desk starts below the screen, top at %v", dtl.y)
	}
	// The settle: the desk passes its place by a pixel or two, then rests on it.
	over := false
	for i := 0; i <= 100; i++ {
		_, _, _, _, tl, _, _, _ := lookDownShapes(float64(i) / 100)
		if tl.y < 0 && tl.y > -2.5 {
			over = true
		}
		if tl.y < -2.5 {
			t.Fatalf("overshoot of %v px is more than 2", -tl.y)
		}
	}
	if !over {
		t.Error("the desk should settle with a 1-2 px overshoot")
	}
}

func TestParseDeskScript(t *testing.T) {
	cmds, err := ParseScript(strings.NewReader("desk\nfolder\nturn\npage sleeplog\nlift admit\napply\nputback\nreception\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(cmds) != 8 || cmds[3].N != int(PageSleepLog) || cmds[4].Stamp != StampAdmit {
		t.Errorf("cmds = %+v", cmds)
	}
	for _, bad := range []string{"page last", "lift nothing", "folder now", "page"} {
		if _, err := ParseScript(strings.NewReader(bad)); err == nil {
			t.Errorf("%q should fail", bad)
		}
	}
}

// The desk scene's clicks: pick a stamp up, press it on the FOR DESK USE
// box. Nothing is drawn; the scene only reads clicks.
func TestDeskSceneClicks(t *testing.T) {
	env := testEnv(t)
	env.Day = 2
	s := &deskScene{setup: deskSetup(true, PageReferral)}
	if err := s.Enter(env); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { env.StopShift() })
	click := func(x, y int) {
		env.Input.InjectClick(x, y)
		if err := s.Update(); err != nil {
			t.Fatal(err)
		}
		env.Input.endFrame()
	}
	sh := env.Shift()
	click(372, 120) // the ADMIT stamp
	if st, ok := sh.Desk().Holding(); !ok || st != StampAdmit {
		t.Fatalf("holding %v %v after clicking the stamp", st, ok)
	}
	click(100, 100) // anywhere else puts it back
	if _, ok := sh.Desk().Holding(); ok {
		t.Fatal("a click elsewhere should put the stamp back")
	}
	click(372, 120)
	click(210, 200) // the FOR DESK USE box
	if sh.Served() != 1 || sh.Score() != pointsRightStamp {
		t.Errorf("served %d score %d, want 1 and %d", sh.Served(), sh.Score(), pointsRightStamp)
	}
	click(100, 260) // BACK
	if s.dir >= 0 {
		t.Error("BACK should start the tilt back up")
	}
}

func TestDragProgress(t *testing.T) {
	cases := []struct {
		startX, x int
		want      float64
	}{
		{318, 318, 0},
		{318, 400, 0}, // dragging the wrong way does nothing
		{318, 318 - flipTravel/2, 0.5},
		{318, 318 - flipTravel, 1},
		{318, 0, 1}, // past the spine: the page is fully turned, no further
	}
	for _, c := range cases {
		if got := dragProgress(c.startX, c.x); got != c.want {
			t.Errorf("dragProgress(%d, %d) = %v, want %v", c.startX, c.x, got, c.want)
		}
	}
}

// Letting go past the middle finishes the flip and the page goes to the back
// of the stack; before the middle it springs back to where it was.
func TestReleaseFinishesOrSpringsBack(t *testing.T) {
	for _, c := range []struct {
		drag     int
		wantPage Page
	}{{flipTravel/2 + 5, PageSleepLog}, {flipTravel/2 - 5, PageReferral}} {
		sh := folderOnDesk(t, 2)
		_ = sh.GoToPage(PageReferral)
		runTicks(sh, openTicks)
		if err := sh.BeginFlipDrag(318, 213); err != nil {
			t.Fatal(err)
		}
		sh.DragFlipTo(318-c.drag, 213)
		runTicks(sh, 5)
		if sh.Desk().Page() != PageReferral {
			t.Fatal("the page must not change while it is held")
		}
		sh.EndFlipDrag()
		runTicks(sh, flipTestTicks)
		if got := sh.Desk().Page(); got != c.wantPage {
			t.Errorf("released after %d px: page %v, want %v", c.drag, got, c.wantPage)
		}
		if _, turning := sh.Turning(); turning {
			t.Error("the flip should be over")
		}
	}
}

func TestDragScript(t *testing.T) {
	cmds, err := ParseScript(strings.NewReader("drag 318 213 200 213\n"))
	if err != nil {
		t.Fatal(err)
	}
	if cmds[0].Op != "down" || cmds[0].X != 318 || cmds[len(cmds)-1].Op != "up" {
		t.Errorf("cmds = %+v", cmds)
	}
	if _, err := ParseScript(strings.NewReader("drag 1 2 3")); err == nil {
		t.Error("drag with three numbers should fail")
	}
}

// The scene: press on the corner, drag left past the middle, let go.
func TestDeskSceneDragFlipsThePage(t *testing.T) {
	env := testEnv(t)
	env.Day = 2
	s := &deskScene{setup: deskSetup(true, PageReferral)}
	if err := s.Enter(env); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { env.StopShift() })
	frame := func() {
		if err := s.Update(); err != nil {
			t.Fatal(err)
		}
		env.Input.endFrame()
	}
	sh := env.Shift()
	runTicks(sh, openTicks)
	env.Input.InjectDown(318, 213)
	frame()
	if !sh.Dragging() {
		t.Fatal("pressing the corner should take the page")
	}
	env.Input.InjectMouse(318-100, 213)
	frame()
	env.Input.InjectUp()
	frame()
	if sh.Dragging() {
		t.Error("the button is up: the page should be let go")
	}
	runTicks(sh, flipTestTicks)
	if sh.Desk().Page() != PageSleepLog {
		t.Errorf("page = %v, want the sleep log", sh.Desk().Page())
	}
}

// flipTestTicks is long enough for any released flip to finish: the curl
// sweeps (about half a second at worst) and the page tucks to the back.
const flipTestTicks = 60

// A released flip takes a fraction of a second, not much more, in both
// directions.
func TestFlipAnimationIsQuick(t *testing.T) {
	for _, target := range []float64{0, 1} {
		f := pageFlip{active: true, progress: 0.5, target: target}
		ticks := 0
		for f.active && !f.done && ticks < 100 {
			f.step()
			ticks++
		}
		if target == 1 && !f.done || target == 0 && f.active {
			t.Fatalf("target %v never finished", target)
		}
		if ticks > 24 { // 0.4 s at most from the middle
			t.Errorf("target %v took %d ticks", target, ticks)
		}
	}
}
