package game

import (
	"math"
	"strings"
	"testing"
	"time"

	"danmingame/night"
)

func TestForeverGain(t *testing.T) {
	const eyes, title = 400, 1300 // the rise runs from 400 to 1294
	cases := []struct {
		name string
		tick int
		eyes int
		want float64
	}{
		{"first line", 0, eyes, foreverQuiet},
		{"just before the eyes", eyes, eyes, foreverQuiet},
		{"halfway up", (eyes + title - foreverRiseStop) / 2, eyes, (1 + foreverQuiet) / 2},
		{"rise is over", title - foreverRiseStop, eyes, 1},
		{"the drop", title, eyes, 1},
		{"under the finale", title + 500, eyes, 1},
		{"no eyes: quiet", 900, -1, foreverQuiet},
		{"no eyes: the drop", title, -1, 1},
	}
	for _, c := range cases {
		if got := foreverGain(c.tick, c.eyes, title); math.Abs(got-c.want) > 1e-9 {
			t.Errorf("%s: foreverGain(%d) = %.4f, want %.4f", c.name, c.tick, got, c.want)
		}
	}
	// It only ever goes up, and slowly at both ends of the rise.
	prev := 0.0
	for tick := 0; tick <= title; tick++ {
		g := foreverGain(tick, eyes, title)
		if g < prev-1e-12 {
			t.Fatalf("the gain went down at tick %d: %.4f after %.4f", tick, g, prev)
		}
		prev = g
	}
}

func TestSyncDecision(t *testing.T) {
	ms := time.Millisecond
	cases := []struct {
		name             string
		expected, actual time.Duration
		gain             float64
		want             syncAction
	}{
		{"on time", 10 * time.Second, 10 * time.Second, 0.1, syncNone},
		{"within tolerance, ahead", 10 * time.Second, 10*time.Second + 39*ms, 0.1, syncNone},
		{"within tolerance, behind", 10 * time.Second, 10*time.Second - 40*ms, 0.1, syncNone},
		{"behind, quiet", 10 * time.Second, 8 * time.Second, 0.1, syncSeek},
		{"ahead, quiet", 10 * time.Second, 10*time.Second + 41*ms, 0.1, syncSeek},
		{"behind, loud, a little", 10 * time.Second, 10*time.Second - 100*ms, 0.9, syncNone},
		{"behind, loud, a lot", 10 * time.Second, 10*time.Second - 200*ms, 0.9, syncReport},
		{"ahead, loud, a lot", 10 * time.Second, 10*time.Second + 200*ms, 0.9, syncReport},
	}
	for _, c := range cases {
		if got := syncDecision(c.expected, c.actual, c.gain); got != c.want {
			t.Errorf("%s: %v, want %v", c.name, got, c.want)
		}
	}
}

// Forever with a music player that starts 1.2 s late (the muted player pretends,
// see sound.SimulateLateStart): the sync seeks while the music is quiet and the
// drop is under the title within one frame. Every route is tried.
func TestForeverDropLandsOnTheTitleWhenTheMusicStartsLate(t *testing.T) {
	for _, flags := range []string{"route_hana", "route_hana,closed_eyes"} {
		for _, late := range []int{0, 72, 150} {
			g, err := New(Options{Enabled: true, Scene: "ending.forever", Flags: flags})
			if err != nil {
				t.Fatal(err)
			}
			g.env.Sound.SimulateLateStart(late)
			if err := g.scenes.Switch("ending.forever"); err != nil { // enter again, now with the late player
				t.Fatal(err)
			}
			s := g.scenes.Current().(*endingScene)
			titleAt := g.env.EndingTitleAt
			seeks, loudSeeks := 0, 0
			var errAtTitle time.Duration
			for tick := 0; s.phase != phaseTitle; tick++ {
				if tick > titleAt+10 {
					t.Fatalf("%s late %d: no title after %d ticks", flags, late, tick)
				}
				g.env.Sound.Update() // the order of Game.Update: the sound first, then the scene
				before := len(g.env.Sound.Log)
				if err := s.Update(); err != nil {
					t.Fatal(err)
				}
				g.env.Input.endFrame()
				for _, l := range g.env.Sound.Log[before:] {
					if strings.HasPrefix(l, "music seek") {
						seeks++
						if g.env.Sound.MusicGain() > syncQuietMax*g.env.Sound.MusicVolume()+1e-9 {
							loudSeeks++
						}
					}
				}
			}
			pos, _ := g.env.Sound.MusicPosition()
			errAtTitle = pos - rockDrops[night.EndingForever]
			t.Logf("%s, music %d ticks late: %d seek(s), title at %d, drop error %v", flags, late, seeks, titleAt, errAtTitle)
			if errAtTitle < -time.Second/60 || errAtTitle > time.Second/60 {
				t.Errorf("%s late %d: the drop is %v from the title, want within one frame", flags, late, errAtTitle)
			}
			if loudSeeks != 0 {
				t.Errorf("%s late %d: %d seek(s) while the music was loud", flags, late, loudSeeks)
			}
			if late > 0 && seeks == 0 {
				t.Errorf("%s late %d: the late start was never corrected", flags, late)
			}
		}
	}
}
