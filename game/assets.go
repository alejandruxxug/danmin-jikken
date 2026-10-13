package game

import (
	"embed"
	"io/fs"
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
