package main

// The trailer's sound, as a list of what plays when (Trailer note -> Music
// plan). Times are trailer frames; a frame is 800 samples. Three things make
// the sound:
//
//   - beds: the music (the jingle with its tape stop, the night, Ironbound),
//     each a stretch of a file with its own playback speed and loudness;
//   - cues: single sounds put on a frame;
//   - game segments: the sound effects the game itself made while a shot
//     was recorded (the stamp, the strobe), moved to where the shot sits in
//     the trailer. The game's own music is never used.

import (
	"fmt"
	"math"

	"danmingame"
	"danmingame/sound"
)

// Bed is a stretch of music.
type Bed struct {
	File      string
	FileStart float64                 // seconds into the file at frame From
	From, To  int                     // trailer frames: it plays from From up to To, then cuts
	Rate      func(f float64) float64 // playback speed at trailer frame f; nil = 1
	Gain      func(f float64) float64 // loudness at trailer frame f
}

// Cue is one sound on one frame.
type Cue struct {
	File  string
	Frame int
	Gain  float64
	Speed float64 // playback speed; 0 means 1. 0.5 is an octave lower
	Until int     // a hard cut: the sound stops on this trailer frame (0 = it plays out)
}

// GameSeg is the game's own sound effects of a shot's frames From to To
// (frame numbers of the shot), placed so that frame From is on trailer frame Start.
type GameSeg struct {
	Shot     string
	From, To int
	Start    int
	Until    int      // cut everything of it at this trailer frame (0 = no cut)
	Skip     []string // cue names to leave out
}

// AudioPlan is the whole mix.
type AudioPlan struct {
	Beds []Bed
	Cues []Cue
	Segs []GameSeg
}

// Ironbound: the build-up starts at 24.7 s in the file, ten seconds before
// its drop (34.7 s), so the drop lands on the frame the dagger lands.
const (
	ironFileStart = 24.7
	ironDropFile  = 34.7
	ironFrom      = slamFrame - 600 // ten seconds earlier: 2661
	ironTo        = 3960
)

// IronboundDropFrame is the trailer frame the drop of the music is on, worked
// out from the bed: where the file reaches 34.7 s.
func IronboundDropFrame() int {
	return ironFrom + int(math.Round((ironDropFile-ironFileStart)*FPS))
}

func linearGain(from, to int, g0, g1 float64) func(float64) float64 {
	return func(f float64) float64 {
		t := math.Min(1, math.Max(0, (f-float64(from))/float64(to-from)))
		return g0 + (g1-g0)*t
	}
}

// BuildAudioPlan is the trailer's sound. reduced picks the strobe's -reduce-flashing sounds.
func BuildAudioPlan(reduced bool) AudioPlan {
	strobe := shotStrobe
	if reduced {
		strobe = shotStrobeRed
	}
	plan := AudioPlan{
		Beds: []Bed{
			// The Driftwell jingle: from frame 30, a tape stop from T08 to T10, then silence.
			{File: "driftwell_jingle.mp3", From: 30, To: t10Start, Rate: TapeRate,
				Gain: func(float64) float64 { return 0.8 }},
			// The night: from the scanline reveal, 1 s fade in, until the kanashibari's own cut.
			{File: "night_ambience.mp3", From: t12Start + 30, To: t16Start,
				Gain: linearGain(t12Start+30, t12Start+90, 0, 0.6)},
			// Ironbound, rising from quiet (inQuad) to the drop, then whole until the hard cut.
			{File: "ending_rock_forever.mp3", FileStart: ironFileStart, From: ironFrom, To: ironTo,
				Gain: func(f float64) float64 {
					t := math.Min(1, (f-ironFrom)/float64(slamFrame-ironFrom))
					return 0.85 * (0.1 + 0.9*inQuad(t)) // 0.1 to 1, with a little headroom for the limiter
				}},
		},
		Cues: []Cue{
			{File: "bell.wav", Frame: 0, Gain: 0.8},
			{File: "stamp_thunk.wav", Frame: Beat(4), Gain: 0.9}, // the hiring stamp
			{File: "click.wav", Frame: t06Start + 4, Gain: 0.5},  // the computer flips up
			{File: "page_1.ogg", Frame: Beat(22), Gain: 0.6},
			// Everything up to T09 stops dead with the jingle: the silence is the point.
			{File: "snore_3.wav", Frame: Beat(40), Gain: 0.7, Until: t10Start},
			{File: "snore_2.wav", Frame: Beat(43) - 10, Gain: 0.6, Until: t10Start},
			{File: "snore.wav", Frame: Beat(45), Gain: 0.6, Until: t10Start},
			{File: "snore_4.wav", Frame: t09Start + 4, Gain: 0.9, Until: t10Start},
			{File: "stamp_thunk.wav", Frame: Beat(48), Gain: 0.9, Speed: 0.55, Until: t10Start}, // NO ROOM, slowed with the tape
			{File: "snore.wav", Frame: 1500, Gain: 0.25},                                        // the only sound in the silence
			{File: "clock_chime.wav", Frame: clockEnd, Gain: 0.8},
			{File: "monitor_beep.wav", Frame: 1952, Gain: 0.6},
			{File: "door_creak.wav", Frame: 2470, Gain: 0.4},
			{File: "breath.wav", Frame: 2580, Gain: 0.7},
			{File: "creak.wav", Frame: 2670, Gain: 0.6},
			{File: "drip.wav", Frame: 2856, Gain: 0.7},
			{File: "stun_hit.wav", Frame: t18Start, Gain: 0.3},
			{File: "stamp_thunk.wav", Frame: slamFrame, Gain: 0.5, Until: ironTo}, // under the drop
			{File: "bell.wav", Frame: 3978, Gain: 0.8},
		},
		Segs: []GameSeg{
			{Shot: shotOrient, From: 61, To: 185, Start: t04Start},
			{Shot: shotStamp, From: 21, To: 145, Start: t05Start},
			{Shot: shotComputer, From: 1, To: 100, Start: t06Start, Skip: []string{"bell", "folder_slide"}},
			{Shot: strobe, From: 81, To: 200, Start: t11Start},
		},
	}
	for _, f := range []int{2820, 2880, 2928, 2970, 3006} { // the heartbeat speeds up
		plan.Cues = append(plan.Cues, Cue{File: "heartbeat.wav", Frame: f, Gain: 0.8})
	}
	for _, f := range ClockTicks() { // the clock race ticks, faster and faster
		plan.Cues = append(plan.Cues, Cue{File: "click.wav", Frame: f, Gain: 0.3})
	}
	return plan
}

