package realtime

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/internal/archon"
)

// =============================================================================
// REALTIME DOMAIN MESSAGES
// =============================================================================
// Messages representing WebSocket real-time events and connection state

// RealtimeConnectedMsg is sent when WebSocket connection is established
type RealtimeConnectedMsg struct{}

// RealtimeDisconnectedMsg is sent when WebSocket connection is lost
type RealtimeDisconnectedMsg struct {
	Error error
}

// RealtimeTaskUpdateMsg is sent when a task is updated via WebSocket
type RealtimeTaskUpdateMsg struct {
	TaskID string
	Task   archon.Task
	Old    *archon.Task
}

// RealtimeTaskCreateMsg is sent when a task is created via WebSocket
type RealtimeTaskCreateMsg struct {
	Task archon.Task
}

// RealtimeTaskDeleteMsg is sent when a task is deleted via WebSocket
type RealtimeTaskDeleteMsg struct {
	TaskID string
	Task   archon.Task
}

// RealtimeProjectUpdateMsg is sent when a project is updated via WebSocket
type RealtimeProjectUpdateMsg struct {
	ProjectID string
	Project   archon.Project
	Old       *archon.Project
}

// Ensure all message types implement tea.Msg
var (
	_ tea.Msg = RealtimeConnectedMsg{}
	_ tea.Msg = RealtimeDisconnectedMsg{}
	_ tea.Msg = RealtimeTaskUpdateMsg{}
	_ tea.Msg = RealtimeTaskCreateMsg{}
	_ tea.Msg = RealtimeTaskDeleteMsg{}
	_ tea.Msg = RealtimeProjectUpdateMsg{}
)
