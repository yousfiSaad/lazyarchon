# Modal System

## Overview

LazyArchon uses a **consistent modal architecture** where all modals follow the same lifecycle pattern. Modals are overlay UI components that temporarily take focus for user input or information display.

## Modal Components

All modals are located in `components/modals/*/` and follow a standard pattern.

### Available Modals

1. **Help Modal** (`help/`) - Keyboard shortcuts reference
2. **Task Edit Modal** (`taskedit/`) - Edit task properties (status, priority, feature)
3. **Status Modal** (`status/`) - Quick task status change
4. **Status Filter Modal** (`statusfilter/`) - Filter tasks by status
5. **Feature Modal** (`feature/`) - Filter tasks by feature tags
6. **Confirmation Modal** (`confirmation/`) - Yes/No confirmation dialogs

## Modal Lifecycle Pattern

All modals follow this standard lifecycle:

```
Inactive → Show Message → Active (Focused) → User Action → Hide Message → Inactive
```

### Lifecycle Messages

Every modal implements these message types:

```go
// Activation
type Show[Modal]Msg struct {
    // Initial data for modal
}

// Deactivation
type Hide[Modal]Msg struct{}

// Confirmation (internal broadcast)
type [Modal]ShownMsg struct{}
type [Modal]HiddenMsg struct{}
```

### Lifecycle Flow Example

```
1. User presses shortcut (e.g., '?')
   ↓
2. Handler creates ShowHelpModalMsg
   ↓
3. Modal.Update receives message
   ↓
4. Modal sets active=true, focused=true
   ↓
5. Modal broadcasts HelpModalShownMsg
   ↓
6. View() renders modal overlay
   ↓
7. User presses 'q' or 'Esc'
   ↓
8. Modal creates HideHelpModalMsg
   ↓
9. Modal sets active=false, focused=false
   ↓
10. Modal broadcasts HelpModalHiddenMsg
```

## Base Modal Pattern

All modals inherit from `base.BaseComponent` and follow this structure:

```go
type Model struct {
    base.BaseComponent

    // UI state
    width  int
    height int
    // ... modal-specific state

    // Dependencies
    // ... injected dependencies
}

// Lifecycle management
func (m *Model) IsActive() bool      // Inherited from BaseComponent
func (m *Model) SetActive(bool)      // Inherited from BaseComponent
func (m *Model) IsFocused() bool     // Inherited from BaseComponent
func (m *Model) SetFocus(bool)       // Inherited from BaseComponent

// Standard methods
func (m *Model) Init() tea.Cmd
func (m *Model) Update(msg tea.Msg) tea.Cmd
func (m *Model) View() string
```

## Modal Implementations

### 1. Help Modal

**Purpose**: Display keyboard shortcuts and help information

**Location**: `components/modals/help/`

**Messages**:
```go
type ShowHelpModalMsg struct{}
type HideHelpModalMsg struct{}
type HelpModalShownMsg struct{}
type HelpModalHiddenMsg struct{}
type HelpModalScrollMsg struct {
    Direction viewport.ScrollDirection
}
```

**State**:
```go
type Model struct {
    base.BaseComponent

    viewport     viewport.Model  // Scrollable help content
    width        int
    height       int
    contentWidth int
}
```

**Key Features**:
- Scrollable viewport with help content
- Arrow keys / j/k for scrolling
- Esc / q to close
- Shows all keyboard shortcuts organized by category

**Usage**:
```go
// Show help
return func() tea.Msg {
    return help.ShowHelpModalMsg{}
}

// Help modal handles its own lifecycle
// User presses 'q' or 'Esc' to close
```

### 2. Task Edit Modal

**Purpose**: Edit task properties (status, priority, feature)

**Location**: `components/modals/taskedit/`

**Messages**:
```go
type ShowTaskEditModalMsg struct {
    TaskID            string
    CurrentStatus     string
    CurrentPriority   int
    CurrentFeature    string
    FocusField        FieldType  // Which field to focus initially
    AvailableFeatures []string
}

type HideTaskEditModalMsg struct{}

type TaskPropertiesUpdatedMsg struct {
    TaskID      string
    NewStatus   *string  // nil = no change
    NewPriority *int     // nil = no change
    NewFeature  *string  // nil = no change
}
```

