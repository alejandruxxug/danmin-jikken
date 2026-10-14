package night

import (
	"fmt"
	"strconv"
	"time"

	"danmingame/sprites"
)

// Cinematic mode (Art Direction -> Cinematic pass): letterbox bars, the
// thoughts as subtitles in the bottom bar, and a camera that can punch in
// (a hard cut to a whole-number zoom), push in (a slow zoom) and pan across
// a wide picture. Everything here is plain arithmetic on ticks, so tests can
// check it without a window; the painter (draw.go) only reads the results.

// Bar sizes and times: the bars are 32 px tall, slide in over 0.6 s and out
// over 0.4 s.
const (
	BarFull      = 32
	barInTicks   = 36
	barOutTicks  = 24
	secondsToTik = TicksPerSecond
)

// barState is a bar height moving from one value to another.
type barState struct {
	from, to float64 // pixels
	start    int     // the engine tick the move began
	dur      int     // ticks; 0 = instant
	white    bool    // the Sleep ending's dream has white bars
}

// BarHeight is the bar height elapsed ticks into a move from one height to
// another over dur ticks. It eases out: fast at first, settling at the end.
func BarHeight(from, to float64, elapsed, dur int) float64 {
	switch {
	case dur <= 0 || elapsed >= dur:
		return to
	case elapsed <= 0:
		return from
	}
	return from + (to-from)*easeOut(float64(elapsed)/float64(dur))
}

// zoomState is the camera's scale moving towards a target, about a centre
// point of the 480x270 picture.
type zoomState struct {
	from, to float64
	cx, cy   float64
	start    int
	dur      int
}

// ZoomScale is the camera scale elapsed ticks into a zoom (eased out).
func ZoomScale(from, to float64, elapsed, dur int) float64 {
	return BarHeight(from, to, elapsed, dur)
}

// CropOrigin is the top-left corner of the part of the picture the camera
// shows at a scale, centred on (cx, cy) but never looking past the picture's
// edge. The picture is drawn at (-x*scale, -y*scale), scaled.
func CropOrigin(scale, cx, cy float64) (x, y float64) {
	w, h := screenW/scale, screenH/scale
	x = min(max(cx-w/2, 0), screenW-w)
	y = min(max(cy-h/2, 0), screenH-h)
	return x, y
}

// panState is the camera's sideways position on a wide picture (hana_floor).
type panState struct {
	from, to float64
	start    int
	dur      int
}

// PanX is the camera's x elapsed ticks into a pan: smooth at both ends.
func PanX(from, to float64, elapsed, dur int) float64 {
	switch {
	case dur <= 0 || elapsed >= dur:
		return to
	case elapsed <= 0:
		return from
	}
	t := float64(elapsed) / float64(dur)
	return from + (to-from)*t*t*(3-2*t)
}

// ---- the engine's side ----

// startBars moves the bars to a height over some ticks.
func (e *Engine) startBars(to float64, dur int) {
	e.bars = barState{from: e.BarsHeight(), to: to, start: e.ticks, dur: dur, white: e.bars.white}
}

// BarsHeight is the height of each bar now, in pixels.
func (e *Engine) BarsHeight() float64 {
	return BarHeight(e.bars.from, e.bars.to, e.ticks-e.bars.start, e.bars.dur)
}

// Cinematic says whether the bars are up (or sliding in): the text box and
// the AUTO toggle give way to subtitles in the bottom bar.
func (e *Engine) Cinematic() bool { return e.bars.to > 0 }

// WhiteBars says whether the bars are the dream's white ones.
func (e *Engine) WhiteBars() bool { return e.bars.white }

// Zoom is the camera now: its scale (1 = the whole picture) and the centre.
func (e *Engine) Zoom() (scale, cx, cy float64) {
	z := e.zoom
	if z.to == 0 {
		return 1, screenW / 2, screenH / 2
	}
	return ZoomScale(z.from, z.to, e.ticks-z.start, z.dur), z.cx, z.cy
}

// PanX is the camera's sideways position on a wide picture now.
func (e *Engine) PanX() float64 {
	return PanX(e.pan.from, e.pan.to, e.ticks-e.pan.start, e.pan.dur)
}

// execCinema runs the bars, punch, push and pan commands.
func (e *Engine) execCinema(n Node) {
	switch n.Op {
	case OpBars:
		e.bars.white = e.bars.white || n.Name == "white"
		if n.Name == "on" {
			e.bars.white = false
		}
		if n.Name == "white" && e.bars.to > 0 {
			return // already up, only the colour changes
		}
		e.startBars(n.Args[0], int(n.Args[1]*secondsToTik))
	case OpPunch, OpPush:
		scale := n.Args[0]
		dur := 0
		if n.Op == OpPush {
			dur = int(n.Args[3] * secondsToTik)
		}
		from, _, _ := e.Zoom()
		e.zoom = zoomState{from: from, to: scale, cx: n.Args[1], cy: n.Args[2], start: e.ticks, dur: dur}
		if scale == 1 {
			e.zoom = zoomState{} // off
		}
	case OpPan:
		e.pan = panState{from: e.PanX(), to: n.Args[0], start: e.ticks, dur: int(n.Args[1] * secondsToTik)}
	}
}

