package game

import (
	"fmt"
	"strings"

	"danmin/hospital"
)

// The RECORDS, SLEEP LOGS and WARD FEED tabs, built to the approved views
// "Day - Records tab", "Day - Sleep logs tab" and "Day - Ward feed tab"
// (Figma px / 4). Capitals' tops are given, like the BEDS tab's.

// RECORDS columns (x) and rows (caps top y).
const (
	recFileX, recIDX, recRefX = 104, 212, 320
	recMarkX                  = 414 // the 24 px wide check / cross icon
	recHeadY                  = 86
	recRow0Y, recRowPitch     = 101, 15
)

var recordsField = rect{40, 62, 236, 13}

// SLEEP LOGS: rows start at logRow0Y, the four episode columns at logTimeX.
const (
	logRow0Y, logRowPitch = 81, 13
	logLevelX, logTimeX   = 136, 210
	logTimePitch          = 52
	logRowsMax            = 7
	logCols               = 4
)

func logRowY(i int) int { return logRow0Y + logRowPitch*i }

// WARD FEED: one line per admitted patient, 16 px apart.
const (
	feedRow0Y, feedRowPitch = 82, 16
	feedRowsMax             = 7
	feedNameX, feedLevelX   = 68, 164
	feedStateX, feedWhereX  = 206, 262
	feedTraceX, feedTraceW  = 352, 75
)

func feedRowY(i int) int { return feedRow0Y + feedRowPitch*i }

// records is the RECORDS tab: a name lookup, Driftwell's file next to the
// Settler ID and the referral, a check or a cross for each row.
func (p *painter) records(sh *Shift, query string) {
	g := sh.FindRecord(query)
	p.outline(recordsField, colPaper)
	p.text("NAME: "+query+"_", recordsField.x+6, 66, colPaper)
	switch {
	case g != nil:
		p.textRight("1 RECORD FOUND", textRight, 66, colPaper)
	case query != "":
		p.textRight("NO RECORD FOUND", textRight, 66, colPaper)
	default:
		p.textRight("TYPE A NAME", textRight, 66, colDim)
	}
	for _, h := range []struct {
		label string
		x     int
	}{{"FIELD", textLeft}, {"DRIFTWELL FILE", recFileX}, {"SETTLER ID", recIDX}, {"REFERRAL", recRefX}} {
		p.text(h.label, h.x, recHeadY, colDim)
	}
	p.fill(rect{textLeft, 94, textRight - textLeft, 1}, colPaper)
	if g == nil {
		p.text("> "+recordsPrompt(sh)+"_", textLeft, 195, colPaper)
		return
	}
	rows := RecordRows(g)
	for i, r := range rows {
		y := recRow0Y + recRowPitch*i
		p.text(r.Field, textLeft, y, colDim)
		p.text(r.File, recFileX, y, colPaper)
		p.text(r.ID, recIDX, y, colPaper)
		p.text(r.Referral, recRefX, y, colPaper)
		if r.Level() {
			p.pipsAfter(g.Level, r.File, recFileX, y, false)
		}
		if r.Checked {
			icon := "match"
			if !r.Match {
				icon = "mismatch"
			}
			p.iconM4(icon, recMarkX, y-5)
		}
	}
	p.text("> "+MismatchLine(rows)+"_", textLeft, 195, colPaper)
}

// Level is true for the LEVEL row, which gets pips after its text.
func (r RecordRow) Level() bool { return r.Field == "LEVEL" }

// recordsPrompt tells the player how to use an empty or unmatched field.
func recordsPrompt(sh *Shift) string {
	if sh.window() != nil {
		return "TYPE A NAME, OR PRESS TAB FOR THE WINDOW"
	}
	return "TYPE THE NAME OF SOMEONE IN THE QUEUE"
}

