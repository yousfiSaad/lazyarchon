# LazyArchon

> A terminal-based task management TUI for Archon, inspired by [lazygit](https://github.com/jesseduffield/lazygit)/[lazydocker](https://github.com/jesseduffield/lazydocker)

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)]()
[![Homebrew](https://img.shields.io/badge/Homebrew-Available-orange?style=flat&logo=homebrew)](https://github.com/yousfisaad/homebrew-lazyarchon)

A powerful terminal UI for managing [Archon](https://github.com/coleam00/Archon) projects and tasks. Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), it provides a vim-like navigation experience for efficient task management directly from your terminal.

## 🎥 Demo

See LazyArchon in action:

![LazyArchon Demo](assets/demo/lazyarchon-demo.gif)

## ✨ Features

- **Two-Panel Interface**: Task list on the left, detailed view on the right
- **Task Management**: View, edit status, manage features, and filter by features
- **Vim-like Navigation**: Use `h/j/k/l` keys for efficient browsing
- **Rich Features**:
  - Change task status with `s` (todo → doing → review → done)
  - Assign features to tasks with `e`
  - Filter by features with `f` (multi-select)
  - Search tasks with `/` and navigate with `n/N`
- **Interactive Modals**: Intuitive modal interfaces for task management
- **Real-time Updates**: Changes reflect immediately after editing
- **Help System**: Press `?` for complete keyboard shortcuts
- **Responsive Design**: Handles terminal resize with automatic content reflow

## 🚀 Quick Install

**Homebrew (recommended, macOS/Linux):**
```bash
brew install yousfisaad/lazyarchon/lazyarchon
```

**Go install (developers):**
```bash
go install github.com/yousfisaad/lazyarchon/cmd/lazyarchon@latest
```

**Direct download:**
Download from [Latest Release](https://github.com/yousfisaad/lazyarchon/releases/latest), then:
```bash
tar -xzf lazyarchon-*.tar.gz
sudo mv lazyarchon /usr/local/bin/
lazyarchon --version
```

**Build from source:**
```bash
git clone https://github.com/yousfisaad/lazyarchon
cd lazyarchon && make build
./bin/lazyarchon
```

### Prerequisites

- **Running [Archon](https://github.com/coleam00/Archon) API** on `localhost:8181`
- **Go 1.24+** (only for go install or build from source)

## ⌨️ Essential Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `?` | Show help (all shortcuts) |
| `h/l` | Switch panels (Tasks ↔ Details) |
| `j/k` | Navigate up/down (1 line) |
| `J/K` | Fast scroll (4 lines) |
| `s` | Change task status |
| `e` | Edit task features |
| `f` | Filter by features |
| `/` | Search tasks |
| `n/N` | Next/previous search result |
| `p` | Select project |
| `r` | Refresh data |
| `q` | Quit |

**For complete keyboard reference, see [Key Bindings](docs/user-guide/key-bindings.md).**

## 🚦 Current Status

**What LazyArchon is good for:**
- ✅ Browsing and exploring Archon tasks and projects
- ✅ Managing task status (todo/doing/review/done)
- ✅ Managing task features and filtering
- ✅ Reading task details in a clean terminal interface
- ✅ Navigating large task lists efficiently

**Not yet available:**
- ❌ Task creation and advanced editing
- ❌ Project management operations
- ❌ Real-time WebSocket updates (manual refresh required)
- ❌ User authentication

## 📚 Documentation

- **[📖 Full Documentation](docs/README.md)** - Complete documentation index
- **[🚀 Getting Started](docs/getting-started/README.md)** - Installation, configuration, quick start
- **[👤 User Guide](docs/user-guide/README.md)** - Features, troubleshooting, advanced usage
- **[🛠️ Development](docs/development/README.md)** - Contributing, testing, build system
- **[⚙️ Configuration](docs/getting-started/configuration.md)** - Customize API endpoint and preferences
- **[🔧 Troubleshooting](docs/user-guide/troubleshooting.md)** - Common issues and solutions

## 🛠️ Development

```bash
# Clone and setup
git clone https://github.com/yousfisaad/lazyarchon
cd lazyarchon

# Build and run
make build && ./bin/lazyarchon

# Or directly run from source
go run ./cmd/lazyarchon

# Run tests
make test

# See all available targets
make help
```

### Built With
- **Go 1.24+** - Programming language
- **Bubble Tea** - Terminal UI framework
- **Lip Gloss** - Terminal styling

## 🎯 Roadmap

- [ ] Task creation forms
- [ ] Advanced task editing (title, description, sources)
- [ ] WebSocket-based real-time updates
- [ ] Configuration files and themes
- [ ] Offline mode and caching

See [Full Roadmap](docs/README.md#roadmap) for more details.

## 🔗 Links

- **Issues**: [GitHub Issues](https://github.com/yousfisaad/lazyarchon/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yousfisaad/lazyarchon/discussions)

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **[Charm Bracelet](https://charm.sh/)** - For the amazing Bubble Tea ecosystem
- **[Archon](https://github.com/coleam00/Archon)** - For the task management platform
- **[lazygit](https://github.com/jesseduffield/lazygit)/[lazydocker](https://github.com/jesseduffield/lazydocker)** - For UI/UX inspiration

---

**Made with ❤️ and Go** | **Happy Task Managing! 🚀**