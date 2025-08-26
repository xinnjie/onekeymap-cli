#!/usr/bin/env bash
set -euo pipefail

# Extract IntelliJ action IDs from a keymap XML, sort and deduplicate.
# Usage:
#   intellij-extract-actions.sh <input-xml|-> [output-file]
#
# Examples:
#   intellij-extract-actions.sh /path/to/intellij-keymap.xml
#   intellij-extract-actions.sh /path/to/intellij-keymap.xml actions.json
#   cat intellij-keymap.xml | intellij-extract-actions.sh - > actions.json

usage() { echo "Usage: $0 <input-xml|-] [output-file]" >&2; }

if [ "$#" -lt 1 ] || [ "$#" -gt 2 ]; then
  usage
  exit 1
fi

INPUT="$1"
OUTPUT="${2:-}"

if [ "$INPUT" != "-" ] && [ ! -f "$INPUT" ]; then
  echo "Error: input file not found: $INPUT" >&2
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "Error: jq is required for JSON output; please install jq" >&2
  exit 1
fi

# Extraction pipeline: find <action id="..."> occurrences and print the id value.
# BSD sed compatible (-E for ERE). Then sort unique.
extract_ids() {
  if [ "$INPUT" = "-" ]; then
    sed -n -E 's@.*<action[^>]*id="([^"]+)".*@\1@p' | sort -u
  else
    sed -n -E 's@.*<action[^>]*id="([^"]+)".*@\1@p' "$INPUT" | sort -u
  fi
}

if [ -z "$OUTPUT" ]; then
  extract_ids | jq -R -s 'split("\n") | map(select(length>0))'
else
  extract_ids | jq -R -s 'split("\n") | map(select(length>0))' > "$OUTPUT"
  count=$(jq 'length' "$OUTPUT")
  echo "Wrote ${count} unique action IDs (JSON) to $OUTPUT"
fi
