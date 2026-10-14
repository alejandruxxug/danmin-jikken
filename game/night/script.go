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
//	layer +face_look -face_down   CG only: switch layers of the CG background on (+) or
//	                              off (-); a new bg switches them all off again
//	fx eyes_close                 a screen effect (see knownFX)
//	sfx name / music name         sound cues (music off stops it; music fade 5 fades it
//	                              out over 5 s)
//	speed 12                      typewriter speed in letters per second
//	wait 1.5                      pause, in seconds (a click cuts it short)
//	hold                          a shot with no line: stays up until a click (AUTO
//	                              goes on after 2 s, the continue arrow shows)
//	lock 5                        pause that clicks cannot cut short
//	auto                          the next line goes on by itself when typed
//	autoplay on                   the whole script plays by itself on a fixed timeline: every line
//	                              goes on after AutoDelay, clicks and keys do nothing (the endings)
//	route                         set route_mayu / route_hana / route_nobody from the flags
//	writer mayu                   whose hand writes the Note:/Letter: lines on the paper
//	                              (hana is the default; a new bg puts it back)
//	bars on|off|snap|white|60 1.5 letterbox bars: slide in (0.6 s), out (0.4 s), snap in at once,
//	                              the dream's white bars, or grow to N pixels in S seconds
//	punch 3 350 89 / punch off    hard cut to a whole-number zoom (2 or 3) centred on x y
//	push 1.1 250 130 6            slow zoom to 1.1 centred on x y over 6 seconds (push off)
//	pan 140 4                     slide the camera across a wide picture to x in 4 seconds
//	paper 4 Words here            write words on the paper at once, on row 4 (no typing; a staging
//	                              command for the Wake ending's letter insert)
//	still procedure_blank         the picture an ending's title comes back to (see Engine.CutToStill)
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
	OpHold
	OpEnd
	OpLock
	OpAuto
	OpRoute
	OpLayer
	OpWriter
	OpBars
	OpPunch
	OpPush
	OpPan
	OpPaper
	OpStill
	OpAutoplay
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
	Parts   []Part // say/pose: parts changed after the expression; layer: {"+", name} or {"-", name}
	Text    string // say/think: the words; jump, if: the label
	Options []Option
	Not     bool      // if: jump when the flag is OFF
	Number  float64   // speed: letters per second; wait: seconds
	Args    []float64 // bars, punch, push, pan: the numbers (see cinema.go)
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
	"eyes_open":    true, // the eyelids open from closed (waking up)
	"eyes_slit":    true, // the eyelids close almost all the way (Forever)
	"splash":       true, // blood on the camera (the Forever cut); the stain stays after it
	"dim":          true, // the screen dims (the game falls asleep, 3:17)
	"figure":       true, // kanashibari: a faceless figure in the doorway
	"figure_gone":  true, // ... gone
	"dark":         true, // the picture goes dark at once (Hana's room before the red)
	"dusk":         true, // a lighter dark: the red room with only the envelope bright
	"light":        true, // ... and back to the light (ends dark and dusk)
	"wash":         true, // the Wake ending: a light-grey dither washes over the picture (2.5 s)
	"sting":        true, // the split-screen sting of the realization (0.8 s)
	"clear":        true, // takes every running effect away (a cut)
}

// knownProps are the things on the counter and the monitor.
var knownProps = map[string]bool{"coffee": true, "can": true, "syringe": true, "note": true, "monitor": true, "clear": true, "envelope": true, "clipboard": true}

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
	case "layer":
		n.Op = OpLayer
		err = parseLayers(n, args)
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
		if len(args) > 0 && args[0] == "fade" { // music fade 5: fade out over 5 seconds
			n.Name = "fade"
			if len(args) != 2 {
				err = fmt.Errorf("music fade: want the seconds, like music fade 5")
			} else if n.Number, err = strconv.ParseFloat(args[1], 64); err != nil || n.Number <= 0 {
				err = fmt.Errorf("music fade %q: want a positive number of seconds", args[1])
			}
		} else {
			n.Name, err = one()
		}
	case "autoplay":
		n.Op = OpAutoplay
		if len(args) != 1 || args[0] != "on" {
			err = fmt.Errorf("autoplay: want `autoplay on`")
		}
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
	case "writer":
		n.Op = OpWriter
		if n.Name, err = one(); err == nil && !IsCharacter(n.Name) {
			err = fmt.Errorf("unknown writer %q (want mayu or hana)", n.Name)
		}
	case "paper":
		n.Op = OpPaper
		row, words, _ := strings.Cut(rest, " ")
		r, convErr := strconv.Atoi(row)
		if convErr != nil || r < 0 || r > 12 || strings.TrimSpace(words) == "" {
			return fmt.Errorf("paper is `paper row words` (row 0 to 12)")
		}
		n.Number, n.Text = float64(r), strings.TrimSpace(words)
	case "still":
		n.Op = OpStill
		n.Name, err = one()
	case "bars":
		err = parseBars(n, args)
	case "punch":
		err = parsePunch(n, args)
	case "push":
		err = parsePush(n, args)
	case "pan":
		err = parsePan(n, args)
	case "hold":
		n.Op = OpHold
		if len(args) != 0 {
			err = fmt.Errorf("hold takes no argument")
		}
	case "end":
		n.Op = OpEnd
		n.Name, err = one()
	default:
		err = fmt.Errorf("unknown command %q", word)
	}
	return err
}

// parseLayers reads `layer +name -name ...`: each word is a sign and a
// layer name. The names are checked against the sheets by a test, as the
// parser does not know which background is up.
func parseLayers(n *Node, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("layer is `layer +name -name ...`")
	}
	for _, a := range args {
		if len(a) < 2 || (a[0] != '+' && a[0] != '-') {
			return fmt.Errorf("layer %q: want +name or -name", a)
		}
		n.Parts = append(n.Parts, Part{a[:1], a[1:]})
	}
	return nil
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
