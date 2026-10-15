package night

import (
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"io/fs"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"danmingame/sprites"
)

// Screen size and the places the approved Figma views put things (Art
// Direction -> Art log, M5 batch 1). All in game pixels (Figma / 4).
const (
	screenW, screenH = 480, 270
	girlX, girlY     = 112, 2 // the girl sprite's top-left corner

	textX, textCapTop   = 44, 216 // dialogue: left edge, first line's capital tops
	lineStep, textWidth = 10, 392
	maxTextLines        = 4

	plateX0, plateX1 = 44, 139 // the name plate's inner span
	plateCapTop      = 198

	choiceX, choiceW, choiceH = 112, 256, 22
	capTopOffset              = 3 // Silkscreen: capitals start 3 px below the y given to Draw

	// The prologue's faces: the girl sprites rotated 180 degrees, only the
	// top 136 rows of the sprite (head and shoulders), hanging from the top edge.
	faceRows        = 136
	mayuFaceX       = 70
	mayuFaceY       = -6
	hanaFaceX       = 236
	hanaFaceY       = -14
	monitorX0       = 337 // the ward monitor's screen (x 333-447, y 47-115) minus a margin
	monitorX1       = 443
	monitorLineStep = 9
	monitorFirstY   = 56
)

// Choice button rows: y of the first button and the step between buttons.
var choiceY = []int{124, 150, 176}

// Colours the night UI draws with (assets/palette/danmin.hex).
var (
	colInk    = color.RGBA{0x0b, 0x0b, 0x0c, 0xff}
	colDark   = color.RGBA{0x2e, 0x2e, 0x33, 0xff}
	colMid    = color.RGBA{0x6b, 0x6b, 0x70, 0xff}
	colLight  = color.RGBA{0xb9, 0xb9, 0xb4, 0xff} // thoughts, normal choice labels
	colPaper  = color.RGBA{0xec, 0xec, 0xe6, 0xff} // speech, hovered choice label
	colBlood  = color.RGBA{0x8a, 0x03, 0x03, 0xff}
	colTraces = color.RGBA{0xb9, 0xb9, 0xb4, 0xff}
	colSage   = color.RGBA{0x9f, 0xb1, 0x9f, 0xff} // deep sage: speakers' names in the white bars
)

// TextDrawer is the game's pixel font: it draws text with its top-left at
// (x, y) and measures it. The game's Font satisfies it, so this package does
// not need to import the game.
type TextDrawer interface {
	Draw(dst *ebiten.Image, s string, x, y int, c color.Color)
	Width(s string) int
}

//go:embed redscale.kage
var redscaleSource []byte

// shakeTable is how far the picture is moved (x, y) on each tick of a
// "shake": a fixed pattern that settles, so the result is the same on every run.
var shakeTable = [...][2]int{{3, 1}, {-3, -2}, {2, 2}, {-2, -1}, {3, -2}, {-3, 1}, {2, -1}, {-2, 2}, {1, 1}, {-1, -1}, {1, -1}, {0, 1}}

// ShakeOffset is the shake at a tick since it started: a new offset every
// 4 ticks, and none once the pattern is used up.
func ShakeOffset(ticks int) (dx, dy int) {
	i := ticks / 4
	if i < 0 || i >= len(shakeTable) {
		return 0, 0
	}
	return shakeTable[i][0], shakeTable[i][1]
}

// Art is every sheet the night scenes draw with.
type Art struct {
	NoWords            bool           // the trailer: no subtitles, text box or menu on top of the picture
	scratch            *ebiten.Image  // the finished scene, when a shake or the redscale must reshape it
	red                *ebiten.Shader // redscale.kage
	mayu, hana         *sprites.Sheet
	station, ceiling   *sprites.Sheet
	ui, choice         *sprites.Sheet
	handMayu, handHana *sprites.Sheet // the name plate's copies: off-white ink
	inkMayu, inkHana   *sprites.Sheet // black ink for the paper
	cinema             *sprites.Sheet // fx_cinema: shade, bars
	// Pictures the camera and the shaders work on.
	world, zoomed, keep *ebiten.Image
	wash                [4]*ebiten.Image // the Wake ending's dither levels, made when first needed
	font                TextDrawer
	bgs                 map[string]*sprites.Sheet // the batch 2 backgrounds that have a sheet, by sheet name
}

// A background name in a script is drawn from the layers of one sheet. A
// name that is not here (and not one of the built-in fills) is a placeholder
// until its sheet exists in game/assets/sprites.
type bgDef struct {
	sheet  string
	layers []string
	// A CG is a full-screen picture whose layers the script switches with
	// "layer": layers lists only the ones that are always on, and the sheet's
	// own layer order is the draw order. A CG boils (frame from the "boil" tag).
	cg bool
}

