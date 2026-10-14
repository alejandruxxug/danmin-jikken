package game

import (
	"image"
	"io/fs"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"danmingame/sprites"
)

// The title screen. The Prologue plays before it (the cold open): the white
// of the dream carries over, the title drops in, and START begins day 1 (the
// day turns on like a CRT). If a "title" sheet is in the sprites folder it is
// drawn (its layer "logo" drops, every other layer stays put); until then the
// title is the ending font, large, on black.

// Title timing, in ticks, and layout, in pixels.
const (
	titleFall     = 16 // the logo falls from above the screen
	titleBounce   = 8  // and hops once after the slam
	titleShake    = 8  // the screen shakes after the slam
	titleWhite    = 90 // the carried-over white fades out over this long
	titleMenuFrom = 30 // the menu answers keys and clicks from here
	titleStartY   = 190
	titleStepY    = 14
	titleLogoY    = 40 // top of the fallback logo's glyph cells
	titleDropFrom = -150
)

// titleDrop is how far (in pixels) the logo is above its place at a tick:
// it falls, speeding up, slams and hops once. Zero when it has settled.
func titleDrop(ticks int) int {
	switch {
	case ticks < 0:
		return titleDropFrom
	case ticks < titleFall:
		t := float64(ticks) / titleFall
		return int(float64(titleDropFrom) * (1 - t*t))
	case ticks < titleFall+titleBounce:
		t := float64(ticks-titleFall) / titleBounce
		return -int(6 * 4 * t * (1 - t))
	}
	return 0
}

// titleShakeOffset is the screen shake after the slam: a fixed pattern that dies away.
func titleShakeOffset(ticks int) (dx, dy int) {
	i := ticks - titleFall
	if i < 0 || i >= titleShake {
		return 0, 0
	}
	amp := 3 * (titleShake - i) / titleShake // 3, then 2, then 1 pixel
	if i%2 == 0 {
		return amp, -amp
	}
	return -amp, amp
}

// titleWhiteAlpha is how much of the dream's white is still over the title.
func titleWhiteAlpha(ticks int) float64 {
	return max(0, 1-float64(ticks)/titleWhite)
}

// Menu entries of the title.
const (
	titleStart = iota
	titleSettings
)

var titleItems = []string{"START", "SETTINGS"}

// titleScene is the title.
type titleScene struct {
	env   *Env
	font  *Font
	bigF  *sprites.Sheet // the ending font, for the fallback logo
	art   *sprites.Sheet // the title sheet, nil until it is drawn
	ticks int
	sel   int
	white bool // the Prologue's white carries over
	drops bool // the logo drops (not when coming back from the settings)
}

func (s *titleScene) Enter(env *Env) error {
	s.env = env
	var err error
	if s.font, err = LoadFont(); err != nil {
		return err
	}
	if s.bigF, err = LoadSheet(env.Sprites, "font_ending"); err != nil {
		return err
	}
	if _, statErr := fs.Stat(env.Sprites, "title.json"); statErr == nil {
		if s.art, err = LoadSheet(env.Sprites, "title"); err != nil {
			return err
		}
	}
	s.white, env.TitleWhite = env.TitleWhite, false
	s.drops, env.TitleSettled = !env.TitleSettled, false
	if !s.drops {
		s.ticks = titleFall + titleBounce + titleShake // already settled
	}
	env.SetMusic("title")
	return nil
}

func (s *titleScene) Update() error {
	s.ticks++
	in := s.env.Input
	if s.ticks < titleMenuFrom {
		return nil
	}
	cx, cy := in.Cursor()
	for i := range titleItems {
		if s.itemRect(i).Min.X <= cx && cx < s.itemRect(i).Max.X && s.itemRect(i).Min.Y <= cy && cy < s.itemRect(i).Max.Y {
			s.sel = i
		}
	}
	if in.KeyPressed(ebiten.KeyArrowUp) {
		s.sel = (s.sel + len(titleItems) - 1) % len(titleItems)
	}
	if in.KeyPressed(ebiten.KeyArrowDown) {
		s.sel = (s.sel + 1) % len(titleItems)
	}
	pick := -1
	if x, y, ok := in.Click(); ok {
		for i := range titleItems {
			if image.Pt(x, y).In(s.itemRect(i)) {
				pick = i
			}
		}
	}
	if in.KeyPressed(ebiten.KeyEnter) || in.KeyPressed(ebiten.KeySpace) {
		pick = s.sel
	}
	switch pick {
	case titleStart:
		s.env.NewRun()
		s.env.StartPowerOn() // day 1 turns on like a CRT
		s.env.SwitchTo(SceneDayReception)
	case titleSettings:
		s.env.TitleSettled = true // coming back, the logo is already there
		s.env.SwitchTo(SceneSettings)
	}
	return nil
}

// itemRect is the mouse area of a menu entry.
func (s *titleScene) itemRect(i int) image.Rectangle {
	w := s.font.Width(titleItems[i])
	y := titleStartY + i*titleStepY
	return image.Rect((ScreenW-w)/2-10, y-3, (ScreenW+w)/2+10, y+11)
}

func (s *titleScene) Draw(screen *ebiten.Image) {
	screen.Fill(paletteInk)
	dx, dy := titleShakeOffset(s.ticks)
	drop := titleDrop(s.ticks)
	if s.art != nil {
		s.drawSheet(screen, dx, dy+drop)
	} else {
		drawBigText(screen, s.bigF, "DANMIN", titleLogoY+dy+drop, dx)
		drawBigText(screen, s.bigF, "JIKKEN", titleLogoY+34+dy+drop, dx)
	}
	if s.ticks >= titleMenuFrom {
		for i, item := range titleItems {
			c := colMidRGBA
			label := item
			if i == s.sel {
				c, label = colPaperRGBA, "> "+item+" <"
			}
			s.font.Draw(screen, label, (ScreenW-s.font.Width(label))/2, titleStartY+i*titleStepY, c)
		}
		warn := "THIS GAME CONTAINS STRONG FLASHING LIGHTS."
		s.font.Draw(screen, warn, (ScreenW-s.font.Width(warn))/2, ScreenH-16, colMidRGBA)
	}
	if s.white {
		if a := titleWhiteAlpha(s.ticks); a > 0 {
			vector.FillRect(screen, 0, 0, ScreenW, ScreenH, colPaperAlpha(a), false)
		}
	}
}

// drawSheet draws the title sheet: every layer in place, the "logo" layer
// moved by (dx, dy).
func (s *titleScene) drawSheet(screen *ebiten.Image, dx, dy int) {
	for _, layer := range s.art.Layers {
		x, y := 0, 0
		if layer == "logo" {
			x, y = dx, dy
		}
		_ = s.art.Draw(screen, []string{layer}, 0, x, y)
	}
}

func (s *titleScene) Exit() {}
