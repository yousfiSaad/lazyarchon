# Navigation System

## Overview

LazyArchon implements a **dual-mode navigation system** that supports both task-focused workflows and project-based organization. Navigation is consistent, intuitive, and follows vim-like keybindings.

## Navigation Modes

### Task Mode (Default)

**Purpose**: Navigate and manage tasks within the selected project

**Layout**:
```
┌────────────────────────────────────────────┐
│ Header: Project Name                       │
├──────────────────┬─────────────────────────┤
│ TaskList (left)  │ TaskDetails (right)     │
│ ○ Task 1      ←  │ Title: Task 1           │
│   Task 2         │ Status: todo            │
│   Task 3         │ Description...          │
│                  │                         │
│                  │                         │
├──────────────────┴─────────────────────────┤
│ Status Bar                                 │
└────────────────────────────────────────────┘
```

**Active Panel**:
- Left Panel: TaskList is focused
- Right Panel: TaskDetails is focused
- Use 'h' and 'l' to switch between panels

### Project Mode

**Purpose**: Select which project to work on

**Layout**:
```
┌────────────────────────────────────────────┐
│ Header: Projects                           │
├──────────────────┬─────────────────────────┤
│ ProjectList      │ ProjectDetails          │
│ → Project 1      │ Title: Project 1        │
│   Project 2      │ Description...          │
│   Project 3      │ Features: auth, api     │
│                  │                         │
│                  │                         │
├──────────────────┴─────────────────────────┤
│ Status Bar                                 │
└────────────────────────────────────────────┘
```

**How to Enter**:
- Press 'p' from task mode
- Arrow keys or j/k to navigate projects
- Enter or 'l' to select project and return to task mode
- Esc, 'q', or 'h' to cancel and return to task mode

## Navigation Handlers

All navigation is handled in `input_handlers_navigation.go`.

### Handler Functions

```go
// Basic navigation
HandleUpNavigationKey(key)              // k, ↑
HandleDownNavigationKey(key)            // j, ↓
HandleLeftNavigationKey(key)            // h
HandleRightNavigationKey(key)           // l

// Jump navigation
HandleJumpToFirstKey(key)               // gg, Home
HandleJumpToLastKey(key)                // G, End

// Fast scrolling
HandleFastScrollUpKey(key)              // K (capital)
HandleFastScrollDownKey(key)            // J (capital)

// Page scrolling
HandleHalfPageUpKey(key)                // Ctrl+U, PgUp
HandleHalfPageDownKey(key)              // Ctrl+D, PgDn
```

## Navigation Keys

### Basic Movement

| Key | Action | Context |
|-----|--------|---------|
| j, ↓ | Move down one item | TaskList, ProjectList, TaskDetails |
| k, ↑ | Move up one item | TaskList, ProjectList, TaskDetails |
| h | Move left / Focus left panel / Exit project mode | All |
| l | Move right / Focus right panel / Select project | All |

### Jump Navigation

| Key | Action | Context |
|-----|--------|---------|
| gg | Jump to first item | TaskList, ProjectList |
| Home | Jump to first item | TaskList, ProjectList |
| G | Jump to last item | TaskList, ProjectList |
| End | Jump to last item | TaskList, ProjectList |

### Fast Scrolling

| Key | Action | Context |
|-----|--------|---------|
| J | Scroll down 4 items | TaskList, ProjectList, TaskDetails |
| K | Scroll up 4 items | TaskList, ProjectList, TaskDetails |
| Ctrl+D | Half-page down | TaskList, ProjectList, TaskDetails |
| PgDn | Half-page down | TaskList, ProjectList, TaskDetails |
| Ctrl+U | Half-page up | TaskList, ProjectList, TaskDetails |
| PgUp | Half-page up | TaskList, ProjectList, TaskDetails |

## Navigation Implementation

### Message-Based Navigation

All navigation sends scroll messages to components:

