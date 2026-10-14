package game

// The music gives way to the day (Alejo, 2026-10-14: "fade the scary music
// before turning on the TV since its now a hard cut"). Every night-to-day
// change, and START on the title, calls StartPowerOn, which starts this fade:
// the playing track fades out over dayFadeTicks while the TV turns on over
// it, and the day's hum waits for the end of the fade instead of cutting the
// track dead. The endings keep their deliberate hard cut: they never come
// through StartPowerOn.

// dayFadeTicks is the fade-out, 1.2 s (the TV's turn-on takes 0.8 s of it).
const dayFadeTicks = 72

// musicFade is a fade into a day in progress, and the track that was asked
// for while it runs.
type musicFade struct {
	active bool
	queued bool   // a track was asked for during the fade
	next   string // that track ("" stops the music)
}

// fadeMusicIntoDay fades out whatever is playing, unless it is already the
// day's hum or nothing plays (then there is nothing to cut).
func (e *Env) fadeMusicIntoDay() {
	if e.Sound == nil {
		return
	}
	if cur := e.Sound.Current(); cur == "" || cur == "day_hum" {
		return
	}
	e.Sound.FadeOutEased(dayFadeTicks)
	e.fade = musicFade{active: true}
}

// tickMusicFade runs once a tick after the player's own update: when the
// fade is over, the track that was asked for during it starts.
func (e *Env) tickMusicFade() {
	if !e.fade.active || e.Sound.Fading() {
		return
	}
	f := e.fade
	e.fade = musicFade{}
	if f.queued {
		e.SetMusic(f.next)
	}
}
