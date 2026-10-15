package main

// The clock digits are Silkscreen, the game's own font, drawn by the
// compositor at x4 (each pixel of the 8 px font becomes a 4x4 block of game
// pixels). The font is rasterised once at its native size, without smoothing:
// Silkscreen sits on an 8 pixel grid, so every glyph pixel is either on or off.

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// FontPx is the font's native size in pixels.
const FontPx = 8

// Glyphs rasterises text with Silkscreen.
type Glyphs struct {
	mu    sync.Mutex // a font face is not safe to share between goroutines
	face  font.Face
	cache map[string]Bitmap
}

// LoadGlyphs reads the font file (game/fonts/Silkscreen-Regular.ttf).
func LoadGlyphs(path string) (*Glyphs, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	f, err := opentype.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	face, err := opentype.NewFace(f, &opentype.FaceOptions{Size: FontPx, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		return nil, err
	}
	return &Glyphs{face: face, cache: map[string]Bitmap{}}, nil
}

// Bitmap is the lit pixels of a string at the font's native size.
type Bitmap struct {
	W, H int
	On   []bool // W*H
}

// Render rasterises s. The bitmap is cropped to its ink: row 0 is the top of
// the tallest letter.
func (g *Glyphs) Render(s string) Bitmap {
	g.mu.Lock()
	defer g.mu.Unlock()
	if b, ok := g.cache[s]; ok {
		return b
	}
	b := g.render(s)
	g.cache[s] = b
	return b
}

func (g *Glyphs) render(s string) Bitmap {
	const margin = 2
	adv := font.MeasureString(g.face, s).Ceil()
	h := FontPx + 2*margin
	mask := image.NewAlpha(image.Rect(0, 0, adv+2*margin, h))
	d := font.Drawer{Dst: mask, Src: image.NewUniform(color.Alpha{A: 255}), Face: g.face,
		Dot: fixed.P(margin, margin+g.face.Metrics().Ascent.Ceil())}
	d.DrawString(s)
	b := Bitmap{W: mask.Bounds().Dx(), H: h, On: make([]bool, mask.Bounds().Dx()*h)}
	minY, maxY := h, -1
	for y := 0; y < h; y++ {
		for x := 0; x < b.W; x++ {
			if mask.AlphaAt(x, y).A >= 128 {
				b.On[y*b.W+x] = true
				minY, maxY = min(minY, y), max(maxY, y)
			}
		}
	}
	if maxY < 0 {
		return Bitmap{}
	}
	return b.crop(minY, maxY+1)
}

// crop keeps the rows from y0 up to y1 and, of those, the columns that have ink.
func (b Bitmap) crop(y0, y1 int) Bitmap {
	minX, maxX := b.W, -1
	for y := y0; y < y1; y++ {
		for x := 0; x < b.W; x++ {
			if b.On[y*b.W+x] {
				minX, maxX = min(minX, x), max(maxX, x)
			}
		}
	}
	out := Bitmap{W: maxX - minX + 1, H: y1 - y0}
	out.On = make([]bool, out.W*out.H)
	for y := 0; y < out.H; y++ {
		for x := 0; x < out.W; x++ {
			out.On[y*out.W+x] = b.On[(y0+y)*b.W+minX+x]
		}
	}
	return out
}

// DrawBitmap puts the bitmap on dst with its top-left corner at game pixel
// (x, y), every font pixel a scale x scale block of game pixels.
func DrawBitmap(dst *image.RGBA, b Bitmap, x, y, scale int, c color.RGBA) {
	for j := 0; j < b.H; j++ {
		for i := 0; i < b.W; i++ {
			if !b.On[j*b.W+i] {
				continue
			}
			for dj := 0; dj < scale; dj++ {
				for di := 0; di < scale; di++ {
					putBlock(dst, x+i*scale+di, y+j*scale+dj, c)
				}
			}
		}
	}
}
