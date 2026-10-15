package main

import (
	"image"
	"os"
	"testing"
)

// testTimeline has made-up sprite boxes: the plan does not care what is in them.
func testTimeline() *Timeline {
	return &Timeline{
		MoonBox: image.Rect(200, 24, 274, 86), SealBox: image.Rect(410, 28, 440, 70),
		NightText: image.Rect(0, 150, 480, 160), NightEdges: []int{175, 186, 200},
		DontText: image.Rect(0, 158, 480, 168), DontEdges: []int{160, 170, 180, 190},
		EndLogo: image.Rect(0, 48, 480, 140), EndMade: image.Rect(0, 170, 480, 180), MadeEdges: []int{150, 160, 170},
		EndKanji: image.Rect(0, 206, 480, 222),
	}
}

// find returns the first layer of the frame that is the sprite name.
func find(t *testing.T, layers []Layer, name string) Layer {
	t.Helper()
	for _, l := range layers {
		if l.Kind == LayerSprite && l.Name == name {
			return l
		}
	}
	t.Fatalf("no %s layer in %d layers", name, len(layers))
	return Layer{}
}

func has(layers []Layer, name string) bool {
	for _, l := range layers {
		if l.Kind == LayerSprite && l.Name == name {
			return true
		}
	}
	return false
}

func gameOf(t *testing.T, layers []Layer) Layer {
	t.Helper()
	for _, l := range layers {
		if l.Kind == LayerGame {
			return l
		}
	}
	t.Fatal("no game layer")
	return Layer{}
}

func TestEveryFrameHasLayers(t *testing.T) {
	tl := testTimeline()
	for f := 0; f < TotalFrames; f++ {
		if len(tl.Plan(f)) == 0 {
			t.Fatalf("frame %d has no layers", f)
		}
	}
	if len(tl.Plan(TotalFrames)) != 0 {
		t.Error("there is no frame past the end")
	}
}

// Every recorded frame the plan asks for must be one shots.txt records.
func TestPlanUsesOnlyRecordedFrames(t *testing.T) {
	file, err := os.Open("shots.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	shots, err := ParseShots(file, ".")
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]Shot{}
	for _, s := range shots {
		byName[s.Name] = s
	}
	for _, reduced := range []bool{false, true} {
		tl := testTimeline()
		tl.Reduced = reduced
		for f := 0; f < TotalFrames; f++ {
			for _, l := range tl.Plan(f) {
				if l.Kind != LayerGame {
					continue
				}
				s, ok := byName[l.Shot]
				if !ok {
					t.Fatalf("frame %d uses shot %q, which shots.txt does not have", f, l.Shot)
				}
				if l.Frame < s.From || l.Frame > s.Frames {
					t.Fatalf("frame %d uses frame %d of %s, which records %d to %d", f, l.Frame, l.Shot, s.From, s.Frames)
				}
			}
		}
	}
}

func TestPlanSoundShotsExist(t *testing.T) {
	file, _ := os.Open("shots.txt")
	defer file.Close()
	shots, _ := ParseShots(file, ".")
	byName := map[string]Shot{}
	for _, s := range shots {
		byName[s.Name] = s
	}
	for _, reduced := range []bool{false, true} {
		for _, seg := range BuildAudioPlan(reduced).Segs {
			s, ok := byName[seg.Shot]
			if !ok || seg.From < s.From || seg.To > s.Frames {
				t.Errorf("sound of %s frames %d-%d is not in the recording", seg.Shot, seg.From, seg.To)
			}
		}
	}
}

