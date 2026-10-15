package main

import "math"

// The easing curves of the motion pass (Trailer note -> Motion pass). Each
// takes t from 0 to 1 and gives the progress from 0 to 1 (outBack overshoots a
// little before it settles).

func outCubic(t float64) float64 { return 1 - math.Pow(1-t, 3) }

func inCubic(t float64) float64 { return t * t * t }

func inQuad(t float64) float64 { return t * t }

func linear(t float64) float64 { return t }

func inOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	return 1 - math.Pow(-2*t+2, 3)/2
}

func outBack(t float64) float64 {
	const c1 = 1.70158
	const c3 = c1 + 1 // 2.70158
	return 1 + c3*math.Pow(t-1, 3) + c1*math.Pow(t-1, 2)
}

// Ease is an easing curve.
type Ease func(t float64) float64

// Move is a position that goes from one whole pixel value to another: the
// animation starts at frame start, lasts dur frames, and uses curve. Before
// it starts the value is From, after it ends it is To. The result is rounded
// to a whole 1x pixel, because the compositor never moves anything by a
// fraction of a pixel.
type Move struct {
	From, To   int
	Start, Dur int
	Curve      Ease
}

// At is the value at frame f.
func (m Move) At(f int) int {
	if f <= m.Start {
		return m.From
	}
	if f >= m.Start+m.Dur {
		return m.To
	}
	t := float64(f-m.Start) / float64(m.Dur)
	return m.From + int(math.Round(float64(m.To-m.From)*m.Curve(t)))
}
