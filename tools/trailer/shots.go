package main

// The shot list: one line per recording, clauses separated by "|".
//
//	name | frames | backdoor args | clause | clause ...
//
//	stamp | 146 | -scene day.desk.referral -day 2 -seed 7 | input stamp.txt | from 21
//
// A shot is one run of the game in record mode. The trailer's timeline
// (timeline.go) says which frames of which shot show where; the cards and the
// motion graphics are not shots, the compositor draws them from the sheets.
//
// The clauses, in any order:
//
//	input <file>   the input script (relative to the shot list's folder)
//	from N         the first frame to save (-record-from): the run before it
//	               is played but not drawn, which makes long set-ups quick
//
// A # starts a comment.

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
)

// Shot is one line of the shot list.
type Shot struct {
	Name   string
	Frames int      // the run's length: the game stops after this many frames
	Args   []string // the game's backdoor flags
	Input  string   // the input script, "" for none
	From   int      // the first saved frame (1 = save them all)
}

// GameArgs are the flags for the game binary, without -record and -record-frames.
func (s Shot) GameArgs() []string {
	args := append([]string{"-agent"}, s.Args...)
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
	shot := Shot{Name: name, Frames: frames, Args: strings.Fields(fields[2]), From: 1}
	for _, clause := range fields[3:] {
		words := strings.Fields(clause)
		if len(words) == 0 {
			continue
		}
		if err := shot.addClause(words, dir); err != nil {
			return Shot{}, err
		}
	}
	if len(shot.Args) == 0 {
		return Shot{}, fmt.Errorf("shot %q has no backdoor args", name)
	}
	if shot.From > shot.Frames {
		return Shot{}, fmt.Errorf("shot %q: from %d is after its last frame (%d)", name, shot.From, shot.Frames)
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
	case "from":
		if len(args) != 1 {
			return fmt.Errorf("from needs a frame number")
		}
		n, err := strconv.Atoi(args[0])
		if err != nil || n < 1 {
			return fmt.Errorf("from %q: want a frame number, 1 or more", args[0])
		}
		s.From = n
	default:
		return fmt.Errorf("unknown clause %q (want input or from)", key)
	}
	return nil
}
