package tasklist

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/internal/archon"
)

// TaskListUpdateMsg is sent to update the task list data
type TaskListUpdateMsg struct {
	Tasks   []archon.Task // Updated task data
	Loading bool          // Whether loading is in progress
	Error   string        // Error message if any
}

// TaskListSelectMsg is sent to select a specific task by index
type TaskListSelectMsg struct {
	Index int // Index to select
}

// TaskListSelectionChangedMsg is sent when the selection changes
type TaskListSelectionChangedMsg struct {
	Index int // New selected index
}

// TaskListSearchMsg is sent to update search state
type TaskListSearchMsg struct {
	Query  string // Search query
	Active bool   // Whether search is active
}

// TaskListFilterMsg is sent to update filter state
type TaskListFilterMsg struct {
	Feature string // Feature filter
	Status  string // Status filter
}

// ScrollDirection represents the direction of scrolling
type ScrollDirection int

const (
	ScrollUp ScrollDirection = iota
	ScrollDown
	ScrollToTop
	ScrollToBottom
	ScrollFastUp   // Fast scroll up (4 lines)
	ScrollFastDown // Fast scroll down (4 lines)
	ScrollPageUp
	ScrollPageDown
)

// TaskListScrollMsg is sent to handle scrolling operations
type TaskListScrollMsg struct {
	Direction ScrollDirection
}

// NOTE: TaskListSetActiveMsg removed - components read active state from UIState directly

// TaskListResizeMsg is sent to update the task list component dimensions
type TaskListResizeMsg struct {
	Width  int // New width for the component
	Height int // New height for the component
}

// Helper functions to create message commands

// UpdateTaskList creates a command to update the task list
func UpdateTaskList(tasks []archon.Task, loading bool, error string) tea.Cmd {
	return func() tea.Msg {
		return TaskListUpdateMsg{
			Tasks:   tasks,
			Loading: loading,
			Error:   error,
		}
	}
}

// SelectTask creates a command to select a task by index
func SelectTask(index int) tea.Cmd {
	return func() tea.Msg {
		return TaskListSelectMsg{Index: index}
	}
}

// UpdateSearch creates a command to update search state
func UpdateSearch(query string, active bool) tea.Cmd {
	return func() tea.Msg {
		return TaskListSearchMsg{
			Query:  query,
			Active: active,
		}
	}
}

// UpdateFilter creates a command to update filter state
func UpdateFilter(feature, status string) tea.Cmd {
	return func() tea.Msg {
		return TaskListFilterMsg{
			Feature: feature,
			Status:  status,
		}
	}
}

// ScrollTaskList creates a command to scroll the task list
func ScrollTaskList(direction ScrollDirection) tea.Cmd {
	return func() tea.Msg {
		return TaskListScrollMsg{Direction: direction}
	}
}

// NOTE: SetTaskListActive helper removed - components read active state from UIState directly

// ResizeTaskList creates a command to resize the task list component
func ResizeTaskList(width, height int) tea.Cmd {
	return func() tea.Msg {
		return TaskListResizeMsg{
			Width:  width,
			Height: height,
		}
	}
}

// Ensure all message types implement tea.Msg
var (
	_ tea.Msg = TaskListUpdateMsg{}
	_ tea.Msg = TaskListSelectMsg{}
	_ tea.Msg = TaskListSelectionChangedMsg{}
	_ tea.Msg = TaskListSearchMsg{}
	_ tea.Msg = TaskListFilterMsg{}
	_ tea.Msg = TaskListScrollMsg{}
	// NOTE: TaskListSetActiveMsg interface check removed - message type deleted
	_ tea.Msg = TaskListResizeMsg{}
)
