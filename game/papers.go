package game

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"sort"
	"strings"
)

// The papers every visitor hands over (Game Design -> The Papers, Please
// layer). The desk, the folder and the computer all draw from this data, and
// Problem works out what is wrong with the papers from the papers
// themselves, so the stamp that scores (RightStamp) can never disagree with
// what the player can read.

// Date is a calendar date in the game's year, 1977, written MM/DD/YY.
type Date struct{ Month, Day, Year int }

func (d Date) String() string { return fmt.Sprintf("%02d/%02d/%02d", d.Month, d.Day, d.Year%100) }

// Short is the date without the year, MM/DD.
func (d Date) Short() string { return fmt.Sprintf("%02d/%02d", d.Month, d.Day) }

// Before says whether d is an earlier day than other.
func (d Date) Before(other Date) bool {
	if d.Year != other.Year {
		return d.Year < other.Year
	}
	if d.Month != other.Month {
		return d.Month < other.Month
	}
	return d.Day < other.Day
}

// dayDate is the date of a shift: day 1 is 10/11/77, day 2 10/12, day 3 10/13.
func dayDate(day int) Date { return Date{10, 10 + day, 77} }

// DoctorStamp says whose stamp the referral carries.
type DoctorStamp int

const (
	StampBarrow DoctorStamp = iota // Dr. Barrow, head of the sleep unit: the only valid one
	StampOther                     // another doctor (Dr. Voss): not valid
	StampNone                      // no stamp at all
)

// LogRow is one line of a sleep log: one episode.
type LogRow struct {
	Date    Date
	At      int // minutes since midnight
	Minutes int // how long it lasted
}

// TimeString is the time of day, "03:17".
func (r LogRow) TimeString() string { return fmt.Sprintf("%02d:%02d", r.At/60, r.At%60) }

// Papers is what one visitor carries. The name on the Settler ID is the
// visitor's own Name.
type Papers struct {
	Today Date // the shift's date (set when the days are built)

	// Settler ID.
	Home    string
	Expires Date

	// Referral.
	RefDate   Date
	RefName   string // the name as the referral writes it; "" means the same as the ID
	Diagnosis string
	Notes     [2]string
	Stamp     DoctorStamp

	Log []LogRow // the sleep log, oldest first

	// Wristband (returning patients only).
	Ward     string
	LastStay Date
}

// RefNameWritten is the name the referral shows.
func (p Papers) RefNameWritten(idName string) string {
	if p.RefName != "" {
		return p.RefName
	}
	return idName
}

// sleepDisorders are the diagnoses that make a referral valid (Game Design
// -> What each stamp means, rule 4).
var sleepDisorders = []string{"NARCOLEPSY", "CATAPLEXY", "SLEEP ATTACKS"}

func isSleepDisorder(diagnosis string) bool {
	for _, d := range sleepDisorders {
		if strings.EqualFold(diagnosis, d) {
			return true
		}
	}
	return false
}

// Problem says what is wrong with the papers, "" when the referral is valid.
// The four checks are the four rules of Game Design -> What each stamp means.
func (v Visitor) Problem() string {
	switch {
	case v.Expires.Before(v.Today):
		return "Settler ID expired"
	case !strings.EqualFold(v.RefNameWritten(v.Name), v.Name):
		return "name does not match the ID"
	case v.Stamp != StampBarrow:
		return "no stamp from Dr. Barrow"
	case !isSleepDisorder(v.Diagnosis):
		return "diagnosis is not a sleep disorder"
	}
	return ""
}

// Reasons written on the citation slip (Game Design -> design log: "Day 1
// is a real tutorial"). Each fits the 86 px line of the slip.
const (
	reasonExpired     = "ID EXPIRED"
	reasonName        = "NAME MISMATCH"
	reasonNoBarrow    = "NO DR. BARROW"
	reasonDiagnosis   = "WRONG DIAGNOSIS"
	reasonPapersOK    = "PAPERS OK"
	reasonWristband   = "HAS WRISTBAND"
	reasonNoWristband = "NO WRISTBAND"
)

// problemReasons turns what Problem says into the words of the slip, so the
// checks are written once, in Problem.
var problemReasons = map[string]string{
	"Settler ID expired":                reasonExpired,
	"name does not match the ID":        reasonName,
	"no stamp from Dr. Barrow":          reasonNoBarrow,
	"diagnosis is not a sleep disorder": reasonDiagnosis,
}

// Reason says why the right stamp is the right one, in the words of the
// slip: the first failed check of Problem, or for valid papers whether the
// visitor wears a wristband. The hints read it too, to know what a patient
// teaches.
func (v Visitor) Reason() string {
	switch p := v.Problem(); {
	case p != "":
		return problemReasons[p]
	case v.Returning:
		return reasonWristband
	}
	return reasonPapersOK
}

// CitationReason is the first line of the slip for a wrong stamp: the
// visitor's Reason, except that a DISCHARGE for somebody with valid papers
// and no wristband says so.
func CitationReason(v Visitor, stamped Stamp) string {
	if r := v.Reason(); r != reasonPapersOK || stamped != StampDischarge {
		return r
	}
	return reasonNoWristband
}

// ReferralOK is true when the referral is valid.
func (v Visitor) ReferralOK() bool { return v.Problem() == "" }

// NewEpisodes counts the sleep log rows after the last stay: what a
// returning patient has had since the wristband was put on.
func (v Visitor) NewEpisodes() int {
	n := 0
	for _, r := range v.Log {
		if v.LastStay.Before(r.Date) {
			n++
		}
	}
	return n
}

// SleepTimes lists the times of day the episodes happen at, earliest first
// and each time once, as the SLEEP LOGS tab shows them.
func (v Visitor) SleepTimes() []string {
	rows := append([]LogRow(nil), v.Log...)
	sort.Slice(rows, func(i, j int) bool { return rows[i].At < rows[j].At })
	var out []string
	for _, r := range rows {
		if t := r.TimeString(); len(out) == 0 || out[len(out)-1] != t {
			out = append(out, t)
		}
	}
	return out
}

// episodeAt317 is 03:17 in minutes since midnight: the pattern every day 2
// sleep log shows (Game Design -> The sleep mechanic).
const episodeAt317 = 3*60 + 17

// makeLog writes a visitor's sleep log. It is plain data made once from the
// name, so the same visitor always has the same log. On day 2 every log has
// an episode at 03:17; a returning patient's log stops at their last stay.
func makeLog(v Visitor, day int) []LogRow {
	h := fnv.New64a()
	h.Write([]byte(v.Name))
	rng := rand.New(rand.NewSource(int64(h.Sum64() >> 1)))

	last := Date{10, 10 + day - 1, 77} // the day before the shift
	if v.Returning {
		last = v.LastStay
	}
	var rows []LogRow
	for d := last.Day - 4; d <= last.Day; d++ {
		date := Date{10, d, 77}
		if day == 2 && d >= last.Day-2 {
			rows = append(rows, LogRow{date, episodeAt317, 30 + rng.Intn(20)})
		}
		if rng.Intn(3) > 0 { // most days have one daytime episode
			rows = append(rows, LogRow{date, 9*60 + rng.Intn(10*60), 8 + rng.Intn(40)})
		}
	}
	if len(rows) == 0 { // a patient always has at least one episode on record
		rows = append(rows, LogRow{last, 13*60 + 5, 9})
	}
	return rows
}