func TestHiringCard(t *testing.T) {
	tl := testTimeline()
	if has(tl.Plan(45), "kt_hiring_line1") || !has(tl.Plan(46), "kt_hiring_line1") {
		t.Error("the first line comes in on beat 0 (frame 46)")
	}
	if dx := find(t, tl.Plan(46), "kt_hiring_line1").Sprite.DX; dx != 96 {
		t.Errorf("line1 starts 96 px to the right, got %d", dx)
	}
	if dx := find(t, tl.Plan(58), "kt_hiring_line1").Sprite.DX; dx != 0 {
		t.Errorf("line1 is in place on frame 58, got %d", dx)
	}
	// The moon drops in 40 px over 18 frames.
	moon := func(f int) int {
		for _, l := range tl.Plan(f) {
			if l.Name == "kt_hiring_bg" && !l.Sprite.Src.Empty() && l.Sprite.Src == tl.MoonBox {
				return l.Sprite.DY
			}
		}
		t.Fatal("no moon layer")
		return 0
	}
	if moon(0) != -40 || moon(18) != 0 || moon(200) != 0 {
		t.Errorf("moon: %d, %d, %d", moon(0), moon(18), moon(200))
	}
	// The stamp slams on beat 4: y -24, -12, 0, with the shake across x.
	for f, want := range map[int]int{142: -24, 144: -12, 146: 0} {
		if got := find(t, tl.Plan(f), "kt_hiring_stamp").Sprite.DY; got != want {
			t.Errorf("stamp y on f%d = %d, want %d", f, got, want)
		}
	}
	if has(tl.Plan(141), "kt_hiring_stamp") {
		t.Error("the stamp is not there before frame 142")
	}
	for f, want := range map[int]int{146: 2, 147: -2, 148: 1, 149: -1, 150: 0} {
		if got := find(t, tl.Plan(f), "kt_hiring_stamp").Sprite.DX; got != want {
			t.Errorf("shake on f%d = %d, want %d", f, got, want)
		}
	}
}

func TestWhipAndReception(t *testing.T) {
	tl := testTimeline()
	for _, l := range tl.Plan(233) {
		if l.Kind == LayerGame {
			t.Error("before the whip the card is alone")
		}
	}
	mid := tl.Plan(240) // half way through the whip: both are on screen
	if g := gameOf(t, mid); g.Shot != shotWake || g.Frame != 7 || g.Game.DX <= 0 || g.Game.DX >= GameW {
		t.Errorf("T02 entering: %+v", g)
	}
	if !has(mid, "sweep") {
		t.Error("the sweep rides the whip")
	}
	if g := gameOf(t, tl.Plan(246)); g.Frame != 13 || g.Game.DX != 0 {
		t.Errorf("T02 settled on f246: %+v", g)
	}
	if lt := find(t, tl.Plan(346), "lt_24hour"); lt.Sprite.DX != -300 {
		t.Errorf("the lower third starts off the left edge, got %d", lt.Sprite.DX)
	}
	if lt := find(t, tl.Plan(360), "lt_24hour"); lt.Sprite.DX != 0 {
		t.Errorf("the lower third is settled on f360, got %d", lt.Sprite.DX)
	}
}

func TestNeverClose(t *testing.T) {
	tl := testTimeline()
	if has(tl.Plan(499), "kt_never_close_text") {
		t.Error("nothing is typed before frame 500")
	}
	// 4 frames a letter: frame 520 has 6 letters, edge 105 from x 121.
	if got := find(t, tl.Plan(520), "kt_never_close_text").Sprite.Clip.Max.X; got != 121+105 {
		t.Errorf("clip at f520 = %d, want %d", got, 121+105)
	}
	if got := find(t, tl.Plan(560), "kt_never_close_text").Sprite.Clip.Max.X; got != 121+243 {
		t.Errorf("clip at f560 = %d, want %d", got, 121+243)
	}
	if got := find(t, tl.Plan(496), "kt_never_close_bg").Sprite.DY; got != 3 {
		t.Errorf("the card pops in 3 px low, got %d", got)
	}
	if got := find(t, tl.Plan(498), "kt_never_close_bg").Sprite.DY; got != 0 {
		t.Errorf("and settles, got %d", got)
	}
}

func TestStampPunchAndSlamFrame(t *testing.T) {
	tl := testTimeline()
	// The game's slam is frame 96 of the stamp shot and lands on beat 30 (f796).
	if g := gameOf(t, tl.Plan(Beat(30))); g.Shot != shotStamp || g.Frame != 96 {
		t.Errorf("f%d shows %s frame %d, want stamp 96", Beat(30), g.Shot, g.Frame)
	}
	if g := gameOf(t, tl.Plan(795)); g.Game.Zoom != 1 {
		t.Error("no punch before the slam")
	}
	if g := gameOf(t, tl.Plan(796)); g.Game.Zoom != 2 || g.Game.CX != 222 || g.Game.CY != 200 {
		t.Errorf("punch x2 on the stamp: %+v", g.Game)
	}
	if g := gameOf(t, tl.Plan(807)); g.Game.Zoom != 2 {
		t.Error("the punch lasts 12 frames")
	}
	if g := gameOf(t, tl.Plan(808)); g.Game.Zoom != 1 {
		t.Error("and goes back on frame 808")
	}
}

