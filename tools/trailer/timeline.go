package main

// The trailer's timeline: for every frame (0 to 4019, 60 per second) the list
// of layers that make it, following the motion pass in the Trailer note. It
// is plain arithmetic on the frame number; nothing here opens a picture.
//
// Frame numbers in this file are trailer frames counted from 0. The recorded
// shots are numbered from 1, so "the stamp shot's frame f-700" is source frame
// number f-700 of the shot called stamp.

import (
	"fmt"
	"image"
)

// TotalFrames is the length of the trailer: 67 seconds.
const TotalFrames = 4020

// The names of the recorded shots (shots.txt).
const (
	shotWake      = "reception_wake"
	shotAsleep    = "reception_asleep"
	shotOrient    = "orientation"
	shotStamp     = "stamp"
	shotComputer  = "computer"
	shotSleepers  = "sleepers"
	shotMemo      = "memo"
	shotStrobe    = "strobe"
	shotStrobeRed = "strobe_reduced"
	shotCoffee    = "coffee"
	shotTired     = "tired"
	shotGap       = "gap"
	shotKana      = "kanashibari"
	shotFlashRed  = "flash_red"
	shotFlashHana = "flash_hana"
	shotTitle     = "title"
)

// Where the trailer's shots start, in trailer frames.
const (
	t02Start = 234 // the whip: T02 starts coming in as T01 leaves
	t03Start = 496
	t04Start = 596
	t05Start = 721
	t06Start = 846
	t07Start = 946
	t08Start = 1096
	t09Start = 1196
	t10Start = 1346
	t10bFrom = 1560
	t11Start = 1680
	t12Start = 1800
	t13Start = 1950
	t14Start = 2190
	t15Start = 2370
	t16Start = 2550
	t17Start = 2820
	t18Start = 3060
	t19Start = 3240
	t20Start = 3600
	t20Black = 3990

	slamFrame = 3261 // the dagger lands on the title and Ironbound drops
)

// Source frames of the shots that do not start at their first frame: the
// trailer frame in the shot's first used frame minus its own number gives the
// offset (source = trailer frame - offset).
const (
	wakeOffset     = t02Start - 1  // trailer frame 234 is frame 1 of the pan
	wakeLeftFrom   = 92            // the awake half of the split starts here
	asleepFrom     = 230           // the asleep half starts here
	orientOffset   = t04Start - 61 // T04 shows frames 61 to 185
	stampOffset    = t05Start - 21 // T05 shows frames 21 to 145: the slam (frame 96) is on trailer frame 796
	computerOffset = t06Start - 1  // T06 shows frames 1 to 100
	sleepersT08    = 702           // T08 shows frames 702 to 801
	sleepersT09    = 802           // T09 shows frames 802 to 951
	memoFrom       = 31            // T10 shows frames 31 to 244
	strobeOffset   = t11Start - 81 // T11 shows frames 81 to 200
	coffeeOffset   = t13Start - 1  // T13 shows frames 1 to 240
	tiredOffset    = t14Start - 1  // T14 shows frames 1 to 180
	gapOffset      = t15Start - 1  // T15 shows frames 1 to 180
	kanaOffset     = t16Start - 1  // T16 shows frames 1 to 270
	titleFrom      = 9             // T19 starts at the title's frame 9, so its frame 30 (the dagger) is on slamFrame
	titleSlamFrame = 30            // the title's own frame where the dagger lands
	titleOffset    = t19Start - titleFrom
)

// The letter edges of "WE NEVER CLOSE." in kt_never_close_text, from the
// art notes (x counted from the text's left edge at x 121; spaces count as
// letters, so there are 15).
var neverCloseEdges = []int{21, 36, 48, 69, 84, 105, 120, 138, 150, 168, 183, 201, 219, 234, 243}

const (
	neverCloseX = 121 // where the typed text starts
	neverCloseY = 124 // its top row
	cursorW     = 8
	cursorH     = 15
)