**Field Types**:
```go
type FieldType int

const (
    FieldStatus   FieldType = iota  // 0
    FieldPriority                    // 1
    FieldFeature                     // 2
)
```

**State**:
```go
type Model struct {
    base.BaseComponent

    // Task info
    taskID string

    // Multi-field form state
    activeField FieldType  // Currently focused field

    // Field values (working state)
    statusValue   string
    priorityValue int
    featureValue  string

    // Original values (for change detection)
    originalStatus   string
    originalPriority int
    originalFeature  string

    // Status field state
    statusIndex int  // Index in statusOptions

    // Priority field state
    priorityEditMode bool
    priorityInput    string

    // Feature field state
    availableFeatures    []string
    selectedFeatureIndex int
    featureSelectionMode bool
    isCreatingNew        bool
    newFeatureName       string
}
```

**Key Features**:
- Multi-field form with Tab navigation
- Status field: Arrow keys cycle through options
- Priority field: Arrow keys adjust ±10, typing enters specific value
- Feature field: Select from existing or create new
- Only sends changes (nil for unchanged fields)
- Enter to confirm, Esc/q to cancel

**Field Behaviors**:

**Status Field**:
```
Arrow Up/Down: Cycle through todo/doing/review/done
Display: [Status] Current Status ▼
```

**Priority Field**:
```
Arrow Up/Down: Adjust by 10
Number keys: Enter specific value (e.g., "75" + Enter)
Display: [Priority] 50 ▲▼
```

**Feature Field**:
```
Space: Open feature selection
Arrow Up/Down: Navigate features
Enter: Select feature
n: Create new feature
Backspace: Clear feature
Display: [Feature] feature-name
```

**Usage Example**:
```go
// Show modal to edit status
showMsg := taskedit.ShowTaskEditModalMsg{
    TaskID:            task.ID,
    CurrentStatus:     task.Status,
    CurrentPriority:   task.TaskOrder,
    CurrentFeature:    task.Feature,
    FocusField:        taskedit.FieldStatus,  // Start on status
    AvailableFeatures: m.GetUniqueFeatures(),
}
return func() tea.Msg { return showMsg }

// Handle result
case taskedit.TaskPropertiesUpdatedMsg:
    // Apply changes to task
    if msg.NewStatus != nil {
        task.Status = *msg.NewStatus
    }
    if msg.NewPriority != nil {
        task.TaskOrder = *msg.NewPriority
    }
    // ... update via API
```

### 3. Status Modal

**Purpose**: Quick single-field status change (simpler than TaskEdit)

**Location**: `components/modals/status/`

**Messages**:
```go
type ShowStatusModalMsg struct {
    CurrentStatus string
}

type HideStatusModalMsg struct{}

type StatusSelectedMsg struct {
    NewStatus string
}
```

**State**:
```go
type Model struct {
    base.BaseComponent

    currentStatus string
    selectedIndex int  // Index in status options
    width         int
    height        int
}
```

**Key Features**:
- Single-purpose: status change only
- Arrow keys or j/k to navigate
- Enter to confirm
- Esc/q to cancel
- Simpler than TaskEdit when only status needed

**Usage**:
```go
// Quick status change
return func() tea.Msg {
    return status.ShowStatusModalMsg{
        CurrentStatus: task.Status,
    }
}

// Handle selection
case status.StatusSelectedMsg:
    // Update task status
    newStatus := msg.NewStatus
    // ... API call
```

### 4. Feature Modal

**Purpose**: Filter tasks by feature tags

**Location**: `components/modals/feature/`

**Messages**:
```go
type ShowFeatureModalMsg struct {
    AllFeatures      []string           // All available features
    SelectedFeatures map[string]bool    // Currently selected features
}

type HideFeatureModalMsg struct{}

type FeatureSelectionAppliedMsg struct {
    SelectedFeatures map[string]bool    // Updated selection
}
```

**State**:
```go
type Model struct {
    base.BaseComponent

    allFeatures      []string
    selectedFeatures map[string]bool
    selectedIndex    int
    viewport         viewport.Model

    // Search within features
    searchActive bool
    searchQuery  string
}
```

