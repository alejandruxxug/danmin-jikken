# 🎬 Danmin Jikken – Trailer

Back to [[Danmin Jikken]] · Music in [[Danmin Jikken – Sound]] · Looks in [[Danmin Jikken – Art Direction]]

> **Status: in production (2026-10-14).** Alejo's picks: rendered by the game (record mode → MP4), ~1 min **with dark hints** (split-second flashes, no endings), English cards with Japanese accents. **Storyboard v2 approved (2026-10-14)** with the main session's picks (red-room flash as is, T15 trimmed to ~1 s of the injection, a real tape-stop, CC BY credits on the end card, the five cards). **Added since:** the blood splash on the logo slam (T19b) and the **motion pass** (Alejo: *"i just want it very motion graphicky! … It will be at the start of it"*: the trailer opens his defense). **The motion pass was approved with all big moves kept, the full strobe and the CC BY credits (Alejo, 2026-10-14); built and rendered, see Rendering.** Figma page *Overnight review*, section `Trailer (2026-10-14)`. The v1 list is kept at the bottom for reference.
>
> **Was: proposal (2026-10-13).** A ~1 minute trailer to present the game to the class, **around Oct 23** (Alejo: "it's in 10 days"). That's after the Oct 19 submission and the Oct 20 defense, so the **finished game** is in the shots: night, title and all. Record on **Oct 21–22**, never during the build days.

---

## 🧭 The idea in one line

**Cheerful company video that goes wrong.** It starts like a Driftwell ad for the job and ends in the night. The trailer follows the same twist the game does, without explaining it.

**Don't spoil:** no gurney, nothing about what 3:17 really means, no endings. Hint at them, don't show them.

## 🎞️ Shot list v2 (68 s, storyboard 2026-10-14)

**The times in this table are v2's. The motion pass below re-times everything to the beat (67 s) and wins where they differ; the commands stay.**

Every shot is a real backdoor run (`danmin -agent … -input shot.txt`, CRT on), rendered by the builder's record mode at 1920×1080, 60 fps. "Render frames a–b" = ticks of that run. Cards come from `assets/src/trailer_cards.aseprite` (one layer per card). Times are seconds in the trailer.

