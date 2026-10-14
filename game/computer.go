package game

import (
	"fmt"
	"strings"
	"time"

	"danmin/hospital"

	"github.com/hajimehoshi/ebiten/v2"

	"danmingame/sprites"
)

// The four tabs of the patient computer.
const (
	TabRecords   = "records"
	TabSleepLogs = "sleep logs"
	TabBeds      = "beds"
	TabWardFeed  = "ward feed"
)

// Where the terminal sits on the screen, and its text. All of these come
// from the approved views "Day - Beds panel" and "Top-10" (Figma px / 4).
const (
	termX, termY = 16, 6 // top-left of the 448 x 236 computer sprite
	textLeft     = 40    // the terminal text column
	textRight    = 436   // where right-aligned text ends
)

// tabPad is the width a tab's box adds to its label: 6 px before the text
// and 13 after, as in the Figma view.
const tabPad = 19

// tabLabels are the terminal's tabs: their name, the text, and its x.
var tabLabels = []struct {
	id, label string
	x         int
}{
	{TabRecords, "RECORDS", 40},
	{TabSleepLogs, "SLEEP LOGS", 108},
	{TabBeds, "BEDS", 194},
	{TabWardFeed, "WARD FEED", 240},
}

var (
	closeRect   = tabRect // the same black tab, now saying CLOSE
	callStaff   = rect{340, 132, 100, 16}
	computerLay = []string{"fill", "lineart", "detail"}
)

// roomCard is where room number i (0, 1, 2) is shown on the BEDS tab.
func roomCard(i int) rect { return rect{40 + 136*i, 64, 128, 44} }

// hallwayRows rows fit between the HALLWAY header and STAFF ON SHIFT; the
// 16 px icons are stacked 12 px apart (their art is only 8 px tall), so the
// second bed ends well above the header at 163. The first row is exactly as
// in the approved view.
const (
	hallwayRows  = 2
	hallwayPitch = 12
)

// hallwayRow is where the k-th sleeper with no bed is shown.
func hallwayRow(k int) rect { return rect{40, 132 + hallwayPitch*k, 290, hallwayPitch} }

// computerScene is the flip-up patient computer over the reception. The
// day goes on behind it: patients still fall asleep while you look at it.
type computerScene struct {
	startTab string // the tab to open on Enter; "" keeps the open one (BEDS by default)
	env      *Env
	art      *DeskArt
	shift    *Shift
	ticks    int
	drawErr  error
}

func (s *computerScene) Enter(env *Env) error {
	s.env = env
	var err error
	if s.art, err = env.Art(); err != nil {
		return err
	}
	if err = env.StartShift(); err != nil {
		return err
	}
	s.shift = env.Shift()
	if s.startTab != "" {
		env.Tab = s.startTab
	}
	if env.Tab == "" {
		env.Tab = TabBeds
	}
	return nil
}

// Exit forgets the open tab, so the next visit (and the dump) starts on BEDS.
func (s *computerScene) Exit() { s.env.Tab = "" }

func (s *computerScene) Update() error {
	if s.drawErr != nil {
		return s.drawErr
	}
	s.ticks++
	if s.env.Tab == TabRecords {
		s.typeQuery()
	}
	if x, y, ok := s.env.Input.Click(); ok {
		s.click(x, y)
	}
	return nil
}

// maxQuery is how many characters the RECORDS name field takes.
const maxQuery = 20

// typeQuery edits the RECORDS name field: typed letters are added,
// Backspace removes one, and Tab fills in the guest at the window.
func (s *computerScene) typeQuery() {
	in := s.env.Input
	for _, r := range in.Typed() {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == ' ' || r == '.' {
			if len(s.env.Query) < maxQuery {
				s.env.Query += strings.ToUpper(string(r))
			}
		}
	}
	if in.KeyPressed(ebiten.KeyBackspace) && s.env.Query != "" {
		s.env.Query = s.env.Query[:len(s.env.Query)-1]
	}
	if in.KeyPressed(ebiten.KeyTab) {
		if g := s.shift.window(); g != nil {
			s.env.Query = strings.ToUpper(g.Name)
		}
	}
}

