# Demo Recording Guide

Three scripts are available for recording LazyArchon demos, from simplest to most advanced.

## 🚀 Quick Start (Recommended)

### Option 1: `quick-demo.sh` - Simplest, fully automated

**Best for: First-time recordings, quick demos**

```bash
cd /Volumes/SSD-NVMe/Dev/lazyarchon
./scripts/quick-demo.sh
```

**What it does:**
- Checks and installs prerequisites automatically
- Shows clear on-screen instructions
- Records your demo
- Converts to GIF automatically
- Done! (~60 lines, no menus)

**When to use:** You want the easiest experience with clear guidance.

---

### Option 2: `demo.sh` - Ultra-simple, no hand-holding

**Best for: Experienced users who know what they're doing**

```bash
cd /Volumes/SSD-NVMe/Dev/lazyarchon
./scripts/demo.sh
```

**What it does:**
- Records immediately (no checks, no prompts)
- Converts to GIF
- Done! (~12 lines, minimal)

**When to use:** You've already recorded demos before and just want the fastest workflow.

---

### Option 3: `record-demo.sh` - Advanced, interactive menu

**Best for: Multiple recordings, experimentation, detailed control**

```bash
cd /Volumes/SSD-NVMe/Dev/lazyarchon
./scripts/record-demo.sh
```

**What it does:**
- Interactive menu with 4 options:
  1. Record new demo
  2. Convert existing .cast file to GIF
  3. Show recording tips
  4. Exit
- Detailed guidance and error handling
- Flexible workflow (~208 lines)

**When to use:** You want to record multiple demos, re-convert existing recordings, or see detailed tips.

---

## 📝 Demo Recording Tips

Regardless of which script you use, follow this sequence for a great demo:

**Setup (before recording):**
- Terminal size: 80x24 or larger
- Font size: 14-16pt for visibility
- Dark theme recommended
- Make sure Archon API is running: `curl http://localhost:8181/health`

**Demo sequence (30-60 seconds):**
1. Clear screen → Launch lazyarchon (5 sec)
2. Navigate with j/k and J/K (10 sec)
3. Search with `/` and type "api" (8 sec)
4. Filter features with `f` (10 sec)
5. Change task status with `t` (10 sec)
6. Show help with `?` (5 sec)
7. Quit with `q` (2 sec)

**Recording tips:**
- Pause 1-2 seconds between actions
- Don't rush - let features be visible
- Press Ctrl+D to stop recording

---

## 🎯 Quick Comparison

| Script | Lines | Setup | Control | Best For |
|--------|-------|-------|---------|----------|
| `quick-demo.sh` | ~60 | Auto | Guided | First-time users |
| `demo.sh` | ~12 | None | None | Experienced users |
| `record-demo.sh` | ~208 | Auto | Full | Multiple recordings |

---

## 🔧 Prerequisites

All scripts require:
- `asciinema` - Terminal recording
- `agg` - GIF conversion
- `lazyarchon` - The app itself

**Install if needed:**
```bash
brew install asciinema agg
make build  # Builds lazyarchon
```

The `quick-demo.sh` and `record-demo.sh` scripts will auto-install these if missing.

---

## 📁 Output Files

All scripts create:
- `assets/demo/lazyarchon-demo.cast` - Raw recording
- `assets/demo/lazyarchon-demo.gif` - Final GIF

**Add to README:**
```markdown
![LazyArchon Demo](assets/demo/lazyarchon-demo.gif)
```

---

## ⚡ Recommended Workflow

1. **First demo:** Use `quick-demo.sh` for guidance
2. **Refining:** Use `record-demo.sh` to re-record or adjust
3. **Quick updates:** Use `demo.sh` when you know exactly what to do

---

**Need help?** See `assets/demo/DEMO_SCRIPT.md` for detailed feature walkthrough.