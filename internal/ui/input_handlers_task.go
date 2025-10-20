package ui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/utils/keys"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/modals/confirmation"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/modals/feature"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/modals/taskedit"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/messages"
)

// =============================================================================
// TASK OPERATION KEY HANDLERS
// =============================================================================
// This file contains all task operation keyboard handlers

// HandleTaskStatusChangeKey handles 't' key - open task properties modal (focused on status)
func (m *MainModel) handleTaskStatusChangeKey(key string) (tea.Cmd, bool) {
	if key == keys.KeyT && !m.uiState.IsProjectView() && len(m.programContext.Tasks) > 0 {
		// CRITICAL: Use GetSelectedTask() to get the actual displayed task
		// Don't use m.GetSortedTasks()[m.selectedIndex] because Model.GetSortedTasks()
		// might not match what TaskList is currently displaying
		selectedTask := m.GetSelectedTask()
		if selectedTask == nil {
			return nil, false
		}

		// DEBUG: Log which task is being edited
		fmt.Fprintf(os.Stderr, "[DEBUG] Opening status change modal: taskID=%s, title=%s, selectedIndex=%d\n",
			selectedTask.ID, selectedTask.Title, m.uiState.SelectedTaskIndex)

		// Get current feature value (handle nil pointer)
		currentFeature := ""
		if selectedTask.Feature != nil {
			currentFeature = *selectedTask.Feature
		}

		// Show unified task properties modal, focused on status field for quick editing
		return func() tea.Msg {
			return taskedit.ShowTaskEditModalMsg{
				TaskID:            selectedTask.ID,
				CurrentStatus:     selectedTask.Status,
				CurrentPriority:   selectedTask.TaskOrder,
				CurrentFeature:    currentFeature,
				FocusField:        taskedit.FieldStatus, // Start on status for quick status changes
				AvailableFeatures: m.GetUniqueFeatures(),
			}
		}, true
	}
	return nil, false
}

// HandleTaskEditKey handles 'e' key - open task properties modal (all fields)
func (m *MainModel) handleTaskEditKey(key string) (tea.Cmd, bool) {
	if key == keys.KeyE && !m.uiState.IsProjectView() && len(m.programContext.Tasks) > 0 {
		// CRITICAL: Use GetSelectedTask() to get the actual displayed task
		// Don't use m.GetSortedTasks()[m.selectedIndex] because Model.GetSortedTasks()
		// might not match what TaskList is currently displaying
		selectedTask := m.GetSelectedTask()
		if selectedTask == nil {
			return nil, false
		}

		// Get current feature value (handle nil pointer)
		currentFeature := ""
		if selectedTask.Feature != nil {
			currentFeature = *selectedTask.Feature
		}

		// DEBUG: Log which task is being edited with all details
		fmt.Fprintf(os.Stderr, "[DEBUG] Opening edit modal:\n")
		fmt.Fprintf(os.Stderr, "  - TaskID: %s\n", selectedTask.ID)
		fmt.Fprintf(os.Stderr, "  - Title: %s\n", selectedTask.Title)
		fmt.Fprintf(os.Stderr, "  - Status: %s\n", selectedTask.Status)
		fmt.Fprintf(os.Stderr, "  - Priority: %d\n", selectedTask.TaskOrder)
		fmt.Fprintf(os.Stderr, "  - Feature: %s\n", currentFeature)
		fmt.Fprintf(os.Stderr, "  - SelectedIndex: %d\n", m.uiState.SelectedTaskIndex)

		// Get available features for the modal
		availableFeatures := m.GetUniqueFeatures()

		// Show unified task properties modal, starting on first field
		showMsg := func() tea.Msg {
			return taskedit.ShowTaskEditModalMsg{
				TaskID:            selectedTask.ID,
				CurrentStatus:     selectedTask.Status,
				CurrentPriority:   selectedTask.TaskOrder,
				CurrentFeature:    currentFeature,
				FocusField:        taskedit.FieldStatus, // Start on first field
				AvailableFeatures: availableFeatures,
			}
		}
		return showMsg, true
	}
	return nil, false
}

// HandleTaskIDCopyKey handles 'y' key - send yank ID message to active component
func (m *MainModel) handleTaskIDCopyKey(key string) (tea.Cmd, bool) {
	if key == keys.KeyY {
		return func() tea.Msg { return messages.YankIDMsg{} }, true
	}
	return nil, false
}

// HandleTaskTitleCopyKey handles 'Y' key - send yank title message to active component
func (m *MainModel) handleTaskTitleCopyKey(key string) (tea.Cmd, bool) {
	if key == keys.KeyYCap {
		return func() tea.Msg { return messages.YankTitleMsg{} }, true
	}
	return nil, false
}

// HandleFeatureSelectionKey handles 'f' key - open feature selection modal
func (m *MainModel) handleFeatureSelectionKey(key string) (tea.Cmd, bool) {
	if key == keys.KeyF && !m.uiState.IsProjectView() {
		// Use the new component-based approach
		// Note: Modal can display "No features available" if GetVisibleFeatures() returns empty

		// Transform featureFilters for modal display:
		// - empty map: No filter active (show all) → display as all features selected
		// - {}: Filter active, nothing selected (show none) → display as nothing selected
		// - populated: Show selected features → display as-is
		selectedFeatures := m.programContext.FeatureFilters
		if len(selectedFeatures) == 0 {
			// Empty map means "no filter, show all" - represent in UI as all features selected
			selectedFeatures = make(map[string]bool)
			for _, feature := range m.GetVisibleFeatures() {
				selectedFeatures[feature] = true
			}
		}

		showMsg := feature.ShowFeatureModalMsg{
			AllFeatures:          m.GetVisibleFeatures(),
			SelectedFeatures:     selectedFeatures, // Never nil - always explicit selection state
			FeatureColorsEnabled: true,             // Enable feature colors
		}
		return func() tea.Msg { return showMsg }, true
	}
	return nil, false
}

// HandleSortModeKey handles 's' key - cycle sort mode forward
//
//nolint:unparam // key parameter intentionally unused - handler is dispatched by routing layer
func (m *MainModel) handleSortModeKey(key string) (tea.Cmd, bool) {
	if !m.uiState.IsProjectView() {
		cmd := m.cycleSortMode()
		return cmd, true
	}
	return nil, false
}

// HandleSortModePreviousKey handles 'S' key - cycle sort mode backward
//
//nolint:unparam // key parameter intentionally unused - handler is dispatched by routing layer
func (m *MainModel) handleSortModePreviousKey(key string) (tea.Cmd, bool) {
	if !m.uiState.IsProjectView() {
		cmd := m.cycleSortModePrevious()
		return cmd, true
	}
	return nil, false
}

// HandleTaskDeleteKey handles 'd' key - delete/archive task with confirmation
func (m *MainModel) handleTaskDeleteKey(key string) (tea.Cmd, bool) {
	if key == keys.KeyD && !m.uiState.IsProjectView() && len(m.programContext.Tasks) > 0 {
		// Get the selected task
		selectedTask := m.GetSelectedTask()
		if selectedTask == nil {
			return nil, false
		}

		// Store the task ID for the confirmation handler
		m.pendingDeleteTaskID = selectedTask.ID

		// Show confirmation modal
		return func() tea.Msg {
			return confirmation.ShowConfirmationModalMsg{
				Message:     "Delete task '" + selectedTask.Title + "'? This cannot be undone.",
				ConfirmText: "Delete",
				CancelText:  "Cancel",
			}
		}, true
	}
	return nil, false
}