// Timeline holds what the plan needs to know about the sprites (found by
// looking at their pixels, see ink.go) and which strobe to use.
type Timeline struct {
	Reduced bool // the backup file: the -reduce-flashing strobe

	MoonBox, SealBox image.Rectangle // parts of kt_hiring_bg that move on their own

	NightText  image.Rectangle // the THE NIGHT SHIFT row of card_night_shift
	NightEdges []int
	DontText   image.Rectangle // the DON'T FALL ASLEEP. row of card_dont_sleep
	DontEdges  []int
	EndLogo    image.Rectangle // card_end: the logo (romaji and kanji)
	EndMade    image.Rectangle // 2026 - MADE IN GO
	MadeEdges  []int
	EndKanji   image.Rectangle // the small 眠らないで
}

// Searching areas for the two parts of the hiring card that move by
// themselves: the moon with its z's (it drops in), the 求人 seal (it pops).
var (
	moonArea = image.Rect(100, 20, 380, 120)
	sealArea = image.Rect(400, 20, 450, 100)
)

// NewTimeline reads what it needs from the two sheets.
func NewTimeline(fx, cards *Sheet, reduced bool) (*Timeline, error) {
	tl := &Timeline{Reduced: reduced}
	bg, err := fx.Layer("kt_hiring_bg", 0)
	if err != nil {
		return nil, err
	}
	paper := [3]uint8{colPaper.R, colPaper.G, colPaper.B}
	if tl.MoonBox = InkBounds(bg, moonArea, paper); tl.MoonBox.Empty() {
		return nil, fmt.Errorf("kt_hiring_bg: no moon found in %v", moonArea)
	}
	if tl.SealBox = InkBounds(bg, sealArea, paper); tl.SealBox.Empty() {
		return nil, fmt.Errorf("kt_hiring_bg: no seal found in %v", sealArea)
	}
	// The lines of text sit in the last row band of their card, except the end card.
	band := func(layer string, i int) (*image.RGBA, image.Rectangle, error) {
		im, err := cards.Layer(layer, 0)
		if err != nil {
			return nil, image.Rectangle{}, err
		}
		bands := InkBands(im)
		if i >= len(bands) {
			return nil, image.Rectangle{}, fmt.Errorf("%s: %d bands of ink, wanted one at %d", layer, len(bands), i)
		}
		return im, bands[i], nil
	}
	im, r, err := band("card_night_shift", 1)
	if err != nil {
		return nil, err
	}
	tl.NightText, tl.NightEdges = r, LetterEdges(im, r)
	if im, r, err = band("card_dont_sleep", 1); err != nil {
		return nil, err
	}
	tl.DontText, tl.DontEdges = r, LetterEdges(im, r)
	end, err := cards.Layer("card_end", 0)
	if err != nil {
		return nil, err
	}
	bands := InkBands(end)
	if len(bands) != 4 {
		return nil, fmt.Errorf("card_end: %d bands of ink, wanted 4 (romaji, kanji, 2026 line, small kanji)", len(bands))
	}
	tl.EndLogo = bands[0].Union(bands[1])
	tl.EndMade, tl.EndKanji = bands[2], bands[3]
	tl.MadeEdges = LetterEdges(end, tl.EndMade)
	return tl, nil
}

