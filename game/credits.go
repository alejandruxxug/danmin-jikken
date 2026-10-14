package game

// CreditLines are the credits, top to bottom. A line starting with "# " is a
// heading. The music credits (musicCredits) name only files that are in
// assets/sound, so the roll never names a file that is not there.
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

// musicCredits are the credit lines of the sound files the game uses (Sound
// note -> Licence summary). Both are CC0, which asks for no credit, but it is
// kind to give one. Add a line here when a file is added to assets/sound.
var musicCredits = []string{
	"Title: Vampire's Piano by TAD (CC0)",
	"Nights: Paranoid Truth by Tsorthan Grove (CC0)",
	"Both from OpenGameArt.org",
}
