#!/usr/bin/env bash
set -euo pipefail

# Convert an asciinema cast file into a GIF for README.
#
# Requirements: `agg` (asciinema/agg)
# - Install (Nix): add `asciinema-agg` to your `~/nix-config` and rebuild
# - Install (non-Nix): cargo install --git https://github.com/asciinema/agg
#
# Input: docs/demo.cast
# Output: docs/demo.gif

IN_FILE="${IN_FILE:-docs/demo.cast}"
OUT_FILE="${OUT_FILE:-docs/demo.gif}"

if ! command -v agg >/dev/null 2>&1; then
  echo "ERROR: agg is not installed." >&2
  echo "Install: (Nix) add 'asciinema-agg' to ~/nix-config, then rebuild" >&2
  echo "  or (non-Nix) cargo install --git https://github.com/asciinema/agg" >&2
  exit 1
fi

if [[ ! -f "$IN_FILE" ]]; then
  echo "ERROR: input cast not found: $IN_FILE" >&2
  exit 1
fi

mkdir -p "$(dirname "$OUT_FILE")"

echo "Rendering GIF: '$IN_FILE' -> '$OUT_FILE'"
agg "$IN_FILE" "$OUT_FILE"

echo "Done: $OUT_FILE"

