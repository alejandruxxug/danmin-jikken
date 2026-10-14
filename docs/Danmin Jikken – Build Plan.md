# 🗓️ Danmin Jikken – Build Plan

Back to [[Danmin Jikken]] · Scope board: https://claude.ai/artifact/8K8TSeebvPn6K8z1d3cdU3

> **Scope: everything In, the full game, by Oct 19 23:59.** Alejo's call (2026-10-13): all 30 board items, built in parallel across several Claude sessions.
> **Build order = grade order.** Each milestone leaves a complete, submittable repo, so whatever happens, what's on `main` can be handed in.

---

## 🧱 Milestones

| # | Day | Milestone | Board items | Done when |
|---|---|---|---|---|
| **M0** ✅ | Tue Oct 13 | **Core model** | r1–r4, r5 started | `go run .` prints the Section 6 scenario; 8 tests pass under `-race`; gofmt + vet clean |
| **M1** ✅ | Wed Oct 14 | **Concurrency + Nurse, then freeze `hospital/`** | b1, d7 | Patient goroutines with random sleep, clean under `go run -race .`; Nurse as a 3rd Attender; **`hospital/` frozen** (the brief says it must not change once the GUI exists). **Frozen 2026-10-13** (Alejo confirmed). |
| **M2** ✅ | Wed–Thu Oct 14–15 | **GUI skeleton + reception view** | d1, d2, d3, a2, d11 | Ebitengine window 480×270 ×4, FNAF pan, patients arrive and fall asleep live, line boil on |
| **M3** ✅ | Thu Oct 15 | **The day loop** | d4, d5, d8, d9 | Click a sleeper → staff dispatched; beds panel + "no room" on screen; score + top-10; 3 days of rising difficulty |
| **M4** | Fri Oct 16 | **The computer + documents** | d6, d12, d10 | Flip-up terminal with the 4 queries + live ward feed; Papers, Please checks (ID, referral, stamps, daily rules) |
| **M5** | Fri–Sat Oct 16–17 | **Night engine + night 1** | n1, n2, n3, a4 | Day → night strobe (inverted) + day CRT **filmed by a camera** (rolling refresh band, moiré, exposure pumping; Art Direction → Day screen CRT); VN text box, choices, flags; the Prologue and night 1 playable with Mayu and Hana at 256 px |
| **M6** | Sat–Sun Oct 17–18 | **Nights 2–3, endings, meta-horror** | n4, n5, a5 | Kanashibari, Hana's room, the gurney, the 3 routes and **4 endings** (Sleep, Wake, Forever, 3:17) from [[Danmin Jikken – Script]]; OS username + `sleep_log_317.txt` |
| **M7** | Sun Oct 18 | **Title, art pass, sound** | a1, a3, d13 | 断眠実験 title with the dagger; all portraits; bell, snores, Driftwell jingle |
| **M8** | Mon Oct 19 | **Freeze + paperwork** | r5, x1 | Feature freeze **20:00**. README + AI_USAGE.md final, clean-clone check, submit before 23:59 |
| **🎤** | Tue Oct 20 | **Defense** | x1 | Drills done on Oct 18–19 |

**Daily rule:** at least one commit a day. The brief says a single commit on the due date gets questioned.

## 🛣️ Parallel lanes (one Claude session each)

Every lane owns its folders and doesn't touch the others', so sessions never overwrite each other.

| Lane | Owns | Starts | Depends on |
|---|---|---|---|
| **A · Core** (this session) | `hospital/`, `main.go`, `simulation/`, AI_USAGE.md | now | none |
| **B · Day GUI** | `game/` (Ebitengine), day scenes, computer, scores | after M1 freeze | `hospital/` frozen |
| **C · Night VN** | `game/night/`, story script files | M5, or earlier for the script | the Story doc |
| **D · Art** | `assets/` (`src/*.aseprite`, `export.sh`, palette) + the exported sheets in `game/assets/sprites/` | now | the Art Direction doc, the `aseprite` MCP |

**Models (Alejo, 2026-10-13):** the Art lane runs on **Opus**; Core, Day GUI and Night VN build on **Sonnet**, and testing is done by the Sonnet `tester` agent through the game's agent backdoor. Agents are defined in `.claude/agents/` (art, builder, tester); see CLAUDE.md → Agent policy.

**Rules for every lane:**
1. **Read the ACC docs first** (Game Design, Story, Art Direction, Architecture, this plan).
2. **Never edit `hospital/` after M1.** If the game needs something, it goes in `game/` or is raised as a question.
3. **No screen's visuals are coded before Alejo approves its Figma view** (sprites in Aseprite → view in Figma → ✅ → code). See [[Danmin Jikken – Art Direction]] → Order of work.
4. **Every AI mistake goes into `AI_USAGE.md`** (date, what, why it was wrong, the fix). The defense checks it against what Alejo says.
5. **Commit small, in English,** one feature per commit.
6. **Read `CLAUDE.local.md`** too, if it exists. It has local-only rules.

