# 🎬 Danmin Jikken – Script

Back to [[Danmin Jikken]] · Story and routes in [[Danmin Jikken – Story]]

> **Status: draft v1 (2026-10-13), written from story v3.** For Alejo to cut and rewrite. English only (Japanese allowed). Short lines: DDLC dread comes from normal talk with one wrong word in it.

---

## 📜 The scenes

| Scene | When | What happens | ~Words |
|---|---|---|---|
| [[Danmin Jikken – Prologue]] | Before day 1 | The gurney. *"Count back from ten."* Eyes close → white | 80 |
| [[Danmin Jikken – Night 1]] | After day 1 | "Go home." Mayu's first slip, Hana's note, one line too many on the monitor | 420 |
| [[Danmin Jikken – Night 2]] | After day 2 | The confession night. Who confesses depends on you | 760 |
| [[Danmin Jikken – Night 3]] | After day 3 | Hana didn't clock in. The gurney again. The endings | 950 |

## 🔀 Every decision in the game

| # | Where | Choice | Sets | Leads to |
|---|---|---|---|---|
| 1 | [[Danmin Jikken – Night 1#Choice 1]] | Stay and talk to Mayu | `liked_mayu` | [[Danmin Jikken – Night 1#1A Stay with Mayu]] |
| | | Go read Hana's note | `liked_hana` | [[Danmin Jikken – Night 1#1B Read Hana's note]] |
| 2 | [[Danmin Jikken – Night 2#Choice 2]] | Stay with Mayu | `stay_mayu` | [[Danmin Jikken – Night 2#Mayu route]] |
| | | Go find Hana | `find_hana` | [[Danmin Jikken – Night 2#Hana route]] |
| | | Call Mayu out | `confront_mayu` | [[Danmin Jikken – Night 2#Hana route]] if `liked_hana`, else [[Danmin Jikken – Night 2#Nobody route]] |
| 3 | Day 3 desk (your own file) | Stamp ADMIT **or** DISCHARGE | `stamped_self` | Only matters for [[Danmin Jikken – Night 3#Secret ending]] |
| 4 | [[Danmin Jikken – Night 3#Choice 4]] | Read Hana's letter **or** leave it | `read_letter` | [[Danmin Jikken – Night 3#4A Read the letter]] / [[Danmin Jikken – Night 3#4B Leave it]] |
| 5 | [[Danmin Jikken – Night 3#Choice 5]] | **Close your eyes** **or** keep them open | — | The ending (below) |

## 🧭 Routes

Decided at the end of [[Danmin Jikken – Night 2#Choice 2]]:

| Route | How | Confession | Hana's death (night 3) |
|---|---|---|---|
| 🌸 **Mayu** | `stay_mayu` | [[Danmin Jikken – Night 2#Mayu route]] | Heartbreak, real letter |
| 🌿 **Hana** | `find_hana`, or `confront_mayu` + `liked_hana` | [[Danmin Jikken – Night 2#Hana route]] | Mayu killed her, fake letter in Mayu's handwriting |
| 🕳️ **Nobody** | `confront_mayu` without `liked_hana` | [[Danmin Jikken – Night 2#Nobody route]] (none) | Heartbreak, real letter |

## 🏁 Endings

| Ending | Reached by | Scene |
|---|---|---|
| **Sleep** 🤍 | Close your eyes (Mayu or Nobody route) | [[Danmin Jikken – Night 3#Sleep ending]] |
| **Wake** 🌅 | Keep them open (Mayu or Nobody route) | [[Danmin Jikken – Night 3#Wake ending]] |
| **Forever** ⛓️ | Hana's route, always | [[Danmin Jikken – Night 3#Forever ending]] |
| **3:17** 🕒 *(secret)* | Keep them open + `stamped_self = DISCHARGE` + `read_letter` (Mayu or Nobody route) | [[Danmin Jikken – Night 3#Secret ending]] |

## 🚩 Flags

| Flag | Set in | Used in |
|---|---|---|
| `liked_mayu` | [[Danmin Jikken – Night 1#1A Stay with Mayu]] | Small lines in [[Danmin Jikken – Night 2]] |
| `liked_hana` | [[Danmin Jikken – Night 1#1B Read Hana's note]] | The route, if you `confront_mayu` ([[Danmin Jikken – Night 2#Choice 2]]) |
| `stay_mayu` / `find_hana` / `confront_mayu` | [[Danmin Jikken – Night 2#Choice 2]] | The route |
| `route` = `mayu` / `hana` / `nobody` | End of [[Danmin Jikken – Night 2#Choice 2]] | All of [[Danmin Jikken – Night 3]] |
| `stamped_self` | Day 3 desk | [[Danmin Jikken – Night 3#Secret ending]] |
| `read_letter` | [[Danmin Jikken – Night 3#Choice 4]] | [[Danmin Jikken – Night 3#Wake ending]], [[Danmin Jikken – Night 3#Secret ending]] |

## ✍️ How to read the scripts

| Looks like | Means |
|---|---|
| `[BG: nurses' station]` | Background |
| `[SFX: monitor beep]` `[MUSIC: off]` | Sound |
| `[FX: eyes close]` | Screen effect (eyelids, white cut, red pixel, text typing by itself) |
| `[FX: STROBE IN]` | The day → night strobe: flicker → build → white → fall into the night ([[Danmin Jikken – Art Direction#⚡ Day to night strobe]]) |
| **Mayu** *(smile)*: line | A character speaks. The word in brackets is the expression = the `eyes_` / `mouth_` layers in Aseprite |
| *Italic lines* | Your thoughts (you never speak aloud; DDLC style) |
| `→ set liked_mayu` | Sets a flag |
| `if route = hana:` | Only plays on that route |
| `{player}` | The player's name. Day 3 uses the OS username |

The game's story-file format is still undecided ([[Danmin Jikken – Architecture]]). These scripts are the source text it will be built from.
