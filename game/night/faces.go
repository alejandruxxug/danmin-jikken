package night

import "fmt"

// The sprites are DDLC style (Art Direction -> M5 batch 1): one layer per
// part. A pose is one layer from each face group, an optional fx layer and
// one left and one right arm. Pose holds the part names without the group
// prefix: Eyes "open" means the layer "eyes_open".
type Pose struct {
	Eyes, Brows, Mouth string
	FX                 string // "" for none
	ArmL, ArmR         string
}

// Layers lists the sprite layers of the pose in the order the sheet stacks
// them (back to front).
func (p Pose) Layers() []string {
	layers := []string{"hair_back", "body", "head",
		"eyes_" + p.Eyes, "brows_" + p.Brows, "mouth_" + p.Mouth}
	if p.FX != "" {
		layers = append(layers, "fx_"+p.FX)
	}
	return append(layers, "arm_l_"+p.ArmL, "arm_r_"+p.ArmR)
}

// parts are the names each group may use, per character. A test checks them
// against the layers in the exported json, so a typo in a script is caught
// when the script is parsed and a missing layer is caught by `go test`.
type parts struct {
	eyes, brows, mouth, fx, armL, armR []string
}

var faceParts = parts{
	eyes:  []string{"open", "half", "wide", "closed", "happy", "away", "cry"},
	brows: []string{"neutral", "up", "worried", "angry", "flat"},
	mouth: []string{"neutral", "smile", "grin", "open", "frown", "pout", "flat", "wobble"},
	fx:    []string{"blush", "sweat", "tears", "shadow"},
}

var characterParts = map[string]parts{
	"mayu": {faceParts.eyes, faceParts.brows, faceParts.mouth, faceParts.fx,
		[]string{"down", "clasp", "chest", "mouth", "wave", "hug"},
		[]string{"down", "clasp", "chest", "mouth", "wave", "hug"}},
	"hana": {faceParts.eyes, faceParts.brows, faceParts.mouth, faceParts.fx,
		[]string{"down", "board", "cross", "chest", "mouth"},
		[]string{"down", "cross", "point", "elbow", "chest", "mouth"}},
}

// IsCharacter says whether name has sprites (a speaker like "Note" has not).
func IsCharacter(name string) bool {
	_, ok := characterParts[name]
	return ok
}

// expressions are the names the scripts use, written as in the docs:
// "Mayu (smile): ...". Each is a whole pose; a script line can override
// single parts after it (see parsePose).
var expressions = map[string]map[string]Pose{
	"mayu": {
		"smile":  {Eyes: "open", Brows: "neutral", Mouth: "smile", ArmL: "clasp", ArmR: "clasp"},
		"bright": {Eyes: "open", Brows: "up", Mouth: "grin", ArmL: "down", ArmR: "wave"},
		"laugh":  {Eyes: "happy", Brows: "up", Mouth: "grin", ArmL: "mouth", ArmR: "down"},
		"happy":  {Eyes: "happy", Brows: "up", Mouth: "smile", ArmL: "clasp", ArmR: "clasp"},
		"soft":   {Eyes: "half", Brows: "worried", Mouth: "smile", ArmL: "chest", ArmR: "down"},
		"blank":  {Eyes: "away", Brows: "flat", Mouth: "flat", FX: "shadow", ArmL: "hug", ArmR: "hug"},
		"pout":   {Eyes: "open", Brows: "worried", Mouth: "pout", ArmL: "clasp", ArmR: "clasp"},
		// Nights 2 and 3.
		"sad":     {Eyes: "half", Brows: "worried", Mouth: "frown", ArmL: "clasp", ArmR: "clasp"},
		"cold":    {Eyes: "open", Brows: "flat", Mouth: "flat", FX: "shadow", ArmL: "down", ArmR: "down"},
		"shy":     {Eyes: "away", Brows: "worried", Mouth: "smile", FX: "blush", ArmL: "chest", ArmR: "down"},
		"whisper": {Eyes: "wide", Brows: "flat", Mouth: "smile", FX: "shadow", ArmL: "mouth", ArmR: "down"},
		"crying":  {Eyes: "cry", Brows: "worried", Mouth: "wobble", FX: "tears", ArmL: "down", ArmR: "down"},
		"calm":    {Eyes: "open", Brows: "neutral", Mouth: "neutral", ArmL: "clasp", ArmR: "clasp"},
	},
	"hana": {
		"tired": {Eyes: "half", Brows: "flat", Mouth: "neutral", ArmL: "board", ArmR: "down"},
		"flat":  {Eyes: "open", Brows: "angry", Mouth: "flat", ArmL: "board", ArmR: "point"},
		"soft":  {Eyes: "half", Brows: "worried", Mouth: "neutral", FX: "blush", ArmL: "down", ArmR: "elbow"},
		// Nights 2 and 3.
		"blush":  {Eyes: "away", Brows: "worried", Mouth: "neutral", FX: "blush", ArmL: "down", ArmR: "chest"},
		"crying": {Eyes: "cry", Brows: "worried", Mouth: "wobble", FX: "tears", ArmL: "down", ArmR: "down"},
	},
}

// expressionKnown says whether the character has that expression.
func expressionKnown(character, name string) bool {
	_, ok := expressions[character][name]
	return ok
}

// startPose is how a character looks before any line gave her an expression.
func startPose(character string) Pose {
	if character == "hana" {
		return expressions["hana"]["tired"]
	}
	return expressions["mayu"]["smile"]
}

// applyPart sets one part of the pose from a "key=value" pair of a script
// line. The key is eyes, brows, mouth, fx, arml, armr or arms (both arms).
func applyPart(p *Pose, character, key, value string) error {
	known := characterParts[character]
	check := func(list []string) error {
		for _, name := range list {
			if name == value {
				return nil
			}
		}
		return fmt.Errorf("%s has no %s %q (want one of %v)", character, key, value, list)
	}
	switch key {
	case "eyes":
		p.Eyes = value
		return check(known.eyes)
	case "brows":
		p.Brows = value
		return check(known.brows)
	case "mouth":
		p.Mouth = value
		return check(known.mouth)
	case "fx":
		p.FX = value
		if value == "none" {
			p.FX = ""
			return nil
		}
		return check(known.fx)
	case "arml":
		p.ArmL = value
		return check(known.armL)
	case "armr":
		p.ArmR = value
		return check(known.armR)
	case "arms":
		p.ArmL, p.ArmR = value, value
		if err := check(known.armL); err != nil {
			return err
		}
		return check(known.armR)
	}
	return fmt.Errorf("unknown pose part %q (want eyes, brows, mouth, fx, arml, armr or arms)", key)
}
