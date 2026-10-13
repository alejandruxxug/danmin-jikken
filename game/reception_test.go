package game

import (
	"strings"
	"testing"
)

func TestParseScriptMouseAndPan(t *testing.T) {
	cmds, err := ParseScript(strings.NewReader("mouse 5 135\npan 240\n"))
	if err != nil {
		t.Fatal(err)
	}
	if cmds[0].Op != "mouse" || cmds[0].X != 5 || cmds[0].Y != 135 {
		t.Errorf("cmd 0 = %+v", cmds[0])
	}
	if cmds[1].Op != "pan" || cmds[1].N != 240 {
		t.Errorf("cmd 1 = %+v", cmds[1])
	}
}

func TestPanOptions(t *testing.T) {
	o, err := ParseOptions([]string{"-agent", "-pan", "0"})
	if err != nil {
		t.Fatal(err)
	}
	if !o.PanSet || o.Pan != 0 {
		t.Errorf("-pan 0 gave %+v", o)
	}
	o, err = ParseOptions([]string{"-agent"})
	if err != nil {
		t.Fatal(err)
	}
	if o.PanSet {
		t.Error("without -pan the camera must keep its default")
	}
}

// The mouse at an edge pans the camera, the middle does not, and the camera
// never leaves the room.
func TestEdgePanClamps(t *testing.T) {
	env := &Env{Input: NewInput(), CameraX: DefaultCameraX}
	s := &receptionScene{env: env}

	env.Input.InjectMouse(ScreenW/2, 100)
	s.pan()
	if env.CameraX != DefaultCameraX {
		t.Errorf("mouse in the middle moved the camera to %v", env.CameraX)
	}
	env.Input.InjectMouse(0, 100)
	for range 200 {
		s.pan()
	}
	if env.CameraX != 0 {
		t.Errorf("left edge: camera = %v, want 0", env.CameraX)
	}
	env.Input.InjectMouse(ScreenW-1, 100)
	for range 200 {
		s.pan()
	}
	if env.CameraX != MaxCameraX {
		t.Errorf("right edge: camera = %v, want %d", env.CameraX, MaxCameraX)
	}
}
