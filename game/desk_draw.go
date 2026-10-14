package game

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"danmingame/sprites"
)

// Drawing the look-down desk, to the approved views "Day - Desk (folder
// arrives)", "(look-down)", "(invalid papers)", "(sleep log page)" and
// "(returning patient)" and the rects in Art Direction -> Art log ("Rects for
// the builder"). Everything is in 1x screen pixels.
const (
	folderW, folderH = 324, 214
	pageW, pageH     = 140, 180
)

var (
	folderPos  = image.Pt(12, 24) // where the folder sheet is drawn
	hingeX     = 174              // the spine: the closed cover is x 174 to 333
	idPos      = image.Pt(29, 58) // the Settler ID
	photoBox   = rect{34, 88, 44, 46}
	pagePos    = image.Pt(186, 41) // the top page of the stack
	cornerRect = rect{310, 205, 16, 16}
	inkTarget  = rect{194, 192, 53, 18} // FOR DESK USE, in screen pixels
	inkAt      = image.Pt(198, 194)
	slipPos    = image.Pt(369, 199)
	filesRect  = rect{64, 253, 88, 17}
	closedRect = rect{174, 26, 160, 210} // where a click opens the folder
)

// deskStamps are the stamps lying on the desk: where each can be picked up
// and which layer of desk_top draws it.
var deskStamps = []struct {
	stamp Stamp
	at    rect
	layer string
}{
	{StampAdmit, rect{357, 106, 30, 30}, "stamp_admit"},
	{StampDischarge, rect{397, 106, 30, 30}, "stamp_discharge"},
	{StampRefer, rect{437, 106, 30, 30}, "stamp_refer"},
}

func stampLayer(st Stamp) string {
	for _, d := range deskStamps {
		if d.stamp == st {
			return d.layer
		}
	}
	return ""
}

// inkLayer is the ink mark of a stamp in doc_stamps.
func inkLayer(st Stamp) string {
	switch st {
	case StampAdmit:
		return "ink_admit"
	case StampDischarge:
		return "ink_discharge"
	}
	return "ink_refer"
}

// The FILES tab pulses through the day palette (no red in the day).
var pulseColours = []color.RGBA{
	{0xec, 0xec, 0xe6, 0xff}, {0xdd, 0xe5, 0xd9, 0xff}, {0xc3, 0xd1, 0xc0, 0xff}, {0x9f, 0xb1, 0x9f, 0xff},
	{0xc3, 0xd1, 0xc0, 0xff}, {0xdd, 0xe5, 0xd9, 0xff},
}

const pulseStep = 180 * time.Millisecond

// filesTab draws the FILES tab beside the computer tab: pulsing through the
// palette in the day 1 tutorial while files wait, a folder mark afterwards
// while files wait, plain otherwise. On the desk the same tab says BACK.
func (p *painter) filesTab(sh *Shift) {
	switch {
	case sh.Pulsing():
		p.fill(filesRect, pulseColours[int(p.elapse/pulseStep)%len(pulseColours)])
		p.textCentered("V FILES V", filesRect.x+filesRect.w/2, 261, colInk)
	case sh.FilesWaiting():
		p.fill(filesRect, colInk)
		p.iconM4("folder", filesRect.x+17, filesRect.y)
		p.text("FILES", filesRect.x+43, 261, colPaper)
	default:
		p.fill(filesRect, colInk)
		p.textCentered("V FILES V", filesRect.x+filesRect.w/2, 261, colPaper)
	}
}

// backTab is the same tab on the desk.
func (p *painter) backTab() {
	p.fill(filesRect, colInk)
	p.textCentered("^ BACK ^", filesRect.x+filesRect.w/2, 261, colPaper)
}

// shadow draws the drop shadow of a layer: a flat deep-sage copy dx, dy
// pixels off to the side, 1 or 2 px for something lying on the desk.
func (p *painter) shadow(s *sprites.Sheet, layer string, x, y, dx, dy int) {
	img, err := s.Shadow(layer, s.FrameAt(sprites.BoilTag, p.elapse), colDim)
	p.keep(err)
	if err != nil {
		return
	}
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(float64(x+dx), float64(y+dy))
	p.screen.DrawImage(img, &op)
}

// into runs draw with a painter that paints onto buf (a scratch picture
// cleared first) instead of the screen, and keeps its error.
func (p *painter) into(buf *ebiten.Image, draw func(q *painter)) {
	buf.Clear()
	q := &painter{art: p.art, screen: buf, day: p.day, elapse: p.elapse}
	draw(q)
	p.keep(q.err)
}

