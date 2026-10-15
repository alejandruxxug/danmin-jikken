// Command trailer renders the trailer: the game records its shots in record
// mode (game/record.go), and this program cuts them together with the cards
// and the motion graphics, mixes the sound and runs ffmpeg.
//
//	trailer render [flags]
//
// See "Rendering the trailer" in docs/Danmin Jikken – Trailer.md.
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] != "render" {
		fmt.Fprintln(os.Stderr, "usage: trailer render [-shots file] [-name name] [-reduced] [-reuse] [-only 0,142,146]")
		os.Exit(2)
	}
	if err := runRender(os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, "trailer:", err)
		os.Exit(1)
	}
}

func runRender(args []string) error {
	fs := flag.NewFlagSet("render", flag.ContinueOnError)
	var cfg Config
	var only string
	fs.StringVar(&cfg.ShotsFile, "shots", "tools/trailer/shots.txt", "the shot list")
	fs.StringVar(&cfg.OutDir, "out", "trailer_out", "output folder")
	fs.StringVar(&cfg.Name, "name", "trailer", "the output is <out>/<name>.mp4")
	fs.StringVar(&cfg.GameBin, "game", "trailer_out/bin/danmin", "the game binary")
	fs.StringVar(&cfg.GameDir, "gamedir", "game", "the folder the game runs in")
	fs.StringVar(&cfg.SoundDir, "sound", "game/assets/sound", "the game's sound folder")
	fs.StringVar(&cfg.SpriteDir, "sprites", "game/assets/sprites", "the exported spritesheets")
	fs.StringVar(&cfg.FontPath, "font", "game/fonts/Silkscreen-Regular.ttf", "the Silkscreen font file")
	fs.StringVar(&cfg.FFmpeg, "ffmpeg", "ffmpeg", "ffmpeg")
	fs.BoolVar(&cfg.Reuse, "reuse", false, "keep shot folders that are already recorded")
	fs.BoolVar(&cfg.Reduced, "reduced", false, "the backup file: the -reduce-flashing strobe")
	fs.StringVar(&only, "only", "", "render only these frames (comma separated) to PNG files, and no mp4")
	if err := fs.Parse(args); err != nil {
		return err
	}
	for _, field := range strings.Split(only, ",") {
		if field = strings.TrimSpace(field); field != "" {
			n, err := strconv.Atoi(field)
			if err != nil {
				return fmt.Errorf("-only %q: want frame numbers", field)
			}
			cfg.Only = append(cfg.Only, n)
		}
	}
	return Render(cfg)
}
