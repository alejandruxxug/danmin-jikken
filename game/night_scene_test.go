package game

import (
	"strings"
	"testing"

	"danmingame/night"
)

func TestNightEndGoesToTheNextDay(t *testing.T) {
	env := &Env{Day: 1, Flags: night.Flags{}}
	s := &nightScene{env: env, script: night.Night1}
	if err := s.finish("day2"); err != nil {
		t.Fatal(err)
	}
	if env.Day != 2 || env.nextScene != SceneDayReception {
		t.Errorf("after night 1: day %d, next %q; want day 2 at the reception", env.Day, env.nextScene)
	}
	if err := s.finish("somewhere"); err == nil {
		t.Error("an end target that is not dayN must be an error")
	}
}

func TestNightScenesAreRegistered(t *testing.T) {
	g, err := New(Options{Enabled: true, Scene: SceneNight1, Line: 1})
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{SceneNightPrologue, SceneNight1, SceneStrobe1} {
		if err := g.scenes.Switch(id); err != nil {
			t.Errorf("scene %q: %v", id, err)
		}
	}
	if err := g.scenes.Switch(SceneNight1); err != nil {
		t.Fatal(err)
	}
	if d := BuildDump(g.env, g.scenes.ID(), 1); d.Night == nil || d.Night.Script != "night1.txt" {
		t.Errorf("the dump has no night state: %+v", d.Night)
	}
}

func TestNightBackdoorCommands(t *testing.T) {
	cmds, err := ParseScript(strings.NewReader("next\nchoose 2\nskipstrobe\n"))
	if err != nil || len(cmds) != 3 || cmds[1].N != 2 {
		t.Fatalf("commands = %+v, %v", cmds, err)
	}
	for _, bad := range []string{"choose", "choose 9", "choose x", "next now"} {
		if _, err := ParseScript(strings.NewReader(bad)); err == nil {
			t.Errorf("%q must be an error", bad)
		}
	}
}

func TestNight3GoesToTheEndingTheFlagsPick(t *testing.T) {
	tests := []struct {
		flags []string
		want  string
	}{
		{[]string{"route_mayu", "closed_eyes"}, "ending.sleep"},
		{[]string{"route_nobody"}, "ending.wake"},
		{[]string{"route_hana"}, "ending.forever"},
		{[]string{"route_nobody", "read_letter", "stamped_discharge"}, "ending.317"},
	}
	for _, tt := range tests {
		env := &Env{Day: 3, Flags: night.Flags{}}
		for _, f := range tt.flags {
			env.Flags.Set(f)
		}
		s := &nightScene{env: env, script: night.Night3}
		if err := s.finish("ending"); err != nil || env.nextScene != tt.want {
			t.Errorf("%v: next scene %q, %v; want %q", tt.flags, env.nextScene, err, tt.want)
		}
	}
}

func TestEveryNightAndEndingSceneRuns(t *testing.T) {
	g, err := New(Options{Enabled: true, Scene: SceneNight1, Line: 1})
	if err != nil {
		t.Fatal(err)
	}
	ids := []string{SceneNight2, SceneNight3, SceneStrobe2, SceneStrobe3}
	for _, e := range night.Endings {
		ids = append(ids, endingSceneID(e))
	}
	for _, id := range ids {
		if err := g.scenes.Switch(id); err != nil {
			t.Errorf("scene %q: %v", id, err)
		}
	}
}

func TestEndingPhasesInOrder(t *testing.T) {
	g, err := New(Options{Enabled: true, Scene: "ending.wake", Line: 1 << 20})
	if err != nil {
		t.Fatal(err)
	}
	s := g.scenes.Current().(*endingScene)
	seen := []string{}
	for i := 0; i < 400 && g.env.nextScene == ""; i++ {
		if err := s.Update(); err != nil {
			t.Fatal(err)
		}
		if n := len(seen); n == 0 || seen[n-1] != s.phase.String() {
			seen = append(seen, s.phase.String())
		}
		g.env.Input.endFrame()
	}
	want := []string{"silence", "card"}
	if len(seen) < 2 || seen[0] != want[0] || seen[1] != want[1] {
		t.Errorf("phases = %v, want to start with %v", seen, want)
	}
	if g.env.MusicCue != "ending_rock_wake" {
		t.Errorf("music cue = %q: the rock starts with the card", g.env.MusicCue)
	}
}
