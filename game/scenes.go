package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// Placeholder scenes. They only clear the screen to a palette colour: the
// real visuals wait for the approved Figma views (Art Direction -> Art log).

var (
	paletteInk   = color.RGBA{0x0b, 0x0b, 0x0c, 0xff} // assets/palette/danmin.hex, entry 1
	paletteSlate = color.RGBA{0x2e, 0x2e, 0x33, 0xff} // entry 2
)

// titleScene is the empty title screen.
type titleScene struct{}

func (titleScene) Enter(*Env) error { return nil }
func (titleScene) Update() error    { return nil }
func (titleScene) Exit()            {}
func (titleScene) Draw(screen *ebiten.Image) {
	screen.Fill(paletteInk)
}

// receptionScene is the empty reception desk. It already runs the ward, so
// patients fall asleep behind the blank screen and the backdoor can dump them.
type receptionScene struct {
	env *Env
}

func (s *receptionScene) Enter(env *Env) error {
	s.env = env
	return env.StartWard()
}

func (s *receptionScene) Update() error {
	// Collect what the patient goroutines reported since the last frame.
	// Later milestones turn these events into patients dropping off.
	s.env.Ward().Drain()
	return nil
}

func (s *receptionScene) Exit() {
	// Stop the goroutines before the scene is gone. An error here would
	// mean a hospital call failed in the background; nothing can fix it now.
	_ = s.env.StopWard()
}

func (s *receptionScene) Draw(screen *ebiten.Image) {
	screen.Fill(paletteSlate)
}
