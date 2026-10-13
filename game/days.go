package game

import (
	"fmt"

	"danmin/hospital"
)

// Stamp is one of the three stamps on the desk.
type Stamp int

const (
	StampAdmit Stamp = iota
	StampRefer
	StampDischarge
)

func (s Stamp) String() string {
	switch s {
	case StampAdmit:
		return "ADMIT"
	case StampRefer:
		return "REFER"
	case StampDischarge:
		return "DISCHARGE"
	}
	return "unknown stamp"
}

// Visitor is one person in a day's queue. The papers check is built in M4;
// until then the desk decides what is right from these fields, and M4's
// documents (Settler ID, referral) will be drawn from the same data.
type Visitor struct {
	Name  string
	Age   int
	Level hospital.NarcolepsyLevel

	// ReferralOK is true when the referral is valid (Game Design -> What
	// each stamp means). Problem says what is wrong when it is not.
	ReferralOK bool
	Problem    string
	// Returning patients wear a wristband. This one has no new sleep
	// episodes, so the right call is DISCHARGE.
	Returning bool
}

// RightStamp is the stamp that scores for this visitor: no valid referral
// is REFER, a returning patient with nothing new is DISCHARGE, everyone
// else is ADMIT.
func (v Visitor) RightStamp() Stamp {
	switch {
	case !v.ReferralOK:
		return StampRefer
	case v.Returning:
		return StampDischarge
	}
	return StampAdmit
}

// DayRules is everything that changes from one day to the next.
type DayRules struct {
	Visitors     []Visitor
	Rooms        int // beds in the ward
	ShiftSeconds int // how long the shift lasts, in game seconds
	ArrivalGap   int // game seconds between two visitors walking up
}

// The three days (Game Design -> the 20-minute budget): day 1 has 5
// patients and 3 rooms, day 2 has 6 and 3 rooms, day 3 has 4 and only 2.
var days = []DayRules{
	{
		Rooms: 3, ShiftSeconds: 180, ArrivalGap: 30,
		Visitors: []Visitor{
			{Name: "Old Cecily", Age: 71, Level: hospital.Severe, ReferralOK: true},
			{Name: "Matty", Age: 12, Level: hospital.Moderate, ReferralOK: true},
			{Name: "Wren Calloway", Age: 34, Level: hospital.Mild, Problem: "Settler ID expired"},
			{Name: "Hank Shoals", Age: 58, Level: hospital.Severe, ReferralOK: true},
			{Name: "Ivy Marsh", Age: 26, Level: hospital.Moderate, ReferralOK: true},
		},
	},
	{
		Rooms: 3, ShiftSeconds: 180, ArrivalGap: 26,
		Visitors: []Visitor{
			{Name: "Tobias Reef", Age: 45, Level: hospital.Moderate, ReferralOK: true},
			{Name: "Nell Garrow", Age: 63, Level: hospital.Severe, ReferralOK: true},
			{Name: "Ozzie Pike", Age: 19, Level: hospital.Mild, Problem: "no stamp from Dr. Barrow"},
			{Name: "Dot Harlan", Age: 52, Level: hospital.Moderate, ReferralOK: true},
			{Name: "Marcus Trell", Age: 41, Level: hospital.Severe, Problem: "name does not match the ID"},
			{Name: "June Alder", Age: 30, Level: hospital.Mild, ReferralOK: true},
		},
	},
	{
		Rooms: 2, ShiftSeconds: 150, ArrivalGap: 32,
		Visitors: []Visitor{
			{Name: "Old Cecily", Age: 71, Level: hospital.Severe, ReferralOK: true, Returning: true},
			{Name: "Silas Brook", Age: 37, Level: hospital.Moderate, ReferralOK: true},
			{Name: "Pearl Wick", Age: 29, Level: hospital.Mild, Problem: "diagnosis is not a sleep disorder"},
			{Name: "Hank Shoals", Age: 58, Level: hospital.Severe, ReferralOK: true},
		},
	},
}

// LastDay is the number of the last day with rules.
func LastDay() int { return len(days) }

// RulesForDay returns the rules of day 1 to LastDay.
func RulesForDay(day int) (DayRules, error) {
	if day < 1 || day > len(days) {
		return DayRules{}, fmt.Errorf("day %d: want a day from 1 to %d", day, len(days))
	}
	return days[day-1], nil
}
