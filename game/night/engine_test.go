package night

import (
	"strings"
	"testing"
)

func newEngine(t *testing.T, src string, flags Flags) *Engine {
	t.Helper()
	e, err := NewEngine(mustParse(t, src), flags)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	return e
}

func TestTypewriterSkipThenAdvance(t *testing.T) {
	e := newEngine(t, "Mayu (smile): Hello there!\n~ Second.\n", Flags{})
	if e.LineDone() || e.ShownCount() != 0 {
		t.Fatal("a new line must start empty")
	}
	for range 3 {
		if err := e.Tick(); err != nil {
			t.Fatal(err)
		}
	}
	if n := e.ShownCount(); n == 0 || n >= len("Hello there!") {
		t.Fatalf("after 3 ticks %d letters shown, want some but not all", n)
	}
	// First click: the whole line appears. The story must not move on.
	if err := e.Advance(); err != nil {
		t.Fatal(err)
	}
	if !e.LineDone() || e.Shown() != "Hello there!" || e.Speaker() != "mayu" {
		t.Fatalf("after the first click: shown %q, speaker %q", e.Shown(), e.Speaker())
	}
	// Second click: the next line (a thought).
	if err := e.Advance(); err != nil {
		t.Fatal(err)
	}
	if !e.Thinking() || e.FullText() != "Second." {
		t.Fatalf("second click: thinking %v, text %q", e.Thinking(), e.FullText())
	}
	_ = e.Advance() // finish typing
	_ = e.Advance() // past the last line
	if _, ended := e.Ended(); !ended {
		t.Fatal("the script should have ended")
	}
}

func TestTypewriterSpeed(t *testing.T) {
	e := newEngine(t, "speed 60\n~ abcdefghij\n", Flags{})
	for range 5 {
		_ = e.Tick()
	}
	if e.ShownCount() != 5 {
		t.Errorf("at 60 letters/s, 5 ticks = %d letters, want 5", e.ShownCount())
	}
}

const menu = `
~ Pick.
choice
option Left | set went_left | jump left
option Right | set went_right | jump right
label left
Mayu (laugh): Left!
jump done
label right
Hana (flat): Right!
label done
end day
`

func TestChoiceSetsFlagAndJumps(t *testing.T) {
	flags := Flags{}
	e := newEngine(t, menu, flags)
	_ = e.Advance() // finish typing "Pick."
	_ = e.Advance() // on to the menu
	if len(e.Options()) != 2 {
		t.Fatalf("menu has %d options, want 2", len(e.Options()))
	}
	if err := e.Advance(); err != nil || len(e.Options()) != 2 {
		t.Fatal("a click must not answer a menu")
	}
	if err := e.Choose(1); err != nil {
		t.Fatal(err)
	}
	if !flags.Has("went_right") || flags.Has("went_left") {
		t.Errorf("flags = %v, want only went_right", flags.Names())
	}
	if e.Speaker() != "hana" || e.FullText() != "Right!" {
		t.Errorf("after choosing: %q says %q", e.Speaker(), e.FullText())
	}
	// The branch that was not taken never ran.
	if e.PoseOf("mayu") != startPose("mayu") {
		t.Error("the left branch changed Mayu's pose")
	}
}

func TestChooseErrors(t *testing.T) {
	e := newEngine(t, "~ x\n", Flags{})
	if err := e.Choose(0); err == nil {
		t.Error("Choose with no menu must fail")
	}
	e = newEngine(t, menu, Flags{})
	_ = e.Advance()
	_ = e.Advance()
	if err := e.Choose(5); err == nil {
		t.Error("Choose(5) on a 2-option menu must fail")
	}
}

func TestIfAndSetJumps(t *testing.T) {
	src := "set seen\nif seen jump yes\n~ no\nlabel yes\nifnot seen jump end\n~ yes\nlabel end\nend day\n"
	e := newEngine(t, src, Flags{})
	if e.FullText() != "yes" {
		t.Errorf("if/ifnot went to %q, want the yes line", e.FullText())
	}
	if !e.Flags().Has("seen") {
		t.Error("set did not set the flag")
	}
}

func TestJumpLoopIsAnErrorNotAFreeze(t *testing.T) {
	_, err := NewEngine(mustParse(t, "label a\njump a\n"), Flags{})
	if err == nil || !strings.Contains(err.Error(), "jump loop") {
		t.Errorf("error = %v, want a jump loop error", err)
	}
}

func TestWaitCanBeSkipped(t *testing.T) {
	e := newEngine(t, "wait 10\n~ after\n", Flags{})
	if e.HasLine() {
		t.Fatal("no line may show during a wait")
	}
	if err := e.Advance(); err != nil || e.FullText() != "after" {
		t.Fatalf("a click should end the wait, got %q (%v)", e.FullText(), err)
	}
	e = newEngine(t, "wait 0.5\n~ after\n", Flags{})
	for range 31 {
		_ = e.Tick()
	}
	if e.FullText() != "after" {
		t.Error("the wait should end by itself after 0.5 s")
	}
}

