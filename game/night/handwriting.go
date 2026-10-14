package night

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"strings"
)

// Handwriting layout (Art Direction -> Handwriting glyph sheets). The two
// nurses' handwriting is drawn from glyph sheets, one layer per character.
// This file only decides where each glyph goes; it has no drawing in it, so
// the night notes can use it later and a test can check it.

// Metrics says how wide a glyph is and whether it can join its neighbour.
// The sheets keep this in the layer user data: "adv=5;join=1".
type Metrics interface {
	Advance(glyph string) (adv int, canJoin, ok bool)
}

// Style is how one nurse writes.
type Style struct {
	Gap       int  // pixels between letters that are not joined
	WordSpace int  // pixels between words
	Seeded    bool // Mayu: joins are a seeded coin flip, letters jitter by a pixel
}

// The two hands. Hana's glyphs carry their own join strokes, so her pairs
// simply touch; Mayu's joins are a separate "join" glyph.
var (
	HanaStyle = Style{Gap: 1, WordSpace: 3}
	MayuStyle = Style{Gap: 2, WordSpace: 4, Seeded: true}
)

// joinGlyph is Mayu's connector stroke; it is 2 px wide, as wide as her gap.
const (
	joinGlyph = "join"
	joinWidth = 2
	maxRun    = 3 // Mayu never joins more than 3 letters in a row
)

// Placed is one glyph to draw: its layer name and where its 12x12 cell goes,
// relative to the start of the text. Y is the jitter (-1, 0 or 1).
type Placed struct {
	Glyph string
	X, Y  int
}

// ParseGlyphData reads the layer user data "adv=5;join=1".
func ParseGlyphData(data string) (adv int, canJoin bool, err error) {
	for _, field := range strings.Split(data, ";") {
		key, value, _ := strings.Cut(field, "=")
		switch key {
		case "adv":
			if _, err := fmt.Sscanf(value, "%d", &adv); err != nil {
				return 0, false, fmt.Errorf("glyph data %q: bad adv: %w", data, err)
			}
		case "join":
			canJoin = value == "1"
		}
	}
	return adv, canJoin, nil
}

// Layout places the glyphs of text and returns them with the text's width.
// Letters the sheet does not have are skipped. All of Mayu's random choices
// come from a generator seeded with a hash of the whole text plus the word's
// number, so the same text looks the same every time it is shown.
func Layout(text string, m Metrics, st Style) (placed []Placed, width int) {
	seed := fnv.New32a()
	seed.Write([]byte(text))
	pen := 0
	for w, word := range strings.Fields(text) {
		if w > 0 {
			pen += st.WordSpace
		}
		rng := rand.New(rand.NewSource(int64(seed.Sum32()) + int64(w)))
		letters := []rune(word)
		run := 1 // letters in the joined run that ends at this letter
		for i, r := range letters {
			glyph := string(r)
			adv, canJoin, ok := m.Advance(glyph)
			if !ok {
				continue
			}
			jitter := 0
			if st.Seeded {
				jitter = rng.Intn(3) - 1
			}
			placed = append(placed, Placed{Glyph: glyph, X: pen, Y: jitter})
			pen += adv
			if i == len(letters)-1 {
				break // nothing follows in this word
			}
			_, nextCan, nextOK := m.Advance(string(letters[i+1]))
			joins := canJoin && nextOK && nextCan
			if st.Seeded {
				coin := rng.Intn(2) == 0 // drawn every time, so the stream never shifts
				joins = joins && coin && run < maxRun
			}
			switch {
			case joins && st.Seeded:
				placed = append(placed, Placed{Glyph: joinGlyph, X: pen, Y: jitter})
				pen += joinWidth
				run++
			case joins:
				run++ // Hana: the glyphs' own strokes touch
			default:
				pen += st.Gap
				run = 1
			}
		}
	}
	return placed, pen
}
