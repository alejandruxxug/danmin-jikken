package night

import (
	"github.com/hajimehoshi/ebiten/v2"

	"danmingame/sprites"
)

// Hana's night report of day 3 (doc_report, 140x180): the page in the desk
// folder and the match cut of the realization draw the same words with the
// same code, so the two signatures are pixel for pixel the same.

// What is written on the report (Alejo approved the words, 2026-10-14).
var reportNotes = []string{"All asleep. All happy!", "Nobody woke up."}

const reportSignature = "- Hana"

// Where things are written on the page, relative to its top-left corner
// (Art log, cinematic pass). The numbers are baselines.
const (
	reportWardX, reportDateX, reportHeadBase = 30, 94, 35
	reportNurseX, reportNurseBase            = 30, 50
	reportSignRight, reportSignBase          = 131, 159
)

// reportNotesPaper is the notes box: rows on baselines 69, 80, 91, ...
var reportNotesPaper = Paper{Left: 8, Right: 132, FirstBase: 69, RowStep: 11}

// ReportHand says whose hand wrote the report: Hana's own, unless Mayu
// forged it (route_hana: "it wasn't Hana who wrote it").
func ReportHand(f Flags) (sheet string, style Style) {
	if f.Has("route_hana") {
		return "hand_mayu", MayuStyle
	}
	return "hand_hana", HanaStyle
}

// WriteReport writes the report's words on a page whose top-left corner is
// at (x, y). hand is the sheet ReportHand named.
func WriteReport(dst *ebiten.Image, hand *sprites.Sheet, style Style, x, y int) error {
	m := SheetMetrics{hand.Meta}
	var first error
	glyph := func(p Placed, left, base int) {
		if err := hand.Draw(dst, []string{p.Glyph}, 0, x+left+p.X, y+base-glyphBaseRow+p.Y); err != nil && first == nil {
			first = err
		}
	}
	put := func(text string, left, base int) {
		placed, _ := Layout(text, m, style)
		for _, p := range placed {
			glyph(p, left, base)
		}
	}
	put("C", reportWardX, reportHeadBase)
	put("Day 3", reportDateX, reportHeadBase)
	put("Hana", reportNurseX, reportNurseBase)
	for _, row := range PaperRows(reportNotes, reportNotesPaper, m, style) {
		for _, p := range row.Glyphs {
			glyph(p, row.X, row.Base)
		}
	}
	_, w := Layout(reportSignature, m, style)
	put(reportSignature, reportSignRight-w, reportSignBase)
	return first
}