// deskTopLayers are the layers of desk_top: the counter, that day's memo,
// and every stamp that is still on the desk.
func deskTopLayers(day int, held bool, heldStamp Stamp) []string {
	layers := []string{"fill", "lineart", "detail", fmt.Sprintf("memo_day%d", day)}
	for _, d := range deskStamps {
		if !(held && heldStamp == d.stamp) {
			layers = append(layers, d.layer)
		}
	}
	return layers
}

// deskObjects draws the look-down desk and everything on it onto p.screen
// (a scratch picture the scene then bends): the counter and stamps, the
// citation slip, the folder, and the stamp in the hand.
func (p *painter) deskObjects(sh *Shift, cursor image.Point) {
	st, holding := sh.Desk().Holding()
	p.sheet(p.art.deskTop, deskTopLayers(p.day, holding, st), 0, 0)
	if stamped, right, ok := sh.Slip(); ok {
		p.slip(stamped, right)
	}
	if pose, ok := sh.FolderPose(); ok {
		p.folder(sh, pose)
	}
	if holding {
		p.heldStamp(st, cursor)
	}
}

// slip is the citation slip for a wrong stamp: flavour, no penalty.
func (p *painter) slip(stamped, right Stamp) {
	p.shadow(p.art.docCitation, "fill", slipPos.X, slipPos.Y, 1, 1)
	p.sheet(p.art.docCitation, []string{"fill", "lineart", "detail"}, slipPos.X, slipPos.Y)
	l1, l2 := citationLines(stamped, right)
	p.text(l1, slipPos.X+5, slipPos.Y+14, colInk)
	p.text(l2, slipPos.X+5, slipPos.Y+22, colInk)
}

// heldStamp draws the lifted stamp at the cursor with a bigger shadow. Its
// top-left corner is 4 px up and left of the cursor, so the stamp hangs below
// and to the right of it: with the cursor on the FOR DESK USE box it never
// covers the SIGNATURE label above it.
func (p *painter) heldStamp(st Stamp, cursor image.Point) {
	layer := stampLayer(st)
	frame := p.art.deskTop.FrameAt(sprites.BoilTag, p.elapse)
	art, err := p.art.deskTop.Frame(layer, frame) // a piece of the sheet: its pixels start at where the layer sits
	p.keep(err)
	where, _ := p.art.deskTop.Rect(layer, frame)
	shade, err2 := p.art.deskTop.Shadow(layer, frame, colDim)
	p.keep(err2)
	if err != nil || err2 != nil {
		return
	}
	var box image.Rectangle
	for _, d := range deskStamps {
		if d.stamp == st {
			box = image.Rect(d.at.x-2, d.at.y-2, d.at.x+d.at.w+2, d.at.y+d.at.h+2)
		}
	}
	x, y := cursor.X-4-2, cursor.Y-4-2 // the crop starts 2 px before the art
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(float64(x+3), float64(y+4))
	p.screen.DrawImage(shade.SubImage(box).(*ebiten.Image), &op)
	op.GeoM.Reset()
	op.GeoM.Translate(float64(x), float64(y))
	p.screen.DrawImage(art.SubImage(box.Add(where.Min)).(*ebiten.Image), &op)
}

// folder draws the folder in the pose the shift says: closed, swinging open
// (bent in code), or open with the Settler ID and a page.
func (p *painter) folder(sh *Shift, pose folderPose) {
	g := pose.guest
	fx, fy := folderPos.X, folderPos.Y+pose.y
	closed := []string{"closed_fill", "closed_lineart", "closed_detail"}
	open := []string{"open_fill", "open_lineart", "open_detail"}
	switch {
	case pose.open <= 0:
		p.shadow(p.art.folder, "closed_fill", fx, fy, 2, 2)
		p.sheet(p.art.folder, closed, fx, fy)
	case pose.open >= 1:
		p.shadow(p.art.folder, "open_fill", fx, fy, 2, 2)
		p.sheet(p.art.folder, open, fx, fy)
		p.settlerID(g, fy-folderPos.Y)
		p.sheet(p.art.folder, []string{"open_clip"}, fx, fy) // the clip is over the ID
		p.pages(sh, g, pose, fy-folderPos.Y)
	default:
		p.swing(fx, fy, pose.open)
	}
	// The name and the severity pips are on the tab, closed or open.
	p.text(g.Name, 231, 28+pose.y, colInk)
	p.pipsAfter(g.Level, g.Name, 231, 28+pose.y, true)
}

