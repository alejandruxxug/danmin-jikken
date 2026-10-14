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
	// Char counts the letters of the text (spaces not counted) up to and
	// including this glyph's letter, so a pen that has written n letters
	// draws the glyphs with Char <= n. A join belongs to the letter before it.
	Char int
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
	pen, chars := 0, 0
	for w, word := range strings.Fields(text) {
		if w > 0 {
			pen += st.WordSpace
		}
		rng := rand.New(rand.NewSource(int64(seed.Sum32()) + int64(w)))
		letters := []rune(word)
		run := 1 // letters in the joined run that ends at this letter
		for i, r := range letters {
			chars++
			glyph := string(r)
			adv, canJoin, ok := m.Advance(glyph)
			if !ok {
				continue
			}
			jitter := 0
			if st.Seeded {
				jitter = rng.Intn(3) - 1
			}
			placed = append(placed, Placed{Glyph: glyph, X: pen, Y: jitter, Char: chars})
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
				placed = append(placed, Placed{Glyph: joinGlyph, X: pen, Y: jitter, Char: chars})
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

// Paper is where a nurse writes on a sheet of paper: the rows sit on the
// paper's ruled lines (Art Direction -> Cinematic pass).
type Paper struct {
	Left, Right int // the writing area; signatures are right-aligned to Right
	FirstBase   int // baseline of the first row
	RowStep     int // distance between two rows (the paper's rules are 11 apart)
	Wrap        int // rows wrap at this width; 0 means Right-Left
}

// The two sheets: the note (night 1) and the letter (night 3).
var (
	// The note wraps a little early: in the Figma view "was" goes down a row (126 px fit in 127).
	NotePaper   = Paper{Left: 172, Right: 299, FirstBase: 88, RowStep: 11, Wrap: 120}
	LetterPaper = Paper{Left: 116, Right: 366, FirstBase: 49, RowStep: 11}
)

// PaperRow is one written row: its glyphs, where the row starts and the
// baseline it sits on. First is how many letters of the whole text come
// before it (so the pen knows when the row is reached).
type PaperRow struct {
	Glyphs []Placed
	X      int
	Base   int
	First  int
}

func (p Paper) wrapWidth() int {
	if p.Wrap > 0 {
		return p.Wrap
	}
	return p.Right - p.Left
}

// IsSignature says whether a line is the writer's signature ("- H.").
func IsSignature(text string) bool { return strings.HasPrefix(text, "- ") }

// WrapGlyphs breaks text into rows no wider than width pixels, at spaces,
// measuring with the real glyph widths.
func WrapGlyphs(text string, width int, m Metrics, st Style) []string {
	return WrapText(text, width, func(s string) int {
		_, w := Layout(s, m, st)
		return w
	})
}

// PaperRows lays out every line written on a paper, top to bottom. Each line
// starts a new row and wraps; a signature gets one blank row above it and is
// right-aligned. The same text always gives the same row, so two signatures
// of the same name are identical.
func PaperRows(lines []string, p Paper, m Metrics, st Style) []PaperRow {
	var rows []PaperRow
	base := p.FirstBase
	first := 0
	for _, line := range lines {
		if IsSignature(line) {
			base += p.RowStep // one blank row
			glyphs, w := Layout(line, m, st)
			rows = append(rows, PaperRow{Glyphs: glyphs, X: p.Right - w, Base: base, First: first})
			first += countLetters(line)
			base += p.RowStep
			continue
		}
		for _, text := range WrapGlyphs(line, p.wrapWidth(), m, st) {
			glyphs, _ := Layout(text, m, st)
			rows = append(rows, PaperRow{Glyphs: glyphs, X: p.Left, Base: base, First: first})
			first += countLetters(text)
			base += p.RowStep
		}
	}
	return rows
}

// countLetters is how many letters a text has, spaces not counted.
func countLetters(text string) int { return len([]rune(strings.Join(strings.Fields(text), ""))) }

// Written is how many letters of the text the pen has written when n
// characters (spaces included) of it are revealed.
func Written(text string, n int) int {
	r := []rune(text)
	return countLetters(string(r[:min(n, len(r))]))
}
