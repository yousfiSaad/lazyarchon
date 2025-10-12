# Search System

## Overview

LazyArchon implements a **powerful inline search system** that allows users to quickly find tasks by typing search queries directly in the task list. The search system supports real-time filtering, match highlighting, and vim-like match navigation.

## Search Features

1. **Inline Search Mode**: Type directly in the task list without opening a modal
2. **Real-time Filtering**: Results update as you type
3. **Match Highlighting**: Search terms highlighted in task list and details
4. **Match Navigation**: Jump between matches with n/N (vim-style)
5. **Multiple Search Targets**: Searches task title, description, feature, and ID
6. **Case-Insensitive**: Searches ignore case for better usability

## Search Activation

### Activating Search

**Keys**: `/` or `Ctrl+F`

**Handler**: `input_handlers_search.go`

```go
func (m *Model) HandleActivateSearchKey(key string) (tea.Cmd, bool) {
    // Only in task mode (not project mode)
    if !m.programContext.IsProjectModeActive() && !m.programContext.SearchMode {
        m.programContext.SetSearchMode(true)
        m.programContext.UpdateSearchInput(m.taskSearchQuery)
        return nil, true
    }
    return nil, false
}
```

**Activation Flow**:
```
1. User presses '/'
   ↓
2. Search mode activated
   ↓
3. Input captures all keystrokes
   ↓
4. Query updates in real-time
   ↓
5. Task list filters to matches
```

### Search Mode State

```go
// ProgramContext holds search mode flag
type ProgramContext struct {
    SearchMode bool  // true when search input is active
}

// Model holds search query and matches
type Model struct {
    taskSearchActive    bool     // true when search has results
    taskSearchQuery     string   // Current search query
    taskMatchingIndices []int    // Indices of matching tasks
    taskTotalMatches    int      // Total number of matches
    taskCurrentMatch    int      // Current match index (for n/N)
}
```

## Search Input Handling

### Input Mode

When search mode is active, all keyboard input is captured:

```go
func (m *Model) HandleKeyPress(key string) tea.Cmd {
    // PRIORITY 1: Global keys (Ctrl+C, ?)
    if cmd, handled := m.handleGlobalKeys(key); handled {
        return cmd
    }

    // PRIORITY 2: Search input mode (captures typing)
    if m.programContext.SearchMode {
        return m.handleInlineSearchInput(key)
    }

    // ... other priorities
}
```

### Search Input Handler

```go
func (m *Model) handleInlineSearchInput(key string) tea.Cmd {
    switch key {
    case "enter":
        // Commit search and exit search mode
        m.programContext.SetSearchMode(false)
        return nil

    case "esc":
        // Cancel search and clear query
        m.programContext.SetSearchMode(false)
        m.taskSearchQuery = ""
        m.updateSearchMatches()
        m.updateTaskListComponents(m.GetSortedTasks())
        return nil

    case "ctrl+u":
        // Clear search query
        m.taskSearchQuery = ""
        m.programContext.UpdateSearchInput("")
        m.updateSearchMatches()
        m.updateTaskListComponents(m.GetSortedTasks())
        return nil

    case "backspace":
        // Delete last character
        if len(m.taskSearchQuery) > 0 {
            m.taskSearchQuery = m.taskSearchQuery[:len(m.taskSearchQuery)-1]
            m.programContext.UpdateSearchInput(m.taskSearchQuery)
            m.updateSearchMatches()
            m.updateTaskListComponents(m.GetSortedTasks())
        }
        return nil

    default:
        // Add typed character to query
        if len(key) == 1 && isTypableCharacter(key) {
            m.taskSearchQuery += key
            m.programContext.UpdateSearchInput(m.taskSearchQuery)
            m.updateSearchMatches()
            m.updateTaskListComponents(m.GetSortedTasks())
        }
        return nil
    }
}

func isTypableCharacter(key string) bool {
    if len(key) != 1 {
        return false
    }
    char := key[0]
    // Alphanumeric, space, and common symbols
    return (char >= 'a' && char <= 'z') ||
           (char >= 'A' && char <= 'Z') ||
           (char >= '0' && char <= '9') ||
           char == ' ' || char == '-' || char == '_'
}
```

