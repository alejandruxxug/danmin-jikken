package game

import (
	"testing"

	"danmingame/night"
)

func TestTiltTopYIsEasedWholePixels(t *testing.T) {
	if tiltTopY(0) != revealTagY || tiltTopY(tiltTicks) != revealFaceY {
		t.Fatalf("the tilt goes %d -> %d, want %d -> %d", tiltTopY(0), tiltTopY(tiltTicks), revealTagY, revealFaceY)
	}
	prev := tiltTopY(0)
	step, biggest := 0, 0
	for tick := 1; tick <= tiltTicks; tick++ {
		y := tiltTopY(tick)
		if y > prev {
			t.Fatalf("tick %d: the view goes back down (%d after %d)", tick, y, prev)
		}
		biggest = max(biggest, prev-y)
		if tick == tiltTicks/2 {
			step = y
		}
		prev = y
	}
	// Ease in and out: slow at both ends, the middle is the halfway point.
	if tiltTopY(6) < revealTagY-1 || tiltTopY(tiltTicks-6) > revealFaceY+1 {
		t.Errorf("the ends are not slow: y(6)=%d y(end-6)=%d", tiltTopY(6), tiltTopY(tiltTicks-6))
	}
	if want := (revealTagY + revealFaceY) / 2; step != want {
		t.Errorf("halfway y = %d, want %d", step, want)
	}
	if biggest > 2 {
		t.Errorf("one tick moves the view %d px, want at most 2 (a smooth tilt)", biggest)
	}
	// The storyboard's frames: top y about 104 at 0.9 s and 86 at 1.5 s.
	if y := tiltTopY(54); y < 104 || y > 106 {
		t.Errorf("y at 0.9 s = %d, want about 104", y)
	}
	if y := tiltTopY(90); y < 85 || y > 87 {
		t.Errorf("y at 1.5 s = %d, want about 86", y)
	}
}

func TestRevealStageMachine(t *testing.T) {
	r := Reveal{Stage: RevealArrival}
	if r.Active() != true || r.InputLocked() {
		t.Fatal("the arrival shows the view with input open (to look down)")
	}
	r.LookDown()
	if r.Stage != RevealDesk {
		t.Fatalf("looking down gave %v", r.Stage)
	}
	r.LookUp(false) // not stamped: the tag shot again, input open
	if r.Stage != RevealArrival {
		t.Fatalf("looking up unstamped gave %v", r.Stage)
	}
	r.LookDown()
	r.LookUp(true)
	if r.Stage != RevealBack || !r.InputLocked() || r.TopY() != revealTagY {
		t.Fatalf("looking up stamped gave %v, locked %v, top %d", r.Stage, r.InputLocked(), r.TopY())
	}
	want := []struct {
		stage RevealStage
		ticks int
	}{{RevealBack, backHoldTicks}, {RevealTilt, tiltTicks}, {RevealFace, faceHoldTicks}}
	blinks := 0
	for _, w := range want {
		if r.Stage != w.stage {
			t.Fatalf("stage %v, want %v", r.Stage, w.stage)
		}
		for range w.ticks {
			if r.Blinking() {
				blinks++
			}
			r.step()
		}
	}
	if r.Stage != RevealStun || r.TopY() != revealFaceY {
		t.Errorf("after the face hold: %v at %d, want stun at the face", r.Stage, r.TopY())
	}
	if blinks != blinkTicks {
		t.Errorf("the blink lasted %d ticks, want %d (150 ms)", blinks, blinkTicks)
	}
	// 1.0 + 2.4 + 1.5 s = 4.9 s.
	if total := backHoldTicks + tiltTicks + faceHoldTicks; total != 294 {
		t.Errorf("the reveal takes %d ticks, want 294 (4.9 s)", total)
	}
}

