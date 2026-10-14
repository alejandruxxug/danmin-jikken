package night

// The blood on the camera at the end of the Forever ending (Game Design
// design log): on the frame of the cut, three splash frames cover the last
// image, then the picture cuts to black but the stain stays, and five drips
// slide slowly down it. The ending card is drawn behind all of it.
//
// Everything is a pure function of the ticks since the cut began, so a test
// can check it and the backdoor can wait to any moment.
const (
	splashFrameTicks = 2 // each of splash_1, splash_2, splash_3 stays 2 ticks
	splashFrames     = 3
	dripEvery        = 24 // a tick of sliding costs one pixel every 24 ticks (2.5 px/s)
	dripStagger      = 40 // each drip starts this many ticks after the one before
)

// BloodSheet is the sheet with the layers splash_1..3, stain and drip_1..5.
const BloodSheet = "fx_blood_splash"

// DripCount is how many drip layers the sheet has.
const DripCount = 5

// Blood is what to draw at one moment.
type Blood struct {
	Splash int            // 1 to 3 while a splash frame shows, else 0
	Stain  bool           // the stain layer is on (from the end of the splash on)
	Drips  [DripCount]int // pixels each drip has slid down (0 = not started yet)
}

// BloodAt is the blood t ticks after the cut began. A negative t (the cut
// has not happened) draws nothing.
func BloodAt(t int) Blood {
	var b Blood
	if t < 0 {
		return b
	}
	if t < splashFrameTicks*splashFrames {
		b.Splash = 1 + t/splashFrameTicks
		return b
	}
	b.Stain = true
	since := t - splashFrameTicks*splashFrames
	for i := range b.Drips {
		if wait := i * dripStagger; since > wait {
			b.Drips[i] = (since - wait) / dripEvery
		}
	}
	return b
}

// SplashTicks is how long the splash frames last.
const SplashTicks = splashFrameTicks * splashFrames
