package main

// The picture the trailer is made on. Everything is placed in whole 1x game
// pixels (the 480x270 canvas); the output is 1920x1080, so one game pixel is a
// 4x4 block. Recorded game frames are already 1920x1080 (the game drew them
// with its CRT), so they are copied block by block; the cards and the motion
// graphics come from sprites and are scaled by drawing each pixel as a block.

import (
	"image"
	"image/color"
)

const (
	Scale = 4 // output pixels per game pixel
	GameW = 480
	GameH = 270
	OutW  = GameW * Scale
	OutH  = GameH * Scale
)

// NewCanvas is a black 1920x1080 picture.
func NewCanvas() *image.RGBA {
	c := image.NewRGBA(image.Rect(0, 0, OutW, OutH))
	Fill(c, image.Rect(0, 0, GameW, GameH), color.RGBA{0, 0, 0, 255})
	return c
}

// outRect turns a rectangle in game pixels into output pixels, kept inside the picture.
func outRect(r image.Rectangle) image.Rectangle {
	return image.Rect(r.Min.X*Scale, r.Min.Y*Scale, r.Max.X*Scale, r.Max.Y*Scale).Intersect(image.Rect(0, 0, OutW, OutH))
}

// Fill paints a rectangle (in game pixels) one colour.
func Fill(dst *image.RGBA, r image.Rectangle, c color.RGBA) {
	o := outRect(r)
	for y := o.Min.Y; y < o.Max.Y; y++ {
		row := dst.Pix[y*dst.Stride:]
		for x := o.Min.X; x < o.Max.X; x++ {
			row[4*x], row[4*x+1], row[4*x+2], row[4*x+3] = c.R, c.G, c.B, c.A
		}
	}
}

// putBlock paints the 4x4 block of game pixel (x, y).
func putBlock(dst *image.RGBA, x, y int, c color.RGBA) {
	if x < 0 || y < 0 || x >= GameW || y >= GameH {
		return
	}
	for j := 0; j < Scale; j++ {
		row := dst.Pix[(y*Scale+j)*dst.Stride+x*Scale*4:]
		for i := 0; i < Scale; i++ {
			row[4*i], row[4*i+1], row[4*i+2], row[4*i+3] = c.R, c.G, c.B, c.A
		}
	}
}

// blockColor is the colour of the 4x4 block of game pixel (x, y): its top-left pixel.
func blockColor(src *image.RGBA, x, y int) color.RGBA {
	i := (y*Scale)*src.Stride + x*Scale*4
	return color.RGBA{src.Pix[i], src.Pix[i+1], src.Pix[i+2], src.Pix[i+3]}
}

// bayer4 is the 4x4 ordered-dither matrix, values 0 to 15. A dither of
// level n (n sixteenths, 4 / 8 / 16 for 25 / 50 / 100 %) lets a pixel through
// when its matrix value is below n.
var bayer4 = [4][4]int{
	{0, 8, 2, 10},
	{12, 4, 14, 6},
	{3, 11, 1, 9},
	{15, 7, 13, 5},
}

// BayerShows says whether game pixel (x, y) shows at a dither level of n sixteenths.
func BayerShows(x, y, n int) bool {
	return bayer4[y&3][x&3] < n
}

// SpriteOp draws a 480x270 sprite layer.
type SpriteOp struct {
	DX, DY  int             // where it sits, in game pixels (whole numbers)
	Src     image.Rectangle // only this part of the sprite (empty = all of it)
	Exclude image.Rectangle // leave this part out (empty = none)
	Clip    image.Rectangle // draw only inside this rectangle of the picture (empty = everywhere)
	Bayer   int             // 0 = whole; else a dither level in sixteenths
}

