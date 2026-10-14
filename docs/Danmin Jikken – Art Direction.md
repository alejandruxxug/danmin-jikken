# 🎨 Danmin Jikken – Art Direction

Back to [[Danmin Jikken]]

---

## 🌸 Style: anime pixel art in two looks

| | ☀️ Day | 🌙 Night |
|---|---|---|
| **Look** | **Black outlines, white and pale-green fills.** Flat, clean, no shading, like a 1950s company pamphlet or a colouring book | **It comes alive:** Slay the Princess greyscale. Heavy shadows, hatching, sketchy texture |
| **Colours** | Black + off-white + 3 pale hospital greens (fills only) | All greys + **blood red** |
| **Feeling** | Cheerful, sterile, corporate. Driftwell's version of the world | The real world under it |
| **Red** | **Never, except the day 3 stamp ink**, where the red first leaks into the day (and the game breaks from there). The only colour in the day is the **violet-blue stamp ink** | Blood, and the twisted stuff only |

### 〰️ Day lines squiggle (line boil)

On the day, **every outline wobbles**: characters, the desk, the room. This is the animation trick called **line boil** (Ed, Edd n Eddy; Courage the Cowardly Dog). Each outline exists in **3 slightly different hand-wobbled versions** that cycle at ~8 fps, so nothing ever sits perfectly still. It's quiet and unsettling: the cheerful, clean world never quite holds its shape.

- **In pixel art:** 3 frames of the line-art layer, each with 1-px wobbles in different places. Fills and eyes stay put; only the outlines move.
- **In code:** the sprite compositor cycles `lineart_0/1/2` on a timer. Cheap, and every character gets it automatically.
- **Two speeds (Alejo, 2026-10-13):** *"characters to have the boil line run at normal fps but backgrounds and all that stuff slower so it shows the instability, focus and cool things"*. **Characters** (patients, staff) boil at **125 ms/frame (8 fps)**; **backgrounds, the desk, props and UI** boil **slower, 375 ms/frame** (starting value, tunable). The people twitch, the room breathes, and the eye goes to the people. Done in code with one per-sprite speed table; the art keeps its 125 ms frames. The screen glass doesn't boil.
- **Escalation idea:** the boil gets faster or wider each day, and on day 3 it's visibly wrong.
- **Night:** steadier, Slay the Princess-style lines, so the night feels *more real* than the day.

**Why the day looks like this (story v2, 2026-10-13):** the day is a **dream**. Flat, white, too perfect, and the lines can't hold still. The night is the real world, which is why the game's **very first frame is greyscale** (the gurney). See [[Danmin Jikken – Story]].

**The switch is the scare:** the same character, same pose, goes from flat line art at the desk to shaded, textured and bloody at night. One sprite set covers both (see the layer system below).

**References:**
- **Slay the Princess** for the night: pencil/charcoal greyscale, white highlights on dark backgrounds, red only for blood and roses. Alejo's reference board, 2026-10-13.
- **DDLC** for the anime faces and the VN framing.
- **Fallout** manuals and Vault-Tec posters for the day: black line art, flat, cheerful. The day look *is* Driftwell's pamphlet.
- **VA-11 Hall-A** for anime pixel art with a job sim and a story told over shifts.

**Translating Slay the Princess into pixels:** StP is hand-drawn charcoal, so the pixel version gets its texture from **hatching and dither patterns** (like Return of the Obra Dinn) and uneven, sketchy outlines, not smooth gradients. That's the hardest part of the art and where most of the review rounds will go.

## 🩸 Hana's room redscale

Alejo, 2026-10-13: when you open Hana's door on night 3, **the greyscale turns into blood-red scale**, for that scene only. It's the goriest image in the game, and the only time the *whole screen* is red.

| Normal night grey | → | Red twin |
|---|---|---|
| Black `#0b0b0c` | → | `#0b0b0c` (stays black) |
| Dark grey `#2e2e33` | → | Dark blood `#4a0000` |
| Mid grey `#6b6b70` | → | Blood red `#8a0303` |
| Light grey `#b9b9b4` | → | *new* `#c41a1a` |
| Off-white `#ecece6` | → | *new* `#e8a6a0` |

- **Art:** the room is drawn normally in greys (every surface soaked: floor, walls, sheets, ceiling), with a `fx_blood` layer for the pools, drips and handprints. Hana is a **silhouette** slumped against the bed, turned away. The aftermath, never the act.
- **The envelope is the only thing that stays white.** It's on its own layer that the remap skips, so the eye goes straight to it.
- **In:** the red bleeds up **from the floor**, ~1 s. **Out:** it drains back to grey **from the top down** when the orderlies grab you.
- **In code:** a `redscale` number in the same night shader (0 = grey, 1 = full red twin), plus a keep-white mask for the envelope layer.
- **Palette file:** add the 2 new reds to `assets/palette/danmin.hex` when the room gets drawn.

## 📺 Day screen CRT

Alejo, 2026-10-13: the whole **day shift** is shown through an old **CRT screen**: like watching Driftwell's orientation reel that never ends. **The night has none of it.** No glass between you and the real world, so the night is flat, sharp and full screen. The jump from curved, glowing day to raw night is part of the strobe's punch.

| Effect | Amount | Notes |
|---|---|---|
| **Scanlines** | 1 dark line per game pixel row (every 4 screen px), ~15 % darker | Over the ×4 image, so they line up with the pixels |
| **Phosphor glow** | Soft bloom on the off-white, small | Makes the white feel like light, not paper |
| **Vignette** | Corners ~35 % darker, centre untouched | Pulls the eye to the desk and the patient |
| **Fisheye (barrel)** | **Small**: the corners bend in a few pixels | Enough to feel like curved glass. Text and stamps in the middle must stay perfectly readable |
| **Flicker** | Barely: brightness ±2 %, now and then | Shares the hum with the strobe's flicker phase |
| **Filmed-CRT: rolling band** | A soft band ~25 % of the screen tall, ~10 % darker with a thin brighter leading edge, crawling down the screen once every ~6 s, always on | Alejo, 2026-10-13: *"the effect that happens when you record a crt with a cam"*. The camera and the screen refresh out of step. Day 2's "rolling bar" is this band getting stronger and faster |
| **Filmed-CRT: moiré** | Faint wavy ripples, a sine at almost the scanline frequency, drifting slowly; ±4 % brightness, weaker in the centre | The camera's sensor grid beating against the scanlines. Text in the middle stays readable |
| **Filmed-CRT: exposure pumping + grain** | Whole-screen brightness breathes ±3 % over a few seconds; 1-screen-px grain at ~3 % | The camera re-adjusting its exposure |
| **No colour fringing** | — | The day stays in its palette (black, off-white, the pale greens). Chromatic aberration would add colours |

### 🪟 Screen glass overlay

Alejo, 2026-10-13: *"add like an overlay so it looks like we are looking through a screen"*. The scanlines alone read as a filter; the day should feel like **you are watching it on a monitor**, with glass between you and the room.

| Layer | What it looks like | Notes |
|---|---|---|
| **Bezel** | Rounded screen corners: a thin black edge that curves in at the 4 corners, like an old monitor's glass meeting its frame | Only the corners and a 2–3 px edge. Nothing covers the desk or the patient |
| **Glare** | One or two soft diagonal light streaks on the glass, top-left, very faint (off-white, ~8–12 % opacity) | Sits *on* the glass, so it doesn't pan with the room. It never crosses text you must read |
| **Dust / smudge** | A few faint specks and one fingerprint smudge | Very subtle. Day 3 can add more (the screen is getting dirty, like the desk breaking) |
| **Reflection (optional)** | A ghost of a dark shape on the glass, barely visible | A horror hook: on day 3 it could be a silhouette standing behind *you* |

**How it's made:** one **480×270 overlay sprite**, `assets/src/screen_glass.aseprite`, drawn in Aseprite with palette colours (layers `bezel`, `glare`, `dust`, `reflection`, hidden ones allowed per day). The game draws it **on top of everything, before the CRT shader**, and it does **not** pan. Night has no glass, same as the rest of the CRT.

**Figma views:** the overlay goes on top of the day views as its own layer, so the approval shows it.

**It breaks down with the days** (same escalation as the line boil):
- Day 1: clean CRT.
- Day 2: a slow **rolling bar** passes every so often.
- Day 3: **tracking errors**, the picture tears sideways for a frame, and red bleeds through (red leaks into day 3 anyway).
- **The strobe** into each night starts as the CRT losing its signal.

**Palette rule still holds:** the *art* (every sprite and background) uses only the palette. The CRT is a **screen filter on top**, so its in-between greys at the edges are fine.

**In code:** one Ebitengine **Kage shader** run on the final ×4 image, day scenes only. It takes 7 numbers (scanline strength, vignette, curve amount, day number for the breakdown, `invert` for the strobe, the time for the filmed-CRT band/moiré/pumping, and how strong the filmed look is), so it stays small enough to explain in the defense.

**Settings → CRT effect:** on by default, with an off switch (some players find curved, flickering text hard to read).

