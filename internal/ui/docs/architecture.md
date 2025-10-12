# Architecture Overview

## System Design

LazyArchon TUI follows a **component-based, message-driven architecture** built on the Bubble Tea framework.

## Core Architecture Patterns

### 1. Elm Architecture (TEA - The Elm Architecture)

```
┌──────────────────────────────────────────────────────┐
│                    User Input                         │
└────────────────────┬─────────────────────────────────┘
                     │
                     ↓
┌──────────────────────────────────────────────────────┐
│                    Update                             │
│  • Handle messages                                    │
│  • Transform state                                    │
│  • Return new state + commands                        │
└────────────────────┬─────────────────────────────────┘
                     │
                     ↓
┌──────────────────────────────────────────────────────┐
│                    Model (State)                      │
│  • Application state                                  │
│  • Component tree                                     │
│  • Managers and context                               │
└────────────────────┬─────────────────────────────────┘
                     │
                     ↓
┌──────────────────────────────────────────────────────┐
│                    View                               │
│  • Render state to string                            │
│  • Compose component views                           │
│  • Apply styles                                       │
└──────────────────────────────────────────────────────┘
```

### 2. Component Hierarchy

```
Model (root)
├── Components (UIComponentSet)
│   ├── Layout
│   │   ├── Header
│   │   ├── MainContent
│   │   │   ├── TaskList (left panel)
│   │   │   ├── TaskDetails (right panel)
│   │   │   ├── ProjectList (project mode)
│   │   │   └── ProjectDetails (project mode)
│   │   └── StatusBar
│   └── Modals
│       ├── Help
│       ├── Confirmation
│       ├── TaskEdit
│       ├── Feature
│       ├── Status
│       └── StatusFilter
├── Managers (ManagerSet)
│   ├── Task
│   ├── Project
│   └── UIUtilities
└── Context (ProgramContext)
    ├── Tasks, Projects (data)
    ├── ArchonClient (API)
    ├── Config, Logger
    └── UI state flags
```

### 3. State Organization

#### Model State (Global)
Located in `model.go`:

```go
type Model struct {
    // Infrastructure
    wsClient interfaces.RealtimeClient
    programContext *context.ProgramContext
    components factories.UIComponentSet
    managers factories.ManagerSet

    // Direct State (migrated from coordinators)
    activeView ActiveView           // Which panel is active
    selectedIndex int                // Currently selected task/project
    sortMode int                     // Current sort mode
    featureFilters map[string]bool   // Feature visibility
    taskSearchActive bool            // Search state
    taskSearchQuery string
    taskMatchingIndices []int
    // ... more state fields
}
```

**Design Decision**: State was migrated from "coordinator" pattern to direct Model fields following lazygit/gh-dash patterns for simplicity.

#### Component State (Local)
Each component owns its UI state:

```go
// TaskList component
type Model struct {
    tasks []archon.Task
    selectedIndex int      // Which task is selected
    viewport viewport.Model  // Scroll position
    searchQuery string     // Local search state
    // ... component-specific state
}
```

**Principle**: Components manage their own rendering state; Model manages global application state.

### 4. Message Flow Architecture

#### Priority-Based Routing

Messages flow through a priority system:

```
1. Global Keys (Ctrl+C, ?) → Always handled first
2. Search Input Mode → Captures all typing
3. Modal Keys → When modal is active
4. Application Keys (p, a, r, q) → Work across all modes
5. Mode-Specific Keys → Navigation, search, task ops
```

#### Message Types

**1. Input Messages** (`tea.KeyMsg`)
- Raw keyboard input
- Routed by `HandleKeyPress` in `input_handlers.go`

**2. Command Messages** (from `commands/`)
- `TasksLoadedMsg` - Tasks fetched from API
- `ProjectModeActivatedMsg` - Enter project selection
- `RefreshDataMsg` - Refresh data from API

**3. Component Messages** (from `components/*/messages.go`)
- `TaskListScrollMsg` - Scroll task list
- `TaskListSelectMsg` - Select specific task
- `TaskDetailsUpdateMsg` - Update task details panel

**4. Modal Messages** (from `components/modals/*/messages.go`)
- `ShowHelpModalMsg` - Show help modal
- `TaskPropertiesUpdatedMsg` - Task was edited
- `FeatureSelectionAppliedMsg` - Features were filtered

**5. Realtime Messages** (from `commands/`)
- `RealtimeTaskUpdateMsg` - Task updated via WebSocket
- `RealtimeConnectedMsg` - WebSocket connected

### 5. Update Flow

```go
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // 1. Type switch on message
    switch msg := msg.(type) {

    case tea.KeyMsg:
        return m.handleKeyInput(msg)

    case commands.TasksLoadedMsg:
        return m.handleTaskMessages(msg)

    case tasklist.TaskListSelectionChangedMsg:
        return m.handleComponentMessages(msg)

    // ... more message handlers
    }

    // 2. Fallback: broadcast to component tree
    return m, m.components.Update(msg)
}
```

**Key Points**:
- Model handles application-level messages
- Components handle their own messages
- Unknown messages broadcast to all components
- Commands returned for async operations

### 6. Command Pattern

