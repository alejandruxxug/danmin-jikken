// Package night is the night shift: a visual novel. It reads a script file,
// runs it (text, choices, flags, effects) and draws it with the exported
// Mayu / Hana sprites. It does not import the game package, so the game can
// import it; the game's night scene is the thin glue between the two.
package night

import "sort"

// Flags is the one store of story flags (Script -> Flags). A choice in the
// night sets one; the day shift reads them later (for example the day 2
// memo rules). It is a map so the agent backdoor and the dump can use it
// like any other map.
type Flags map[string]bool

// Set turns a flag on.
func (f Flags) Set(name string) { f[name] = true }

// Has says whether a flag is on.
func (f Flags) Has(name string) bool { return f[name] }

// Names lists the flags that are on, sorted so dumps and tests compare well.
func (f Flags) Names() []string {
	names := []string{}
	for name, on := range f {
		if on {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}