**Key Features**:
- Multi-select: Toggle features with Space
- Arrow keys / j/k to navigate
- 'a' to select all
- 'c' to clear all
- '/' to search features
- Enter to apply selection
- Esc/q to cancel without applying

**Usage**:
```go
// Show feature filter
showMsg := feature.ShowFeatureModalMsg{
    AllFeatures:      m.GetUniqueFeatures(),
    SelectedFeatures: m.featureFilters,  // Current selection
}
return func() tea.Msg { return showMsg }

// Apply selection
case feature.FeatureSelectionAppliedMsg:
    m.featureFilters = msg.SelectedFeatures
    m.updateTaskListComponents(m.GetSortedTasks())
```

### 5. Status Filter Modal

**Purpose**: Filter tasks by status (multi-select)

**Location**: `components/modals/statusfilter/`

**Messages**:
```go
type ShowStatusFilterModalMsg struct {
    SelectedStatuses map[string]bool
}

type HideStatusFilterModalMsg struct{}

type StatusFilterAppliedMsg struct {
    SelectedStatuses map[string]bool
}
```

**State**:
```go
type Model struct {
    base.BaseComponent

    selectedStatuses map[string]bool
    selectedIndex    int
    width            int
    height           int
}
```

**Key Features**:
- Multi-select: Toggle statuses with Space
- Shows todo/doing/review/done options
- Arrow keys / j/k to navigate
- 'a' to select all
- 'c' to clear all
- Enter to apply
- Esc/q to cancel

**Usage**:
```go
// Show status filter
showMsg := statusfilter.ShowStatusFilterModalMsg{
    SelectedStatuses: m.statusFilters,
}
return func() tea.Msg { return showMsg }

// Apply filter
case statusfilter.StatusFilterAppliedMsg:
    m.statusFilters = msg.SelectedStatuses
    m.statusFilterActive = true
    m.updateTaskListComponents(m.GetSortedTasks())
```

### 6. Confirmation Modal

**Purpose**: Generic yes/no confirmation dialogs

**Location**: `components/modals/confirmation/`

**Messages**:
```go
type ShowConfirmationModalMsg struct {
    Title   string
    Message string
    Type    ConfirmationType
}

type HideConfirmationModalMsg struct{}

type ConfirmationResultMsg struct {
    Confirmed bool
    Type      ConfirmationType
}
```

**Confirmation Types**:
```go
type ConfirmationType int

const (
    ConfirmQuit ConfirmationType = iota
    // ... add more types as needed
)
```

**State**:
```go
type Model struct {
    base.BaseComponent

    title        string
    message      string
    confirmType  ConfirmationType
    selectedYes  bool  // true=Yes, false=No
    width        int
    height       int
}
```

**Key Features**:
- Generic confirmation dialog
- Left/right or h/l to toggle Yes/No
- Enter to confirm selection
- Esc/q defaults to No
- Customizable title and message

**Usage**:
```go
// Show quit confirmation
showMsg := confirmation.ShowConfirmationModalMsg{
    Title:   "Quit LazyArchon",
    Message: "Are you sure you want to quit?",
    Type:    confirmation.ConfirmQuit,
}
return func() tea.Msg { return showMsg }

// Handle result
case confirmation.ConfirmationResultMsg:
    if msg.Confirmed && msg.Type == confirmation.ConfirmQuit {
        return tea.Quit
    }
```

## Modal State Management

### Active State

Modals have two important flags:
- **active**: Modal is visible
- **focused**: Modal can receive input

```go
func (m *Model) Update(msg tea.Msg) tea.Cmd {
    switch msg := msg.(type) {
    case ShowModalMsg:
        m.SetActive(true)   // Make visible
        m.SetFocus(true)    // Receive input
        return m.BroadcastMessage(ModalShownMsg{})

    case HideModalMsg:
        m.SetActive(false)  // Hide
        m.SetFocus(false)   // Stop receiving input
        return m.BroadcastMessage(ModalHiddenMsg{})
    }
}
```

### View Rendering

Modals only render when active:

```go
func (m *Model) View() string {
    if !m.IsActive() {
        return ""  // Don't render when inactive
    }

    // Render modal content
    return m.renderModal()
}
```

### Input Handling

Modals only handle input when focused:

```go
func (m *Model) Update(msg tea.Msg) tea.Cmd {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if !m.IsActive() || !m.IsFocused() {
            return nil  // Ignore input when not focused
        }
        return m.handleKeyPress(msg)
    }
}
```

## Modal Rendering

### Overlay Pattern

Modals render as centered overlays:

```go
func (m *Model) View() string {
    // Create modal content
    content := m.renderContent()

    // Wrap in styled box
    modalBox := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("blue")).
        Padding(1).
        Width(m.width).
        Render(content)

    // Center on screen
    return lipgloss.Place(
        m.programContext.Width,
        m.programContext.Height,
        lipgloss.Center,
        lipgloss.Center,
        modalBox,
    )
}
```

### Dimension Management

Modals handle window resize:

```go
func (m *Model) Update(msg tea.Msg) tea.Cmd {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.updateDimensions(msg.Width, msg.Height)
        if m.IsActive() {
            m.updateContent()  // Regenerate content
        }
        return nil
    }
}
```

## Modal Communication Patterns

### Pattern 1: Show → Action → Hide

**Most common pattern**:

```
1. User presses shortcut
2. Show modal with initial data
3. User interacts with modal
4. User confirms action
5. Modal sends action message
6. Modal hides itself
7. Model handles action
```

**Example**:
```go
// Step 2: Show
case tea.KeyMsg:
    if key == "t" {
        return func() tea.Msg {
            return taskedit.ShowTaskEditModalMsg{
                TaskID: currentTask.ID,
                CurrentStatus: currentTask.Status,
                // ...
            }
        }
    }

// Step 5-6: Action and hide
case taskedit.TaskPropertiesUpdatedMsg:
    // Step 7: Handle action
    return m.updateTaskProperties(msg)
```

### Pattern 2: Show → Display → Hide

**Information display**:

```
1. User presses '?'
2. Show help modal
3. User reads information
4. User presses 'q' or Esc
5. Modal hides
```

**Example**:
```go
case tea.KeyMsg:
    if key == "?" {
        return func() tea.Msg {
            return help.ShowHelpModalMsg{}
        }
    }
```

### Pattern 3: Show → Confirm → Action

**Confirmation flow**:

```
1. User requests dangerous action
2. Show confirmation modal
3. User selects Yes/No
4. Modal sends result
5. Modal hides
6. Model performs or cancels action
```

**Example**:
```go
// Request quit
case tea.KeyMsg:
    if key == "q" {
        return func() tea.Msg {
            return confirmation.ShowConfirmationModalMsg{
                Title: "Quit",
                Message: "Are you sure?",
                Type: confirmation.ConfirmQuit,
            }
        }
    }

// Handle confirmation
case confirmation.ConfirmationResultMsg:
    if msg.Confirmed && msg.Type == confirmation.ConfirmQuit {
        return tea.Quit
    }
```

## Creating a New Modal

### Step 1: Create Directory Structure

```bash
mkdir -p components/modals/mymodal
touch components/modals/mymodal/component.go
touch components/modals/mymodal/messages.go
```

### Step 2: Define Messages

```go
// components/modals/mymodal/messages.go
package mymodal

// Lifecycle messages
type ShowMyModalMsg struct {
    // Initial data
    InitialValue string
}

type HideMyModalMsg struct{}

// Confirmation messages (optional)
type MyModalShownMsg struct{}
type MyModalHiddenMsg struct{}

// Action messages
type MyModalActionMsg struct {
    Result string
}
```

### Step 3: Implement Component

