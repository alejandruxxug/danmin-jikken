package sprites

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// Remapped returns a copy of the sheet in which every pixel whose colour is
// a key of swap gets the colour it maps to. The frames and layers are the
// same. The handwriting sheets are black ink for paper; the name plate is
// black, so the game draws them from a copy with the ink swapped for
// off-white. Both colours come from the palette.
func (s *Sheet) Remapped(swap map[color.RGBA]color.RGBA) *Sheet {
	b := s.pic.Bounds()
	out := image.NewRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := color.RGBAModel.Convert(s.pic.At(x, y)).(color.RGBA)
			if to, ok := swap[c]; ok {
				c = to
			}
			out.SetRGBA(x, y, c)
		}
	}
	return &Sheet{Meta: s.Meta, img: ebiten.NewImageFromImage(out), pic: out}
}
