package sprites

import (
	"image/color"
	"os"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	data, err := os.ReadFile("testdata/tiny.json")
	if err != nil {
		t.Fatal(err)
	}
	m, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	r, err := m.Rect("lineart", 1)
	if err != nil {
		t.Fatal(err)
	}
	if r.Min.X != 4 || r.Dx() != 4 || r.Dy() != 4 {
		t.Errorf("lineart#1 rect = %v, want x=4 and 4x4", r)
	}
	if len(m.Layers) != 2 || m.Layers[1] != "eyes_sleepy" {
		t.Errorf("layers = %v", m.Layers)
	}
	if _, err := m.Rect("lineart", 9); err == nil {
		t.Error("asking for a missing frame should fail")
	}
}

func TestParseRejectsBadFrameName(t *testing.T) {
	bad := `{"frames":[{"filename":"lineart","frame":{"x":0,"y":0,"w":1,"h":1},"duration":125}],"meta":{}}`
	if _, err := Parse([]byte(bad)); err == nil {
		t.Error("a frame name without #<frame> should fail")
	}
	if _, err := Parse([]byte("not json")); err == nil {
		t.Error("broken json should fail")
	}
}

func TestFrameAtCyclesTheBoil(t *testing.T) {
	data, _ := os.ReadFile("testdata/tiny.json")
	m, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		elapsed time.Duration
		want    int
	}{
		{0, 0},
		{124 * time.Millisecond, 0},
		{125 * time.Millisecond, 1},
		{300 * time.Millisecond, 2},
		{375 * time.Millisecond, 0}, // 3 x 125 ms, back to the start
		{500 * time.Millisecond, 1},
	}
	for _, tc := range tests {
		if got := m.FrameAt(BoilTag, tc.elapsed); got != tc.want {
			t.Errorf("FrameAt(%v) = %d, want %d", tc.elapsed, got, tc.want)
		}
	}
	if got := m.FrameAt("nope", time.Second); got != 0 {
		t.Errorf("unknown tag gave frame %d, want 0", got)
	}
}

// BoilScale 3 plays the same tag three times slower: frame 1 starts at 375 ms.
func TestBoilScaleSlowsTheBoil(t *testing.T) {
	data, _ := os.ReadFile("testdata/tiny.json")
	m, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	m.BoilScale = 3
	tests := []struct {
		elapsed time.Duration
		want    int
	}{
		{0, 0},
		{374 * time.Millisecond, 0},
		{375 * time.Millisecond, 1},
		{750 * time.Millisecond, 2},
		{1125 * time.Millisecond, 0}, // a full cycle is 3 x 375 ms
	}
	for _, tc := range tests {
		if got := m.FrameAt(BoilTag, tc.elapsed); got != tc.want {
			t.Errorf("scale 3: FrameAt(%v) = %d, want %d", tc.elapsed, got, tc.want)
		}
	}
}

func TestLoadCutsFrames(t *testing.T) {
	sheet, err := Load(os.DirFS("testdata"), "tiny")
	if err != nil {
		t.Fatal(err)
	}
	img, err := sheet.Frame("lineart", 2)
	if err != nil {
		t.Fatal(err)
	}
	if b := img.Bounds(); b.Dx() != 4 || b.Dy() != 4 {
		t.Errorf("frame size = %v, want 4x4", b)
	}
	if _, err := Load(os.DirFS("testdata"), "missing"); err == nil {
		t.Error("loading a missing sheet should fail")
	}
}

func TestParseLayerOpacity(t *testing.T) {
	data := `{"frames":[],"meta":{"layers":[{"name":"bezel","opacity":255},{"name":"glare","opacity":28},{"name":"plain"}]}}`
	m, err := Parse([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Opacity("bezel"); got != 1 {
		t.Errorf("bezel opacity = %v, want 1", got)
	}
	if got := m.Opacity("glare"); got < 0.10 || got > 0.12 {
		t.Errorf("glare opacity = %v, want about 0.11", got)
	}
	if got := m.Opacity("plain"); got != 1 {
		t.Errorf("a layer with no opacity field should be opaque, got %v", got)
	}
	if got := m.Opacity("missing"); got != 1 {
		t.Errorf("an unknown layer should be opaque, got %v", got)
	}
}

func TestDarkerStaysOnThePalette(t *testing.T) {
	ramp := []color.RGBA{
		{0xec, 0xec, 0xe6, 0xff}, {0xdd, 0xe5, 0xd9, 0xff}, {0xc3, 0xd1, 0xc0, 0xff}, {0x9f, 0xb1, 0x9f, 0xff},
	}
	black := color.RGBA{0x0b, 0x0b, 0x0c, 0xff}
	for i, c := range ramp {
		for steps := 0; steps <= 4; steps++ {
			want := ramp[min(i+steps, 3)]
			if got := Darker(c, steps); got != want {
				t.Errorf("Darker(ramp[%d], %d) = %v, want %v", i, steps, got, want)
			}
		}
	}
	if got := Darker(black, 3); got != black {
		t.Errorf("outline colour changed to %v", got)
	}
}
