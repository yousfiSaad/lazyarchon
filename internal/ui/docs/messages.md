# Message Flow

## Overview

LazyArchon uses a **message-driven architecture** where all state changes happen through messages. This document explains how messages flow through the system.

## Message Categories

### 1. Input Messages

**Source**: Bubble Tea runtime from user keyboard input

**Type**: `tea.KeyMsg`

**Flow**:
```
User Keyboard → tea.KeyMsg → HandleKeyPress → Specific Handler → Message/Command
```

**Example**:
```go
case tea.KeyMsg:
    // User pressed 'j' key
    key := msg.String()
    return m.handleKeyInput(key)
```

### 2. Command Messages

**Source**: Async command execution (API calls, timers)

**Location**: `commands/` package

**Types**:
- `TasksLoadedMsg` - Tasks fetched from API
- `TaskUpdateMsg` - Single task updated
- `ProjectsLoadedMsg` - Projects fetched
- `ErrorMsg` - Error occurred
- `RealtimeConnectedMsg` - WebSocket connected
- `RealtimeTaskUpdateMsg` - Task updated via WebSocket

**Flow**:
```
Command Execution → API Call → Result → Message → Update Handler
```

**Example**:
```go
// Create command
func LoadTasks(client interfaces.ArchonClient, projectID *string) tea.Cmd {
    return func() tea.Msg {
        tasks, err := client.GetTasks(projectID)
        if err != nil {
            return ErrorMsg{Error: err}
        }
        return TasksLoadedMsg{Tasks: tasks}
    }
}

// Handle message
case commands.TasksLoadedMsg:
    m.programContext.Tasks = msg.Tasks
    m.programContext.Loading = false
    m.updateTaskListComponents()
```

### 3. Component Messages

**Source**: Components communicating with each other

**Location**: `components/*/messages.go`

**Types**:

#### Data Update Messages
```go
// Update component data
type TaskListUpdateMsg struct {
    Tasks []archon.Task
    Loading bool
    Error string
}

type TaskDetailsUpdateMsg struct {
    SelectedTask *archon.Task
    SearchQuery string
    SearchActive bool
}
```

#### Action Messages
```go
// User took action in component
type TaskListSelectionChangedMsg struct {
    Index int
}

type ProjectListSelectionChangedMsg struct {
    Index int
    ProjectID string
}
```

#### Control Messages
```go
// Control component behavior
type TaskListScrollMsg struct {
    Direction ScrollDirection
}

type TaskListSelectMsg struct {
    Index int  // Set specific selection
}
```

**Flow**:
```
Component A → Update → Return Cmd → Message → Component B/Model receives
```

### 4. Modal Messages

**Source**: Modal components and modal activation

**Location**: `components/modals/*/messages.go`

**Pattern**: All modals follow Show/Hide/Action pattern

**Types**:

#### Lifecycle Messages
```go
// Show modal (activate)
type ShowHelpModalMsg struct{}

type ShowTaskEditModalMsg struct {
    TaskID string
    CurrentStatus string
    // ... initial values
}

// Hide modal (deactivate)
type HideHelpModalMsg struct{}
type HideTaskEditModalMsg struct{}

// Confirmation messages (internal)
type HelpModalShownMsg struct{}
type TaskEditModalHiddenMsg struct{}
```

#### Action Messages
```go
// User completed action in modal
type TaskPropertiesUpdatedMsg struct {
    TaskID string
    NewStatus *string
    NewPriority *int
}

type FeatureSelectionAppliedMsg struct {
    SelectedFeatures map[string]bool
}

type StatusSelectedMsg struct {
    NewStatus string
}
```

**Flow**:
```
Handler → ShowModalMsg → Modal.Update → Modal.active = true → View renders modal
   User interacts with modal
Modal → ActionMsg → Hide Modal → Model.Update handles action
```

### 5. System Messages

**Source**: Bubble Tea runtime

**Types**:
- `tea.WindowSizeMsg` - Terminal resized
- `tea.MouseMsg` - Mouse input (not currently used)
- `tea.QuitMsg` - Application quit

**Example**:
```go
case tea.WindowSizeMsg:
    m.programContext.UpdateScreenDimensions(msg.Width, msg.Height)
    m.components.Update(msg)  // Broadcast to all components
```

## Message Routing

### Priority-Based Router

Messages are routed through a **priority system** in `HandleKeyPress`:

```go
func (m *Model) HandleKeyPress(key string) tea.Cmd {
    // PRIORITY 1: Global keys (Ctrl+C, ?)
    if cmd, handled := m.handleGlobalKeys(key); handled {
        return cmd
    }

    // PRIORITY 2: Search input mode
    if m.programContext.SearchMode {
        return m.handleInlineSearchInput(key)
    }

    // PRIORITY 3: Modal keys
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

**Why Priority Matters**: Ensures emergency keys (Ctrl+C) always work, even when modals are open.

### Message Dispatchers

Messages are dispatched by type:

```go
// Application key dispatcher
func (m *Model) handleApplicationKey(key string) (tea.Cmd, bool) {
    switch key {
    case keys.KeyQ:
        return m.HandleQuitKey(key)
    case keys.KeyR:
        return m.HandleRefreshKey(key)
    case keys.KeyP:
        return m.HandleProjectModeKey(key)
    // ... more keys
    }
}