```go
func (m *Model) HandleDownNavigationKey(key string) (tea.Cmd, bool) {
    cmd := m.handleDownNavigation()
    return cmd, true
}

func (m *Model) handleDownNavigation() tea.Cmd {
    if m.programContext.IsProjectModeActive() {
        // Project mode: scroll project list
        scrollMsg := projectlist.ProjectListScrollMsg{
            Direction: ScrollDown,
        }
        return m.components.Layout.MainContent.Update(scrollMsg)
    } else {
        // Task mode: scroll active panel
        if m.IsLeftPanelActive() {
            scrollMsg := tasklist.TaskListScrollMsg{
                Direction: ScrollDown,
            }
            return m.components.Layout.MainContent.Update(scrollMsg)
        } else if m.IsRightPanelActive() {
            scrollMsg := taskdetails.TaskDetailsScrollMsg{
                Direction: ScrollDown,
            }
            return m.components.Layout.MainContent.Update(scrollMsg)
        }
    }
    return nil
}
```

### Scroll Directions

```go
type ScrollDirection int

const (
    ScrollUp ScrollDirection = iota
    ScrollDown
    ScrollToTop
    ScrollToBottom
    ScrollFastUp
    ScrollFastDown
    ScrollPageUp
    ScrollPageDown
)
```

### Panel Focus

```go
type ActiveView int

const (
    LeftPanel ActiveView = iota
    RightPanel
)

func (m *Model) IsLeftPanelActive() bool {
    return m.activeView == LeftPanel
}

func (m *Model) IsRightPanelActive() bool {
    return m.activeView == RightPanel
}

func (m *Model) SetActiveView(view ActiveView) tea.Cmd {
    m.activeView = view
    // Notify components of focus change
    // ...
    return nil
}
```

## Task Mode Navigation

### Task List Navigation

**Component**: `components/tasklist/`

**Selection Management**:
```go
type Model struct {
    tasks         []archon.Task
    sortedTasks   []archon.Task
    selectedIndex int      // Currently selected task
    viewport      viewport.Model
}

// Critical method - ensures cursor updates
func (m *Model) setSelectedIndex(newIndex int) {
    if newIndex < 0 || newIndex >= len(m.sortedTasks) {
        return
    }

    m.selectedIndex = newIndex
    m.updateViewportContent()  // Regenerate with cursor
    m.followSelection()         // Scroll to show cursor
}
```

**Scroll Handling**:
```go
func (m *Model) handleScrollMessages(msg TaskListScrollMsg) tea.Cmd {
    switch msg.Direction {
    case ScrollUp:
        m.setSelectedIndex(m.selectedIndex - 1)
    case ScrollDown:
        m.setSelectedIndex(m.selectedIndex + 1)
    case ScrollToTop:
        m.setSelectedIndex(0)
    case ScrollToBottom:
        m.setSelectedIndex(len(m.sortedTasks) - 1)
    case ScrollFastUp:
        m.setSelectedIndex(max(0, m.selectedIndex-4))
    case ScrollFastDown:
        m.setSelectedIndex(min(len(m.sortedTasks)-1, m.selectedIndex+4))
    case ScrollPageUp:
        m.setSelectedIndex(max(0, m.selectedIndex-m.maxLines))
    case ScrollPageDown:
        m.setSelectedIndex(min(len(m.sortedTasks)-1, m.selectedIndex+m.maxLines))
    }

    // Return selection changed message
    return func() tea.Msg {
        return TaskListSelectionChangedMsg{Index: m.selectedIndex}
    }
}
```

**Viewport Following**:
```go
func (m *Model) followSelection() {
    if len(m.sortedTasks) == 0 {
        return
    }

    // Calculate viewport bounds
    viewportTop := m.viewport.YOffset
    viewportBottom := viewportTop + m.viewport.Height - 1

    // If selection is above viewport, scroll up
    if m.selectedIndex < viewportTop {
        m.viewport.SetYOffset(m.selectedIndex)
    }

    // If selection is below viewport, scroll down
    if m.selectedIndex > viewportBottom {
        newOffset := m.selectedIndex - m.viewport.Height + 1
        m.viewport.SetYOffset(newOffset)
    }
}
```

### Task Details Navigation

**Component**: `components/taskdetails/`

