// Package sprites reads the spritesheets that assets/export.sh writes: one
// PNG with every layer of every frame packed in it, and an Aseprite
// "json-array" file that says where each "<layer>#<frame>" sits.
//
// The game never draws art in code. It cuts frames out of these sheets with
// SubImage and stacks them. See Art Direction -> Pipeline.
package sprites

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io/fs"
	"strconv"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// BoilTag is the animation tag of the line boil: 3 frames, 125 ms each.
const BoilTag = "boil"

// Tag is a named run of frames, for example "boil" = frames 0 to 2.
type Tag struct {
	Name string
	From int
	To   int
}

// Meta is what we keep from the json file. Parse fills it.
type Meta struct {
	frames    map[string]image.Rectangle // "<layer>#<frame>" -> rectangle in the sheet
	durations []time.Duration            // per frame number, from the "duration" field
	Layers    []string
	Tags      map[string]Tag
}

// The structs below only describe the json file, so encoding/json can fill them.
type jsonFile struct {
	Frames []jsonFrame `json:"frames"`
	Meta   jsonMeta    `json:"meta"`
}

type jsonFrame struct {
	Filename string `json:"filename"`
	Frame    struct {
		X, Y, W, H int
	} `json:"frame"`
	Duration int `json:"duration"` // milliseconds
}

type jsonMeta struct {
	Layers []struct {
		Name string `json:"name"`
	} `json:"layers"`
	FrameTags []struct {
		Name string `json:"name"`
		From int    `json:"from"`
		To   int    `json:"to"`
	} `json:"frameTags"`
}

// Parse reads the json half of a sheet. It does not need a window or the png,
// so tests can call it anywhere.
func Parse(data []byte) (*Meta, error) {
	var f jsonFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("sprites: reading json: %w", err)
	}
	m := &Meta{
		frames: make(map[string]image.Rectangle),
		Tags:   make(map[string]Tag),
	}
	for _, jf := range f.Frames {
		m.frames[jf.Filename] = image.Rect(jf.Frame.X, jf.Frame.Y, jf.Frame.X+jf.Frame.W, jf.Frame.Y+jf.Frame.H)
		// Names look like "lineart#2": the layer, then '#', then the frame number.
		n, err := frameNumber(jf.Filename)
		if err != nil {
			return nil, err
		}
		for len(m.durations) <= n {
			m.durations = append(m.durations, 0)
		}
		m.durations[n] = time.Duration(jf.Duration) * time.Millisecond
	}
	for _, l := range f.Meta.Layers {
		m.Layers = append(m.Layers, l.Name)
	}
	for _, t := range f.Meta.FrameTags {
		m.Tags[t.Name] = Tag{Name: t.Name, From: t.From, To: t.To}
	}
	return m, nil
}

// frameNumber reads the 2 out of "lineart#2".
func frameNumber(name string) (int, error) {
	i := strings.LastIndex(name, "#")
	if i < 0 {
		return 0, fmt.Errorf("sprites: frame name %q is not <layer>#<frame>", name)
	}
	n, err := strconv.Atoi(name[i+1:])
	if err != nil {
		return 0, fmt.Errorf("sprites: frame name %q is not <layer>#<frame>: %w", name, err)
	}
	return n, nil
}

// Rect returns where a layer's frame sits in the sheet.
func (m *Meta) Rect(layer string, frame int) (image.Rectangle, error) {
	r, ok := m.frames[fmt.Sprintf("%s#%d", layer, frame)]
	if !ok {
		return image.Rectangle{}, fmt.Errorf("sprites: no frame %s#%d", layer, frame)
	}
	return r, nil
}

// FrameAt says which frame of a tag to show after elapsed time has passed.
// It cycles forever, so FrameAt(BoilTag, t) gives the line boil. A tag that
// does not exist gives frame 0.
func (m *Meta) FrameAt(tag string, elapsed time.Duration) int {
	t, ok := m.Tags[tag]
	if !ok {
		return 0
	}
	var total time.Duration
	for n := t.From; n <= t.To; n++ {
		total += m.durationOf(n)
	}
	if total <= 0 {
		return t.From
	}
	left := elapsed % total
	for n := t.From; n <= t.To; n++ {
		if left < m.durationOf(n) {
			return n
		}
		left -= m.durationOf(n)
	}
	return t.To
}

func (m *Meta) durationOf(n int) time.Duration {
	if n < 0 || n >= len(m.durations) {
		return 0
	}
	return m.durations[n]
}

// Sheet is a loaded spritesheet: the picture plus its Meta.
type Sheet struct {
	*Meta
	img *ebiten.Image
}

// Load reads <name>.json and <name>.png from fsys (the game passes its
// embedded assets/sprites folder, tests pass a testdata folder).
func Load(fsys fs.FS, name string) (*Sheet, error) {
	data, err := fs.ReadFile(fsys, name+".json")
	if err != nil {
		return nil, fmt.Errorf("sprites: %w", err)
	}
	meta, err := Parse(data)
	if err != nil {
		return nil, fmt.Errorf("sprites: sheet %s: %w", name, err)
	}
	file, err := fsys.Open(name + ".png")
	if err != nil {
		return nil, fmt.Errorf("sprites: %w", err)
	}
	defer file.Close()
	pic, err := png.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("sprites: sheet %s: %w", name, err)
	}
	return &Sheet{Meta: meta, img: ebiten.NewImageFromImage(pic)}, nil
}

// Frame cuts one layer's frame out of the sheet. The result shares pixels
// with the sheet, so it costs nothing to call every frame.
func (s *Sheet) Frame(layer string, frame int) (*ebiten.Image, error) {
	r, err := s.Rect(layer, frame)
	if err != nil {
		return nil, err
	}
	return s.img.SubImage(r).(*ebiten.Image), nil
}
