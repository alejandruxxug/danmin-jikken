package hospital

// PatientState is where a patient is in the sleep cycle.
type PatientState int

const (
	Awake PatientState = iota
	AsleepInHallway
	AsleepInBed
)

func (s PatientState) String() string {
	switch s {
	case Awake:
		return "awake"
	case AsleepInHallway:
		return "asleep in hallway"
	case AsleepInBed:
		return "asleep in bed"
	}
	return "unknown state"
}

// NarcolepsyLevel is how strong a patient's narcolepsy is.
type NarcolepsyLevel int

const (
	Mild NarcolepsyLevel = iota
	Moderate
	Severe
)

func (l NarcolepsyLevel) String() string {
	switch l {
	case Mild:
		return "mild"
	case Moderate:
		return "moderate"
	case Severe:
		return "severe"
	}
	return "unknown level"
}

// RoomState says whether a room can take another patient.
type RoomState int

const (
	Available RoomState = iota
	Occupied
)

func (s RoomState) String() string {
	switch s {
	case Available:
		return "available"
	case Occupied:
		return "occupied"
	}
	return "unknown state"
}
