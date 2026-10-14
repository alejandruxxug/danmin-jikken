package game

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
)

// The files on the look-down desk (Game Design -> design log: "Patients hand
// over a folder"). This is the logic only: which folder is on the desk, how
// far it is open, which page shows, which stamp is in the hand. It has no
// drawing code; desk_scene.go reads it and draws. Times are in ticks of the
// shift (60 a second), which do not speed up with -speed: they are
// animations, not game time.

// Page is a page of the open folder, in the order they are turned.
type Page int

const (
	PageReferral Page = iota
	PageSleepLog
	PageWristband // returning patients only
	PageReport    // Hana's night report: clipped to the first folder of day 3
)

func (p Page) String() string {
	switch p {
	case PageReferral:
		return "referral"
	case PageSleepLog:
		return "sleep log"
	case PageWristband:
		return "wristband"
	case PageReport:
		return "report"
	}
	return "unknown page"
}

// Ticks each desk move takes.
const (
	slideInTicks  = 28  // the folder slides onto the desk
	openTicks     = 14  // the cover swings open
	turnTicks     = 14  // a page turns
	inkTicks      = 24  // the ink mark shows before the folder closes
	closeTicks    = 14  // the cover swings shut
	slideOutTicks = 24  // the folder slides back to the window
	citationTicks = 300 // how long the citation slip stays on the desk
)

// Errors the desk reports for clicks that do nothing.
var (
	ErrNoFolder     = errors.New("there is no folder on the desk")
	ErrFolderClosed = errors.New("open the folder first")
	ErrFolderBusy   = errors.New("the folder is still moving")
	ErrWrongPage    = errors.New("stamp the referral page")
	ErrNoStampHeld  = errors.New("no stamp in hand")
	ErrNoSuchPage   = errors.New("this folder has no such page")
)

// sent is a folder on its way back to the window after a stamp.
type sent struct {
	guest *Guest
	stamp Stamp
	page  Page // the page that showed; the ink mark is only on the referral
	open  bool // the folder was open
	at    int  // tick the ink appears: the impact plus the hit-stop
	// impact is the tick the stamp hit the page (the stamp logic ran then),
	// and points what it scored; both are for the landing effects.
	impact int
	points int
}

// citation is the slip handed out for a wrong stamp (flavour, no penalty).
type citation struct {
	stamped, right Stamp
	until          int
}

// Desk is the state of the look-down desk.
type Desk struct {
	guest     *Guest // whose folder lies on the desk, nil when there is none
	arrivedAt int    // tick the folder started sliding in
	open      bool
	openedAt  int
	page      Page // the page on top of the stack
	flip      pageFlip
	holding   bool // a stamp is lifted off the desk
	held      Stamp
	visited   bool // the player has opened the desk at least once
	out       *sent
	slip      *citation
}

// Desk is the desk's state, for the scenes to draw.
func (s *Shift) Desk() *Desk { return &s.desk }

// FolderGuest is the guest whose folder lies on the desk, or nil.
func (d *Desk) FolderGuest() *Guest { return d.guest }

// IsOpen says whether the folder is open (or opening).
func (d *Desk) IsOpen() bool { return d.open }

// Page is the page showing.
func (d *Desk) Page() Page { return d.page }

// Holding is the stamp in the player's hand, and whether there is one.
func (d *Desk) Holding() (Stamp, bool) { return d.held, d.holding }

// Visited is true once the desk has been opened (the tutorial pulse stops).
func (d *Desk) Visited() bool { return d.visited }

// MarkVisited records that the player has opened the desk.
func (d *Desk) MarkVisited() { d.visited = true }

// pages is how many pages the guest's folder has.
func pages(g *Guest) int { return len(pageOrder(g)) }

// pageOrder is the order the pages lie in the guest's folder: the referral
// and the sleep log, then the wristband (returning patients) and Hana's
// report (the one folder that has it).
func pageOrder(g *Guest) []Page {
	order := []Page{PageReferral, PageSleepLog}
	if g.Returning {
		order = append(order, PageWristband)
	}
	if g.Report {
		order = append(order, PageReport)
	}
	return order
}

// hasPage says whether the guest's folder has the page.
func hasPage(g *Guest, p Page) bool { return slices.Contains(pageOrder(g), p) }

// FilesWaiting is true while a folder lies on the desk: the FILES tab shows
// it (pulsing in the day 1 tutorial, a folder mark afterwards).
func (s *Shift) FilesWaiting() bool { return s.desk.guest != nil }

