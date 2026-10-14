package game

import "math"

// The stamp landing (Game Design -> design log, "The stamp landing is
// JUICY"). Every number that shapes it is in the table below, and every
// motion is a pure function of a frame count, so it can be tuned in one place
// and tested without a window. Frames are ticks (60 a second).
const (
	windupTicks  = 2  // the stamp lifts and stretches
	slamTicks    = 3  // it drops fast and squashes on contact
	hitStopTicks = 3  // the desk freezes on impact
	springTicks  = 8  // the stamp springs back up
	shakeTicks   = 9  // the screen shakes, decaying (0.15 s)
	dipTicks     = 4  // the folder dips 1 px and comes back
	popTicks     = 7  // the ink mark settles from 115% to 100% (0.12 s)
	popupTicks   = 45 // the score floats up and fades
	speckTicks   = 18 // the ink specks scatter and fade (0.3 s)

	windupLift   = 3.0  // px the stamp rises in the wind-up
	springRise   = 14.0 // px it rises when it springs back (the overshoot is on top)
	shakeAmp     = 2.0  // px of shake at the moment of impact
	popScale     = 1.15 // the ink mark's size when it lands
	popupRise    = 12.0 // px the score floats up
	impactFrame  = windupTicks + slamTicks
	landingTotal = windupTicks + slamTicks + hitStopTicks + springTicks // press to done
)

// stampPose is how the stamp in the hand looks on a frame of its landing:
// how far above the page it is (negative is up), and how it is squashed.
type stampPose struct {
	lift   float64 // px above the page, 0 on contact
	sx, sy float64 // scale: wider and shorter when squashed, narrower and taller when stretched
	done   bool    // the landing is over
}

// landingPose is the stamp at frame f since the press (0 is the press).
func landingPose(f int) stampPose {
	switch {
	case f < 0:
		return stampPose{sx: 1, sy: 1}
	case f < windupTicks: // wind-up: up 3 px and stretched taller
		return stampPose{lift: windupLift, sx: 0.92, sy: 1.15}
	case f < impactFrame: // slam: falls fast (it speeds up), squashing as it arrives
		i := float64(f-windupTicks+1) / slamTicks // 1/3, 2/3, 1
		fall := 1 - i*i                           // 0.89, 0.56, 0: most of the drop is at the end
		squash := 0.0
		if f == impactFrame-1 {
			squash = 1 // the contact frame
		}
		return stampPose{lift: windupLift * fall, sx: 1 + 0.15*squash, sy: 1 - 0.15*squash}
	case f < impactFrame+hitStopTicks: // hit-stop: held in the squash
		return stampPose{sx: 1.15, sy: 0.85}
	case f < landingTotal: // spring back up with a little overshoot, un-squashing
		t := float64(f-impactFrame-hitStopTicks+1) / springTicks
		return stampPose{lift: springRise * easeOutBack(t), sx: 1 + 0.15*(1-t), sy: 1 - 0.15*(1-t)}
	}
	return stampPose{sx: 1, sy: 1, done: true}
}

// easeOutBack rises to 1 with a small overshoot (about 10%) and settles.
func easeOutBack(t float64) float64 {
	t = min(max(t, 0), 1)
	const c1 = 1.70158
	const c3 = c1 + 1
	return 1 + c3*math.Pow(t-1, 3) + c1*math.Pow(t-1, 2)
}

// shakePattern is the fixed zig-zag of the screen shake (no random numbers,
// so a screenshot of a frame is always the same): x flips every frame, y
// every other frame.
var shakePattern = [shakeTicks][2]int{{1, 0}, {-1, 1}, {1, 0}, {-1, -1}, {1, 0}, {-1, 1}, {1, 0}, {-1, 0}, {1, 0}}

// shakeOffset is how far the desk is moved at frame f since the impact: up
// to shakeAmp px, decaying to nothing after shakeTicks.
func shakeOffset(f int) (dx, dy int) {
	if f < 0 || f >= shakeTicks {
		return 0, 0
	}
	amp := shakeAmp * (1 - float64(f)/shakeTicks)
	pattern := shakePattern[f]
	return int(math.Round(amp * float64(pattern[0]))), int(math.Round(amp / 2 * float64(pattern[1])))
}

// folderDip is how many pixels the folder sits lower at frame f since the
// impact: 1 px for the first frames, then back.
func folderDip(f int) int {
	if f >= 0 && f < dipTicks {
		return 1
	}
	return 0
}

// inkScale is the size of the ink mark at frame f since it appeared: it
// pops in at popScale and settles to 1 over popTicks, in two whole steps so
// the pixels stay square rather than smeared.
func inkScale(f int) float64 {
	switch {
	case f < 0:
		return 0
	case f < popTicks/2:
		return popScale
	case f < popTicks:
		return 1 + (popScale-1)/2
	}
	return 1
}

// popup is the floating score at frame f since the impact: how many pixels
// it has risen and how opaque it is (1 to 0, fading over the second half).
func popup(f int) (rise float64, alpha float64) {
	if f < 0 || f >= popupTicks {
		return 0, 0
	}
	t := float64(f) / popupTicks
	return popupRise * ease(t), min(1, 2*(1-t))
}

// speck is how far the ink specks have scattered at frame f since the ink
// appeared (0 to 1) and how opaque they are. The art's ink_specks layer is
// drawn with it once it is exported (desk_draw.go checks it exists).
func speck(f int) (spread, alpha float64) {
	if f < 0 || f >= speckTicks {
		return 0, 0
	}
	t := float64(f) / speckTicks
	return ease(t), 1 - t
}

// playSFX is the hook for the sound pass (M7): the game calls it where a
// sound belongs ("stamp_lift", "stamp_thunk") and it does nothing until the
// sounds exist.
func playSFX(name string) {}
