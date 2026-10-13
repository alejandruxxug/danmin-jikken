package game

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"danmin/hospital"
	"danmin/simulation"
)

// Points: only correct choices score, wrong ones just do not (Game Design ->
// Score + top-10). There is no game over at the desk.
const (
	pointsRightStamp = 100
	pointsAttend     = 50

	busySeconds  = 8 // how long a staff member shows BUSY after attending
	toastSeconds = 3 // how long a message stays on the screen
	endDelay     = 4 // seconds between the last patient and the end of the shift
	walkStep     = 3 // pixels a guest walks per tick

	// The day 1 tutorial: Old Cecily has been at the window this many game
	// seconds when she nods off.
	nodOffAfter = 2.0
)

// Errors the desk reports for clicks that do nothing.
var (
	ErrNoOneAtWindow   = errors.New("nobody is waiting at the window")
	ErrNothingToAttend = errors.New("that patient does not need staff")
	ErrDozing          = errors.New("the patient at the window is asleep: wake them first")
)

// guestStatus is where a visitor is in their visit.
type guestStatus int

const (
	guestComing   guestStatus = iota // has not walked up yet
	guestWaiting                     // in the queue, not stamped yet
	guestAdmitted                    // stamped ADMIT, in the hospital
	guestLeft                        // stamped REFER or DISCHARGE, walking out
)

func (s guestStatus) String() string {
	switch s {
	case guestComing:
		return "coming"
	case guestWaiting:
		return "waiting"
	case guestAdmitted:
		return "admitted"
	case guestLeft:
		return "left"
	}
	return "unknown"
}

// Guest is a visitor in today's queue, with the patient the hospital will
// get if they are admitted, and where they stand.
type Guest struct {
	Visitor
	patient *hospital.Patient
	status  guestStatus
	chair   int     // index in chairSpots while they sit in the lobby
	x, y    float64 // where they are in the room (sprite top-left)
	steps   int     // ticks they have been walking without a stop; 0 = standing
	blink   blinker
}

// ID is the patient id ("P-001"), also the id in the hospital.
func (g *Guest) ID() string { return g.patient.ID() }

// Sleeper is a patient who fell asleep and is waiting for staff or a bed.
type Sleeper struct {
	Location string
	Since    float64 // game seconds into the shift when they fell asleep
	Attended bool    // staff came (the hospital recorded the episode)
}

// Alert is the "no room" box.
type Alert struct {
	Guest *Guest
}

// StampRecord is one stamp the player used.
type StampRecord struct {
	Patient string
	Stamp   Stamp
	Right   bool
}

// Attendance says what happened when staff were sent to a sleeper.
type Attendance struct {
	Patient  string
	Attender string // "" when the patient had been attended before
	Room     int    // 0 when there is no bed
	NoRoom   bool
}

// Shift is one day at the desk: the queue, the ward, the score. It has no
// drawing code; the desk scenes read it and draw. Env keeps one while a
// shift runs, so the reception, the computer and the top-10 share it.
type Shift struct {
	Day    int
	Rules  DayRules
	Guests []*Guest

	ward   *Ward
	rng    *rand.Rand // the guests' blink timers; only the game thread uses it
	speed  float64
	scores map[string]int // Env.Scores, written under the key "day<N>"
	ticks  int
	snap   hospital.Snapshot

	sleepers  map[string]*Sleeper // by patient id
	busyUntil map[string]float64  // by staff name
	stamps    []StampRecord
	served    int
	message   string
	messageAt float64
	alert     *Alert
	dozing    *Guest  // the day 1 tutorial guest while they are asleep at the window
	windowAt  float64 // when the first guest reached the window, 0 = not yet
	tutorial  bool    // the nod-off has happened (it only happens once)
	allDoneAt float64 // seconds when the last guest was served, 0 = not yet
	over      bool
}

// NewShift builds the day's queue and starts an empty ward for it.
func NewShift(day int, seed int64, speed float64, sleepy bool, scores map[string]int) (*Shift, error) {
	rules, err := RulesForDay(day)
	if err != nil {
		return nil, err
	}
	if speed <= 0 {
		speed = 1
	}
	ward, err := NewWard(seed+int64(day)*100, speed, rules.Rooms, sleepy)
	if err != nil {
		return nil, err
	}
	ward.Start()
	s := &Shift{
		Day: day, Rules: rules, ward: ward, speed: speed, scores: scores,
		rng:       rand.New(rand.NewSource(seed + int64(day)*100 - 1)),
		sleepers:  make(map[string]*Sleeper),
		busyUntil: make(map[string]float64),
	}
	for i, v := range rules.Visitors {
		id := fmt.Sprintf("P-%03d", i+1)
		s.Guests = append(s.Guests, &Guest{
			Visitor: v,
			patient: hospital.NewPatient(id, v.Name, v.Age, v.Level),
		})
	}
	s.scores[s.scoreKey()] = 0
	return s, nil
}

