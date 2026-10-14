package game

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"danmingame/night"
)

// An ending is four things in a row (Sound -> The ending rule): the ending's
// script, a hard cut to black and a beat of silence, the ending card with the
// loud rock slamming in, and the credits rolling under the same track. Then
// the top-10, once, and the title.

type endingPhase int

const (
	phaseScript endingPhase = iota
	phaseSilence
	phaseCard
	phaseCredits
)

func (p endingPhase) String() string {
	return [...]string{"script", "silence", "card", "credits"}[p]
}

const (
	silenceTicks     = 90  // black and quiet before the rock
	cardTicks        = 360 // the card stays up 6 s
	secretCardTicks  = 600 // the secret ending's card stays 10 s, the glitch needs the time
	creditsSkipAfter = 30  // a click skips a phase only after half a second
	creditsPixelsPer = 0.3 // the credits scroll this many pixels a tick
)

// endingSubtitle is the line under every card (the one in the Sound note).
const endingSubtitle = "Halcyon Rest, Ward B - 3:17 AM"

// endingScene plays one ending (ending.sleep, ending.wake, ending.forever,
// ending.317).
type endingScene struct {
	env     *Env
	name    string
	script  *nightScene
	font    *Font
	phase   endingPhase
	ticks   int // ticks into the phase
	credits []string
	blood   int // ticks since the blood splashed (Forever); -1 when it has not
	card    *cardArt
	glitch  *Glitch       // the secret ending's lag and glitch (nil for the others)
	buf     *ebiten.Image // the card drawn aside, before the glitch tears it
}

func (s *endingScene) Enter(env *Env) error {
	s.env = env
	s.blood = -1
	var err error
	if s.font, err = LoadFont(); err != nil {
		return err
	}
	if s.card, err = loadCardArt(env.Sprites, s.name); err != nil {
		return err
	}
	if s.name == night.Ending317 {
		if s.glitch, err = NewGlitch(env.Sprites, env.Seed); err != nil {
			return err
		}
	}
	s.credits = CreditLines()
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
		s.blood = t // the stain stays over the silence and the card
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
	case phaseCard:
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
	if s.glitch != nil && s.phase >= phaseCard {
		s.env.EndingGlitch = s.glitchLevel()
	}
	if s.blood >= 0 {
		s.blood++
	}
	skip := false
	if _, _, clicked := s.env.Input.Click(); clicked || s.env.Input.KeyPressed(ebiten.KeySpace) {
		skip = s.ticks > creditsSkipAfter
	}
	switch s.phase {
	case phaseSilence:
		if s.ticks >= silenceTicks {
			s.setPhase(phaseCard)
		}
	case phaseCard:
		if s.ticks >= s.cardLength() || skip {
			s.setPhase(phaseCredits)
		}
	case phaseCredits:
		if s.ticks == creditsTotalTicks(len(s.credits))-creditsFadeTicks && s.env.Sound != nil {
			s.env.Sound.FadeOut(creditsFadeTicks) // the rock fades in the last seconds
		}
		if s.creditsDone() || skip {
			s.env.SwitchTo(SceneDayTop10)
		}
	}
	return nil
}

// creditsY is the y of the first credits line at a tick: the block starts
// below the screen and rolls up.
func creditsY(ticks int) int { return ScreenH - int(float64(ticks)*creditsPixelsPer) }

const creditsLineStep = 12

// creditsTotalTicks is the tick of the credits phase when the last line has
// scrolled off the top.
func creditsTotalTicks(lines int) int {
	return int(float64(ScreenH+lines*creditsLineStep)/creditsPixelsPer) + 1
}

func (s *endingScene) creditsDone() bool {
	return creditsY(s.ticks)+len(s.credits)*creditsLineStep < 0
}

func (s *endingScene) Draw(screen *ebiten.Image) {
	defer s.drawBlood(screen) // on top of the card
	screen.Fill(paletteInk)
	if s.phase == phaseScript {
		s.script.Draw(screen)
		return
	}
	draw := s.drawCard
	if s.phase == phaseCredits {
		draw = s.drawCredits
	}
	if s.glitch == nil || s.phase == phaseSilence {
		if s.phase != phaseSilence {
			draw(screen)
		}
		return
	}
	// The secret ending: draw aside, then let the glitch tear it onto the screen.
	if s.buf == nil {
		s.buf = ebiten.NewImage(ScreenW, ScreenH)
	}
	s.buf.Fill(paletteInk)
	draw(s.buf)
	s.glitch.Apply(screen, s.buf, s.ticks, s.glitchLevel(), s.env.ReduceFlashing)
}

// cardLength is how long the card stays up: the secret ending's glitch needs
// longer to ramp up.
func (s *endingScene) cardLength() int {
	if s.glitch != nil {
		return secretCardTicks
	}
	return cardTicks
}

// glitchLevel is the secret ending's glitch strength now: it ramps up over
// the card and stays strong under the credits.
func (s *endingScene) glitchLevel() float64 {
	if s.phase == phaseCredits {
		return 0.7
	}
	return glitchLevel(s.ticks)
}

// drawCard paints the ending card.
func (s *endingScene) drawCard(dst *ebiten.Image) {
	shown := s.ticks
	if s.glitch != nil {
		shown = lagTick(s.ticks, s.env.Seed, s.glitchLevel())
	}
	s.card.draw(dst, s.name, shown)
	label, subtitle := cardSmall(s.name)
	s.centered(dst, label, cardLabelY, colMidRGBA)
	s.centered(dst, subtitle, cardSubY, colMidRGBA)
}

// drawCredits paints the rolling credits.
func (s *endingScene) drawCredits(dst *ebiten.Image) {
	y := creditsY(s.ticks)
	for i, line := range s.credits {
		c := color.Color(colMidRGBA)
		if strings.HasPrefix(line, "#") {
			line, c = strings.TrimPrefix(line, "# "), colPaperRGBA
		}
		if ly := y + i*creditsLineStep; ly > -creditsLineStep && ly < ScreenH {
			s.centered(dst, line, ly, c)
		}
	}
}

// drawBlood puts the stain over whatever the phase drew.
func (s *endingScene) drawBlood(screen *ebiten.Image) {
	if s.phase != phaseScript && s.blood >= 0 {
		s.script.art.DrawBlood(screen, s.blood)
	}
}

func (s *endingScene) centered(screen *ebiten.Image, text string, y int, c color.Color) {
	s.font.Draw(screen, text, (ScreenW-s.font.Width(text))/2, y, c)
}

func (s *endingScene) Exit() { s.script.Exit() }

// endingSceneID is the scene id of an ending.
func endingSceneID(name string) string { return fmt.Sprintf("ending.%s", name) }

// Night palette colours the card and the credits use.
var (
	colPaperRGBA = color.RGBA{0xec, 0xec, 0xe6, 0xff}
	colMidRGBA   = color.RGBA{0x6b, 0x6b, 0x70, 0xff}
)
