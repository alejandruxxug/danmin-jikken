package game

import (
	"fmt"
	"image"
	"image/color"
	"io/fs"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"danmingame/sprites"
)

// The SCENES gallery (Figma: Overnight review -> Scenes gallery, 2026-10-14).
// Twelve cards in a 4 x 3 grid, in story order. A card the player has reached
// shows its thumbnail and replays on a click; a locked one is a dark card with
// a closed eye. The 3:17 card is a tear in the page: a hairline crack while
// locked, ripped open and glitching once found. Every picture comes from the
// scenes_gallery sheet; the title sheet's "bg" is the off-white page.

// Layout, in 1x pixels (Figma notes 52:472 and 54:488).
const (
	cardX0, cardY0         = 17, 24  // top-left of the first card
	cardPitchX, cardPitchY = 114, 74 // from one card to the next
	cardCols               = 4
	cardPad                = 20 // the sheet leaves 20 px around the card for the tear
	cardLift               = 2  // a selected card is drawn this much higher
	headerY                = 7
	footerY                = 251
	backX                  = 17
	counterRight           = 463
	tearX, tearY           = 339, 152 // the tear overlay's top-left
	backSel                = 12       // the selection on the "< BACK" button (cards are 0 to 11)
)

// cardOrigin is the top-left corner of card i's slot.
func cardOrigin(i int) image.Point {
	return image.Pt(cardX0+i%cardCols*cardPitchX, cardY0+i/cardCols*cardPitchY)
}

// cardRect is the mouse area of card i: the frame, at rest (a lifted card
// keeps the area it had, so the hover does not flicker at its edge).
func cardRect(i int) image.Rectangle {
	o := cardOrigin(i)
	return image.Rect(o.X+1, o.Y+1, o.X+103, o.Y+59)
}

// cardAt is the card under a point, or -1.
func cardAt(p image.Point) int {
	for i := range sceneCards {
		if p.In(cardRect(i)) {
			return i
		}
	}
	return -1
}

// moveSel moves the selection one step on the grid (dx, dy are -1, 0 or 1).
// The edges stop it, except that Down on the last row goes to BACK and Up
// from BACK goes back to the first card of the last row.
func moveSel(sel, dx, dy int) int {
	rows := len(sceneCards) / cardCols
	if sel == backSel {
		if dy < 0 {
			return (rows - 1) * cardCols
		}
		return sel
	}
	col, row := sel%cardCols+dx, sel/cardCols+dy
	switch {
	case row >= rows:
		return backSel
	case col < 0 || col >= cardCols || row < 0:
		return sel
	}
	return row*cardCols + col
}

// ---- The 3:17 tear: timing (Figma note 54:488) ----

// One game tick, and the tag of the glitch animation in the sheet.
const (
	tearTick = time.Second / 60
	tearTag  = "glitch"
)

// The locked card's glint plays frames 2 to 4 of the tag once, every 3 to 5
// seconds. Frame 1 lasts 140 ms, the three after it 90 + 60 + 180 ms.
const (
	glintStart  = 140 * time.Millisecond
	glintLength = 330 * time.Millisecond
	glintMin    = 180 // ticks between glints: 3 s ...
	glintSpread = 120 // ... to 5 s
)

// The hover stutter: every 72 ticks (1.2 s) the tear freezes for 6 ticks, then
// is drawn 2 px to the right for 2 ticks. On the hover the animation starts on
// frame 3, "the rip" (the first two frames last 140 + 90 ms).
const (
	stutterEvery  = 72
	stutterFreeze = 6
	stutterJump   = 2
	ripStartTicks = 14 // 230 ms
)

// stutterAt says what the hovered tear does at a tick of the hover: frozen
// (the frame holds) or jumped (drawn 2 px to the right). Reduce flashing keeps
// the freeze and drops the jump.
func stutterAt(tick int, reduce bool) (frozen, jumped bool) {
	phase := tick % stutterEvery
	start := stutterEvery - stutterFreeze - stutterJump
	frozen = phase >= start && phase < start+stutterFreeze
	jumped = !reduce && phase >= start+stutterFreeze
	return frozen, jumped
}

