package game

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// TestSmokeAgentRun builds the real binary and runs it through the backdoor.
// It opens a window, so it needs a display: it is skipped with -short and on
// Linux without DISPLAY or WAYLAND_DISPLAY.
func TestSmokeAgentRun(t *testing.T) {
	if testing.Short() {
		t.Skip("opens a window")
	}
	if runtime.GOOS == "linux" && os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
		t.Skip("no display")
	}
	dir := t.TempDir()
	bin := filepath.Join(dir, "danmin")
	if out, err := exec.Command("go", "build", "-o", bin, "./cmd/danmin").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}

	dump := filepath.Join(dir, "state.json")
	shot := filepath.Join(dir, "shot.png")
	cmd := exec.Command(bin, "-agent", "-scene", "day.reception", "-seed", "7",
		"-frames", "30", "-dump", dump, "-shot", shot, "-quit")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("run: %v\n%s", err, out)
	}

	data, err := os.ReadFile(dump)
	if err != nil {
		t.Fatal(err)
	}
	var d Dump
	if err := json.Unmarshal(data, &d); err != nil {
		t.Fatal(err)
	}
	if d.Scene != SceneDayReception || d.Seed != 7 || len(d.Patients) != 5 {
		t.Errorf("dump = %+v", d)
	}
	if _, err := os.Stat(shot); err != nil {
		t.Errorf("no screenshot: %v", err)
	}
}

// The input script from testdata drives keys, a click and two outputs.
func TestSmokeInputScript(t *testing.T) {
	if testing.Short() {
		t.Skip("opens a window")
	}
	if runtime.GOOS == "linux" && os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
		t.Skip("no display")
	}
	dir := t.TempDir()
	bin := filepath.Join(dir, "danmin")
	if out, err := exec.Command("go", "build", "-o", bin, "./cmd/danmin").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	script, err := filepath.Abs("testdata/smoke.txt")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(bin, "-agent", "-scene", "day.reception", "-seed", "7", "-input", script, "-quit")
	cmd.Dir = dir // the script writes its shot and dump next to the binary
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("run: %v\n%s", err, out)
	}
	for _, name := range []string{"smoke_shot.png", "smoke_dump.json"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("script output missing: %v", err)
		}
	}
}
