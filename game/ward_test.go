package game

import (
	"errors"
	"testing"
	"time"

	"danmin/hospital"
	"danmin/simulation"
)

// With a huge speed-up an admitted patient falls asleep within
// milliseconds, so the test sees real events without waiting. Run it with
// -race: Drain and Snapshot are what the game calls while the goroutines work.
func TestWardDrainsEventsAndStops(t *testing.T) {
	w, err := NewWard(7, 100000, 3, false)
	if err != nil {
		t.Fatal(err)
	}
	w.Start()
	w.Start() // second call must do nothing
	if n := len(w.Snapshot().Patients); n != 0 {
		t.Fatalf("%d patients before anybody was admitted, want 0", n)
	}
	if err := w.Admit(hospital.NewPatient("P-001", "Old Cecily", 71, hospital.Severe)); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for len(w.Log()) < 5 && time.Now().Before(deadline) {
		w.Drain()
		w.Snapshot()
		time.Sleep(time.Millisecond)
	}
	if err := w.Stop(); err != nil {
		t.Fatalf("Stop returned %v", err)
	}
	w.Drain()
	if len(w.Log()) < 5 {
		t.Fatalf("only %d events arrived", len(w.Log()))
	}
	if w.Log()[0].Kind != simulation.FellAsleep {
		t.Errorf("first event = %v, want a patient falling asleep", w.Log()[0].Kind)
	}
	if err := w.Stop(); err != nil { // stopping twice is fine
		t.Errorf("second Stop returned %v", err)
	}
}

func TestAdmitBeforeStartFails(t *testing.T) {
	w, err := NewWard(7, 1, 3, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Admit(hospital.NewPatient("P-001", "X", 30, hospital.Mild)); err == nil {
		t.Error("Admit before Start should fail")
	}
}

// Staff are dispatched in round-robin order and a full ward answers with
// ErrNoRoomAvailable, which the game finds with errors.Is.
func TestWardAttendRoundRobinAndNoRoom(t *testing.T) {
	w, err := NewWard(7, 1, 1, false) // one room
	if err != nil {
		t.Fatal(err)
	}
	w.Start()
	defer w.Stop()
	for _, id := range []string{"P-001", "P-002", "P-003"} {
		if err := w.Admit(hospital.NewPatient(id, id, 30, hospital.Mild)); err != nil {
			t.Fatal(err)
		}
	}
	var who []string
	for _, id := range []string{"P-001", "P-002", "P-003"} {
		name, err := w.Attend(id, "reception")
		if err != nil {
			t.Fatal(err)
		}
		who = append(who, name)
	}
	if who[0] != "Dr. Barrow" || who[1] != "Nurse Mayu" || who[2] != "Gus" {
		t.Errorf("attenders = %v, want the staff in round-robin order", who)
	}
	if room, err := w.AssignRoom("P-001"); room != 101 || !errors.Is(err, hospital.ErrAlreadyInBed) {
		t.Errorf("P-001: room %d, err %v; want room 101 and ErrAlreadyInBed", room, err)
	}
	if _, err := w.AssignRoom("P-002"); !errors.Is(err, hospital.ErrNoRoomAvailable) {
		t.Errorf("P-002: err %v, want ErrNoRoomAvailable", err)
	}
	if _, err := w.Attend("P-404", "reception"); !errors.Is(err, hospital.ErrNotAdmitted) {
		t.Errorf("unknown patient: err %v, want ErrNotAdmitted", err)
	}
}

func TestSceneManagerStartsAndStopsShift(t *testing.T) {
	env := testEnv(t)
	m := registerScenes(env)

	if err := m.Switch("nope"); err == nil {
		t.Error("switching to an unknown scene should fail")
	}
	if err := m.Switch(SceneDayReception); err != nil {
		t.Fatal(err)
	}
	if env.Shift() == nil {
		t.Fatal("the reception should have started the shift")
	}
	if err := m.Switch(SceneDayComputer); err != nil {
		t.Fatal(err)
	}
	if env.Shift() == nil {
		t.Error("the computer shares the reception's shift: it must keep running")
	}
	if err := m.Switch(SceneDayTop10); err != nil {
		t.Fatal(err)
	}
	if env.Shift() != nil {
		t.Error("the top-10 should have stopped the shift")
	}
	if m.ID() != SceneDayTop10 {
		t.Errorf("scene = %q", m.ID())
	}
}

func TestSwitchToIsAppliedAfterUpdate(t *testing.T) {
	env := testEnv(t)
	m := registerScenes(env)
	if err := m.Switch(SceneTitle); err != nil {
		t.Fatal(err)
	}
	env.SwitchTo(SceneDayReception)
	if err := m.Update(); err != nil {
		t.Fatal(err)
	}
	defer env.StopShift()
	if m.ID() != SceneDayReception {
		t.Errorf("scene = %q, want the reception", m.ID())
	}
}

func registerScenes(env *Env) *SceneManager {
	m := NewSceneManager(env)
	m.Register(SceneTitle, func() Scene { return titleScene{} })
	m.Register(SceneDayReception, func() Scene { return &receptionScene{} })
	m.Register(SceneDayComputer, func() Scene { return &computerScene{} })
	m.Register(SceneDayTop10, func() Scene { return &top10Scene{} })
	return m
}

// testEnv is an Env with what the day scenes need to start: the real
// embedded sprites and an input.
func testEnv(t *testing.T) *Env {
	t.Helper()
	fsys, err := SpriteFS()
	if err != nil {
		t.Fatal(err)
	}
	return &Env{Seed: 1, Day: 1, Speed: 1, Clock: NewClock(9, 0), Input: NewInput(),
		Sprites: fsys, Scores: map[string]int{}, ClockRate: 1}
}