func TestOrientationPunchAndStatic(t *testing.T) {
	tl := testTimeline()
	if !has(tl.Plan(596), "channel_static") || !has(tl.Plan(601), "channel_static") || has(tl.Plan(602), "channel_static") {
		t.Error("the channel change is frames 596 to 601")
	}
	if i := find(t, tl.Plan(598), "channel_static").Index; i != 1 {
		t.Errorf("static picture on f598 = %d, want 1 (2 frames each)", i)
	}
	if gameOf(t, tl.Plan(645)).Game.Zoom != 1 || gameOf(t, tl.Plan(646)).Game.Zoom != 2 || gameOf(t, tl.Plan(696)).Game.Zoom != 1 {
		t.Error("punch x2 from beat 24 (646) to beat 26 (696)")
	}
}

func TestSplitScreen(t *testing.T) {
	tl := testTimeline()
	p := tl.Plan(t07Start)
	var left, right Layer
	for _, l := range p {
		if l.Kind == LayerGame && l.Shot == shotWake {
			left = l
		}
		if l.Kind == LayerGame && l.Shot == shotAsleep {
			right = l
		}
	}
	if left.Game.Clip != image.Rect(0, 0, 240, GameH) || right.Game.Clip != image.Rect(240, 0, GameW, GameH) {
		t.Errorf("the halves: %v and %v", left.Game.Clip, right.Game.Clip)
	}
	// Each half shows the crop x 120 to 359 of its reception.
	if left.Game.DX != -120 || right.Game.DX != 120 {
		t.Errorf("crop offsets %d and %d, want -120 and 120", left.Game.DX, right.Game.DX)
	}
	if div := find(t, p, "split_divider"); div.Sprite.DY != -GameH || div.Sprite.DX != 0 {
		t.Errorf("the divider starts above the picture: %+v", div.Sprite)
	}
	if div := find(t, tl.Plan(t07Start+8), "split_divider"); div.Sprite.DY != 0 {
		t.Errorf("the divider has dropped by f%d", t07Start+8)
	}
	// The collapse: the divider runs to the left edge between f1046 and f1058.
	if div := find(t, tl.Plan(1046), "split_divider"); div.Sprite.DX != 0 {
		t.Errorf("the collapse starts from the middle, divider at %d", div.Sprite.DX)
	}
	mid := find(t, tl.Plan(1052), "split_divider").Sprite.DX
	if mid >= 0 || mid <= -GameW/2 {
		t.Errorf("half way the divider is between the middle and the edge, got %d", mid)
	}
	end := tl.Plan(1058)
	if has(end, "split_divider") {
		t.Error("the divider is gone at the end of the collapse")
	}
	if len(end) != 1 || end[0].Shot != shotAsleep || end[0].Game.Clip != fullScreen || end[0].Game.DX != 0 {
		t.Errorf("then the asleep side fills the picture: %+v", end)
	}
	// The awake half is cut from frame 92 on, the asleep half from frame 230.
	if left.Frame != wakeLeftFrom || right.Frame != asleepFrom {
		t.Errorf("source frames %d and %d", left.Frame, right.Frame)
	}
}

func TestBedsPlateAndNoRoom(t *testing.T) {
	tl := testTimeline()
	if got := find(t, tl.Plan(t08Start), "hud_beds"); got.Index != 0 || got.Sprite.DY != -60 {
		t.Errorf("the plate starts above the picture at 0/3: %+v", got)
	}
	if got := find(t, tl.Plan(t08Start+8), "hud_beds").Sprite.DY; got != 0 {
		t.Errorf("the plate is down by f1104, got %d", got)
	}
	for f, want := range map[int]int{1120: 0, 1121: 1, 1146: 2, 1171: 3, 1300: 3} {
		if got := find(t, tl.Plan(f), "hud_beds").Index; got != want {
			t.Errorf("beds count on f%d = %d, want %d", f, got, want)
		}
	}
	if got := find(t, tl.Plan(1121), "hud_beds").Sprite.DY; got != -2 {
		t.Errorf("the plate bumps up 2 px when the count changes, got %d", got)
	}
	if got := find(t, tl.Plan(1123), "hud_beds").Sprite.DY; got != 0 {
		t.Errorf("and settles, got %d", got)
	}
	// NO ROOM slams on beat 48 (f1246), over the floor.
	if st := find(t, tl.Plan(1242), "stamp_no_room"); st.Sprite.DX != 90 || st.Sprite.DY != 70-24 {
		t.Errorf("stamp_no_room starts high: %+v", st.Sprite)
	}
	if st := find(t, tl.Plan(1246), "stamp_no_room"); st.Sprite.DY != 70 || st.Sprite.DX != 90+2 {
		t.Errorf("stamp_no_room lands with the shake: %+v", st.Sprite)
	}
	if has(tl.Plan(1241), "stamp_no_room") {
		t.Error("no stamp before frame 1242")
	}
	if g := gameOf(t, tl.Plan(t09Start)); g.Shot != shotSleepers || g.Frame != sleepersT09 {
		t.Errorf("T09 starts at sleepers frame 802, got %+v", g)
	}
}

