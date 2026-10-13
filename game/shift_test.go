package game

import (
	"math/rand"
	"testing"
)

func TestWalkBob(t *testing.T) {
	if WalkBob(0) != 0 {
		t.Error("standing must be exactly on the resting position")
	}
	// One step is 7.5 ticks: 0, -1, -2, -1, then the cycle repeats.
	want := map[int]int{1: 0, 8: -1, 15: -2, 23: -1, 30: 0, 38: -1}
	for steps, off := range want {
		if got := WalkBob(steps); got != off {
			t.Errorf("WalkBob(%d) = %d, want %d", steps, got, off)
		}
	}
	for steps := 1; steps < 200; steps++ {
		if got := WalkBob(steps); got > 0 || got < -2 {
			t.Fatalf("WalkBob(%d) = %d, outside 0..-2", steps, got)
		}
	}
}

func TestBlinker(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	var b blinker
	b.update(0, rng, true) // next is 0: the first blink starts now
	if !b.closed(0.05) || b.closed(blinkFor+0.01) {
		t.Errorf("blink should last %v s: closed(0.05)=%v closed(after)=%v", blinkFor, b.closed(0.05), b.closed(blinkFor+0.01))
	}
	gap := b.next
	if gap < blinkMin || gap > blinkMin+blinkSpread {
		t.Errorf("next blink in %v s, want %v to %v", gap, blinkMin, blinkMin+blinkSpread)
	}

	// Walking: the eyes stay open and the timer waits.
	b.update(gap+1, rng, false)
	if b.closed(gap + 1) {
		t.Error("blinked while walking")
	}
	if b.next <= gap+1 {
		t.Error("the timer should be pushed past now while walking")
	}
}

// Same seed, same blinks; different guests blink at different times.
func TestBlinkIsSeededAndNotInSync(t *testing.T) {
	first := func(seed int64) float64 {
		var b blinker
		rng := rand.New(rand.NewSource(seed))
		b.update(0, rng, true)
		return b.next
	}
	if first(3) != first(3) {
		t.Error("the same seed must give the same gap")
	}
	if first(3) == first(4) {
		t.Error("different seeds should give different gaps")
	}
}
