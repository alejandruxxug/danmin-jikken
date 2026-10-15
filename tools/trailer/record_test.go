package main

import (
	"crypto/sha256"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"danmingame"
)

// The same command must give the same frames and the same audio.json, CRT on.
// It builds the game and opens its window twice, so -short skips it.
func TestRecordIsDeterministic(t *testing.T) {
	if testing.Short() {
		t.Skip("opens the game window")
	}
	bin := filepath.Join(t.TempDir(), "danmin")
	build := exec.Command("go", "build", "-o", bin, "./cmd/danmin")
	build.Dir = "../../game"
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("building the game: %v\n%s", err, out)
	}
	script := filepath.Join(t.TempDir(), "in.txt")
	if err := os.WriteFile(script, []byte("wait 20\nstamp admit\nwait 40\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	record := func() [sha256.Size]byte {
		dir := t.TempDir()
		cmd := exec.Command(bin, "-agent", "-scene", "day.reception", "-day", "1", "-seed", "7",
			"-input", script, "-record", dir, "-record-frames", "40")
		cmd.Dir = "../../game"
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("recording: %v\n%s", err, out)
		}
		h := sha256.New()
		for n := 1; n <= 40; n++ {
			data, err := os.ReadFile(filepath.Join(dir, game.FrameFile(n)))
			if err != nil {
				t.Fatal(err)
			}
			h.Write(data)
		}
		audio, err := os.ReadFile(filepath.Join(dir, "audio.json"))
		if err != nil {
			t.Fatal(err)
		}
		h.Write(audio)
		var sum [sha256.Size]byte
		copy(sum[:], h.Sum(nil))
		return sum
	}
	if a, b := record(), record(); a != b {
		t.Error("two recordings of the same command differ")
	}
}
