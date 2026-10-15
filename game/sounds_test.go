package game

import (
	"bytes"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"testing"
	"time"

	"danmingame/night"
	"danmingame/sound"
)

// soundPlayer is a muted player reading the real embedded sound folder, so
// FileName says which cues have a file.
func soundPlayer(t *testing.T) *sound.Player {
	t.Helper()
	fsys, err := SoundFS()
	if err != nil {
		t.Fatal(err)
	}
	return sound.New(fsys, true)
}

// Every sfx and music cue in every night script resolves to a file.
func TestEveryScriptCueHasAFile(t *testing.T) {
	p := soundPlayer(t)
	names, err := night.ScriptNames()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		script, err := night.LoadScript(name)
		if err != nil {
			t.Fatal(err)
		}
		for _, n := range script.Nodes {
			switch {
			case n.Op == night.OpSfx && n.Name != "off":
				if p.FileName(sfxFile(n.Name)) == "" {
					t.Errorf("%s: sfx %q has no file (looked for %q)", name, n.Name, sfxFile(n.Name))
				}
			case n.Op == night.OpMusic && n.Name != "off" && n.Name != "fade":
				if p.FileName(musicFile(n.Name)) == "" {
					t.Errorf("%s: music %q has no file (looked for %q)", name, n.Name, musicFile(n.Name))
				}
			}
		}
	}
}

// The cues the Go code plays by itself also have files.
func TestEveryCodeCueHasAFile(t *testing.T) {
	p := soundPlayer(t)
	cues := []string{
		"title", "night_ambience", "day_hum", // music
		"bell", "snore", "stamp_thunk", "stamp_lift", "chime_right", // the day
		"stun_hit", "stun_ring", "heartbeat", // the strobe's stun
		"driftwell_jingle",
	}
	cues = append(cues, "folder_open", "folder_close", "folder_slide")
	for v := 1; v <= pageVariants; v++ {
		cues = append(cues, fmt.Sprintf("page_%d", v))
	}
	for v := 1; v <= snoreVariants; v++ {
		cues = append(cues, snoreName(v))
	}
	for _, name := range night.Endings {
		cues = append(cues, "ending_rock_"+name)
	}
	for _, c := range cues {
		if p.FileName(c) == "" {
			t.Errorf("cue %q has no sound file", c)
		}
	}
}

// Every ending has a start time for its rock, and none runs past its file's
// quiet part by mistake (the files are cut already, so the start is small).
func TestRockStartsCoverEveryEnding(t *testing.T) {
	for _, name := range night.Endings {
		start, ok := rockStarts[name]
		if !ok {
			t.Errorf("ending %q has no rock start", name)
		}
		if start < 0 || start.Seconds() > 10 {
			t.Errorf("ending %q starts its rock at %v", name, start)
		}
	}
}

// The rock must last through what plays after it enters (title and finale)
// or the finale would play in silence. Each file's real length is read. For
// Forever the rock enters earlier, so what counts is the part of the file
// after the drop, which lands on the title.
func TestRockOutlastsTheFinale(t *testing.T) {
	p := soundPlayer(t)
	need := time.Duration(secretTitleTicks+FinaleStageStart("black")) * time.Second / 60 // it fades out into the black
	for _, name := range night.Endings {
		length, err := p.Length("ending_rock_" + name)
		if err != nil {
			t.Fatal(err)
		}
		left := length - rockStarts[name] // from the slam to the end of the file
		if drop, timed := rockDrops[name]; timed {
			left = length - drop
		}
		if left < need {
			t.Errorf("%s: %v of rock after the title enters, but the title and finale last %v", name, left, need)
		}
	}
	if FinaleStageStart("black")-finaleFadeOut < FinaleStageStart("logo") {
		t.Error("the music fade starts before the logo")
	}
}

// Where the timed rock starts: the drop minus the time to the title, never
// before the start of the file.
func TestSyncedStart(t *testing.T) {
	for _, c := range []struct {
		drop    time.Duration
		titleAt int
		start   time.Duration
		clamped bool
	}{
		{34 * time.Second, 600, 24 * time.Second, false}, // 10 s of ticks
		{34 * time.Second, 60 * 34, 0, false},            // lands exactly at the file's start
		{34 * time.Second, 60*34 + 1, 0, true},           // one tick too long: starts at 0, late
		{34 * time.Second, 0, 34 * time.Second, false},   // no lead-up: the drop is the start
	} {
		start, clamped := syncedStart(c.drop, c.titleAt)
		if start != c.start || clamped != c.clamped {
			t.Errorf("syncedStart(%v, %d) = %v, %v; want %v, %v", c.drop, c.titleAt, start, clamped, c.start, c.clamped)
		}
	}
}

