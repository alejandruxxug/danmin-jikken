package game

import (
	"errors"
	"path/filepath"
	"testing"
)

func newTestShift(t *testing.T, day int) *Shift {
	t.Helper()
	sh, err := NewShift(day, 7, 1, false, map[string]int{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sh.Stop() })
	return sh
}

// runTicks plays n ticks of the shift.
func runTicks(sh *Shift, n int) {
	for range n {
		sh.Update()
	}
}

// Patients are not in the hospital until the player stamps ADMIT, and only
// a right stamp scores.
func TestStampAdmitsAndScores(t *testing.T) {
	sh := newTestShift(t, 1)
	if err := sh.Stamp(StampAdmit); !errors.Is(err, ErrNoOneAtWindow) {
		t.Fatalf("stamp with an empty window: %v, want ErrNoOneAtWindow", err)
	}
	runTicks(sh, 200) // Old Cecily walks up to the window
	if n := len(sh.Snapshot().Patients); n != 0 {
		t.Fatalf("%d patients admitted before the stamp, want 0", n)
	}
	if err := sh.Stamp(StampAdmit); err != nil {
		t.Fatal(err)
	}
	sh.Update()
	if n := len(sh.Snapshot().Patients); n != 1 {
		t.Errorf("%d patients in the hospital after ADMIT, want 1", n)
	}
	if sh.Score() != pointsRightStamp || sh.Served() != 1 {
		t.Errorf("score %d served %d, want %d and 1", sh.Score(), sh.Served(), pointsRightStamp)
	}
}

func TestWrongStampDoesNotScore(t *testing.T) {
	sh := newTestShift(t, 1)
	runTicks(sh, 200)
	if err := sh.Stamp(StampRefer); err != nil { // Cecily's referral is valid
		t.Fatal(err)
	}
	sh.Update()
	if sh.Score() != 0 {
		t.Errorf("score = %d after a wrong stamp, want 0", sh.Score())
	}
	if n := len(sh.Snapshot().Patients); n != 0 {
		t.Errorf("a referred patient must not enter the hospital (%d there)", n)
	}
	if st := sh.Stamps(); len(st) != 1 || st[0].Right {
		t.Errorf("stamps = %+v", st)
	}
	if sh.Message() == "" {
		t.Error("a wrong stamp should leave a citation message")
	}
}

func TestRightStampFromData(t *testing.T) {
	cases := []struct {
		v    Visitor
		want Stamp
	}{
		{Visitor{ReferralOK: true}, StampAdmit},
		{Visitor{ReferralOK: false}, StampRefer},
		{Visitor{ReferralOK: true, Returning: true}, StampDischarge},
		{Visitor{ReferralOK: false, Returning: true}, StampRefer},
	}
	for _, c := range cases {
		if got := c.v.RightStamp(); got != c.want {
			t.Errorf("%+v: %v, want %v", c.v, got, c.want)
		}
	}
}

func TestDaysTable(t *testing.T) {
	want := []struct{ patients, rooms, secs int }{{5, 3, 180}, {6, 3, 180}, {4, 2, 150}}
	for i, w := range want {
		r, err := RulesForDay(i + 1)
		if err != nil {
			t.Fatal(err)
		}
		if len(r.Visitors) != w.patients || r.Rooms != w.rooms || r.ShiftSeconds != w.secs {
			t.Errorf("day %d = %d patients, %d rooms, %ds; want %+v", i+1, len(r.Visitors), r.Rooms, r.ShiftSeconds, w)
		}
	}
	for _, bad := range []int{0, 4} {
		if _, err := RulesForDay(bad); err == nil {
			t.Errorf("day %d should fail", bad)
		}
	}
}