func (s *Shift) scoreKey() string { return fmt.Sprintf("day%d", s.Day) }

// Score is today's points.
func (s *Shift) Score() int { return s.scores[s.scoreKey()] }

func (s *Shift) addScore(n int) { s.scores[s.scoreKey()] += n }

// Seconds is how long the shift has run, in game seconds.
func (s *Shift) Seconds() float64 { return float64(s.ticks) / 60 * s.speed }

// Served is how many guests have been stamped; Total how many there are.
func (s *Shift) Served() int { return s.served }
func (s *Shift) Total() int  { return len(s.Guests) }

// Over is true once the shift has ended.
func (s *Shift) Over() bool { return s.over }

// Snapshot is the ward as of the last tick.
func (s *Shift) Snapshot() hospital.Snapshot { return s.snap }

// Stamps lists the stamps used so far.
func (s *Shift) Stamps() []StampRecord { return s.stamps }

// Message is the line shown after the last click, "" when it has faded.
func (s *Shift) Message() string {
	if s.dozing != nil { // the tutorial hint stays until the player wakes them
		return strings.ToUpper(s.dozing.Name) + " DOZED OFF MID-JOKE. CLICK HER."
	}
	if s.message == "" || s.Seconds()-s.messageAt > toastSeconds {
		return ""
	}
	return s.message
}

func (s *Shift) say(format string, args ...any) {
	s.message = strings.ToUpper(fmt.Sprintf(format, args...))
	s.messageAt = s.Seconds()
}

// Alert is the open "no room" box, or nil.
func (s *Shift) Alert() *Alert { return s.alert }

// DismissAlert closes the "no room" box.
func (s *Shift) DismissAlert() { s.alert = nil }

// Stop ends the patients' goroutines. It returns what they reported.
func (s *Shift) Stop() error { return s.ward.Stop() }

// Update moves the shift one tick forward. Call it once per frame.
func (s *Shift) Update() {
	s.ticks++
	// Collect what the patient goroutines reported since the last frame.
	// The picture itself comes from the snapshot, never from the events.
	for _, ev := range s.ward.Drain() {
		switch ev.Kind {
		case simulation.FellAsleep:
			s.sleepers[ev.PatientID] = &Sleeper{Location: ev.Location, Since: s.Seconds()}
		case simulation.WokeUp:
			delete(s.sleepers, ev.PatientID)
		}
	}
	s.snap = s.ward.Snapshot()

	now := s.Seconds()
	for i, g := range s.Guests {
		if g.status == guestComing && now >= float64(1+s.Rules.ArrivalGap*i) {
			s.arrive(g)
		}
	}
	s.layout()
	s.tutorialStep(now)

	if s.alert != nil && s.stateOf(s.alert.Guest) != hospital.AsleepInHallway {
		s.alert = nil // they woke up or got a bed: nothing left to warn about
	}
	s.checkEnd(now)
}

// tutorialStep is the scripted moment of day 1 (Game Design -> design log):
// a few seconds after Old Cecily reaches the window she nods off mid-joke.
// It happens once, and only on day 1.
func (s *Shift) tutorialStep(now float64) {
	if s.Day != 1 || s.tutorial || s.served > 0 {
		return
	}
	g := s.Guests[0]
	if g.status != guestWaiting || !s.atWindow(g) {
		return
	}
	if s.windowAt == 0 {
		s.windowAt = now
	}
	if now-s.windowAt >= nodOffAfter {
		s.dozing = g
		s.tutorial = true
	}
}

// Dozing is the guest who nodded off at the window, or nil.
func (s *Shift) Dozing() *Guest { return s.dozing }

// Wake is the player's click on the guest who nodded off. This is a moment
// of the game, not of the hospital: the guest has not been admitted yet, so
// the hospital knows nothing about them and has no episode to attend. That
// is why this only clears a flag and calls no Hospital method (Attend would
// fail with ErrPatientNotFound for a patient who is not admitted).
func (s *Shift) Wake(g *Guest) error {
	if s.dozing != g {
		return fmt.Errorf("%s: %w", g.Name, ErrNothingToAttend)
	}
	s.dozing = nil
	s.say("%s WAKES UP", g.Name)
	return nil
}

