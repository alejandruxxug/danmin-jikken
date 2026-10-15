package main

// The pure pieces of the motion pass: the beat, the slam, the typewriter, the
// clock race, the glitch rows. Each is a function of the frame number only, so
// the tests can say exactly what happens on exactly which frame.

import "math"

// The jingle is 144 BPM: a beat every 25 frames. Beat k lands on frame 46 + 25k.
const (
	beatFirst = 46
	beatLen   = 25
)

// Beat is the trailer frame of beat k.
func Beat(k int) int { return beatFirst + beatLen*k }

// Slam is a stamp coming down: hidden before frame start, 24 pixels above its
// place for 2 frames, 12 above for 2 more, then down.
func Slam(f, start int) (dy int, visible bool) {
	switch {
	case f < start:
		return 0, false
	case f < start+2:
		return -24, true
	case f < start+4:
		return -12, true
	}
	return 0, true
}

// slamShake is the picture's shake after a stamp lands on frame landing: it
// swings 2, -2, 1, -1 pixels across x and is still after.
var slamShake = []int{2, -2, 1, -1}

// ShakeX is the horizontal shake on frame f of a slam that lands on frame landing.
func ShakeX(f, landing int) int {
	i := f - landing
	if i < 0 || i >= len(slamShake) {
		return 0
	}
	return slamShake[i]
}

// Typed is how many letters of a typed line show on frame f: the first appears
// on frame start and a new one every per frames, up to total.
func Typed(f, start, per, total int) int {
	if f < start {
		return 0
	}
	return min(total, (f-start)/per+1)
}

// The clock race (T10b): the digits change on the frames from clockFrom, with
// a gap that shrinks from 6 frames to 1 (inQuad), and show 03:17 from clockEnd on.
const (
	clockFrom     = 1568
	clockEnd      = 1668
	clockStartMin = 17*60 + 58 // 17:58
	clockEndMin   = 3*60 + 17  // 03:17
	clockSpan     = (24*60 - clockStartMin) + clockEndMin
)

// clockChanges are the frames the digits change on; the last is clockEnd.
func clockChanges() []int {
	var changes []int
	for f := clockFrom; f < clockEnd; {
		changes = append(changes, f)
		t := float64(f-clockFrom) / float64(clockEnd-clockFrom)
		f += max(1, int(math.Round(6-5*inQuad(t))))
	}
	return append(changes, clockEnd)
}

var clockFrames = clockChanges()

// ClockAt is the time the race shows on frame f. Before the first change it is
// 17:58. The minutes run through the night, spread evenly over the changes.
func ClockAt(f int) (hour, minute int) {
	n := 0 // changes so far
	for n < len(clockFrames) && clockFrames[n] <= f {
		n++
	}
	total := clockStartMin + int(math.Round(float64(clockSpan)*float64(n)/float64(len(clockFrames))))
	total %= 24 * 60
	return total / 60, total % 60
}

// ClockTicks are the frames the digits change on before 03:17 lands.
func ClockTicks() []int { return clockFrames[:len(clockFrames)-1] }

// GlitchRow is one 8-pixel row of the glitch-slice cut: which of the two
// pictures it shows and how far it is shifted sideways.
type GlitchRow struct {
	Y     int
	FromB bool
	Shift int
}

const glitchRowH = 8

// mix32 scrambles three numbers into one: a small deterministic hash, so the
// glitch looks the same on every render.
func mix32(a, b, c uint32) uint32 {
	h := a*0x9E3779B1 ^ b*0x85EBCA6B ^ c*0xC2B2AE35
	h ^= h >> 15
	h *= 0x2C1B3C6D
	h ^= h >> 12
	h *= 0x297A2D39
	h ^= h >> 15
	return h
}

// GlitchRows is the glitch-slice cut on its step-th frame (0 to steps-1): every
// 8-pixel row shows picture A (the old shot) or B (the new one), moved by 0, 8
// or 16 pixels either way. B takes over more rows with every step.
func GlitchRows(step, steps int, seed uint32) []GlitchRow {
	shifts := []int{-16, -8, 0, 0, 8, 16}
	var rows []GlitchRow
	for y := 0; y < GameH; y += glitchRowH {
		r := mix32(seed, uint32(step), uint32(y/glitchRowH))
		rows = append(rows, GlitchRow{
			Y:     y,
			FromB: int(r%uint32(steps+1)) <= step,
			Shift: shifts[(r>>8)%uint32(len(shifts))],
		})
	}
	return rows
}
