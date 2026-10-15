package game

// Every day turns on like a CRT (Art Direction -> Day screen CRT): black, a
// bright dot in the middle, a thin line, the line opens into the picture with
// a white flash, then the picture settles. About 0.8 s. The phases are this
// table; the shader (crt.kage) draws them from one number, the overall
// progress, so its three boundary constants must match the table.
var powerOnPhases = []struct {
	Name  string
	Ticks int
}{
	{"dot", 6},     // a bright dot in the centre
	{"line", 4},    // it opens into a thin horizontal line
	{"open", 8},    // the line stretches into the full picture, with a white flash
	{"settle", 30}, // over-bright fading to normal, one pass of the band, the barrel wobbles once
}

// PowerOnState is where the turn-on is at one tick.
type PowerOnState struct {
	Phase    string
	Phased   float64 // 0 to 1 inside the phase
	Progress float64 // 0 to 1 over the whole turn-on: the shader's uniform
}

// PowerOnTicks is the length of the whole turn-on.
func PowerOnTicks() int {
	total := 0
	for _, p := range powerOnPhases {
		total += p.Ticks
	}
	return total
}

// PowerOnAt gives the state at a tick; ok is false before the start and
// after the end (the picture is then simply normal).
func PowerOnAt(tick int) (st PowerOnState, ok bool) {
	total := PowerOnTicks()
	if tick < 0 || tick >= total {
		return PowerOnState{}, false
	}
	left := tick
	for _, p := range powerOnPhases {
		if left < p.Ticks {
			return PowerOnState{Phase: p.Name, Phased: float64(left) / float64(p.Ticks), Progress: float64(tick) / float64(total)}, true
		}
		left -= p.Ticks
	}
	return PowerOnState{}, false
}

// PowerOnLevel is the brightness of the turn-on with "reduce flashing" on:
// no dot, no white flash, no over-bright settle, only a slow rise from black
// to the normal picture (an eased 0 to 1, so it never overshoots). Over the
// 48 ticks its steepest step is under 4 % of the full range.
func PowerOnLevel(progress float64) float64 {
	t := max(0, min(progress, 1))
	return t * t * (3 - 2*t)
}

// StartPowerOn begins the turn-on; the day scene that follows a night calls
// it through the night scene (and -poweron in test mode).
func (e *Env) StartPowerOn() {
	e.PoweringOn, e.PowerOnTick = true, 0
	e.fadeMusicIntoDay() // the TV does not turn on over a music track that stops dead
}

// tickPowerOn moves the turn-on one tick on.
func (e *Env) tickPowerOn() {
	if !e.PoweringOn {
		return
	}
	e.PowerOnTick++
	if e.PowerOnTick >= PowerOnTicks() {
		e.PoweringOn = false
	}
}
