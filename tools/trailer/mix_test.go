package main

import (
	"fmt"
	"math"
	"testing"

	"danmingame/sound"
)

func TestFrameToSample(t *testing.T) {
	for frame, want := range map[int]int{0: 0, 1: 0, 2: 800, 61: 48000, 181: 144000} {
		if got := frameToSample(frame); got != want {
			t.Errorf("frameToSample(%d) = %d, want %d", frame, got, want)
		}
	}
}

// constClip is n sample pairs of the same value.
func constClip(n int, v float32) Audio {
	a := Audio{Samples: make([]float32, 2*n)}
	for i := range a.Samples {
		a.Samples[i] = v
	}
	return a
}

// rampClip counts up: sample i has the value i/1000, so a test can tell which
// sample of the file is playing where.
func rampClip(n int) Audio {
	a := Audio{Samples: make([]float32, 2*n)}
	for i := 0; i < n; i++ {
		a.Samples[2*i], a.Samples[2*i+1] = float32(i)/1000, float32(i)/1000
	}
	return a
}

func fakeLoader(clips map[string]Audio) Loader {
	return func(file string) (Audio, error) {
		if c, ok := clips[file]; ok {
			return c, nil
		}
		return Audio{}, fmt.Errorf("no %s", file)
	}
}

func TestMixPlacesSFX(t *testing.T) {
	load := fakeLoader(map[string]Audio{"hit.wav": constClip(100, 0.5)})
	log := []sound.Event{{Frame: 3, Kind: "sfx", File: "hit.wav", Gain: 0.5}, {Frame: 4, Kind: "sfx", Cue: "none"}}
	out, err := Mix(log, 5, load, MixOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if out.Frames() != 5*SamplesPerFrame {
		t.Fatalf("length = %d samples", out.Frames())
	}
	at := 2 * SamplesPerFrame // frame 3 starts at 1600
	for i, want := range map[int]float32{at - 1: 0, at: 0.25, at + 99: 0.25, at + 100: 0} {
		if got := out.Samples[2*i]; got != want {
			t.Errorf("sample %d = %v, want %v", i, got, want)
		}
	}
}

func TestMixMissingFileIsAnError(t *testing.T) {
	_, err := Mix([]sound.Event{{Frame: 1, Kind: "sfx", File: "gone.wav", Gain: 1}}, 2, fakeLoader(nil), MixOptions{})
	if err == nil {
		t.Error("a file that cannot be loaded should be an error")
	}
}

func TestMixMusicOffsetLoopSeekStop(t *testing.T) {
	load := fakeLoader(map[string]Audio{"bed.wav": rampClip(5000)})
	log := []sound.Event{
		{Frame: 1, Kind: "music", File: "bed.wav", Offset: 1000.0 / SampleRate, Loop: true}, // starts at file sample 1000
		{Frame: 1, Kind: "gain", Gain: 1},
		{Frame: 3, Kind: "seek", Offset: 200.0 / SampleRate}, // at sample 1600 the file jumps to 200
		{Frame: 5, Kind: "stop"},                             // at sample 3200
	}
	out, err := Mix(log, 6, load, MixOptions{})
	if err != nil {
		t.Fatal(err)
	}
	check := func(i int, want float32) {
		t.Helper()
		if got := out.Samples[2*i]; math.Abs(float64(got-want)) > 1e-6 {
			t.Errorf("sample %d = %v, want %v", i, got, want)
		}
	}
	check(0, 1.0)      // file sample 1000
	check(799, 1.799)  // still counting up
	check(1599, 2.599) // just before the seek
	check(1600, 0.2)   // the seek: file sample 200
	check(3199, 1.799) // 200 + 1599
	check(3200, 0)     // stopped
	check(4000, 0)
	// The loop: the file is 5000 long and loops from 1000; run it past the end.
	load = fakeLoader(map[string]Audio{"bed.wav": rampClip(1500)})
	log = []sound.Event{{Frame: 1, Kind: "music", File: "bed.wav", Offset: 1000.0 / SampleRate, Loop: true}}
	if out, err = Mix(log, 2, load, MixOptions{}); err != nil {
		t.Fatal(err)
	}
	check(499, 1.499) // the last sample of the file
	check(500, 1.0)   // back to the loop start
	// Without a loop it ends.
	log[0].Loop = false
	if out, err = Mix(log, 2, load, MixOptions{}); err != nil {
		t.Fatal(err)
	}
	check(500, 0)
}

func TestMixGainEnvelopeAndDropMusic(t *testing.T) {
	load := fakeLoader(map[string]Audio{"bed.wav": constClip(48000, 1), "hit.wav": constClip(10, 1)})
	log := []sound.Event{
		{Frame: 1, Kind: "music", File: "bed.wav"},
		{Frame: 1, Kind: "gain", Gain: 0.5}, // the starting gain, no ramp
		{Frame: 2, Kind: "sfx", File: "hit.wav", Gain: 1},
		{Frame: 3, Kind: "gain", Gain: 0}, // from sample 1600, reached over one frame
	}
	out, err := Mix(log, 4, load, MixOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for i, want := range map[int]float64{0: 0.5, 1599: 0.5, 1600: 0.5, 2000: 0.25, 2400: 0, 3000: 0} {
		if got := float64(out.Samples[2*i]); math.Abs(got-want) > 1e-3 {
			t.Errorf("sample %d = %v, want %v", i, got, want)
		}
	}
	if got := out.Samples[2*SamplesPerFrame]; math.Abs(float64(got)-1.5) > 1e-6 {
		t.Errorf("music and sfx should add: %v", got)
	}
	dropped, err := Mix(log, 4, load, MixOptions{DropMusic: true})
	if err != nil {
		t.Fatal(err)
	}
	if got := dropped.Samples[2*SamplesPerFrame]; got != 1 {
		t.Errorf("DropMusic should leave the sfx only: %v", got)
	}
	if got := dropped.Samples[0]; got != 0 {
		t.Errorf("DropMusic should leave silence where only music played: %v", got)
	}
}

func TestSoftLimit(t *testing.T) {
	if got := SoftLimit(0.5); got != 0.5 {
		t.Errorf("below the knee nothing changes: %v", got)
	}
	for _, x := range []float32{0.9, 1, 3, 100, -3} {
		if got := SoftLimit(x); math.Abs(float64(got)) > 1 {
			t.Errorf("SoftLimit(%v) = %v passes 1", x, got)
		}
	}
	if SoftLimit(2) <= SoftLimit(0.9) || SoftLimit(-2) != -SoftLimit(2) {
		t.Error("the limiter should rise with the input and be symmetric")
	}
}
