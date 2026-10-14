package night

import "testing"

func TestRouteFor(t *testing.T) {
	tests := []struct {
		name  string
		flags []string
		want  string
	}{
		{"stay with Mayu", []string{"stay_mayu"}, RouteMayu},
		{"find Hana", []string{"find_hana"}, RouteHana},
		{"call Mayu out after reading the note", []string{"confront_mayu", "liked_hana"}, RouteHana},
		{"call Mayu out without it", []string{"confront_mayu", "liked_mayu"}, RouteNobody},
		{"no choice at all", nil, RouteNobody},
	}
	for _, tt := range tests {
		f := Flags{}
		for _, name := range tt.flags {
			f.Set(name)
		}
		if got := RouteFor(f); got != tt.want {
			t.Errorf("%s: route = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestEveryEndingIsReachable(t *testing.T) {
	tests := []struct {
		flags []string
		want  string
	}{
		{[]string{"route_mayu", "closed_eyes"}, EndingSleep},
		{[]string{"route_nobody", "closed_eyes", "read_letter", "stamped_discharge"}, EndingSleep},
		{[]string{"route_mayu"}, EndingWake},
		{[]string{"route_nobody", "read_letter"}, EndingWake},
		{[]string{"route_nobody", "stamped_discharge"}, EndingWake}, // needs the letter too
		{[]string{"route_hana"}, EndingForever},
		{[]string{"route_hana", "closed_eyes"}, EndingForever},
		{[]string{"route_hana", "read_letter", "stamped_discharge"}, EndingForever},
		{[]string{"route_mayu", "read_letter", "stamped_discharge"}, Ending317},
		{[]string{"route_nobody", "read_letter", "stamped_discharge"}, Ending317},
	}
	reached := map[string]bool{}
	for _, tt := range tests {
		f := Flags{}
		for _, name := range tt.flags {
			f.Set(name)
		}
		got := EndingFor(f)
		if got != tt.want {
			t.Errorf("flags %v: ending = %q, want %q", tt.flags, got, tt.want)
		}
		reached[got] = true
	}
	for _, e := range Endings {
		if !reached[e] {
			t.Errorf("ending %q is never reached", e)
		}
	}
}

// playThrough runs a script to its end, answering menus by the flags.
func playThrough(t *testing.T, name string, flags Flags) (*Engine, string) {
	t.Helper()
	s, err := LoadScript(name)
	if err != nil {
		t.Fatal(err)
	}
	e, err := NewEngine(s, flags)
	if err != nil {
		t.Fatal(err)
	}
	if err := e.FastForward(1 << 30); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	to, ended := e.Ended()
	if !ended {
		t.Fatalf("%s did not end", name)
	}
	return e, to
}

func flagsOf(names ...string) Flags {
	f := Flags{}
	for _, n := range names {
		f.Set(n)
	}
	return f
}

func TestNight2SetsTheRoute(t *testing.T) {
	tests := []struct {
		flags []string
		want  string
	}{
		{[]string{"stay_mayu"}, RouteMayu},
		{[]string{"find_hana"}, RouteHana},
		{[]string{"confront_mayu", "liked_hana"}, RouteHana},
		{[]string{"confront_mayu"}, RouteNobody},
	}
	for _, tt := range tests {
		e, to := playThrough(t, Night2, flagsOf(tt.flags...))
		if to != "day3" || !e.Flags().Has("route_"+tt.want) {
			t.Errorf("%v: ends to %q with flags %v, want day3 and route_%s", tt.flags, to, e.Flags().Names(), tt.want)
		}
	}
}

func TestNight3EndsInAnEndingScene(t *testing.T) {
	for _, route := range []string{"route_mayu", "route_hana", "route_nobody"} {
		e, to := playThrough(t, Night3, flagsOf(route, "read_letter"))
		if to != "ending" || !e.Flags().Has("read_letter") {
			t.Errorf("%s: ends to %q", route, to)
		}
	}
	// Choosing "leave it" must not set read_letter: the second option is picked
	// when no option's flags are on, so ask for it by hand.
	s, _ := LoadScript(Night3)
	e, err := NewEngine(s, Flags{})
	if err != nil {
		t.Fatal(err)
	}
	for len(e.Options()) == 0 {
		_ = e.Advance()
	}
	_ = e.Choose(1)
	if e.Flags().Has("read_letter") {
		t.Error("leaving the letter must not set read_letter")
	}
}

func TestEndingScriptsEndInTheCard(t *testing.T) {
	for _, ending := range Endings {
		for _, route := range []string{"route_mayu", "route_nobody", "route_hana"} {
			_, to := playThrough(t, EndingScript(ending), flagsOf(route, "read_letter"))
			if to != "card" {
				t.Errorf("ending %s (%s) ends to %q, want card", ending, route, to)
			}
		}
	}
}

func TestPlayerNameInTheWords(t *testing.T) {
	s := mustParse(t, "~ Hello {player}.\n")
	e2, err := NewEngineFor(s, Flags{}, "river")
	if err != nil {
		t.Fatal(err)
	}
	if e2.FullText() != "Hello river." {
		t.Errorf("text = %q", e2.FullText())
	}
}

func TestRedscaleAmounts(t *testing.T) {
	e := newEngine(t, "fx redscale\n~ red\nfx redscale_off\n~ grey again\n", Flags{})
	if r, _ := e.RedscaleAmounts(); r != 0 {
		t.Errorf("at tick 0 rise = %v, want 0", r)
	}
	for range 60 {
		_ = e.Tick()
	}
	if r, d := e.RedscaleAmounts(); r != 1 || d != 0 {
		t.Errorf("after 1 s: rise %v drain %v, want 1 and 0", r, d)
	}
	_ = e.Advance()
	_ = e.Advance()
	for range 60 {
		_ = e.Tick()
	}
	if r, d := e.RedscaleAmounts(); r != 1 || d != 1 {
		t.Errorf("after the drain: rise %v drain %v, want 1 and 1", r, d)
	}
}

func TestLockIgnoresClicks(t *testing.T) {
	e := newEngine(t, "lock 1\n~ after\n", Flags{})
	_ = e.Advance()
	if e.FullText() == "after" {
		t.Fatal("a click cut a lock short")
	}
	for range 61 {
		_ = e.Tick()
	}
	if e.FullText() != "after" {
		t.Error("the lock never ended")
	}
}

func TestBloodTimeline(t *testing.T) {
	if b := BloodAt(-1); b != (Blood{}) {
		t.Errorf("before the cut: %+v", b)
	}
	// splash_1, splash_2, splash_3 in order, two ticks each, then the stain.
	for tick, want := range []int{1, 1, 2, 2, 3, 3} {
		if b := BloodAt(tick); b.Splash != want || b.Stain {
			t.Errorf("tick %d: %+v, want splash %d and no stain", tick, b, want)
		}
	}
	after := BloodAt(SplashTicks)
	if after.Splash != 0 || !after.Stain || after.Drips != [DripCount]int{} {
		t.Errorf("right after the splash: %+v, want the stain and no drips yet", after)
	}
	// Later every drip has started, each after the one before, and only goes down.
	late := BloodAt(SplashTicks + 600)
	for i := 1; i < DripCount; i++ {
		if late.Drips[i] >= late.Drips[i-1] {
			t.Errorf("drip %d slid %d, not less than drip %d (%d)", i+1, late.Drips[i], i, late.Drips[i-1])
		}
	}
	if BloodAt(SplashTicks+300).Drips[0] > BloodAt(SplashTicks+301).Drips[0] {
		t.Error("a drip moved up")
	}
}
