package sound

import (
	"bytes"
	"encoding/binary"
	"testing"
	"testing/fstest"
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
		p.SetVolume(tt.in)
		if p.Volume() != tt.want {
			t.Errorf("SetVolume(%v) = %v, want %v", tt.in, p.Volume(), tt.want)
		}
	}
}