// Pulsing is true in the day 1 tutorial while files wait and the player has
// not opened the desk yet.
func (s *Shift) Pulsing() bool {
	return s.Day == 1 && !s.desk.visited && s.FilesWaiting()
}

// updateDesk runs every tick: the next guest's folder arrives when they
// stand at the window and the last folder is gone, and a folder that has
// finished leaving is forgotten.
func (s *Shift) updateDesk() {
	d := &s.desk
	if d.out != nil && s.ticks-d.out.at >= inkTicks+closeTicks+slideOutTicks {
		d.out = nil
	}
	d.flip.step()
	if d.flip.done {
		d.page = nextPage(d.guest, d.page)
		d.flip = pageFlip{}
	}
	if d.slip != nil && s.ticks >= d.slip.until {
		d.slip = nil
	}
	w := s.window()
	if d.guest == nil && d.out == nil && w != nil && s.atWindow(w) {
		*d = Desk{guest: w, arrivedAt: s.ticks, visited: d.visited, slip: d.slip}
	}
}

// LastGuestAtDesk is true while the day's last guest is at the desk and not
// yet stamped (the strobe's flicker starts then).
func (s *Shift) LastGuestAtDesk() bool {
	return s.desk.guest != nil && s.served == len(s.Guests)-1
}

// FolderReady is true once the folder has finished sliding in.
func (s *Shift) FolderReady() bool {
	return s.desk.guest != nil && s.ticks-s.desk.arrivedAt >= slideInTicks
}

// OpenFolder opens the closed folder on the desk, like a book.
func (s *Shift) OpenFolder() error {
	d := &s.desk
	switch {
	case d.guest == nil:
		return ErrNoFolder
	case !s.FolderReady():
		return ErrFolderBusy
	case d.open:
		return nil
	}
	d.open, d.openedAt, d.page = true, s.ticks, PageReferral
	return nil
}

// FolderOpen is how far open the folder is, 0 (closed) to 1 (open).
func (s *Shift) FolderOpen() float64 {
	if !s.desk.open {
		return 0
	}
	return min(float64(s.ticks-s.desk.openedAt)/openTicks, 1)
}

// flipTravel is how far (in pixels) the page corner is dragged to turn the
// whole page: the page's width, so the page's edge follows the cursor.
const flipTravel = pageW

// dragProgress is how far a page has turned when the corner that was
// pressed at startX is now at x: 0 (flat) to 1 (turned right over).
func dragProgress(startX, x int) float64 {
	return min(max(float64(startX-x)/flipTravel, 0), 1)
}

// pageFlip is the top page being turned, by a drag or by TurnPage. The
// grabbed corner is at progress along the way to the spine and dy pixels
// above or below the corner's own height (the cursor's up-and-down), which
// curl.go turns into the fold.
type pageFlip struct {
	active   bool
	dragging bool // the player is holding the corner
	startX   int  // where the corner was pressed
	startY   int
	progress float64 // 0 to 1
	dy       float64 // the cursor's height above the grab point, pixels
	target   float64 // after letting go: 1 finishes the flip, 0 springs back
	tuck     float64 // 0 to 1: the turned page slides to the back of the stack
	done     bool    // the flip finished this tick: the next page is on top
}

// finishAt is the progress at which letting go finishes the flip.
const finishAt = 0.5

// tuckTicks is how long the turned page takes to slide to the back.
const tuckTicks = 6

// release lets go of the corner: past the middle the page finishes its
// turn, before the middle it springs back.
func (f *pageFlip) release() {
	f.dragging = false
	f.target = 0
	if f.progress >= finishAt {
		f.target = 1
	}
}

// step moves a released flip toward its target: quickly at first, then
// slowing down (a fifth of what is left each tick, but never slower than 1/40),
// and the corner's height settles back too. Once a flip has turned all the
// way the page tucks to the back; once it is flat again the flip is over.
func (f *pageFlip) step() {
	if !f.active || f.dragging {
		return
	}
	left := f.target - f.progress
	move := math.Max(math.Abs(left)*0.2, 1.0/40)
	switch {
	case math.Abs(left) <= move:
		f.progress = f.target
	case left > 0:
		f.progress += move
	default:
		f.progress -= move
	}
	f.dy *= 0.8
	if f.progress != f.target {
		return
	}
	if f.target == 0 {
		*f = pageFlip{}
		return
	}
	f.tuck += 1.0 / tuckTicks
	if f.tuck >= 1-1e-9 {
		f.done = true
	}
}

