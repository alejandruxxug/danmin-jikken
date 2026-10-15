package game

import "fmt"

// The tutorial (Game Design -> design log: "Day 1 is a real tutorial"). A
// small box of two lines from Driftwell Orientation tells the player what to
// do next. The words and the highlighted spots are in the builder spec card
// of the Figma section "Day 1 tutorial (2026-10-14)".
//
// This file is the logic only. It works from tutFacts, a plain copy of what
// the desk looks like right now, so every rule can be tested without a
// window: Current says which hint shows, Observe notices the actions the
// hints ask for. tutorial_draw.go draws the box and the highlights.

// hintScene says where a hint shows. The box sits in a different place on the
// reception and on the look-down desk.
type hintScene int

const (
	hintAnywhere hintScene = iota // the hint shows wherever the player is
	hintReception
	hintDesk
)

// The ids of the day 1 steps that other code needs to know.
const (
	hintFiles       = "d1.files" // step 1: open her files
	hintOrientation = "d1.sheet" // step 2: the sheet; the next click turns to the referral
)

// tutFacts is what the hints look at. Shift.tutFacts fills it in.
type tutFacts struct {
	day    int
	served int // guests stamped so far

	window       string // who stands at the window, "" for nobody
	reason       string // that guest's Reason()
	wardC        bool   // the day's memo sends that guest to Ward C
	returning    bool   // that guest wears a wristband
	filesWaiting bool   // their folder lies on the desk
	folderOpen   bool   // ... and it is open, with the cover all the way back
	visited      bool   // the player has looked down at the desk

	holding bool  // a stamp is in the hand
	held    Stamp // which one

	dozing   bool // the day 1 guest asleep at the window: the "click her" line has the screen
	sleeper  bool // somebody sleeps and no staff has been sent yet
	attended bool // staff have been sent to somebody
	slip     bool // a citation slip lies on the desk
}

// Hint is one box of the tutorial.
type Hint struct {
	ID      string
	Step    int       // 1 to 10 for day 1's guided steps, 0 for the others
	Lines   [2]string // at most 174 px each
	Scene   hintScene
	Target  string // what is framed, in words (the dump shows it)
	Targets []rect // what is framed, in screen pixels
	Sleeper bool   // the first sleeper is framed too (the reception)

	after   string // the hint that must be done first ("" for none)
	oneShot bool   // gone for good once it has been shown and its moment has passed
	when    func(f tutFacts) bool
	done    func(f tutFacts) bool // the action the hint asks for has been done
}

// Where the things the hints frame are on the desk, in screen pixels (builder
// spec card, "TARGET RECTS").
var (
	rectIDName     = rect{33, 70, 120, 15}
	rectIDExpiry   = rect{82, 104, 72, 15}
	rectRefDate    = rect{193, 62, 58, 15}
	rectRefName    = rect{193, 77, 126, 14}
	rectRefDiag    = rect{193, 107, 126, 14}
	rectBarrow     = rect{197, 156, 27, 27}
	rectForDeskUse = rect{194, 192, 53, 18}
	rectMemo       = rect{357, 27, 110, 63}
	rectAdmit      = rect{357, 106, 30, 30}
	rectDischarge  = rect{397, 106, 30, 30}
	rectRefer      = rect{437, 106, 30, 30}
	rectSlip       = rect{370, 200, 94, 38}
)

// The moments hints wait for.
func atCecily(f tutFacts) bool { return f.day == 1 && f.window == "Old Cecily" && f.served == 0 }
func cecilyOpen(f tutFacts) bool {
	return atCecily(f) && f.folderOpen
}
func atWren(f tutFacts) bool { return f.day == 1 && f.window != "" && f.reason == reasonExpired }

// holdingStamp is true while the player holds that stamp.
func holdingStamp(f tutFacts, st Stamp) bool { return f.holding && f.held == st }