func (s *computerScene) click(x, y int) {
	if closeRect.has(x, y) {
		s.env.SwitchTo(SceneDayReception)
		return
	}
	for _, t := range tabLabels {
		if (rect{t.x - 6, 37, s.art.font.Width(t.label) + tabPad, 14}).has(x, y) {
			s.env.Tab = t.id
			return
		}
	}
	switch s.env.Tab {
	case TabSleepLogs:
		for i, g := range s.shift.Arrived() {
			if i < logRowsMax && (rect{textLeft, logRowY(i) - 3, 90, 12}).has(x, y) {
				s.env.LogPick = g.Name
			}
		}
		return
	case TabWardFeed:
		for i, g := range s.shift.Admitted() {
			if i < feedRowsMax && (rect{textLeft, feedRowY(i) - 5, textRight - textLeft, 14}).has(x, y) && s.shift.needsStaff(g) {
				att, err := s.shift.Dispatch(g)
				s.afterDispatch(att, err)
			}
		}
		return
	case TabRecords:
		return
	}
	waiting := s.shift.Waiting()
	if callStaff.has(x, y) && len(waiting) > 0 {
		att, err := s.shift.CallStaff()
		s.afterDispatch(att, err)
		return
	}
	for k, g := range waiting {
		if k < hallwayRows && hallwayRow(k).has(x, y) {
			att, err := s.shift.Dispatch(g)
			s.afterDispatch(att, err)
			return
		}
	}
}

// afterDispatch shows the alert when there was no room: it is drawn over
// the reception, so the computer closes to let you see it.
func (s *computerScene) afterDispatch(att Attendance, err error) {
	if err == nil && att.NoRoom {
		showAlert(s.env)
		s.env.SwitchTo(SceneDayReception)
	}
}

func (s *computerScene) Draw(screen *ebiten.Image) {
	p := &painter{art: s.art, screen: screen, day: s.env.Day, elapse: time.Duration(s.ticks) * time.Second / 60}
	p.terminal(s.shift, int(s.env.CameraX), "V CLOSE COMPUTER V")
	p.terminalHeader(fmt.Sprintf("DAY %d  %s", s.shift.Day, s.env.Clock), s.env.Tab)
	switch s.env.Tab {
	case TabRecords:
		p.records(s.shift, s.env.Query)
	case TabSleepLogs:
		p.sleepLogs(s.shift, s.env.LogPick)
	case TabWardFeed:
		p.wardFeed(s.shift)
	default:
		p.beds(s.shift)
	}
	p.glass()
	if p.err != nil && s.drawErr == nil {
		s.drawErr = p.err
	}
}

// terminal draws the reception behind, the computer over it and the close
// tab. sh may be nil.
func (p *painter) terminal(sh *Shift, cam int, closeLabel string) {
	p.room(sh, cam)
	p.sheet(p.art.computer, computerLay, termX, termY)
	p.tab(closeLabel)
}

// terminalHeader draws the title, the day and clock, and the tabs.
func (p *painter) terminalHeader(clock, tab string) {
	p.text("DRIFTWELL PATIENT TERMINAL", textLeft, 27, colPaper)
	p.textRight(clock, textRight, 27, colPaper)
	for _, t := range tabLabels {
		if t.id == tab {
			p.fill(rect{t.x - 6, 37, p.art.font.Width(t.label) + tabPad, 14}, colPaper)
			p.text(t.label, t.x, 43, colInk)
		} else {
			p.text(t.label, t.x, 43, colDim)
		}
	}
	p.fill(rect{textLeft, 57, textRight - textLeft + 4, 1}, colPaper)
}

// icon draws one frame-boiling icon of ui_icons.
func (p *painter) icon(name string, x, y int) {
	p.keep(p.art.icons.Draw(p.screen, []string{name}, p.art.icons.FrameAt(sprites.BoilTag, p.elapse), x, y))
}

// iconM4 draws one icon of ui_icons_m4 (eyes, check and cross, folder).
func (p *painter) iconM4(name string, x, y int) {
	p.keep(p.art.iconsM4.Draw(p.screen, []string{name}, p.art.iconsM4.FrameAt(sprites.BoilTag, p.elapse), x, y))
}