// nextPage is the page after p in the guest's folder; the stack cycles, the
// turned page goes to the back.
func nextPage(g *Guest, p Page) Page {
	order := pageOrder(g)
	return order[(slices.Index(order, p)+1)%len(order)]
}

// flipReady checks the folder is open and still before a flip starts.
func (s *Shift) flipReady() error {
	d := &s.desk
	switch {
	case d.guest == nil:
		return ErrNoFolder
	case !d.open:
		return ErrFolderClosed
	case d.flip.active || s.FolderOpen() < 1:
		return ErrFolderBusy
	}
	return nil
}

// BeginFlipDrag takes the corner of the top page at screen point (x, y).
func (s *Shift) BeginFlipDrag(x, y int) error {
	if err := s.flipReady(); err != nil {
		return err
	}
	s.desk.flip = pageFlip{active: true, dragging: true, startX: x, startY: y}
	return nil
}

// DragFlipTo moves the held corner to the screen point (x, y).
func (s *Shift) DragFlipTo(x, y int) {
	if f := &s.desk.flip; f.active && f.dragging {
		f.progress = dragProgress(f.startX, x)
		f.dy = min(max(float64(y-f.startY), -70), 10) // the corner may rise, hardly sink
	}
}

// EndFlipDrag lets go of the corner.
func (s *Shift) EndFlipDrag() {
	if f := &s.desk.flip; f.active && f.dragging {
		f.release()
	}
}

// Dragging says whether the player holds a page corner.
func (s *Shift) Dragging() bool { return s.desk.flip.active && s.desk.flip.dragging }

// TurnPage turns the top page by itself (the backdoor's "turn"), the same
// flip a drag finishes.
func (s *Shift) TurnPage() error {
	if err := s.flipReady(); err != nil {
		return err
	}
	s.desk.flip = pageFlip{active: true, target: 1, dy: -30} // a slanted curl looks best
	return nil
}

// GoToPage shows a page at once, with no turning (the backdoor's scene ids
// and script use it). It opens the folder if it is closed.
func (s *Shift) GoToPage(p Page) error {
	d := &s.desk
	if d.guest == nil {
		return ErrNoFolder
	}
	if !hasPage(d.guest, p) {
		return fmt.Errorf("%s: %w", p, ErrNoSuchPage)
	}
	if !d.open {
		d.open, d.openedAt = true, s.ticks-openTicks
	}
	d.page, d.flip = p, pageFlip{}
	return nil
}

// Turning says whether the top page is turning and how far (0 to 1).
func (s *Shift) Turning() (progress float64, ok bool) {
	return s.desk.flip.progress, s.desk.flip.active
}

// Curl is where the held corner is, in the page's own pixels (the page's
// bottom-right corner is (pageW, pageH)), and how far the turned page has
// tucked to the back (0 to 1).
func (s *Shift) Curl() (corner point, tuck float64) {
	f := s.desk.flip
	return point{float32(pageW - f.progress*flipTravel), float32(pageH + f.dy)}, f.tuck
}

// LiftStamp takes a stamp off the desk into the hand.
func (s *Shift) LiftStamp(st Stamp) {
	s.desk.holding, s.desk.held = true, st
}

// PutBackStamp puts the stamp in the hand back on the desk.
func (s *Shift) PutBackStamp() { s.desk.holding = false }

// ApplyStamp presses the stamp in the hand onto the open folder's referral
// page. It is the same stamping as the reception's (Shift.Stamp): the
// guest is admitted, referred or discharged and a wrong stamp costs the
// points and brings a citation. Then the folder closes and slides back.
func (s *Shift) ApplyStamp() error {
	if err := s.CheckApply(); err != nil {
		return err
	}
	d := &s.desk
	g, before := d.guest, s.Score()
	if err := s.Stamp(d.held); err != nil {
		return err
	}
	if o := d.out; o != nil && o.guest == g { // the desk freezes for the hit-stop
		o.impact, o.points = s.ticks, s.Score()-before
		o.at = o.impact + hitStopTicks
	}
	return nil
}

// CheckApply says whether the stamp in the hand can be pressed on the folder
// now, with the same errors ApplyStamp gives. The desk scene asks before it
// starts the landing, so a stamp that cannot land is not slammed.
func (s *Shift) CheckApply() error {
	d := &s.desk
	switch {
	case !d.holding:
		return ErrNoStampHeld
	case d.guest == nil:
		return ErrNoFolder
	case !d.open || s.FolderOpen() < 1:
		return ErrFolderClosed
	case d.page != PageReferral:
		return ErrWrongPage
	case d.guest == s.dozing:
		s.say("%s IS ASLEEP. WAKE THEM FIRST", d.guest.Name)
		return ErrDozing
	}
	return nil
}