## Search Query Processing

### Update Search Matches

```go
func (m *Model) updateSearchMatches() {
    if m.taskSearchQuery == "" {
        // Clear search
        m.taskSearchActive = false
        m.taskMatchingIndices = nil
        m.taskTotalMatches = 0
        m.taskCurrentMatch = 0
        return
    }

    // Find matching tasks
    sortedTasks := m.GetSortedTasks()
    var matchingIndices []int

    query := strings.ToLower(m.taskSearchQuery)

    for i, task := range sortedTasks {
        if m.taskMatchesQuery(task, query) {
            matchingIndices = append(matchingIndices, i)
        }
    }

    // Update state
    m.taskMatchingIndices = matchingIndices
    m.taskTotalMatches = len(matchingIndices)
    m.taskSearchActive = m.taskTotalMatches > 0

    // Set current match
    if m.taskTotalMatches > 0 {
        m.taskCurrentMatch = 0
        // Jump to first match
        firstMatchIndex := m.taskMatchingIndices[0]
        m.setSelectedTask(firstMatchIndex)
    }
}
```

### Task Matching Logic

```go
func (m *Model) taskMatchesQuery(task archon.Task, query string) bool {
    // Search in multiple fields
    searchableText := strings.ToLower(strings.Join([]string{
        task.Title,
        task.Description,
        task.Feature,
        task.ID,
    }, " "))

    return strings.Contains(searchableText, query)
}
```

## Match Navigation

### Next Match (n)

**Handler**:
```go
func (m *Model) HandleNextSearchMatchKey(key string) (tea.Cmd, bool) {
    if !m.programContext.IsProjectModeActive() && m.taskSearchActive && m.taskTotalMatches > 0 {
        cmd := m.nextSearchMatch()
        return cmd, true
    }
    return nil, false
}

func (m *Model) nextSearchMatch() tea.Cmd {
    // Cycle to next match
    m.taskCurrentMatch = (m.taskCurrentMatch + 1) % m.taskTotalMatches
    matchIndex := m.taskMatchingIndices[m.taskCurrentMatch]

    // Jump to match
    return m.setSelectedTask(matchIndex)
}
```

### Previous Match (N)

**Handler**:
```go
func (m *Model) HandlePrevSearchMatchKey(key string) (tea.Cmd, bool) {
    if !m.programContext.IsProjectModeActive() && m.taskSearchActive && m.taskTotalMatches > 0 {
        cmd := m.previousSearchMatch()
        return cmd, true
    }
    return nil, false
}

func (m *Model) previousSearchMatch() tea.Cmd {
    // Cycle to previous match
    m.taskCurrentMatch = (m.taskCurrentMatch - 1 + m.taskTotalMatches) % m.taskTotalMatches
    matchIndex := m.taskMatchingIndices[m.taskCurrentMatch]

    // Jump to match
    return m.setSelectedTask(matchIndex)
}
```

### Match Cycling

Matches wrap around:
```
Matches: [2, 5, 8, 12]
Current: 3 (index 12)

Press 'n' → Current: 0 (index 2) - wraps to first
Press 'N' → Current: 3 (index 12) - wraps to last
```

## Search Highlighting

### Task List Highlighting

**Component**: `components/tasklist/`

Tasks are highlighted if they match the search query:

```go
func (m *Model) renderTask(task archon.Task, index int) string {
    // Check if task matches search
    isMatch := false
    if m.searchActive && m.searchQuery != "" {
        isMatch = m.taskMatchesQuery(task, m.searchQuery)
    }

    // Highlight matching tasks
    style := m.getTaskStyle(task, index == m.selectedIndex, isMatch)

    return style.Render(m.formatTaskLine(task))
}

func (m *Model) getTaskStyle(task archon.Task, selected bool, isMatch bool) lipgloss.Style {
    baseStyle := lipgloss.NewStyle()

    if selected {
        // Selected task - bold
        baseStyle = baseStyle.Bold(true)
    }

    if isMatch {
        // Matching task - highlighted background
        baseStyle = baseStyle.Background(lipgloss.Color("235"))
    }

    return baseStyle
}
```