```go
// components/modals/mymodal/component.go
package mymodal

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/yousfisaad/lazyarchon/internal/ui/components/base"
)

const ComponentID = "my-modal"

type Model struct {
    base.BaseComponent

    // State
    value string
    width int
    height int
}

func NewModel(context *base.ComponentContext) *Model {
    baseComponent := base.NewBaseComponent(
        ComponentID,
        base.MyModalComponent,  // Add to ComponentType enum
        context,
    )

    return &Model{
        BaseComponent: baseComponent,
        width:         50,
        height:        10,
    }
}

func (m *Model) Init() tea.Cmd {
    return nil
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
    switch msg := msg.(type) {
    case ShowMyModalMsg:
        m.SetActive(true)
        m.SetFocus(true)
        m.value = msg.InitialValue
        return m.BroadcastMessage(MyModalShownMsg{})

    case HideMyModalMsg:
        m.SetActive(false)
        m.SetFocus(false)
        return m.BroadcastMessage(MyModalHiddenMsg{})

    case tea.KeyMsg:
        if !m.IsActive() || !m.IsFocused() {
            return nil
        }
        return m.handleKeyPress(msg)
    }
    return nil
}

func (m *Model) handleKeyPress(msg tea.KeyMsg) tea.Cmd {
    switch msg.String() {
    case "enter":
        // Confirm action
        return tea.Batch(
            func() tea.Msg {
                return MyModalActionMsg{Result: m.value}
            },
            func() tea.Msg {
                return HideMyModalMsg{}
            },
        )

    case "q", "esc":
        // Cancel
        return func() tea.Msg {
            return HideMyModalMsg{}
        }
    }
    return nil
}

func (m *Model) View() string {
    if !m.IsActive() {
        return ""
    }

    // Render modal content
    content := "My Modal Content: " + m.value

    // Style and center
    // ... lipgloss styling

    return content
}
```

### Step 4: Register in Component Factory

```go
// factories/component_factory.go

type ModalComponents struct {
    // ... existing modals
    MyModal *mymodal.Model
}

func CreateModalComponents(context *base.ComponentContext) ModalComponents {
    return ModalComponents{
        // ... existing modals
        MyModal: mymodal.NewModel(context),
    }
}
```

### Step 5: Wire Up in Model

```go
// model.go

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case mymodal.ShowMyModalMsg:
        return m, m.components.Modals.MyModal.Update(msg)

    case mymodal.MyModalActionMsg:
        return m.handleMyModalAction(msg)
    }
    // ...
}
```

### Step 6: Add Keyboard Shortcut

```go
// input_handlers.go

func (m *Model) handleApplicationKey(key string) (tea.Cmd, bool) {
    switch key {
    case keys.KeyMyShortcut:
        return func() tea.Msg {
            return mymodal.ShowMyModalMsg{
                InitialValue: "default",
            }
        }, true
    }
    return nil, false
}
```

## Modal Best Practices

### 1. Consistent Lifecycle

Always follow the Show → Active → Action → Hide pattern:

```go
// ✅ GOOD: Consistent pattern
case ShowModalMsg:
    m.SetActive(true)
    m.SetFocus(true)
    return m.BroadcastMessage(ModalShownMsg{})

case HideModalMsg:
    m.SetActive(false)
    m.SetFocus(false)
    return m.BroadcastMessage(ModalHiddenMsg{})
```

### 2. Guard Input Handling

Always check active/focused before handling input:

```go
// ✅ GOOD: Guarded input
case tea.KeyMsg:
    if !m.IsActive() || !m.IsFocused() {
        return nil
    }
    return m.handleKeyPress(msg)
```

### 3. Broadcast State Changes

Use BroadcastMessage for lifecycle events:

```go
// ✅ GOOD: Broadcast for observers
return m.BroadcastMessage(ModalShownMsg{})

// ❌ BAD: Direct return
return nil
```

### 4. Separate Action from Hide

Send action message, then hide separately:

```go
// ✅ GOOD: Separate action and hide
return tea.Batch(
    func() tea.Msg { return ActionMsg{} },
    func() tea.Msg { return HideModalMsg{} },
)

// ❌ BAD: Hide without action
return func() tea.Msg { return HideModalMsg{} }
```

### 5. Use Pointers for Optional Changes

Use nil to indicate "no change":

```go
// ✅ GOOD: Nil means unchanged
type UpdateMsg struct {
    NewStatus *string  // nil = don't change
    NewValue  *int     // nil = don't change
}

// ❌ BAD: Can't distinguish empty from unchanged
type UpdateMsg struct {
    NewStatus string  // "" = unchanged or cleared?
}
```

