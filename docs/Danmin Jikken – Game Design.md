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
| 2026-10-13 | **Every ending finishes on loud, intense rock** (like the COD death screen: the music swells, then the card shows where you "died"). **Forever above all**: hard cut to black → silence → rock slams in → ending card | Alejo: "i want the ending to always finish with rock music and very intense … especially in the forever ending since all the endings cut". Music sourced royalty-free, not composed |
| 2026-10-13 | **The rock plays over the credits**: ending card first, then the credits roll under the same track | Alejo: "that would be the credit screen" |
| 2026-10-13 | **Day 1 tutorial is scripted: Old Cecily nods off at the window mid-joke**; you click her (learn the mechanic), she wakes, then you stamp her. Needed because only admitted patients have goroutines now | Alejo: yes to the scripted nod-off |
| 2026-10-13 | **Score: 100 per right stamp, 50 per attended sleeper; one top-10 row per shift** | Builder's proposal, Alejo: "keep it" |
| 2026-10-13 | **Fullscreen**: F11 / Alt+Enter toggles fullscreen; whole-number (integer) scaling with black bars so pixels stay crisp; windowed stays the default for the agent backdoor | Alejo: "make the screen able to be viewed fullsize" |
| 2026-10-13 | **Slower blinks**: every 4–8 s (was 2–5 s), eyes closed ~0.25 s (was 0.14 s); a slow sleepy blink | Alejo: "the chars are blinking too quicky" |
| 2026-10-13 | **The papers get their own look-down desk view, a flip-down overlay**: like the computer but downward (the "v FILES v" tab, see below) the view tilts down to a top-down desk with the papers, the memo and the stamps. The approved reception view stays as it is. While you are on the desk you are not watching the room | Alejo: "we also need to make the desk view like the look down view"; picked the flip-down overlay over a Papers, Please split screen |
| 2026-10-13 | **Look-down animation, warped in code**: ~0.3 s; the reception slides up and squashes, `desk_top` rises from the bottom as a trapezoid (narrow far edge) and flattens, with a 1-2 px settle; reversed going up. The game bends the Aseprite sprites (Ebitengine `DrawTriangles`), no in-between frames drawn. Papers get a 1-2 px deep-sage drop shadow and lift when dragged (shadow grows), with a small bounce on drop | Alejo: "I thought of a too" (option A, warp in code, over hand-drawn frames) |
| 2026-10-13 | **Patients hand over a folder; you literally look through it**: a closed manila folder (name on the tab) slides onto the look-down desk; click it to open like a book. The Settler ID is clipped to the inside left cover (always visible to compare); the right side is a stack of pages you flip with a corner click: referral, sleep log, wristband (returning patients only). Page turns are bent in code like the look-down. You stamp the open folder, it closes and slides back | Alejo: "They should hand over a folder and you will look through the papers in the folder literally". Clip-the-ID layout is the main session's proposal |
| 2026-10-13 | **The desk opens with its own "v FILES v" tab** at the bottom, beside the "^ PATIENT COMPUTER ^" tab (the bottom edge already belongs to the computer tab, so no edge trigger). **In the mini tutorial the FILES tab pulses through changing colours** (cycling the day palette: off-white, mint, sage, deep sage; no red in the day) to show there are files on the desk; it stops once you open the desk. After the tutorial a small static folder mark on the tab shows files are waiting | Alejo: "the mini tutorial should make the button pulse in changing colors to indicate that we have files on the desk". The separate tab and the static mark after the tutorial are the main session's proposal |
| 2026-10-13 | **Marking very sick (Severe) patients**: (1) **severity pips** (Mild ·, Moderate ··, Severe ···, black) on the folder tab, the wristband, the BEDS cards and the WARD FEED; (2) a **nurse call light** over the ward doors in the reception that blinks while a Severe patient sleeps unattended (you miss it while on the computer or the desk; on day 3 it tempts you against *"DO NOT WAKE THEM"*); (3) **score bonus**: attending a Severe sleeper within ~10 s gives +50 on top of the 50; (4) **Severe patients get darker as they succumb**: each episode today moves their fills one day-palette step darker (off-white, mint, sage, deep sage), a palette-to-palette swap done in code, no new colours | Alejo picked pips, call light and score bonus (not the computer highlight), and added: "tinting patients that are severe darker ? so they are sucumbing to the narcolepsy by becomming darker". One step per episode is the main session's proposal |
| 2026-10-13 | **Severe darkening is built straight in code**, no preview strip or Figma view first; judged in the game. The pips and the call light are still drawn in Aseprite and approved as usual | Alejo: "do it in code there is no need to redraw" (confirmed: only the darkening) |
| 2026-10-13 | **The day CRT looks filmed by a camera**: a rolling refresh band crawling down the screen, faint drifting moiré, exposure pumping + grain, in the M5 CRT shader. No colour fringing (the day stays in its palette). Off with `-nocrt` and the settings switch | Alejo: "i want the effect that happens when you record a crt with a cam" |
| 2026-10-13 | **M4 views approved** (folder + look-down desk, FILES tab and pulse, Records / Sleep logs / Ward feed tabs, nurse call light) | Alejo: "i love the views!" |
| 2026-10-13 | **M4 view details**: the held stamp moves so it no longer covers SIGNATURE on the referral. Kept as drawn: the Records tab's ✓/✗ column and mismatch line, and Dr. Barrow's light deep-sage stamp | Alejo picked only the SIGNATURE fix |
| 2026-10-13 | **Pages are flipped by dragging**, not clicking: grab the dog-eared corner of the top page and drag it left; the page curls after the mouse (bent in code). Let go past the middle of the page and it flips and slides to the back of the stack (so the ID on the left is never covered, and the pages cycle); let go before that and it springs back | Alejo: "make the page draggeable for it to flip". Back-of-the-stack cycling is the main session's proposal |
| 2026-10-13 | **The page turn is a real page curl**: the grabbed corner follows the mouse, the page folds along the line halfway between the corner and the cursor, the folded flap shows the back of the page, and the page underneath shows through. Let go past the middle and the curl sweeps through until the page has fully turned over (then it goes to the back of the stack); before that it curls back flat | Alejo: "i want the animation to be more like when you grab a page and it like flips back (until if flips back fully)" |
| 2026-10-13 | **Night sprites are DDLC-style and very expressive**: half-body (waist up) at 256 px, several body poses per character, DDLC's split-arm system (torso + swappable left/right arm layers + head), a layered face (eyes, brows, many mouths) with overlays (blush, sweat, tears, shadow over the eyes), so a line can use pose + face, not just a face swap | Alejo: "i need the night art to be more sophisticated and more ddlc like with the characters poses and all that stuff! so they can be very expressive not just simple facial expressions" |
| 2026-10-13 | **The stamp ink is the only colour in the day**: violet-blue stamp-pad ink on days 1-2 (all three stamps, only the ink mark on the paper); on **day 3 the ink comes out blood red**, so the red leaks into the day through your own stamp. Everything else stays in the day palette | Alejo: "can we make it so the only thing with color is the stamp we give on DAY?" → "i like the color on the stamp ink!!". Violet-blue + red on day 3 is the main session's proposal, approved |
| 2026-10-13 | **Bigger hit areas on the desk**: grab the page anywhere along its right edge strip (~40 px wide, full height) instead of a 16x16 corner; a stamp lands anywhere on the open referral page and the ink snaps into the FOR DESK USE box; lift a stamp from its whole cubby | Alejo: "make the draggeable area larger and the stamp area larger? i find that its too precise" |
| 2026-10-13 | **The stamp landing is JUICY**: (1) wind-up: the held stamp lifts 3 px and stretches tall for 2 frames; (2) slam: it drops fast and squashes wide; (3) **hit-stop**: the game freezes ~3 frames on impact; (4) **screen shake** 2 px for ~0.15 s, the folder dips 1 px; (5) the ink mark **pops** in a touch too big and settles to size, with a few ink specks flying out (drawn in Aseprite as an `ink_specks` layer in the ink colour); (6) the stamp springs back up with a little overshoot; (7) the score pops up (+100 / QUICK +50) and floats away; (8) a heavy rubber **THUNK** sound (M7, Sound note). Wrong stamps get the same slam, then the citation slip slides in | Alejo: "add the note for the stamp landing to be JUICY" |
| 2026-10-13 | **Night batch 1 approved** (Mayu, Hana, nurses' station, gurney ceiling, VN text box and choices) | Alejo: "Wow... i ve never seen something so unique. its super increible im speechless" |
| 2026-10-13 | **Name plates in the speaker's handwriting**: "Mayu" is written with `hand_mayu` (jagged, half joined, ±1 px jitter) and "Hana" with `hand_hana` (looped cursive), centred in the name plate; anyone without a handwriting sheet keeps Silkscreen | Alejo: "can you make the characters names in their handwriting?" |
| 2026-10-14 | **Talk hop**: when a character starts a new line (or the speaker changes to her), her sprite hops up ~3 px and settles back with a tiny bounce (~0.33 s, slowed from 0.2 s: Alejo "i think too quick"). Only on the first line of a run, not every line, so it reads as "she starts talking". No hop on thoughts | Alejo: "Can you add a little when a character talks it jumps a bit?" |
| 2026-10-14 | **The strobe kicks in while you finish the last patient**: the flicker starts with the last patient of the day still at the desk, the strobe takes over as their folder goes back, and the shift ends inside it. The strobe speed stays as built (`-reduce-flashing` stays as the safe option). **Nothing of the night shows through the strobe**: it flashes the day / inverted day / white only, and lands on black, so the night opens on its black thoughts and the station appears only when the script shows it. The shift score is saved silently and the top-10 shows once, at the end of the run | Alejo: "the strobe should kick in while youre finishing the last patient. its fine and the flash its cool but its leaking some of what is behind before hopping into the black thoughts before waking up". Top-10 at the end of the run is the main session's proposal |
| 2026-10-14 | **Every day turns on like a CRT**: when you fall asleep at the end of a night (and after the Prologue into day 1), the day starts on black, a bright dot opens into a thin horizontal line, the line stretches up into the full picture with a white flash, then the image settles (slightly over-bright, a short roll of the band, the curve wobbles once). ~0.8 s, in the CRT shader. It mirrors the strobe: the day ends by breaking, and starts by switching on | Alejo: "make also when we fall asleep make the day turn on like a crt !" |
| 2026-10-14 | **Overnight run: the approval gate is relaxed for tonight only.** Art draws the M5 batch 2 placeholders, the batch 1 polish and the M6/M7 art, exports it straight away (Art log: *built overnight, review in the morning*) so the builders can wire it in; M6, then M7, get built without waiting. Anything Alejo doesn't like in the morning is redrawn and re-exported (the code reads layer names, so art swaps don't touch code). Sound uses the Sound note's top recommendations as **provisional picks** until Alejo fills in *Your picks* | Alejo: "start working on everything that you can since ill go to sleep. Finish the placeholders too we need this DONE by morning. if you finish m6 start M7" |
| 2026-10-14 | **Forever ending: blood splashes on the camera at the cut.** She leaps, you see her falling, and on the cut frame blood splashes across the lens (2-3 frames of the splatter spreading over the last image). The picture cuts to black but the blood stays on the glass, dripping slowly in the silence; then the rock slams in and the ending card comes up behind the blood. Art: a 480x270 `fx_blood_splash` sheet in blood red (splash frames + separate drip layers); the game slides the drips | Alejo: "for the forever ending can we splash blood on the camera just when it cuts? So when she leaps into the rope and the cut as shes falling" |
| 2026-10-14 | **AUTO play in the VN**: an AUTO toggle on the text box (and the A key) advances lines by itself. After a line finishes typing it waits **1.2 s + 0.05 s per character, capped at 6 s**. It stops at choices (never picks for you) and at big FX beats; clicking still advances; the setting is remembered | Alejo: "add a playing button so the text auto plays decide it o the character amount". The exact timing is the main session's proposal |
| 2026-10-14 | **Nights open on black thoughts, then the eyes open**: the night's first thoughts play on black; then the eyes open onto Ward C for a beat before the corridor. (The Night 1 doc had Ward C behind the opening thoughts; changed to match what Alejo described) | Alejo: "before hopping into the black thoughts before waking up". Main session's reading, review in the morning |
| 2026-10-14 | **Ending cards**: the ending title is **LARGE, ALL CAPS, in an 8-bit serif italic** (Times-New-Roman-like capitals in pixels, more definition than Silkscreen). Alejo's exact font is **That That New Pixel (Italic)** by That That Creative (seen in CapCut). Its free version is personal-use only, so tonight art draws a matching glyph sheet `font_ending`; if Alejo buys a licence, the builder switches the titles to the real .otf. **Forever**: the title is blood red. **Each girl's ending**: that girl stands to the side of the card in a fitting pose. **The secret ending**: the screen lags out (stutters, freezes, tears) and a TON of glitch effects everywhere | Alejo: "make the ending font a special 8 bit italic … like times new roman but 8 bit (in capcut its called NOVO PIXEL) … make the ending LARGE and in all caps. in the forever one make it red in each of the girls make the girl pose in the side. and in the secret make the screen lag out and but a TON glitcheefect everywhere" |
| 2026-10-14 | **Cold open: the Prologue plays before the title.** Launch → the Prologue (gurney) → the 断眠実験 title drops → start → day 1. On later launches the Prologue can be skipped with a key (a saved flag remembers you've seen it) | Alejo: "i want the prologue before the title". Skip-on-later-runs is the main session's proposal |
| 2026-10-14 | **The Prologue is the "dream": white-ish.** The whole Prologue is washed out and overexposed: blacks lifted to a light grey, a soft white glow/bloom, a white haze at the edges, the same art underneath (a screen filter in code, like the CRT, Prologue only). **Closing your eyes fades to white, not black**, there. The real nights stay dark | Alejo: "when you close your eyes it turns white and the prologue should be whiteish. since its in the dream" |
| 2026-10-14 | **Cinematic pass, no script changes.** Notes and letters are written on the paper in the writer's own hand (the Hana-route letter in Mayu's hand, because she forged it). The "it wasn't Hana" beat, Hana's room and the four endings get staged like a film, with the ending title and credits laid over the last image instead of separate screens. Shots with no line wait for a click (`hold`). | Playtest: *"i didnt really like the ending i want it more of a cinematic"*, *"I dont want yo change the script but make those scenes more contundent"* |
| 2026-10-14 | **Full-screen CGs, like DDLC.** New CGs: Mayu handing you the coffee (night 1), Hana injecting herself with Mayu's smile in the door gap (night 2, Hana route), the kanashibari figure seen from the bed (night 2). Hana's room, the gurney and each ending's last image come out of the cinematic pass as CGs too. Maybe later: Hana watching from the corridor, Mayu on the counter, a CG gallery. | Alejo: *"where could we integrate like full on scenes? like on ddlc"*, then *"YEaaaa"* to that pick |
| 2026-10-14 | **CGs approved; the Forever chair shot is redone as a close-up.** The three night CGs (coffee, supply room, kanashibari) are approved. In Forever, Mayu on the chair is no longer a small far-away silhouette: we see her face up close through the gap of your eyelids, smiling. It stays implied: no rope, nothing shown after the cut. | Alejo: *"i like the cg but remake the mayu death scene. make us see her more up close"* |
| 2026-10-14 | **Forever close-up shows the rope.** Through the eyelid gap, Mayu settles a loose rope around her neck while she smiles at you. Then one short frame with no face: only the rope, pulled tight. Then the hard cut and the blood splash, as before. Replaces the earlier "no rope" version. | Alejo: *"show her placing the rope while she smiles (looking loose) then a frame where she falls (doesnt show her face but the rope now tight) just there the cut to the forever animation"* |
| 2026-10-14 | **Cinematic pass picks.** (1) Hana's day 3 reports get a folder page in Mayu's hand (*"All asleep. All happy!"*, *"Nobody woke up."*, *"- Hana"*). (2) Mayu's spiky hand stays for the forged letter. (3) The realization is Option 1: ×3 punch-in on "- Hana", match cut to the day 3 report, split-screen sting. (4) The separate ending-card screen goes; the title and credits sit on the last image. (5) AUTO hidden in letterbox scenes (A key still works). (6) The small push-in wobble is accepted. (7) Sleep gets white bars and a black title. Plus a content warning on the title screen (suicide, self-harm, death, gore). | Alejo: *"i like the picks"* |
| 2026-10-14 | **Forever: the pipe breaks.** After the tight rope, the ceiling pipe it's tied to cracks and snaps (crack, snap, debris, about 0.3 s). The snap sound runs straight into the hard cut and the blood splash, so the blood now has a cause: she falls. Whole Forever sequence approved. | Alejo: *"show the rope and then the railing breaks since it breaks it would cause a break sound into the blood splash and the forever thing"*, then *"WOW wow wow i love i t"* |
| 2026-10-14 | **Fullscreen: the visual novel fills the screen's width.** In fullscreen the night scenes, the Prologue and the endings are scaled to the whole width (a non-whole scale, drawn "sharp bilinear": nearest up to the next whole number, then linear down) so there are black lines only above and below. Day scenes, the title, the strobes and windowed mode keep the whole-number scale with black bars (the day CRT is untouched). | Alejo: *"i want to fit the WIDTH of the screen when we are on fullscreen measure it on the visual novel. so the visual novel only has black lines on top and on the bottom"* |
| 2026-10-14 | **Only Esc skips the Prologue** on later launches. Clicks, Space and Enter read it line by line like any scene. Before, any click skipped it, so the first click threw the whole scene away. | Alejo: *"the first scene when we click it takes us straight to the title"* |
| 2026-10-14 | **A chime for a right call, and snoring that is alive.** A soft chime on the +100 of a right stamp. Snores: several recordings picked at random, random pauses, more sleepers snore more often, and the volume depends on the place: quiet in a room (behind a wall), loud in the hallway. | Alejo: *"add a chime when we get good patients. also the snoring is repeating itself"*; *"when they are snoring on a room make it less audible. but when they are on the hall make it sound loud"* |
| 2026-10-14 | **The Prologue is the real world, so it is no longer white.** It draws with the normal night look (no dream filter, no lifted blacks) and its eyelids and fades go black like the nights. After the eyes close it **cuts to white in one frame** (no fade); that white carries into the title drop as before. **Replaces** the earlier row that made the Prologue the white "dream" (the `dream.kage` filter is removed). | Alejo: *"the prologue is in white which as its reality it shouldnt be! it should be a cut to white"* |
| 2026-10-14 | **The desk makes paper noises.** The folder slides, its cover opens and shuts, a lifted page rustles (one of three, at random), the turn has its own sound, and a page that springs back flat only whispers. | Alejo: *"when we lift a paper or open a folder a sound for it should play!"* |
| 2026-10-14 | **SCENES gallery on the title.** START / SCENES / SETTINGS. SCENES shows the key moments (Prologue, coffee, Hana's note, supply room, kanashibari, Hana's room, the forged letter, the procedure, the four endings). Seen ones replay from their start; unseen ones stay locked; unlocks are saved between launches. 3:17 stays hidden until found. | Alejo: *"can you add SCENES to the game so we can replay those moments like. add this hahha"* |
| 2026-10-14 | **The music fades before the TV turns on, every time.** Title to day 1, night 1 to day 2, night 2 to day 3: the scary track fades out over about 1.2 s while the CRT turns on, then the day hum comes in. The endings keep their hard cut. | Alejo: *"fade the scary music before turning on the TV since its now a hard cut"*, then *"not only at the start, everytime it transitions from night to day"* |
| 2026-10-14 | **Night 3 has music that fades before the ending.** A dark, slow track starts when the eyes open, is silent in Hana's room, returns at the gurney and fades out over about 5 s during Barrow's last lines, so the ending starts in silence (then the rock, as before). | Alejo: *"add music on the third night but it fades as we approach the ending (just before the ending CG)"* |
| 2026-10-14 | **SCENES gallery design approved** (off-white page, 4×3 shrunk thumbnails, closed-eye locked cards). Mayu's confession (night 2) replaces The procedure; Hana's real letter gets no card; route variants replay the route seen last; Settings moves to the off-white look too. **The 3:17 card is a tear in the gallery**: a hairline crack while locked, ripped open and glitching once found. | Alejo: *"can the secret one be like a break in the scene. so its like bugged but also like tearing through?"*, *"i like the other ones"* |
| 2026-10-14 | **New ending finale, all four endings.** Once you close your eyes (choice 5), the ending plays by itself on a fixed timeline (no clicks). Then the credits **scroll over the last image** (the blood on Forever), the **DANMIN JIKKEN logo** holds a while and fades to black, then **the TV turns on and shows the top-10**. **Forever music:** Ironbound plays continuously, starting earlier so its build-up runs under the scene and the default drop at **1:24.7** (his pick after trying 2:12) lands exactly when the ending title pops; no silence gap. | Alejo: *"i want the ending animation to auto play (when we are closing our eyes). also make the credits scroll on top of the blood at the end make the title DANMIN JIKKEN stay on the screen for a while until it fades to black. Make the tv turn on and display the high score"*; picked all four endings and "build up, drop on the title"; *"i like ironbound default drop"* |
| 2026-10-14 | **The Sleep ending turns into a day, so it gets the day CRT.** From the white of the Driftwell reel onward (the reel line, then `desk_day`, the title and the finale) the Sleep ending is drawn through the same CRT shader as the day shift: scanlines, glow, vignette, barrel, flicker, rolling band, moiré, exposure and grain at day 1 strength. The whole frame is on the glass, bars, title and finale included, the way the day scenes draw their own UI. It follows Settings → CRT effect, `-nocrt` and F10. The procedure room before the reel, and the other three endings (all night), stay without it. *"on the endings that are about going to sleep (ending up in day) can you make the crt like the day?"* (Alejo) |
| 2026-10-14 | **Day 1 is a real tutorial.** (1) A Driftwell "HOW TO READ A REFERRAL" orientation sheet is page 1 of the first folder and the last page of every folder: the 4 checks (ID not expired, name matches, Dr. Barrow's stamp, diagnosis is a sleep disorder) → all ✓ ADMIT, any ✗ REFER, wristband DISCHARGE. (2) Guided first patients: Cecily step by step with pulsing highlights, Wren's expired ID teaches REFER, then Matty/Hank/Ivy alone. (3) First sleeper: hint to click them. (4) One-time hints when a new rule first appears (day 2: name mismatch, no stamp, Ward C; day 3: wristband → DISCHARGE, do not wake them). (5) The citation slip says why (e.g. ID EXPIRED → REFER). Hints come from **Driftwell Orientation** in a small memo box; skippable on later runs. **Built 2026-10-14** (Build Plan progress log): Esc turns the hints off for good, Settings has TUTORIAL HINTS, they are off after the first finished run. | Alejo: *"im having some problems getting to know what are the things or the criterion on what to do with patients"*, *"do it !!!!"* |
| 2026-10-14 | **YOU at the window (day 3) wears a hospital gown with a patient tag on the chest.** You first see only the chest and the tag through the glass. You still stamp your own file (the secret ending needs it). When you look back up from the desk, the camera tilts up to the face: it's you. Stun, then the strobe into night 3. | Alejo: *"i want the hospital gown BUT BUT with like a tag on the chest. it makes you look up and there it stuns you and transitions to night"* |
| 2026-10-14 | **The Sleep ending turns into a day, so it gets the day CRT.** From the white of the Driftwell reel onward (the reel line, then `desk_day`, the title and the finale) the Sleep ending is drawn through the same CRT shader as the day shift: scanlines, glow, vignette, barrel, flicker, rolling band, moiré, exposure and grain at day 1 strength. The whole frame is on the glass, bars, title and finale included, the way the day scenes draw their own UI. It follows Settings → CRT effect, `-nocrt` and F10. The procedure room before the reel, and the other three endings (all night), stay without it. *"on the endings that are about going to sleep (ending up in day) can you make the crt like the day?"* (Alejo) |
| 2026-10-14 | **Day 1 is a real tutorial.** (1) A Driftwell "HOW TO READ A REFERRAL" orientation sheet is page 1 of the first folder and the last page of every folder: the 4 checks (ID not expired, name matches, Dr. Barrow's stamp, diagnosis is a sleep disorder) → all ✓ ADMIT, any ✗ REFER, wristband DISCHARGE. (2) Guided first patients: Cecily step by step with pulsing highlights, Wren's expired ID teaches REFER, then Matty/Hank/Ivy alone. (3) First sleeper: hint to click them. (4) One-time hints when a new rule first appears (day 2: name mismatch, no stamp, Ward C; day 3: wristband → DISCHARGE, do not wake them). (5) The citation slip says why (e.g. ID EXPIRED → REFER). Hints come from **Driftwell Orientation** in a small memo box; skippable on later runs. | Alejo: *"im having some problems getting to know what are the things or the criterion on what to do with patients"*, *"do it !!!!"* |

