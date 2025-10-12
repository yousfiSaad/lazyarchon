# LazyArchon TUI Documentation

> Comprehensive documentation for the LazyArchon Terminal User Interface

## Table of Contents

1. [Architecture Overview](./architecture.md) - System design and patterns
2. [Component System](./components.md) - Component hierarchy and lifecycle
3. [Message Flow](./messages.md) - How messages propagate through the system
4. [State Management](./state-management.md) - State organization and updates
5. [Input Handling](./input-handling.md) - Keyboard input routing and handlers
6. [Modal System](./modals.md) - Modal components and workflows
7. [Navigation System](./navigation.md) - Task/project navigation
8. [Search System](./search.md) - Search functionality and match navigation
9. [Styling System](./styling.md) - Theme and style management

## Quick Start

### Understanding the Codebase

LazyArchon TUI is built on the **Bubble Tea framework**, following the Elm Architecture:
- **Model**: Application state
- **Update**: Message handlers that transform state
- **View**: Renders state to terminal output

### Key Directories

```
internal/ui/
├── model.go              # Main application model
├── components/           # Reusable UI components
│   ├── layout/          # Header, status bar, main content
│   ├── modals/          # Modal dialogs
│   ├── tasklist/        # Task list component
│   ├── taskdetails/     # Task details panel
│   └── projectlist/     # Project list component
├── input_handlers*.go    # Keyboard input handlers (by feature)
├── model_handlers_*.go   # Message handlers (by category)
├── context/             # Shared context and dependencies
├── factories/           # Component and manager creation
├── managers/            # Business logic managers
├── helpers/             # Pure utility functions
├── messages/            # Custom message types
├── commands/            # Command constructors
└── styling/             # Theme and style providers
```

### Core Patterns

1. **Message-Based Architecture**: All state changes happen through messages
2. **Component Composition**: UI built from composable, reusable components
3. **Single Source of Truth**: State lives in one place (Model or component)
4. **Pure Functions**: Helpers and utilities are side-effect free
5. **Interface-Based Design**: Dependencies injected via interfaces

### Common Operations

#### Adding a New Modal
See [Modal System](./modals.md#creating-a-new-modal)

#### Adding a New Keyboard Shortcut
See [Input Handling](./input-handling.md#adding-new-shortcuts)

#### Creating a New Component
See [Component System](./components.md#creating-components)

#### Adding State to the Model
See [State Management](./state-management.md#adding-state)

## Architecture Principles

### 1. Component-Based Architecture

Every UI element is a component with:
- **State**: Internal component state
- **Update**: Message handler
- **View**: Rendering function
- **Messages**: Custom message types for communication

### 2. Message Flow

```
User Input → HandleKeyPress → Handler → Message → Update → View
                                           ↓
                                      Components
```

### 3. State Ownership

- **Model**: Global application state (selected index, sort mode, filters)
- **Components**: Local UI state (scroll position, dimensions, viewport)
- **Context**: Shared dependencies (API client, logger, config)

### 4. Dependency Injection

All dependencies flow through:
- **ProgramContext**: Runtime state and API clients
- **ManagerSet**: Business logic managers
- **UIComponentSet**: UI component tree

## Getting Help

- For architecture questions, see [Architecture Overview](./architecture.md)
- For component questions, see [Component System](./components.md)
- For message routing questions, see [Message Flow](./messages.md)
- For styling questions, see [Styling System](./styling.md)

## Contributing

When adding new features:

1. ✅ Follow existing patterns (component-based, message-driven)
2. ✅ Add messages for state changes (no direct mutations in handlers)
3. ✅ Use helpers for pure logic (keep handlers focused)
4. ✅ Document complex flows (comments explaining "why")
5. ✅ Test component isolation (components should work standalone)

## Recent Improvements

- **Input Handler Organization**: Handlers split by feature area (navigation, search, task operations)
- **Helper Consolidation**: Removed dead code, organized utilities
- **Selection Change Encapsulation**: Generic `setSelectedIndex()` prevents cursor update bugs
- **Message Handler Extraction**: Model handlers organized by message category
