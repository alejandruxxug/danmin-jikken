package game

// Record mode (-agent -record <dir>): the trailer is rendered by the game
// itself. Every frame is saved as a 1920x1080 PNG, exactly what a player sees
// on a screen that size (the 480x270 canvas x4, through the CRT when the scene
// has it), and the sound events go to audio.json for tools/trailer to mix.
//
// The run is fixed-step: one update per saved frame, whatever the machine's
// speed. Ebitengine may run several updates before a draw when it falls
// behind; the recorder refuses that by skipping an update until the last one
// has been drawn and saved.

import (
	"encoding/json"
	"fmt"
	"image/png"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"

	"danmingame/sound"
)

// Record output size: the canvas x4.
const (
	RecordScale = 4
	RecordW     = ScreenW * RecordScale
	RecordH     = ScreenH * RecordScale
	RecordFPS   = 60
)

// FrameFile is the file name of frame n (the first frame is 1).
func FrameFile(n int) string { return fmt.Sprintf("f%06d.png", n) }

// FrameSeconds is when frame n is shown, in seconds from the start: frame 1
// is at 0 s. A sound event stamped with frame n starts at the same time.
func FrameSeconds(n int) float64 { return float64(n-1) / RecordFPS }

// AudioLog is audio.json: the sound events of the whole run.
type AudioLog struct {
	FPS    int           `json:"fps"`
	Frames int           `json:"frames"`
	Events []sound.Event `json:"events"`
}

// recorder saves the frames.
type recorder struct {
	dir     string
	frame   int // frames updated so far (= saved, once drawn)
	drawn   bool
	started bool          // an update has run
	screen  *ebiten.Image // the 1920x1080 picture
	err     error
}

func newRecorder(dir string) (*recorder, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("record: %w", err)
	}
	return &recorder{dir: dir, drawn: true, screen: ebiten.NewImage(RecordW, RecordH)}, nil
}

// mustWait says whether the last update is still waiting for its frame to be
// drawn and saved; the game skips its update meanwhile.
func (r *recorder) mustWait() bool { return r.started && !r.drawn }

// startFrame is called by the update that goes ahead.
func (r *recorder) startFrame() {
	r.started, r.drawn = true, false
	r.frame++
}

// save writes the picture of the frame that was just updated.
func (r *recorder) save(g *Game) {
	if !r.started || r.drawn || r.err != nil {
		return
	}
	r.drawn = true
	g.putOnScreen(r.screen, false)
	img := readScreen(r.screen)
	f, err := os.Create(filepath.Join(r.dir, FrameFile(r.frame)))
	if err != nil {
		r.err = fmt.Errorf("record: %w", err)
		return
	}
	defer f.Close()
	enc := png.Encoder{CompressionLevel: png.BestSpeed}
	if err := enc.Encode(f, img); err != nil {
		r.err = fmt.Errorf("record: frame %d: %w", r.frame, err)
	}
}

// finish writes audio.json.
func (r *recorder) finish(events []sound.Event) error {
	if r.err != nil {
		return r.err
	}
	if events == nil {
		events = []sound.Event{} // "events": [] and not null
	}
	log := AudioLog{FPS: RecordFPS, Frames: r.frame, Events: events}
	data, err := json.MarshalIndent(log, "", "  ")
	if err != nil {
		return fmt.Errorf("record: audio.json: %w", err)
	}
	if err := os.WriteFile(filepath.Join(r.dir, "audio.json"), data, 0o644); err != nil {
		return fmt.Errorf("record: audio.json: %w", err)
	}
	return nil
}