// pips draws the severity dots (Mild 1, Moderate 2, Severe 3) with their
// top-left corner at (x, y). ink is true on paper (black dots) and false on
// the terminal (light dots).
func (p *painter) pips(l hospital.NarcolepsyLevel, x, y int, ink bool) {
	name := fmt.Sprintf("sev_%d", Pips(l))
	if ink {
		name += "_ink"
	}
	p.icon(name, x, y)
}

// pipsAfter draws the pips 4 px after a text that starts at textX and whose
// capitals start at capsY (the Art log's rule: 2 px below the caps).
func (p *painter) pipsAfter(l hospital.NarcolepsyLevel, text string, textX, capsY int, ink bool) {
	p.pips(l, textX+p.art.font.Width(text)+4, capsY+2, ink)
}

// beds is the BEDS tab: the rooms, the sleepers with no bed, the staff.
func (p *painter) beds(sh *Shift) {
	snap := sh.Snapshot()
	for i, room := range snap.Rooms {
		card := roomCard(i)
		p.outline(card, colPaper)
		p.text(fmt.Sprintf("ROOM %d", room.Number), card.x+38, 73, colPaper)
		if len(room.Occupants) == 0 {
			p.icon("bed_free", card.x+6, 70)
			p.text("FREE", card.x+38, 85, colPaper)
			p.text("READY FOR A PATIENT", card.x+6, 97, colPaper)
			continue
		}
		p.icon("bed_occupied", card.x+6, 70)
		p.icon("fx_z", card.x+6, 70)
		p.text("ASLEEP", card.x+38, 85, colPaper)
		if g := sh.GuestByPatient(room.Occupants[0]); g != nil {
			// The pips are right-aligned in the card (10 px wide, 6 px from its
			// edge); when the name and level would run into them the level
			// word is left out and the pips say it.
			label := g.Name + "  " + levelName(g.Level)
			if p.art.font.Width(label) > 100 {
				label = g.Name
			}
			p.text(label, card.x+6, 97, colPaper)
			p.pips(g.Level, card.x+card.w-16, 99, false)
		}
	}

	waiting := sh.Waiting()
	p.text(fmt.Sprintf("HALLWAY  ASLEEP, NO BED (%d)", len(waiting)), textLeft, 121, colPaper)
	if more := len(waiting) - hallwayRows; more > 0 {
		p.textRight(fmt.Sprintf("+%d MORE", more), textRight, 121, colDim)
	}
	for k, g := range waiting {
		if k >= hallwayRows {
			break
		}
		row := hallwayRow(k)
		p.icon("sleeper", row.x, row.y)
		secs := int(sh.SleptFor(g))
		label := g.Name + "  " + levelName(g.Level)
		p.text(label, 72, row.y+7, colPaper)
		p.pipsAfter(g.Level, label, 72, row.y+7, false)
		// The time follows the pips (3 dots are 10 px wide at most).
		after := 72 + p.art.font.Width(label) + 4 + 4*Pips(g.Level) + 4
		p.text(fmt.Sprintf("%02d:%02d", secs/60, secs%60), after, row.y+7, colPaper)
	}
	if len(waiting) > 0 {
		p.fill(callStaff, colPaper)
		p.text("CALL STAFF", callStaff.x+10, callStaff.y+7, colInk)
	}

	p.text("STAFF ON SHIFT", textLeft, 163, colPaper)
	var staff []string
	for _, name := range sh.StaffNames() {
		state := "FREE"
		if sh.Busy(name) {
			state = "BUSY"
		}
		staff = append(staff, name+": "+state)
	}
	p.text(strings.Join(staff, "  "), textLeft, 177, colPaper)

	prompt := "CLICK A SLEEPER TO SEND STAFF"
	if msg := sh.Message(); msg != "" {
		prompt = msg
	}
	p.text("> "+prompt+"_", textLeft, 195, colPaper)
}

// levelName is the narcolepsy level in capitals.
func levelName(l hospital.NarcolepsyLevel) string {
	return strings.ToUpper(l.String())
}
