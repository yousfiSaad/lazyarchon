#!/bin/bash

# LazyArchon Quick Demo Recording - Simple as possible
# Just run this script and follow the prompts

set -e

# Setup paths
DEMO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CAST_FILE="${DEMO_DIR}/assets/demo/lazyarchon-demo.cast"
GIF_FILE="${DEMO_DIR}/assets/demo/lazyarchon-demo.gif"

# Check prerequisites
if ! command -v asciinema &> /dev/null; then
    echo "Installing asciinema..."
    brew install asciinema || { echo "Please install asciinema first: brew install asciinema"; exit 1; }
fi

if ! command -v agg &> /dev/null; then
    echo "Installing agg..."
    brew install agg || { echo "Please install agg first: brew install agg"; exit 1; }
fi

if ! command -v lazyarchon &> /dev/null; then
    echo "Building LazyArchon..."
    cd "$DEMO_DIR" && make build &> /dev/null
    export PATH="${DEMO_DIR}/bin:$PATH"
fi

# Clear and show instructions
clear
echo "LazyArchon Demo Recording"
echo "========================="
echo ""
echo "You'll now record a 30-60 second demo showing these features:"
echo ""
echo "  1. Launch lazyarchon                    (5 sec)"
echo "  2. Navigate with j/k and J/K            (10 sec)"
echo "  3. Search with / (type 'api')           (8 sec)"
echo "  4. Filter features with f               (10 sec)"
echo "  5. Change status with t                 (10 sec)"
echo "  6. Show help with ?                     (5 sec)"
echo "  7. Quit with q                          (2 sec)"
echo ""
echo "Tips:"
echo "  • Pause 1-2 seconds between actions"
echo "  • Keep terminal at 80x24 or larger"
echo "  • Total time: aim for 30-60 seconds"
echo ""
echo "Press ENTER to start recording (Ctrl+D to stop when done)..."
read

cd

# Start recording
echo "🎬 Recording started..."
asciinema rec --overwrite --idle-time-limit 2 "$CAST_FILE"

# Convert to GIF
echo ""
echo "Converting to GIF (this takes ~30 seconds)..."
agg --speed 1.5 --font-size 14 "$CAST_FILE" "$GIF_FILE"

# Done!
echo ""
echo "✅ Demo created successfully!"
echo ""
echo "Files created:"
echo "  • Recording: $CAST_FILE"
echo "  • GIF: $GIF_FILE"
echo ""
echo "Next step: Add to README.md:"
echo "  ![LazyArchon Demo](assets/demo/lazyarchon-demo.gif)"
echo ""