// arrive lets a guest walk in: from the right if the window is free, from
// the left to a chair if somebody is already being served.
func (s *Shift) arrive(g *Guest) {
	g.status = guestWaiting
	if s.window() == g {
		g.x, g.y = float64(windowSpot.x+walkFrom), float64(windowSpot.y)
		return
	}
	g.chair = s.freeChair()
	spot := chairSpots[g.chair]
	g.x, g.y = float64(spot.x-walkFrom), float64(spot.y)
}

// window is the guest being served: the first one waiting in the queue.
func (s *Shift) window() *Guest {
	for _, g := range s.Guests {
		if g.status == guestWaiting {
			return g
		}
	}
	return nil
}

// freeChair is the first chair nobody sits in. Chairs 1 and 2 come first
// because they are the ones in the middle of the default view.
func (s *Shift) freeChair() int {
	used := make(map[int]bool)
	front := s.window()
	for _, g := range s.Guests {
		if g.onChair() && g != front { // the window guest is not sitting
			used[g.chair] = true
		}
	}
	for _, c := range []int{1, 2, 3, 4, 0} {
		if !used[c] {
			return c
		}
	}
	return 0
}

// onChair is true for guests who are in the lobby (not at the window).
func (g *Guest) onChair() bool {
	return g.status == guestAdmitted || g.status == guestWaiting
}

// stateOf is the hospital's view of a guest: awake until they are admitted.
func (s *Shift) stateOf(g *Guest) hospital.PatientState {
	for _, p := range s.snap.Patients {
		if p.ID == g.ID() {
			return p.State
		}
	}
	return hospital.Awake
}

// Look is how a guest is drawn.
type Look int

const (
	LookHidden   Look = iota // not here: not arrived, gone, or in a bed
	LookStanding             // awake, standing or sitting
	LookAsleep               // asleep where they sit, nobody has come yet
	LookHallway              // asleep on the hallway floor with no bed
	LookDozing               // asleep at the window (the day 1 tutorial)
)

// LookOf decides how to draw a guest from the hospital's snapshot.
func (s *Shift) LookOf(g *Guest) Look {
	switch g.status {
	case guestComing:
		return LookHidden
	case guestLeft:
		if g.x >= RoomW {
			return LookHidden
		}
		return LookStanding
	case guestWaiting:
		if g == s.dozing {
			return LookDozing
		}
		return LookStanding
	}
	switch s.stateOf(g) {
	case hospital.AsleepInBed:
		return LookHidden
	case hospital.AsleepInHallway:
		return LookHallway
	}
	if s.sleepers[g.ID()] != nil {
		return LookAsleep // fell asleep, no episode recorded yet
	}
	return LookStanding
}

// hallwaySlot is the position of this guest among those on the hallway floor.
func (s *Shift) hallwaySlot(g *Guest) int {
	n := 0
	for _, other := range s.Guests {
		if other == g {
			break
		}
		if s.LookOf(other) == LookHallway {
			n++
		}
	}
	return min(n, len(hallwaySpots)-1)
}

// layout walks every guest toward the spot they should be at.
func (s *Shift) layout() {
	for _, g := range s.Guests {
		var tx, ty float64
		switch s.LookOf(g) {
		case LookHidden:
			if g.status == guestLeft {
				continue
			}
			// Out of sight (not arrived, or in a bed): wait at the chair.
			if g.status == guestAdmitted {
				spot := chairSpots[g.chair]
				g.x, g.y = float64(spot.x), float64(spot.y)
			}
			continue
		case LookHallway:
			spot := hallwaySpots[s.hallwaySlot(g)]
			g.x, g.y = float64(spot.x), float64(spot.y) // they dropped there
			continue
		}
		switch {
		case g.status == guestLeft:
			tx, ty = RoomW+walkFrom, g.y
		case s.window() == g:
			tx, ty = float64(windowSpot.x), float64(windowSpot.y)
		default:
			spot := chairSpots[g.chair]
			tx, ty = float64(spot.x), float64(spot.y)
		}
		nx, ny := approach(g.x, tx), approach(g.y, ty)
		if nx != g.x || ny != g.y {
			g.steps++
		} else {
			g.steps = 0
		}
		g.x, g.y = nx, ny
		g.blink.update(s.Seconds(), s.rng, g.steps == 0)
	}
}

