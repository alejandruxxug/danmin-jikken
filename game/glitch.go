package game

import (
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"io/fs"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"danmingame/sprites"
)

// The secret ending lags out (Game Design -> design log, "Ending cards"): the
// picture stutters and freezes, rows tear sideways, blocks smear, and the
// glitch pieces of the fx_glitch sheet fly around, more and more until it is
// a mess. Everything here is a pure function of the tick and the seed, so a
// screenshot taken at the same tick always looks the same.
//
// -reduce-flashing keeps the lag, the tearing and the smears, and drops the
// hard flicker, the white flashes and the inverted strips.

//go:embed invert.kage
var invertSource []byte

const (
	glitchRamp   = 420 // ticks until the glitch is at full strength
	glitchWindow = 24  // the lag is decided once per window of ticks
)

// glitchLevel is how strong the glitch is, 0 to 1, ticks after it began.
// It starts at 0.2 (the lag is already there) and climbs to 1.
func glitchLevel(ticks int) float64 {
	return min(1, 0.2+0.8*float64(ticks)/glitchRamp)
}

// lagTick is the tick of the picture that is really shown: the game "lags".
// Per window of 24 ticks a coin decides whether it freezes at the start of
// the window for a few ticks and then jumps forward (the dropped frames).
func lagTick(ticks int, seed int64, level float64) int {
	w, offset := ticks/glitchWindow, ticks%glitchWindow
	rng := rand.New(rand.NewSource(seed + int64(w)*7919))
	if rng.Float64() >= 0.3+0.4*level {
		return ticks
	}
	freeze := 4 + rng.Intn(14)
	if offset < freeze {
		return w * glitchWindow
	}
	return ticks
}

// pieceLayers are the layers of the fx_glitch sheet.
var pieceLayers = []string{"bands", "blocks", "scanlines", "glyphs", "noise"}

// Glitch draws the secret ending's glitch on top of a finished picture.
type Glitch struct {
	pieces *sprites.Sheet // fx_glitch
	invert *ebiten.Shader
	seed   int64
}

// NewGlitch loads the pieces sheet and compiles the invert shader.
func NewGlitch(fsys fs.FS, seed int64) (*Glitch, error) {
	pieces, err := LoadSheet(fsys, "fx_glitch")
	if err != nil {
		return nil, err
	}
	invert, err := ebiten.NewShader(invertSource)
	if err != nil {
		return nil, fmt.Errorf("compiling the invert shader: %w", err)
	}
	return &Glitch{pieces: pieces, invert: invert, seed: seed}, nil
}

// Apply draws src (the finished card) onto dst, glitched. level is 0 to 1
// (glitchLevel). The same ticks always give the same picture; the effects
// change every second tick so the glitch reads as frames, not as a blur.
func (g *Glitch) Apply(dst, src *ebiten.Image, ticks int, level float64, reduceFlashing bool) {
	rng := rand.New(rand.NewSource(g.seed + int64(ticks/2)*104729))
	dst.Fill(paletteInk)
	g.tear(dst, src, rng, level)
	g.smear(dst, src, rng, level)
	g.drawPieces(dst, rng, level)
	if reduceFlashing {
		return
	}
	g.invertStrips(dst, src, rng, level)
	switch roll := rng.Float64(); {
	case roll < 0.04*level: // a whole frame flipped
		g.invertRect(dst, src, image.Rect(0, 0, ScreenW, ScreenH))
	case roll < 0.04*level+0.1*level: // a white flash frame
		vector.FillRect(dst, 0, 0, ScreenW, ScreenH, color.NRGBA{0xec, 0xec, 0xe6, 0x60}, false)
	}
}

// tear draws src in horizontal strips; some strips slide sideways.
func (g *Glitch) tear(dst, src *ebiten.Image, rng *rand.Rand, level float64) {
	for y := 0; y < ScreenH; {
		h := 1 + rng.Intn(14)
		h = min(h, ScreenH-y)
		dx := 0
		if rng.Float64() < 0.55*level {
			dx = rng.Intn(int(10+60*level)*2+1) - int(10+60*level)
		}
		strip := src.SubImage(image.Rect(0, y, ScreenW, y+h)).(*ebiten.Image)
		var op ebiten.DrawImageOptions
		op.GeoM.Translate(float64(dx), float64(y))
		dst.DrawImage(strip, &op)
		// What slid off one side comes back on the other.
		if dx != 0 {
			op.GeoM.Reset()
			wrap := ScreenW
			if dx > 0 {
				wrap = -ScreenW
			}
			op.GeoM.Translate(float64(dx+wrap), float64(y))
			dst.DrawImage(strip, &op)
		}
		y += h
	}
}

// smear is the data-mosh: blocks of the picture are pasted again and again
// along a line, like a video that lost its key frame.
func (g *Glitch) smear(dst, src *ebiten.Image, rng *rand.Rand, level float64) {
	for i := 0; i < int(1+level*5); i++ {
		w, h := 30+rng.Intn(100), 8+rng.Intn(36)
		x, y := rng.Intn(ScreenW-w), rng.Intn(ScreenH-h)
		block := src.SubImage(image.Rect(x, y, x+w, y+h)).(*ebiten.Image)
		stepX, stepY := rng.Intn(9)-4, rng.Intn(5)
		for k := 1; k <= 3+rng.Intn(5); k++ {
			var op ebiten.DrawImageOptions
			op.GeoM.Translate(float64(x+k*stepX*3), float64(y+k*stepY*2))
			dst.DrawImage(block, &op)
		}
	}
}

// drawPieces throws the fx_glitch layers on top at random offsets.
func (g *Glitch) drawPieces(dst *ebiten.Image, rng *rand.Rand, level float64) {
	for _, layer := range pieceLayers {
		if rng.Float64() > 0.25+0.7*level {
			continue
		}
		frame, err := g.pieces.Frame(layer, 0)
		if err != nil {
			continue // a layer missing from the sheet is simply skipped
		}
		var op ebiten.DrawImageOptions
		op.GeoM.Translate(float64(rng.Intn(61)-30), float64(rng.Intn(41)-20))
		op.ColorScale.ScaleAlpha(float32(0.5 + 0.5*rng.Float64()))
		dst.DrawImage(frame, &op)
	}
}

// invertStrips flips a few strips of the picture to their opposite colours.
func (g *Glitch) invertStrips(dst, src *ebiten.Image, rng *rand.Rand, level float64) {
	for i := 0; i < int(level*4); i++ {
		if rng.Float64() > 0.6 {
			continue
		}
		y, h := rng.Intn(ScreenH-4), 3+rng.Intn(18)
		h = min(h, ScreenH-y)
		g.invertRect(dst, src, image.Rect(0, y, ScreenW, y+h))
	}
}

// invertRect draws the part of src inside r, inverted, onto dst.
func (g *Glitch) invertRect(dst, src *ebiten.Image, r image.Rectangle) {
	op := &ebiten.DrawRectShaderOptions{}
	op.Images[0] = src.SubImage(r).(*ebiten.Image)
	op.GeoM.Translate(float64(r.Min.X), float64(r.Min.Y))
	dst.DrawRectShader(r.Dx(), r.Dy(), g.invert, op)
}
