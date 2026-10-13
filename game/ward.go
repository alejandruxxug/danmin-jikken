package game

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"danmin/hospital"
	"danmin/simulation"
)

// Ward is the bridge between the hospital model and the game.
//
// A patient only enters the hospital when the player stamps ADMIT (Admit),
// and only then gets a goroutine (simulation.Run, one call per patient).
// The goroutines never touch the game: they send events, and Drain
// collects them once per frame inside Update. The picture comes from
// Snapshot, which is safe to call at any time.
//
// Who attends a sleeper: the goroutine only announces that the patient
// fell asleep (simulation.Config.ManualAttend). The player's click calls
// Attend, which is Hospital.RegisterEpisode: the hospital gives the
// patient the first free room and the next attender in line (round robin)
// records the episode. Both sides go through the hospital's mutex, so the
// click and a patient waking up at the same moment cannot corrupt it.
type Ward struct {
	hospital *hospital.Hospital
	byID     map[string]*hospital.Patient // admitted patients, for the clicks
	events   chan simulation.Event
	seed     int64
	speed    float64
	sleepy   bool // test mode: patients fall asleep fast and sleep long

	ctx      context.Context
	cancel   context.CancelFunc
	running  sync.WaitGroup // one per patient goroutine
	runErrs  chan error     // what simulation.Run returned, one slot per patient
	admitted int
	wakeErr  error // a wake-up that failed in Drain
	log      []simulation.Event
}

// maxVisitors bounds the error channel; no day has more patients.
const maxVisitors = 16

// NewWard builds an empty hospital for a day: the staff and the rooms.
// speed multiplies how fast the patients' timers run (1 = normal).
func NewWard(seed int64, speed float64, rooms int, sleepy bool) (*Ward, error) {
	if speed <= 0 {
		speed = 1
	}
	h := hospital.NewHospital("Halcyon Rest")
	// The order is the round-robin order the model dispatches in.
	hires := []error{
		h.HireDoctor(hospital.NewDoctor("D-01", "Dr. Barrow", 52, "sleep medicine")),
		h.HireStaff(hospital.NewNurse("N-01", "Nurse Mayu", 24, "Ward A")),
		h.HireStaff(hospital.NewOrderly("O-01", "Gus", 29)),
	}
	for n := 101; n < 101+rooms; n++ {
		hires = append(hires, h.AddRoom(hospital.NewRoom(n, 1)))
	}
	for _, err := range hires {
		if err != nil {
			return nil, fmt.Errorf("building the ward: %w", err)
		}
	}
	return &Ward{
		hospital: h,
		byID:     make(map[string]*hospital.Patient),
		events:   make(chan simulation.Event, 64),
		runErrs:  make(chan error, maxVisitors),
		seed:     seed,
		speed:    speed,
		sleepy:   sleepy,
	}, nil
}

// Start gets the ward ready to take patients. Calling it twice does nothing.
func (w *Ward) Start() {
	if w.cancel != nil {
		return
	}
	w.ctx, w.cancel = context.WithCancel(context.Background())
}

// Admit registers the patient in the hospital and starts their goroutine.
func (w *Ward) Admit(p *hospital.Patient) error {
	if w.cancel == nil {
		return errors.New("ward: Admit before Start")
	}
	if err := w.hospital.AdmitPatient(p); err != nil {
		return err
	}
	w.byID[p.ID()] = p
	cfg := simulation.Config{
		// Each patient gets their own seed, so they do not nod off in step.
		Seed:         w.seed + int64(w.admitted),
		MaxAwake:     25 * time.Second,
		MaxAsleep:    30 * time.Second,
		Locations:    []string{"reception", "waiting room", "hallway", "cafeteria"},
		Sleep:        w.sleep,
		Events:       w.events,
		ManualAttend: true,
	}
	if w.sleepy {
		cfg.MaxAwake, cfg.MaxAsleep = time.Second, 120*time.Second
	}
	w.admitted++
	w.running.Add(1)
	go func() {
		defer w.running.Done()
		if err := simulation.Run(w.ctx, w.hospital, []*hospital.Patient{p}, cfg); err != nil {
			w.runErrs <- err
		}
	}()
	return nil
}

// Stop cancels the patients' goroutines and waits until every one is
// gone. It returns what they reported, if anything went wrong.
func (w *Ward) Stop() error {
	if w.cancel == nil {
		return nil
	}
	w.cancel()
	w.running.Wait()
	w.cancel = nil
	close(w.runErrs)
	errs := []error{w.wakeErr}
	for err := range w.runErrs {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// Drain collects the events that happened since the last call and keeps
// them in the log. It never blocks. Call it from Update.
func (w *Ward) Drain() []simulation.Event {
	var got []simulation.Event
	for {
		select {
		case ev := <-w.events:
			got = append(got, ev)
			if ev.Kind == simulation.WokeUp {
				w.getUp(ev.PatientID)
			}
		default:
			w.log = append(w.log, got...)
			return got
		}
	}
}

// getUp makes sure a patient who woke up is really up. The goroutine woke
// them before it sent the event, so this normally finds nothing to do. But
// a click can attend a patient in the instant between the goroutine's wake
// and the game hearing about it; then the patient would sleep in a bed with
// nobody left to wake them, and this call gets them out.
func (w *Ward) getUp(id string) {
	p := w.byID[id]
	if p == nil {
		return
	}
	if err := w.hospital.WakePatient(p); err != nil && !errors.Is(err, hospital.ErrNotAsleep) {
		w.wakeErr = errors.Join(w.wakeErr, err)
	}
}

// Log is every event Drain has returned so far.
func (w *Ward) Log() []simulation.Event { return w.log }

// Snapshot is a consistent copy of the hospital, safe to read at any time.
func (w *Ward) Snapshot() hospital.Snapshot { return w.hospital.Snapshot() }

// StaffNames lists who can be dispatched, in round-robin order.
func (w *Ward) StaffNames() []string {
	var names []string
	for _, a := range w.hospital.Staff() {
		names = append(names, a.Name())
	}
	return names
}

// Attend dispatches staff to a patient who fell asleep at location: the
// hospital records the episode, gives them a room if one is free, and the
// next attender in line attends. It returns that attender's name. A patient
// left without a room is asleep in the hallway (see AssignRoom).
func (w *Ward) Attend(id, location string) (string, error) {
	p := w.byID[id]
	if p == nil {
		return "", fmt.Errorf("patient %s: %w", id, hospital.ErrNotAdmitted)
	}
	if err := w.hospital.RegisterEpisode(p, location); err != nil {
		return "", err
	}
	// Only the game's thread calls RegisterEpisode, so the newest record is
	// the one just made.
	history := w.hospital.History()
	return history[len(history)-1].Attender().Name(), nil
}

// AssignRoom tries to put an attended patient in a bed and returns the room
// number. A patient who already has one gets it back with
// hospital.ErrAlreadyInBed; with no room free the error wraps
// hospital.ErrNoRoomAvailable and the patient stays in the hallway.
func (w *Ward) AssignRoom(id string) (int, error) {
	p := w.byID[id]
	if p == nil {
		return 0, fmt.Errorf("patient %s: %w", id, hospital.ErrNotAdmitted)
	}
	room, err := w.hospital.AssignRoom(p)
	if room != nil {
		return room.Number(), err
	}
	return 0, err
}

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
