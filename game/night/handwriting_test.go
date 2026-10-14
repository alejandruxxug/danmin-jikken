package night

import (
	"os"
	"strings"
	"testing"

	"danmingame/sprites"
)

// fakeMetrics: every letter 5 wide, every letter can join.
type fakeMetrics struct{ noJoin string }

func (f fakeMetrics) Advance(g string) (int, bool, bool) {
	if g == "?" {
		return 0, false, false
	}
	return 5, !strings.Contains(f.noJoin, g), g != "join"
}

func TestHanaJoinsTouch(t *testing.T) {
	placed, width := Layout("Hana", fakeMetrics{}, HanaStyle)
	if len(placed) != 4 || width != 20 {
		t.Fatalf("placed %d glyphs, width %d; want 4 and 20 (no gaps: every pair joins)", len(placed), width)
	}
	for i, p := range placed {
		if p.X != 5*i || p.Y != 0 {
			t.Errorf("glyph %d at (%d,%d), want (%d,0)", i, p.X, p.Y, 5*i)
		}
	}
}

func TestHanaGapWhenAGlyphCannotJoin(t *testing.T) {
	_, width := Layout("Hana", fakeMetrics{noJoin: "n"}, HanaStyle)
	// H-a joins, a-n does not join (n cannot), n-a does not: two 1 px gaps.
	if width != 22 {
		t.Errorf("width = %d, want 22", width)
	}
}

func TestWordSpace(t *testing.T) {
	placed, _ := Layout("ab cd", fakeMetrics{}, HanaStyle)
	if got := placed[2].X; got != 10+3 {
		t.Errorf("second word starts at %d, want 13 (two letters + a 3 px word space)", got)
	}
}

func TestMayuIsSeededAndBounded(t *testing.T) {
	text := "Did you hear about Cecily"
	a, wa := Layout(text, fakeMetrics{}, MayuStyle)
	b, wb := Layout(text, fakeMetrics{}, MayuStyle)
	if wa != wb || len(a) != len(b) {
		t.Fatal("the same text must lay out the same way every time")
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("glyph %d differs between runs: %+v vs %+v", i, a[i], b[i])
		}
	}
	run := 1
	for i, p := range a {
		if p.Y < -1 || p.Y > 1 {
			t.Errorf("glyph %d jitter %d outside -1..1", i, p.Y)
		}
		if p.Glyph == "join" {
			run++
			if run > maxRun {
				t.Errorf("more than %d letters joined in a row at glyph %d", maxRun, i)
			}
		} else if i > 0 && a[i-1].Glyph != "join" {
			run = 1
		}
	}
	// A different text may look different (the seed is the text).
	c, _ := Layout("Did you hear about Cecilia", fakeMetrics{}, MayuStyle)
	_ = c
}

func TestMayuWidthIgnoresTheCoin(t *testing.T) {
	// A join and a gap are both 2 px, so the coin flips never change the width.
	_, w1 := Layout("abcde", fakeMetrics{}, MayuStyle)
	if w1 != 5*5+4*2 {
		t.Errorf("width = %d, want 33", w1)
	}
}

func TestSkipsMissingGlyphs(t *testing.T) {
	placed, _ := Layout("a?b", fakeMetrics{}, HanaStyle)
	if len(placed) != 2 {
		t.Errorf("placed %d, want 2 (the ? has no glyph)", len(placed))
	}
}

func TestParseGlyphData(t *testing.T) {
	adv, join, err := ParseGlyphData("adv=6;join=1")
	if err != nil || adv != 6 || !join {
		t.Errorf("got %d %v %v", adv, join, err)
	}
	if _, _, err := ParseGlyphData("adv=x"); err == nil {
		t.Error("adv=x must be an error")
	}
}

// The real sheets: every name used in the plate must have a glyph.
func TestRealHandSheets(t *testing.T) {
	for _, name := range []string{"hand_hana", "hand_mayu"} {
		data, err := os.ReadFile("../assets/sprites/" + name + ".json")
		if err != nil {
			t.Fatal(err)
		}
		meta, err := sprites.Parse(data)
		if err != nil {
			t.Fatal(err)
		}
		m := SheetMetrics{meta}
		for _, ch := range "MayuHan" {
			if _, _, ok := m.Advance(string(ch)); !ok {
				t.Errorf("%s has no glyph %q", name, ch)
			}
		}
	}
}
