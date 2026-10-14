package game

import (
	"strings"
	"testing"
)

// arriveAll lets every guest of the day walk up.
func arriveAll(sh *Shift) {
	for _, g := range sh.Guests {
		g.status = guestWaiting
	}
}

func TestFindRecord(t *testing.T) {
	sh := newTestShift(t, 2)
	if sh.FindRecord("marc") != nil {
		t.Fatal("found someone who has not arrived")
	}
	arriveAll(sh)
	if g := sh.FindRecord("marc"); g == nil || g.Name != "Marcus Trell" {
		t.Errorf("FindRecord(marc) = %v", g)
	}
	if sh.FindRecord("") != nil || sh.FindRecord("zzz") != nil {
		t.Error("an empty or unknown name should find nobody")
	}
}

// The RECORDS table flags exactly the fields that disagree: Marcus Trell's
// referral spells his name differently, and everyone else's table is clean
// except the one whose diagnosis is not on Driftwell's file.
func TestRecordRowsMismatch(t *testing.T) {
	sh := newTestShift(t, 2)
	marcus := sh.Guests[4]
	bad := Mismatches(RecordRows(marcus))
	if len(bad) != 1 || bad[0] != "NAME" {
		t.Errorf("Marcus mismatches = %v, want [NAME]", bad)
	}
	if line := MismatchLine(RecordRows(marcus)); !strings.HasPrefix(line, "1 MISMATCH: THE REFERRAL NAME") {
		t.Errorf("line = %q", line)
	}
	if line := MismatchLine(RecordRows(sh.Guests[0])); !strings.HasPrefix(line, "NO MISMATCH") {
		t.Errorf("a clean record says %q", line)
	}
	day3 := newTestShift(t, 3)
	if bad := Mismatches(RecordRows(day3.Guests[2])); len(bad) != 1 || bad[0] != "DIAGNOSIS" {
		t.Errorf("Pearl Wick mismatches = %v, want [DIAGNOSIS]", bad)
	}
}

// The ward feed keeps the last 60 seconds of each admitted patient.
func TestTraceKeepsSixtySeconds(t *testing.T) {
	sh := newTestShift(t, 1)
	g := sh.Guests[1]
	putToSleep(t, sh, g, 0)
	runTicks(sh, 100*60) // 100 game seconds
	tr := sh.Trace(g.ID())
	if len(tr) != traceLen {
		t.Fatalf("trace has %d seconds, want %d", len(tr), traceLen)
	}
	if !tr[traceLen-1] {
		t.Error("the newest second should show them asleep")
	}
	if sh.Trace("P-999") != nil {
		t.Error("an unknown patient has no trace")
	}
}

func TestEpisodeTotal(t *testing.T) {
	sh := newTestShift(t, 2)
	arriveAll(sh)
	want := 0
	for _, g := range sh.Guests {
		want += len(g.Log)
	}
	if got := sh.EpisodeTotal(); got != want {
		t.Errorf("EpisodeTotal = %d, want %d", got, want)
	}
}

func TestParseTypeAndAdvance(t *testing.T) {
	cmds, err := ParseScript(strings.NewReader("type marcus trell\nadvance 30\n"))
	if err != nil {
		t.Fatal(err)
	}
	if cmds[0].Op != "type" || cmds[0].Text != "marcus trell" || cmds[1].N != 30 {
		t.Errorf("cmds = %+v", cmds)
	}
	if _, err := ParseScript(strings.NewReader("advance x")); err == nil {
		t.Error("advance x should fail")
	}
}

// Typing into the RECORDS field: letters add, Tab fills in the window guest.
func TestRecordsTyping(t *testing.T) {
	env := testEnv(t)
	env.Day = 2
	s := &computerScene{startTab: TabRecords}
	if err := s.Enter(env); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { env.StopShift() })
	env.Input.InjectText("oz1")
	if err := s.Update(); err != nil {
		t.Fatal(err)
	}
	env.Input.endFrame()
	if env.Query != "OZ" { // the digit is ignored
		t.Errorf("query = %q, want OZ", env.Query)
	}
}
