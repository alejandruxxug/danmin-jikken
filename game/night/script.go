package night

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// A script is a plain text file, one command per line. Blank lines and
// everything after a '#' are ignored. The files are in scripts/ and the
// format is explained in Architecture -> Night script format.
//
//	label name                    a place to jump to
//	jump name                     go there
//	if flag jump name             jump when the flag is on (ifnot: when it is off)
//	set flag                      turn a flag on
//	bg station                    background (station, ceiling, black, white, or any
//	                              other name: drawn as a marked placeholder)
//	prop can                      put a prop on the counter (coffee, can, syringe,
//	                              note, monitor); "prop clear" takes them all off
//	show mayu                     a character steps in (one at a time)
//	hide                          she leaves
//	faces                         both faces lean in, upside down (the Prologue)
//	pose mayu smile arms=clasp    change a pose without a line
//	Mayu (smile): Hello!          she says a line; the expression is optional
//	Mayu (smile eyes=half): Hi.   ... and single parts can be changed
//	~ Cold.                       your thought (no name plate)
//	choice                        a menu; the lines under it start with "option"
//	option Text | set flag | jump name
//	fx eyes_close                 a screen effect (see knownFX)
//	sfx name / music name         sound cues (the sound comes in M7: only logged)
//	speed 12                      typewriter speed in letters per second
//	wait 1.5                      pause, in seconds (a click cuts it short)
//	lock 5                        pause that clicks cannot cut short
//	auto                          the next line goes on by itself when typed
//	route                         set route_mayu / route_hana / route_nobody from the flags
//	end day                       the night is over; the game decides what follows
//
// Every error names the file and the line.

// Op is the kind of a command.
type Op int

const (
	OpSay Op = iota
	OpThink
	OpChoice
	OpJump
	OpIf
	OpSet
	OpBG
	OpProp
	OpShow
	OpHide
	OpFaces
	OpPose
	OpFx
	OpSfx
	OpMusic
	OpSpeed
	OpWait
	OpEnd
	OpLock
	OpAuto
	OpRoute
)

// Part is one "key=value" change to a pose.
type Part struct{ Key, Value string }

// Option is one choice of a menu.
type Option struct {
	Text string
	Set  []string // flags the choice turns on
	Jump string   // label the choice goes to ("" = go on with the next line)
}

// Node is one command of the script.
type Node struct {
	Op      Op
	Line    int    // the line of the file, for errors and the backdoor
	Name    string // say/pose/show: the speaker; bg, prop, fx, sfx, music, end: the name; set, if: the flag
	Expr    string // say/pose: the expression ("" = keep the pose)
	Parts   []Part // say/pose: parts changed after the expression
	Text    string // say/think: the words; jump, if: the label
	Options []Option
	Not     bool    // if: jump when the flag is OFF
	Number  float64 // speed: letters per second; wait: seconds
}

// Script is a parsed file.
type Script struct {
	Name   string
	Nodes  []Node
	Labels map[string]int // label -> index of the node that follows it
}

// knownFX are the screen effects the painter understands.
var knownFX = map[string]bool{
	"fade_in":    true, // black fades away (1.5 s)
	"eyes_close": true, // eyelids close from top and bottom (1.2 s)
	"red_pixel":  true, // the first red of the game: the 7th monitor line blinks red once
	"flat_line":  true, // Matty's monitor line goes flat
	"line_back":  true, // ... and comes back
	"seventh":    true, // the 7th (nameless) line appears on the monitor
	// Nights 2 and 3.
	"shake":        true, // the picture shakes for 0.75 s
	"redscale":     true, // Hana's room: grey turns red from the floor up
	"redscale_off": true, // ... and drains back to grey from the top down
	"eyes_slit":    true, // the eyelids close almost all the way (Forever)
	"splash":       true, // blood on the camera (the Forever cut); the stain stays after it
	"dim":          true, // the screen dims (the game falls asleep, 3:17)
	"figure":       true, // kanashibari: a faceless figure in the doorway
	"figure_gone":  true, // ... gone
	"clear":        true, // takes every running effect away (a cut)
}

// knownProps are the things on the counter and the monitor.
var knownProps = map[string]bool{"coffee": true, "can": true, "syringe": true, "note": true, "monitor": true, "clear": true, "envelope": true}

