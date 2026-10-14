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

// Visitor is one person in a day's queue: who they are, how they look at
// the window, and the papers they carry. What is right to do with them is
// worked out from the papers (Problem, RightStamp).
type Visitor struct {
	Name  string
	Age   int
	Level hospital.NarcolepsyLevel

	// Sprite is the sheet of this person at the window (assets/sprites/
	// <Sprite>.json) and Mouth the layer they wear while awake.
	Sprite string
	Mouth  string
	// FaceY is the row of the window sprite where the Settler ID photo
	// starts (the photo is a crop of the sprite; the column is always 26).
	// 0 means the usual 14.
	FaceY int

	Papers

	// Returning patients wear a wristband. This one has no new sleep
	// episodes, so the right call is DISCHARGE.
	Returning bool
}

// RightStamp is the stamp that scores for this visitor: no valid referral
// is REFER, a returning patient with nothing new is DISCHARGE, everyone
// else is ADMIT.
func (v Visitor) RightStamp() Stamp {
	switch {
	case !v.ReferralOK():
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

// The places on the Settler IDs.
const (
	gullhaven = "GULLHAVEN"
	saltReach = "SALT REACH"
	theShoals = "THE SHOALS"
)

// person builds a visitor with a valid set of papers: the Settler ID
// matches the referral, it has not expired, Dr. Barrow stamped it and the
// diagnosis is a sleep disorder. A problem is then made by changing one
// field of the result (see the days below).
func person(name, sprite, mouth string, age int, level hospital.NarcolepsyLevel, home string, expires Date, diagnosis, note1, note2 string) Visitor {
	return Visitor{
		Name: name, Sprite: sprite, Mouth: mouth, Age: age, Level: level,
		Papers: Papers{Home: home, Expires: expires, Diagnosis: diagnosis, Notes: [2]string{note1, note2}, Stamp: StampBarrow},
	}
}

// The three days (Game Design -> the 20-minute budget): day 1 has 5
// patients and 3 rooms, day 2 has 6 and 3 rooms, day 3 has 4 and only 2.
var days = buildDays()

func buildDays() []DayRules {
	const (
		neutral = "mouth_neutral"
		smile   = "mouth_smile"
	)
	wren := person("Wren Calloway", "wren_calloway", neutral, 34, hospital.Mild, theShoals, Date{6, 30, 77}, "NARCOLEPSY", "SLEEP ATTACKS AT", "THE CANNERY.") // ID expired

	ozzie := person("Ozzie Pike", "ozzie_pike", smile, 19, hospital.Mild, saltReach, Date{2, 14, 80}, "NARCOLEPSY", "DOZES OFF MID-", "SENTENCE.")
	ozzie.Stamp = StampNone
	ozzie.FaceY = 18

	marcus := person("Marcus Trell", "marcus_trell", neutral, 41, hospital.Severe, saltReach, Date{8, 3, 79}, "NARCOLEPSY", "COLLAPSES AT THE", "SALT PANS.")
	marcus.RefName = "MARTIN TRELL"

	pearl := person("Pearl Wick", "pearl_wick", smile, 29, hospital.Mild, gullhaven, Date{4, 22, 79}, "SEASICKNESS", "FEELS ILL ON THE", "FERRY.")

	cecily := person("Old Cecily", "cecily", smile, 71, hospital.Severe, gullhaven, Date{6, 30, 79}, "NARCOLEPSY", "FALLS ASLEEP AT HER", "FISH STALL.")
	cecilyBack := cecily
	cecilyBack.Returning, cecilyBack.Ward, cecilyBack.LastStay = true, "WARD B", dayDate(1)

	d := []DayRules{
		{Rooms: 3, ShiftSeconds: 180, ArrivalGap: 30, Visitors: []Visitor{
			cecily,
			person("Matty", "matty", neutral, 12, hospital.Moderate, saltReach, Date{3, 15, 80}, "CATAPLEXY", "COLLAPSES WHEN HE", "LAUGHS."),
			wren,
			person("Hank Shoals", "hank_shoals", neutral, 58, hospital.Severe, theShoals, Date{11, 20, 79}, "SLEEP ATTACKS", "NODS OFF WHILE", "MENDING NETS."),
			person("Ivy Marsh", "ivy_marsh", smile, 26, hospital.Moderate, gullhaven, Date{1, 9, 80}, "NARCOLEPSY", "FALLS ASLEEP ON THE", "FERRY."),
		}},
		{Rooms: 3, ShiftSeconds: 180, ArrivalGap: 26, Visitors: []Visitor{
			person("Tobias Reef", "tobias_reef", neutral, 45, hospital.Moderate, saltReach, Date{5, 2, 79}, "NARCOLEPSY", "NODS OFF AT THE", "DOCKS."),
			person("Nell Garrow", "nell_garrow", neutral, 63, hospital.Severe, gullhaven, Date{9, 1, 78}, "SLEEP ATTACKS", "COLLAPSES AT HER", "LOOM."),
			ozzie,
			person("Dot Harlan", "dot_harlan", smile, 52, hospital.Moderate, theShoals, Date{12, 12, 79}, "CATAPLEXY", "DROPS HER TRAY WHEN", "SHE LAUGHS."),
			marcus,
			person("June Alder", "june_alder", smile, 30, hospital.Mild, gullhaven, Date{7, 7, 79}, "NARCOLEPSY", "NODS OFF AT THE", "LIGHTHOUSE."),
		}},
		{Rooms: 2, ShiftSeconds: 150, ArrivalGap: 32, Visitors: []Visitor{
			cecilyBack,
			person("Silas Brook", "silas_brook", neutral, 37, hospital.Moderate, saltReach, Date{10, 30, 79}, "SLEEP ATTACKS", "NODS OFF AT THE", "HELM."),
			pearl,
			person("Hank Shoals", "hank_shoals", neutral, 58, hospital.Severe, theShoals, Date{11, 20, 79}, "SLEEP ATTACKS", "NODS OFF WHILE", "MENDING NETS."),
		}},
	}
	// The date and the sleep log depend on the day, so they are filled in here.
	for i := range d {
		for j := range d[i].Visitors {
			v := &d[i].Visitors[j]
			v.Today = dayDate(i + 1)
			v.RefDate = v.Today
			v.Log = makeLog(*v, i+1)
		}
	}
	return d
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

// windowSprites lists the sheet name of every visitor of every day, once
// each, so the art loader can read them all.
func windowSprites() []string {
	var names []string
	seen := make(map[string]bool)
	for _, d := range days {
		for _, v := range d.Visitors {
			if !seen[v.Sprite] {
				seen[v.Sprite] = true
				names = append(names, v.Sprite)
			}
		}
	}
	return names
}
