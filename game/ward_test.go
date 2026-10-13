package game

import (
	"testing"
	"time"

	"danmin/simulation"
)

// With a huge speed-up the patients fall asleep within milliseconds, so the
// test sees real events without waiting. Run it with -race: Drain and
// Snapshot are what the game calls while the goroutines work.
func TestWardDrainsEventsAndStops(t *testing.T) {
	w, err := NewWard(7, 100000)
	if err != nil {
		t.Fatal(err)
	}
	w.Start()
	w.Start() // second call must do nothing

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

func TestSceneManagerStartsAndStopsWard(t *testing.T) {
	env := &Env{Seed: 1, Speed: 1, Clock: NewClock(9, 0)}
	m := NewSceneManager(env)
	m.Register(SceneTitle, func() Scene { return titleScene{} })
	m.Register(SceneDayReception, func() Scene { return &receptionScene{} })

	if err := m.Switch("nope"); err == nil {
		t.Error("switching to an unknown scene should fail")
	}
	if err := m.Switch(SceneDayReception); err != nil {
		t.Fatal(err)
	}
	if env.Ward() == nil {
		t.Fatal("the reception should have started the ward")
	}
	if err := m.Switch(SceneTitle); err != nil {
		t.Fatal(err)
	}
	if env.Ward() != nil {
		t.Error("leaving the reception should have stopped the ward")
	}
	if m.ID() != SceneTitle {
		t.Errorf("scene = %q", m.ID())
	}
}

func TestSwitchToIsAppliedAfterUpdate(t *testing.T) {
	env := &Env{Seed: 1, Speed: 1, Clock: NewClock(9, 0)}
	m := NewSceneManager(env)
	m.Register(SceneTitle, func() Scene { return titleScene{} })
	m.Register(SceneDayReception, func() Scene { return &receptionScene{} })
	if err := m.Switch(SceneTitle); err != nil {
		t.Fatal(err)
	}
	env.SwitchTo(SceneDayReception)
	if err := m.Update(); err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if m.ID() != SceneDayReception {
		t.Errorf("scene = %q, want the reception", m.ID())
	}
}
