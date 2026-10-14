package game

import (
	"math"
	"testing"
)

func near(a, b float32) bool { return math.Abs(float64(a-b)) < 1e-3 }

// The fold is the perpendicular bisector: every point of it is as far from
// the rest corner as from the cursor, and mirroring the rest corner across it
// lands exactly on the cursor.
func TestFoldIsThePerpendicularBisector(t *testing.T) {
	rest, cursor := point{140, 180}, point{60, 150}
	f, ok := newFold(rest, cursor)
	if !ok {
		t.Fatal("no fold")
	}
	if !near(f.side(f.m), 0) {
		t.Error("the middle of the segment must be on the fold")
	}
	along := point{-f.n.y, f.n.x} // along the fold
	for _, k := range []float32{-50, 0, 37} {
		p := point{f.m.x + k*along.x, f.m.y + k*along.y}
		dr := math.Hypot(float64(p.x-rest.x), float64(p.y-rest.y))
		dc := math.Hypot(float64(p.x-cursor.x), float64(p.y-cursor.y))
		if math.Abs(dr-dc) > 1e-3 {
			t.Errorf("point %v is %v from the corner but %v from the cursor", p, dr, dc)
		}
	}
	if got := f.reflect(rest); !near(got.x, cursor.x) || !near(got.y, cursor.y) {
		t.Errorf("reflect(rest) = %v, want the cursor %v", got, cursor)
	}
	if _, ok := newFold(rest, rest); ok {
		t.Error("a corner that has not moved has no fold")
	}
}

func TestReflectIsItsOwnInverse(t *testing.T) {
	f, _ := newFold(point{140, 180}, point{20, 90})
	p := point{33, 71}
	if back := f.reflect(f.reflect(p)); !near(back.x, p.x) || !near(back.y, p.y) {
		t.Errorf("reflecting twice gave %v, want %v", back, p)
	}
}

// A horizontal drag to the page's middle folds the right half over: the
// fold is the vertical line x = 105 for a cursor at x = 70.
func TestClipByAVerticalFold(t *testing.T) {
	flat, peeled, flap, f, ok := curl(140, 180, point{70, 180})
	if !ok {
		t.Fatal("no curl")
	}
	if !near(f.m.x, 105) || !near(area(flat), 105*180) || !near(area(peeled), 35*180) {
		t.Errorf("fold at %v, flat %v, peeled %v", f.m, area(flat), area(peeled))
	}
	// The peeled part and the flap are mirror images: same area, and the
	// flap lies on the flat side, between x = 70 and 105.
	if !near(area(flap), area(peeled)) {
		t.Errorf("flap %v vs peeled %v", area(flap), area(peeled))
	}
	for _, p := range flap {
		if p.x < 70-1e-3 || p.x > 105+1e-3 {
			t.Errorf("flap corner %v is outside x 70..105", p)
		}
	}
}

// A slanted fold still splits the page without losing any of it.
func TestClipKeepsTheWholePage(t *testing.T) {
	for _, cursor := range []point{{100, 170}, {40, 120}, {10, 60}, {0, 180}} {
		flat, peeled, _, _, ok := curl(140, 180, cursor)
		if !ok {
			t.Fatalf("no curl for %v", cursor)
		}
		if total := area(flat) + area(peeled); !near(total/10, 140*180/10) {
			t.Errorf("cursor %v: flat+peeled = %v, want %v", cursor, total, 140*180)
		}
	}
}

func TestNoFoldLeavesThePageFlat(t *testing.T) {
	flat, peeled, flap, _, ok := curl(140, 180, point{140, 180})
	if ok || len(peeled) != 0 || len(flap) != 0 || !near(area(flat), 140*180) {
		t.Errorf("ok %v, peeled %d, flap %d, flat %v", ok, len(peeled), len(flap), area(flat))
	}
}
