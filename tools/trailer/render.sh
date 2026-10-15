#!/bin/sh
# Renders the trailer: the game records every shot of the shot list (record
# mode), the compositor cuts them together with the cards and the motion
# graphics, mixes the sound, and ffmpeg encodes the mp4.
#
#   tools/trailer/render.sh [extra trailer render flags]
#
#   tools/trailer/render.sh                       # records everything, writes trailer_out/trailer.mp4
#   tools/trailer/render.sh -reuse                # keeps the shot folders that are already recorded
#   tools/trailer/render.sh -reuse -reduced -name trailer_reduced   # the backup with the -reduce-flashing strobe
#   tools/trailer/render.sh -reuse -only 142,146,796                # only these frames, as PNGs, to look at
#
# Output: trailer_out/<name>.mp4 (shots, wav and frames are under trailer_out/).
# Run it from anywhere; it needs a screen (the game opens a window while it
# records) and ffmpeg (brew install ffmpeg). More in the Trailer note.
set -eu
root=$(cd "$(dirname "$0")/../.." && pwd)

command -v ffmpeg >/dev/null || { echo "ffmpeg is missing: brew install ffmpeg" >&2; exit 1; }
mkdir -p "$root/trailer_out/bin"
(cd "$root/game" && go build -o "$root/trailer_out/bin/danmin" ./cmd/danmin)
(cd "$root/tools/trailer" && go build -o "$root/trailer_out/bin/trailer" .)
cd "$root"
trailer_out/bin/trailer render "$@"
