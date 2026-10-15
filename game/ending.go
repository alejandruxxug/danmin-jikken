package game

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"danmingame/night"
)

// An ending is four things in a row (Sound -> The ending rule): the ending's
// script (in cinematic mode, the bars up), a hard cut to black and a beat of
// silence, then the rock slams in and the last picture comes back with the
// ending's title written over it, and then the finale (finale.go): the credits
// roll up over the same picture, the logo holds, the screen fades to black and
// the TV turns on onto the top-10, once, and the title screen after it.

type endingPhase int

const (
	phaseScript endingPhase = iota
	phaseSilence
	phaseTitle
	phaseFinale
)

func (p endingPhase) String() string {
	return [...]string{"script", "silence", "title", "finale"}[p]
}

const (
	silenceTicks     = 90  // black and quiet before the rock
	titleTicks       = 360 // the title stays up 6 s
	secretTitleTicks = 600 // the secret ending's stays 10 s, the glitch needs the time
	titleSkipAfter   = 30  // a click skips a replay's title only after half a second
)

// endingScene plays one ending (ending.sleep, ending.wake, ending.forever,
// ending.317).
//
// A gallery replay (replay.sleep ...) plays the same scene: the script, the
// cut, the silence, the rock and the title, glitch and stain included. It
// differs only at the edges: no finale and no top-10 (it goes back to the
// gallery once the title has been up its time, or on a click, or on Esc), and
// no side effects (no sleep_log, no 3:17 quit, nothing saved).
type endingScene struct {
	env    *Env
	name   string
	replay *sceneCard // set when the SCENES gallery replays this ending (see below)
	script *nightScene
	font   *Font
	phase  endingPhase
	ticks  int // ticks into the phase
	cards  []creditCard
	finale *finaleArt    // the roll, the logo and the fade
	over   *ebiten.Image // 3:17: the finale's layer drawn aside, so a row band can slip
	blood  int           // ticks since the blood splashed (Forever); -1 when it has not
	title  *titleArt
	timed  bool // the rock plays from the first line, timed to the title (Forever)
	clock  int  // ticks since the ending began (every phase), for the music's sync
	eyesAt int  // tick the eyes start to close (-1: never), where Forever's music starts to rise
	// the clock tick before which the music is not judged again (it just jumped)
	settleUntil int
	musicStart  time.Duration // where the music was asked to start in its file
	musicMoved  bool          // the sound card has really started playing it
	glitch      *Glitch       // the secret ending's lag and glitch (nil for the others)
	buf         *ebiten.Image // the picture drawn aside, before the glitch tears it
}

func (s *endingScene) Enter(env *Env) error {
	s.env = env
	s.blood = -1
	var err error
	if s.font, err = LoadFont(); err != nil {
		return err
	}
	if s.title, err = loadTitleArt(env.Sprites, s.name); err != nil {
		return err
	}
	if s.name == night.Ending317 {
		if s.glitch, err = NewGlitch(env.Sprites, env.Seed); err != nil {
			return err
		}
	}
	s.cards = CreditCards()
	if s.finale, err = newFinaleArt(env.Sprites, s.name, s.cards); err != nil {
		return err
	}
	env.Ending = s.name
	s.script = &nightScene{script: night.EndingScript(s.name), onEnd: s.scriptEnded, replay: s.replay}
	if err := s.script.Enter(env); err != nil {
		return err
	}
	return s.planMusic()
}

// planMusic works out the ending's fixed timeline (the script plays by itself,
// see `autoplay on`) and, for an ending whose music is timed to it, starts the
// music now. The title pops scriptTicks + silenceTicks after the first line;
// Forever's rock is started so that its drop lands on that very tick.
func (s *endingScene) planMusic() error {
	script, err := night.LoadScript(night.EndingScript(s.name))
	if err != nil {
		return err
	}
	scriptTicks, eyes, err := night.TimelineMarks(script, s.script.engine.Flags(), Username())
	if err != nil {
		return err
	}
	s.env.EndingTitleAt = scriptTicks + silenceTicks
	s.eyesAt = eyes
	drop, timed := rockDrops[s.name]
	s.timed = timed
	if !timed || s.env.Sound == nil {
		return nil
	}
	start, clamped := syncedStart(drop, s.env.EndingTitleAt)
	s.env.EndingRockFrom = start.Seconds()
	s.musicStart = start
	s.settleUntil = syncSettle // the first reading of a fresh stream is noisy: look from half a second on
	if clamped {
		s.env.Sound.Note(fmt.Sprintf("ending %s: the rock file starts too late to reach its drop by tick %d; starting at 0", s.name, s.env.EndingTitleAt))
	}
	s.env.MusicCue = "ending_rock_" + s.name
	s.env.Sound.PlayMusic("ending_rock_"+s.name, false, start)
	s.env.Sound.SetMusicShape(foreverQuiet) // very quiet until the eyes close (ending_sync.go)
	logSync("%s: music asked for at tick 0, from %.3f s in the file; title planned at tick %d, eyes at %d", s.name, start.Seconds(), s.env.EndingTitleAt, eyes)
	return nil
}