// tearElapsed is how far into the animation a number of playing ticks is.
// Reduce flashing plays it at half speed.
func tearElapsed(ticks int, reduce bool) time.Duration {
	d := time.Duration(ticks) * tearTick
	if reduce {
		d /= 2
	}
	return d
}

// glintFrame is the frame of the locked crack a number of ticks after a glint
// started, and whether the glint is over (then the crack rests on frame 1).
func glintFrame(meta *sprites.Meta, ticks int, reduce bool) (frame int, over bool) {
	elapsed := tearElapsed(ticks, reduce)
	if elapsed >= glintLength {
		return meta.Tags[tearTag].From, true
	}
	return meta.FrameAt(tearTag, glintStart+elapsed), false
}

// ---- The scene ----

// galleryScene is the SCENES gallery.
type galleryScene struct {
	env   *Env
	font  *Font
	page  *sprites.Sheet // the title sheet: its "bg" layer is the page
	cards *sprites.Sheet // scenes_gallery
	sel   int
	ticks int

	lastCursor image.Point // the mouse only selects when it moves, so the arrow keys win

	// The tear. tearTicks counts the ticks the animation has played (a frozen
	// tick does not count); hoverTicks counts since the 3:17 card was selected.
	tearTicks, hoverTicks int
	glintWait             int // locked: ticks until the next glint
	glintTicks            int // ticks into the glint playing now, -1 when none
	rng                   *rand.Rand
	tearFrame, tearShift  int  // what the last Update chose, for Draw and the dump
	tearFrozen            bool // ... and whether the frame is held
}

func (s *galleryScene) Enter(env *Env) error {
	s.env = env
	var err error
	if s.font, err = LoadFont(); err != nil {
		return err
	}
	if _, statErr := fs.Stat(env.Sprites, "title.json"); statErr == nil {
		if s.page, err = LoadSheet(env.Sprites, "title"); err != nil {
			return err
		}
	}
	if s.cards, err = LoadSheet(env.Sprites, "scenes_gallery"); err != nil {
		return err
	}
	s.sel = env.GallerySel
	cx, cy := env.Input.Cursor()
	s.lastCursor = image.Pt(cx, cy)
	s.rng = rand.New(rand.NewSource(env.Seed))
	s.glintTicks, s.glintWait = -1, s.nextGlint()
	if s.sel == secretCard {
		s.tearTicks = ripStartTicks
	}
	env.SetMusic("title") // the title's music goes on; nothing changes if it is playing
	s.publish()
	return nil
}

// nextGlint draws the wait until the locked crack glints again.
func (s *galleryScene) nextGlint() int { return glintMin + s.rng.Intn(glintSpread+1) }

func (s *galleryScene) Update() error {
	s.ticks++
	in := s.env.Input
	was := s.sel

	if cx, cy := in.Cursor(); image.Pt(cx, cy) != s.lastCursor {
		s.lastCursor = image.Pt(cx, cy)
		if i := cardAt(s.lastCursor); i >= 0 {
			s.sel = i
		} else if s.lastCursor.In(s.backRect()) {
			s.sel = backSel
		}
	}
	switch {
	case in.KeyPressed(ebiten.KeyArrowLeft):
		s.sel = moveSel(s.sel, -1, 0)
	case in.KeyPressed(ebiten.KeyArrowRight):
		s.sel = moveSel(s.sel, 1, 0)
	case in.KeyPressed(ebiten.KeyArrowUp):
		s.sel = moveSel(s.sel, 0, -1)
	case in.KeyPressed(ebiten.KeyArrowDown):
		s.sel = moveSel(s.sel, 0, 1)
	}
	if s.sel == secretCard && was != secretCard { // the hover starts on the rip
		s.hoverTicks, s.tearTicks = 0, ripStartTicks
	}

	pick := -1
	if x, y, ok := in.Click(); ok {
		if i := cardAt(image.Pt(x, y)); i >= 0 {
			pick = i
		} else if image.Pt(x, y).In(s.backRect()) {
			pick = backSel
		}
	}
	if in.KeyPressed(ebiten.KeyEnter) || in.KeyPressed(ebiten.KeySpace) {
		pick = s.sel
	}
	s.updateTear()
	s.publish()
	switch {
	case in.KeyPressed(ebiten.KeyEscape), in.RightClicked(), pick == backSel:
		s.leave()
	case pick >= 0 && s.env.SceneUnlocked(sceneCards[pick].ID): // a locked card does nothing
		s.sel = pick
		s.env.GallerySel = pick
		s.env.SwitchTo("replay." + sceneCards[pick].ID)
	}
	return nil
}