| # | In–out | Shot | Command (+ input) | Camera | Card | Sound |
|---|---|---|---|---|---|---|
| T01 | 0.0–3.0 | Hiring card | card `card_hiring` | 0.3 s fade up from black, hold | HALCYON REST SANATORIUM IS HIRING! · 求人 · the Driftwell moon | `bell.wav` at 0.0 · `driftwell_jingle.mp3` from 0.0, starting at 0.5 |
| T02 | 3.0–7.0 | The reception, Cecily walks up | `-scene day.reception -day 1 -seed 7 -pan 0` · `mouse 476 135` until `cameraX` 120, then `mouse 240 135` · frames 0–240 | the game's own edge pan | – | jingle |
| T03 | 7.0–9.0 | Never-close card | card `card_never_close` | hold | WE NEVER CLOSE. · 年中無休 · 24-HOUR CARE FOR THE SHOALS | jingle |
| T04 | 9.0–12.0 | Orientation sheet + Driftwell hint (*"FOUR CHECKS. EASY AS SLEEPING."*) | `-scene day.desk.orientation -day 1 -seed 7 -tutorial-step 2` · frames 60–240 | hold | – | jingle · `page_1.ogg` at 9.0 |
| T05 | 12.0–14.5 | The juicy stamp landing | `-scene day.desk.referral -day 1 -seed 7` · `wait 20`, `lift admit`, `wait 30`, `click 250 60`, `wait 90` · render from the lift | the game's hit-stop + shake | – | `stamp_lift.wav` 12.3 · `stamp_thunk.wav` on the slam (~12.9, put it on a beat) · `chime_right.mp3` on the +100 |
| T06 | 14.5–16.5 | The computer flips up (BEDS) | `-scene day.computer -day 1 -seed 7` · frames 0–120 | the game's flip-up | – | `click.wav` 14.5 · jingle |
| T07 | 16.5–20.0 | Cecily nods off mid-joke | `-scene day.reception -day 1 -seed 7 -pan 120` · frames 200–410 (she dozes ~260) | hold | – | the jingle starts to sag (tape-stop from 18.0) · `snore_3.wav` 18.5 |
| T08 | 20.0–22.5 | Five sleepers in a row under DRIFT OFF WELL! | `-scene day.reception -day 2 -seed 7 -sleepy -speed 2` · `mouse 240 135`, then 25 × (`wait 150`, `click 240 110`, `mouse 240 135`, `stamp admit`), `pan 0` · the last 150 frames | hold | – | jingle slowing · `snore.wav` 20.2 · `snore_2.wav` 21.0 |
| T09 | 22.5–25.0 | NO ROOM: *"DRIFT OFF WELL, FRIEND!"*, a sleeper on the hallway floor | the T08 run + `click 139 140`, `167 140`, `192 140`, `219 140` (each followed by `mouse 240 135`, `pan 0`, `wait 20`), then `pan 240` · 150 frames | hold | – | `snore_4.wav` loud 22.6 · the jingle down to ~0.3× speed by 25.0 |
| T10 | 25.0–29.0 | The memo: DAY 3 · DO NOT WAKE THEM. | `-scene day.desk.closed -day 3 -seed 7` · frames 30–270 | hold (optional slow push to the memo, 1× x 356–461, y 30–90) | (the memo is the card) | **hard cut to silence at 25.0** · one quiet `snore.wav` at 27.5 |
| T11 | 29.0–31.0 | The strobe | `-scene transition.night1 -seed 7` · frames 80–200 (strobe from ~84, then white, then the fade) | the game's strobe | – | `stun_hit.wav` on the first invert (~29.07) · `stun_ring.wav` 29.1, fading out by 32.0 · `heartbeat.wav` 29.5 |
| T12 | 31.0–33.5 | Night-shift card | card `card_night_shift` | hold | 夜勤 · THE NIGHT SHIFT | ring tail · `night_ambience.mp3` from 0.0 at 31.5 (1 s fade-in) |
| T13 | 33.5–37.5 | Coffee CG, Mayu's shaking hands | `-scene night.1 -seed 7 -line 27` · frames 0–240 | the CG's own shake | – | night music · `monitor_beep.wav` 33.6 |
| T14 | 37.5–40.5 | *"Hey. You look so tired."* | `-scene night.1 -seed 7 -line 97` · frames 0–180 | hold | – | night music |
| T15 | 40.5–43.5 | Mayu's black silhouette in the door gap (the injection shows only ~1 s first) | `-scene night.2 -seed 7 -flags find_hana -line 102` · render the last 60 frames before `layer +mayu_gap` (line 104), then on through line 105 | hold | – | `door_creak.wav` 40.6, low |
| T16 | 43.5–48.0 | Kanashibari: the figure, *"don't wake up don't wake up"* | `-scene night.2 -seed 7 -flags find_hana -line 134` · `wait 90`, then `next` until line 138 · 270 frames | hold | – | music cuts at 43.5 (the script's own `music off`) · `breath.wav` 44.0 · `creak.wav` 45.5 |
| T17 | 48.0–52.0 | **Dark-hint flashes** on black | 48.6 Hana's red room (`-scene night.3 -seed 7 -flags route_hana -line 40`, frame 40), **4 frames** · 49.4 the forged *"- Hana"* punch-in (`-flags route_hana,read_letter -line 101`, frames 0–1, before the subtitle types), **3 frames** · 50.3 Mayu in the gap (T15), **2 frames** · 51.2 the red room again, **2 frames** | cuts | – | `heartbeat.wav` 48.0 / 49.0 / 49.8 / 50.5 (speeding up) · `drip.wav` 48.6 · Ironbound building underneath |
| T18 | 52.0–55.0 | Don't-fall-asleep card | card `card_dont_sleep`, tag `flicker` (1 steady 0.5 s, 2 red slip 60 ms, 3 dim 80 ms, back to 1, hold) | hold | 眠らないで · DON'T FALL ASLEEP. | quiet `stun_hit.wav` at 52.0 · Ironbound rising |
| T19 | 55.0–61.0 | The logo drop, the dagger slam, the blood; **blood splashes onto the glass on the slam** (T19b) and stays | `-scene title -seed 7` · frames 0–360, **menu and warnings hidden** (needs a record option) | the game's drop + slam shake | – | **the drop**: see the music plan |
| T20 | 61.0–68.0 | End card, the stain still on the glass (dim) + the music credits | cards `card_end` + `end_stain` + `card_end_credits` | cut in, hold; black 67.5–68.0 | 断眠実験 · DANMIN JIKKEN · 2026 - MADE IN GO · 眠らないで | Ironbound until 66.0, hard cut · `bell.wav` alone at 66.3 |

**Don't-spoil check:** no ending, no rope / pipe / Forever, no gurney or Barrow, no YOU at the window (the day 3 shot is the desk memo only; the strobe is night 1's, so no YOU frame can leak into it). The dark parts are only the four T17 flashes (2–4 frames each).