**Figma views:** Figma can't bend the image, so day views get a **vignette + scanline overlay** layer on top and a note *"+ fisheye in game"*. The approval is about the layout; the bend gets checked in the game.

## ⚡ Day to night strobe

Alejo, 2026-10-13: the cut from the day shift into each night **flickers a bit, then a stun sound hits and the screen goes into a full strobe, with the colours inverted**: the day is almost all white, so inverting turns it black. Then you fall into the night through white-ish flashes.

| Phase | Time | What you see | Sound |
|---|---|---|---|
| **1 · Flicker** | ~1.5 s | The day keeps going, but the picture **inverts for a single frame**, once or twice, and the CRT jitters. Half the players won't be sure they saw it | The fluorescent hum skips |
| **2 · Stun + full strobe** | ~1 s | **The stun sound hits, and the strobe starts: the picture flips normal ↔ inverted, over and over**. White desk ↔ **black desk with white lines**. Some flips land on a frame of the night instead (the grey ward, bed 4, a face), inverted too | **The stun:** a sharp hit, then a high **ringing whine** (like after an explosion in a film), with every other sound muffled underneath it, and a heartbeat |
| **3 · White** | ~0.5 s | Full white | Only the ringing, fading |
| **4 · Fall** | ~2 s | The night comes up **through over-exposed white**: grey shapes bleeding in, the view **drifting downward** as if you're sinking into the bed, 2–3 fading afterimage flashes | The ringing fades out; the ward monitor's beep comes in, far away, getting closer |

**The strobe *is* the inversion** (Alejo, 2026-10-13): every flash, from the first flicker to the last, is the picture flipping between normal and inverted. Nothing else flashes, no white or black frames laid on top.

**Why inverting:** the day is ~90 % off-white, so swapping black ↔ white turns the whole screen dark in one frame. It's the hardest-hitting flash the day can do, and it's free: the same image with the palette swapped.

**It gets worse every night:**

| | Strobe length | Feel |
|---|---|---|
| Into night 1 | ~0.6 s | The stun lands, short, and it's over before you understand it |
| Into night 2 | ~0.8 s | The inverted night frames last long enough to almost read |
| Into night 3 | ~1 s | The most violent. The inverted frames are red-tinged (red has leaked into day 3) |

**⚠️ Flashing safety** (full-screen black/white inversion is the flash type most likely to trigger photosensitive seizures; Alejo chose the full strobe, 2026-10-13, so these two stay):
- **Title-screen warning**, next to the content warning: *"This game contains strong flashing lights."*
- **Settings → Reduce flashing:** swaps the strobe for a single slow invert-and-fade with the same timing and the same stun sound, so nobody misses a beat.
- Keep the full strobe itself **≤ ~1 s** (the table above already does).

**In code:** one `Transition` scene between day and night, driven by a small table of keyframes per night (time → normal / inverted / white / night frame). Inverting is the CRT shader's job (one `invert` number), so no extra art. The night frames are ordinary exported sprites. Reachable with the agent backdoor: `-scene transition.night1` (2, 3).

**In the scripts:** `[FX: STROBE IN]` at the top of each night. The other direction, night → day, stays the **eyes close → hard cut to white** from the scripts.

## 📐 Resolution

| Thing | Size |
|---|---|
| Internal resolution | **480 × 270** (scaled ×4 to 1920×1080, crisp, no blur) |
| VN sprite (night scenes, the "high-res" ones) | **max 256 px tall**, ~**160 × 256** (knees-up, DDLC framing) |
| Desk portrait (patient at the window) | ~**64 × 64** px |
| Documents (Settler ID, sleep log) | ~**96 × 64** px, readable text with a pixel font |

**256 px is the cap** (Alejo, 2026-10-13). On a 270-px-tall screen a 256 sprite stands floor to almost ceiling, and the text box covers the legs, like DDLC. The face gets roughly 50–60 px, enough for anime eyes.

**Night models are very detailed** (Alejo, 2026-10-13): the night VN sprites (Mayu, Hana, …, 160×256) get hair strands, cloth folds, detailed anime faces and full greyscale shading, the Slay the Princess look. Nothing like the simple day line art: the jump from flat day to rich night is the point.

**The reception is a panorama:** the day desk background is **720×270** (the FNAF 2 pan, see [[Danmin Jikken – Game Design]]); the screen shows 480 of it at a time.

**Same pixel scale everywhere:** sprites are drawn at 1× in the 480×270 world and scaled ×4 with everything else. Mixing pixel sizes looks pasted on.

## 🎨 Palette: `assets/palette/danmin.hex`

| Role | Hex | Day | Night |
|---|---|---|---|
| Black | `#0b0b0c` | ✅ outlines | ✅ |
| Dark grey | `#2e2e33` | ❌ | ✅ |
| Mid grey | `#6b6b70` | ❌ | ✅ |
| Light grey | `#b9b9b4` | ❌ | ✅ |
| Off-white | `#ecece6` | ✅ fill | ✅ highlights |
| 🩸 **Blood red** | `#8a0303` | ❌ | ✅ blood only |
| 🩸 Dark blood | `#4a0000` | ❌ | ✅ blood shading |
| 🌿 Pale mint | `#dde5d9` | ✅ fills: walls, glass, apron | ❌ |
| 🌿 Sage | `#c3d1c0` | ✅ fills: floor, clothes | ❌ |
| 🌿 Deep sage | `#9fb19f` | ✅ fills: seats, frames, shadow side, the floor grid | ❌ |
| 🖋️ Stamp ink | `#3c3a8f` | ✅ **only** the ADMIT / REFER / DISCHARGE ink marks and their specks | ❌ |
| 🖋️ Stamp ink, thin | `#8784c4` | ✅ a few lighter specks inside the ink marks | ❌ |
| 🩸 Blood red on day 3 | `#8a0303` | ✅ **only** the day 3 ink marks (`*_red`) | (see above) |

**Day greens** (Alejo, 2026-10-13: *"something is getting very lost on the white on white … allow some whiteish desaturated greens"*). Pale, desaturated, slightly cool hospital greens, used **only as fills** to separate planes (walls vs floor vs desk vs people). Outlines stay black, off-white stays the brightest value, nothing saturated. Appended to the end of `danmin.hex` so the old indexes don't move.
- **Under the CRT:** the scanlines and vignette darken them a little; the planes stay readable.
- **In the strobe (inverted):** they don't turn pink. Inverting a pale colour makes it dark, so the greens become **dark muted plum** (`#221a26`, `#3c2e3f`, `#604e60`) under white lines. It reads as a bruised, night-like version of the room, which suits the strobe. Preview: `assets/preview/reception_v2_view_inverted_x4.png`.

*Proposal: add 1–2 more greys if the StP shading needs them.*

## 🧩 Sprite system: layers

Each character is **one `.aseprite` file whose layers** are composed at runtime (exported as one spritesheet, see the pipeline below):

```
lineart (pose) ×3 boil frames  →  eyes [normal | sleepy | closed | wide | ...]  →  mouth [neutral | smile | open | slack]
                →  NIGHT ONLY: shading (greys + hatching)  →  overlays [blood, tears, shadow]
```

- **Day = the line-art layers on white.** Night = the same line art + the shading layer + overlays. Drawing a character once gives both looks.
- **Expressions = swapping 2 layers**, so the same character never drifts.
- This is also OOP in the code (see [[Danmin Jikken – Architecture]]).

## 🛠️ Pipeline: everything is drawn in Aseprite, exported as spritesheets

**Rule (Alejo, 2026-10-13): every piece of art is designed in Aseprite, exported to the game folder as a spritesheet, and the game only reads the sheets.** No code-drawn pixels, no hand-made PNGs, no editing exported files.

```
assets/src/mina.aseprite  ──draw──  Claude (aseprite MCP) or Alejo (Aseprite app)
        │
        ▼  ./assets/export.sh
game/assets/sprites/mina.png    one packed sheet: every layer × every frame
game/assets/sprites/mina.json   where each "<layer>#<frame>" sits in the sheet + frame times + tags
        │
        ▼  //go:embed
the game cuts frames out of the sheet with SubImage and stacks layers (internal/art)
```

### 🚦 Order of work: approve the look BEFORE building (Alejo, 2026-10-13)

**Sprites in Aseprite → views in Figma → Alejo approves → only then into the game.** No game screen gets coded until Alejo has approved its Figma view.

| Step | Who / tool | Output | Gate |
|---|---|---|---|
| 1. **Sprites** | Claude, `aseprite` MCP | `assets/src/<name>.aseprite` | Alejo OKs each sprite (preview PNG or the Aseprite app) |
| 2. **Views** | Claude, **Figwright** (`mcp__figwright__*`, never the official Figma MCP) | One Figma frame per game screen, built from the real sprites | — |
| 3. **Approval** | Alejo | ✅ / changes, logged in the Art log below | **Nothing is built before ✅** |
| 4. **Into the game** | Claude, `./assets/export.sh` + Day GUI / Night VN lanes | Sheets in `game/assets/sprites/`, the screen coded to match the approved view | Matches the Figma view |

