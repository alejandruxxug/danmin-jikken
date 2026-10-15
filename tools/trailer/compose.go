package main

// The compositor: draws the layers of a frame (layers.go, timeline.go) with
// the recorded game frames and the sprites, on a 1920x1080 picture.

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"sync"

	"danmingame"
)

// Frames gives the recorded game frames: frame number n (from 1) of a shot.
type Frames interface {
	Frame(shot string, n int) (*image.RGBA, error)
}

// DiskFrames reads the frames from the shot folders, <root>/<shot>/f000001.png,
// and keeps the most recent ones, because a still shot asks for the same frame again and again.
type DiskFrames struct {
	Root string

	mu    sync.Mutex
	cache map[string]*image.RGBA
}

const frameCacheSize = 48 // about 400 MB of 1920x1080 frames

func (d *DiskFrames) Frame(shot string, n int) (*image.RGBA, error) {
	key := fmt.Sprintf("%s/%d", shot, n)
	d.mu.Lock()
	im, ok := d.cache[key]
	d.mu.Unlock()
	if ok {
		return im, nil
	}
	path := filepath.Join(d.Root, shot, game.FrameFile(n))
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("shot %q has no frame %d (%w): is shots.txt's frames/from right?", shot, n, err)
	}
	defer f.Close()
	decoded, err := png.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	im, ok = decoded.(*image.RGBA)
	if !ok {
		im = image.NewRGBA(decoded.Bounds())
		for y := 0; y < im.Bounds().Dy(); y++ {
			for x := 0; x < im.Bounds().Dx(); x++ {
				im.Set(x, y, decoded.At(x, y))
			}
		}
	}
	if im.Bounds() != image.Rect(0, 0, OutW, OutH) {
		return nil, fmt.Errorf("%s is %v, want %dx%d", path, im.Bounds(), OutW, OutH)
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.cache == nil {
		d.cache = map[string]*image.RGBA{}
	}
	if len(d.cache) >= frameCacheSize {
		for k := range d.cache { // drop an arbitrary one: good enough for a cache this size
			delete(d.cache, k)
			break
		}
	}
	d.cache[key] = im
	return im, nil
}

// Composer draws frames.
type Composer struct {
	Sheets map[string]*Sheet // SheetFX, SheetCards
	Glyphs *Glyphs
	Frames Frames
}

// Compose clears dst to black and draws the layers on it.
func (c *Composer) Compose(dst *image.RGBA, layers []Layer) error {
	Fill(dst, fullScreen, colBlack)
	for _, l := range layers {
		switch l.Kind {
		case LayerGame:
			src, err := c.Frames.Frame(l.Shot, l.Frame)
			if err != nil {
				return err
			}
			DrawGame(dst, src, l.Game)
		case LayerSprite:
			sheet, ok := c.Sheets[l.Sheet]
			if !ok {
				return fmt.Errorf("no sheet %q", l.Sheet)
			}
			sprite, err := sheet.Layer(l.Name, l.Index)
			if err != nil {
				return err
			}
			DrawSprite(dst, sprite, l.Sprite)
		case LayerFill:
			Fill(dst, l.Rect, l.Color)
		case LayerChecker:
			Checker(dst, l.Color)
		case LayerText:
			b := c.Glyphs.Render(l.Text)
			x := l.X
			if l.Centered {
				x -= b.W * l.Scale / 2
			}
			DrawBitmap(dst, b, x, l.Y, l.Scale, l.Color)
		}
	}
	return nil
}