## 🎬 Motion pass (v3, 2026-10-14) — waiting for Alejo's approval

Alejo: *"i just want it very motion graphicky! if possible since it will win us the dissertation! It will be at the start of it."* The trailer becomes a motion-graphics piece in the game's own pixel style: everything is composited on the **480×270 canvas and moves in whole 1× pixels** (= the 4-px grid at 1080p), then scaled ×4 nearest. Zooms are **whole-number punch-ins only** (×2, ×3, cut, never a smooth scale). No full-white frames anywhere except the game's own strobe (T11).

**How the compositor works:** for each trailer frame *f* (60 fps), take the recorded game frame (or a card), apply the shot's camera (crop for a punch-in), then stack the overlay layers from `trailer_fx` / `trailer_cards` at their offsets for frame *f*, ×4. Offsets come from an easing curve sampled per frame and **rounded to whole 1× px**.

**Easings** (t = 0…1 over the stated frames): `outCubic` 1−(1−t)³ · `inCubic` t³ · `inOutCubic` · `outBack` 1+2.70158(t−1)³+1.70158(t−1)² (a small overshoot) · `inQuad` t².

**The beat:** the Driftwell jingle (Hot Swing) is **~144 BPM = one beat every 25 frames**. It starts at trailer frame 30 (file 0.0); its first beat is in the file at 0.273 s, so **beat k lands on trailer frame 46 + 25·k** (k0 = 46, k4 = 146, k8 = 246 …). All first-half cuts and pops sit on these frames. Ironbound's drop (file 34.7 s) lands on the slam, **frame 3261 (54.35 s)**.

