// Command trailer builds the trailer from recordings the game makes with
// -agent -record (game/record.go). Two jobs:
//
//	trailer mix <dir>        mix <dir>/audio.json into <dir>/audio.wav
//	trailer render [flags]   record every shot of a shot list, mix, and encode the mp4
//
// See "Rendering the trailer" in docs/Danmin Jikken – Trailer.md.
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	var err error
	switch os.Args[1] {
	case "mix":
		err = runMix(os.Args[2:])
	case "render":
		err = runRender(os.Args[2:])
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "trailer:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: trailer mix <dir> [-sound dir] | trailer render -shots file [flags]")
	os.Exit(2)
}

func runMix(args []string) error {
	fs := flag.NewFlagSet("mix", flag.ContinueOnError)
	soundDir := fs.String("sound", "../game/assets/sound", "the game's sound folder")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("mix needs one folder")
	}
	a, err := mixFolder(fs.Arg(0), *soundDir)
	if err != nil {
		return err
	}
	fmt.Printf("%s/audio.wav: %.2f s, %v\n", fs.Arg(0), float64(a.Frames())/SampleRate, Measure(a))
	return nil
}

func runRender(args []string) error {
	fs := flag.NewFlagSet("render", flag.ContinueOnError)
	var cfg Config
	fs.StringVar(&cfg.ShotsFile, "shots", "tools/trailer/shots.txt", "the shot list")
	fs.StringVar(&cfg.OutDir, "out", "trailer_out", "output folder")
	fs.StringVar(&cfg.Name, "name", "trailer", "the output is <out>/<name>.mp4")
	fs.StringVar(&cfg.GameBin, "game", "trailer_out/bin/danmin", "the game binary")
	fs.StringVar(&cfg.GameDir, "gamedir", "game", "the folder the game runs in")
	fs.StringVar(&cfg.SoundDir, "sound", "game/assets/sound", "the game's sound folder")
	fs.StringVar(&cfg.FFmpeg, "ffmpeg", "ffmpeg", "ffmpeg")
	fs.BoolVar(&cfg.Reuse, "reuse", false, "keep shot folders that are already recorded")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return Render(cfg)
}