var bgTable = map[string]bgDef{
	"ward_c":    {"ward_c", []string{"room", "sleepers", "blanket"}, false},
	"corridor":  {"corridor", []string{"corridor", "sign_station"}, false},
	"hana_far":  {"corridor", []string{"corridor", "sign_station", "hana_far"}, false},
	"hana_door": {"corridor", []string{"corridor", "sign_quarters", "door_crack"}, false},
	// The day 3 report on the desk, for the match cut of the realization.
	"report": {"doc_report", []string{"fill", "lineart", "detail"}, false},
	// The Wake ending's last picture: the procedure room with nobody in it.
	"procedure_empty": {"procedure", []string{"room", "lamp_glow", "rail"}, false},
	// The secret ending's: Barrow with no expression at all.
	"procedure_blank": {"procedure", []string{"room", "lamp_glow", "barrow", "barrow_blank", "rail"}, false},
	"supply_room":     {"supply", []string{"wall", "bulb_glow", "table", "prop_cans", "prop_syringe"}, false},
	"hana_room":       {"hana_room", []string{"room", "fx_blood", "cans", "hana_silhouette"}, false},
	// Hana's room in close-ups (Art log 2026-10-14, cinematic pass).
	"door_closeup": {"hana_door", []string{"door", "card"}, false},                      // her name is written on the card
	"hana_floor":   {"hana_floor", []string{"room", "fx_blood", "hand", "cans"}, false}, // 720 wide: the camera pans
	"hana_desk":    {"hana_desk", []string{"desk", "fx_blood"}, false},                  // the envelope is drawn after the red
	// The paper's writing is drawn by the game (drawPaper), in the writer's hand.
	"note":        {"note", []string{"backdrop", "paper_note", "hands_note"}, false},
	"letter":      {"note", []string{"backdrop", "paper_letter", "hands_letter"}, false},
	"ward_c_dark": {"ward_c", []string{"room", "sleepers", "blanket", "dark"}, false},
	"ward_c_up":   {"ward_c", []string{"room", "sitters", "red_eyes", "clock_317"}, false},
	// Night 3: Dr. Barrow smiling at the foot of the procedure table.
	"procedure_room": {"procedure", []string{"room", "lamp_glow", "barrow", "barrow_smile", "rail"}, false},
	// Sleep ending: the reception desk by day. drawDeskDay adds Mayu (or
	// the empty desk) between "room" and the front of the next desk.
	"desk_day": {"desk_day", []string{"room", "next_desk_front"}, false},
	// CGs (Art log 2026-10-14): the script adds the rest with "layer".
	"cg_coffee":      {"cg_coffee", []string{"backdrop", "mayu", "counter", "prop_pot", "hands_cup", "steam"}, true},
	"cg_supply":      {"cg_supply", []string{"wall", "bulb_glow", "hana", "table", "prop_cans", "arms"}, true},
	"cg_kanashibari": {"cg_kanashibari", []string{"room", "bed", "hand"}, true},
	// Forever (Art log 2026-10-14): Mayu close up through the slit of your
	// eyelids, then the drop. The rope and the smile are switched with "layer".
	"cg_forever":      {"cg_forever", []string{"ceiling", "mayu", "eyes_soft", "brows_worried", "fx_blush", "fx_tears"}, true},
	"cg_forever_drop": {"cg_forever_drop", []string{"ceiling", "lamp", "pipe", "rope", "dust"}, true}, // plays once, see DropFrame
}

func bgTableDefs() []bgDef {
	var defs []bgDef
	for _, def := range bgTable {
		defs = append(defs, def)
	}
	return defs
}

// fills are the backgrounds the painter makes itself.
var fills = map[string]bool{"black": true, "white": true, "grey": true, "station": true, "ceiling": true}

// DrawBlood paints the blood t ticks after the cut (BloodAt). It does
// nothing until the blood sheet is in game/assets/sprites.
func (a *Art) DrawBlood(dst *ebiten.Image, t int) {
	sheet := a.bgs[BloodSheet]
	if sheet == nil || t < 0 {
		return
	}
	b := BloodAt(t)
	if b.Splash > 0 {
		_ = sheet.Draw(dst, []string{fmt.Sprintf("splash_%d", b.Splash)}, 0, 0, 0)
		return
	}
	_ = sheet.Draw(dst, []string{"stain"}, 0, 0, 0)
	for i, slid := range b.Drips {
		_ = sheet.Draw(dst, []string{fmt.Sprintf("drip_%d", i+1)}, 0, 0, slid)
	}
}

// IsPlaceholder says whether a background name has no art yet.
func (a *Art) IsPlaceholder(bg string) bool {
	if fills[bg] {
		return false
	}
	def, ok := bgTable[bg]
	return !ok || a.bgs[def.sheet] == nil
}

// LoadArt reads the night sheets from the embedded sprites folder.
func LoadArt(fsys fs.FS, font TextDrawer) (*Art, error) {
	a := &Art{font: font}
	shader, err := ebiten.NewShader(redscaleSource)
	if err != nil {
		return nil, fmt.Errorf("night art: compiling the redscale shader: %w", err)
	}
	a.red = shader
	a.bgs = map[string]*sprites.Sheet{}
	extra := []bgDef{{sheet: BloodSheet}, {sheet: "mayu_day"}, {sheet: "desk_top"}}
	for _, def := range append(extra, bgTableDefs()...) {
		if _, loaded := a.bgs[def.sheet]; loaded {
			continue
		}
		if _, err := fs.Stat(fsys, def.sheet+".json"); err != nil {
			continue // not drawn yet: its backgrounds stay placeholders
		}
		sheet, err := sprites.Load(fsys, def.sheet)
		if err != nil {
			return nil, fmt.Errorf("night art: %w", err)
		}
		a.bgs[def.sheet] = sheet
	}
	for name, dst := range map[string]**sprites.Sheet{
		"mayu": &a.mayu, "hana": &a.hana, "bg_station": &a.station, "bg_ceiling": &a.ceiling,
		"vn_ui": &a.ui, "vn_choice": &a.choice, "fx_cinema": &a.cinema, "hand_mayu": &a.handMayu, "hand_hana": &a.handHana,
	} {
		sheet, err := sprites.Load(fsys, name)
		if err != nil {
			return nil, fmt.Errorf("night art: %w", err)
		}
		*dst = sheet
	}
	// The handwriting is black ink for paper; the name plate is black, so the
	// plate uses copies with the ink off-white and the joins mid grey.
	a.inkMayu, a.inkHana = a.handMayu, a.handHana
	onPlate := map[color.RGBA]color.RGBA{colInk: colPaper, colDark: colMid}
	a.handMayu, a.handHana = a.handMayu.Remapped(onPlate), a.handHana.Remapped(onPlate)
	return a, nil
}

