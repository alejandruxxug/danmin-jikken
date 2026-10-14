package night

// The routes and the endings are chosen from flags (Script -> Routes,
// Endings). The scripts set the flags; these two functions are the only
// place that decides what they mean, so the tests can check every case.

// The routes, decided at the end of night 2's Choice 2.
const (
	RouteMayu   = "mayu"
	RouteHana   = "hana"
	RouteNobody = "nobody"
)

// The four endings.
const (
	EndingSleep   = "sleep"
	EndingWake    = "wake"
	EndingForever = "forever"
	Ending317     = "317"
)

// Endings lists them in the order the Script note does.
var Endings = []string{EndingSleep, EndingWake, EndingForever, Ending317}

// RouteFor reads the choices of nights 1 and 2: staying with Mayu is her
// route; going to find Hana is Hana's; calling Mayu out is Hana's route if
// you liked Hana (she heard you) and nobody's otherwise. No choice made
// (a test jumping straight to night 3) is nobody's route.
func RouteFor(f Flags) string {
	switch {
	case f.Has("stay_mayu"):
		return RouteMayu
	case f.Has("find_hana"):
		return RouteHana
	case f.Has("confront_mayu") && f.Has("liked_hana"):
		return RouteHana
	}
	return RouteNobody
}

// EndingFor reads the end of night 3. Hana's route always ends Forever.
// Otherwise closing your eyes is Sleep; keeping them open is Wake, unless
// you discharged your own file on day 3 and read Hana's letter: that is
// the secret 3:17.
func EndingFor(f Flags) string {
	switch {
	case f.Has("route_hana"):
		return EndingForever
	case f.Has("closed_eyes"):
		return EndingSleep
	case f.Has("stamped_discharge") && f.Has("read_letter"):
		return Ending317
	}
	return EndingWake
}