// Plan lists the layers of trailer frame f, first to last.
func (tl *Timeline) Plan(f int) []Layer {
	switch {
	case f < t02Start:
		return tl.hiring(f, 0)
	case f < t02Start+12: // the whip
		push := Move{From: 0, To: -480, Start: t02Start, Dur: 12, Curve: inCubic}.At(f)
		layers := tl.reception(f, push+480)
		layers = append(layers, tl.hiring(f, push)...)
		return append(layers, spriteLayer(SheetFX, "sweep", 0, SpriteOp{DY: GameH/2 - 8}))
	case f < t03Start:
		return tl.reception(f, 0)
	case f < t04Start:
		return tl.neverClose(f)
	case f < t05Start:
		return tl.orientation(f)
	case f < t06Start:
		return tl.stamp(f)
	case f < t07Start:
		return tl.computer(f)
	case f < t08Start:
		return tl.split(f)
	case f < t09Start:
		return tl.sleepers(f)
	case f < t10Start:
		return tl.noRoom(f)
	case f < t10bFrom:
		return tl.memo(f)
	case f < t11Start:
		return tl.clock(f)
	case f < t12Start:
		return tl.strobe(f)
	case f < t13Start:
		return tl.nightCard(f)
	case f < t14Start:
		return tl.coffee(f)
	case f < t15Start:
		return tl.tired(f)
	case f < t16Start:
		return tl.supply(f)
	case f < t17Start:
		return tl.kanashibari(f)
	case f < t18Start:
		return tl.flashes(f)
	case f < t19Start:
		return tl.dontSleep(f)
	case f < t20Start:
		return tl.title(f)
	case f < TotalFrames:
		return tl.endCard(f)
	}
	return nil
}

// ---- T01: the hiring card ----

// The hiring card slides off to the left by dx pixels during the whip; every
// part of it, and the shake, move together.
func (tl *Timeline) hiring(f, dx int) []Layer {
	sx := ShakeX(f, 146) + dx
	moonY := Move{From: -40, To: 0, Start: 0, Dur: 18, Curve: outBack}.At(f)
	sealY := 0
	if f == 96 || f == 97 {
		sealY = -3
	}
	layers := []Layer{
		fillLayer(fullScreen.Add(image.Pt(dx, 0)), colPaper),
		spriteLayer(SheetFX, "kt_hiring_bg", 0, SpriteOp{DX: sx, Exclude: tl.MoonBox.Union(tl.SealBox)}),
		spriteLayer(SheetFX, "kt_hiring_bg", 0, SpriteOp{DX: sx, DY: moonY, Src: tl.MoonBox}),
		spriteLayer(SheetFX, "kt_hiring_bg", 0, SpriteOp{DX: sx, DY: sealY, Src: tl.SealBox}),
	}
	if f >= Beat(0) {
		slide := Move{From: 96, To: 0, Start: Beat(0), Dur: 12, Curve: outCubic}.At(f)
		layers = append(layers, spriteLayer(SheetFX, "kt_hiring_line1", 0, SpriteOp{DX: sx + slide}))
	}
	if dy, ok := Slam(f, Beat(4)-4); ok {
		layers = append(layers, spriteLayer(SheetFX, "kt_hiring_stamp", 0, SpriteOp{DX: sx, DY: dy}))
	}
	return layers
}

// ---- T02: the reception ----

// reception is the game's reception, moved dx pixels sideways (the whip), with
// the 24-hour lower third.
func (tl *Timeline) reception(f, dx int) []Layer {
	layers := []Layer{gameLayer(shotWake, f-wakeOffset, GameOp{Zoom: 1, DX: dx})}
	return append(layers, lowerThird("lt_24hour", f, Beat(12), Beat(17))...)
}

// lowerThird slides a lower third in from the left (outBack, 10 frames) on
// frame in and out again (inCubic, 8 frames) on frame out.
func lowerThird(name string, f, in, out int) []Layer {
	if f < in {
		return nil
	}
	x := Move{From: -300, To: 0, Start: in, Dur: 10, Curve: outBack}.At(f)
	if f >= out {
		x = Move{From: 0, To: -300, Start: out, Dur: 8, Curve: inCubic}.At(f)
	}
	return []Layer{spriteLayer(SheetFX, name, 0, SpriteOp{DX: x})}
}

// ---- T03: WE NEVER CLOSE. ----

