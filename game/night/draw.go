package night

import (
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"io/fs"
	"math"
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
	scratch            *ebiten.Image  // the finished scene, when a shake or the redscale must reshape it
	red                *ebiten.Shader // redscale.kage
	mayu, hana         *sprites.Sheet
	station, ceiling   *sprites.Sheet
	ui, choice         *sprites.Sheet
	handMayu, handHana *sprites.Sheet
	font               TextDrawer
	bgs                map[string]*sprites.Sheet // the batch 2 backgrounds that have a sheet, by sheet name
}

// A background name in a script is drawn from the layers of one sheet. A
// name that is not here (and not one of the built-in fills) is a placeholder
// until its sheet exists in game/assets/sprites.
type bgDef struct {
	sheet  string
	layers []string
}

var bgTable = map[string]bgDef{
	"ward_c":      {"ward_c", []string{"room", "sleepers", "blanket"}},
	"corridor":    {"corridor", []string{"corridor", "sign_station"}},
	"hana_far":    {"corridor", []string{"corridor", "sign_station", "hana_far"}},
	"hana_door":   {"corridor", []string{"corridor", "sign_quarters", "door_crack"}},
	"supply_room": {"supply", []string{"wall", "bulb_glow", "table", "prop_cans", "prop_syringe"}},
	"hana_room":   {"hana_room", []string{"room", "fx_blood", "cans", "hana_silhouette"}},
	"note":        {"note", []string{"backdrop", "paper_note", "text_note", "hands_note"}},
	"letter":      {"note", []string{"backdrop", "paper_letter", "hands_letter"}},
	"ward_c_dark": {"ward_c", []string{"room", "sleepers", "blanket", "dark"}},
	"ward_c_up":   {"ward_c", []string{"room", "sitters", "red_eyes", "clock_317"}},
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
	for _, def := range append([]bgDef{{sheet: BloodSheet}}, bgTableDefs()...) {
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
		"vn_ui": &a.ui, "vn_choice": &a.choice, "hand_mayu": &a.handMayu, "hand_hana": &a.handHana,
	} {
		sheet, err := sprites.Load(fsys, name)
		if err != nil {
			return nil, fmt.Errorf("night art: %w", err)
		}
		*dst = sheet
	}
	// The handwriting is black ink for paper; the name plate is black, so the
	// plate uses copies with the ink off-white and the joins mid grey.
	onPlate := map[color.RGBA]color.RGBA{colInk: colPaper, colDark: colMid}
	a.handMayu, a.handHana = a.handMayu.Remapped(onPlate), a.handHana.Remapped(onPlate)
	return a, nil
}

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

// Draw paints the whole night screen for the engine's current state. Order
// (Art log): wall + monitor, traces, the girl, the counter and props, then
// the effects and the UI on top.
func (a *Art) Draw(dst *ebiten.Image, e *Engine) {
	_, shaking := e.FxTicks("shake")
	shaking = shaking && e.FxProgress("shake") < 1
	rise, drain := e.RedscaleAmounts()
	if !shaking && rise == 0 {
		a.drawScene(dst, e)
		a.drawKeepWhite(dst, e)
		return
	}
	// The shake and the redscale work on the finished picture: draw it
	// aside, then put it through the shader.
	if a.scratch == nil {
		a.scratch = ebiten.NewImage(screenW, screenH)
	}
	a.drawScene(a.scratch, e)
	dx, dy := 0, 0
	if shaking {
		ticks, _ := e.FxTicks("shake")
		dx, dy = ShakeOffset(ticks)
	}
	dst.Fill(colInk)
	op := &ebiten.DrawRectShaderOptions{}
	op.Images[0] = a.scratch
	op.GeoM.Translate(float64(dx), float64(dy))
	op.Uniforms = map[string]any{"Rise": float32(rise), "Drain": float32(drain)}
	dst.DrawRectShader(screenW, screenH, a.red, op)
	a.drawKeepWhite(dst, e)
}