// Where the AUTO label is written (caps top 252).
const autoLabelX, autoLabelCapTop = 400, 252

// AutoRect is the AUTO toggle on the text box, for the mouse (the pixels of
// the auto_off / auto_on layers, with a little room around them).
var AutoRect = image.Rect(386, 246, 432, 264)

// ChoiceRects are the buttons of a menu with n options, for the mouse.
func ChoiceRects(n int) []image.Rectangle {
	rects := make([]image.Rectangle, 0, n)
	for i := 0; i < n && i < len(choiceY); i++ {
		rects = append(rects, image.Rect(choiceX, choiceY[i], choiceX+choiceW, choiceY[i]+choiceH))
	}
	return rects
}

// Draw paints the whole night screen for the engine's current state. Two
// steps: the picture (the wall, the girl, the props, the screen effects),
// which the camera can zoom and the shake and the redscale can work on, then
// the overlay on top that never moves (the bars, the subtitles or the text
// box, the menu).
func (a *Art) Draw(dst *ebiten.Image, e *Engine) {
	_, shaking := e.FxTicks("shake")
	shaking = shaking && e.FxProgress("shake") < 1
	rise, drain := e.RedscaleAmounts()
	scale, cx, cy := e.Zoom()
	if !shaking && rise == 0 && scale == 1 {
		a.drawPicture(dst, e)
		a.drawKeepWhite(dst, e, scale, cx, cy)
		a.drawOverlay(dst, e)
		return
	}
	if a.world == nil {
		a.world = ebiten.NewImage(screenW, screenH)
		a.zoomed = ebiten.NewImage(screenW, screenH)
		a.keep = ebiten.NewImage(screenW, screenH)
	}
	a.world.Clear()
	a.drawPicture(a.world, e)
	view := a.world
	if scale != 1 {
		a.zoomed.Clear()
		blit(a.zoomed, a.world, scale, cx, cy)
		view = a.zoomed
	}
	if !shaking && rise == 0 {
		dst.DrawImage(view, nil)
	} else {
		// The shake and the redscale work on the finished picture.
		dx, dy := 0, 0
		if shaking {
			ticks, _ := e.FxTicks("shake")
			dx, dy = ShakeOffset(ticks)
		}
		dst.Fill(colInk)
		op := &ebiten.DrawRectShaderOptions{}
		op.Images[0] = view
		op.GeoM.Translate(float64(dx), float64(dy))
		op.Uniforms = map[string]any{"Rise": float32(rise), "Drain": float32(drain)}
		dst.DrawRectShader(screenW, screenH, a.red, op)
	}
	a.drawKeepWhite(dst, e, scale, cx, cy)
	a.drawOverlay(dst, e)
}

// blit draws src (a whole 480x270 picture) into dst as the camera sees it:
// the part around (cx, cy) at the given scale. Nearest-neighbour, so a
// whole-number scale stays crisp.
func blit(dst, src *ebiten.Image, scale, cx, cy float64) {
	ox, oy := CropOrigin(scale, cx, cy)
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(-ox, -oy)
	op.GeoM.Scale(scale, scale)
	dst.DrawImage(src, &op)
}

// drawKeepWhite draws what stays white in Hana's red room, after the redscale
// (Art Direction -> Hana's room redscale): the envelope, and "Four" written
// on it. It follows the camera, so a zoom keeps it in place.
func (a *Art) drawKeepWhite(dst *ebiten.Image, e *Engine, scale, cx, cy float64) {
	var sheet *sprites.Sheet
	switch {
	case e.BG() == "hana_room" && e.Prop("envelope"):
		sheet = a.bgs["hana_room"]
	case e.BG() == "hana_desk":
		sheet = a.bgs["hana_desk"]
	}
	if sheet == nil {
		return
	}
	target := dst
	if scale != 1 {
		a.keep.Clear()
		target = a.keep
	}
	_ = sheet.Draw(target, []string{"envelope"}, 0, 0, 0)
	if e.BG() == "hana_desk" {
		a.writeCentred(target, a.inkHana, HanaStyle, "Four", envelopeX, envelopeBase)
	}
	if scale != 1 {
		blit(dst, a.keep, scale, cx, cy)
	}
}

// Where the name goes on the envelope (hana_desk) and on the door card.
const (
	envelopeX, envelopeBase = 242, 156
	cardX, cardBase         = 222, 96
)

// writeCentred writes text in a handwriting sheet, centred on x, with its
// baseline at the given y.
func (a *Art) writeCentred(dst *ebiten.Image, hand *sprites.Sheet, style Style, text string, x, base int) {
	placed, w := Layout(text, SheetMetrics{hand.Meta}, style)
	for _, p := range placed {
		_ = hand.Draw(dst, []string{p.Glyph}, 0, x-w/2+p.X, base-glyphBaseRow+p.Y)
	}
}

// drawPicture paints the picture without the camera, the shake or the
// redscale: the background, the girl, the props and the screen effects.
func (a *Art) drawPicture(dst *ebiten.Image, e *Engine) {
	tick := e.Ticks()
	a.drawBackground(dst, e, tick)
	switch {
	case e.Faces():
		a.drawFaces(dst, e)
	case a.isCG(e.BG()):
		// A CG replaces the background and the girl in one picture.
	case e.BG() == "desk_day":
		// A day scene: no night portrait, Mayu is part of the background.
	case e.GirlShown():
		sheet := a.mayu
		if e.Visible() == "hana" {
			sheet = a.hana
		}
		_ = sheet.Draw(dst, e.PoseOf(e.Visible()).Layers(), 0, girlX, girlY+e.HopOffset())
	}
	if e.BG() == "station" {
		layers := []string{"counter", "prop_coffee"}
		for _, p := range []string{"can", "syringe", "note"} {
			if e.Prop(p) {
				layers = append(layers, "prop_"+p)
			}
		}
		_ = a.station.Draw(dst, layers, 0, 0, 0)
	}
	a.drawEffects(dst, e)
}

