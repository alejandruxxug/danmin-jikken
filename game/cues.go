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
// the SFX volume). The default is 1. The stun's ringing is held down (a high
// tone). The snores set their own gain by where the sleeper is (snore.go).
var sfxGain = map[string]float64{
	"chime_right": 0.7,
	"stun_ring":   0.6,
	"stamp_lift":  0.8,
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
	"forever": 0, // Ironbound: the real start is computed (syncedStart), this is only the fallback
	"317":     0, // Unholy Surge, from the start
}

// rockDrops is, for the endings whose music is timed to the picture, the
// moment inside the rock file that must land exactly when the title pops.
// Only Forever has one: Ironbound's crunch, 1:24.7 of the original, in a file
// cut from foreverCutFrom of the original (assets/sound/README.txt).
var rockDrops = map[string]time.Duration{
	"forever": foreverCrunch - foreverCutFrom,
}

const (
	foreverCrunch = 84700 * time.Millisecond // the measured crunch in the original
	// ending_rock_forever.mp3 is the original cut at MP3 frame 1914 (frames are
	// 1152 samples at 44.1 kHz), about 50.0 s, to the end of the song.
	foreverCutFrom = time.Duration(1914*1152) * time.Second / 44100
)

// foreverFadeInTicks is the short fade-in of Forever's music (3 s).
const foreverFadeInTicks = 180

// syncedStart is where in its file a track must start so that its drop lands
// after titleAt ticks of playing. If the file does not reach back that far
// the track starts at 0 and clamped says so (the drop will then be late).
func syncedStart(drop time.Duration, titleAt int) (start time.Duration, clamped bool) {
	start = drop - time.Duration(titleAt)*time.Second/60
	if start < 0 {
		return 0, true
	}
	return start, false
}

// creditsFadeTicks is how long the rock fades out at the very end of the
// credits roll (4 s).
const creditsFadeTicks = 240

// playStun is the stun of the day-to-night strobe: a hit, a ringing whine and
// a heartbeat (Art Direction -> Day to night strobe).
func playStun() {
	playSFX("stun_hit")
	playSFX("stun_ring")
	playSFX("heartbeat")
}
