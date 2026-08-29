package tasks

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/interfaces"
)

// =============================================================================
// TASK DOMAIN MESSAGES
// =============================================================================
// Messages representing task-related domain events and operations

// TasksLoadedMsg is sent when tasks are loaded from the API
type TasksLoadedMsg struct {
	Tasks []interfaces.Task
	Error error
}

// TaskCreateMsg is sent when a task is created
type TaskCreateMsg struct {
	Task  *interfaces.Task
	Error error
}

// TaskUpdateMsg is sent when a task is updated
type TaskUpdateMsg struct {
	Task  *interfaces.Task
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
	_ tea.Msg = TaskCreateMsg{}
	_ tea.Msg = TaskUpdateMsg{}
	_ tea.Msg = TaskDeleteMsg{}
)
