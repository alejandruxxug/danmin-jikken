# 断眠実験 DANMIN JIKKEN

**Hospital de los Costeños con Narcolepsia**: OOP in Go, Universidad EIA (Prof. Sebastián Zapata Ramírez).
Student: **Alejandro Urrego Giraldo**. Due 2026-10-19, defense 2026-10-20.

## What this is

Two things in one repository.

1. **The assignment.** A hospital management system for patients with narcolepsy, written as the brief asks: patients, doctors, orderlies and nurses, rooms, sleep episodes, the four required queries, the Section 6 demo scenario and the concurrency bonus (patients as goroutines). It lives in `hospital/` and `simulation/` and runs with `go run .`.
2. **A game on top of it** (the GUI bonus). *断眠実験 DANMIN JIKKEN* is a dark, anime-style pixel-art game. The **day shift** is the assignment, playable: you run the hospital's reception desk, and every click calls the real `hospital.Hospital` methods. The **night shift** is a visual novel and has nothing to do with the grade. The game is its own Go module in `game/`.

**Content warning:** the game contains depictions of suicide, self-harm, death and graphic gore, and strong flashing lights (a strobe). The title screen shows this warning, and Settings has a **Reduce flashing** option.

## How to run

### The assignment (root module)

```bash
go run .             # the Section 6 scenario and the four queries, printed
go test -race ./...  # all tests, with the race detector
go vet ./...         # reports nothing
gofmt -l .           # lists nothing
```

### The game

```bash
cd game
go run ./cmd/danmin            # play from the title
go test ./...                  # the game's tests
```

| Key | Does |
|---|---|
| **F11** (or Alt/Option+Enter, Cmd+F) | Fullscreen. The 480x270 picture is scaled by a whole number and centred, so pixels stay crisp |
| **F10** | CRT filter on/off for the day screens (also **Settings**) |
| **Esc** | Skips the Prologue (from the second launch on) and the tutorial hints |

**Settings** (title menu) has the CRT filter, reduce flashing, AUTO play, the tutorial hints and the music and sound volumes. They are saved in `danmin_settings.json` next to the executable. High scores go in `danmin_scores.txt` in the same place. With `go run`, "next to the executable" is a temporary folder, so build a binary (`go build -o danmin ./cmd/danmin`) if you want to keep scores between runs.

## Requirements

