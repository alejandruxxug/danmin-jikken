package game

import (
	"image"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// deskScene is the look-down desk: the files, the stamps and the day's memo
// seen from above. It is a flip-down overlay like the computer: the reception
// goes on behind it (patients still fall asleep) and the view tilts down and
// back with a 0.3 s warp.
type deskScene struct {
	// setup, when set, runs on Enter and the desk starts settled (no tilt):
	// the backdoor's scene ids use it to get a folder on the desk at once.
	setup func(*Shift) error

	env     *Env
	art     *DeskArt
	shift   *Shift
	ticks   int
	t       float64 // how far down the look is, 0 = reception, 1 = desk
	dir     int     // +1 tilting down, -1 tilting back, 0 settled
	drawErr error
	slam    *slamAnim // the stamp landing, while it plays
}

// slamAnim is a stamp landing on the open referral: frame counts from the
// press (landingPose), from is where the stamp's top-left was when pressed.
type slamAnim struct {
	stamp Stamp
	frame int
	from  image.Point
}

// landedAt is where the stamp's top-left lands: centred on the ink mark, just
// under the SIGNATURE label.
var landedAt = image.Pt(inkAt.X+22-15, inkAt.Y+7-15+1)

func (s *deskScene) Enter(env *Env) error {
	s.env = env
	var err error
	if s.art, err = env.Art(); err != nil {
		return err
	}
	if err = env.StartShift(); err != nil {
		return err
	}
	s.shift = env.Shift()
	s.shift.Desk().MarkVisited() // the FILES tab stops pulsing
	if s.setup != nil {
		s.t = 1
		return s.setup(s.shift)
	}
	s.dir = 1
	return nil
}

// Exit puts a lifted stamp back: it does not stay in the hand off the desk.
func (s *deskScene) Exit() {
	s.shift.PutBackStamp()
	s.shift.EndFlipDrag()
}

func (s *deskScene) Update() error {
	if s.drawErr != nil {
		return s.drawErr
	}
	s.ticks++
	if s.slam != nil {
		s.stepSlam()
		return nil
	}
	switch {
	case s.dir > 0:
		if s.t += 1.0 / lookDownTicks; s.t >= 1 {
			s.t, s.dir = 1, 0
		}
	case s.dir < 0:
		if s.t -= 1.0 / lookDownTicks; s.t <= 0 {
			s.env.SwitchTo(SceneDayReception)
		}
	default:
		if s.shift.Dragging() {
			s.drag()
		} else if x, y, ok := s.env.Input.Click(); ok {
			s.click(x, y)
		}
	}
	return nil
}

// drag follows the held page corner: the fold goes where the cursor is, and
// letting go finishes the turn or springs the page back.
func (s *deskScene) drag() {
	in := s.env.Input
	x, y := in.Cursor()
	s.shift.DragFlipTo(x, y)
	if in.Released() || !in.Held() {
		s.shift.EndFlipDrag()
	}
}

// stepSlam plays one frame of the stamp landing. The stamp logic runs on the
// contact frame (the same ApplyStamp as ever); the hit-stop is that the ink,
// the folder and the shake wait three frames for it.
func (s *deskScene) stepSlam() {
	f := s.slam.frame
	if f == impactFrame-1 {
		if err := s.shift.ApplyStamp(); err != nil { // it cannot land after all
			s.shift.PutBackStamp()
			s.slam = nil
			return
		}
		playSFX("stamp_thunk")
	}
	s.slam.frame++
	if landingPose(s.slam.frame).done {
		s.slam = nil
	}
}

// landing is the stamp in its landing, for the painter, or nil. Before the
// contact it glides from where it was pressed to where it lands.
func (s *deskScene) landing() *slamDraw {
	if s.slam == nil {
		return nil
	}
	f := min(s.slam.frame, impactFrame-1)
	glide := ease(float64(f) / float64(impactFrame-1))
	x := s.slam.from.X + int(float64(landedAt.X-s.slam.from.X)*glide)
	y := s.slam.from.Y + int(float64(landedAt.Y-s.slam.from.Y)*glide)
	return &slamDraw{stamp: s.slam.stamp, x: x, y: y, pose: landingPose(s.slam.frame)}
}

// click handles a click on the settled desk.
func (s *deskScene) click(x, y int) {
	sh := s.shift
	if filesRect.has(x, y) { // BACK
		sh.PutBackStamp()
		s.dir = -1
		return
	}
	if _, holding := sh.Desk().Holding(); holding {
		// A click anywhere on the open page presses the stamp (the ink
		// snaps into the FOR DESK USE box); a click elsewhere puts it back.
		if pageRect.has(x, y) {
			if err := sh.CheckApply(); err == nil {
				st, _ := sh.Desk().Holding()
				s.slam = &slamAnim{stamp: st, from: image.Pt(x-4, y-4)}
				playSFX("stamp_lift")
				return
			}
		}
		sh.PutBackStamp()
		return
	}
	for _, d := range deskStamps {
		if d.cubby.has(x, y) {
			sh.LiftStamp(d.stamp)
			return
		}
	}
	desk := sh.Desk()
	switch {
	case desk.FolderGuest() == nil:
	case !desk.IsOpen():
		if closedRect.has(x, y) {
			_ = sh.OpenFolder() // still sliding in: nothing happens yet
		}
	case grabRect.has(x, y):
		_ = sh.BeginFlipDrag(x, y) // the page is taken; drag() follows the mouse
	}
}

func (s *deskScene) Draw(screen *ebiten.Image) {
	elapse := time.Duration(s.ticks) * time.Second / 60
	p := &painter{art: s.art, screen: screen, day: s.env.Day, elapse: elapse, forged: s.env.Flags.Has("route_hana")}
	cx, cy := s.env.Input.Cursor()
	if s.t < 1 {
		// The reception goes into its own picture first, so it can be squashed.
		p.into(s.art.back, func(q *painter) {
			q.room(s.shift, int(s.env.CameraX))
			q.hud(s.shift, s.env.Clock.String(), "^ PATIENT COMPUTER ^")
			q.toast(s.shift.Message())
		})
	}
	p.into(s.art.top, func(q *painter) { q.deskObjects(s.shift, image.Pt(cx, cy), s.landing()) })
	if s.t >= 1 {
		var op ebiten.DrawImageOptions // the screen shakes after a stamp lands
		if since, ok := s.shift.SinceImpact(); ok {
			if dx, dy := shakeOffset(since); dx != 0 || dy != 0 {
				screen.Fill(colDim) // the strip the shake uncovers at the edge
				op.GeoM.Translate(float64(dx), float64(dy))
			}
		}
		screen.DrawImage(s.art.top, &op)
	} else {
		rtl, rtr, rbl, rbr, dtl, dtr, dbl, dbr := lookDownShapes(s.t)
		screen.Clear()
		full := image.Rect(0, 0, ScreenW, ScreenH)
		drawWarped(screen, s.art.back, full, rtl, rtr, rbl, rbr, 1, 12)
		drawWarped(screen, s.art.top, full, dtl, dtr, dbl, dbr, 1, 34)
	}
	if s.t > 0.5 {
		p.backTab()
	} else {
		p.filesTab(s.shift)
	}
	p.deskToast(s.shift.Message())
	p.glass()
	if p.err != nil && s.drawErr == nil {
		s.drawErr = p.err
	}
}

// deskToast shows the last message in a black bar at the bottom, right of the
// BACK tab (the top strip of the desk holds the window glass).
func (p *painter) deskToast(msg string) {
	if msg == "" {
		return
	}
	w := p.art.font.Width(msg) + 12
	r := rect{ScreenW/2 - w/2 + 60, 252, w, 14}
	p.fill(r, colInk)
	p.textCentered(msg, r.x+r.w/2, 256, colPaper)
}