// Render makes the bed's sound on a timeline of total frames.
func (b Bed) Render(clip Audio, total int) Audio {
	out := Audio{Samples: make([]float32, 2*total*SamplesPerFrame)}
	n := (b.To - b.From) * SamplesPerFrame
	frameAt := func(i int) float64 { return float64(b.From) + float64(i)/SamplesPerFrame }
	rate := func(i int) float64 {
		if b.Rate == nil {
			return 1
		}
		return b.Rate(frameAt(i))
	}
	seg := Varispeed(clip, b.FileStart*SampleRate, n, rate)
	for i := 0; i < n; i++ {
		g := float32(b.Gain(frameAt(i)))
		seg.Samples[2*i] *= g
		seg.Samples[2*i+1] *= g
	}
	TaperEnd(seg, cutTaper)
	addAt(out, seg, b.From*SamplesPerFrame, 1)
	return out
}

// MixTrailer mixes the plan on a timeline of total frames. logs holds the
// audio.json of each recorded shot, by shot name.
func MixTrailer(plan AudioPlan, total int, load Loader, logs map[string]game.AudioLog) (Audio, error) {
	out := Audio{Samples: make([]float32, 2*total*SamplesPerFrame)}
	cache := map[string]Audio{}
	get := func(file string) (Audio, error) {
		if a, ok := cache[file]; ok {
			return a, nil
		}
		a, err := load(file)
		if err != nil {
			return Audio{}, fmt.Errorf("%s: %w", file, err)
		}
		cache[file] = a
		return a, nil
	}
	for _, b := range plan.Beds {
		clip, err := get(b.File)
		if err != nil {
			return Audio{}, err
		}
		addAt(out, b.Render(clip, total), 0, 1)
	}
	for _, c := range plan.Cues {
		clip, err := get(c.File)
		if err != nil {
			return Audio{}, err
		}
		speed := c.Speed
		if speed == 0 {
			speed = 1
		}
		n := int(float64(clip.Frames()) / speed)
		if c.Until > 0 {
			n = min(n, (c.Until-c.Frame)*SamplesPerFrame)
		}
		a := Varispeed(clip, 0, n, func(int) float64 { return speed })
		if c.Until > 0 && n == (c.Until-c.Frame)*SamplesPerFrame {
			TaperEnd(a, cutTaper) // cut short: a hard cut
		}
		addAt(out, a, c.Frame*SamplesPerFrame, float32(c.Gain))
	}
	for _, s := range plan.Segs {
		log, ok := logs[s.Shot]
		if !ok {
			return Audio{}, fmt.Errorf("no recording of shot %q for its sound", s.Shot)
		}
		a, err := gameSegAudio(s, log, load)
		if err != nil {
			return Audio{}, err
		}
		addAt(out, a, s.Start*SamplesPerFrame, 1)
	}
	return out, nil
}

// gameSegAudio mixes the game's sound effects of the frames the segment uses,
// the first of them at sample 0.
func gameSegAudio(s GameSeg, log game.AudioLog, load Loader) (Audio, error) {
	frames := s.To - s.From + 1
	if s.Until > 0 {
		frames = min(frames, s.Until-s.Start)
	}
	skip := map[string]bool{}
	for _, name := range s.Skip {
		skip[name] = true
	}
	var events []sound.Event
	for _, e := range log.Events {
		if e.Kind != "sfx" || e.Frame < s.From || e.Frame > s.To || skip[e.Cue] {
			continue
		}
		e.Frame = e.Frame - s.From + 1 // the mixer counts frames from 1
		events = append(events, e)
	}
	a, err := Mix(events, frames, load, MixOptions{DropMusic: true})
	if err != nil {
		return Audio{}, fmt.Errorf("shot %q: %w", s.Shot, err)
	}
	if s.Until > 0 && s.Until-s.Start < s.To-s.From+1 {
		TaperEnd(a, cutTaper) // a hard cut
	}
	return a, nil
}
