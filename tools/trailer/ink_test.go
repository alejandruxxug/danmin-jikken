package main

import (
	"image"
	"image/color"
	"reflect"
	"testing"
)

// strip is a 30x20 picture: black background, two lines of "letters" (white
// blocks): one at rows 3-5 with letters at columns 2-4, 7-8 and 12-12, one at rows 10-11.
func strip() *image.RGBA {
	im := image.NewRGBA(image.Rect(0, 0, 30, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 30; x++ {
			im.SetRGBA(x, y, color.RGBA{11, 11, 12, 255})
		}
	}
	white := color.RGBA{236, 236, 230, 255}
	for y := 3; y <= 5; y++ {
		for _, x := range []int{2, 3, 4, 7, 8, 12} {
			im.SetRGBA(x, y, white)
		}
	}
	for x := 5; x <= 20; x++ {
		im.SetRGBA(x, 10, white)
		im.SetRGBA(x, 11, white)
	}
	return im
}

func TestInkBands(t *testing.T) {
	got := InkBands(strip())
	want := []image.Rectangle{image.Rect(0, 3, 30, 6), image.Rect(0, 10, 30, 12)}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("InkBands = %v, want %v", got, want)
	}
}

func TestLetterEdges(t *testing.T) {
	im := strip()
	got := LetterEdges(im, image.Rect(0, 3, 30, 6))
	if want := []int{5, 9, 13}; !reflect.DeepEqual(got, want) {
		t.Errorf("LetterEdges = %v, want %v", got, want)
	}
	if got := LetterEdges(im, image.Rect(0, 10, 30, 12)); !reflect.DeepEqual(got, []int{21}) {
		t.Errorf("one run touching: got %v", got)
	}
}

func TestInkBounds(t *testing.T) {
	im := strip()
	bg := backgroundOf(im)
	if got, want := InkBounds(im, image.Rect(0, 0, 30, 8), bg), image.Rect(2, 3, 13, 6); got != want {
		t.Errorf("InkBounds = %v, want %v", got, want)
	}
	if got := InkBounds(im, image.Rect(14, 0, 30, 8), bg); !got.Empty() {
		t.Errorf("no ink there, got %v", got)
	}
}
