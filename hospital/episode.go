package hospital

import (
	"fmt"
	"sync"
	"time"
)

// now is the clock every record reads. Tests replace it to get fixed times.
var now = time.Now

// EpisodeRecord is one sudden sleep episode. It is a value: once created it
// keeps the room and attender of that moment, even if the patient later
// wakes up or moves.
type EpisodeRecord struct {
	id       string
	at       time.Time
	patient  *Patient
	attender Attender
	location string
	room     *Room
}

// newEpisode reads the patient's room at this moment (nil if none).
func newEpisode(id string, p *Patient, a Attender, location string) EpisodeRecord {
	return EpisodeRecord{id: id, at: now(), patient: p, attender: a, location: location, room: p.Room()}
}

func (e EpisodeRecord) ID() string         { return e.id }
func (e EpisodeRecord) At() time.Time      { return e.at }
func (e EpisodeRecord) Patient() *Patient  { return e.patient }
func (e EpisodeRecord) Attender() Attender { return e.attender }
func (e EpisodeRecord) Location() string   { return e.location }

// Room returns the room assigned during the episode, or nil.
func (e EpisodeRecord) Room() *Room { return e.room }

// Summary is one readable line, e.g.
// "2026-10-13 14:15 P-001 cafeteria -> room 101 (Dr. Barrow)".
func (e EpisodeRecord) Summary() string {
	dest := "stays in hallway"
	if e.room != nil {
		dest = fmt.Sprintf("room %d", e.room.Number())
	}
	return fmt.Sprintf("%s %s %s -> %s (%s)",
		e.at.Format("2006-01-02 15:04"), e.patient.ID(), e.location, dest, e.attender.Name())
}

// episodeLog is the history every kind of attender keeps (Doctor, Orderly,
// Nurse), so the checks and the id numbering are written once. It has its own
// mutex because MyEpisodes reads it without holding the hospital's lock while
// RegisterEpisode appends to it from another goroutine. Lock order is always
// hospital first, then log, so the two can never deadlock.
type episodeLog struct {
	mu       sync.Mutex
	episodes []EpisodeRecord
}

// attend checks that p is asleep, then records the episode for by.
func (l *episodeLog) attend(by Attender, p *Patient, location string) (EpisodeRecord, error) {
	if p == nil {
		return EpisodeRecord{}, ErrNilPatient
	}
	if p.State() == Awake {
		return EpisodeRecord{}, fmt.Errorf("%s: %w", p.ID(), ErrNotAsleep)
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	id := fmt.Sprintf("%s-E%02d", by.ID(), len(l.episodes)+1)
	rec := newEpisode(id, p, by, location)
	l.episodes = append(l.episodes, rec)
	return rec, nil
}

// all returns a copy, so callers can't change the log.
func (l *episodeLog) all() []EpisodeRecord {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]EpisodeRecord(nil), l.episodes...)
}
