// Package ui implements the Terminal User Interface (TUI) for LazyArchon using Bubble Tea.
//
// This package provides a comprehensive, keyboard-driven terminal interface for managing
// tasks and projects in the Archon system. It follows the Elm architecture pattern
// with Model-View-Update (MVU) for predictable state management.
//
// # Core Architecture
//
// The UI is built around several key components:
//
//   - Model: Central application state and data management
//   - Views: Rendering functions that convert state to terminal output
//   - Input handlers: Keyboard event processing and navigation
//   - Commands: Asynchronous operations for API calls and real-time updates
//   - Styling: Consistent visual presentation using Lip Gloss
//
// # Model Structure
//
// The Model struct uses modern component-based architecture with coordinators
// and managers for separation of concerns:
//
//	type Model struct {
//		// Core infrastructure
//		wsClient       interfaces.RealtimeClient
//		programContext *context.ProgramContext
//
//		// Component-based UI (8 modal/layout components)
//		helpModalComponent         base.Component
//		statusModalComponent       base.Component
//		confirmationModalComponent base.Component
//		taskEditModalComponent     base.Component
//		featureModalComponent      base.Component
//		headerComponent            base.Component
//		statusBarComponent         base.Component
//		layoutViewComponent        base.Component
//
//		// Coordinators - embedded as values for compile-time safety
//		coordinators factories.CoordinatorSet  // Layout, Navigation, Sorting, Search, Feature, CoreState
//
//		// Managers - embedded as values for compile-time safety
//		managers factories.ManagerSet          // Task, Project, UIUtilities
//	}
//
// The coordinators handle UI state (navigation, layout, search, features),
// while managers handle business logic (tasks, projects) and UI utilities.
// All shared state and dependencies are centralized in ProgramContext.
//
// # Key Features
//
// ## Task Management
//   - View tasks filtered by project or show all tasks
//   - Sort by status+priority, priority, creation time, or alphabetical
//   - Task detail view with scrolling support
//   - Real-time status updates via WebSocket
//   - Task status changes and feature assignment
//
// ## Navigation
//   - Vim-like keyboard shortcuts (j/k, g/G, Ctrl+D/Ctrl+U)
//   - Panel switching with Tab key
//   - Modal navigation with consistent patterns
//   - Search functionality with n/N navigation
//
// ## Project Management
//   - Project selection mode (p key)
//   - Filter tasks by specific projects
//   - Project-aware task loading and display
//
// ## Real-time Features
//   - WebSocket connection for live updates
//   - Connection status indicator
//   - Automatic UI refresh on remote changes
//   - Graceful fallback to HTTP polling if WebSocket fails
//
// # Usage Example
//
// Creating and running the TUI application:
//
//	cfg, err := config.Load()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	model := ui.NewModel(cfg)
//	program := tea.NewProgram(model, tea.WithAltScreen())
//	if err := program.Start(); err != nil {
//		log.Fatal(err)
//	}
//
// # Keyboard Controls
//
// The interface supports extensive keyboard navigation:
//
//   - Navigation: j/k (up/down), g/G (first/last), Ctrl+D/Ctrl+U (page up/down)
//   - Actions: Enter (view details), s (change status), e (edit feature)
//   - Modes: p (project mode), / (search), ? (help)
//   - General: r (refresh), q (quit), Tab (switch panels)
//
// # State Management
//
// The application follows the Elm architecture with unidirectional data flow:
//
//  1. User Input -> Update function
//  2. Update function -> New state + Commands
//  3. Commands -> Async operations (API calls, WebSocket events)
//  4. Command results -> Update function (via messages)
//  5. New state -> View function -> Terminal output
//
// # Architecture Pattern
//
// The UI follows a coordinator-based architecture:
//
//   - ProgramContext: Centralized shared state and dependencies
//   - Coordinators: UI state management (navigation, layout, search, sorting, features, core state)
//   - Managers: Business logic (tasks, projects) and UI utilities (spinner, scrolling)
//   - Components: Self-contained UI elements with Update/View/Init methods
//   - Message Router: Routes messages between components and coordinators
//
// This architecture provides:
//
//   - Compile-time nil safety (coordinators/managers embedded as values)
//   - Clear separation of concerns (UI state vs business logic)
//   - Single source of truth (ProgramContext for shared state)
//   - Testability (components and coordinators can be tested independently)
//
// # Styling and Theming
//
// The UI uses Lip Gloss for consistent styling with support for:
//
//   - Multiple color schemes (blue, gray, warm_gray, cool_gray)
//   - Terminal capability detection and graceful degradation
//   - Responsive layout based on terminal size
//   - Consistent component styling across the application
//
// # Testing
//
// The package includes comprehensive test coverage:
//
//   - Unit tests for individual functions and components
//   - Integration tests for user workflows
//   - Performance tests for large datasets
//   - Mock implementations for external dependencies
//
// # Performance Optimizations
//
// The UI is optimized for large datasets and responsive interaction:
//
//   - Efficient task sorting with caching
//   - Minimal re-renders through careful state management
//   - Viewport-based rendering for large lists
//   - Background API operations that don't block the UI
//
// For detailed documentation of specific components, see the individual
// type and function documentation below.
package ui
