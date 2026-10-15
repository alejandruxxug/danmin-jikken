package main

// Finding things in a sprite by looking at its pixels: the rows a line of text
// sits in, where each letter ends, the box around the moon. The compositor
// uses them to type a line out letter by letter and to move one part of a card
// without a second layer for it.

import "image"

// isInk says whether a pixel is part of the picture: not transparent and not
// the background colour bg.
func isInk(im *image.RGBA, x, y int, bg [3]uint8) bool {
	i := y*im.Stride + x*4
	if im.Pix[i+3] == 0 {
		return false
	}
	return [3]uint8{im.Pix[i], im.Pix[i+1], im.Pix[i+2]} != bg
}

// backgroundOf is the colour of the top-left pixel, taken as the card's background.
func backgroundOf(im *image.RGBA) [3]uint8 {
	return [3]uint8{im.Pix[0], im.Pix[1], im.Pix[2]}
}

// InkBands are the runs of rows that have ink, as full-width rectangles, top
// to bottom. Rows with no ink between two runs separate them.
func InkBands(im *image.RGBA) []image.Rectangle {
	bg := backgroundOf(im)
	w, h := im.Bounds().Dx(), im.Bounds().Dy()
	var bands []image.Rectangle
	start := -1
	for y := 0; y <= h; y++ {
		has := false
		if y < h {
			for x := 0; x < w && !has; x++ {
				has = isInk(im, x, y, bg)
			}
		}
		switch {
		case has && start < 0:
			start = y
		case !has && start >= 0:
			bands = append(bands, image.Rect(0, start, w, y))
			start = -1
		}
	}
	return bands
}

// LetterEdges gives, for each run of ink columns inside a band (a letter, or
// two letters that touch), the x just past its right edge. Typing the line
// out is then: show every column left of Edges[i] for letter i.
func LetterEdges(im *image.RGBA, band image.Rectangle) []int {
	bg := backgroundOf(im)
	var edges []int
	inRun := false
	for x := 0; x <= band.Max.X; x++ {
		has := false
		if x < band.Max.X {
			for y := band.Min.Y; y < band.Max.Y && !has; y++ {
				has = isInk(im, x, y, bg)
			}
		}
		switch {
		case has:
			inRun = true
		case inRun:
			edges = append(edges, x)
			inRun = false
		}
	}
	return edges
}

// InkBounds is the smallest rectangle that holds all the ink inside area
// (empty when there is none).
func InkBounds(im *image.RGBA, area image.Rectangle, bg [3]uint8) image.Rectangle {
	var box image.Rectangle
	for y := area.Min.Y; y < area.Max.Y; y++ {
		for x := area.Min.X; x < area.Max.X; x++ {
			if !isInk(im, x, y, bg) {
				continue
			}
			p := image.Rect(x, y, x+1, y+1)
			if box.Empty() {
				box = p
			} else {
				box = box.Union(p)
			}
		}
	}
	return box
}
