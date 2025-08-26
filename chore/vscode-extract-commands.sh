#!/bin/bash

# This script extracts, filters, and sorts command values from a VS Code keybindings file.

SOURCE_FILE="/Users/xinnjie/Dev/watchbeats/onekeymap/chore/vscode-mac-default.json"
DEST_FILE="/Users/xinnjie/Dev/watchbeats/onekeymap/chore/vscode-valid.json"

# Check if source file exists
if [ ! -f "$SOURCE_FILE" ]; then
    echo "Error: Source file not found at $SOURCE_FILE"
    exit 1
fi

# Use jq to:
# 1. Extract all 'command' values.
# 2. Filter out nulls and commands starting with '-'.
# 3. Get unique values and sort them.
# 4. Write the resulting JSON array to the destination file.
jq '[.[] | .command | select(. != null and (startswith("-") | not))] | unique | sort' "$SOURCE_FILE" > "$DEST_FILE"

echo "Processed commands from '$SOURCE_FILE' and saved to '$DEST_FILE'"
