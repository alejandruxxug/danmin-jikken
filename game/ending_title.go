package game

import (
	"image/color"
	"io/fs"

	"github.com/hajimehoshi/ebiten/v2"

	"danmingame/night"
	"danmingame/sprites"
)

// The ending title (Art Direction -> Cinematic pass): when the rock slams in,
// the ending's last picture comes back and the title is written over it, on a
// dither shade, in the font_ending glyph sheet. There is no separate card.

const (
	titleCellTop = 176 // top of the glyph cells: the capitals' baseline is at y 203
	labelCapTop  = 170 // "ENDING" caps top
	subtitleTop  = 212 // the place line's caps top
	titleWord    = 12  // pixels between the words of a title
	titleGap     = 1
)

// titleWords are the large titles, and titleLabels the small line above them.
var (
	titleWords = map[string]string{
		night.EndingSleep: "SLEEP", night.EndingWake: "WAKE", night.EndingForever: "FOREVER", night.Ending317: "3:17",
	}
	titleLabels = map[string]string{
		night.EndingSleep: "ENDING", night.EndingWake: "ENDING", night.EndingForever: "ENDING", night.Ending317: "SECRET ENDING",
	}
)

// endingSubtitle is the line under every title (the one in the Sound note).
const endingSubtitle = "Halcyon Rest, Ward B - 3:17 AM"

// The font_ending glyphs are off-white and light grey. Each ending recolours
// them from the palette: Sleep's are ink on the white dream, Forever's blood
// red (#c41a1a over #8a0303) with a black outline.
var (
	colBrightRed = color.RGBA{0xc4, 0x1a, 0x1a, 0xff}
	colBlood     = color.RGBA{0x8a, 0x03, 0x03, 0xff}
	colDarkGrey  = color.RGBA{0x2e, 0x2e, 0x33, 0xff}
	colOffWhite  = color.RGBA{0xec, 0xec, 0xe6, 0xff}
	colLightGrey = color.RGBA{0xb9, 0xb9, 0xb4, 0xff}
)

func remap(from1, to1, from2, to2 color.RGBA) map[color.RGBA]color.RGBA {
	return map[color.RGBA]color.RGBA{from1: to1, from2: to2}
}

// titleArt is everything the title is drawn with.
type titleArt struct {
	font    *sprites.Sheet // the title glyphs in this ending's colours
	outline *sprites.Sheet // Forever only: the same glyphs in black
	shade   *sprites.Sheet // fx_cinema: its "shade" layer is the dither band
	twice   bool           // Forever's shade is drawn twice, to be denser
	small   color.RGBA     // colour of the small lines
}

func loadTitleArt(fsys fs.FS, ending string) (*titleArt, error) {
	font, err := LoadSheet(fsys, "font_ending")
	if err != nil {
		return nil, err
	}
	cinema, err := LoadSheet(fsys, "fx_cinema")
	if err != nil {
		return nil, err
	}
	a := &titleArt{font: font, shade: cinema, small: colMidRGBA}
	switch ending {
	case night.EndingSleep: // ink on the white dream
		a.font = font.Remapped(remap(colOffWhite, colInk, colLightGrey, colDarkGrey))
		a.shade = cinema.Remapped(map[color.RGBA]color.RGBA{colInk: colOffWhite}) // the black dither becomes off-white
		a.small = colDarkGrey
	case night.EndingForever:
		a.font = font.Remapped(remap(colOffWhite, colBrightRed, colLightGrey, colBlood))
		a.outline = font.Remapped(remap(colOffWhite, colInk, colLightGrey, colInk))
		a.twice = true
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
var endingStyle = night.Style{Gap: titleGap, WordSpace: titleWord}

// titleWidth is how wide a title is, in pixels.
func (a *titleArt) titleWidth(title string) int {
	_, w := night.Layout(title, endingMetrics{a.font.Meta}, endingStyle)
	return w
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

// draw paints the shade and the title. Forever's word has a black outline
// (the glyphs in black one pixel to each side) and a 2 px drop, then the red
// on top, so it reads on the red stain.
func (a *titleArt) draw(dst *ebiten.Image, ending string) {
	_ = a.shade.Draw(dst, []string{"shade"}, 0, 0, 0)
	if a.twice {
		_ = a.shade.Draw(dst, []string{"shade"}, 0, 0, 0)
	}
	word := titleWords[ending]
	if a.outline != nil {
		for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}, {1, 1}, {2, 2}} {
			drawBigText(dst, a.outline, word, titleCellTop+d[1], d[0])
		}
	}
	drawBigText(dst, a.font, word, titleCellTop, 0)
}
