package game

import (
	_ "embed"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

// The dream look of the Prologue (dream.kage): a screen filter on the
// finished picture, like the CRT, for the Prologue only. The nights stay dark.
//
//go:embed dream.kage
var dreamSource []byte

// Dream is the compiled filter and the picture it reads.
type Dream struct {
	shader *ebiten.Shader
	buf    *ebiten.Image // the Prologue is drawn here first
}

// The filter's two amounts (see dream.kage).
const (
	dreamLift = 0.36 // blacks come up to a light-ish grey; faces and text stay readable
	dreamHaze = 0.7
)

// NewDream compiles the filter.
func NewDream() (*Dream, error) {
	shader, err := ebiten.NewShader(dreamSource)
	if err != nil {
		return nil, fmt.Errorf("compiling the dream shader: %w", err)
	}
	return &Dream{shader: shader, buf: ebiten.NewImage(ScreenW, ScreenH)}, nil
}

// Target is the picture to draw the Prologue on; Apply then puts it on the screen.
func (d *Dream) Target() *ebiten.Image {
	d.buf.Clear()
	return d.buf
}

// Apply draws the filtered picture onto dst.
func (d *Dream) Apply(dst *ebiten.Image) {
	op := &ebiten.DrawRectShaderOptions{}
	op.Images[0] = d.buf
	op.Uniforms = map[string]any{"Lift": float32(dreamLift), "Haze": float32(dreamHaze)}
	dst.DrawRectShader(ScreenW, ScreenH, d.shader, op)
}
