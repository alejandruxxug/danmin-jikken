package game

import (
	"os"
	"strings"
	"testing"
)

func TestPowerOnPhases(t *testing.T) {
	if got := PowerOnTicks(); got < 44 || got > 52 {
		t.Errorf("the turn-on is %d ticks, want about 0.8 s (48)", got)
	}
	want := map[int]string{0: "dot", 5: "dot", 6: "line", 9: "line", 10: "open", 17: "open", 18: "settle", 47: "settle"}
	for tick, phase := range want {
		st, ok := PowerOnAt(tick)
		if !ok || st.Phase != phase {
			t.Errorf("tick %d = %+v %v, want phase %s", tick, st, ok, phase)
		}
	}
	if _, ok := PowerOnAt(-1); ok {
		t.Error("before the start there is no turn-on")
	}
	if _, ok := PowerOnAt(PowerOnTicks()); ok {
		t.Error("after the end there is no turn-on")
	}
	prev := -1.0
	for tick := 0; tick < PowerOnTicks(); tick++ {
		st, _ := PowerOnAt(tick)
		if st.Progress <= prev || st.Progress < 0 || st.Progress >= 1 {
			t.Fatalf("tick %d progress %v after %v: want a steady rise inside [0,1)", tick, st.Progress, prev)
		}
		prev = st.Progress
	}
}

// The shader hard-codes the phase boundaries as fractions of the progress.
func TestShaderBoundariesMatchTheTable(t *testing.T) {
	src, err := os.ReadFile("crt.kage")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []string{"dotEnd = 0.125", "lineEnd = 0.2083", "openEnd = 0.375"} {
		if !strings.Contains(string(src), c) {
			t.Errorf("crt.kage lacks %q: it must match powerOnPhases (6, 4, 8, 30 of 48)", c)
		}
	}
}

func TestPowerOnRunsOut(t *testing.T) {
	env := &Env{}
	env.StartPowerOn()
	for range PowerOnTicks() {
		env.tickPowerOn()
	}
	if env.PoweringOn {
		t.Error("the turn-on must end")
	}
}
