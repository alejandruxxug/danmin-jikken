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
// note -> Licence summary). Ironbound (CC BY 3.0) and Hot Swing (CC BY 4.0)
// must be credited with these exact lines; the CC0 files ask for nothing but
// it is kind to give one. Add a line here when a file is added to assets/sound.
var musicCredits = []string{
	"Title: Vampire's Piano by TAD (CC0)",
	"Nights: Paranoid Truth by Tsorthan Grove (CC0)",
	"Sleep: Sad orchestral theme by Gichco (CC0)",
	"Wake: Realm of Torment by Vitalezzz (CC0)",
	"3:17: Unholy Surge by Vitalezzz (CC0)",
	"Forever: Ironbound by Zander Noriega",
	"CC BY 3.0, creativecommons.org/licenses/by/3.0",
	"Hot Swing by Kevin MacLeod (incompetech.com)",
	"Licensed under Creative Commons: By Attribution 4.0",
	"creativecommons.org/licenses/by/4.0",
	"Sound effects (CC0): rubberduck, Brian MacIntosh,",
	"yd, estevinho, Spring Spring, AntumDeluge,",
	"w.forster.1, Independent.nu",
	"Beeps, ringing and hum made for this game",
	"All from OpenGameArt.org and incompetech.com",
}
