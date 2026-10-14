package game

import (
	"strings"
	"testing"

	"danmin/hospital"
)

func TestNextSnoreVariant(t *testing.T) {
	tests := []struct {
		name string
		last int
		roll float64
		want int
	}{
		{"first, low roll", 0, 0, 1},
		{"first, high roll", 0, 0.99, snoreVariants},
		{"skips last at the bottom", 1, 0, 2},
		{"skips last at the top", snoreVariants, 0.99, snoreVariants - 1},
		{"below last stays", 3, 0, 1},
		{"above last shifts up", 2, 0.99, snoreVariants},
	}
	for _, tt := range tests {
		if got := nextSnoreVariant(tt.last, tt.roll); got != tt.want {
			t.Errorf("%s: got %d, want %d", tt.name, got, tt.want)
		}
	}
	// Never the same twice, over every last and a sweep of rolls.
	for last := 1; last <= snoreVariants; last++ {
		for i := 0; i < 100; i++ {
			v := nextSnoreVariant(last, float64(i)/100)
			if v == last || v < 1 || v > snoreVariants {
				t.Fatalf("last %d roll %d%%: got %d", last, i, v)
			}
		}
	}
}

func TestSnoreGain(t *testing.T) {
	tests := []struct {
		place snorePlace
		roll  float64
		want  float64
	}{
		{snoreInRoom, 0, 0.20},
		{snoreInRoom, 1, 0.30},
		{snoreInOpen, 0, 0.80},
		{snoreInOpen, 1, 1.00},
	}
	for _, tt := range tests {
		if got := snoreGain(tt.place, tt.roll); got < tt.want-1e-9 || got > tt.want+1e-9 {
			t.Errorf("%v roll %v: got %v, want %v", tt.place, tt.roll, got, tt.want)
		}
	}
	if snoreGain(snoreInRoom, 1) >= snoreGain(snoreInOpen, 0) {
		t.Error("a room snore must always be quieter than a hallway one")
	}
}

func TestSnoreGap(t *testing.T) {
	tests := []struct {
		name     string
		sleepers int
		open     bool
		roll     float64
		want     int
	}{
		{"one in a room, short", 1, false, 0, 252}, // 180 * 1.4
		{"one in a room, long", 1, false, 0.9999, 588},
		{"one in the open, short", 1, true, 0, 144}, // 180 * 0.8
		{"one in the open, long", 1, true, 0.9999, 336},
		{"many in the open hits the floor", 10, true, 0, snoreGapFloor},
		{"zero counts as one", 0, true, 0, 144},
	}
	for _, tt := range tests {
		got := snoreGap(tt.sleepers, tt.open, tt.roll)
		if got < tt.want-1 || got > tt.want+1 {
			t.Errorf("%s: got %d, want %d", tt.name, got, tt.want)
		}
	}
	if snoreGap(3, true, 0.5) >= snoreGap(1, true, 0.5) {
		t.Error("more sleepers must snore more often")
	}
	if snoreGap(2, true, 0.5) >= snoreGap(2, false, 0.5) {
		t.Error("a sleeper in the open must snore more often than bed sleepers")
	}
}

func TestSnoreName(t *testing.T) {
	if snoreName(1) != "snore" || snoreName(3) != "snore_3" {
		t.Errorf("names: %q %q", snoreName(1), snoreName(3))
	}
}

// With nobody asleep nothing snores; with a sleeper in the hallway it snores
// loudly, never the same variant twice in a row, and the log says where.
func TestShiftSnores(t *testing.T) {
	p := soundPlayer(t)
	old := activeSound
	activeSound = p
	defer func() { activeSound = old }()

	sh, err := NewShift(1, 7, 1, false, map[string]int{})
	if err != nil {
		t.Fatal(err)
	}
	defer sh.Stop()
	for i := 0; i < 1000; i++ {
		sh.snoreStep()
		sh.ticks++
	}
	if len(p.Log) != 0 {
		t.Fatalf("snored with nobody asleep: %v", p.Log)
	}
	sh.sleepers["P-001"] = &Sleeper{Location: "hallway"}
	for i := 0; i < 6000; i++ {
		sh.snoreStep()
		sh.ticks++
	}
	last := ""
	plays := 0
	for _, e := range p.Log {
		if !strings.HasPrefix(e, "sfx snore") {
			continue
		}
		plays++
		if e == last {
			t.Errorf("the same snore twice in a row: %q", e)
		}
		last = e
	}
	if plays < 10 {
		t.Errorf("only %d snores in 100 s", plays)
	}
	notes := strings.Join(p.Log, "\n")
	if !strings.Contains(notes, "note snore P-001 hallway variant") {
		t.Errorf("the log does not show the place:\n%s", notes)
	}
	sh.sleepers = map[string]*Sleeper{}
	p.Log = nil
	for i := 0; i < 1000; i++ {
		sh.snoreStep()
		sh.ticks++
	}
	if len(p.Log) != 0 {
		t.Errorf("kept snoring after everyone woke: %v", p.Log)
	}
}

func TestPlaceOfBedIsRoom(t *testing.T) {
	sh, err := NewShift(1, 7, 1, false, map[string]int{})
	if err != nil {
		t.Fatal(err)
	}
	defer sh.Stop()
	g := sh.Guests[0]
	if sh.placeOf(g.ID()) != snoreInOpen {
		t.Error("an awake or unknown guest is not in a room")
	}
	sh.snap.Patients = append(sh.snap.Patients, hospital.PatientView{ID: g.ID(), State: hospital.AsleepInBed})
	if sh.placeOf(g.ID()) != snoreInRoom {
		t.Error("a guest asleep in a bed is in a room")
	}
}

// A right stamp chimes when the +100 floats up (three ticks after the
// impact), a wrong one never does.
func TestChimeOnRightStampOnly(t *testing.T) {
	for _, tt := range []struct {
		name  string
		stamp Stamp
		want  int
	}{{"right", StampAdmit, 1}, {"wrong", StampDischarge, 0}} {
		p := soundPlayer(t)
		old := activeSound
		activeSound = p
		sh := folderOnDesk(t, 2)
		if err := sh.GoToPage(PageReferral); err != nil {
			t.Fatal(err)
		}
		runTicks(sh, openTicks)
		p.Log = nil
		sh.LiftStamp(tt.stamp)
		if err := sh.ApplyStamp(); err != nil {
			t.Fatal(err)
		}
		runTicks(sh, hitStopTicks-1)
		if strings.Contains(strings.Join(p.Log, ","), "chime_right") {
			t.Errorf("%s: the chime came before the +100", tt.name)
		}
		runTicks(sh, 2)
		got := strings.Count(strings.Join(p.Log, ","), "sfx chime_right")
		activeSound = old
		if got != tt.want {
			t.Errorf("%s stamp: %d chimes, want %d", tt.name, got, tt.want)
		}
	}
}
