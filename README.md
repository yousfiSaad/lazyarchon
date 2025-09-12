# LazyArchon

> A terminal-based task management TUI for Archon, inspired by lazygit/lazydocker

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)]()

LazyArchon is a powerful terminal user interface (TUI) for managing Archon projects and tasks. Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), it provides a vim-like navigation experience for efficient task management directly from your terminal.

## ✨ Features

> **📋 Current Status**: LazyArchon now includes comprehensive task management! View, navigate, edit task features, filter by features, and manage task status with intuitive modal interfaces.

### ✅ Implemented Features

**Core Functionality**
- **Two-Panel Interface**: Task list on the left, detailed view on the right
- **Project Browsing**: Browse and filter tasks by project
- **Task Viewing**: View task details, status, assignee, and metadata
- **Task Status Management**: Change task status with interactive modal (press `s` key)
- **Task Feature Management**: Edit and assign features to tasks (press `e` key)
- **Feature-based Filtering**: Filter tasks by features with multi-select modal (press `f` key)
- **Task Status Indicators**: Visual symbols for todo/doing/review/done states
- **Markdown Rendering**: Rich text display for task descriptions with syntax highlighting

**Navigation & UX**
- **Comprehensive Help System**: Press `?` for modal with all shortcuts organized by category
- **Clean Status Bar**: Minimal design showing only essential info (active panel, help, quit)
- **Active Panel System**: Vim-style `h/l` keys to switch between task list and details panels
- **Progressive Scrolling**: Three speed levels - `j/k` (precise), `J/K` (fast), `ctrl+u/d` (half-page)
- **Unified Navigation**: Same scrolling system works in main interface AND help modal
- **Visual Panel Feedback**: Active panels highlighted with bright borders, inactive dimmed
- **Project Selection Mode**: Press `p` to browse and select projects
- **Smart Scroll Bars**: Visual position indicators with percentage feedback
- **Responsive Layout**: Adapts to terminal size with proper content wrapping
- **Error Handling**: Graceful API failure handling with retry options

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
- Text-based search (beyond feature filtering)
- Configuration files and themes
- Offline mode and caching
- Enhanced help system and tutorials

## 🚀 Installation

### Prerequisites
- **Go 1.24+** (required for building)
- **Running Archon API** server on `localhost:8181`
- **Note**: LazyArchon supports comprehensive task management - view, browse, update status, and manage features

### ⚠️ Current Limitations

The following features are **not yet available**:

- ❌ **Task Creation**: Cannot create new tasks (forms not yet implemented)
- ❌ **Advanced Task Editing**: Cannot modify task titles, descriptions, sources, or code examples  
- ❌ **Project Management**: Cannot create or edit projects
- ❌ **Real-time Updates**: No WebSocket connection (manual refresh required for external changes)
- ❌ **User Authentication**: No login/auth system (uses anonymous API access)
- ❌ **Offline Mode**: Requires active API connection
- ❌ **Configuration**: No config files or customization options yet
- ❌ **Text Search**: No search functionality (only feature filtering available)

**What LazyArchon IS good for:**
- ✅ Browsing and exploring your Archon tasks and projects
- ✅ **Managing task status** - change todo/doing/review/done states
- ✅ **Managing task features** - assign existing features or create new ones
- ✅ **Filtering by features** - multi-select feature filtering with task counts
- ✅ Getting an overview of project status and task distribution  
- ✅ Reading task details and descriptions in a clean terminal interface
- ✅ Navigating large task lists efficiently with vim-like controls

### Build from Source

**Using Makefile (Recommended)**
```bash
# Clone the repository
git clone https://github.com/yousfisaad/lazyarchon
cd lazyarchon

# Build for current platform
make build

# Run LazyArchon
./bin/lazyarchon

# Or build and run in one step
make run
```

**Manual Build**
```bash
# Build the application manually
go build -o bin/lazyarchon ./cmd/lazyarchon

# Run LazyArchon
./bin/lazyarchon
```

**Cross-platform Builds**
```bash
# Build for all platforms (Linux, macOS, Windows)
make build-all

# Build for specific platform
make build-linux    # Linux AMD64
make build-darwin   # macOS AMD64 
make build-windows  # Windows AMD64
```

### Quick Start
```bash
# Install Go dependencies
go mod download

# Build and run in one step
go run ./cmd/lazyarchon
```

## 📖 Usage

### Main Interface

```
LazyArchon - Project Name (15)                                    
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
[Tasks] Ready | ?: help | q: quit
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
9. **Change task status** with `s` key - todo/doing/review/done
10. **Scroll at different speeds**: `j/k` (line), `J/K` (fast), `Ctrl+u/d` (half-page)

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
| `t` | **Change task status** (Todo/Doing/Review/Done) |

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

# Development build with debugging
make dev

# Run tests
make test

# Run linting
make lint

# Clean build artifacts
make clean

# Run with live reload (using air)
air

# Build for production with optimizations
make build-release
```

### Available Make Targets
```bash
make help          # Show all available targets
make build         # Build for current platform
make build-all     # Cross-platform builds
make build-release # Optimized production build
make test          # Run all tests
make lint          # Run code linting
make clean         # Clean build artifacts
make run           # Build and run
make dev           # Development mode
```

### Contributing
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 🔧 Troubleshooting

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
- [ ] **Configuration System**: Config files for API endpoint and preferences
- [x] **Help System**: Built-in help with `?` key ✅
- [x] **Build System**: Cross-platform builds with Makefile ✅
- [ ] **Debug Mode**: Verbose logging and troubleshooting options

### 🚀 Medium Priority
- [ ] **Real-time Updates**: WebSocket connection for live data
- [ ] **Text-based Search**: Find tasks by content, title, description
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
- **[Archon Project](https://github.com/archon-project)** - For the task management platform
- **lazygit/lazydocker** - For UI/UX inspiration
- **Go Community** - For the robust tooling and ecosystem

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/yousfisaad/lazyarchon/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yousfisaad/lazyarchon/discussions)
- **Documentation**: [Wiki](https://github.com/yousfisaad/lazyarchon/wiki)

---

**Made with ❤️ and Go** | **Happy Task Managing! 🚀**