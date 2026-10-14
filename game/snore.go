package game

import (
	"fmt"
	"math/rand"
	"sort"

	"danmin/hospital"
)

// The ward's snoring (Game Design -> design log, 2026-10-14). Each snore
// belongs to one sleeper: it picks one of snoreVariants recordings (never the
// same one twice in a row), plays at a volume that depends on where the
// sleeper is, and the next snore comes after a random pause. The choices are
// pure functions of a dice roll (0 up to, not including, 1), so the table
// tests can pin every number; Shift draws the rolls from its own seeded
// random source, so a -seed run snores the same way every time.

// snoreVariants is how many snore recordings there are: the cues snore,
// snore_2, snore_3 ... in assets/sound.
const snoreVariants = 4

// snorePlace is where a sleeper is, as far as the sound goes.
type snorePlace int

const (
	snoreInRoom snorePlace = iota // asleep in a bed, behind a wall: quiet
	snoreInOpen                   // no bed (hallway, waiting room...): loud and close
)

func (p snorePlace) String() string {
	if p == snoreInRoom {
		return "room"
	}
	return "hallway"
}

const (
	snoreRoomGainMin, snoreRoomGainMax = 0.20, 0.30
	snoreOpenGainMin, snoreOpenGainMax = 0.80, 1.00

	snoreGapMin    = 180 // ticks: 3 s, the shortest base pause
	snoreGapSpread = 240 // ticks: up to 4 s more, so 3 to 7 s
	snoreGapFloor  = 120 // ticks: never closer than 2 s, so it is not a wall of noise
)

// snoreName is the cue of a variant (1 is "snore", then "snore_2" ...).
func snoreName(variant int) string {
	if variant <= 1 {
		return "snore"
	}
	return fmt.Sprintf("snore_%d", variant)
}

// nextSnoreVariant picks a variant from 1 to snoreVariants that is not last
// (last 0 means none yet).
func nextSnoreVariant(last int, roll float64) int {
	if last < 1 || last > snoreVariants {
		return int(roll*snoreVariants) + 1
	}
	v := int(roll*(snoreVariants-1)) + 1 // one of the other variants
	if v >= last {
		v++ // skip over the last one
	}
	return v
}

// snoreGain is the volume of one snore: quiet through a wall, loud in the open.
func snoreGain(place snorePlace, roll float64) float64 {
	if place == snoreInRoom {
		return snoreRoomGainMin + roll*(snoreRoomGainMax-snoreRoomGainMin)
	}
	return snoreOpenGainMin + roll*(snoreOpenGainMax-snoreOpenGainMin)
}

// snoreGap is the pause in ticks before the next snore. More sleepers means
// shorter pauses (down to the floor); someone in the open makes them a bit
// shorter, and a ward of bed sleepers only is a bit slower.
func snoreGap(sleepers int, anyInOpen bool, roll float64) int {
	gap := (snoreGapMin + roll*snoreGapSpread) / (1 + 0.3*float64(max(sleepers, 1)-1))
	if anyInOpen {
		gap *= 0.8
	} else {
		gap *= 1.4
	}
	return max(int(gap), snoreGapFloor)
}

// snoreState is the shift's snoring: its own random source (the blink timers
// keep theirs, so adding sounds never changes how the guests blink).
type snoreState struct {
	rng  *rand.Rand
	next int // tick of the next snore, 0 = nothing scheduled
	last int // the variant that played last, 0 = none yet
}

// placeOf is where a sleeper is: in a bed or not.
func (s *Shift) placeOf(id string) snorePlace {
	for _, g := range s.Guests {
		if g.ID() == id && s.stateOf(g) == hospital.AsleepInBed {
			return snoreInRoom
		}
	}
	return snoreInOpen
}

// snoreNow makes one sleeper snore: picks a variant, sets the volume from
// where they are, plays it and notes it in the sound log.
func (s *Shift) snoreNow(id string) {
	place := s.placeOf(id)
	variant := nextSnoreVariant(s.snore.last, s.snore.rng.Float64())
	s.snore.last = variant
	gain := snoreGain(place, s.snore.rng.Float64())
	playSFXGain(snoreName(variant), gain)
	if activeSound != nil {
		activeSound.Note(fmt.Sprintf("snore %s %s variant %d gain %.2f", id, place, variant, gain))
	}
}

// snoreStep runs once a tick: while anyone sleeps, a random sleeper snores
// whenever the pause is over.
func (s *Shift) snoreStep() {
	if len(s.sleepers) == 0 {
		s.snore.next = 0 // silence: the next sleeper starts a fresh pause
		return
	}
	if s.snore.next == 0 {
		s.scheduleSnore()
		return
	}
	if s.ticks < s.snore.next {
		return
	}
	// Map order is random, so sort the ids: the same seed picks the same sleeper.
	ids := make([]string, 0, len(s.sleepers))
	for id := range s.sleepers {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	s.snoreNow(ids[s.snore.rng.Intn(len(ids))])
	s.scheduleSnore()
}

// scheduleSnore sets the tick of the next snore.
func (s *Shift) scheduleSnore() {
	open := false
	for id := range s.sleepers {
		if s.placeOf(id) == snoreInOpen {
			open = true
		}
	}
	s.snore.next = s.ticks + snoreGap(len(s.sleepers), open, s.snore.rng.Float64())
}
