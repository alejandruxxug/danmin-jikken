package game

import (
	"strings"
	"testing"

	"danmin/hospital"
)

// Every visitor's papers must show the problem that RightStamp acts on:
// what the player can read at the desk and what scores are the same thing.
func TestPapersAgreeWithRightStamp(t *testing.T) {
	problems := map[string]string{ // visitors with a problem and what it is
		"Wren Calloway": "Settler ID expired",
		"Ozzie Pike":    "no stamp from Dr. Barrow",
		"Marcus Trell":  "name does not match the ID",
		"Pearl Wick":    "diagnosis is not a sleep disorder",
	}
	for day := 1; day <= LastDay(); day++ {
		rules, _ := RulesForDay(day)
		for _, v := range rules.Visitors {
			want, broken := problems[v.Name]
			if got := v.Problem(); got != want {
				t.Errorf("day %d %s: problem %q, want %q", day, v.Name, got, want)
			}
			switch {
			case broken && v.RightStamp() != StampRefer:
				t.Errorf("day %d %s has a problem, right stamp is %v", day, v.Name, v.RightStamp())
			case !broken && v.Returning && v.RightStamp() != StampDischarge:
				t.Errorf("day %d %s is returning, right stamp is %v", day, v.Name, v.RightStamp())
			case !broken && !v.Returning && v.RightStamp() != StampAdmit:
				t.Errorf("day %d %s has valid papers, right stamp is %v", day, v.Name, v.RightStamp())
			}
		}
	}
}

// ADMIT needs episodes in the sleep log; a returning patient has none new.
func TestSleepLogsMatchTheCalls(t *testing.T) {
	for day := 1; day <= LastDay(); day++ {
		rules, _ := RulesForDay(day)
		for _, v := range rules.Visitors {
			if len(v.Log) == 0 || len(v.Log) > 12 { // the page has 12 rows
				t.Errorf("day %d %s: %d log rows", day, v.Name, len(v.Log))
			}
			if v.Returning && v.NewEpisodes() != 0 {
				t.Errorf("day %d %s is returning with %d new episodes", day, v.Name, v.NewEpisodes())
			}
			if v.Returning && (v.Ward == "" || v.LastStay == (Date{})) {
				t.Errorf("day %d %s is returning without a wristband", day, v.Name)
			}
			for _, r := range v.Log {
				if !r.Date.Before(v.Today) {
					t.Errorf("day %d %s: episode %v is not in the past", day, v.Name, r.Date)
				}
			}
		}
	}
}

func TestDay2LogsAllShow0317(t *testing.T) {
	rules, _ := RulesForDay(2)
	for _, v := range rules.Visitors {
		if times := v.SleepTimes(); len(times) == 0 || times[0] != "03:17" {
			t.Errorf("%s: sleep times %v, want 03:17 first", v.Name, times)
		}
	}
	day1, _ := RulesForDay(1)
	for _, v := range day1.Visitors {
		if strings.Contains(strings.Join(v.SleepTimes(), " "), "03:17") {
			t.Errorf("%s: day 1 must not show 03:17 yet", v.Name)
		}
	}
}

func TestDateFormat(t *testing.T) {
	if got := dayDate(2).String(); got != "10/12/77" {
		t.Errorf("day 2 = %s, want 10/12/77", got)
	}
	if !(Date{6, 30, 77}).Before(dayDate(1)) || dayDate(1).Before(dayDate(1)) {
		t.Error("Date.Before is wrong")
	}
}

func TestProblemOrder(t *testing.T) {
	v := person("A", "cecily", "mouth_smile", 30, hospital.Mild, gullhaven, Date{1, 1, 70}, "SEASICKNESS", "", "")
	v.Today = dayDate(1)
	if got := v.Problem(); got != "Settler ID expired" {
		t.Errorf("first problem = %q, want the expired ID", got)
	}
}