### 6. Clear State on Hide

Reset modal state when hiding:

```go
// ✅ GOOD: Clear state
case HideMyModalMsg:
    m.SetActive(false)
    m.SetFocus(false)
    m.value = ""          // Reset
    m.selectedIndex = 0   // Reset
    return m.BroadcastMessage(MyModalHiddenMsg{})
```

## Modal Testing

### Test Lifecycle

```go
func TestModalLifecycle(t *testing.T) {
    modal := NewModel(testContext)

    // Show
    cmd := modal.Update(ShowMyModalMsg{})
    assert.True(t, modal.IsActive())
    assert.True(t, modal.IsFocused())

    // Hide
    cmd = modal.Update(HideMyModalMsg{})
    assert.False(t, modal.IsActive())
    assert.False(t, modal.IsFocused())
}
```

### Test Input Handling

```go
func TestModalIgnoresInputWhenInactive(t *testing.T) {
    modal := NewModel(testContext)

    // Input when inactive should be ignored
    cmd := modal.Update(tea.KeyMsg{Type: tea.KeyEnter})
    assert.Nil(t, cmd)
}
```

### Test Action Messages

```go
func TestModalSendsActionMessage(t *testing.T) {
    modal := NewModel(testContext)
    modal.Update(ShowMyModalMsg{InitialValue: "test"})

    // Confirm action
    cmd := modal.Update(tea.KeyMsg{Type: tea.KeyEnter})
    msg := cmd()

    actionMsg, ok := msg.(MyModalActionMsg)
    assert.True(t, ok)
    assert.Equal(t, "test", actionMsg.Result)
}
```

## Common Modal Patterns

### Pattern: Single-Select List

```go
type Model struct {
    options       []string
    selectedIndex int
}

func (m *Model) handleKeyPress(msg tea.KeyMsg) tea.Cmd {
    switch msg.String() {
    case "up", "k":
        m.selectedIndex = max(0, m.selectedIndex-1)
    case "down", "j":
        m.selectedIndex = min(len(m.options)-1, m.selectedIndex+1)
    case "enter":
        return m.confirmSelection()
    }
    return nil
}
```

**Examples**: Status Modal

### Pattern: Multi-Select List

```go
type Model struct {
    options  []string
    selected map[string]bool
    index    int
}

func (m *Model) handleKeyPress(msg tea.KeyMsg) tea.Cmd {
    switch msg.String() {
    case "space":
        option := m.options[m.index]
        m.selected[option] = !m.selected[option]
    case "a":
        for _, opt := range m.options {
            m.selected[opt] = true
        }
    case "c":
        m.selected = make(map[string]bool)
    case "enter":
        return m.applySelection()
    }
    return nil
}
```

**Examples**: Feature Modal, Status Filter Modal

### Pattern: Form with Multiple Fields

```go
type Model struct {
    activeField int
    fields      []FieldState
}

func (m *Model) handleKeyPress(msg tea.KeyMsg) tea.Cmd {
    switch msg.String() {
    case "tab":
        m.activeField = (m.activeField + 1) % len(m.fields)
    case "shift+tab":
        m.activeField = (m.activeField - 1 + len(m.fields)) % len(m.fields)
    case "enter":
        return m.submitForm()
    default:
        // Delegate to active field
        return m.fields[m.activeField].HandleInput(msg)
    }
    return nil
}
```

**Examples**: Task Edit Modal

### Pattern: Scrollable Content

```go
type Model struct {
    viewport viewport.Model
}

func (m *Model) handleKeyPress(msg tea.KeyMsg) tea.Cmd {
    switch msg.String() {
    case "up", "k":
        m.viewport.LineUp(1)
    case "down", "j":
        m.viewport.LineDown(1)
    case "pgup":
        m.viewport.ViewUp()
    case "pgdown":
        m.viewport.ViewDown()
    }
    return nil
}
```

**Examples**: Help Modal

## References

- [Base Component](../components/base/component.go)
- [Modal Components](../components/modals/)
- [Message Patterns](./messages.md)
- [Input Handling](./input-handling.md)
