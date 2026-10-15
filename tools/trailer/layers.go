package main

// A frame of the trailer is a list of layers, drawn first to last. The
// timeline (timeline.go) works out the list for a frame as plain data, with
// no pictures involved, so it can be tested by looking at the numbers; the
// compositor (compose.go) then draws the list.

import (
	"image"
	"image/color"
)

// LayerKind says what a layer is.
type LayerKind int

const (
	LayerGame    LayerKind = iota // a frame of a recorded game shot
	LayerSprite                   // a layer of a spritesheet
	LayerFill                     // a flat rectangle
	LayerChecker                  // a chessboard of one colour over the whole picture
	LayerText                     // a string in Silkscreen, drawn at a whole-number scale
)

// The two sheets the trailer draws from.
const (
	SheetFX    = "fx"    // trailer_fx: the motion graphics
	SheetCards = "cards" // trailer_cards: the cards, the splash, the stain
)

// Layer is one thing to draw. Only the fields of its Kind are used.
type Layer struct {
	Kind LayerKind

	// LayerGame: frame number Frame (from 1) of the recorded shot Shot.
	Shot  string
	Frame int
	Game  GameOp

	// LayerSprite: frame Index of the layer Name in the sheet Sheet.
	Sheet  string
	Name   string
	Index  int
	Sprite SpriteOp

	// LayerFill and LayerChecker.
	Rect  image.Rectangle
	Color color.RGBA

	// LayerText: Text with its top-left corner at (X, Y), each pixel of the
	// font a Scale x Scale block. Color is the ink.
	Text        string
	X, Y, Scale int
	Centered    bool // X is the middle of the text, not its left edge
}

// The palette colours the compositor uses (assets/palette/danmin.hex).
var (
	colBlack = color.RGBA{0x0b, 0x0b, 0x0c, 0xff}
	colPaper = color.RGBA{0xec, 0xec, 0xe6, 0xff}
)

func gameLayer(shot string, frame int, op GameOp) Layer {
	return Layer{Kind: LayerGame, Shot: shot, Frame: frame, Game: op}
}

func spriteLayer(sheet, name string, index int, op SpriteOp) Layer {
	return Layer{Kind: LayerSprite, Sheet: sheet, Name: name, Index: index, Sprite: op}
}

func fillLayer(r image.Rectangle, c color.RGBA) Layer {
	return Layer{Kind: LayerFill, Rect: r, Color: c}
}

// fullScreen is the whole 480x270 picture.
var fullScreen = image.Rect(0, 0, GameW, GameH)
