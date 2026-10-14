package game

import (
	"fmt"
	"slices"
	"strings"

	"danmingame/night"
)

// The SCENES gallery's data: which moments the gallery shows, what each one
// replays, when it unlocks and which story flags a replay needs. No drawing
// here, so every rule can be tested with a table (gallery_cards_test.go).

// sceneCard is one card of the gallery: a moment of the story.
type sceneCard struct {
	ID     string // the backdoor's name for it (-scenes coffee, -scene replay.coffee)
	Thumb  string // the thumbnail layer in the scenes_gallery sheet (thumb_<Thumb>)
	Title  string // under the card
	Detail string // the line at the bottom while the card is selected

	// The stretch of the story it replays. The replay starts at the first
	// command on or after From (the engine plays the script silently up to it,
	// like the backdoor's -line N) and stops before the first command on or
	// after To. To 0 plays to the script's own end. The same range is what
	// unlocks the card: the player has reached it in normal play.
	Script   string
	From, To int

	// Routes lists the routes the replay can be played on (nil: any). It uses
	// the route of the run the player saw last if that one is allowed, else
	// the first one listed.
	Routes []string
	Set    []string // flags the replay always has on
	Clear  []string // flags the replay always has off
	Needs  string   // a flag the run needs for the card to unlock ("" for none)
}

// sceneCards is the grid in story order, row by row (4 x 3). The line
// numbers are the lines of the script files (a test plays every one of them).
var sceneCards = []sceneCard{
	{ID: "gurney", Thumb: "prologue", Title: "THE GURNEY", Detail: "PROLOGUE", Script: night.Prologue, From: 3},
	{ID: "coffee", Thumb: "coffee", Title: "COFFEE", Detail: "NIGHT 1", Script: night.Night1, From: 16, To: 47},
	{ID: "note", Thumb: "note", Title: "HANA'S NOTE", Detail: "NIGHT 1", Script: night.Night1, From: 62, To: 75,
		Set: []string{"liked_hana"}, Clear: []string{"liked_mayu"}}, // so choice 1 goes to the note
	{ID: "confession", Thumb: "confession", Title: "THE CONFESSION", Detail: "NIGHT 2 - MAYU ROUTE", Script: night.Night2, From: 57, To: 74,
		Routes: []string{night.RouteMayu}},
	{ID: "supply", Thumb: "supply", Title: "SUPPLY ROOM", Detail: "NIGHT 2 - HANA ROUTE", Script: night.Night2, From: 77, To: 107,
		Routes: []string{night.RouteHana}},
	{ID: "kanashibari", Thumb: "kanashibari", Title: "KANASHIBARI", Detail: "NIGHT 2", Script: night.Night2, From: 125, To: 154},
	{ID: "hana_room", Thumb: "hana_room", Title: "HANA'S ROOM", Detail: "NIGHT 3", Script: night.Night3, From: 22, To: 72},
	{ID: "forged", Thumb: "forged", Title: "IT WASN'T HANA", Detail: "NIGHT 3 - HANA ROUTE", Script: night.Night3, From: 75, To: 111,
		Routes: []string{night.RouteHana}, Set: []string{"read_letter"}, Needs: "route_hana"},
	{ID: "sleep", Thumb: "sleep", Title: "SLEEP", Detail: "ENDING", Script: night.EndingScript(night.EndingSleep),
		Routes: []string{night.RouteNobody, night.RouteMayu}}, // Hana's route always ends Forever
	{ID: "wake", Thumb: "wake", Title: "WAKE", Detail: "ENDING", Script: night.EndingScript(night.EndingWake),
		Routes: []string{night.RouteNobody, night.RouteMayu}},
	{ID: "forever", Thumb: "forever", Title: "FOREVER", Detail: "ENDING", Script: night.EndingScript(night.EndingForever),
		Routes: []string{night.RouteHana}},
	{ID: "317", Thumb: "s317", Title: "3:17", Detail: "SECRET ENDING", Script: night.EndingScript(night.Ending317)},
}

