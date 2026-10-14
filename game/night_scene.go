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
	onEnd func(to string) error
	// replay, when set, makes this a replay from the SCENES gallery: one
	// stretch of the script, played with a copy of the saved flags. It never
	// unlocks anything or saves anything, and its end (or Esc) goes back to
	// the gallery.
	replay *sceneCard
	art    *night.Art
	engine *night.Engine

	lastCursorX, lastCursorY int // the mouse only changes the highlight when it moves
}

func (s *nightScene) Enter(env *Env) error {
	s.env = env
	var err error
	if s.art, err = env.NightArt(); err != nil {
		return err
	}
	if s.replay != nil {
		if s.engine, err = newReplayEngine(*s.replay, env.SavedFlags); err != nil {
			return err
		}
		s.engine.TakeCues() // the sounds of the skipped lines are not played
		env.Replaying = s.replay.ID
	} else {
		script, err := night.LoadScript(s.script)
		if err != nil {
			return err
		}
		if s.engine, err = night.NewEngineFor(script, env.Flags, Username()); err != nil {
			return err
		}
	}
	if env.NightLine > 0 && s.replay == nil { // the backdoor's -line N
		if err := s.engine.FastForward(env.NightLine); err != nil {
			return err
		}
		env.NightLine = 0
	}
	s.engine.AutoPlay = env.Auto
	switch {
	case s.replay != nil:
		s.replayMusic()
	case s.onEnd != nil: // the endings bring their own music
	case s.script == night.Night3:
		// Night 3 is silent in its black opening; the script starts its music.
		// A backdoor jump (-line N) skipped those cues, so start the one that
		// should be playing by that line.
		env.SetMusic("")
		if m := s.engine.LastMusic; m != "" && m != "off" && m != "fade" {
			env.SetMusic(m)
		}
	default:
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
	if s.replay != nil && in.KeyPressed(ebiten.KeyEscape) {
		s.env.SwitchTo(SceneScenes) // Esc leaves a replay at any time
		return nil
	}
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
	if s.replay == nil { // reaching a gallery card unlocks it (a replay never does)
		if err := s.env.NoteScene(s.script, e.Line()); err != nil {
			return err
		}
	}
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
		case c.Kind == "music" && c.Name == "fade":
			s.env.FadeMusic(int(c.Seconds * 60))
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

// replayMusic starts the music a replay needs: night 3 and the endings are
// silent until their script asks for a track (the one the skipped lines asked
// for last), the other nights have the ambience.
func (s *nightScene) replayMusic() {
	env := s.env
	if !quietScript(s.script) {
		env.SetMusic("night_ambience")
		return
	}
	env.SetMusic("")
	if m := s.engine.LastMusic; m != "" && m != "off" && m != "fade" {
		env.SetMusic(m)
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
	// The end of a replay: back to the gallery, nothing else changes. An
	// ending's replay has an onEnd: the ending scene plays on to its title.
	if s.replay != nil && s.onEnd == nil {
		s.env.SwitchTo(SceneScenes)
		return nil
	}
	if s.onEnd != nil {
		return s.onEnd(to)
	}
	if s.script == night.Prologue { // the cold open: the cut to white ends in the title
		s.env.SeenPrologue = true
		s.env.TitleWhite = true // its white carries into the title drop
		s.env.SwitchTo(SceneTitle)
		return s.env.SaveSettings()
	}
	if to == "ending" { // the end of night 3: the flags pick the ending
		if s.env.Sound != nil && s.env.Sound.Fading() {
			s.env.Sound.StopMusic() // a fade not yet finished ends here: the ending starts in silence
		}
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
	s.art.Draw(screen, s.engine)
}

func (s *nightScene) Exit() {
	if s.replay != nil {
		s.env.Replaying = ""
	}
	if s.env.Night == s.engine {
		s.env.Night = nil
	}
}