// drawOverlay paints what sits on top of the picture and never moves: the
// bars in cinematic mode, the blood on the lens, then the
// words (subtitles in the bottom bar, or the text box), the arrow and the menu.
func (a *Art) drawOverlay(dst *ebiten.Image, e *Engine) {
	a.drawBars(dst, e)
	if _, ok := e.FxTicks("sting"); ok && e.FxProgress("sting") < 1 {
		a.drawSting(dst, e)
		return
	}
	if t, ok := e.FxTicks("splash"); ok {
		a.DrawBlood(dst, t)
	}
	if a.NoWords {
		return
	}
	switch {
	case e.OnPaper():
		// A note or a letter is written on the paper: no text box, only the arrow.
		if e.LineDone() {
			a.drawContinue(dst, e)
		}
	case e.HasLine() && e.Cinematic():
		a.drawSubtitle(dst, e)
		if e.LineDone() && len(e.Options()) == 0 {
			a.drawContinue(dst, e)
		}
	case e.HasLine():
		a.drawTextBox(dst, e)
	}
	if e.Holding() {
		// A shot held for a click has no text box, only the arrow.
		a.drawContinue(dst, e)
	}
	if len(e.Options()) > 0 {
		a.drawChoices(dst, e)
	}
}

// drawContinue is the little bobbing arrow.
func (a *Art) drawContinue(dst *ebiten.Image, e *Engine) {
	frame := a.ui.FrameAt("bob", time.Duration(e.Ticks())*time.Second/TicksPerSecond)
	_ = a.ui.Draw(dst, []string{"continue"}, frame, 0, 0)
}

// drawBars paints the letterbox bars at their current height. The sheet's
// bar is a solid block, so one row of it stretched to the height is the same
// pixels; that lets the bars slide and grow to any size.
func (a *Art) drawBars(dst *ebiten.Image, e *Engine) {
	h := int(math.Round(e.BarsHeight()))
	if h <= 0 {
		return
	}
	layer := "bars"
	if e.WhiteBars() {
		layer = "bars_white"
	}
	frame, err := a.cinema.Frame(layer, 0)
	if err != nil {
		return
	}
	b := frame.Bounds()
	top := frame.SubImage(image.Rect(b.Min.X, b.Min.Y, b.Max.X, b.Min.Y+1)).(*ebiten.Image)
	bottom := frame.SubImage(image.Rect(b.Min.X, b.Max.Y-1, b.Max.X, b.Max.Y)).(*ebiten.Image)
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(1, float64(h))
	dst.DrawImage(top, &op)
	op.GeoM.Reset()
	op.GeoM.Scale(1, float64(h))
	op.GeoM.Translate(0, float64(screenH-h))
	dst.DrawImage(bottom, &op)
}

// Subtitles (Art Direction -> Cinematic pass): Silkscreen, centred on x 240,
// wrapped to 360 px; the caps tops are 250 for one line, 245 and 255 for two.
const (
	subtitleWidth = 360
	subtitleStep  = 10
)

// SubtitleTops are the y of each subtitle line's capital tops.
func SubtitleTops(lines int) []int {
	tops := make([]int, lines)
	for i := range tops {
		tops[i] = 250 - 5*(lines-1) + i*subtitleStep
	}
	return tops
}

// drawSubtitle writes the line in the bottom bar. Your thoughts are light
// grey; speech is the speaker's name (mid grey) and two spaces, then the
// words (off-white). In the dream's white bars the colours are inked.
func (a *Art) drawSubtitle(dst *ebiten.Image, e *Engine) {
	words, name := colPaper, color.Color(colMid)
	if e.Thinking() {
		words = colLight
	}
	if e.WhiteBars() {
		words, name = colInk, colSage
	}
	text := strings.ReplaceAll(e.FullText(), "…", "...") // the font has no ellipsis glyph
	label := ""
	if !e.Thinking() && e.Speaker() != "" {
		label = strings.ToUpper(e.Speaker())
	}
	whole := text
	if label != "" {
		whole = label + " " + text // the wrap sees one space; two are drawn
	}
	reveal := len([]rune(strings.ReplaceAll(e.Shown(), "…", "...")))
	if label != "" {
		reveal += len([]rune(label)) + 1
	}
	lines := WrapText(whole, subtitleWidth, a.font.Width)
	tops := SubtitleTops(len(lines))
	for i, line := range lines {
		runes := []rune(line)
		visible := min(max(reveal, 0), len(runes))
		reveal -= len(runes) + 1
		if visible == 0 {
			continue
		}
		y := tops[i] - capTopOffset
		if i == 0 && label != "" {
			rest := string(runes[len([]rune(label))+1:])
			full := label + "  " + rest
			x := screenW/2 - a.font.Width(full)/2
			nameRunes := min(visible, len([]rune(label)))
			a.font.Draw(dst, string(runes[:nameRunes]), x, y, name)
			if visible > len([]rune(label))+1 {
				shown := string(runes[len([]rune(label))+1 : visible])
				a.font.Draw(dst, shown, x+a.font.Width(label+"  "), y, words)
			}
			continue
		}
		x := screenW/2 - a.font.Width(line)/2
		a.font.Draw(dst, string(runes[:visible]), x, y, words)
	}
}