// secretCard is the index of the 3:17 card: it is not counted until found.
const secretCard = 11

// cardIndex finds a card by its id, or -1.
func cardIndex(id string) int {
	for i, c := range sceneCards {
		if c.ID == id {
			return i
		}
	}
	return -1
}

// reaches says whether a story at this line of this script is inside the
// card's range. line 0 means the script has ended or shows nothing.
func (c sceneCard) reaches(script string, line int, flags night.Flags) bool {
	if script != c.Script || line <= 0 || line < c.From {
		return false
	}
	if c.To > 0 && line >= c.To {
		return false
	}
	return c.Needs == "" || flags.Has(c.Needs)
}

// cardsReached lists the cards (by id) a story at this line of this script
// has reached. Several cards can share a script; a line is in at most one range.
func cardsReached(script string, line int, flags night.Flags) []string {
	var ids []string
	for _, c := range sceneCards {
		if c.reaches(script, line, flags) {
			ids = append(ids, c.ID)
		}
	}
	return ids
}

// The flags a replay may need from the run the player saw last: the choices
// of nights 1 and 2, the route they lead to, and the two of night 3.
var sceneFlagNames = []string{
	"liked_mayu", "liked_hana", "stay_mayu", "find_hana", "confront_mayu",
	"route_mayu", "route_hana", "route_nobody", "read_letter", "closed_eyes",
}

// routeFlags are the flags that decide the route; the replay builds them
// again when it has to play another route than the saved one.
var routeFlagNames = []string{"stay_mayu", "find_hana", "confront_mayu", "route_mayu", "route_hana", "route_nobody"}

// mergeSceneFlags keeps the flags of the run the player saw most recently. A
// run only counts once its route is decided (the route_ flags): before that
// the saved flags of the earlier run stay.
func mergeSceneFlags(saved []string, now night.Flags) []string {
	if !now.Has("route_mayu") && !now.Has("route_hana") && !now.Has("route_nobody") {
		return saved
	}
	var out []string
	for _, name := range sceneFlagNames {
		if now.Has(name) {
			out = append(out, name)
		}
	}
	return out
}

// choiceFlag is the choice of night 2 that leads to a route.
var choiceFlag = map[string]string{
	night.RouteMayu: "stay_mayu", night.RouteHana: "find_hana", night.RouteNobody: "confront_mayu",
}

// replayFlags are the story flags a card's replay runs with: a copy built
// from the saved ones, so nothing a replay does reaches the player's run.
func replayFlags(c sceneCard, saved []string) night.Flags {
	f := make(night.Flags)
	for _, name := range saved {
		f.Set(name)
	}
	route := savedRoute(f)
	want := route
	if len(c.Routes) > 0 && !slices.Contains(c.Routes, route) {
		want = c.Routes[0]
	}
	if want != route || !hasChoice(f) {
		// Play the card's own route: rebuild the choices that lead to it.
		for _, name := range routeFlagNames {
			delete(f, name)
		}
		f.Set(choiceFlag[want])
		if want == night.RouteNobody {
			delete(f, "liked_hana") // with it, calling Mayu out would be Hana's route
		}
	}
	for _, name := range []string{"route_mayu", "route_hana", "route_nobody"} {
		delete(f, name) // exactly one route_ flag, the `route` command sets the same one
	}
	f.Set("route_" + want)
	for _, name := range c.Set {
		f.Set(name)
	}
	for _, name := range c.Clear {
		delete(f, name)
	}
	return f
}

// savedRoute is the route a set of saved flags stands for: the route_ flag the
// run ended up with, else what its choices say ("nobody" when it has none).
func savedRoute(f night.Flags) string {
	switch {
	case f.Has("route_mayu"):
		return night.RouteMayu
	case f.Has("route_hana"):
		return night.RouteHana
	case f.Has("route_nobody"):
		return night.RouteNobody
	}
	return night.RouteFor(f)
}

// hasChoice says whether a night 2 choice flag is on.
func hasChoice(f night.Flags) bool {
	return f.Has("stay_mayu") || f.Has("find_hana") || f.Has("confront_mayu")
}