**How the Figma views are made:**
- Frame per screen at **1920×1080** (the 480×270 world at ×4). Name it after the screen: `Day – Reception desk`, `Night – VN scene`, `Title`.
- Sprites go in as **×4 PNG exports straight from Aseprite** (`export_frame` scale 4, or `--scale 4` on the CLI) into `assets/preview/` (gitignored), then `import_image` at 100%. **Never scale pixel art inside Figma**: it blurs. Every position is a multiple of 4 so it lands on the game's pixel grid.
- Text boxes, the computer UI and the HUD are drawn in Figma at the same ×4 grid, using the palette as Figma colour styles, so they can be moved into Aseprite and exported later.
- Each view has a note listing the sprites + layers it uses (e.g. `mina: lineart, eyes_sleepy, mouth_slack`), so building it later is just wiring.
- Figma is **only for the approval mock-up**. The game never reads from Figma: anything that ends up in the game is drawn/redrawn in Aseprite and exported.

### 🔁 The sprite loop (step 1 in detail)

1. Alejo sends references and the mood.
2. **Claude draws in `assets/src/<name>.aseprite`** with the `aseprite` MCP tools (`mcp__aseprite__*`).
3. Claude checks its own work: `export_frame` at scale 4–8 or `render_onion_skin` (for boil frames) into the scratchpad, then *looks at the PNG*. `get_color_stats` must show only palette colours.
4. Alejo reviews. He can open the same `.aseprite` in the Aseprite app (`/Applications/Aseprite.app`) and fix pixels by hand. That's fine: the file is the source of truth for both of us.
5. `./assets/export.sh` (or `./assets/export.sh mina` for one file) **once the view is approved**. Commit **both** the `.aseprite` and the exported sheet + json.

### 📁 Files and names

| Thing | Rule | Example |
|---|---|---|
| Source file | `assets/src/<name>.aseprite`, lowercase, one per character / prop / background | `assets/src/mina.aseprite`, `assets/src/desk.aseprite` |
| Exported sheet | `game/assets/sprites/<name>.png` + `.json`. Inside `game/` because `go:embed` can't reach `../` | `game/assets/sprites/mina.png` |
| Layers | The layer stack from the sprite system above. `kind` or `kind_variant` | `lineart`, `eyes_normal`, `eyes_sleepy`, `mouth_smile`, `shading`, `fx_blood` |
| Guide / reference layers | Start with `_`. **Never exported** | `_ref`, `_guide` |
| Expression variants | Separate layers, **hidden** while drawing is fine (hidden layers still export) | `eyes_closed` hidden |
| Frames | Line boil = **3 frames, 125 ms each** (8 fps), tag `boil`. Other animations get their own tag | tags `boil`, `nod_off`, `blink` |
| Static layers | Same art on every frame: use **linked cels**; the export stores a repeated frame only once | `eyes_normal` |
| Canvas | 1× size. VN sprite **160×256**, portrait **64×64**, document **96×64**, background **480×270** (the reception panorama is **720×270**) | |
| Colours | Only `assets/palette/danmin.hex`. Set it with `set_palette` right after `create_canvas` (first entry transparent `#00000000`) | |

### 📦 What the game reads (`mina.json`)

Aseprite's standard **json-array** format. `encoding/json` reads it directly:

```json
{ "frames": [ { "filename": "lineart#0", "frame": {"x":0,"y":0,"w":160,"h":256}, "duration": 125 }, ... ],
  "meta": { "size": {"w":320,"h":512},
            "layers":    [ {"name":"lineart"}, {"name":"eyes_sleepy"}, ... ],
            "frameTags": [ {"name":"boil","from":0,"to":2,"direction":"forward"} ] } }
```

- Frame numbers in the json start at **0**. (The MCP tools count frames from **1**.)
- Look a frame up by `"<layer>#<frame>"`. Every layer has an entry for every frame, so the game never needs special cases. Repeated frames point to the same rectangle.

### ✍️ Handwriting glyph sheets (the nurses' notes)

Alejo, 2026-10-13: the UI uses the Silkscreen pixel font; **the only lettering drawn in Aseprite is the two nurses' handwriting**, so a player knows whose note it is at a glance (day 3: Hana's reports keep coming, in Mayu's hand).

| | `assets/src/hand_hana.aseprite` | `assets/src/hand_mayu.aseprite` |
|---|---|---|
| Look | Neat, small, **joined** cursive with **looped capitals**. Tidy, clinical, tired | Her own erratic hand: **some letters apart, some joined**, sharp, uneven sizes, tilted capitals, jumping baselines. Still readable |
| Ink | Black `#0b0b0c`; joins dark grey `#2e2e33` (v3; v2's light grey vanished at 1×) | Black only |
| x-height | 5 px | 5 px (some letters 6) |
| Word space | 3 px | 4 px |
| Gap between letters that don't join | 1 px | 2 px |

Why Mayu mixes them (Alejo, 2026-10-13): *"mayu should have some letters in cursive"*, *"since the psycho minds tend to write some letters apart and some letters joined"*. Her joins are **her own** (a sharp rising stroke), never Hana's forms, so the two hands stay easy to tell apart on day 3.

**Night look:** read on off-white paper at night, so black ink + greys, **no red** (red is blood only). No boil.

**Size: cap height 7 px.** On a 96×64 note, 7-px capitals with 5-px lowercase is the smallest size where a, e, s and o stay apart with a 1-px pen. Lines sit **11 px apart** (5 lines, about 15–18 characters each).

**Format: one layer per glyph, 1 frame, a 12×12 cell.** The layer name is the character itself, so `./assets/export.sh` gives json keys like `"a#0"`, `"Y#0"`, `"?#0"`, `"…#0"`. Both sheets have A–Z, a–z, 0–9 and `. , ! ? ' : - …` (70 glyphs). Mayu's also has a 71st layer, **`join`**: her connector stroke. There's no space glyph: use the word space above.

**Metrics** (both hands):
- **Baseline = cell row 8.** Draw a glyph cell at `(penX, baselineY − 8 + jitterY)`.
- Capitals and ascenders reach cell row 2 (Mayu's taller letters reach row 1). Descenders go down to row 10.
- Each layer's **user data** is `adv=N;join=0|1` (in the json's `meta.layers[].data`). `adv` is how far the pen moves after the glyph. `join=1` means the glyph *can* join its neighbour.

**Layout, letter by letter** (inside a word; the word space resets everything):
1. Draw the glyph, then `pen += adv`.
2. Decide whether it joins the next letter. A pair can only join when **both** glyphs have `join=1`.
   - **Hana:** always joins then. Her lowercase letters (and the capitals with a tail) carry their own lead-in pixel at column 0, **cell row 7**, and exit pixel on their last column, row 7. `adv` ends right at the exit, so the next lead-in lands next to it and they connect. Row 7 is one row above the baseline, so letter bottoms stay apart and the word stays readable.
   - **Mayu:** joins only when **a seeded coin flip says yes and the current joined run is shorter than 3 letters**. When she joins, draw her `join` glyph at the pen (a 2-px rising stroke from the baseline to row 7), then `pen += 2`. Her letters carry no lead-in or exit of their own.
3. If they don't join: `pen += gap` (Hana 1, Mayu 2). For Mayu, a join and a gap are the same width, so the coin never changes how a line wraps.
4. **Mayu's jitter (Alejo: yes):** every letter gets a random `jitterY` of −1, 0 or +1 px.
5. **Seeding:** all of Mayu's random choices come from one RNG seeded with a hash of the note's text (FNV-1a) plus the word's index. The same note looks exactly the same every time it's shown, and word wrap can measure a word with the same choices it later draws.
6. Word wrap: greedy, by word width.

**What's baked into the sheets** (not the game's job): Hana's slant (capitals lean 1 px for every 3 rows; lowercase only at the ascender tops). Mayu's uneven sizes (some letters 1 px taller) and the tilt of her capitals, digits and tall letters.

**Previews:** `assets/preview/hand_hana_note_x4.png` (*"Your stamp on bed 4 was wrong."*), `hand_mayu_note_x4.png` (*"Did you hear about Cecily?"*), `hand_compare_x4.png` (the same line in both hands), `hand_hana_glyphs_x4.png`, `hand_mayu_glyphs_x4.png`. v1 copies are in `assets/preview/v1/`. The previews use exactly the rules above.

### ⚠️ MCP gotchas (found while testing, 2026-10-13)

| Gotcha | Do this |
|---|---|
| `create_canvas` adds a layer called `Layer 1` | Rename it to `_ref` (use it for references) or `delete_layer` it |
| `create_if_missing: true` creates a missing **cel**, not a missing **layer**; it fails with "Layer not found" | `add_layer` first |
| A failed tool can still come back as success, with the error only in its text ("Failed to …") | Read every result message |
| MCP `export_spritesheet` **flattens all layers into one picture** | Don't use it for game sprites. Always `./assets/export.sh`. (It's fine for a quick look) |
| Only one session should write to a given `.aseprite` at a time | It's one binary file; two sessions writing it means one loses its work |
| Figwright `apply_style_to_node` times out ("Unable to establish connection to Figma") | The palette colour styles exist in the Figma file (`Danmin/…`), but HUD shapes are painted with the exact palette hex for now |
| A fill under a boiling outline must boil too | A linked fill shows white specks or grey gaps when the outline wobbles 1–2 px. Make the fill a 3-frame layer computed from each frame's own outline |

