package hospital

import "fmt"

// Patient is someone admitted for narcolepsy. Their state only moves through
// methods: SufferSleepAttack puts them to sleep where they stand, a room
// puts them in a bed, and WakeUp gets them back on their feet.
type Patient struct {
	Person
	level    NarcolepsyLevel
	state    PatientState
	location string
	doctor   *Doctor
	room     *Room
}

// NewPatient creates an awake patient at the reception.
func NewPatient(id, name string, age int, level NarcolepsyLevel) *Patient {
	return &Patient{
		Person:   NewPerson(id, name, age),
		level:    level,
		state:    Awake,
		location: "reception",
	}
}

func (p *Patient) Level() NarcolepsyLevel  { return p.level }
func (p *Patient) State() PatientState     { return p.state }
func (p *Patient) Location() string        { return p.location }
func (p *Patient) AssignedDoctor() *Doctor { return p.doctor }

// Room returns the room the patient sleeps in, or nil if they have none.
func (p *Patient) Room() *Room { return p.room }

// SufferSleepAttack makes an awake patient fall asleep where they are.
// Until a room is assigned they block the hallway.
func (p *Patient) SufferSleepAttack(location string) error {
	if p.state != Awake {
		return fmt.Errorf("%s at %s: %w", p.ID(), location, ErrAlreadyAsleep)
	}
	p.state = AsleepInHallway
	p.location = location
	return nil
}

// WakeUp wakes a sleeping patient. If they were in a bed, the room is
// released so the next sleeper can have it.
func (p *Patient) WakeUp() error {
	if p.state == Awake {
		return fmt.Errorf("%s: %w", p.ID(), ErrNotAsleep)
	}
	if p.room != nil {
		if err := p.room.Release(p); err != nil {
			return err
		}
	}
	p.state = Awake
	return nil
}

// putInBed is called by Room.Occupy, the only way a patient gets a bed.
func (p *Patient) putInBed(r *Room) {
	p.room = r
	p.state = AsleepInBed
	p.location = fmt.Sprintf("room %d", r.Number())
}

// leaveBed is called by Room.Release.
func (p *Patient) leaveBed() {
	p.room = nil
}
