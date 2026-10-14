package game

import (
	"image"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

// The settings screen (Build Plan -> M7): CRT on/off, reduce flashing, AUTO
// play, music and sound volume. Every change is saved at once, in the
// settings file next to the scores file. Plain text rows on black; the real
// look waits for an approved Figma view.

// The rows, top to bottom.
const (
	rowCRT = iota
	rowFlashing
	rowAuto
	rowMusic
	rowSFX
	rowBack
	rowCount
)

const (
	settingsFirstY = 70
	settingsStepY  = 22
	volumeSteps    = 10 // the volumes move in tenths
)

// settingsScene is the settings screen.
type settingsScene struct {
	env  *Env
	font *Font
	sel  int

	// The cursor's spot on the last frame: a row is picked by hovering only
	// when the cursor moves, so a cursor resting on a row does not undo the
	// arrow keys.
	lastCursor image.Point
}

func (s *settingsScene) Enter(env *Env) error {
	s.env = env
	s.sel = 0
	cx, cy := env.Input.Cursor()
	s.lastCursor = image.Pt(cx, cy)
	var err error
	s.font, err = LoadFont()
	return err
}

// rowRect is the mouse area of a row: nearly the whole width.
func rowRect(i int) image.Rectangle {
	y := settingsFirstY + i*settingsStepY
	return image.Rect(100, y-5, ScreenW-100, y+13)
}

func (s *settingsScene) Update() error {
	in := s.env.Input
	cx, cy := in.Cursor()
	if cur := image.Pt(cx, cy); cur != s.lastCursor {
		for i := 0; i < rowCount; i++ {
			if cur.In(rowRect(i)) {
				s.sel = i
			}
		}
		s.lastCursor = cur
	}
	if in.KeyPressed(ebiten.KeyArrowUp) {
		s.sel = (s.sel + rowCount - 1) % rowCount
	}
	if in.KeyPressed(ebiten.KeyArrowDown) {
		s.sel = (s.sel + 1) % rowCount
	}
	change := 0 // -1 and +1 move a volume; a toggle ignores the sign
	switch {
	case in.KeyPressed(ebiten.KeyArrowLeft):
		change = -1
	case in.KeyPressed(ebiten.KeyArrowRight), in.KeyPressed(ebiten.KeyEnter), in.KeyPressed(ebiten.KeySpace):
		change = 1
	}
	if x, y, ok := in.Click(); ok {
		for i := 0; i < rowCount; i++ {
			if image.Pt(x, y).In(rowRect(i)) {
				s.sel = i
				change = 1
				if x < ScreenW/2 { // the left half of a volume row turns it down
					change = -1
				}
			}
		}
	}
	if change != 0 {
		return s.change(s.sel, change)
	}
	if in.KeyPressed(ebiten.KeyEscape) {
		return s.leave()
	}
	return nil
}

// change applies one press to a row and saves the settings.
func (s *settingsScene) change(row, dir int) error {
	e := s.env
	switch row {
	case rowCRT:
		e.CRT = !e.CRT
	case rowFlashing:
		e.ReduceFlashing = !e.ReduceFlashing
	case rowAuto:
		e.Auto = !e.Auto
	case rowMusic:
		e.MusicVolume = stepVolume(e.MusicVolume, dir)
		e.Apply(e.Settings())
	case rowSFX:
		e.SFXVolume = stepVolume(e.SFXVolume, dir)
		e.Apply(e.Settings())
		playSFX("stamp_thunk") // so the new volume can be heard
	case rowBack:
		return s.leave()
	}
	if err := e.SaveSettings(); err != nil {
		return err
	}
	return nil
}

// leave goes back to the title.
func (s *settingsScene) leave() error {
	s.env.SwitchTo(SceneTitle)
	return nil
}

// stepVolume moves a volume one tenth up or down, between 0 and 1.
func stepVolume(v float64, dir int) float64 {
	n := int(v*volumeSteps+0.5) + dir
	return float64(min(max(n, 0), volumeSteps)) / volumeSteps
}

// onOff is the label of a toggle.
func onOff(on bool) string {
	if on {
		return "ON"
	}
	return "OFF"
}

// volumeBar draws a volume as ten cells: "#####-----".
func volumeBar(v float64) string {
	n := int(v*volumeSteps + 0.5)
	return strings.Repeat("#", n) + strings.Repeat("-", volumeSteps-n)
}

// rowText is a row's label and its value.
func (s *settingsScene) rowText(i int) (label, value string) {
	e := s.env
	switch i {
	case rowCRT:
		return "CRT EFFECT", onOff(e.CRT)
	case rowFlashing:
		return "REDUCE FLASHING", onOff(e.ReduceFlashing)
	case rowAuto:
		return "AUTO PLAY", onOff(e.Auto)
	case rowMusic:
		return "MUSIC", "< " + volumeBar(e.MusicVolume) + " >"
	case rowSFX:
		return "SOUND EFFECTS", "< " + volumeBar(e.SFXVolume) + " >"
	}
	return "BACK", ""
}

func (s *settingsScene) Draw(screen *ebiten.Image) {
	screen.Fill(paletteInk)
	title := "SETTINGS"
	s.font.Draw(screen, title, (ScreenW-s.font.Width(title))/2, 30, colPaperRGBA)
	for i := 0; i < rowCount; i++ {
		label, value := s.rowText(i)
		y := settingsFirstY + i*settingsStepY
		c := colMidRGBA
		if i == s.sel {
			c = colPaperRGBA
			s.font.Draw(screen, ">", 108, y, c)
		}
		if i == rowBack {
			s.font.Draw(screen, label, (ScreenW-s.font.Width(label))/2, y, c)
			continue
		}
		s.font.Draw(screen, label, 124, y, c)
		s.font.Draw(screen, value, ScreenW-124-s.font.Width(value), y, c)
	}
	hint := "UP / DOWN: CHOOSE   LEFT / RIGHT: CHANGE   ESC: BACK"
	s.font.Draw(screen, hint, (ScreenW-s.font.Width(hint))/2, ScreenH-20, colMidRGBA)
}

func (s *settingsScene) Exit() {}
