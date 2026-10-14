package game

import (
	"testing"

	"danmingame/night"
)

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
		// Strobe: normal and inverted day frames only. The night must not leak.
		var normal, inverted int
		for tick := flickerTicks; tick < flickerTicks+strobe; tick++ {
			st := StrobeAt(n, tick, false)
			if st.Layer != layerDay {
				t.Errorf("night %d tick %d shows %v, want only the day", n, tick, st.Layer)
			}
			if st.Invert == 1 {
				inverted++
			} else {
				normal++
			}
		}
		if normal == 0 || inverted == 0 {
			t.Errorf("night %d strobe: %d normal, %d inverted frames", n, normal, inverted)
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

// The leak bug: the strobe must end on solid black, and the night script
// must open on black with no girl and no station.
func TestStrobeEndsOnBlackAndTheNightOpensOnBlack(t *testing.T) {
	for n := 1; n <= 3; n++ {
		end := StrobeAt(n, StrobeLength(n)-1, false)
		if end.Layer != layerFall || fallWhite(end.Fall) > 1.0/255 {
			t.Errorf("night %d ends with layer %v, white %v: want solid black", n, end.Layer, fallWhite(end.Fall))
		}
		if fallWhite(1) != 0 {
			t.Error("the white over the fall must be exactly 0 at its end")
		}
	}
	script, err := night.LoadScript(night.Night1)
	if err != nil {
		t.Fatal(err)
	}
	e, err := night.NewEngine(script, night.Flags{})
	if err != nil {
		t.Fatal(err)
	}
	st := e.Snapshot()
	if st.BG == "station" || st.BG == "ceiling" || st.BG == "white" || st.Visible != "" || len(st.Props) != 0 {
		t.Errorf("night 1 opens on %+v: want black with nobody and no props", st)
	}
}

// The strobe starts while the last guest is at the desk, flickers until the
// stamp lands, then the full strobe takes over and hands over at the white.
func TestLiveStrobeFollowsTheLastGuest(t *testing.T) {
	sh := newTestShift(t, 1)
	env := &Env{Day: 1}
	started := false
	for guest := 0; guest < sh.Total(); guest++ {
		for i := 0; i < 3000 && sh.desk.guest == nil; i++ {
			sh.Update()
			env.UpdateLiveStrobe(sh)
		}
		if sh.desk.guest == nil {
			t.Fatalf("guest %d never reached the desk", guest)
		}
		last := guest == sh.Total()-1
		if last {
			sh.Update()
			env.UpdateLiveStrobe(sh)
			if env.StrobeNight != 1 {
				t.Fatal("the flicker must start when the last guest is at the desk")
			}
			// Still not stamped: the flicker repeats and never reaches the strobe.
			for range 3 * flickerTicks {
				if env.UpdateLiveStrobe(sh) || env.StrobeTick >= flickerTicks {
					t.Fatal("the strobe must wait for the last stamp")
				}
			}
			started = true
		} else if env.StrobeNight != 0 {
			t.Fatalf("the strobe started with %d guests still to come", sh.Total()-guest)
		}
		if err := sh.Stamp(StampAdmit); err != nil {
			t.Fatal(err)
		}
		sh.Update()
	}
	if !started {
		t.Fatal("no last guest")
	}
	handOver := false
	for range StrobeLength(1) {
		if env.UpdateLiveStrobe(sh) {
			handOver = true
			break
		}
	}
	if !handOver || env.StrobeTick < flickerTicks+strobeTicks[1] {
		t.Errorf("hand over %v at tick %d: want it at the white", handOver, env.StrobeTick)
	}
}