// scriptEnded is the ending script's "end card". The secret ending leaves its
// file behind here, once, when the game has "fallen asleep".
func (s *endingScene) scriptEnded(string) error {
	if s.name == night.Ending317 && s.replay == nil && !s.env.Audition {
		path, err := WriteSleepLog(s.env.ScoresPath)
		if err != nil {
			s.env.SleepLogError = err.Error() // the ending still plays; the dump shows why
		}
		s.env.SleepLogPath = path
		s.env.QuitAfterTop10 = true // the game closes itself at the very end
	}
	s.blood = -1
	if t, ok := s.script.engine.FxTicks("splash"); ok {
		s.blood = t // the stain stays over the silence and the title
	}
	s.env.Night = nil
	s.setPhase(phaseSilence)
	return nil
}

func (s *endingScene) setPhase(p endingPhase) {
	s.phase, s.ticks = p, 0
	s.env.EndingPhase = p.String()
	switch p {
	case phaseSilence:
		if s.timed {
			return // Forever: the music plays on through the black, the cut is only visual
		}
		s.env.SetMusic("") // hard cut: nothing plays under the black
		if s.env.Sound != nil {
			s.env.Sound.Preload("ending_rock_" + s.name) // decode now, in the quiet, so the slam is not late
		}
	case phaseTitle:
		if s.timed && s.env.Sound != nil {
			pos, _ := s.env.Sound.MusicPosition()
			logSync("title pops at tick %d (planned %d); music at %.3f s, drop at %.3f s (error %+.1f ms)", s.clock, s.env.EndingTitleAt, pos.Seconds(), rockDrops[s.name].Seconds(), (pos-rockDrops[s.name]).Seconds()*1000)
		}
		s.script.engine.CutToStill() // the last picture comes back (some endings name theirs)
		if s.timed {
			return // the rock has been playing since the first line
		}
		s.env.MusicCue = "ending_rock_" + s.name
		if s.env.Sound != nil { // the rock slams in, mid-riff, no fade-in
			s.env.Sound.PlayMusic("ending_rock_"+s.name, false, rockStarts[s.name])
		}
	}
}

func (s *endingScene) Update() error {
	s.clock++
	defer s.syncMusic()
	if s.env.JumpToFinale { // the backdoor's "finale N": straight to tick N of the finale
		s.env.JumpToFinale = false
		if err := s.jumpToFinale(s.env.FinaleJumpTick); err != nil {
			return err
		}
	}
	if s.phase == phaseScript {
		if s.env.SkipToCard { // the backdoor's "card": jump to the end of the script
			s.env.SkipToCard = false
			return s.scriptEnded("")
		}
		return s.script.Update()
	}
	if s.replay != nil && s.env.Input.KeyPressed(ebiten.KeyEscape) {
		s.env.SwitchTo(SceneScenes) // Esc leaves a replay at any time
		return nil
	}
	s.ticks++
	s.env.EndingGlitch = 0
	if s.glitch != nil && s.phase >= phaseTitle {
		s.env.EndingGlitch = s.glitchLevel()
	}
	if s.blood >= 0 {
		s.blood++
	}
	// The picture goes on living under the title: the boil, the drips.
	if err := s.script.engine.Tick(); err != nil {
		return err
	}
	switch s.phase {
	case phaseSilence:
		if s.ticks >= silenceTicks {
			s.setPhase(phaseTitle)
		}
	case phaseTitle:
		// Only a replay can be clicked past its title: the real ending runs
		// on its own, so Forever's music stays in step with the picture.
		_, _, clicked := s.env.Input.Click()
		skip := s.replay != nil && (clicked || s.env.Input.KeyPressed(ebiten.KeySpace)) && s.ticks > titleSkipAfter
		if s.ticks >= s.titleLength() || skip {
			if s.replay != nil { // a replay ends with the title: no finale, no top-10
				s.env.SwitchTo(SceneScenes)
				return nil
			}
			s.setPhase(phaseFinale)
		}
	case phaseFinale:
		return s.updateFinale()
	}
	return nil
}

// jumpToFinale is the backdoor's way to any tick of the finale: it plays the
// rest of the script away first if it is still running.
func (s *endingScene) jumpToFinale(tick int) error {
	if s.phase == phaseScript {
		if err := s.scriptEnded(""); err != nil {
			return err
		}
	}
	s.script.engine.CutToStill()
	s.setPhase(phaseFinale)
	s.ticks = tick - 1 // this frame's Update counts the tick
	return nil
}