func (tl *Timeline) neverClose(f int) []Layer {
	pop := 0
	if f < t03Start+2 {
		pop = 3 // the card pops in 3 pixels low for 2 frames
	}
	n := Typed(f, t03Start+4, 4, len(neverCloseEdges))
	layers := []Layer{
		fillLayer(fullScreen, colPaper),
		spriteLayer(SheetFX, "kt_never_close_bg", 0, SpriteOp{DY: pop}),
	}
	edge := 0
	if n > 0 {
		edge = neverCloseX + neverCloseEdges[n-1]
		layers = append(layers, spriteLayer(SheetFX, "kt_never_close_text", 0,
			SpriteOp{DY: pop, Clip: image.Rect(0, 0, edge, GameH)}))
	}
	if n > 0 && cursorOn(f, t03Start+4+4*(len(neverCloseEdges)-1)) {
		x := edge + 2
		layers = append(layers, fillLayer(image.Rect(x, neverCloseY+pop, x+cursorW, neverCloseY+cursorH+pop), colBlack))
	}
	return layers
}

// cursorOn says whether the typing cursor shows: always while the text is
// being typed, then it blinks 15 frames on, 15 off from the frame typing ends.
func cursorOn(f, typedAt int) bool {
	if f <= typedAt {
		return true
	}
	return (f-typedAt-1)/15%2 == 1
}

// ---- T04, T05, T06: the desk, the stamp, the computer ----

func (tl *Timeline) orientation(f int) []Layer {
	op := GameOp{Zoom: 1}
	if f >= Beat(24) && f < Beat(26) { // punch x2 on the hint
		op = GameOp{Zoom: 2, CX: 255, CY: 230}
	}
	layers := []Layer{gameLayer(shotOrient, f-orientOffset, op)}
	if i := (f - t04Start) / 2; i < 3 { // the channel change: 3 pictures, 2 frames each
		layers = append(layers, spriteLayer(SheetFX, "channel_static", i, SpriteOp{}))
	}
	return layers
}

func (tl *Timeline) stamp(f int) []Layer {
	op := GameOp{Zoom: 1}
	if f >= Beat(30) && f < Beat(30)+12 { // punch x2 on the stamp
		op = GameOp{Zoom: 2, CX: 222, CY: 200}
	}
	return []Layer{gameLayer(shotStamp, f-stampOffset, op)}
}

func (tl *Timeline) computer(f int) []Layer {
	layers := []Layer{gameLayer(shotComputer, f-computerOffset, GameOp{Zoom: 1})}
	return append(layers, lowerThird("lt_trusted", f, Beat(34), t06Start+92)...)
}

// ---- T07: the split screen ----

// split shows Cecily awake on the left and asleep on the right, each half a
// 240-pixel-wide crop of the reception. At f1046 the divider runs to the left
// edge and the asleep side widens to the whole picture.
func (tl *Timeline) split(f int) []Layer {
	d := GameW/2 + Move{From: 0, To: -GameW / 2, Start: Beat(40), Dur: 12, Curve: inOutCubic}.At(f)
	var layers []Layer
	if d > 0 { // the awake half is [0, d), its picture centred in it
		layers = append(layers, gameLayer(shotWake, f-t07Start+wakeLeftFrom,
			GameOp{Zoom: 1, DX: d/2 - GameW/2, Clip: image.Rect(0, 0, d, GameH)}))
	}
	layers = append(layers, gameLayer(shotAsleep, f-t07Start+asleepFrom,
		GameOp{Zoom: 1, DX: d / 2, Clip: image.Rect(d, 0, GameW, GameH)}))
	if d > 0 {
		drop := Move{From: -GameH, To: 0, Start: t07Start, Dur: 8, Curve: outCubic}.At(f)
		layers = append(layers, spriteLayer(SheetFX, "split_divider", 0, SpriteOp{DX: d - GameW/2, DY: drop}))
	}
	return layers
}

// ---- T08, T09: the sleepers and NO ROOM ----

