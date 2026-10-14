package night

import (
	"strings"
	"testing"
)

func TestPaperRowsWrapAndSign(t *testing.T) {
	paper := Paper{Left: 0, Right: 60, FirstBase: 49, RowStep: 11}
	// Every letter is 5 px wide, so 12 letters (60 px) fit a row.
	rows := PaperRows([]string{"aaaaaaaa bbbbbbbb", "- Hana"}, paper, fakeMetrics{}, HanaStyle)
	if len(rows) != 3 {
		t.Fatalf("got %d rows, want 3 (two for the wrapped line, one signature)", len(rows))
	}
	if rows[0].Base != 49 || rows[1].Base != 60 {
		t.Errorf("rows sit on baselines %d and %d, want 49 and 60", rows[0].Base, rows[1].Base)
	}
	if rows[2].Base != 82 {
		t.Errorf("signature baseline %d, want 82 (one blank row under the text)", rows[2].Base)
	}
	if rows[1].First != 8 || rows[2].First != 16 {
		t.Errorf("rows start at letters %d and %d, want 8 and 16", rows[1].First, rows[2].First)
	}
}

func TestWrittenCountsLettersNotSpaces(t *testing.T) {
	if got := Written("ab cd", 4); got != 3 { // "ab c"
		t.Errorf("Written = %d, want 3", got)
	}
	if got := Written("ab", 99); got != 2 {
		t.Errorf("Written past the end = %d, want 2", got)
	}
}

// The two signatures of the realization (the letter and the day 3 report)
// are the same text in the same hand, so they lay out pixel for pixel alike.
func TestSignaturesAreIdentical(t *testing.T) {
	a, wa := Layout("- Hana", fakeMetrics{}, MayuStyle)
	b, wb := Layout("- Hana", fakeMetrics{}, MayuStyle)
	if wa != wb || len(a) != len(b) {
		t.Fatal("signatures differ in width")
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("glyph %d differs", i)
		}
	}
}

func TestPaperTextHeart(t *testing.T) {
	if got := PaperText([]string{"so tired <3"})[0]; got != "so tired ♥" {
		t.Errorf("got %q", got)
	}
}

func TestPaperEngine(t *testing.T) {
	src := "bg letter\nwriter mayu\nLetter: One.\nLetter: Two.\n~ A thought.\n"
	e := newEngine(t, src, Flags{})
	if e.Writer() != "mayu" || !e.OnPaper() {
		t.Fatalf("writer %q, onPaper %v", e.Writer(), e.OnPaper())
	}
	if err := e.Advance(); err != nil { // finish typing
		t.Fatal(err)
	}
	if err := e.Advance(); err != nil { // next line
		t.Fatal(err)
	}
	if got := e.Paper(); len(got) != 2 || got[0] != "One." || got[1] != "Two." {
		t.Fatalf("paper = %v, earlier lines must stay", got)
	}
	_ = e.Advance()
	_ = e.Advance()
	if e.OnPaper() || !e.Thinking() {
		t.Error("a thought is not written on the paper")
	}
	if len(e.Paper()) != 2 {
		t.Error("the paper keeps its lines under the thought")
	}
}

func TestNewBGEmptiesPaperAndHand(t *testing.T) {
	e := newEngine(t, "bg note\nwriter mayu\nNote: Hi.\nbg black\n~ Done.\n", Flags{})
	_ = e.Advance()
	_ = e.Advance()
	if e.Writer() != "hana" || len(e.Paper()) != 0 {
		t.Errorf("writer %q, paper %v after a new bg", e.Writer(), e.Paper())
	}
}

func TestWriterNeedsACharacter(t *testing.T) {
	if _, err := Parse("x.txt", strings.NewReader("writer barrow\n")); err == nil {
		t.Error("writer barrow must be an error")
	}
}
