# LazyArchon Architecture

## Overview

LazyArchon is a terminal-based task management interface (TUI) for Archon, built using the Bubble Tea framework in Go. It provides a clean, keyboard-driven interface for managing tasks and projects.

## Project Structure

```
lazyarchon/
├── cmd/lazyarchon/          # Application entry point
├── internal/
│   ├── archon/              # Archon API client
│   ├── config/              # Configuration management
│   └── ui/                  # TUI implementation
├── configs/                 # Configuration files
├── scripts/                 # Build and utility scripts
└── docs/                   # Documentation
```

## Core Components

### 1. Archon API Client (`internal/archon/`)
- **`client.go`** - HTTP client for Archon API
- **`models.go`** - Data structures and API response types

### 2. Configuration Management (`internal/config/`)
- **`config.go`** - Configuration loading and management
- Supports YAML files and environment variables
- Fallback to sensible defaults

### 3. TUI Implementation (`internal/ui/`)
- **`app.go`** - Core Bubble Tea lifecycle (Init, Update)
- **`model.go`** - Application state management
- **`views.go`** - UI rendering methods
- **`keyhandler.go`** - Keyboard input handling
- **`commands.go`** - Async operations and API calls
- **`sorting.go`** - Task sorting logic with caching
- **`styles.go`** - LipGloss styling and color definitions

## Key Features

### Task Management
- View tasks filtered by project or show all tasks
- Sort by status+priority, priority only, creation time, or alphabetical
- Task detail view with scrolling support
- Real-time status updates

### Project Management
- Project selection mode
- Filter tasks by specific projects
- Project-aware task loading

### UI/UX Features
- Vim-like keyboard navigation
- Responsive layout with dynamic sizing
- Markdown rendering for task descriptions
- Status indicators and color coding
- Scrollable content areas

## Data Flow

1. **Application Start**: Load configuration, initialize API client, fetch initial data
2. **User Input**: Keyboard events processed by key handler
3. **State Updates**: Model state updated based on user actions
4. **API Calls**: Async commands fetch data from Archon server
5. **UI Refresh**: Views re-rendered based on updated state

## Configuration

Configuration is loaded in order of preference:
1. `./config.yaml`
2. `./configs/default.yaml`
3. `~/.config/lazyarchon/config.yaml`
4. `/etc/lazyarchon/config.yaml`
5. Built-in defaults

Environment variables can override any configuration value.

## Build System

- **Makefile** - Common development tasks
- **scripts/build.sh** - Simple build script
- Go modules for dependency management

## Future Enhancements

- Task creation and editing
- Bulk task operations
- Custom keybinding configuration
- Theme customization
- Plugin system for extensions