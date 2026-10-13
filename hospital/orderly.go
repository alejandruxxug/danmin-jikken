package hospital

import "fmt"

// Orderly moves sleeping patients. Unlike a Doctor, an orderly has no
// DiagnosePatient: they can attend an episode but can't take anyone under
// their care.
type Orderly struct {
	Person
	episodes []EpisodeRecord
}

func NewOrderly(id, name string, age int) *Orderly {
	return &Orderly{Person: NewPerson(id, name, age)}
}

// Attend records that this orderly attended a sleeping patient.
func (o *Orderly) Attend(p *Patient, location string) (EpisodeRecord, error) {
	if p == nil {
		return EpisodeRecord{}, ErrNilPatient
	}
	if p.State() == Awake {
		return EpisodeRecord{}, fmt.Errorf("%s: %w", p.ID(), ErrNotAsleep)
	}
	id := fmt.Sprintf("%s-E%02d", o.ID(), len(o.episodes)+1)
	rec := newEpisode(id, p, o, location)
	o.episodes = append(o.episodes, rec)
	return rec, nil
}

// Episodes returns a copy of every episode this orderly attended.
func (o *Orderly) Episodes() []EpisodeRecord {
	return append([]EpisodeRecord(nil), o.episodes...)
}
