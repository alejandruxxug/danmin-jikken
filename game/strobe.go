package game

import (
	"fmt"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"danmingame/night"
)

// The day-to-night strobe (Art Direction -> Day to night strobe). The
// timeline is a pure function of the tick, so a test can check it and the
// agent can wait to any moment of it.
//
//	1 flicker   1.5 s   the day, inverted for a single frame once or twice
//	2 strobe    0.6-1 s the picture flips normal <-> inverted; some flips are a frame of the night
//	3 white     0.5 s   full white
//	4 fall      2 s     the night comes up through white, drifting, with afterimage flashes
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
	layerNight
	layerWhite
	layerFall
)

func (l strobeLayer) String() string {
	return [...]string{"day", "night", "white", "fall"}[l]
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
// each: the day inverted, the day, the night inverted, the day, ...
var strobeFrames = []StrobeState{
	{layerDay, 1, 0}, {layerDay, 0, 0}, {layerNight, 1, 0}, {layerDay, 0, 0},
	{layerDay, 1, 0}, {layerNight, 0, 0}, {layerDay, 1, 0}, {layerNight, 1, 0},
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
	nart  *night.Art // the night frame that flashes through
	ticks int
}

func (s *strobeScene) Enter(env *Env) error {
	s.env = env
	var err error
	if s.art, err = env.Art(); err != nil {
		return err
	}
	if s.nart, err = env.NightArt(); err != nil {
		return err
	}
	// The shift is over: stop its goroutines, as the top-10 does.
	if err := env.StopShift(); err != nil {
		return fmt.Errorf("the shift ended with an error: %w", err)
	}
	env.SkipStrobe = false
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
	case layerNight:
		screen.Fill(paletteInk)
		s.nart.DrawStationFrame(screen, 0)
	case layerWhite:
		screen.Fill(colPaper)
	case layerFall:
		screen.Fill(colPaper)
		s.nart.DrawStationFrame(screen, int(24*(1-st.Fall)))
		s.whiteOver(screen, st)
	}
}

// whiteOver lays the over-exposed white on the rising night: it thins out
// as the fall goes on, and three short afterimage flashes come back.
func (s *strobeScene) whiteOver(screen *ebiten.Image, st StrobeState) {
	alpha := (1 - st.Fall) * (1 - st.Fall)
	t := int(st.Fall * fallTicks)
	for _, at := range afterimageTicks {
		if t >= at && t < at+4 {
			alpha = max(alpha, 0.35)
		}
	}
	vector.FillRect(screen, 0, 0, ScreenW, ScreenH, color.NRGBA{0xec, 0xec, 0xe6, uint8(alpha * 255)}, false)
}

func (s *strobeScene) Exit() { s.env.Invert, s.env.StrobeNight, s.env.StrobeTick = 0, 0, 0 }