// hintList is every hint, in the order they are tried: the first one that
// is due shows.
var hintList = []Hint{
	// Day 1, Old Cecily: the guided first patient. Her queue place is 1, so
	// "served > 0" means she has been stamped and the steps are over.
	{ID: hintFiles, Step: 1, Scene: hintReception, Target: "FILES TAB",
		Lines:   [2]string{"GOOD MORNING, CLERK! YOUR FIRST", "PATIENT IS HERE. OPEN HER FILES."},
		Targets: []rect{filesRect},
		when:    func(f tutFacts) bool { return atCecily(f) && f.filesWaiting },
		done:    func(f tutFacts) bool { return f.visited || f.served > 0 }},
	{ID: hintOrientation, Step: 2, Scene: hintDesk, Target: "NONE (ORIENTATION SHEET OPEN)", after: hintFiles,
		Lines: [2]string{"THIS IS YOUR ORIENTATION SHEET!", "FOUR CHECKS. EASY AS SLEEPING."},
		when:  cecilyOpen, done: func(f tutFacts) bool { return f.served > 0 }},
	{ID: "d1.name", Step: 3, Scene: hintDesk, Target: "ID NAME + PATIENT", after: "d1.sheet",
		Lines:   [2]string{"COMPARE THE NAME ON THE ID", "WITH THE REFERRAL. SAME? LOVELY!"},
		Targets: []rect{rectIDName, rectRefName},
		when:    cecilyOpen, done: func(f tutFacts) bool { return f.served > 0 }},
	{ID: "d1.expiry", Step: 4, Scene: hintDesk, Target: "ID EXPIRY + DATE", after: "d1.name",
		Lines:   [2]string{"NOW THE DATE UNDER EXPIRES.", "STILL AFTER TODAY? WONDERFUL!"},
		Targets: []rect{rectIDExpiry, rectRefDate},
		when:    cecilyOpen, done: func(f tutFacts) bool { return f.served > 0 }},
	{ID: "d1.barrow", Step: 5, Scene: hintDesk, Target: "BARROW STAMP", after: "d1.expiry",
		Lines:   [2]string{"NOW FIND DR. BARROW'S ROUND", "STAMP. ONLY HIS STAMP COUNTS!"},
		Targets: []rect{rectBarrow},
		when:    cecilyOpen, done: func(f tutFacts) bool { return f.served > 0 }},
	{ID: "d1.diagnosis", Step: 6, Scene: hintDesk, Target: "DIAGNOSIS", after: "d1.barrow",
		Lines:   [2]string{"THE DIAGNOSIS MUST BE A SLEEP", "DISORDER. NARCOLEPSY? PERFECT!"},
		Targets: []rect{rectRefDiag},
		when:    cecilyOpen, done: func(f tutFacts) bool { return f.served > 0 }},
	{ID: "d1.lift", Step: 7, Scene: hintDesk, Target: "ADMIT STAMP", after: "d1.diagnosis",
		Lines:   [2]string{"ALL FOUR CHECKS PASS! LIFT THE", "ADMIT STAMP. WELL DONE, YOU!"},
		Targets: []rect{rectAdmit},
		when:    cecilyOpen, done: func(f tutFacts) bool { return holdingStamp(f, StampAdmit) || f.served > 0 }},
	{ID: "d1.press", Step: 8, Scene: hintDesk, Target: "FOR DESK USE", after: "d1.lift",
		Lines:   [2]string{"PRESS IT ON FOR DESK USE.", "THUNK! THAT'S THE SOUND OF CARE."},
		Targets: []rect{rectForDeskUse},
		when:    cecilyOpen, done: func(f tutFacts) bool { return f.served > 0 }},

	// Day 1, Matty: one line, then free play. Matty is second in the queue.
	{ID: "d1.matty", Scene: hintAnywhere, Target: "NONE", oneShot: true,
		Lines: [2]string{"YOU'RE A NATURAL! THE NEXT", "PATIENTS ARE ALL YOURS."},
		when:  func(f tutFacts) bool { return f.day == 1 && f.window == "Matty" },
		done:  func(f tutFacts) bool { return f.served >= 2 }},

	// Day 1, Wren: the expired ID teaches REFER. Wren is third.
	{ID: "d1.wren1", Step: 9, Scene: hintDesk, Target: "ID EXPIRY + DATE",
		Lines:   [2]string{"CHECK THE DATE ON HER ID...", "EXPIRED BEFORE TODAY? REFER!"},
		Targets: []rect{rectIDExpiry, rectRefDate},
		when:    func(f tutFacts) bool { return atWren(f) && f.folderOpen },
		done:    func(f tutFacts) bool { return f.served >= 3 }},
	{ID: "d1.wren2", Step: 10, Scene: hintDesk, Target: "REFER STAMP", after: "d1.wren1",
		Lines:   [2]string{"ANY CHECK FAILS = REFER. SHE CAN", "COME BACK WITH FRESH PAPERS!"},
		Targets: []rect{rectRefer},
		when:    func(f tutFacts) bool { return atWren(f) && f.folderOpen },
		done:    func(f tutFacts) bool { return holdingStamp(f, StampRefer) || f.served >= 3 }},

	// The first sleeper (day 1: send staff; day 3: leave them be).
	{ID: "d1.sleeper", Scene: hintReception, Target: "THE SLEEPER", Sleeper: true, oneShot: true,
		Lines: [2]string{"OH! SOMEONE NODDED OFF. CLICK A", "SLEEPING PATIENT TO SEND STAFF."},
		when:  func(f tutFacts) bool { return f.day == 1 && f.sleeper },
		done:  func(f tutFacts) bool { return f.day == 1 && f.attended }},
	{ID: "d3.sleeper", Scene: hintReception, Target: "THE SLEEPER", Sleeper: true, oneShot: true,
		Lines: [2]string{"NEW MEMO: DO NOT WAKE THEM.", "SLEEPERS SLEEP. LET THEM REST!"},
		when:  func(f tutFacts) bool { return f.day == 3 && f.sleeper }},

	// The first citation, whatever the day.
	{ID: "citation", Scene: hintDesk, Target: "CITATION SLIP", Targets: []rect{rectSlip}, oneShot: true,
		Lines: [2]string{"A CITATION! THE SLIP SAYS WHY.", "NO PENALTY. WE BELIEVE IN YOU!"},
		when:  func(f tutFacts) bool { return f.slip }},

	// Day 2: each new rule, the first time it shows up.
	{ID: "d2.name", Scene: hintDesk, Target: "ID NAME + PATIENT", Targets: []rect{rectIDName, rectRefName}, oneShot: true,
		Lines: [2]string{"READ NAMES LETTER BY LETTER!", "NEARLY THE SAME NAME IS A REFER."},
		when:  func(f tutFacts) bool { return f.day == 2 && f.reason == reasonName && f.folderOpen }},
	{ID: "d2.nostamp", Scene: hintDesk, Target: "BARROW STAMP SPOT", Targets: []rect{rectBarrow}, oneShot: true,
		Lines: [2]string{"NO STAMP, NO BED! A REFERRAL", "NEEDS DR. BARROW'S ROUND STAMP."},
		when:  func(f tutFacts) bool { return f.day == 2 && f.reason == reasonNoBarrow && f.folderOpen }},
	{ID: "d2.wardc", Scene: hintDesk, Target: "DESK MEMO", Targets: []rect{rectMemo}, oneShot: true,
		Lines: [2]string{"NEW MEMO! SHOAL PATIENTS GO TO", "WARD C. ADMIT THEM AS USUAL!"},
		when:  func(f tutFacts) bool { return f.day == 2 && f.wardC && f.folderOpen }},

	// Day 3: Cecily comes back with a wristband.
	{ID: "d3.wristband", Scene: hintDesk, Target: "DISCHARGE STAMP", Targets: []rect{rectDischarge}, oneShot: true,
		Lines: [2]string{"A WRISTBAND? WELCOME BACK!", "PAPERS OK + BAND? DISCHARGE!"},
		when:  func(f tutFacts) bool { return f.day == 3 && f.returning && f.reason == reasonWristband && f.folderOpen }},
}