func (a *Art) drawBackground(dst *ebiten.Image, e *Engine, tick int) {
	switch bg := e.BG(); bg {
	case "station":
		_ = a.station.Draw(dst, []string{"wall", "monitor_screen"}, 0, 0, 0)
		if e.Prop("monitor") {
			drawTraces(dst, e)
		}
	case "ceiling":
		offset := ceilingOffset(tick)
		layers := []string{"ceiling"}
		if LampOn(tick) {
			layers = append(layers, "light_on")
		}
		_ = a.ceiling.Draw(dst, layers, 0, 0, offset)
		_ = a.ceiling.Draw(dst, layers, 0, 0, offset-screenH)
	case "black":
		dst.Fill(a.lid())
	case "white":
		dst.Fill(colPaper)
	case "grey":
		dst.Fill(colMid)
	default:
		if def, ok := bgTable[bg]; ok && a.bgs[def.sheet] != nil {
			layers := def.layers
			if bg == "hana_room" && e.Prop("clipboard") { // Hana's route: the clipboard is snapped
				layers = append(append([]string(nil), layers...), "clipboard_snapped")
			}
			if def.cg {
				a.drawCG(dst, e, def, tick)
				return
			}
			if bg == "desk_day" {
				a.drawDeskDay(dst, e)
				return
			}
			switch bg {
			case "hana_floor": // a wide picture: the camera pans across it
				if e.Prop("clipboard") {
					layers = append(append([]string(nil), layers...), "clipboard_snapped")
				}
				_ = a.bgs[def.sheet].Draw(dst, layers, 0, -int(e.PanX()), 0)
				return
			case "report":
				a.drawReport(dst, e, def)
				return
			case "door_closeup":
				_ = a.bgs[def.sheet].Draw(dst, layers, 0, 0, 0)
				a.writeCentred(dst, a.inkHana, HanaStyle, "Hana", cardX, cardBase)
				return
			}
			if bg == "note" || bg == "letter" {
				a.drawPaper(dst, e, def)
				return
			}
			_ = a.bgs[def.sheet].Draw(dst, layers, 0, 0, 0)
			return
		}
		// Batch 2 art is not drawn yet: a black screen with the name.
		dst.Fill(colInk)
		label := "PLACEHOLDER: " + strings.ToUpper(strings.ReplaceAll(bg, "_", " "))
		a.font.Draw(dst, label, 8, 8, colLight)
	}
}

// isCG says whether a background name is a CG.
func (a *Art) isCG(bg string) bool {
	def, ok := bgTable[bg]
	return ok && def.cg && a.bgs[def.sheet] != nil
}

// figureSuffix says which doorway figure the "figure" effect shows: with a
// clipboard and cap (Hana, Mayu's route), head tilted (Mayu, Hana's route),
// or plain (nobody's route).
func figureSuffix(e *Engine) string {
	switch {
	case e.Flags().Has("route_mayu"):
		return "_board"
	case e.Flags().Has("route_hana"):
		return "_tilt"
	}
	return ""
}

// drawCG paints a CG: the sheet's layers in the sheet's own order (back to
// front), those that are always on, those the script switched on, and the
// kanashibari figure when the "figure" effect runs. The frame boils.
func (a *Art) drawCG(dst *ebiten.Image, e *Engine, def bgDef, tick int) {
	sheet := a.bgs[def.sheet]
	figure := ""
	if _, on := e.FxTicks("figure"); on && def.sheet == "cg_kanashibari" {
		figure = "figure" + figureSuffix(e)
	}
	var layers []string
	for _, l := range sheet.Layers {
		if slices.Contains(def.layers, l) || e.HasLayer(l) || l == figure {
			layers = append(layers, l)
		}
	}
	frame := sheet.FrameAt(sprites.BoilTag, time.Duration(tick)*time.Second/TicksPerSecond)
	if def.sheet == "cg_forever_drop" {
		frame = DropFrame(sheet.Meta, time.Duration(e.BGTicks())*time.Second/TicksPerSecond)
	}
	_ = sheet.Draw(dst, layers, frame, 0, 0)
}

// Where Mayu's day sprite (96x128) sits at the next desk in desk_day.
const (
	deskMayuX = 318
	deskMayuY = 92
)

// drawDeskDay paints the Sleep ending's desk: the room, then the next desk
// empty, or Mayu sitting there on her route, then the front of that desk
// (the room has a gap where the desk goes, so the front is always last).
func (a *Art) drawDeskDay(dst *ebiten.Image, e *Engine) {
	desk := a.bgs["desk_day"]
	_ = desk.Draw(dst, []string{"room"}, 0, 0, 0)
	mayu := a.bgs["mayu_day"]
	if e.Flags().Has("route_mayu") && mayu != nil {
		mouth := "mouth_smile"
		if e.Speaker() == "mayu" && !e.LineDone() {
			mouth = "mouth_open" // she is talking
		}
		layers := []string{"fill", "lineart", "detail", "eyes_happy", mouth}
		_ = mayu.Draw(dst, layers, 0, deskMayuX, deskMayuY+e.HopOffset())
	} else {
		_ = desk.Draw(dst, []string{"next_desk_empty"}, 0, 0, 0)
	}
	_ = desk.Draw(dst, []string{"next_desk_front"}, 0, 0, 0)
}

// lid is the colour the screen goes to when it "goes black": always ink (the
// Prologue is the real world, so it goes black like the nights).
func (a *Art) lid() color.RGBA { return colInk }

// ceilingOffset is how far the ceiling has slid down, 0 to 269. It tiles
// every 270 px, so it wraps without a jump.
func ceilingOffset(tick int) int { return (tick * 2) % screenH }

// LampOn is the Prologue's flickering light: on most of the time, with a
// stutter before it goes out and a short blink after.
func LampOn(tick int) bool {
	phase := tick % 60
	return phase < 36 || (phase >= 44 && phase < 48)
}