// Navigation key dispatcher
func (m *Model) handleNavigationKey(key string) (tea.Cmd, bool) {
    switch key {
    case keys.KeyArrowUp, keys.KeyK:
        return m.HandleUpNavigationKey(key)
    case keys.KeyArrowDown, keys.KeyJ:
        return m.HandleDownNavigationKey(key)
    // ... more keys
    }
}
```

**Organized by Feature**: Different dispatchers for navigation, search, task operations (see `input_handlers_*.go` files).

## Message Flow Patterns

### Pattern 1: User Input → Handler → Component

**Example**: User presses 'j' to move down

```
1. User presses 'j'
   ↓
2. tea.KeyMsg{Runes: ['j']}
   ↓
3. HandleKeyPress → handleNavigationKey → HandleDownNavigationKey
   ↓
4. Creates TaskListScrollMsg{Direction: ScrollDown}
   ↓
5. Sends to TaskList component via MainContent.Update(msg)
   ↓
6. TaskList.handleScrollMessages updates selectedIndex
   ↓
7. Returns TaskListSelectionChangedMsg
   ↓
8. Model.Update receives selection change, updates state
```

### Pattern 2: User Input → Modal → API → Update

**Example**: User changes task status

```
1. User presses 't'
   ↓
2. HandleTaskStatusChangeKey
   ↓
3. Creates ShowTaskEditModalMsg with current task data
   ↓
4. TaskEdit modal activates, shows UI
   ↓
5. User selects new status, presses Enter
   ↓
6. TaskEdit creates TaskPropertiesUpdatedMsg
   ↓
7. Model.handleModalActions receives message
   ↓
8. Makes API call via UpdateTaskStatus command
   ↓
9. API returns TaskUpdateMsg
   ↓
10. Model updates task in programContext
    ↓
11. Broadcasts TaskListUpdateMsg to refresh UI
```

### Pattern 3: WebSocket → Realtime Message → Update

**Example**: Task updated by another user

```
1. WebSocket receives event
   ↓
2. Creates RealtimeTaskUpdateMsg
   ↓
3. Model.handleRealtimeMessages
   ↓
4. Updates task in programContext.Tasks
   ↓
5. Broadcasts TaskListUpdateMsg to refresh UI
   ↓
6. Components re-render with new data
```

### Pattern 4: Command Chain

**Example**: Load tasks after selecting project

```
1. User selects project
   ↓
2. SetSelectedProject(projectID)
   ↓
3. Returns LoadTasks command
   ↓
4. Command executes async
   ↓
5. Returns TasksLoadedMsg
   ↓
6. Update handler processes new tasks
   ↓
7. Returns RefreshUI command
   ↓
8. UI updates with new tasks
```

## Message Handler Organization

### Model Message Handlers

Extracted to separate files by category:

**`model_handlers_task.go`**: Task and project data messages
```go
func (m *Model) handleTaskMessages(msg tea.Msg) (tea.Model, tea.Cmd)
func (m *Model) handleProjectMessages(msg tea.Msg) (tea.Model, tea.Cmd)
```

**`model_handlers_modal.go`**: Modal lifecycle and actions
```go
func (m *Model) handleModalLifecycle(msg tea.Msg) (tea.Model, tea.Cmd)
func (m *Model) handleModalActions(msg tea.Msg) (tea.Model, tea.Cmd)
```

**`model_handlers_realtime.go`**: WebSocket and animation
```go
func (m *Model) handleRealtimeMessages(msg tea.Msg) (tea.Model, tea.Cmd)
func (m *Model) handleTickAnimation(msg tickMsg) (tea.Model, tea.Cmd)
```

### Input Handlers

Organized by feature area:

**`input_handlers_navigation.go`**: Navigation keys (j/k/h/l/gg/G)
```go
func (m *Model) HandleUpNavigationKey(key string) (tea.Cmd, bool)
func (m *Model) HandleDownNavigationKey(key string) (tea.Cmd, bool)
// ... 10 navigation handlers
```

**`input_handlers_search.go`**: Search keys (/,n,N,Ctrl+X)
```go
func (m *Model) HandleActivateSearchKey(key string) (tea.Cmd, bool)
func (m *Model) HandleNextSearchMatchKey(key string) (tea.Cmd, bool)
// ... 4 search handlers
```

**`input_handlers_task.go`**: Task operation keys (t,e,y,f,s)
```go
func (m *Model) HandleTaskStatusChangeKey(key string) (tea.Cmd, bool)
func (m *Model) HandleTaskEditKey(key string) (tea.Cmd, bool)
// ... 7 task operation handlers
```

## Creating Custom Messages

### Step 1: Define Message Type

```go
// components/mycomponent/messages.go
package mycomponent

