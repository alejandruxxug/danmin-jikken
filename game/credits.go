package game

import "danmingame/night"

// creditCard is one card of the credits: a heading and a line. The roll
// (finale.go) sets them in one column; cards with the same heading share it.
type creditCard struct {
	Heading, Line string
}

// dedications are the closing card of each ending's credits (Alejo, 2026-10-14):
// the heading is always "Made with love" and the line says for whom. Every one
// is a single line, so the roll is the same length for all four endings.
var dedications = map[string]string{
	night.EndingForever: "as a depiction of the most toxic love there is",
	night.EndingSleep:   "for the comfort that keeps you asleep",
	night.EndingWake:    "for everyone who kept their eyes open",
	night.Ending317:     "for whoever is still awake at 3:17",
}

// CreditCards are the credits of one ending, in order. The music cards name only files
// that are in assets/sound (musicCredits), so the roll never names a file
// that is not there. The last card, with no line, is the thanks.
func CreditCards(ending string) []creditCard {
	cards := []creditCard{
		{"DANMIN JIKKEN", ""},
		{"A game for the course", "Languages and Compilers, Universidad EIA"},
		{"Made by", "Alejo"},
		{"Written in", "Go and Ebitengine"},
		{"Art", "Drawn in Aseprite. Font: Silkscreen (SIL Open Font License)"},
		{"Made with", "Claude Code by Anthropic"},
		{"Made with", "the pixel-game-studio skill (ClaudeWD40)"},
	}
	for _, line := range musicCredits {
		cards = append(cards, creditCard{"Music and sound", line})
	}
	return append(cards,
		creditCard{"Halcyon Rest", "Drift off well."},
		creditCard{"Made with love", dedications[ending]},
		creditCard{"Thank you for staying awake", ""})
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
