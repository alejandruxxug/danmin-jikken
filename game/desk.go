package game

import (
	"fmt"
	"image"
	"image/color"
	"io/fs"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"danmingame/sprites"
)

// The reception desk, built from the approved Figma views "Day - Reception
// desk", "Day - Beds panel" and "Day - No room alert" (Art Direction -> Art
// log). Everything is in world pixels: the room is 720 wide and the camera
// shows 480 of it. The positions below come from those views (Figma px / 4).
const (
	RoomW          = 720
	MaxCameraX     = RoomW - ScreenW // 240
	DefaultCameraX = MaxCameraX / 2  // 120: the Figma frame's crop

	walkFrom = 120 // guests walk in from this far away
)

// A spot a patient can stand, sit or lie at, in world pixels (sprite top-left).
type spot struct{ x, y int }

var (
	windowSpot = spot{312, 64} // behind the counter glass; the desk hides below y 186
	// The five chairs of the waiting room, left to right. A seated sprite's
	// body is 24 px wide, so the slots are at least 24 px apart (no two
	// patients overlap), and the last slot keeps the patient and their "z"
	// bubble left of x 258, where the post of the counter frame starts. That
	// puts the seated patients up to 5 px left of their chairs.
	chairSpots = []spot{{122, 110}, {150, 110}, {176, 110}, {202, 110}, {226, 110}}
	// Where sleepers lie on the hallway floor, in front of the ward doors.
	hallwaySpots = []spot{{548, 176}, {598, 176}, {648, 176}, {548, 198}, {598, 198}}
)

// A rectangle on the screen or in the room, for clicks.
type rect struct{ x, y, w, h int }

func (r rect) has(x, y int) bool { return x >= r.x && x < r.x+r.w && y >= r.y && y < r.y+r.h }

// The stamps on the desk, in world pixels. Each stamp has two places to
// click: its cubby on the counter (labelled ADMIT, DISCHARGE, REFER) and the
// stamp itself on the desk (the letters A, D and R).
var stampRects = []struct {
	stamp Stamp
	rects []rect
}{
	{StampAdmit, []rect{{312, 199, 48, 33}, {481, 241, 20, 27}}},
	{StampDischarge, []rect{{360, 199, 48, 33}, {502, 241, 20, 27}}},
	{StampRefer, []rect{{408, 199, 49, 33}, {525, 241, 20, 27}}},
}

// The tab at the bottom of the screen opens and closes the computer. It is
// on the screen, not in the room, so it never pans.
var tabRect = rect{160, 253, 160, 17}

// The glass opening of the counter window, in world pixels: the posts' centre
// lines are x 261 and 459, the top bar y 61, and the counter covers y 190
// and below. Every guest on the window row is drawn only inside it.
var windowGlass = rect{261, 61, 198, 129}

// The layers that make up a patient, back to front. Awake and asleep differ
// only in the eyes, the mouth and the zzz.
var (
	fisherAwake   = []string{"fill", "lineart", "detail", "eyes_normal", "mouth_neutral"}
	patientAsleep = []string{"fill", "lineart", "detail", "eyes_closed", "mouth_slack", "fx_zzz"}
	hallwayLayers = []string{"fill", "lineart", "eyes_closed", "fx_zzz"}
)

// windowAwake is the awake layer list of a window sprite: every visitor has
// their own mouth (Visitor.Mouth).
func windowAwake(v Visitor) []string {
	return []string{"fill", "lineart", "detail", "eyes_normal", v.Mouth}
}

// eyesFor is the layer list with the eyes shut when blink is true: the
// eyes_normal layer is swapped for eyes_closed.
func eyesFor(layers []string, blink bool) []string {
	if !blink {
		return layers
	}
	out := append([]string(nil), layers...)
	for i, l := range out {
		if l == "eyes_normal" {
			out[i] = "eyes_closed"
		}
	}
	return out
}

var (
	roomLayers  = []string{"fill", "lineart", "detail"}
	glassLayers = []string{"dust", "glare", "bezel"} // "reflection" is for day 3
)