| # | Frames (s) | What moves | Easing, durations (frames) | Beat / sound |
|---|---|---|---|---|
| T01 | 0–246 (0–4.1) | **Kinetic hiring card.** `kt_hiring_bg` on from 0. The moon drops y −40 → 0. `kt_hiring_line1` slides in from x +96. The 求人 seal pops (2 frames at y −3, then 0). **`kt_hiring_stamp` slams**: y −24 (f142), −12 (f144), 0 (f146) + shake x +2, −2, +1, −1, 0 (f146–151) | moon `outBack` 18 (f0–18) · line1 `outCubic` 12 (f46–58) · seal f96 · stamp f142–151 | bell f0 · line1 on k0 (46) · seal on k2 (96) · **stamp + `stamp_thunk.wav` on k4 (146)** |
| T01→T02 | 234–246 | **Whip push:** the card slides x 0 → −480 while T02 enters x +480 → 0; `sweep` rides the seam | `inCubic` 12 | lands on k8 (246) |
| T02 | 246–496 (4.1–8.3) | The reception with the game's own pan. **Lower-third `lt_24hour`** slides in x −300 → 0, out 0 → −300 | in `outBack` 10 (f346) · out `inCubic` 8 (f471) | in on k12 · out on k17 |
| T03 | 496–596 (8.3–9.9) | **`WE NEVER CLOSE.` types itself** over `kt_never_close_bg`: letter *i* shows at f500 + 4·i (15 letters, done f556), reveal = clip `kt_never_close_text` at x 121 + edge[i]; a black 2×4-px block cursor (8×15 at 1×) after the last letter blinks 15 on / 15 off. The card itself pops in (2 frames at y +3) | 4 frames per letter | card on k18 (496) · `click.wav` 25 % per letter |
| T03→T04 | 596–602 | **CRT channel change:** `channel_static` frames 1, 2, 3 (2 frames each) | cut | k22 |
| T04 | 596–721 (9.9–12.0) | Orientation sheet. **Punch ×2** on the hint (centre 1× 255, 230), back to ×1 | cut on k24 (646), back on k26 (696) | `page_1.ogg` on k22 |
| T05 | 721–846 (12.0–14.1) | The stamp. Offset the render so **the game's slam frame lands on k30 (796)**; punch ×2 on the stamp (222, 200) for 12 frames | cut f796, back f808 | `stamp_thunk` + `chime_right` from the game, on k30 |
| T06 | 846–946 (14.1–15.8) | The computer flips up. **`lt_trusted`** slides in / out | in `outBack` 10 (f896) · out `inCubic` 8 (f938) | k32 cut · k34 in |
| T07 | 946–1096 (15.8–18.3) | **Split screen:** left = Cecily awake (T02 run, 1× crop x 120–359), right = Cecily asleep (T07 run, same crop). `split_divider` drops y −270 → 0. Then the split **collapses**: the divider runs x 0 → −240 and the asleep side widens to full | drop `outCubic` 8 (f946) · collapse `inOutCubic` 12 (f1046–1058) | k36 · k40 · `snore_3.wav` on k40 |
| T08 | 1096–1196 (18.3–19.9) | Five sleepers. **`hud_beds`** plate drops y −60 → 0, then counts 0/3 → 3/3 (frames 1–4), each change bumps the plate y −2 for 2 frames | drop `outBack` 8 · counts on k42, k43, k44, k45 | **tape-stop starts f1096**: jingle speed 1.0 → 0.3, `inQuad`, to f1346 · snores |
| T09 | 1196–1346 (19.9–22.4) | NO ROOM, the hallway sleeper. `hud_beds` 3/3 stays. **`stamp_no_room` slams** (offset +90, +70: over the floor by the sleeper), same slam curve as T01 | f1242–1251 | **k48 (1246)** · `stamp_thunk` (slowed with the tape) |
| T10 | 1346–1560 (22.4–26.0) | **Hard cut to silence.** The memo, **step zoom**: ×1, then ×2, then ×3 on the memo (1× centre 408, 60), cuts only | ×1 f1346 · ×2 f1406 · ×3 f1466 | silence · one `snore.wav` f1500 |
| T10b | 1560–1680 (26.0–28.0) | **The shift clock races to 03:17:** T10's last frame at ×1 under a 50 % black checker dim; `hud_clock_plate` drops in; the digits (Silkscreen ×4, off-white, centred x 240, caps top y 128) roll from 17:58 through the night, changing every 6 frames at first and every frame by the end; **03:17 lands at f1668** as `hud_clock_317` (blood red) and holds | plate `outBack` 8 · interval 6 → 1 `inQuad` | `click.wav` ticks speeding up · `clock_chime.wav` f1668 |
| T11 | 1680–1800 (28.0–30.0) | The game's strobe (frames 80–200 of `transition.night1`). For a flash-safe render: the same frames with `-reduce-flashing` | – | `stun_hit` f1684 · `stun_ring` |
| T12 | 1800–1950 (30.0–32.5) | **Scanline reveal:** `sweep` slides y −16 → 270; the card shows only above the band. Then `THE NIGHT SHIFT` types 2 frames per letter | sweep linear 30 (f1800–1830) · type from f1830 | `night_ambience.mp3` from f1830, 60-frame fade-in |
| T13 | 1950–2190 (32.5–36.5) | Coffee CG. **Punch ×2** on Mayu's hands (1× 240, 150) | cut f2100 to the end | `monitor_beep` f1952 |
| T14 | 2190–2370 (36.5–39.5) | "Hey. You look so tired." Out: **crescent iris closes** on her (`iris` frames 1 → 6, 4 frames each) | f2346–2370 | – |
| T15 | 2370–2550 (39.5–42.5) | **Iris opens** on the supply room (frames 6 → 1); ~1 s of the injection, then Mayu's gap; **punch ×2** on the gap (1× 50, 60) | iris f2370–2394 · punch f2490 | `door_creak` f2470 |
| T16 | 2550–2820 (42.5–47.0) | In: **glitch-slice cut** (8-px rows, each from T15 or T16, shifted ±8 / ±16 px, seed 3) for 6 frames. Kanashibari, the voice types. Out: `channel_static` 2 frames | in f2550–2556 · out f2808–2820 | music cut f2550 · `breath` f2580 · `creak` f2670 |
| T17 | 2820–3060 (47.0–51.0) | The dark-hint flashes on black (cuts only): red room f2856 (4 frames), "- Hana" f2904 (3), the gap f2958 (2), red room f3012 (2) | cuts | `heartbeat.wav` f2820, 2880, 2928, 2970, 3006 · `drip` f2856 · Ironbound from f2661 (file 24.7 s), gain 0.1 → 1.0 `inQuad` to f3261 |
| T18 | 3060–3240 (51.0–54.0) | **眠らないで tears in**: `kt_dont_sleep_glitch` frames 1–4 (3 frames each), then the clean card; `DON'T FALL ASLEEP.` types 2 frames per letter; the `flicker` tag (f2 4 frames, f3 5 frames) at f3180 | glitch f3060–3072 · type f3090 · flicker f3180 | quiet `stun_hit` f3060 |
| T19 | 3240–3600 (54.0–60.0) | The game's logo drop; **the slam at f3261** (title tick 21). **T19b, the splash:** `title_splash` f1 on slam+0, f2 on slam+2, f3 on slam+4 → `title_stain` from slam+6 to the end of T19 → `title_drips` revealed top-down 1 px/frame from slam+10. The menu and warnings stay hidden | the game's shake | **Ironbound drop on f3261** · `stamp_thunk` 50 % under it |
| T20 | 3600–4020 (60.0–67.0) | End card with `end_stain` (the same stain, dimmed) from the first frame. Builds: the logo on f3600, `2026 - MADE IN GO` types 2 frames per letter from f3630, 眠らないで pops f3690, **`card_end_credits`** steps in (Bayer 25 / 50 / 100 %, 4 frames each) at f3720. Cut to black f3990 | – | Ironbound hard cut f3960 · `bell.wav` f3978 |

