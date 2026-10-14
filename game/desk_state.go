package game

import (
	"errors"
	"fmt"
	"math"
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
)

func (p Page) String() string {
	switch p {
	case PageReferral:
		return "referral"
	case PageSleepLog:
		return "sleep log"
	case PageWristband:
		return "wristband"
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
	at    int  // tick the stamp landed
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
func pages(g *Guest) int {
	if g.Returning {
		return 3
	}
	return 2
}

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

// pageFlip is the top page being turned, by a drag or by TurnPage.
type pageFlip struct {
	active   bool
	dragging bool    // the player is holding the corner
	startX   int     // where the corner was pressed
	progress float64 // 0 to 1
	target   float64 // after letting go: 1 finishes the flip, 0 springs back
	done     bool    // the flip finished this tick: the next page is on top
}

// finishAt is the progress at which letting go finishes the flip.
const finishAt = 0.5

// release lets go of the corner: past the middle the page finishes its
// turn, before the middle it springs back.
func (f *pageFlip) release() {
	f.dragging = false
	f.target = 0
	if f.progress >= finishAt {
		f.target = 1
	}
}

// step moves a released flip toward its target, one turnTicks-th of the
// way a tick, and ends it there.
func (f *pageFlip) step() {
	if !f.active || f.dragging {
		return
	}
	d := 1.0 / turnTicks
	switch {
	case f.progress < f.target:
		f.progress = min(f.progress+d, f.target)
	case f.progress > f.target:
		f.progress = max(f.progress-d, f.target)
	}
	if math.Abs(f.progress-f.target) < 1e-9 { // sums of 1/14 are not exact
		f.progress = f.target
		f.done = f.target == 1
		if !f.done {
			*f = pageFlip{}
		}
	}
}

// nextPage is the page after p in the guest's folder; the stack cycles, the
// turned page goes to the back.
func nextPage(g *Guest, p Page) Page { return Page((int(p) + 1) % pages(g)) }

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

// BeginFlipDrag takes the corner of the top page at screen column x.
func (s *Shift) BeginFlipDrag(x int) error {
	if err := s.flipReady(); err != nil {
		return err
	}
	s.desk.flip = pageFlip{active: true, dragging: true, startX: x}
	return nil
}

// DragFlipTo moves the held corner to screen column x.
func (s *Shift) DragFlipTo(x int) {
	if f := &s.desk.flip; f.active && f.dragging {
		f.progress = dragProgress(f.startX, x)
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
	s.desk.flip = pageFlip{active: true, target: 1}
	return nil
}

// GoToPage shows a page at once, with no turning (the backdoor's scene ids
// and script use it). It opens the folder if it is closed.
func (s *Shift) GoToPage(p Page) error {
	d := &s.desk
	if d.guest == nil {
		return ErrNoFolder
	}
	if int(p) < 0 || int(p) >= pages(d.guest) {
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
	}
	g := d.guest
	if err := s.Stamp(d.held); err != nil {
		if errors.Is(err, ErrDozing) {
			s.say("%s IS ASLEEP. WAKE THEM FIRST", g.Name)
		}
		return err
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
			pose.open, pose.ink, pose.stamp = 1, o.page == PageReferral, o.stamp
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

// citationLines are the two lines written on the slip; the rest is printed.
func citationLines(stamped, right Stamp) (string, string) {
	return "STAMPED: " + strings.ToUpper(stamped.String()), "RIGHT CALL: " + strings.ToUpper(right.String())
}
