# 💤🗡️ 断眠実験 DANMIN JIKKEN

*"Sleep Deprivation Experiment."* 断 (*dan*) means "to sever", so the dagger in the logo cuts through 断 and through **DEPRIVATION** in the English subtitle. Title locked 2026-10-13.

> A dark, anime-style pixel visual novel set in a Fallout-style company sanatorium on a ruined coast, where the patients keep falling asleep. It's the Langs & Compilers workshop project (OOP in Go). Born in the daily note [[2026-10-13]].

![[siesta-test-patient-v0.png]]
*First sprite test, v0 (2026-10-13). Not anime. Superseded, but it proved the red rule.*

---

## 🗂️ Knowledge base

| Note | What's inside |
|---|---|
| [[Danmin Jikken – Game Design]] | Core loop, the 20-minute budget, the day desk, the night scenes, endings, the **design log** |
| [[Danmin Jikken – Story]] | Premise, cast, the twist options, a beat sheet per night, word budget |
| [[Danmin Jikken – Script]] | The night scripts (Prologue, Night 1–3, endings) + every decision, route and flag, all linked |
| [[Danmin Jikken – Art Direction]] | Anime pixel art, resolution, palette + the red rule, the sprite layer system, references |
| [[Danmin Jikken – Build Plan]] | 🗓️ Milestones M0–M8 to Oct 19, parallel lanes for several sessions, progress log |
| [[Danmin Jikken – Architecture]] | Go + Ebitengine, how every OOP concept maps to the game, goroutine patients, packages, the assignment checklist |

**Repo:** `~/Documents/Uni/4th Semester/Langs&Compilers/danmin-jikken`. **These notes live in the repo's `docs/` folder;** the Obsidian folder `ACC/Danmin Jikken` is a symlink to it, so both are the same files. The full brief is `docs/brief.pdf` (local only).

---

## ⚡ The game in 30 seconds

| Layer | What it is |
|---|---|
| ☀️ **Day shift** | You're the admissions clerk. Papers, Please at a hospital desk: check documents, follow the day's rules, stamp **ADMIT / DISCHARGE / REFER**. Patients fall asleep mid-interview. |
| 🌙 **Night shift** | Closing hours. The DDLC part: the story, the hospital after dark, choices. |
| 🎨 **Day vs night** | Day is **black outlines on white**, clean and flat like a company pamphlet. Night **comes alive** in Slay the Princess greyscale with **blood red**. Red means something is wrong (and it leaks into the day only when the game breaks). |
| ☢️ **Setting** | Fallout-style retro-futurist 1950s: a cheerful corporation runs a seaside sanatorium for the coastal settlements |
| ⏱️ **Length** | **20 minutes max**, target ~18. Three days, three nights. |
| 🧠 **The lore** | The diagnosis is narcolepsy. It goes much deeper than that. |
| 🧑‍🏫 **Why it wins the room** | The assignment's "sleepy" entities are goroutines, so the Go code is the horror. The professor sees his requirement working as a mechanic. |

## 🧭 Non-negotiables
1. **20 minutes max.** Cut story before cutting the time limit.
2. **Red never appears on the day shift**, until the one moment it's meant to break.
3. **Every assignment requirement is visible in the game**, not only in the code.
4. **The coastal patients are people, not a joke.** Inland folk say Shoal folk are "just lazy." The twist has to turn that prejudice around, not confirm it.
5. **Only English or Japanese** in the game: code, UI, names, terms, dialogue.

## 📋 The brief (`trabajo go.pdf`, Prof. Zapata)

| | |
|---|---|
| **Due** | **Oct 19 2026, 23:59** (repo link). Defense **Oct 20**, individual, ~15 min |
| **Solo** | 1 student |
| **Weight** | 15% of the course grade |
| **Rubric** | Model 20 · Interfaces + composition 15 · 4 queries 30 · Demo 15 · Idiomatic Go 20 · **+10% GUI** (Fyne / Wails / REST+HTML) · **+5% goroutines + mutex, race-clean** |
| **Defense multiplier** | ×1.0 to ×0 on your own grade. Explain any line, including the GUI |
| **Must ship** | README.md, AI_USAGE.md (honest, with 2+ cases of wrong AI code), git history with real commits |
| **GUI rule** | `hospital/` can't change when the GUI is added and can't contain display code |

**Day = the assignment, playable. Night = the visual novel, 0 points.** Alejo's split (2026-10-13).

🗳️ **Scope board:** https://claude.ai/artifact/8K8TSeebvPn6K8z1d3cdU3 (In / Later / Cut per feature, hours budget; picks saved for Claude to read)

## ❓ Blocking questions (for Alejo)
- [x] **Deadline:** Oct 19, 23:59 · defense Oct 20
- [x] **The brief:** read (see above)
- [x] **Solo or team:** solo
- [x] **GUI engine:** Ebitengine. Not on the brief's list, but Alejo says the professor doesn't mind (2026-10-13)
- [x] **Scope:** everything In, the full game (2026-10-13). See [[Danmin Jikken – Build Plan]]
- [x] **Sprite tool:** ✅ **Aseprite** + the `aseprite` MCP (2026-10-13). Pipeline in [[Danmin Jikken – Art Direction]]
- [x] **Title:** 断眠実験 Danmin Jikken (2026-10-13)
- [x] **Disease name:** ✅ **Yumetsuki (夢憑き)** (what the Shoals call it) + **S-317** (Driftwell's code). Alejo, 2026-10-13. See [[Danmin Jikken – Story]].