// deskLayers are the layers of the desk sheet for a day: the same three as
// the room, plus that day's memo ("memo_day1" ... "memo_day3").
func deskLayers(day int) []string {
	return []string{"fill", "lineart", "detail", fmt.Sprintf("memo_day%d", day)}
}

// DeskArt is every sheet the day scenes draw, plus the font.
type DeskArt struct {
	room, desk, glass, waiting, hallway, computer, icons, iconsM4                                     *sprites.Sheet
	deskTop, folder, docID, docReferral, docSleepLog, docWristband, docStamps, docCitation, inkSpecks *sprites.Sheet
	docReport, handHana, handMayu                                                                     *sprites.Sheet            // Hana's report and the two hands that can write it
	docOrientation, tutorial                                                                          *sprites.Sheet            // the orientation sheet and the tutorial's hint box and highlight pieces
	window                                                                                            map[string]*sprites.Sheet // by Visitor.Sprite
	font                                                                                              *Font

	// Scratch pictures the look-down desk draws into before bending them.
	back, top, closed, open, page *ebiten.Image
	white                         *ebiten.Image // one white pixel, the source of flat-coloured triangles
}

// LoadDeskArt loads the sheets and the font.
func LoadDeskArt(fsys fs.FS) (*DeskArt, error) {
	a := &DeskArt{}
	for _, l := range []struct {
		sheet **sprites.Sheet
		name  string
	}{
		{&a.room, "reception_bg"}, {&a.desk, "desk"}, {&a.glass, "screen_glass"},
		{&a.waiting, "waiting_fisher"}, {&a.hallway, "hallway_sleeper"},
		{&a.computer, "computer"}, {&a.icons, "ui_icons"}, {&a.iconsM4, "ui_icons_m4"},
		{&a.deskTop, "desk_top"}, {&a.folder, "folder"}, {&a.docID, "doc_settler_id"},
		{&a.docReferral, "doc_referral"}, {&a.docSleepLog, "doc_sleep_log"}, {&a.docWristband, "doc_wristband"},
		{&a.docStamps, "doc_stamps"}, {&a.inkSpecks, "ink_specks"}, {&a.docCitation, "doc_citation"},
		{&a.docReport, "doc_report"}, {&a.handHana, "hand_hana"}, {&a.handMayu, "hand_mayu"},
		{&a.docOrientation, "doc_orientation"}, {&a.tutorial, "ui_tutorial"},
	} {
		var err error
		if *l.sheet, err = LoadSheet(fsys, l.name); err != nil {
			return nil, fmt.Errorf("day art: %w", err)
		}
	}
	a.window = make(map[string]*sprites.Sheet)
	for _, name := range windowSprites() {
		sheet, err := LoadSheet(fsys, name)
		if err != nil {
			return nil, fmt.Errorf("day art: %w", err)
		}
		a.window[name] = sheet
	}
	a.back, a.top = ebiten.NewImage(ScreenW, ScreenH), ebiten.NewImage(ScreenW, ScreenH)
	a.closed, a.open = ebiten.NewImage(folderW, folderH), ebiten.NewImage(folderW, folderH)
	a.page = ebiten.NewImage(pageW, pageH)
	a.white = ebiten.NewImage(1, 1)
	a.white.Fill(color.White)
	var err error
	if a.font, err = LoadFont(); err != nil {
		return nil, err
	}
	return a, nil
}

// painter draws one frame and remembers the first error, because Draw
// cannot return one.
type painter struct {
	art      *DeskArt
	screen   *ebiten.Image
	day      int           // which day's memo the desk shows
	elapse   time.Duration // time since the scene started, for the boil
	dark     int           // palette steps darker the next sheets are drawn (a Severe patient)
	forged   bool          // Mayu's hand wrote Hana's report (route_hana), see night.ReportHand
	tutTicks int           // how many checks of the orientation sheet are ticked (the day 1 tutorial)
	err      error
}

