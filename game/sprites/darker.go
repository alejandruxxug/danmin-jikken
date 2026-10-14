package sprites

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// The day palette's fills, light to dark (assets/palette/danmin.hex). A
// patient who is succumbing to their illness gets darker one step at a
// time (Game Design -> Marking very sick patients): every fill pixel moves
// one place down this ramp, and the darkest stays where it is. No colour
// outside the palette is ever made.
var fillRamp = []color.RGBA{
	{0xec, 0xec, 0xe6, 0xff}, // off-white
	{0xdd, 0xe5, 0xd9, 0xff}, // mint
	{0xc3, 0xd1, 0xc0, 0xff}, // sage
	{0x9f, 0xb1, 0x9f, 0xff}, // deep sage
}

// Darker moves a colour steps places down the fill ramp. A colour that is
// not on the ramp (the black outlines, the eyes) comes back unchanged.
func Darker(c color.RGBA, steps int) color.RGBA {
	for i, f := range fillRamp {
		if c == f {
			return fillRamp[min(i+steps, len(fillRamp)-1)]
		}
	}
	return c
}

// darkLayers are the layers that hold fills. The outlines (lineart) and the
// eyes live on layers of their own, so they stay readable.
var darkLayers = map[string]bool{"fill": true, "detail": true}

type darkKey struct {
	layer        string
	frame, steps int
}

// darkFrame is a layer's frame with every fill pixel steps darker. It is
// made once and kept, so drawing a dark patient costs nothing extra.
func (s *Sheet) darkFrame(layer string, frame, steps int) (*ebiten.Image, error) {
	key := darkKey{layer, frame, steps}
	if img, ok := s.dark[key]; ok {
		return img, nil
	}
	r, err := s.Rect(layer, frame)
	if err != nil {
		return nil, err
	}
	out := image.NewRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	for y := 0; y < r.Dy(); y++ {
		for x := 0; x < r.Dx(); x++ {
			c := color.RGBAModel.Convert(s.pic.At(r.Min.X+x, r.Min.Y+y)).(color.RGBA)
			out.SetRGBA(x, y, Darker(c, steps))
		}
	}
	img := ebiten.NewImageFromImage(out)
	if s.dark == nil {
		s.dark = make(map[darkKey]*ebiten.Image)
	}
	s.dark[key] = img
	return img, nil
}

// DrawDarker is Draw with the fills steps darker (0 is the plain Draw).
func (s *Sheet) DrawDarker(dst *ebiten.Image, layers []string, frame, x, y, steps int) error {
	for _, layer := range layers {
		var img *ebiten.Image
		var err error
		if steps > 0 && darkLayers[layer] {
			img, err = s.darkFrame(layer, frame, steps)
		} else {
			img, err = s.Frame(layer, frame)
		}
		if err != nil {
			return fmt.Errorf("drawing %s: %w", layer, err)
		}
		var op ebiten.DrawImageOptions
		op.GeoM.Translate(float64(x), float64(y))
		op.ColorScale.ScaleAlpha(s.Opacity(layer))
		dst.DrawImage(img, &op)
	}
	return nil
}