// folderLeaves is called by Stamp: whatever way the guest was stamped, their
// folder goes back to the window, with a citation slip when the call was wrong.
func (s *Shift) folderLeaves(g *Guest, st Stamp) {
	d := &s.desk
	if d.guest == g {
		d.out = &sent{guest: g, stamp: st, page: d.page, open: d.open && s.FolderOpen() == 1, at: s.ticks}
		d.guest, d.open, d.flip, d.holding = nil, false, pageFlip{}, false
	}
	if right := g.RightStamp(); right != st {
		d.slip = &citation{stamped: st, right: right, until: s.ticks + citationTicks}
	}
}

// folderPose is how the folder looks this tick.
type folderPose struct {
	guest *Guest
	y     int     // pixels the folder is above its resting place (0 or less)
	open  float64 // 0 closed, 1 open
	page  Page
	ink   bool  // the ink mark shows on the referral
	stamp Stamp // which stamp the ink is of
}

// slideDistance is how far above its place a folder starts and ends: past
// the top of the screen.
const slideDistance = 260

// FolderPose works out the look of the folder on the desk (or on its way
// out): where it is, how far open, which page, and whether the ink shows.
// ok is false when there is no folder to draw.
func (s *Shift) FolderPose() (pose folderPose, ok bool) {
	d := &s.desk
	if g := d.guest; g != nil {
		in := float64(s.ticks-d.arrivedAt) / slideInTicks
		return folderPose{guest: g, y: -int(slideDistance * (1 - ease(in))), open: s.FolderOpen(), page: d.page}, true
	}
	if o := d.out; o != nil {
		dt := s.ticks - o.at
		pose = folderPose{guest: o.guest, page: o.page, open: 0}
		switch {
		case dt < inkTicks && o.open: // the stamp has landed, the folder is still open
			// (dt is negative during the hit-stop: no ink yet; the folder dips 1 px)
			pose.open, pose.ink, pose.stamp = 1, o.page == PageReferral && dt >= 0, o.stamp
			pose.y = folderDip(s.ticks - o.impact)
		case dt < inkTicks+closeTicks && o.open:
			pose.open = 1 - float64(dt-inkTicks)/closeTicks
			pose.ink, pose.stamp = o.page == PageReferral && pose.open > 0.5, o.stamp
		default:
			start := inkTicks + closeTicks
			if !o.open {
				start = 0
			}
			pose.y = -int(slideDistance * ease(float64(dt-start)/slideOutTicks))
		}
		return pose, true
	}
	return folderPose{}, false
}

// Slip is the citation slip on the desk: what was stamped, what was right,
// and whether there is one.
func (s *Shift) Slip() (stamped, right Stamp, ok bool) {
	if c := s.desk.slip; c != nil {
		return c.stamped, c.right, true
	}
	return 0, 0, false
}

// citationLineWidth is how wide a line written on the slip may be: the lines
// of the doc_citation sheet are 86 px.
const citationLineWidth = 86

// citationLines are the two lines written on the slip; the rest is printed.
// "YOU:" and "RIGHT:" are short so that "DISCHARGE" still fits the line.
func citationLines(stamped, right Stamp) (string, string) {
	return "YOU: " + strings.ToUpper(stamped.String()), "RIGHT: " + strings.ToUpper(right.String())
}

// SinceImpact is how many ticks ago a stamp hit the page, while the effects
// that follow it (shake, dip, ink pop, score) are still running.
func (s *Shift) SinceImpact() (int, bool) {
	if o := s.desk.out; o != nil && o.impact > 0 {
		return s.ticks - o.impact, true
	}
	return 0, false
}

// InkAge is how many ticks the ink mark has been showing (negative during
// the hit-stop), for its pop-in.
func (s *Shift) InkAge() int {
	if o := s.desk.out; o != nil {
		return s.ticks - o.at
	}
	return 0
}

// ScorePopup is the floating "+100" next to the mark, with how far it has
// risen and how opaque it is; ok is false when nothing floats.
func (s *Shift) ScorePopup() (text string, rise, alpha float64, ok bool) {
	o := s.desk.out
	if o == nil || o.points <= 0 {
		return "", 0, 0, false
	}
	rise, alpha = popup(s.ticks - o.at) // it starts when the ink appears
	return fmt.Sprintf("+%d", o.points), rise, alpha, alpha > 0
}
