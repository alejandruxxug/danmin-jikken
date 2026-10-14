// Package sound plays the game's music and sound effects with Ebitengine's
// audio package. The game asks for a cue by name ("stamp_thunk"); the
// package finds the file in the embedded sound folder (name.ogg, name.mp3 or
// name.wav). A cue with no file stays silent and is listed in Missing, so
// the game works with whatever sounds exist.
//
// In test mode (the agent backdoor) the Player is muted: nothing is decoded
// or played, but every request is still recorded in Log, so a test can check
// that the right cue was asked for.
package sound

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"path"
	"sort"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

// SampleRate is the rate every file is converted to.
const SampleRate = 44100

// extensions are tried in this order when a cue is looked up.
var extensions = []string{".ogg", ".mp3", ".wav"}

// Player is the game's sound system.
type Player struct {
	fsys     fs.FS
	ctx      *audio.Context // nil when muted or when no audio device was found
	muted    bool
	musicVol float64 // 0 to 1, the music
	sfxVol   float64 // 0 to 1, the sound effects

	effects map[string][]byte // decoded sound effects by cue name
	playing []*audio.Player   // effects that may still be playing (kept so they are not collected)
	music   *audio.Player
	current string // the music that was asked for last ("" for none)

	preloaded map[string][]byte // decoded music kept by Preload, so a track can start without a hitch

	// A fade-out in progress (FadeOut): fadeLeft counts ticks down to 0.
	fadeLeft, fadeTotal int
	fadeEased           bool // the fade follows a smooth S curve (FadeOutEased)

	missing map[string]bool
	Log     []string // every request, like "sfx stamp_thunk" or "music title"
}

// New makes a player reading files from fsys. muted gives a player that only
// records requests (test mode). If the computer has no audio device the
// player mutes itself instead of stopping the game.
func New(fsys fs.FS, muted bool) *Player {
	p := &Player{fsys: fsys, muted: muted, musicVol: 1, sfxVol: 1, effects: map[string][]byte{}, preloaded: map[string][]byte{}, missing: map[string]bool{}}
	if !muted {
		// Ebitengine allows one audio context per program.
		p.ctx = audio.CurrentContext()
		if p.ctx == nil {
			p.ctx = audio.NewContext(SampleRate)
		}
	}
	return p
}

// MusicVolume and SFXVolume are the two volumes, 0 to 1 (Settings screen).
func (p *Player) MusicVolume() float64 { return p.musicVol }
func (p *Player) SFXVolume() float64   { return p.sfxVol }

// SetMusicVolume sets the music volume (clamped to 0..1) and applies it to
// the music that is playing.
func (p *Player) SetMusicVolume(v float64) {
	p.musicVol = min(max(v, 0), 1)
	if p.music != nil {
		p.music.SetVolume(p.musicVol * p.fadeGain())
	}
}

// SetSFXVolume sets the volume of the sound effects played from now on.
func (p *Player) SetSFXVolume(v float64) { p.sfxVol = min(max(v, 0), 1) }

// Current is the music asked for last ("" when it was stopped).
func (p *Player) Current() string { return p.current }

