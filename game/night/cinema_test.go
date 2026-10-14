package night

import (
	"math"
	"strings"
	"testing"
)

func TestBarHeight(t *testing.T) {
	if got := BarHeight(0, 32, 0, 36); got != 0 {
		t.Errorf("at the start: %v, want 0", got)
	}
	if got := BarHeight(0, 32, 36, 36); got != 32 {
		t.Errorf("at the end: %v, want 32", got)
	}
	half := BarHeight(0, 32, 18, 36)
	if half <= 16 || half >= 32 {
		t.Errorf("halfway %v: an ease-out is past half-way already", half)
	}
	if got := BarHeight(10, 40, 5, 0); got != 40 {
		t.Errorf("a move of 0 ticks is instant, got %v", got)
	}
	if got := BarHeight(32, 0, 12, 24); got <= 0 || got >= 16 {
		t.Errorf("sliding out, halfway: %v", got)
	}
}

func TestCropOrigin(t *testing.T) {
	x, y := CropOrigin(3, 350, 89) // the realization's punch-in
	if x != 270 || y != 44 {
		t.Errorf("crop at (%v,%v), want (270,44): a 160x90 window centred on the signature", x, y)
	}
	x, y = CropOrigin(2, 10, 5)
	if x != 0 || y != 0 {
		t.Errorf("near a corner the crop stops at the edge, got (%v,%v)", x, y)
	}
	x, y = CropOrigin(2, 479, 269)
	if x != 240 || y != 135 {
		t.Errorf("far corner: (%v,%v), want (240,135)", x, y)
	}
	if x, y = CropOrigin(1, 100, 100); x != 0 || y != 0 {
		t.Errorf("scale 1 shows everything, got (%v,%v)", x, y)
	}
}

func TestPanX(t *testing.T) {
	if PanX(0, 140, 0, 240) != 0 || PanX(0, 140, 240, 240) != 140 {
		t.Error("the pan must start and end on its numbers")
	}
	if mid := PanX(0, 140, 120, 240); math.Abs(mid-70) > 0.001 {
		t.Errorf("halfway %v, want 70 (smooth both ends)", mid)
	}
}

func TestSubtitleTops(t *testing.T) {
	one, two := SubtitleTops(1), SubtitleTops(2)
	if one[0] != 250 || two[0] != 245 || two[1] != 255 {
		t.Errorf("tops %v %v, want [250] [245 255]", one, two)
	}
}

func TestBarsAndCameraCommands(t *testing.T) {
	src := "bars on\n~ One.\nbars 62 1.5\n~ Two.\npunch 3 350 89\n~ Three.\npunch off\npush 1.1 250 130 6\n~ Four.\npan 140 4\n~ Five.\nbars off\n~ Six.\n"
	e := newEngine(t, src, Flags{})
	if !e.Cinematic() {
		t.Fatal("bars on must start the cinematic mode")
	}
	for range 40 {
		_ = e.Tick()
	}
	if got := e.BarsHeight(); got != BarFull {
		t.Fatalf("after 40 ticks the bars are %v tall, want %d", got, BarFull)
	}
	_ = e.Advance()
	_ = e.Advance() // bars 62 1.5 -> Two
	for range 90 {
		_ = e.Tick()
	}
	if got := e.BarsHeight(); got != 62 {
		t.Errorf("grown bars = %v, want 62", got)
	}
	_ = e.Advance()
	_ = e.Advance() // punch -> Three
	if s, cx, cy := e.Zoom(); s != 3 || cx != 350 || cy != 89 {
		t.Errorf("punch: scale %v at (%v,%v)", s, cx, cy)
	}
	_ = e.Advance()
	_ = e.Advance() // punch off, push -> Four
	if s, _, _ := e.Zoom(); s != 1 {
		t.Errorf("a push starts from 1, got %v", s)
	}
	for range 360 {
		_ = e.Tick()
	}
	if s, _, _ := e.Zoom(); math.Abs(s-1.1) > 0.001 {
		t.Errorf("after 6 s the push is at %v, want 1.1", s)
	}
	_ = e.Advance()
	_ = e.Advance() // pan -> Five
	for range 240 {
		_ = e.Tick()
	}
	if e.PanX() != 140 {
		t.Errorf("pan = %v, want 140", e.PanX())
	}
	_ = e.Advance()
	_ = e.Advance() // bars off -> Six
	if e.Cinematic() {
		t.Error("bars off ends the cinematic mode (the subtitles give way to the text box)")
	}
}

func TestCinemaCommandErrors(t *testing.T) {
	for _, src := range []string{"bars\n", "bars sideways\n", "punch 5 1 1\n", "punch 3 1\n", "push 3 1 1 2\n", "pan\n", "pan x\n"} {
		if _, err := Parse("x.txt", strings.NewReader(src)); err == nil {
			t.Errorf("%q must be an error", src)
		}
	}
}

func TestFastForwardSettlesTheCamera(t *testing.T) {
	e := newEngine(t, "bars on\npunch 3 350 89\n~ One.\n~ Two.\n", Flags{})
	if err := e.FastForward(4); err != nil {
		t.Fatal(err)
	}
	if got := e.BarsHeight(); got != BarFull {
		t.Errorf("a backdoor jump lands on settled bars, got %v", got)
	}
}

func TestWashLevels(t *testing.T) {
	for _, c := range []struct {
		p    float64
		want int
	}{{0, 0}, {0.24, 0}, {0.3, 1}, {0.6, 2}, {0.9, 3}, {1, 3}} {
		if got := WashLevel(c.p); got != c.want {
			t.Errorf("WashLevel(%v) = %d, want %d", c.p, got, c.want)
		}
	}
	pix := WashPixels(2, 4, 2, [4]byte{1, 2, 3, 255})
	covered := 0
	for i := 3; i < len(pix); i += 4 {
		if pix[i] == 255 {
			covered++
		}
	}
	if covered != 4 { // half of 8 pixels: a checker
		t.Errorf("level 2 covers %d of 8 pixels, want 4", covered)
	}
}

func TestPaperCommandWritesOnARow(t *testing.T) {
	e := newEngine(t, "bg letter\npaper 4 Please.\n~ Ok.\n", Flags{})
	p := e.Paper()
	if len(p) != 5 || p[4] != "Please." || p[0] != "" {
		t.Fatalf("paper = %q: four blank rows, then the words", p)
	}
	rows := PaperRows(p, LetterPaper, fakeMetrics{}, HanaStyle)
	if len(rows) != 1 || rows[0].Base != 93 {
		t.Errorf("rows %+v: the words sit on baseline 93 (row 4 of the letter)", rows)
	}
}
