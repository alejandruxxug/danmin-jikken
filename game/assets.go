package game

import (
	"embed"
	"io/fs"

	"danmingame/sprites"
)

// The exported spritesheets are baked into the binary. "all:" also takes
// .gitkeep, so the folder exists before the first sheet is exported.
//
//go:embed all:assets/sprites
var spriteFiles embed.FS

// SpriteFS is the assets/sprites folder as a file system root, so
// sprites.Load(fsys, "mina") finds mina.json and mina.png.
func SpriteFS() (fs.FS, error) {
	return fs.Sub(spriteFiles, "assets/sprites")
}

// Line boil speed (Alejo, 2026-10-13): the people twitch at the speed in
// the sheet (125 ms a frame), the room breathes slower. boilScales is the
// one place to tune it: a sheet listed here plays its boil that many times
// slower, a sheet that is not listed plays at the normal speed.
// The rule is "everything that is not a character is slow".
var boilScales = map[string]float64{
	"reception_bg":   3,
	"desk":           3,
	"computer":       3,
	"ui_icons":       3,
	"ui_icons_m4":    3,
	"desk_top":       3,
	"folder":         3,
	"doc_settler_id": 3,
	"doc_referral":   3,
	"doc_sleep_log":  3,
	"doc_wristband":  3,
	"doc_citation":   3,
	// cecily, waiting_fisher and hallway_sleeper are characters: normal speed.
	// screen_glass has one frame and never boils.
}

// LoadSheet loads <name>.json and <name>.png and applies the boil speed
// from boilScales.
func LoadSheet(fsys fs.FS, name string) (*sprites.Sheet, error) {
	sheet, err := sprites.Load(fsys, name)
	if err != nil {
		return nil, err
	}
	sheet.BoilScale = boilScales[name]
	return sheet, nil
}