## ⚙️ Open technical decisions

| Decision | Plan | Status |
|---|---|---|
| GUI engine | **Ebitengine** | ✅ Not on the brief's list, but Alejo says the professor doesn't mind (2026-10-13) |
| Go version | Core stays `go 1.21` (std lib only). The game goes in its own module (`game/go.mod`) because Ebitengine needs a newer Go | ✅ Done 2026-10-13: `game/go.mod` is `go 1.27` |
| Separation | The game reads the model only through `Hospital`'s methods and never changes `hospital/` | Rule |

## ✅ Progress log

| Date | Done |
|---|---|
| 2026-10-13 | M0: `hospital/` (Person, Patient, Doctor, Orderly, Room, EpisodeRecord, Hospital, Attender, typed states, sentinel errors), `main.go` demo, 8 tests (3 required + MyEpisodes, SevereReport by day, polymorphic dispatch, wake frees room, 10 goroutines for 3 rooms), README, AI_USAGE.md started |
| 2026-10-13 | Art pipeline: Aseprite built + `aseprite` MCP connected and tested; `assets/export.sh` exports every `assets/src/*.aseprite` to a per-layer spritesheet + json in `game/assets/sprites/` |
| 2026-10-13 | M1: `simulation/` (one goroutine per patient, injected seed + sleep, `context` to stop, events channel), `Nurse` as the 3rd Attender, shared `episodeLog` with its own mutex, `Hospital.Snapshot()` for race-free reads, `main.go` runs the concurrent demo after Section 6. 18 tests (12 hospital, 6 simulation); `go run -race .` and `go test -race ./...` clean. **`hospital/` frozen** (Alejo confirmed, 2026-10-13): the game lane reads it only through `Hospital` methods and `Snapshot` |
| 2026-10-13 | M2 skeleton (no visuals yet): `game/` module (Ebitengine v2.10.5, `replace danmin => ../`), 480×270 window at ×4 with nearest-neighbour, Scene + SceneManager (`title`, `day.reception` as blank placeholders), Ward bridge (goroutines → events drained in `Update`, `Snapshot` for reads, stopped by context on scene exit), Aseprite sheet+json loader with `//go:embed`, the whole agent backdoor in `game/agent.go`. Tests for the loader, the input-script parser, the dump and the ward; `-agent -scene day.reception -seed 7 -frames 30 -dump … -quit` runs. Reception visuals wait for the approved Figma view; **M2 not ticked** |
| 2026-10-13 | M2 reception scene (`game/reception.go`) built to the approved v3 Figma view: layered sprite compositing (`Sheet.Draw`, layer opacity from the json, `boil` tag at 125 ms), 720-wide room with the 480 camera (default x 120) and the desk on the same origin, FNAF edge pan (mouse within 48 px of a side, clamped 0–240, no arrows), the day's queue walks up (window patient with the cecily sprite, two waiting_fisher in the chairs) and every patient is drawn awake or asleep from `Snapshot()` each frame, `screen_glass` on top without panning. Backdoor: `-pan X`, script `mouse X Y` and `pan X`, `cameraX` in the dump. **Skipped:** HUD text (no pixel font in the repo yet), CRT shader (M5), the patient-computer tab. d1, d2, d3 and a2 are met; d11 is only listed by id in the plan (its text is on the scope board, not in the docs), so **M2 is not ticked** until Alejo confirms what d11 needs |
| 2026-10-13 | **M2 ticked.** Its *Done when* is met: window 480×270 ×4, FNAF edge pan, patients arrive and fall asleep live, line boil on (d11 = squiggly outlines, done in Aseprite frames, not in code). **Moved to M3:** d2's "you admit them" (the ADMIT stamp calling `AdmitPatient`), the HUD text (needs a pixel font), and a ward fix: right now the ward admits all 5 patients at start, so they're asleep in beds before they reach the desk; they should only enter the hospital when stamped ADMIT. CRT shader stays in M5 |
| 2026-10-13 | **M3 the day loop, ticked.** Ward fixed: patients queue, walk to the window and enter the hospital (`AdmitPatient`) only on ADMIT; only admitted patients get a goroutine (`simulation.Config.ManualAttend`: the goroutine announces the sleep, the player's click calls `RegisterEpisode`). d2 stamps (ADMIT / REFER / DISCHARGE) decided from each visitor's data (`ReferralOK`, `Returning`); d4 click a sleeper, staff attend round robin, "NURSE MAYU ATTENDS ..."; d5 the patient computer (BEDS tab) + the no-room alert + hallway sleeper; d8 score (100 per right stamp, 50 per attended sleeper) + top-10 saved to `danmin_scores.txt` with OS usernames and the 03:17 ghost row; d9 three days from a table (5/3 rooms/3:00, 6/3/3:00, 4/2/2:30), after each shift the top-10 then the next day. HUD and all text in Silkscreen (OFL). Boil speed per sheet (characters 125 ms, room/props 375 ms), walk bob and blinking added. Backdoor: `day.computer`, `day.top10`, `-scores`, `-score`, `-sleepy`, script `stamp` and `shot4`, fuller dump. Screens compared with the three approved Figma previews. **Not built (M4):** the papers check, RECORDS / SLEEP LOGS / WARD FEED tabs; window patients all use Cecily's sprite |
| 2026-10-13 | **M3 fix + polish batch** (tester report). Window guests are clipped to the glass opening (x 261, y 61, 198 x 129); desk memo per day; every visitor has their own window sheet and awake mouth (`Visitor.Sprite` / `Mouth` in `days.go`); blinks every 4-8 s for 0.25 s; fullscreen (F11, Alt+Enter, `-fullscreen`, whole-number scaling with black bars); the scripted day 1 nod-off (Old Cecily dozes mid-joke, click wakes her, game side only, no hospital call); beds panel rows no longer overlap STAFF ON SHIFT (`+N MORE`), the toast moved to the top strip, chair slots spaced; shift clock stops on the top-10, computer tab resets on close, names cut by letters, backdoor `shot`/`dump` show the state before later commands. Decided, unchanged: a wrong ADMIT still admits, attending a hallway sleeper still scores 50. Reported as VISUAL, not fixed: ceiling-fixture strokes beside the HUD plates at pan 0 and 240 (the plates match the approved views, the strokes are in `reception_bg`) |
| 2026-10-13 | **M4 the computer + documents, built** (`game/` only; `hospital/` untouched). **Papers data** (`papers.go`, `days.go`): every visitor has a Settler ID (home, expiry), a referral (date, name as written, diagnosis, notes, Barrow stamp / none), a sleep log (day 2: all have 03:17) and, for returning patients, a wristband; `Visitor.Problem()` is derived from those papers, so `RightStamp` is the truth (tests check every visitor). **FILES tab + look-down desk**: tab at (64,253,88,17), pulses through the day palette in the day 1 tutorial until opened, a folder mark afterwards; the tilt is a 0.3 s `DrawTriangles` warp (`warp.go`) with a 1-2 px settle; the world keeps running. **Folder**: slides in, click opens it (cover swing bent in code), Settler ID with the window sprite cropped into the photo box, pages flip by **dragging the dog-eared corner** (curl follows the cursor; past the middle finishes, before springs back; pure functions `dragProgress` / `pageFlip` are tested), stamps are lifted and pressed on FOR DESK USE (held stamp hangs below-right of the cursor, so SIGNATURE stays clear), then the folder closes and slides back; a wrong stamp hands out the citation slip; reception stamp clicks now look down to the desk (`stamp` script command kept as a shortcut). **Computer**: RECORDS (type a name; file vs ID vs referral, check/cross, mismatch line), SLEEP LOGS (times earliest first, 03:17 not highlighted, click a name), WARD FEED (a line per admitted patient from `Snapshot`/sleepers, 60 s trace, click a sleeper to send staff). **Severe**: pips (terminal, paper, BEDS, folder tab, wristband), nurse call light blinks while a Severe patient sleeps unattended, +50 for attending within 10 s (`quickBonus`), darkening one palette step per episode today (`Hospital.SevereReport` + `sprites.Darker`). **Backdoor**: scene ids `day.desk`, `day.desk.closed|referral|sleeplog|wristband`, `day.computer.records|sleeplogs|wardfeed`; script `desk reception folder turn page lift apply putback drag down up type advance`; the dump has `desk`, `callLight` and per-guest pips/episodes/darkness. **Not done / why:** the day 2 rule "Shoal patients go to Ward C" and the day 3 rule "Do not wake them" are memo text only (nothing enforces them, they belong with the story flags in M5/M6), so **M4 is not ticked**; the CRT filter is M5; the Records tab shows only people who have reached the desk |
| 2026-10-13 | **M4: page turn is now a real page curl** (`game/curl.go`, `desk_draw.go`). The grabbed corner follows the cursor (x and y); the fold is the perpendicular bisector of the segment from the corner's rest place to the cursor, the part of the page past the fold is mirrored across it and drawn as the back (mint), the flat part shows the page, the next page shows through where the top page peeled away, with a thin deep-sage shadow along the fold and a drop shadow under the flap, all drawn as `DrawTriangles` fans. Release past the middle: the curl sweeps to the end, then the page fades down to the back of the stack; before the middle it curls back flat (about 0.3 s, eased). The geometry is small pure tested functions (`newFold`, `side`, `reflect`, `clip`, `curl`). Backdoor: `drag` still works; freeze mid-drag with `down X Y`, `mouse X Y`, `shot4` (no `up`) |
