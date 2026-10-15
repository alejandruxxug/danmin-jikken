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