// sleepLogs is the SLEEP LOGS tab: every arrived patient's episodes, as
// times of the day, earliest first. 03:17 is simply the first time in each
// row on day 2; nothing points at it.
func (p *painter) sleepLogs(sh *Shift, pick string) {
	p.text("PATIENT", textLeft, 66, colDim)
	p.text("LEVEL", logLevelX, 66, colDim)
	p.text("EPISODES, EARLIEST FIRST", logTimeX, 66, colDim)
	p.fill(rect{textLeft, 74, textRight - textLeft, 1}, colPaper)
	var picked *Guest
	for i, g := range sh.Arrived() {
		if i >= logRowsMax {
			break
		}
		y := logRowY(i)
		colour := colPaper
		if strings.EqualFold(g.Name, pick) { // the picked name is shown inverted
			picked, colour = g, colInk
			p.fill(rect{textLeft - 3, y - 2, p.art.font.Width(g.Name) + 6, 9}, colPaper)
		}
		p.text(g.Name, textLeft, y, colour)
		p.pips(g.Level, logLevelX, y+2, false)
		for k, t := range g.SleepTimes() {
			if k >= logCols {
				break
			}
			p.text(t, logTimeX+logTimePitch*k, y, colPaper)
		}
	}
	p.text(fmt.Sprintf("%d LOGS  %d EPISODES SINCE DAY 1", len(sh.Arrived()), sh.EpisodeTotal()), textLeft, 172, colDim)
	prompt := "CLICK A NAME TO OPEN ITS LOG"
	if picked != nil {
		prompt = fmt.Sprintf("%s: %d ON FILE, %d TODAY, LONGEST %d MIN", strings.ToUpper(picked.Name), len(picked.Log), sh.EpisodesOf(picked), longest(picked.Log))
	}
	p.text("> "+prompt+"_", textLeft, 195, colPaper)
}

// longest is the length in minutes of the longest episode of a log.
func longest(log []LogRow) int {
	most := 0
	for _, r := range log {
		most = max(most, r.Minutes)
	}
	return most
}

// wardFeed is the WARD FEED tab: one live line per admitted patient, from
// the hospital's snapshot, with the last 60 seconds as a trace.
func (p *painter) wardFeed(sh *Shift) {
	p.text("PATIENT", feedNameX, 66, colDim)
	p.text("LEVEL", feedLevelX, 66, colDim)
	p.text("STATE", feedStateX, 66, colDim)
	p.text("WHERE", feedWhereX, 66, colDim)
	p.text("LAST 60 S", feedTraceX, 66, colDim)
	p.fill(rect{textLeft, 74, textRight - textLeft, 1}, colPaper)
	lines := 0
	for i, g := range sh.Admitted() {
		if i >= feedRowsMax {
			break
		}
		lines++
		y := feedRowY(i)
		asleep := sh.sleepers[g.ID()] != nil
		eye, state, stateColour := "eye_open", "AWAKE", colDim
		if asleep {
			eye, state, stateColour = "eye_closed", "ASLEEP", colPaper
		}
		p.iconM4(eye, textLeft, y-6)
		p.text(g.Name, feedNameX, y, colPaper)
		p.pips(g.Level, feedLevelX, y+2, false)
		p.text(state, feedStateX, y, stateColour)
		p.text(sh.whereIs(g), feedWhereX, y, colPaper)
		p.trace(sh.Trace(g.ID()), feedTraceX, y+2)
	}
	prompt := fmt.Sprintf("%d LINES LIVE. CLICK A SLEEPER TO SEND STAFF", lines)
	if lines == 0 {
		prompt = "NOBODY IS ADMITTED YET"
	}
	p.text("> "+prompt+"_", textLeft, 195, colPaper)
}

// whereIs is where the ward feed says a patient is: their room, the
// hallway (every sleeper with no bed), or the ward when awake.
func (s *Shift) whereIs(g *Guest) string {
	for _, pv := range s.snap.Patients {
		if pv.ID != g.ID() {
			continue
		}
		switch pv.State {
		case hospital.AsleepInBed:
			return fmt.Sprintf("ROOM %d", pv.Room)
		case hospital.AsleepInHallway:
			return "HALLWAY"
		}
		if s.sleepers[g.ID()] != nil {
			return "HALLWAY" // asleep and no bed yet: they lie in the hallway
		}
	}
	return "WARD A"
}

// trace draws one patient's last seconds: a low line while awake, a high
// one while asleep, joined by short strokes. y is the centre of the row.
func (p *painter) trace(t []bool, x, y int) {
	prev := false
	for i, asleep := range t {
		px := x + i*feedTraceW/traceLen
		w := x + (i+1)*feedTraceW/traceLen - px // 1 or 2 px: 60 seconds over 75 px
		top := y + 2
		colour := colDim
		if asleep {
			top, colour = y-2, colPaper
		}
		p.fill(rect{px, top, w, 1}, colour)
		if i > 0 && asleep != prev {
			p.fill(rect{px, y - 2, 1, 5}, colPaper)
		}
		prev = asleep
	}
}