// Blinking says whether a guest has their eyes shut right now.
func (s *Shift) Blinking(g *Guest) bool { return g.blink.closed(s.Seconds()) }

// stepOffsets is the up-and-down of a walk: up one pixel, up two, back up
// one, down. The sprite is only moved; no art is drawn.
var stepOffsets = [4]int{0, -1, -2, -1}

// WalkBob is how many pixels a guest is lifted after walking for steps
// ticks. One step lasts 125 ms (7.5 ticks, the same as a boil frame), so
// four steps make the cycle; 0 steps (standing) is exactly the resting
// position, so a guest that stops ends on the pixel they should be on.
func WalkBob(steps int) int {
	if steps <= 0 {
		return 0
	}
	return stepOffsets[(steps*2/15)%len(stepOffsets)]
}

// approach moves v toward target by at most walkStep.
func approach(v, target float64) float64 {
	switch {
	case v < target-walkStep:
		return v + walkStep
	case v > target+walkStep:
		return v - walkStep
	}
	return target
}

// atSpot is true when the guest stands exactly where they should.
func (s *Shift) atWindow(g *Guest) bool {
	return g.x == float64(windowSpot.x) && g.y == float64(windowSpot.y)
}

// checkEnd ends the shift when time is up, or a few seconds after the last
// guest was served and nobody is still asleep waiting for staff.
func (s *Shift) checkEnd(now float64) {
	if s.over {
		return
	}
	if s.served == len(s.Guests) && s.allDoneAt == 0 {
		s.allDoneAt = now
	}
	waiting := false
	for _, g := range s.Guests {
		if s.needsStaff(g) {
			waiting = true
		}
	}
	timeUp := now >= float64(s.Rules.ShiftSeconds)
	allDone := s.allDoneAt > 0 && now-s.allDoneAt >= endDelay && !waiting
	if timeUp || allDone {
		s.over = true
	}
}

// needsStaff is true for a guest asleep with no staff and no bed yet.
func (s *Shift) needsStaff(g *Guest) bool {
	switch s.LookOf(g) {
	case LookAsleep, LookHallway:
		return true
	}
	return false
}

// Stamp uses a stamp on the guest at the window. A right stamp scores. ADMIT
// puts the patient into the hospital (and starts their goroutine); REFER
// and DISCHARGE send them out. Wrong stamps still happen, they just do not
// score: there is no game over at the desk.
func (s *Shift) Stamp(st Stamp) error {
	g := s.window()
	if g == nil || !s.atWindow(g) {
		return ErrNoOneAtWindow
	}
	if g == s.dozing {
		return ErrDozing
	}
	right := g.RightStamp() == st
	if right {
		s.addScore(pointsRightStamp)
	}
	// Decided (Alejo, 2026-10-13): a wrong ADMIT still admits the patient.
	// The citation is flavour, and the wrong admission costs a bed, which is
	// its own consequence; only the points differ.
	if st == StampAdmit {
		if err := s.ward.Admit(g.patient); err != nil {
			return fmt.Errorf("admitting %s: %w", g.Name, err)
		}
		g.status = guestAdmitted
		g.chair = s.freeChair()
	} else {
		g.status = guestLeft
	}
	s.served++
	s.stamps = append(s.stamps, StampRecord{Patient: g.Name, Stamp: st, Right: right})
	if right {
		s.say("%s: %s", st, g.Name)
	} else {
		s.say("CITATION: WRONG STAMP FOR %s", g.Name)
	}
	return nil
}

