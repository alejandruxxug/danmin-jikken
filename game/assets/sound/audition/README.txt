Rock audition tracks for the Forever ending (listening only).
The audio is NOT committed (see .gitignore) and NOT embedded in the game (assets.go embeds assets/sound/*.* only).
To rebuild the folder: download each source below, cut it at the "cut" time (an mp3 is cut at frame boundaries,
no re-encoding; the wav is cut at the sample) and keep 80 s. Each track also has a preview_<name> file of its first 30 s.
Run the game from game/:  go run ./cmd/danmin -rock list

name              track (author)                          source page                                          licence     cut (of the original)  file
sad_orchestral    Sad orchestral theme (Gichco)           https://opengameart.org/node/29751                   CC0         2:22.5                 mp3
ironbound         Ironbound (Zander Noriega)              https://opengameart.org/content/ironbound            CC-BY 3.0   1:24.7                 mp3
ending_scene      Ending Scene (nene), "Ending.wav"       https://opengameart.org/content/ending-scene         CC0         0:00                   wav (no mp3 of this version exists)
unholy_surge      Unholy Surge (Vitalezzz)                https://opengameart.org/content/unholy-surge         CC0         0:00                   mp3
realm_of_torment  Realm of Torment (Vitalezzz)            https://opengameart.org/content/realm-of-torment     CC0         0:00                   mp3
shadows_awaken    Shadows Awaken Within (Vitalezzz)       https://opengameart.org/content/shadows-awaken-within CC0        0:09                   mp3
aggressor         Aggressor (Kevin MacLeod)               https://incompetech.com/music/royalty-free/mp3-royaltyfree/Aggressor.mp3  CC-BY 4.0  0:01  mp3

Licences: the OpenGameArt pages show the licence in their sidebar (read 2026-10-14). The Ironbound download has no licence text inside
(the zip holds the mp3 and wav only), so the page is the record. Aggressor's mp3 tags and incompetech's page say
"Creative Commons: By Attribution 4.0". CC-BY tracks (ironbound, aggressor) need a credit line if one of them is chosen.
Not downloaded: "Credits (Slow Heavy Rock)" (Pixabay blocks automated downloads; not worked around), and the Sound note's
honourable mentions (Alt Rock Song, Tell us Green Man, Grunge Rock, Post-Rock Crescendo) have no measured cut.