func TestPoseSelectsLayers(t *testing.T) {
	e := newEngine(t, "Mayu (blank): ...\nMayu (smile eyes=closed arml=wave): next\n", Flags{})
	want := []string{"hair_back", "body", "head", "eyes_away", "brows_flat", "mouth_flat", "fx_shadow", "arm_l_hug", "arm_r_hug"}
	if got := e.PoseOf("mayu").Layers(); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("blank layers = %v, want %v", got, want)
	}
	_ = e.Advance()
	_ = e.Advance()
	p := e.PoseOf("mayu")
	if p.Eyes != "closed" || p.ArmL != "wave" || p.Mouth != "smile" || p.FX != "" {
		t.Errorf("override pose = %+v", p)
	}
	if e.Visible() != "mayu" {
		t.Errorf("visible = %q, want mayu", e.Visible())
	}
}

func TestFacesKeepBothGirlsAndPoses(t *testing.T) {
	e := newEngine(t, "faces\nMayu (smile): a\nHana (soft): b\n", Flags{})
	if !e.Faces() || e.Visible() != "" {
		t.Fatalf("faces: %v visible %q", e.Faces(), e.Visible())
	}
	_ = e.Advance()
	_ = e.Advance()
	if e.PoseOf("hana").FX != "blush" || e.PoseOf("mayu").Mouth != "smile" {
		t.Error("both poses should be kept for the two faces")
	}
	if e.Visible() != "" {
		t.Error("a speaker must not step in over the faces")
	}
}

func TestFastForward(t *testing.T) {
	flags := Flags{"went_right": true}
	e := newEngine(t, menu, flags)
	// Line 10 is `Hana (flat): Right!` (the raw string starts with a blank line)
	if err := e.FastForward(10); err != nil {
		t.Fatal(err)
	}
	if e.Line() != 10 || e.Speaker() != "hana" {
		t.Errorf("at line %d, speaker %q; want line 10, hana", e.Line(), e.Speaker())
	}
}

func TestSnapshot(t *testing.T) {
	e := newEngine(t, "bg station\nprop can\nshow hana\nHana (tired): hi\n", Flags{})
	st := e.Snapshot()
	if st.BG != "station" || st.Visible != "hana" || st.Speaker != "hana" || st.Line != 4 {
		t.Errorf("snapshot = %+v", st)
	}
	if len(st.Props) != 1 || st.Props[0] != "can" || len(st.Layers) != 8 {
		t.Errorf("props %v layers %v", st.Props, st.Layers)
	}
}

func TestRealScriptsPlayThrough(t *testing.T) {
	for _, name := range []string{Prologue, Night1} {
		s, err := LoadScript(name)
		if err != nil {
			t.Fatal(err)
		}
		e, err := NewEngine(s, Flags{})
		if err != nil {
			t.Fatal(err)
		}
		for step := 0; step < 1000; step++ {
			if _, ended := e.Ended(); ended {
				break
			}
			if len(e.Options()) > 0 {
				_ = e.Choose(0)
			} else {
				_ = e.Advance()
				_ = e.Tick()
			}
		}
		if to, ended := e.Ended(); !ended || to == "" {
			t.Errorf("%s did not reach its end command", name)
		}
	}
}

func TestNight1BranchesSetTheirFlags(t *testing.T) {
	s, err := LoadScript(Night1)
	if err != nil {
		t.Fatal(err)
	}
	for i, flag := range []string{"liked_mayu", "liked_hana"} {
		flags := Flags{}
		e, _ := NewEngine(s, flags)
		for step := 0; step < 1000 && !isEnded(e); step++ {
			if len(e.Options()) > 0 {
				_ = e.Choose(i)
			} else {
				_ = e.Advance()
				_ = e.Tick()
			}
		}
		if names := flags.Names(); len(names) != 1 || names[0] != flag {
			t.Errorf("option %d set %v, want only %s", i+1, names, flag)
		}
	}
}

func isEnded(e *Engine) bool { _, ended := e.Ended(); return ended }

func TestWrapText(t *testing.T) {
	measure := func(s string) int { return len(s) * 5 }
	got := WrapText("one two three four", 45, measure)
	want := []string{"one two", "three", "four"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("WrapText = %q, want %q", got, want)
	}
	if got := WrapText("", 50, measure); len(got) != 0 {
		t.Errorf("empty text gave %q", got)
	}
}

func TestLampFlickers(t *testing.T) {
	on, off := 0, 0
	for tick := 0; tick < 60; tick++ {
		if LampOn(tick) {
			on++
		} else {
			off++
		}
	}
	if on == 0 || off == 0 || on <= off {
		t.Errorf("lamp on %d ticks, off %d: it should flicker but stay mostly on", on, off)
	}
	if ceilingOffset(0) != 0 || ceilingOffset(135) != 0 {
		t.Error("the ceiling must wrap every 270 px")
	}
}