// Tutorial is the state of the hints for one shift.
type Tutorial struct {
	On    bool            // hints are showing (Esc and the Settings turn them off)
	seen  map[string]bool // hints that are over
	shown map[string]bool // hints that were on the screen at least once
}

func newTutorial(on bool) *Tutorial {
	return &Tutorial{On: on, seen: make(map[string]bool), shown: make(map[string]bool)}
}

// ready says whether the hint may show: it is not over, the hint before it
// in its chain is, and it belongs to this scene.
func (t *Tutorial) ready(h *Hint, scene hintScene) bool {
	if t.seen[h.ID] || (h.after != "" && !t.seen[h.after]) {
		return false
	}
	return h.Scene == hintAnywhere || h.Scene == scene
}

// Current is the hint to show in scene right now, or nil. While the day 1
// guest is asleep at the window nothing shows: her own "click her" line has
// the screen.
func (t *Tutorial) Current(f tutFacts, scene hintScene) *Hint {
	if !t.On || f.dozing {
		return nil
	}
	for i := range hintList {
		h := &hintList[i]
		if t.ready(h, scene) && h.when(f) {
			return h
		}
	}
	return nil
}

// markShown remembers that a hint was on the screen.
func (t *Tutorial) markShown(h *Hint) {
	if h != nil {
		t.shown[h.ID] = true
	}
}

