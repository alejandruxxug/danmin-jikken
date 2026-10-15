package main

// The audio mixer. It reads audio.json (the sound events the game recorded,
// each with its frame number) and the sound files, and places everything on
// one 48 kHz stereo timeline: sound effects at their frame, the music with its
// start offset, seeks, loop, gain changes (the game's fades) and stops. The
// result is soft-limited and written as a wav.

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"

	"danmingame/sound"
)

const (
	SampleRate       = 48000
	FPS              = 60
	SamplesPerFrame  = SampleRate / FPS // 800: a frame is a whole number of samples
	bytesPerSample16 = 2
)

// Audio is stereo audio: left and right samples interleaved, from -1 to 1.
type Audio struct {
	Samples []float32
}

// Frames is the number of sample pairs.
func (a Audio) Frames() int { return len(a.Samples) / 2 }

// Loader finds and decodes a sound file by name (as written in audio.json).
type Loader func(file string) (Audio, error)

// frameToSample is the sample where an event of this frame starts. Frame 1 is
// the first picture, shown at 0 s, so an event of frame n starts at (n-1)/60 s.
// The game also logs events from before its first update as frame 0: they go
// to the very start too.
func frameToSample(frame int) int {
	return max(frame-1, 0) * SamplesPerFrame
}

// track is one run of music: from a "music" event to the next "music" or
// "stop" event (or the end). Seeks add parts; gain events make the envelope.
type track struct {
	file     string
	loop     bool
	loopFrom int // sample of the file where the loop starts (the play offset)
	start    int // timeline sample where it starts
	end      int // timeline sample where it stops
	parts    []part
	gains    []gainPoint
	k        int // gainAt's cursor: the index of the last gain point reached, -1 before the first
}

// part says "from this timeline sample the file plays from this sample".
type part struct{ at, pos int }

type gainPoint struct {
	at   int
	gain float64
}

// buildTracks turns the music events into tracks. total is the length of the
// timeline in samples.
func buildTracks(events []sound.Event, total int) []*track {
	var tracks []*track
	var cur *track
	closeCur := func(at int) {
		if cur != nil {
			cur.end = min(at, total)
			cur = nil
		}
	}
	for _, e := range events {
		at := frameToSample(e.Frame)
		switch e.Kind {
		case "music":
			closeCur(at)
			off := int(math.Round(e.Offset * SampleRate))
			cur = &track{file: e.File, loop: e.Loop, loopFrom: off, start: at, end: total, parts: []part{{at, off}}, k: -1}
			tracks = append(tracks, cur)
		case "seek":
			if cur != nil {
				cur.parts = append(cur.parts, part{at, int(math.Round(e.Offset * SampleRate))})
			}
		case "gain":
			if cur != nil {
				cur.gains = append(cur.gains, gainPoint{at, e.Gain})
			}
		case "stop":
			closeCur(at)
		}
	}
	return tracks
}

// gainAt is the track's gain at a sample: the last gain event at or before it,
// (call it with growing i) reached by a straight ramp over one frame from the one before (the game
// changes the volume once per frame; the ramp keeps the steps from clicking).
// The first event is the starting gain, with no ramp. Before any event: 1.
func (t *track) gainAt(i int) float64 {
	for t.k+1 < len(t.gains) && t.gains[t.k+1].at <= i {
		t.k++ // the samples are asked for in order, so the cursor only moves forward
	}
	k := t.k
	switch {
	case k < 0:
		return 1
	case k == 0 || i-t.gains[k].at >= SamplesPerFrame:
		return t.gains[k].gain
	}
	f := float64(i-t.gains[k].at) / SamplesPerFrame
	return t.gains[k-1].gain + (t.gains[k].gain-t.gains[k-1].gain)*f
}

// filePos is the sample of the file that plays at timeline sample i, from the
// part that is active then. ok is false past the end of a file that does not loop.
func (t *track) filePos(i, fileLen int) (pos int, ok bool) {
	p := t.parts[0]
	for _, q := range t.parts {
		if q.at <= i {
			p = q
		}
	}
	pos = p.pos + (i - p.at)
	if pos >= fileLen {
		if !t.loop || fileLen <= t.loopFrom {
			return 0, false
		}
		pos = t.loopFrom + (pos-t.loopFrom)%(fileLen-t.loopFrom)
	}
	return pos, pos >= 0
}

// MixOptions says what to leave out of the mix.
type MixOptions struct {
	DropMusic bool // a trailer-level music bed replaces the game's own music
}

// Mix places every event of the log on a timeline of frames frames and
// returns it (not limited yet). Files the loader cannot find are an error: a
// silent hole in a trailer should not go unnoticed. A cue with no file in the
// game (File "") is silent in the game too, so it is skipped.
func Mix(log []sound.Event, frames int, load Loader, opt MixOptions) (Audio, error) {
	total := frames * SamplesPerFrame
	out := Audio{Samples: make([]float32, total*2)}
	cache := map[string]Audio{}
	get := func(file string) (Audio, error) {
		if a, ok := cache[file]; ok {
			return a, nil
		}
		a, err := load(file)
		if err != nil {
			return Audio{}, err
		}
		cache[file] = a
		return a, nil
	}
	for _, e := range log {
		if e.Kind != "sfx" || e.File == "" {
			continue
		}
		clip, err := get(e.File)
		if err != nil {
			return Audio{}, err
		}
		addAt(out, clip, frameToSample(e.Frame), float32(e.Gain))
	}
	if opt.DropMusic {
		return out, nil
	}
	for _, t := range buildTracks(log, total) {
		if t.file == "" {
			continue
		}
		clip, err := get(t.file)
		if err != nil {
			return Audio{}, err
		}
		for i := t.start; i < t.end && i < total; i++ {
			pos, ok := t.filePos(i, clip.Frames())
			if !ok {
				continue
			}
			g := float32(t.gainAt(i))
			out.Samples[2*i] += clip.Samples[2*pos] * g
			out.Samples[2*i+1] += clip.Samples[2*pos+1] * g
		}
	}
	return out, nil
}

