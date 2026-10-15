package main

// Playing a sound at a changing speed. A tape stopping is a playback speed that
// falls to a crawl: the sound slows down and drops in pitch together, which is
// what resampling does. The speed curve is a plain function; Varispeed walks
// through the file by that speed and reads between samples.

import "math"

// Varispeed returns n sample pairs read from clip, starting at sample position
// start. Output sample i advances the position by rate(i) (1 = normal speed,
// 0.5 = an octave down at half speed). Reading between two samples is linear.
// Past the end of the clip it is silent.
func Varispeed(clip Audio, start float64, n int, rate func(i int) float64) Audio {
	out := Audio{Samples: make([]float32, 2*n)}
	pos := start
	for i := 0; i < n; i++ {
		k := int(math.Floor(pos))
		if k >= 0 && k+1 < clip.Frames() {
			frac := float32(pos - float64(k))
			for c := 0; c < 2; c++ {
				a, b := clip.Samples[2*k+c], clip.Samples[2*(k+1)+c]
				out.Samples[2*i+c] = a + (b-a)*frac
			}
		}
		pos += rate(i)
	}
	return out
}

// The jingle's tape stop: from frame tapeFrom the speed falls from 1 to
// tapeSpeed by frame tapeTo, slowly at first and faster at the end (inQuad).
const (
	tapeFrom  = t08Start
	tapeTo    = t10Start
	tapeSpeed = 0.3
)

// TapeRate is the jingle's playback speed at trailer frame f (fractions allowed).
func TapeRate(f float64) float64 {
	if f <= tapeFrom {
		return 1
	}
	t := min(1, (f-tapeFrom)/float64(tapeTo-tapeFrom))
	return 1 - (1-tapeSpeed)*inQuad(t)
}

// TaperEnd fades the last n samples of a to nothing: a hard cut that does not click.
func TaperEnd(a Audio, n int) {
	n = min(n, a.Frames())
	for k := 0; k < n; k++ {
		i := a.Frames() - 1 - k
		f := float32(k) / float32(n)
		a.Samples[2*i] *= f
		a.Samples[2*i+1] *= f
	}
}

// cutTaper is the length of the tiny fade at a hard cut: 10 ms.
const cutTaper = 480
