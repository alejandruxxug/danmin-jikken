package hospital

import (
	"fmt"
	"sync"
)

// Hospital owns the patients, staff and rooms, and is the only entry point
// that is safe to call from several goroutines: every exported method takes
// mu. Methods ending in Locked expect mu to be held already; Go's Mutex is
// not reentrant, so an exported method must never call another one.
type Hospital struct {
	mu       sync.Mutex
	name     string
	doctors  []*Doctor
	staff    []Attender
	patients []*Patient
	rooms    []*Room
	history  []EpisodeRecord
	next     int // index in staff of the next attender to dispatch
}

func NewHospital(name string) *Hospital {
	return &Hospital{name: name}
}

func (h *Hospital) Name() string { return h.name }

// HireDoctor adds a doctor to the staff that can be dispatched.
func (h *Hospital) HireDoctor(d *Doctor) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.addStaffLocked(d); err != nil {
		return err
	}
	h.doctors = append(h.doctors, d)
	return nil
}

// HireStaff adds any Attender (an Orderly, or a type added later).
func (h *Hospital) HireStaff(a Attender) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.addStaffLocked(a)
}

func (h *Hospital) addStaffLocked(a Attender) error {
	for _, s := range h.staff {
		if s.ID() == a.ID() {
			return fmt.Errorf("staff %s: %w", a.ID(), ErrDuplicateID)
		}
	}
	h.staff = append(h.staff, a)
	return nil
}

// AddRoom registers a room.
func (h *Hospital) AddRoom(r *Room) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, existing := range h.rooms {
		if existing.Number() == r.Number() {
			return fmt.Errorf("room %d: %w", r.Number(), ErrDuplicateID)
		}
	}
	h.rooms = append(h.rooms, r)
	return nil
}

// AdmitPatient registers a patient in the hospital.
func (h *Hospital) AdmitPatient(p *Patient) error {
	if p == nil {
		return ErrNilPatient
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, existing := range h.patients {
		if existing.ID() == p.ID() {
			return fmt.Errorf("patient %s: %w", p.ID(), ErrDuplicateID)
		}
	}
	h.patients = append(h.patients, p)
	return nil
}

// RegisterEpisode handles a sudden sleep attack: the patient falls asleep at
// location, gets the first free room if there is one, and the next staff
// member in line attends and records it. No free room is not an error here:
// the episode is recorded without a room and the patient stays in the
// hallway, where PatientsInHallway will find them.
func (h *Hospital) RegisterEpisode(p *Patient, location string) error {
	if p == nil {
		return ErrNilPatient
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.admittedLocked(p) {
		return fmt.Errorf("%s: %w", p.ID(), ErrNotAdmitted)
	}
	if len(h.staff) == 0 {
		return ErrNoStaff
	}
	if err := p.SufferSleepAttack(location); err != nil {
		return err
	}
	_, _ = h.assignRoomLocked(p) // no room: they stay in the hallway
	attender := h.staff[h.next%len(h.staff)]
	h.next++
	rec, err := attender.Attend(p, location)
	if err != nil {
		return err
	}
	h.history = append(h.history, rec)
	return nil
}

// AssignRoom gives a patient asleep in the hallway the first free room.
// With no free room it returns an error wrapping ErrNoRoomAvailable and the
// patient stays in the hallway.
func (h *Hospital) AssignRoom(p *Patient) (*Room, error) {
	if p == nil {
		return nil, ErrNilPatient
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.admittedLocked(p) {
		return nil, fmt.Errorf("%s: %w", p.ID(), ErrNotAdmitted)
	}
	return h.assignRoomLocked(p)
}

func (h *Hospital) assignRoomLocked(p *Patient) (*Room, error) {
	switch p.State() {
	case Awake:
		return nil, fmt.Errorf("%s: %w", p.ID(), ErrNotAsleep)
	case AsleepInBed:
		return p.Room(), fmt.Errorf("%s: %w", p.ID(), ErrAlreadyInBed)
	}
	for _, r := range h.rooms {
		if r.IsAvailable() {
			if err := r.Occupy(p); err != nil {
				return nil, err
			}
			return r, nil
		}
	}
	return nil, fmt.Errorf("%s stays in %s: %w", p.ID(), p.Location(), ErrNoRoomAvailable)
}

// WakePatient wakes a patient and frees their room, if they had one.
func (h *Hospital) WakePatient(p *Patient) error {
	if p == nil {
		return ErrNilPatient
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.admittedLocked(p) {
		return fmt.Errorf("%s: %w", p.ID(), ErrNotAdmitted)
	}
	return p.WakeUp()
}

// PatientsInHallway returns the patients asleep in a hallway, in admission
// order, so an orderly can be dispatched to them.
func (h *Hospital) PatientsInHallway() []*Patient {
	h.mu.Lock()
	defer h.mu.Unlock()
	var out []*Patient
	for _, p := range h.patients {
		if p.State() == AsleepInHallway {
			out = append(out, p)
		}
	}
	return out
}

// SevereReport pairs every Severe patient with how many episodes they had
// today. A Severe patient with no episodes still appears, with 0.
func (h *Hospital) SevereReport() map[*Patient]int {
	h.mu.Lock()
	defer h.mu.Unlock()
	report := make(map[*Patient]int)
	for _, p := range h.patients {
		if p.Level() == Severe {
			report[p] = 0
		}
	}
	y, m, d := now().Date()
	for _, rec := range h.history {
		ry, rm, rd := rec.At().Date()
		if ry != y || rm != m || rd != d {
			continue
		}
		if _, severe := report[rec.Patient()]; severe {
			report[rec.Patient()]++
		}
	}
	return report
}

// History returns a copy of every episode recorded in the hospital.
func (h *Hospital) History() []EpisodeRecord {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]EpisodeRecord(nil), h.history...)
}

// Patients returns a copy of the admitted patients, in admission order.
func (h *Hospital) Patients() []*Patient {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]*Patient(nil), h.patients...)
}

