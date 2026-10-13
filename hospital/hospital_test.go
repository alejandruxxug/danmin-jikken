package hospital

import (
	"errors"
	"sync"
	"testing"
	"time"
)

// fixedNow pins the record clock for the duration of a test.
func fixedNow(t *testing.T, at time.Time) {
	t.Helper()
	old := now
	now = func() time.Time { return at }
	t.Cleanup(func() { now = old })
}

// newTestHospital builds a hospital with one doctor, one orderly, the given
// number of single-bed rooms and the given patients, all admitted.
func newTestHospital(t *testing.T, rooms int, patients ...*Patient) *Hospital {
	t.Helper()
	h := NewHospital("Test General")
	if err := h.HireDoctor(NewDoctor("D-1", "Dr. Test", 40, "sleep")); err != nil {
		t.Fatal(err)
	}
	if err := h.HireStaff(NewOrderly("O-1", "Orderly Test", 30)); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < rooms; i++ {
		if err := h.AddRoom(NewRoom(101+i, 1)); err != nil {
			t.Fatal(err)
		}
	}
	for _, p := range patients {
		if err := h.AdmitPatient(p); err != nil {
			t.Fatal(err)
		}
	}
	return h
}

// Required test 1: a sleeping patient gets the free room.
func TestAssignRoomWithFreeRoom(t *testing.T) {
	p := NewPatient("P-1", "Ana", 30, Mild)
	h := newTestHospital(t, 1, p)
	if err := p.SufferSleepAttack("cafeteria"); err != nil {
		t.Fatal(err)
	}

	room, err := h.AssignRoom(p)
	if err != nil {
		t.Fatalf("AssignRoom: unexpected error %v", err)
	}
	if room.Number() != 101 {
		t.Errorf("room = %d, want 101", room.Number())
	}
	if p.State() != AsleepInBed {
		t.Errorf("patient state = %v, want %v", p.State(), AsleepInBed)
	}
	if room.State() != Occupied {
		t.Errorf("room state = %v, want %v", room.State(), Occupied)
	}
}

// Required test 2: with every room taken, AssignRoom fails and the patient
// stays in the hallway.
func TestAssignRoomWithoutFreeRoom(t *testing.T) {
	first := NewPatient("P-1", "Ana", 30, Mild)
	second := NewPatient("P-2", "Beto", 45, Severe)
	h := newTestHospital(t, 1, first, second)
	if err := h.RegisterEpisode(first, "cafeteria"); err != nil {
		t.Fatal(err)
	}
	if err := second.SufferSleepAttack("hallway 2"); err != nil {
		t.Fatal(err)
	}

	room, err := h.AssignRoom(second)
	if !errors.Is(err, ErrNoRoomAvailable) {
		t.Fatalf("err = %v, want ErrNoRoomAvailable", err)
	}
	if room != nil {
		t.Errorf("room = %v, want nil", room)
	}
	if second.State() != AsleepInHallway {
		t.Errorf("state = %v, want %v", second.State(), AsleepInHallway)
	}
}

// Required test 3 (a query): PatientsInHallway finds only hallway sleepers.
func TestPatientsInHallway(t *testing.T) {
	inBed := NewPatient("P-1", "Ana", 30, Mild)
	inHallway := NewPatient("P-2", "Beto", 45, Severe)
	awake := NewPatient("P-3", "Caro", 22, Moderate)
	h := newTestHospital(t, 1, inBed, inHallway, awake)
	for _, p := range []*Patient{inBed, inHallway} {
		if err := h.RegisterEpisode(p, "lobby"); err != nil {
			t.Fatal(err)
		}
	}

	got := h.PatientsInHallway()
	if len(got) != 1 || got[0] != inHallway {
		t.Fatalf("PatientsInHallway = %v, want only P-2", got)
	}
}

func TestWakeUpFreesTheRoom(t *testing.T) {
	first := NewPatient("P-1", "Ana", 30, Mild)
	second := NewPatient("P-2", "Beto", 45, Severe)
	h := newTestHospital(t, 1, first, second)
	_ = h.RegisterEpisode(first, "lobby")
	_ = h.RegisterEpisode(second, "hallway 2")

	if err := h.WakePatient(first); err != nil {
		t.Fatal(err)
	}
	room, err := h.AssignRoom(second)
	if err != nil {
		t.Fatalf("AssignRoom after wake-up: %v", err)
	}
	if room.Number() != 101 || first.Room() != nil {
		t.Errorf("room %d, first.Room() = %v; want 101 and nil", room.Number(), first.Room())
	}
}

func TestMyEpisodesOnlyListsThatDoctor(t *testing.T) {
	a := NewPatient("P-1", "Ana", 30, Mild)
	b := NewPatient("P-2", "Beto", 45, Severe)
	h := newTestHospital(t, 2, a, b)
	_ = h.RegisterEpisode(a, "lobby")     // round robin: the doctor
	_ = h.RegisterEpisode(b, "radiology") // then the orderly

	doc := h.Doctors()[0]
	eps := doc.MyEpisodes()
	if len(eps) != 1 || eps[0].Patient() != a {
		t.Fatalf("MyEpisodes = %d records, want 1 for P-1", len(eps))
	}
}

func TestSevereReportCountsOnlyToday(t *testing.T) {
	today := time.Date(2026, 10, 13, 14, 15, 0, 0, time.Local)
	severe := NewPatient("P-1", "Ana", 30, Severe)
	quiet := NewPatient("P-2", "Beto", 45, Severe)
	mild := NewPatient("P-3", "Caro", 22, Mild)
	h := newTestHospital(t, 3, severe, quiet, mild)

	fixedNow(t, today.AddDate(0, 0, -1))
	_ = h.RegisterEpisode(severe, "yesterday")
	_ = h.WakePatient(severe)
	fixedNow(t, today)
	_ = h.RegisterEpisode(severe, "lobby")
	_ = h.RegisterEpisode(mild, "lobby")

	report := h.SevereReport()
	if len(report) != 2 || report[severe] != 1 || report[quiet] != 0 {
		t.Fatalf("SevereReport = %v, want P-1:1 and P-2:0 only", report)
	}
}

func TestStaffIsDispatchedPolymorphically(t *testing.T) {
	a := NewPatient("P-1", "Ana", 30, Mild)
	b := NewPatient("P-2", "Beto", 45, Severe)
	h := newTestHospital(t, 2, a, b)
	_ = h.RegisterEpisode(a, "lobby")
	_ = h.RegisterEpisode(b, "lobby")

	hist := h.History()
	if _, ok := hist[0].Attender().(*Doctor); !ok {
		t.Errorf("first attender is %T, want *Doctor", hist[0].Attender())
	}
	if _, ok := hist[1].Attender().(*Orderly); !ok {
		t.Errorf("second attender is %T, want *Orderly", hist[1].Attender())
	}
}

// Many goroutines race for few rooms: exactly len(rooms) must win.
// Run with go test -race.
func TestConcurrentAssignRoom(t *testing.T) {
	const sleepers, rooms = 10, 3
	var ps []*Patient
	for i := 0; i < sleepers; i++ {
		ps = append(ps, NewPatient(string(rune('A'+i)), "P", 30, Mild))
	}
	h := newTestHospital(t, rooms, ps...)
	for _, p := range ps {
		if err := p.SufferSleepAttack("hallway"); err != nil {
			t.Fatal(err)
		}
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	won := 0
	for _, p := range ps {
		p := p
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := h.AssignRoom(p); err == nil {
				mu.Lock()
				won++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if won != rooms {
		t.Fatalf("%d patients got a room, want %d", won, rooms)
	}
}