// ---- parsing ----

// parseBars reads `bars on|off|snap|white|<px> [seconds]` into the node:
// Name is "white" or "", Args are the height in pixels and the seconds.
func parseBars(n *Node, args []string) error {
	if len(args) == 0 || len(args) > 2 {
		return fmt.Errorf("bars is `bars on|off|snap|white|<pixels> [seconds]`")
	}
	n.Op = OpBars
	switch args[0] {
	case "on":
		n.Name, n.Args = "on", []float64{BarFull, 0.6}
	case "off":
		n.Args = []float64{0, 0.4}
	case "snap":
		n.Args = []float64{BarFull, 0}
	case "white":
		n.Name, n.Args = "white", []float64{BarFull, 0.6}
	default:
		px, err := strconv.ParseFloat(args[0], 64)
		if err != nil || px < 0 || px > 100 {
			return fmt.Errorf("bars %q: want on, off, snap, white or a height of 0 to 100", args[0])
		}
		n.Args = []float64{px, 1}
		if len(args) == 2 {
			sec, err := strconv.ParseFloat(args[1], 64)
			if err != nil || sec < 0 {
				return fmt.Errorf("bars %q: want the seconds as a number", args[1])
			}
			n.Args[1] = sec
		}
	}
	return nil
}

// parseNumbers reads exactly count numbers.
func parseNumbers(word string, args []string, count int) ([]float64, error) {
	if len(args) != count {
		return nil, fmt.Errorf("%s needs %d numbers", word, count)
	}
	out := make([]float64, count)
	for i, a := range args {
		v, err := strconv.ParseFloat(a, 64)
		if err != nil {
			return nil, fmt.Errorf("%s %q: want a number", word, a)
		}
		out[i] = v
	}
	return out, nil
}

// parsePunch reads `punch 2|3 cx cy` or `punch off`: a hard cut to a
// whole-number zoom centred on (cx, cy).
func parsePunch(n *Node, args []string) error {
	n.Op = OpPunch
	if len(args) == 1 && args[0] == "off" {
		n.Args = []float64{1, screenW / 2, screenH / 2}
		return nil
	}
	nums, err := parseNumbers("punch", args, 3)
	if err != nil {
		return err
	}
	if nums[0] != 2 && nums[0] != 3 {
		return fmt.Errorf("punch %v: the zoom is a whole number, 2 or 3", nums[0])
	}
	n.Args = nums
	return nil
}

// parsePush reads `push scale cx cy seconds` or `push off`: a slow zoom.
func parsePush(n *Node, args []string) error {
	n.Op = OpPush
	if len(args) == 1 && args[0] == "off" {
		n.Args = []float64{1, screenW / 2, screenH / 2, 0}
		return nil
	}
	nums, err := parseNumbers("push", args, 4)
	if err != nil {
		return err
	}
	if nums[0] < 1 || nums[0] > 2 || nums[3] <= 0 {
		return fmt.Errorf("push: the scale is 1 to 2 and the time above 0")
	}
	n.Args = nums
	return nil
}

// parsePan reads `pan x [seconds]`: the camera glides to x (no seconds = a cut).
func parsePan(n *Node, args []string) error {
	n.Op = OpPan
	if len(args) < 1 || len(args) > 2 {
		return fmt.Errorf("pan is `pan x [seconds]`")
	}
	nums, err := parseNumbers("pan", args, len(args))
	if err != nil {
		return err
	}
	n.Args = []float64{nums[0], 0}
	if len(nums) == 2 {
		n.Args[1] = nums[1]
	}
	return nil
}

// ---- the Wake ending's wash ----

// WashLevel says how much of a 2x2 dither is light grey at a progress from 0
// to 1: 0 (none), 1 (one pixel in four), 2 (a checker) or 3 (three in four).
func WashLevel(progress float64) int { return min(3, max(0, int(progress*4))) }

// WashPixels is the RGBA picture of a dither level: every 2x2 block has
// `level` pixels of c, filled in this order (0,0), (1,1), (1,0).
func WashPixels(level, w, h int, c [4]byte) []byte {
	order := [3][2]int{{0, 0}, {1, 1}, {1, 0}}
	pix := make([]byte, w*h*4)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			for i := 0; i < level; i++ {
				if x%2 == order[i][0] && y%2 == order[i][1] {
					copy(pix[(y*w+x)*4:], c[:])
				}
			}
		}
	}
	return pix
}

// ---- Forever's drop ----

// The drop plays once: the tag "sway" (the rope taut, 3 frames), then "break"
// (the pipe cracks, snaps, falls), then the last frame stays. Their lengths
// are the frame times in cg_forever_drop.json: 3 x 150 ms and 3 x 100 ms.
const (
	dropSway  = 450 * time.Millisecond
	dropBreak = 300 * time.Millisecond
)

// DropFrame is the frame of the drop shot elapsed after it came up.
func DropFrame(m *sprites.Meta, elapsed time.Duration) int {
	switch {
	case elapsed < dropSway:
		return m.FrameAt("sway", elapsed)
	case elapsed < dropSway+dropBreak:
		return m.FrameAt("break", elapsed-dropSway)
	}
	return m.Tags["break"].To // it hangs there: the cut and the splash land on it
}
