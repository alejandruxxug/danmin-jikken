package game

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

// The day is shown through an old CRT (Art Direction -> Day screen CRT). One
// Kage shader (crt.kage) runs on the finished picture, only in day scenes.
// Every number below is an amount from that note.
//
//go:embed crt.kage
var crtSource []byte

// CRTParams are the numbers the shader takes (the seven of the note, plus
// PowerOn for the turn-on), named as in crt.kage.
type CRTParams struct {
	Scanline, Vignette, Curve, Day, Invert, Time, Filmed, PowerOn float64
}

// DefaultCRT gives the amounts of the Art Direction note for a day, at a
// time in seconds. invert is 1 during the strobe's inverted frames.
func DefaultCRT(day int, seconds, invert float64) CRTParams {
	return CRTParams{
		Scanline: 0.15,  // ~15 % darker rows
		Vignette: 0.35,  // corners ~35 % darker
		Curve:    0.025, // a small barrel bend: the corners move in a few pixels
		Day:      float64(max(1, min(day, 3))),
		Invert:   invert,
		Time:     seconds,
		Filmed:   1,
		PowerOn:  -1, // not turning on
	}
}

// CRT holds the compiled shader and the picture it draws into.
type CRT struct {
	shader *ebiten.Shader
	buf    *ebiten.Image
}

// NewCRT compiles the shader.
func NewCRT() (*CRT, error) {
	shader, err := ebiten.NewShader(crtSource)
	if err != nil {
		return nil, fmt.Errorf("compiling the CRT shader: %w", err)
	}
	return &CRT{shader: shader}, nil
}

// Apply draws src (the 480x270 canvas) through the shader into an image
// scale times bigger and returns it. The image is reused between calls.
func (c *CRT) Apply(src *ebiten.Image, scale int, p CRTParams) *ebiten.Image {
	w, h := src.Bounds().Dx()*scale, src.Bounds().Dy()*scale
	if c.buf == nil || c.buf.Bounds().Dx() != w || c.buf.Bounds().Dy() != h {
		c.buf = ebiten.NewImage(w, h)
	}
	// The rectangle is the source's size and GeoM stretches it over the buffer;
	// the shader works out its own positions from the buffer's size.
	op := &ebiten.DrawRectShaderOptions{}
	op.Images[0] = src
	op.GeoM.Scale(float64(scale), float64(scale))
	op.Uniforms = map[string]any{
		"Scanline": float32(p.Scanline), "Vignette": float32(p.Vignette), "Curve": float32(p.Curve),
		"Day": float32(p.Day), "Invert": float32(p.Invert), "Time": float32(p.Time), "Filmed": float32(p.Filmed), "PowerOn": float32(p.PowerOn),
	}
	c.buf.DrawRectShader(src.Bounds().Dx(), src.Bounds().Dy(), c.shader, op)
	return c.buf
}

// usesCRT says whether the scene is shown through the CRT: the day scenes
// are, the nights never are. A scene can decide for itself (the strobe's
// day half does); otherwise the id says.
func usesCRT(id string, s Scene) bool {
	if own, ok := s.(interface{ UsesCRT() bool }); ok {
		return own.UsesCRT()
	}
	return strings.HasPrefix(id, "day.")
}