// Rooms returns a copy of the rooms, in the order they were added.
func (h *Hospital) Rooms() []*Room {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]*Room(nil), h.rooms...)
}

// Doctors returns a copy of the hired doctors.
func (h *Hospital) Doctors() []*Doctor {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]*Doctor(nil), h.doctors...)
}

// Staff returns a copy of everyone who can be dispatched.
func (h *Hospital) Staff() []Attender {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]Attender(nil), h.staff...)
}

// PatientView is a read-only copy of one patient taken under the lock.
// Code that runs while patient goroutines are active (the game's Update
// loop, for example) reads these instead of calling methods on *Patient,
// which would race with the goroutines.
type PatientView struct {
	ID       string
	Name     string
	Level    NarcolepsyLevel
	State    PatientState
	Location string
	Room     int    // room number, 0 when the patient has no room
	Doctor   string // doctor's name, "" when none
}

// RoomView is a read-only copy of one room taken under the lock.
type RoomView struct {
	Number    int
	Capacity  int
	State     RoomState
	Occupants []string // patient ids
}

// Snapshot is a consistent picture of the whole hospital at one instant.
type Snapshot struct {
	Patients []PatientView
	Rooms    []RoomView
}

// Snapshot copies the state of every patient and room while holding the
// lock, so it is safe to call while other goroutines are changing them.
func (h *Hospital) Snapshot() Snapshot {
	h.mu.Lock()
	defer h.mu.Unlock()
	var s Snapshot
	for _, p := range h.patients {
		v := PatientView{ID: p.ID(), Name: p.Name(), Level: p.level, State: p.state, Location: p.location}
		if p.room != nil {
			v.Room = p.room.number
		}
		if p.doctor != nil {
			v.Doctor = p.doctor.Name()
		}
		s.Patients = append(s.Patients, v)
	}
	for _, r := range h.rooms {
		v := RoomView{Number: r.number, Capacity: r.capacity, State: r.state}
		for _, o := range r.occupants {
			v.Occupants = append(v.Occupants, o.ID())
		}
		s.Rooms = append(s.Rooms, v)
	}
	return s
}

func (h *Hospital) admittedLocked(p *Patient) bool {
	for _, existing := range h.patients {
		if existing == p {
			return true
		}
	}
	return false
}
