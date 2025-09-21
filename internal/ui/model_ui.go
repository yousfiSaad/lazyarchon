package ui

import (
	"strings"
	"github.com/yousfisaad/lazyarchon/internal/ui/styling"
)

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
		// Initialize modal only if this is the first time or no features are set
		if m.Modals.featureMode.selectedFeatures == nil || len(m.Modals.featureMode.selectedFeatures) == 0 {
			m.InitFeatureModal()
		}
		// Save backup of current state before opening modal (after potential initialization)
		m.backupFeatureState()

		// Initialize search state
		m.Modals.featureMode.searchMode = false
		m.Modals.featureMode.searchInput = ""
		m.Modals.featureMode.searchQuery = ""
		m.Modals.featureMode.filteredFeatures = nil
		m.Modals.featureMode.matchingIndices = nil
		m.Modals.featureMode.currentMatchIndex = 0

		// Update search matches (will populate with all features since search is empty)
		m.updateFeatureSearchMatches()
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

// Inline Search Management Methods

// ActivateInlineSearch enters inline search mode in the status bar
func (m *Model) ActivateInlineSearch() {
	m.Data.searchMode = true
	m.Data.searchInput = m.Data.searchQuery // Start with current search
}

// CancelInlineSearch exits search mode without applying changes
func (m *Model) CancelInlineSearch() {
	m.Data.searchMode = false
	m.Data.searchInput = ""
}

// CommitInlineSearch applies the current search input and exits search mode
func (m *Model) CommitInlineSearch() {
	m.SetSearchQuery(m.Data.searchInput)
	m.Data.searchMode = false
	m.Data.searchInput = ""
}

// UpdateRealTimeSearch applies search filtering as user types
func (m *Model) UpdateRealTimeSearch() {
	// Temporarily update search query for real-time filtering
	m.SetSearchQuery(m.Data.searchInput)
}

// SetSearchQuery sets the current search query and updates search state
func (m *Model) SetSearchQuery(query string) {
	// Trim whitespace
	query = strings.TrimSpace(query)

	// Remember currently selected task before search changes (same pattern as ClearSearch)
	var selectedTaskID string
	sortedTasks := m.GetSortedTasks()
	if len(sortedTasks) > 0 && m.Navigation.selectedIndex < len(sortedTasks) {
		selectedTaskID = sortedTasks[m.Navigation.selectedIndex].ID
	}

	// Update search state
	m.Data.searchQuery = query
	m.Data.searchActive = (query != "")

	// Add to search history if non-empty and not duplicate
	if query != "" {
		m.addToSearchHistory(query)
	}

	// Update search matches for n/N navigation
	m.updateSearchMatches()

	// Preserve task selection when possible (same pattern as ClearSearch)
	m.findAndSelectTask(selectedTaskID)

	// Update task details viewport to refresh highlighting in content panel
	m.updateTaskDetailsViewport()
}

// ClearSearch clears the current search query
func (m *Model) ClearSearch() {
	// Remember currently selected task
	var selectedTaskID string
	sortedTasks := m.GetSortedTasks()
	if len(sortedTasks) > 0 && m.Navigation.selectedIndex < len(sortedTasks) {
		selectedTaskID = sortedTasks[m.Navigation.selectedIndex].ID
	}

	m.Data.searchQuery = ""
	m.Data.searchActive = false

	// Update search matches (will clear them since search is now inactive)
	m.updateSearchMatches()

	// Find the same task in the (now unfiltered) list and select it
	m.findAndSelectTask(selectedTaskID)
}

// addToSearchHistory adds a query to search history, avoiding duplicates
func (m *Model) addToSearchHistory(query string) {
	// Remove existing instance if present
	for i, existing := range m.Data.searchHistory {
		if existing == query {
			// Remove existing instance
			m.Data.searchHistory = append(m.Data.searchHistory[:i], m.Data.searchHistory[i+1:]...)
			break
		}
	}

	// Add to front of history
	m.Data.searchHistory = append([]string{query}, m.Data.searchHistory...)

	// Limit history size
	if len(m.Data.searchHistory) > 10 {
		m.Data.searchHistory = m.Data.searchHistory[:10]
	}
}

// Helper Methods

// Status Filter Modal Management Methods

// IsStatusFilterModeActive returns true if the status filter modal is currently open
func (m Model) IsStatusFilterModeActive() bool {
	return m.Modals.statusFilter.active
}

// SetStatusFilterMode toggles the status filter modal state
func (m *Model) SetStatusFilterMode(show bool) {
	m.Modals.statusFilter.active = show
	if show {
		// Initialize with current status filters or default to all enabled
		if m.Data.statusFilters == nil {
			m.Modals.statusFilter.selectedStatuses = map[string]bool{
				"todo": true, "doing": true, "review": true, "done": true,
			}
		} else {
			// Copy current filters
			m.Modals.statusFilter.selectedStatuses = make(map[string]bool)
			for status, enabled := range m.Data.statusFilters {
				m.Modals.statusFilter.selectedStatuses[status] = enabled
			}
		}

		// Backup for cancel functionality
		m.Modals.statusFilter.backupStatuses = make(map[string]bool)
		for status, enabled := range m.Modals.statusFilter.selectedStatuses {
			m.Modals.statusFilter.backupStatuses[status] = enabled
		}

		m.Modals.statusFilter.selectedIndex = 0
	}
}

// ApplyStatusFilters applies the status filter selections
func (m *Model) ApplyStatusFilters() {
	// Check if all statuses are enabled (default state)
	allEnabled := true
	for _, enabled := range m.Modals.statusFilter.selectedStatuses {
		if !enabled {
			allEnabled = false
			break
		}
	}

	if allEnabled {
		// All statuses enabled = no custom filtering
		m.Data.statusFilters = nil
		m.Data.statusFilterActive = false
	} else {
		// Copy to data state
		m.Data.statusFilters = make(map[string]bool)
		for status, enabled := range m.Modals.statusFilter.selectedStatuses {
			m.Data.statusFilters[status] = enabled
		}
		m.Data.statusFilterActive = true
	}

	// Reset task selection since filtering changed
	m.setSelectedTask(0)
}

// HasActiveModal returns true if any modal is currently active
func (m Model) HasActiveModal() bool {
	return m.IsHelpMode() || m.IsStatusChangeMode() || m.Modals.projectMode.active || m.IsConfirmationMode() || m.IsFeatureModeActive() || m.IsTaskEditModeActive() || m.IsStatusFilterModeActive()
}

// Styling Helper Methods

// CreateStyleContext creates a StyleContext for UI components with current model state
func (m Model) CreateStyleContext(isSelected bool) *styling.StyleContext {
	themeAdapter := &styling.ThemeAdapter{
		TodoColor:     CurrentTheme.TodoColor,
		DoingColor:    CurrentTheme.DoingColor,
		ReviewColor:   CurrentTheme.ReviewColor,
		DoneColor:     CurrentTheme.DoneColor,
		HeaderColor:   CurrentTheme.HeaderColor,
		MutedColor:    CurrentTheme.MutedColor,
		AccentColor:   CurrentTheme.AccentColor,
		StatusColor:   CurrentTheme.StatusColor,
		FeatureColors: CurrentTheme.FeatureColors,
		Name:          CurrentTheme.Name,
	}

	return styling.NewStyleContext(themeAdapter, m.config).
		WithSelection(isSelected).
		WithSearch(m.Data.searchQuery, m.Data.searchActive)
}
