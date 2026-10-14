package sprites

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type shadowKey struct {
	layer   string
	frame   int
	r, g, b uint8
}

// Shadow is a flat copy of a layer's frame: every pixel that has any colour
// is painted c, the rest stays clear. The desk draws it a pixel or two off
// to the side under a paper or a folder, so it casts the drop shadow without
// any extra art (Art Direction: the game draws the shadows). The copy is made
// once and kept.
func (s *Sheet) Shadow(layer string, frame int, c color.RGBA) (*ebiten.Image, error) {
	key := shadowKey{layer, frame, c.R, c.G, c.B}
	if img, ok := s.shadows[key]; ok {
		return img, nil
	}
	r, err := s.Rect(layer, frame)
	if err != nil {
		return nil, err
	}
	out := image.NewRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	for y := 0; y < r.Dy(); y++ {
		for x := 0; x < r.Dx(); x++ {
			if _, _, _, a := s.pic.At(r.Min.X+x, r.Min.Y+y).RGBA(); a > 0 {
				out.SetRGBA(x, y, c)
			}
		}
	}
	img := ebiten.NewImageFromImage(out)
	if s.shadows == nil {
		s.shadows = make(map[shadowKey]*ebiten.Image)
	}
	s.shadows[key] = img
	return img, nil
}