// bedsPlate is the BEDS 0/3 .. 3/3 plate: it drops in on frame t08Start, and
// its count goes up on beats 42 to 45, each time with a small bump.
func bedsPlate(f, dx int) Layer {
	y := Move{From: -60, To: 0, Start: t08Start, Dur: 8, Curve: outBack}.At(f)
	count := 0
	for k := 43; k <= 45; k++ {
		if f >= Beat(k) {
			count++
			if f < Beat(k)+2 {
				y -= 2
			}
		}
	}
	return spriteLayer(SheetFX, "hud_beds", count, SpriteOp{DX: dx, DY: y})
}

func (tl *Timeline) sleepers(f int) []Layer {
	return []Layer{
		gameLayer(shotSleepers, f-t08Start+sleepersT08, GameOp{Zoom: 1}),
		bedsPlate(f, 0),
	}
}

func (tl *Timeline) noRoom(f int) []Layer {
	landing := Beat(48)
	sx := ShakeX(f, landing)
	layers := []Layer{
		gameLayer(shotSleepers, f-t09Start+sleepersT09, GameOp{Zoom: 1, DX: sx}),
		bedsPlate(f, sx),
	}
	if dy, ok := Slam(f, landing-4); ok {
		// Over the floor by the sleeper in the hallway.
		layers = append(layers, spriteLayer(SheetFX, "stamp_no_room", 0, SpriteOp{DX: 90 + sx, DY: 70 + dy}))
	}
	return layers
}

// ---- T10, T10b: the memo and the clock ----

// memoLast is the last frame of the memo shot.
const memoLast = 244

func (tl *Timeline) memo(f int) []Layer {
	op := GameOp{Zoom: 1}
	switch {
	case f >= t10Start+120: // x3 from f1466
		op = GameOp{Zoom: 3, CX: 408, CY: 60}
	case f >= t10Start+60: // x2 from f1406
		op = GameOp{Zoom: 2, CX: 408, CY: 60}
	}
	return []Layer{gameLayer(shotMemo, f-t10Start+memoFrom, op)}
}

// The clock plate in Silkscreen digits: off-white, centred on x 240, the
// capitals' top at y 128, each font pixel 4 game pixels.
const (
	clockDigitsY     = 128
	clockDigitsScale = 4
)

func (tl *Timeline) clock(f int) []Layer {
	layers := []Layer{
		gameLayer(shotMemo, memoLast, GameOp{Zoom: 1}),
		{Kind: LayerChecker, Color: colBlack},
	}
	if f >= clockEnd { // 03:17 lands in blood red
		return append(layers, spriteLayer(SheetFX, "hud_clock_317", 0, SpriteOp{}))
	}
	y := Move{From: -170, To: 0, Start: t10bFrom, Dur: 8, Curve: outBack}.At(f)
	layers = append(layers, spriteLayer(SheetFX, "hud_clock_plate", 0, SpriteOp{DY: y}))
	if f >= t10bFrom+8 {
		h, m := ClockAt(f)
		layers = append(layers, Layer{Kind: LayerText, Text: fmt.Sprintf("%02d:%02d", h, m),
			X: GameW / 2, Y: clockDigitsY, Scale: clockDigitsScale, Color: colPaper, Centered: true})
	}
	return layers
}

// ---- T11: the strobe ----

func (tl *Timeline) strobe(f int) []Layer {
	shot := shotStrobe
	if tl.Reduced {
		shot = shotStrobeRed
	}
	return []Layer{gameLayer(shot, f-strobeOffset, GameOp{Zoom: 1})}
}

// ---- T12: the night-shift card ----

