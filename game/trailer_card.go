package game

// The trailer's full-screen cards (scene "trailer.card", used by
// tools/trailer): one layer of a sheet drawn over the whole 480x270 picture,
// on black, with no CRT. Test mode only; a player never reaches it.

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"danmingame/sprites"
)

// SceneTrailerCard is the scene id; -card says which layer it shows.
const SceneTrailerCard = "trailer.card"

// DefaultCardSheet is the sheet a card spec without a sheet name uses.
const DefaultCardSheet = "trailer_cards"

// ParseCardSpec splits "sheet:layer", or "layer" alone (the default sheet).
func ParseCardSpec(spec string) (sheet, layer string, err error) {
	sheet, layer, found := strings.Cut(spec, ":")
	if !found {
		sheet, layer = DefaultCardSheet, spec
	}
	if sheet == "" || layer == "" {
		return "", "", fmt.Errorf("-card %q: want sheet:layer or a layer name", spec)
	}
	return sheet, layer, nil
}

type cardScene struct {
	spec  string
	sheet *sprites.Sheet
	layer string
}

func (s *cardScene) Enter(env *Env) error {
	name, layer, err := ParseCardSpec(s.spec)
	if err != nil {
		return err
	}
	if s.sheet, err = LoadSheet(env.Sprites, name); err != nil {
		return fmt.Errorf("card: %w", err)
	}
	if _, err := s.sheet.Frame(layer, 0); err != nil {
		return fmt.Errorf("card: %w", err)
	}
	s.layer = layer
	return nil
}

func (s *cardScene) Update() error { return nil }
func (s *cardScene) Exit()         {}

func (s *cardScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)
	_ = s.sheet.Draw(screen, []string{s.layer}, 0, 0, 0) // Enter checked the layer exists
}

// UsesCRT: the cards are flat.
func (s *cardScene) UsesCRT() bool { return false }
