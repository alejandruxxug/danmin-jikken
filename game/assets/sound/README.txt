Sound files for the game (see docs/Danmin Jikken - Sound.md for the sources and licences). A cue with no file here is silent.
Every name is a cue: the game looks for name.ogg, name.mp3 or name.wav. All the choices below were confirmed by ear by Alejo (2026-10-14: "audio wise its good too").

MUSIC
title.mp3               Vampire's Piano, TAD, CC0, opengameart.org/content/vampires-piano
night_ambience.mp3      Paranoid Truth, Tsorthan Grove, CC0, opengameart.org/content/paranoid-truth
ending_rock_forever.mp3 Ironbound, Zander Noriega, CC BY 3.0 (CREDIT REQUIRED), opengameart.org/content/ironbound
                        https://creativecommons.org/licenses/by/3.0/  Cut: from MP3 frame 1914 (about 0:50.0) of the original to the end of the song. The crunch (1:24.7) is 34.7 s into the file;
                        the game starts the file earlier or later so the crunch lands on the ending title (rockDrops in cues.go).
ending_rock_sleep.mp3   Sad orchestral theme, Gichco, CC0, opengameart.org/node/29751. Cut: starts at 2:22.5 of the original, 80 s kept.
ending_rock_wake.mp3    Realm of Torment, Vitalezzz, CC0, opengameart.org/content/realm-of-torment. From 0:00, 80 s kept.
ending_rock_317.mp3     Unholy Surge, Vitalezzz, CC0, opengameart.org/content/unholy-surge. From 0:00, 80 s kept.
night3_tension.mp3 Narrow Corridors (narrow_corridors_short), TinyWorlds, CC0, opengameart.org/content/narrow-corridors. Night 3 music, DEFAULT PICK, not auditioned
driftwell_jingle.mp3    Hot Swing, Kevin MacLeod (incompetech.com), CC BY 4.0 (CREDIT REQUIRED)
                        https://creativecommons.org/licenses/by/4.0/  Cut: 1.0 s to 25.0 s of the original (the first second is silence).

The rock tracks were cut at MP3 frame boundaries with a small script (no re-encoding, so no quality loss). The measured slam
time of each is the "cut" column of the Sound note. The Sleep, Wake and 3:17 files start on the slam, so the game plays them from 0 (rockStarts in cues.go). Forever's starts well before it.

SOUND EFFECTS (short, 16-bit mono wav, volume-normalised and trimmed; every one CC0)
bell.wav          Pleasing Bell Sound Effect, Julie Damsgaard / Spring Spring (opengameart.org/content/pleasing-bell-sound-effect)
snore.wav         "snore", 80 CC0 creature SFX, rubberduck (opengameart.org/content/80-cc0-creature-sfx)
snore_2.wav       snore_02, 80 CC0 creature SFX 2, rubberduck (opengameart.org/content/80-cc0-creture-sfx-2)
snore_3.wav       Breathing tired, CC0 (opengameart.org/content/breathing-tired)
snore_4.wav       littlegrunts, crueltyinc, CC0 (opengameart.org/content/little-grunts)
                  All four snores (2026-10-14): 50 Hz high-pass, spectral noise gate, then
                  levelled to the same loudness (about -20 dBFS RMS of the loud part, soft
                  limit at -1 dBFS), mono 16-bit. Done outside the game with a Python script.
folder_open.wav   Paper Sound - 3, Various paper sound effects, Luckius, CC0 (opengameart.org/content/various-paper-sound-effects), mono, peak 0.6
folder_close.wav  Paper Sound - 1, same pack
folder_slide.wav  Paper Sound - 2, same pack
page_1.ogg        book_flip.8, 10 book page flips, StarNinjas, CC0 (opengameart.org/content/10-book-page-flips), the pack's own ogg
page_2.ogg        book_flip.4, same pack
page_3.ogg        book_flip.1, same pack
chime_right.mp3   Chime_Confirm, Chimey UI sounds, MouseBYTE, CC0 "no rights reserved" (opengameart.org/content/chimey-ui-sounds), the right-stamp ding
breath.wav        "breath", same pack
stamp_thunk.wav   thud3, Metal Impact Sounds, Brian MacIntosh (opengameart.org/content/metal-impact-sounds)
clock_chime.wav   bong1, same pack
stamp_lift.wav    wood_hit_01, 100 CC0 metal and wood SFX, rubberduck (opengameart.org/content/100-cc0-metal-and-wood-sfx)
stun_hit.wav      metal_slam_01, same pack
click.wav         lock_open_01, same pack
gurney.wav        wood_squeak_01 and _02 repeated three times (a squeaking wheel), same pack
heartbeat.wav     heartbeat_slow_0, Heartbeat sounds, estevinho (opengameart.org/content/heartbeat-sounds)
footsteps.wav     steps_platform and steps_platform2 alternated six times, yd (opengameart.org/content/platformer-sounds-terminal-interaction-door-shots-bang-and-footsteps)
drip.wav          one drop cut from "atmo", Dripping water loop, Independent.nu (opengameart.org/content/dripping-water-loop)
door_creak.wav    Creaky Door Hinge Spooky, w.forster.1 (opengameart.org/content/creaky-door-hinge-spooky)
pipe_snap.ogg     wood_breaking_01, 100 CC0 metal and wood SFX, rubberduck (opengameart.org/content/100-cc0-metal-and-wood-sfx, CC0 on the page), the ceiling pipe snapping in Forever (kept as the pack's own ogg: no converter here). DEFAULT PICK, not auditioned: other candidates in the same pack are wood_breaking_02, wood_cracking_01..04 and metal_hit_01..05
creak.wav         tree_creak, Tree Creaking, AntumDeluge (opengameart.org/content/tree-creaking), used as the floorboard

MADE FOR THE GAME (pure tones written with a small program, so no licence at all)
monitor_beep.wav  0.2 s sine at 1000 Hz
flatline.wav      4 s sine at 1000 Hz
stun_ring.wav     3 s of 6000 + 6600 + 3100 Hz sines, peak 0.22 (kept quiet on purpose)
day_hum.wav       3 s of 60 Hz and its harmonics, a whole number of cycles so it loops without a click

Each download was checked on its OpenGameArt page (licence and author) before it was used. The Ironbound zip carries no licence
file; its licence is the one on the page (CC BY 3.0). The OpenGameArt packs carry no licence file or only a CC0 note (yd's readme says cc0).

NOT FOUND / SILENT
All cues the scripts and the code ask for have a file. Pixabay was NOT used: its pages answer an automated download with a 403 challenge page,
which we do not get around. Nothing from Sonniss.
