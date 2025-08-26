#!/bin/bash

# This script extracts all "action" values from zed-mac.json
# and writes them into a new JSON array in zed-valid-action.json.

INPUT_FILE="/Users/xinnjie/Dev/watchbeats/onekeymap/chore/zed-mac.json"
OUTPUT_FILE="/Users/xinnjie/Dev/watchbeats/onekeymap/chore/zed-valid-action.json"

if [ ! -f "$INPUT_FILE" ]; then
    echo "Error: Input file not found at $INPUT_FILE"
    exit 1
fi

# Use jq to extract all values for the key "action"
# The '..' recursively descends through the JSON structure.
# '.action?' safely gets the value of 'action' if it exists.
# 'select(. != null)' filters out any nulls for objects without an 'action' key.
# The surrounding '[]' collects all results into a single array.
# Remove comments from the JSON file before processing with jq
sed 's|//.*||g' "$INPUT_FILE" | jq '[.. | .bindings? | objects | .[] | if type == "string" then . else .[0] end] | map(select(. != null and . != "")) | unique | sort' > "$OUTPUT_FILE"

echo "Actions extracted from '$INPUT_FILE' and saved to '$OUTPUT_FILE'"