// Parse reads a script. name is only used in error messages.
func Parse(name string, r io.Reader) (*Script, error) {
	s := &Script{Name: name, Labels: make(map[string]int)}
	type ref struct {
		label string
		line  int
	}
	var refs []ref // jumps to check once every label is known
	scanner := bufio.NewScanner(r)
	var choice *Node // the open menu, while its option lines are read
	fail := func(line int, format string, args ...any) error {
		return fmt.Errorf("%s line %d: %s", name, line, fmt.Sprintf(format, args...))
	}
	for line := 1; scanner.Scan(); line++ {
		text := stripComment(scanner.Text())
		if text == "" {
			continue
		}
		word, rest, _ := strings.Cut(text, " ")
		rest = strings.TrimSpace(rest)

		if word == "option" {
			if choice == nil {
				return nil, fail(line, "option outside a choice")
			}
			opt, err := parseOption(rest)
			if err != nil {
				return nil, fail(line, "%v", err)
			}
			if opt.Jump != "" {
				refs = append(refs, ref{opt.Jump, line})
			}
			choice.Options = append(choice.Options, opt)
			continue
		}
		if choice != nil { // any other line closes the menu
			if len(choice.Options) < 2 {
				return nil, fail(choice.Line, "a choice needs at least two options")
			}
			s.Nodes = append(s.Nodes, *choice)
			choice = nil
		}

		node := Node{Line: line}
		switch {
		case word == "~":
			if rest == "" {
				return nil, fail(line, "a thought needs some text")
			}
			node.Op, node.Text = OpThink, rest
		case word[0] >= 'A' && word[0] <= 'Z':
			if err := parseSay(&node, text); err != nil {
				return nil, fail(line, "%v", err)
			}
		case word == "label":
			if rest == "" || strings.Contains(rest, " ") {
				return nil, fail(line, "label needs one name")
			}
			if _, dup := s.Labels[rest]; dup {
				return nil, fail(line, "label %q is defined twice", rest)
			}
			s.Labels[rest] = len(s.Nodes)
			continue
		case word == "choice":
			if rest != "" {
				return nil, fail(line, "choice takes no argument")
			}
			choice = &Node{Op: OpChoice, Line: line}
			continue
		default:
			if err := parseCommand(&node, word, rest); err != nil {
				return nil, fail(line, "%v", err)
			}
			if node.Op == OpJump || node.Op == OpIf {
				refs = append(refs, ref{node.Text, line})
			}
		}
		s.Nodes = append(s.Nodes, node)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("%s: reading: %w", name, err)
	}
	if choice != nil {
		if len(choice.Options) < 2 {
			return nil, fail(choice.Line, "a choice needs at least two options")
		}
		s.Nodes = append(s.Nodes, *choice)
	}
	for _, r := range refs {
		if _, ok := s.Labels[r.label]; !ok {
			return nil, fail(r.line, "jump to unknown label %q", r.label)
		}
	}
	return s, nil
}

// stripComment cuts a '#' comment off a line and trims it. A '#' only
// starts a comment at the start of the line or after a space, so it can
// still appear inside the words ("bed #4" is written "bed 4" anyway).
func stripComment(line string) string {
	if strings.HasPrefix(line, "#") {
		return ""
	}
	if i := strings.Index(line, " #"); i >= 0 {
		line = line[:i]
	}
	return strings.TrimSpace(line)
}

// parseOption reads "Text | set a,b | jump label".
func parseOption(rest string) (Option, error) {
	parts := strings.Split(rest, "|")
	opt := Option{Text: strings.TrimSpace(parts[0])}
	if opt.Text == "" {
		return opt, fmt.Errorf("an option needs its text first")
	}
	for _, part := range parts[1:] {
		key, value, _ := strings.Cut(strings.TrimSpace(part), " ")
		value = strings.TrimSpace(value)
		switch key {
		case "set":
			if value == "" {
				return opt, fmt.Errorf("set needs a flag name")
			}
			for _, flag := range strings.Split(value, ",") {
				opt.Set = append(opt.Set, strings.TrimSpace(flag))
			}
		case "jump":
			if value == "" || strings.Contains(value, " ") {
				return opt, fmt.Errorf("jump needs one label")
			}
			opt.Jump = value
		default:
			return opt, fmt.Errorf("unknown option part %q (want set or jump)", key)
		}
	}
	return opt, nil
}

// parseSay reads `Mayu (smile eyes=half): the words`.
func parseSay(n *Node, text string) error {
	head, words, ok := strings.Cut(text, ": ")
	if !ok {
		return fmt.Errorf("a line is `Name (expression): words`, got %q", text)
	}
	words = strings.TrimSpace(words)
	if words == "" {
		return fmt.Errorf("a line needs some words")
	}
	speaker, inside, hasParens := strings.Cut(head, "(")
	speaker = strings.TrimSpace(speaker)
	n.Op, n.Text = OpSay, words
	n.Name = strings.ToLower(speaker)
	if !hasParens {
		return nil
	}
	inside, closed := strings.CutSuffix(strings.TrimSpace(inside), ")")
	if !closed {
		return fmt.Errorf("missing ')' in %q", head)
	}
	if !IsCharacter(n.Name) {
		return fmt.Errorf("%q has no sprites, so it has no expressions", speaker)
	}
	return parsePose(n, inside)
}

