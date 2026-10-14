package game

import (
	"fmt"
	"strings"
)

// What the patient computer knows about the people in today's queue: the
// RECORDS lookup, the SLEEP LOGS table and the WARD FEED traces. Nothing
// here draws; computer.go reads it.

// Arrived lists the guests who have walked up so far, in queue order. The
// computer only knows people who have come to the desk.
func (s *Shift) Arrived() []*Guest {
	var out []*Guest
	for _, g := range s.Guests {
		if g.status != guestComing {
			out = append(out, g)
		}
	}
	return out
}

// FindRecord is the first arrived guest whose name starts with query,
// ignoring case. An empty query finds nobody.
func (s *Shift) FindRecord(query string) *Guest {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}
	for _, g := range s.Arrived() {
		if strings.HasPrefix(strings.ToUpper(g.Name), strings.ToUpper(query)) {
			return g
		}
	}
	return nil
}

// RecordRow is one line of the RECORDS table: the same field as Driftwell's
// file has it, as the Settler ID shows it and as the referral writes it.
type RecordRow struct {
	Field, File, ID, Referral string
	Checked                   bool // the row has a check mark column
	Match                     bool // the file and the papers agree
}

// fileDiagnosis is what Driftwell's own file says: the diagnosis of every
// patient of the sleep unit. Someone whose referral names something that
// is not a sleep disorder has no such file entry.
func fileDiagnosis(v Visitor) string {
	if isSleepDisorder(v.Diagnosis) {
		return v.Diagnosis
	}
	return "NOT ON FILE"
}

// RecordRows compares what Driftwell's file says about a guest with their
// papers, field by field.
func RecordRows(g *Guest) []RecordRow {
	name := strings.ToUpper(g.Name)
	ref := strings.ToUpper(g.RefNameWritten(g.Name))
	age := fmt.Sprint(g.Age)
	diag := fileDiagnosis(g.Visitor)
	return []RecordRow{
		{"NAME", name, name, ref, true, name == ref},
		{"AGE", age, age, age, true, true},
		{"HOME", g.Home, g.Home, g.Home, true, true},
		{"DIAGNOSIS", diag, "-", g.Diagnosis, true, diag == g.Diagnosis},
		{"LEVEL", levelName(g.Level), "-", "-", false, true},
	}
}

// Mismatches names the rows whose file and papers disagree.
func Mismatches(rows []RecordRow) []string {
	var out []string
	for _, r := range rows {
		if !r.Match {
			out = append(out, r.Field)
		}
	}
	return out
}

// MismatchLine is the sentence under the table.
func MismatchLine(rows []RecordRow) string {
	bad := Mismatches(rows)
	switch len(bad) {
	case 0:
		return "NO MISMATCH IN THESE FIELDS"
	case 1:
		return "1 MISMATCH: THE REFERRAL " + bad[0] + ". CHECK THE PAPERS"
	}
	return fmt.Sprintf("%d MISMATCHES: THE REFERRAL %s. CHECK THE PAPERS", len(bad), strings.Join(bad, ", "))
}

// EpisodeTotal counts every episode on the sleep logs of the arrived guests,
// plus the ones the hospital recorded today.
func (s *Shift) EpisodeTotal() int {
	n := 0
	for _, g := range s.Arrived() {
		n += len(g.Log) + s.EpisodesOf(g)
	}
	return n
}

// traceLen is how many seconds of history the ward feed keeps per patient.
const traceLen = 60

// Trace is a patient's last traceLen game seconds: true where they slept.
// The newest second is last.
func (s *Shift) Trace(id string) []bool { return s.trace[id] }

// sampleTraces adds one second to every admitted patient's trace. Update
// calls it once per game second, from the snapshot.
func (s *Shift) sampleTraces() {
	sec := int(s.Seconds())
	if sec == s.traceAt {
		return
	}
	s.traceAt = sec
	for _, p := range s.snap.Patients {
		t := append(s.trace[p.ID], s.sleepers[p.ID] != nil)
		if len(t) > traceLen {
			t = t[len(t)-traceLen:]
		}
		s.trace[p.ID] = t
	}
}

// Admitted lists the guests who were stamped ADMIT, in queue order: the
// people the ward feed has a line for.
func (s *Shift) Admitted() []*Guest {
	var out []*Guest
	for _, g := range s.Guests {
		if g.status == guestAdmitted {
			out = append(out, g)
		}
	}
	return out
}