func TestMemoZoomAndClock(t *testing.T) {
	tl := testTimeline()
	for f, zoom := range map[int]int{1346: 1, 1405: 1, 1406: 2, 1465: 2, 1466: 3, 1559: 3} {
		g := gameOf(t, tl.Plan(f))
		if g.Game.Zoom != zoom {
			t.Errorf("memo zoom on f%d = %d, want %d", f, g.Game.Zoom, zoom)
		}
		if zoom > 1 && (g.Game.CX != 408 || g.Game.CY != 60) {
			t.Errorf("the zoom is on the memo (408, 60), got %d, %d", g.Game.CX, g.Game.CY)
		}
	}
	// T10b: the memo's last frame, dimmed, with the plate and the digits.
	p := tl.Plan(1600)
	if g := gameOf(t, p); g.Frame != memoLast || g.Game.Zoom != 1 {
		t.Errorf("T10b base: %+v", g)
	}
	var checker, digits bool
	for _, l := range p {
		checker = checker || l.Kind == LayerChecker
		digits = digits || (l.Kind == LayerText && len(l.Text) == 5 && l.Scale == 4 && l.Y == 128 && l.X == 240 && l.Centered)
	}
	if !checker || !digits || !has(p, "hud_clock_plate") {
		t.Errorf("T10b needs the dim, the plate and the digits: %v %v", checker, digits)
	}
	if has(tl.Plan(1667), "hud_clock_317") || !has(tl.Plan(clockEnd), "hud_clock_317") {
		t.Error("03:17 lands on frame 1668")
	}
	if has(tl.Plan(clockEnd), "hud_clock_plate") {
		t.Error("the red 03:17 sprite replaces the plate")
	}
}

func TestIris(t *testing.T) {
	for _, c := range []struct {
		f       int
		closing bool
		start   int
		want    int
		ok      bool
	}{
		{2345, true, 2346, 0, false}, {2346, true, 2346, 0, true}, {2354, true, 2346, 2, true}, {2369, true, 2346, 5, true},
		{2370, false, 2370, 5, true}, {2382, false, 2370, 2, true}, {2393, false, 2370, 0, true}, {2394, false, 2370, 0, false},
	} {
		got, ok := irisIndex(c.f, c.start, c.closing)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("irisIndex(%d) = %d, %v; want %d, %v", c.f, got, ok, c.want, c.ok)
		}
	}
	tl := testTimeline()
	if !has(tl.Plan(2369), "iris") || !has(tl.Plan(2370), "iris") || has(tl.Plan(2394), "iris") {
		t.Error("the iris closes to f2369 and opens from f2370 to f2393")
	}
}

func TestGlitchCutAndFlashes(t *testing.T) {
	tl := testTimeline()
	rows := 0
	for _, l := range tl.Plan(t16Start) {
		if l.Kind == LayerGame {
			rows++
			if l.Game.Clip.Dy() != 8 {
				t.Errorf("glitch rows are 8 px tall, got %v", l.Game.Clip)
			}
		}
	}
	if rows != 34 {
		t.Errorf("%d glitch rows, want 34", rows)
	}
	if g := gameOf(t, tl.Plan(t16Start+glitchSteps)); g.Shot != shotKana || g.Game.Clip != (image.Rectangle{}) {
		t.Errorf("after the glitch the new shot is whole: %+v", g)
	}
	if !has(tl.Plan(2808), "channel_static") || has(tl.Plan(2807), "channel_static") {
		t.Error("the static covers frames 2808 to 2819")
	}
	for f, shot := range map[int]string{2856: shotFlashRed, 2859: shotFlashRed, 2904: shotFlashHana, 2906: shotFlashHana, 2958: shotGap, 3012: shotFlashRed} {
		if g := gameOf(t, tl.Plan(f)); g.Shot != shot {
			t.Errorf("flash on f%d shows %q, want %q", f, g.Shot, shot)
		}
	}
	for _, f := range []int{2860, 2907, 2960, 3014, 2830} {
		for _, l := range tl.Plan(f) {
			if l.Kind == LayerGame {
				t.Errorf("frame %d should be black, found a game frame", f)
			}
		}
	}
}

