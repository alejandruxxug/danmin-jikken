package game

import "testing"

func TestCRTShaderCompiles(t *testing.T) {
	if _, err := NewCRT(); err != nil {
		t.Fatal(err)
	}
}

func TestDefaultCRTAmounts(t *testing.T) {
	p := DefaultCRT(2, 1.5, 0)
	if p.Scanline != 0.15 || p.Vignette != 0.35 || p.Day != 2 || p.Time != 1.5 || p.Filmed != 1 {
		t.Errorf("params = %+v", p)
	}
	if got := DefaultCRT(9, 0, 0).Day; got != 3 {
		t.Errorf("day 9 clamps to %v, want 3", got)
	}
}

func TestUsesCRTByIdAndByScene(t *testing.T) {
	if !usesCRT("day.reception", &titleScene{}) || usesCRT("night.1", &titleScene{}) {
		t.Error("day ids use the CRT, night ids do not")
	}
	if usesCRT("day.reception", &crtOffScene{}) {
		t.Error("a scene that says no must win over its id")
	}
}

type crtOffScene struct{ titleScene }

func (*crtOffScene) UsesCRT() bool { return false }
