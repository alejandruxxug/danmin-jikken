# 🎬 Danmin Jikken – Trailer

Back to [[Danmin Jikken]] · Music in [[Danmin Jikken – Sound]] · Looks in [[Danmin Jikken – Art Direction]]

> **Status: in production (2026-10-14).** Alejo's picks: rendered by the game (record mode → MP4), ~1 min **with dark hints** (split-second flashes, no endings), English cards with Japanese accents. The shot list below is v1; the storyboard v2 is in Figma (section `Trailer`).
>
> **Was: proposal (2026-10-13).** A ~1 minute trailer to present the game to the class, **around Oct 23** (Alejo: "it's in 10 days"). That's after the Oct 19 submission and the Oct 20 defense, so the **finished game** is in the shots: night, title and all. Record on **Oct 21–22**, never during the build days.

---

## 🧭 The idea in one line

**Cheerful company video that goes wrong.** It starts like a Driftwell ad for the job and ends in the night. The trailer follows the same twist the game does, without explaining it.

**Don't spoil:** no gurney, nothing about what 3:17 really means, no endings. Hint at them, don't show them.

## 🎞️ Shot list (~60–75 s)

| # | Time | Shot | Scene / how | Sound |
|---|---|---|---|---|
| 1 | 0:00 | Black. Bell *ding*. **"Halcyon Rest Sanatorium is hiring."** | title card | bell |
| 2 | 0:05 | The Driftwell poster up close: the smiling moon, "DRIFT OFF WELL!" | `day.reception`, pan to the poster | Driftwell jingle starts |
| 3 | 0:10 | Slow pan across the reception. Old Cecily walks up to the glass | `day.reception -day 1 -seed 7` | jingle |
| 4 | 0:18 | Fast cuts: a stamp coming down, the folder of papers, the Settler ID, the computer flipping up | `day.reception`, `day.computer` | stamp thuds on the beat |
| 5 | 0:28 | Cecily **nods off mid-sentence**. The chairs fill with sleepers, then the hallway | `-sleepy` | jingle slows down, snores |
| 6 | 0:35 | Hold on the desk memo: **"DAY 3 – DO NOT WAKE THEM."** | `-day 3` | **silence** |
| 7 | 0:40 | The clock hits **3:17**. Strobe, the colours invert | `transition.night1` | strobe sound |
| 8 | 0:45 | Night: 3–4 quick VN frames. Mayu smiling too wide. One line: *"You look tired. You should sleep."* Hana half in shadow | `night.<beat>` | night music |
| 9 | 0:55 | Logo: **断眠実験 DANMIN JIKKEN**, the dagger drops through 断 | `title` | hit, then silence |
| 10 | 1:00 | *"Don't fall asleep."* · 2026 · made in Go | card | – |

## 🎥 How to record it

1. Each shot is a short **input script** for the agent backdoor (`-agent -scene … -seed … -input …`), so a take can be repeated exactly. See [[Danmin Jikken – Architecture#🤖 Agent backdoor (test mode)]].
2. Run it with `-fullscreen` and record the screen (**Cmd+Shift+5**). Keep the CRT filter **on** (no `-nocrt`): it's the look.
3. Cut the clips in any editor (CapCut, iMovie). The sound comes from the picks in [[Danmin Jikken – Sound]], with the same credit lines.
4. Export 1080p. 1920×1080 is exactly 480×270 ×4, so the pixels stay sharp.

**Bonus for the class:** the trailer shows the real game running on the real `hospital` model. The sleepers are goroutines.

## ❓ Open

| Question | Options |
|---|---|
| Length | 30 s (shots 2, 5, 6, 7, 9) / ~1 min (all) |
| Presentation date | Around Oct 23. Confirm the exact day |
| Text cards | English only, or English + Japanese (眠らないで, "don't sleep") |

## 🎞️ Rendering the trailer

Built 2026-10-14 (builder). The game renders the trailer **itself**, frame by frame, with the CRT and every shader exactly as a player sees them. No screen recording.

**One command** (from the repo root; needs ffmpeg and a screen, because the game opens a window while it records):

```bash
tools/trailer/render.sh tools/trailer/shots.txt trailer     # -> trailer_out/trailer.mp4
tools/trailer/render.sh tools/trailer/sample_shots.txt sample   # the 3-shot proof
```

| Part | What it does |
|---|---|
| `-agent -record <dir> -record-frames N` | The game runs fixed-step (one update per saved frame, 60 a second of output, no real time) and saves `<dir>/f000001.png …` at 1920x1080 (the 480x270 canvas x4, nearest) plus `<dir>/audio.json`. Works with `-input`, `-scene`, `-day`, `-seed`, `-flags`, `-line` and the rest. Writes no save files |
| `audio.json` | Every sound event with its frame: `sfx` (file, gain), `music` (file, start offset, loop), `seek`, `gain` (the music's loudness each frame, so fades are exact), `stop`. Frame 1 is at 0 s; an event of frame n sounds at (n-1)/60 s |
| `tools/trailer` (`trailer mix <dir>`) | Decodes the files from `game/assets/sound` to 48 kHz stereo, places every event, soft-limits, writes `<dir>/audio.wav` |
| `trailer render` | Records each shot, mixes it, joins the shots (frames as links, fades as darkened copies), applies the music beds, and runs ffmpeg: H.264, yuv420p, 60 fps, CRF 16, AAC 192k |

**The shot list** is one line per shot, clauses separated by `|`:

```
name | frames | backdoor args | clause | clause ...
desk | 180 | -scene day.desk.referral -day 1 -seed 7 | input stamp.txt | fadein 15
```

| Clause | Meaning |
|---|---|
| `input <file>` | The input script, path relative to the shot list |
| `card <layer>` | A full-screen card instead of the game: a layer of the sheet `trailer_cards`, or `sheet:layer`. Scene `trailer.card`, flat, no CRT |
| `music <file> <offset> [gain]` | A music bed that replaces the game's own music for the shot. Shots in a row with the same `music` clause share one bed that plays on across them. The game's sound effects stay |
| `fadein N` / `fadeout N` | Fade the picture and the sound from or to black over N frames |
| `cut` | Hard cut: the shot's sound is not tapered (10 ms) at its end |

Checks without listening (printed per shot): peak and RMS in dBFS, in the shot as the game sounded and as it sits in the trailer. The sample renders 9 s, 540 frames, 1920x1080, 60 fps, AAC 48 kHz stereo.

**Limits.**
- The window shows nothing while it records (drawing the CRT twice made the frames differ from run to run).
- The ward's timers (when a patient falls asleep) run in real time inside `hospital/`, which is frozen, so shots that depend on a sleeper nodding off can differ between runs. Everything seeded and everything on the VN side is identical every run (tested with `go test` in `tools/trailer`).
- A shot is as long as its `frames`, exactly; an sfx tail that runs past the end is cut.
- The rock audition's `-rock` files and the `audition/` folder are not mixed (only `game/assets/sound`).

