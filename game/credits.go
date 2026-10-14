package game

// creditCard is one card of the credits: a heading and a line, shown in the
// bottom bar (Art Direction -> Cinematic pass). Ticks is how long it stays.
type creditCard struct {
	Heading, Line string
	Ticks         int
}

const (
	cardTicks      = 180 // a card stays 3 s
	musicCardTicks = 120 // the licence lines are many, so 2 s each
	cardGapTicks   = 15  // 0.25 s of empty bar between two cards
	thanksTicks    = 240 // the last card stays 4 s
)

// CreditCards are the credits, in order. The music cards name only files
// that are in assets/sound (musicCredits), so the roll never names a file
// that is not there.
func CreditCards() []creditCard {
	cards := []creditCard{
		{"DANMIN JIKKEN", "", cardTicks},
		{"A game for the course", "Languages and Compilers, Universidad EIA", cardTicks},
		{"Made by", "Alejo", cardTicks},
		{"Written in", "Go and Ebitengine", cardTicks},
		{"Art", "Drawn in Aseprite. Font: Silkscreen (SIL Open Font License)", cardTicks},
	}
	for _, line := range musicCredits {
		cards = append(cards, creditCard{"Music and sound", line, musicCardTicks})
	}
	return append(cards,
		creditCard{"Halcyon Rest", "Drift off well.", cardTicks},
		creditCard{"Thank you for staying awake", "", thanksTicks})
}

// CreditsTicks is how long the whole roll takes, gaps included.
func CreditsTicks(cards []creditCard) int {
	total := 0
	for _, c := range cards {
		total += c.Ticks + cardGapTicks
	}
	return total
}

// CardAt says which card is up ticks into the roll, and whether it is showing
// (false in the gap after it). Past the last card it returns len(cards).
func CardAt(cards []creditCard, ticks int) (index int, showing bool) {
	for i, c := range cards {
		if ticks < c.Ticks {
			return i, true
		}
		ticks -= c.Ticks
		if ticks < cardGapTicks {
			return i, false
		}
		ticks -= cardGapTicks
	}
	return len(cards), false
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
