package game

import (
	"encoding/json"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestParseScript(t *testing.T) {
	script := `
# comment line
wait 30            # frames
key Space
click 212 140
shot after.png
dump after.json
`
	cmds, err := ParseScript(strings.NewReader(script))
	if err != nil {
		t.Fatal(err)
	}
	if len(cmds) != 5 {
		t.Fatalf("got %d commands, want 5: %+v", len(cmds), cmds)
	}
	if cmds[0].Op != "wait" || cmds[0].N != 30 {
		t.Errorf("cmd 0 = %+v", cmds[0])
	}
	if cmds[1].Op != "key" || cmds[1].Key != ebiten.KeySpace {
		t.Errorf("cmd 1 = %+v", cmds[1])
	}
	if cmds[2].Op != "click" || cmds[2].X != 212 || cmds[2].Y != 140 {
		t.Errorf("cmd 2 = %+v", cmds[2])
	}
	if cmds[3].Path != "after.png" || cmds[4].Path != "after.json" {
		t.Errorf("paths = %q, %q", cmds[3].Path, cmds[4].Path)
	}
}

func TestParseScriptErrors(t *testing.T) {
	bad := map[string]string{
		"unknown command": "jump 3",
		"wait no number":  "wait",
		"wait not number": "wait soon",
		"unknown key":     "key NoSuchKey",
		"click one arg":   "click 5",
		"click not nums":  "click a b",
		"shot no file":    "shot",
		"mouse one arg":   "mouse 5",
		"pan too far":     "pan 241",
		"pan negative":    "pan -1",
	}
	for name, line := range bad {
		if _, err := ParseScript(strings.NewReader("wait 1\n" + line)); err == nil {
			t.Errorf("%s: %q should fail", name, line)
		} else if !strings.Contains(err.Error(), "line 2") {
			t.Errorf("%s: error %q should name line 2", name, err)
		}
	}
}

func TestParseScriptFixture(t *testing.T) {
	f, err := os.Open("testdata/smoke.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := ParseScript(f); err != nil {
		t.Fatal(err)
	}
}

func TestParseOptions(t *testing.T) {
	o, err := ParseOptions([]string{"-agent", "-scene", "day.reception", "-day", "2", "-clock", "15:40",
		"-seed", "7", "-flags", "a,b", "-speed", "4", "-nocrt", "-quit", "-frames", "30"})
	if err != nil {
		t.Fatal(err)
	}
	if !o.Enabled || o.Scene != "day.reception" || o.Day != 2 || o.Clock != "15:40" ||
		o.Seed != 7 || o.Flags != "a,b" || o.Speed != 4 || !o.NoCRT || !o.Quit || o.Frames != 30 {
		t.Errorf("options = %+v", o)
	}
}

func TestOptionsIgnoredWithoutAgent(t *testing.T) {
	o, err := ParseOptions([]string{"-scene", "day.reception", "-day", "3", "-quit"})
	if err != nil {
		t.Fatal(err)
	}
	if o != (Options{}) {
		t.Errorf("without -agent every option must be ignored, got %+v", o)
	}
}

func TestParseClock(t *testing.T) {
	c, err := ParseClock("03:17")
	if err != nil {
		t.Fatal(err)
	}
	if c.String() != "03:17" {
		t.Errorf("clock = %s", c)
	}
	c.Advance(60 * 21) // past midnight wraps
	if c.String() != "00:17" {
		t.Errorf("after wrap clock = %s, want 00:17", c)
	}
	for _, bad := range []string{"", "9", "25:00", "10:61", "ab:cd"} {
		if _, err := ParseClock(bad); err == nil {
			t.Errorf("ParseClock(%q) should fail", bad)
		}
	}
}

func TestBuildDumpWithoutWard(t *testing.T) {
	env := &Env{Day: 2, Seed: 7, Clock: NewClock(15, 40),
		Flags: map[string]bool{"b": true, "a": true, "off": false}, Scores: map[string]int{"served": 3}}
	d := BuildDump(env, SceneTitle, 12)
	if d.Scene != "title" || d.Day != 2 || d.Clock != "15:40" || d.Frame != 12 {
		t.Errorf("dump = %+v", d)
	}
	if strings.Join(d.Flags, ",") != "a,b" {
		t.Errorf("flags = %v, want sorted and without 'off'", d.Flags)
	}
	if d.Patients == nil || d.Rooms == nil {
		t.Error("patients and rooms must be empty lists, not null, in the json")
	}
}

func TestBuildDumpWithWard(t *testing.T) {
	w, err := NewWard(7, 1)
	if err != nil {
		t.Fatal(err)
	}
	env := &Env{Day: 1, Seed: 7, Clock: NewClock(9, 0), Flags: map[string]bool{}, Scores: map[string]int{}, ward: w}
	d := BuildDump(env, SceneDayReception, 1)
	if len(d.Patients) != 5 || len(d.Rooms) != 3 {
		t.Fatalf("got %d patients and %d rooms, want 5 and 3", len(d.Patients), len(d.Rooms))
	}
	if d.Patients[0].State != "awake" || d.Patients[0].Level != "severe" {
		t.Errorf("patient 0 = %+v", d.Patients[0])
	}
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var back Dump
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatal(err)
	}
	if back.Rooms[0].Number != 101 || back.Rooms[0].State != "available" {
		t.Errorf("room 0 = %+v", back.Rooms[0])
	}
}

func TestWriteDump(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	d := BuildDump(&Env{Clock: NewClock(9, 0), Flags: map[string]bool{}, Scores: map[string]int{}}, SceneTitle, 1)
	if err := writeDump(path, d); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(data) {
		t.Errorf("not valid json: %s", data)
	}
}

func TestScaleUp(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 2, 1))
	red := color.RGBA{255, 0, 0, 255}
	blue := color.RGBA{0, 0, 255, 255}
	src.SetRGBA(0, 0, red)
	src.SetRGBA(1, 0, blue)
	dst := scaleUp(src, 4)
	if dst.Bounds().Dx() != 8 || dst.Bounds().Dy() != 4 {
		t.Fatalf("size = %v, want 8x4", dst.Bounds())
	}
	if dst.RGBAAt(3, 3) != red || dst.RGBAAt(4, 0) != blue {
		t.Error("pixels were not repeated 4 times each way")
	}
}
