package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Input is what scenes ask for instead of calling Ebitengine directly. It
// mixes the real keyboard and mouse with the inputs the agent backdoor
// injects, so a scene cannot tell a script from a person.
type Input struct {
	keys    map[ebiten.Key]bool // injected this frame only
	clicked bool
	clickX  int
	clickY  int

	// The agent backdoor holds the cursor still at a chosen spot. Without it
	// the real mouse is used, except in test mode, where a stray mouse must
	// not pan the camera: then the cursor rests in the middle of the screen.
	fakeMouse  bool
	mouseX     int
	mouseY     int
	IgnoreReal bool // set in test mode
}

// NewInput makes an Input with nothing injected.
func NewInput() *Input {
	return &Input{keys: make(map[ebiten.Key]bool)}
}

// KeyPressed is true on the frame a key goes down.
func (in *Input) KeyPressed(k ebiten.Key) bool {
	return in.keys[k] || inpututil.IsKeyJustPressed(k)
}

// Click returns the 1x game position of a left click on this frame.
func (in *Input) Click() (x, y int, ok bool) {
	if in.clicked {
		return in.clickX, in.clickY, true
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y = ebiten.CursorPosition() // already in 480x270 game pixels
		return x, y, true
	}
	return 0, 0, false
}

// Cursor returns the 1x game position of the mouse.
func (in *Input) Cursor() (x, y int) {
	switch {
	case in.fakeMouse:
		return in.mouseX, in.mouseY
	case in.IgnoreReal:
		return ScreenW / 2, ScreenH / 2
	}
	return ebiten.CursorPosition()
}

// InjectMouse moves the pretend cursor to (x, y) and keeps it there.
func (in *Input) InjectMouse(x, y int) {
	in.fakeMouse = true
	in.mouseX, in.mouseY = x, y
}

// InjectKey pretends a key went down on the next frame.
func (in *Input) InjectKey(k ebiten.Key) { in.keys[k] = true }

// InjectClick pretends a left click happened at (x, y) on the next frame.
func (in *Input) InjectClick(x, y int) {
	in.clicked = true
	in.clickX, in.clickY = x, y
}

// endFrame forgets the injected inputs once the scene has seen them.
func (in *Input) endFrame() {
	clear(in.keys)
	in.clicked = false
}
