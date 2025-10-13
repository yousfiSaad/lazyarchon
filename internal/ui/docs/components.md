# Component System

## Overview

LazyArchon uses a **component-based architecture** where each UI element is a self-contained component with its own state, update logic, and rendering.

## Base Component Pattern

All components implement the `base.Component` interface:

```go
type Component interface {
    Update(msg tea.Msg) tea.Cmd
    View() string
}
```

## Component Hierarchy

```
Root (Model)
│
├── Layout Components
│   ├── Header (top bar with title)
│   ├── MainContent (dynamic panel container)
│   │   ├── TaskList (left panel in task mode)
│   │   ├── TaskDetails (right panel in task mode)
│   │   ├── ProjectList (left panel in project mode)
│   │   └── ProjectDetails (right panel in project mode)
│   └── StatusBar (bottom bar with status/errors)
│
└── Modal Components (overlay UI)
    ├── Help (keyboard shortcuts)
    ├── Confirmation (yes/no dialogs)
    ├── TaskEdit (edit task properties)
    ├── Feature (filter by features)
    ├── Status (change task status)
    └── StatusFilter (filter by status)
```

## Core Components

### 1. TaskList Component

**Location**: `components/tasklist/`

**Responsibility**: Display filterable, sortable list of tasks

**State**:
```go
type Model struct {
    tasks []archon.Task
    selectedIndex int
    viewport viewport.Model
    searchQuery string
    searchActive bool
    sortedTasks []archon.Task
    // ... more fields
}
```

**Key Methods**:
- `setSelectedIndex(int)` - Change selection (with viewport update)
- `updateViewportContent()` - Regenerate viewport display
- `followSelection()` - Adjust scroll to keep selection visible
- `handleScrollMessages()` - Process navigation keys
- `handleDataMessages()` - Process data updates

**Messages**:
```go
// Incoming
type TaskListUpdateMsg struct {
    Tasks []archon.Task
    Loading bool
    Error string
}

type TaskListSelectMsg struct {
    Index int
}

type TaskListScrollMsg struct {
    Direction ScrollDirection
}

// Outgoing
type TaskListSelectionChangedMsg struct {
    Index int
}
```

**Design Pattern**: TaskList owns the **selection cursor** and **viewport scrolling**. It's the source of truth for which task is selected.

**Critical Invariant**: When `selectedIndex` changes, viewport MUST be regenerated via `updateViewportContent()`. This is enforced by `setSelectedIndex()` helper.

### 2. TaskDetails Component

**Location**: `components/taskdetails/`

**Responsibility**: Display detailed view of selected task

**State**:
```go
type Model struct {
    selectedTask *archon.Task
    panelCore *detailspanel.Core  // Handles scrolling
    contentGenerator *ContentGenerator  // Renders markdown
    searchQuery string
    searchActive bool
}
```

**Key Methods**:
- `updateContent()` - Regenerate task detail content
- `broadcastScrollPosition()` - Notify parent of scroll changes

**Messages**:
```go
// Incoming
type TaskDetailsUpdateMsg struct {
    SelectedTask *archon.Task
    SearchQuery string
    SearchActive bool
}

type TaskDetailsScrollMsg struct {
    Direction viewport.ScrollDirection
}

// Outgoing
type TaskDetailsScrollPositionMsg struct {
    YOffset int
    MaxOffset int
}
```

**Design Pattern**: TaskDetails is a **read-only view**. It doesn't modify tasks, just displays them.

### 3. ProjectList Component

**Location**: `components/projectlist/`

**Responsibility**: Display list of projects (in project mode)

**State**:
```go
type Model struct {
    projects []archon.Project
    selectedIndex int
    viewport viewport.Model
}
```

**Messages**:
```go
type ProjectListUpdateMsg struct {
    Projects []archon.Project
}

type ProjectListScrollMsg struct {
    Direction ScrollDirection
}

type ProjectListSelectionChangedMsg struct {
    Index int
    ProjectID string
}
```

### 4. Modal Components

**Location**: `components/modals/*/`

**Shared Pattern**: All modals follow the same lifecycle:

```
Inactive → Show Message → Active → User Action → Hide Message → Inactive
```

**Base Modal Pattern**:
```go
type Model struct {
    active bool
    width int
    height int
    // ... modal-specific fields
}

// Lifecycle
func (m *Model) IsActive() bool
func (m *Model) SetActive(bool)
func (m *Model) Show(msg ShowModalMsg)  // Activate with data
func (m *Model) Hide()                   // Deactivate
```

**Example: TaskEdit Modal**:

