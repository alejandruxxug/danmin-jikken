package simulation

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"testing"
	"time"

	"danmin/hospital"
)

// instant is a Sleep that never waits, so tests run in microseconds. It
// yields so the goroutines still interleave.
func instant(ctx context.Context, d time.Duration) error {
	runtime.Gosched()
	return ctx.Err()
}

func newWard(t *testing.T, rooms, patients int, staff ...hospital.Attender) (*hospital.Hospital, []*hospital.Patient) {
	t.Helper()
	h := hospital.NewHospital("Test General")
	for _, s := range staff {
		if err := h.HireStaff(s); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < rooms; i++ {
		if err := h.AddRoom(hospital.NewRoom(101+i, 1)); err != nil {
			t.Fatal(err)
		}
	}
	var ps []*hospital.Patient
	for i := 0; i < patients; i++ {
		p := hospital.NewPatient(fmt.Sprintf("P-%02d", i+1), "Patient", 30, hospital.Severe)
		if err := h.AdmitPatient(p); err != nil {
			t.Fatal(err)
		}
		ps = append(ps, p)
	}
	return h, ps
}

func threeStaff() []hospital.Attender {
	return []hospital.Attender{
		hospital.NewDoctor("D-1", "Dr. Test", 40, "sleep"),
		hospital.NewOrderly("O-1", "Orderly Test", 30),
		hospital.NewNurse("N-1", "Nurse Test", 35, "east ward"),
	}
}

// Many patients, few rooms: every attack is recorded, nobody is left asleep
// and no room is left occupied. Run with go test -race.
func TestRunEveryAttackIsRecorded(t *testing.T) {
	const patients, attacks = 8, 5
	h, ps := newWard(t, 3, patients, threeStaff()...)

	err := Run(context.Background(), h, ps, Config{Seed: 1, Attacks: attacks, Sleep: instant,
		Locations: []string{"cafeteria", "lobby", "radiology"}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if got := len(h.History()); got != patients*attacks {
		t.Errorf("history has %d episodes, want %d", got, patients*attacks)
	}
	snap := h.Snapshot()
	for _, p := range snap.Patients {
		if p.State != hospital.Awake {
			t.Errorf("%s is %v at the end, want awake", p.ID, p.State)
		}
	}
	for _, r := range snap.Rooms {
		if len(r.Occupants) != 0 || r.State != hospital.Available {
			t.Errorf("room %d not empty at the end: %v", r.Number, r)
		}
	}
}

// The nurse is dispatched like any other attender.
func TestRunDispatchesAllThreeKinds(t *testing.T) {
	h, ps := newWard(t, 2, 4, threeStaff()...)
	if err := Run(context.Background(), h, ps, Config{Attacks: 3, Sleep: instant}); err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, rec := range h.History() {
		seen[fmt.Sprintf("%T", rec.Attender())] = true
	}
	if len(seen) != 3 {
		t.Errorf("attender types used = %v, want Doctor, Orderly and Nurse", seen)
	}
}

// The same seed gives the same locations for one patient, in the same order.
func TestRunSameSeedSameLocations(t *testing.T) {
	locations := func(seed int64) []string {
		h, ps := newWard(t, 1, 1, threeStaff()...)
		events := make(chan Event, 100)
		cfg := Config{Seed: seed, Attacks: 6, Sleep: instant, Events: events,
			Locations: []string{"a", "b", "c", "d"}}
		if err := Run(context.Background(), h, ps, cfg); err != nil {
			t.Fatal(err)
		}
		close(events)
		var out []string
		for ev := range events {
			if ev.Kind == FellAsleep {
				out = append(out, ev.Location)
			}
		}
		return out
	}
	a, b := locations(42), locations(42)
	if len(a) != 6 || fmt.Sprint(a) != fmt.Sprint(b) {
		t.Errorf("seed 42 gave %v then %v, want the same 6 locations", a, b)
	}
}

// Cancelling stops patients that would otherwise run forever, with no error.
func TestRunStopsWhenCancelled(t *testing.T) {
	h, ps := newWard(t, 1, 3, threeStaff()...)
	ctx, cancel := context.WithCancel(context.Background())
	events := make(chan Event) // nobody reads it: send must not block forever
	done := make(chan error)
	go func() { done <- Run(ctx, h, ps, Config{MaxAwake: time.Hour, Events: events}) }()

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run after cancel = %v, want nil", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not stop after cancel")
	}
}

// A failure comes back wrapped, so errors.Is can still find the cause.
func TestRunReturnsWrappedErrors(t *testing.T) {
	h, ps := newWard(t, 1, 2) // no staff hired
	err := Run(context.Background(), h, ps, Config{Attacks: 1, Sleep: instant})
	if !errors.Is(err, hospital.ErrNoStaff) {
		t.Fatalf("err = %v, want it to wrap ErrNoStaff", err)
	}
}

// Reading the hospital through Snapshot while patients change it is safe.
func TestSnapshotWhileRunning(t *testing.T) {
	h, ps := newWard(t, 2, 6, threeStaff()...)
	done := make(chan error)
	go func() { done <- Run(context.Background(), h, ps, Config{Attacks: 20, Sleep: instant}) }()
	for {
		select {
		case err := <-done:
			if err != nil {
				t.Fatal(err)
			}
			return
		default:
			s := h.Snapshot()
			if len(s.Patients) != 6 {
				t.Fatalf("snapshot has %d patients, want 6", len(s.Patients))
			}
			for _, d := range h.Doctors() {
				_ = d.MyEpisodes()
			}
		}
	}
}