### Task Details Highlighting

**Component**: `components/taskdetails/`

Search terms are highlighted in the detail view:

```go
func (m *Model) updateContent() {
    if m.selectedTask == nil {
        m.panelCore.SetContent("No task selected")
        return
    }

    content := m.contentGenerator.GenerateContent(
        m.selectedTask,
        m.searchQuery,  // Pass search query for highlighting
        m.searchActive,
    )

    m.panelCore.SetContent(content)
}
```

**Content Generator**:
```go
func (g *ContentGenerator) GenerateContent(task *archon.Task, searchQuery string, searchActive bool) string {
    var sections []string

    // Title
    titleSection := g.renderTitle(task.Title, searchQuery, searchActive)
    sections = append(sections, titleSection)

    // Description
    descSection := g.renderDescription(task.Description, searchQuery, searchActive)
    sections = append(sections, descSection)

    // ... other sections

    return strings.Join(sections, "\n\n")
}

func (g *ContentGenerator) renderDescription(text string, query string, searchActive bool) string {
    if !searchActive || query == "" {
        return text
    }

    // Highlight search terms
    highlightStyle := lipgloss.NewStyle().
        Background(lipgloss.Color("yellow")).
        Foreground(lipgloss.Color("black"))

    highlighted := strings.ReplaceAll(
        text,
        query,
        highlightStyle.Render(query),
    )

    return highlighted
}
```

## Clearing Search

### Clear Search Key

**Keys**: `Ctrl+X` or `Ctrl+L`

**Handler**:
```go
func (m *Model) HandleClearSearchKey(key string) (tea.Cmd, bool) {
    if !m.programContext.IsProjectModeActive() && (m.taskSearchActive || m.programContext.SearchMode) {
        m.clearSearch()
        return nil, true
    }
    return nil, false
}

func (m *Model) clearSearch() {
    // Clear all search state
    m.taskSearchActive = false
    m.taskSearchQuery = ""
    m.taskMatchingIndices = nil
    m.taskTotalMatches = 0
    m.taskCurrentMatch = 0

    // Exit search mode if active
    m.programContext.SetSearchMode(false)

    // Update components to remove highlighting
    m.updateTaskListComponents(m.GetSortedTasks())
    m.updateTaskDetailsComponent()
}
```

## Search Messages

### Search State Messages

Components receive search state via update messages:

```go
// TaskList receives search state
type TaskListUpdateMsg struct {
    Tasks        []archon.Task
    SearchQuery  string
    SearchActive bool
    Loading      bool
    Error        string
}

// TaskDetails receives search state
type TaskDetailsUpdateMsg struct {
    SelectedTask *archon.Task
    SearchQuery  string
    SearchActive bool
}
```

### Update Flow

```
1. User types in search mode
   ↓
2. taskSearchQuery updated
   ↓
3. updateSearchMatches() finds matches
   ↓
4. updateTaskListComponents() sends search state to TaskList
   ↓
5. TaskList highlights matching tasks
   ↓
6. updateTaskDetailsComponent() sends search state to TaskDetails
   ↓
7. TaskDetails highlights search terms
```

## Search UI Indicators

### Status Bar Search Indicator

When search is active, the status bar shows:

```go
func (m *Model) renderStatusBar() string {
    var segments []string

    // Search indicator
    if m.programContext.SearchMode {
        searchText := fmt.Sprintf("Search: %s", m.taskSearchQuery)
        segments = append(segments, searchText)
    } else if m.taskSearchActive {
        matchText := fmt.Sprintf("Match %d/%d", m.taskCurrentMatch+1, m.taskTotalMatches)
        segments = append(segments, matchText)
    }

    // ... other segments

    return strings.Join(segments, " | ")
}
```

**Display Examples**:
```
Search: auth          (while typing)
Match 2/5            (after committing search)
```

### Header Search Indicator

The header can also show search state:

```go
func (m *Model) renderHeader() string {
    title := "Tasks"

    if m.taskSearchActive {
        title += fmt.Sprintf(" (filtered: %d matches)", m.taskTotalMatches)
    }

    return headerStyle.Render(title)
}
```

## Search Edge Cases

### Empty Query

```go
if m.taskSearchQuery == "" {
    // No search - show all tasks
    m.taskSearchActive = false
    m.updateTaskListComponents(m.GetSortedTasks())
    return
}
```

### No Matches

```go
if m.taskTotalMatches == 0 {
    // No matches found
    m.taskSearchActive = true  // Still searching
    m.taskCurrentMatch = 0
    // Don't change selection
    return
}
```

### Search While Filtered

Search works with other filters (status, feature):

```go
func (m *Model) GetSortedTasks() []archon.Task {
    tasks := m.programContext.Tasks

    // Apply status filter
    if m.statusFilterActive {
        tasks = filterByStatus(tasks, m.statusFilters)
    }

    // Apply feature filter
    if m.featureFilters != nil {
        tasks = filterByFeature(tasks, m.featureFilters)
    }

    // Apply search filter
    if m.taskSearchActive && m.taskSearchQuery != "" {
        tasks = filterBySearch(tasks, m.taskSearchQuery)
    }

    // Sort
    tasks = sortTasks(tasks, m.sortMode)

    return tasks
}
```

## Search Performance

### Incremental Search

Search updates as you type (no debouncing needed):

```go
func (m *Model) handleInlineSearchInput(key string) tea.Cmd {
    // Update query
    m.taskSearchQuery += key

    // Immediately update matches
    m.updateSearchMatches()

    // Update UI
    m.updateTaskListComponents(m.GetSortedTasks())

    return nil
}
```

**Why No Debouncing**:
- Task lists are typically small (< 1000 tasks)
- Filtering is fast (simple string contains)
- Real-time feedback is better UX

### Case-Insensitive Search

All searches are case-insensitive for better UX:

```go
func (m *Model) taskMatchesQuery(task archon.Task, query string) bool {
    // Convert both to lowercase
    searchableText := strings.ToLower(strings.Join([]string{
        task.Title,
        task.Description,
        task.Feature,
        task.ID,
    }, " "))

    query = strings.ToLower(query)

    return strings.Contains(searchableText, query)
}
```

## Search Handlers

All search handlers are in `input_handlers_search.go`:

```go
// Search activation
HandleActivateSearchKey(key)      // /, Ctrl+F

// Search clearing
HandleClearSearchKey(key)          // Ctrl+X, Ctrl+L

// Match navigation
HandleNextSearchMatchKey(key)      // n
HandlePrevSearchMatchKey(key)      // N
```

## Search Keys Reference

| Key | Action | Context |
|-----|--------|---------|
| `/` | Activate search | Task mode |
| `Ctrl+F` | Activate search | Task mode |
| `a-z, 0-9, space` | Type search query | Search mode |
| `Backspace` | Delete character | Search mode |
| `Ctrl+U` | Clear query | Search mode |
| `Enter` | Commit search | Search mode |
| `Esc` | Cancel search | Search mode |
| `n` | Next match | After search |
| `N` | Previous match | After search |
| `Ctrl+X` | Clear search | After search |
| `Ctrl+L` | Clear search | After search |

## Search Flow Diagram