// DrawSprite puts the sprite on dst. Transparent pixels are skipped; the
// layers are palette-only, so a pixel is either there or not (a pixel with
// some alpha is blended).
func DrawSprite(dst, sprite *image.RGBA, op SpriteOp) {
	src := op.Src
	if src.Empty() {
		src = image.Rect(0, 0, GameW, GameH)
	}
	for y := src.Min.Y; y < src.Max.Y; y++ {
		for x := src.Min.X; x < src.Max.X; x++ {
			if image.Pt(x, y).In(op.Exclude) {
				continue
			}
			dx, dy := x+op.DX, y+op.DY
			if !op.Clip.Empty() && !image.Pt(dx, dy).In(op.Clip) {
				continue
			}
			if op.Bayer > 0 && !BayerShows(dx, dy, op.Bayer) {
				continue
			}
			i := y*sprite.Stride + x*4
			a := sprite.Pix[i+3]
			if a == 0 {
				continue
			}
			c := color.RGBA{sprite.Pix[i], sprite.Pix[i+1], sprite.Pix[i+2], 255}
			if a < 255 {
				under := blockColor(dst, clampInt(dx, 0, GameW-1), clampInt(dy, 0, GameH-1))
				c = blend(c, under, a)
			}
			putBlock(dst, dx, dy, c)
		}
	}
}

func blend(top, under color.RGBA, a uint8) color.RGBA {
	mix := func(t, u uint8) uint8 { return uint8((int(t)*int(a) + int(u)*(255-int(a)) + 127) / 255) }
	return color.RGBA{mix(top.R, under.R), mix(top.G, under.G), mix(top.B, under.B), 255}
}

func clampInt(v, lo, hi int) int { return max(lo, min(hi, v)) }

// GameOp draws a recorded game frame (1920x1080) onto dst.
//
// The view: the frame is shown whole-number magnified by Zoom (1 = as it is,
// 2 or 3 = a punch-in) around the game pixel (CX, CY): the window of
// 480/Zoom x 270/Zoom game pixels that has (CX, CY) in its middle, kept
// inside the frame. Then it is moved by (DX, DY) game pixels and clipped.
type GameOp struct {
	Zoom   int
	CX, CY int
	DX, DY int
	Clip   image.Rectangle // draw only inside this rectangle of the picture (empty = everywhere)
}

// ViewOrigin is the top-left game pixel of the window a punch-in shows.
func ViewOrigin(zoom, cx, cy int) (vx, vy int) {
	if zoom <= 1 {
		return 0, 0
	}
	w, h := GameW/zoom, GameH/zoom
	return clampInt(cx-w/2, 0, GameW-w), clampInt(cy-h/2, 0, GameH-h)
}

// DrawGame copies src, a recorded frame, onto dst as op says.
func DrawGame(dst, src *image.RGBA, op GameOp) {
	zoom := max(op.Zoom, 1)
	vx, vy := ViewOrigin(zoom, op.CX, op.CY)
	area := image.Rect(0, 0, OutW, OutH)
	if !op.Clip.Empty() {
		area = area.Intersect(outRect(op.Clip))
	}
	for y := area.Min.Y; y < area.Max.Y; y++ {
		sy := vy*Scale + floorDiv(y-op.DY*Scale, zoom)
		if sy < 0 || sy >= OutH {
			continue
		}
		drow := dst.Pix[y*dst.Stride:]
		srow := src.Pix[sy*src.Stride:]
		if zoom == 1 {
			x0, x1 := max(area.Min.X, op.DX*Scale), min(area.Max.X, OutW+op.DX*Scale)
			if x0 < x1 {
				copy(drow[4*x0:4*x1], srow[4*(x0-op.DX*Scale):4*(x1-op.DX*Scale)])
			}
			continue
		}
		for x := area.Min.X; x < area.Max.X; x++ {
			sx := vx*Scale + floorDiv(x-op.DX*Scale, zoom)
			if sx < 0 || sx >= OutW {
				continue
			}
			copy(drow[4*x:4*x+4], srow[4*sx:4*sx+4])
		}
	}
}

// floorDiv divides rounding towards minus infinity (Go's / rounds to zero).
func floorDiv(a, b int) int {
	q := a / b
	if a%b != 0 && (a < 0) != (b < 0) {
		q--
	}
	return q
}

// Checker darkens the picture the way a dither does: every other game pixel
// (a chessboard) turns to colour c.
func Checker(dst *image.RGBA, c color.RGBA) {
	for y := 0; y < GameH; y++ {
		for x := 0; x < GameW; x++ {
			if (x+y)&1 == 0 {
				putBlock(dst, x, y, c)
			}
		}
	}
}
