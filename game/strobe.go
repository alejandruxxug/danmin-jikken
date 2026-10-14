package game

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// The day-to-night strobe (Art Direction -> Day to night strobe). The
// timeline is a pure function of the tick, so a test can check it and the
// agent can wait to any moment of it.
//
//	1 flicker   1.5 s   the day, inverted for a single frame once or twice
//	2 strobe    0.6-1 s the picture flips normal <-> inverted
//	3 white     0.5 s   full white
//	4 fall      2 s     the white thins out to solid black, with afterimage flashes; the night script opens on black
const (
	flickerTicks = 90
	whiteTicks   = 30
	fallTicks    = 120
)

// strobeTicks is how long the full strobe lasts into each night: it gets
// longer, so the stun lands hardest on night 3.
var strobeTicks = map[int]int{1: 36, 2: 48, 3: 60}

// strobeLayer says what the strobe shows at a moment.
type strobeLayer int

const (
	layerDay strobeLayer = iota
	layerWhite
	layerFall
)

func (l strobeLayer) String() string {
	return [...]string{"day", "white", "fall"}[l]
}

// StrobeState is what to draw at one tick.
type StrobeState struct {
	Layer  strobeLayer
	Invert float64 // 0 normal, 1 inverted; between the two only in "reduce flashing"
	Fall   float64 // layerFall: progress from 0 to 1
}

// flickerFlips are the ticks of the first phase where the picture inverts
// for a single frame. Half the players should not be sure they saw it.
var flickerFlips = map[int]bool{38: true, 71: true}

// strobeFrames is the order of the frames of the full strobe, three ticks
// each. Only the day flashes (normal, inverted): nothing of the night may
// show before its script opens on black (Alejo: the leak bug).
var strobeFrames = []StrobeState{
	{layerDay, 1, 0}, {layerDay, 0, 0}, {layerDay, 1, 0}, {layerDay, 0, 0},
	{layerDay, 1, 0}, {layerDay, 0, 0}, {layerDay, 1, 0}, {layerDay, 1, 0},
}

const ticksPerStrobeFrame = 3

// StrobeAt is the state at a tick of the strobe into night n (1 to 3). With
// reduce on (Settings -> Reduce flashing) the flips are replaced by one slow
// invert-and-fade over the same time.
func StrobeAt(n, tick int, reduce bool) StrobeState {
	strobe := strobeTicks[n]
	switch {
	case tick < flickerTicks:
		if !reduce && flickerFlips[tick] {
			return StrobeState{Layer: layerDay, Invert: 1}
		}
		return StrobeState{Layer: layerDay}
	case tick < flickerTicks+strobe:
		t := tick - flickerTicks
		if reduce {
			return StrobeState{Layer: layerDay, Invert: float64(t) / float64(strobe)}
		}
		return strobeFrames[(t/ticksPerStrobeFrame)%len(strobeFrames)]
	case tick < flickerTicks+strobe+whiteTicks:
		return StrobeState{Layer: layerWhite}
	}
	fall := float64(tick-flickerTicks-strobe-whiteTicks) / fallTicks
	return StrobeState{Layer: layerFall, Fall: min(fall, 1)}
}

// StrobeLength is the whole strobe into night n, in ticks.
func StrobeLength(n int) int { return flickerTicks + strobeTicks[n] + whiteTicks + fallTicks }

// afterimageTicks are the ticks of the fall where a faint white flash comes back.
var afterimageTicks = []int{22, 52, 84}

// strobeScene is the day-to-night transition. It plays, then goes on to the night.
type strobeScene struct {
	env   *Env
	night int    // 1 to 3
	next  string // the night scene it ends in
	art   *DeskArt
	ticks int
}

