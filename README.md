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

## 🎥 Demo

See LazyArchon in action! Here's a quick preview of the interface and key features:

![LazyArchon Demo](assets/demo/lazyarchon-demo.gif)

For a detailed walkthrough, view the [full asciinema recording](https://asciinema.org/a/YOUR_ID_HERE) or check out the [Quick Start Guide](docs/getting-started/quick-start.md).

## 📖 Usage

Get started quickly with LazyArchon! For detailed documentation, see our comprehensive [User Guide](docs/user-guide/README.md).

### Quick Start

**New to LazyArchon?** Check out our **[Quick Start Guide](docs/getting-started/quick-start.md)** to get up and running in 5 minutes.

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

### Essential Navigation
- **Get Help**: Press `?` for complete keyboard shortcuts
- **Switch Panels**: `h` (tasks) / `l` (details)
- **Navigate**: `j/k` (precise) / `J/K` (fast) / `Ctrl+u/d` (jump)
- **Task Management**: `t` (change status) / `e` (edit features) / `f` (filter)
- **Search**: `/` (search) / `n/N` (navigate results)

**For complete navigation details, see [Key Bindings](docs/user-guide/key-bindings.md).**

## ⌨️ Keyboard Shortcuts

LazyArchon provides comprehensive keyboard shortcuts for efficient navigation. Here are the most essential ones:

### Quick Reference
| Key | Action |
|-----|--------|
| `?` | **Show complete help** (all shortcuts) |
| `h/l` | Switch between panels (Tasks/Details) |
| `j/k` | Navigate up/down (1 line) |
| `J/K` | Fast scroll (4 lines) |
| `t` | Change task status |
| `f` | Filter by features |
| `/` | Search tasks |
| `p` | Select project |
| `r` | Refresh data |
| `q` | Quit |

**For the complete keyboard reference with all shortcuts, navigation patterns, and modal controls, see [Key Bindings Documentation](docs/user-guide/key-bindings.md).**

## ⚙️ Configuration

LazyArchon supports comprehensive configuration through YAML files and environment variables.

### Quick Setup

**Default Connection:** LazyArchon connects to `http://localhost:8181` by default.

**Basic Configuration:**
```bash
# Create config directory
mkdir -p ~/.config/lazyarchon

# Create basic config
cat > ~/.config/lazyarchon/config.yaml << EOF
archon:
  url: "http://localhost:8181"
  api_key: ""

ui:
  theme:
    name: "default"
  display:
    status_color_scheme: "blue"
EOF
```

**Environment Override:**
```bash
export LAZYARCHON_API_URL="http://localhost:8181"
export LAZYARCHON_API_KEY="your-api-key"
```

**For complete configuration options including themes, performance tuning, and advanced settings, see [Configuration Guide](docs/getting-started/configuration.md).**

## 🏗️ Development

Contributing to LazyArchon? Check out our comprehensive development documentation!

### Quick Start for Developers
```bash
# Clone and setup
git clone https://github.com/yousfisaad/lazyarchon
cd lazyarchon
make setup && make build

# Run tests and start developing
make test && make dev
```

### Key Information
- **Built with:** Go 1.24+, [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **Architecture:** Terminal UI with API integration
- **Testing:** Comprehensive test suite with coverage reporting
- **Building:** Make-based build system with GoReleaser

**For complete development information including architecture, testing, contribution guidelines, and build processes, see [Development Documentation](docs/development/README.md).**

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

### Quick Fixes for Common Issues

| Problem | Quick Solution |
|---------|----------------|
| `command not found: lazyarchon` | Check PATH, reinstall |
| `connection refused` | Start Archon API server |
| `permission denied` | Fix file permissions |
| Display corruption | Try different terminal |

### Installation Help
```bash
# Verify installation
lazyarchon --version

# Test API connectivity
curl http://localhost:8181/health

# Reinstall if needed
curl -sSL https://raw.githubusercontent.com/yousfisaad/lazyarchon/main/scripts/install.sh | bash
```

**For comprehensive troubleshooting including detailed solutions, debug techniques, and platform-specific fixes, see [Troubleshooting Guide](docs/user-guide/troubleshooting.md).**

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

## 📚 Documentation

### Complete Documentation
- **[📖 Documentation Home](docs/README.md)** - Complete documentation index
- **[🚀 Getting Started](docs/getting-started/README.md)** - Installation, quick start, configuration
- **[👤 User Guide](docs/user-guide/README.md)** - Features, key bindings, troubleshooting
- **[🛠️ Development](docs/development/README.md)** - Contributing, testing, build system
- **[🏗️ Architecture](docs/architecture/README.md)** - Design, patterns, evolution plans
- **[📋 Reference](docs/reference/README.md)** - API, CLI options, configuration schema

### Quick Links
- **[Installation Guide](docs/getting-started/installation.md)** - Get LazyArchon installed
- **[Quick Start](docs/getting-started/quick-start.md)** - 5-minute getting started guide
- **[Key Bindings](docs/user-guide/key-bindings.md)** - Complete keyboard shortcuts
- **[Configuration](docs/getting-started/configuration.md)** - Customize your setup
- **[Features](docs/user-guide/features.md)** - All current and planned features
- **[Troubleshooting](docs/user-guide/troubleshooting.md)** - Solve common issues

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/yousfisaad/lazyarchon/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yousfisaad/lazyarchon/discussions)
- **Documentation**: [Complete Docs](docs/README.md)

---

**Made with ❤️ and Go** | **Happy Task Managing! 🚀**