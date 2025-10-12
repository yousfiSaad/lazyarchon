package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/internal/domain/projects"
	"github.com/yousfisaad/lazyarchon/internal/domain/realtime"
	"github.com/yousfisaad/lazyarchon/internal/domain/tasks"
)

// =============================================================================
// REALTIME MESSAGE HANDLERS
// =============================================================================
// This file contains handlers for WebSocket real-time events

// handleRealtimeMessages processes WebSocket real-time events
func (m *MainModel) handleRealtimeMessages(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case realtime.RealtimeConnectedMsg:
		// Connection established - update UI status
		m.setConnectionStatus(true)
		return m, realtime.ListenForRealtimeEvents(m.wsClient)

	case realtime.RealtimeDisconnectedMsg:
		// Connection lost - update UI status
		m.setConnectionStatus(false)
		return m, nil

	case realtime.RealtimeTaskUpdateMsg:
		// Task was updated - refresh the task list to show changes
		return m, tea.Batch(
			tasks.LoadTasksInterface(m.programContext.ArchonClient, m.programContext.SelectedProjectID),
			realtime.ListenForRealtimeEvents(m.wsClient),
		)

	case realtime.RealtimeTaskCreateMsg:
		// New task was created - refresh the task list
		return m, tea.Batch(
			tasks.LoadTasksInterface(m.programContext.ArchonClient, m.programContext.SelectedProjectID),
			realtime.ListenForRealtimeEvents(m.wsClient),
		)

	case realtime.RealtimeTaskDeleteMsg:
		// Task was deleted - refresh the task list
		return m, tea.Batch(
			tasks.LoadTasksInterface(m.programContext.ArchonClient, m.programContext.SelectedProjectID),
			realtime.ListenForRealtimeEvents(m.wsClient),
		)

	case realtime.RealtimeProjectUpdateMsg:
		// Project was updated - refresh both tasks and projects
		return m, tea.Batch(
			tasks.LoadTasksInterface(m.programContext.ArchonClient, m.programContext.SelectedProjectID),
			projects.LoadProjectsInterface(m.programContext.ArchonClient),
			realtime.ListenForRealtimeEvents(m.wsClient),
		)
	}
	return m, nil
}