// Missing lists the cues that were asked for but have no file, sorted.
func (p *Player) Missing() []string {
	var names []string
	for name := range p.missing {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// find reads the file of a cue, trying each extension.
func (p *Player) find(name string) (data []byte, ext string, ok bool) {
	for _, ext := range extensions {
		data, err := fs.ReadFile(p.fsys, name+ext)
		if err == nil {
			return data, ext, true
		}
	}
	return nil, "", false
}

// decode turns an ogg, mp3 or wav file into raw 16-bit stereo samples at
// SampleRate.
func decode(data []byte, ext string) ([]byte, error) {
	r := bytes.NewReader(data)
	var stream io.Reader
	var err error
	switch ext {
	case ".ogg":
		stream, err = vorbis.DecodeWithSampleRate(SampleRate, r)
	case ".mp3":
		stream, err = mp3.DecodeWithSampleRate(SampleRate, r)
	case ".wav":
		stream, err = wav.DecodeWithSampleRate(SampleRate, r)
	default:
		return nil, fmt.Errorf("unknown sound format %q", ext)
	}
	if err != nil {
		return nil, fmt.Errorf("decoding: %w", err)
	}
	return io.ReadAll(stream)
}

// PlaySFX plays a short sound once at full gain.
func (p *Player) PlaySFX(name string) { p.PlaySFXGain(name, 1) }

// Note adds a line to the log that is not a sound request ("note snore P-001
// room variant 2 gain 0.25"), so the test dump can show why a sound played.
func (p *Player) Note(text string) { p.Log = append(p.Log, "note "+text) }

// PlaySFXGain plays a short sound once, scaled by gain (0 to 1) on top of the
// SFX volume. Several can play at the same time.
func (p *Player) PlaySFXGain(name string, gain float64) {
	p.Log = append(p.Log, "sfx "+name)
	if p.muted || p.ctx == nil {
		return
	}
	pcm, ok := p.effects[name]
	if !ok {
		data, ext, found := p.find(name)
		if !found {
			p.missing[name] = true
			return
		}
		var err error
		if pcm, err = decode(data, ext); err != nil {
			p.missing[name] = true // a broken file counts as missing: the game goes on
			return
		}
		p.effects[name] = pcm
	}
	pl := p.ctx.NewPlayerFromBytes(pcm)
	pl.SetVolume(p.sfxVol * min(max(gain, 0), 1))
	pl.Play()
	// Forget the effects that have finished.
	live := p.playing[:0]
	for _, old := range p.playing {
		if old.IsPlaying() {
			live = append(live, old)
		}
	}
	p.playing = append(live, pl)
}

// PlayMusic starts a track, replacing the one that is playing. loop repeats
// it; start skips the quiet part of a track (the ending rock enters
// "mid-riff"). Asking for the track that is already playing does nothing.
func (p *Player) PlayMusic(name string, loop bool, start time.Duration) {
	p.Log = append(p.Log, "music "+name)
	if name == p.current && p.music != nil && p.music.IsPlaying() {
		return
	}
	p.stop()
	p.current = name
	if p.muted || p.ctx == nil {
		return
	}
	pcm, ok := p.preloaded[name]
	delete(p.preloaded, name)
	if !ok {
		data, ext, found := p.find(name)
		if !found {
			p.missing[name] = true
			return
		}
		var err error
		if pcm, err = decode(data, ext); err != nil {
			p.missing[name] = true
			return
		}
	}
	bytesPerSecond := SampleRate * 4 // 16-bit stereo
	skip := int(start.Seconds()*float64(bytesPerSecond)) / 4 * 4
	skip = min(skip, len(pcm))
	var src io.Reader = bytes.NewReader(pcm[skip:])
	if loop {
		src = audio.NewInfiniteLoop(bytes.NewReader(pcm[skip:]), int64(len(pcm)-skip))
	}
	pl, err := p.ctx.NewPlayer(src)
	if err != nil {
		p.missing[name] = true
		return
	}
	pl.SetVolume(p.musicVol)
	pl.Play()
	p.music = pl
}

// Preload decodes a track ahead of time so PlayMusic can start it at once.
// Decoding a minute of mp3 takes a moment; the ending does it during its
// beat of silence. Nothing is recorded in Log: it is not a request to play.
func (p *Player) Preload(name string) {
	if p.muted || p.ctx == nil || p.preloaded[name] != nil {
		return
	}
	data, ext, found := p.find(name)
	if !found {
		return // PlayMusic will report it as missing
	}
	if pcm, err := decode(data, ext); err == nil {
		p.preloaded[name] = pcm
	}
}

// FadeOut lowers the music to silence over the given number of ticks (one
// per frame) and then stops it. Update must be called every frame for it to run.
func (p *Player) FadeOut(ticks int) {
	p.Log = append(p.Log, "music fade")
	p.fadeLeft, p.fadeTotal, p.fadeEased = ticks, ticks, false
}

// FadeOutEased is FadeOut with a smooth start and end (EasedFadeGain), for the
// music that gives way to the next scene.
func (p *Player) FadeOutEased(ticks int) {
	p.FadeOut(ticks)
	p.fadeEased = true
}

// fadeGain is the fade's multiplier on the music volume right now.
func (p *Player) fadeGain() float64 {
	if p.fadeEased {
		return EasedFadeGain(p.fadeLeft, p.fadeTotal)
	}
	return FadeGain(p.fadeLeft, p.fadeTotal)
}

// MusicGain is how loud the music plays right now: the volume times the fade.
// The test dump shows it.
func (p *Player) MusicGain() float64 {
	if p.current == "" {
		return 0
	}
	return p.musicVol * p.fadeGain()
}

// Fading reports whether a fade-out is in progress.
func (p *Player) Fading() bool { return p.fadeLeft > 0 }

// FadeGain is how loud the music is, from 1 down to 0, with left ticks of a
// fade of total ticks remaining.
func FadeGain(left, total int) float64 {
	if total <= 0 {
		return 1
	}
	return min(max(float64(left)/float64(total), 0), 1)
}

// EasedFadeGain is FadeGain with smoothstep applied: it leaves 1 slowly,
// drops fastest in the middle and lands softly on 0, so there is no corner at
// either end of the fade.
func EasedFadeGain(left, total int) float64 {
	x := FadeGain(left, total)
	return x * x * (3 - 2*x)
}

// Update moves a fade-out one tick forward. Call it once per frame.
func (p *Player) Update() {
	if p.fadeLeft <= 0 {
		return
	}
	p.fadeLeft--
	if p.music != nil {
		p.music.SetVolume(p.musicVol * p.fadeGain())
	}
	if p.fadeLeft == 0 {
		p.stop()
	}
}

// StopMusic cuts the music (a script's "music off").
func (p *Player) StopMusic() {
	p.Log = append(p.Log, "music off")
	p.stop()
}

func (p *Player) stop() {
	p.fadeLeft, p.fadeTotal, p.fadeEased = 0, 0, false
	if p.music != nil {
		p.music.Pause()
		p.music = nil
	}
	p.current = ""
}

// FileName is the cue's file name when it exists, for the tests and the dump.
func (p *Player) FileName(name string) string {
	if _, ext, ok := p.find(name); ok {
		return path.Base(name + ext)
	}
	return ""
}

// Length decodes a cue's file and says how long it plays, for the tests: a
// file that is empty or broken shows up as an error or 0 s.
func (p *Player) Length(name string) (time.Duration, error) {
	data, ext, ok := p.find(name)
	if !ok {
		return 0, fmt.Errorf("no file for %q", name)
	}
	pcm, err := decode(data, ext)
	if err != nil {
		return 0, fmt.Errorf("%s%s: %w", name, ext, err)
	}
	return time.Duration(len(pcm)/4) * time.Second / SampleRate, nil
}