// A click on a sleeper dispatches the next attender and scores; when every
// bed is taken the no-room alert opens and the patient lies in the hallway.
func TestDispatchAndNoRoom(t *testing.T) {
	sh := newTestShift(t, 3) // two rooms
	var guests []*Guest
	for _, g := range sh.Guests[:3] {
		if err := sh.ward.Admit(g.patient); err != nil {
			t.Fatal(err)
		}
		g.status = guestAdmitted
		sh.sleepers[g.ID()] = &Sleeper{Location: "waiting room"} // they nodded off
		guests = append(guests, g)
	}
	sh.Update()
	if sh.LookOf(guests[0]) != LookAsleep {
		t.Fatalf("look = %v, want asleep before anyone attends", sh.LookOf(guests[0]))
	}
	want := []string{"Dr. Barrow", "Nurse Mayu", "Gus"}
	for i, g := range guests {
		att, err := sh.Dispatch(g)
		if err != nil {
			t.Fatal(err)
		}
		if att.Attender != want[i] {
			t.Errorf("guest %d attended by %q, want %q", i, att.Attender, want[i])
		}
		if i < 2 && (att.NoRoom || att.Room != 101+i) {
			t.Errorf("guest %d: %+v, want room %d", i, att, 101+i)
		}
		sh.Update()
	}
	if sh.Score() != 3*pointsAttend {
		t.Errorf("score = %d, want %d", sh.Score(), 3*pointsAttend)
	}
	if sh.Alert() == nil || sh.Alert().Guest != guests[2] {
		t.Fatal("the third sleeper should have opened the no-room alert")
	}
	if sh.LookOf(guests[2]) != LookHallway || sh.LookOf(guests[0]) != LookHidden {
		t.Errorf("looks: hallway guest %v, bed guest %v", sh.LookOf(guests[2]), sh.LookOf(guests[0]))
	}
	if w := sh.Waiting(); len(w) != 1 || w[0] != guests[2] {
		t.Errorf("waiting = %v, want only the hallway sleeper", w)
	}
	if !sh.Busy("Dr. Barrow") {
		t.Error("an attender who just attended should be busy")
	}

	// A second click on the hallway sleeper tries for a bed again, and does
	// not pay for staff twice.
	att, err := sh.Dispatch(guests[2])
	if err != nil || !att.NoRoom || att.Attender != "" {
		t.Errorf("retry = %+v, %v; want no room and no new attender", att, err)
	}
	if sh.Score() != 3*pointsAttend {
		t.Errorf("score changed to %d on a retry", sh.Score())
	}

	// A bed frees up when the first sleeper wakes: CALL STAFF gives it away.
	if err := sh.ward.hospital.WakePatient(guests[0].patient); err != nil {
		t.Fatal(err)
	}
	delete(sh.sleepers, guests[0].ID())
	att, err = sh.CallStaff()
	if err != nil || att.NoRoom || att.Room != 101 {
		t.Errorf("CallStaff = %+v, %v; want room 101", att, err)
	}
	sh.Update()
	if sh.Alert() != nil {
		t.Error("the alert should close once the sleeper has a bed")
	}
}

func TestDispatchAwakeGuestDoesNothing(t *testing.T) {
	sh := newTestShift(t, 1)
	if _, err := sh.Dispatch(sh.Guests[0]); !errors.Is(err, ErrNothingToAttend) {
		t.Errorf("err = %v, want ErrNothingToAttend", err)
	}
	if _, err := sh.CallStaff(); !errors.Is(err, ErrNothingToAttend) {
		t.Errorf("CallStaff err = %v, want ErrNothingToAttend", err)
	}
}

func TestShiftEndsWhenTimeIsUp(t *testing.T) {
	sh := newTestShift(t, 3)
	sh.ticks = sh.Rules.ShiftSeconds*60 - 2
	sh.Update()
	if sh.Over() {
		t.Fatal("over one tick early")
	}
	sh.Update()
	sh.Update()
	if !sh.Over() {
		t.Error("the shift should be over after ShiftSeconds")
	}
}

// Ending a shift saves today's score and shows the top-10, with a ghost row
// from an earlier run.
func TestTop10SavesTheScore(t *testing.T) {
	env := testEnv(t)
	env.ScoresPath = filepath.Join(t.TempDir(), ScoresFile)
	if err := AppendScore(env.ScoresPath, ScoreRow{Name: "alejo", Score: 500, Time: "10:00"}); err != nil {
		t.Fatal(err)
	}
	if err := EnsureGhost(env.ScoresPath, "alejo"); err != nil {
		t.Fatal(err)
	}
	env.Scores["day1"] = 700
	s := &top10Scene{}
	if err := s.Enter(env); err != nil {
		t.Fatal(err)
	}
	if !s.saved || !s.newRecord {
		t.Errorf("saved %v, newRecord %v; want both", s.saved, s.newRecord)
	}
	rows, _ := LoadScores(env.ScoresPath)
	if len(rows) != 3 || rows[2].Score != 700 {
		t.Errorf("file rows = %+v", rows)
	}
	if len(s.rows) != 3 || s.rows[0].Score != 700 || !s.rows[2].Ghost {
		t.Errorf("ranked = %+v, want 700 first and the ghost last", s.rows)
	}
}