// updateFinale moves the finale one tick on. The music fades out over the
// 3 s before the black; after the black the TV turns on onto the top-10 (an
// audition ends here instead: no top-10, nothing saved).
func (s *endingScene) updateFinale() error {
	st, ok := FinaleAt(s.ticks)
	s.env.EndingStage, s.env.EndingTick = st.Stage, st.At
	if s.ticks == FinaleStageStart("black")-finaleFadeOut && s.env.Sound != nil {
		s.env.Sound.FadeOut(finaleFadeOut)
	}
	if ok {
		return nil
	}
	if s.env.Audition {
		return ebiten.Termination
	}
	s.env.StartPowerOn()
	s.env.SwitchTo(SceneDayTop10)
	return nil
}

func (s *endingScene) Draw(screen *ebiten.Image) {
	screen.Fill(paletteInk)
	switch s.phase {
	case phaseScript:
		s.script.Draw(screen)
		return
	case phaseSilence: // black; Forever's stain stays on the glass
		if s.blood >= 0 {
			s.script.art.DrawBlood(screen, s.blood)
		}
		return
	}
	st, running := FinaleAt(s.ticks)
	if s.phase == phaseFinale && (!running || fadeFrame(st) == 3) {
		return // black and silence, until the TV turns on
	}
	if s.glitch == nil {
		s.drawPicture(screen)
		if s.phase == phaseFinale {
			s.finale.draw(screen, st)
			s.finale.drawFade(screen, st)
		}
		return
	}
	// The secret ending: draw aside, then let the glitch tear it onto the screen.
	if s.buf == nil {
		s.buf = ebiten.NewImage(ScreenW, ScreenH)
	}
	s.buf.Fill(paletteInk)
	s.drawPicture(s.buf)
	s.glitch.Apply(screen, s.buf, s.ticks, s.glitchLevel(), s.env.ReduceFlashing)
	if s.phase != phaseFinale {
		return
	}
	// The roll and the logo stay readable: they are drawn aside as well, and
	// only now and then one row band of them slips sideways.
	if s.over == nil {
		s.over = ebiten.NewImage(ScreenW, ScreenH)
	}
	s.over.Clear()
	s.finale.draw(s.over, st)
	drawSlipped(screen, s.over, s.ticks, s.env.Seed)
	s.finale.drawFade(screen, st)
}

// drawPicture paints the ending's last picture (with its bars) and, in the
// title phase, the title over it.
func (s *endingScene) drawPicture(dst *ebiten.Image) {
	s.script.art.Draw(dst, s.script.engine)
	if s.phase == phaseTitle {
		s.drawTitle(dst)
	}
}

// titleLength is how long the title stays up: the secret ending's glitch
// needs longer to ramp up.
func (s *endingScene) titleLength() int {
	if s.glitch != nil {
		return secretTitleTicks
	}
	return titleTicks
}

// glitchLevel is the secret ending's glitch strength now: it ramps up over
// the title and stays strong under the finale.
func (s *endingScene) glitchLevel() float64 {
	if s.phase == phaseFinale {
		return 0.7
	}
	return glitchLevel(s.ticks)
}

// drawTitle writes the ending's name over the picture: the small label, the
// large title on its shade, and the place line.
func (s *endingScene) drawTitle(dst *ebiten.Image) {
	s.title.draw(dst, s.name)
	small := color.Color(s.title.small)
	s.centered(dst, strings.ToUpper(titleLabels[s.name]), labelCapTop-capTop, small)
	s.centered(dst, strings.ToUpper(endingSubtitle), subtitleTop-capTop, small)
}

// UsesCRT says whether this frame goes through the day's CRT. The Sleep
// ending ends in a day (Alejo, 2026-10-14: "on the endings that are about
// going to sleep (ending up in day) can you make the crt like the day?"):
// from the white of the Driftwell reel on, the picture is the day reception
// desk, so the whole frame (bars, title and finale included, as the day
// scenes' own UI is) is shown on the glass. The other endings stay night.
func (s *endingScene) UsesCRT() bool {
	if s.name != night.EndingSleep || s.script == nil || s.script.engine == nil {
		return false
	}
	switch s.script.engine.BG() {
	case "white", "desk_day":
		return true
	}
	return false
}

func (s *endingScene) centered(screen *ebiten.Image, text string, y int, c color.Color) {
	s.font.Draw(screen, text, (ScreenW-s.font.Width(text))/2, y, c)
}

func (s *endingScene) Exit() { s.script.Exit() }

// endingSceneID is the scene id of an ending.
func endingSceneID(name string) string { return fmt.Sprintf("ending.%s", name) }

// Night palette colours the title and the credits use.
var (
	colPaperRGBA = color.RGBA{0xec, 0xec, 0xe6, 0xff}
	colMidRGBA   = color.RGBA{0x6b, 0x6b, 0x70, 0xff}
)

// endingOf gives the ending a gallery card replays ("" if it is not one).
func endingOf(c sceneCard) string {
	for _, name := range night.Endings {
		if c.Script == night.EndingScript(name) {
			return name
		}
	}
	return ""
}
