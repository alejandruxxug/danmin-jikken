package main

import (
	"math"
	"testing"

	"danmingame"
)

func TestIronboundDropIsOnTheSlam(t *testing.T) {
	if got := IronboundDropFrame(); got != slamFrame {
		t.Errorf("the drop is on frame %d, the dagger lands on %d", got, slamFrame)
	}
	// 24.7 s into the file at frame 2661, and the file moves at normal speed: 34.7 s at the slam.
	secs := ironFileStart + float64(slamFrame-ironFrom)/FPS
	if math.Abs(secs-ironDropFile) > 1e-9 {
		t.Errorf("the file is at %v s on the slam frame, want %v", secs, ironDropFile)
	}
	var iron Bed
	for _, b := range BuildAudioPlan(false).Beds {
		if b.File == "ending_rock_forever.mp3" {
			iron = b
		}
	}
	if iron.Rate != nil {
		t.Error("Ironbound must play at normal speed or its drop would drift off the slam")
	}
	if g := iron.Gain(float64(ironFrom)); g < 0.05 || g > 0.12 {
		t.Errorf("Ironbound starts quiet, got %v", g)
	}
	if iron.Gain(float64(slamFrame)) < 0.8 || iron.Gain(float64(ironFrom)) >= iron.Gain(float64(slamFrame-100)) {
		t.Error("Ironbound rises to the drop")
	}
}

func TestMixTrailerHardCuts(t *testing.T) {
	load := func(string) (Audio, error) { return constClip(SampleRate, 0.5), nil } // one second of 0.5
	plan := AudioPlan{
		Beds: []Bed{{File: "m", From: 10, To: 40, Gain: func(float64) float64 { return 1 }}},
		Cues: []Cue{{File: "c", Frame: 5, Gain: 1, Until: 20}},
	}
	a, err := MixTrailer(plan, 100, load, map[string]game.AudioLog{})
	if err != nil {
		t.Fatal(err)
	}
	at := func(frame int) float32 { return a.Samples[2*frame*SamplesPerFrame+2] }
	if at(4) != 0 || at(6) != 0.5 || at(15) != 1 { // the cue alone, then the cue and the bed
		t.Errorf("levels %v %v %v", at(4), at(6), at(15))
	}
	if got := at(25); got != 0.5 { // the cue is cut on frame 20, the bed plays on
		t.Errorf("after the cue's cut only the bed sounds: %v", got)
	}
	if got := at(45); got != 0 {
		t.Errorf("after the bed's cut there is silence: %v", got)
	}
}

func TestTapeStopSlowsTheBed(t *testing.T) {
	// A bed on a ramp (the sample value is its own position) plays through the tape
	// stop: how far the file has got by the end follows the speed curve.
	n := 40 * SampleRate
	clip := Audio{Samples: make([]float32, 2*n)}
	for i := 0; i < n; i++ {
		clip.Samples[2*i] = float32(i) / float32(n)
	}
	b := Bed{File: "x", From: tapeFrom - 60, To: tapeTo, Rate: TapeRate, Gain: func(float64) float64 { return 1 }}
	out := b.Render(clip, tapeTo+10)
	end := (tapeTo - 1) * SamplesPerFrame
	got := float64(out.Samples[2*end]) * float64(n) / SampleRate // seconds into the file
	// 1 s at full speed, then the integral of 1 - 0.7 t^2 over 250 frames.
	want := 1 + float64(tapeTo-tapeFrom)/FPS*(1-0.7/3)
	if math.Abs(got-want) > 0.02 {
		t.Errorf("the file has played %.3f s, want %.3f s", got, want)
	}
}
