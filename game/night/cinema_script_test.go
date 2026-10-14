package night

import (
	"os"
	"slices"
	"testing"
	"time"

	"danmingame/sprites"
)

// playTo clicks through night 3 until the line `text` is on screen, with
// the flags set, and returns the engine.
func playTo(t *testing.T, flags []string, text string) *Engine {
	t.Helper()
	f := Flags{}
	for _, name := range flags {
		f.Set(name)
	}
	script, err := LoadScript(Night3)
	if err != nil {
		t.Fatal(err)
	}
	e, err := NewEngine(script, f)
	if err != nil {
		t.Fatal(err)
	}
	for range 400 {
		if e.FullText() == text {
			return e
		}
		if len(e.Options()) > 0 {
			_ = e.Choose(e.pickedOption())
			continue
		}
		for range 3 {
			_ = e.Tick()
		}
		_ = e.Advance()
		_ = e.Advance()
	}
	t.Fatalf("never reached %q", text)
	return nil
}

func TestHanaRoomIsCinematicFromTheDoorToTheChoice(t *testing.T) {
	door := playTo(t, []string{"route_hana"}, "Her name is on the door, in her small, neat letters.")
	if !door.Cinematic() || door.BG() != "door_closeup" {
		t.Errorf("at the door: cinematic %v, bg %q", door.Cinematic(), door.BG())
	}
	punch := playTo(t, []string{"route_hana"}, "Hana is there.")
	if s, cx, cy := punch.Zoom(); s != 2 || cx != 282 || cy != 165 {
		t.Errorf("on 'Hana is there.' the camera is x%v at (%v,%v), want x2 at (282,165)", s, cx, cy)
	}
	floor := playTo(t, []string{"route_hana"}, "Her clipboard is on the floor, snapped in half. Her nails are broken.")
	if floor.BG() != "hana_floor" || !floor.Prop("clipboard") {
		t.Errorf("clipboard shot: bg %q, clipboard %v", floor.BG(), floor.Prop("clipboard"))
	}
	envelope := playTo(t, []string{"route_hana"}, "On the desk there's an envelope. It's the only clean thing in here. It says Four.")
	if envelope.BG() != "hana_desk" || !envelope.Cinematic() {
		t.Errorf("envelope shot: bg %q, cinematic %v", envelope.BG(), envelope.Cinematic())
	}
}

// On Mayu's and nobody's routes the clipboard shot is skipped.
func TestNoClipboardOutsideHanaRoute(t *testing.T) {
	e := playTo(t, []string{"route_mayu"}, "The newest ones are still sealed. She stopped.")
	for range 2 {
		_ = e.Advance()
		_ = e.Advance()
	}
	if e.Prop("clipboard") {
		t.Error("the clipboard is only snapped on Hana's route")
	}
}

// The letter is read without the bars and in the forger's hand on Hana's route.
func TestLetterWriters(t *testing.T) {
	forged := playTo(t, []string{"route_hana", "read_letter"}, "I'm so sorry!! I was just SO tired <3")
	if forged.Writer() != "mayu" || forged.Cinematic() {
		t.Errorf("forged letter: writer %q, cinematic %v", forged.Writer(), forged.Cinematic())
	}
	hers := playTo(t, []string{"route_mayu", "read_letter"}, "Four. You chose her. I understand.")
	if hers.Writer() != "hana" {
		t.Errorf("Hana writes the letter on Mayu's route, got %q", hers.Writer())
	}
}

// Forever's drop plays once: the taut rope, then the break, and the last
// frame stays for the splash. The frames come from the real sheet.
func TestDropFrames(t *testing.T) {
	data, err := os.ReadFile("../assets/sprites/cg_forever_drop.json")
	if err != nil {
		t.Fatal(err)
	}
	meta, err := sprites.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	ms := time.Millisecond
	for _, c := range []struct {
		at   time.Duration
		want int
	}{{0, 0}, {160 * ms, 1}, {310 * ms, 2}, {449 * ms, 2}, {450 * ms, 3}, {560 * ms, 4}, {660 * ms, 5}, {749 * ms, 5}, {5000 * ms, 5}} {
		if got := DropFrame(meta, c.at); got != c.want {
			t.Errorf("DropFrame at %v = %d, want %d", c.at, got, c.want)
		}
	}
}

// The Forever script shows the rope in the order of the Figma row: the hands
// first, then the loose loop with the smile; every layer is in cg_forever.json
// (TestCGLayersExistInTheSheets checks the names).
func TestForeverRopeSteps(t *testing.T) {
	script, err := LoadScript(EndingScript(EndingForever))
	if err != nil {
		t.Fatal(err)
	}
	f := Flags{}
	f.Set("route_hana")
	en, err := NewEngine(script, f)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string][]string{}
	for range 200 {
		if _, ended := en.Ended(); ended || len(en.Options()) > 0 {
			break
		}
		seen[en.FullText()] = en.Snapshot().CGLayers
		_ = en.Advance()
		_ = en.Advance()
		for range 200 {
			_ = en.Tick()
		}
	}
	has := func(text, layer string) bool { return slices.Contains(seen[text], layer) }
	if !has("Through the gap, I can still see her.", "mouth_still") || has("Through the gap, I can still see her.", "rope_hands") {
		t.Error("through the gap: her face with the still mouth, no rope yet")
	}
	if !has("She's standing on the chair.", "rope_hands") {
		t.Error("standing on the chair: her hands raise the rope")
	}
	smile := "She's smiling at me."
	if !has(smile, "rope_loose") || !has(smile, "mouth_smile") || has(smile, "rope_hands") || has(smile, "mouth_still") {
		t.Errorf("smiling: the rope lies loose and she smiles, got %v", seen[smile])
	}
}

// Nobody is drawn in front of a paper or a CG: the painter asks GirlShown.
// Every script is played through on a few routes and every cut is checked.
func TestNobodyStandsInFrontOfAPaperOrACG(t *testing.T) {
	names, err := ScriptNames()
	if err != nil {
		t.Fatal(err)
	}
	routes := [][]string{
		{"route_mayu", "read_letter", "closed_eyes"}, {"route_hana", "read_letter"},
		{"route_nobody", "read_letter"}, {"route_mayu"}, {"route_hana", "closed_eyes"},
	}
	for _, name := range names {
		script, err := LoadScript(name)
		if err != nil {
			t.Fatal(err)
		}
		for _, flags := range routes {
			f := Flags{}
			for _, fl := range flags {
				f.Set(fl)
			}
			e, err := NewEngine(script, f)
			if err != nil {
				t.Fatal(err)
			}
			for range 2000 {
				if _, ended := e.Ended(); ended {
					break
				}
				if def, ok := bgTable[e.BG()]; (e.OnPaperBG() || (ok && def.cg)) && e.GirlShown() {
					t.Fatalf("%s %v line %d: %s would be drawn over %q", name, flags, e.Line(), e.Visible(), e.BG())
				}
				if len(e.Options()) > 0 {
					_ = e.Choose(e.pickedOption())
					continue
				}
				_ = e.Tick()
				_ = e.Advance()
				_ = e.Advance()
			}
		}
	}
}
