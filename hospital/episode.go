package hospital

import (
	"fmt"
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
