package main

import (
	"image"
	"image/color"
)

// FadeFactor is how much of the picture (and the sound) is left at a position
// in a shot of total frames, with a fade in and a fade out of the given
// lengths. pos counts frames from the start of the shot and may be fractional
// (the sound uses the position between two frames): 0 is the first frame, so
// a fade in starts at black and is whole after fadeIn frames; a fade out
// reaches black at the last frame.
func FadeFactor(pos float64, total, fadeIn, fadeOut int) float64 {
	f := 1.0
	if fadeIn > 0 && pos < float64(fadeIn) {
		f = min(f, pos/float64(fadeIn))
	}
	if left := float64(total-1) - pos; fadeOut > 0 && left < float64(fadeOut) {
		f = min(f, left/float64(fadeOut))
	}
	return max(f, 0)
}

// Darken multiplies the picture's colour by f (0 = black, 1 = unchanged).
func Darken(src image.Image, f float64) *image.RGBA {
	b := src.Bounds()
	out := image.NewRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := color.RGBAModel.Convert(src.At(x, y)).(color.RGBA)
			out.SetRGBA(x, y, color.RGBA{uint8(float64(c.R)*f + 0.5), uint8(float64(c.G)*f + 0.5), uint8(float64(c.B)*f + 0.5), 255})
		}
	}
	return out
}

// FadeAudio applies the shot's fades to its sound, per sample.
func FadeAudio(a Audio, total, fadeIn, fadeOut int) {
	for i := 0; i < a.Frames(); i++ {
		f := float32(FadeFactor(float64(i)/SamplesPerFrame, total, fadeIn, fadeOut))
		if f < 1 {
			a.Samples[2*i] *= f
			a.Samples[2*i+1] *= f
		}
	}
}

// endTaper is the tiny fade at the end of a shot's sound (a few ms) that keeps
// a cut from clicking; a shot marked "cut" does not get it.
const endTaper = 480 // samples, 10 ms

func TaperEnd(a Audio) {
	n := min(endTaper, a.Frames())
	for k := 0; k < n; k++ {
		i := a.Frames() - 1 - k
		f := float32(k) / float32(n)
		a.Samples[2*i] *= f
		a.Samples[2*i+1] *= f
	}
}
