package ui

// Panel and View Management Methods

// IsLeftPanelActive returns true if the left panel (task list) is currently active
func (m Model) IsLeftPanelActive() bool {
	return m.Window.activeView == LeftPanel
}

// IsRightPanelActive returns true if the right panel (task details) is currently active
func (m Model) IsRightPanelActive() bool {
	return m.Window.activeView == RightPanel
}

// SetActiveView sets the currently active panel
func (m *Model) SetActiveView(view ActiveView) {
	m.Window.activeView = view
}

// GetActiveViewName returns a human-readable name of the currently active view
func (m Model) GetActiveViewName() string {
	switch m.Window.activeView {
	case LeftPanel:
		return "Tasks"
	case RightPanel:
		return "Details"
	default:
		return "Unknown"
	}
}

// Help Modal Management Methods

// IsHelpMode returns true if the help modal is currently open
func (m Model) IsHelpMode() bool {
	return m.Modals.help.active
}

// SetHelpMode toggles the help modal state
func (m *Model) SetHelpMode(show bool) {
	m.Modals.help.active = show
	if show {
		// Calculate modal dimensions for viewport sizing
		modalWidth := Min(m.Window.width-4, 70)   // Maximum 70 chars wide, with margins
		modalHeight := Min(m.Window.height-4, 25) // Maximum 25 lines high, with margins
		contentHeight := modalHeight - 4          // Account for border and padding
		contentWidth := modalWidth - 4            // Account for border and padding

		// Resize viewport to fit modal content area
		m.helpModalViewport.Width = contentWidth
		m.helpModalViewport.Height = contentHeight

		// Update content and reset scroll when opening help
		m.updateHelpModalViewport()
		m.helpModalViewport.GotoTop()
	}
}

// Status Change Modal Management Methods

// IsStatusChangeMode returns true if the status change modal is currently open
func (m Model) IsStatusChangeMode() bool {
	return m.Modals.statusChange.active
}

// SetStatusChangeMode toggles the status change modal state
func (m *Model) SetStatusChangeMode(show bool) {
	m.Modals.statusChange.active = show
	if show {
		// Initialize to current task's status index when opening modal
		if len(m.Data.tasks) > 0 && m.Navigation.selectedIndex < len(m.GetSortedTasks()) {
			currentTask := m.GetSortedTasks()[m.Navigation.selectedIndex]
			m.Modals.statusChange.selectedIndex = m.getStatusIndex(currentTask.Status)
		}
	}
}

// Project Mode Management Methods

// IsProjectMode returns true if the project selection modal is currently open
func (m Model) IsProjectMode() bool {
	return m.Modals.projectMode.active
}

// SetProjectMode toggles the project selection modal state
func (m *Model) SetProjectMode(show bool) {
	m.Modals.projectMode.active = show
	if show {
		m.Modals.projectMode.selectedIndex = 0 // Reset to first project when opening
	}
}

// Status Utility Methods

// getStatusIndex returns the index (0-3) for a given status string
func (m Model) getStatusIndex(status string) int {
	switch status {
	case "todo":
		return 0
	case "doing":
		return 1
	case "review":
		return 2
	case "done":
		return 3
	default:
		return 0 // Default to todo
	}
}

// getStatusFromIndex returns the status string for a given index (0-3)
func (m Model) getStatusFromIndex(index int) string {
	switch index {
	case 0:
		return "todo"
	case 1:
		return "doing"
	case 2:
		return "review"
	case 3:
		return "done"
	default:
		return "todo"
	}
}

// Confirmation Modal Management Methods

// IsConfirmationMode returns true if the confirmation modal is currently open
func (m Model) IsConfirmationMode() bool {
	return m.Modals.confirmation.active
}

// SetConfirmationMode toggles the confirmation modal state
func (m *Model) SetConfirmationMode(show bool, message, confirmText, cancelText string) {
	m.Modals.confirmation.active = show
	if show {
		m.Modals.confirmation.message = message
		m.Modals.confirmation.confirmText = confirmText
		m.Modals.confirmation.cancelText = cancelText
		m.Modals.confirmation.selectedOption = 0 // Default to confirm option
	}
}

// ShowQuitConfirmation shows the quit confirmation modal
func (m *Model) ShowQuitConfirmation() {
	m.SetConfirmationMode(true, "Are you sure you want to quit LazyArchon?", "Yes", "No")
}

// Feature Selection Modal Management Methods

// IsFeatureModeActive returns true if the feature selection modal is currently open
func (m Model) IsFeatureModeActive() bool {
	return m.Modals.featureMode.active
}

// SetFeatureMode toggles the feature selection modal state
func (m *Model) SetFeatureMode(show bool) {
	m.Modals.featureMode.active = show
	if show {
		// Save backup of current state before opening modal
		m.backupFeatureState()
		// Initialize modal when opening
		m.InitFeatureModal()
	}
}

// Task Edit Modal Management Methods

// IsTaskEditModeActive returns true if the task edit modal is currently open
func (m Model) IsTaskEditModeActive() bool {
	return m.Modals.taskEdit.active
}

// SetTaskEditMode toggles the task edit modal state
func (m *Model) SetTaskEditMode(show bool) {
	m.Modals.taskEdit.active = show
	if show {
		// Initialize modal when opening
		m.Modals.taskEdit.selectedIndex = 0
		m.Modals.taskEdit.newFeatureName = ""
		m.Modals.taskEdit.isCreatingNew = false
		m.Modals.taskEdit.currentField = "feature" // Start with feature field

		// If current task has a feature, find its index to pre-select it
		if len(m.Data.tasks) > 0 && m.Navigation.selectedIndex < len(m.GetSortedTasks()) {
			currentTask := m.GetSortedTasks()[m.Navigation.selectedIndex]
			if currentTask.Feature != nil && *currentTask.Feature != "" {
				features := m.GetUniqueFeatures()
				for i, feature := range features {
					if feature == *currentTask.Feature {
						m.Modals.taskEdit.selectedIndex = i
						break
					}
				}
			}
		}
	}
}

// Helper Methods

// HasActiveModal returns true if any modal is currently active
func (m Model) HasActiveModal() bool {
	return m.IsHelpMode() || m.IsStatusChangeMode() || m.Modals.projectMode.active || m.IsConfirmationMode() || m.IsFeatureModeActive() || m.IsTaskEditModeActive()
}
