package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/v2/internal/domain/tasks"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/interfaces"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/utils/keys"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/modals/confirmation"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/modals/feature"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/modals/taskcreate"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/modals/taskedit"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/modals/worktree"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/messages"
)

// =============================================================================
// TASK OPERATION KEY HANDLERS
// =============================================================================
// This file contains all task operation keyboard handlers

// HandleTaskCreateKey handles 'c' key - open task creation modal
func (m *MainModel) handleTaskCreateKey(key string) (tea.Cmd, bool) {
	if key == keys.KeyC && !m.uiState.IsProjectView() {
		// Default to the selected project; in the "All Tasks" view fall back
		// to the first known project (the local backend seeds an Inbox).
		defaultProjectID := ""
		if m.programContext.SelectedProjectID != nil {
			defaultProjectID = *m.programContext.SelectedProjectID
		} else if len(m.programContext.Projects) > 0 {
			defaultProjectID = m.programContext.Projects[0].ID
		}

		// Get available features
		availableFeatures := m.GetUniqueFeatures()

		// Get default feature from current context if any
		defaultFeature := ""
		if len(m.programContext.FeatureFilters) == 1 {
			// If exactly one feature filter is active, use it as default
			for feature := range m.programContext.FeatureFilters {
				defaultFeature = feature
				break
			}
		}

		// Show task creation modal
		showMsg := func() tea.Msg {
			return taskcreate.ShowTaskCreateModalMsg{
				DefaultProjectID:  defaultProjectID,
				AvailableFeatures: availableFeatures,
				DefaultFeature:    defaultFeature,
			}
		}
		return showMsg, true
	}
	return nil, false
}

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

		// Get current feature/tag value (use first tag if available)
		currentFeature := ""
		if len(selectedTask.Tags) > 0 {
			currentFeature = selectedTask.Tags[0]
		}

		// Show unified task properties modal, focused on status field for quick editing
		return func() tea.Msg {
			return taskedit.ShowTaskEditModalMsg{
				TaskID:            selectedTask.ID,
				CurrentStatus:     selectedTask.Status,
				CurrentPriority:   selectedTask.Priority,
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

		// Get current feature/tag value (use first tag if available)
		currentFeature := ""
		if len(selectedTask.Tags) > 0 {
			currentFeature = selectedTask.Tags[0]
		}

		// Get available features for the modal
		availableFeatures := m.GetUniqueFeatures()

		// Show unified task properties modal, starting on first field
		showMsg := func() tea.Msg {
			return taskedit.ShowTaskEditModalMsg{
				TaskID:            selectedTask.ID,
				CurrentStatus:     selectedTask.Status,
				CurrentPriority:   selectedTask.Priority,
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

		// Get all features from current project (without feature filter applied)
		// so user can see all available options to select/deselect
		allProjectFeatures := m.GetFeaturesForProjectSelection()

		// Transform featureFilters for modal display:
		// - empty map: No filter active (show all) → display as all features selected
		// - {}: Filter active, nothing selected (show none) → display as nothing selected
		// - populated: Show selected features → display as-is
		selectedFeatures := m.programContext.FeatureFilters
		if selectedFeatures == nil {
			// nil means "no filter, show all" - represent in UI as all features selected.
			// An empty non-nil map means "filter active, nothing selected" and must
			// pass through unchanged so the modal reopens with nothing checked.
			selectedFeatures = make(map[string]bool)
			for _, feature := range allProjectFeatures {
				selectedFeatures[feature] = true
			}
		}

		showMsg := feature.ShowFeatureModalMsg{
			AllFeatures:          allProjectFeatures, // All project features (ignore current feature filter)
			SelectedFeatures:     selectedFeatures,   // Never nil - always explicit selection state
			FeatureColorsEnabled: true,               // Enable feature colors
		}
		return func() tea.Msg { return showMsg }, true
	}
	return nil, false
}

// HandleWorktreeSelectionKey handles 'w' key - open worktree selection modal
func (m *MainModel) handleWorktreeSelectionKey(key string) (tea.Cmd, bool) {
	if key == keys.KeyW && !m.uiState.IsProjectView() {
		// Get all worktrees from current project (without worktree filter applied)
		// so user can see all available options to select/deselect
		allProjectWorktrees := m.GetWorktreesForProjectSelection()

		// Transform worktreeFilters for modal display:
		// - nil: No filter active (show all) → display as all worktrees selected
		// - {}: Filter active, nothing selected (show none) → display as nothing selected
		// - populated: Show selected worktrees → display as-is
		selectedWorktrees := m.programContext.WorktreeFilters
		if selectedWorktrees == nil {
			// nil means "no filter, show all" - represent in UI as all worktrees selected.
			// An empty non-nil map means "filter active, nothing selected" and must
			// pass through unchanged so the modal reopens with nothing checked.
			selectedWorktrees = make(map[string]bool)
			for _, worktree := range allProjectWorktrees {
				selectedWorktrees[worktree] = true
			}
		}

		showMsg := worktree.ShowWorktreeModalMsg{
			AllWorktrees:      allProjectWorktrees, // All project worktrees (ignore current worktree filter)
			SelectedWorktrees: selectedWorktrees,   // Never nil - always explicit selection state
		}
		return func() tea.Msg { return showMsg }, true
	}
	return nil, false
}

// HandleReadyFilterKey handles 'b' key - toggle ready-only filtering
func (m *MainModel) handleReadyFilterKey(key string) (tea.Cmd, bool) {
	if key == keys.KeyB && !m.uiState.IsProjectView() {
		m.programContext.ReadyOnly = !m.programContext.ReadyOnly
		m.refreshUIAfterFilterChange()
		return nil, true
	}
	return nil, false
}

// HandleArchivedFilterKey handles 'x' key - toggle archived-task visibility
func (m *MainModel) handleArchivedFilterKey(key string) (tea.Cmd, bool) {
	if key == keys.KeyX && !m.uiState.IsProjectView() {
		m.programContext.ToggleShowArchivedTasks()
		m.refreshUIAfterFilterChange()
		return nil, true
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

// HandleTaskArchiveKey handles 'X' key - archive/unarchive the selected task.
// Reversible, so no confirmation: the resulting TaskUpdateMsg reloads all
// tasks and the row disappears (or gains its archived chip when 'x' is on).
func (m *MainModel) handleTaskArchiveKey(key string) (tea.Cmd, bool) {
	if key == keys.KeyXCap && !m.uiState.IsProjectView() && len(m.programContext.Tasks) > 0 {
		selectedTask := m.GetSelectedTask()
		if selectedTask == nil {
			return nil, false
		}

		newArchived := !selectedTask.Archived
		return tasks.UpdateTaskWithRequest(
			m.programContext.TaskClient,
			selectedTask.ID,
			interfaces.UpdateTaskRequest{Archived: &newArchived},
		), true
	}
	return nil, false
}
