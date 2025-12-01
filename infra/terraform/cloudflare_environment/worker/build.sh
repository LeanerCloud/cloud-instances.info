#!/bin/bash
set -e

# Change to the worker directory
cd "$1"

# Install dependencies and build.
# Discard npm output because the output of the build script needs to just be a JSON object that contains the base64 encoded string.
npm install >/dev/null 2>&1
npm run build >/dev/null 2>&1

# Base64 encode the file, in a way portable across OSes (hopefully)
B64=$(cat dist/index.js | base64 -w 0 2>/dev/null || cat dist/index.js | base64 | tr -d '\n')

# Output only the JSON (nothing else should go to stdout)
echo "{\"content\": \"$B64\"}"