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
| Cold open | 0:45 | First day on the job. A cheerful corporate orientation reel (Fallout-style), then the desk. |
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

**Stamps:** ADMIT · DISCHARGE · REFER. Wrong stamps get a citation slip (Papers, Please style), but that's flavour. **No game over at the desk**, because the run has to fit in 20 minutes.

**The sleep mechanic, i.e. the assignment working in front of the room:**
- Every patient is a goroutine with its own random sleep timer (see [[Danmin Jikken – Architecture]]).
- They can **fall asleep mid-interview**: eyes close, a 💤 bubble, the papers drop.
- You can **nudge** them awake (click), but each nudge costs time.
- Their sleep times get written into their sleep log. That's how players notice things: **on day 2 every log shows an episode at 3:17 AM.**

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

These depend on the twist, so they're **locked only after [[Danmin Jikken – Story]]'s twist is picked.**

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
