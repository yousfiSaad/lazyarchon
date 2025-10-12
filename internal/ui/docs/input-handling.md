# Input Handling

## Overview

LazyArchon uses a **priority-based routing system** for keyboard input. All keyboard input flows through a single entry point and is routed based on priority and context.

## Input Flow

```
User Keyboard
     ↓
tea.KeyMsg
     ↓
HandleKeyPress (priority router)
     ↓
Specific Handler (by feature)
     ↓
Message/Command
     ↓
Component/Model Update
```

## Priority System

Input is processed in **strict priority order**:

```go
func (m *Model) HandleKeyPress(key string) tea.Cmd {
    // PRIORITY 1: Emergency keys (always work)
    if cmd, handled := m.handleGlobalKeys(key); handled {
        return cmd
    }

    // PRIORITY 2: Search input mode (captures typing)
    if m.programContext.SearchMode {
        return m.handleInlineSearchInput(key)
    }

    // PRIORITY 3: Modal keys (when modal is active)
    if m.HasActiveModal() {
        if cmd, handled := m.routeToActiveModal(key); handled {
            return cmd
        }
    }

    // PRIORITY 4: Application keys (work everywhere)
    if cmd, handled := m.handleApplicationKey(key); handled {
        return cmd
    }

    // PRIORITY 5: Mode-specific keys
    if m.programContext.IsProjectModeActive() {
        return m.handleProjectModeKeys(key)
    } else {
        return m.handleTaskModeKeys(key)
    }
}
```

**Why Priority Matters**:
- Ensures Ctrl+C always quits, even when modal is open
- Search mode can capture all typing without conflicts
- Modals can override normal key behavior
- Application keys work consistently across all modes

## Handler Organization

Handlers are organized by **feature area** in separate files:

### Navigation Handlers (`input_handlers_navigation.go`)

**Responsibility**: Move between tasks/projects, scroll, jump

**Handlers** (10 total):
```go
HandleUpNavigationKey(key)           // k, ↑
HandleDownNavigationKey(key)         // j, ↓
HandleLeftNavigationKey(key)         // h
HandleRightNavigationKey(key)        // l
HandleJumpToFirstKey(key)            // gg, Home
HandleJumpToLastKey(key)             // G, End
HandleFastScrollUpKey(key)           // K
HandleFastScrollDownKey(key)         // J
HandleHalfPageUpKey(key)             // Ctrl+U, PgUp
HandleHalfPageDownKey(key)           // Ctrl+D, PgDn
```

**Pattern**: Navigation handlers send scroll messages to components

```go
func (m *Model) HandleDownNavigationKey(key string) (tea.Cmd, bool) {
    cmd := m.handleDownNavigation()
    return cmd, true
}

func (m *Model) handleDownNavigation() tea.Cmd {
    if m.programContext.IsProjectModeActive() {
        scrollMsg := projectlist.ProjectListScrollMsg{Direction: ScrollDown}
        return m.components.Layout.MainContent.Update(scrollMsg)
    } else {
        if m.IsLeftPanelActive() {
            scrollMsg := tasklist.TaskListScrollMsg{Direction: ScrollDown}
            return m.components.Layout.MainContent.Update(scrollMsg)
        } else if m.IsRightPanelActive() {
            scrollMsg := taskdetails.TaskDetailsScrollMsg{Direction: ScrollDown}
            return m.components.Layout.MainContent.Update(scrollMsg)
        }
    }
    return nil
}
```

### Search Handlers (`input_handlers_search.go`)

**Responsibility**: Activate search, navigate matches, clear search

**Handlers** (4 total):
```go
HandleActivateSearchKey(key)      // /, Ctrl+F
HandleClearSearchKey(key)         // Ctrl+X, Ctrl+L
HandleNextSearchMatchKey(key)     // n
HandlePrevSearchMatchKey(key)     // N
```

**Pattern**: Search handlers manage search state

```go
func (m *Model) HandleActivateSearchKey(key string) (tea.Cmd, bool) {
    if !m.programContext.IsProjectModeActive() && !m.programContext.SearchMode {
        m.programContext.SetSearchMode(true)
        m.programContext.UpdateSearchInput(m.taskSearchQuery)
        return nil, true
    }
    return nil, false
}

func (m *Model) HandleNextSearchMatchKey(key string) (tea.Cmd, bool) {
    if !m.programContext.IsProjectModeActive() && m.taskSearchActive && m.taskTotalMatches > 0 {
        cmd := m.nextSearchMatch()
        return cmd, true
    }
    return nil, false
}
```

