package game

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// Bending the sprites in code. The look-down desk and the page turns are not
// drawn frame by frame: the game takes a finished picture and squeezes it
// into a different four-sided shape with Ebitengine's DrawTriangles (Game
// Design -> design log, "Look-down animation, warped in code").

// point is a position on the screen, in pixels.
type point struct{ x, y float32 }

// lerp is the point a fraction f of the way from a to b.
func lerp(a, b point, f float32) point {
	return point{a.x + (b.x-a.x)*f, a.y + (b.y-a.y)*f}
}

// drawWarped draws the part of src inside srcRect so that it fills the
// four-sided shape with corners tl, tr, bl and br. The shape is cut into a
// grid of nx by ny cells, each a flat rectangle of the source drawn as two
// triangles, so a long edge bends smoothly instead of showing a diagonal
// seam. src should be an ordinary image (not a sub-image).
func drawWarped(dst, src *ebiten.Image, srcRect image.Rectangle, tl, tr, bl, br point, nx, ny int) {
	vertices := make([]ebiten.Vertex, 0, (nx+1)*(ny+1))
	for j := 0; j <= ny; j++ {
		v := float32(j) / float32(ny)
		left, right := lerp(tl, bl, v), lerp(tr, br, v)
		for i := 0; i <= nx; i++ {
			u := float32(i) / float32(nx)
			at := lerp(left, right, u)
			vertices = append(vertices, ebiten.Vertex{
				DstX: at.x, DstY: at.y,
				SrcX:   float32(srcRect.Min.X) + u*float32(srcRect.Dx()),
				SrcY:   float32(srcRect.Min.Y) + v*float32(srcRect.Dy()),
				ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1,
			})
		}
	}
	var indices []uint16
	for j := 0; j < ny; j++ {
		for i := 0; i < nx; i++ {
			a := uint16(j*(nx+1) + i) // top-left corner of the cell
			b, c, d := a+1, a+uint16(nx+1), a+uint16(nx+1)+1
			indices = append(indices, a, b, c, b, d, c)
		}
	}
	dst.DrawTriangles(vertices, indices, src, &ebiten.DrawTrianglesOptions{Filter: ebiten.FilterNearest})
}

// ease is how far along a move is after a fraction t (0 to 1) of its time:
// fast at first, slowing down.
func ease(t float64) float64 {
	t = min(max(t, 0), 1)
	return 1 - (1-t)*(1-t)*(1-t)
}

// settle is ease with a small overshoot at the end: it passes 1 by 0.6% at
// 80% of the time and comes back to 1, so the desk, 270 px tall, drops about
// a pixel and a half past its place and settles back.
func settle(t float64) float64 {
	t = min(max(t, 0), 1)
	if t < 0.8 {
		return ease(t/0.8) * 1.006
	}
	back := (t - 0.8) / 0.2
	return 1.006 - 0.006*back*back*(3-2*back) // smoothstep back down to 1
}

// lookDownTicks is how long the look-down takes: 18 ticks, 0.3 s.
const lookDownTicks = 18

// lookDownShapes says where the reception and the desk are drawn when the
// look-down has gone t of the way (0 = reception, 1 = desk): each as the
// corners of its quad. The reception slides up and squashes; the desk rises
// from the bottom as a trapezoid with a narrow far edge and flattens. At 0
// and 1 they are exactly the full screen.
func lookDownShapes(t float64) (recTL, recTR, recBL, recBR, deskTL, deskTR, deskBL, deskBR point) {
	u := float32(settle(t))
	w, h := float32(ScreenW), float32(ScreenH)
	// The reception: its top edge slides up 40 px, its bottom edge rises to the top.
	top, bottom := -40*u, h*(1-min(u, 1))
	recTL, recTR = point{0, top}, point{w, top}
	recBL, recBR = point{0, bottom}, point{w, bottom}
	// The desk: its top edge is where the reception's bottom edge is; its
	// height grows from about half to the whole screen, and the far edge is
	// pulled in by up to 40 px on each side.
	deskTop := h * (1 - u)
	deskBottom := deskTop + h*(0.5+0.5*min(u, 1))
	inset := 40 * (1 - min(u, 1))
	deskTL, deskTR = point{inset, deskTop}, point{w - inset, deskTop}
	deskBL, deskBR = point{0, deskBottom}, point{w, deskBottom}
	return
}
