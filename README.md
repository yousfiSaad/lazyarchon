# LazyArchon

> A terminal-based task management TUI for Archon, inspired by [lazygit](https://github.com/jesseduffield/lazygit)/[lazydocker](https://github.com/jesseduffield/lazydocker)

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)]()

[![Homebrew](https://img.shields.io/badge/Homebrew-Available-orange?style=flat&logo=homebrew)](https://github.com/yousfisaad/homebrew-lazyarchon)
[![Go Install](https://img.shields.io/badge/Go%20Install-Latest-blue?style=flat&logo=go)](https://pkg.go.dev/github.com/yousfisaad/lazyarchon)
[![Script Install](https://img.shields.io/badge/Script%20Install-curl%20%7C%20bash-green?style=flat&logo=gnu-bash)](https://raw.githubusercontent.com/yousfisaad/lazyarchon/main/scripts/install.sh)

LazyArchon is a powerful terminal user interface (TUI) for managing [Archon](https://github.com/coleam00/Archon) projects and tasks. Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), it provides a vim-like navigation experience for efficient task management directly from your terminal.

## ✨ Features

> **📋 Current Status**: LazyArchon now includes comprehensive task management! View, navigate, edit task features, filter by features, and manage task status with intuitive modal interfaces.

### ✅ Implemented Features

**Core Functionality**
- **Two-Panel Interface**: Task list on the left, detailed view on the right
- **Project Browsing**: Browse and filter tasks by project
- **Task Viewing**: View task details, status, assignee, and metadata
- **Task Status Management**: Change task status with interactive modal (press `s` key)
- **Task Feature Management**: Edit and assign features to tasks (press `e` key)
- **Feature-based Filtering**: Filter tasks by features with multi-select modal (press `f` key) and high-contrast checkboxes (■/□)
- **Task Status Indicators**: Visual symbols for todo/doing/review/done states
- **Markdown Rendering**: Rich text display for task descriptions with syntax highlighting

**Navigation & UX**
- **Comprehensive Help System**: Press `?` for modal with all shortcuts organized by category
- **Enhanced Status Bar**: Displays comprehensive keyboard shortcuts contextually based on active modal
- **Active Panel System**: Vim-style `h/l` keys to switch between task list and details panels
- **Progressive Scrolling**: Three speed levels - `j/k` (precise), `J/K` (fast), `ctrl+u/d` (half-page)
- **Advanced Modal Navigation**: Full vim navigation in feature modal (gg/G, J/K, ctrl+u/d, home/end)
- **Unified Navigation**: Same scrolling system works in main interface AND all modals
- **Visual Panel Feedback**: Active panels highlighted with bright borders, inactive dimmed
- **Project Selection Mode**: Press `p` to browse and select projects
- **Smart Scroll Bars**: Visual position indicators with percentage feedback
- **Responsive Design**: Real-time window resize handling with automatic content reflow
- **Filter Preservation**: Feature modal preserves selected filters when reopened
- **Error Handling**: Graceful API failure handling with retry options
- **Enhanced Search System**: Inline search with real-time highlighting and vim-style navigation
- **Text Search & Highlighting**: Search task titles with yellow highlighting of matching terms
- **Search Navigation**: n/N keys for cycling through search results with position indicators
- **Multiple Clear Options**: Ctrl+L or Ctrl+X to clear active searches

**Data Integration**
- **API Connectivity**: Fetches current data from Archon API server
- **Manual Refresh**: Press `r` to reload data from server
- **Project Filtering**: View tasks filtered by specific projects
- **Real-time Updates**: Task changes reflect immediately after editing
- **Feature Management**: Create new features or assign existing ones to tasks

### 🚧 Planned Features

**Task Management**
- Task creation (forms and workflows)
- Advanced task editing (title, description, sources, code examples)
- Project management operations  
- Bulk task operations

**Real-time Features**
- WebSocket-based live updates
- Automatic refresh when data changes
- Real-time collaboration indicators

**Advanced Features**
- Configuration files and themes
- Offline mode and caching
- Enhanced task editing and creation forms

## 🚀 Installation

**Quick Install:**
```bash
curl -sSL https://raw.githubusercontent.com/yousfisaad/lazyarchon/main/scripts/install.sh | bash
```

**Other Methods:** Go install • Homebrew • Binary downloads • Build from source

### Prerequisites
- **Go 1.24+** (required for go install and building from source)
- **Running [Archon](https://github.com/coleam00/Archon) API** server on `localhost:8181`
- **Note**: LazyArchon supports comprehensive task management - view, browse, update status, and manage features

### ⚠️ Current Limitations

The following features are **not yet available**:

- ❌ **Task Creation**: Cannot create new tasks (forms not yet implemented)
- ❌ **Advanced Task Editing**: Cannot modify task titles, descriptions, sources, or code examples  
- ❌ **Project Management**: Cannot create or edit projects
- ❌ **Real-time Updates**: No WebSocket connection (manual refresh required for external changes)
- ❌ **User Authentication**: No login/auth system (uses anonymous API access)
- ❌ **Offline Mode**: Requires active API connection
- ✅ **Configuration**: Config files and customization options available
- ✅ **Text Search**: Enhanced inline search with real-time highlighting and n/N navigation

**What LazyArchon IS good for:**
- ✅ Browsing and exploring your Archon tasks and projects
- ✅ **Managing task status** - change todo/doing/review/done states
- ✅ **Managing task features** - assign existing features or create new ones
- ✅ **Filtering by features** - multi-select feature filtering with task counts
- ✅ Getting an overview of project status and task distribution  
- ✅ Reading task details and descriptions in a clean terminal interface
- ✅ Navigating large task lists efficiently with vim-like controls

### Method 1: One-Line Install (Recommended)

**Install latest version:**
```bash
curl -sSL https://raw.githubusercontent.com/yousfisaad/lazyarchon/main/scripts/install.sh | bash
```

**System-wide install (requires sudo):**
```bash
curl -sSL https://raw.githubusercontent.com/yousfisaad/lazyarchon/main/scripts/install.sh | bash -s -- --dir /usr/local/bin
```

**Install specific version:**
```bash
curl -sSL https://raw.githubusercontent.com/yousfisaad/lazyarchon/main/scripts/install.sh | bash -s -- --version v1.0.0
```

**Verification:**
```bash
lazyarchon --version  # Check version
which lazyarchon      # Check installation path
lazyarchon --help     # Test basic functionality
```

**Uninstall:**
```bash
curl -sSL https://raw.githubusercontent.com/yousfisaad/lazyarchon/main/scripts/uninstall.sh | bash
```

### Method 2: Homebrew (macOS/Linux)

**Quick Install:**
```bash
brew install yousfisaad/lazyarchon/lazyarchon
```

**First-time Setup (if tap not found):**
```bash
# Add the tap
brew tap yousfisaad/lazyarchon

# Install LazyArchon
brew install lazyarchon
```

**Update to Latest:**
```bash
brew update && brew upgrade lazyarchon
```

**Verification:**
```bash
lazyarchon --version
which lazyarchon  # Should show: /opt/homebrew/bin/lazyarchon (Apple Silicon) or /usr/local/bin/lazyarchon (Intel)
```

**Platform Support:**
- ✅ **macOS**: Both Intel and Apple Silicon (ARM64)
- ✅ **Linux**: AMD64 and ARM64 architectures
- ❌ **Windows**: Use Method 1 (Script Install) instead

**Troubleshooting:**
- **Tap not found**: Ensure you have access to the internet and GitHub
- **Permission denied**: Try `brew doctor` and fix any issues
- **Old version**: Run `brew update && brew upgrade lazyarchon`
- **Command not found**: Restart terminal or check `brew --prefix`/bin is in PATH

### Method 3: Go Install (Go Developers)

**Prerequisites**: Go 1.24+ installed

```bash
# Install latest version
go install github.com/yousfisaad/lazyarchon/cmd/lazyarchon@latest
```

**Verification:**
```bash
lazyarchon --version  # Check version
which lazyarchon      # Should show: $(go env GOPATH)/bin/lazyarchon
go env GOPATH         # Check Go path is correct
lazyarchon --help     # Test basic functionality
```

**Troubleshooting:**
- Ensure `$GOPATH/bin` or `$GOBIN` is in your PATH
- If command not found, check: `go env GOPATH` and add `$(go env GOPATH)/bin` to PATH
- Restart terminal after adding to PATH

### Method 4: Download Binary Releases

Visit the [GitHub Releases](https://github.com/yousfisaad/lazyarchon/releases) page to download pre-built binaries for your platform:

- **Linux**: `lazyarchon-linux-amd64.tar.gz`
- **macOS**: `lazyarchon-darwin-amd64.tar.gz`
- **Windows**: `lazyarchon-windows-amd64.zip`

**Installation:**
```bash
# Extract and install (Linux/macOS example)
tar -xzf lazyarchon-linux-amd64.tar.gz
sudo mv lazyarchon /usr/local/bin/

# For Windows (extract .zip file)
# Move lazyarchon.exe to a directory in your PATH
```

**Verification:**
```bash
lazyarchon --version  # Check version
which lazyarchon      # Should show: /usr/local/bin/lazyarchon
ls -la /usr/local/bin/lazyarchon  # Check permissions
lazyarchon --help     # Test basic functionality
```

### Method 4: Build from Source

**Using Makefile (Development)**
```bash
# Clone the repository
git clone https://github.com/yousfisaad/lazyarchon
cd lazyarchon

# Quick start: build and run
make run

# Or step by step
make build          # Build for current platform
./bin/lazyarchon    # Run the binary
```

**Using Go Commands (Alternative)**
```bash
# Direct Go commands (no Makefile needed)
go run ./cmd/lazyarchon                      # Run directly
go build -o lazyarchon ./cmd/lazyarchon      # Build binary
```

**Cross-Platform Builds**

For cross-platform builds, use GoReleaser:
```bash
# Install GoReleaser
go install github.com/goreleaser/goreleaser@latest

# Build for all platforms without publishing
goreleaser build --snapshot --clean

# Binaries will be in: dist/
```

### 📱 Platform-Specific Installation Notes

#### macOS
**Recommended Method:** Homebrew
```bash
brew install yousfisaad/lazyarchon/lazyarchon
```
- ✅ **Apple Silicon (M1/M2/M3)**: Native ARM64 support
- ✅ **Intel Macs**: Native AMD64 support
- 🔧 **Troubleshooting**: If Homebrew fails, use script install as fallback

#### Linux
**Recommended Method:** One-line script install
```bash
curl -sSL https://raw.githubusercontent.com/yousfisaad/lazyarchon/main/scripts/install.sh | bash
```
- ✅ **Ubuntu/Debian**: All methods work
- ✅ **CentOS/RHEL/Fedora**: All methods work
- ✅ **ARM64 (Raspberry Pi, etc.)**: All methods supported
- 🔧 **Note**: Homebrew on Linux requires additional setup

#### Windows
**Recommended Method:** Script install via WSL
```bash
# In WSL/PowerShell:
curl -sSL https://raw.githubusercontent.com/yousfisaad/lazyarchon/main/scripts/install.sh | bash
```
- ✅ **WSL 1/2**: Script install or Go install
- ✅ **PowerShell**: Use script install
- ❌ **Native Windows**: Download binary manually
- 🔧 **Note**: Homebrew not supported on Windows

#### Architecture Support
- ✅ **AMD64 (x86_64)**: All platforms
- ✅ **ARM64 (aarch64)**: macOS, Linux
- ❌ **ARM32**: Not currently supported
- ❌ **32-bit systems**: Not supported

### Quick Start

**For Homebrew Users:**
```bash
# Install and run in one command
brew install yousfisaad/lazyarchon/lazyarchon && lazyarchon
```

**For Go Developers:**
```bash
# Install and run in one command
go install github.com/yousfisaad/lazyarchon/cmd/lazyarchon@latest && lazyarchon
```

**For Everyone Else:**
```bash
# One-line install script
curl -sSL https://raw.githubusercontent.com/yousfisaad/lazyarchon/main/scripts/install.sh | bash && lazyarchon
```

**For Development:**
```bash
# Clone and run from source
git clone https://github.com/yousfisaad/lazyarchon
cd lazyarchon
go run ./cmd/lazyarchon
```

## 📖 Usage

### Main Interface

```
LazyArchon - Project Name • 🔍 "api" (15)
┌─────────────────────────┐  ┌─────────────────────────────────┐
│ ○ Initialize project    │  │ # Task Title                    │
│ ◐ Implement API client  │  │                                 │
│ ◉ Add error handling    │  │ Detailed task description with  │
│ ● Create documentation  │  │ markdown formatting support.    │
│   [15 more tasks...]  ▓ │  │                               ░ │
│                       ░ │  │ **Status:** doing             ░ │
│                       ░ │  │ **Assignee:** AI IDE Agent   ▓ │
│                       ░ │  │ **Feature:** ui-navigation    ░ │
└─────────────────────────┘  └─────────────────────────────────┘
[Tasks] Connected • Match 2/3 • Sort: Status | /: search | n/N: next/prev match | Ctrl+L: clear search | ?: help
```

### Task Status Indicators
- `○` **Todo** - Not started
- `◐` **Doing** - In progress  
- `◉` **Review** - Under review
- `●` **Done** - Completed

### Basic Navigation
1. **Launch** LazyArchon in your terminal
2. **Get help** with `?` key - shows all available shortcuts
3. **Browse** tasks with `j/k` or arrow keys (left panel active by default)
4. **Switch panels** with `h` (left/tasks) and `l` (right/details)
5. **Navigate active panel** - all movement keys work on highlighted panel
6. **Select projects** with `p` key for filtering
7. **Edit task features** with `e` key - assign or create features
8. **Filter by features** with `f` key - multi-select filtering
9. **Search tasks** with `/` key - inline search with highlighting and n/N navigation
10. **Change task status** with `t` key - todo/doing/review/done
11. **Scroll at different speeds**: `j/k` (line), `J/K` (fast), `Ctrl+u/d` (half-page)

## ⌨️ Keyboard Shortcuts

### Panel Navigation & Movement
| Key | Action |
|-----|--------|
| `h` | Switch to left panel (Tasks) |
| `l` | Switch to right panel (Details) |
| `↑/↓` or `j/k` | Navigate/scroll in active panel (1 line) |
| `J/K` | Fast scroll in active panel (4 lines) |
| `Ctrl+u/Ctrl+d` | Half-page scroll in active panel |
| `gg` | Jump to first item in active panel |
| `G` | Jump to last item in active panel |
| `Home/End` | Jump to start/end in active panel |

### Project Management
| Key | Action |
|-----|--------|
| `p` | Enter project selection mode |
| `a` | Show all tasks (exit project filter) |
| `Enter` | Select project in project mode |
| `Esc` | Exit project mode |

### Task Management
| Key | Action |
|-----|--------|
| `s` | **Change sorting criteria** (cycle through sort modes) |
| `t` | **Change task status** (Todo/Doing/Review/Done) |
| `e` | **Edit task features** (assign/create features) |
| `f` | **Filter by features** (multi-select modal) |

### Search & Navigation
| Key | Action |
|-----|--------|
| `/` or `Ctrl+f` | **Start inline search** (search in task titles) |
| `Enter` | **Apply search** and exit search mode |
| `Esc` | **Cancel search** and exit search mode |
| `Ctrl+L` or `Ctrl+X` | **Clear search** when search is active |
| `n` | **Next match** - jump to next search result |
| `N` | **Previous match** - jump to previous search result |

### Feature Modal Navigation (when f is pressed)
| Key | Action |
|-----|--------|
| `j/k` or `↑/↓` | Navigate feature list (1 item) |
| `J/K` | Fast scroll feature list (5 items) |
| `Ctrl+u/Ctrl+d` | Half-page scroll feature list |
| `gg` | Jump to first feature |
| `G` | Jump to last feature |
| `Home/End` | Jump to first/last feature |
| `Space` | Toggle feature selection |
| `a` | Select all features |
| `n` | Select no features |
| `Enter` | Apply filter and close modal |
| `Esc/q` | Cancel and close modal |

### Help & Application Controls
| Key | Action |
|-----|--------|
| `?` | **Toggle help modal** with all shortcuts |
| `r` or `F5` | Refresh data from API |
| `q` | Quit application |

### Help Modal Navigation (when ? is pressed)
| Key | Action |
|-----|--------|
| `j/k` or `↑/↓` | Scroll help content (1 line) |
| `J/K` | Fast scroll help (4 lines) |
| `Ctrl+u/d` | Half-page scroll help |
| `gg/G` | Jump to help top/bottom |
| `?` or `Esc` | Close help modal |

### Progressive Scrolling System

LazyArchon features a three-tier scrolling system for precise navigation:

| Speed | Keys | Movement | Best For |
|-------|------|----------|----------|
| **Precise** | `j/k` or `↑/↓` | 1 line | Fine positioning and reading |
| **Fast** | `J/K` (Shift+j/k) | 4 lines | Quick browsing through lists |  
| **Jump** | `Ctrl+u/Ctrl+d` | Half-page | Rapid navigation in long content |

**Key Benefits:**
- **No Key Conflicts**: Each key has a distinct, useful purpose
- **Natural Progression**: Intuitive speed scaling with familiar vim patterns
- **Active Panel Aware**: All scrolling works on the highlighted panel
- **Consistent Behavior**: Same speeds work in both task list and details

### Visual Indicators
| Element | Meaning |
|---------|---------|
| **Bright cyan border** | Active panel (receives input) |
| **Dim gray border** | Inactive panel |
| **[Tasks]** / **[Details]** | Active panel name in status bar |
| **▓░** scroll bar | Position indicator with thumb (▓) and track (░) |

## ⚙️ Configuration

### API Connection
LazyArchon connects to your Archon API server at `http://localhost:8181` by default.

**Future Configuration Options:**
- Environment variables for API endpoint
- Configuration file support
- Custom keybindings
- Theme customization

### Environment Variables (Future)
```bash
export ARCHON_API_URL="http://localhost:8181"
export ARCHON_API_KEY="your-api-key"
```

## 🏗️ Development

### Project Structure
```
lazyarchon/
├── cmd/lazyarchon/          # Application entry point
│   └── main.go
├── internal/
│   ├── ui/                  # TUI interface logic
│   │   └── app.go
│   └── archon/              # API client and models
│       ├── client.go
│       └── models.go
├── go.mod                   # Go module definition
└── README.md
```

### Key Dependencies
- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** - TUI framework
- **[Lip Gloss](https://github.com/charmbracelet/lipgloss)** - Styling and layout
- **[Glamour](https://github.com/charmbracelet/glamour)** - Markdown rendering

### Building for Development
```bash
# Install dependencies
go mod download

# Development build
make build

# Run tests
make test

# Run linting
make lint

# Clean build artifacts
make clean

# Run with live reload (using air)
air

# Generate test coverage
make test-coverage
```

### Release Management

#### Automated Releases (GitHub Actions)
LazyArchon uses GoReleaser with GitHub Actions for fully automated releases:

```bash
# Create and push a new tag to trigger release
git tag v1.2.0
git push origin v1.2.0

# GitHub Actions will automatically:
# 1. Build cross-platform binaries
# 2. Create GitHub release with changelog
# 3. Upload release assets
# 4. Update Homebrew tap (yousfisaad/homebrew-lazyarchon)
```

#### Manual Testing (Development)
```bash
# Test GoReleaser configuration
goreleaser check

# Build snapshot (without publishing)
goreleaser build --snapshot --clean

# Manual release (local testing only)
goreleaser release --clean --skip=publish
```

#### GitHub Actions Workflows
- **Release** (`.github/workflows/release.yml`) - Automated release on git tags
- **GoReleaser Test** (`.github/workflows/goreleaser-test.yml`) - Validates config on PRs

### Available Make Targets
```bash
make help          # Show all available targets
make run           # Build and run (quick development)
make build         # Build for current platform
make test          # Run all tests
make test-coverage # Run tests with coverage report
make lint          # Run code linting and formatting
make clean         # Clean build artifacts
make deps          # Tidy Go modules
```

**Note**: Cross-platform builds are handled by GoReleaser, not Makefile.

### Repository Setup (Maintainers)

For automated Homebrew tap updates, configure these repository secrets:

#### Required Secrets
1. **HOMEBREW_TAP_GITHUB_TOKEN** - Personal Access Token with `repo` scope
   - Used to update the `yousfisaad/homebrew-lazyarchon` repository
   - Generate at: https://github.com/settings/tokens
   - Grant access to: `yousfisaad/homebrew-lazyarchon` repository

#### Repository Configuration
- **Create Homebrew tap repository**: `yousfisaad/homebrew-lazyarchon`
  - Initialize as public repository
  - Add basic README and license
  - GoReleaser will automatically populate with formula
- **Configure repository settings**:
  - Enable issues and discussions (optional)
  - Set repository description: "Homebrew tap for LazyArchon"
- **Automatic updates**: GoReleaser will create/update the Homebrew formula
- **Release trigger**: Workflow runs on any `v*` tag push

### Contributing
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 🔧 Troubleshooting

### Installation Issues

**Homebrew: Tap not found**
```
Error: No available formula with name "yousfisaad/lazyarchon/lazyarchon"
```
- Add the tap first: `brew tap yousfisaad/lazyarchon`
- Update Homebrew: `brew update`
- Check tap status: `brew tap-info yousfisaad/lazyarchon`

**Go Install: Command not found after installation**
```bash
lazyarchon: command not found
```
- Check if `$GOPATH/bin` is in PATH: `echo $PATH | grep $(go env GOPATH)/bin`
- Add to PATH: `export PATH="$(go env GOPATH)/bin:$PATH"`
- Add to shell profile: `echo 'export PATH="$(go env GOPATH)/bin:$PATH"' >> ~/.bashrc`

**Script Install: Permission denied**
```
Permission denied (creating directory)
```
- Use system-wide install: `curl -sSL ... | bash -s -- --dir /usr/local/bin`
- Or install to user directory: `curl -sSL ... | bash -s -- --dir ~/.local/bin`
- Ensure ~/.local/bin is in PATH: `export PATH="$HOME/.local/bin:$PATH"`

**Platform-Specific Issues:**
- **macOS**: If Homebrew not working, try script install or Go install
- **Linux ARM64**: All methods supported, Homebrew recommended
- **Windows**: Use script install (WSL) or download binary directly
- **Older systems**: May need Go 1.24+ for go install method

### Common Issues

**Connection refused / API not available**
```
Error: Failed to load tasks: dial tcp [::1]:8181: connect: connection refused
```
- Ensure Archon API server is running on localhost:8181
- Check if the server is accessible: `curl http://localhost:8181/health`

**Build failures**
```
go: module requires Go 1.24 or later
```
- Update Go to version 1.24 or later
- Check version: `go version`

**Terminal display issues**
- Ensure terminal supports ANSI colors
- Try running in a different terminal emulator
- Check terminal size: LazyArchon adapts to screen size

**Performance issues**
- Large task lists may cause slower rendering
- Use project filtering (`p` key) to reduce displayed items
- Refresh data (`r` key) if UI becomes unresponsive

### Debug Mode *(Planned Feature)*
```bash
# Debug mode not yet implemented - planned for future release
# ./bin/lazyarchon --debug

# Check API connectivity manually
curl -v http://localhost:8181/api/tasks
```

## 🎯 Roadmap

### 🔥 High Priority (Next Release)
- [ ] **Task Creation Forms**: Create new tasks with full form interface
- [x] **Status Management**: Change task status (todo → doing → review → done) ✅
- [x] **Feature Assignment**: Assign and manage task features/tags ✅
- [x] **Feature Filtering**: Multi-select feature-based task filtering ✅
- [x] **Enhanced Modal Navigation**: Advanced vim navigation in all modals ✅
- [x] **Responsive Design**: Real-time window resize with content reflow ✅
- [x] **High-contrast UI**: Improved checkbox styling and visual feedback ✅
- [x] **Enhanced Text Search**: Inline search with real-time highlighting and n/N navigation ✅
- [ ] **Configuration System**: Config files for API endpoint and preferences
- [x] **Help System**: Built-in help with `?` key ✅
- [x] **Build System**: GoReleaser automation + minimal dev Makefile ✅
- [ ] **Debug Mode**: Verbose logging and troubleshooting options

### 🚀 Medium Priority
- [ ] **Real-time Updates**: WebSocket connection for live data
- [ ] **Advanced Task Editing**: Modify titles, descriptions, sources, code examples
- [ ] **Themes & Customization**: Color schemes and layout options
- [ ] **Project Management**: Create and edit projects
- [ ] **User Authentication**: Login system and API key management

### 🌟 Long-term Vision
- [ ] **Cross-platform Releases**: Pre-built binaries for all platforms
- [ ] **Plugin System**: Extensible architecture for custom features
- [ ] **Multiple Server Support**: Connect to different Archon instances
- [ ] **Offline Mode**: Local caching and sync capabilities
- [ ] **Git Integration**: Link tasks to commits and branches

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **[Charm Bracelet](https://charm.sh/)** - For the amazing Bubble Tea ecosystem
- **[Archon](https://github.com/coleam00/Archon)** - For the task management platform
- **[lazygit](https://github.com/jesseduffield/lazygit)/[lazydocker](https://github.com/jesseduffield/lazydocker)** - For UI/UX inspiration
- **Go Community** - For the robust tooling and ecosystem

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/yousfisaad/lazyarchon/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yousfisaad/lazyarchon/discussions)
- **Documentation**: [Wiki](https://github.com/yousfisaad/lazyarchon/wiki)

---

**Made with ❤️ and Go** | **Happy Task Managing! 🚀**