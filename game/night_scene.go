package game

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"danmingame/night"
)

// The night scenes play a script from game/night/scripts. The engine, the
// script format and the drawing live in the night package; this file is the
// glue that feeds it the player's input and says what comes next.

// nightScene plays one night script (the Prologue or Night 1).
type nightScene struct {
	env    *Env
	script string // night.Prologue or night.Night1
	art    *night.Art
	engine *night.Engine

	lastCursorX, lastCursorY int // the mouse only changes the highlight when it moves
}

func (s *nightScene) Enter(env *Env) error {
	s.env = env
	var err error
	if env.nightArt == nil {
		font, err := LoadFont()
		if err != nil {
			return err
		}
		if env.nightArt, err = night.LoadArt(env.Sprites, font); err != nil {
			return err
		}
	}
	s.art = env.nightArt
	script, err := night.LoadScript(s.script)
	if err != nil {
		return err
	}
	if s.engine, err = night.NewEngine(script, env.Flags); err != nil {
		return err
	}
	if env.NightLine > 0 { // the backdoor's -line N
		if err := s.engine.FastForward(env.NightLine); err != nil {
			return err
		}
		env.NightLine = 0
	}
	env.Night = s.engine
	s.lastCursorX, s.lastCursorY = env.Input.Cursor()
	return nil
}

func (s *nightScene) Update() error {
	in := s.env.Input
	e := s.engine
	if n := len(e.Options()); n > 0 {
		s.updateMenu(n)
	} else if _, _, clicked := in.Click(); clicked || in.KeyPressed(ebiten.KeySpace) || in.KeyPressed(ebiten.KeyEnter) {
		if err := e.Advance(); err != nil {
			return err
		}
	}
	if err := e.Tick(); err != nil {
		return err
	}
	if to, ended := e.Ended(); ended {
		return s.finish(to)
	}
	return nil
}

// updateMenu moves the highlight (mouse or arrow keys) and picks an option
// (click on a button, Enter or Space).
func (s *nightScene) updateMenu(n int) {
	in, e := s.env.Input, s.engine
	rects := night.ChoiceRects(n)
	cx, cy := in.Cursor()
	if cx != s.lastCursorX || cy != s.lastCursorY {
		s.lastCursorX, s.lastCursorY = cx, cy
		for i, r := range rects {
			if cx >= r.Min.X && cx < r.Max.X && cy >= r.Min.Y && cy < r.Max.Y {
				e.SetHover(i)
			}
		}
	}
	if in.KeyPressed(ebiten.KeyArrowUp) {
		e.MoveHover(-1)
	}
	if in.KeyPressed(ebiten.KeyArrowDown) {
		e.MoveHover(1)
	}
	pick := -1
	if x, y, ok := in.Click(); ok {
		for i, r := range rects {
			if x >= r.Min.X && x < r.Max.X && y >= r.Min.Y && y < r.Max.Y {
				pick = i
			}
		}
	}
	if in.KeyPressed(ebiten.KeySpace) || in.KeyPressed(ebiten.KeyEnter) {
		pick = e.Hover()
	}
	if pick >= 0 {
		_ = e.Choose(pick) // pick is always inside the menu
	}
}

// finish leaves the night: the script's "end" says which day comes next.
// The next day starts at the reception, as it did before the nights existed.
func (s *nightScene) finish(to string) error {
	var day int
	if _, err := fmt.Sscanf(to, "day%d", &day); err != nil {
		return fmt.Errorf("night script %s ends with %q: want day1, day2 or day3: %w", s.script, to, err)
	}
	s.env.Day = day
	s.env.SwitchTo(SceneDayReception)
	return nil
}

func (s *nightScene) Draw(screen *ebiten.Image) { s.art.Draw(screen, s.engine) }

func (s *nightScene) Exit() {
	if s.env.Night == s.engine {
		s.env.Night = nil
	}
}
