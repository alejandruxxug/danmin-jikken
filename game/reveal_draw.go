package game

import "github.com/hajimehoshi/ebiten/v2"

// revealView draws the x4 view of the YOU reveal: the window, you and the
// counter, with no HUD and no glass, 120 x 67.5 world pixels blown up to the
// whole 480 x 270 picture. The room is first drawn at 1x into a scratch
// picture with the camera at its far right (world x 240 to 720), then the
// scratch is scaled by 4 with nearest-neighbour so every pixel stays a block.
func (p *painter) revealView(r *Reveal) {
	const cam = MaxCameraX
	view := p.art.view
	view.Clear()
	q := &painter{art: p.art, screen: view, day: p.day, elapse: p.elapse}
	q.sheet(p.art.room, roomLayers, -cam, 0)
	q.sheet(p.art.room, []string{q.callLayer(nil)}, -cam, 0)
	if g := r.Guest; g != nil {
		x, y := windowSpot.x, windowSpot.y // after the arrival you stand still at the window
		bob := 0
		if r.Stage == RevealArrival {
			x, y, bob = int(g.x), int(g.y), WalkBob(g.steps)
		}
		layers := eyesFor(windowAwake(g.Visitor), r.Blinking())
		q.windowGuest(g, layers, x-cam, y+bob, cam)
	}
	q.sheet(p.art.desk, deskLayers(p.day), -cam, 0)
	p.keep(q.err)

	var op ebiten.DrawImageOptions
	op.GeoM.Translate(-revealLeft+cam, -float64(r.TopY())) // the view's corner to the scratch's origin
	op.GeoM.Scale(revealZoom, revealZoom)
	op.Filter = ebiten.FilterNearest
	p.screen.DrawImage(view, &op)
}