// nightCard is revealed by a band of scanlines running down the picture; then
// THE NIGHT SHIFT types itself, 2 frames a letter.
func (tl *Timeline) nightCard(f int) []Layer {
	y := Move{From: -16, To: GameH, Start: t12Start, Dur: 30, Curve: linear}.At(f)
	layers := []Layer{fillLayer(fullScreen, colBlack)}
	if y > 0 {
		layers = append(layers, spriteLayer(SheetCards, "card_night_shift", 0,
			SpriteOp{Exclude: tl.NightText, Clip: image.Rect(0, 0, GameW, y)}))
	}
	if n := Typed(f, t12Start+30, 2, len(tl.NightEdges)); n > 0 {
		layers = append(layers, spriteLayer(SheetCards, "card_night_shift", 0,
			SpriteOp{Src: tl.NightText, Clip: image.Rect(0, 0, tl.NightEdges[n-1], GameH)}))
	}
	if y < GameH {
		layers = append(layers, spriteLayer(SheetFX, "sweep", 0, SpriteOp{DY: y}))
	}
	return layers
}

// ---- T13 to T16: the night ----

func (tl *Timeline) coffee(f int) []Layer {
	op := GameOp{Zoom: 1}
	if f >= 2100 { // punch x2 on Mayu's shaking hands
		op = GameOp{Zoom: 2, CX: 240, CY: 150}
	}
	return []Layer{gameLayer(shotCoffee, f-coffeeOffset, op)}
}

// irisIndex gives the frame of the iris sprite for frame f of a step that runs
// from frame start: 4 frames each, 6 pictures, the crescent closing (0 to 5)
// or opening (5 to 0).
func irisIndex(f, start int, closing bool) (index int, ok bool) {
	i := (f - start) / 4
	if f < start || i > 5 {
		return 0, false
	}
	if closing {
		return i, true
	}
	return 5 - i, true
}

func (tl *Timeline) tired(f int) []Layer {
	layers := []Layer{gameLayer(shotTired, f-tiredOffset, GameOp{Zoom: 1})}
	if i, ok := irisIndex(f, t15Start-24, true); ok {
		layers = append(layers, spriteLayer(SheetFX, "iris", i, SpriteOp{}))
	}
	return layers
}

func (tl *Timeline) supply(f int) []Layer {
	op := GameOp{Zoom: 1}
	if f >= 2490 { // punch x2 on the gap in the door
		op = GameOp{Zoom: 2, CX: 50, CY: 60}
	}
	layers := []Layer{gameLayer(shotGap, f-gapOffset, op)}
	if i, ok := irisIndex(f, t15Start, false); ok {
		layers = append(layers, spriteLayer(SheetFX, "iris", i, SpriteOp{}))
	}
	return layers
}

// glitchSteps is how many frames the glitch-slice cut into T16 lasts.
const glitchSteps = 6

func (tl *Timeline) kanashibari(f int) []Layer {
	var layers []Layer
	if step := f - t16Start; step < glitchSteps {
		// Rows of the old shot (its last frame) and of the new one, shifted.
		for _, row := range GlitchRows(step, glitchSteps, 3) {
			shot, frame := shotGap, 180
			if row.FromB {
				shot, frame = shotKana, f-kanaOffset
			}
			layers = append(layers, gameLayer(shot, frame,
				GameOp{Zoom: 1, DX: row.Shift, Clip: image.Rect(0, row.Y, GameW, row.Y+glitchRowH)}))
		}
	} else {
		layers = append(layers, gameLayer(shotKana, f-kanaOffset, GameOp{Zoom: 1}))
	}
	if f >= t17Start-12 { // the channel changes on the way out
		layers = append(layers, spriteLayer(SheetFX, "channel_static", (f-(t17Start-12))/2, SpriteOp{}))
	}
	return layers
}

// ---- T17: the flashes on black ----

func (tl *Timeline) flashes(f int) []Layer {
	black := fillLayer(fullScreen, colBlack)
	switch {
	case f >= 2856 && f < 2860: // Hana's red room, 4 frames
		return []Layer{black, gameLayer(shotFlashRed, f-2855, GameOp{Zoom: 1})}
	case f >= 2904 && f < 2907: // the forged "- Hana", 3 frames
		return []Layer{black, gameLayer(shotFlashHana, f-2903, GameOp{Zoom: 1})}
	case f >= 2958 && f < 2960: // Mayu in the gap, 2 frames
		return []Layer{black, gameLayer(shotGap, 150+f-2958, GameOp{Zoom: 1})}
	case f >= 3012 && f < 3014: // the red room again, 2 frames
		return []Layer{black, gameLayer(shotFlashRed, 1+f-3012, GameOp{Zoom: 1})}
	}
	return []Layer{black}
}

