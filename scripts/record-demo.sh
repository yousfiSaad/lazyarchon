#!/bin/bash

# LazyArchon Interactive Demo Recording Script
# This script helps you record a professional demo and convert it to GIF

set -e

DEMO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CAST_FILE="${DEMO_DIR}/assets/demo/lazyarchon-demo.cast"
GIF_FILE="${DEMO_DIR}/assets/demo/lazyarchon-demo.gif"
THUMB_FILE="${DEMO_DIR}/assets/demo/lazyarchon-demo-thumb.png"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}================================${NC}"
echo -e "${BLUE}LazyArchon Demo Recording${NC}"
echo -e "${BLUE}================================${NC}"
echo ""

# Check prerequisites
echo -e "${YELLOW}Checking prerequisites...${NC}"

if ! command -v asciinema &> /dev/null; then
    echo -e "${RED}❌ asciinema not found. Install with: brew install asciinema${NC}"
    exit 1
fi

if ! command -v agg &> /dev/null; then
    echo -e "${RED}❌ agg not found. Install with: brew install agg${NC}"
    exit 1
fi

if ! command -v lazyarchon &> /dev/null; then
    echo -e "${YELLOW}⚠️  LazyArchon not in PATH. Building...${NC}"
    cd "$DEMO_DIR"
    make build
    export PATH="${DEMO_DIR}/bin:$PATH"
fi

echo -e "${GREEN}✅ All prerequisites met${NC}"
echo ""

# Function to record demo
record_demo() {
    echo ""
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}Recording Demo${NC}"
    echo -e "${BLUE}================================${NC}"
    echo ""
    echo -e "${YELLOW}Tips:${NC}"
    echo "- Use a 24x80 terminal or similar"
    echo "- Increase font size (14pt+) for visibility"
    echo "- Pause 1-2 seconds between actions"
    echo "- Total target: 30-60 seconds"
    echo ""
    echo -e "${YELLOW}Key Demo Features to Show:${NC}"
    echo "1. App launch and connection"
    echo "2. Navigation (j/k, J/K, h/l)"
    echo "3. Search (/) and highlighting"
    echo "4. Feature filtering (f)"
    echo "5. Status change (t)"
    echo "6. Project switching (p)"
    echo "7. Help system (?)"
    echo ""
    echo -e "${YELLOW}Press Enter to start recording...${NC}"
    read

    echo -e "${GREEN}🎬 Recording started...${NC}"
    echo "Press Ctrl+D when done."
    echo ""

    asciinema rec \
        --overwrite \
        --title "LazyArchon - Terminal UI for Archon" \
        --idle-time-limit 2 \
        "$CAST_FILE"

    echo ""
    echo -e "${GREEN}✅ Recording saved to: $CAST_FILE${NC}"
    echo ""
    read -p "Convert to GIF now? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        convert_to_gif
    fi
}

# Function to convert to GIF
convert_to_gif() {
    if [ ! -f "$CAST_FILE" ]; then
        echo -e "${RED}❌ No recording found at: $CAST_FILE${NC}"
        echo "Please record a demo first."
        exit 1
    fi

    echo ""
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}Converting to GIF${NC}"
    echo -e "${BLUE}================================${NC}"
    echo ""
    echo -e "${YELLOW}Converting $CAST_FILE to GIF...${NC}"
    echo "This may take a minute..."
    echo ""

    # Convert with good compression settings
    agg \
        --speed 1.5 \
        --font-size 14 \
        --padding 10 \
        "$CAST_FILE" \
        "$GIF_FILE"

    if [ -f "$GIF_FILE" ]; then
        SIZE=$(du -h "$GIF_FILE" | cut -f1)
        echo -e "${GREEN}✅ GIF created: $GIF_FILE (${SIZE})${NC}"
        echo ""
        echo -e "${YELLOW}Next steps:${NC}"
        echo "1. Review the GIF: open $GIF_FILE"
        echo "2. If happy, it's ready for README.md"
        echo "3. Add to README: ![LazyArchon Demo](assets/demo/lazyarchon-demo.gif)"
        echo "4. Commit and push!"
    else
        echo -e "${RED}❌ Failed to create GIF${NC}"
        exit 1
    fi
}

# Function to show recording tips
show_tips() {
    echo ""
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}Demo Recording Tips${NC}"
    echo -e "${BLUE}================================${NC}"
    echo ""
    echo -e "${YELLOW}Terminal Setup:${NC}"
    echo "- Use macOS Terminal, iTerm2, or similar"
    echo "- Font size: 14-16pt for clarity"
    echo "- Terminal size: 24x80 or 24x120"
    echo "- Dark theme (improves readability)"
    echo ""
    echo -e "${YELLOW}Recording Quality:${NC}"
    echo "- Record at normal human speed (not too fast)"
    echo "- Pause 1-2 seconds between interactions"
    echo "- Total duration: 30-60 seconds ideal"
    echo "- Keep GIF < 5MB for GitHub"
    echo ""
    echo -e "${YELLOW}Demo Sequence:${NC}"
    echo "1. Clear screen"
    echo "2. Type: lazyarchon"
    echo "3. Wait for connection"
    echo "4. Show task list (scroll j/k)"
    echo "5. Fast scroll (J/K)"
    echo "6. Search (/ api)"
    echo "7. Feature filter (f)"
    echo "8. Status change (t)"
    echo "9. Project switch (p)"
    echo "10. Help (?) and quit (q)"
    echo ""
    echo -e "${YELLOW}File Size Optimization:${NC}"
    echo "- Use --speed 1.5 to speed up boring parts"
    echo "- Adjust --font-size if too large"
    echo "- agg compression is automatic"
    echo "- Result should be < 5MB"
    echo ""
    echo -e "${YELLOW}Troubleshooting:${NC}"
    echo "- If terminal looks corrupted, try different terminal app"
    echo "- If GIF is too large, increase speed or reduce font size"
    echo "- If too fast/slow, adjust video speed in editing"
    echo ""
}

# Menu
echo -e "${BLUE}What would you like to do?${NC}"
echo "1) Record new demo"
echo "2) Convert existing .cast file to GIF"
echo "3) Show recording tips"
echo "4) Exit"
echo ""
read -p "Choose option (1-4): " choice

case $choice in
    1)
        record_demo
        ;;
    2)
        convert_to_gif
        ;;
    3)
        show_tips
        ;;
    4)
        echo "Goodbye!"
        exit 0
        ;;
    *)
        echo -e "${RED}Invalid option${NC}"
        exit 1
        ;;
esac
