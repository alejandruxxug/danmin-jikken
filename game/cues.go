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
}

// sfxFile is the file name a script's sfx cue plays.
func sfxFile(cue string) string {
	if file, ok := sfxAliases[cue]; ok {
		return file
	}
	return cue
}

// rockStarts is where each ending's rock track starts: the game skips the
// quiet part so the track slams in mid-riff (Sound -> The ending rule).
var rockStarts = map[string]time.Duration{
	"sleep":   0,
	"wake":    0,
	"forever": 0,
	"317":     0,
}

// playStun is the stun of the day-to-night strobe: a hit, a ringing whine and
// a heartbeat (Art Direction -> Day to night strobe).
func playStun() {
	playSFX("stun_hit")
	playSFX("stun_ring")
	playSFX("heartbeat")
}