// parsePose reads "smile eyes=half arms=clasp" into n.Expr and n.Parts and
// checks every name against the character's layers.
func parsePose(n *Node, inside string) error {
	scratch := startPose(n.Name)
	for i, field := range strings.Fields(inside) {
		key, value, isPart := strings.Cut(field, "=")
		if !isPart {
			if i != 0 {
				return fmt.Errorf("%q: the expression comes first", field)
			}
			if !expressionKnown(n.Name, field) {
				return fmt.Errorf("%s has no expression %q", n.Name, field)
			}
			n.Expr = field
			continue
		}
		if err := applyPart(&scratch, n.Name, key, value); err != nil {
			return err
		}
		n.Parts = append(n.Parts, Part{key, value})
	}
	return nil
}

func parseCommand(n *Node, word, rest string) error {
	args := strings.Fields(rest)
	one := func() (string, error) {
		if len(args) != 1 {
			return "", fmt.Errorf("%s needs exactly one word", word)
		}
		return args[0], nil
	}
	var err error
	switch word {
	case "jump":
		n.Op = OpJump
		n.Text, err = one()
	case "set":
		n.Op = OpSet
		n.Name, err = one()
	case "if", "ifnot":
		n.Op, n.Not = OpIf, word == "ifnot"
		if len(args) != 3 || args[1] != "jump" {
			return fmt.Errorf("%s is `%s flag jump label`", word, word)
		}
		n.Name, n.Text = args[0], args[2]
	case "bg":
		n.Op = OpBG
		n.Name, err = one()
	case "prop":
		n.Op = OpProp
		if n.Name, err = one(); err == nil && !knownProps[n.Name] {
			err = fmt.Errorf("unknown prop %q", n.Name)
		}
	case "show":
		n.Op = OpShow
		if n.Name, err = one(); err == nil && !IsCharacter(n.Name) {
			err = fmt.Errorf("unknown character %q (want mayu or hana)", n.Name)
		}
	case "hide":
		n.Op = OpHide
		if len(args) != 0 {
			err = fmt.Errorf("hide takes no argument")
		}
	case "faces":
		n.Op = OpFaces
		if len(args) != 0 {
			err = fmt.Errorf("faces takes no argument")
		}
	case "pose":
		n.Op = OpPose
		if len(args) < 2 {
			return fmt.Errorf("pose is `pose name expression parts...`")
		}
		n.Name = args[0]
		if !IsCharacter(n.Name) {
			return fmt.Errorf("unknown character %q (want mayu or hana)", n.Name)
		}
		err = parsePose(n, strings.Join(args[1:], " "))
	case "fx":
		n.Op = OpFx
		if n.Name, err = one(); err == nil && !knownFX[n.Name] {
			err = fmt.Errorf("unknown fx %q", n.Name)
		}
	case "sfx":
		n.Op = OpSfx
		n.Name, err = one()
	case "music":
		n.Op = OpMusic
		n.Name, err = one()
	case "route":
		n.Op = OpRoute
		if len(args) != 0 {
			err = fmt.Errorf("route takes no argument")
		}
	case "auto":
		n.Op = OpAuto
		if len(args) != 0 {
			err = fmt.Errorf("auto takes no argument")
		}
	case "speed", "wait", "lock":
		n.Op = OpSpeed
		if word == "wait" {
			n.Op = OpWait
		}
		if word == "lock" {
			n.Op = OpLock
		}
		var v string
		if v, err = one(); err == nil {
			if n.Number, err = strconv.ParseFloat(v, 64); err != nil || n.Number <= 0 {
				err = fmt.Errorf("%s %q: want a positive number", word, v)
			}
		}
	case "end":
		n.Op = OpEnd
		n.Name, err = one()
	default:
		err = fmt.Errorf("unknown command %q", word)
	}
	return err
}

// BGNames lists the backgrounds the script uses, once each, in order.
func (s *Script) BGNames() []string {
	var names []string
	seen := map[string]bool{}
	for _, n := range s.Nodes {
		if n.Op == OpBG && !seen[n.Name] {
			seen[n.Name] = true
			names = append(names, n.Name)
		}
	}
	return names
}