// leave goes back to the title, which is already settled.
func (s *galleryScene) leave() {
	s.env.GallerySel = 0
	s.env.TitleSettled = true
	s.env.SwitchTo(SceneTitle)
}

// backRect is the mouse area of the "< BACK" button.
func (s *galleryScene) backRect() image.Rectangle {
	return image.Rect(backX-3, footerY-3, backX+s.font.Width("< BACK")+3, footerY+11)
}

// updateTear moves the 3:17 animation on by one tick.
func (s *galleryScene) updateTear() {
	reduce := s.env.ReduceFlashing
	s.tearShift, s.tearFrozen = 0, false
	if !s.env.SceneUnlocked(sceneCards[secretCard].ID) {
		s.updateGlint(reduce)
		return
	}
	if s.sel == secretCard { // hovered: the stutter
		frozen, jumped := stutterAt(s.hoverTicks, reduce)
		s.hoverTicks++
		s.tearFrozen = frozen
		if jumped {
			s.tearShift = 2
		}
		if frozen {
			s.tearFrame = s.cards.FrameAt(tearTag, tearElapsed(s.tearTicks, reduce))
			return
		}
	}
	s.tearTicks++
	s.tearFrame = s.cards.FrameAt(tearTag, tearElapsed(s.tearTicks, reduce))
}

// updateGlint runs the locked crack: mostly frame 1, a glint now and then.
func (s *galleryScene) updateGlint(reduce bool) {
	s.tearFrame = s.cards.Tags[tearTag].From
	if s.glintTicks < 0 {
		if s.glintWait--; s.glintWait > 0 {
			return
		}
		s.glintTicks = 0
	}
	frame, over := glintFrame(s.cards.Meta, s.glintTicks, reduce)
	if over {
		s.glintTicks, s.glintWait = -1, s.nextGlint()
		return
	}
	s.tearFrame = frame
	s.glintTicks++
}

// publish tells the dump where the gallery is.
func (s *galleryScene) publish() {
	seen, total := s.env.SceneCounts()
	sel := "back"
	if s.sel != backSel {
		sel = sceneCards[s.sel].ID
	}
	mode := "locked"
	if s.env.SceneUnlocked(sceneCards[secretCard].ID) {
		mode = "open"
		if s.sel == secretCard {
			mode = "hover"
		}
	}
	s.env.GalleryState = &DumpGallery{Selected: sel, Seen: seen, Total: total,
		Tear: mode, TearFrame: s.tearFrame, TearShift: s.tearShift, TearFrozen: s.tearFrozen}
}

func (s *galleryScene) Draw(screen *ebiten.Image) {
	if s.page != nil {
		_ = s.page.Draw(screen, []string{"bg"}, 0, 0, 0)
	} else {
		screen.Fill(colPaper)
	}
	s.drawCentered(screen, "SCENES", 0, ScreenW, headerY, colInk)
	for i := range sceneCards {
		s.drawCard(screen, i)
	}
	for i := range sceneCards {
		s.drawTitle(screen, i)
	}
	s.drawTear(screen) // last: it tears into its neighbours
	s.drawFooter(screen)
}