// Forever's rock reaches back far enough for every path of its script.
func TestForeverRockReachesItsDrop(t *testing.T) {
	script, err := night.LoadScript(night.EndingScript(night.EndingForever))
	if err != nil {
		t.Fatal(err)
	}
	for _, flags := range []night.Flags{{"route_hana": true}, {"route_hana": true, "closed_eyes": true}} {
		ticks, err := night.Timeline(script, flags, "Test")
		if err != nil {
			t.Fatal(err)
		}
		start, clamped := syncedStart(rockDrops[night.EndingForever], ticks+silenceTicks)
		t.Logf("forever %v: title at %d ticks, rock starts %.2f s into the file", flags.Names(), ticks+silenceTicks, start.Seconds())
		if clamped || start < 2*time.Second {
			t.Errorf("forever %v: the rock starts at %v: cut the file earlier", flags.Names(), start)
		}
	}
}

// Forever from the backdoor with nobody clicking: the music starts with the
// first line, the title comes by itself at the planned tick, and the music
// was not restarted or stopped on the way.
func TestForeverPlaysItselfToTheTitle(t *testing.T) {
	g, err := New(Options{Enabled: true, Scene: "ending.forever", Flags: "route_hana"})
	if err != nil {
		t.Fatal(err)
	}
	s := g.scenes.Current().(*endingScene)
	want := g.env.EndingTitleAt
	if want == 0 || g.env.EndingRockFrom <= 0 {
		t.Fatalf("no plan: title at %d, rock from %v", want, g.env.EndingRockFrom)
	}
	for tick := 0; s.phase != phaseTitle; tick++ {
		if tick > want+10 {
			t.Fatalf("no title after %d ticks (planned %d), phase %v", tick, want, s.phase)
		}
		if err := s.Update(); err != nil {
			t.Fatal(err)
		}
		g.env.Input.endFrame()
	}
	if g.env.MusicCue != "ending_rock_forever" {
		t.Errorf("music cue %q", g.env.MusicCue)
	}
	plays := 0
	for _, l := range g.env.Sound.Log {
		if l == "music ending_rock_forever" {
			plays++
		}
		if l == "music off" {
			t.Error("the music was stopped on the way to the title")
		}
	}
	if plays != 1 {
		t.Errorf("the rock was started %d times, want once", plays)
	}
}

// The thanks is the last card and the roll is one column: every line of it
// fits the 336 px column, the widest in uppercase like the roll prints it.
func TestCreditCards(t *testing.T) {
	cards := CreditCards()
	last := cards[len(cards)-1]
	if last.Heading != "Thank you for staying awake" || last.Line != "" {
		t.Errorf("last card %+v", last)
	}
	font, err := LoadFont()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cards {
		if w := font.Width(strings.ToUpper(c.Line)); w > 336 {
			t.Errorf("card line %q is %d px wide, the column takes 336", c.Line, w)
		}
	}
}

// The credits name the licences that require it.
func TestCreditsHaveTheRequiredLines(t *testing.T) {
	all := strings.Join(musicCredits, "\n")
	for _, want := range []string{"Ironbound by Zander Noriega", "CC BY 3.0", "Hot Swing by Kevin MacLeod", "Creative Commons: By Attribution 4.0"} {
		if !strings.Contains(all, want) {
			t.Errorf("credits miss %q", want)
		}
	}
}

// The stun is three sounds and the snore is held down.
func TestStunCuesAndGains(t *testing.T) {
	p := soundPlayer(t)
	old := activeSound
	activeSound = p
	defer func() { activeSound = old }()
	playStun()
	got := strings.Join(p.Log, ",")
	if got != "sfx stun_hit,sfx stun_ring,sfx heartbeat" {
		t.Errorf("stun log = %q", got)
	}
	if sfxGain["stun_ring"] >= 1 {
		t.Error("the ringing must be quieter than its file")
	}
}

// Every file in assets/sound is real: the small ones decode and play for a
// moment (an empty ogg once got into the folder), the long mp3 tracks are
// checked by size and header only, because decoding minutes of mp3 under the
// race detector takes a minute of test time. A rock track is a 320 kbps mp3
// of about 80 s, so at least 2.5 MB.
func TestEverySoundFileIsReal(t *testing.T) {
	fsys, err := SoundFS()
	if err != nil {
		t.Fatal(err)
	}
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		t.Fatal(err)
	}
	p := soundPlayer(t)
	for _, e := range entries {
		ext := path.Ext(e.Name())
		if ext == ".txt" {
			continue
		}
		data, err := fs.ReadFile(fsys, e.Name())
		if err != nil {
			t.Fatal(err)
		}
		name := strings.TrimSuffix(e.Name(), ext)
		if strings.HasPrefix(name, "ending_rock_") && len(data) < 2_500_000 { // Forever's runs from about 0:50 to the end of the song, 6.2 MB
			t.Errorf("%s is only %d bytes: too short for the credits", e.Name(), len(data))
		}
		if len(data) > 1<<20 {
			if ext != ".mp3" || !(bytes.HasPrefix(data, []byte("ID3")) || data[0] == 0xFF) {
				t.Errorf("%s: a long file that does not look like an mp3", e.Name())
			}
			continue
		}
		length, err := p.Length(name)
		if err != nil {
			t.Errorf("%s: %v", e.Name(), err)
		} else if length < 100*time.Millisecond {
			t.Errorf("%s plays for %v: empty?", e.Name(), length)
		}
	}
}
