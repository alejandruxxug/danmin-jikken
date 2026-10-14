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

	missing map[string]bool
	Log     []string // every request, like "sfx stamp_thunk" or "music title"
}

// New makes a player reading files from fsys. muted gives a player that only
// records requests (test mode). If the computer has no audio device the
// player mutes itself instead of stopping the game.
func New(fsys fs.FS, muted bool) *Player {
	p := &Player{fsys: fsys, muted: muted, musicVol: 1, sfxVol: 1, effects: map[string][]byte{}, missing: map[string]bool{}}
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
		p.music.SetVolume(p.musicVol)
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

// PlaySFX plays a short sound once. Several can play at the same time.
func (p *Player) PlaySFX(name string) {
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
	pl.SetVolume(p.sfxVol)
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
	data, ext, found := p.find(name)
	if !found {
		p.missing[name] = true
		return
	}
	pcm, err := decode(data, ext)
	if err != nil {
		p.missing[name] = true
		return
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

// StopMusic cuts the music (a script's "music off").
func (p *Player) StopMusic() {
	p.Log = append(p.Log, "music off")
	p.stop()
}

func (p *Player) stop() {
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
