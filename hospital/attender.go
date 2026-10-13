package hospital

// Attender is anyone the hospital can send to a sleeping patient. Hospital
// keeps its staff in a []Attender and dispatches them without knowing
// whether it got a Doctor, an Orderly, a Nurse or any type added later.
type Attender interface {
	ID() string
	Name() string
	Attend(p *Patient, location string) (EpisodeRecord, error)
}

// Compile-time checks that all three staff types satisfy Attender.
var (
	_ Attender = (*Doctor)(nil)
	_ Attender = (*Orderly)(nil)
	_ Attender = (*Nurse)(nil)
)
