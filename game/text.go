package game

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// The UI font is Silkscreen (SIL Open Font License, see fonts/OFL.txt).
// It is a pixel font drawn on an 8-pixel grid, so it is used at size 8 and
// placed on whole pixels: the glyphs then land exactly on the game's pixels
// and stay crisp at 1x. It has capitals only; lowercase prints as capitals.
//
//go:embed fonts/Silkscreen-Regular.ttf
var silkscreenTTF []byte

const fontSize = 8

// capTop is how many pixels the top of a capital letter sits below the y
// that Font.Draw is given, so the painter can place text by its capitals.
const capTop = 3

// Palette colours the UI draws with (assets/palette/danmin.hex).
var (
	colInk   = color.RGBA{0x0b, 0x0b, 0x0c, 0xff}
	colPaper = color.RGBA{0xec, 0xec, 0xe6, 0xff}
	colDim   = color.RGBA{0x9f, 0xb1, 0x9f, 0xff} // inactive tabs, column headers, faded lines
)

// Font draws Silkscreen text.
type Font struct {
	face *text.GoTextFace
}

// LoadFont reads the embedded Silkscreen font at its native size.
func LoadFont() (*Font, error) { return LoadFontAt(fontSize) }

// LoadFontAt reads it at another size. Silkscreen is a pixel font, so only
// whole multiples of 8 keep it crisp (the ending's thanks line uses 16).
func LoadFontAt(size float64) (*Font, error) {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(silkscreenTTF))
	if err != nil {
		return nil, fmt.Errorf("loading the font: %w", err)
	}
	return &Font{face: &text.GoTextFace{Source: src, Size: size}}, nil
}

// Draw prints s with its top-left corner at (x, y), in game pixels.
func (f *Font) Draw(dst *ebiten.Image, s string, x, y int, c color.Color) {
	var op text.DrawOptions
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(c)
	text.Draw(dst, s, f.face, &op)
}

// Width is how many pixels wide s is, for right-aligning and centring.
func (f *Font) Width(s string) int {
	w, _ := text.Measure(s, f.face, 0)
	return int(w + 0.5)
}

// DrawRight prints s so that it ends at x.
func (f *Font) DrawRight(dst *ebiten.Image, s string, x, y int, c color.Color) {
	f.Draw(dst, s, x-f.Width(s), y, c)
}

// DrawCentered prints s centred on x.
func (f *Font) DrawCentered(dst *ebiten.Image, s string, x, y int, c color.Color) {
	f.Draw(dst, s, x-f.Width(s)/2, y, c)
}