// newReplayEngine builds the engine of a card's replay: the script from the
// start, played silently up to the card's first line (FastForward answers the
// menus on the way with the flags above), and told to stop at its last.
func newReplayEngine(c sceneCard, saved []string) (*night.Engine, error) {
	script, err := night.LoadScript(c.Script)
	if err != nil {
		return nil, err
	}
	engine, err := night.NewEngineFor(script, replayFlags(c, saved), Username())
	if err != nil {
		return nil, err
	}
	if c.From > 0 {
		if err := engine.FastForward(c.From); err != nil {
			return nil, fmt.Errorf("replay %s: %w", c.ID, err)
		}
	}
	engine.StopAtLine(c.To) // 0 = no stop: the script's own end
	return engine, nil
}

// quietScript says whether a replay of this script starts its own music
// (night 3 and the endings are silent until the script asks for a track).
func quietScript(script string) bool {
	return script == night.Night3 || strings.HasPrefix(script, "ending_")
}

// ---- What the player has unlocked ----

// SceneUnlocked says whether a card is open.
func (e *Env) SceneUnlocked(id string) bool { return slices.Contains(e.Unlocked, id) }

// UnlockScene opens a card and reports whether it was closed before. The list
// stays in grid order, so the settings file reads the same every time.
func (e *Env) UnlockScene(id string) bool {
	if cardIndex(id) < 0 || e.SceneUnlocked(id) {
		return false
	}
	e.Unlocked = append(e.Unlocked, id)
	slices.SortFunc(e.Unlocked, func(a, b string) int { return cardIndex(a) - cardIndex(b) })
	return true
}

// NoteScene is called every tick by a night script that is really being
// played (never by a replay). Reaching a card's range unlocks it and saves
// the flags of this run; the settings file is only written when something new
// has happened.
func (e *Env) NoteScene(script string, line int) error {
	ids := cardsReached(script, line, e.Flags)
	if len(ids) == 0 {
		return nil
	}
	changed := false
	for _, id := range ids {
		if e.UnlockScene(id) {
			changed = true
		}
	}
	if merged := mergeSceneFlags(e.SavedFlags, e.Flags); !slices.Equal(merged, e.SavedFlags) {
		e.SavedFlags = merged
		changed = true
	}
	if changed {
		return e.SaveSettings()
	}
	return nil
}

// SceneCounts is the gallery's counter: how many cards are seen, and how many
// there are to see (the secret one only counts once it is found).
func (e *Env) SceneCounts() (seen, total int) {
	for _, c := range sceneCards {
		if e.SceneUnlocked(c.ID) {
			seen++
		}
	}
	total = len(sceneCards) - 1
	if e.SceneUnlocked(sceneCards[secretCard].ID) {
		total++
	}
	return seen, total
}

// UnlockFromOption is the backdoor's -scenes: "all", or card ids separated by commas.
func (e *Env) UnlockFromOption(list string) error {
	for _, id := range strings.Split(list, ",") {
		id = strings.TrimSpace(id)
		switch {
		case id == "":
		case id == "all":
			for _, c := range sceneCards {
				e.UnlockScene(c.ID)
			}
		case cardIndex(id) < 0:
			return fmt.Errorf("-scenes: no card %q (want all or: %s)", id, cardIDs())
		default:
			e.UnlockScene(id)
		}
	}
	return nil
}

// knownCards drops the ids that are not cards (a hand-edited settings file)
// and puts the rest in grid order.
func knownCards(ids []string) []string {
	var out []string
	for _, id := range ids {
		if cardIndex(id) >= 0 && !slices.Contains(out, id) {
			out = append(out, id)
		}
	}
	slices.SortFunc(out, func(a, b string) int { return cardIndex(a) - cardIndex(b) })
	return out
}

// cardIDs lists the card ids for an error message.
func cardIDs() string {
	ids := make([]string, len(sceneCards))
	for i, c := range sceneCards {
		ids[i] = c.ID
	}
	return strings.Join(ids, ", ")
}
