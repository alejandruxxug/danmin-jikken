package game

import (
	"strings"
	"testing"
)

func TestDeskPaperSoundsPlayOncePerAction(t *testing.T) {
	p := soundPlayer(t)
	old := activeSound
	activeSound = p
	defer func() { activeSound = old }()
	n := func(cue string) int {
		c := 0
		for _, e := range p.Log {
			if strings.HasPrefix(e, "sfx "+cue) {
				c++
			}
		}
		return c
	}

	sh := newTestShift(t, 2)
	runTicks(sh, 5*60)
	if n("folder_slide") != 1 {
		t.Fatalf("the folder sliding in: %d slide sounds, want 1 (%v)", n("folder_slide"), p.Log)
	}
	if err := sh.OpenFolder(); err != nil {
		t.Fatal(err)
	}
	runTicks(sh, openTicks)
	if n("folder_open") != 1 {
		t.Errorf("opening: %d open sounds, want 1", n("folder_open"))
	}

	// A drag that is let go early: one lift sound at the press (not one per
	// frame of the drag) and one whisper when the page springs back.
	if err := sh.BeginFlipDrag(300, 100); err != nil {
		t.Fatal(err)
	}
	for x := 299; x > 280; x-- {
		sh.DragFlipTo(x, 100)
		sh.Update()
	}
	if n("page_") != 1 {
		t.Errorf("during a drag: %d page sounds, want 1", n("page_"))
	}
	sh.EndFlipDrag()
	runTicks(sh, 200)
	if n("page_") != 2 {
		t.Errorf("after the spring-back: %d page sounds, want 2", n("page_"))
	}

	// A drag all the way: lift, then the turn finishing.
	if err := sh.BeginFlipDrag(300, 100); err != nil {
		t.Fatal(err)
	}
	sh.DragFlipTo(300-pageW, 100)
	sh.EndFlipDrag()
	runTicks(sh, 200)
	if n("page_") != 4 {
		t.Errorf("after a full turn: %d page sounds, want 4", n("page_"))
	}
	last := ""
	for _, e := range p.Log {
		if strings.HasPrefix(e, "sfx page_") && e == last {
			t.Errorf("the same page sound twice in a row: %s", e)
		}
		if strings.HasPrefix(e, "sfx page_") {
			last = e
		}
	}

	// Stamping: the cover shuts once, the folder slides out once.
	if err := sh.GoToPage(PageReferral); err != nil {
		t.Fatal(err)
	}
	sh.LiftStamp(StampAdmit)
	if err := sh.ApplyStamp(); err != nil {
		t.Fatal(err)
	}
	runTicks(sh, inkTicks+closeTicks+slideOutTicks+10)
	if n("folder_close") != 1 {
		t.Errorf("closing: %d close sounds, want 1", n("folder_close"))
	}
	if n("folder_slide") != 2 {
		t.Errorf("sliding in and out: %d slide sounds, want 2", n("folder_slide"))
	}
}
