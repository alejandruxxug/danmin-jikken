package game

import (
	"testing"

	"danmingame/night"
)

func TestGlitchLevelRamps(t *testing.T) {
	prev := 0.0
	for ticks := 0; ticks <= glitchRamp+60; ticks += 10 {
		lvl := glitchLevel(ticks)
		if lvl < prev || lvl > 1 {
			t.Fatalf("level at %d = %v after %v: want a rise that stops at 1", ticks, lvl, prev)
		}
		prev = lvl
	}
	if glitchLevel(0) <= 0 || glitchLevel(glitchRamp) != 1 {
		t.Errorf("the lag is there from the start (%v) and full at the end (%v)", glitchLevel(0), glitchLevel(glitchRamp))
	}
}

func TestLagTickStuttersAndNeverRunsAhead(t *testing.T) {
	frozen := 0
	for ticks := 0; ticks < 600; ticks++ {
		shown := lagTick(ticks, 5, 1)
		if shown > ticks || ticks-shown >= glitchWindow {
			t.Fatalf("tick %d shows %d: the picture may only lag by less than one window", ticks, shown)
		}
		if shown != lagTick(ticks, 5, 1) {
			t.Fatalf("tick %d: the same seed must give the same picture", ticks)
		}
		if shown < ticks {
			frozen++
		}
	}
	if frozen == 0 {
		t.Error("the secret ending never froze")
	}
}

func TestEndingTitlesAreInTheFont(t *testing.T) {
	fsys, err := SpriteFS()
	if err != nil {
		t.Fatal(err)
	}
	art, err := loadCardArt(fsys, night.EndingForever)
	if err != nil {
		t.Fatal(err)
	}
	for _, ending := range night.Endings {
		title := cardTitles[ending]
		placed, _ := night.Layout(title, endingMetrics{art.font.Meta}, endingStyle)
		if len(placed) != len([]rune(title)) {
			t.Errorf("%s: %d glyphs placed for %q: the font is missing a letter", ending, len(placed), title)
		}
		if w := art.titleWidth(title); w > 320 {
			t.Errorf("%s: the title is %d px wide, it must leave room for the girl", ending, w)
		}
	}
}
