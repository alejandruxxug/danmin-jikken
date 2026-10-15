package main

import "testing"

func TestGlyphsRenderDigits(t *testing.T) {
	g, err := LoadGlyphs("../../game/fonts/Silkscreen-Regular.ttf")
	if err != nil {
		t.Skip("font not found:", err)
	}
	for _, s := range []string{"0", "03:17", "23:41", "17:58"} {
		b := g.Render(s)
		if b.W == 0 || b.H < 4 || b.H > FontPx {
			t.Errorf("Render(%q) is %dx%d: want a glyph no taller than the font", s, b.W, b.H)
		}
	}
	one, two := g.Render("7"), g.Render("77")
	if two.W <= one.W {
		t.Error("two digits should be wider than one")
	}
}
