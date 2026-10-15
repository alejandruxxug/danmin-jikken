package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"

	"danmingame"
)

// Config is what render needs from the command line.
type Config struct {
	ShotsFile string
	OutDir    string // trailer_out
	Name      string // the output is <OutDir>/<Name>.mp4
	GameBin   string
	GameDir   string // the game's working folder
	SoundDir  string
	SpriteDir string
	FontPath  string
	FFmpeg    string
	Reuse     bool  // keep shot folders that are already recorded
	Reduced   bool  // the backup file, with the -reduce-flashing strobe
	Only      []int // render only these frames, as PNGs in <OutDir>/<Name>_frames, and no mp4
}

// readAudioLog reads the audio.json of a recorded folder.
func readAudioLog(dir string) (game.AudioLog, error) {
	var log game.AudioLog
	data, err := os.ReadFile(filepath.Join(dir, "audio.json"))
	if err != nil {
		return log, err
	}
	if err := json.Unmarshal(data, &log); err != nil {
		return log, fmt.Errorf("%s: %w", dir, err)
	}
	return log, nil
}

// recorded says whether a shot folder already holds the whole recording.
func recorded(dir string, s Shot) bool {
	log, err := readAudioLog(dir)
	if err != nil || log.Frames != s.Frames {
		return false
	}
	_, err = os.Stat(filepath.Join(dir, game.FrameFile(s.Frames)))
	return err == nil
}

func recordShot(cfg Config, s Shot, dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	args := append(s.GameArgs(), "-record", dir, "-record-frames", fmt.Sprint(s.Frames), "-record-from", fmt.Sprint(s.From))
	cmd := exec.Command(cfg.GameBin, args...)
	cmd.Dir = cfg.GameDir
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("the game (%v): %w", args, err)
	}
	return nil
}

// Render records the shots, composes every frame, mixes the sound and encodes the mp4.
func Render(cfg Config) error {
	f, err := os.Open(cfg.ShotsFile)
	if err != nil {
		return err
	}
	shots, err := ParseShots(f, filepath.Dir(cfg.ShotsFile))
	f.Close()
	if err != nil {
		return fmt.Errorf("%s: %w", cfg.ShotsFile, err)
	}
	if cfg.GameBin, err = filepath.Abs(cfg.GameBin); err != nil { // the game runs in another folder
		return err
	}
	shotsDir, _ := filepath.Abs(filepath.Join(cfg.OutDir, "shots"))
	logs := map[string]game.AudioLog{}
	for i, s := range shots {
		dir := filepath.Join(shotsDir, s.Name)
		if cfg.Reuse && recorded(dir, s) {
			fmt.Printf("shot %d %s: reused\n", i+1, s.Name)
		} else {
			fmt.Printf("shot %d %s: recording frames %d to %d\n", i+1, s.Name, s.From, s.Frames)
			if err := recordShot(cfg, s, dir); err != nil {
				return fmt.Errorf("shot %q: %w", s.Name, err)
			}
		}
		if logs[s.Name], err = readAudioLog(dir); err != nil {
			return err
		}
	}

	fx, err := LoadSheet(cfg.SpriteDir, "trailer_fx")
	if err != nil {
		return err
	}
	cards, err := LoadSheet(cfg.SpriteDir, "trailer_cards")
	if err != nil {
		return err
	}
	glyphs, err := LoadGlyphs(cfg.FontPath)
	if err != nil {
		return err
	}
	tl, err := NewTimeline(fx, cards, cfg.Reduced)
	if err != nil {
		return err
	}
	if err := checkFrames(tl, shotsDir); err != nil {
		return err
	}
	comp := &Composer{Sheets: map[string]*Sheet{SheetFX: fx, SheetCards: cards}, Glyphs: glyphs,
		Frames: &DiskFrames{Root: shotsDir}}

	if len(cfg.Only) > 0 {
		return writeFrames(comp, tl, cfg)
	}
	work := filepath.Join(cfg.OutDir, cfg.Name)
	if err := os.MkdirAll(work, 0o755); err != nil {
		return err
	}
	wav, err := mixToWAV(cfg, work, logs)
	if err != nil {
		return err
	}
	mp4, _ := filepath.Abs(filepath.Join(cfg.OutDir, cfg.Name+".mp4"))
	return encode(cfg, comp, tl, wav, mp4)
}

// checkFrames makes sure every recorded frame the timeline asks for exists,
// before the long part starts.
func checkFrames(tl *Timeline, shotsDir string) error {
	seen := map[string]bool{}
	for f := 0; f < TotalFrames; f++ {
		for _, l := range tl.Plan(f) {
			if l.Kind != LayerGame {
				continue
			}
			path := filepath.Join(shotsDir, l.Shot, game.FrameFile(l.Frame))
			if seen[path] {
				continue
			}
			seen[path] = true
			if _, err := os.Stat(path); err != nil {
				return fmt.Errorf("trailer frame %d needs %s/%s: %w (check frames and from in shots.txt)", f, l.Shot, game.FrameFile(l.Frame), err)
			}
		}
	}
	return nil
}

