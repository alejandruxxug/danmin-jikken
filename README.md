# 断眠実験 DANMIN JIKKEN

**Hospital de los Costeños con Narcolepsia**: OOP in Go, Universidad EIA (Prof. Sebastián Zapata Ramírez).
Student: **Alejandro Urrego Giraldo**.

The core is the hospital management system the brief asks for. On top of it sits a game: a dark, anime-style pixel-art hospital where the **day shift** is the assignment, playable (you run the reception desk), and the **night shift** is a visual novel.

## Run

```bash
go run .          # the Section 6 demo scenario and the four queries
go test ./...     # tests (add -race to check the concurrency)
go vet ./...      # reports nothing
gofmt -l .        # lists nothing
```

Core: Go 1.21 or later, standard library only.

## Layout

```
main.go            demo scenario + printing only
hospital/          the model: no printing, no drawing
  person.go          Person, embedded by value in Patient, Doctor, Orderly and Nurse
  patient.go         Patient
  doctor.go          Doctor (an Attender)
  orderly.go         Orderly (an Attender that can't diagnose)
  nurse.go           Nurse (the third Attender, assigned to a ward)
  attender.go        the Attender interface
  room.go            Room
  episode.go         EpisodeRecord
  states.go          PatientState, NarcolepsyLevel, RoomState (typed constants)
  hospital.go        Hospital: dispatch, the lock, the four queries, Snapshot
  errors.go          sentinel errors
  hospital_test.go   tests (the 3 required ones come first)
simulation/        patient goroutines: random sleep, falls asleep, wakes up (bonus)
assets/            game art and palette
docs/              design docs (game design, story, art direction, architecture, build plan)
```

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

## License

- **Code:** MIT, see [LICENSE](LICENSE).
- **Art, story and characters:** CC BY-NC 4.0, see [LICENSE-ART.md](LICENSE-ART.md). Share and adapt with credit, not for commercial use.
