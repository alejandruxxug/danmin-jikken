package game

import "math"

// The page curl, in small pure functions (no drawing), so it can be tested
// and explained: the grabbed corner follows the cursor; the fold is the
// perpendicular bisector of the segment from the corner's rest place to the
// cursor; everything on the corner's side of the fold is mirrored across it
// and drawn as the back of the page.

// fold is a line: the point m on it and the unit vector n across it that
// points from the corner's rest place toward the cursor.
type fold struct{ m, n point }

// newFold makes the fold of a corner that rests at rest and is held at
// cursor. ok is false when the two are the same place (no fold yet).
func newFold(rest, cursor point) (f fold, ok bool) {
	dx, dy := float64(cursor.x-rest.x), float64(cursor.y-rest.y)
	length := math.Hypot(dx, dy)
	if length < 1e-3 {
		return fold{}, false
	}
	return fold{
		m: point{(rest.x + cursor.x) / 2, (rest.y + cursor.y) / 2}, // the middle of the segment
		n: point{float32(dx / length), float32(dy / length)},
	}, true
}

// side is how far p is from the fold, positive on the cursor's side and
// negative on the corner's side (where the page peels away).
func (f fold) side(p point) float32 {
	return (p.x-f.m.x)*f.n.x + (p.y-f.m.y)*f.n.y
}

// reflect is the mirror image of p across the fold.
func (f fold) reflect(p point) point {
	d := 2 * f.side(p)
	return point{p.x - d*f.n.x, p.y - d*f.n.y}
}

// clip keeps the part of a convex polygon on the cursor's side of the fold
// (positive) or on the corner's side, cutting edges where they cross it
// (Sutherland-Hodgman, one line).
func (f fold) clip(poly []point, positive bool) []point {
	sign := float32(1)
	if !positive {
		sign = -1
	}
	var out []point
	for i, a := range poly {
		b := poly[(i+1)%len(poly)]
		da, db := sign*f.side(a), sign*f.side(b)
		if da >= 0 {
			out = append(out, a)
		}
		if (da >= 0) != (db >= 0) { // this edge crosses the fold
			t := da / (da - db)
			out = append(out, lerp(a, b, t))
		}
	}
	return out
}

// rectPoly is a w by h rectangle as a polygon, corners clockwise.
func rectPoly(w, h float32) []point {
	return []point{{0, 0}, {w, 0}, {w, h}, {0, h}}
}

// curl splits a w by h page whose bottom-right corner is held at cursor:
// the part still lying flat, the part peeled away (where the page below
// shows) and the flap, which is the peeled part mirrored across the fold.
// ok is false when there is no fold yet.
func curl(w, h float32, cursor point) (flat, peeled, flap []point, f fold, ok bool) {
	f, ok = newFold(point{w, h}, cursor)
	if !ok {
		return rectPoly(w, h), nil, nil, f, false
	}
	page := rectPoly(w, h)
	flat = f.clip(page, true)
	peeled = f.clip(page, false)
	for _, p := range peeled {
		flap = append(flap, f.reflect(p))
	}
	return flat, peeled, flap, f, true
}

// area is the area of a polygon (shoelace formula).
func area(poly []point) float32 {
	var sum float32
	for i, a := range poly {
		b := poly[(i+1)%len(poly)]
		sum += a.x*b.y - b.x*a.y
	}
	if sum < 0 {
		sum = -sum
	}
	return sum / 2
}