// ---- T18: DON'T FALL ASLEEP. ----

// dontSleepCard is the card with its kanji and the first n letters of the line.
func (tl *Timeline) dontSleepCard(index, n int) []Layer {
	layers := []Layer{spriteLayer(SheetCards, "card_dont_sleep", index, SpriteOp{Exclude: tl.DontText})}
	if n > 0 {
		layers = append(layers, spriteLayer(SheetCards, "card_dont_sleep", index,
			SpriteOp{Src: tl.DontText, Clip: image.Rect(0, 0, tl.DontEdges[n-1], GameH)}))
	}
	return layers
}

func (tl *Timeline) dontSleep(f int) []Layer {
	layers := []Layer{fillLayer(fullScreen, colBlack)}
	if step := (f - t18Start) / 3; step < 4 { // the kanji tears in: 4 pictures, 3 frames each
		return append(layers, spriteLayer(SheetFX, "kt_dont_sleep_glitch", step, SpriteOp{}))
	}
	index := 0
	switch { // the card's own flicker: a red slip for 4 frames, a dim one for 5
	case f >= 3180 && f < 3184:
		index = 1
	case f >= 3184 && f < 3189:
		index = 2
	}
	return append(layers, tl.dontSleepCard(index, Typed(f, 3090, 2, len(tl.DontEdges)))...)
}

// ---- T19, T19b: the title and the blood ----

func (tl *Timeline) title(f int) []Layer {
	layers := []Layer{gameLayer(shotTitle, f-titleOffset, GameOp{Zoom: 1})}
	// The splash, 2 frames each for its 3 pictures, then the stain, then the
	// drips running down 1 pixel a frame.
	if i := (f - slamFrame) / 2; f >= slamFrame && i < 3 {
		layers = append(layers, spriteLayer(SheetCards, "title_splash", i, SpriteOp{}))
	}
	if f >= slamFrame+6 {
		layers = append(layers, spriteLayer(SheetCards, "title_stain", 0, SpriteOp{}))
	}
	if f >= slamFrame+10 {
		rows := min(GameH, 34+f-(slamFrame+10))
		layers = append(layers, spriteLayer(SheetCards, "title_drips", 0, SpriteOp{Clip: image.Rect(0, 0, GameW, rows)}))
	}
	return layers
}

// ---- T20: the end card ----

func (tl *Timeline) endCard(f int) []Layer {
	if f >= t20Black {
		return []Layer{fillLayer(fullScreen, colBlack)}
	}
	card := func(op SpriteOp) Layer { return spriteLayer(SheetCards, "card_end", 0, op) }
	layers := []Layer{
		fillLayer(fullScreen, colBlack),
		card(SpriteOp{Src: tl.EndLogo}),
	}
	if n := Typed(f, 3630, 2, len(tl.MadeEdges)); n > 0 {
		layers = append(layers, card(SpriteOp{Src: tl.EndMade, Clip: image.Rect(0, 0, tl.MadeEdges[n-1], GameH)}))
	}
	if f >= 3690 {
		layers = append(layers, card(SpriteOp{Src: tl.EndKanji}))
	}
	layers = append(layers, spriteLayer(SheetCards, "end_stain", 0, SpriteOp{}))
	if f >= 3720 { // the credits step in through a dither: 25 %, 50 %, then whole
		level := 0
		switch {
		case f < 3724:
			level = 4
		case f < 3728:
			level = 8
		}
		layers = append(layers, spriteLayer(SheetCards, "card_end_credits", 0, SpriteOp{Bayer: level}))
	}
	return layers
}
