# CLAUDE.md

Guidance for Claude Code sessions in this repo. Alejo (the student) gives short instructions like "do M1" or "build the beds panel". This file and `docs/` hold everything you need to act on them without asking for context.

## What this is

**断眠実験 DANMIN JIKKEN**: a university assignment that grows into a game.

- **The assignment:** "Hospital de los Costeños con Narcolepsia", OOP in Go, Universidad EIA, Prof. Sebastián Zapata. **Due 2026-10-19 23:59** (repo link); **individual oral defense 2026-10-20**. Solo. Full brief: `docs/brief.pdf` (local only; the key rules are summarized below and in `docs/Danmin Jikken.md`).
- **The game:** a dark, anime-style pixel-art hospital. **Day shift = the assignment, playable** (FNAF 2-style reception desk; every action calls the real `hospital` package). **Night shift = a visual novel** (Slay the Princess look, worth 0 points). Scope: **the full game, everything on the scope board, by Oct 19** (Alejo's decision).

## The knowledge base: `docs/`

`docs/` is the design knowledge base. It's the same folder Alejo sees in Obsidian (`ACC(AfterClaudeCode)/Danmin Jikken` is a symlink to it), so **anything you write there shows up in his Obsidian immediately.** Notes use Obsidian `[[WikiLinks]]`: `[[Danmin Jikken – Story]]` means `docs/Danmin Jikken – Story.md` (the dash is an en dash `–`).

| Note | Read it when |
|---|---|
| `docs/Danmin Jikken.md` | Always first: the hub, the brief's rules, open questions |
| `docs/Danmin Jikken – Build Plan.md` | Picking what to do next: milestones M0–M8, lanes, the progress log |
| `docs/Danmin Jikken – Game Design.md` | Anything about gameplay: loop, timing, the day desk, the night, the design log |
| `docs/Danmin Jikken – Architecture.md` | Writing code: packages, OOP mapping, goroutines, the assignment checklist |
| `docs/Danmin Jikken – Art Direction.md` | Anything visual: resolution, palette, day/night looks, line boil, sprite layers |
| `docs/Danmin Jikken – Story.md` | Anything narrative: setting, cast, the illness, the twist, beat sheet |

Scope board (Alejo's In/Later/Cut picks): https://claude.ai/artifact/8K8TSeebvPn6K8z1d3cdU3

**Keep the docs current as you work:**
- Finished something → add a row to the **progress log** in the Build Plan and tick the milestone.
- Alejo decided something → add a row to the **design log** in Game Design (date, decision, why, in his words when possible).
- Notes are written for Alejo: plain English, tables, short sentences. Match the existing style.

## Commands

```bash
go run .                 # Section 6 demo scenario + the four queries
go test -race ./...      # all tests, with the race detector
go vet ./...             # must report nothing
gofmt -l .               # must list nothing
./assets/export.sh       # re-export every assets/src/*.aseprite to game/assets/sprites/ (sheet + json)
```

## Art: Aseprite only

All art is drawn in **Aseprite** and exported as **spritesheets**; the game only reads the exported sheets (Alejo's rule). Draw with the `aseprite` MCP tools (`mcp__aseprite__*`) into `assets/src/<name>.aseprite`, check your work with `export_frame`/`render_onion_skin` previews, then run `./assets/export.sh`. Never draw art in code, never hand-edit files in `game/assets/sprites/`, and never use the MCP's `export_spritesheet` for game sprites (it flattens the layers). Naming, sizes, the json format and the MCP gotchas: **Art Direction → Pipeline**.

**Visual approval gate (Alejo's rule):** sprites in Aseprite → a view of each screen in Figma (Figwright, `mcp__figwright__*`; `ping` first; never the official Figma MCP) → **Alejo approves** → only then code that screen. If a screen has no approved Figma view yet (check the Art log in Art Direction), don't build its visuals: make the view and ask for approval.

## Agent policy: Opus for visuals, Sonnet for building and testing (Alejo, 2026-10-13)

Three project subagents in `.claude/agents/`. Delegate to them (Agent tool, by name) instead of doing that kind of work in the main session:

| Agent | Model | Use it for |
|---|---|---|
| `art` | **Opus** | Everything visual: Aseprite sprites, Figma views, art reviews, Art log |
| `builder` | **Sonnet** | Go code: approved screens, scenes, the agent backdoor, the model, tests |
| `tester` | **Sonnet** | Checks + driving the game through the agent backdoor (Architecture → Agent backdoor); reports, never fixes |

Usual cycle for a screen: `art` (sprites → Figma view) → **Alejo approves** → `builder` → `tester` → fixes by `builder`. A whole lane session can also be started on a model directly: `claude --model opus` (art) or `claude --model sonnet` (build/test).

Every scene must be reachable with `-agent -scene <id>`; testing goes through the backdoor, not by playing from the title.

## Hard rules from the brief (the grade depends on them)

1. **`hospital/` is standard library only, Go 1.21** (`go.mod` says `go 1.21`). No printing, no drawing, no display code in it.
2. **After milestone M1, `hospital/` is frozen.** The brief requires that the model doesn't change when the GUI is added. The game goes in its own folder/module and uses `Hospital`'s methods only. If the game truly needs a model change, stop and ask Alejo.
3. **`main.go` only builds the Section 6 scenario and prints.** No business logic in it.
4. Unexported fields + methods. Typed constants for states. `Person` embedded by value. Staff in a `[]Attender`. Errors returned (never `panic` for foreseeable cases), wrapped with `%w`, compared with `errors.Is`.
5. **gofmt, `go vet` and `go test -race ./...` must be clean before every commit.**
6. **Everything in English**: code, comments, UI, names, dialogue. In-game text may also use Japanese. **No Spanish** anywhere in the game (Alejo's rule).
7. **Alejo must be able to defend every line** (his grade is multiplied ×0 to ×1.0 by the defense, GUI code included). Write plain, idiomatic Go, matching the code that's already there. No clever tricks. Explain non-obvious choices in a short comment.

## AI_USAGE.md is mandatory and must be honest

The brief requires `AI_USAGE.md` with real cases where the AI's code was wrong or not idiomatic, and the defense checks it against what Alejo says. **Whenever you (or Alejo) catch a mistake in AI-written code or advice, add a row to its table right away**: date, what the AI did, why it was wrong, the fix. Never invent entries.

## Git

- Commit after each finished feature or milestone, with checks green. Messages in English, one feature per commit. The brief says a single commit on the due date gets questioned, so commits should be spread over the days.
- **Never push** unless Alejo asks.
- Several Claude sessions work in parallel. Each owns one **lane** (see the Build Plan): Core = `hospital/`, `main.go`, `simulation/`; Day GUI = `game/`; Night VN = `game/night/` + story files; Art = `assets/`. Stay in your lane's folders. Pull/check `git status` before starting so you don't clobber another session's work.

## Layout

```
main.go          demo scenario + printing only
hospital/        the model (frozen after M1)
docs/            the knowledge base (= Obsidian ACC/Danmin Jikken)
assets/          art sources: src/*.aseprite, palette/danmin.hex, export.sh
game/assets/sprites/  exported sheets + json (generated, embedded by the game)
AI_USAGE.md      mandatory AI log
README.md        how to run + design decisions (for the professor)
```
