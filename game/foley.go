package game

import (
	"fmt"
	"math/rand"
)

// The desk's paper sounds (Game Design -> design log, 2026-10-14). Each one
// is played where the action happens, once, never on every frame of a drag:
// the folder slides in or out, its cover opens or closes, a page corner is
// lifted, a page finishes turning, or a lifted page springs back flat.

// pageVariants is how many page rustles there are: page_1, page_2, page_3.
const pageVariants = 3

// Gains of the page rustle: lifting a corner and the turn that finishes are
// clear, the spring-back is only a whisper.
const (
	pageLiftGain = 0.5
	pageTurnGain = 0.7
	pageBackGain = 0.25
)

// paperState picks the page rustles: its own random source (so the snores
// and the blink timers keep theirs) and the variant that played last.
type paperState struct {
	rng  *rand.Rand
	last int
}

// pageSound plays one page rustle at gain, never the same one twice in a row.
func (s *Shift) pageSound(gain float64) {
	v := nextVariant(s.paper.last, pageVariants, s.paper.rng.Float64())
	s.paper.last = v
	playSFXGain(fmt.Sprintf("page_%d", v), gain)
}

// folderSounds plays the sounds of a folder going back to the window: the
// cover shuts (if it was open), then the folder slides away. Each plays once
// when its move begins.
func (s *Shift) folderSounds() {
	o := s.desk.out
	if o == nil {
		return
	}
	dt := s.ticks - o.at
	if o.open && !o.closeHeard && dt >= inkTicks {
		o.closeHeard = true
		playSFX("folder_close")
	}
	slideAt := 0
	if o.open {
		slideAt = inkTicks + closeTicks
	}
	if !o.slideHeard && dt >= slideAt {
		o.slideHeard = true
		playSFX("folder_slide")
	}
}