// Data message
type MyComponentUpdateMsg struct {
    Data []Item
    Config Settings
}

// Action message
type MyComponentActionMsg struct {
    SelectedItem Item
    Action ActionType
}

// Control message
type MyComponentScrollMsg struct {
    Direction ScrollDirection
}
```

### Step 2: Create Message Constructor (Optional)

```go
// Helper function to create message
func UpdateData(data []Item, config Settings) tea.Cmd {
    return func() tea.Msg {
        return MyComponentUpdateMsg{
            Data: data,
            Config: config,
        }
    }
}
```

### Step 3: Send Message

```go
// From handler
return UpdateData(newData, config)

// From component
return func() tea.Msg {
    return MyComponentActionMsg{
        SelectedItem: m.items[m.selectedIndex],
        Action: ActionConfirm,
    }
}
```

### Step 4: Handle Message

```go
// In Model.Update
case mycomponent.MyComponentActionMsg:
    return m.handleMyComponentAction(msg)

// In Component.Update
case MyComponentUpdateMsg:
    m.data = msg.Data
    m.config = msg.Config
    m.refresh()
    return nil
```

## Message Best Practices

### 1. Use Descriptive Names

```go
// ✅ GOOD: Clear what message does
type TaskListSelectionChangedMsg struct {
    Index int
}

// ❌ BAD: Vague purpose
type UpdateMsg struct {
    Index int
}
```

### 2. Include Necessary Context

```go
// ✅ GOOD: Has all needed info
type TaskPropertiesUpdatedMsg struct {
    TaskID string
    NewStatus *string
    NewPriority *int
}

// ❌ BAD: Missing which task
type PropertiesUpdatedMsg struct {
    NewStatus string
}
```

### 3. Use Pointer for Optional Fields

```go
// ✅ GOOD: Can tell if field was set
type UpdateMsg struct {
    NewStatus *string  // nil = not changed
}

// ❌ BAD: Can't distinguish empty from unchanged
type UpdateMsg struct {
    NewStatus string  // "" = unchanged or cleared?
}
```

### 4. Document Message Flow

```go
// TaskListSelectMsg is sent by the Model to select a specific task.
// The TaskList component handles this by updating its selectedIndex
// and returning TaskListSelectionChangedMsg to notify the Model.
type TaskListSelectMsg struct {
    Index int
}
```

### 5. Group Related Messages

```go
// Modal lifecycle messages grouped together
type (
    ShowHelpModalMsg struct{}
    HideHelpModalMsg struct{}
    HelpModalShownMsg struct{}
    HelpModalHiddenMsg struct{}
)
```

## Debugging Messages

### 1. Log Messages

```go
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Log all messages
    m.programContext.Logger.Debug("Update", "msg", fmt.Sprintf("%T", msg))

    switch msg := msg.(type) {
    // ... handle messages
    }
}
```

### 2. Message Inspector

Add temporary message viewer:

```go
// In View()
if m.debugMode {
    return fmt.Sprintf("Last message: %T\n%+v", m.lastMsg, m.lastMsg)
}
```

### 3. Command Tracing

Wrap commands to trace execution:

```go
func traceCmd(name string, cmd tea.Cmd) tea.Cmd {
    if cmd == nil {
        return nil
    }
    return func() tea.Msg {
        log.Printf("Executing command: %s", name)
        msg := cmd()
        log.Printf("Command %s returned: %T", name, msg)
        return msg
    }
}

// Use in handlers
return traceCmd("LoadTasks", LoadTasks(client, projectID))
```

## Common Message Patterns

### Pattern: Request-Response

```go
// Request
type LoadDataRequestMsg struct {
    Filter string
}

// Response
type LoadDataResponseMsg struct {
    Data []Item
    Error error
}
```

### Pattern: Event Notification

```go
// Something happened, no response expected
type SelectionChangedMsg struct {
    OldIndex int
    NewIndex int
}
```

### Pattern: Command Result

```go
// Async operation completed
type TaskUpdateCompletedMsg struct {
    TaskID string
    Success bool
    Error error
}
```

### Pattern: Batch Update

```go
// Multiple changes at once
type BatchUpdateMsg struct {
    TaskUpdates []TaskUpdate
    ProjectUpdates []ProjectUpdate
}
```

## References

- [Bubble Tea Messages](https://github.com/charmbracelet/bubbletea#messages)
- [Command Pattern](https://github.com/charmbracelet/bubbletea#commands)
- [Message Definitions](../components/)
- [Handler Organization](../input_handlers_*.go)
