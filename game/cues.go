package game

import "time"

// The sound cues. A script says `sfx footsteps_many`; the file is named by
// the cue unless sfxAliases points several cues at one file. A cue with no
// file is silent (sound.Player.Missing lists it).

var sfxAliases = map[string]string{
	"monitor_far":     "monitor_beep",
	"beep_slow":       "monitor_beep",
	"beep":            "monitor_beep",
	"long_tone":       "flatline",
	"footsteps_many":  "footsteps",
	"footsteps_stop":  "footsteps",
	"wheels_squeak":   "gurney",
	"drip_fast":       "drip",
	"door_close_soft": "door_creak",
	"floorboard":      "creak",
	"breathe_again":   "breath",
	"rail_clang":      "click",
	"wheels_slow":     "gurney",
	"wheels_stop":     "gurney",
}

// musicFile is the file a script's music cue plays. The Sleep ending says
// `music jingle`, the Prologue `music driftwell_jingle`: one file.
func musicFile(cue string) string {
	if cue == "jingle" {
		return "driftwell_jingle"
	}
	return cue
}

// isJingle reports whether a music cue is the Driftwell jingle, which plays
// once instead of looping.
func isJingle(cue string) bool { return musicFile(cue) == "driftwell_jingle" }

// sfxGain is how loud a cue plays compared with its file (0 to 1, on top of
// the SFX volume). The default is 1. The snore and the stun's ringing are held
// down: the snore repeats under the whole shift, the ringing is a high tone.
var sfxGain = map[string]float64{
	"snore":      0.5,
	"stun_ring":  0.6,
	"stamp_lift": 0.8,
}

// sfxFile is the file name a script's sfx cue plays.
func sfxFile(cue string) string {
	if file, ok := sfxAliases[cue]; ok {
		return file
	}
	return cue
}

// rockStarts is where each ending's rock track starts inside its file. The
// files in assets/sound are already cut at the measured slam (Sound -> The
// ending rule; the cut times are in assets/sound/README.txt), so every start is
// 0 and the rock hits the moment the card shows. Cutting the file, not the
// code, keeps the game from decoding minutes of music it never plays.
var rockStarts = map[string]time.Duration{
	"sleep":   0, // Sad orchestral theme, cut at 2:22.5 of the original
	"wake":    0, // Realm of Torment, from the start
	"forever": 0, // Ironbound, cut at 1:24.7 of the original
	"317":     0, // Unholy Surge, from the start
}

// creditsFadeTicks is how long the rock fades out at the very end of the
// credits roll (4 s).
const creditsFadeTicks = 240

// snoreEvery is how often a sleeping patient snores again (3.5 s), as long
// as someone is asleep. One cue for the whole ward, so ten sleepers are not
// ten times as loud.
const snoreEvery = 210

// playStun is the stun of the day-to-night strobe: a hit, a ringing whine and
// a heartbeat (Art Direction -> Day to night strobe).
func playStun() {
	playSFX("stun_hit")
	playSFX("stun_ring")
	playSFX("heartbeat")
}
