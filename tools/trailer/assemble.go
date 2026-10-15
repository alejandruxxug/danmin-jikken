package main

import (
	"encoding/json"
	"fmt"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"

	"danmingame"
)

// FrameRef says which recorded frame is final frame number j (the plan the
// assembler follows) and how bright it is after the fades.
type FrameRef struct {
	Shot   int     // index in the shot list
	Frame  int     // frame number in the shot's folder, from 1
	Factor float64 // 1 = as recorded
}

// Plan lists every frame of the trailer in order.
func Plan(shots []Shot) []FrameRef {
	var plan []FrameRef
	for i, s := range shots {
		for f := 1; f <= s.Frames; f++ {
			plan = append(plan, FrameRef{i, f, FadeFactor(float64(f-1), s.Frames, s.FadeIn, s.FadeOut)})
		}
	}
	return plan
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

// mixFolder mixes the audio.json of a recorded folder and writes its audio.wav.
func mixFolder(dir, soundDir string) (Audio, error) {
	log, err := readAudioLog(dir)
	if err != nil {
		return Audio{}, err
	}
	a, err := Mix(log.Events, log.Frames, DiskLoader(soundDir), MixOptions{})
	if err != nil {
		return Audio{}, err
	}
	return a, WriteWAV(filepath.Join(dir, "audio.wav"), a)
}

// Config is what render needs from the command line.
type Config struct {
	ShotsFile string
	OutDir    string // trailer_out
	Name      string // the output is <OutDir>/<Name>.mp4
	GameBin   string
	GameDir   string // the game's working folder
	SoundDir  string
	FFmpeg    string
	Reuse     bool // keep shot folders that are already recorded
}

// Render records every shot, mixes the sound, and encodes the mp4.
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
	work := filepath.Join(cfg.OutDir, cfg.Name)
	var dirs []string
	for i, s := range shots {
		dir, _ := filepath.Abs(filepath.Join(work, "shots", fmt.Sprintf("%02d_%s", i+1, s.Name)))
		dirs = append(dirs, dir)
		if cfg.Reuse && recorded(dir, s.Frames) {
			fmt.Printf("shot %d %s: reused\n", i+1, s.Name)
			continue
		}
		fmt.Printf("shot %d %s: recording %d frames\n", i+1, s.Name, s.Frames)
		if err := recordShot(cfg, s, dir); err != nil {
			return fmt.Errorf("shot %q: %w", s.Name, err)
		}
	}

	// The sound: each shot on its own, then all of them on one timeline.
	loader := DiskLoader(cfg.SoundDir)
	beds := map[string]Audio{}
	starts := BedStarts(shots)
	var all Audio
	for i, s := range shots {
		log, err := readAudioLog(dirs[i])
		if err != nil {
			return err
		}
		if log.Frames != s.Frames {
			return fmt.Errorf("shot %q: recorded %d frames, wanted %d", s.Name, log.Frames, s.Frames)
		}
		own, err := mixFolder(dirs[i], cfg.SoundDir) // the shot as the game sounded, in audio.wav
		if err != nil {
			return fmt.Errorf("shot %q: %w", s.Name, err)
		}
		a, err := Mix(log.Events, log.Frames, loader, MixOptions{DropMusic: s.Bed != nil})
		if err != nil {
			return fmt.Errorf("shot %q: %w", s.Name, err)
		}
		if !s.Cut && s.FadeOut == 0 {
			TaperEnd(a)
		}
		if s.Bed != nil {
			if err := addBed(a, s.Bed, starts[i], cfg.SoundDir, beds); err != nil {
				return fmt.Errorf("shot %q: %w", s.Name, err)
			}
		}
		FadeAudio(a, s.Frames, s.FadeIn, s.FadeOut)
		fmt.Printf("shot %d %s: game %v; in the trailer %v\n", i+1, s.Name, Measure(own), Measure(a))
		all.Samples = append(all.Samples, a.Samples...)
	}
	fmt.Printf("trailer: %d frames (%.2f s), %v\n", all.Frames()/SamplesPerFrame, float64(all.Frames())/SampleRate, Measure(all))
	wavPath, _ := filepath.Abs(filepath.Join(work, "audio.wav"))
	if err := WriteWAV(wavPath, all); err != nil {
		return err
	}

	// The picture: the frames in order, as links (or a darkened copy where a fade is).
	framesDir, _ := filepath.Abs(filepath.Join(work, "frames"))
	if err := os.RemoveAll(framesDir); err != nil {
		return err
	}
	if err := os.MkdirAll(framesDir, 0o755); err != nil {
		return err
	}
	for j, ref := range Plan(shots) {
		src := filepath.Join(dirs[ref.Shot], game.FrameFile(ref.Frame))
		dst := filepath.Join(framesDir, game.FrameFile(j+1))
		if err := placeFrame(src, dst, ref.Factor); err != nil {
			return err
		}
	}
	mp4, _ := filepath.Abs(filepath.Join(cfg.OutDir, cfg.Name+".mp4"))
	return encode(cfg.FFmpeg, framesDir, wavPath, mp4)
}

// recorded says whether a shot folder already holds a whole recording.
func recorded(dir string, frames int) bool {
	log, err := readAudioLog(dir)
	if err != nil || log.Frames != frames {
		return false
	}
	_, err = os.Stat(filepath.Join(dir, game.FrameFile(frames)))
	return err == nil
}

func recordShot(cfg Config, s Shot, dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	args := append(s.GameArgs(), "-record", dir, "-record-frames", fmt.Sprint(s.Frames))
	cmd := exec.Command(cfg.GameBin, args...)
	cmd.Dir = cfg.GameDir
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("the game (%v): %w", args, err)
	}
	return nil
}

// addBed mixes the shot's slice of the music bed into a. startSec is how far
// into the bed file the shot begins.
func addBed(a Audio, bed *Bed, startSec float64, soundDir string, cache map[string]Audio) error {
	clip, ok := cache[bed.File]
	if !ok {
		path, err := ResolveSound(soundDir, bed.File)
		if err != nil {
			return err
		}
		if clip, err = DecodeFile(path); err != nil {
			return err
		}
		cache[bed.File] = clip
	}
	from := int(startSec * SampleRate)
	if from >= clip.Frames() {
		return nil // the bed has ended: silence
	}
	slice := Audio{Samples: clip.Samples[2*from:]}
	addAt(a, slice, 0, float32(bed.Gain))
	return nil
}

// placeFrame puts a recorded frame at its place in the final sequence.
func placeFrame(src, dst string, factor float64) error {
	if factor >= 1 {
		return os.Symlink(src, dst)
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	img, err := png.Decode(in)
	if err != nil {
		return fmt.Errorf("%s: %w", src, err)
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	enc := png.Encoder{CompressionLevel: png.BestSpeed}
	return enc.Encode(out, Darken(img, factor))
}

// encode makes the mp4: H.264, yuv420p, 60 fps, CRF 16, AAC 192k.
func encode(ffmpeg, framesDir, wav, out string) error {
	cmd := exec.Command(ffmpeg, "-y", "-loglevel", "error",
		"-framerate", "60", "-i", filepath.Join(framesDir, "f%06d.png"), "-i", wav,
		"-c:v", "libx264", "-preset", "slow", "-crf", "16", "-pix_fmt", "yuv420p", "-r", "60",
		"-c:a", "aac", "-b:a", "192k", "-movflags", "+faststart", out)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg: %w", err)
	}
	fmt.Println("wrote", out)
	return nil
}
