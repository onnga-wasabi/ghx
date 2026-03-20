#!/usr/bin/env bash
set -euo pipefail

# Record a short terminal session (asciinema cast) for README demo GIF.
# Requirements: `asciinema` and `ghx` command should work in your environment.
#
# Output: docs/demo.cast by default

DURATION_SECONDS="${DURATION_SECONDS:-25}"
OUT_FILE="${OUT_FILE:-docs/demo.cast}"
CMD="${CMD:-ghx}"

if ! command -v asciinema >/dev/null 2>&1; then
  echo "ERROR: asciinema is not installed." >&2
  echo "Install: (Nix) ensure `asciinema` is in ~/nix-config and rebuild" >&2
  echo "  or (non-Nix) brew install asciinema" >&2
  exit 1
fi

mkdir -p "$(dirname "$OUT_FILE")"

echo "Recording: command='$CMD' duration=${DURATION_SECONDS}s output='$OUT_FILE'"
asciinema rec -c "$CMD" -t "$DURATION_SECONDS" "$OUT_FILE"

echo "Done: $OUT_FILE"