**ComfyUI (optional, later):** night backgrounds only, downscaled and forced onto the palette by a script, then **imported into an `.aseprite` file** (`import_image_as_layer`) and cleaned up there, so it goes through the same export. Not for characters.

## 🖼️ Art log

| Date | What | Verdict |
|---|---|---|
| 2026-10-13 | v0 test patient, 40×48, day + night ![[siesta-test-patient-v0.png]] | Red rule works. **Not anime, too small, and the day is now line art. Superseded.** |
| 2026-10-13 | Style locked: day = black outlines + white fill; night = Slay the Princess greyscale + red | From Alejo's StP reference board |
| 2026-10-13 | Day outlines **squiggle** (line boil, 3 frames ~8 fps) | Alejo: "to give the unsettling effect" |
| 2026-10-13 | **Tool connected: Aseprite 1.3.18.6** (built from source) + `aseprite` MCP (diivi/aseprite-mcp, 116 tools). Pipeline tested end to end on a 32×32 boil test head: draw → preview → `export.sh` → per-layer sheet + json | Works. Alejo: everything in Aseprite, exported as spritesheets |
| 2026-10-13 | **Day – Reception desk v1** (M2). Sprites: `reception_bg` 720×270 (fill, lineart ×3, detail: waiting room with benches + window to the sea + DRIFT OFF WELL! poster, HALCYON REST entrance, notice board, fountain, WARDS A-B doors, PALE MSTR poster), `desk` 720×270 (fill, lineart ×3, detail: glass window, counter with the day-1 rules memo + ADMIT/DISCHARGE/REFER cubbies, phone, in-tray, stamps, bell, mug), `cecily` 96×128 (fill ×3, lineart ×3, detail, eyes_normal, eyes_closed, mouth_smile, mouth_slack, fx_zzz), `waiting_fisher` 32×64 (same layer set, mouth_neutral). All tag `boil`, 3 × 125 ms, only `#0b0b0c` + `#ecece6`. Figma file *Danmin-Jikken*, page *Day*: frame `Day – Reception desk` (2:9) + `Day – Reception desk · full pan (reference)` (3:25) + note (3:32) | **Waiting for Alejo's approval** |
| 2026-10-13 | v1 review (Alejo): white on white gets lost, Cecily doesn't read old, the floor grid pulls the eye | Not approved. Decided: window patient stays 96×128, only outlines boil (text/signs/eyes/mouths stay still), panorama stays 720×270, no pan arrows (edge panning like FNAF) |
| 2026-10-13 | **Reception desk v2.** 3 day greens added to the palette as fills (walls mint, floor sage, seats/frames/kickplates deep sage, glass mint, clothes sage/mint). Old Cecily redrawn: messy white hair out of the scarf, deep smile and marionette lines, jowls, small tired eyes with heavy lids and bags, sage cardigan with rolled deep-sage cuffs, stained mint apron, striped rag over the shoulder. Floor grid now thin deep-sage lines (2 across + 7 radial) instead of black. Fills under outlines boil with them. Same Figma frames updated (2:9, 3:25, note 3:32). Before: `assets/preview/v1/`, after: `assets/preview/*_x4.png` | **Waiting for Alejo's approval** |
| 2026-10-13 | v2 review: the day palette with the 3 greens is approved (*"i like the pallete now"*). Cecily: hair like flames, wrinkles like cracks, not anime enough, rag hard to read | Palette ✅. Cecily → v3 |
| 2026-10-13 | **Cecily v3** (cute-grandma anime, still old): soft cloud-shaped white tufts out of the scarf, round face, gentle tired eyes with heavy lids and a highlight, kind smile, only smile lines + under-eye bags + crow's feet (the cheek lines are gone), light deep-sage blush, a longer rag with two clear stripes and a frayed end. Same 96×128 and layers. Before: `assets/preview/v2/cecily_*_x4.png`, after: `assets/preview/cecily_*_x4.png` | **Waiting for Alejo's approval** |
| 2026-10-13 | **Screen glass overlay** `screen_glass` 480×270, 1 frame (glass, no boil, no pan): `bezel` (2-px black edge + radius-12 rounded corners, 100 %), `glare` (two dithered off-white diagonal streaks top-left, fading out above y 100, layer 11 %), `dust` (specks + a fingerprint smudge on the right wall, 25 %), `reflection` (hidden, 6 %: a dithered head-and-shoulders behind you, for day 3). Added to both Figma day frames under the scanlines/vignette. Previews: `assets/preview/screen_glass_x4.png`, `screen_glass_day3_x4.png`, whole view `reception_v3_view_x4.png` | **Waiting for Alejo's approval** |
| 2026-10-13 | v3 review: **Cecily v3 approved** (Alejo: *"cecily v3 is super good"*). With the palette and v2 room/desk already OK'd, **`Day – Reception desk` (v3, with the screen glass overlay) is approved** | ✅ Exported: `reception_bg`, `desk`, `cecily`, `waiting_fisher`, `screen_glass`. M2 visuals can be built |
| 2026-10-13 | **M3 sprites:** `computer` 448×236 (the flip-up patient terminal, FNAF 2's tablet: fill + lineart boil, detail; sage housing, mint bezel, black screen, DRIFTWELL plate, speaker, knobs), `ui_icons` 24×16 (light terminal icons, one per layer, boil: `bed_free`, `bed_occupied`, `fx_z`, `sleeper`, `alert`), `hallway_sleeper` 48×24 (a patient asleep on the hallway floor: fill + lineart boil, `eyes_closed`, `fx_zzz`). Palette only. Previews: `assets/preview/computer_x4.png`, `icon_*_x4.png`, `hallway_sleeper_x4.png` | **Waiting for Alejo's approval** |
| 2026-10-13 | **M3 views** in Figma (page *Day*): `Day – Beds panel` (6:50: the terminal over the room, tabs, 3 room cards, hallway list with CALL STAFF, staff on shift), `Day – No room alert` (6:51: room panned right, a patient asleep in front of the ward doors, NO ROOM AVAILABLE box with CALL STAFF / OPEN BEDS), `Top-10` (6:52: OS usernames, scores, times, and the 03:17 *"why did you let me sleep"* row). Notes 6:142–6:144. UI text is a Figma placeholder (Silkscreen 32). Exports: `assets/preview/figma/day_beds_panel.png`, `day_no_room_alert.png`, `top10.png` | **Waiting for Alejo's approval** |
| 2026-10-13 | M3 review: **`Day – Beds panel` and `Top-10` approved** (Alejo). Fingerprint smudge on the terminal: keep it. Ghost 03:17 row: identical to the real rows | ✅ Exported `computer`, `ui_icons`. `Day – No room alert` was rebuilt after a Figwright batch timeout; waiting for Alejo |
| 2026-10-13 | **`Day – No room alert` approved** (Alejo) | ✅ Exported `hallway_sleeper`. All 3 M3 views approved |
| 2026-10-13 | **Handwriting sheets:** `hand_hana` (neat slanted joined cursive, mid-grey join hairlines at cell row 7) and `hand_mayu` (apart, jagged, jumping baselines, heavy strokes with grey smear and bleed). 70 glyphs each, one layer per glyph in a 12×12 cell, cap height 7, baseline cell row 8, `adv`/`join` in layer user data. Night ink only, no red. Layout rules in Pipeline → Handwriting glyph sheets. Previews: `assets/preview/hand_hana_note_x4.png`, `hand_mayu_note_x4.png`, `hand_compare_x4.png`, `hand_*_glyphs_x4.png` | **Waiting for Alejo's approval** |
| 2026-10-13 | Handwriting v1 review (Alejo): Hana needs flourishes, looped cursive capitals, not slanted print; her lowercase merged at 1× (*"wrong"* read as *"wcong"*, *"was"* as *"wus"*). Mayu: *"mayu should have some letters in cursive"*, *"since the psycho minds tend to write some letters apart and some letters joined"*; her grey pressure blocks read as glitches; the game jitters her letters ±1 px (*yes*) | Not approved → v2 |
| 2026-10-13 | **Handwriting v2.** Hana: 26 looped cursive capitals (the ones with a tail join the next letter), lowercase opened up (a, e, o, r, s, w, n, u redrawn), join hairlines moved to cell row 7 and made light grey, slant kept off the small letters. Mayu: her own erratic joins through a separate `join` connector glyph (seeded coin per pair, runs of at most 3), no grey pressure, uneven sizes + tilted capitals/tall letters, ±1 px jitter at layout. Layout rules rewritten in Pipeline → Handwriting glyph sheets. Previews: `assets/preview/hand_compare_x4.png`, `hand_hana_note_x4.png`, `hand_mayu_note_x4.png`, `hand_*_glyphs_x4.png` (v1 in `assets/preview/v1/`) | **Waiting for Alejo's approval** |
| 2026-10-13 | **VISUAL: window patient drawn outside the window** while walking in from the right (over the wall, notice board and frame). Confirmed (`-day 1 -seed 7 -frames 70`). Cause: **game code**, the window sprite isn't clipped. Fix for `builder`: draw every window-row guest into the screen clipped to the glass opening, world rect **x 261, y 61, w 198, h 129** (posts' centre lines x 261 / 459, top bar y 61; the counter covers y ≥ 186), shifted by −camera. No art change | Code fix, handed to builder |
| 2026-10-13 | **Desk memo per day:** the memo text moved out of `desk` `detail` into `memo_day1` (*ADMIT ONLY WITH VALID REFERRAL.*), `memo_day2` (*SHOAL PATIENTS GO TO WARD C.*), `memo_day3` (*DO NOT WAKE THEM.*), same 3×5 lettering, DAY n header; the paper, pin and underline stay. Day 1 is pixel-identical to the approved view. Preview: `assets/preview/desk_memos_x8.png` | ✅ Fix to an approved view, exported `desk` |
| 2026-10-13 | `hallway_sleeper` boil frame 2 had an empty (see-through) head and shoe, so the head flickered white in game. Filled black like frames 1 and 3, as in the approved No-room view | ✅ Fix, exported `hallway_sleeper` |
| 2026-10-13 | **Mayu v2 approved (Alejo)** | ✅ Exported `hand_mayu` |
| 2026-10-13 | **Hana v3:** v2's letters and looped capitals kept; joins now dark grey `#2e2e33` instead of light grey (they vanished at 1×, so it read as print). Black was tried too: both read as joined, but black fuses neighbours (*wrong* → a blob) while the dark grey still shows a faint seam between letters. Previews: `assets/preview/hand_compare_x4.png`, `hand_hana_note_x4.png`, `hand_hana_glyphs_x4.png` (v2 in `assets/preview/v2/`) | **Waiting for Alejo's approval** (not exported) |
| 2026-10-13 | **Day 1 window patients** (96×128, Cecily's layer set + `mouth_neutral`; fill/lineart boil 3 × 125 ms, day palette): `matty` (12, fisherman's kid: messy black hair with a cowlick, big eyes, plaster, oversized cable-knit sweater), `wren_calloway` (34: black bob, hair clip, round glasses, worried brows, mint raincoat, belt, papers in the pocket), `hank_shoals` (58: beanie, walrus moustache, stubble, heavy lids, oilskin over bib overalls), `ivy_marsh` (26: wavy white hair, black hairband, sleepy half-lids, shell necklace, cardigan over a striped top). Previews: `assets/preview/<name>_awake_x4.png`, `<name>_asleep_x4.png`, all at the window: `window_patients_day1_x3.png`. Seated in the lobby they reuse `waiting_fisher` for now | **Waiting for Alejo's approval** (not exported) |
| 2026-10-13 | **Hana v3 approved** (Alejo: approve), dark-grey `#2e2e33` joins | ✅ Exported `hand_hana` |
| 2026-10-13 | **Day 1 window patients approved** as they are (Alejo: approve): `matty`, `wren_calloway`, `hank_shoals`, `ivy_marsh` | ✅ Exported all four |
| 2026-10-13 | **Day 2 window patients** (same spec as day 1: 96×128, `fill`/`lineart` boil 3 × 125 ms, `detail`, `eyes_normal`, `eyes_closed`, `mouth_neutral`, `mouth_smile`, `mouth_slack`, `fx_zzz`; day palette): `tobias_reef` (45, net mender: flat cap, sideburns, goatee, neckerchief, dark waistcoat), `nell_garrow` (63: stern, white hair in a tall bun, pearl earrings, high collar + brooch, fringed shawl), `ozzie_pike` (19: bleached shaggy hair over one eye, smirk, earring, hoodie with a smiley badge), `dot_harlan` (52: plump, black curly perm, smiling eyes, rosy cheeks, pearls, polka-dot dress with a round collar), `marcus_trell` (41: slick side-parted hair, pencil moustache, shifty sideways eyes, a bead of sweat, padded suit, black tie), `june_alder` (30: high black ponytail with a white ribbon, sailor collar + scarf knot). Previews: `assets/preview/<name>_awake_x4.png`, `<name>_asleep_x4.png`, at the window: `window_patients_day2_x4.png` | **Waiting for Alejo's approval** (not exported) |
| 2026-10-13 | **Day 3 window patients**, same spec, normal day palette (the day 3 breakdown is done by effects): `silas_brook` (37: sou'wester hat with chin straps, full black beard, kind crinkled eyes, oilskin with toggles), `pearl_wick` (29: choppy white pixie with a pearl pin, wide unblinking eyes with dark bags, deep-sage turtleneck, pearl pendant). Previews as above; at the window with Cecily and Hank: `window_patients_day3_x4.png` (day 1 strip: `window_patients_day1_x4.png`) | **Waiting for Alejo's approval** (not exported) |
| 2026-10-13 | **Day 2–3 window patients approved** (Alejo: approve all): `tobias_reef`, `nell_garrow`, `ozzie_pike`, `dot_harlan`, `marcus_trell`, `june_alder`, `silas_brook`, `pearl_wick` | ✅ Exported all eight |
| 2026-10-13 | **VISUAL: ceiling lights poked out beside the HUD plates** at pan 0 / 240. Confirmed: the lights sat at y 8–23, the same band as the plates (y 8–26), so panning always slid part of one under a plate. Cause: **art**. Fix in `reception_bg` `lineart` (all 3 boil frames): the 4 lights are now flat panels flush with the ceiling at **y 2–6** (same x centres, outer edge + one tube line, corners boil). Nothing of the room is drawn in x 41–679, y 8–27 any more, so no pan can put a plate over a light. The only lines left in that band are the room corner edges at x ≤ 40 / ≥ 680, part of the approved room. Before/after: `assets/preview/ceiling_fix_before_after_x4.png` | ✅ Fix to an approved view, exported `reception_bg` |
| 2026-10-13 | **M4 desk sprites** (the papers get their own look-down desk, a folder you look through; flat, seen from straight above; outlines boil, text/form lines/stamps don't; no baked shadows, the game draws them). `desk_top` 480×270 (the counter from above: top strip y 0–18 = bottom of the window glass + the folder slot x 204–276, y 9–14; desk surface mint; `memo_day1..3` as in `desk`; stamp tray with `stamp_admit` / `stamp_discharge` / `stamp_refer` as their own layers; bell + mug; the outer 24 px of the top edge kept empty for the warp). `folder` 324×214 (manila = sage: `open_fill/open_lineart/open_detail` two-page spread with a pocket on the inside left cover, 2 page edges under the stack; `open_clip` paper clip drawn over the ID; `closed_fill/closed_lineart/closed_detail` front cover with a name tab). Pages 140×180 with a dog-eared corner: `doc_referral` (letterhead, DATE/AGE/PATIENT/HOME/DIAGNOSIS/NOTES, `stamp_barrow` round deep-sage stamp, `stamp_other` Dr. Voss square stamp (hidden), `signature`, FOR DESK USE box), `doc_sleep_log` (FORM S-317, 12 rows), `doc_wristband` (band taped to a page, WARD + name + LAST STAY). `doc_settler_id` 128×80 (grew from 96×64 so the photo box fits the window sprite's face). `doc_stamps` 44×14 (`ink_admit/refer/discharge`), `doc_citation` 96×40. `ui_icons_m4` 24×16 (`eye_open`, `eye_closed`, `match`, `mismatch`, `folder`). `ui_icons` + `sev_1..3` (light, terminal) and `sev_1_ink..sev_3_ink` (paper): 2×2 dots 4 px apart, still. Palette checked on every layer and frame. Previews: `assets/preview/desk_top_x4.png`, `folder_open_x4.png`, `folder_closed_x4.png`, `doc_*_x4.png`, `icon_*_x4.png` | **Waiting for Alejo's approval** (not exported) |
| 2026-10-13 | **Rects for the builder** (1× screen px; field = the blank value row, Silkscreen caps top = row y + 1). **Tabs:** `V FILES V` (64, 253, 88, 17) next to the computer tab; on `desk_top` the same rect says `^ BACK ^`. **Folder** drawn at (12, 24): open spread x 14–333, y 26–235, spine x 174; name tab x 226–330, y 26–33, name text at (231, 28), pips 4 px after the name (names ≤ 77 px fit). **Settler ID** at (29, 58): photo (34, 88, 44, 46) = the window sprite clipped, its face at sprite (26, 14) (per-sprite offset, e.g. Ozzie (26, 18)); NAME (34, 77, 118, 7), HOME (83, 94, 70, 7), EXPIRES (83, 111, 70, 7). `open_clip` drawn after the ID. **Top page** at (186, 41), 140×180; turn-corner hotspot (310, 205, 16, 16). Page-local fields: referral DATE (8, 28, 56, 7), AGE (72, 28, 60, 7), PATIENT (8, 43, 124, 7), HOME (8, 58, 124, 7), DIAGNOSIS (8, 73, 124, 7), NOTES (8, 88, 124, 7) + (8, 97, 124, 7); physician stamp centred at (24, 128); **stamp target** = FOR DESK USE box (8, 151, 53, 18), the ink mark lands at (12, 153). Sleep log PATIENT (8, 28, 124, 7); row i = 0–11 at y 51 + 10·i: DATE (8, ·, 36, 7), TIME (50, ·, 36, 7), MINUTES (92, ·, 40, 7). Wristband WARD (51, 34, 62, 7), NAME (33, 43, 80, 7), LAST STAY (8, 70, 124, 7). Citation line 1 (5, 13, 86, 7), line 2 (5, 21, 86, 7). **Stamps on the desk** (lift targets): ADMIT (357, 106, 30, 30), DISCHARGE (397, 106, 30, 30), REFER (437, 106, 30, 30); lifted = hide the layer, draw it at the cursor with a deep-sage shadow. **Pips:** draw `sev_n` at (end of text + 4, caps top + 2) | Notes for M4 build |
| 2026-10-13 | **Nurse call light** in `reception_bg` (changes the approved reception): `call_off` (unlit dome, deep sage) and `call_on` (hidden: lit off-white dome + 6 short rays), dome centre world (572, 41), just above WARDS A–B, under the ceiling, clear of the HUD plates. Palette only, no red; the game blinks `call_on`. Previews: `assets/preview/call_light_off_x4.png`, `call_light_on_x4.png` | **Waiting for Alejo's approval** (not exported) |
| 2026-10-13 | **M4 views** in Figma (page *Day*), each a 1920×1080 frame built from the real sprites + the game's own text drawn in Silkscreen at 1× (baked into the frame image), CRT overlays on top, a note on the right: `Day – FILES tab` (14:2, reception with both tabs, FILES in a pulse colour) + `Day – FILES tab pulse` (14:15, the 5 colour states and the folder-mark tab), `Day – Desk (folder arrives)` (14:6), `Day – Desk (look-down)` (14:7, Old Cecily valid, ADMIT lifted), `Day – Desk (invalid papers)` (14:8, Ozzie Pike, no Barrow stamp, REFER lifted), `Day – Desk (sleep log page)` (14:9, day 2, 03:17 rows, citation slip), `Day – Desk (returning patient)` (14:10, wristband page), `Day – Records tab` (14:11), `Day – Sleep logs tab` (14:12, 03:17 first in every row, not highlighted), `Day – Ward feed tab` (14:13, a line + 60-s trace per patient), `Day – Nurse call light` (14:14). Notes 15:45–15:54. Exports: `assets/preview/figma/day_files_tab.png`, `day_files_tab_pulse.png`, `day_desk_*.png`, `day_records_tab.png`, `day_sleep_logs_tab.png`, `day_ward_feed_tab.png`, `day_nurse_call_light.png` | **Waiting for Alejo's approval** |
| 2026-10-13 | M4 review: **all 11 M4 views approved** (Alejo: *"i love the views!"*): `Day – FILES tab`, `FILES tab pulse`, `Desk (folder arrives)`, `Desk (look-down)`, `Desk (invalid papers)`, `Desk (sleep log page)`, `Desk (returning patient)`, `Records tab`, `Sleep logs tab`, `Ward feed tab`, `Nurse call light`. Raised by the main session, not decided yet: Records' ✓/✗ column and mismatch line give the answer away; the lifted stamp covers "SIGNATURE" on the referral; Dr. Barrow's stamp reads faint | ✅ Approved, not exported yet |
| 2026-10-13 | **M5 batch 1: night VN sprites** (Slay the Princess greyscale: auto form shading from the upper left, diagonal charcoal hatching, black outlines heavier on the shadow side; **no red anywhere**, the first red is the monitor in Night 1). **`mayu` and `hana`, 192×256, waist up, DDLC split layers** (Alejo: *"more DDLC like with the characters poses"*). 1 frame, **no boil** (Art Direction: night lines are steadier). **Layer scheme, same names on both girls**, bottom to top: `hair_back`, `body`, `head` (face + front hair + cap), `eyes_open/half/wide/closed/happy/away/cry`, `brows_neutral/up/worried/angry/flat`, `mouth_neutral/smile/grin/open/frown/pout/flat/wobble`, `fx_blush/sweat/tears/shadow` (shadow = DDLC's dark look over the eyes), then `arm_l_<pose>` and `arm_r_<pose>` (on top, so a hand can cover the mouth). A pose = one `arm_l_*` + one `arm_r_*` + one of each face group + any `fx_*`. Mayu (light messy bob, crossed clip, dark cardigan, sweater sleeves): arms `down`, `clasp`, `chest`, `mouth`, `wave`, `hug`. Hana (long black hair, blunt bangs, mole, white long sleeves, black cuffs): arms `down`, `board` (clipboard, left only makes sense), `cross`, `chest`, `mouth` (left) and `down`, `cross`, `point`, `elbow`, `chest`, `mouth` (right). Previews: `assets/preview/mayu_poses_x4.png`, `hana_poses_x4.png`, `mayu_faces_x4.png`, `hana_faces_x4.png` | **Waiting for Alejo's approval** (not exported) |
| 2026-10-13 | **M5 batch 1: night backgrounds + VN UI.** `bg_station` 480×270 (the nurses' station: corridor to WARD C lit only at the far end, shelves, clock at 3:15, the DRIFTWELL ward monitor): layers `wall`, `monitor_screen` (screen rect x 333–447, y 47–115; the game draws the traces on it), **`counter` drawn by the game IN FRONT of the girls** (top edge y 195), `prop_coffee`, `prop_can` (PALE MSTR), `prop_syringe`, `prop_note` (Hana's folded note) as separate layers. `bg_ceiling` 480×270 (Prologue: corridor ceiling from the gurney; tiles every 270 px so the game scrolls it; `ceiling` = lamp off, `light_on` = lamp + glow, toggled for the flicker). `vn_ui` 480×270: `textbox` (x 32–447, y 204–263), `nameplate` (x 44–139, y 193–207), `continue` (arrow at x 431, y 253; 3 frames × 250 ms, tag `bob`). `vn_choice` 256×22: `normal`, `hover`. Text is the game's (Silkscreen 8): name centred in the plate, caps top y 198; dialogue x 44, caps top 216, 10 px per line, 392 px wide (4 lines), off-white; your thoughts light grey without the plate; choice labels centred, caps top = button y + 9. Palette checked on every layer: only the 5 greys. Previews: `assets/preview/bg_station_x4.png`, `bg_ceiling_x4.png`, `bg_ceiling_off_x4.png`, `vn_ui_x4.png` | **Waiting for Alejo's approval** (not exported) |
| 2026-10-13 | **M5 batch 1 views** in Figma, page *Night* (1920×1080, composited at 1× from the real sprites + Silkscreen text, ×4; no CRT, no glass; a note under each): `Night – VN scene (Mayu)` (26:66, smile, hands clasped), `· bright` (26:67, waving), `· blank` (26:68, slip #1, hugging herself, shadow over the eyes), `Night – VN scene (Hana)` (26:69, tired, clipboard), `· stern` (26:70, "Go back to bed, Four.", pointing), `· soft` (26:71, holding her elbow, blush), `Night – Choice` (26:72, Choice 1 over Mayu's pout, first button hovered), `Night – Prologue (gurney)` (26:73, both faces upside down over the scrolling ceiling: the game rotates the same sprites 180°). Reference sheets ×2: 26:74–26:77. Exports: `assets/preview/figma/night_vn_mayu.png`, `night_vn_mayu_2.png`, `night_vn_mayu_3.png`, `night_vn_hana.png`, `night_vn_hana_2.png`, `night_vn_hana_3.png`, `night_choice.png`, `night_prologue_gurney.png` | **Waiting for Alejo's approval** |
| 2026-10-13 | **Stamp ink in colour** (Alejo: the stamp ink is the only colour in the day). `doc_stamps`: `ink_admit/refer/discharge` recoloured to violet-blue stamp-pad ink `#3c3a8f` with ~1 in 6 pixels in the thin-ink tint `#8784c4`; the letters are now solid (the old knock-outs ate pixels of I and H). New hidden layers `ink_admit_red`, `ink_refer_red`, `ink_discharge_red` in blood red `#8a0303` for day 3. **Juicy landing:** new sheet `ink_specks` 64×34 (`ink_specks`, `ink_specks_red`): 10 droplets of 1–5 px in a loose ring around the mark; the 44×14 mark sits at (10, 10) inside it, so draw the specks at (mark x − 10, mark y − 10); the game scatters and fades them over ~0.3 s. A separate sheet because the specks fall outside the mark's 44×14 canvas, and growing `doc_stamps` would move the exported marks the game already places. Palette file: the two inks appended to `danmin.hex`. Handles, cubbies, `desk_top` unchanged. Preview: `assets/preview/stamp_ink_x4.png`; Figma `Day – Stamp ink` (27:102, export `assets/preview/figma/day_stamp_ink.png`) | **Waiting for Alejo's approval** (not exported) |
| 2026-10-13 | **Stamp ink approved** (Alejo: approve): violet-blue `#3c3a8f` + thin-ink tint `#8784c4`, day 3 red `#8a0303`, `ink_specks` sheet | ✅ Exported `doc_stamps`, `ink_specks` |
| 2026-10-13 | **Night batch 1 approved** (Alejo: *"Wow... i ve never seen something so unique. its super increible im speechless"*): `mayu`, `hana`, `bg_station`, `bg_ceiling`, `vn_ui`, `vn_choice` and the 8 *Night* views. Polish for batch 2: hands less blocky, Hana's arms/sleeves less flat, Hana's pointing hand, a lean / head-turn pose, bigger differences between expressions. Kept: 192 px wide, no boil at night, counter in front, clock at 3:15 | ✅ Exported all six |
| 2026-10-14 | **Overnight run, night sheets (logged after the fact; drawn by the first overnight art agent):** `ward_c` (room, sleepers, blanket, dark, doorway_figure / _board / _tilt, sitters, red_eyes, clock_315 / clock_317), `corridor` (corridor, sign_station, sign_quarters, hana_far, door_crack), `note` (backdrop, paper_note, text_note, hands_note, paper_letter, hands_letter), `vn_ui` + `auto_off` / `auto_on`, `supply` (wall, bulb_glow, table, prop_cans, prop_syringe), `hana_room` (room, fx_blood, cans, hana_silhouette, envelope; palette gets the two red twins for the redscale), `procedure` (room, lamp_glow, barrow, barrow_smile, barrow_blank, rail), `fx_hands`, `fx_eyelids` (lid_top, lid_bottom), `forever` (room, mayu_chair, mayu_smile), `fx_blood_splash` (splash_1..3, stain, drip_1..5), `font_ending` (43 glyphs, cap 23), `ending_cards` (card_bg, 4 emblems), `os_dialog`, `fx_glitch`, `mayu_day`. Figma: page *Overnight review* (33:105) | Built overnight, review in the morning |
| 2026-10-14 | **Batch 1 polish of `mayu` / `hana`** (same layer names, so no code change): hands rebuilt (rounded palm that narrows to the wrist, tapered capsule fingers with round tips, middle finger longest, knuckle bumps on fists, thumb off the side of the palm, fingers can point at any angle); **`arm_r_point` now reads as a pointing index finger** (wrist moved in to (166, 138) so the hand stays on the canvas); a core shadow along the underside of every sleeve (Hana's white sleeves no longer flat); white-sleeve folds in mid grey; brows angled harder and drawn over the bangs with a 1-px mid-grey halo on Hana (they were invisible on her black hair); `mouth_neutral` is now a short soft line with a lower-lip hint, `mouth_flat` a long hard line with dark corners; `eyes_half` lower. Before: `assets/preview/v_batch1/`. After: `assets/preview/mayu_poses_b2_x2.png`, `hana_poses_b2_x2.png`, `mayu_faces_b2_x4.png`, `hana_faces_b2_x4.png`. Figma: `Review – hana_polish` (33:132), `mayu_polish` (33:133), `portrait_poses` (33:134) | Built overnight, review in the morning |
| 2026-10-14 | **`title`** 480×270, 1 frame (M7): `bg` (off-white, the corners fall into light grey by Bayer dither: the white of the dream thinning), `logo_shadow` (light-grey ink offset 2 px), `logo_kanji` (断眠実験, Hiragino Mincho rasterised and cleaned, 279×64 at (100, 44)), `logo_romaji` (DANMIN JIKKEN, Didot caps, at y 20), `logo_sub` (SLEEP DEPRIVATION EXPERIMENT, hand-made 5×7 caps, at (147, 150)), `press_rule` (two mid-grey rules at y 229, x 168–195 and 284–311: the game writes PRESS ANY KEY in Silkscreen between them, caps top y 226), `dagger` (greyscale steel, wrapped grip, guard; axis 39.9° through the centre of 断 (132, 76) and of DEPRIVATION (224, 153); pommel (71, 25), tip (234, 161); to drop it in, slide it along (0.767, 0.642)), `blood` (**hidden**: blood red on the last third of the edge + 3 drips under the tip). The letters have a 1-px gap where the blade passes (断 and DEPRIVATION are cut). For the cold open, fade `bg` in from white, then the logo, then drop the dagger. Preview `assets/preview/title_x4.png`. Figma: `Review – title` (33:135), `title_layers` (33:136) | Built overnight, review in the morning |
| 2026-10-14 | **`desk_day`** 480×270, 3 boil frames (the night engine uses frame 0), day look, for the Sleep ending: `room` (the reception from behind the counter: mint wall, DRIFTWELL *A rested Shoal is a happy Shoal!* poster, filing cabinet, a clock at 9:00, an *Employee of the month* frame with a blank face, the window to the sea, sage floor, your desk in the foreground with IN tray, stamp and DW mug), `next_desk_empty` (the empty chair back, only above the desk top), `next_desk_front` (the whole next desk with a blank name plate, lamp and tray; **always drawn last**). Draw order: empty = `room`, `next_desk_empty`, `next_desk_front`; Mayu's route = `room`, **`mayu_day` at (318, 92)**, `next_desk_front` (her chest goes behind the desk top at y 188). Previews `assets/preview/desk_day_empty_x4.png`, `desk_day_mayu_x4.png`. Figma: `Review – desk_day` (34:193). Index of the whole page: `Review – index` (34:188) | Built overnight, review in the morning |
| 2026-10-14 | **Three CGs, DDLC style** (full-screen 480×270 illustrations that replace background + characters; the normal text box sits on top; everything that matters is above y 192). Each is 3 frames × 125 ms, tag `boil` (the night engine draws frame 0 today); the girls are the same builds as the VN sprites, at 1×. **`cg_coffee`** (night1 l.23–30, your POV at the station, Mayu leaning on the counter holding the paper cup out with both hands, sweater paws): `backdrop` (bg_station wall, pushed into the dark), `mayu` (boils), `face_smile` / `face_laugh` (happy eyes, grin, sweat) / `face_watch` (wide eyes, smile), `counter`, `prop_pot`, `prop_can`, `prop_syringe` (bg_station's props, moved), `hands_cup` (arms, hands, cup: jumps 1 px per frame = the shaking), `steam`, `fx_shake` (tremble marks). **`cg_supply`** (night2 l.79–101): `wall` (supply wall + a partition with a STAFF door on the left), `bulb_glow`, `door_crack` (the door ajar, corridor light in the gap), `mayu_gap` (her real face clipped to the gap: wide eye, grin), `hana` (seated, boils), `face_down` / `face_look` / `face_soft`, `table` (top y 166–189), `prop_cans`, `arms` (the injection: sleeve rolled up, tourniquet, old needle marks, the syringe in the crook, one blood-red pixel). **`cg_kanashibari`** (night2 l.123–147, from the pillow down your body to the Ward C doorway): `room` (lit doorway + corridor, window, next bed, IV stand with the line to your hand), `figure` / `figure_board` (clipboard + cap: looks like Hana) / `figure_tilt` (tilted head, messy bob, the smile catching the light: Mayu), `bed` (footboard + blanket), `hand` (yours, limp, IV taped on); no figure layer = `figure_gone`. Which layer on which line: the notes in Figma. Palette checked on every layer and frame. Previews: `assets/preview/cg_*_x4.png`; views `assets/preview/figma/cg_*.png`. Figma: page *Night*, section `CGs` (38:324), frames 38:325–38:336, notes 38:349–38:351 (the file can't take a 4th page). Scripts: `assets/preview/lua/night/cg/` | ✅ **Approved** (Alejo, 2026-10-14: *"i like the cg"*). Not exported yet |
| 2026-10-14 | **`cg_supply` `mayu_gap` v2** (Alejo: *"i like it but at the final one (the mayu peeping through the door) make her face black only show her eyes and her smile"*). Mayu in the gap is now a solid black silhouette; only her wide eyes (the inner halves of both) and a wide off-white crescent grin show. Only that layer changed (same name, full canvas, 3 frames). Source: `assets/preview/lua/night/cg/supply_gap_v2.lua`. View `assets/preview/figma/cg_supply_4.png`, Figma `CG – Supply 4` (38:332, image 38:440) | ✅ Approved (Alejo, 2026-10-14: *"i like the art one"*), exported `cg_supply` |
| 2026-10-14 | **Cinematic pass: new sprites** (Alejo after the playtest: *"i want it more of a cinematic … the note that mayu wrote for mayu handwritten … the ending the credits and the ending name should be overlaid on top"*; script words unchanged, staging only). `fx_cinema` 480×270: `vignette`, `shade` (black dither band behind a title, centre y 192 ±46), `bars` (black letterbox, y 0–31 and 238–269), `bars_white` (hidden; Sleep's dream). `hana_door` 480×270: `door` (close-up of her door, open a crack), `card` (name card; the game writes *Hana* in `hand_hana`, card x 188–255, y 80–103). `hana_floor` **720×270** (a pan): `room`, `fx_blood`, `hand` (her hand palm up, white sleeve, black cuff; no wounds), `cans` (front ones sealed), `clipboard_snapped` (hidden, Hana route). `hana_desk` 480×270: `desk`, `fx_blood`, `envelope` (keep-white; the game writes *Four* at x 242, baseline 156). `doc_report` 140×180 (Hana's night report, day look, boil; Mayu's hand goes on the ruled lines). `hand_mayu` + glyph `♥` (adv=9;join=0) for *<3*. Palette checked on every layer. Previews: `assets/preview/hana_door_x4.png`, `hana_floor_x4.png`, `hana_floor_red_x4.png`, `hana_desk_x4.png`, `doc_report_x4.png`, `fx_cinema_x4.png`, `hand_mayu_heart_x8.png` | ✅ Approved (Alejo, 2026-10-14: *"i like the picks"*); the builder exports them |
| 2026-10-14 | **Cinematic pass: Figma storyboards**, page *Overnight review* (the file is at the Starter plan's 3-page limit), section `Cinematic pass (2026-10-14)` (37:217), 49 frames `Cine – …` (38:218–38:266) with a note under each (sprites, layers, rects at 1×, timings) and a spec card for the builder. **1 · Letters** handwritten on the paper, text box hidden, line by line (night 1 note, Hana's letter on both routes, Mayu's forged letter in her own hand with the heart). **2 · The realization**: Option 1 = punch-in ×3 on *- Hana* → match cut to the day 3 report → split-screen sting (the two signatures are pixel-identical); Option 2 = double exposure. **3 · Hana's room** in cinematic mode: bars, subtitles in the bottom bar, door close-up, the dark, the red rising, push-in, ×2 on her, a slow pan across the floor (hand → cans → clipboard), the envelope close-up. **4 · Endings** (Sleep, Wake, Forever, 3:17): bars from the start, the black + silence kept, then the rock slams in with the final still and the title overlaid, credits as cards in the bottom bar; the ending card screen, its emblems and the girl beside it are dropped. Exports: `assets/preview/figma/cinematic/*.png` | ✅ Approved (Alejo, 2026-10-14: *"i like the picks"*); Forever row being redone with the rope |
| 2026-10-14 | **Forever close-up CG `cg_forever`** 480×270, 1 frame (Alejo: *"remake the mayu death scene. make us see her more up close"*, then *"show her placing the rope while she smiles (looking loose) then a frame where she falls (doesnt show her face but the rope now tight) just there the cut"*). Replaces the far silhouette (`forever` sheet / `bg mayu_chair`). Mayu's VN design redrawn at 2×, head and shoulders, seen from below under the lamp, through the eyelid slit. Layers: `ceiling`, `mayu`, `eyes_soft` (looking at you, wet), `eyes_happy` (hidden alt), `brows_worried`, `mouth_still`, `mouth_smile` (hidden), `fx_blush`, `fx_tears`, `fx_shadow` (hidden alt), `rope_loose` (hidden: the slack loop round her neck, going up behind her jaw), `rope_hands` (hidden: her forearms raised beside her head, the loop between her fists across her bangs). Steps: *Through the gap* = face + `mouth_still`; *standing on the chair* = + `rope_hands`; *smiling at me* = `rope_loose` + `mouth_smile`; 2-s push-in 1.00→1.35. **New sheet `cg_forever_drop`** 480×270, 6 frames, layers `ceiling` (static), `lamp`, `pipe`, `rope`, `dust` (per frame). Tag `sway` (frames 1–3 × 150 ms): the rope taut, straight down out of frame, bottom 0/+2/−1 px. Tag `break` (frames 4–6 × 100 ms; Alejo: *"show the rope and then the railing breaks … a break sound into the blood splash and the forever thing"*): the pipe and plaster crack → the pipe snaps and bends down, the rope whips loose, a burst of plaster dust, the lamp jolts → the pipe hangs lower, the rope falls slack, debris sinking. A sharp metal/plaster **snap** starts on frame 4 and runs into the splash, which lands on frame 6 as before. No face, no body, no knot. Forever title: bright red `#c41a1a` with a black outline over a denser shade. Figma: `Cine – Forever · 2` (38:257), `2b` (38:434), `3a` (38:435), `3` (38:258), `3b` (40:446), `3c` (40:451), `3d` (40:452), `3e` (40:453), `4` (38:259), `6` (38:261). Previews: `assets/preview/cg_forever_1_still_x4.png`, `cg_forever_2_rope_hands_x4.png`, `cg_forever_3_rope_loose_smile_x4.png`, `cg_forever_drop_f1..f6_x4.png` | ✅ Approved (Alejo, 2026-10-14: *"WOW wow wow i love i t"*); the builder exports it |
| 2026-10-14 | **VISUAL review of the cinematic build (tester's issues 2–6).** (2, 4) Dark elliptical dither round the Wake wash/title, the floor pan, the door close-up and Forever's pipe: real, **code**: the game lays `fx_cinema` `vignette` over every cinematic shot, but no approved frame uses it. Fix for builder: stop drawing `vignette` (night/draw.go `drawOverlay`); the title's `shade` band is as approved. (3) Title warnings: real, **code**: the last line ends 1 px above the screen edge, mid grey on the dither. Fix: `warningBase` 9 → 12, `warningStep` 8 → 9, colour dark grey `#2e2e33`. (5) CRT moiré on the day 3 report and the split sting: half real, **code**. The CRT on the report shot is intended (B3), but at a ×3/×4 punch-in the filmed bands get as big as the letters. Fix: `Filmed` 0.3 and `Day` 1 for the report shot; **no CRT on the sting** (B4 has none, and its top half is the night letter). (6) AUTO toggle with no label: real, **code**: the `auto_off` / `auto_on` art is the box and the triangle only (UI words are Silkscreen, written by the game). Fix: write `AUTO` in Silkscreen 8 at x 400, caps top 252 (Draw y 249), mid grey `#6b6b70` off / off-white `#ecece6` on. Also seen: on the Wake letter insert (route_mayu) Mayu's sprite is still drawn over the letter; the insert must hide her (code) | No art change; numbers handed to builder |
| 2026-10-14 | **SCENES gallery** (Alejo: *"can you add SCENES to the game so we can replay those moments like. add this hahha"*). **Sprite** `scenes_gallery` **136×92** (card at (20, 20): draw every layer at card origin − (20, 20)), 4 frames, tag `glitch` 1–4 (140/90/60/180 ms), palette only (0 off-palette pixels), not exported: `shadow`; thumbs `thumb_prologue`, `thumb_coffee`, `thumb_note`, `thumb_confession` (new: Mayu's blank pose at the station), `thumb_supply`, `thumb_kanashibari`, `thumb_hana_room`, `thumb_forged`, `thumb_procedure` (kept, unused), `thumb_sleep`, `thumb_wake`, `thumb_forever`, `thumb_s317` (100×56, 2:1 reductions of the real art snapped to the palette); `locked` (dark grey + closed eye), `locked_317` (the same card with two slipped row bands), `frame`, `frame_hover`, and the 3:17 overlays `tear_317_locked` (hairline crack that breaks the frame; frame 3 glints with a red pixel), `tear_317` (the slot ripped open into a black void past its rect, biting Forever's corner and It wasn't Hana's title; paper-core rim, curled flaps; glitch rows, a sliver of another card, half of "03:17" and "LET ME SLEEP", one blood-red pixel), `tear_317_hover` (torn wider, more glitch). **Title**: START / SCENES / SETTINGS, menu up 16 px (`titleStartY` 223 → 207, `press_rule` 229 → 213), warnings unchanged. **Grid** in story order: Gurney, Coffee, Hana's note, **The confession** (replaces The procedure; night2 57–73, Mayu route), Supply room, Kanashibari, Hana's room, It wasn't Hana, Sleep, Wake, Forever, 3:17. Route variants replay the route seen most recently. **Settings** goes off-white too: same layout, title `bg`, ink title and selected row, dark grey hint. Figma, page *Overnight review*, section `Scenes gallery (2026-10-14)` (52:460): `Title – SCENES in the menu` (52:461), `Scenes – first run` (52:462), `Scenes – after one run` (52:463), `Scenes – all unlocked (3:17 hover)` (52:464), `Scenes – card layers` (52:465), `Scenes – 3:17 tear frames` (54:481), `Scenes – all unlocked (3:17 found, Forever hover)` (54:482), `Settings – off-white` (54:483), `Settings – off-white (MUSIC selected)` (54:484); notes 52:471–52:475, 54:488, 54:489. Previews: `assets/preview/figma/scenes/*.png` | ✅ Gallery approved (Alejo, 2026-10-14: *"i like the other ones"*). 3:17 tear, THE CONFESSION card and off-white Settings ✅ approved (Alejo, 2026-10-14: *"i really love it <3"*); the builder exports them |
