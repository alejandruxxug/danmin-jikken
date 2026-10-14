package game

import (
	"bytes"
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
			case n.Op == night.OpMusic && n.Name != "off":
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

// The rock must last through the silence-free part of the ending (title and
// credits) plus the fade, or the credits would roll in silence. The files are
// 80 s long.
func TestRockOutlastsTheCredits(t *testing.T) {
	const rockTicks = 80 * 60
	need := secretTitleTicks + CreditsTicks(CreditCards())
	if need > rockTicks {
		t.Errorf("title + credits last %d ticks but the rock tracks only %d", need, rockTicks)
	}
	if CreditsTicks(CreditCards())-creditsFadeTicks < 0 {
		t.Error("the fade starts before the credits")
	}
}

// The roll is over exactly when the last card and its gap are done.
func TestCreditsDoneMatchesCreditsTicks(t *testing.T) {
	s := &endingScene{cards: CreditCards()}
	total := CreditsTicks(s.cards)
	s.ticks = total - 5
	if s.creditsDone() {
		t.Error("credits done 5 ticks early")
	}
	s.ticks = total + 5
	if !s.creditsDone() {
		t.Error("credits not done 5 ticks late")
	}
}

// CardAt walks the cards: each shows for its ticks, then a quarter second of
// empty bar, and the roll ends after the last.
func TestCardAt(t *testing.T) {
	cards := []creditCard{{"A", "a", 100}, {"B", "b", 50}}
	for _, c := range []struct {
		ticks   int
		index   int
		showing bool
	}{{0, 0, true}, {99, 0, true}, {100, 0, false}, {100 + cardGapTicks - 1, 0, false}, {100 + cardGapTicks, 1, true}, {100 + cardGapTicks + 49, 1, true}, {100 + cardGapTicks + 50, 1, false}, {CreditsTicks(cards), 2, false}} {
		if i, showing := CardAt(cards, c.ticks); i != c.index || showing != c.showing {
			t.Errorf("CardAt(%d) = %d, %v; want %d, %v", c.ticks, i, showing, c.index, c.showing)
		}
	}
}

// The last card is the thank-you, 4 s; every card line fits the bottom bar.
func TestCreditCards(t *testing.T) {
	cards := CreditCards()
	last := cards[len(cards)-1]
	if last.Heading != "Thank you for staying awake" || last.Ticks != 240 {
		t.Errorf("last card %+v", last)
	}
	font, err := LoadFont()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cards {
		if w := font.Width(c.Line); w > 360 {
			t.Errorf("card line %q is %d px wide, the bar takes 360", c.Line, w)
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
		if strings.HasPrefix(name, "ending_rock_") && len(data) < 2_500_000 {
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