**Total: 4020 frames = 67.0 s.** Style frames for every key moment are in Figma (row `Motion pass`, under the shots), previews in `assets/preview/figma/trailer/motion/`.

**New sprites:**
- `assets/src/trailer_fx.aseprite` (480×270, 6 frames, palette only): `kt_hiring_bg`, `kt_hiring_line1`, `kt_hiring_stamp` (violet-blue stamp ink, box 1× 140,134–339,170), `kt_never_close_bg`, `kt_never_close_text` (at 121,124, 237×15; letter right edges 21 36 48 69 84 105 120 138 150 168 183 201 219 234 243), `kt_dont_sleep_glitch` (frames 1–4 tearing in, 5–6 clean), `lt_24hour`, `lt_trusted` (lower-thirds, y 208–240, x 0–300), `hud_beds` (frames 1–4 = 0/3 … 3/3), `stamp_no_room`, `hud_clock_plate` (1× 160,104–319,165), `hud_clock_317`, `split_divider`, `sweep` (16-px band at y 0), `wipe_dither` (6 frames), `channel_static` (3 frames ×2), `iris` (crescent, frames 1 open → 6 closed).
- `assets/src/trailer_cards.aseprite`: plus `title_splash` (3 frames), `title_stain`, `title_drips`, `end_stain`, `card_end_credits` (*MUSIC: IRONBOUND - ZANDER NORIEGA (CC BY 3.0)* / *HOT SWING - KEVIN MACLEOD, INCOMPETECH.COM (CC BY 4.0)*, mid grey, y 240 and 250).

### 🎵 Music plan (v2 times; the motion pass table above has the v3 frames)

| Time | Music | Why |
|---|---|---|
| 0.5–18.0 | `driftwell_jingle.mp3` (Hot Swing) from 0.0, full | The cheerful company ad |
| 18.0–25.0 | the same, **tape-stop**: speed (and pitch) ramps down to ~0.3× | The ad falls asleep with Cecily. If the mixer can't change speed: a low-pass + slow fade instead |
| 25.0 | **hard cut to silence** | The memo lands in silence (one far snore at 27.5) |
| 29.07 | `stun_hit.wav` + `stun_ring.wav` | The game's own strobe sound |
| 31.5–43.5 | `night_ambience.mp3` (Paranoid Truth) from 0.0, 1 s fade-in, cut at 43.5 | The night; the cut matches the kanashibari's own `music off` |
| 43.5–48.0 | silence, breath, creak | Sleep paralysis |
| ~45.35–55.35 | `ending_rock_forever.mp3` (Ironbound) **from 24.7 s in the file**, gain 0.1 → 1.0 | Its build-up under the flashes and the card |
| ~55.35 | Ironbound **34.7 s in the file = the drop (1:24.7 of the original)** on the dagger's slam tick | The logo slam |
| 66.0 | hard cut; `bell.wav` alone at 66.3 | Bookend: the trailer opens and closes on the reception bell |

The slam tick is ~frame 21 of `-scene title` (logo lands, the dagger 14 ticks later): the builder takes the exact tick from the title code and starts Ironbound at `slam − 10.0 s` from file 24.7 s. **Credit line** (CC BY): Ironbound, Zander Noriega; Hot Swing, Kevin MacLeod (incompetech.com). Both go in the video description or a last line on the end card.

