package hospital

import "errors"

// Sentinel errors. Callers compare them with errors.Is; the methods wrap
// them with %w to add which patient, room or staff member was involved.
var (
	ErrNilPatient       = errors.New("patient is nil")
	ErrDuplicateID      = errors.New("id already registered")
	ErrNotAdmitted      = errors.New("patient is not admitted")
	ErrAlreadyAsleep    = errors.New("patient is already asleep")
	ErrNotAsleep        = errors.New("patient is not asleep")
	ErrAlreadyInBed     = errors.New("patient is already in a bed")
	ErrNoRoomAvailable  = errors.New("no room available")
	ErrRoomFull         = errors.New("room is full")
	ErrNotInRoom        = errors.New("patient is not in this room")
	ErrNoStaff          = errors.New("no staff to dispatch")
	ErrAlreadyDiagnosed = errors.New("patient already has a doctor")
)