// ---- the girl ----

// drawFaces draws both girls upside down over the ceiling, for the gurney.
// They lean in from above: the slide uses the "faces" effect's progress.
func (a *Art) drawFaces(dst *ebiten.Image, e *Engine) {
	slide := (1 - easeOut(e.FxProgress("faces"))) * -float64(faceRows)
	drawRotated(dst, a.mayu, e.PoseOf("mayu").Layers(), mayuFaceX, float64(mayuFaceY)+slide+float64(e.FaceHop("mayu")))
	drawRotated(dst, a.hana, e.PoseOf("hana").Layers(), hanaFaceX, float64(hanaFaceY)+slide+float64(e.FaceHop("hana")))
}

func easeOut(t float64) float64 { return 1 - (1-t)*(1-t) }

// drawRotated draws the top faceRows rows of a sprite turned half a circle,
// with its top-left corner (before the turn) at (x, y).
func drawRotated(dst *ebiten.Image, sheet *sprites.Sheet, layers []string, x, y float64) {
	for _, layer := range layers {
		frame, err := sheet.Frame(layer, 0)
		if err != nil {
			continue
		}
		b := frame.Bounds()
		top := frame.SubImage(image.Rect(b.Min.X, b.Min.Y, b.Max.X, b.Min.Y+faceRows)).(*ebiten.Image)
		var op ebiten.DrawImageOptions
		op.GeoM.Translate(-float64(b.Dx())/2, -float64(faceRows)/2)
		op.GeoM.Rotate(math.Pi)
		op.GeoM.Translate(float64(b.Dx())/2, float64(faceRows)/2)
		op.GeoM.Translate(x, y)
		op.ColorScale.ScaleAlpha(sheet.Opacity(layer))
		dst.DrawImage(top, &op)
	}
}

// ---- the ward monitor ----

