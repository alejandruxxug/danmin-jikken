package main

// The shot list: one line per shot, clauses separated by "|".
//
//	name | frames | backdoor args | clause | clause ...
//
//	stamp | 180 | -scene day.reception -day 1 -seed 7 | input stamp.txt | fadein 20
//	title | 120 | | card trailer_cards:logo | music driftwell_jingle 4.5 0.8 | fadeout 30
//
// The clauses, in any order:
//
//	input <file>              the input script (relative to the shot list's folder)
//	card <layer>              a full-screen card, from trailer_cards or "sheet:layer", in place of the game
//	music <file> <offset> [gain]  a trailer-level music bed (see below)
//	fadein N, fadeout N       fade the picture and the sound from or to black, N frames
//	cut                       a hard cut: the shot's sound stops at its last frame with no tiny fade
//
// A music bed replaces the game's own music for the shot. Consecutive shots
// with the same music clause share one bed: it starts at the first shot's
// offset and plays on across them without a break. A # starts a comment.

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
)

// Bed is a trailer-level music bed.
type Bed struct {
	File   string
	Offset float64 // seconds into the file where it starts
	Gain   float64
}

// Shot is one line of the shot list.
type Shot struct {
	Name    string
	Frames  int
	Args    []string // the game's backdoor flags
	Input   string   // the input script, "" for none
	Card    string   // a card layer, "" for a game shot
	Bed     *Bed
	FadeIn  int
	FadeOut int
	Cut     bool
}

// GameArgs are the flags for the game binary (without -record and -record-frames).
func (s Shot) GameArgs() []string {
	args := []string{"-agent"}
	if s.Card != "" {
		args = append(args, "-scene", "trailer.card", "-card", s.Card)
	} else {
		args = append(args, s.Args...)
	}
	if s.Input != "" {
		args = append(args, "-input", s.Input)
	}
	return args
}

// ParseShots reads a shot list. dir is the folder input files are relative to.
// Errors name the line.
func ParseShots(r io.Reader, dir string) ([]Shot, error) {
	var shots []Shot
	seen := map[string]bool{}
	scanner := bufio.NewScanner(r)
	for lineNo := 1; scanner.Scan(); lineNo++ {
		line := scanner.Text()
		if i := strings.Index(line, "#"); i >= 0 {
			line = line[:i]
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		shot, err := parseShotLine(line, dir)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNo, err)
		}
		if seen[shot.Name] {
			return nil, fmt.Errorf("line %d: two shots are called %q", lineNo, shot.Name)
		}
		seen[shot.Name] = true
		shots = append(shots, shot)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return shots, nil
}

func parseShotLine(line, dir string) (Shot, error) {
	fields := strings.Split(line, "|")
	if len(fields) < 3 {
		return Shot{}, fmt.Errorf("want name | frames | args [| clauses], got %d fields", len(fields))
	}
	name := strings.TrimSpace(fields[0])
	if name == "" || strings.ContainsAny(name, " /\\") {
		return Shot{}, fmt.Errorf("shot name %q: want one word", name)
	}
	frames, err := strconv.Atoi(strings.TrimSpace(fields[1]))
	if err != nil || frames < 1 {
		return Shot{}, fmt.Errorf("frames %q: want a number of frames, 1 or more", strings.TrimSpace(fields[1]))
	}
	shot := Shot{Name: name, Frames: frames, Args: strings.Fields(fields[2])}
	for _, clause := range fields[3:] {
		words := strings.Fields(clause)
		if len(words) == 0 {
			continue
		}
		if err := shot.addClause(words, dir); err != nil {
			return Shot{}, err
		}
	}
	if shot.Card == "" && len(shot.Args) == 0 {
		return Shot{}, fmt.Errorf("shot %q has no backdoor args and no card", name)
	}
	if shot.FadeIn+shot.FadeOut > shot.Frames {
		return Shot{}, fmt.Errorf("shot %q: the fades (%d + %d) are longer than the shot (%d)", name, shot.FadeIn, shot.FadeOut, shot.Frames)
	}
	return shot, nil
}

func (s *Shot) addClause(words []string, dir string) error {
	key, args := words[0], words[1:]
	switch key {
	case "input":
		if len(args) != 1 {
			return fmt.Errorf("input needs one file")
		}
		s.Input = args[0]
		if !filepath.IsAbs(s.Input) {
			s.Input = filepath.Join(dir, s.Input)
		}
		// The game runs in its own folder, so the path must not depend on where we are.
		abs, err := filepath.Abs(s.Input)
		if err != nil {
			return err
		}
		s.Input = abs
	case "card":
		if len(args) != 1 {
			return fmt.Errorf("card needs one layer")
		}
		s.Card = args[0]
	case "music":
		if len(args) < 2 || len(args) > 3 {
			return fmt.Errorf("music needs a file, an offset in seconds and, optionally, a gain")
		}
		offset, err := strconv.ParseFloat(args[1], 64)
		if err != nil || offset < 0 {
			return fmt.Errorf("music offset %q: want seconds, 0 or more", args[1])
		}
		gain := 1.0
		if len(args) == 3 {
			if gain, err = strconv.ParseFloat(args[2], 64); err != nil || gain < 0 {
				return fmt.Errorf("music gain %q: want a number, 0 or more", args[2])
			}
		}
		s.Bed = &Bed{File: args[0], Offset: offset, Gain: gain}
	case "fadein", "fadeout":
		if len(args) != 1 {
			return fmt.Errorf("%s needs a number of frames", key)
		}
		n, err := strconv.Atoi(args[0])
		if err != nil || n < 0 {
			return fmt.Errorf("%s %q: want a number of frames", key, args[0])
		}
		if key == "fadein" {
			s.FadeIn = n
		} else {
			s.FadeOut = n
		}
	case "cut":
		if len(args) != 0 {
			return fmt.Errorf("cut takes no argument")
		}
		s.Cut = true
	default:
		return fmt.Errorf("unknown clause %q (want input, card, music, fadein, fadeout or cut)", key)
	}
	return nil
}

// sameBed says whether two shots are on the same music bed: both have one and
// it is the very same clause.
func sameBed(a, b *Bed) bool {
	return a != nil && b != nil && *a == *b
}

// BedStarts gives, for every shot with a bed, how many seconds into the bed
// file the shot begins: the bed's offset, plus the length of the earlier shots
// of the same run. A shot with no bed gets -1.
func BedStarts(shots []Shot) []float64 {
	starts := make([]float64, len(shots))
	for i, s := range shots {
		switch {
		case s.Bed == nil:
			starts[i] = -1
		case i > 0 && sameBed(shots[i-1].Bed, s.Bed):
			starts[i] = starts[i-1] + float64(shots[i-1].Frames)/FPS
		default:
			starts[i] = s.Bed.Offset
		}
	}
	return starts
}
