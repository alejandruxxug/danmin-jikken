package game

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"danmingame/night"
)

// An ending is four things in a row (Sound -> The ending rule): the ending's
// script (in cinematic mode, the bars up), a hard cut to black and a beat of
// silence, then the rock slams in and the last picture comes back with the
// ending's title written over it, and the credits run in the bottom bar over
// the same picture. Then the top-10, once, and the title screen.

type endingPhase int

const (
	phaseScript endingPhase = iota
	phaseSilence
	phaseTitle
	phaseCredits
)

func (p endingPhase) String() string {
	return [...]string{"script", "silence", "title", "credits"}[p]
}

const (
	silenceTicks      = 90  // black and quiet before the rock
	titleTicks        = 360 // the title stays up 6 s
	secretTitleTicks  = 600 // the secret ending's stays 10 s, the glitch needs the time
	creditsSkipAfter  = 30  // a click skips a phase only after half a second
	creditsCardHeadY  = 245 // caps top of a card's heading in the bottom bar
	creditsCardLineY  = 256 // ... and of its line
	creditsCardAloneY = 250 // a card with only a heading
)

// endingScene plays one ending (ending.sleep, ending.wake, ending.forever,
// ending.317).
type endingScene struct {
	env    *Env
	name   string
	script *nightScene
	font   *Font
	phase  endingPhase
	ticks  int // ticks into the phase
	cards  []creditCard
	blood  int // ticks since the blood splashed (Forever); -1 when it has not
	title  *titleArt
	glitch *Glitch       // the secret ending's lag and glitch (nil for the others)
	buf    *ebiten.Image // the picture drawn aside, before the glitch tears it
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
	env.Ending = s.name
	s.script = &nightScene{script: night.EndingScript(s.name), onEnd: s.scriptEnded}
	return s.script.Enter(env)
}

// scriptEnded is the ending script's "end card". The secret ending leaves its
// file behind here, once, when the game has "fallen asleep".
func (s *endingScene) scriptEnded(string) error {
	if s.name == night.Ending317 {
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
		s.env.SetMusic("") // hard cut: nothing plays under the black
		if s.env.Sound != nil {
			s.env.Sound.Preload("ending_rock_" + s.name) // decode now, in the quiet, so the slam is not late
		}
	case phaseTitle:
		s.script.engine.CutToStill() // the last picture comes back (some endings name theirs)
		s.env.MusicCue = "ending_rock_" + s.name
		if s.env.Sound != nil { // the rock slams in, mid-riff, no fade-in
			s.env.Sound.PlayMusic("ending_rock_"+s.name, false, rockStarts[s.name])
		}
	}
}

func (s *endingScene) Update() error {
	if s.phase == phaseScript {
		if s.env.SkipToCard { // the backdoor's "card": jump to the end of the script
			s.env.SkipToCard = false
			return s.scriptEnded("")
		}
		return s.script.Update()
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
	skip := false
	if _, _, clicked := s.env.Input.Click(); clicked || s.env.Input.KeyPressed(ebiten.KeySpace) {
		skip = s.ticks > creditsSkipAfter
	}
	switch s.phase {
	case phaseSilence:
		if s.ticks >= silenceTicks {
			s.setPhase(phaseTitle)
		}
	case phaseTitle:
		if s.ticks >= s.titleLength() || skip {
			s.setPhase(phaseCredits)
		}
	case phaseCredits:
		if s.ticks == CreditsTicks(s.cards)-creditsFadeTicks && s.env.Sound != nil {
			s.env.Sound.FadeOut(creditsFadeTicks) // the rock fades in the last seconds
		}
		if s.creditsDone() || skip {
			s.env.SwitchTo(SceneDayTop10)
		}
	}
	return nil
}

// creditsDone says whether the last card has gone.
func (s *endingScene) creditsDone() bool { return s.ticks >= CreditsTicks(s.cards) }

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
	if s.glitch == nil {
		s.drawPicture(screen)
		return
	}
	// The secret ending: draw aside, then let the glitch tear it onto the screen.
	if s.buf == nil {
		s.buf = ebiten.NewImage(ScreenW, ScreenH)
	}
	s.buf.Fill(paletteInk)
	s.drawPicture(s.buf)
	s.glitch.Apply(screen, s.buf, s.ticks, s.glitchLevel(), s.env.ReduceFlashing)
}

// drawPicture paints the ending's last picture (with its bars) and then the
// title over it, or the credit card in the bottom bar.
func (s *endingScene) drawPicture(dst *ebiten.Image) {
	s.script.art.Draw(dst, s.script.engine)
	if s.phase == phaseTitle {
		s.drawTitle(dst)
		return
	}
	s.drawCard(dst)
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
// the title and stays strong under the credits.
func (s *endingScene) glitchLevel() float64 {
	if s.phase == phaseCredits {
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

// drawCard writes the credit card that is up in the bottom bar. The white
// bars of the Sleep ending take ink and deep sage, the others off-white and
// mid grey. A glitched ending shows its cards a few frames late.
func (s *endingScene) drawCard(dst *ebiten.Image) {
	ticks := s.ticks
	if s.glitch != nil {
		ticks = lagTick(ticks, s.env.Seed, s.glitchLevel())
	}
	i, showing := CardAt(s.cards, ticks)
	if !showing {
		return
	}
	card := s.cards[i]
	head, line := color.Color(colMidRGBA), color.Color(colPaperRGBA)
	if s.name == night.EndingSleep {
		head, line = colDim, colInk
	}
	if card.Line == "" {
		s.centered(dst, strings.ToUpper(card.Heading), creditsCardAloneY-capTop, head)
		return
	}
	s.centered(dst, strings.ToUpper(card.Heading), creditsCardHeadY-capTop, head)
	s.centered(dst, card.Line, creditsCardLineY-capTop, line)
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