```go
// Show message carries initial data
type ShowTaskEditModalMsg struct {
    TaskID string
    CurrentStatus string
    CurrentPriority int
    CurrentFeature string
    FocusField FieldType
    AvailableFeatures []string
}

// Action message carries result
type TaskPropertiesUpdatedMsg struct {
    TaskID string
    NewStatus *string
    NewPriority *int
    NewFeature *string
}
```

**Modal Lifecycle Flow**:
1. User presses 't' → `HandleTaskStatusChangeKey`
2. Handler creates `ShowTaskEditModalMsg` with current values
3. Modal receives message, sets `active = true`, initializes fields
4. User edits fields, presses Enter
5. Modal creates `TaskPropertiesUpdatedMsg` with changes
6. Modal sets `active = false`
7. Model receives action message, updates task via API

## Component Communication

### 1. Parent-to-Child (Downward)

**Via Messages**: Parent sends messages to children

```go
// Model sends update to TaskList
updateMsg := tasklist.TaskListUpdateMsg{
    Tasks: newTasks,
    Loading: false,
}
m.components.Layout.MainContent.Update(updateMsg)
```

### 2. Child-to-Parent (Upward)

**Via Return Messages**: Child returns messages that bubble up

```go
// TaskList selection changes
func (m *Model) Update(msg TaskListSelectMsg) tea.Cmd {
    m.setSelectedIndex(msg.Index)

    // Return message that parent will see
    return func() tea.Msg {
        return TaskListSelectionChangedMsg{Index: m.selectedIndex}
    }
}
```

**Model receives it**:
```go
case tasklist.TaskListSelectionChangedMsg:
    // Update model state based on selection change
    m.selectedIndex = msg.Index
```

### 3. Sibling Communication

**Through Parent**: Siblings communicate via the parent (Model)

```go
// TaskList selection changes
TaskList → TaskListSelectionChangedMsg → Model
                                           ↓
Model updates selectedTask, sends TaskDetailsUpdateMsg
                                           ↓
                                      TaskDetails
```

## Creating a New Component

### Step 1: Define Component Structure

```go
// components/mycomponent/component.go
package mycomponent

type Model struct {
    // Embed base component for common functionality
    *base.BaseComponent

    // Component-specific state
    data []DataItem
    selectedIndex int
    active bool

    // UI state
    width int
    height int
}
```

### Step 2: Define Messages

```go
// components/mycomponent/messages.go
package mycomponent

// Incoming messages
type MyComponentUpdateMsg struct {
    Data []DataItem
}

type MyComponentScrollMsg struct {
    Direction ScrollDirection
}

// Outgoing messages
type MyComponentActionMsg struct {
    SelectedItem DataItem
}
```

### Step 3: Implement Component Interface

```go
// Update handles messages
func (m *Model) Update(msg tea.Msg) tea.Cmd {
    switch msg := msg.(type) {
    case MyComponentUpdateMsg:
        m.data = msg.Data
        m.render()
        return nil

    case MyComponentScrollMsg:
        m.handleScroll(msg.Direction)
        return nil
    }
    return nil
}

// View renders component
func (m *Model) View() string {
    if !m.active {
        return ""
    }
    return m.renderContent()
}
```

### Step 4: Add to Component Tree

```go
// factories/component_factory.go
type UIComponentSet struct {
    MyComponent *mycomponent.Model
    // ... other components
}

func CreateUIComponents(...) UIComponentSet {
    myComp := mycomponent.NewModel(...)

    return UIComponentSet{
        MyComponent: myComp,
        // ... other components
    }
}
```

### Step 5: Wire Up in Model

```go
// model.go
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case mycomponent.MyComponentActionMsg:
        return m.handleMyComponentAction(msg)
    }
    // ...
}
```

## Component Best Practices

### 1. Single Responsibility

Each component should do **one thing well**:
- ✅ TaskList: Display and navigate tasks
- ❌ TaskList: Display tasks, edit tasks, manage API calls

### 2. Clear Ownership