func (p *painter) keep(err error) {
	if err != nil && p.err == nil {
		p.err = err
	}
}

// sheet draws layers of a sheet; the boil frame follows the sheet's own speed.
func (p *painter) sheet(s *sprites.Sheet, layers []string, x, y int) {
	p.keep(s.DrawDarker(p.screen, layers, s.FrameAt(sprites.BoilTag, p.elapse), x, y, p.dark))
}

// room draws the reception: the room, the guests and the desk, as the
// camera at cam sees them. sh may be nil.
func (p *painter) room(sh *Shift, cam int) {
	p.sheet(p.art.room, roomLayers, -cam, 0)
	p.sheet(p.art.room, []string{p.callLayer(sh)}, -cam, 0)
	// Guests stand between the room and the desk, so the counter hides the
	// window patient's body below its top edge. With no shift (the top-10
	// after the day ended) the room is empty.
	if sh != nil {
		// Guests who stand or sit first, then those still walking, so a
		// guest passing a chair is drawn over the one sitting there.
		for _, walking := range []bool{false, true} {
			for _, g := range sh.Guests {
				if (g.steps > 0) == walking {
					p.guest(sh, g, cam)
				}
			}
		}
	}
	// The desk shares the room's origin, so it pans with it.
	p.sheet(p.art.desk, deskLayers(p.day), -cam, 0)
}

// sheetIn draws like sheet, but only inside the clip rectangle (screen
// pixels): everything outside it is left alone.
func (p *painter) sheetIn(clip rect, s *sprites.Sheet, layers []string, x, y int) {
	r := image.Rect(clip.x, clip.y, clip.x+clip.w, clip.y+clip.h).Intersect(p.screen.Bounds())
	if r.Empty() {
		return
	}
	full := p.screen
	p.screen = full.SubImage(r).(*ebiten.Image) // draws keep screen coordinates
	p.sheet(s, layers, x, y)
	p.screen = full
}

// callBlink is how long the nurse call light stays lit, then dark.
const callBlink = 500 * time.Millisecond

// callLayer is the nurse call light over the ward doors: lit every other
// half second while a Severe patient sleeps with nobody attending.
func (p *painter) callLayer(sh *Shift) string {
	if sh != nil && sh.SevereCalling() && (p.elapse/callBlink)%2 == 0 {
		return "call_on"
	}
	return "call_off"
}

// guest draws one guest as the shift says they look.
func (p *painter) guest(sh *Shift, g *Guest, cam int) {
	p.dark = sh.Darkness(g) // a Severe patient is darker for every episode
	defer func() { p.dark = 0 }()
	x, y := int(g.x)-cam, int(g.y)+WalkBob(g.steps)
	switch sh.LookOf(g) {
	case LookStanding:
		blink := sh.Blinking(g)
		if int(g.y) == windowSpot.y { // on the window row: the big sprite
			p.windowGuest(g, eyesFor(windowAwake(g.Visitor), blink), x, y, cam)
		} else {
			p.sheet(p.art.waiting, eyesFor(fisherAwake, blink), x, y)
		}
	case LookDozing:
		p.windowGuest(g, patientAsleep, x, y, cam)
	case LookAsleep:
		p.sheet(p.art.waiting, patientAsleep, x, y)
	case LookHallway:
		p.sheet(p.art.hallway, hallwayLayers, x, y)
	}
}

// windowGuest draws a guest on the window row, clipped to the glass opening
// so they slide in and out from behind the posts instead of walking over
// the wall.
func (p *painter) windowGuest(g *Guest, layers []string, x, y, cam int) {
	clip := rect{windowGlass.x - cam, windowGlass.y, windowGlass.w, windowGlass.h}
	p.sheetIn(clip, p.art.window[g.Sprite], layers, x, y)
}

// glass is the dust and bezel over the whole screen. It never pans.
func (p *painter) glass() {
	p.keep(p.art.glass.Draw(p.screen, glassLayers, 0, 0, 0))
}

