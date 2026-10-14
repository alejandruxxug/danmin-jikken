package game

import (
	"strings"
	"testing"

	"danmingame/night"
)

// Day 2's memo: Shoal patients go to Ward C.
func TestDay2ShoalPatientsGoToWardC(t *testing.T) {
	sh := newTestShift(t, 2)
	var shoal []string
	for _, g := range sh.Guests {
		if g.SendToWardC(2) {
			shoal = append(shoal, g.Name)
		}
	}
	if len(shoal) != 1 || shoal[0] != "Dot Harlan" {
		t.Fatalf("Ward C patients on day 2 = %v, want only Dot Harlan", shoal)
	}
	if sh.Guests[0].SendToWardC(1) || sh.Guests[0].SendToWardC(3) {
		t.Error("the rule is day 2 only")
	}
	dot := sh.Guests[3]
	putToSleep(t, sh, dot, 0)
	att, err := sh.Dispatch(dot)
	if err != nil {
		t.Fatal(err)
	}
	if att.NoRoom || sh.Alert() != nil {
		t.Error("Ward C has no beds on the map: no \"no room\" alert")
	}
	if want := pointsAttend + pointsWardC; sh.Score() != want {
		t.Errorf("score = %d, want %d (attend + Ward C)", sh.Score(), want)
	}
	if !strings.Contains(sh.Message(), "WARD C") {
		t.Errorf("message = %q, want it to say WARD C", sh.Message())
	}
}

// Day 3's memo: do not wake them.
func TestDay3WakingThemCostsPoints(t *testing.T) {
	sh := newTestShift(t, 3)
	sh.addScore(100)
	putToSleep(t, sh, sh.Guests[1], 0)
	if _, err := sh.Dispatch(sh.Guests[1]); err != nil {
		t.Fatal(err)
	}
	if sh.Score() != 100-pointsAttend {
		t.Errorf("score = %d after waking a sleeper, want %d", sh.Score(), 100-pointsAttend)
	}
	// The score never goes below zero.
	sh2 := newTestShift(t, 3)
	putToSleep(t, sh2, sh2.Guests[1], 0)
	if _, err := sh2.Dispatch(sh2.Guests[1]); err != nil {
		t.Fatal(err)
	}
	if sh2.Score() != 0 {
		t.Errorf("score = %d, want 0", sh2.Score())
	}
}

func TestDay3SleepersLeftAloneScoreAtTheEnd(t *testing.T) {
	sh := newTestShift(t, 3)
	putToSleep(t, sh, sh.Guests[1], 0)
	putToSleep(t, sh, sh.Guests[2], 0)
	sh.sleepers[sh.Guests[2].ID()].Attended = true // this one was woken
	sh.checkEnd(float64(sh.Rules.ShiftSeconds))    // time is up
	if !sh.Over() || sh.Score() != pointsLetSleep {
		t.Errorf("over %v score %d, want the shift over and %d for the one left asleep", sh.Over(), sh.Score(), pointsLetSleep)
	}
}

// Your own file on day 3 is the choice that decides the secret ending.
func TestStampingYourOwnFileSetsFlags(t *testing.T) {
	for _, tt := range []struct {
		stamp Stamp
		want  []string
	}{
		{StampDischarge, []string{"stamped_discharge", "stamped_self"}},
		{StampAdmit, []string{"stamped_self"}},
		{StampRefer, []string{}},
	} {
		sh := newTestShift(t, 3)
		sh.Flags = night.Flags{}
		last := len(sh.Guests) - 1
		if !sh.Guests[last].Self {
			t.Fatal("the last patient of day 3 must be you")
		}
		for _, g := range sh.Guests[:last] { // everyone before you has been served
			g.status = guestLeft
			sh.served++
		}
		runTicks(sh, 130*60) // you walk up to the window
		if err := sh.Stamp(tt.stamp); err != nil {
			t.Fatalf("%v: %v", tt.stamp, err)
		}
		got := sh.Flags.Names()
		if strings.Join(got, ",") != strings.Join(tt.want, ",") {
			t.Errorf("%v: flags %v, want %v", tt.stamp, got, tt.want)
		}
	}
}

func TestSelfNameIsTheUsername(t *testing.T) {
	sh := newTestShift(t, 3)
	if got, want := sh.Guests[len(sh.Guests)-1].Name, strings.ToUpper(Username()); got != want {
		t.Errorf("your name at the window = %q, want %q", got, want)
	}
}
