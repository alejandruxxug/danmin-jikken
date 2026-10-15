package main

import (
	"math"
	"testing"
)

func TestEaseEnds(t *testing.T) {
	curves := map[string]Ease{
		"outCubic": outCubic, "inCubic": inCubic, "inQuad": inQuad, "linear": linear,
		"inOutCubic": inOutCubic, "outBack": outBack,
	}
	for name, c := range curves {
		if got := c(0); math.Abs(got) > 1e-9 {
			t.Errorf("%s(0) = %v, want 0", name, got)
		}
		if got := c(1); math.Abs(got-1) > 1e-9 {
			t.Errorf("%s(1) = %v, want 1", name, got)
		}
	}
}

func TestEaseValues(t *testing.T) {
	for _, c := range []struct {
		name string
		f    Ease
		t    float64
		want float64
	}{
		{"outCubic half", outCubic, 0.5, 0.875},
		{"inCubic half", inCubic, 0.5, 0.125},
		{"inQuad half", inQuad, 0.5, 0.25},
		{"inOutCubic half", inOutCubic, 0.5, 0.5},
		{"inOutCubic quarter", inOutCubic, 0.25, 0.0625},
		{"inOutCubic three quarters", inOutCubic, 0.75, 0.9375},
	} {
		if got := c.f(c.t); math.Abs(got-c.want) > 1e-9 {
			t.Errorf("%s = %v, want %v", c.name, got, c.want)
		}
	}
	// outBack overshoots 1 and then comes back.
	peak := 0.0
	for i := 0; i <= 100; i++ {
		peak = math.Max(peak, outBack(float64(i)/100))
	}
	if peak <= 1 || peak > 1.15 {
		t.Errorf("outBack peaks at %v, want a small overshoot above 1", peak)
	}
}

func TestMove(t *testing.T) {
	m := Move{From: -40, To: 0, Start: 10, Dur: 18, Curve: outBack}
	if m.At(0) != -40 || m.At(10) != -40 {
		t.Error("before the start the value is From")
	}
	if m.At(28) != 0 || m.At(500) != 0 {
		t.Error("after the end the value is To")
	}
	if got := m.At(19); got <= -40 || got >= 5 {
		t.Errorf("At(19) = %d, want a value between the ends", got)
	}
	slide := Move{From: 96, To: 0, Start: 46, Dur: 12, Curve: outCubic}
	if slide.At(52) != 12 { // outCubic(0.5) = 0.875, so 96 - 84
		t.Errorf("slide.At(52) = %d, want 12", slide.At(52))
	}
}
