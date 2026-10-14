# 🔊 Sound: music and effects shortlist

Research done 2026-10-13. Everything here is royalty-free with terms that allow it inside a game. **Update 2026-10-14: the builder downloaded default picks and wired them (see [[#🔌 What is in the game now]]).** Every file is a *default pick, Alejo to confirm by ear*; changing a pick means replacing one file in `game/assets/sound/`.

> **Read this first (honest limits).** I can't listen to audio. The "with the cut" timestamps below come from **measuring the loudness of the real files** (I downloaded them to a scratch folder and found where the volume jumps). The *sound* descriptions come from the track tags and the author's page, not from my ears. **Listen to the preview clips before you pick.** Where I couldn't download a file, I say so.

Related: [[Danmin Jikken – Game Design]] (the design log has the ending-rock decision), [[Danmin Jikken – Art Direction#⚡ Day to night strobe]] (the strobe sound), [[Danmin Jikken – Script]].

---

## 📏 The ending rule (your decisions, short)

| Step | What happens |
|---|---|
| 1 | The ending **hard-cuts** to black. Same for all four: Sleep, Wake, Forever, 3:17. |
| 2 | About **1.5 s of total silence**. Not even the room tone. |
| 3 | **Loud rock slams in, mid-riff, no fade-in** (the COD death-screen swell). |
| 4 | The **ending card** shows ("FOREVER ENDING · Halcyon Rest, Ward B · 3:17 AM"). |
| 5 | **Credits roll** under the same track. Track length target: about 2 to 3 minutes. |

Rock is used **only** here. It must sound **angry or grieving, never triumphant**. Forever (Mayu's suicide, hard cut to black at the drop) is the one that matters most.

**How the cut works in code:** the game starts the file at the slam timestamp (the `cut` column), so the player never hears the quiet part. If a track is longer than the credits, fade the volume over the last few seconds. All of this is the builder's job; you only choose.

---

## 🎸 Ending rock options

Three flavours, like you asked: **G** = gothic (Evanescence: strings or piano against heavy guitars), **S** = sludgy grunge (*Bleach*), **C** = quiet then crunch (Creep).

| # | Track | Flavour | Licence | Length | **Cut (start here)** | Left after cut | Best for | Your pick |
|---|---|---|---|---|---|---|---|---|
| 1 | **Sad orchestral theme**, Gichco | G | CC0 | 5:16 | **2:22.5** | ~2:53 | **Forever** | |
| 2 | **Ironbound**, Zander Noriega | C + S | CC-BY 3.0 | 3:25 | **1:24.7** | ~2:00 | **Forever** or Wake | |
| 3 | **Ending Scene**, nene | G | CC0 | 2:26 | **0:00** (or 0:22) | 2:26 | **Sleep** or Forever | |
| 4 | **Unholy Surge**, Vitalezzz | S | CC0 | 3:18 | **0:00** | 3:18 | **3:17** | |
| 5 | **Realm of Torment**, Vitalezzz | S | CC0 | 3:18 | **0:00** | 3:18 | Wake | |
| 6 | **Shadows Awaken Within**, Vitalezzz | S | CC0 | 3:43 | **0:09** | ~3:34 | Sleep | |
| 7 | **Aggressor**, Kevin MacLeod | S | CC-BY 4.0 | 3:59 | **0:01** | ~3:58 | 3:17 | |
| 8 | **Credits (Slow Heavy Rock)**, 9JackJack8 | S + C | Pixabay | 1:48 | **not measured** | n/a | Forever (if it builds) | |

Honorable mentions, not in the main eight: **Alt Rock Song** (CC-BY 3.0, 2:55, loud from 0:00, plain alt-rock, probably too cheerful), **Tell us Green Man** (CC0, 8:42, gothic doom metal, sad and somber tags but I can't tell if it has vocals, too long unless trimmed), **Grunge Rock** by SigmaMusicArt (Pixabay, 3:05, "aggressive, atmospheric", not measured), **Post-Rock Crescendo** by alex-morgan (Pixabay, 3:25, a slow build, labelled AI generated, not measured).

### Preview clips (so you can hear the cut)

Each clip is **1.5 s of silence, then 30 s of the track from the cut.** No ffmpeg on this Mac, so they are `.m4a` (AAC), not mp3. Double-click to play in QuickTime.

Folder: `/private/tmp/claude-501/-Users-urrexx-Documents-Uni-4th-Semester-Langs-Compilers-danmin-jikken/ba3a13c1-5fad-4946-a3da-d1079099094b/scratchpad/audio/previews/`

| Clip file | Option |
|---|---|
| `gothic_gichco_sad-orchestral-theme.m4a` | #1 |
| `creep_ironbound.m4a` | #2 |
| `gothic_nene_ending-scene.m4a` | #3 |
| `grunge_unholy-surge.m4a` | #4 |
| `grunge_realm-of-torment.m4a` | #5 |
| `grunge_shadows-awaken-within.m4a` (starts at 0:00, not 0:09) | #6 |
| `kevin_aggressor.m4a` | #7 |
| `gothic_green-man.m4a` | honorable mention |
| `altrock_professorlamp.m4a` | honorable mention |

That folder lives in a temporary scratch area, so copy the clips somewhere safe if you want to keep them.

### With the cut: what you'd hear

**#1 Sad orchestral theme (Gichco), start at 2:22.5, gothic.** Black screen, 1.5 s of silence, then the track lands at the moment its volume jumps about 13 dB: the full, loudest section of the song, which stays at that level to the end. The tags are post-rock, orchestral, sad, piano, violin, strings, **angry**, **depressing**, so expect strings and piano carrying a big, grieving wall of sound. The card shows over it. Only the last ~3 minutes are needed, so fade out around 3:00 into the clip (the real file keeps going until 5:08 and then fades by itself). I could not confirm there are electric guitars, so this is the closest to Evanescence *in mood*, maybe not in instruments. **Page:** https://opengameart.org/node/29751

**#2 Ironbound (Zander Noriega), start at 1:24.7, Creep crunch.** In the full song the first 25 seconds are a quiet, cold guitar verse, then it builds, and at **1:24.7** it jumps ~18 dB in a tenth of a second: that is the "crunch". You skip all of that. Black, silence, then the crunch hits mid-riff, a dark, cold, cinematic guitar-bass-drums instrumental stays loud through the card, and the final ~2 minutes carry the credits. It ends with a long fade-out, so it needs no extra fade. The author warns "use with caution" on the page, so listen before you commit. Best fit for Forever: it sounds like something giving way. **Page:** https://opengameart.org/content/ironbound

**#3 Ending Scene (nene), start at 0:00 (or 0:22), gothic.** The author describes it as "metal/rock guitar with orchestral music". It is loud from the first second: a heavy rhythmic hit pattern with the orchestra on top. At 0:22 there is a short gap and then a huge re-entry (about +36 dB). Starting at 0:00 gives you the slam straight away; starting at 0:22 gives a "second slam" after a beat of the track itself. The whole file is 2:26, which fits the credits exactly, and it fades out at the end. Version 1 is the one measured; there is also a brighter V2 and an orchestra-only version. **Page:** https://opengameart.org/content/ending-scene

**#4 Unholy Surge (Vitalezzz), start at 0:00, sludge.** Loud from the first beat (it starts at its average volume and gets denser until about 1:30). Tags: heavy rock, metal, dark. Black, silence, heavy riffing straight away, steady all the way through the credits, and a short fade-out at 3:18. Probably the angriest of the CC0 ones, so I'd keep it for 3:17. **Page:** https://opengameart.org/content/unholy-surge

**#5 Realm of Torment (Vitalezzz), start at 0:00, sludge.** Tags: heavy rock, metal, dark, electric guitar. Starts loud, with the main riff taking over at 0:08. There is a dip and a re-entry at 0:24 (+15 dB), another "second slam" if you want it. Ends with a ~10 second wind-down. **Page:** https://opengameart.org/content/realm-of-torment

**#6 Shadows Awaken Within (Vitalezzz), start at 0:09, sludge.** Quiet-ish for 8 seconds, then at about 0:09 it jumps to its full, loudest volume and stays there. Slightly slower and thicker than the other two. 3:43 is long, so fade it at 3:00. **Page:** https://opengameart.org/content/shadows-awaken-within

**#7 Aggressor (Kevin MacLeod), start at 0:01, sludge.** Album "Destruction Device". Loud from the very first second (about +19 dB right at the start), then a nearly constant wall of sound for 4 minutes. I only know it from the loudness and the album name. Needs a fade-out at ~3:00. **Page:** https://incompetech.com/music/royalty-free/index.html?isrc=USUAN1700051&Search=Search (the page loads its info with JavaScript, so open it in a browser)

**#8 Credits (Slow Heavy Rock) (9JackJack8, Pixabay), cut unknown.** I could not download it (Pixabay blocks automated downloads, I did not try to get around that). Tags: post-rock, heavy guitar, slow, melancholic, mournful, doom, **tension, release**, end credits, 1:48. It sounds like exactly the right idea (slow, heavy, grieving) and it was made for credits, but I can't give a slam timestamp. You'd listen on the page and tell me the second where the first big hit lands. **Page:** https://pixabay.com/music/hard-rock-credits-slow-heavy-rock-514006/

### My top three

| Rank | Pick | Why |
|---|---|---|
| 1 | **#2 Ironbound** | The only one with a proven, huge, clean "crunch" jump (18 dB in 0.1 s) at a known second, so the cut is a real event, not just "loud from the start". Sludgy and cold, which is closer to *Bleach* and Creep than a metal track. Free to use with one credit line. |
| 2 | **#1 Sad orchestral theme** | CC0, tagged "angry" and "depressing", and the build into the final loud section is the most grief-like shape on the list. Best match for the Evanescence side. Only risk: maybe no real guitars. |
| 3 | **#3 Ending Scene** | CC0, the right length for credits already, guitars *and* orchestra (the Evanescence mix on paper). Loud from the first second, so the slam is guaranteed. |

Suggested split if you want a different track per ending: **Forever = Ironbound**, **Sleep = Sad orchestral theme**, **Wake = Realm of Torment**, **3:17 = Unholy Surge**. Or one track for all four, which is also valid.

---

## 🎛️ Everything else, by cue

**Where the cues come from:** the `[SFX]` and `[MUSIC]` tags in the scripts ([[Danmin Jikken – Prologue]], [[Danmin Jikken – Night 1]], [[Danmin Jikken – Night 2]], [[Danmin Jikken – Night 3]]) plus what the design implies.

**How sure am I about these?** The Pixabay lists below come from Pixabay's search pages (title, artist, length are right). I checked the licence **summary** page once, not every sound's own page. Many Pixabay sounds are by the uploader "freesound_community" (re-uploads of Freesound sounds), and their original Freesound licence is not shown on Pixabay. I could not preview any of them.

**Pixabay licence in one line:** free for any use, no credit needed, you can edit the files, but you can't sell or hand out the files *by themselves*. See the licence section for the repo question.

### 🎼 Music

| Cue | Where in the game | Option A | Option B | Option C | Your pick |
|---|---|---|---|---|---|
| **Driftwell jingle** (cheerful 1950s corporate, last note held a beat too long) | Prologue reel; Night 3 line 176 | **Hot Swing**, Kevin MacLeod, CC-BY 4.0, 0:50, swing. Cut a few bars and stretch the last note ourselves. https://incompetech.com/music/royalty-free/ (search the title) | **Cheery Monday**, Kevin MacLeod, CC-BY 4.0, 1:20, bouncy and bright. Same site. | **Commercial Jingle 04**, alanajordan, Pixabay, 2:26. Not checked for vocals or AI. https://pixabay.com/music/beats-commercial-jingle-04-252796/ | |
| **Title screen** (not rock) | Title with the dagger | **Vampire's Piano**, TAD, CC0, loopable sad piano, WAV 15.8 MB / MP3 2.9 MB. https://opengameart.org/content/vampires-piano | **Sadness**, Kistol, CC0, seamless loop piano. https://opengameart.org/content/sadness | **Slow dark gothic piano**, Dark__, Pixabay, 2:46. https://pixabay.com/music/solo-piano-slow-dark-gothic-piano-599906/ | |
| **Night ambience** | All three nights (`[MUSIC: off]` cuts it) | **Paranoid Truth**, Tsorthan Grove, CC0, eerie drone, mounting tension (length not on the page). https://opengameart.org/content/paranoid-truth | **Abandoned passages**, congusbongus, CC0, seamless loop, 409 KB, cramped dark corridor feel. https://opengameart.org/content/abandoned-passages-horror-ambience-loop | **Ambience Pack 1, Sci Fi Horror**, Joth, CC0, five 1-minute loops. https://opengameart.org/content/ambience-pack-1-sci-fi-horror | |

Note on Paranoid Truth: it comes as a `.7z` archive, so unpack it in a throwaway folder first.

### 🌞 Day shift

| Cue | Option A | Option B | Option C | Your pick |
|---|---|---|---|---|
| **Day desk ambience / loop** (fluorescent hum under the desk) | **Fluorescent Light Hum**, freesound_community, Pixabay, 0:19. https://pixabay.com/sound-effects/household-fluorescent-light-hum-31094/ | **Factory Fluorescent light buzz**, Pixabay, 2:49 (long, loop-friendly). https://pixabay.com/sound-effects/film-special-effects-factory-fluorescent-light-buzz-6871/ | **Hospital ambience, Kwun Tong**, Pixabay, 2:02 (a real ward, quiet voices). https://pixabay.com/sound-effects/city-012174-kwun-tong-hospital-ambience9wav-58596/ | |
| **Reception bell** | **Service/Receptionist bell**, OxidVideos, Pixabay, 0:02. https://pixabay.com/sound-effects/film-special-effects-servicereceptionist-bell-418758/ | **Reception Bell**, freesound_community, Pixabay, 0:07. https://pixabay.com/sound-effects/film-special-effects-reception-bell-14620/ | **Bell Ring**, Universfield, Pixabay, 0:02. https://pixabay.com/sound-effects/film-special-effects-bell-ring-123742/ | |
| **Snores** (sleeping patients) | **A person snoring**, Mollyroselee, Pixabay, 0:08. https://pixabay.com/sound-effects/people-a-person-snoring-468533/ | **snoring single**, u_23l344d16t, Pixabay, 0:14. https://pixabay.com/sound-effects/people-snoring-single-343603/ | **Male snoring**, AudioPapkin, Pixabay, 0:27. https://pixabay.com/sound-effects/people-male-snoring-297875/ | |

### ⚡ Day to night strobe (stun = hit + ringing whine + heartbeat)

| Part | Option A | Option B | Option C | Your pick |
|---|---|---|---|---|
| **Sharp hit** | **Horror Impact Hit**, Universfield, Pixabay, 0:04. https://pixabay.com/sound-effects/horror-horror-impact-hit-567238/ | **Cinematic Impact Hit**, Universfield, 0:03. https://pixabay.com/sound-effects/film-special-effects-cinematic-impact-hit-352702/ | **Horror Hit**, Universfield, 0:02. https://pixabay.com/sound-effects/horror-horror-hit-443980/ | |
| **Ringing whine (tinnitus)** | **High pitched ringing**, freesound_community, 0:06. https://pixabay.com/sound-effects/film-special-effects-high-pitched-ringing-88473/ | **Tinnitus**, freesound_community, 0:10. https://pixabay.com/sound-effects/film-special-effects-tinnitus-7048/ | **High pitched tone [quiet to loud]**, 0:11 (a built-in fade-in). https://pixabay.com/sound-effects/film-special-effects-high-pitched-tone-quiet-to-loud-14549/ | |
| **Heartbeat** | **Thudding Heartbeat**, DRAGON-STUDIO, 0:08. https://pixabay.com/sound-effects/film-special-effects-thudding-heartbeat-372487/ | **Heartbeat 02**, BRVHRTZ, 0:08. https://pixabay.com/sound-effects/film-special-effects-heartbeat-02-225103/ | **Intense Cinematic Heartbeat**, LensPulse, 0:06. https://pixabay.com/sound-effects/film-special-effects-intense-cinematic-heartbeat-sound-effect-584627/ | |

The "Reduce flashing" setting uses the same stun sound, so keep the ringing at a safe volume. The game mixes these; we only need the three source files.

### 🌙 Night SFX

| Cue (script) | Option A | Option B | Option C | Your pick |
|---|---|---|---|---|
| **Monitor beep**, beep… beep… (N1, N3) | **Heart Monitor Beep**, Universfield, 0:01 (a single beep, loop it). https://pixabay.com/sound-effects/film-special-effects-heart-monitor-beep-151929/ | **Heart Monitor Beep**, Universfield, 0:08. https://pixabay.com/sound-effects/technology-heart-monitor-beep-153306/ | **heart monitor beep**, freesound_community, 0:09. https://pixabay.com/sound-effects/film-special-effects-heart-monitor-beep-96554/ | |
| **Flatline**, "the beep becomes one long tone" (N1) | **The terror of flatlining**, freesound_community, 0:55. https://pixabay.com/sound-effects/technology-the-terror-of-flatlining-34031/ | **Cardiac Arrest**, Liecio, 0:27. https://pixabay.com/sound-effects/film-special-effects-cardiac-arrest-189742/ | **Heartbeat with beep**, freesound_community, 0:11. https://pixabay.com/sound-effects/film-special-effects-heartbeat-with-beep-100833/ | |
| **Footsteps** (N1, N2, N3 "many, fast") | **Footsteps in a hallway**, freesound_community, 0:10. https://pixabay.com/sound-effects/household-footsteps-in-a-hallway-47842/ | **Creepy Distant Footsteps**, DRAGON-STUDIO, 0:19. https://pixabay.com/sound-effects/film-special-effects-creepy-distant-footsteps-482883/ | **Echoing footsteps**, PWLPL, 0:05. https://pixabay.com/sound-effects/film-special-effects-echoing-footsteps-377259/ | |
| **Squeaking gurney wheels** (Prologue, N3) | **Metal_Cart_Stretcher_Rolling**, u_ivtxcu8cix, 0:38. https://pixabay.com/sound-effects/film-special-effects-metal-cart-stretcher-rolling-285242/ | **Person Pushing Cart down hallway**, freesound_community, 0:18. https://pixabay.com/sound-effects/film-special-effects-person-pushing-cart-down-hallway-29620/ | **rolling cart 002**, freesound_community, 0:06. https://pixabay.com/sound-effects/film-special-effects-rolling-cart-002-86702/ | |
| **Slow drip**, then another (N3) | **Water Drip**, freesound_community, 0:01 (one drip, we time the gaps). https://pixabay.com/sound-effects/film-special-effects-water-drip-45622/ | **Water Drip 7**, 0:03. https://pixabay.com/sound-effects/film-special-effects-water-drip-7-39622/ | **Water Dripping**, DRAGON-STUDIO, 0:22. https://pixabay.com/sound-effects/nature-water-dripping-364450/ | |
| **Door creak** (N3) | **Creaky Old Door**, DRAGON-STUDIO, 0:01. https://pixabay.com/sound-effects/household-creaky-old-door-472357/ | **Creaky Door Hinge Spooky**, uploader w.forster.1, **CC0**, "Recorded by me". https://opengameart.org/content/creaky-door-hinge-spooky | **Creepy Creaking Door**, Universfield, 0:03. https://pixabay.com/sound-effects/horror-creepy-creaking-door-402155/ | |
| **A floorboard** (N2) | **Heavy Creaking**, DRAGON-STUDIO, 0:01. https://pixabay.com/sound-effects/film-special-effects-heavy-creaking-515252/ | **Creaking Door 03**, Universfield, 0:01 (short, could pass). https://pixabay.com/sound-effects/household-creaking-door-03-487855/ | n/a: I did not find a clean "floorboard" sound | |
| **Click** ("a small click", N1, N2) | **Light Switch Flip**, Homemade_SFX, Pixabay. https://pixabay.com/sound-effects/household-light-switch-flip-272436/ | **Switch**, sumaga123, 0:01. https://pixabay.com/sound-effects/film-special-effects-switch-440413/ | **Switch Click and Beep 001a**, EdR. https://pixabay.com/sound-effects/household-switch-click-and-beep-001a-11602/ | |
| **Clock chime** (N3, "somewhere, chimes") | **Clock Chime**, freesound_community, 0:04. https://pixabay.com/sound-effects/film-special-effects-clock-chime-88027/ | **Old Clock Chimes**, 0:22. https://pixabay.com/sound-effects/household-old-clock-chimes-74624/ | **Grandfather Clock, Half Hour Chimes**, designerschoice, 0:13. https://pixabay.com/sound-effects/city-clockchim-cu-grandfather-clock-half-hour-chimes-nicholas-judy-tdc-493266/ | |

**Not found:** the line "[SFX: I can breathe again]" in Night 2 is a breath, and I did not search for it. A long breath-in from the snore set or a Freesound CC0 breath would do. I did not search Freesound itself for CC0 monitor sounds; its search is a login-gated site and the web search tool returned nothing from it. Say the word and I'll look at it in the browser.

---

## 📜 Licence summary and credits

| Source | Licence | Credit needed? | Verified on the page? |
|---|---|---|---|
| **OpenGameArt, CC0 items** (Gichco, nene, Vitalezzz, Tsorthan Grove, congusbongus, Joth, TAD, Kistol, the creaky door) | CC0 1.0, public domain | **No.** Kind to credit, not required | Yes, each track page opened |
| **Ironbound**, Zander Noriega | CC-BY 3.0 | **Yes** | Yes, page opened |
| **Kevin MacLeod / Incompetech** | CC-BY 4.0 | **Yes** | Yes, the FAQ page (not each track page) |
| **Pixabay music and sound effects** | Pixabay Content Licence | **No** | Licence summary page and the music pages I opened; the SFX pages not individually |
| **Sonniss GDC bundles** (not used, see below) | Royalty-free, no credit | No | Yes, but see the repo caveat |

### Credit lines for the credits screen

Only these two need to appear. Everything else is optional thanks.

```
Ironbound by Zander Noriega (CC BY 3.0, creativecommons.org/licenses/by/3.0/)
Hot Swing by Kevin MacLeod (incompetech.com)
Licensed under Creative Commons: By Attribution 4.0
https://creativecommons.org/licenses/by/4.0/
```

Swap the track name in the Kevin line for each Kevin track you end up using (Hot Swing, Cheery Monday, Aggressor). Ironbound's page also asks you to link the name to the author's SoundCloud or Twitter where possible, so put a link there in the README. Optional courtesy lines (not required): "Gichco", "nene", "Vitalezzz", "Kistol", "TAD", "Pixabay and its artists".

### Caveats you should know before the defense

1. **Pixabay and the public repo.** The Pixabay licence bans giving out the files "on a standalone basis". Embedding a sound in a game is a creative work, but the raw `.mp3` also sits in a public git repo. The licence page does not say whether that counts. **Safest: prefer the CC0 and CC-BY items; use Pixabay only for the SFX that have no CC0 option** (bell, ringing, heartbeat, snores). If you want zero risk, we make those few sounds ourselves.
2. **Sonniss bundles are not recommended.** Their licence forbids redistributing the raw sound files, which a public repo does. They are also 23.5 GB.
3. **AI-generated tracks exist on Pixabay.** "Post-Rock Crescendo", "Quiet and Dim" and "Jingle Pop!" carry an AI label. I did not pick them as main options.
4. **Check the licence inside downloads.** One OpenGameArt pack I came across said CC0 on the page but "no reselling" inside the zip. For the picks, we read the licence file in the download before using it.
5. **Not verified:** exact guitar/vocal content of every track (no ears); Pixabay downloads (blocked by a challenge page, not bypassed); the per-sound licences of "freesound_community" re-uploads; Freesound itself.

---

## 🔌 What is in the game now

Built 2026-10-14 by the builder. I can't listen, so **every row below is a default pick, Alejo to confirm by ear.** Files, authors and licences are also in `game/assets/sound/README.txt`. Each download was opened on its OpenGameArt page (licence and author checked) before use; Pixabay was not used because its pages answer an automated download with a 403 challenge (not bypassed) and Sonniss was not used.

### Music

| Cue | When | File | Source and licence | Pick |
|---|---|---|---|---|
| `title` | Title (loops) | `title.mp3` | Vampire's Piano, TAD, CC0 | kept |
| `night_ambience` | All nights (loops) | `night_ambience.mp3` | Paranoid Truth, Tsorthan Grove, CC0 | kept |
| `day_hum` | The day desk (loops, stops under the strobe) | `day_hum.wav` | a 60 Hz hum with harmonics written for the game, 3 s, loops without a click | **default, confirm** |
| `ending_rock_forever` | Forever ending | `ending_rock_forever.mp3` | Ironbound, Zander Noriega, **CC BY 3.0 (credit)**. Cut at 1:24.7, 80 s kept | **default, confirm** |
| `ending_rock_sleep` | Sleep ending | `ending_rock_sleep.mp3` | Sad orchestral theme, Gichco, CC0. Cut at 2:22.5, 80 s kept | **default, confirm** |
| `ending_rock_wake` | Wake ending | `ending_rock_wake.mp3` | Realm of Torment, Vitalezzz, CC0. From 0:00, 80 s kept | **default, confirm** |
| `ending_rock_317` | 3:17 ending | `ending_rock_317.mp3` | Unholy Surge, Vitalezzz, CC0. From 0:00, 80 s kept | **default, confirm** |
| `driftwell_jingle` (`music jingle` in the Sleep ending, `music driftwell_jingle` in the Prologue) | Plays once | `driftwell_jingle.mp3` | Hot Swing, Kevin MacLeod, **CC BY 4.0 (credit)**. 1.0 s to 25.0 s of the original. The "last note held too long" is NOT made: the track just ends. | **default, confirm** |
| `night3_tension` (night 3: starts when the eyes open on Ward C, off in Hana's room, back at the gurney, fades out over 5 s from "Tonight we make your rest permanent" so it is silent when the ending starts; the black opening has no music) | `night3_tension.mp3` (3 MB, about 3 min, looped) | Narrow Corridors, TinyWorlds, CC0 (opengameart.org/content/narrow-corridors), "a suspension building piece". Not heard by the builder. Other CC0 candidates: Deep Humidity (opengameart.org/content/deep-humidity), Nighttime Solitude | **default, confirm by ear** |

The rock files are **cut at the slam** (frame-level mp3 cut, no re-encoding), so the game plays them from 0 (`rockStarts` in `cues.go` is all 0 and says why). Reasons: no decoding of minutes of music nobody hears, and the repo stays about 13 MB smaller. 80 s is enough: card 6 s (10 s for 3:17) + credits roll about 50 s; a test fails if the roll ever outgrows the track.

### The ending, as built

| Step | Code |
|---|---|
| Hard cut, 1.5 s silence | `phaseSilence`: `SetMusic("")`; the rock is **decoded now** (`Player.Preload`) so the slam is not late |
| Slam, no fade-in | `phaseCard`: `PlayMusic("ending_rock_<name>", false, rockStarts[name])` |
| Credits under the same track | the track just keeps playing |
| Fade in the last seconds | when 4 s of credits are left, `Player.FadeOut(240)` (a tested linear ramp, `FadeGain`); skipping the credits hard-cuts at the top-10 |

### Sound effects (all short, 16-bit mono wav)

| Cue (script or code) | File | Source and licence | Pick |
|---|---|---|---|
| reception bell (`bell`, day) | `bell.wav` | Pleasing Bell Sound Effect, Julie Damsgaard / Spring Spring, CC0 | **default, confirm** |
| `snore`, `snore_2`, `snore_3`, `snore_4` (a sleeper snores: random variant, never the same twice; gain 0.2 to 0.3 in a bed, 0.8 to 1.0 in the hallway; every 3 to 7 s, shorter with more sleepers; none with 0 sleepers; `game/snore.go`) | `snore.wav` (0.44 s), `snore_2.wav` (0.93 s), `snore_3.wav` (3.2 s), `snore_4.wav` (6.3 s) | `snore`: 80 CC0 creature SFX, rubberduck, CC0. `snore_2`: `snore_02` from 80 CC0 creature SFX 2, rubberduck, CC0. `snore_3`: Breathing tired, CC0 (opengameart.org/content/breathing-tired). `snore_4`: littlegrunts, crueltyinc, CC0 ("noises in my sleep", opengameart.org/content/little-grunts). 2, 3 and 4 are mono 16-bit, peak normalised to 0.5. Real long snores are rare on OpenGameArt, so these are short snorts, a breathing and sleep grunts | **default, weak, confirm by ear** |
| `chime_right` (the +100 floats up after a right stamp; gain 0.7; a wrong stamp, the +50s and QUICK are silent) | `chime_right.mp3` (1.4 s) | Chime_Confirm from "Chimey UI sounds", MouseBYTE, CC0 (opengameart.org/content/chimey-ui-sounds). Not heard by the builder: may not be a two-note ding | **default, confirm by ear** |
| `folder_open` (the cover swings open), `folder_close` (it swings shut after the stamp), `folder_slide` (the folder slides in, and again when it slides out) | `folder_open.wav` 0.7 s, `folder_close.wav` 0.49 s, `folder_slide.wav` 0.55 s | "Paper Sound" 3, 1 and 2 from Various paper sound effects, Luckius, CC0 (opengameart.org/content/various-paper-sound-effects). Mono, peak 0.6. Paper, not cardboard; the builder cannot hear which fits which | **default, weak, confirm by ear** |
| `page_1`, `page_2`, `page_3` (a page corner is lifted at gain 0.5, a turn finishes at 0.7, a lifted page springs back flat at 0.25; random, never the same twice; `game/foley.go`) | `page_1.ogg` 0.60 s, `page_2.ogg` 0.65 s, `page_3.ogg` 0.75 s | `book_flip.8`, `.4`, `.1` from "10 book page flips", StarNinjas, CC0 (opengameart.org/content/10-book-page-flips). The pack's own ogg files | **default, confirm by ear** |
| `stamp_thunk`, also the SFX volume test in Settings | `stamp_thunk.wav` | thud3, Metal Impact Sounds, Brian MacIntosh, CC0 | **default, confirm** |
| `stamp_lift` (gain 0.8) | `stamp_lift.wav` | wood_hit_01, 100 CC0 metal and wood SFX, rubberduck, CC0 | **default, confirm** |
| `pipe_snap` (Forever: the ceiling pipe snaps on the first frame of the break, then the splash) | `pipe_snap.ogg` | wood_breaking_01, 100 CC0 metal and wood SFX, rubberduck, CC0 (opengameart.org/content/100-cc0-metal-and-wood-sfx). Not auditioned; alternatives in the same pack: wood_breaking_02, wood_cracking_01..04, metal_hit_01..05 | **default, confirm** |
| stun part 1, `stun_hit` | `stun_hit.wav` | metal_slam_01, same pack | **default, confirm** |
| stun part 2, `stun_ring` (gain 0.6) | `stun_ring.wav` | 3 s of three high sines, peak 0.22, written for the game | **default, confirm** |
| stun part 3, `heartbeat` | `heartbeat.wav` | heartbeat_slow_0, Heartbeat sounds, estevinho, CC0 | **default, confirm** |
| `monitor_beep` (aliases `beep`, `beep_slow`, `monitor_far`) | `monitor_beep.wav` | a 1000 Hz sine, 0.2 s, written for the game | **default, confirm** |
| `flatline` (alias `long_tone`) | `flatline.wav` | a 1000 Hz sine, 4 s, written for the game | **default, confirm** |
| `footsteps` (aliases `footsteps_many`, `footsteps_stop`) | `footsteps.wav` | steps_platform and steps_platform2 from yd's pack, six steps, CC0 | **default, confirm** |
| `gurney` (aliases `wheels_slow`, `wheels_squeak`, `wheels_stop`) | `gurney.wav` | the two wood_squeak files of the 100 CC0 pack repeated three times, CC0. Not a real wheel | **default, weak, confirm** |
| `drip` (alias `drip_fast`) | `drip.wav` | one drop cut from Dripping water loop, Independent.nu, CC0 | **default, confirm** |
| `door_creak` (alias `door_close_soft`) | `door_creak.wav` | Creaky Door Hinge Spooky, w.forster.1, CC0 | **default, confirm** |
| `creak` (alias `floorboard`) | `creak.wav` | tree_creak, Tree Creaking, AntumDeluge, CC0. A tree, not a floorboard | **default, weak, confirm** |
| `click` (alias `rail_clang`) | `click.wav` | lock_open_01, 100 CC0 pack. A lock, not a bed rail | **default, weak, confirm** |
| `clock_chime` | `clock_chime.wav` | bong1, Metal Impact Sounds, CC0 | **default, confirm** |
| `breath` (alias `breathe_again`) | `breath.wav` | breath, 80 CC0 creature SFX, CC0. Short (0.7 s) | **default, weak, confirm** |

All cues in the scripts and in the Go code have a file (two tests check it, every night script included). Nothing is silent. **Not done:** no real "fluorescent hum" or "hospital ambience" recording (the hum is a synthesized buzz), no real wheeled-gurney, floorboard or bed-rail sound. Those four are the likeliest to sound wrong; they are the first to swap if a better CC0 file turns up.

### Rules the code follows

- Music and SFX go through the Settings volumes; SFX also have a per-cue gain (`sfxGain` in `cues.go`): the ringing is held down; the snores choose their own gain by where the sleeper is.
- `-reduce-flashing` changes only the picture; the stun sound is the same and is quiet by design.
- In `-agent` test mode the player is muted and plays nothing; every request is still in the dump (`soundLog`) and `soundFiles` shows the file each cue resolves to.

---

## ✅ Your picks

Fill in the "Your pick" columns above, or just write it here.

*(The builder filled these with defaults on 2026-10-14: Forever Ironbound, Sleep Sad orchestral theme, Wake Realm of Torment, 3:17 Unholy Surge, jingle Hot Swing. Tick or change after listening.)*

- [ ] Forever ending track:
- [ ] Sleep ending track:
- [ ] Wake ending track:
- [ ] 3:17 ending track:
- [ ] Title music:
- [ ] Night ambience:
- [ ] Driftwell jingle:
- [ ] Stun sound (hit / ringing / heartbeat):
- [ ] Day desk ambience:
- [ ] SFX set (bell, snore, beep, flatline, footsteps, gurney, drip, creak, click, chime):
