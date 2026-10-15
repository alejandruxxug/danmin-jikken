package game

import (
	"image"
	"image/color"
	"io/fs"
	"math/rand"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"danmingame/sprites"
)

// finaleArt is everything the finale is drawn with.
type finaleArt struct {
	sheet *sprites.Sheet // finale: the shade, the light logos, the fade
	title *sprites.Sheet // title: the logo, the dagger and the blood
	small *Font          // Silkscreen 8: the roll
	big   *Font          // Silkscreen 16: the thanks
	style finaleStyle
	lines []rollLine
}

const bigCapTop = 2 * capTop // a size 16 capital starts twice as far down

func newFinaleArt(fsys fs.FS, ending string, cards []creditCard) (*finaleArt, error) {
	var err error
	f := &finaleArt{style: finaleStyleFor(ending), lines: rollLayout(cards)}
	if f.sheet, err = LoadSheet(fsys, "finale"); err != nil {
		return nil, err
	}
	if f.title, err = LoadSheet(fsys, "title"); err != nil {
		return nil, err
	}
	if f.small, err = LoadFont(); err != nil {
		return nil, err
	}
	if f.big, err = LoadFontAt(2 * fontSize); err != nil {
		return nil, err
	}
	return f, nil
}

// draw paints the finale's own layers over the picture: the shade, then the
// roll, the thanks or the logo. The fade is drawn on top by drawFade.
func (f *finaleArt) draw(dst *ebiten.Image, st FinaleState) {
	if frame := shadeFrame(st, f.style.ShadeMax); frame >= 0 {
		_ = f.sheet.Draw(dst, []string{f.style.ShadeLayer}, frame, 0, 0)
	}
	switch st.Stage {
	case "shade-in", "roll", "thanks":
		f.drawRoll(dst, rollDY(st.At))
	case "logo", "fade": // the logo holds, then the fade goes over it
		f.drawLogo(dst)
	}
}

// drawFade lays the dither of the fade over everything (frame 3 is solid).
func (f *finaleArt) drawFade(dst *ebiten.Image, st FinaleState) {
	switch frame := fadeFrame(st); {
	case frame == 3:
		dst.Fill(paletteInk)
	case frame >= 0:
		_ = f.sheet.Draw(dst, []string{"fade_black"}, frame, 0, 0)
	}
}

// drawRoll writes the lines that are on the screen, the roll climbed dy pixels.
func (f *finaleArt) drawRoll(dst *ebiten.Image, dy int) {
	for _, l := range f.lines {
		y := rollStartY + l.Y - dy // caps top
		if y > ScreenH || y < -12 {
			continue
		}
		font, top, c := f.small, capTop, f.style.Name
		switch {
		case l.Thanks:
			font, top, c = f.big, bigCapTop, f.style.Thanks
		case l.Heading:
			c = f.style.Head
		}
		text := strings.ToUpper(l.Text)
		x, ty := (ScreenW-font.Width(text))/2, y-top
		switch {
		case l.Thanks && f.style.ThanksEdge, l.Heading && f.style.HeadOutline:
			drawOutlined(dst, font, text, x, ty, c)
		case l.Heading && f.style.HeadShadow:
			font.Draw(dst, text, x+1, ty+1, colDim)
			font.Draw(dst, text, x, ty, c)
		default:
			font.Draw(dst, text, x, ty, c)
		}
	}
}

// drawOutlined prints text with a 1 px black outline: the same text in black
// at the eight neighbours, then the colour on top.
func drawOutlined(dst *ebiten.Image, f *Font, text string, x, y int, c color.Color) {
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx != 0 || dy != 0 {
				f.Draw(dst, text, x+dx, y+dy, colInk)
			}
		}
	}
	f.Draw(dst, text, x, y, c)
}

// drawLogo is the game's logo, still, logoDY pixels lower than on the title
// screen. Forever and 3:17 use the light logo layers of the finale sheet with
// a coloured shadow; Sleep and Wake the title's own black logo. All four use
// the title's dagger and blood.
func (f *finaleArt) drawLogo(dst *ebiten.Image) {
	layers := func(sheet *sprites.Sheet, names ...string) {
		for _, name := range names {
			_ = sheet.Draw(dst, []string{name}, 0, 0, logoDY)
		}
	}
	switch f.style.Logo {
	case logoBlood:
		layers(f.sheet, "logo_shadow_blood", "logo_kanji_light", "logo_romaji_light")
	case logoDark:
		layers(f.sheet, "logo_shadow_dark", "logo_kanji_light", "logo_romaji_light")
	default:
		layers(f.title, "logo_shadow", "logo_kanji", "logo_romaji")
	}
	layers(f.title, "dagger", "blood")
}

// The secret ending's finale slips now and then: one 8 px row band of the roll
// moves 3 to 6 px sideways for two frames (fx_glitch's timing).
const (
	slipWindow = 45 // ticks between two chances
	slipBand   = 8
)

// bandSlip says which band (its top y) is slipped at a tick and by how many
// pixels (negative: to the left); ok is false when none is. It is a pure
// function of the tick and the seed, so a screenshot is repeatable.
func bandSlip(tick int, seed int64) (y, dx int, ok bool) {
	if tick%slipWindow >= 2 {
		return 0, 0, false
	}
	rng := rand.New(rand.NewSource(seed + int64(tick/slipWindow)*6151))
	if rng.Float64() >= 0.6 {
		return 0, 0, false
	}
	y = rng.Intn(ScreenH/slipBand) * slipBand
	dx = 3 + rng.Intn(4)
	if rng.Intn(2) == 0 {
		dx = -dx
	}
	return y, dx, true
}

// drawSlipped copies the finale layer onto dst; the slipped band moves sideways.
func drawSlipped(dst, layer *ebiten.Image, tick int, seed int64) {
	slipY, slipDX, slipped := bandSlip(tick, seed)
	var op ebiten.DrawImageOptions
	if !slipped {
		dst.DrawImage(layer, &op)
		return
	}
	for _, r := range []image.Rectangle{
		image.Rect(0, 0, ScreenW, slipY),
		image.Rect(0, slipY, ScreenW, slipY+slipBand),
		image.Rect(0, slipY+slipBand, ScreenW, ScreenH),
	} {
		if r.Empty() {
			continue
		}
		op.GeoM.Reset()
		x := 0
		if r.Min.Y == slipY {
			x = slipDX
		}
		op.GeoM.Translate(float64(x), float64(r.Min.Y))
		dst.DrawImage(layer.SubImage(r).(*ebiten.Image), &op)
	}
}