// fill paints a flat rectangle: HUD plates and buttons, as in the views.
func (p *painter) fill(r rect, c color.Color) {
	p.screen.SubImage(image.Rect(r.x, r.y, r.x+r.w, r.y+r.h)).(*ebiten.Image).Fill(c)
}

// outline paints a one-pixel frame.
func (p *painter) outline(r rect, c color.Color) {
	p.fill(rect{r.x, r.y, r.w, 1}, c)
	p.fill(rect{r.x, r.y + r.h - 1, r.w, 1}, c)
	p.fill(rect{r.x, r.y, 1, r.h}, c)
	p.fill(rect{r.x + r.w - 1, r.y, 1, r.h}, c)
}

// text prints with the top of the capitals at y.
func (p *painter) text(s string, x, y int, c color.Color) {
	p.art.font.Draw(p.screen, s, x, y-capTop, c)
}

func (p *painter) textRight(s string, x, y int, c color.Color) {
	p.art.font.DrawRight(p.screen, s, x, y-capTop, c)
}

func (p *painter) textCentered(s string, x, y int, c color.Color) {
	p.art.font.DrawCentered(p.screen, s, x, y-capTop, c)
}

// hud draws the DAY plate, the PATIENTS plate and the computer tab.
func (p *painter) hud(sh *Shift, clock, tabLabel string) {
	// The plates are solid, so the ceiling lights do not run through the text.
	p.fill(rect{8, 8, 76, 18}, colPaper)
	p.outline(rect{8, 8, 76, 18}, colInk)
	p.text(fmt.Sprintf("DAY %d  %s", sh.Day, clock), 14, 16, colInk)
	p.fill(rect{400, 8, 72, 18}, colPaper)
	p.outline(rect{400, 8, 72, 18}, colInk)
	p.text(fmt.Sprintf("PATIENTS %d/%d", sh.Served(), sh.Total()), 406, 16, colInk)
	p.tab(tabLabel)
	p.filesTab(sh)
}

// tab draws the black tab at the bottom of the screen.
func (p *painter) tab(label string) {
	p.fill(tabRect, colInk)
	p.textCentered(label, tabRect.x+tabRect.w/2, 261, colPaper)
}

// toast shows the last message in a black bar at the top, between the two HUD
// plates. It is the one strip of the screen where the approved views have
// nothing but ceiling at every camera position (the bottom strip holds the
// stamps, the tray and the bell, and they pan).
func (p *painter) toast(msg string) {
	if msg == "" {
		return
	}
	w := p.art.font.Width(msg) + 12
	r := rect{ScreenW/2 - w/2, 10, w, 14}
	p.fill(r, colInk)
	p.textCentered(msg, ScreenW/2, 14, colPaper)
}

// noRoomBox is the "no room" alert (the Figma view's box: 223 x 92).
var (
	alertBox    = rect{129, 70, 223, 92}
	alertHeader = rect{129, 70, 223, 18}
	callButton  = rect{140, 146, 100, 16}
	bedsButton  = rect{250, 146, 90, 16}
)

// alert draws the "no room" box over the room.
func (p *painter) alert(a *Alert) {
	p.fill(alertBox, colPaper)
	p.fill(alertHeader, colInk)
	p.keep(p.art.icons.Draw(p.screen, []string{"alert"}, p.art.icons.FrameAt(sprites.BoilTag, p.elapse), 133, 71))
	p.text("NO ROOM AVAILABLE", 162, 78, colPaper)
	p.text("EVERY BED IS TAKEN.", 140, 99, colInk)
	p.text(a.Guest.Name+" FELL ASLEEP", 140, 111, colInk)
	p.text("IN THE HALLWAY.", 140, 123, colInk)
	p.text("DRIFT OFF WELL, FRIEND!", 140, 135, colDim)
	p.fill(callButton, colInk)
	p.text("CALL STAFF", callButton.x+10, callButton.y+7, colPaper)
	p.outline(bedsButton, colInk)
	p.text("OPEN BEDS", bedsButton.x+10, bedsButton.y+7, colInk)
}
