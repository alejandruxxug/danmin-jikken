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
	sh := folderOnDesk(t, 3) // day 3 starts with Old Cecily: returning, and Hana's report is clipped to her folder
	if err := sh.TurnPage(); !errors.Is(err, ErrFolderClosed) {
		t.Errorf("turning a closed folder: %v, want ErrFolderClosed", err)
	}
	if err := sh.OpenFolder(); err != nil {
		t.Fatal(err)
	}
	runTicks(sh, openTicks)
	var seen []Page
	for range 5 {
		seen = append(seen, sh.Desk().Page())
		if err := sh.TurnPage(); err != nil {
			t.Fatal(err)
		}
		runTicks(sh, flipTestTicks)
	}
	want := []Page{PageReferral, PageSleepLog, PageWristband, PageReport, PageOrientation}
	for i := range want {
		if seen[i] != want[i] {
			t.Fatalf("pages seen %v, want %v", seen, want)
		}
	}
	if sh.Desk().Page() != PageReferral {
		t.Error("after the last page the stack turns back to the referral")
	}
	// A visitor without a wristband has the referral, the sleep log and the
	// orientation sheet only.
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
	if pose, ok := sh.FolderPose(); !ok || pose.ink || pose.open != 1 {
		t.Errorf("during the hit-stop: %+v, want the open folder with no ink yet", pose)
	}
	runTicks(sh, hitStopTicks)
	pose, ok := sh.FolderPose()
	if !ok || !pose.ink || pose.open != 1 || pose.stamp != StampAdmit {
		t.Errorf("after the hit-stop: %+v, want the open folder with the ink mark", pose)
	}
	if _, _, _, slip := sh.Slip(); slip {
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
	stamped, right, reason, ok := sh.Slip()
	if !ok || stamped != StampRefer || right != StampAdmit || reason != "PAPERS OK" {
		t.Errorf("slip = %v %v %q %v", stamped, right, reason, ok)
	}
	if sh.Score() != 0 {
		t.Errorf("score = %d after a wrong stamp", sh.Score())
	}
	l1, l2 := citationLines(reason, right)
	if l1 != "PAPERS OK" || l2 != "RIGHT: ADMIT" {
		t.Errorf("lines %q, %q", l1, l2)
	}
	runTicks(sh, citationTicks+1)
	if _, _, _, ok := sh.Slip(); ok {
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
	click(250, 60)               // anywhere on the open page: the ink snaps into the box
	for range landingTotal + 1 { // the landing plays; the stamp logic runs on contact
		if err := s.Update(); err != nil {
			t.Fatal(err)
		}
		env.Input.endFrame()
	}
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

// The hit areas are big: the whole cubby lifts a stamp, the whole open page
// takes it, and the right-edge strip grabs the page. Just inside works, just
// outside does not.
func TestDeskHitAreas(t *testing.T) {
	for _, d := range deskStamps {
		if !d.cubby.has(d.at.x, d.at.y) || !d.cubby.has(d.at.x+d.at.w-1, d.at.y+d.at.h-1) {
			t.Errorf("the cubby of %v must hold its stamp", d.stamp)
		}
	}
	admit := deskStamps[0].cubby
	for _, c := range []struct {
		area rect
		x, y int
		in   bool
	}{
		{admit, admit.x, admit.y, true}, {admit, admit.x - 1, admit.y, false},
		{admit, admit.x + admit.w - 1, admit.y + admit.h - 1, true}, {admit, admit.x + admit.w, admit.y, false},
		{grabRect, grabRect.x, grabRect.y, true}, {grabRect, grabRect.x - 1, grabRect.y + 50, false},
		{grabRect, grabRect.x + grabRect.w - 1, grabRect.y + grabRect.h - 1, true}, {grabRect, grabRect.x, grabRect.y + grabRect.h, false},
	} {
		if got := c.area.has(c.x, c.y); got != c.in {
			t.Errorf("%v at (%d, %d): inside = %v, want %v", c.area, c.x, c.y, got, c.in)
		}
	}
	if grabRect.x+grabRect.w != pageRect.x+pageRect.w || grabRect.h != pageRect.h {
		t.Error("the grab strip is the right edge of the page, full height")
	}
}

// Press anywhere on the strip, drag: the page turns; just outside it does not.
func TestDeskSceneGrabAnywhereOnTheEdge(t *testing.T) {
	env := testEnv(t)
	env.Day = 2
	s := &deskScene{setup: deskSetup(true, PageReferral)}
	if err := s.Enter(env); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { env.StopShift() })
	sh := env.Shift()
	runTicks(sh, openTicks)
	press := func(x, y int) {
		env.Input.InjectClick(x, y)
		if err := s.Update(); err != nil {
			t.Fatal(err)
		}
		env.Input.endFrame()
	}
	press(grabRect.x-1, 100)
	if sh.Dragging() {
		t.Error("a press just left of the strip must not take the page")
	}
	press(grabRect.x, 100) // mid-height on the edge, not the corner
	if !sh.Dragging() {
		t.Error("a press just inside the strip should take the page")
	}
}

// The landing's motion, frame by frame: lift and stretch, slam and squash,
// hold, spring back with an overshoot, done. Pure functions of the frame.
func TestLandingPose(t *testing.T) {
	if p := landingPose(0); p.lift != windupLift || p.sy <= 1 || p.sx >= 1 {
		t.Errorf("wind-up %+v: want lifted 3 px, taller and narrower", p)
	}
	prev := landingPose(windupTicks).lift
	for f := windupTicks + 1; f < impactFrame; f++ {
		p := landingPose(f)
		if p.lift >= prev {
			t.Errorf("frame %d: lift %v did not fall (was %v)", f, p.lift, prev)
		}
		prev = p.lift
	}
	contact := landingPose(impactFrame - 1)
	if contact.lift != 0 || contact.sx <= 1 || contact.sy >= 1 {
		t.Errorf("contact %+v: want on the page, wider and shorter", contact)
	}
	for f := impactFrame; f < impactFrame+hitStopTicks; f++ {
		if p := landingPose(f); p != (stampPose{sx: 1.15, sy: 0.85}) {
			t.Errorf("hit-stop frame %d moved: %+v", f, p)
		}
	}
	peak := 0.0
	for f := impactFrame + hitStopTicks; f < landingTotal; f++ {
		peak = max(peak, landingPose(f).lift)
	}
	if peak <= springRise {
		t.Errorf("the spring-back peaks at %v, want an overshoot past %v", peak, springRise)
	}
	if !landingPose(landingTotal).done || landingPose(landingTotal-1).done {
		t.Error("the landing should be done exactly at landingTotal")
	}
}

func TestShakeDecaysAndIsFixed(t *testing.T) {
	maxFirst, last := 0, 0
	for f := 0; f < shakeTicks+3; f++ {
		dx, dy := shakeOffset(f)
		if a := abs(dx); f == 0 {
			maxFirst = a
		} else if a > maxFirst {
			t.Errorf("frame %d shakes %d px, more than the first (%d)", f, a, maxFirst)
		}
		if abs(dy) > 1 {
			t.Errorf("frame %d: vertical shake %d", f, dy)
		}
		last = abs(dx) + abs(dy)
	}
	if maxFirst != 2 || last != 0 {
		t.Errorf("first frame %d px (want 2), after the shake %d (want 0)", maxFirst, last)
	}
	a1, a2 := fmtShake(), fmtShake()
	if a1 != a2 {
		t.Error("the shake must be the same every run")
	}
}

func fmtShake() [shakeTicks][2]int {
	var out [shakeTicks][2]int
	for f := range out {
		out[f][0], out[f][1] = shakeOffset(f)
	}
	return out
}

func TestInkPopFolderDipAndPopup(t *testing.T) {
	if inkScale(-1) != 0 || inkScale(0) != popScale || inkScale(popTicks) != 1 {
		t.Errorf("ink scale: %v %v %v", inkScale(-1), inkScale(0), inkScale(popTicks))
	}
	if mid := inkScale(popTicks/2 + 1); mid <= 1 || mid >= popScale {
		t.Errorf("the settle has a middle step, got %v", mid)
	}
	if folderDip(0) != 1 || folderDip(dipTicks) != 0 || folderDip(-1) != 0 {
		t.Error("the folder dips 1 px for dipTicks frames")
	}
	r0, a0 := popup(0)
	r1, a1 := popup(popupTicks - 5)
	if r1 <= r0 || a1 >= a0 || a0 != 1 {
		t.Errorf("the score should rise and fade: %v %v then %v %v", r0, a0, r1, a1)
	}
	if _, a := popup(popupTicks); a != 0 {
		t.Error("the score is gone after popupTicks")
	}
	if s, a := speck(speckTicks); s != 0 || a != 0 {
		t.Error("the specks are gone after speckTicks")
	}
}

// The score popup shows what the stamp scored, after the impact.
func TestScorePopupAfterTheStamp(t *testing.T) {
	sh := folderOnDesk(t, 2)
	_ = sh.GoToPage(PageReferral)
	runTicks(sh, openTicks)
	sh.LiftStamp(StampAdmit)
	if err := sh.ApplyStamp(); err != nil {
		t.Fatal(err)
	}
	runTicks(sh, hitStopTicks) // the score floats up when the ink appears
	text, _, alpha, ok := sh.ScorePopup()
	if !ok || text != "+100" || alpha != 1 {
		t.Errorf("popup %q %v %v", text, alpha, ok)
	}
	if since, ok := sh.SinceImpact(); !ok || since != hitStopTicks {
		t.Errorf("since impact %d %v", since, ok)
	}
}

// Every citation line fits the lines printed on the slip.
func TestCitationLinesFit(t *testing.T) {
	font, err := LoadFont()
	if err != nil {
		t.Fatal(err)
	}
	reasons := []string{reasonExpired, reasonName, reasonNoBarrow, reasonDiagnosis, reasonPapersOK, reasonWristband, reasonNoWristband}
	for _, reason := range reasons {
		for _, right := range []Stamp{StampAdmit, StampRefer, StampDischarge} {
			l1, l2 := citationLines(reason, right)
			for _, l := range []string{l1, l2} {
				if w := font.Width(l); w > citationLineWidth {
					t.Errorf("%q is %d px wide, the slip line is %d", l, w, citationLineWidth)
				}
			}
		}
	}
}

// Hana's report is the last page of the first folder of day 3 and of no other.
func TestReportPageOnlyInTheFirstFolderOfDay3(t *testing.T) {
	d3, err := RulesForDay(3)
	if err != nil {
		t.Fatal(err)
	}
	first := &Guest{Visitor: d3.Visitors[0]}
	order := pageOrder(first)
	want := []Page{PageReferral, PageSleepLog, PageWristband, PageReport, PageOrientation}
	if len(order) != len(want) {
		t.Fatalf("pages = %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("pages = %v, want %v", order, want)
		}
	}
	if got := nextPage(first, PageOrientation); got != PageReferral {
		t.Errorf("after the orientation sheet the stack cycles to %v", got)
	}
	if got := nextPage(first, PageWristband); got != PageReport {
		t.Errorf("after the wristband comes %v", got)
	}
	for _, v := range d3.Visitors[1:] {
		if hasPage(&Guest{Visitor: v}, PageReport) {
			t.Errorf("%s has the report but only the first folder should", v.Name)
		}
	}
	d1, _ := RulesForDay(1)
	for _, v := range d1.Visitors {
		if v.Report {
			t.Errorf("%s: no report before day 3", v.Name)
		}
	}
}
