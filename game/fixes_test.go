package game

import (
	"strings"
	"testing"
)

// The "no room" alert stays up for as long as the sleeper lies in the
// hallway (M3 regression: it must survive the ticks after the click).
func TestNoRoomAlertStaysWhileTheSleeperIsInTheHallway(t *testing.T) {
	sh := newTestShift(t, 3) // two rooms
	sh.Day = 1               // day 3's memo changes attending (see memo_rules_test.go)
	for i, g := range sh.Guests[:3] {
		putToSleep(t, sh, g, float64(i))
	}
	for _, g := range sh.Guests[:3] {
		if _, err := sh.Dispatch(g); err != nil {
			t.Fatal(err)
		}
	}
	runTicks(sh, 120)
	if sh.Alert() == nil || sh.Alert().Guest != sh.Guests[2] {
		t.Fatal("the alert disappeared while the third sleeper is still in the hallway")
	}
	if len(sh.Waiting()) != 1 {
		t.Errorf("waiting = %d, want the hallway sleeper", len(sh.Waiting()))
	}
}

// WHERE on the ward feed says HALLWAY for every sleeper with no bed, and the
// banner names the Severe patient the call light blinks for.
func TestWardFeedWhereAndNurseCallBanner(t *testing.T) {
	sh := newTestShift(t, 1)
	cecily := sh.Guests[0]
	putToSleep(t, sh, cecily, 0)
	sh.Update()
	if got := sh.whereIs(cecily); got != "HALLWAY" {
		t.Errorf("an unattended sleeper is %q, want HALLWAY", got)
	}
	if got := sh.Banner(); got != "NURSE CALL: OLD CECILY IS DOWN" {
		t.Errorf("banner = %q", got)
	}
	sh.say("a message")
	if got := sh.Banner(); !strings.Contains(got, "A MESSAGE") {
		t.Errorf("a fresh message should win over the call: %q", got)
	}
	matty := sh.Guests[1]
	if got := sh.whereIs(matty); got != "WARD A" {
		t.Errorf("an awake patient is %q, want WARD A", got)
	}
}

func TestForceDarkShowsEveryStep(t *testing.T) {
	sh := newTestShift(t, 1)
	if sh.Darkness(sh.Guests[1]) != 0 {
		t.Error("nobody is dark by default")
	}
	for steps := 0; steps <= maxDarkness; steps++ {
		sh.ForceDark(steps)
		if got := sh.Darkness(sh.Guests[1]); got != steps {
			t.Errorf("ForceDark(%d): Darkness = %d", steps, got)
		}
	}
	sh.ForceDark(-1)
	if sh.Darkness(sh.Guests[1]) != 0 {
		t.Error("ForceDark(-1) should give control back to the game")
	}
	if _, err := ParseScript(strings.NewReader("dark 4")); err == nil {
		t.Error("dark 4 should fail")
	}
	cmds, err := ParseScript(strings.NewReader("dark 2"))
	if err != nil || cmds[0].N != 2 {
		t.Errorf("dark 2: %v %+v", err, cmds)
	}
	if o, err := ParseOptions([]string{"-agent"}); err != nil || o.Dark != -1 {
		t.Errorf("-dark defaults to %d (%v), want -1", o.Dark, err)
	}
}

// With no scores file the top-10 keeps the rows in memory for the run.
func TestTop10KeepsScoresInMemory(t *testing.T) {
	env := testEnv(t)
	env.Scores["day1"] = 450
	s := &top10Scene{}
	if err := s.Enter(env); err != nil {
		t.Fatal(err)
	}
	if !s.saved || len(env.MemScores) != 1 || env.MemScores[0].Score != 450 {
		t.Errorf("saved %v, memory %+v", s.saved, env.MemScores)
	}
	if len(s.rows) != 1 || s.rows[0].Score != 450 {
		t.Errorf("rows = %+v", s.rows)
	}
}
