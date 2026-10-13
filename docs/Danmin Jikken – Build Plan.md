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
| **M3** | Thu Oct 15 | **The day loop** | d4, d5, d8, d9 | Click a sleeper → staff dispatched; beds panel + "no room" on screen; score + top-10; 3 days of rising difficulty |
| **M4** | Fri Oct 16 | **The computer + documents** | d6, d12, d10 | Flip-up terminal with the 4 queries + live ward feed; Papers, Please checks (ID, referral, stamps, daily rules) |
| **M5** | Fri–Sat Oct 16–17 | **Night engine + night 1** | n1, n2, n3, a4 | Day → night strobe (inverted) + day CRT; VN text box, choices, flags; the Prologue and night 1 playable with Mayu and Hana at 256 px |
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
