package game

import (
	"fmt"
	"strings"
	"time"

	"danmin/hospital"

	"github.com/hajimehoshi/ebiten/v2"

	"danmingame/sprites"
)

// The tabs of the patient computer. Only BEDS exists in M3; the others come
// with the documents and the queries in M4 and show a placeholder.
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
// 16 px icons are stacked 14 px apart so the second one ends at y 162, just
// above the header at 163. The first row is exactly as in the approved view.
const (
	hallwayRows  = 2
	hallwayPitch = 14
)

// hallwayRow is where the k-th sleeper with no bed is shown.
func hallwayRow(k int) rect { return rect{40, 132 + hallwayPitch*k, 290, hallwayPitch} }

// computerScene is the flip-up patient computer over the reception. The
// day goes on behind it: patients still fall asleep while you look at it.
type computerScene struct {
	env     *Env
	art     *DeskArt
	shift   *Shift
	ticks   int
	drawErr error
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
	if x, y, ok := s.env.Input.Click(); ok {
		s.click(x, y)
	}
	return nil
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
	if s.env.Tab != TabBeds {
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
	if s.env.Tab == TabBeds {
		p.beds(s.shift)
	} else {
		p.offline(s.env.Tab)
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

// offline is the placeholder of the tabs that M4 builds.
func (p *painter) offline(tab string) {
	p.text(strings.ToUpper(tab)+": MODULE OFFLINE", textLeft, 75, colPaper)
	p.text("DRIFTWELL SYSTEMS IS CONNECTING THIS", textLeft, 95, colDim)
	p.text("TERMINAL TO THE WARD. PLEASE WAIT.", textLeft, 107, colDim)
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
			p.text(g.Name+"  "+levelName(g.Level), card.x+6, 97, colPaper)
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
		p.text(fmt.Sprintf("%s  %s  %02d:%02d", g.Name, levelName(g.Level), secs/60, secs%60), 72, row.y+7, colPaper)
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