// Dispatch sends staff to a guest who is asleep. A guest nobody has
// attended yet gets the next attender in line (one scoring click); a guest
// already in the hallway is only tried for a bed again. When there is no
// room the "no room" alert opens.
func (s *Shift) Dispatch(g *Guest) (Attendance, error) {
	a := Attendance{Patient: g.Name}
	sl := s.sleepers[g.ID()]
	if sl == nil {
		return a, fmt.Errorf("%s: %w", g.Name, ErrNothingToAttend)
	}
	if !sl.Attended {
		who, err := s.ward.Attend(g.ID(), sl.Location)
		if err != nil {
			return a, err
		}
		sl.Attended = true
		a.Attender = who
		s.busyUntil[who] = s.Seconds() + busySeconds
		// Decided (Alejo, 2026-10-13): attending a sleeper in the hallway
		// still scores the 50 points; getting them staff is the right call
		// whether or not there was a bed.
		s.addScore(pointsAttend)
	}
	// RegisterEpisode already gave a room if one was free; AssignRoom tells
	// us which one, or that there is none (hospital.ErrNoRoomAvailable).
	room, err := s.ward.AssignRoom(g.ID())
	switch {
	case err == nil || errors.Is(err, hospital.ErrAlreadyInBed):
		a.Room = room
	case errors.Is(err, hospital.ErrNoRoomAvailable):
		a.NoRoom = true
	default:
		return a, err
	}
	switch {
	case a.NoRoom:
		s.alert = &Alert{Guest: g}
		s.say("NO ROOM FOR %s", g.Name)
	case a.Attender != "":
		s.alert = nil
		s.say("%s attends %s", a.Attender, g.Name)
	default:
		s.alert = nil
		s.say("%s goes to room %d", g.Name, a.Room)
	}
	return a, nil
}

// CallStaff is the CALL STAFF button: help the sleeper who has waited the
// longest, someone nobody attended before anyone already in the hallway.
func (s *Shift) CallStaff() (Attendance, error) {
	var pick *Guest
	for _, g := range s.Guests {
		if !s.needsStaff(g) {
			continue
		}
		if pick == nil || s.rank(g) < s.rank(pick) {
			pick = g
		}
	}
	if pick == nil {
		s.say("NO ONE NEEDS STAFF")
		return Attendance{}, ErrNothingToAttend
	}
	return s.Dispatch(pick)
}

// rank orders sleepers for CallStaff: unattended first, then by who fell
// asleep earlier.
func (s *Shift) rank(g *Guest) float64 {
	sl := s.sleepers[g.ID()]
	if sl == nil {
		return 1e9
	}
	if sl.Attended {
		return 1e6 + sl.Since
	}
	return sl.Since
}

// Waiting lists the sleepers with no bed, oldest first, for the beds panel.
func (s *Shift) Waiting() []*Guest {
	var out []*Guest
	for _, g := range s.Guests {
		if s.needsStaff(g) {
			out = append(out, g)
		}
	}
	for i := 1; i < len(out); i++ { // insertion sort by Since: a handful of guests
		for j := i; j > 0 && s.sleepers[out[j].ID()].Since < s.sleepers[out[j-1].ID()].Since; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

// SleptFor is how long a sleeper has been asleep, in game seconds.
func (s *Shift) SleptFor(g *Guest) float64 {
	if sl := s.sleepers[g.ID()]; sl != nil {
		return s.Seconds() - sl.Since
	}
	return 0
}

// Busy says whether a staff member attended someone a moment ago.
func (s *Shift) Busy(name string) bool { return s.Seconds() < s.busyUntil[name] }

// StaffNames lists the staff in round-robin order.
func (s *Shift) StaffNames() []string { return s.ward.StaffNames() }

// GuestByPatient finds the guest for a patient id.
func (s *Shift) GuestByPatient(id string) *Guest {
	for _, g := range s.Guests {
		if g.ID() == id {
			return g
		}
	}
	return nil
}

// Blinking (Alejo, 2026-10-13, "the chars are blinking too quickly"): awake
// guests close their eyes for a quarter of a second every 4 to 8 seconds
// while they wait. The drawing swaps the eyes_normal
// layer for eyes_closed.
const (
	blinkFor    = 0.25 // game seconds the eyes stay shut
	blinkMin    = 4.0  // shortest time between two blinks
	blinkSpread = 4.0  // blinks come every blinkMin to blinkMin+blinkSpread seconds (4 to 8)
)

// blinker is one guest's blink timer. The gap to the next blink comes from
// the shift's seeded random source, so a -seed run blinks the same way every
// time, and two guests do not blink together.
type blinker struct {
	next  float64 // when the next blink starts (game seconds)
	until float64 // the eyes are shut until then
}

// update runs once a tick. canBlink is false while the guest walks or is
// asleep: then the eyes stay open and the timer waits.
func (b *blinker) update(now float64, rng *rand.Rand, canBlink bool) {
	if !canBlink {
		b.until = 0
		if b.next <= now {
			b.next = now + blinkMin + rng.Float64()*blinkSpread
		}
		return
	}
	if now >= b.next {
		b.until = now + blinkFor
		b.next = now + blinkMin + rng.Float64()*blinkSpread
	}
}

// closed is true while the eyes are shut.
func (b *blinker) closed(now float64) bool { return now < b.until }
