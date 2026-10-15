package game

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	edgeZone = 48 // the mouse this close to a side of the screen pans
	panSpeed = 3  // camera pixels per tick while panning
)

// receptionScene is the front reception desk: the room, the queue and the
// patients the ward is putting to sleep, the stamps, the HUD, and the "no
// room" alert when it is open. The day itself lives in env.Shift().
type receptionScene struct {
	env     *Env
	art     *DeskArt
	shift   *Shift
	ticks   int // updates since the scene started
	drawErr error
}

func (s *receptionScene) Enter(env *Env) error {
	s.env = env
	var err error
	if s.art, err = env.Art(); err != nil {
		return err
	}
	if err = env.StartShift(); err != nil {
		return err
	}
	s.shift = env.Shift()
	return nil
}

// Exit leaves the shift running: the computer and the top-10 share it.
func (s *receptionScene) Exit() {}

func (s *receptionScene) Update() error {
	if s.drawErr != nil {
		return s.drawErr
	}
	s.ticks++
	s.pan()
	if err := tutorialInput(s.env, s.shift, hintReception); err != nil {
		return err
	}
	if x, y, ok := s.env.Input.Click(); ok {
		s.click(x, y)
	}
	return nil
}

// pan moves the camera when the mouse is near the left or right edge of
// the screen, like the FNAF 2 office. There are no arrows.
func (s *receptionScene) pan() {
	x, _ := s.env.Input.Cursor()
	switch {
	case x < edgeZone:
		s.env.CameraX -= panSpeed
	case x >= ScreenW-edgeZone:
		s.env.CameraX += panSpeed
	}
	s.env.CameraX = min(max(s.env.CameraX, 0), MaxCameraX)
}

// click handles a left click at (x, y) in screen pixels.
func (s *receptionScene) click(x, y int) {
	if hintClicked(s.shift, hintReception, x, y) { // the tutorial box takes the click
		return
	}
	if alert := s.shift.Alert(); alert != nil { // the alert is a box you must answer
		s.clickAlert(x, y)
		return
	}
	if tabRect.has(x, y) {
		s.env.Tab = TabBeds
		s.env.SwitchTo(SceneDayComputer)
		return
	}
	if filesRect.has(x, y) {
		s.env.SwitchTo(SceneDayDesk)
		return
	}
	wx := x + int(s.env.CameraX) // the same click in the room's pixels
	// The stamps are used on the look-down desk, on the papers (Game Design
	// -> design log): a click on a stamp or its cubby looks down to the desk.
	for _, st := range stampRects {
		for _, r := range st.rects {
			if r.has(wx, y) {
				s.env.SwitchTo(SceneDayDesk)
				return
			}
		}
	}
	for _, g := range s.shift.Guests {
		if hit := guestRect(s.shift, g); hit.has(wx, y) {
			s.attend(g)
			return
		}
	}
}

// attend sends staff to a sleeping guest the player clicked (or wakes the
// guest who dozed off at the window).
func (s *receptionScene) attend(g *Guest) {
	if g == s.shift.Dozing() { // the day 1 tutorial: a click wakes her up
		_ = s.shift.Wake(g)
		return
	}
	if !s.shift.needsStaff(g) {
		return // awake, in a bed or not here: nothing to do
	}
	if att, err := s.shift.Dispatch(g); err == nil && att.NoRoom {
		showAlert(s.env)
	}
}

// showAlert pans the camera to the hallway, where the patient lies.
func showAlert(env *Env) { env.CameraX = MaxCameraX }

// clickAlert answers the "no room" box.
func (s *receptionScene) clickAlert(x, y int) {
	switch {
	case callButton.has(x, y):
		if att, err := s.shift.CallStaff(); err == nil && att.NoRoom {
			showAlert(s.env)
		}
	case bedsButton.has(x, y):
		s.shift.DismissAlert()
		s.env.Tab = TabBeds
		s.env.SwitchTo(SceneDayComputer)
	}
}

// guestRect is where a guest can be clicked in the room, nothing for guests
// who are not visibly asleep.
func guestRect(sh *Shift, g *Guest) rect {
	x, y := int(g.x), int(g.y)
	switch sh.LookOf(g) {
	case LookDozing: // the visible part of the window sprite
		return rect{x, y, 96, windowGlass.y + windowGlass.h - y}
	case LookAsleep:
		return rect{x, y, 32, 64}
	case LookHallway:
		return rect{x, y, 48, 24}
	}
	return rect{}
}

func (s *receptionScene) Draw(screen *ebiten.Image) {
	p := &painter{art: s.art, screen: screen, day: s.env.Day, elapse: time.Duration(s.ticks) * time.Second / 60, tutTicks: s.shift.TutorialTicks()}
	p.room(s.shift, int(s.env.CameraX))
	p.hud(s.shift, s.env.Clock.String(), "^ PATIENT COMPUTER ^")
	p.toast(s.shift.Banner())
	if a := s.shift.Alert(); a != nil {
		p.alert(a)
	}
	p.tutorial(s.shift, hintReception, int(s.env.CameraX))
	// The glass is on the monitor, not in the room: it never pans.
	p.glass()
	if p.err != nil && s.drawErr == nil {
		s.drawErr = p.err
	}
}
