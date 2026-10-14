package game

import "image/color"

// paletteInk is the night's black (assets/palette/danmin.hex, entry 1).
var paletteInk = color.RGBA{0x0b, 0x0b, 0x0c, 0xff}

// colPaperAlpha is the palette's off-white at a strength from 0 to 1 (the
// dream's white fading out).
func colPaperAlpha(a float64) color.NRGBA {
	return color.NRGBA{0xec, 0xec, 0xe6, uint8(a * 255)}
}
