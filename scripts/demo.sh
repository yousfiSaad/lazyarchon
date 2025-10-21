#!/bin/bash

# Ultra-simple demo recording - one command, no questions asked

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "Recording demo (press Ctrl+D when done)..."
asciinema rec --overwrite --idle-time-limit 2 "$DIR/assets/demo/lazyarchon-demo.cast"

echo "Converting to GIF..."
agg --speed 1.5 "$DIR/assets/demo/lazyarchon-demo.cast" "$DIR/assets/demo/lazyarchon-demo.gif"

echo "Done! GIF saved to: assets/demo/lazyarchon-demo.gif"