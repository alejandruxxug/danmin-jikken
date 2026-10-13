package game

import (
	"context"
	"fmt"
	"time"

	"danmin/hospital"
	"danmin/simulation"
)

// Ward is the bridge between the hospital model and the game.
//
// The patients' goroutines (simulation.Run) change the hospital in the
// background. The game never reads the hospital while they do that, except
// through Hospital's own locked methods: Snapshot for the picture, and the
// events that Drain collects. Update calls Drain once per frame, so every
// change the game reacts to arrives on the game's own thread.
type Ward struct {
	hospital *hospital.Hospital
	patients []*hospital.Patient
	events   chan simulation.Event
	seed     int64
	speed    float64

	cancel   context.CancelFunc
	finished chan struct{} // closed when simulation.Run has returned
	runErr   error         // written before finished is closed, read after
	log      []simulation.Event
}

// NewWard builds the hospital for a day: staff, rooms and five patients.
// speed multiplies how fast the patients' timers run (1 = normal).
func NewWard(seed int64, speed float64) (*Ward, error) {
	if speed <= 0 {
		speed = 1
	}
	h := hospital.NewHospital("Halcyon Rest")
	hires := []error{
		h.HireDoctor(hospital.NewDoctor("D-01", "Dr. Barrow", 52, "sleep medicine")),
		h.HireDoctor(hospital.NewDoctor("D-02", "Dr. Sato", 38, "neurology")),
		h.HireStaff(hospital.NewOrderly("O-01", "Gus the orderly", 29)),
	}
	for _, n := range []int{101, 102, 103} {
		hires = append(hires, h.AddRoom(hospital.NewRoom(n, 1)))
	}
	patients := []*hospital.Patient{
		hospital.NewPatient("P-001", "Old Cecily", 71, hospital.Severe),
		hospital.NewPatient("P-002", "Matty", 12, hospital.Moderate),
		hospital.NewPatient("P-003", "Wren Calloway", 34, hospital.Mild),
		hospital.NewPatient("P-004", "Hank Shoals", 58, hospital.Severe),
		hospital.NewPatient("P-005", "Ivy Marsh", 26, hospital.Moderate),
	}
	for _, p := range patients {
		hires = append(hires, h.AdmitPatient(p))
	}
	for _, err := range hires {
		if err != nil {
			return nil, fmt.Errorf("building the ward: %w", err)
		}
	}
	return &Ward{
		hospital: h,
		patients: patients,
		events:   make(chan simulation.Event, 64),
		seed:     seed,
		speed:    speed,
	}, nil
}

// Start launches the simulation in the background. It stops when Stop is
// called. Calling Start twice does nothing.
func (w *Ward) Start() {
	if w.cancel != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel
	w.finished = make(chan struct{})
	cfg := simulation.Config{
		Seed:      w.seed,
		MaxAwake:  20 * time.Second,
		MaxAsleep: 10 * time.Second,
		Locations: []string{"reception", "waiting room", "hallway", "cafeteria"},
		Sleep:     w.sleep,
		Events:    w.events,
	}
	go func() {
		defer close(w.finished)
		w.runErr = simulation.Run(ctx, w.hospital, w.patients, cfg)
	}()
}

// Stop cancels the simulation and waits until every patient goroutine is
// gone. It returns the error simulation.Run reported, if any.
func (w *Ward) Stop() error {
	if w.cancel == nil {
		return nil
	}
	w.cancel()
	<-w.finished
	w.cancel = nil
	return w.runErr
}

// Drain collects the events that happened since the last call and keeps
// them in the log. It never blocks. Call it from Update.
func (w *Ward) Drain() []simulation.Event {
	var got []simulation.Event
	for {
		select {
		case ev := <-w.events:
			got = append(got, ev)
		default:
			w.log = append(w.log, got...)
			return got
		}
	}
}

// Log is every event Drain has returned so far.
func (w *Ward) Log() []simulation.Event { return w.log }

// Snapshot is a consistent copy of the hospital, safe to read at any time.
func (w *Ward) Snapshot() hospital.Snapshot { return w.hospital.Snapshot() }

// sleep is the simulation's timer, shortened by the game speed.
func (w *Ward) sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(time.Duration(float64(d) / w.speed))
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
