package game

import (
	"image/color"

	"danmingame/night"
)

// The ending finale (Art log: Ending finale storyboard). After the ending's
// title has held, the credits roll up the screen over the ending's last
// picture on a dither shade, THANK YOU FOR STAYING AWAKE stops in the middle,
// the game's logo cuts in and holds, the screen fades to black through a
// dither, and the TV turns on onto the top-10. This file is the timing and
// the layout, all pure functions of a tick; finale_draw.go paints them.

// The stages, in order, and their lengths in ticks (60 per second). The
// numbers are the spec's seconds: the shade steps in 0.25 s at a time, the
// roll runs 33.4 s, the thanks holds 3 s, the logo 5 s.
var finaleStages = []struct {
	Name  string
	Ticks int
}{
	{"shade-in", 30}, // the dither column steps in (and the roll starts to climb)
	{"roll", rollTicks(CreditCards(night.EndingForever))}, // the credits climb 16 px/s until THANK YOU is at its resting y (derived from the layout)
	{"thanks", 180}, // THANK YOU FOR STAYING AWAKE holds in the middle
	{"shade-out", 30},
	{"logo", 300}, // the logo holds, still
	{"fade", 90},  // three dither steps to black, 0.5 s each
	{"black", 60}, // black and silence; the TV turns on after it
}

// Layout numbers, in pixels (caps top of a line is the y we speak of).
const (
	rollStartY     = 270 // the first line enters at the bottom edge
	rollPxPerSec   = 16
	thanksStopY    = 130 // the thanks line stops with its caps top here
	rollHeadGap    = 10  // heading to its first line, and line to line
	rollCardGap    = 24  // the last line of a card to the next heading
	rollThanksGap  = 50  // the last line to the thanks
	shadeStepTicks = 15  // the shade (and the logo's fade in) change every 0.25 s
	fadeStepTicks  = 30  // each dither of the fade lasts 0.5 s
	logoDY         = 52  // the title sheet's logo is drawn this far down
	finaleFadeOut  = 180 // the music fades out over the 3 s before the black
)

// FinaleState is where the finale is at one tick.
type FinaleState struct {
	Stage string
	Tick  int // ticks into the stage
	At    int // ticks into the whole finale
}

// FinaleTicks is the length of the whole finale, up to the turn-on.
func FinaleTicks() int {
	total := 0
	for _, s := range finaleStages {
		total += s.Ticks
	}
	return total
}

// FinaleStageStart is the tick where a stage begins (-1 if there is none by
// that name).
func FinaleStageStart(name string) int {
	start := 0
	for _, s := range finaleStages {
		if s.Name == name {
			return start
		}
		start += s.Ticks
	}
	return -1
}

// FinaleAt gives the stage at a tick; ok is false once the finale is over.
func FinaleAt(tick int) (st FinaleState, ok bool) {
	if tick < 0 {
		return FinaleState{}, false
	}
	left := tick
	for _, s := range finaleStages {
		if left < s.Ticks {
			return FinaleState{Stage: s.Name, Tick: left, At: tick}, true
		}
		left -= s.Ticks
	}
	return FinaleState{}, false
}

// shadeFrame is the frame of the shade sheet that shows at a point of the
// finale (-1: no shade). max is the strongest frame the ending uses: 1 (50 %)
// for Forever and 3:17, which keep their picture showing through, 2 (solid)
// for Sleep and Wake. It steps in one frame per 0.25 s and out the same way,
// skipping the solid frame on the way out so the logo can cut in on time.
func shadeFrame(st FinaleState, max int) int {
	switch st.Stage {
	case "shade-in":
		return min(st.Tick/shadeStepTicks, max)
	case "roll", "thanks":
		return max
	case "shade-out":
		if f := max - 1 - st.Tick/shadeStepTicks; f >= 0 {
			return f
		}
	}
	return -1
}

// fadeFrame is the frame of fade_black over everything: -1 before the fade,
// 0 to 2 while it steps (25, 50, 75 %), 3 for solid black.
func fadeFrame(st FinaleState) int {
	switch st.Stage {
	case "fade":
		return st.Tick / fadeStepTicks
	case "black":
		return 3
	}
	return -1
}

