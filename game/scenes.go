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

// titleScene is the empty title screen. Until the real title exists (M7),
// any key or click starts the game: the Prologue, then day 1.
type titleScene struct{ env *Env }

func (s *titleScene) Enter(env *Env) error { s.env = env; return nil }
func (s *titleScene) Update() error {
	if _, _, clicked := s.env.Input.Click(); clicked || s.env.Input.AnyKeyPressed() {
		s.env.Day = 1
		s.env.SwitchTo(SceneNightPrologue) // the gurney comes first
	}
	return nil
}
func (s *titleScene) Exit() {}
func (s *titleScene) Draw(screen *ebiten.Image) {
	screen.Fill(paletteInk)
}