```
┌─────────────────────────────────────────────────┐
│ User presses '/'                                │
└────────────────┬────────────────────────────────┘
                 ↓
┌─────────────────────────────────────────────────┐
│ Search mode activated                           │
│ - programContext.SearchMode = true              │
│ - Keyboard input captured                       │
└────────────────┬────────────────────────────────┘
                 ↓
┌─────────────────────────────────────────────────┐
│ User types query                                │
│ - Each keystroke appends to taskSearchQuery     │
│ - updateSearchMatches() called                  │
└────────────────┬────────────────────────────────┘
                 ↓
┌─────────────────────────────────────────────────┐
│ Find matches                                    │
│ - Filter tasks by query                         │
│ - Store matching indices                        │
│ - Jump to first match                           │
└────────────────┬────────────────────────────────┘
                 ↓
┌─────────────────────────────────────────────────┐
│ Update UI                                       │
│ - TaskList highlights matches                   │
│ - TaskDetails highlights search terms           │
│ - Status bar shows "Match 1/N"                  │
└────────────────┬────────────────────────────────┘
                 ↓
┌─────────────────────────────────────────────────┐
│ User presses Enter or Esc                       │
│ - Search mode deactivated                       │
│ - Search remains active (can use n/N)           │
└────────────────┬────────────────────────────────┘
                 ↓
┌─────────────────────────────────────────────────┐
│ Navigate matches                                │
│ - Press 'n' for next match                      │
│ - Press 'N' for previous match                  │
│ - Selection jumps to match                      │
└────────────────┬────────────────────────────────┘
                 ↓
┌─────────────────────────────────────────────────┐
│ Clear search (Ctrl+X or Ctrl+L)                 │
│ - taskSearchActive = false                      │
│ - Clear all search state                        │
│ - Show all tasks                                │
└─────────────────────────────────────────────────┘
```

## Best Practices

### 1. Always Update Matches After Query Change

```go
// ✅ GOOD: Update matches immediately
m.taskSearchQuery += key
m.updateSearchMatches()
m.updateTaskListComponents(m.GetSortedTasks())

// ❌ BAD: Update query without refreshing matches
m.taskSearchQuery += key
```

### 2. Send Search State to Components

```go
// ✅ GOOD: Include search state
updateMsg := tasklist.TaskListUpdateMsg{
    Tasks:        sortedTasks,
    SearchQuery:  m.taskSearchQuery,
    SearchActive: m.taskSearchActive,
}

// ❌ BAD: Components can't highlight
updateMsg := tasklist.TaskListUpdateMsg{
    Tasks: sortedTasks,
}
```

### 3. Validate Match Index Before Navigation

```go
// ✅ GOOD: Check for matches
if m.taskSearchActive && m.taskTotalMatches > 0 {
    return m.nextSearchMatch()
}

// ❌ BAD: No validation
return m.nextSearchMatch()  // Might crash!
```

### 4. Clear All Search State

```go
// ✅ GOOD: Clear everything
m.taskSearchActive = false
m.taskSearchQuery = ""
m.taskMatchingIndices = nil
m.taskTotalMatches = 0
m.taskCurrentMatch = 0
m.programContext.SetSearchMode(false)

// ❌ BAD: Partial clear
m.taskSearchQuery = ""
```

### 5. Use Case-Insensitive Matching

```go
// ✅ GOOD: Case-insensitive
query := strings.ToLower(m.taskSearchQuery)
text := strings.ToLower(task.Title)
return strings.Contains(text, query)

// ❌ BAD: Case-sensitive
return strings.Contains(task.Title, m.taskSearchQuery)
```

## Search Testing

### Test Search Activation

```go
func TestActivateSearch(t *testing.T) {
    m := NewTestModel()

    cmd, handled := m.HandleActivateSearchKey("/")

    assert.True(t, handled)
    assert.True(t, m.programContext.SearchMode)
}
```

### Test Match Finding

```go
func TestFindMatches(t *testing.T) {
    m := NewTestModel()
    m.programContext.Tasks = []archon.Task{
        {Title: "Fix auth bug"},
        {Title: "Add feature"},
        {Title: "Fix auth test"},
    }

    m.taskSearchQuery = "auth"
    m.updateSearchMatches()

    assert.Equal(t, 2, m.taskTotalMatches)
    assert.Equal(t, []int{0, 2}, m.taskMatchingIndices)
}
```

### Test Match Navigation

```go
func TestNextMatch(t *testing.T) {
    m := NewTestModel()
    m.taskMatchingIndices = []int{1, 3, 7}
    m.taskTotalMatches = 3
    m.taskCurrentMatch = 0

    m.nextSearchMatch()

    assert.Equal(t, 1, m.taskCurrentMatch)
}
```

## References

- [Input Handling](./input-handling.md)
- [Navigation System](./navigation.md)
- [Component System](./components.md)
- [State Management](./state-management.md)