func TestDontSleepCard(t *testing.T) {
	tl := testTimeline()
	for f, want := range map[int]int{3060: 0, 3062: 0, 3063: 1, 3069: 3, 3071: 3} {
		if got := find(t, tl.Plan(f), "kt_dont_sleep_glitch").Index; got != want {
			t.Errorf("glitch picture on f%d = %d, want %d", f, got, want)
		}
	}
	if has(tl.Plan(3072), "kt_dont_sleep_glitch") || !has(tl.Plan(3072), "card_dont_sleep") {
		t.Error("the clean card from frame 3072")
	}
	for f, want := range map[int]int{3179: 0, 3180: 1, 3183: 1, 3184: 2, 3188: 2, 3189: 0} {
		if got := find(t, tl.Plan(f), "card_dont_sleep").Index; got != want {
			t.Errorf("card flicker picture on f%d = %d, want %d", f, got, want)
		}
	}
	// The line types 2 frames a letter from f3090.
	typed := func(f int) int {
		n := 0
		for _, l := range tl.Plan(f) {
			if l.Name == "card_dont_sleep" && !l.Sprite.Src.Empty() {
				n = l.Sprite.Clip.Max.X
			}
		}
		return n
	}
	if typed(3089) != 0 || typed(3090) != 160 || typed(3092) != 170 {
		t.Errorf("typing: %d, %d, %d", typed(3089), typed(3090), typed(3092))
	}
}

func TestTitleSplash(t *testing.T) {
	tl := testTimeline()
	// Frame 3261 is the title's frame 30, where the dagger lands.
	if g := gameOf(t, tl.Plan(slamFrame)); g.Shot != shotTitle || g.Frame != titleSlamFrame {
		t.Errorf("slam frame shows %+v", g)
	}
	for f, want := range map[int]int{3261: 0, 3262: 0, 3263: 1, 3264: 1, 3265: 2, 3266: 2} {
		if got := find(t, tl.Plan(f), "title_splash").Index; got != want {
			t.Errorf("splash on f%d = %d, want %d", f, got, want)
		}
	}
	if has(tl.Plan(3260), "title_splash") || has(tl.Plan(3267), "title_splash") {
		t.Error("the splash is frames 3261 to 3266")
	}
	if has(tl.Plan(3266), "title_stain") || !has(tl.Plan(3267), "title_stain") || !has(tl.Plan(3599), "title_stain") {
		t.Error("the stain stays from slam+6")
	}
	if has(tl.Plan(3270), "title_drips") {
		t.Error("the drips start at slam+10")
	}
	for f, rows := range map[int]int{3271: 34, 3272: 35, 3300: 63, 3599: 270} {
		if got := find(t, tl.Plan(f), "title_drips").Sprite.Clip.Max.Y; got != rows {
			t.Errorf("drips on f%d reach row %d, want %d", f, got, rows)
		}
	}
}

func TestEndCard(t *testing.T) {
	tl := testTimeline()
	bayer := func(f int) int {
		for _, l := range tl.Plan(f) {
			if l.Name == "card_end_credits" {
				return l.Sprite.Bayer
			}
		}
		return -1
	}
	for f, want := range map[int]int{3719: -1, 3720: 4, 3723: 4, 3724: 8, 3727: 8, 3728: 0, 3989: 0} {
		if got := bayer(f); got != want {
			t.Errorf("credits dither on f%d = %d, want %d (-1 = not there)", f, got, want)
		}
	}
	if !has(tl.Plan(3600), "end_stain") || !has(tl.Plan(3600), "card_end") {
		t.Error("the logo and the stain are there from the first frame")
	}
	if n := len(tl.Plan(3629)); n != 3 { // black, logo, stain
		t.Errorf("before the line types there are 3 layers, got %d", n)
	}
	for _, l := range tl.Plan(3990) {
		if l.Kind != LayerFill {
			t.Error("black from frame 3990")
		}
	}
}
