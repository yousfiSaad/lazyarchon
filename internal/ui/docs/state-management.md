# State Management

## Overview

LazyArchon follows a **clear state ownership model** where each piece of state has exactly one owner. This document explains how state is organized and managed.

## State Hierarchy

```
┌─────────────────────────────────────────────────┐
│ Model (Global Application State)                │
│  ├── activeView: Which panel is active          │
│  ├── selectedIndex: Currently selected item     │
│  ├── sortMode: How tasks are sorted             │
│  ├── featureFilters: Feature visibility         │
│  └── searchState: Search query and matches      │
└─────────────────────────────────────────────────┘
         │
         ├─────────────────────────────────────┐
         ↓                                      ↓
┌──────────────────────────┐      ┌──────────────────────────┐
│ ProgramContext           │      │ Components               │
│  ├── Tasks []Task        │      │  Each owns local state:  │
│  ├── Projects []Project  │      │  ├── viewport position   │
│  ├── SelectedProjectID   │      │  ├── dimensions          │
│  ├── Loading bool        │      │  ├── rendered content    │
│  ├── Error string        │      │  └── UI-specific flags   │
│  └── SearchMode bool     │      └──────────────────────────┘
└──────────────────────────┘
```

## State Categories

### 1. Model State (Application-Level)

**Location**: `model.go` fields

**Purpose**: Global application state that affects multiple components

```go
type Model struct {
    // Layout state
    activeView ActiveView  // LeftPanel or RightPanel

    // Navigation state
    selectedIndex int  // Currently selected task/project index

    // Sorting/filtering state
    sortMode           int
    sortProjectID      *string
    featureFilters     map[string]bool
    // Note: statusFilters and showCompletedTasks moved to ProgramContext (user preferences)

    // Search state
    taskSearchActive    bool
    taskSearchQuery     string
    taskMatchingIndices []int
    taskTotalMatches    int

    // Feature search (for modal)
    featureSearchActive  bool
    featureSearchQuery   string
    featureSelectedIndex int
}
```

**Key Principle**: Model state is the **source of truth** for application behavior.

### 2. Program Context State

**Location**: `context/context.go`

**Purpose**: Shared runtime data and dependencies

```go
type ProgramContext struct {
    // Data
    Tasks    []archon.Task
    Projects []archon.Project

    // Selection
    SelectedProjectID *string

    // UI State
    Loading    bool
    Error      string
    SearchMode bool  // Inline search input active

    // User Preferences (persistent settings)
    SortMode           int
    StatusFilters      map[string]bool
    StatusFilterActive bool
    ShowCompletedTasks bool
    SearchHistory      []string

    // Dimensions
    Width  int
    Height int

    // Dependencies
    ArchonClient interfaces.ArchonClient
    Logger       interfaces.Logger
    Config       *config.Config
    StyleContext *styling.Context
}
```

**Key Principle**: Context holds **shared** data that many components need.

### 3. Component State

**Location**: Each component's `Model` struct

**Purpose**: Local UI state specific to that component

#### TaskList State
```go
type Model struct {
    // Data (received from parent)
    tasks []archon.Task
    sortedTasks []archon.Task

    // Selection (synced with Model)
    selectedIndex int

    // UI State (component-owned)
    viewport viewport.Model
    width int
    height int
    maxLines int

    // Search highlighting (component-owned)
    searchQuery string
    searchActive bool

    // Rendering cache
    effectiveWidth int
}
```

#### TaskDetails State
```go
type Model struct {
    // Data (received from parent)
    selectedTask *archon.Task

    // UI State (component-owned)
    panelCore *detailspanel.Core
    contentGenerator *ContentGenerator

    // Dimensions
    width int
    height int

    // Search highlighting
    searchQuery string
    searchActive bool
}
```

**Key Principle**: Components own **only what they need for rendering**.

### 4. Manager State

**Location**: `managers/*_manager.go`

**Purpose**: Business logic state

#### TaskManager
```go
type TaskManager struct {
    tasks []archon.Task
    logger interfaces.Logger
}

func (m *TaskManager) UpdateTasks(tasks []archon.Task)
func (m *TaskManager) GetTaskByID(id string) (*archon.Task, bool)
```

#### ProjectManager
```go
type ProjectManager struct {
    selectedProjectID *string
    logger interfaces.Logger
}

func (m *ProjectManager) SetSelectedProject(id *string)
func (m *ProjectManager) GetSelectedProjectID() *string
```

**Key Principle**: Managers encapsulate **business logic**, not UI logic.

## State Ownership Rules

### Rule 1: Single Source of Truth

Each piece of state has **exactly one owner**:

```go
// ✅ GOOD: Clear ownership
// Model owns selectedIndex (application level)
m.selectedIndex = 5

// TaskList owns its local selectedIndex (UI level)
taskList.selectedIndex = 5

// They stay in sync via messages

// ❌ BAD: Shared mutable state
var globalSelectedIndex int  // Who owns this?
```

### Rule 2: Parent Owns, Child Receives

Components receive data from parents, don't fetch it themselves:

```go
// ✅ GOOD: Received via message
case TaskListUpdateMsg:
    m.tasks = msg.Tasks
    m.updateSortedTasks()

// ❌ BAD: Fetching from global state
func (m *TaskList) Update() {
    m.tasks = globalState.GetTasks()  // Don't do this!
}
```

### Rule 3: State Changes via Messages

Never mutate state directly from outside:

```go
// ✅ GOOD: Send message
updateMsg := TaskListUpdateMsg{Tasks: newTasks}
m.components.Layout.MainContent.Update(updateMsg)

// ❌ BAD: Direct mutation
m.components.TaskList.tasks = newTasks  // Don't do this!
```

### Rule 4: Derived State Computed, Not Stored

Don't store what you can compute:

```go
// ✅ GOOD: Compute when needed
func (m *Model) GetSortedTasks() []archon.Task {
    return helpers.FilterAndSortTasks(
        m.programContext.Tasks,
        m.sortMode,
        m.getFilters(),
    )
}

// ❌ BAD: Store derived state
type Model struct {
    tasks []Task
    sortedTasks []Task  // Redundant! Can be computed
}
```

## State Synchronization

### Model ↔ Components

State flows in both directions:

**Downward (Model → Component)**:
```go
// Model has new data, send to component
updateMsg := tasklist.TaskListUpdateMsg{
    Tasks: m.GetSortedTasks(),
    Loading: m.programContext.Loading,
}
m.components.Layout.MainContent.Update(updateMsg)
```

**Upward (Component → Model)**:
```go
// Component selection changed, notify Model
return func() tea.Msg {
    return TaskListSelectionChangedMsg{Index: m.selectedIndex}
}

// Model receives and updates
case tasklist.TaskListSelectionChangedMsg:
    m.selectedIndex = msg.Index
    m.updateTaskDetailsComponent()
```

### Keeping State in Sync

**Pattern**: Use helper method to update both:

```go
func (m *Model) setSelectedTask(index int) tea.Cmd {
    // Update Model state
    m.selectedIndex = index

    // Update component state
    selectMsg := tasklist.TaskListSelectMsg{Index: index}
    return m.components.Layout.MainContent.Update(selectMsg)
}
```

**Pattern**: Component returns confirmation message:

```go
// TaskList updates its state
case TaskListSelectMsg:
    m.setSelectedIndex(msg.Index)

    // Confirm to Model
    return func() tea.Msg {
        return TaskListSelectionChangedMsg{Index: m.selectedIndex}
    }
```

## State Updates

### Updating Model State

```go
// Direct state update (synchronous)
m.selectedIndex = newIndex
m.sortMode = SortByPriority
m.featureFilters = newFilters

// State update with component notification
m.selectedIndex = newIndex
m.updateTaskListComponents(m.GetSortedTasks())
```

### Updating Context State

```go
// Update shared data
m.programContext.Tasks = newTasks
m.programContext.Loading = false
m.programContext.Error = ""

// Update configuration
m.programContext.UpdateSearchMode(true)
m.programContext.SetSelectedProject(projectID)
```

### Updating Component State

```go
// Via message (preferred)
updateMsg := TaskListUpdateMsg{Tasks: newTasks}
component.Update(updateMsg)

// Direct (only within component)
func (m *Model) handleInternalEvent() {
    m.selectedIndex++  // OK: component owns this
    m.updateViewportContent()
}
```

## State Persistence

### What Gets Persisted

Currently, no state is persisted between sessions. All state is runtime-only.

**Future**: Could persist:
- Sort mode
- Filter preferences
- Selected project
- Window dimensions

### How to Add Persistence

```go
// 1. Save state on change
func (m *Model) setSortMode(mode int) {
    m.sortMode = mode
    m.savePreferences()  // Save to file/config
}

// 2. Load state on startup
func NewModel(...) *Model {
    m := &Model{}
    m.loadPreferences()  // Load from file/config
    return m
}
```

## State Validation

### Bounds Checking

Always validate state changes:

```go
// ✅ GOOD: Validate before setting
func (m *Model) setSelectedIndex(index int) {
    if index < 0 || index >= len(m.tasks) {
        return  // Invalid, don't change
    }
    m.selectedIndex = index
}

// ❌ BAD: No validation
m.selectedIndex = index  // Might be out of bounds!
```

### Consistency Checks