### 🃏 Text cards (`assets/src/trailer_cards.aseprite`)

480×270, palette only, 3 frames (500 / 60 / 80 ms, tag `flicker`); single-frame cards are linked across the 3 frames. English in Silkscreen (drawn at 8 px, scaled ×1–×4 by whole numbers), Japanese rasterised from Hiragino Mincho and thresholded like the 断眠実験 logo; the end card reuses the title's own kanji and romaji. Not exported yet.

| Layer | Look | Copy |
|---|---|---|
| `card_hiring` | Day pamphlet: off-white, mint bands, double black rule; the Driftwell moon in a sage nightcap with z's; a black 求人 seal top right | HALCYON REST SANATORIUM / **IS HIRING!** / RECEPTION CLERK - DAY SHIFT - START TODAY / A DRIFTWELL SLEEP UNIT |
| `card_never_close` | Same pamphlet | 年中無休 ("open every day of the year") / **WE NEVER CLOSE.** / 24-HOUR CARE FOR THE SHOALS / DRIFT OFF WELL! |
| `card_night_shift` | Black, light-grey kanji | 夜勤 ("night shift") / THE NIGHT SHIFT |
| `card_dont_sleep` | Black, off-white kanji; frame 2 a blood-red ghost slips left, frame 3 dims to scanlines | 眠らないで ("don't fall asleep") / DON'T FALL ASLEEP. |
| `card_end` | Black, the title logo in off-white with a dark-grey shadow | DANMIN JIKKEN / 断眠実験 / 2026 - MADE IN GO / 眠らないで |

Previews: `assets/preview/figma/trailer/T01…T20_*.png` (the Figma images, ×4).

### 🛠️ What record mode needs (for the builder)

1. Render a run's frames a–b to images, plus card layers from `trailer_cards` (with the `flicker` tag on T18).
2. **Hide the title's menu and warnings** for T19 (they appear ~33 frames in).
3. T17 frames 0–1 of the `-line 101` run show the punch-in before the subtitle types; if that is too tight, a "no subtitles" option.
4. The mix: per-shot sfx at the times above, the jingle tape-stop (speed ramp), Ironbound's start offset and gain ramp.

## 🎥 How to record it