// ecg is one heartbeat, 30 px wide: how far above the line (in pixels) the
// trace goes at each column.
var ecg = [30]int{0, 0, 0, 0, 1, 1, 0, 0, 0, 0, -1, 3, 4, -1, 0, 0, 0, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

// drawTraces draws the patients' lines in the monitor's screen rect: six
// lines, Matty's (the third) can go flat, and a seventh with no name.
func drawTraces(dst *ebiten.Image, e *Engine) {
	const matty = 2
	lines := 6
	if _, ok := e.FxTicks("seventh"); ok {
		lines = 7
	}
	flat := false
	if _, ok := e.FxTicks("flat_line"); ok {
		flat = true
	}
	scroll := e.Ticks() / 3
	for i := 0; i < lines; i++ {
		base := monitorFirstY + i*monitorLineStep
		c := colTraces
		if i == 6 {
			c = colMid
			if t, ok := e.FxTicks("red_pixel"); ok && t < 30 && t >= 4 {
				c = colBlood // the first red of the game, once
			}
		}
		prev := base
		for x := monitorX0; x < monitorX1; x++ {
			y := base
			if !(flat && i == matty) {
				y -= ecg[(x+scroll*1+i*7)%len(ecg)]
			}
			top, h := min(prev, y), absInt(prev-y)+1
			vector.FillRect(dst, float32(x), float32(top), 1, float32(h), c, false)
			prev = y
		}
	}
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// ---- effects ----

// darkness are the effects that darken the picture at once: "dark" is the
// dark room, "dusk" the red room with only the envelope bright.
var darkness = map[string]float64{"dark": 0.7, "dusk": 0.55}

func (a *Art) drawEffects(dst *ebiten.Image, e *Engine) {
	if _, ok := e.FxTicks("fade_in"); ok {
		if alpha := 1 - e.FxProgress("fade_in"); alpha > 0 {
			c := a.lid()
			vector.FillRect(dst, 0, 0, screenW, screenH, color.NRGBA{c.R, c.G, c.B, uint8(alpha * 255)}, false)
		}
	}
	if _, ok := e.FxTicks("wash"); ok {
		if level := WashLevel(e.FxProgress("wash")); level > 0 {
			if a.wash[level] == nil {
				img := ebiten.NewImage(screenW, screenH)
				img.WritePixels(WashPixels(level, screenW, screenH, [4]byte{colLight.R, colLight.G, colLight.B, 0xff}))
				a.wash[level] = img
			}
			dst.DrawImage(a.wash[level], nil)
		}
	}
	for fx, alpha := range darkness {
		if _, ok := e.FxTicks(fx); ok {
			vector.FillRect(dst, 0, 0, screenW, screenH, color.NRGBA{0x0b, 0x0b, 0x0c, uint8(alpha * 255)}, false)
		}
	}
	if _, ok := e.FxTicks("dim"); ok {
		vector.FillRect(dst, 0, 0, screenW, screenH, color.NRGBA{0x0b, 0x0b, 0x0c, uint8(e.FxProgress("dim") * 0.75 * 255)}, false)
	}
	if _, ok := e.FxTicks("figure"); ok {
		// The faceless figure in the doorway: with a clipboard (Hana, Mayu's
		// route), head tilted (Mayu, Hana's route), or plain (nobody's route).
		layer := "doorway_figure" + figureSuffix(e)
		if sheet := a.bgs["ward_c"]; sheet != nil {
			_ = sheet.Draw(dst, []string{layer}, 0, 0, 0)
		}
	}
	if _, ok := e.FxTicks("eyes_slit"); ok {
		// Almost closed: a thin gap is left in the middle.
		h := float32(e.FxProgress("eyes_slit") * screenH * 0.46)
		vector.FillRect(dst, 0, 0, screenW, h, a.lid(), false)
		vector.FillRect(dst, 0, screenH-h, screenW, h, a.lid(), false)
	}
	if _, ok := e.FxTicks("eyes_open"); ok {
		// The eyelids part: closed at the start, gone at the end.
		if h := float32((1 - e.FxProgress("eyes_open")) * screenH / 2); h > 0 {
			vector.FillRect(dst, 0, 0, screenW, h, a.lid(), false)
			vector.FillRect(dst, 0, screenH-h, screenW, h, a.lid(), false)
		}
	}
	if _, ok := e.FxTicks("eyes_close"); ok {
		// The eyelids come in from the top and the bottom and meet in the middle.
		h := float32(e.FxProgress("eyes_close") * screenH / 2)
		vector.FillRect(dst, 0, 0, screenW, h, a.lid(), false)
		vector.FillRect(dst, 0, screenH-h, screenW, h, a.lid(), false)
	}
}

// ---- text box, name plate, choices ----

func (a *Art) drawTextBox(dst *ebiten.Image, e *Engine) {
	_ = a.ui.Draw(dst, []string{"textbox"}, 0, 0, 0)
	colour := color.Color(colPaper)
	if e.Thinking() {
		colour = colLight // your thoughts: light grey, no name plate
	} else if e.Speaker() != "" {
		_ = a.ui.Draw(dst, []string{"nameplate"}, 0, 0, 0)
		a.drawName(dst, e.Speaker())
	}
	full := strings.ReplaceAll(e.FullText(), "…", "...") // the font has no ellipsis glyph
	// The reveal is counted on the replaced text, so "…" types as three dots.
	reveal := len([]rune(strings.ReplaceAll(e.Shown(), "…", "...")))
	for i, line := range WrapText(full, textWidth, a.font.Width) {
		if i >= maxTextLines || reveal <= 0 {
			break
		}
		runes := []rune(line)
		n := min(reveal, len(runes))
		a.font.Draw(dst, string(runes[:n]), textX, textCapTop-capTopOffset+i*lineStep, colour)
		reveal -= len(runes) + 1 // the space the wrap dropped
	}
	autoLayer := "auto_off"
	if e.AutoPlay {
		autoLayer = "auto_on"
	}
	_ = a.ui.Draw(dst, []string{autoLayer}, a.ui.FrameAt("bob", time.Duration(e.Ticks())*time.Second/TicksPerSecond), 0, 0)
	// The toggle's label: mid grey when AUTO is off, off-white when it is on.
	label := color.Color(colMid)
	if e.AutoPlay {
		label = colPaper
	}
	a.font.Draw(dst, "AUTO", autoLabelX, autoLabelCapTop-capTopOffset, label)
	if e.LineDone() && len(e.Options()) == 0 {
		frame := a.ui.FrameAt("bob", time.Duration(e.Ticks())*time.Second/TicksPerSecond)
		_ = a.ui.Draw(dst, []string{"continue"}, frame, 0, 0)
	}
}

// drawName writes the speaker on the name plate: in her own handwriting if
// she has a sheet (Mayu, Hana), in the pixel font otherwise.
func (a *Art) drawName(dst *ebiten.Image, speaker string) {
	name := strings.ToUpper(speaker[:1]) + speaker[1:]
	sheet, style := a.handFor(speaker)
	if sheet == nil {
		a.font.Draw(dst, strings.ToUpper(speaker), (plateX0+plateX1)/2-a.font.Width(strings.ToUpper(speaker))/2,
			plateCapTop-capTopOffset, colPaper)
		return
	}
	placed, width := Layout(name, SheetMetrics{sheet.Meta}, style)
	x0 := plateX0 + (plateX1-plateX0+1-width)/2
	// A glyph cell's baseline is row 8 and capitals start at row 2, so the
	// cell's top sits 2 rows above the capital tops.
	y0 := plateCapTop - 2
	for _, p := range placed {
		_ = sheet.Draw(dst, []string{p.Glyph}, 0, x0+p.X, y0+p.Y)
	}
}

func (a *Art) handFor(speaker string) (*sprites.Sheet, Style) {
	switch speaker {
	case "mayu":
		return a.handMayu, MayuStyle
	case "hana":
		return a.handHana, HanaStyle
	}
	return nil, Style{}
}

func (a *Art) drawChoices(dst *ebiten.Image, e *Engine) {
	for i, rect := range ChoiceRects(len(e.Options())) {
		layer, colour := "normal", color.Color(colLight)
		if i == e.Hover() {
			layer, colour = "hover", colPaper
		}
		_ = a.choice.Draw(dst, []string{layer}, 0, rect.Min.X, rect.Min.Y)
		label := strings.ToUpper(e.Options()[i].Text)
		a.font.Draw(dst, label, rect.Min.X+(choiceW-a.font.Width(label))/2, rect.Min.Y+9-capTopOffset, colour)
	}
}

// SheetMetrics lets Layout read glyph widths from a handwriting sheet's
// layer user data.
type SheetMetrics struct{ meta *sprites.Meta }

func (m SheetMetrics) Advance(glyph string) (adv int, canJoin, ok bool) {
	data := m.meta.LayerData(glyph)
	if data == "" {
		return 0, false, false
	}
	adv, canJoin, err := ParseGlyphData(data)
	return adv, canJoin, err == nil
}

// WrapText breaks text into lines no wider than width, at spaces. measure
// gives the pixel width of a string. A word wider than a line stays whole.
func WrapText(text string, width int, measure func(string) int) []string {
	var lines []string
	line := ""
	for _, word := range strings.Fields(text) {
		try := word
		if line != "" {
			try = line + " " + word
		}
		if line != "" && measure(try) > width {
			lines = append(lines, line)
			line = word
			continue
		}
		line = try
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

// ---- notes and letters, written on the paper ----

// drawPaper paints a note or a letter: the paper, the writer's ink and then
// the hands holding it. The ink is every line so far in the writer's hand;
// the line being typed is written glyph by glyph.
func (a *Art) drawPaper(dst *ebiten.Image, e *Engine, def bgDef) {
	lines := PaperText(e.Paper())
	written := 0 // letters the pen has written so far
	for i, line := range lines {
		if i < len(lines)-1 || !e.OnPaper() || e.LineDone() {
			written += countLetters(line)
			continue
		}
		reveal := len([]rune(PaperText([]string{e.Shown()})[0]))
		written += Written(line, reveal)
	}
	a.paperPicture(dst, def, e.BG(), e.Writer(), lines, written, e.LineDone())
}

// paperPicture paints a sheet of paper with the first `written` letters of
// the lines on it. signed says whether the note carries Hana's "- H.".
func (a *Art) paperPicture(dst *ebiten.Image, def bgDef, bg, writer string, lines []string, written int, signed bool) {
	sheet := a.bgs[def.sheet]
	_ = sheet.Draw(dst, def.layers[:2], 0, 0, 0) // backdrop, paper
	hand, style := a.inkHana, HanaStyle
	if writer == "mayu" {
		hand, style = a.inkMayu, MayuStyle
	}
	spec := LetterPaper
	if bg == "note" {
		spec = NotePaper
	}
	metrics := SheetMetrics{hand.Meta}
	for _, row := range PaperRows(lines, spec, metrics, style) {
		for _, p := range row.Glyphs {
			if row.First+p.Char <= written {
				_ = hand.Draw(dst, []string{p.Glyph}, 0, row.X+p.X, row.Base-glyphBaseRow+p.Y)
			}
		}
	}
	if bg == "note" && signed {
		// Hana signs her note; the signature is not a line of the script.
		sign, w := Layout(noteSignature, metrics, style)
		for _, p := range sign {
			_ = hand.Draw(dst, []string{p.Glyph}, 0, spec.Right-w+p.X, noteSignBase-glyphBaseRow+p.Y)
		}
	}
	_ = sheet.Draw(dst, def.layers[2:], 0, 0, 0) // hands
}

// PaperText is what is written for the lines of the script: "<3" is drawn
// as the heart glyph.
func PaperText(lines []string) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = strings.ReplaceAll(l, "<3", "♥")
	}
	return out
}

const (
	glyphBaseRow  = 8 // a glyph cell's baseline is its row 8 (see drawName)
	noteSignature = "- H."
	noteSignBase  = 138
)

// ---- the day 3 report (the realization's match cut) ----

// Where the report page lies on the look-down desk (the same place the
// folder's page has), and the desk's look.
const reportX, reportY = 186, 41

// drawReport paints the desk from above with Hana's report on it, written
// in the hand ReportHand names. It is a day picture, so the scene puts the
// CRT on it (DayLook).
func (a *Art) drawReport(dst *ebiten.Image, e *Engine, def bgDef) {
	desk := a.bgs["desk_top"]
	frame := desk.FrameAt(sprites.BoilTag, time.Duration(e.Ticks())*time.Second/TicksPerSecond)
	_ = desk.Draw(dst, []string{"fill", "lineart", "detail", "memo_day3"}, frame, 0, 0)
	page := a.bgs[def.sheet]
	_ = page.Draw(dst, def.layers, page.FrameAt(sprites.BoilTag, time.Duration(e.Ticks())*time.Second/TicksPerSecond), reportX, reportY)
	hand, style := a.inkHana, HanaStyle
	if sheet, _ := ReportHand(e.Flags()); sheet == "hand_mayu" {
		hand, style = a.inkMayu, MayuStyle
	}
	_ = WriteReport(dst, hand, style, reportX, reportY)
}

// ---- the split-screen sting ----

// Sting: the letter's signature above, the report's below, both at x4
// (Figma: Cine - Opt 1 . B4). The strip shown is rows 68-201 of the x4 crop.
const (
	stingScale = 4
	stingRow0  = 68
	stingRows  = 134
)

// stingSignature are the centres the two crops are taken on: the letter's
// and the report's signature.
var (
	stingLetterAt = [2]float64{350, 89}
	stingReportAt = [2]float64{301, 196}
)

// drawSting fills the screen with the two signatures one above the other, a
// 2 px divider between. They are the same words in the same hand, so they
// match pixel for pixel. It covers the whole screen (bars and words too).
func (a *Art) drawSting(dst *ebiten.Image, e *Engine) {
	if a.world == nil {
		a.world = ebiten.NewImage(screenW, screenH)
		a.zoomed = ebiten.NewImage(screenW, screenH)
		a.keep = ebiten.NewImage(screenW, screenH)
	}
	dst.Fill(colInk)
	letter, report := bgTable["letter"], bgTable["report"]
	lines := PaperText(e.LastPaper())
	all := 0
	for _, l := range lines {
		all += countLetters(l)
	}
	a.world.Clear()
	a.paperPicture(a.world, letter, "letter", e.LastWriter(), lines, all, true)
	a.stingStrip(dst, a.world, stingLetterAt, 0)
	a.world.Clear()
	a.drawReport(a.world, e, report)
	a.stingStrip(dst, a.world, stingReportAt, stingRows+2)
}

// stingStrip draws the middle rows of src at x4 around the point into dst,
// starting at row y.
func (a *Art) stingStrip(dst, src *ebiten.Image, at [2]float64, y int) {
	strip := dst.SubImage(image.Rect(0, y, screenW, y+stingRows)).(*ebiten.Image)
	ox, oy := CropOrigin(stingScale, at[0], at[1])
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(-ox, -oy)
	op.GeoM.Scale(stingScale, stingScale)
	op.GeoM.Translate(0, float64(y-stingRow0))
	strip.DrawImage(src, &op)
}
