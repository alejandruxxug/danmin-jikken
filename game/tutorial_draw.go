package game

import (
	"image"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// Drawing the tutorial, to the approved Figma views "Day 1 tutorial
// (2026-10-14)" and the builder spec card in it. Everything is in 1x screen
// pixels, and it is drawn on the picture the CRT filter then takes, like the
// rest of the day.

const (
	hintW, hintH = 200, 36 // the ui_tutorial canvas
	hintTextX    = 8       // the text's left edge, from the box's
	hintLine1    = 15      // the tops of the capitals, from the box's top
	hintLine2    = 24
	hintMaxLine  = 174 // the widest a line may be
)

// hintBoxAt is where the box sits: low on the left of the reception, under
// the folder on the desk.
func hintBoxAt(scene hintScene) image.Point {
	if scene == hintDesk {
		return image.Pt(156, 232)
	}
	return image.Pt(4, 58)
}

// hintRect is the box as a click area.
func hintRect(scene hintScene) rect {
	at := hintBoxAt(scene)
	return rect{at.X, at.Y, hintW, hintH}
}

// stepTime is how long one frame of the highlight lasts (the sheet's 125 ms).
const stepTime = 125 * time.Millisecond

// cornerPulse is the order the corner brackets play their three frames in.
var cornerPulse = [4]int{0, 1, 2, 1}

// hint draws the box with its two lines and frames the targets.
func (p *painter) hint(h *Hint, scene hintScene, targets []rect) {
	for _, t := range targets {
		p.highlight(t)
	}
	at := hintBoxAt(scene)
	sheet := p.art.tutorial
	frame := sheet.FrameAt("boil", p.elapse)
	p.keep(sheet.Draw(p.screen, []string{"hint_fill", "hint_lineart", "hint_detail", "hint_next"}, frame, at.X, at.Y))
	p.text(h.Lines[0], at.X+hintTextX, at.Y+hintLine1, colInk)
	p.text(h.Lines[1], at.X+hintTextX, at.Y+hintLine2, colInk)
}

// highlight frames a rectangle: a one-pixel ring of marching ants just
// outside it, walked clockwise from its top-left corner (pixel k is pixel
// k mod 6 of the 6-pixel hl_ants tile, which moves 2 px a frame), and four
// corner brackets that pulse outward.
func (p *painter) highlight(r rect) {
	step := int(p.elapse / stepTime)
	ants, err := p.art.tutorial.Frame("hl_ants", step%3)
	p.keep(err)
	if err != nil {
		return
	}
	tile := ants.Bounds().Min
	ring := rect{r.x - 1, r.y - 1, r.w + 2, r.h + 2}
	k := 0
	put := func(x, y int) {
		src := ants.SubImage(image.Rect(tile.X+k%6, tile.Y, tile.X+k%6+1, tile.Y+1)).(*ebiten.Image)
		var op ebiten.DrawImageOptions
		op.GeoM.Translate(float64(x), float64(y))
		p.screen.DrawImage(src, &op)
		k++
	}
	for x := ring.x; x < ring.x+ring.w; x++ { // top, left to right
		put(x, ring.y)
	}
	for y := ring.y + 1; y < ring.y+ring.h; y++ { // right side, down
		put(ring.x+ring.w-1, y)
	}
	for x := ring.x + ring.w - 2; x >= ring.x; x-- { // bottom, right to left
		put(x, ring.y+ring.h-1)
	}
	for y := ring.y + ring.h - 2; y > ring.y; y-- { // left side, up
		put(ring.x, y)
	}

	pulse := cornerPulse[step%len(cornerPulse)]
	for _, c := range []struct {
		layer string
		x, y  int
	}{
		{"hl_tl", r.x - 6, r.y - 6}, {"hl_tr", r.x + r.w - 3, r.y - 6},
		{"hl_bl", r.x - 6, r.y + r.h - 3}, {"hl_br", r.x + r.w - 3, r.y + r.h - 3},
	} {
		img, err := p.art.tutorial.Frame(c.layer, pulse)
		p.keep(err)
		if err != nil {
			return
		}
		m := img.Bounds().Min // the 9 by 9 bracket is at the corner of the canvas
		var op ebiten.DrawImageOptions
		op.GeoM.Translate(float64(c.x), float64(c.y))
		p.screen.DrawImage(img.SubImage(image.Rect(m.X, m.Y, m.X+9, m.Y+9)).(*ebiten.Image), &op)
	}
}

// tutorial draws the hint of the scene, if there is one.
func (p *painter) tutorial(sh *Shift, scene hintScene, cameraX int) {
	if h := sh.CurrentHint(scene); h != nil {
		p.hint(h, scene, sh.HintTargets(h, cameraX))
	}
}

// tutorialInput is what the player does to the hint box without clicking it:
// Esc turns the hints off for good, Space or Enter goes to the next line. It
// also tells the shift the hint was on the screen. A scene calls it every tick.
func tutorialInput(env *Env, sh *Shift, scene hintScene) error {
	if !sh.TutorialOn() {
		return nil
	}
	in := env.Input
	switch {
	case in.KeyPressed(ebiten.KeyEscape):
		return env.SkipTutorial()
	case in.KeyPressed(ebiten.KeySpace) || in.KeyPressed(ebiten.KeyEnter):
		sh.AdvanceHint(scene)
	}
	sh.NoteHintShown(scene)
	return nil
}

// hintClicked says whether a click at (x, y) landed on the hint box. If so
// the hint is over and the click does nothing else.
func hintClicked(sh *Shift, scene hintScene, x, y int) bool {
	if sh.CurrentHint(scene) == nil || !hintRect(scene).has(x, y) {
		return false
	}
	sh.AdvanceHint(scene)
	return true
}