// drawKeepWhite draws the envelope on Hana's desk after the redscale, so it
// is the one thing in the room that stays white (Art Direction -> Hana's
// room redscale).
func (a *Art) drawKeepWhite(dst *ebiten.Image, e *Engine) {
	if sheet := a.bgs["hana_room"]; sheet != nil && e.BG() == "hana_room" && e.Prop("envelope") {
		_ = sheet.Draw(dst, []string{"envelope"}, 0, 0, 0)
	}
}

// drawScene paints the night screen without the shake or the redscale.
func (a *Art) drawScene(dst *ebiten.Image, e *Engine) {
	tick := e.Ticks()
	a.drawBackground(dst, e, tick)
	switch {
	case e.Faces():
		a.drawFaces(dst, e)
	case e.Visible() != "":
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
	if t, ok := e.FxTicks("splash"); ok {
		a.DrawBlood(dst, t)
	}
	if e.HasLine() {
		a.drawTextBox(dst, e)
	}
	if len(e.Options()) > 0 {
		a.drawChoices(dst, e)
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
		dst.Fill(colInk)
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
			_ = a.bgs[def.sheet].Draw(dst, layers, 0, 0, 0)
			return
		}
		// Batch 2 art is not drawn yet: a black screen with the name.
		dst.Fill(colInk)
		label := "PLACEHOLDER: " + strings.ToUpper(strings.ReplaceAll(bg, "_", " "))
		a.font.Draw(dst, label, 8, 8, colLight)
	}
}

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

func (a *Art) drawEffects(dst *ebiten.Image, e *Engine) {
	if _, ok := e.FxTicks("fade_in"); ok {
		if alpha := 1 - e.FxProgress("fade_in"); alpha > 0 {
			vector.FillRect(dst, 0, 0, screenW, screenH, color.NRGBA{0x0b, 0x0b, 0x0c, uint8(alpha * 255)}, false)
		}
	}
	if _, ok := e.FxTicks("dim"); ok {
		vector.FillRect(dst, 0, 0, screenW, screenH, color.NRGBA{0x0b, 0x0b, 0x0c, uint8(e.FxProgress("dim") * 0.75 * 255)}, false)
	}
	if _, ok := e.FxTicks("figure"); ok {
		// The faceless figure in the doorway: with a clipboard (Hana, Mayu's
		// route), head tilted (Mayu, Hana's route), or plain (nobody's route).
		layer := "doorway_figure"
		switch {
		case e.Flags().Has("route_mayu"):
			layer = "doorway_figure_board"
		case e.Flags().Has("route_hana"):
			layer = "doorway_figure_tilt"
		}
		if sheet := a.bgs["ward_c"]; sheet != nil {
			_ = sheet.Draw(dst, []string{layer}, 0, 0, 0)
		}
	}
	if _, ok := e.FxTicks("eyes_slit"); ok {
		// Almost closed: a thin gap is left in the middle.
		h := float32(e.FxProgress("eyes_slit") * screenH * 0.46)
		vector.FillRect(dst, 0, 0, screenW, h, colInk, false)
		vector.FillRect(dst, 0, screenH-h, screenW, h, colInk, false)
	}
	if _, ok := e.FxTicks("eyes_open"); ok {
		// The eyelids part: closed at the start, gone at the end.
		if h := float32((1 - e.FxProgress("eyes_open")) * screenH / 2); h > 0 {
			vector.FillRect(dst, 0, 0, screenW, h, colInk, false)
			vector.FillRect(dst, 0, screenH-h, screenW, h, colInk, false)
		}
	}
	if _, ok := e.FxTicks("eyes_close"); ok {
		// The eyelids come in from the top and the bottom and meet in the middle.
		h := float32(e.FxProgress("eyes_close") * screenH / 2)
		vector.FillRect(dst, 0, 0, screenW, h, colInk, false)
		vector.FillRect(dst, 0, screenH-h, screenW, h, colInk, false)
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
	placed, width := Layout(name, sheetMetrics{sheet.Meta}, style)
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

// sheetMetrics lets Layout read glyph widths from a handwriting sheet's
// layer user data.
type sheetMetrics struct{ meta *sprites.Meta }

func (m sheetMetrics) Advance(glyph string) (adv int, canJoin, ok bool) {
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