// swing draws the cover opening like a book, bent in code: the closed cover
// narrows toward the spine (it swings up on its hinge), then the inside
// left cover widens away from it. e is how far open, 0 to 1.
func (p *painter) swing(fx, fy int, e float64) {
	p.into(p.art.closed, func(q *painter) {
		q.sheet(q.art.folder, []string{"closed_fill", "closed_lineart", "closed_detail"}, 0, 0)
	})
	p.into(p.art.open, func(q *painter) {
		q.sheet(q.art.folder, []string{"open_fill", "open_lineart", "open_detail"}, 0, 0)
	})
	half := hingeX - folderPos.X // 162: the spine, in the sheet
	hx, top, bottom := float32(hingeX), float32(fy), float32(fy+folderH)
	// The right half of the inside is always there, under the moving cover.
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(float64(fx+half), float64(fy))
	p.screen.DrawImage(p.art.open.SubImage(image.Rect(half, 0, folderW, folderH)).(*ebiten.Image), &op)
	span := float32(folderW - half)
	if e < 0.5 {
		s := float32(1 - 2*e) // the cover narrows from full width to nothing
		lift := 6 * (1 - s)   // and its far edge is a little shorter
		drawWarped(p.screen, p.art.closed, image.Rect(half, 0, folderW, folderH),
			point{hx, top}, point{hx + span*s, top + lift}, point{hx, bottom}, point{hx + span*s, bottom - lift}, 10, 1)
		return
	}
	s := float32(2*e - 1) // the left half widens away from the spine
	lift := 6 * (1 - s)
	drawWarped(p.screen, p.art.open, image.Rect(0, 0, half, folderH),
		point{hx - float32(half)*s, top + lift}, point{hx, top}, point{hx - float32(half)*s, bottom - lift}, point{hx, bottom}, 10, 1)
}

// settlerID draws the Settler ID clipped inside the left cover: the fields,
// and the photo, which is the visitor's window sprite cut to the photo box.
func (p *painter) settlerID(g *Guest, dy int) {
	x, y := idPos.X, idPos.Y+dy
	p.shadow(p.art.docID, "fill", x, y, 1, 1)
	p.sheet(p.art.docID, []string{"fill", "lineart", "detail"}, x, y)
	p.text(g.Name, 34, 78+dy, colInk)
	p.text(g.Home, 83, 95+dy, colInk)
	p.text(g.Expires.String(), 83, 112+dy, colInk)
	faceY := g.FaceY
	if faceY == 0 {
		faceY = 14
	}
	// The photo is a crop of the sprite whose top-left is (26, faceY).
	box := rect{photoBox.x, photoBox.y + dy, photoBox.w, photoBox.h}
	p.sheetIn(box, p.art.window[g.Sprite], windowAwake(g.Visitor), box.x-26, box.y-faceY)
}

// pages draws the stack of pages on the right: the page showing, and while
// one turns, the old page bending away over the new one.
func (p *painter) pages(sh *Shift, g *Guest, pose folderPose, dy int) {
	x, y := pagePos.X, pagePos.Y+dy
	if _, turning := sh.Turning(); turning && pose.guest == sh.Desk().FolderGuest() {
		p.curledPage(sh, g, pose.page, x, y)
		return
	}
	p.shadow(p.pageSheet(pose.page), "fill", x, y, 1, 1)
	p.pageAt(g, pose.page, x, y)
	if pose.ink {
		p.keep(p.art.docStamps.Draw(p.screen, []string{inkLayer(pose.stamp)},
			p.art.docStamps.FrameAt(sprites.BoilTag, p.elapse), inkAt.X, inkAt.Y+dy))
	}
}

// pageSheet is the sheet of a page.
func (p *painter) pageSheet(pg Page) *sprites.Sheet {
	switch pg {
	case PageSleepLog:
		return p.art.docSleepLog
	case PageWristband:
		return p.art.docWristband
	}
	return p.art.docReferral
}

// pageAt draws one page with its writing at (x, y). The text is drawn
// at the field rows listed in the Art log, 1 px below each row's top.
func (p *painter) pageAt(g *Guest, pg Page, x, y int) {
	layers := []string{"fill", "lineart", "detail"}
	switch pg {
	case PageSleepLog:
		p.sheet(p.art.docSleepLog, layers, x, y)
		p.text(g.Name, x+8, y+29, colInk)
		for i, r := range g.Log {
			if i >= 12 {
				break
			}
			row := y + 52 + 10*i
			p.text(r.Date.String(), x+8, row, colInk)
			p.text(r.TimeString(), x+50, row, colInk)
			p.text(fmt.Sprint(r.Minutes), x+92, row, colInk)
		}
	case PageWristband:
		p.sheet(p.art.docWristband, layers, x, y)
		p.text(g.Ward, x+51, y+35, colInk)
		p.pipsAfter(g.Level, g.Ward, x+51, y+35, true)
		p.text(g.Name, x+33, y+44, colInk)
		p.text(g.LastStay.String(), x+8, y+71, colInk)
	default:
		stamp := "stamp_barrow"
		switch g.Stamp {
		case StampOther:
			stamp = "stamp_other"
		case StampNone:
			stamp = ""
		}
		layers = append(layers, "signature")
		if stamp != "" {
			layers = append(layers, stamp)
		}
		p.sheet(p.art.docReferral, layers, x, y)
		p.text(g.RefDate.String(), x+8, y+29, colInk)
		p.text(fmt.Sprint(g.Age), x+72, y+29, colInk)
		p.text(g.RefNameWritten(g.Name), x+8, y+44, colInk)
		p.text(g.Home, x+8, y+59, colInk)
		p.text(g.Diagnosis, x+8, y+74, colInk)
		p.text(g.Notes[0], x+8, y+89, colInk)
		p.text(g.Notes[1], x+8, y+98, colInk)
	}
}

