package game

import "math"

// The YOU reveal (Game Design -> design log 2026-10-14; Art Direction ->
// Art log, M7 portraits, storyboard SB1 to SB7). On day 3 the last patient
// is you, in a hospital gown with a tag on the chest. The camera punches in
// on the tag when you walk up, and after you stamp your own file it tilts
// up to your face, and the day goes into the strobe from that picture.
//
//	arrival   SB1  x4 view on the tag; the only input is looking down
//	desk      SB2  the desk as always; the folder and the stamps
//	back      SB3  back up, the same tag shot, input locked, 1.0 s
//	tilt      SB4  the view's top goes from y 120 to y 70 in 2.4 s (eased)
//	face      SB5  face hold 1.5 s, one blink at 0.8 s
//	stun      SB6  the day-to-night strobe starts from this picture
type RevealStage int

const (
	RevealOff RevealStage = iota
	RevealArrival
	RevealDesk
	RevealBack
	RevealTilt
	RevealFace
	RevealStun
)

func (s RevealStage) String() string {
	return [...]string{"off", "arrival", "desk", "back", "tilt", "face", "stun"}[s]
}

const (
	// The view is 120 x 67.5 world pixels, shown 4 times bigger (480 x 270),
	// centred on world x 360. Its top edge starts at the tag and ends at the face.
	revealZoom  = 4
	revealLeft  = 300 // world x of the view's left edge (360 - 120/2)
	revealTagY  = 120 // top edge on the tag: the chin stays just above it
	revealFaceY = 70  // top edge on the face

	backHoldTicks = 60  // 1.0 s
	tiltTicks     = 144 // 2.4 s
	faceHoldTicks = 90  // 1.5 s
	blinkAtTick   = 48  // 0.8 s into the face hold
	blinkTicks    = 9   // 150 ms
)

// Reveal is where the YOU reveal is. It lives in Env, because the reception,
// the desk and the strobe scene all draw it.
type Reveal struct {
	Stage RevealStage
	Tick  int    // ticks into the stage (counted in back, tilt and face)
	Guest *Guest // you; stays valid after the shift is stopped
}

// Active says whether the x4 view is on the screen.
func (r *Reveal) Active() bool { return r.Stage != RevealOff }

// InputLocked says whether the player's input is off: from SB3 to the night.
func (r *Reveal) InputLocked() bool { return r.Stage >= RevealBack }

// tiltTopY is the top edge of the view tick ticks into the tilt: a sine
// ease-in-out from revealTagY to revealFaceY, in whole pixels so the
// pixel art never shimmers.
func tiltTopY(tick int) int {
	if tick <= 0 {
		return revealTagY
	}
	if tick >= tiltTicks {
		return revealFaceY
	}
	t := float64(tick) / tiltTicks
	eased := (1 - math.Cos(math.Pi*t)) / 2
	return revealTagY - int(math.Round(eased*(revealTagY-revealFaceY)))
}

// TopY is the world y of the view's top edge right now.
func (r *Reveal) TopY() int {
	switch r.Stage {
	case RevealTilt:
		return tiltTopY(r.Tick)
	case RevealFace, RevealStun:
		return revealFaceY
	}
	return revealTagY
}

// Blinking is true for the one blink of the face hold.
func (r *Reveal) Blinking() bool {
	return r.Stage == RevealFace && r.Tick >= blinkAtTick && r.Tick < blinkAtTick+blinkTicks
}

// LookDown is the player looking down at the desk.
func (r *Reveal) LookDown() {
	if r.Stage == RevealArrival {
		r.Stage = RevealDesk
	}
}

// LookUp is the player looking back up. After the stamp it is SB3; with
// the folder not stamped yet it is the tag shot again.
func (r *Reveal) LookUp(stamped bool) {
	if r.Stage != RevealDesk {
		return
	}
	if stamped {
		r.Stage, r.Tick = RevealBack, 0
		return
	}
	r.Stage = RevealArrival
}

// step runs one tick of the timed stages: back, tilt, face, then stun.
func (r *Reveal) step() {
	limits := map[RevealStage]int{RevealBack: backHoldTicks, RevealTilt: tiltTicks, RevealFace: faceHoldTicks}
	limit, timed := limits[r.Stage]
	if !timed {
		return
	}
	if r.Tick++; r.Tick >= limit {
		r.Stage, r.Tick = r.Stage+1, 0
	}
}

// UpdateReveal runs once per tick on day 3, after the shift's own update.
// It starts the reveal when you walk up to the window, and runs its clock.
func (e *Env) UpdateReveal(sh *Shift) {
	r := &e.Reveal
	switch r.Stage {
	case RevealOff:
		if g := sh.SelfAtWindow(); g != nil && !sh.Over() {
			*r = Reveal{Stage: RevealArrival, Guest: g}
		}
	case RevealArrival, RevealDesk:
		// The shift ran out before the folder was stamped: the reveal still
		// happens. From the desk the view goes back up first.
		// A stamp still landing at the desk is let land first: it sets the
		// stamped_self flag when it hits, and leaving now would drop it.
		if sh.Over() && !e.StampLanding {
			if r.Stage == RevealDesk {
				e.SwitchTo(SceneDayReception)
			}
			r.Stage, r.Tick = RevealBack, 0
		}
	default:
		r.step()
	}
}

// SelfAtWindow is you, once you are the one being served (nil on every
// other day and before that).
func (s *Shift) SelfAtWindow() *Guest {
	if g := s.window(); g != nil && g.Self {
		return g
	}
	return nil
}
