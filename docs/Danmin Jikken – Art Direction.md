# 🎨 Danmin Jikken – Art Direction

Back to [[Danmin Jikken]]

---

## 🌸 Style: anime pixel art in two looks

| | ☀️ Day | 🌙 Night |
|---|---|---|
| **Look** | **Black outlines, white and pale-green fills.** Flat, clean, no shading, like a 1950s company pamphlet or a colouring book | **It comes alive:** Slay the Princess greyscale. Heavy shadows, hatching, sketchy texture |
| **Colours** | Black + off-white + 3 pale hospital greens (fills only) | All greys + **blood red** |
| **Feeling** | Cheerful, sterile, corporate. Driftwell's version of the world | The real world under it |
| **Red** | **Never.** (Until day 3, when it leaks in and the game breaks.) | Blood, and the twisted stuff only |

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

**In code:** one Ebitengine **Kage shader** run on the final ×4 image, day scenes only. It takes 5 numbers (scanline strength, vignette, curve amount, day number for the breakdown, and `invert` for the strobe), so it stays small enough to explain in the defense.

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
| Ink | Black `#0b0b0c`; join hairlines light grey `#b9b9b4` | Black only |
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
