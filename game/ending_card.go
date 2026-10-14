package game

import (
	"image"
	"image/color"
	"io/fs"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"danmingame/night"
	"danmingame/sprites"
)

// The ending card (Game Design -> design log, "Ending cards"): the empty
// card with the ending's emblem, the title LARGE and in capitals in the
// font_ending glyph sheet, a small line under the rule, and the girl of that
// ending standing to the side. Forever's title is blood red.

const (
	cardTitleY    = 153 // top of the glyph cells: the capitals sit on a baseline at y 180, just over the card's rule
	cardLabelY    = 194 // small line under the rule
	cardSubY      = 208
	cardWordSpace = 12
	cardGap       = 1
)

// cardTitles are the large titles, and cardLabels the small line under them.
var (
	cardTitles = map[string]string{
		night.EndingSleep: "SLEEP", night.EndingWake: "WAKE", night.EndingForever: "FOREVER", night.Ending317: "3:17",
	}
	cardLabels = map[string]string{
		night.EndingSleep: "ENDING", night.EndingWake: "ENDING", night.EndingForever: "ENDING", night.Ending317: "SECRET ENDING",
	}
)

// cardGirl is the girl who stands beside an ending's card: her sheet, the
// layers of the pose and where her top-left corner goes. The secret ending
// has nobody.
type cardGirl struct {
	sheet  string
	layers []string
	x, y   int
	boils  bool // the sheet has the 3-frame line boil
	// backing is a pale panel drawn behind her: the day look is black
	// lineart made for a light room, so it needs a light wall on the dark card.
	backing image.Rectangle
}

var cardGirls = map[string]cardGirl{
	// Mayu in the day look, asleep at the window, smiling.
	night.EndingSleep: {"mayu_day", []string{"fill", "lineart", "detail", "eyes_closed", "mouth_smile", "fx_zzz"}, 344, 124, true, image.Rect(344, 140, 440, 252)},
	// Mayu awake, smiling, waving.
	night.EndingWake: {"mayu", []string{"hair_back", "body", "head", "eyes_open", "brows_neutral", "mouth_smile",
		"arm_l_down", "arm_r_wave"}, 296, 14, false, image.Rectangle{}},
	// Hana with her arms crossed, crying.
	night.EndingForever: {"hana", []string{"hair_back", "body", "head", "eyes_closed", "brows_worried", "mouth_frown",
		"fx_tears", "arm_l_cross", "arm_r_cross"}, 312, 14, false, image.Rectangle{}},
}

// Blood red for Forever's title: the font's two off-whites become the blood
// palette's red and dark red.
var bloodRemap = map[color.RGBA]color.RGBA{
	{0xec, 0xec, 0xe6, 0xff}: {0x8a, 0x03, 0x03, 0xff},
	{0xb9, 0xb9, 0xb4, 0xff}: {0x4a, 0x00, 0x00, 0xff},
}

// cardArt is everything the card draws with.
type cardArt struct {
	cards     *sprites.Sheet
	font      *sprites.Sheet // the ending font, off-white
	fontRed   *sprites.Sheet // the same glyphs in blood red
	girl      *sprites.Sheet // nil when this ending has nobody beside the card
	girlStyle cardGirl
}

func loadCardArt(fsys fs.FS, ending string) (*cardArt, error) {
	a := &cardArt{}
	var err error
	if a.cards, err = LoadSheet(fsys, "ending_cards"); err != nil {
		return nil, err
	}
	if a.font, err = LoadSheet(fsys, "font_ending"); err != nil {
		return nil, err
	}
	a.fontRed = a.font.Remapped(bloodRemap)
	if g, ok := cardGirls[ending]; ok {
		a.girlStyle = g
		if a.girl, err = LoadSheet(fsys, g.sheet); err != nil {
			return nil, err
		}
	}
	return a, nil
}

// endingMetrics gives night.Layout the width of each ending-font glyph.
type endingMetrics struct{ meta *sprites.Meta }

func (m endingMetrics) Advance(glyph string) (adv int, canJoin, ok bool) {
	data := m.meta.LayerData(glyph)
	if data == "" {
		return 0, false, false
	}
	adv, canJoin, err := night.ParseGlyphData(data)
	return adv, canJoin, err == nil
}

// endingStyle is the title's spacing: plain, no joins and no jitter.
var endingStyle = night.Style{Gap: cardGap, WordSpace: cardWordSpace}

// titleWidth is how wide a large title is, in pixels.
func (a *cardArt) titleWidth(title string) int {
	_, w := night.Layout(title, endingMetrics{a.font.Meta}, endingStyle)
	return w
}

// drawTitle writes the large title centred, with the glyph cells' top at y.
func (a *cardArt) drawTitle(dst *ebiten.Image, title string, y int, red bool) {
	sheet := a.font
	if red {
		sheet = a.fontRed
	}
	drawBigText(dst, sheet, title, y, 0)
}

// drawBigText writes text in the ending font, centred (then moved dx
// pixels), with the glyph cells' top at y. It is the handwriting layout
// without joins or jitter.
func drawBigText(dst *ebiten.Image, font *sprites.Sheet, text string, y, dx int) {
	placed, w := night.Layout(text, endingMetrics{font.Meta}, endingStyle)
	x0 := (ScreenW-w)/2 + dx
	for _, p := range placed {
		_ = font.Draw(dst, []string{p.Glyph}, 0, x0+p.X, y)
	}
}

// draw paints the card for an ending, ticks after it came up. Like the rest
// of the game's drawing it ignores a layer the sheet does not have.
func (a *cardArt) draw(dst *ebiten.Image, ending string, ticks int) {
	_ = a.cards.Draw(dst, []string{"card_bg"}, 0, 0, 0)
	_ = a.cards.Draw(dst, []string{"emblem_" + ending}, 0, 0, 0)
	if a.girl != nil {
		frame := 0
		if a.girlStyle.boils {
			frame = (ticks / 22) % 3
		}
		if r := a.girlStyle.backing; !r.Empty() {
			vector.FillRect(dst, float32(r.Min.X), float32(r.Min.Y), float32(r.Dx()), float32(r.Dy()), colPaperRGBA, false)
		}
		_ = a.girl.Draw(dst, a.girlStyle.layers, frame, a.girlStyle.x, a.girlStyle.y)
	}
	a.drawTitle(dst, cardTitles[ending], cardTitleY, ending == night.EndingForever)
}

// cardSmall is the small text under the rule: the label and the subtitle.
func cardSmall(ending string) (label, subtitle string) {
	return strings.ToUpper(cardLabels[ending]), endingSubtitle
}