func (s *strobeScene) Enter(env *Env) error {
	s.env = env
	var err error
	if s.art, err = env.Art(); err != nil {
		return err
	}
	// The shift ends inside the strobe: stop its goroutines and save the
	// score without showing the top-10 (it comes once, at the end of the run).
	if err := env.EndShiftSilently(); err != nil {
		return err
	}
	env.SkipStrobe = false
	// Coming from the live day, the strobe goes on where it was.
	if env.StrobeNight == s.night {
		s.ticks = env.StrobeTick
	}
	return nil
}

func (s *strobeScene) state() StrobeState {
	return StrobeAt(s.night, s.ticks, s.env.ReduceFlashing)
}

func (s *strobeScene) Update() error {
	s.ticks++
	if s.env.SkipStrobe || s.ticks >= StrobeLength(s.night) {
		s.env.SwitchTo(s.next)
		return nil
	}
	s.env.Invert = s.state().Invert
	s.env.StrobeNight, s.env.StrobeTick = s.night, s.ticks
	return nil
}

// UsesCRT: the day half is still the CRT (the shader does the inversion);
// once the night comes up through the white there is no glass any more.
func (s *strobeScene) UsesCRT() bool {
	return s.state().Layer != layerFall && s.state().Layer != layerWhite
}

func (s *strobeScene) Draw(screen *ebiten.Image) {
	st := s.state()
	switch st.Layer {
	case layerDay:
		p := &painter{art: s.art, screen: screen, day: s.env.Day, elapse: time.Duration(s.ticks) * time.Second / 60}
		p.room(nil, DefaultCameraX)
		p.glass()
	case layerWhite:
		screen.Fill(colPaper)
	case layerFall:
		screen.Fill(paletteInk) // the fall ends on solid black
		s.whiteOver(screen, st)
	}
}

// fallWhite is how much white lies over the black at a moment of the fall
// (0 to 1): it thins out to nothing, and three short afterimage flashes
// come back on the way. At the end of the fall it is exactly 0.
func fallWhite(fall float64) float64 {
	alpha := (1 - fall) * (1 - fall)
	t := int(fall * fallTicks)
	for _, at := range afterimageTicks {
		if t >= at && t < at+4 {
			alpha = max(alpha, 0.35*(1-fall))
		}
	}
	return alpha
}

// whiteOver lays that white on the black.
func (s *strobeScene) whiteOver(screen *ebiten.Image, st StrobeState) {
	alpha := fallWhite(st.Fall)
	vector.FillRect(screen, 0, 0, ScreenW, ScreenH, color.NRGBA{0xec, 0xec, 0xe6, uint8(alpha * 255)}, false)
}

func (s *strobeScene) Exit() { s.env.Invert, s.env.StrobeNight, s.env.StrobeTick = 0, 0, 0 }

// UpdateLiveStrobe runs the strobe over the live day (Alejo: it starts while
// you finish the last patient). The flicker begins when the last guest is at
// the desk and repeats until their stamp lands; then the full strobe takes
// over, and when it reaches the white the strobe scene carries on and the
// shift ends inside it. It returns true at that hand-over. Only days that
// lead into a night have one (day 1 for now; nights 2 and 3 come in M6).
func (e *Env) UpdateLiveStrobe(sh *Shift) (handOver bool) {
	if _, ok := strobeTicks[e.Day]; !ok || e.Day != 1 {
		return false
	}
	last := sh.Served() == sh.Total() || sh.Over() // the last stamp has landed
	if e.StrobeNight == 0 {
		if !last && !sh.LastGuestAtDesk() {
			return false
		}
		e.StrobeNight, e.StrobeTick = e.Day, 0
	}
	e.StrobeTick++
	if !last && e.StrobeTick >= flickerTicks {
		e.StrobeTick = 0 // still stamping: keep flickering
	}
	if last && e.StrobeTick < flickerTicks {
		e.StrobeTick = flickerTicks // the stamp landed: the full strobe takes over
	}
	st := StrobeAt(e.StrobeNight, e.StrobeTick, e.ReduceFlashing)
	e.Invert = st.Invert
	return st.Layer == layerWhite
}
