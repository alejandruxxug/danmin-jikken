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
	want := []string{"silence", "title"}
	if len(seen) < 2 || seen[0] != want[0] || seen[1] != want[1] {
		t.Errorf("phases = %v, want to start with %v", seen, want)
	}
	if g.env.MusicCue != "ending_rock_wake" {
		t.Errorf("music cue = %q: the rock starts with the title", g.env.MusicCue)
	}
}

// Lists the backgrounds the scripts use that still have no art. It only
// reports (run it with -v): the count is how far the art is.
func TestPlaceholdersInUse(t *testing.T) {
	g, err := New(Options{Enabled: true, Scene: SceneNight1, Line: 1})
	if err != nil {
		t.Fatal(err)
	}
	art, err := g.env.NightArt()
	if err != nil {
		t.Fatal(err)
	}
	names, _ := night.ScriptNames()
	count := 0
	for _, name := range names {
		s, err := night.LoadScript(name)
		if err != nil {
			t.Fatal(err)
		}
		for _, bg := range s.BGNames() {
			if art.IsPlaceholder(bg) {
				count++
				t.Logf("PLACEHOLDER %-16s used in %s", bg, name)
			}
		}
	}
	t.Logf("%d placeholder backgrounds in use", count)
}

func TestNightCuesReachTheSoundPlayer(t *testing.T) {
	g, err := New(Options{Enabled: true, Scene: SceneNight1, Line: 1})
	if err != nil {
		t.Fatal(err)
	}
	for range 5 {
		if err := g.scenes.Update(); err != nil {
			t.Fatal(err)
		}
	}
	log := strings.Join(g.env.Sound.Log, ",")
	if !strings.Contains(log, "music night_ambience") {
		t.Errorf("sound log %q: the night should start its ambience", log)
	}
	if got := sfxFile("footsteps_many"); got != "footsteps" {
		t.Errorf("alias = %q", got)
	}
}

// The Sleep ending has no placeholder background left (same script, both routes).
func TestDeskDayHasArt(t *testing.T) {
	g, err := New(Options{Enabled: true, Scene: SceneNight1, Line: 1})
	if err != nil {
		t.Fatal(err)
	}
	art, err := g.env.NightArt()
	if err != nil {
		t.Fatal(err)
	}
	if art.IsPlaceholder("desk_day") {
		t.Error("desk_day is still a placeholder")
	}
	s, err := night.LoadScript("ending_sleep")
	if err != nil {
		t.Fatal(err)
	}
	for _, bg := range s.BGNames() {
		if art.IsPlaceholder(bg) {
			t.Errorf("ending_sleep uses placeholder %q", bg)
		}
	}
}

// Each ending's title comes back over its own last picture, and the credits
// run over the same picture until the top-10.
func TestEndingTitleComesBackOverTheLastPicture(t *testing.T) {
	for _, c := range []struct {
		ending, flags, bg string
	}{
		{night.EndingSleep, "route_mayu", "desk_day"},
		{night.EndingWake, "route_mayu,read_letter", "procedure_empty"},
		{night.EndingForever, "route_hana", "black"}, // the stain stays on the glass
		{night.Ending317, "route_nobody", "procedure_blank"},
	} {
		g, err := New(Options{Enabled: true, Scene: endingSceneID(c.ending), Line: 1 << 20, Flags: c.flags})
		if err != nil {
			t.Fatal(err)
		}
		s := g.scenes.Current().(*endingScene)
		phases := []string{}
		for i := 0; i < 6000 && g.env.nextScene == ""; i++ {
			if err := s.Update(); err != nil {
				t.Fatal(err)
			}
			if n := len(phases); n == 0 || phases[n-1] != s.phase.String() {
				phases = append(phases, s.phase.String())
				if s.phase == phaseTitle && s.script.engine.BG() != c.bg {
					t.Errorf("%s: the title comes back over %q, want %q", c.ending, s.script.engine.BG(), c.bg)
				}
			}
			g.env.Input.endFrame()
		}
		want := []string{"silence", "title", "credits"}
		if len(phases) != 3 || phases[0] != want[0] || phases[1] != want[1] || phases[2] != want[2] {
			t.Errorf("%s: phases %v, want %v then the top-10", c.ending, phases, want)
		}
		if g.env.nextScene != SceneDayTop10 {
			t.Errorf("%s: after the credits the game goes to %q, want the top-10", c.ending, g.env.nextScene)
		}
		if !s.script.engine.Cinematic() {
			t.Errorf("%s: the bars must still be up in the title and the credits", c.ending)
		}
	}
}
