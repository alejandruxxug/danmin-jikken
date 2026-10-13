// Package simulation makes the patients sleepy on their own: each patient
// runs in its own goroutine, waits a random time, falls asleep, sleeps a
// random time, wakes up, and repeats.
//
// The goroutines never touch a Patient or Room themselves. Everything goes
// through hospital.Hospital's methods, which take the hospital's mutex, so
// two patients falling asleep at the same moment cannot take the same room.
package simulation

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"danmin/hospital"
)

// EventKind says what happened to a patient.
type EventKind int

const (
	FellAsleep EventKind = iota
	WokeUp
)

func (k EventKind) String() string {
	switch k {
	case FellAsleep:
		return "fell asleep"
	case WokeUp:
		return "woke up"
	}
	return "unknown event"
}

// Event reports one change. It carries the patient's id, not the *Patient,
// so whoever receives it can't read the patient while a goroutine changes it.
type Event struct {
	PatientID string
	Kind      EventKind
	Location  string // where the patient fell asleep (FellAsleep only)
}

// Config controls the simulation. The zero value is valid but boring: no
// waiting, one attack per patient, one location.
type Config struct {
	Seed      int64         // each patient's random source is derived from it
	MaxAwake  time.Duration // longest time awake between attacks
	MaxAsleep time.Duration // longest time asleep
	Attacks   int           // attacks per patient; 0 or less means until ctx is cancelled
	Locations []string      // where patients can fall asleep; empty means "hallway"

	// ManualAttend leaves the attack to whoever listens to Events. The
	// patient still waits, sends FellAsleep, sleeps and sends WokeUp, but the
	// goroutine does not call RegisterEpisode: the listener decides if and
	// when staff come (the game does it when the player clicks). If nobody
	// attended, the patient is still awake in the hospital's eyes and the
	// wake-up finds nothing to do, which is fine.
	ManualAttend bool

	// Sleep waits for d or until ctx is cancelled. Nil uses a real timer;
	// tests replace it so they never wait.
	Sleep func(ctx context.Context, d time.Duration) error

	// Events, if set, receives every fall-asleep and wake-up. The simulation
	// stops sending when ctx is cancelled, so an unread channel can't block it.
	Events chan<- Event
}

// Run starts one goroutine per patient and blocks until all of them finish:
// when every patient has had cfg.Attacks attacks, or when ctx is cancelled.
// Cancelling is a normal way to stop and is not an error. If a hospital call
// fails, that patient's goroutine stops and Run returns the errors, joined,
// each wrapping the cause so errors.Is still works.
func Run(ctx context.Context, h *hospital.Hospital, patients []*hospital.Patient, cfg Config) error {
	if cfg.Sleep == nil {
		cfg.Sleep = timerSleep
	}
	if len(cfg.Locations) == 0 {
		cfg.Locations = []string{"hallway"}
	}

	var wg sync.WaitGroup
	errs := make(chan error, len(patients)) // one slot each, so sends never block
	for i, p := range patients {
		wg.Add(1)
		// rand.Rand is not safe for concurrent use, so every goroutine gets its
		// own, seeded from the config seed and the patient's position.
		rng := rand.New(rand.NewSource(cfg.Seed + int64(i)))
		go func(p *hospital.Patient) {
			defer wg.Done()
			if err := runPatient(ctx, h, p, cfg, rng); err != nil {
				errs <- fmt.Errorf("patient %s: %w", p.ID(), err)
			}
		}(p)
	}
	wg.Wait()
	close(errs)

	var all []error
	for err := range errs {
		all = append(all, err)
	}
	return errors.Join(all...)
}

// runPatient is the life of one patient. It returns nil when ctx is cancelled.
func runPatient(ctx context.Context, h *hospital.Hospital, p *hospital.Patient, cfg Config, rng *rand.Rand) error {
	for n := 0; cfg.Attacks <= 0 || n < cfg.Attacks; n++ {
		if err := cfg.Sleep(ctx, randomUpTo(rng, cfg.MaxAwake)); err != nil {
			return cancelled(err)
		}
		location := cfg.Locations[rng.Intn(len(cfg.Locations))]
		if !cfg.ManualAttend {
			if err := h.RegisterEpisode(p, location); err != nil {
				return err
			}
		}
		send(ctx, cfg.Events, Event{PatientID: p.ID(), Kind: FellAsleep, Location: location})

		if err := cfg.Sleep(ctx, randomUpTo(rng, cfg.MaxAsleep)); err != nil {
			return cancelled(err)
		}
		err := h.WakePatient(p)
		if cfg.ManualAttend && errors.Is(err, hospital.ErrNotAsleep) {
			err = nil // nobody attended this attack, so there is nobody to wake
		}
		if err != nil {
			return err
		}
		send(ctx, cfg.Events, Event{PatientID: p.ID(), Kind: WokeUp})
	}
	return nil
}

// cancelled turns a context cancellation into nil, the normal way to stop.
func cancelled(err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return nil
	}
	return err
}

// randomUpTo returns a duration in [1ns, max], or 0 if max is not positive.
func randomUpTo(rng *rand.Rand, max time.Duration) time.Duration {
	if max <= 0 {
		return 0
	}
	return time.Duration(rng.Int63n(int64(max))) + 1
}

// send delivers an event unless nobody listens or ctx is cancelled first.
func send(ctx context.Context, ch chan<- Event, ev Event) {
	if ch == nil {
		return
	}
	select {
	case ch <- ev:
	case <-ctx.Done():
	}
}

// timerSleep is the real wait: d, or until ctx is cancelled.
func timerSleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