Ensure related state stays consistent:

```go
// ✅ GOOD: Keep related state in sync
func (m *Model) clearSearch() {
    m.taskSearchActive = false
    m.taskSearchQuery = ""
    m.taskMatchingIndices = nil  // Clear related state
    m.taskTotalMatches = 0
}

// ❌ BAD: Partial update
m.taskSearchActive = false
// Forgot to clear search query and matches!
```

## State Reset

### Resetting on Mode Change

```go
func (m *Model) activateProjectMode() {
    // Reset task-specific state
    m.selectedIndex = 0
    m.clearSearch()

    // Set project mode flag
    m.programContext.SetProjectMode(true)
}
```

### Resetting on Error

```go
func (m *Model) handleError(err error) {
    m.programContext.Error = err.Error()
    m.programContext.Loading = false
    // Keep other state intact
}
```

## State Debugging

### Logging State Changes

```go
func (m *Model) setSelectedTask(index int) tea.Cmd {
    m.logger.Debug("Selection changed",
        "old", m.selectedIndex,
        "new", index,
    )
    m.selectedIndex = index
    return m.updateTaskSelectionComponents(index)
}
```

### State Inspector

Add debug view:

```go
func (m *Model) debugView() string {
    return fmt.Sprintf(`
Model State:
  selectedIndex: %d
  sortMode: %d
  searchActive: %v
  searchQuery: %s
  featureFilters: %+v
`,
        m.selectedIndex,
        m.sortMode,
        m.taskSearchActive,
        m.taskSearchQuery,
        m.featureFilters,
    )
}
```

### State Assertions

In development, assert invariants:

```go
func (m *Model) validateState() {
    if m.selectedIndex < 0 {
        panic("selectedIndex cannot be negative")
    }
    if m.selectedIndex >= len(m.GetSortedTasks()) {
        panic("selectedIndex out of bounds")
    }
    if m.taskSearchActive && m.taskSearchQuery == "" {
        panic("search active but query empty")
    }
}
```

## Common State Patterns

### Pattern: Toggle Boolean

```go
func (m *Model) toggleSearch() {
    m.taskSearchActive = !m.taskSearchActive
    if !m.taskSearchActive {
        m.clearSearchResults()
    }
}
```

### Pattern: Cycle Through Options

```go
func (m *Model) cycleSortMode() {
    m.sortMode = (m.sortMode + 1) % 4
    m.updateTaskListComponents(m.GetSortedTasks())
}
```

### Pattern: Conditional State

```go
func (m *Model) setSearchQuery(query string) {
    m.taskSearchQuery = query

    if query == "" {
        // Clear search when query empty
        m.taskSearchActive = false
        m.taskMatchingIndices = nil
        m.taskTotalMatches = 0
    } else {
        // Activate search when query present
        m.taskSearchActive = true
        m.updateSearchMatches()
    }
}
```

### Pattern: Batched State Update

```go
func (m *Model) applyFilters(features map[string]bool, statuses map[string]bool) {
    // Update all filter state together
    m.featureFilters = features
    m.statusFilters = statuses
    m.statusFilterActive = true

    // Single refresh after all updates
    m.updateTaskListComponents(m.GetSortedTasks())
}
```

## Best Practices

### 1. Minimize State

Only store what you can't compute:

```go
// ✅ GOOD: Minimal state
type Model struct {
    tasks []Task
    sortMode int
}

func (m *Model) GetSortedTasks() []Task {
    return sortTasks(m.tasks, m.sortMode)
}

// ❌ BAD: Redundant state
type Model struct {
    tasks []Task
    sortMode int
    sortedTasks []Task  // Can be computed!
}
```

### 2. Use Helpers for State Changes

```go
// ✅ GOOD: Helper ensures consistency
func (m *Model) setSelectedIndex(index int) {
    m.selectedIndex = index
    m.updateViewportContent()  // Always happens
    m.followSelection()
}

// ❌ BAD: Easy to forget steps
m.selectedIndex = index
// Forgot to update viewport!
```

### 3. Document State Dependencies

```go
// featureFilters state:
// - nil: No filter active (show all tasks)
// - {}: Filter active, nothing selected (show no tasks)
// - populated: Filter active with selections
featureFilters map[string]bool
```

### 4. Use Types for Safety

```go
// ✅ GOOD: Type-safe
type SortMode int

const (
    SortByStatusPriority SortMode = iota
    SortByPriority
    SortByTime
    SortByAlphabetical
)

// ❌ BAD: Magic numbers
sortMode := 2  // What does 2 mean?
```

## References

- [Model Definition](../model.go)
- [Program Context](../context/context.go)
- [Component State Examples](../components/)
- [Manager Pattern](../managers/)
