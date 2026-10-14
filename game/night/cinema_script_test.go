package night

import "testing"

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
