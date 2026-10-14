package game

// CreditLines are the credits, top to bottom. A line starting with "# " is a
// heading. Music credits are added by the sound system (musicCredits) for the
// tracks the game really uses, so the roll never names a file that is not there.
func CreditLines() []string {
	lines := []string{
		"# DANMIN JIKKEN",
		"",
		"# A game for the course",
		"Languages and Compilers",
		"Universidad EIA",
		"",
		"# Made by",
		"Alejo",
		"",
		"# Written in",
		"Go and Ebitengine",
		"Art drawn in Aseprite",
		"Font: Silkscreen (SIL Open Font License)",
		"",
	}
	if len(musicCredits) > 0 {
		lines = append(lines, "# Music and sound")
		lines = append(lines, musicCredits...)
		lines = append(lines, "")
	}
	lines = append(lines,
		"# Halcyon Rest",
		"Drift off well.",
		"",
		"# Thank you for staying awake",
		"")
	return lines
}

// musicCredits are the credit lines of the sound files the game uses (the
// Sound note: only CC-BY tracks need one). Filled in by the sound system.
var musicCredits []string
