package realtime

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/internal/archon"
	"github.com/yousfisaad/lazyarchon/internal/shared/interfaces"
)

// =============================================================================
// REALTIME DOMAIN COMMANDS
// =============================================================================
// Command functions for WebSocket real-time operations

// InitializeRealtimeCmd sets up the WebSocket connection and event handlers
func InitializeRealtimeCmd(wsClient interfaces.RealtimeClient) tea.Cmd {
	return func() tea.Msg {
		// Attempt to connect
		if err := wsClient.Connect(); err != nil {
			return RealtimeDisconnectedMsg{Error: err}
		}

		return RealtimeConnectedMsg{}
	}
}

// ListenForRealtimeEvents creates a command that listens for WebSocket events
// and converts them to Bubble Tea messages
func ListenForRealtimeEvents(wsClient interfaces.RealtimeClient) tea.Cmd {
	return func() tea.Msg {
		// Block and wait for the next event from the WebSocket client
		eventCh := wsClient.GetEventChannel()

		select {
		case event := <-eventCh:
			// Convert archon message types to ui message types
			switch e := event.(type) {
			case archon.RealtimeTaskCreateMsg:
				return RealtimeTaskCreateMsg{Task: e.Task}
			case archon.RealtimeTaskUpdateMsg:
				return RealtimeTaskUpdateMsg{
					TaskID: e.TaskID,
					Task:   e.Task,
					Old:    e.Old,
				}
			case archon.RealtimeTaskDeleteMsg:
				return RealtimeTaskDeleteMsg{
					TaskID: e.TaskID,
					Task:   e.Task,
				}
			case archon.RealtimeProjectUpdateMsg:
				return RealtimeProjectUpdateMsg{
					ProjectID: e.ProjectID,
					Project:   e.Project,
					Old:       e.Old,
				}
			case archon.RealtimeConnectedMsg:
				return RealtimeConnectedMsg{}
			case archon.RealtimeDisconnectedMsg:
				return RealtimeDisconnectedMsg{Error: e.Error}
			default:
				// Unknown event type, return nil to ignore
				return nil
			}
		}
	}
}