// mixToWAV makes the trailer's whole sound and writes it, with checks on the
// numbers (the sound cannot be listened to here): the drop against the slam,
// and the level of each part.
func mixToWAV(cfg Config, work string, logs map[string]game.AudioLog) (string, error) {
	plan := BuildAudioPlan(cfg.Reduced)
	a, err := MixTrailer(plan, TotalFrames, DiskLoader(cfg.SoundDir), logs)
	if err != nil {
		return "", err
	}
	wav, _ := filepath.Abs(filepath.Join(work, "audio.wav"))
	if err := WriteWAV(wav, a); err != nil {
		return "", err
	}
	fmt.Printf("sound: %d frames (%.2f s), %v\n", a.Frames()/SamplesPerFrame, float64(a.Frames())/SampleRate, Measure(a))
	fmt.Printf("sync: Ironbound's drop is on trailer frame %d, the dagger lands on frame %d (the title's frame %d)\n",
		IronboundDropFrame(), slamFrame, titleSlamFrame)
	for _, part := range []struct {
		name     string
		from, to int
	}{{"jingle", 30, t10Start}, {"silence", t10Start, t11Start - 20}, {"night", t12Start + 90, t16Start},
		{"ironbound", slamFrame, ironTo}, {"after the cut", ironTo + 5, 3975}} {
		seg := Audio{Samples: a.Samples[2*part.from*SamplesPerFrame : 2*part.to*SamplesPerFrame]}
		fmt.Printf("  %-14s frames %4d-%4d: %v\n", part.name, part.from, part.to, Measure(seg))
	}
	return wav, nil
}

// writeFrames renders the frames in cfg.Only to PNG files, to look at them.
func writeFrames(c *Composer, tl *Timeline, cfg Config) error {
	dir := filepath.Join(cfg.OutDir, cfg.Name+"_frames")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	dst := image.NewRGBA(image.Rect(0, 0, OutW, OutH))
	for _, f := range cfg.Only {
		if f < 0 || f >= TotalFrames {
			return fmt.Errorf("frame %d: the trailer has frames 0 to %d", f, TotalFrames-1)
		}
		if err := c.Compose(dst, tl.Plan(f)); err != nil {
			return fmt.Errorf("frame %d: %w", f, err)
		}
		path := filepath.Join(dir, fmt.Sprintf("f%04d.png", f))
		out, err := os.Create(path)
		if err != nil {
			return err
		}
		if err := png.Encode(out, dst); err != nil {
			out.Close()
			return err
		}
		out.Close()
	}
	fmt.Println("wrote", len(cfg.Only), "frames to", dir)
	return nil
}

// encode composes every frame, a batch at a time on all the cores, and feeds
// them in order to ffmpeg, which makes the mp4: H.264, yuv420p, 60 fps, CRF 16, AAC 192k.
func encode(cfg Config, c *Composer, tl *Timeline, wav, out string) error {
	cmd := exec.Command(cfg.FFmpeg, "-y", "-loglevel", "error",
		"-f", "rawvideo", "-pix_fmt", "rgba", "-video_size", fmt.Sprintf("%dx%d", OutW, OutH), "-framerate", "60", "-i", "-",
		"-i", wav,
		"-vf", "scale=in_range=full:out_range=tv:out_color_matrix=bt709,format=yuv420p",
		"-colorspace", "bt709", "-color_primaries", "bt709", "-color_trc", "bt709",
		"-c:v", "libx264", "-preset", "slow", "-crf", "16", "-r", "60",
		"-c:a", "aac", "-b:a", "192k", "-shortest", "-movflags", "+faststart", out)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("ffmpeg: %w", err)
	}
	w := bufio.NewWriterSize(stdin, 1<<22)
	err = composeAll(c, tl, w)
	if flushErr := w.Flush(); err == nil {
		err = flushErr
	}
	stdin.Close()
	if waitErr := cmd.Wait(); err == nil && waitErr != nil {
		err = fmt.Errorf("ffmpeg: %w", waitErr)
	}
	if err != nil {
		return err
	}
	fmt.Println("wrote", out)
	return nil
}

// composeAll draws the trailer's frames in batches (one goroutine per core,
// each with its own picture) and writes each batch in frame order.
func composeAll(c *Composer, tl *Timeline, w io.Writer) error {
	workers := runtime.NumCPU()
	bufs := make([]*image.RGBA, workers)
	for i := range bufs {
		bufs[i] = image.NewRGBA(image.Rect(0, 0, OutW, OutH))
	}
	for start := 0; start < TotalFrames; start += workers {
		n := min(workers, TotalFrames-start)
		errs := make([]error, n)
		var wg sync.WaitGroup
		for i := 0; i < n; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := c.Compose(bufs[i], tl.Plan(start+i)); err != nil {
					errs[i] = fmt.Errorf("frame %d: %w", start+i, err)
				}
			}()
		}
		wg.Wait()
		for i := 0; i < n; i++ {
			if errs[i] != nil {
				return errs[i]
			}
			if _, err := w.Write(bufs[i].Pix); err != nil {
				return fmt.Errorf("ffmpeg: %w", err)
			}
		}
		if start%(workers*20) == 0 {
			fmt.Printf("\rcomposing: frame %d of %d", start, TotalFrames)
		}
	}
	fmt.Println()
	return nil
}
