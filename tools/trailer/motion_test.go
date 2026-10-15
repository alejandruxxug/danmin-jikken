package main

import "testing"

func TestBeat(t *testing.T) {
	for k, want := range map[int]int{0: 46, 2: 96, 4: 146, 8: 246, 12: 346, 30: 796, 48: 1246} {
		if got := Beat(k); got != want {
			t.Errorf("Beat(%d) = %d, want %d", k, got, want)
		}
	}
}

func TestSlamAndShake(t *testing.T) {
	// The hiring stamp: y -24 on f142, -12 on f144, 0 on f146.
	for _, c := range []struct {
		f, dy int
		vis   bool
	}{{141, 0, false}, {142, -24, true}, {143, -24, true}, {144, -12, true}, {145, -12, true}, {146, 0, true}, {400, 0, true}} {
		if dy, vis := Slam(c.f, 142); dy != c.dy || vis != c.vis {
			t.Errorf("Slam(%d) = %d, %v; want %d, %v", c.f, dy, vis, c.dy, c.vis)
		}
	}
	want := map[int]int{145: 0, 146: 2, 147: -2, 148: 1, 149: -1, 150: 0, 151: 0}
	for f, dx := range want {
		if got := ShakeX(f, 146); got != dx {
			t.Errorf("ShakeX(%d) = %d, want %d", f, got, dx)
		}
	}
}

func TestTyped(t *testing.T) {
	// WE NEVER CLOSE.: 15 letters, one every 4 frames from f500, done on f556.
	for f, want := range map[int]int{499: 0, 500: 1, 503: 1, 504: 2, 520: 6, 555: 14, 556: 15, 900: 15} {
		if got := Typed(f, 500, 4, 15); got != want {
			t.Errorf("Typed(%d) = %d, want %d", f, got, want)
		}
	}
}

func TestClockRace(t *testing.T) {
	if h, m := ClockAt(1560); h != 17 || m != 58 {
		t.Errorf("before the race: %02d:%02d, want 17:58", h, m)
	}
	if h, m := ClockAt(clockEnd); h != 3 || m != 17 {
		t.Errorf("03:17 must land on frame %d, got %02d:%02d", clockEnd, h, m)
	}
	if h, m := ClockAt(clockEnd - 1); h == 3 && m == 17 {
		t.Error("03:17 must not show before its frame")
	}
	if h, m := ClockAt(4000); h != 3 || m != 17 {
		t.Errorf("after the race: %02d:%02d, want 03:17", h, m)
	}
	// The gaps shrink from 6 frames to 1.
	ticks := ClockTicks()
	if gap := ticks[1] - ticks[0]; gap != 6 {
		t.Errorf("first gap = %d, want 6", gap)
	}
	if gap := ticks[len(ticks)-1] - ticks[len(ticks)-2]; gap != 1 {
		t.Errorf("last gap = %d, want 1", gap)
	}
	// The time only moves forward through the night: 17:58 ... 23:59, 00:00 ... 03:17.
	prev := clockStartMin
	wrapped := false
	for f := clockFrom; f <= clockEnd; f++ {
		h, m := ClockAt(f)
		now := h*60 + m
		if now < prev {
			if wrapped {
				t.Fatalf("the clock went back on frame %d", f)
			}
			wrapped = true
		}
		prev = now
	}
	if !wrapped {
		t.Error("the clock should pass midnight")
	}
}

func TestGlitchRows(t *testing.T) {
	rows := GlitchRows(0, 6, 3)
	if len(rows) != 34 { // 270 / 8, rounded up
		t.Fatalf("%d rows, want 34", len(rows))
	}
	again := GlitchRows(0, 6, 3)
	for i := range rows {
		if rows[i] != again[i] {
			t.Fatal("the same step and seed must give the same rows")
		}
	}
	count := func(step int) int {
		n := 0
		for _, r := range GlitchRows(step, 6, 3) {
			if r.FromB {
				n++
			}
			if r.Shift%8 != 0 || r.Shift < -16 || r.Shift > 16 {
				t.Errorf("shift %d: want 0, 8 or 16 either way", r.Shift)
			}
		}
		return n
	}
	if count(5) <= count(0) {
		t.Errorf("the new shot should take over: %d rows of it at step 0, %d at step 5", count(0), count(5))
	}
}
