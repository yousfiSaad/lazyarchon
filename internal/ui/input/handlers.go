package input

// This file serves as the main documentation and overview of the input system modularization.
//
// The input package provides a clean separation of concerns for handling different types
// of keyboard input in the LazyArchon TUI application. Each module focuses on a specific
// category of input handling:
//
// - package.go: Core interfaces and key classification helpers
// - modal_*.go: Modal-specific input key mappings and validation
// - search.go: Search input handling and navigation
// - navigation.go: Movement and scrolling commands
// - task_operations.go: Task manipulation commands
// - application.go: Application-level commands (quit, refresh, etc.)
//
// This modular approach provides several benefits:
// 1. Single Responsibility: Each file handles one type of input
// 2. Easy Testing: Individual modules can be tested in isolation
// 3. Clear Documentation: Input behavior is documented per category
// 4. Maintainable: Changes to specific input types are localized
// 5. Extensible: New input categories can be added without affecting existing code
//
// Integration Pattern:
// The actual input handling implementations remain in the ui package as methods
// on the Model struct, since they need access to private fields. This package
// provides the key mapping structures and validation functions that the ui
// package methods can use to determine what actions to take.
//
// Usage Example:
//   if input.IsNavigationKey(key) {
//       action := input.GetNavigationAction(key)
//       return m.handleNavigationAction(action)
//   }
//
// This approach respects Go's encapsulation while providing clean modularization.