func TestYouIsTheWindowPatientInAGown(t *testing.T) {
	d3, err := RulesForDay(3)
	if err != nil {
		t.Fatal(err)
	}
	you := d3.Visitors[len(d3.Visitors)-1]
	if !you.Self || you.Sprite != "you_patient" || you.FaceY != 11 {
		t.Errorf("YOU = sprite %q, FaceY %d, self %v; want you_patient, 11, true", you.Sprite, you.FaceY, you.Self)
	}
	cecilies := 0
	for day := 1; day <= 3; day++ {
		rules, _ := RulesForDay(day)
		for _, v := range rules.Visitors {
			if v.Sprite == "cecily" {
				cecilies++
				if v.FaceY != 5 {
					t.Errorf("day %d: Cecily's FaceY = %d, want 5", day, v.FaceY)
				}
			}
		}
	}
	if cecilies != 2 {
		t.Errorf("found %d Cecilys, want 2 (day 1 and day 3)", cecilies)
	}
}

// Day 3 played as a person does it: the strobe waits for the reveal, the
// reveal runs its stages after the stamp, and the strobe takes over from the
// face and hands over to night 3 with stamped_self set.
func TestDay3GoesThroughTheRevealIntoNight3(t *testing.T) {
	sh := newTestShift(t, 3)
	sh.Flags = night.Flags{}
	env := &Env{Day: 3, Flags: sh.Flags}
	seen := map[RevealStage]bool{}
	stamped, handOver := false, false
	for tick := 0; tick < 60*200 && !handOver; tick++ {
		sh.Update()
		env.UpdateReveal(sh)
		r := &env.Reveal
		seen[r.Stage] = true
		if r.Stage != RevealOff && r.Stage != RevealStun && env.StrobeNight != 0 {
			t.Fatalf("the strobe started in the %v stage", r.Stage)
		}
		if g := sh.window(); g != nil && sh.atWindow(g) && !stamped {
			switch {
			case !g.Self:
				_ = sh.Stamp(StampRefer)
			case r.Stage == RevealArrival:
				r.LookDown() // the player opens the desk
				seen[r.Stage] = true
				if err := sh.Stamp(StampAdmit); err != nil {
					t.Fatal(err)
				}
				stamped = true
				r.LookUp(true) // and looks back up
			}
		}
		handOver = env.UpdateLiveStrobe(sh)
	}
	for _, st := range []RevealStage{RevealArrival, RevealDesk, RevealBack, RevealTilt, RevealFace, RevealStun} {
		if !seen[st] {
			t.Errorf("the reveal never reached %v", st)
		}
	}
	if !handOver || env.StrobeNight != 3 {
		t.Fatalf("hand over %v, strobe night %d: want the strobe into night 3", handOver, env.StrobeNight)
	}
	if !sh.Flags.Has("stamped_self") {
		t.Error("stamped_self not set")
	}
}

// The shift runs out with the folder never opened: the reveal still plays
// and the strobe still follows it.
func TestRevealStillPlaysWhenTheShiftRunsOut(t *testing.T) {
	sh := newTestShift(t, 3)
	env := &Env{Day: 3, Flags: night.Flags{}}
	sh.Flags = env.Flags
	handOver := false
	for tick := 0; tick < 60*400 && !handOver; tick++ {
		sh.Update()
		env.UpdateReveal(sh)
		if g := sh.window(); g != nil && sh.atWindow(g) && !g.Self {
			_ = sh.Stamp(StampRefer)
		}
		if env.Reveal.Stage == RevealArrival && !sh.Over() {
			sh.ticks += 60 // let the clock run to the end of the shift
		}
		handOver = env.UpdateLiveStrobe(sh)
	}
	if !handOver || env.Reveal.Stage != RevealStun {
		t.Fatalf("hand over %v at stage %v: want the strobe after the reveal", handOver, env.Reveal.Stage)
	}
	if sh.Flags.Has("stamped_self") {
		t.Error("stamped_self set without a stamp")
	}
}