// addAt adds clip, scaled by gain, into dst starting at sample start; what
// runs past the end of dst is dropped.
func addAt(dst, clip Audio, start int, gain float32) {
	n := min(clip.Frames(), dst.Frames()-start)
	for i := 0; i < n; i++ {
		dst.Samples[2*(start+i)] += clip.Samples[2*i] * gain
		dst.Samples[2*(start+i)+1] += clip.Samples[2*i+1] * gain
	}
}

// SoftLimit squeezes a sample into -1..1: unchanged up to the knee, then a
// smooth curve that never passes 1, so a loud pile-up rounds off instead of
// clipping.
func SoftLimit(x float32) float32 {
	const knee = 0.8
	a := float64(x)
	if math.Abs(a) <= knee {
		return x
	}
	over := (math.Abs(a) - knee) / (1 - knee)
	return float32(math.Copysign(knee+(1-knee)*math.Tanh(over), a))
}

// Stats are the peak and the RMS of audio, to check a mix without listening.
type Stats struct {
	Peak, RMS float64
}

func Measure(a Audio) Stats {
	var peak, sum float64
	for _, s := range a.Samples {
		v := math.Abs(float64(s))
		peak = math.Max(peak, v)
		sum += v * v
	}
	if len(a.Samples) == 0 {
		return Stats{}
	}
	return Stats{Peak: peak, RMS: math.Sqrt(sum / float64(len(a.Samples)))}
}

// dB is a level as decibels relative to full scale, "-inf" for silence.
func dB(v float64) string {
	if v <= 0 {
		return "-inf"
	}
	return fmt.Sprintf("%.1f", 20*math.Log10(v))
}

func (s Stats) String() string {
	return fmt.Sprintf("peak %s dBFS, rms %s dBFS", dB(s.Peak), dB(s.RMS))
}

// WriteWAV writes 16-bit stereo 48 kHz PCM, soft-limited.
func WriteWAV(path string, a Audio) error {
	data := make([]byte, 0, len(a.Samples)*bytesPerSample16)
	for _, s := range a.Samples {
		v := int16(math.Round(float64(SoftLimit(s)) * 32767))
		data = binary.LittleEndian.AppendUint16(data, uint16(v))
	}
	var b bytes.Buffer
	b.WriteString("RIFF")
	binary.Write(&b, binary.LittleEndian, uint32(36+len(data)))
	b.WriteString("WAVEfmt ")
	binary.Write(&b, binary.LittleEndian, uint32(16))
	binary.Write(&b, binary.LittleEndian, uint16(1)) // PCM
	binary.Write(&b, binary.LittleEndian, uint16(2)) // stereo
	binary.Write(&b, binary.LittleEndian, uint32(SampleRate))
	binary.Write(&b, binary.LittleEndian, uint32(SampleRate*2*bytesPerSample16))
	binary.Write(&b, binary.LittleEndian, uint16(2*bytesPerSample16))
	binary.Write(&b, binary.LittleEndian, uint16(16))
	b.WriteString("data")
	binary.Write(&b, binary.LittleEndian, uint32(len(data)))
	b.Write(data)
	return os.WriteFile(path, b.Bytes(), 0o644)
}

// DiskLoader reads files from the game's sound folder and decodes them to
// 48 kHz stereo with the decoders the game itself uses (Ebitengine's).
func DiskLoader(soundDir string) Loader {
	return func(file string) (Audio, error) {
		path := file
		if !filepath.IsAbs(path) {
			path = filepath.Join(soundDir, file)
		}
		return DecodeFile(path)
	}
}

// ResolveSound finds a sound by name with or without its extension, in the
// sound folder, or as a path of its own.
func ResolveSound(soundDir, name string) (string, error) {
	candidates := []string{name}
	if filepath.Ext(name) == "" {
		for _, ext := range []string{".ogg", ".mp3", ".wav"} {
			candidates = append(candidates, name+ext)
		}
	}
	for _, c := range candidates {
		for _, p := range []string{filepath.Join(soundDir, c), c} {
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				return filepath.Abs(p)
			}
		}
	}
	return "", fmt.Errorf("no sound file %q in %s", name, soundDir)
}

// DecodeFile decodes an ogg, mp3 or wav file to 48 kHz stereo.
func DecodeFile(path string) (Audio, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Audio{}, err
	}
	r := bytes.NewReader(raw)
	var stream io.Reader
	switch strings.ToLower(filepath.Ext(path)) {
	case ".ogg":
		stream, err = vorbis.DecodeWithSampleRate(SampleRate, r)
	case ".mp3":
		stream, err = mp3.DecodeWithSampleRate(SampleRate, r)
	case ".wav":
		stream, err = wav.DecodeWithSampleRate(SampleRate, r)
	default:
		return Audio{}, fmt.Errorf("%s: unknown sound format", path)
	}
	if err != nil {
		return Audio{}, fmt.Errorf("%s: %w", path, err)
	}
	pcm, err := io.ReadAll(stream) // 16-bit little-endian stereo
	if err != nil {
		return Audio{}, fmt.Errorf("%s: %w", path, err)
	}
	a := Audio{Samples: make([]float32, len(pcm)/2)}
	for i := range a.Samples {
		a.Samples[i] = float32(int16(binary.LittleEndian.Uint16(pcm[2*i:]))) / 32768
	}
	return a, nil
}
