package tasks

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/internal/archon"
)

// =============================================================================
// TASK DOMAIN MESSAGES
// =============================================================================
// Messages representing task-related domain events and operations

// TasksLoadedMsg is sent when tasks are loaded from the API
type TasksLoadedMsg struct {
	Tasks []archon.Task
	Error error
}

// TaskUpdateMsg is sent when a task is updated
type TaskUpdateMsg struct {
	Task  *archon.Task
	Error error
}

// TaskDeleteMsg is sent when a task is deleted/archived
type TaskDeleteMsg struct {
	TaskID string
	Error  error
}

// Ensure all message types implement tea.Msg
var (
	_ tea.Msg = TasksLoadedMsg{}
	_ tea.Msg = TaskUpdateMsg{}
	_ tea.Msg = TaskDeleteMsg{}
)