// Dismiss ends a hint (the player clicked it).
func (t *Tutorial) Dismiss(id string) { t.seen[id] = true }

// Observe runs every tick: a hint is over when the action it asks for has
// been done, or when a one-shot hint has been shown and its moment has
// passed. The list is walked in order, so one tick can finish a whole chain
// (the player stamps Cecily without reading the steps).
func (t *Tutorial) Observe(f tutFacts) {
	if !t.On {
		return
	}
	for i := range hintList {
		h := &hintList[i]
		if t.seen[h.ID] || (h.after != "" && !t.seen[h.after]) {
			continue
		}
		if (h.done != nil && h.done(f)) || (h.oneShot && t.shown[h.ID] && !h.when(f)) {
			t.seen[h.ID] = true
		}
	}
}

// Ticks is how many of the four checks of the orientation sheet are ticked:
// one for each of day 1's steps 3 to 6 that has been explained.
func (t *Tutorial) Ticks() int {
	n := 0
	for _, id := range []string{"d1.name", "d1.expiry", "d1.barrow", "d1.diagnosis"} {
		if t.seen[id] {
			n++
		}
	}
	return n
}

// Skip is Esc: no more hints this run.
func (t *Tutorial) Skip() { t.On = false }

// ---- The shift's side: facts, the current hint, what a click does. ----

// tutFacts copies what the hints look at out of the shift.
func (s *Shift) tutFacts() tutFacts {
	d := &s.desk
	f := tutFacts{
		day: s.Day, served: s.served, dozing: s.dozing != nil,
		filesWaiting: s.FilesWaiting(), visited: d.visited, slip: d.slip != nil,
		holding: d.holding, held: d.held,
	}
	if w := s.window(); w != nil && s.atWindow(w) {
		f.window, f.reason, f.wardC, f.returning = w.Name, w.Reason(), w.SendToWardC(s.Day), w.Returning
		f.folderOpen = d.guest == w && d.open && s.FolderOpen() == 1
	}
	f.sleeper = s.firstSleeper() != nil
	for _, sl := range s.sleepers {
		if sl.Attended {
			f.attended = true
		}
	}
	return f
}

// firstSleeper is the sleeper who has waited longest and nobody has attended,
// or nil. The hints frame them.
func (s *Shift) firstSleeper() *Guest {
	for _, g := range s.Waiting() {
		if !s.sleepers[g.ID()].Attended {
			return g
		}
	}
	return nil
}

// EnableTutorial turns the hints on or off when the shift is built. Without
// them the first folder of day 1 is like any other (the orientation sheet is
// its last page, not its first).
func (s *Shift) EnableTutorial(on bool) {
	s.tut.On = on
	for i, g := range s.Guests { // the day's rules still say which folder has it first
		g.OrientationFirst = on && s.Rules.Visitors[i].OrientationFirst
	}
}