### Task Operation Handlers (`input_handlers_task.go`)

**Responsibility**: Task actions (edit, copy, filter, sort)

**Handlers** (7 total):
```go
HandleTaskStatusChangeKey(key)    // t
HandleTaskEditKey(key)            // e
HandleTaskIDCopyKey(key)          // y
HandleTaskTitleCopyKey(key)       // Y
HandleFeatureSelectionKey(key)    // f
HandleSortModeKey(key)            // s
HandleSortModePreviousKey(key)    // S
```

**Pattern**: Task handlers show modals or change modes

```go
func (m *Model) HandleTaskStatusChangeKey(key string) (tea.Cmd, bool) {
    if key == keys.KeyT && !m.programContext.IsProjectModeActive() && len(m.programContext.Tasks) > 0 {
        selectedTask := m.GetSortedTasks()[m.selectedIndex]

        return func() tea.Msg {
            return taskedit.ShowTaskEditModalMsg{
                TaskID: selectedTask.ID,
                CurrentStatus: selectedTask.Status,
                FocusField: taskedit.FieldStatus,
                // ... more fields
            }
        }, true
    }
    return nil, false
}
```

### Application Handlers (`input_handlers.go`)

**Responsibility**: Cross-mode operations (quit, refresh, help)

**Handlers** (8 total):
```go
HandleQuitKey(key)              // q
HandleEmergencyQuitKey(key)     // Ctrl+C
HandleRefreshKey(key)           // r, F5
HandleProjectModeKey(key)       // p
HandleShowAllTasksKey(key)      // a
HandleEscapeKey(key)            // Esc
HandleConfirmKey(key)           // Enter
HandleToggleHelpKey(key)        // ?
```

## Key Mappings

### Global Keys (Priority 1)

| Key | Action | Works When |
|-----|--------|------------|
| Ctrl+C | Emergency quit | Always |
| ? | Toggle help | Always |

### Search Input Mode (Priority 2)

When search mode is active (`/` pressed):

| Key | Action |
|-----|--------|
| [a-z0-9] | Type search query |
| Backspace | Delete character |
| Ctrl+U | Clear query |
| Enter | Commit search |
| Esc | Cancel search |

### Modal Keys (Priority 3)

Handled by active modal component. Examples:

**Help Modal**:
- ↑/↓, j/k: Scroll help
- q, Esc: Close help

**TaskEdit Modal**:
- Tab: Next field
- Shift+Tab: Previous field
- Enter: Confirm changes
- Esc, q: Cancel

**Feature Modal**:
- ↑/↓, j/k: Navigate features
- Space: Toggle feature
- Enter: Apply selection
- a: Select all
- c: Clear all

### Application Keys (Priority 4)

| Key | Action | Mode |
|-----|--------|------|
| q | Quit (with confirmation) | All |
| r, F5 | Refresh data | All |
| p | Toggle project mode | All |
| a | Show all tasks | All |
| Esc | Cancel/Go back | All |
| Enter | Confirm/Select | All |

### Task Mode Keys (Priority 5)

**Navigation**:
| Key | Action |
|-----|--------|
| j, ↓ | Move down |
| k, ↑ | Move up |
| h | Focus left panel |
| l | Focus right panel |
| gg, Home | Jump to first |
| G, End | Jump to last |
| J | Fast scroll down (4 lines) |
| K | Fast scroll up (4 lines) |
| Ctrl+D, PgDn | Half-page down |
| Ctrl+U, PgUp | Half-page up |

**Search**:
| Key | Action |
|-----|--------|
| /, Ctrl+F | Activate search |
| n | Next match |
| N | Previous match |
| Ctrl+X, Ctrl+L | Clear search |

**Task Operations**:
| Key | Action |
|-----|--------|
| t | Change status |
| e | Edit properties |
| y | Copy task ID |
| Y | Copy task title |
| f | Filter by feature |
| s | Cycle sort forward |
| S | Cycle sort backward |

### Project Mode Keys (Priority 5)

| Key | Action |
|-----|--------|
| j, ↓ | Move down in projects |
| k, ↑ | Move up in projects |
| h | Focus left panel |
| l | Focus right panel |
| Enter, l | Select project |
| q, Esc, h | Exit project mode |
| gg, Home | Jump to first project |
| G, End | Jump to last project |
| y | Copy project ID |
| Y | Copy project title |