1. Each shot is a short **input script** for the agent backdoor (`-agent -scene … -seed … -input …`), so a take can be repeated exactly. See [[Danmin Jikken – Architecture#🤖 Agent backdoor (test mode)]].
2. Run it with `-fullscreen` and record the screen (**Cmd+Shift+5**). Keep the CRT filter **on** (no `-nocrt`): it's the look.
3. Cut the clips in any editor (CapCut, iMovie). The sound comes from the picks in [[Danmin Jikken – Sound]], with the same credit lines.
4. Export 1080p. 1920×1080 is exactly 480×270 ×4, so the pixels stay sharp.

**Bonus for the class:** the trailer shows the real game running on the real `hospital` model. The sleepers are goroutines.

## ❓ Open

| Question | Options |
|---|---|
| Presentation date | It opens Alejo's project defense. Confirm the day |
| T11 strobe in the defense room | Full strobe (the game's) / the `-reduce-flashing` render (a projector in a lecture room is the classic photosensitive risk) |
| Motion pass | Approve / change per shot (T01 stamp slam, T07 split screen, T10b clock race, T14→T15 crescent iris are the big ones) |

*(Answered: ~1 min with dark hints; English cards with Japanese accents; red-room flash as is; T15 trimmed; real tape-stop; credits on the end card; the five cards.)* `How to record it` above is the v1 screen-recording plan; the game's record mode below replaces it.

## 🎞️ Rendering the trailer

Built 2026-10-14 (builder). The game records its shots **itself** (CRT and every shader as a player sees them); a compositor in `tools/trailer` cuts them together with the cards and the motion graphics. No screen recording.

**One command** (from the repo root; needs ffmpeg and a screen, the game opens a window while it records):

```bash
tools/trailer/render.sh                                          # records every shot, writes trailer_out/trailer.mp4 (full strobe)
tools/trailer/render.sh -reuse                                   # same shots as last time: skips the recording (about 80 s)
tools/trailer/render.sh -reuse -reduced -name trailer_reduced    # the backup: the -reduce-flashing strobe only
tools/trailer/render.sh -reuse -only 142,146,796                 # just these frames, as PNGs in trailer_out/trailer_frames/
```

| Part | What it does |
|---|---|
| `shots.txt` | The recordings, one line each: `name \| frames \| backdoor args \| input file \| from N`. 16 game runs; the inputs are in `tools/trailer/inputs/`. Cards and motion graphics are not recorded |
| `-agent -record <dir> -record-frames N [-record-from N]` | Fixed-step run, one update per saved frame, `f000001.png ...` at 1920x1080 plus `audio.json`. `-record-from` runs the first frames without drawing or saving them (long set-ups are quick). Record-only options: `-bare-title` (the title without menu and warnings), `-no-words` (the VN without subtitles and text box) |
| `timeline.go` | For each of the 4020 frames, the list of layers (which recorded frame, which sprite at which offset, clips, zoom). Pure arithmetic, table-tested (`timeline_test.go`); easings, slam, typewriter, clock race and glitch rows are in `ease.go` and `motion.go` |
| `compose.go`, `canvas.go` | The compositor. **It works on the 1920x1080 picture:** the recorded frames already have the game's CRT, so they are copied as they are (offsets in whole game pixels = multiples of 4, punch-ins whole x2 / x3 by nearest copy); sprites are drawn pixel by pixel as 4x4 blocks, so the motion layers stay crisp and the CRT look is identical to the game's. The cards and graphics have no CRT, as in the Figma views |
| `audioplan.go`, `tape.go` | The mix: beds, cues and the game's own sound effects at their frames. **Tape stop**: the jingle is read by a resampling ramp (speed 1 to 0.3, `inQuad`, frames 1096 to 1346), so speed and pitch fall together |
| `render.go` | Records, composes on all cores, pipes raw frames to ffmpeg (H.264, yuv420p, 60 fps, CRF 16, AAC 192k) |

**Outputs:** `trailer_out/trailer.mp4` (full strobe, as Alejo chose) and `trailer_out/trailer_reduced.mp4` (backup, only the strobe differs). 4020 frames, 67.000 s, 1920x1080, 60 fps, yuv420p, AAC 48 kHz stereo.

**Sync, in numbers** (the render prints them):
- **Ironbound's drop is on trailer frame 3261 and the dagger lands on frame 3261.** The bed starts at file 24.7 s on frame 2661 at normal speed, so file 34.7 s (the drop) is exactly 600 frames later. The dagger lands on the title's **frame 30** (it starts 14 ticks after the logo and falls for 16), not "tick 21" as the motion table said; T19 starts at the title's frame 9 so that frame 30 is on frame 3261.
- The game's slam in the stamp shot is its frame 96; T05 shows frames 21 to 145, so it lands on frame 796 (beat 30).
- Tape stop: the jingle's zero crossings per second fall from 1086 (frames 1100 to 1140) to 656 (1200 to 1240) to 318 (1290 to 1330) while the speed ramp ends at 0.3. Frames 1346 to 1495 are digital silence (-180 dB) except the one quiet snore on 1500; 3960 to 3978 is silent until the bell.

**Take pinning.** Everything seeded is identical every run. The sleepers (T08, T09) depend on the ward's real-time timers, so they can differ between recordings: the take that is in `trailer_out/shots/` is the take; `-reuse` keeps it, and the game's own snores for that shot are **not** used (the snores are fixed cues on fixed frames). Delete `trailer_out/shots/sleepers` to take it again and look at the result before rendering.

**Things that differ from the note's table, on purpose**
- The v3 frames win over the v2 seconds (the tape stop is 1096 to 1346, the jingle cut 1346, Ironbound's hard cut 3960, the bell 3978).
- T05 is recorded on **day 2** (`-day 2`): on day 1 Old Cecily is asleep at the window and the game refuses the stamp until she is woken; the shot is the same desk.
- Old Cecily is woken at frame 224 of the T02 run, so the "dozed off" joke is not given away before the T07 split screen.
- `sweep` (a full-width dither band) rides the T01 to T02 whip at mid height.
- The T10b clock runs from 17:58 through midnight to 03:17 in about 40 changes (gap 6 frames to 1, inQuad); the digits are Silkscreen rasterised without smoothing at x4.

**Limits.** The window shows nothing while it records. A shot is as long as its frames. The rock audition's files are not used.

## 🗄️ Shot list v1 (2026-10-13, superseded by v2)

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
