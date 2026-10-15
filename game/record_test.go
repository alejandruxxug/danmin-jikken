package game

import "testing"

func TestFrameNamesAndTimes(t *testing.T) {
	if got := FrameFile(1); got != "f000001.png" {
		t.Errorf("FrameFile(1) = %q", got)
	}
	if got := FrameFile(123456); got != "f123456.png" {
		t.Errorf("FrameFile(123456) = %q", got)
	}
	for _, c := range []struct {
		frame int
		want  float64
	}{{1, 0}, {2, 1.0 / 60}, {61, 1}, {181, 3}} {
		if got := FrameSeconds(c.frame); got != c.want {
			t.Errorf("FrameSeconds(%d) = %v, want %v", c.frame, got, c.want)
		}
	}
}

func TestParseCardSpec(t *testing.T) {
	for _, c := range []struct{ spec, sheet, layer string }{
		{"logo", DefaultCardSheet, "logo"},
		{"cards2:hiring", "cards2", "hiring"},
	} {
		sheet, layer, err := ParseCardSpec(c.spec)
		if err != nil || sheet != c.sheet || layer != c.layer {
			t.Errorf("ParseCardSpec(%q) = %q, %q, %v", c.spec, sheet, layer, err)
		}
	}
	for _, bad := range []string{"", "a:", ":b"} {
		if _, _, err := ParseCardSpec(bad); err == nil {
			t.Errorf("ParseCardSpec(%q) should fail", bad)
		}
	}
}

func TestRecordNeedsAgent(t *testing.T) {
	if _, err := ParseOptions([]string{"-record", "x"}); err == nil {
		t.Error("-record without -agent should fail")
	}
	o, err := ParseOptions([]string{"-agent", "-record", "x", "-record-frames", "30"})
	if err != nil || o.Frames != 30 || !o.Quit {
		t.Errorf("got %+v, %v", o, err)
	}
}
