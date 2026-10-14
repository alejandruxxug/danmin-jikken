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

// Night 3 plays its own track from the opening eyes, is silent in Hana's
// room, comes back at the gurney and fades out over the last lines, so the
// volume is 0 when the ending starts. The test clicks through the real
// script, giving every line two seconds of ticks.
func TestNight3MusicPlaysAndFadesBeforeTheEnding(t *testing.T) {
	env := &Env{Day: 3, Flags: night.Flags{}, Sound: soundPlayer(t)}
	env.Flags.Set("route_mayu")
	script, err := night.LoadScript(night.Night3)
	if err != nil {
		t.Fatal(err)
	}
	e, err := night.NewEngineFor(script, env.Flags, "Alejo")
	if err != nil {
		t.Fatal(err)
	}
	s := &nightScene{env: env, script: night.Night3, engine: e}

	playing := map[string]bool{} // what played on which line of the script
	var fadeGains []float64
	ended := func() bool { _, ok := e.Ended(); return ok }
	for step := 0; step < 3000 && !ended(); step++ {
		if err := e.Tick(); err != nil {
			t.Fatal(err)
		}
		switch {
		case len(e.Options()) > 0:
			err = e.Choose(0)
		default:
			if err = e.Advance(); err == nil {
				err = e.Advance()
			}
		}
		if err != nil {
			t.Fatal(err)
		}
		s.playCues()
		playing[env.Sound.Current()] = true
		for range 120 {
			before := env.Sound.Fading()
			env.Sound.Update()
			env.tickMusicFade()
			if before || env.Sound.Fading() {
				fadeGains = append(fadeGains, env.Sound.MusicGain())
			}
		}
	}
	if !ended() {
		t.Fatal("night 3 did not reach its end")
	}
	if !playing["night3_tension"] {
		t.Error("night3_tension never played")
	}
	tension := 0
	for _, entry := range env.Sound.Log {
		if entry == "music night3_tension" {
			tension++
		}
	}
	if tension != 2 {
		t.Errorf("night3_tension was started %d times, want 2 (the eyes opening, the gurney)", tension)
	}
	if len(fadeGains) < dayFadeTicks/2 {
		t.Fatalf("the fade ran only %d ticks", len(fadeGains))
	}
	prev := 1.0
	for i, g := range fadeGains {
		if g > prev || prev-g > 0.1 {
			t.Fatalf("fade tick %d: %v after %v", i, g, prev)
		}
		prev = g
	}
	if env.Sound.MusicGain() != 0 || env.Sound.Current() != "" {
		t.Errorf("at the ending: gain %v, playing %q; want silence", env.Sound.MusicGain(), env.Sound.Current())
	}
	if len(fadeGains) < 290 || len(fadeGains) > 310 {
		t.Errorf("the fade took %d ticks, want 5 s (300)", len(fadeGains))
	}
}

// Clicking through the last lines faster than the fade: the ending still
// starts in silence.
func TestNight3FadeIsFinishedWhenTheEndingStarts(t *testing.T) {
	env := &Env{Day: 3, Flags: night.Flags{}, Sound: soundPlayer(t)}
	env.SetMusic("night3_tension")
	env.FadeMusic(300)
	s := &nightScene{env: env, script: night.Night3}
	if err := s.finish("ending"); err != nil {
		t.Fatal(err)
	}
	if env.Sound.MusicGain() != 0 || env.Sound.Fading() {
		t.Errorf("gain %v fading %v at the ending", env.Sound.MusicGain(), env.Sound.Fading())
	}
}