## Adding New Shortcuts

### Step 1: Choose Handler File

Based on feature area:
- Navigation → `input_handlers_navigation.go`
- Search → `input_handlers_search.go`
- Task operations → `input_handlers_task.go`
- Application-wide → `input_handlers.go`

### Step 2: Define Key Constant

```go
// internal/shared/utils/keys/keys.go
const (
    KeyMyNewKey = "m"
)
```

### Step 3: Create Handler Function

```go
// Example: input_handlers_task.go
func (m *Model) HandleMyNewActionKey(key string) (tea.Cmd, bool) {
    if key == keys.KeyMyNewKey && !m.programContext.IsProjectModeActive() {
        // Check preconditions
        if len(m.programContext.Tasks) == 0 {
            return nil, false
        }

        // Perform action
        // ... do something ...

        return someCommand, true
    }
    return nil, false
}
```

### Step 4: Wire Up in Dispatcher

```go
// input_handlers.go
func (m *Model) handleTaskKey(key string) (tea.Cmd, bool) {
    switch key {
    // ... existing cases ...
    case keys.KeyMyNewKey:
        return m.HandleMyNewActionKey(key)
    default:
        return nil, false
    }
}
```

### Step 5: Document in Help

```go
// components/modals/help/content.go
const helpContent = `
...
Task Operations:
  m    My new action
...
`
```

## Handler Patterns

### Pattern 1: Simple Action

**No conditions, just do it**:

```go
func (m *Model) HandleRefreshKey(key string) (tea.Cmd, bool) {
    m.SetLoadingWithMessage(true, "Refreshing...")
    return commands.RefreshData(m.programContext.ArchonClient), true
}
```

### Pattern 2: Conditional Action

**Check state before acting**:

```go
func (m *Model) HandleNextSearchMatchKey(key string) (tea.Cmd, bool) {
    // Check all conditions
    if !m.programContext.IsProjectModeActive() &&
       m.taskSearchActive &&
       m.taskTotalMatches > 0 {

        cmd := m.nextSearchMatch()
        return cmd, true
    }
    return nil, false
}
```

### Pattern 3: Mode-Dependent Action

**Different behavior based on mode**:

```go
func (m *Model) HandleLeftNavigationKey(key string) (tea.Cmd, bool) {
    if m.programContext.IsProjectModeActive() {
        // In project mode: go back
        cmd := func() tea.Msg {
            return commands.ProjectModeDeactivatedMsg{ShouldLoadTasks: false}
        }
        return cmd, true
    } else {
        // In task mode: switch to left panel
        cmd := m.SetActiveView(LeftPanel)
        return cmd, true
    }
}
```

### Pattern 4: Show Modal

**Open modal with data**:

```go
func (m *Model) HandleFeatureSelectionKey(key string) (tea.Cmd, bool) {
    if key == keys.KeyF && !m.programContext.IsProjectModeActive() {
        // Prepare modal data
        selectedFeatures := m.featureFilters
        if selectedFeatures == nil {
            selectedFeatures = make(map[string]bool)
            for _, feature := range m.GetUniqueFeatures() {
                selectedFeatures[feature] = true
            }
        }

        // Show modal
        showMsg := feature.ShowFeatureModalMsg{
            AllFeatures: m.GetUniqueFeatures(),
            SelectedFeatures: selectedFeatures,
        }
        return func() tea.Msg { return showMsg }, true
    }
    return nil, false
}
```

### Pattern 5: Smart Behavior

**Context-aware actions**:

```go
func (m *Model) HandleQuitKey(key string) (tea.Cmd, bool) {
    // If search active, close search first
    if m.programContext.SearchMode {
        m.programContext.SetSearchMode(false)
        return nil, true
    }

    // If modal active, close modal first
    if m.HasActiveModal() {
        // Close appropriate modal
        // ...
        return nil, true
    }

    // Otherwise, show quit confirmation
    return m.ShowQuitConfirmation(), true
}
```

## Input Routing Details

### Multi-Key Sequences

**Example**: `gg` to jump to first task

```go
func (m *Model) handleMultiKeySequence(key string) (tea.Cmd, bool) {
    if key == keys.KeyG {
        if m.lastKey == keys.KeyG {
            // Second 'g' - execute gg command
            cmd := m.handleJumpToFirst()
            m.lastKey = ""
            return cmd, true
        }
        // First 'g' - wait for second
        m.lastKey = keys.KeyG
        return nil, true
    }
    m.lastKey = ""
    return nil, false
}
```