// rollDY is how many pixels the roll has climbed at a tick of the finale. It
// climbs from the end of the shade-in (the roll's first line is then at the
// bottom edge) and stops when the thanks line has arrived.
func rollDY(tick int) int {
	climbing := tick - finaleStages[0].Ticks
	return min(max(climbing, 0), finaleStages[1].Ticks) * rollPxPerSec / 60
}

// rollTicks is how long the roll climbs for these cards: until the thanks line
// has reached thanksStopY, rounded up to a whole tick. The stage table uses it,
// so a change to the credits can never leave the two out of step. The four
// endings' credits have the same layout, so one length serves all of them
// (TestRollIsTheSameForEveryEnding).
func rollTicks(cards []creditCard) int {
	pixels := rollStartY + thanksY(rollLayout(cards)) - thanksStopY
	return (pixels*60 + rollPxPerSec - 1) / rollPxPerSec
}

// rollLine is one line of the roll: its text, its caps-top y counted from
// the first line (which is 0), and what kind of line it is.
type rollLine struct {
	Text    string
	Y       int
	Heading bool
	Thanks  bool // the closing line, drawn large
}

// rollLayout turns the credit cards into the roll. Cards that share a heading
// (the music lines) sit under one heading. The last card, if it is only a
// heading, is the thanks line.
func rollLayout(cards []creditCard) []rollLine {
	var lines []rollLine
	y, heading := 0, ""
	for i, c := range cards {
		if i > 0 && i == len(cards)-1 && c.Line == "" { // the thanks
			lines = append(lines, rollLine{Text: c.Heading, Y: y + rollThanksGap, Thanks: true})
			break
		}
		switch {
		case i == 0:
			lines = append(lines, rollLine{Text: c.Heading, Y: 0, Heading: true})
		case c.Heading != heading:
			y += rollCardGap
			lines = append(lines, rollLine{Text: c.Heading, Y: y, Heading: true})
		}
		heading = c.Heading
		if c.Line != "" {
			y += rollHeadGap // heading to line, and line to line
			lines = append(lines, rollLine{Text: c.Line, Y: y})
		}
	}
	return lines
}

// thanksY is where the thanks line sits in the roll's own coordinates.
func thanksY(lines []rollLine) int {
	for _, l := range lines {
		if l.Thanks {
			return l.Y
		}
	}
	return 0
}

// finaleStyle is how one ending dresses its finale.
type finaleStyle struct {
	ShadeLayer  string // finale sheet layer under the roll
	ShadeMax    int    // the strongest shade frame (see shadeFrame)
	Head, Name  color.RGBA
	Thanks      color.RGBA
	HeadOutline bool // headings get a 1 px black outline
	HeadShadow  bool // headings get a deep-sage shadow at (+1, +1)
	ThanksEdge  bool // the thanks gets a black outline
	Logo        string
}

// Logo looks: the light logo with the blood shadow (Forever), with the dark
// grey shadow (3:17), or the title's own black logo (Sleep, Wake).
const (
	logoBlood = "blood"
	logoDark  = "dark"
	logoTitle = "title"
)

// finaleStyleFor gives an ending's look (Art log: per ending colours).
// Forever and 3:17 print red headings with a black outline on a half shade;
// Wake prints light grey on a solid black one; Sleep prints black on a solid
// off-white one, the headings with a deep-sage shadow.
func finaleStyleFor(ending string) finaleStyle {
	switch ending {
	case night.EndingSleep:
		return finaleStyle{ShadeLayer: "shade_light", ShadeMax: 2, Head: colInk, Name: colInk, Thanks: colInk, HeadShadow: true, Logo: logoTitle}
	case night.EndingWake:
		return finaleStyle{ShadeLayer: "shade_dark", ShadeMax: 2, Head: colLightGrey, Name: colOffWhite, Thanks: colOffWhite, Logo: logoTitle}
	case night.Ending317:
		return finaleStyle{ShadeLayer: "shade_dark", ShadeMax: 1, Head: colBrightRed, Name: colOffWhite, Thanks: colOffWhite, HeadOutline: true, ThanksEdge: true, Logo: logoDark}
	}
	return finaleStyle{ShadeLayer: "shade_dark", ShadeMax: 1, Head: colBrightRed, Name: colOffWhite, Thanks: colOffWhite, HeadOutline: true, ThanksEdge: true, Logo: logoBlood}
}