Commands represent **future work** (async operations):

```go
// Command to load tasks
func LoadTasks(client interfaces.ArchonClient) tea.Cmd {
    return func() tea.Msg {
        tasks, err := client.GetTasks()
        if err != nil {
            return commands.ErrorMsg{Error: err}
        }
        return commands.TasksLoadedMsg{Tasks: tasks}
    }
}
```

**Usage**:
```go
// In Update, return command
return m, LoadTasks(m.programContext.ArchonClient)

// Command executes asynchronously
// Result comes back as message in future Update call
```

## Design Decisions

### Why Component-Based?

**Before**: Monolithic model with mixed responsibilities
**After**: Focused components with clear boundaries

**Benefits**:
1. **Testability**: Components can be tested in isolation
2. **Reusability**: Components can be used in different contexts
3. **Maintainability**: Changes localized to components
4. **Clarity**: Each component has single responsibility

### Why Message-Driven?

**Alternatives Considered**:
- Direct method calls
- Callback functions
- Event emitters

**Why Messages Won**:
1. **Traceable**: All state changes go through Update
2. **Debuggable**: Can log all messages
3. **Testable**: Easy to test message handlers
4. **Composable**: Messages naturally compose

### Why Direct State (Not Coordinators)?

**Before**: Coordinators (NavigationCoordinator, SortingCoordinator, SearchCoordinator)

**Problems**:
- Extra indirection
- Unclear ownership
- More complexity than benefit

**After**: Direct state fields on Model

**Benefits**:
- Simpler mental model
- Follows lazygit/gh-dash patterns
- Less code to understand
- Clearer state ownership

### Why Helper Functions?

**Pattern**: Pure functions in `helpers/` package

**Purpose**:
1. **Testability**: Pure functions easy to test
2. **Reusability**: Used across components
3. **Clarity**: Business logic separate from UI logic

**Example**:
```go
// Pure function - no side effects
func FilterAndSortTasks(tasks []archon.Task, sortMode int, filters TaskFilters) []archon.Task {
    // ... filtering and sorting logic
    return sortedTasks
}

// Used in Model
func (m *Model) GetSortedTasks() []archon.Task {
    filters := helpers.TaskFilters{
        ProjectID: m.sortProjectID,
        // ... other filters
    }
    return helpers.FilterAndSortTasks(m.programContext.Tasks, m.sortMode, filters)
}
```

## Performance Considerations

### 1. Viewport Optimization

Only render visible lines:

```go
// TaskList only renders tasks in viewport
func (m *Model) updateViewportContent() {
    // Only renders tasks that fit in viewport height
    // Viewport handles scrolling without re-rendering
}
```

### 2. Message Batching

Combine multiple commands:

```go
return m, tea.Batch(
    command1,
    command2,
    command3,
)
```

### 3. Conditional Rendering

Skip rendering when not visible:

```go
func (m *Model) View() string {
    if !m.isVisible {
        return ""
    }
    // ... render
}
```

## Error Handling Strategy

### 1. Error Messages

Errors become messages:

```go
type ErrorMsg struct {
    Error error
    Context string
}
```

### 2. Error State

Model tracks error state:

```go
m.programContext.Error = "Failed to load tasks"
m.programContext.Loading = false
```

### 3. User Feedback

Errors shown in UI:

```go
// Status bar shows errors
if m.programContext.Error != "" {
    return m.renderError()
}
```

## Testing Strategy

### 1. Component Tests

Test components in isolation:

```go
func TestTaskListSelection(t *testing.T) {
    model := NewModel(/* test dependencies */)

    msg := TaskListSelectMsg{Index: 5}
    cmd := model.Update(msg)

    assert.Equal(t, 5, model.selectedIndex)
}
```

### 2. Message Handler Tests

Test message handling:

```go
func TestHandleTasksLoaded(t *testing.T) {
    model := NewModel()

    msg := commands.TasksLoadedMsg{Tasks: testTasks}
    newModel, _ := model.Update(msg)

    assert.Len(t, newModel.programContext.Tasks, 10)
}
```

### 3. Integration Tests

Test full workflows:

```go
func TestSearchWorkflow(t *testing.T) {
    model := NewModel()

    // Activate search
    model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})

    // Type query
    model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})

    // Navigate matches
    model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

    // Verify state
    assert.True(t, model.taskSearchActive)
}
```

## Migration Path

### From Old Architecture

The codebase recently migrated from:

**Coordinator Pattern** → **Direct State**
- Removed NavigationCoordinator, SortingCoordinator, SearchCoordinator
- Moved state directly to Model
- Simplified mental model

**Monolithic Files** → **Organized by Feature**
- Split `input_handlers.go` into feature files
- Extracted message handlers from `model.go`
- Better code organization

**Dead Code** → **Clean Codebase**
- Removed unused modal wrappers
- Cleaned up utility managers
- Consolidated helpers

## References

- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)
- [Elm Architecture](https://guide.elm-lang.org/architecture/)
- [lazygit Architecture](https://github.com/jesseduffield/lazygit) (inspiration)
- [gh-dash Architecture](https://github.com/dlvhdr/gh-dash) (inspiration)
