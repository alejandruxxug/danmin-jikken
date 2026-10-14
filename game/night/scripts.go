package night

import (
	"embed"
	"fmt"
	"strings"
)

// The night scripts are baked into the binary, like the sprites.
//
//go:embed scripts/*.txt
var scriptFiles embed.FS

// Names of the scripts, as the scene ids use them (night.prologue, night.1).
const (
	Prologue = "prologue"
	Night1   = "night1"
	Night2   = "night2"
	Night3   = "night3"
)

// EndingScript is the name of an ending's script (ending_sleep, ...).
func EndingScript(ending string) string { return "ending_" + ending }

// LoadScript parses one of the embedded scripts by name.
func LoadScript(name string) (*Script, error) {
	f, err := scriptFiles.Open("scripts/" + name + ".txt")
	if err != nil {
		return nil, fmt.Errorf("night: no script %q: %w", name, err)
	}
	defer f.Close()
	return Parse(name+".txt", f)
}

// ScriptNames lists the embedded scripts, for the tests.
func ScriptNames() ([]string, error) {
	entries, err := scriptFiles.ReadDir("scripts")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		names = append(names, strings.TrimSuffix(e.Name(), ".txt"))
	}
	return names, nil
}
