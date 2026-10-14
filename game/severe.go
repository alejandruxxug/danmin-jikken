package game

import "danmin/hospital"

// Very sick patients (Game Design -> Marking very sick patients). Severe
// patients get four things: severity pips on every list, a nurse call light
// that blinks while one of them sleeps with nobody attending, a score bonus
// for being quick, and a body that darkens with every episode.

const (
	pointsSevereQuick = 50 // on top of pointsAttend, for a quick answer
	severeQuickSecs   = 10 // game seconds a Severe sleeper may wait for the bonus
	maxDarkness       = 3  // palette steps a patient can darken (off-white to deep sage)
)

// Pips is how many dots show a level: Mild 1, Moderate 2, Severe 3.
func Pips(l hospital.NarcolepsyLevel) int {
	switch l {
	case hospital.Severe:
		return 3
	case hospital.Moderate:
		return 2
	}
	return 1
}

// Episodes counts the episodes the hospital recorded for each patient id.
// It reads History, which is safe while the patient goroutines run.
func (w *Ward) Episodes() map[string]int {
	counts := make(map[string]int)
	for _, rec := range w.hospital.History() {
		counts[rec.Patient().ID()]++
	}
	return counts
}

// SevereToday asks the hospital how many episodes each Severe patient had
// today (Hospital.SevereReport), keyed by patient id. A Severe patient with
// none still appears with 0.
func (w *Ward) SevereToday() map[string]int {
	counts := make(map[string]int)
	for p, n := range w.hospital.SevereReport() {
		counts[p.ID()] = n
	}
	return counts
}

// Darkness is how many palette steps darker a guest is drawn: one for each
// episode a Severe patient had today, at most maxDarkness. Everyone else is
// never darkened.
func (s *Shift) Darkness(g *Guest) int {
	if g.Level != hospital.Severe {
		return 0
	}
	return min(s.severeToday[g.ID()], maxDarkness)
}

// EpisodesOf is how many episodes the hospital recorded for a guest today.
func (s *Shift) EpisodesOf(g *Guest) int { return s.episodes[g.ID()] }

// SevereCalling is true while a Severe patient is asleep and no staff has
// come: the nurse call light blinks for it.
func (s *Shift) SevereCalling() bool {
	for _, g := range s.Guests {
		if sl := s.sleepers[g.ID()]; sl != nil && !sl.Attended && g.Level == hospital.Severe {
			return true
		}
	}
	return false
}

// quickBonus is the bonus for attending a sleeper: pointsSevereQuick when
// the guest is Severe and has been asleep no longer than severeQuickSecs.
func (s *Shift) quickBonus(g *Guest, sl *Sleeper) int {
	if g.Level == hospital.Severe && s.Seconds()-sl.Since <= severeQuickSecs {
		return pointsSevereQuick
	}
	return 0
}
