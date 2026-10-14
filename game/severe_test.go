package game

import (
	"testing"

	"danmin/hospital"
)

// putToSleep admits a guest and makes them fall asleep as the patient
// goroutine would announce it, at the given game second.
func putToSleep(t *testing.T, sh *Shift, g *Guest, at float64) {
	t.Helper()
	if err := sh.ward.Admit(g.patient); err != nil {
		t.Fatal(err)
	}
	g.status = guestAdmitted
	sh.sleepers[g.ID()] = &Sleeper{Location: "waiting room", Since: at}
}

func TestSevereBonusOnlyWhenQuick(t *testing.T) {
	cases := []struct {
		name  string
		guest int  // index in day 1: 0 Cecily (Severe), 1 Matty (Moderate)
		wait  int  // ticks (60 = 1 s) the sleeper waits
		bonus bool // is the bonus expected
	}{
		{"Severe, quick", 0, 5 * 60, true},
		{"Severe, just inside the window", 0, severeQuickSecs * 60, true},
		{"Severe, too slow", 0, (severeQuickSecs + 2) * 60, false},
		{"Moderate, quick", 1, 60, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sh := newTestShift(t, 1)
			g := sh.Guests[c.guest]
			putToSleep(t, sh, g, 0)
			runTicks(sh, c.wait)
			att, err := sh.Dispatch(g)
			if err != nil {
				t.Fatal(err)
			}
			want := pointsAttend
			if c.bonus {
				want += pointsSevereQuick
			}
			if (att.Bonus > 0) != c.bonus || sh.Score() != want {
				t.Errorf("bonus %d, score %d; want bonus %v, score %d", att.Bonus, sh.Score(), c.bonus, want)
			}
		})
	}
}

func TestSevereCallLightWhileUnattended(t *testing.T) {
	sh := newTestShift(t, 1)
	if sh.SevereCalling() {
		t.Fatal("the light is on with nobody asleep")
	}
	putToSleep(t, sh, sh.Guests[1], 0) // Moderate: no light
	if sh.SevereCalling() {
		t.Error("a Moderate sleeper must not light the call light")
	}
	putToSleep(t, sh, sh.Guests[0], 0) // Severe
	if !sh.SevereCalling() {
		t.Error("a Severe sleeper with nobody attending should light it")
	}
	if _, err := sh.Dispatch(sh.Guests[0]); err != nil {
		t.Fatal(err)
	}
	if sh.SevereCalling() {
		t.Error("the light should go off once staff attended")
	}
}

// Every episode today moves a Severe patient one palette step darker, up to
// the darkest fill; other patients never change.
func TestSevereDarkensPerEpisode(t *testing.T) {
	sh := newTestShift(t, 1)
	cecily, matty := sh.Guests[0], sh.Guests[1]
	putToSleep(t, sh, cecily, 0)
	putToSleep(t, sh, matty, 0)
	for episode := 1; episode <= 5; episode++ {
		if err := sh.ward.hospital.RegisterEpisode(cecily.patient, "hallway"); err != nil {
			t.Fatal(err)
		}
		if err := sh.ward.hospital.RegisterEpisode(matty.patient, "hallway"); err != nil {
			t.Fatal(err)
		}
		sh.Update()
		if want := min(episode, maxDarkness); sh.Darkness(cecily) != want {
			t.Errorf("after %d episodes Severe darkness = %d, want %d", episode, sh.Darkness(cecily), want)
		}
		if sh.Darkness(matty) != 0 {
			t.Errorf("a Moderate patient was darkened (%d)", sh.Darkness(matty))
		}
		if sh.EpisodesOf(matty) != episode {
			t.Errorf("episodes of Matty = %d, want %d", sh.EpisodesOf(matty), episode)
		}
		// Free the bed again so the next episode can be recorded.
		_ = sh.ward.hospital.WakePatient(cecily.patient)
		_ = sh.ward.hospital.WakePatient(matty.patient)
	}
}

func TestPipsPerLevel(t *testing.T) {
	for l, want := range map[hospital.NarcolepsyLevel]int{hospital.Mild: 1, hospital.Moderate: 2, hospital.Severe: 3} {
		if got := Pips(l); got != want {
			t.Errorf("Pips(%v) = %d, want %d", l, got, want)
		}
	}
}
