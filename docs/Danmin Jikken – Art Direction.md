# 🎨 Danmin Jikken – Art Direction

Back to [[Danmin Jikken]]

---

## 🌸 Style: anime pixel art in two looks

| | ☀️ Day | 🌙 Night |
|---|---|---|
| **Look** | **Black outlines, white fill.** Flat, clean, no shading, like a 1950s company pamphlet or a colouring book | **It comes alive:** Slay the Princess greyscale. Heavy shadows, hatching, sketchy texture |
| **Colours** | 2: black + off-white | All greys + **blood red** |
| **Feeling** | Cheerful, sterile, corporate. Driftwell's version of the world | The real world under it |
| **Red** | **Never.** (Until day 3, when it leaks in and the game breaks.) | Blood, and the twisted stuff only |

### 〰️ Day lines squiggle (line boil)

On the day, **every outline wobbles**: characters, the desk, the room. This is the animation trick called **line boil** (Ed, Edd n Eddy; Courage the Cowardly Dog). Each outline exists in **3 slightly different hand-wobbled versions** that cycle at ~8 fps, so nothing ever sits perfectly still. It's quiet and unsettling: the cheerful, clean world never quite holds its shape.

- **In pixel art:** 3 frames of the line-art layer, each with 1-px wobbles in different places. Fills and eyes stay put; only the outlines move.
- **In code:** the sprite compositor cycles `lineart_0/1/2` on a timer. Cheap, and every character gets it automatically.
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

## ⚡ Day → night: the strobe (Alejo, 2026-10-13)

The cut from the day shift into each night is a **strobe that starts as almost nothing, builds until it's overwhelming, and then you fall into the night** through white-ish flashes.

| Phase | Time | What you see | Sound |
|---|---|---|---|
| **1 · Flicker** | ~1.5 s | The day keeps going, but the desk lamp dips **once or twice**, one frame each. Half the players won't be sure they saw it. The line boil speeds up a little | The fluorescent hum skips |
| **2 · Build** | ~2 s | Flashes come in, each **brighter, bigger and longer** than the last: white frames, then **single frames of the night** (the grey ward, bed 4, a face) cut into the day | The hum turns into a rising buzz, a heartbeat under it |
| **3 · Peak** | ~0.5 s | **Full white** | Silence |
| **4 · Fall** | ~2 s | The night comes up **through over-exposed white**: grey shapes bleeding in, the view **drifting downward** as if you're sinking into the bed, 2–3 fading afterimage flashes | The ward monitor's beep, far away, getting closer |

**It gets worse every night:**

| | Length | Feel |
|---|---|---|
| Into night 1 | ~5 s | Gentle. A couple of flickers, a short build |
| Into night 2 | ~6 s | The night frames last longer; you can almost read them |
| Into night 3 | ~7 s | Violent. The flashes are red-tinged (red has leaked into day 3 anyway) and the fall is the longest |

**⚠️ Flashing safety (non-negotiable for a public game):**
- **At most 3 big flashes per second, ever.** The build gets more intense through **brightness, size and length**, not speed. (The 3-per-second limit is the standard rule for avoiding photosensitive seizures.)
- **Title-screen warning**, next to the content warning: *"This game contains flashing lights."*
- **Settings → Reduce flashing:** swaps the strobe for a slow white fade with the same timing and sound, so nobody misses a story beat.

**In code:** one `Transition` scene between day and night, driven by a small table of keyframes per night (time → flash brightness/size, which night frame to show). The night frames are ordinary exported sprites. Reachable with the agent backdoor: `-scene transition.night1` (2, 3).

**In the scripts:** `[FX: STROBE IN]` at the top of each night. The other direction, night → day, stays the **eyes close → hard cut to white** from the scripts.

## 📐 Resolution

| Thing | Size |
|---|---|
| Internal resolution | **480 × 270** (scaled ×4 to 1920×1080, crisp, no blur) |
| VN sprite (night scenes, the "high-res" ones) | **max 256 px tall**, ~**160 × 256** (knees-up, DDLC framing) |
| Desk portrait (patient at the window) | ~**64 × 64** px |
| Documents (Settler ID, sleep log) | ~**96 × 64** px, readable text with a pixel font |

**256 px is the cap** (Alejo, 2026-10-13). On a 270-px-tall screen a 256 sprite stands floor to almost ceiling, and the text box covers the legs, like DDLC. The face gets roughly 50–60 px, enough for anime eyes.

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
| Canvas | 1× size. VN sprite **160×256**, portrait **64×64**, document **96×64**, background **480×270** | |
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

### ⚠️ MCP gotchas (found while testing, 2026-10-13)

| Gotcha | Do this |
|---|---|
| `create_canvas` adds a layer called `Layer 1` | Rename it to `_ref` (use it for references) or `delete_layer` it |
| `create_if_missing: true` creates a missing **cel**, not a missing **layer**; it fails with "Layer not found" | `add_layer` first |
| A failed tool can still come back as success, with the error only in its text ("Failed to …") | Read every result message |
| MCP `export_spritesheet` **flattens all layers into one picture** | Don't use it for game sprites. Always `./assets/export.sh`. (It's fine for a quick look) |
| Only one session should write to a given `.aseprite` at a time | It's one binary file; two sessions writing it means one loses its work |

**ComfyUI (optional, later):** night backgrounds only, downscaled and forced onto the palette by a script, then **imported into an `.aseprite` file** (`import_image_as_layer`) and cleaned up there, so it goes through the same export. Not for characters.

## 🖼️ Art log

| Date | What | Verdict |
|---|---|---|
| 2026-10-13 | v0 test patient, 40×48, day + night ![[siesta-test-patient-v0.png]] | Red rule works. **Not anime, too small, and the day is now line art. Superseded.** |
| 2026-10-13 | Style locked: day = black outlines + white fill; night = Slay the Princess greyscale + red | From Alejo's StP reference board |
| 2026-10-13 | Day outlines **squiggle** (line boil, 3 frames ~8 fps) | Alejo: "to give the unsettling effect" |
| 2026-10-13 | **Tool connected: Aseprite 1.3.18.6** (built from source) + `aseprite` MCP (diivi/aseprite-mcp, 116 tools). Pipeline tested end to end on a 32×32 boil test head: draw → preview → `export.sh` → per-layer sheet + json | Works. Alejo: everything in Aseprite, exported as spritesheets |