| Module | Go | Dependencies |
|---|---|---|
| Root (`danmin`: `hospital/`, `simulation/`, `main.go`) | **1.21** or later | standard library only |
| Game (`danmingame`, in `game/`) | **1.27** (see `game/go.mod`) | [Ebitengine](https://ebitengine.org) v2.10.5 and its dependencies, downloaded by `go` on the first build |
| Trailer tool (`danmintrailer`, in `tools/trailer/`) | 1.27 | optional, only for making the trailer video; needs `ffmpeg`. Not needed to grade or play |

Ebitengine needs:

- **macOS:** nothing extra (no cgo).
- **Windows:** nothing extra (no cgo).
- **Linux:** a C compiler and the system headers, for example on Debian/Ubuntu `sudo apt install gcc libc6-dev libgl1-mesa-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev libxxf86vm-dev libasound2-dev pkg-config`. It needs a display (X11 or Wayland) and a sound device to run; `go test` in `game/` works without opening a window.

The art and sound are embedded in the binary; there are no files to install. The root module never imports the game, so `go run .` and `go test ./...` at the root need nothing but Go.

## Project structure

```
main.go            Section 6 demo + printing only. No business logic
hospital/          THE MODEL. Standard library only, no printing, no drawing. FROZEN
simulation/        patient goroutines (bonus): uses only Hospital's methods
game/              the game (own module, requires danmin via "replace danmin => ../")
  cmd/danmin/        main: flags in, game.Run out
  night/             the visual novel engine and the night scripts
  sprites/, sound/   loaders for the Aseprite sheets and the audio
  assets/, fonts/    exported spritesheets (+ json), music, sound effects, the font
assets/            art sources (Aseprite files), palette and export script
tools/trailer/     optional: renders the trailer video from the game's record mode
docs/              design notes (game design, story, art, architecture, build plan)
```

The dependency only goes one way: `game/` -> `simulation/` -> `hospital/`. **`hospital/` is separate and frozen on purpose.** The brief says the model must not change when the GUI is added. `git log -- hospital/` shows the last change on 2026-10-13, before any game code existed. Everything the game needs (scores, the shift, the stamps, the beds panel) is built on top of `Hospital`'s public methods, in `game/`.

## Design decisions

- **Person is embedded by value**, so each patient or doctor owns their identity and the zero value is usable. With `*Person`, a `Patient{}` would hold a nil pointer and two structs could share one identity.
- **Staff lives in a `[]Attender`.** `RegisterEpisode` takes the next one in line (round robin) and calls `Attend` without knowing the concrete type. A new staff type only needs `ID`, `Name` and `Attend`, and `HireStaff` accepts it.
- **Each attender keeps its own episodes,** so `Doctor.MyEpisodes()` needs no hospital. Episode ids come from the attender (`D-01-E01`).
- **A missing room isn't an error for `RegisterEpisode`:** the episode is recorded without a room and the patient stays `AsleepInHallway`. `AssignRoom` is what returns `ErrNoRoomAvailable`.
- **Records are values.** An `EpisodeRecord` keeps the room and attender of that moment, even after the patient wakes up.
- **Only `Patient`'s own methods and `Room.Occupy`/`Release` change patient state.** `Room` puts patients in beds through unexported helpers, so outside code can't set `AsleepInBed` by hand.
- **One mutex in `Hospital`.** Every exported method locks it. Internal `...Locked` helpers assume it's held, because Go's mutex isn't reentrant. Entities (`Patient`, `Room`, `Doctor`) aren't safe alone; concurrent code goes through `Hospital`.
- **Patients are goroutines (bonus).** `simulation.Run` starts one goroutine per patient. Each waits a random time, calls `RegisterEpisode`, sleeps a random time and calls `WakePatient`. A goroutine never touches a `Patient` or `Room` directly, only `Hospital`'s methods, so the mutex decides who gets the last free room. `go run -race .` and `go test -race ./...` are clean.
- **Randomness and time are injected.** Each goroutine has its own `rand.Rand` (it isn't safe to share), seeded from `Config.Seed`, and `Config.Sleep` replaces the real timer, so tests run instantly and repeat exactly. A `context` stops the goroutines.
- **Reading while goroutines run goes through `Hospital.Snapshot()`,** which copies patients and rooms under the lock. Calling `p.State()` on a `*Patient` from another goroutine would race. The staff's episode history has its own small mutex (always taken after the hospital's), so `MyEpisodes` is safe too.
- **Nurse is the third `Attender`.** Adding it needed no change to `Hospital`: `HireStaff` takes any `Attender`. Doctor, Orderly and Nurse share one unexported `episodeLog` instead of repeating the same `Attend` code.
- **Slices for patients, rooms and staff** because order matters (admission order, "first free room"). **A map for `SevereReport`** because it answers "how many episodes for this patient" directly. `main` sorts it, since maps have no order.

- **Typed constants for states.** `PatientState`, `NarcolepsyLevel` and `RoomState` are their own types with `iota` and a `String` method, so a room state can't be passed where a patient state is expected. Every field of the model's types is unexported and reached by methods. The only exported-field structs are the read-only copies in `Snapshot` (`PatientView`, `RoomView`).
- **Errors are returned, never `panic`.** Sentinel errors live in `hospital/errors.go` (`ErrNoRoomAvailable`, ...). Methods wrap them with context using `%w` (`fmt.Errorf("patient %s: %w", p.ID(), ErrNotAdmitted)`) and callers test with `errors.Is`. There is no `panic` in `hospital/` or `simulation/`.
- **The GUI only calls `Hospital`'s methods.** The game's `Ward` (in `game/ward.go`) wraps one `Hospital`: ADMIT calls `AdmitPatient`, a click on a sleeper calls `RegisterEpisode`, the beds panel reads `Snapshot()` and the computer's queries call the four query methods. The game's own rules (score, stamps, days) are in `game/` and never reach into a `Patient` or `Room` field.

## Testing

- **Unit tests** in every package, run with `go test -race ./...` (root) and `go test ./...` (game). The model's tests start with the three the brief requires, then cover `MyEpisodes`, `SevereReport` by day, polymorphic dispatch over `[]Attender`, waking a patient frees the room and 10 goroutines competing for 3 rooms (clean under `-race`). The simulation takes its random source and its sleep function as parameters, so tests are instant and repeatable. The game's logic (scores, stamps, papers, finale timing) is tested without opening a window.
- **The agent backdoor.** `go run ./cmd/danmin -agent ...` is a test mode (all in `game/agent.go`, off unless `-agent` is passed). It jumps straight to any scene, sets the day, the clock and the random seed, plays a script of clicks and keys, saves a screenshot and dumps the game state as JSON. Example:

```bash
cd game
go run ./cmd/danmin -agent -scene day.reception -day 2 -clock 15:40 -seed 7 -frames 60 \
    -shot /tmp/out.png -dump /tmp/state.json -quit
```

  The flags and the list of scene ids are in `docs/Danmin Jikken – Architecture.md` (Agent backdoor). Each screen was compared with its approved design view this way.

## AI use

Claude Code (Anthropic) was used throughout. The honest log, including every case where the AI was wrong or not idiomatic, is in **[AI_USAGE.md](AI_USAGE.md)**.

## Game notes

- **Font:** the game's UI text uses **Silkscreen** (Jason Kottke), licensed under the SIL Open Font License 1.1. The font and its licence text are in `game/fonts/` (`Silkscreen-Regular.ttf`, `OFL.txt`).
- **Fullscreen:** F11, Alt(Option)+Enter or Cmd+F toggles it (macOS keeps F11), or start with `-fullscreen`. The 480x270 picture is scaled by a whole number (x8 on a 4K screen) and centred with black bars, so the pixels stay crisp at any screen size.
- **Scores:** `danmin_scores.txt`, next to the executable, one row per line: `<name> <score> <HH:MM> [message]`. The name is the OS username (`os/user`, "clerk" if unknown). A row with score `---` is the ghost row the game adds between runs.
- **Ward and goroutines in the game:** a patient only enters the hospital when the player stamps ADMIT, and only then gets a goroutine. The goroutine announces that the patient fell asleep; the player's click calls `RegisterEpisode`, so the next staff member in line attends and a room is assigned if one is free.
- **Music and sound credits:** the files, authors and licences are listed in `game/assets/sound/README.txt` and rolled in the game's credits. Two need an attribution line: **Ironbound** by Zander Noriega ([CC BY 3.0](https://creativecommons.org/licenses/by/3.0/), https://opengameart.org/content/ironbound, the Forever ending) and **Hot Swing** by Kevin MacLeod ([incompetech.com](https://incompetech.com), licensed under [Creative Commons: By Attribution 4.0](https://creativecommons.org/licenses/by/4.0/), the Driftwell jingle). The other music and the sound effects are CC0 from OpenGameArt.org (TAD, Tsorthan Grove, TinyWorlds, Gichco, Vitalezzz, rubberduck, Brian MacIntosh, yd, estevinho, Spring Spring, AntumDeluge, w.forster.1, Independent.nu); the beeps, the ringing and the hum are tones made for the game.
- **Made with:** Claude Code by Anthropic and the pixel-game-studio skill ([ClaudeWD40](https://github.com/alejandruxxug/ClaudeWD40)). Both are named in the game's credits, and each ending closes its credits with its own "Made with love" line.

## License

- **Code:** MIT, see [LICENSE](LICENSE).
- **Art, story and characters:** CC BY-NC 4.0, see [LICENSE-ART.md](LICENSE-ART.md). Share and adapt with credit, not for commercial use.
