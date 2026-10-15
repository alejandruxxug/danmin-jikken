#!/bin/sh
# Renders the trailer: every shot of a shot list is recorded by the game
# (record mode), mixed, joined and encoded with ffmpeg.
#
#   tools/trailer/render.sh [shot list] [name]      # defaults: tools/trailer/shots.txt trailer
#
# Output: trailer_out/<name>.mp4 (frames, wavs and shot folders are under
# trailer_out/<name>/). Run it from anywhere; it needs a screen (the game opens
# a window) and ffmpeg (brew install ffmpeg). More in the Trailer note.
set -eu
root=$(cd "$(dirname "$0")/../.." && pwd)
shots=${1:-$root/tools/trailer/shots.txt}
name=${2:-trailer}

command -v ffmpeg >/dev/null || { echo "ffmpeg is missing: brew install ffmpeg" >&2; exit 1; }
mkdir -p "$root/trailer_out/bin"
(cd "$root/game" && go build -o "$root/trailer_out/bin/danmin" ./cmd/danmin)
(cd "$root/tools/trailer" && go build -o "$root/trailer_out/bin/trailer" .)
cd "$root"
trailer_out/bin/trailer render -shots "$shots" -name "$name"
