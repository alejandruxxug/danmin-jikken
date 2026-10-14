package sound

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"
)

// wavBytes builds a tiny 16-bit mono wav of n samples of silence.
func wavBytes(n int) []byte {
	var b bytes.Buffer
	data := make([]byte, n*2)
	b.WriteString("RIFF")
	binary.Write(&b, binary.LittleEndian, uint32(36+len(data)))
	b.WriteString("WAVEfmt ")
	binary.Write(&b, binary.LittleEndian, uint32(16))
	binary.Write(&b, binary.LittleEndian, uint16(1))     // PCM
	binary.Write(&b, binary.LittleEndian, uint16(1))     // mono
	binary.Write(&b, binary.LittleEndian, uint32(44100)) // rate
	binary.Write(&b, binary.LittleEndian, uint32(88200)) // bytes per second
	binary.Write(&b, binary.LittleEndian, uint16(2))     // block align
	binary.Write(&b, binary.LittleEndian, uint16(16))    // bits
	b.WriteString("data")
	binary.Write(&b, binary.LittleEndian, uint32(len(data)))
	b.Write(data)
	return b.Bytes()
}

func TestDecodeWav(t *testing.T) {
	pcm, err := decode(wavBytes(100), ".wav")
	if err != nil {
		t.Fatal(err)
	}
	// Mono is turned into stereo: 100 samples are 400 bytes.
	if len(pcm) != 400 {
		t.Errorf("decoded %d bytes, want 400", len(pcm))
	}
	if _, err := decode([]byte("nope"), ".wav"); err == nil {
		t.Error("a broken file must be an error")
	}
	if _, err := decode(nil, ".xyz"); err == nil {
		t.Error("an unknown format must be an error")
	}
}

func TestMutedPlayerRecordsRequests(t *testing.T) {
	p := New(fstest.MapFS{"stamp_thunk.wav": {Data: wavBytes(10)}}, true)
	p.PlaySFX("stamp_thunk")
	p.PlayMusic("title", true, 0)
	if p.Current() != "title" {
		t.Errorf("current music = %q", p.Current())
	}
	p.StopMusic()
	want := []string{"sfx stamp_thunk", "music title", "music off"}
	if len(p.Log) != len(want) {
		t.Fatalf("log = %v, want %v", p.Log, want)
	}
	for i := range want {
		if p.Log[i] != want[i] {
			t.Errorf("log[%d] = %q, want %q", i, p.Log[i], want[i])
		}
	}
	if p.Current() != "" {
		t.Error("StopMusic must clear the current track")
	}
}

func TestFindTriesEachExtensionAndReportsMissing(t *testing.T) {
	p := New(fstest.MapFS{"a.mp3": {Data: []byte("x")}, "b.wav": {Data: []byte("x")}}, true)
	if got := p.FileName("a"); got != "a.mp3" {
		t.Errorf("a = %q", got)
	}
	if got := p.FileName("b"); got != "b.wav" {
		t.Errorf("b = %q", got)
	}
	if got := p.FileName("c"); got != "" {
		t.Errorf("c = %q, want no file", got)
	}
}

func TestVolumeIsClamped(t *testing.T) {
	p := New(fstest.MapFS{}, true)
	for _, tt := range []struct{ in, want float64 }{{-1, 0}, {0.4, 0.4}, {3, 1}} {
		p.SetMusicVolume(tt.in)
		if p.MusicVolume() != tt.want {
			t.Errorf("SetMusicVolume(%v) = %v, want %v", tt.in, p.MusicVolume(), tt.want)
		}
	}
}

func TestFadeGain(t *testing.T) {
	cases := []struct {
		left, total int
		want        float64
	}{
		{240, 240, 1}, {120, 240, 0.5}, {0, 240, 0}, {-3, 240, 0}, {300, 240, 1}, {5, 0, 1},
	}
	for _, c := range cases {
		if got := FadeGain(c.left, c.total); got != c.want {
			t.Errorf("FadeGain(%d, %d) = %v, want %v", c.left, c.total, got, c.want)
		}
	}
}

// A fade-out runs for its ticks and then the music is stopped.
func TestFadeOutStopsTheMusic(t *testing.T) {
	p := New(fstest.MapFS{"rock.wav": {Data: wavBytes(10)}}, true)
	p.PlayMusic("rock", false, 0)
	p.FadeOut(3)
	if !p.Fading() {
		t.Fatal("not fading after FadeOut")
	}
	for range 3 {
		p.Update()
	}
	if p.Fading() || p.Current() != "" {
		t.Errorf("after the fade: fading=%v current=%q", p.Fading(), p.Current())
	}
	if p.Log[len(p.Log)-1] != "music fade" {
		t.Errorf("log = %v", p.Log)
	}
}

func TestSFXGainIsClampedAndLogged(t *testing.T) {
	p := New(fstest.MapFS{"snore.wav": {Data: wavBytes(10)}}, true)
	p.PlaySFXGain("snore", 7) // a muted player records it and plays nothing
	if len(p.Log) != 1 || p.Log[0] != "sfx snore" {
		t.Errorf("log = %v", p.Log)
	}
}

// The eased fade has no corner at either end and never jumps.
func TestEasedFadeGain(t *testing.T) {
	if EasedFadeGain(10, 10) != 1 || EasedFadeGain(0, 10) != 0 || EasedFadeGain(5, 10) != 0.5 {
		t.Errorf("ends %v %v, middle %v", EasedFadeGain(10, 10), EasedFadeGain(0, 10), EasedFadeGain(5, 10))
	}
	if EasedFadeGain(9, 10) <= FadeGain(9, 10) { // slower than linear at the start
		t.Error("an eased fade should leave 1 more slowly than a linear one")
	}
	prev := 1.0
	for left := 71; left >= 0; left-- {
		g := EasedFadeGain(left, 72)
		if g > prev || prev-g > 0.1 {
			t.Fatalf("left %d: %v after %v", left, g, prev)
		}
		prev = g
	}
}

// Override plays a file from disk in place of the embedded one.
func TestOverrideReadsADiskFile(t *testing.T) {
	embedded := fstest.MapFS{"rock.wav": {Data: wavBytes(44100)}} // 1 s
	p := New(embedded, true)
	path := filepath.Join(t.TempDir(), "other.wav")
	if err := os.WriteFile(path, wavBytes(44100*3), 0o644); err != nil { // 3 s
		t.Fatal(err)
	}
	p.Override("rock", path)
	d, err := p.Length("rock")
	if err != nil || d != 3*time.Second {
		t.Errorf("Length = %v, %v; want 3s from the disk file", d, err)
	}
	p.Override("rock", filepath.Join(t.TempDir(), "missing.wav"))
	if _, err := p.Length("rock"); err == nil {
		t.Error("a missing override file must not fall back silently")
	}
}
