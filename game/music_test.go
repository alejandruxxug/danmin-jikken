package game

import (
	"testing"

	"danmingame/night"
)

// Every way into a day with music playing fades the music: the volume falls
// in small steps (never to 0 in one tick), reaches 0 on the last tick, and
// only then does the day's hum start.
func TestMusicFadesIntoEveryDay(t *testing.T) {
	transitions := []struct {
		name  string
		music string // what plays before
		go_   func(env *Env)
	}{
		{"title to day 1", "title", func(env *Env) {
			env.NewRun()
			env.StartPowerOn() // what START does
		}},
		{"night 1 to day 2", "night_ambience", func(env *Env) {
			env.Flags = night.Flags{}
			s := &nightScene{env: env, script: night.Night1}
			if err := s.finish("day2"); err != nil {
				t.Fatal(err)
			}
		}},
		{"night 2 to day 3", "night_ambience", func(env *Env) {
			env.Flags = night.Flags{}
			s := &nightScene{env: env, script: night.Night2}
			if err := s.finish("day3"); err != nil {
				t.Fatal(err)
			}
		}},
	}
	for _, tt := range transitions {
		env := &Env{Day: 1, Sound: soundPlayer(t)}
		env.SetMusic(tt.music)
		tt.go_(env)
		env.SetMusic("day_hum") // the day scene's Enter: it must wait

		prev := env.Sound.MusicGain()
		if prev < 0.99 {
			t.Fatalf("%s: the fade starts at %v, want 1", tt.name, prev)
		}
		for tick := 1; tick <= dayFadeTicks; tick++ {
			env.Sound.Update()
			env.tickMusicFade()
			g := env.Sound.MusicGain()
			if tick < dayFadeTicks && (g >= prev || g <= 0) {
				t.Fatalf("%s: tick %d gain %v after %v: want a steady fall that is not 0 yet", tt.name, tick, g, prev)
			}
			if prev-g > 0.1 {
				t.Fatalf("%s: tick %d gain fell %v in one tick", tt.name, tick, prev-g)
			}
			if tick < dayFadeTicks && env.Sound.Current() == "day_hum" {
				t.Fatalf("%s: the hum started during the fade", tt.name)
			}
			prev = g
		}
		if env.Sound.Current() != "day_hum" {
			t.Errorf("%s: after the fade %q plays, want day_hum", tt.name, env.Sound.Current())
		}
		if g := env.Sound.MusicGain(); g < 0.99 {
			t.Errorf("%s: the hum plays at %v, want full volume", tt.name, g)
		}
	}
}

// No music, or the hum already playing: nothing to fade, no cut either.
func TestNoFadeWhenThereIsNothingToFade(t *testing.T) {
	env := &Env{Day: 1, Sound: soundPlayer(t)}
	env.StartPowerOn()
	if env.Sound.Fading() {
		t.Error("fading with no music")
	}
	env.SetMusic("day_hum")
	env.StartPowerOn()
	if env.Sound.Fading() || env.Sound.Current() != "day_hum" {
		t.Error("the hum must carry on from day to day")
	}
}

// The ending's hard cut is not touched: asking for silence stops at once.
func TestEndingStillCutsDead(t *testing.T) {
	env := &Env{Day: 1, Sound: soundPlayer(t)}
	env.SetMusic("night_ambience")
	env.SetMusic("")
	if env.Sound.Current() != "" || env.Sound.Fading() {
		t.Error("SetMusic(\"\") must still cut at once")
	}
}
