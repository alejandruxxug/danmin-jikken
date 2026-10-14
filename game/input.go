package game

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Input is what scenes ask for instead of calling Ebitengine directly. It
// mixes the real keyboard and mouse with the inputs the agent backdoor
// injects, so a scene cannot tell a script from a person.
type Input struct {
	keys     map[ebiten.Key]bool // injected this frame only
	typed    []rune              // text injected this frame only (the script's "type")
	fakeDown bool                // the backdoor holds the left button down
	released bool                // the backdoor let go this frame
	clicked  bool
	clickX   int
	clickY   int

	// The agent backdoor holds the cursor still at a chosen spot. Without it
	// the real mouse is used, except in test mode, where a stray mouse must
	// not pan the camera: then the cursor rests in the middle of the screen.
	fakeMouse  bool
	mouseX     int
	mouseY     int
	IgnoreReal bool // set in test mode

	// Where the 480x270 picture sits on the real screen (see fitScreen):
	// scale 0 means the screen is the picture itself.
	scale, offX, offY float64
}

// NewInput makes an Input with nothing injected.
func NewInput() *Input {
	return &Input{keys: make(map[ebiten.Key]bool)}
}

// KeyPressed is true on the frame a key goes down.
func (in *Input) KeyPressed(k ebiten.Key) bool {
	return in.keys[k] || inpututil.IsKeyJustPressed(k)
}

// Typed returns the characters typed on this frame: the real keyboard's and
// the ones the backdoor injected.
func (in *Input) Typed() []rune {
	return append(ebiten.AppendInputChars(nil), in.typed...)
}

// InjectText pretends the characters of s were typed on the next frame.
func (in *Input) InjectText(s string) { in.typed = append(in.typed, []rune(s)...) }

// Held is true while the left mouse button is down.
func (in *Input) Held() bool {
	return in.fakeDown || ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
}

// Released is true on the frame the left button comes up.
func (in *Input) Released() bool {
	return in.released || inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
}

// InjectDown presses the left button at (x, y) and keeps it down (and the
// cursor there) until InjectUp.
func (in *Input) InjectDown(x, y int) {
	in.InjectClick(x, y)
	in.InjectMouse(x, y)
	in.fakeDown = true
}

// InjectUp lets the left button go.
func (in *Input) InjectUp() {
	in.fakeDown = false
	in.released = true
}

// SetView tells Input how the game picture is placed on the screen.
func (in *Input) SetView(scale, offX, offY float64) {
	in.scale, in.offX, in.offY = scale, offX, offY
}

// mouse is the real mouse in 1x game pixels. Outside the picture (on the
// black bars) the numbers fall outside 0..479 and 0..269, so no button is hit.
func (in *Input) mouse() (x, y int) {
	px, py := ebiten.CursorPosition()
	if in.scale <= 0 {
		return px, py
	}
	return screenToGame(px, py, in.scale, in.offX, in.offY)
}

// screenToGame turns a point on the real screen into a game pixel, given
// where the picture sits (its scale and the offset of its top-left corner).
// math.Floor, so a point just left of or above the picture is column or row
// -1 and not 0.
func screenToGame(px, py int, scale, offX, offY float64) (x, y int) {
	return int(math.Floor((float64(px) - offX) / scale)), int(math.Floor((float64(py) - offY) / scale))
}

// AnyKeyPressed is true on the frame any key goes down, except the
// fullscreen keys (F11, Alt), which must not start the game.
func (in *Input) AnyKeyPressed() bool {
	for k := range in.keys {
		if k != ebiten.KeyF11 {
			return true
		}
	}
	for _, k := range inpututil.AppendJustPressedKeys(nil) {
		if k != ebiten.KeyF11 && k != ebiten.KeyAlt && k != ebiten.KeyAltLeft && k != ebiten.KeyAltRight {
			return true
		}
	}
	return false
}

// FullscreenPressed is true on the frame F11, Alt+Enter or Cmd+F goes down.
// macOS keeps F11 for itself, so Option+Enter and Cmd+F are there too.
func (in *Input) FullscreenPressed() bool {
	altEnter := ebiten.IsKeyPressed(ebiten.KeyAlt) && in.KeyPressed(ebiten.KeyEnter)
	cmdF := ebiten.IsKeyPressed(ebiten.KeyMeta) && in.KeyPressed(ebiten.KeyF)
	return in.KeyPressed(ebiten.KeyF11) || altEnter || cmdF
}

// Click returns the 1x game position of a left click on this frame.
func (in *Input) Click() (x, y int, ok bool) {
	if in.clicked {
		return in.clickX, in.clickY, true
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y = in.mouse()
		return x, y, true
	}
	return 0, 0, false
}

// RightClicked is true on the frame the right button goes down (the gallery's
// way back). The backdoor does not inject it: Esc does the same there.
func (in *Input) RightClicked() bool {
	return inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight)
}

// Cursor returns the 1x game position of the mouse.
func (in *Input) Cursor() (x, y int) {
	switch {
	case in.fakeMouse:
		return in.mouseX, in.mouseY
	case in.IgnoreReal:
		return ScreenW / 2, ScreenH / 2
	}
	return in.mouse()
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
	in.typed = nil
	in.released = false
	in.clicked = false
}
