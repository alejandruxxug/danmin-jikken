# AI usage

> Running log, written as the project happens so it matches the defense. Finished before submission.

## Tools
- **Claude Code (Claude Opus 5.5)**, several sessions/accounts: design docs, the `hospital` model, tests, the demo, the game.
- Student: Alejandro Urrego Giraldo

## Decisive prompts
*(to fill in: 2–3 prompts that shaped the project)*

## Cases where the AI was wrong or not idiomatic

| Date | What the AI did | Why it was wrong | Fix |
|---|---|---|---|
| 2026-10-13 | Recommended **Ebitengine** for the GUI before reading the brief | The brief only lists Fyne, Wails or REST + HTML/JS for the GUI bonus | Checked with the brief; kept Ebitengine because the professor accepts other GUI libraries |
| 2026-10-13 | Labelled the final room listing in `main.go` as "Query 5.3" | Query 5.3 is `AssignRoom` itself, not a room listing | Relabelled: the `AssignRoom` calls are 5.3, the listing is "Rooms at the end of the shift" |
| 2026-10-13 | In M0 the AI wrote `Doctor.MyEpisodes` and `Orderly.Episodes` to read their `episodes` slice with no lock, while `RegisterEpisode` appends to it under the hospital's mutex, and `doc.go` said goroutines were safe if they used `Hospital`'s methods | Fine with one goroutine, but once patient goroutines exist a reader calling `MyEpisodes` while an episode is recorded is a data race (the lock only protected the writer). Found while designing M1, before running the race detector | Moved the history into one small `episodeLog` with its own mutex (lock order: hospital, then log), shared by Doctor, Orderly and Nurse; added `Hospital.Snapshot()` for reading patients and rooms while goroutines run; rewrote the `doc.go` claim |
| 2026-10-13 | When asking for pale day greens, the main Claude session told the art agent the greens would turn **pale pink** in the strobe's colour inversion | Inverting a colour flips its brightness too (255 − value), so a pale green becomes a **dark** plum, not a pale pink. The art agent caught it by testing the inversion | The art direction notes record the real result: dark muted plum (`#221a26`, `#3c2e3f`, `#604e60`) under white lines |
| 2026-10-13 | Put the entry point in `cmd/danmin/` | The brief needs `go run .` from the repo root | Moved `main.go` to the root, set `go 1.21` in `go.mod` |
| 2026-10-13 | While adding the `-pan` flag the AI first gave it a default of 0 with -1 meaning "not given", then saw that any `Options{Enabled: true}` written in a test or in code would silently start the camera at 0 instead of the centre | A zero value must mean "use the default"; a magic -1 is easy to get wrong and 0 is a real camera position | Added `PanSet bool`, set with `flag.Visit` only when `-pan` was really passed, plus a test for both cases. Caught by the AI before the commit |
| 2026-10-13 | While designing the click that sends staff to a sleeper (M3), the AI's first idea was to call `Attender.Attend(p, location)` directly from the game's click handler on the `[]Attender` from `Hospital.Staff()` | `Attend` reads `p.State()` with no lock while the patient's goroutine can be changing it, so that is a data race, and it would also skip the hospital's own round-robin and episode history. Caught on paper before any code was written | The click calls `Hospital.RegisterEpisode` (through `Ward.Attend`), which holds the hospital's mutex; the simulation got a `ManualAttend` option so the goroutine only announces the sleep and does not attend by itself |
| 2026-10-13 | In `scorefile.go` the AI named a slice `real` | It shadows Go's built-in `real` function, which compiles but is confusing to read and to defend | Renamed to `played` before the commit |
| 2026-10-13 | In the M3 reception scene the AI drew the window guest with a plain `sheet` call and no clip, so a patient walking in from the right was drawn over the wall, the notice board and the post. The same pass put the toast over the stamp handles, let two hallway rows run into the "STAFF ON SHIFT" line, and kept the shift clock running on the top-10 | I only looked at the default camera and one guest at the window, never at the guest while walking or at other camera positions. Found by the tester agent and confirmed by the art agent | Clipped window-row guests to the glass opening (`sheetIn` with a `SubImage`), moved the toast to the top strip, 14 px row pitch with `+N MORE`, and stopped the clock on the top-10 scene |
| 2026-10-13 | The top-10 cut long names with `name[:maxNameChars]` | That slices bytes, so an OS username with an accent (for example `Zoë`) can be cut in the middle of a letter and print garbage | `shortName` converts to `[]rune` first and cuts letters; test with accents in `polish_test.go` |

## What I learned
*(to fill in before the defense)*
