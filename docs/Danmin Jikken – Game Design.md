# ☀️🌙 Danmin Jikken – Game Design

Back to [[Danmin Jikken]] · Story in [[Danmin Jikken – Story]] · Code in [[Danmin Jikken – Architecture]]

---

## 🔁 Core loop

```
COLD OPEN ──► DAY 1 ──► NIGHT 1 ──► DAY 2 ──► NIGHT 2 ──► DAY 3 ──► NIGHT 3 ──► ENDING
              desk      story       desk      story       desk      story
```

What you do at the desk changes the night. Who you admitted is in the ward, and who you discharged is gone. What you noticed in the papers unlocks lines and choices that night.

## ⏱️ The 20-minute budget

Hard cap of **20:00**, target **~18:00**. The ~2 minutes left over absorb slow readers and fumbled stamps.

| Segment | Time | What happens |
|---|---|---|
| Cold open | 0:45 | **Greyscale: you on a gurney, the nurses count you down, your eyes close.** Then white: a cheerful corporate orientation reel (Fallout-style), then the desk. |
| ☀️ Day 1 | 3:00 | 5 patients. Rules are easy. One patient falls asleep mid-interview (teaches the mechanic). |
| 🌙 Night 1 | 2:30 | Walking out at closing. First sign something's wrong. **First red pixel.** |
| ☀️ Day 2 | 3:00 | 5–6 patients. New rule. The papers start disagreeing with each other. |
| 🌙 Night 2 | 2:45 | You stay late on purpose. The ward. |
| ☀️ Day 3 | 2:30 | 4 patients. The desk itself starts to break (DDLC-style). |
| 🌙 Night 3 + ending | 3:30 | The truth, one final choice, the ending. |
| **Total** | **18:00** | |

**Word budget:** reading pace ≈ 200 words/min with pauses, so the nights get **~1,500 words of script in total** (~400 / 450 / 650). That's tight on purpose. DDLC-style horror works better short.

## ☀️ Day shift: the admissions desk

### 🖥️ The view: FNAF 2 office, but it's the reception desk

You're **seated at the front reception desk, facing the entrance**, first-person and fixed in place, like the FNAF 2 office. Instead of things coming down a hall at you, **patients walk up to your counter.** You never stand up during the day.

```
 ◄── pan left ──────────────── CENTER ──────────────── pan right ──►
 WAITING ROOM             THE COUNTER / WINDOW            WARD DOORS
 benches, patients        the patient you're serving,     staff-only doors,
 nodding off (live         papers slide in here,           Ward C sign?
 goroutines you can        stamps, the rules memo          (the hallway the
 see sleeping)                                             nurses come from)
                    ┌──────────────────────────┐
                    │  💻 flip-up COMPUTER      │  ← FNAF 2's camera tablet
                    └──────────────────────────┘
```

| FNAF 2 piece | In Danmin Jikken |
|---|---|
| Wide office you **pan left/right** with the mouse | The reception panorama (~720 px wide in a 480 px view) |
| The **camera tablet** you flip up | The **💻 patient computer**: a Driftwell terminal (Fallout style, white text on black) |
| Things approaching while you're on the tablet | **While you're on the computer, you're not watching the room.** Patients fall asleep, papers drop, someone new is standing at the counter, or something is standing at the ward doors |
| Mask / flashlight | Open slot. The **nudge** (wake a patient) or a desk bell could fill it |

### 💻 The patient computer (day)

Everything patient-related lives here. *(Exact screens wait on the hospital part Alejo is sending.)*
- **Patient records:** look up a name and compare it with their Settler ID and referral.
- **Sleep logs database:** every patient's episodes, where the **3:17** pattern shows up.
- **Bed / ward assignment:** where you send them after ADMIT (and on day 2, **Ward C** appears in the list but not on the map).
- **Live ward feed:** each admitted patient's line, i.e. the goroutines. At night this becomes the ward monitor.


**The Papers, Please layer.** One patient at a time walks up and hands over papers:

| Document | What to check |
|---|---|
| 🪪 **Settler ID** | Name, photo, home settlement, expiry |
| 📋 **Referral** | From which clinic, which doctor's stamp, which diagnosis |
| 📈 **Sleep log** | When the episodes happen. **This is where the lore leaks.** |
| 🏷️ **Wristband** | Only for returning patients. Which ward they were in. |

**The day's rules** go on a memo pinned to the desk, one new rule per day (like Papers, Please):
- Day 1: *"Admit only patients with a valid referral."*
- Day 2: *"Patients from [town] go to Ward C."* (There is no Ward C on the map.)
- Day 3: *"Do not wake them."*

### 📜 What each stamp means

> ✅ **Approved by Alejo (2026-10-13).** ⚠️ **The player must be told these rules before Day 1** (Alejo, 2026-10-13). Right now the orientation reel is one line, so nobody would know what a "valid referral" is. See *How the rules get explained* below.

**A referral is valid when all of these hold:**
1. The name on the referral matches the **Settler ID**.
2. The Settler ID is **not expired**.
3. It carries **Dr. Barrow's stamp** (Driftwell's head of the sleep unit; he writes every referral).
4. The diagnosis is a **sleep disorder** (narcolepsy, cataplexy, "sleep attacks").

| Stamp | When | What happens in the model |
|---|---|---|
| **ADMIT** | Valid referral **and** the sleep log shows episodes | `AdmitPatient`, then `AssignRoom` when they fall asleep; if every room is full → "no room", they wait in the hallway (`PatientsInHallway`) |
| **REFER** | Something is wrong with the papers (no referral, no stamp, names don't match, expired ID) | Sent back to their clinic for proper papers. Not admitted |
| **DISCHARGE** | A **returning** patient (wristband) whose sleep log has no new episodes | The model has **no discharge method** (and `hospital/` is frozen), so this is game-side only: they leave the queue. If they're asleep in a room, `WakePatient` frees it first |

Each day's memo adds one rule on top of these (day 2: *Shoal patients → Ward C*; day 3: *Do not wake them*).

**How the rules get explained (before Day 1):**
- **The orientation reel** (end of the [[Danmin Jikken – Prologue]]) grows from one line to ~4 short cards, Driftwell's cheerful voice: what the desk is, what makes a referral valid, the three stamps, "click a sleeping patient to call staff". ~30 s, skippable on a second run.
- **The rules memo** stays pinned on the desk, so the player can re-read it any time.
- **Day 1's first patient is the tutorial**: a clean, valid case to ADMIT, then Old Cecily falls asleep mid-joke (teaches the click).

**Stamps:** ADMIT · DISCHARGE · REFER. Wrong stamps get a citation slip (Papers, Please style), but that's flavour. **No game over at the desk**, because the run has to fit in 20 minutes.

**The sleep mechanic, i.e. the assignment working in front of the room:**
- Every patient is a goroutine with its own random sleep timer (see [[Danmin Jikken – Architecture]]).
- They can **fall asleep mid-interview**: eyes close, a 💤 bubble, the papers drop.
- You can **nudge** them awake (click), but each nudge costs time.
- Their sleep times get written into their sleep log. That's how players notice things: **on day 2 every log shows an episode at 3:17 AM.**

### 🏆 Score + top-10 (M3)

- **Points = correct choices.** Every right call at the desk scores: right stamp, right room, attending the sleeper. Wrong calls just don't score (no game over).
- **The top-10 uses OS usernames** (`os/user` in Go), not typed names.
- **The save file talks back (DDLC-style).** Scores are saved to a real file next to the game, `danmin_scores.txt`. Between runs, a row you never played appears: your username, no score, the time **03:17**, and a line like *"why did you let me sleep"*. Same idea as DDLC's character files; it pairs with `sleep_log_317.txt`.

```
danmin_scores.txt
-----------------
alejo   880  14:02
alejo   912  16:40
alejo   ---  03:17  why did you let me sleep
```

All of this is game code (`game/`). `hospital/` stays frozen.

## 🌙 Night shift: the story

DDLC-style VN: a background, character sprites, a text box, and a choice now and then.
- **Explorable?** No, it's linear scenes with choices. Exploration doesn't fit in 20 minutes.
- **Choices:** 2–3 per night, plus **flags carried from the desk** (who you admitted, which discrepancies you caught).
- **The ward monitor:** a heart-monitor screen at the nurses' station that shows every patient's line live. **That's the goroutines drawn on screen.** For the sustentación it doubles as the debug view. 🧑‍🏫

## 🎭 Meta-horror ideas (DDLC-style, pick 2–3 max)
- The game "falls asleep": the screen dims, input is ignored for a few seconds, the text box types *"…"*.
- The day 3 desk shows a patient whose Settler ID has **the player's OS username** (`os/user` in Go).
- After the ending, a file appears in the game folder: `sleep_log_317.txt`.
- The ward monitor shows one more line than there are patients. Yours.

## 🏁 Endings (2–3)

| Ending | Reached by | Feel |
|---|---|---|
| **Wake** | You caught the key discrepancies and refused the day 3 rule | Bitter, you got out |
| **Sleep** | You followed every rule | You become a patient |
| **3:17** *(secret?)* | Specific flags | The meta ending |

**Superseded by story v3:** four endings, Sleep / Wake / Forever / 3:17 secret, decided by the final gurney choice (close your eyes or keep them open) + flags. Full table in [[Danmin Jikken – Story]] → Endings.

## 📜 Design log

| Date | Decision | Why |
|---|---|---|
| 2026-10-13 | Hospital VN: day = Papers, Please desk, night = DDLC story | Alejo's concept ([[2026-10-13]]) |
| 2026-10-13 | Stay in **Go** (no language haggle) | The assignment is OOP in Go, and goroutine patients make the code the horror. Ren'Py would make the VN easy but bolt the OOP on |
| 2026-10-13 | **20 min max**, target 18 | Alejo: "maybe max 20 minutes" (first said <25) |
| 2026-10-13 | No game over at the desk | A failed run can't eat the time budget |
| 2026-10-13 | Anime style (DDLC), pixel art | Alejo. See [[Danmin Jikken – Art Direction]] |
| 2026-10-13 | **Day = black outlines + white fill only. Night = Slay the Princess look: greyscale shading + blood red** | Alejo, with Slay the Princess references. The night "comes alive" |
| 2026-10-13 | **Setting = Fallout-style** retro-futurist 1950s Americana: a company sanatorium on a ruined coast | Alejo |
| 2026-10-13 | **Only English or Japanese** in the game: names, terms, dialogue. No Spanish | Alejo. The old Spanish cast and terms were replaced |
| 2026-10-13 | The nurses are going mad because they've caught the illness and refuse to sleep | Alejo's idea; the sleep-deprivation explanation was Claude's |
| 2026-10-13 | VN character sprites **max 256 px** tall (~160×256), same pixel scale as the world | Alejo. See [[Danmin Jikken – Art Direction]] |
| 2026-10-13 | Claude makes the sprites; Alejo art-directs | Alejo's call. Pipeline in [[Danmin Jikken – Art Direction]] |
| 2026-10-13 | **Title: 断眠実験 DANMIN JIKKEN** ("Sleep Deprivation Experiment"). Logo: a dagger through 断 and through DEPRIVATION | Alejo picked it from Claude's 断眠 suggestion. "Experiment" hints at Driftwell's Ward C before the player knows. That's dread, not a spoiler |
| 2026-10-13 | **Day view = FNAF 2 office at the front reception desk**: fixed first-person, pan left/right, patients walk up to the counter, a flip-up **patient computer** in place of the camera tablet | Alejo. Time on the computer = not watching the room |
| 2026-10-13 | **Scope: the full game, all 30 board items In, by Oct 19** | Alejo: "i dont care i have the time", building with several Claude sessions. Claude flagged the 74 h vs 18 h budget once; his call |
| 2026-10-13 | **Day = the assignment playable; night = VN (0 points)** | Alejo |
| 2026-10-13 | **GUI engine: Ebitengine** | Not on the brief's list (Fyne/Wails/REST). Alejo: the professor "doesnt care" |
| 2026-10-13 | **All art is made in Aseprite and exported as spritesheets** to the game folder; the game only reads the sheets | Alejo: "preferably everything is designed in asesprite and exported to the game folder then the game uses it. Prefereably spritesheets". Pipeline in [[Danmin Jikken – Art Direction]] |
| 2026-10-13 | **Approve the look before building: sprites in Aseprite → screen views in Figma → Alejo approves → then into the game** | Alejo: "so i can approve the visual side BEFORE building anything" |
| 2026-10-13 | **Agent policy: Opus for visuals, Sonnet for building and testing; the game has an agent backdoor** (`-agent` flags: jump to a scene, set day/clock/seed, scripted input, screenshots, state dump) | Alejo: "so the agent can go exactly where it needs as to have the testing workflow as easy as possible" |
| 2026-10-13 | **Disease name: Yumetsuki (夢憑き)** for the Shoals, **S-317** for Driftwell | Alejo's pick from the Story options. Yumetsuki = "dream-possessed", explains the nurses; S-317 keeps the 3:17 pattern |
| 2026-10-13 | **Story: a nurse (Hana) stops reporting; night 2 we find she killed herself before Yumetsuki could take her, so she wouldn't hurt anyone. Plus a DDLC love triangle (Mayu / Hana / you)** | Alejo's ideas. Draft v1 of the full storyline in [[Danmin Jikken – Story]] |
| 2026-10-13 | **Open source, public repo from the first commit. Code MIT, art + story CC BY-NC 4.0** | Alejo. Claude flagged that the solution is public before the Oct 19 deadline (classmates can copy); his call |
| 2026-10-13 | **A Despair ending: you end your own life, mirroring Hana** (reached by never opening her letter + pushing Mayu away) | Alejo: "an ending where despair sits in and we end up killing ourselves". Off-screen, under the title-screen content warning |
| 2026-10-13 | **The day is the dream.** The game opens greyscale on a gurney, nurses count you down, your eyes close → the white day. You're only truly awake at night; the last scene is the gurney again (close your eyes or keep them open) | Alejo: "the time we're really awake would be at night… the day would be like a false". Story v2 in [[Danmin Jikken – Story]] |
| 2026-10-13 | **Love routes (story v3):** choices decide who confesses on night 2 (Mayu / Hana / nobody). Hana dies on every route (heartbreak, found night 3). Mayu's love slips and she corrects herself. On Hana's route Mayu kills Hana, sedates you for good and hangs herself → **FOREVER** ending (seen through your closing eyes, hard cut at the drop) | Alejo: "if she knows that she can't have you, we also get it… and then she kills herself". 5 endings now |
| 2026-10-13 | **Endings cut to 4: Sleep, Wake, Forever, 3:17 secret.** Despair is gone | Alejo: "could we reduce it to sleep wake forever and secret" |
| 2026-10-13 | **The nurses fight Yumetsuki by injecting PALE MSTR**, a white energy drink, straight into their veins. It's a Driftwell product | Alejo: "they inject white monster into their veins directly!" Name changed to a parody because the repo is public (trademark) |
| 2026-10-13 | **Day → night = a strobe**: barely there, then it builds until it's overwhelming, full white, then you fall into the night through white-ish flashes. Worse every night | Alejo. Claude added the flashing safety (max 3 big flashes/s, warning, Reduce flashing setting) because it's a public game |
| 2026-10-13 | **The day is shown through a CRT: scanlines, vignette, a little fisheye.** The night has no filter | Alejo. Fits the story: the day is Driftwell's broadcast-dream. It breaks down day by day, with an off switch in Settings |
| 2026-10-13 | **Strobe v2: flicker → stun sound (ringing whine, everything muffled) → full strobe with inverted colours (white day ↔ black)** → white → fall into the night | Alejo: "i want the full strobe effect… inverting the colors would be a good way since day is mostly white". He chose no flash-rate cap; Claude flagged the seizure risk; the warning + Reduce flashing setting stay, strobe ≤ ~1 s |
| 2026-10-13 | **Hana's room is SUPER gory, and the greyscale turns into blood-red scale for that scene only** | Alejo: "i want grescales to turn into BLOOD Red scales in just that scene". Shown as aftermath, never the act; content warning adds "graphic gore" |
| 2026-10-13 | **Agent loop: if Sonnet (tester/builder) sees something visually fishy, it reports a VISUAL issue; Opus (art) confirms it and fixes exactly that thing** | Alejo. Relayed by the main session (subagents can't message each other). Rules in CLAUDE.md → Agent policy |
| 2026-10-13 | **`hospital/` frozen** after M1 | Alejo confirmed. The brief says the model must not change once the GUI exists; the game uses only `Hospital` methods and `Snapshot()` |
| 2026-10-13 | `go run .` prints Section 6, then the concurrent demo (6 patients, 24 episodes) | Alejo: keep both. Section 6 output stays unchanged; the goroutines are visible without a flag |
| 2026-10-13 | **Score = correct choices; top-10 uses OS usernames** | Alejo: "Points are correct choices and top 10 is OS usernames" |
| 2026-10-13 | **Top-10 twist: the save file talks back.** `danmin_scores.txt` gets a 03:17 row you never played, with a message | Alejo wanted "something very DDLC style"; picked this over the board slowly filling with his username |
| 2026-10-13 | **Night models are very detailed**: the 256 px VN sprites get full detail (hair strands, folds, faces), not the simple day line art | Alejo: "on night cycle i want the models to be very detailed" |
| 2026-10-13 | **Window patient = 96×128 waist-up** (not 64×64); **text, signs, eyes and mouths don't boil**, only outlines do | Alejo, reviewing reception desk v1. Readability; the day 3 breakdown can make text wobble later |
| 2026-10-13 | **Old Cecily must read clearly old**: white/grey hair, deep wrinkles, a fish-seller apron | Alejo: v1 looked too young |
| 2026-10-13 | **Day palette gets 2–3 whitish desaturated greens** (fills only; outlines stay black; night unchanged) | Alejo: "something is getting very lost on the white on white" |
| 2026-10-13 | **Desk rules approved**: valid referral = name matches ID, ID not expired, Dr. Barrow's stamp, sleep-disorder diagnosis; ADMIT / REFER / DISCHARGE as in *What each stamp means*; explained by the orientation reel before Day 1 | Alejo: "yes those rules work" |
| 2026-10-13 | **Day palette with the three greens approved** | Alejo: "i like the pallete now" |
| 2026-10-13 | **Reception desk view approved** (v3: greens, old anime Cecily, screen glass overlay) | Alejo: "cecily v3 is super good continue building" |
| 2026-10-13 | **Ward: 3 rooms on days 1–2, 2 on day 3** | Alejo picked it: "no room" hits harder on day 3 |
| 2026-10-13 | **Beds panel and Top-10 views approved**; the glass smudge stays on the terminal; the 03:17 ghost row looks exactly like the real rows | Alejo |
| 2026-10-13 | **Fonts: the UI uses a ready-made open-licensed pixel font (Silkscreen, the Figma placeholder). Only the nurses' handwriting is drawn in Aseprite: Hana = neat joined cursive, Mayu = psycho style, jagged separate letters** | Alejo: "Only drawn are the letters (to distinguish hana (cursive) with mayus (pshycho style) so joined and apart letters". Lets day 3's "Hana's reports in Mayu's handwriting" read at a glance |
| 2026-10-13 | **No room alert view approved** | Alejo |
| 2026-10-13 | **Day 1 tutorial is scripted: Old Cecily nods off at the window mid-joke**; you click her (learn the mechanic), she wakes, then you stamp her. Needed because only admitted patients have goroutines now | Alejo: yes to the scripted nod-off |
| 2026-10-13 | **Score: 100 per right stamp, 50 per attended sleeper; one top-10 row per shift** | Builder's proposal, Alejo: "keep it" |