### Modal Input Delegation

**Modals handle their own input**:

```go
// Model routes to modal
func (m *Model) routeToActiveModal(key string) (tea.Cmd, bool) {
    if m.components.Modals.Help.IsActive() {
        // Help modal handles key internally
        return nil, false  // Let component handle it
    }
    // ... other modals
}

// Modal handles input
func (modal *HelpModal) Update(msg tea.Msg) tea.Cmd {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        return modal.handleKeyPress(msg)
    }
}
```

### Panel-Aware Routing

**Actions depend on active panel**:

```go
func (m *Model) handleDownNavigation() tea.Cmd {
    if m.IsLeftPanelActive() {
        // Scroll task list
        return m.components.Layout.MainContent.Update(
            tasklist.TaskListScrollMsg{Direction: ScrollDown})
    } else if m.IsRightPanelActive() {
        // Scroll task details
        return m.components.Layout.MainContent.Update(
            taskdetails.TaskDetailsScrollMsg{Direction: ScrollDown})
    }
    return nil
}
```

## Testing Input Handlers

### Unit Test Handler

```go
func TestHandleDownNavigationKey(t *testing.T) {
    m := NewTestModel()
    m.selectedIndex = 5

    cmd, handled := m.HandleDownNavigationKey("j")

    assert.True(t, handled)
    assert.NotNil(t, cmd)
}
```

### Test Handler Conditions

```go
func TestNextSearchMatch_OnlyWhenSearchActive(t *testing.T) {
    m := NewTestModel()

    // Should not work when search inactive
    cmd, handled := m.HandleNextSearchMatchKey("n")
    assert.False(t, handled)

    // Activate search
    m.taskSearchActive = true
    m.taskTotalMatches = 5

    // Should work now
    cmd, handled = m.HandleNextSearchMatchKey("n")
    assert.True(t, handled)
}
```

### Test Priority

```go
func TestGlobalKeysOverrideEverything(t *testing.T) {
    m := NewTestModel()
    m.showModal(HelpModal)  // Modal active

    // Ctrl+C should still quit
    cmd := m.HandleKeyPress("ctrl+c")

    quitMsg := cmd()
    assert.IsType(t, tea.QuitMsg{}, quitMsg)
}
```

## Best Practices

### 1. Return Early

```go
// ✅ GOOD: Early return for invalid conditions
func (m *Model) HandleSomeKey(key string) (tea.Cmd, bool) {
    if m.programContext.IsProjectModeActive() {
        return nil, false  // Not applicable
    }
    if len(m.tasks) == 0 {
        return nil, false  // Nothing to act on
    }

    // Main logic here
    // ...
    return cmd, true
}
```

### 2. Check Preconditions

```go
// ✅ GOOD: Explicit precondition checks
if key == keys.KeyT &&
   !m.programContext.IsProjectModeActive() &&
   len(m.programContext.Tasks) > 0 &&
   m.selectedIndex < len(m.GetSortedTasks()) {
    // Handle key
}
```

### 3. Return (Cmd, Bool)

```go
// ✅ GOOD: Consistent return pattern
func (m *Model) HandleKey(key string) (tea.Cmd, bool) {
    if canHandle {
        return someCmd, true   // Handled
    }
    return nil, false          // Not handled
}
```

### 4. Delegate to Components

```go
// ✅ GOOD: Send message to component
scrollMsg := tasklist.TaskListScrollMsg{Direction: ScrollDown}
return m.components.Layout.MainContent.Update(scrollMsg), true

// ❌ BAD: Direct component manipulation
m.components.TaskList.selectedIndex++
m.components.TaskList.updateViewport()
```

### 5. Document Handler Purpose

```go
// HandleTaskStatusChangeKey handles 't' key - opens task properties modal
// focused on the status field for quick status changes.
func (m *Model) HandleTaskStatusChangeKey(key string) (tea.Cmd, bool) {
    // ...
}
```

## References

- [Input Handler Files](../input_handlers*.go)
- [Key Constants](../../shared/utils/keys/keys.go)
- [Help Documentation](../components/modals/help/content.go)
- [Bubble Tea Key Messages](https://github.com/charmbracelet/bubbletea#key-messages)