Each piece of state has **one owner**:
- TaskList owns `selectedIndex` for task list
- Model owns global `selectedIndex` (synced with TaskList)
- TaskDetails is read-only (doesn't own selection)

### 3. Message-Driven Updates

**Never** directly mutate parent state from child:

```go
// ❌ BAD: Direct mutation
m.parent.selectedIndex = 5

// ✅ GOOD: Send message
return func() tea.Msg {
    return TaskListSelectionChangedMsg{Index: 5}
}
```

### 4. Encapsulate Invariants

Use helper methods to maintain invariants:

```go
// ✅ GOOD: Helper ensures viewport updates
func (m *Model) setSelectedIndex(index int) {
    m.selectedIndex = index
    m.updateViewportContent()  // ALWAYS happens
    m.followSelection()
}

// ❌ BAD: Easy to forget viewport update
m.selectedIndex = index
// Forgot to call updateViewportContent()!
```

### 5. Stateless Where Possible

Prefer stateless components that receive all data via messages:

```go
// ✅ GOOD: Receives current task via message
type TaskDetailsUpdateMsg struct {
    SelectedTask *archon.Task
}

// ❌ BAD: Fetches task from global state
func (m *Model) getCurrentTask() *archon.Task {
    return globalState.tasks[globalState.selectedIndex]
}
```

## Component Lifecycle

### Initialization

```go
func NewModel(deps Dependencies) *Model {
    m := &Model{
        // Initialize with safe defaults
        selectedIndex: 0,
        active: false,
        viewport: viewport.New(0, 0),
    }
    return m
}
```

### Activation/Deactivation

Components can be active/inactive:

```go
// Set active state
m.SetActive(true)

// Skip rendering when inactive
func (m *Model) View() string {
    if !m.active {
        return ""
    }
    return m.render()
}
```

### Resizing

Handle terminal resize:

```go
func (m *Model) Update(msg tea.Msg) tea.Cmd {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.recalculateLayout()
        return nil
    }
}
```

## Component Testing

### Unit Tests

Test component in isolation:

```go
func TestMyComponent(t *testing.T) {
    // Create component with test dependencies
    comp := NewModel(testDeps)

    // Send message
    msg := MyComponentUpdateMsg{Data: testData}
    cmd := comp.Update(msg)

    // Verify state
    assert.Equal(t, testData, comp.data)
    assert.NotNil(t, cmd)
}
```

### Integration Tests

Test component interaction:

```go
func TestTaskListTaskDetailsInteraction(t *testing.T) {
    taskList := tasklist.NewModel(...)
    taskDetails := taskdetails.NewModel(...)

    // Select task in list
    selectMsg := tasklist.TaskListSelectMsg{Index: 5}
    cmd := taskList.Update(selectMsg)

    // Get outgoing message
    msg := cmd()
    selectionMsg := msg.(tasklist.TaskListSelectionChangedMsg)

    // Update details
    updateMsg := taskdetails.TaskDetailsUpdateMsg{
        SelectedTask: tasks[selectionMsg.Index],
    }
    taskDetails.Update(updateMsg)

    // Verify
    assert.Equal(t, tasks[5].Title, taskDetails.selectedTask.Title)
}
```

## Common Patterns

### Pattern: Viewport with Selection

**Problem**: Need scrollable list with cursor

**Solution**: Component owns viewport + selection index

```go
type Model struct {
    items []Item
    selectedIndex int
    viewport viewport.Model
}

func (m *Model) setSelectedIndex(index int) {
    m.selectedIndex = index
    m.updateViewportContent()  // Regenerate with cursor
    m.followSelection()         // Scroll to show selection
}
```

**Examples**: TaskList, ProjectList, Feature Modal

### Pattern: Read-Only Display

**Problem**: Show data without editing

**Solution**: Component receives data via messages, doesn't modify

```go
func (m *Model) Update(msg tea.Msg) tea.Cmd {
    switch msg := msg.(type) {
    case UpdateMsg:
        m.data = msg.Data  // Store
        m.regenerateView() // Display
        return nil
    }
}
```

**Examples**: TaskDetails, ProjectDetails

### Pattern: Modal Dialog

**Problem**: Temporary overlay UI for user input

**Solution**: Active/inactive component with lifecycle messages

```go
// Show
func (m *Model) Update(msg ShowModalMsg) tea.Cmd {
    m.active = true
    m.initializeFields(msg.InitialData)
    return nil
}

// Hide
func (m *Model) Hide() {
    m.active = false
}

// View only renders when active
func (m *Model) View() string {
    if !m.active {
        return ""
    }
    return m.renderModal()
}
```

**Examples**: All modals (Help, TaskEdit, Feature, etc.)

### Pattern: Panel Core

**Problem**: Common scrolling/sizing logic across detail panels

**Solution**: Extract to reusable `detailspanel.Core`

```go
type Core struct {
    width int
    height int
    yOffset int
    content string
}

func (c *Core) HandleScroll(direction ScrollDirection)
func (c *Core) SetContent(content string)
func (c *Core) View() string
```

**Examples**: TaskDetails, ProjectDetails

## References

- [Base Component](../components/base/component.go)
- [TaskList Implementation](../components/tasklist/component.go)
- [Modal Pattern](../components/modals/)
- [Bubble Tea Components](https://github.com/charmbracelet/bubbletea#components)
