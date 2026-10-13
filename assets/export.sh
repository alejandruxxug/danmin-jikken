#!/usr/bin/env bash
# Exports every assets/src/<name>.aseprite to a spritesheet the game embeds:
#   game/assets/sprites/<name>.png   all layers x all frames, packed
#   game/assets/sprites/<name>.json  Aseprite JSON: one entry per "<layer>#<frame>", plus tags and layers
#
# Layers whose name starts with "_" (references, guides) are left out.
# Hidden layers ARE exported (expression variants like eyes_sleepy are usually hidden while drawing).
# Run from anywhere: ./assets/export.sh   (or ./assets/export.sh mina   to export just assets/src/mina.aseprite)
set -euo pipefail

ASEPRITE="${ASEPRITE_PATH:-/Applications/Aseprite.app/Contents/MacOS/aseprite}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SRC="$ROOT/assets/src"
OUT="$ROOT/game/assets/sprites" # inside game/ because go:embed can't reach ../
mkdir -p "$OUT"

if [ $# -gt 0 ]; then
  files=()
  for n in "$@"; do files+=("$SRC/$n.aseprite"); done
else
  files=("$SRC"/*.aseprite)
fi

for f in "${files[@]}"; do
  [ -e "$f" ] || { echo "no .aseprite files in $SRC"; exit 1; }
  name="$(basename "$f" .aseprite)"

  ignore=()
  while IFS= read -r layer; do
    case "$layer" in _*) ignore+=(--ignore-layer "$layer") ;; esac
  done < <("$ASEPRITE" -b --all-layers --list-layers "$f")

  # Options must come BEFORE the file: Aseprite applies them to the files that follow.
  "$ASEPRITE" -b \
    --all-layers ${ignore[@]+"${ignore[@]}"} \
    --split-layers --list-layers --list-tags \
    --sheet-pack --merge-duplicates \
    --filename-format '{layer}#{frame}' --format json-array \
    "$f" \
    --sheet "$OUT/$name.png" --data "$OUT/$name.json" >/dev/null

  echo "exported $name -> game/assets/sprites/$name.{png,json}"
done
