package ui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/v2/internal/archon"
	"github.com/yousfisaad/lazyarchon/v2/internal/domain/tasks"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/modals/confirmation"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/modals/feature"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/modals/status"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/modals/taskcreate"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/modals/taskedit"
)

// =============================================================================
// MODAL MESSAGE HANDLERS
// =============================================================================
// This file contains handlers for modal lifecycle and action messages

// handleModalLifecycle processes modal show/hide messages
//
//nolint:ireturn // Required by Bubble Tea framework - must return tea.Model interface
func (m *MainModel) handleModalLifecycle(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Relevant modal will handle its message
	return m, m.components.Update(msg)
}

// handleModalActions processes modal action messages that need parent handling
// High complexity (16) due to routing 6+ modal action message types with validation
//
//nolint:gocyclo,ireturn // Modal requires routing status/priority/feature updates with validation; ireturn required by Bubble Tea
func (m *MainModel) handleModalActions(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case status.StatusSelectedMsg:
		// Legacy status modal handler - kept for backwards compatibility
		// New code should use TaskPropertiesUpdatedMsg from taskedit modal
		return m, tasks.UpdateTaskStatusInterface(m.programContext.ArchonClient, msg.TaskID, msg.Status)

	case taskedit.TaskPropertiesUpdatedMsg:
		// Handle unified task properties update (status, priority, feature)
		updates := archon.UpdateTaskRequest{}
		hasChanges := false

		if msg.Status != nil {
			updates.Status = msg.Status
			hasChanges = true
		}
		if msg.Priority != nil {
			updates.TaskOrder = msg.Priority
			hasChanges = true
		}
		if msg.Feature != nil {
			updates.Feature = msg.Feature
			hasChanges = true
		}

		// DEBUG: Log which task is being updated via API
		if hasChanges {
			fmt.Fprintf(os.Stderr, "[DEBUG] Sending task update to API: taskID=%s, hasStatus=%v, hasPriority=%v, hasFeature=%v\n",
				msg.TaskID, msg.Status != nil, msg.Priority != nil, msg.Feature != nil)
		}

		// Only send update if something changed
		if hasChanges {
			return m, tasks.UpdateTaskWithRequest(
				m.programContext.ArchonClient,
				msg.TaskID,
				updates,
			)
		}
		return m, nil

	case confirmation.ConfirmationSelectedMsg:
		// Handle confirmation selection
		// Check if this is a task deletion confirmation
		if m.pendingDeleteTaskID != "" {
			taskID := m.pendingDeleteTaskID
			m.pendingDeleteTaskID = "" // Clear pending state

			if msg.Confirmed {
				// User confirmed deletion - execute delete command
				return m, tasks.DeleteTaskInterface(m.programContext.ArchonClient, taskID)
			}
			// User canceled - just return
			return m, nil
		}

		// Default confirmation (quit)
		if msg.Confirmed {
			return m, tea.Quit
		}
		return m, nil

	case taskcreate.TaskCreatedMsg:
		// Handle task creation from task create modal
		// Build CreateTaskRequest and send to API
		createRequest := archon.CreateTaskRequest{
			ProjectID:   msg.ProjectID,
			Title:       msg.Title,
			Description: msg.Description,
			Status:      msg.Status,
			TaskOrder:   msg.Priority,
			Feature:     msg.Feature,
		}

		// DEBUG: Log task creation
		fmt.Fprintf(os.Stderr, "[DEBUG] Creating new task: title=%s, projectID=%s, feature=%v, priority=%d\n",
			msg.Title, msg.ProjectID, msg.Feature, msg.Priority)

		// Send create command
		return m, tasks.CreateTaskInterface(m.programContext.ArchonClient, createRequest)

	case taskedit.FeatureSelectedMsg:
		// Legacy feature selection handler - kept for backwards compatibility
		// New code should use TaskPropertiesUpdatedMsg instead
		return m, tasks.UpdateTaskFeatureInterface(m.programContext.ArchonClient, msg.TaskID, msg.Feature)

	case feature.FeatureSelectionAppliedMsg:
		// Handle feature selection application - update task filtering in ProgramContext
		// This is a client-side filter change - no server fetch needed, just refresh UI
		m.programContext.FeatureFilters = msg.SelectedFeatures
		m.programContext.FeatureFilterActive = len(msg.SelectedFeatures) > 0
		m.refreshUIAfterFilterChange() // Refresh UI immediately with current data
		return m, nil
	}
	return m, nil
}
