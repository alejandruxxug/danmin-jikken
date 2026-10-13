package game

import (
	"fmt"
	"time"

	"danmin/hospital"

	"github.com/hajimehoshi/ebiten/v2"

	"danmingame/sprites"
)

// The reception desk, built from the approved Figma view "Day - Reception
// desk" (Art Direction -> Art log). Everything is in world pixels: the room
// is 720 wide and the camera shows 480 of it. The positions below are the
// ones in the Figma note.
const (
	RoomW          = 720
	MaxCameraX     = RoomW - ScreenW // 240
	DefaultCameraX = MaxCameraX / 2  // 120: the Figma frame's crop

	edgeZone  = 48 // the mouse this close to a side of the screen pans
	panSpeed  = 3  // camera pixels per tick while panning
	walkTicks = 60 // a patient takes one second to walk to their spot
	walkFrom  = 120

	firstArrival  = 1 // seconds (game time) until the first patient walks up
	arrivalPeriod = 6 // seconds between patients
)

// A spot a patient can stand or sit at, in world pixels (sprite top-left).
type spot struct{ x, y int }

var (
	windowSpot = spot{312, 64} // behind the counter glass; the desk hides below y 186
	chairSpots = []spot{{154, 110}, {180, 110}}
)

// The layers that make up a patient, back to front. Awake and asleep differ
// only in the eyes, the mouth and the zzz.
var (
	cecilyAwake   = []string{"fill", "lineart", "detail", "eyes_normal", "mouth_smile"}
	fisherAwake   = []string{"fill", "lineart", "detail", "eyes_normal", "mouth_neutral"}
	patientAsleep = []string{"fill", "lineart", "detail", "eyes_closed", "mouth_slack", "fx_zzz"}
)

var (
	roomLayers  = []string{"fill", "lineart", "detail"}
	glassLayers = []string{"dust", "glare", "bezel"} // "reflection" is for day 3
)

// receptionScene is the front reception desk: the room, the patients the
// ward is putting to sleep, the desk and the glass over the screen.
type receptionScene struct {
	env *Env

	room, desk, glass, window, waiting *sprites.Sheet

	ticks     int               // updates since the scene started
	arrivedAt []int             // arrivedAt[i] is the tick patient i walked in
	snap      hospital.Snapshot // the ward as of this tick
	drawErr   error             // first error from drawing; Update returns it
}

func (s *receptionScene) Enter(env *Env) error {
	s.env = env
	var err error
	for _, l := range []struct {
		sheet **sprites.Sheet
		name  string
	}{
		{&s.room, "reception_bg"}, {&s.desk, "desk"}, {&s.glass, "screen_glass"},
		{&s.window, "cecily"}, {&s.waiting, "waiting_fisher"},
	} {
		if *l.sheet, err = sprites.Load(env.Sprites, l.name); err != nil {
			return fmt.Errorf("reception: %w", err)
		}
	}
	return env.StartWard()
}

func (s *receptionScene) Update() error {
	if s.drawErr != nil {
		return s.drawErr
	}
	s.ticks++
	s.pan()

	// Collect what the patient goroutines reported since the last frame.
	// The picture itself comes from the snapshot, never from the events.
	s.env.Ward().Drain()
	s.snap = s.env.Ward().Snapshot()

	// The day's queue walks up one patient at a time. Game time is the
	// ticks scaled by the speed, so -speed shortens the wait.
	gameSeconds := float64(s.ticks) / 60 * s.env.Speed
	for n := len(s.arrivedAt); n < len(s.snap.Patients); n++ {
		if gameSeconds < float64(firstArrival+arrivalPeriod*n) {
			break
		}
		s.arrivedAt = append(s.arrivedAt, s.ticks)
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

func (s *receptionScene) Exit() {
	// Stop the goroutines before the scene is gone. An error here would
	// mean a hospital call failed in the background; nothing can fix it now.
	_ = s.env.StopWard()
}

func (s *receptionScene) Draw(screen *ebiten.Image) {
	cam := int(s.env.CameraX)
	boil := s.room.FrameAt(sprites.BoilTag, time.Duration(s.ticks)*time.Second/60)

	s.keep(s.room.Draw(screen, roomLayers, boil, -cam, 0))
	// Patients stand between the room and the desk, so the counter hides
	// the window patient's body below its top edge.
	s.drawPatients(screen, cam, boil)
	// The desk shares the room's origin, so it pans with it.
	s.keep(s.desk.Draw(screen, roomLayers, boil, -cam, 0))
	// The glass is on the monitor, not in the room: it never pans.
	s.keep(s.glass.Draw(screen, glassLayers, 0, 0, 0))
}

// drawPatients puts the first patient at the window and the next ones in
// the chairs, awake or asleep as the snapshot says. Patients still in the
// queue have not walked in yet and are not drawn. The window patient uses
// Cecily's sprite for now, whoever it is.
func (s *receptionScene) drawPatients(screen *ebiten.Image, cam, boil int) {
	for i, arrived := range s.arrivedAt {
		if i >= len(s.snap.Patients) || i > len(chairSpots) {
			break
		}
		walked := min(s.ticks-arrived, walkTicks)
		behind := walkFrom * (walkTicks - walked) / walkTicks // pixels left to walk

		sheet, at, awake := s.waiting, spot{}, fisherAwake
		if i == 0 {
			sheet, at, awake = s.window, windowSpot, cecilyAwake
			at.x += behind // walks in from the right
		} else {
			at = chairSpots[i-1]
			at.x -= behind // walks in from the left
		}
		layers := awake
		if s.snap.Patients[i].State != hospital.Awake {
			layers = patientAsleep
		}
		s.keep(sheet.Draw(screen, layers, boil, at.x-cam, at.y))
	}
}

// keep remembers the first drawing error, because Draw cannot return one.
func (s *receptionScene) keep(err error) {
	if err != nil && s.drawErr == nil {
		s.drawErr = err
	}
}
