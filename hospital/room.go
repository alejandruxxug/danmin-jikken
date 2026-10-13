package hospital

import "fmt"

// Room holds up to capacity sleeping patients.
type Room struct {
	number    int
	capacity  int
	state     RoomState
	occupants []*Patient
}

// NewRoom creates an empty room. A capacity below 1 is raised to 1.
func NewRoom(number, capacity int) *Room {
	if capacity < 1 {
		capacity = 1
	}
	return &Room{number: number, capacity: capacity, state: Available}
}

func (r *Room) Number() int      { return r.number }
func (r *Room) Capacity() int    { return r.capacity }
func (r *Room) State() RoomState { return r.state }

// Occupants returns a copy, so callers can't change who is in the room.
func (r *Room) Occupants() []*Patient {
	return append([]*Patient(nil), r.occupants...)
}

// IsAvailable reports whether one more patient fits.
func (r *Room) IsAvailable() bool { return len(r.occupants) < r.capacity }

// Occupy puts a sleeping patient in a bed in this room.
func (r *Room) Occupy(p *Patient) error {
	if p == nil {
		return ErrNilPatient
	}
	if !r.IsAvailable() {
		return fmt.Errorf("room %d: %w", r.number, ErrRoomFull)
	}
	switch p.State() {
	case Awake:
		return fmt.Errorf("%s: %w", p.ID(), ErrNotAsleep)
	case AsleepInBed:
		return fmt.Errorf("%s: %w", p.ID(), ErrAlreadyInBed)
	}
	r.occupants = append(r.occupants, p)
	p.putInBed(r)
	r.updateState()
	return nil
}

// Release takes a patient out of this room.
func (r *Room) Release(p *Patient) error {
	for i, o := range r.occupants {
		if o == p {
			r.occupants = append(r.occupants[:i], r.occupants[i+1:]...)
			p.leaveBed()
			r.updateState()
			return nil
		}
	}
	return fmt.Errorf("room %d: %w", r.number, ErrNotInRoom)
}

func (r *Room) updateState() {
	if r.IsAvailable() {
		r.state = Available
	} else {
		r.state = Occupied
	}
}
