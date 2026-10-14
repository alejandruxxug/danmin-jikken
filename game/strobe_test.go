package game

import "testing"

func TestStrobePhases(t *testing.T) {
	for n := 1; n <= 3; n++ {
		strobe := strobeTicks[n]
		// Flicker: the day, with single inverted frames.
		if st := StrobeAt(n, 0, false); st.Layer != layerDay || st.Invert != 0 {
			t.Errorf("night %d tick 0 = %+v, want the plain day", n, st)
		}
		flips := 0
		for tick := 0; tick < flickerTicks; tick++ {
			if StrobeAt(n, tick, false).Invert == 1 {
				flips++
			}
		}
		if flips < 1 || flips > 2 {
			t.Errorf("night %d: %d inverted frames in the flicker, want one or two", n, flips)
		}
		// Strobe: both normal and inverted frames, and at least one night frame.
		var normal, inverted, nightFrames int
		for tick := flickerTicks; tick < flickerTicks+strobe; tick++ {
			st := StrobeAt(n, tick, false)
			if st.Invert == 1 {
				inverted++
			} else {
				normal++
			}
			if st.Layer == layerNight {
				nightFrames++
			}
		}
		if normal == 0 || inverted == 0 || nightFrames == 0 {
			t.Errorf("night %d strobe: %d normal, %d inverted, %d night frames", n, normal, inverted, nightFrames)
		}
		// White, then the fall that ends at progress 1.
		if st := StrobeAt(n, flickerTicks+strobe, false); st.Layer != layerWhite {
			t.Errorf("night %d: after the strobe comes %v, want white", n, st.Layer)
		}
		end := StrobeLength(n)
		if st := StrobeAt(n, end-1, false); st.Layer != layerFall || st.Fall < 0.99 {
			t.Errorf("night %d last tick = %+v, want the end of the fall", n, st)
		}
	}
}

func TestStrobeGetsLongerEachNight(t *testing.T) {
	// Art Direction: ~0.6 s, ~0.8 s, ~1 s of full strobe.
	want := map[int]int{1: 36, 2: 48, 3: 60}
	for n, ticks := range want {
		if strobeTicks[n] != ticks {
			t.Errorf("night %d strobe is %d ticks, want %d", n, strobeTicks[n], ticks)
		}
	}
	if !(StrobeLength(1) < StrobeLength(2) && StrobeLength(2) < StrobeLength(3)) {
		t.Error("each night's strobe must be longer than the one before")
	}
}

func TestReduceFlashingFadesOnce(t *testing.T) {
	// Same timing, but no flip: the invert only ever grows.
	prev := -1.0
	for tick := 0; tick < flickerTicks+strobeTicks[1]; tick++ {
		st := StrobeAt(1, tick, true)
		if st.Layer != layerDay || st.Invert < prev {
			t.Fatalf("tick %d: %+v after invert %v: want one slow rise on the day frame", tick, st, prev)
		}
		prev = st.Invert
	}
	if prev < 0.9 {
		t.Errorf("the fade ends at %v, want almost fully inverted", prev)
	}
	if StrobeAt(1, flickerTicks+strobeTicks[1], true).Layer != layerWhite {
		t.Error("the white comes at the same time as in the full strobe")
	}
}

func TestStrobeFlashRate(t *testing.T) {
	// The strobe may not flip faster than every 3 ticks (10 flips a second).
	last, run := StrobeState{}, 0
	for tick := flickerTicks; tick < flickerTicks+strobeTicks[3]; tick++ {
		st := StrobeAt(3, tick, false)
		if st != last {
			if run != 0 && run < ticksPerStrobeFrame {
				t.Fatalf("a frame lasted %d ticks, want at least %d", run, ticksPerStrobeFrame)
			}
			last, run = st, 0
		}
		run++
	}
}
