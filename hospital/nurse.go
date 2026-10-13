package hospital

// Nurse watches over one ward. Like an Orderly, a nurse attends sleep
// episodes but can't diagnose; what sets her apart is the ward she is
// assigned to. She is the third Attender, added without touching Hospital:
// HireStaff accepts anything that satisfies the interface.
type Nurse struct {
	Person
	ward string
	log  episodeLog
}

func NewNurse(id, name string, age int, ward string) *Nurse {
	return &Nurse{Person: NewPerson(id, name, age), ward: ward}
}

// Ward is the ward this nurse is responsible for.
func (n *Nurse) Ward() string { return n.ward }

// Attend records that this nurse attended a sleeping patient.
func (n *Nurse) Attend(p *Patient, location string) (EpisodeRecord, error) {
	return n.log.attend(n, p, location)
}

// Episodes returns a copy of every episode this nurse attended.
func (n *Nurse) Episodes() []EpisodeRecord { return n.log.all() }