// drawCard draws one card's layers from the sheet, back to front. A selected
// card (except 3:17) lifts 2 px and leaves its shadow behind.
func (s *galleryScene) drawCard(screen *ebiten.Image, i int) {
	card := sceneCards[i]
	o := cardOrigin(i)
	x, y := o.X-cardPad, o.Y-cardPad
	lifted := i == s.sel && i != secretCard
	if lifted {
		_ = s.cards.Draw(screen, []string{"shadow"}, 0, x, y)
		y -= cardLift
	}
	picture := "thumb_" + card.Thumb
	switch {
	case s.env.SceneUnlocked(card.ID):
	case i == secretCard:
		picture = "locked_317"
	default:
		picture = "locked"
	}
	frame := "frame"
	if lifted {
		frame = "frame_hover"
	}
	_ = s.cards.Draw(screen, []string{picture, frame}, 0, x, y)
}

// drawTitle writes the name under a card: dark grey, or ink with arrows when
// the card is selected. A locked card has question marks instead.
func (s *galleryScene) drawTitle(screen *ebiten.Image, i int) {
	card := sceneCards[i]
	title, c := card.Title, colDarkGrey
	if !s.env.SceneUnlocked(card.ID) {
		title, c = "? ? ?", colMidRGBA
		if i == secretCard {
			title, c = "?:??", colDarkGrey
		}
	}
	if i == s.sel {
		title, c = "> "+title+" <", colInk
	}
	o := cardOrigin(i)
	s.font.DrawCentered(screen, title, o.X+52, o.Y+59, c) // centred on the card, as in Figma
}

// drawCentered centres text in a strip, rounding the left edge down like the
// title screen does with its menu (the header and the footer line use it).
func (s *galleryScene) drawCentered(screen *ebiten.Image, text string, left, width, y int, c color.Color) {
	s.font.Draw(screen, text, left+(width-s.font.Width(text))/2, y, c)
}

// drawTear draws the 3:17 overlay: the crack, the open tear or the wider one.
func (s *galleryScene) drawTear(screen *ebiten.Image) {
	layer := "tear_317_locked"
	if s.env.SceneUnlocked(sceneCards[secretCard].ID) {
		layer = "tear_317"
		if s.sel == secretCard {
			layer = "tear_317_hover"
		}
	}
	_ = s.cards.Draw(screen, []string{layer}, s.tearFrame, tearX+s.tearShift, tearY)
}

// drawFooter writes < BACK, what the selected card is, and the counter.
func (s *galleryScene) drawFooter(screen *ebiten.Image) {
	back := colDarkGrey
	if s.sel == backSel {
		back = colInk
	}
	s.font.Draw(screen, "< BACK", backX, footerY, back)

	if s.sel != backSel {
		card := sceneCards[s.sel]
		text, c := card.Detail, colInk
		if !s.env.SceneUnlocked(card.ID) {
			text, c = "NOT SEEN YET", colMidRGBA
			if s.sel == secretCard {
				text = "? ? ?"
			}
		}
		s.drawCentered(screen, text, 0, ScreenW, footerY, c)
	}
	seen, total := s.env.SceneCounts()
	s.font.DrawRight(screen, fmt.Sprintf("%d / %d", seen, total), counterRight, footerY, colDarkGrey)
}

func (s *galleryScene) Exit() { s.env.GalleryState = nil }

// DumpGallery is the gallery's state for the dump.
type DumpGallery struct {
	Selected   string `json:"selected"`   // the card id, or "back"
	Seen       int    `json:"seen"`       // the counter: seen / total
	Total      int    `json:"total"`      // 11, or 12 once 3:17 is found
	Tear       string `json:"tear"`       // locked, open (found) or hover (found and selected)
	TearFrame  int    `json:"tearFrame"`  // the frame of the glitch tag showing
	TearShift  int    `json:"tearShift"`  // 2 on a stutter tick (drawn 2 px to the right)
	TearFrozen bool   `json:"tearFrozen"` // the frame is held (the stutter's freeze)
}
