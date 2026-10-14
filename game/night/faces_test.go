package night

import (
	"os"
	"slices"
	"testing"

	"danmingame/sprites"
)

// Every layer a pose can name must exist in the exported sheet.
func TestPoseNamesExistInTheSheets(t *testing.T) {
	for character, p := range characterParts {
		data, err := os.ReadFile("../assets/sprites/" + character + ".json")
		if err != nil {
			t.Fatal(err)
		}
		meta, err := sprites.Parse(data)
		if err != nil {
			t.Fatal(err)
		}
		has := func(layer string) {
			if !slices.Contains(meta.Layers, layer) {
				t.Errorf("%s has no layer %q", character, layer)
			}
		}
		for _, n := range p.eyes {
			has("eyes_" + n)
		}
		for _, n := range p.brows {
			has("brows_" + n)
		}
		for _, n := range p.mouth {
			has("mouth_" + n)
		}
		for _, n := range p.fx {
			has("fx_" + n)
		}
		for _, n := range p.armL {
			has("arm_l_" + n)
		}
		for _, n := range p.armR {
			has("arm_r_" + n)
		}
		for name, pose := range expressions[character] {
			for _, layer := range pose.Layers() {
				if !slices.Contains(meta.Layers, layer) {
					t.Errorf("%s expression %q uses missing layer %q", character, name, layer)
				}
			}
		}
	}
}
