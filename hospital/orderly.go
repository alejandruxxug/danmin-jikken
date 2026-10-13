package hospital

// Orderly moves sleeping patients. Unlike a Doctor, an orderly has no
// DiagnosePatient: they can attend an episode but can't take anyone under
// their care.
type Orderly struct {
	Person
	log episodeLog
}

func NewOrderly(id, name string, age int) *Orderly {
	return &Orderly{Person: NewPerson(id, name, age)}
}

// Attend records that this orderly attended a sleeping patient.
func (o *Orderly) Attend(p *Patient, location string) (EpisodeRecord, error) {
	return o.log.attend(o, p, location)
}

// Episodes returns a copy of every episode this orderly attended.
func (o *Orderly) Episodes() []EpisodeRecord { return o.log.all() }
