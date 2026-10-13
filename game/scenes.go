package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// Placeholder scene. It only clears the screen to a palette colour: the
// real visuals wait for the approved Figma views (Art Direction -> Art log).

var (
	paletteInk = color.RGBA{0x0b, 0x0b, 0x0c, 0xff} // assets/palette/danmin.hex, entry 1
)

// titleScene is the empty title screen.
type titleScene struct{}

func (titleScene) Enter(*Env) error { return nil }
func (titleScene) Update() error    { return nil }
func (titleScene) Exit()            {}
func (titleScene) Draw(screen *ebiten.Image) {
	screen.Fill(paletteInk)
}
