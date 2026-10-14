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

func TestDayOneEndsInTheStrobe(t *testing.T) {
	env := &Env{Day: 1}
	(&top10Scene{env: env}).next()
	if env.nextScene != SceneStrobe1 {
		t.Errorf("after the day 1 top-10 comes %q, want %q", env.nextScene, SceneStrobe1)
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