**Scroll Handling**:
```go
type Model struct {
    selectedTask *archon.Task
    panelCore    *detailspanel.Core  // Handles scrolling
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
    switch msg := msg.(type) {
    case TaskDetailsScrollMsg:
        return m.handleScroll(msg)
    }
}

func (m *Model) handleScroll(msg TaskDetailsScrollMsg) tea.Cmd {
    m.panelCore.HandleScroll(msg.Direction)
    return m.broadcastScrollPosition()
}

func (m *Model) broadcastScrollPosition() tea.Cmd {
    return func() tea.Msg {
        return TaskDetailsScrollPositionMsg{
            YOffset:   m.panelCore.GetYOffset(),
            MaxOffset: m.panelCore.GetMaxYOffset(),
        }
    }
}
```

### Panel Switching

**Left (h)**:
```go
func (m *Model) HandleLeftNavigationKey(key string) (tea.Cmd, bool) {
    if m.programContext.IsProjectModeActive() {
        // In project mode: exit to task mode
        cmd := func() tea.Msg {
            return commands.ProjectModeDeactivatedMsg{
                ShouldLoadTasks: false,
            }
        }
        return cmd, true
    } else {
        // In task mode: focus left panel
        cmd := m.SetActiveView(LeftPanel)
        return cmd, true
    }
}
```

**Right (l)**:
```go
func (m *Model) HandleRightNavigationKey(key string) (tea.Cmd, bool) {
    if m.programContext.IsProjectModeActive() {
        // In project mode: select project
        return m.handleProjectSelection(), true
    } else {
        // In task mode: focus right panel
        cmd := m.SetActiveView(RightPanel)
        return cmd, true
    }
}
```

## Project Mode Navigation

### Entering Project Mode

**Trigger**: Press 'p' from task mode

```go
func (m *Model) HandleProjectModeKey(key string) (tea.Cmd, bool) {
    if key == keys.KeyP {
        return func() tea.Msg {
            return commands.ProjectModeActivatedMsg{}
        }, true
    }
    return nil, false
}
```

**Activation Flow**:
```
1. User presses 'p'
   ↓
2. ProjectModeActivatedMsg sent
   ↓
3. Model sets project mode flag
   ↓
4. ProjectList component activated
   ↓
5. View renders project list
```

### Project List Navigation

**Component**: `components/projectlist/`

**Selection Management**:
```go
type Model struct {
    projects      []archon.Project
    selectedIndex int
    viewport      viewport.Model
}

func (m *Model) handleScrollMessages(msg ProjectListScrollMsg) tea.Cmd {
    switch msg.Direction {
    case ScrollUp:
        m.selectedIndex = max(0, m.selectedIndex-1)
    case ScrollDown:
        m.selectedIndex = min(len(m.projects)-1, m.selectedIndex+1)
    case ScrollToTop:
        m.selectedIndex = 0
    case ScrollToBottom:
        m.selectedIndex = len(m.projects) - 1
    // ... other directions
    }

    m.updateViewportContent()
    m.followSelection()

    return func() tea.Msg {
        return ProjectListSelectionChangedMsg{
            Index:     m.selectedIndex,
            ProjectID: m.projects[m.selectedIndex].ID,
        }
    }
}
```

### Project Selection

**Selection Flow**:
```
1. User presses Enter or 'l' on a project
   ↓
2. ProjectListSelectMsg sent with project ID
   ↓
3. Model saves selected project ID
   ↓
4. Load tasks for selected project
   ↓
5. Exit project mode
   ↓
6. Return to task mode with filtered tasks
```

**Implementation**:
```go
func (m *Model) handleProjectSelection() tea.Cmd {
    if len(m.programContext.Projects) == 0 {
        return nil
    }

    selectedProject := m.programContext.Projects[m.selectedIndex]

    return tea.Batch(
        // Set selected project
        func() tea.Msg {
            return commands.SetSelectedProjectMsg{
                ProjectID: &selectedProject.ID,
            }
        },
        // Load tasks for project
        commands.LoadTasks(
            m.programContext.ArchonClient,
            &selectedProject.ID,
        ),
        // Exit project mode
        func() tea.Msg {
            return commands.ProjectModeDeactivatedMsg{
                ShouldLoadTasks: false, // Already loading above
            }
        },
    )
}
```

### Exiting Project Mode

**Triggers**:
- Press 'h' (left/back)
- Press 'q' (quit project mode)
- Press Esc (cancel)
- Select a project (automatically returns to task mode)