// TutorialOn says whether hints are showing.
func (s *Shift) TutorialOn() bool { return s.tut.On }

// SkipTutorial is Esc: the hints stop for the rest of the run.
func (s *Shift) SkipTutorial() { s.tut.Skip() }

// CurrentHint is the hint to draw in scene, or nil.
func (s *Shift) CurrentHint(scene hintScene) *Hint { return s.tut.Current(s.tutFacts(), scene) }

// NoteHintShown is called every tick by the scene that draws the hint, so a
// one-shot hint knows it was seen.
func (s *Shift) NoteHintShown(scene hintScene) { s.tut.markShown(s.CurrentHint(scene)) }

// HintTargets is where the hint's highlights go on the screen. The sleeper
// is in the room, so its place moves with the camera.
func (s *Shift) HintTargets(h *Hint, cameraX int) []rect {
	out := append([]rect(nil), h.Targets...)
	if g := s.firstSleeper(); h.Sleeper && g != nil {
		r := guestRect(s, g)
		r.x -= cameraX
		out = append(out, r)
	}
	return out
}

// AdvanceHint is a click on the hint box (or Space): the hint is over. After
// the orientation sheet the folder turns to the referral, where the next
// steps point.
func (s *Shift) AdvanceHint(scene hintScene) {
	h := s.CurrentHint(scene)
	if h == nil {
		return
	}
	s.tut.Dismiss(h.ID)
	if h.ID == hintOrientation {
		_ = s.GoToPage(PageReferral)
	}
}

// TutorialTicks is how many checks of the orientation sheet are ticked.
func (s *Shift) TutorialTicks() int { return s.tut.Ticks() }

// tutorialPaused is true while Old Cecily's guided steps run, from the moment
// the player opened her files until she is stamped. The shift's clock stops
// then (the arrivals, the time limit, the HUD clock): reading the steps must
// not use up the shift. Nothing else needs stopping on day 1: nobody is in the
// ward yet, so nobody can fall asleep, and the next guests wait in the chairs.
func (s *Shift) tutorialPaused() bool {
	return s.tut.On && s.Day == 1 && s.served == 0 && s.tut.seen[hintFiles]
}

// Paused is true while the tutorial holds the shift's clock.
func (s *Shift) Paused() bool { return s.tutorialPaused() }

// jumpToTutorialStep is the backdoor's -tutorial-step N (1 to 10): it runs
// day 1 up to the moment of that step, so a tester can see it at once. It
// serves Old Cecily and Matty to get to Wren's steps.
func (s *Shift) jumpToTutorialStep(n int) error {
	if s.Day != 1 || n < 1 || n > 10 {
		return fmt.Errorf("tutorial step %d: want a step from 1 to 10 on day 1", n)
	}
	s.tut.On = true
	if n >= 9 { // serve Cecily and Matty first
		for _, g := range s.Guests[:2] {
			if err := s.runUntilFolder(g); err != nil {
				return err
			}
			_ = s.Wake(g) // an error just means they were awake
			if err := s.Stamp(StampAdmit); err != nil {
				return err
			}
		}
		if err := s.runUntilFolder(s.Guests[2]); err != nil {
			return err
		}
	} else if err := s.runUntilFolder(s.Guests[0]); err != nil {
		return err
	}
	if n >= 2 && n <= 8 {
		s.tutorial = true // she was woken already: no nod-off on top of the steps
		s.desk.visited = true
	}
	for _, h := range hintList {
		if h.Step > 0 && h.Step < n {
			s.tut.seen[h.ID] = true
		}
	}
	return nil
}

// runUntilFolder runs the shift until g stands at the window with their folder
// on the desk.
func (s *Shift) runUntilFolder(g *Guest) error {
	for range 60 * 90 {
		if s.window() == g && s.FolderReady() {
			return nil
		}
		s.Update()
	}
	return fmt.Errorf("tutorial: %s never reached the desk", g.Name)
}
