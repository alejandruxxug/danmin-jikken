// Command danmin opens the game window. All of the work is in the game
// package; this file only reads the flags and starts it.
package main

import (
	"fmt"
	"os"

	game "danmingame"
)

func main() {
	opts, err := game.ParseOptions(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "danmin:", err)
		os.Exit(2)
	}
	if err := game.Run(opts); err != nil {
		fmt.Fprintln(os.Stderr, "danmin:", err)
		os.Exit(1)
	}
}