```go
func (m *Model) HandleLeftNavigationKey(key string) (tea.Cmd, bool) {
    if m.programContext.IsProjectModeActive() {
        return func() tea.Msg {
            return commands.ProjectModeDeactivatedMsg{
                ShouldLoadTasks: false,
            }
        }, true
    }
    // ...
}
```

## Multi-Key Sequences

### gg - Jump to First

**Implementation**:
```go
type Model struct {
    lastKey string  // Track previous key
}

func (m *Model) HandleJumpToFirstKey(key string) (tea.Cmd, bool) {
    // Check for 'gg' sequence
    if key == keys.KeyG {
        if m.lastKey == keys.KeyG {
            // Second 'g' - execute jump
            cmd := m.handleJumpToFirst()
            m.lastKey = ""
            return cmd, true
        }
        // First 'g' - wait for second
        m.lastKey = keys.KeyG
        return nil, true
    }

    // Home key - immediate jump
    if key == keys.KeyHome {
        return m.handleJumpToFirst(), true
    }

    return nil, false
}

func (m *Model) handleJumpToFirst() tea.Cmd {
    if m.programContext.IsProjectModeActive() {
        scrollMsg := projectlist.ProjectListScrollMsg{
            Direction: ScrollToTop,
        }
        return m.components.Layout.MainContent.Update(scrollMsg)
    } else {
        if m.IsLeftPanelActive() {
            scrollMsg := tasklist.TaskListScrollMsg{
                Direction: ScrollToTop,
            }
            return m.components.Layout.MainContent.Update(scrollMsg)
        }
    }
    return nil
}
```

**Sequence Timeout**: The `lastKey` is reset when any other key is pressed:

```go
func (m *Model) handleKeyInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    key := msg.String()

    // Handle 'g' specially for gg sequence
    if key == keys.KeyG && m.lastKey == keys.KeyG {
        // Execute gg
        return m.handleJumpToFirst()
    }

    // Any other key resets the sequence
    if key != keys.KeyG {
        m.lastKey = ""
    }

    // Continue with normal key handling
    // ...
}
```

## Viewport Management

### Viewport Regeneration

**Critical Pattern**: Always regenerate viewport content when selection changes

```go
// ✅ GOOD: Uses helper that ensures regeneration
m.setSelectedIndex(newIndex)

// ❌ BAD: Direct assignment misses regeneration
m.selectedIndex = newIndex  // Cursor won't appear!
```

**Why Regeneration is Necessary**:
- Cursor indicator (→) is part of rendered content
- Content is pre-rendered into viewport
- Changing `selectedIndex` doesn't automatically update display
- Must call `updateViewportContent()` to re-render with new cursor position

### Viewport Content Update

```go
func (m *Model) updateViewportContent() {
    if len(m.sortedTasks) == 0 {
        m.viewport.SetContent("No tasks")
        return
    }

    var lines []string
    for i, task := range m.sortedTasks {
        // Render cursor indicator
        indicator := styling.NoSelection
        if i == m.selectedIndex {
            indicator = styling.SelectionIndicator  // "→ "
        }

        // Render task line
        line := indicator + m.renderTask(task)
        lines = append(lines, line)
    }

    content := strings.Join(lines, "\n")
    m.viewport.SetContent(content)
}
```

### Scroll Following

**Purpose**: Keep selected item visible in viewport

```go
func (m *Model) followSelection() {
    viewportTop := m.viewport.YOffset
    viewportBottom := viewportTop + m.viewport.Height - 1

    // Selection above viewport - scroll up
    if m.selectedIndex < viewportTop {
        m.viewport.SetYOffset(m.selectedIndex)
    }

    // Selection below viewport - scroll down
    if m.selectedIndex > viewportBottom {
        newOffset := m.selectedIndex - m.viewport.Height + 1
        m.viewport.SetYOffset(max(0, newOffset))
    }
}
```

## Selection State Synchronization

### Model ↔ Component Sync

**Problem**: Selection state exists in both Model and components

**Solution**: Single source of truth with message synchronization

```go
// Model (global state)
type Model struct {
    selectedIndex int  // Source of truth for selection
}

// Component (local UI state)
type TaskList struct {
    selectedIndex int  // Synchronized from Model
}
```

**Synchronization Flow**:

**Downward (Model → Component)**:
```go
// Model sends selection update
selectMsg := tasklist.TaskListSelectMsg{Index: newIndex}
m.components.Layout.MainContent.Update(selectMsg)

// Component receives and updates
case TaskListSelectMsg:
    m.setSelectedIndex(msg.Index)
```

**Upward (Component → Model)**:
```go
// Component selection changed
func (m *TaskList) setSelectedIndex(index int) tea.Cmd {
    m.selectedIndex = index
    m.updateViewportContent()
    m.followSelection()

    // Notify Model
    return func() tea.Msg {
        return TaskListSelectionChangedMsg{Index: m.selectedIndex}
    }
}

// Model receives notification
case tasklist.TaskListSelectionChangedMsg:
    m.selectedIndex = msg.Index
    m.updateTaskDetailsComponent()
```

## Navigation Performance

### Viewport Optimization

Only visible lines are rendered:

```go
func (m *Model) updateViewportContent() {
    // Only render tasks that fit in viewport
    // Viewport handles virtual scrolling

    var lines []string
    for i, task := range m.sortedTasks {
        lines = append(lines, m.renderTask(task, i))
    }

    // Viewport only displays visible portion
    content := strings.Join(lines, "\n")
    m.viewport.SetContent(content)
}
```

### Lazy Rendering

TaskDetails only regenerates when selection changes:

```go
case TaskDetailsUpdateMsg:
    if m.selectedTask == nil ||
       (msg.SelectedTask != nil && m.selectedTask.ID != msg.SelectedTask.ID) {
        // Task changed - regenerate
        m.selectedTask = msg.SelectedTask
        m.updateContent()
    }
```

## Navigation Edge Cases

### Empty Lists

Handle empty task/project lists gracefully:

```go
func (m *Model) handleDownNavigation() tea.Cmd {
    // Check for empty list
    sortedTasks := m.GetSortedTasks()
    if len(sortedTasks) == 0 {
        return nil  // No navigation possible
    }

    // Safe to navigate
    scrollMsg := tasklist.TaskListScrollMsg{Direction: ScrollDown}
    return m.components.Layout.MainContent.Update(scrollMsg)
}
```

### Bounds Checking

Always validate navigation bounds:

```go
func (m *Model) setSelectedIndex(newIndex int) {
    // Validate bounds
    if newIndex < 0 || newIndex >= len(m.sortedTasks) {
        return  // Invalid, don't change
    }

    m.selectedIndex = newIndex
    m.updateViewportContent()
    m.followSelection()
}
```

### Selection After Delete

Adjust selection when items are removed:

```go
func (m *Model) handleTaskDeleted(taskID string) {
    // Remove task from list
    m.programContext.Tasks = removeTask(m.programContext.Tasks, taskID)

    // Get updated sorted list
    sortedTasks := m.GetSortedTasks()

    // Adjust selection if out of bounds
    if m.selectedIndex >= len(sortedTasks) && len(sortedTasks) > 0 {
        m.selectedIndex = len(sortedTasks) - 1
    }

    // Update components
    m.updateTaskListComponents(sortedTasks)
}
```

## Best Practices

### 1. Use Helper Methods

```go
// ✅ GOOD: Helper ensures invariants
m.setSelectedIndex(newIndex)

// ❌ BAD: Manual update error-prone
m.selectedIndex = newIndex
m.updateViewportContent()
m.followSelection()
```

### 2. Send Messages for Navigation

```go
// ✅ GOOD: Message-based
scrollMsg := TaskListScrollMsg{Direction: ScrollDown}
return m.components.Update(scrollMsg)

// ❌ BAD: Direct manipulation
m.components.TaskList.selectedIndex++
```

### 3. Always Validate Bounds

```go
// ✅ GOOD: Validated
func (m *Model) setSelectedIndex(index int) {
    if index < 0 || index >= len(m.items) {
        return
    }
    // ...
}
```

### 4. Sync Selection State

```go
// ✅ GOOD: Notify on change
return func() tea.Msg {
    return SelectionChangedMsg{Index: m.selectedIndex}
}
```

### 5. Handle Empty States

```go
// ✅ GOOD: Check for empty
if len(m.tasks) == 0 {
    return nil
}
```

## References

- [Input Handling](./input-handling.md)
- [Component System](./components.md)
- [State Management](./state-management.md)
- [Message Flow](./messages.md)
