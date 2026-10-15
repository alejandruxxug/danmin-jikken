package main

// Reading the spritesheets the compositor draws from (game/assets/sprites):
// the png and the Aseprite json-array that goes with it (frames named
// "<layer>#<frame>"). Nothing is drawn in code: the cards and the motion
// graphics come from trailer_cards and trailer_fx.

import (
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"sync"
)

// Sheet is one exported spritesheet.
type Sheet struct {
	img    *image.RGBA
	frames map[string]image.Rectangle // "layer#frame" -> where it is in the png
	mu     sync.Mutex                 // the compositor draws from several goroutines
	cache  map[string]*image.RGBA     // the layers already cut out, 480x270 each
}

type sheetJSON struct {
	Frames []struct {
		Filename string `json:"filename"`
		Frame    struct{ X, Y, W, H int }
	} `json:"frames"`
}

// LoadSheet reads <dir>/<name>.png and <dir>/<name>.json.
func LoadSheet(dir, name string) (*Sheet, error) {
	f, err := os.Open(filepath.Join(dir, name+".png"))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	decoded, err := png.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("%s.png: %w", name, err)
	}
	img := image.NewRGBA(decoded.Bounds())
	draw.Draw(img, img.Bounds(), decoded, image.Point{}, draw.Src)
	data, err := os.ReadFile(filepath.Join(dir, name+".json"))
	if err != nil {
		return nil, err
	}
	var meta sheetJSON
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("%s.json: %w", name, err)
	}
	s := &Sheet{img: img, frames: map[string]image.Rectangle{}, cache: map[string]*image.RGBA{}}
	for _, fr := range meta.Frames {
		s.frames[fr.Filename] = image.Rect(fr.Frame.X, fr.Frame.Y, fr.Frame.X+fr.Frame.W, fr.Frame.Y+fr.Frame.H)
	}
	return s, nil
}

// Layer is a frame of a layer as a 480x270 picture with transparent pixels.
// The same layer frame is cut out once.
func (s *Sheet) Layer(layer string, frame int) (*image.RGBA, error) {
	key := fmt.Sprintf("%s#%d", layer, frame)
	s.mu.Lock()
	defer s.mu.Unlock()
	if im, ok := s.cache[key]; ok {
		return im, nil
	}
	r, ok := s.frames[key]
	if !ok {
		return nil, fmt.Errorf("the sheet has no %q", key)
	}
	im := image.NewRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	draw.Draw(im, im.Bounds(), s.img, r.Min, draw.Src)
	s.cache[key] = im
	return im, nil
}
