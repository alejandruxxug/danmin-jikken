package night

import (
	"os"
	"slices"
	"testing"

	"danmingame/sprites"
)

// next finishes the line being typed and goes on to the following one.
func next(t *testing.T, e *Engine) {
	t.Helper()
	for range 2 { // the first click shows the whole line, the second goes on
		if err := e.Advance(); err != nil {
			t.Fatal(err)
		}
	}
}

// "layer" switches layers on and off, and a new "bg" switches them all off.
func TestLayerCommand(t *testing.T) {
	e := newEngine(t, "bg cg_coffee\nlayer +face_smile +fx_shake\n~ one\nlayer -face_smile +face_laugh\n~ two\nbg station\n~ three\n", Flags{})
	if !e.HasLayer("face_smile") || !e.HasLayer("fx_shake") || e.HasLayer("face_laugh") {
		t.Fatalf("after the first layer line: %v", e.Snapshot().CGLayers)
	}
	next(t, e)
	if e.HasLayer("face_smile") || !e.HasLayer("face_laugh") || !e.HasLayer("fx_shake") {
		t.Fatalf("after switching the face: %v", e.Snapshot().CGLayers)
	}
	next(t, e)
	if got := e.Snapshot().CGLayers; len(got) != 0 {
		t.Fatalf("a new bg must clear the layers, got %v", got)
	}
}

// Every layer the scripts switch on must exist in the CG's exported sheet,
// and so must the always-on layers and the figure layers in bgTable / draw.go.
func TestCGLayersExistInTheSheets(t *testing.T) {
	metas := map[string]*sprites.Meta{}
	meta := func(sheet string) *sprites.Meta {
		if m, ok := metas[sheet]; ok {
			return m
		}
		data, err := os.ReadFile("../assets/sprites/" + sheet + ".json")
		if err != nil {
			t.Fatal(err)
		}
		m, err := sprites.Parse(data)
		if err != nil {
			t.Fatal(err)
		}
		metas[sheet] = m
		return m
	}
	for name, def := range bgTable {
		if !def.cg {
			continue
		}
		for _, layer := range def.layers {
			if !slices.Contains(meta(def.sheet).Layers, layer) {
				t.Errorf("%s: always-on layer %q is not in %s.json", name, layer, def.sheet)
			}
		}
	}
	for _, layer := range []string{"figure", "figure_board", "figure_tilt"} {
		if !slices.Contains(meta("cg_kanashibari").Layers, layer) {
			t.Errorf("cg_kanashibari.json has no layer %q", layer)
		}
	}
	names, err := ScriptNames()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		s, err := LoadScript(name)
		if err != nil {
			t.Fatal(err)
		}
		bg := ""
		for _, n := range s.Nodes {
			switch n.Op {
			case OpBG:
				bg = n.Name
			case OpLayer:
				def, ok := bgTable[bg]
				if !ok || !def.cg {
					t.Errorf("%s line %d: layer used while bg is %q, which is not a CG", name, n.Line, bg)
					continue
				}
				for _, p := range n.Parts {
					if !slices.Contains(meta(def.sheet).Layers, p.Value) {
						t.Errorf("%s line %d: %s.json has no layer %q", name, n.Line, def.sheet, p.Value)
					}
				}
			}
		}
	}
}
