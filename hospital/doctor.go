package hospital

import "fmt"

// Doctor diagnoses patients and attends sleep emergencies. A doctor keeps
// their own episode history, so MyEpisodes needs no hospital.
type Doctor struct {
	Person
	specialty string
	patients  []*Patient
	log       episodeLog
}

func NewDoctor(id, name string, age int, specialty string) *Doctor {
	return &Doctor{Person: NewPerson(id, name, age), specialty: specialty}
}

func (d *Doctor) Specialty() string { return d.specialty }

// Patients returns a copy of the patients under this doctor's care.
func (d *Doctor) Patients() []*Patient {
	return append([]*Patient(nil), d.patients...)
}

// DiagnosePatient takes a patient under this doctor's care.
func (d *Doctor) DiagnosePatient(p *Patient) error {
	if p == nil {
		return ErrNilPatient
	}
	if p.doctor != nil {
		return fmt.Errorf("%s is with %s: %w", p.ID(), p.doctor.Name(), ErrAlreadyDiagnosed)
	}
	p.doctor = d
	d.patients = append(d.patients, p)
	return nil
}

// Attend records that this doctor attended a sleeping patient.
func (d *Doctor) Attend(p *Patient, location string) (EpisodeRecord, error) {
	return d.log.attend(d, p, location)
}

// AttendSleepEmergency attends a patient where they fell asleep.
func (d *Doctor) AttendSleepEmergency(p *Patient) error {
	if p == nil {
		return ErrNilPatient
	}
	_, err := d.Attend(p, p.Location())
	return err
}

// MyEpisodes returns a copy of every episode this doctor attended.
func (d *Doctor) MyEpisodes() []EpisodeRecord {
	return d.log.all()
}
