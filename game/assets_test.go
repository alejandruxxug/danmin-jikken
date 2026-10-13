package game

import "testing"

// The room and the props boil slower than the people (boilScales).
func TestLoadSheetAppliesBoilScale(t *testing.T) {
	fsys, err := SpriteFS()
	if err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]float64{"reception_bg": 3, "desk": 3, "cecily": 0, "waiting_fisher": 0} {
		sheet, err := LoadSheet(fsys, name)
		if err != nil {
			t.Fatal(err)
		}
		if sheet.BoilScale != want {
			t.Errorf("%s: BoilScale = %v, want %v", name, sheet.BoilScale, want)
		}
	}
}
