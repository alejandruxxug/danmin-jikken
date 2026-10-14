package game

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"

	"danmingame/night"
)

// The night scenes play a script from game/night/scripts. The engine, the
// script format and the drawing live in the night package; this file is the
// glue that feeds it the player's input and says what comes next.

// nightScene plays one night script (the Prologue or Night 1).
type nightScene struct {
	env    *Env
	script string // night.Prologue, night.Night1 ... or an ending's script
	// onEnd, when set, takes over the end of the script (the ending scene
	// uses it for "end card"); otherwise finish decides.
	onEnd  func(to string) error
	art    *night.Art
	engine *night.Engine
	dream  *Dream // the Prologue's washed-out filter; nil for the nights

	lastCursorX, lastCursorY int // the mouse only changes the highlight when it moves
}

func (s *nightScene) Enter(env *Env) error {
	s.env = env
	var err error
	if s.art, err = env.NightArt(); err != nil {
		return err
	}
	s.art.Dream = s.script == night.Prologue // the Prologue is a dream: white, not black
	if s.art.Dream {
		if s.dream, err = NewDream(); err != nil {
			return err
		}
	}
	script, err := night.LoadScript(s.script)
	if err != nil {
		return err
	}
	if s.engine, err = night.NewEngineFor(script, env.Flags, Username()); err != nil {
		return err
	}
	if env.NightLine > 0 { // the backdoor's -line N
		if err := s.engine.FastForward(env.NightLine); err != nil {
			return err
		}
		env.NightLine = 0
	}
	s.engine.AutoPlay = env.Auto
	if s.onEnd == nil { // the endings bring their own music
		env.SetMusic("night_ambience")
	}
	s.playCues() // cues of the lines a backdoor jump skipped are dropped
	env.Night = s.engine
	s.lastCursorX, s.lastCursorY = env.Input.Cursor()
	return nil
}

func (s *nightScene) Update() error {
	in := s.env.Input
	e := s.engine
	if s.script == night.Prologue && s.env.SeenPrologue && s.engine.Ticks() > 0 {
		// A later launch: Esc skips the Prologue. Clicks, Space and Enter
		// still read it line by line, so a click never throws it away.
		if in.KeyPressed(ebiten.KeyEscape) {
			return s.finish("title")
		}
	}
	if in.KeyPressed(ebiten.KeyA) {
		s.toggleAuto()
	}
	storyClick := false // a click that is for the story, not for the AUTO toggle
	if x, y, ok := in.Click(); ok {
		if image.Pt(x, y).In(night.AutoRect) && !e.Cinematic() { // the toggle is hidden under the bars
			s.toggleAuto()
		} else {
			storyClick = true
		}
	}
	if n := len(e.Options()); n > 0 {
		s.updateMenu(n)
	} else if storyClick || in.KeyPressed(ebiten.KeySpace) || in.KeyPressed(ebiten.KeyEnter) {
		if err := e.Advance(); err != nil {
			return err
		}
	}
	if err := e.Tick(); err != nil {
		return err
	}
	s.playCues()
	if to, ended := e.Ended(); ended {
		return s.finish(to)
	}
	return nil
}

// UsesCRT says whether this frame goes through the day's CRT: only the
// realization's match cut to the day 3 report (a day picture inside a night),
// not the split-screen sting that follows it.
func (s *nightScene) UsesCRT() bool {
	if s.engine == nil || !s.engine.DayLook() {
		return false
	}
	_, sting := s.engine.FxTicks("sting")
	return !(sting && s.engine.FxProgress("sting") < 1)
}

// playCues plays the sounds the script asked for since the last frame.
func (s *nightScene) playCues() {
	for _, c := range s.engine.TakeCues() {
		switch {
		case c.Kind == "sfx" && c.Name != "off":
			playSFX(sfxFile(c.Name))
		case c.Kind == "music" && c.Name == "off":
			s.env.SetMusic("")
		case c.Kind == "music" && isJingle(c.Name):
			if s.env.Sound != nil { // plays once, it does not loop
				s.env.Sound.PlayMusic(musicFile(c.Name), false, 0)
			}
		case c.Kind == "music":
			s.env.SetMusic(c.Name)
		}
	}
}

// toggleAuto flips AUTO (the A key or a click on the toggle). The choice is
// kept in Env so the next night starts with it.
func (s *nightScene) toggleAuto() {
	s.engine.AutoPlay = !s.engine.AutoPlay
	s.env.Auto = s.engine.AutoPlay
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
	if s.onEnd != nil {
		return s.onEnd(to)
	}
	if s.script == night.Prologue { // the cold open: the dream ends in the title
		s.env.SeenPrologue = true
		s.env.TitleWhite = true // its white carries into the title drop
		s.env.SwitchTo(SceneTitle)
		return s.env.SaveSettings()
	}
	if to == "ending" { // the end of night 3: the flags pick the ending
		s.env.SwitchTo(endingSceneID(night.EndingFor(s.env.Flags)))
		return nil
	}
	var day int
	if _, err := fmt.Sscanf(to, "day%d", &day); err != nil {
		return fmt.Errorf("night script %s ends with %q: want day1, day2 or day3: %w", s.script, to, err)
	}
	s.env.Day = day
	s.env.StartPowerOn() // the day turns on like a CRT
	s.env.SwitchTo(SceneDayReception)
	return nil
}

func (s *nightScene) Draw(screen *ebiten.Image) {
	if s.dream == nil {
		s.art.Draw(screen, s.engine)
		return
	}
	s.art.Draw(s.dream.Target(), s.engine)
	s.dream.Apply(screen)
}

func (s *nightScene) Exit() {
	s.art.Dream = false
	if s.env.Night == s.engine {
		s.env.Night = nil
	}
}