// backOfPage is the colour of the reverse of a page: mint, so it reads as
// the other side (no mirrored writing needed).
var backOfPage = color.RGBA{0xdd, 0xe5, 0xd9, 0xff}

// curledPage draws the top page being turned like a real page (curl.go): the
// grabbed corner follows the cursor, the part still flat shows the page, the
// part peeled away shows the next page underneath with a thin shadow along the
// fold, and the peeled part is mirrored across the fold as the back of the
// page, with a drop shadow under it. When the turn is finished the page fades
// down to the back of the stack.
func (p *painter) curledPage(sh *Shift, g *Guest, top Page, x, y int) {
	under := nextPage(g, top)
	p.shadow(p.pageSheet(under), "fill", x, y, 1, 1)
	p.pageAt(g, under, x, y)
	p.into(p.art.page, func(q *painter) { q.pageAt(g, top, 0, 0) })

	corner, tuck := sh.Curl()
	flat, peeled, flap, f, ok := curl(pageW, pageH, corner)
	origin := point{float32(x), float32(y) + float32(8*tuck)} // the tucked page sinks a little
	alpha := float32(1 - tuck)
	if ok {
		// A thin shadow along the fold, on the page underneath: the strip of
		// the peeled part within 3 px of the fold.
		near := fold{m: point{f.m.x - 3*f.n.x, f.m.y - 3*f.n.y}, n: f.n}
		p.poly(near.clip(peeled, true), origin, nil, colDim, 1)
	}
	p.poly(flat, origin, p.art.page, color.RGBA{0xff, 0xff, 0xff, 0xff}, alpha)
	if ok {
		flap = clipToPage(flap)
		shade := make([]point, len(flap))
		for i, v := range flap {
			shade[i] = point{v.x + 1, v.y + 2} // the usual drop shadow, a little bigger
		}
		p.poly(clipToPage(shade), origin, nil, colDim, alpha)
		p.poly(flap, origin, nil, backOfPage, alpha)
	}
}

// clipToPage keeps the part of a polygon inside the page rectangle, so a
// flap or a shadow never leaves the page (and never covers the Settler ID).
func clipToPage(poly []point) []point {
	for _, edge := range []fold{
		{point{0, 0}, point{1, 0}}, {point{0, 0}, point{0, 1}},
		{point{pageW, pageH}, point{-1, 0}}, {point{pageW, pageH}, point{0, -1}},
	} {
		if len(poly) == 0 {
			return nil
		}
		poly = edge.clip(poly, true)
	}
	return poly
}

// poly draws a convex polygon (page pixels, at origin on the screen) as a
// fan of triangles. With src it is textured from that picture (the polygon's
// own coordinates are the picture's); with nil src it is one flat colour.
func (p *painter) poly(pts []point, origin point, src *ebiten.Image, c color.RGBA, alpha float32) {
	if len(pts) < 3 {
		return
	}
	img := src
	if img == nil {
		img = p.art.white
	}
	r, g, b := float32(c.R)/255*alpha, float32(c.G)/255*alpha, float32(c.B)/255*alpha // premultiplied
	vertices := make([]ebiten.Vertex, len(pts))
	for i, v := range pts {
		vertices[i] = ebiten.Vertex{DstX: origin.x + v.x, DstY: origin.y + v.y, SrcX: v.x, SrcY: v.y, ColorR: r, ColorG: g, ColorB: b, ColorA: alpha}
		if src == nil {
			vertices[i].SrcX, vertices[i].SrcY = 0.5, 0.5
		}
	}
	var indices []uint16
	for i := 1; i < len(pts)-1; i++ {
		indices = append(indices, 0, uint16(i), uint16(i+1))
	}
	p.screen.DrawTriangles(vertices, indices, img, &ebiten.DrawTrianglesOptions{Filter: ebiten.FilterNearest})
}
