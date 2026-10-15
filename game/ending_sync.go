package game

import (
	"fmt"
	"os"
	"time"
)

// Forever's music is tied to the picture in two ways (Sound -> The ending
// rule): its loudness follows the story (very quiet, then rising from the
// moment the eyes close, full at the drop), and its position is checked
// against the real audio clock so the drop lands on the title.

const (
	foreverQuiet    = 0.09 // the music's volume at the first line, times the music setting
	foreverRiseStop = 6    // the rise ends this many ticks before the drop, full volume from then on
)

// foreverGain is how loud Forever's music is (0 to 1, times the music
// setting) at tick t of the ending. It sits at foreverQuiet until the eyes
// start to close (eyesAt), rises along a smoothstep curve (slow start, slow
// landing) and is at 1 for the drop at titleAt.
func foreverGain(t, eyesAt, titleAt int) float64 {
	riseEnd := titleAt - foreverRiseStop
	if eyesAt < 0 || riseEnd <= eyesAt { // no eyes in this script: quiet, then the drop
		if t >= titleAt {
			return 1
		}
		return foreverQuiet
	}
	if t <= eyesAt {
		return foreverQuiet
	}
	if t >= riseEnd {
		return 1
	}
	x := float64(t-eyesAt) / float64(riseEnd-eyesAt)
	return foreverQuiet + (1-foreverQuiet)*x*x*(3-2*x)
}

// How the music is kept on time. Every syncEvery ticks the real position of
// the music is compared with where it should be; a drift over syncTolerance
// is fixed by a seek, but only while the music is quiet (a jump in loud music
// is a click). Past the quiet limit a big drift can only be reported.
const (
	syncEvery     = 6
	syncTolerance = 40 * time.Millisecond
	syncLoudDrift = 150 * time.Millisecond
	syncQuietMax  = 0.3 // music gain up to which a seek is allowed
	syncSettle    = 30  // ticks to wait after a seek before judging again
)

type syncAction int

const (
	syncNone   syncAction = iota // on time
	syncSeek                     // off, and quiet: jump to the right place
	syncReport                   // far off while loud: a seek would click, only say so
)

// syncDecision says what to do when the music is at actual and should be at
// expected, with the music playing at gain. Pure, so it has a table test.
func syncDecision(expected, actual time.Duration, gain float64) syncAction {
	drift := actual - expected
	if drift < 0 {
		drift = -drift
	}
	switch {
	case drift <= syncTolerance:
		return syncNone
	case gain <= syncQuietMax:
		return syncSeek
	case drift > syncLoudDrift:
		return syncReport
	}
	return syncNone // a little off while loud: leave it alone
}

// syncLog turns on the clock log on stderr (DANMIN_SYNCLOG=1): the tick the
// music starts, where it really is, and the tick and position of the title.
var syncLog = os.Getenv("DANMIN_SYNCLOG") != ""

func logSync(format string, args ...any) {
	if syncLog {
		fmt.Fprintf(os.Stderr, "synclog: "+format+"\n", args...)
	}
}

// musicShouldBeAt is where in its file Forever's music must be at tick clock,
// so that the drop is under it when the title pops at titleAt.
func musicShouldBeAt(drop time.Duration, clock, titleAt int) time.Duration {
	return drop - time.Duration(titleAt-clock)*time.Second/60
}

// syncMusic runs once per tick of Forever's script, black beat and title.
func (s *endingScene) syncMusic() {
	snd := s.env.Sound
	if !s.timed || snd == nil || s.clock > s.env.EndingTitleAt {
		return
	}
	snd.SetMusicShape(foreverGain(s.clock, s.eyesAt, s.env.EndingTitleAt))
	pos, ok := snd.MusicPosition()
	if !ok {
		return
	}
	expected := musicShouldBeAt(rockDrops[s.name], s.clock, s.env.EndingTitleAt)
	if s.clock%60 == 0 {
		logSync("tick %d gain %.2f music at %.3f s, should be %.3f s (drift %+.0f ms)", s.clock, snd.MusicGain(), pos.Seconds(), expected.Seconds(), (pos-expected).Seconds()*1000)
	}
	// The sound card can take a moment to start (the position stays where the
	// track began): there is nothing to judge until the music has moved.
	if !s.musicMoved {
		if pos == s.musicStart {
			return
		}
		s.musicMoved = true
		logSync("tick %d: the music starts to move", s.clock)
	}
	if s.clock%syncEvery != 0 || s.clock < s.settleUntil {
		return
	}
	switch syncDecision(expected, pos, snd.MusicGain()) {
	case syncSeek:
		logSync("tick %d: seek from %.3f s to %.3f s (drift %+.0f ms)", s.clock, pos.Seconds(), expected.Seconds(), (pos-expected).Seconds()*1000)
		snd.SeekMusic(expected)
		s.settleUntil = s.clock + syncSettle
	case syncReport:
		note := fmt.Sprintf("ending %s: music %+.0f ms off at tick %d while loud, not seeking", s.name, (pos-expected).Seconds()*1000, s.clock)
		snd.Note(note)
		logSync("%s", note)
	}
}
