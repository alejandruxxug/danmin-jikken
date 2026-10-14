package night

import (
	"strings"
	"testing"
)

func mustParse(t *testing.T, src string) *Script {
	t.Helper()
	s, err := Parse("test", strings.NewReader(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return s
}

func TestParseGood(t *testing.T) {
	s := mustParse(t, `
# a comment
bg station
show mayu
Mayu (smile arms=clasp): Hello! # trailing comment
~ A thought.
label a
choice
option First | set one,two | jump a
option Second
end day
`)
	if len(s.Nodes) != 6 {
		t.Fatalf("got %d nodes, want 6", len(s.Nodes))
	}
	say := s.Nodes[2]
	if say.Op != OpSay || say.Name != "mayu" || say.Expr != "smile" || say.Text != "Hello!" || say.Line != 5 {
		t.Errorf("say node = %+v", say)
	}
	if len(say.Parts) != 1 || say.Parts[0] != (Part{"arms", "clasp"}) {
		t.Errorf("parts = %v", say.Parts)
	}
	choice := s.Nodes[4]
	if len(choice.Options) != 2 || choice.Options[0].Jump != "a" || len(choice.Options[0].Set) != 2 {
		t.Errorf("choice = %+v", choice)
	}
	if s.Labels["a"] != 4 {
		t.Errorf("label a = %d, want 4", s.Labels["a"])
	}
}

func TestParseErrorsNameTheLine(t *testing.T) {
	tests := []struct{ name, src, want string }{
		{"unknown command", "bg station\nflip it", "line 2: unknown command"},
		{"unknown expression", "Mayu (grumpy): hi", "line 1: mayu has no expression"},
		{"bad part", "Hana (tired arms=clasp): hi", "hana has no arms"},
		{"unknown label", "jump nowhere", "line 1: jump to unknown label"},
		{"duplicate label", "label a\nlabel a", "line 2: label \"a\" is defined twice"},
		{"option alone", "option x", "line 1: option outside a choice"},
		{"one option", "choice\noption only\nbg black", "line 1: a choice needs at least two options"},
		{"unknown fx", "fx explode", "unknown fx"},
		{"bad wait", "wait soon", "want a positive number"},
		{"hold with a number", "hold 2", "hold takes no argument"},
		{"no sprites", "Note (angry): x", "has no sprites"},
		{"empty line", "Mayu:", "line 1"},
		{"layer without a sign", "layer face_look", "want +name or -name"},
		{"layer alone", "layer", "layer is `layer +name -name ...`"},
	}
	for _, tt := range tests {
		_, err := Parse("t.txt", strings.NewReader(tt.src))
		if err == nil || !strings.Contains(err.Error(), tt.want) || !strings.HasPrefix(err.Error(), "t.txt line") {
			t.Errorf("%s: error = %v, want it to contain %q", tt.name, err, tt.want)
		}
	}
}

func TestEmbeddedScriptsParse(t *testing.T) {
	names, err := ScriptNames()
	if err != nil || len(names) < 2 {
		t.Fatalf("ScriptNames = %v, %v", names, err)
	}
	for _, name := range names {
		if _, err := LoadScript(name); err != nil {
			t.Errorf("script %s: %v", name, err)
		}
	}
}
