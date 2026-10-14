package night

import "testing"

// Every ending plays by itself: Timeline works for each one, with every
// flag combination that changes its path, and gives a plain length.
func TestEndingTimelines(t *testing.T) {
	cases := []struct {
		ending string
		flags  Flags
	}{
		{EndingSleep, Flags{}},
		{EndingSleep, Flags{"route_mayu": true}},
		{EndingSleep, Flags{"route_nobody": true}},
		{EndingWake, Flags{}},
		{EndingWake, Flags{"route_mayu": true}},
		{EndingWake, Flags{"read_letter": true}},
		{EndingForever, Flags{"route_hana": true}},
		{EndingForever, Flags{"route_hana": true, "closed_eyes": true}},
		{Ending317, Flags{}},
	}
	for _, c := range cases {
		s, err := LoadScript(EndingScript(c.ending))
		if err != nil {
			t.Fatal(err)
		}
		got, err := Timeline(s, c.flags, "Test")
		if err != nil {
			t.Fatalf("%s %v: %v", c.ending, c.flags.Names(), err)
		}
		t.Logf("%s %v: %d ticks = %.2f s", c.ending, c.flags.Names(), got, float64(got)/TicksPerSecond)
		if got < 10*TicksPerSecond {
			t.Errorf("%s %v: only %d ticks, an ending is longer than that", c.ending, c.flags.Names(), got)
		}
		again, _ := Timeline(s, c.flags, "Test")
		if again != got {
			t.Errorf("%s: the timeline changed between runs (%d then %d)", c.ending, got, again)
		}
	}
}

// Clicking does not change the timeline: an engine that is clicked on every
// tick ends at the same tick as the one nobody touches.
func TestClicksDoNotChangeTheTimeline(t *testing.T) {
	for _, ending := range Endings {
		s, err := LoadScript(EndingScript(ending))
		if err != nil {
			t.Fatal(err)
		}
		flags := Flags{"route_hana": true}
		want, err := Timeline(s, flags, "Test")
		if err != nil {
			t.Fatal(err)
		}
		e, err := NewEngineFor(s, Flags{"route_hana": true}, "Test")
		if err != nil {
			t.Fatal(err)
		}
		for e.Ticks() < maxTimeline {
			if _, ended := e.Ended(); ended {
				break
			}
			_ = e.Advance()    // a click on every tick
			e.AutoPlay = false // and AUTO switched off
			if err := e.Tick(); err != nil {
				t.Fatal(err)
			}
		}
		if e.Ticks() != want {
			t.Errorf("%s: clicked it ends at tick %d, untouched at %d", ending, e.Ticks(), want)
		}
	}
}

// A script without `autoplay on` has no fixed length, so Timeline refuses it.
func TestTimelineNeedsAutoplay(t *testing.T) {
	s, err := LoadScript(Night1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Timeline(s, Flags{}, "Test"); err == nil {
		t.Error("Timeline accepted a script with no autoplay")
	}
}
