package taskdetails

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/internal/archon"
	"github.com/yousfisaad/lazyarchon/internal/shared/viewport"
)

// TaskDetailsUpdateMsg is sent to update the task details data
type TaskDetailsUpdateMsg struct {
	SelectedTask *archon.Task // The task to display
	SearchQuery  string       // Current search query for highlighting
	SearchActive bool         // Whether search highlighting is active
}

// NOTE: TaskDetailsSetActiveMsg removed - components read active state from UIState directly

// TaskDetailsResizeMsg is sent to update the component dimensions
type TaskDetailsResizeMsg struct {
	Width  int // New panel width
	Height int // New panel height
}

// Helper functions to create message commands

// UpdateTaskDetails creates a command to update the task details
func UpdateTaskDetails(selectedTask *archon.Task, searchQuery string, searchActive bool) tea.Cmd {
	return func() tea.Msg {
		return TaskDetailsUpdateMsg{
			SelectedTask: selectedTask,
			SearchQuery:  searchQuery,
			SearchActive: searchActive,
		}
	}
}

// NOTE: SetTaskDetailsActive helper removed - components read active state from UIState directly

// ResizeTaskDetails creates a command to resize the task details component
func ResizeTaskDetails(width, height int) tea.Cmd {
	return func() tea.Msg {
		return TaskDetailsResizeMsg{
			Width:  width,
			Height: height,
		}
	}
}

// TaskDetailsScrollMsg is sent to scroll the task details content
type TaskDetailsScrollMsg struct {
	Direction viewport.ScrollDirection // Direction and amount to scroll
}

// ScrollTaskDetails creates a command to scroll the task details
func ScrollTaskDetails(direction viewport.ScrollDirection) tea.Cmd {
	return func() tea.Msg {
		return TaskDetailsScrollMsg{Direction: direction}
	}
}

// TaskDetailsScrollPositionChangedMsg is broadcast when scroll position changes
type TaskDetailsScrollPositionChangedMsg struct {
	Position string // Use detailspanel.ScrollPosition* constants
}

// Ensure all message types implement tea.Msg
var (
	_ tea.Msg = TaskDetailsUpdateMsg{}
	// NOTE: TaskDetailsSetActiveMsg interface check removed - message type deleted
	_ tea.Msg = TaskDetailsResizeMsg{}
	_ tea.Msg = TaskDetailsScrollMsg{}
	_ tea.Msg = TaskDetailsScrollPositionChangedMsg{}
)
