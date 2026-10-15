package main

import (
	"math"
	"testing"
)

// sine is a clip of a sine wave, 1000 Hz at 48 kHz, one second long.
func sine() Audio {
	a := Audio{Samples: make([]float32, 2*SampleRate)}
	for i := 0; i < SampleRate; i++ {
		v := float32(math.Sin(2 * math.Pi * 1000 * float64(i) / SampleRate))
		a.Samples[2*i], a.Samples[2*i+1] = v, v
	}
	return a
}

// crossings counts the rising zero crossings of the left channel: the number of cycles.
func crossings(a Audio) int {
	n := 0
	for i := 1; i < a.Frames(); i++ {
		if a.Samples[2*(i-1)] < 0 && a.Samples[2*i] >= 0 {
			n++
		}
	}
	return n
}

func TestVarispeedPitch(t *testing.T) {
	clip := sine()
	same := Varispeed(clip, 0, 24000, func(int) float64 { return 1 })
	if got := crossings(same); got < 498 || got > 502 { // half a second of 1000 Hz
		t.Errorf("speed 1: %d cycles in 0.5 s, want 500", got)
	}
	slow := Varispeed(clip, 0, 24000, func(int) float64 { return 0.5 })
	if got := crossings(slow); got < 248 || got > 252 { // an octave down
		t.Errorf("speed 0.5: %d cycles, want 250", got)
	}
}

func TestVarispeedStartAndEnd(t *testing.T) {
	clip := Audio{Samples: []float32{0, 0, 1, 1, 2, 2, 3, 3}}
	got := Varispeed(clip, 1.5, 5, func(int) float64 { return 1 })
	want := []float32{1.5, 2.5, 0, 0, 0} // reads between samples, silent past the end
	for i, w := range want {
		if math.Abs(float64(got.Samples[2*i]-w)) > 1e-6 {
			t.Errorf("sample %d = %v, want %v", i, got.Samples[2*i], w)
		}
	}
}

func TestTapeRate(t *testing.T) {
	if TapeRate(0) != 1 || TapeRate(tapeFrom) != 1 {
		t.Error("the tape runs at full speed until the stop begins")
	}
	if got := TapeRate(tapeTo); math.Abs(got-tapeSpeed) > 1e-9 {
		t.Errorf("the speed at the end is %v, want %v", got, tapeSpeed)
	}
	if got := TapeRate(tapeTo + 100); math.Abs(got-tapeSpeed) > 1e-12 {
		t.Errorf("after the stop the speed stays at %v, got %v", tapeSpeed, got)
	}
	prev := 1.0
	for f := float64(tapeFrom); f <= tapeTo; f += 5 {
		r := TapeRate(f)
		if r > prev+1e-12 {
			t.Fatalf("the speed went up at frame %v", f)
		}
		prev = r
	}
	// Half way through, inQuad has only lost a quarter of the way down.
	mid := TapeRate((tapeFrom + tapeTo) / 2)
	if want := 1 - (1-tapeSpeed)*0.25; math.Abs(mid-want) > 1e-9 {
		t.Errorf("mid-way speed %v, want %v", mid, want)
	}
}

func TestTaperEnd(t *testing.T) {
	a := constClip(1000, 1)
	TaperEnd(a, 100)
	if a.Samples[0] != 1 || a.Samples[2*999] != 0 || a.Samples[2*950] >= 1 {
		t.Error("the last samples should fall to 0 and the early ones stay")
	}
}
