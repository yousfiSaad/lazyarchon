package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"time"
)

// handleHelpModeInput handles keyboard input when help modal is open
func (m Model) handleHelpModeInput(key string) (Model, tea.Cmd) {
	switch key {
	case "?", "esc":
		m.SetHelpMode(false)
		return m, nil
	case "j", "down":
		m.helpModalViewport.LineDown(1)
		return m, nil
	case "k", "up":
		m.helpModalViewport.LineUp(1)
		return m, nil
	case "J":
		m.helpModalViewport.LineDown(4)
		return m, nil
	case "K":
		m.helpModalViewport.LineUp(4)
		return m, nil
	case "ctrl+u", "pgup":
		m.helpModalViewport.HalfViewUp()
		return m, nil
	case "ctrl+d", "pgdown":
		m.helpModalViewport.HalfViewDown()
		return m, nil
	case "gg":
		m.helpModalViewport.GotoTop()
		return m, nil
	case "G":
		m.helpModalViewport.GotoBottom()
		return m, nil
	case "home":
		m.helpModalViewport.GotoTop()
		return m, nil
	case "end":
		m.helpModalViewport.GotoBottom()
		return m, nil
	case "q":
		m.SetHelpMode(false)
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	default:
		return m, nil // Ignore other keys when help is open
	}
}

// handleStatusChangeModeInput handles keyboard input when status change modal is open
func (m Model) handleStatusChangeModeInput(key string) (Model, tea.Cmd) {
	switch key {
	case "esc":
		m.SetStatusChangeMode(false)
		return m, nil
	case "j", "down":
		return m.handleStatusChangeNavigation(1), nil
	case "k", "up":
		return m.handleStatusChangeNavigation(-1), nil
	case "enter":
		// Apply status change and close modal
		return m.handleStatusChangeConfirm()
	case "q":
		m.SetStatusChangeMode(false)
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	default:
		return m, nil // Ignore other keys when status change modal is open
	}
}

// handleConfirmationModeInput handles keyboard input when confirmation modal is open
func (m Model) handleConfirmationModeInput(key string) (Model, tea.Cmd) {
	switch key {
	case "esc", "n":
		m.SetConfirmationMode(false, "", "", "")
		return m, nil
	case "y":
		// User pressed 'y' - always quit
		return m, tea.Quit
	case "enter":
		// Enter key - act based on selected option
		if m.Modals.confirmation.selectedOption == 0 {
			// Confirm option selected - quit
			return m, tea.Quit
		} else {
			// Cancel option selected - close modal
			m.SetConfirmationMode(false, "", "", "")
			return m, nil
		}
	case "j", "down":
		// Navigate to cancel option
		m.Modals.confirmation.selectedOption = 1
		return m, nil
	case "k", "up":
		// Navigate to confirm option
		m.Modals.confirmation.selectedOption = 0
		return m, nil
	case "q":
		// Close confirmation modal instead of recursive confirmation
		m.SetConfirmationMode(false, "", "", "")
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	default:
		return m, nil // Ignore other keys when confirmation is open
	}
}

// handleMultiKeySequence handles multi-key sequences like 'gg'
func (m Model) handleMultiKeySequence(key string) (Model, tea.Cmd, bool) {
	if key == "g" {
		// Check if this is the second 'g' in quick succession
		if m.Navigation.keySequence.lastKeyPressed == "g" && time.Since(m.Navigation.keySequence.lastKeyTime) < 500*time.Millisecond {
			// Double 'g' detected - clear sequence and handle gg command
			m.Navigation.keySequence.lastKeyPressed = ""
			if m.IsHelpMode() {
				m.helpModalViewport.GotoTop()
				return m, nil, true
			} else {
				return m.handleJumpToFirst(), nil, true
			}
		} else {
			// First 'g' - start sequence tracking
			m.Navigation.keySequence.lastKeyPressed = "g"
			m.Navigation.keySequence.lastKeyTime = time.Now()
			return m, nil, true
		}
	}

	// Clear key sequence if different key pressed
	m.Navigation.keySequence.lastKeyPressed = ""
	return m, nil, false
}

// Status change modal navigation methods
func (m Model) handleStatusChangeNavigation(direction int) Model {
	newIndex := m.Modals.statusChange.selectedIndex + direction
	if newIndex < 0 {
		newIndex = 0
	} else if newIndex > 3 { // 4 status options (Todo, Doing, Review, Done)
		newIndex = 3
	}
	m.Modals.statusChange.selectedIndex = newIndex
	return m
}

func (m Model) handleStatusChangeConfirm() (Model, tea.Cmd) {
	// Get the selected task
	sortedTasks := m.GetSortedTasks()
	if len(sortedTasks) == 0 || m.Navigation.selectedIndex >= len(sortedTasks) {
		m.SetStatusChangeMode(false)
		return m, nil
	}

	task := sortedTasks[m.Navigation.selectedIndex]
	statusOptions := []string{"todo", "doing", "review", "done"}

	if m.Modals.statusChange.selectedIndex >= 0 && m.Modals.statusChange.selectedIndex < len(statusOptions) {
		newStatus := statusOptions[m.Modals.statusChange.selectedIndex]

		// Close modal
		m.SetStatusChangeMode(false)

		// Update task status via API
		return m, UpdateTaskStatusCmd(m.client, task.ID, newStatus)
	}

	m.SetStatusChangeMode(false)
	return m, nil
}

// handleFeatureModeInput handles keyboard input when feature selection modal is open
func (m Model) handleFeatureModeInput(key string) (Model, tea.Cmd) {
	availableFeatures := m.GetUniqueFeatures()
	if len(availableFeatures) == 0 {
		// No features available, close modal
		m.SetFeatureMode(false)
		return m, nil
	}

	switch key {
	case "esc", "h":
		// Cancel - restore previous state and close modal
		m.restoreFeatureState()
		m.SetFeatureMode(false)
		return m, nil
	case "j", "down":
		return m.handleFeatureModeNavigation(1), nil
	case "k", "up":
		return m.handleFeatureModeNavigation(-1), nil
	case "enter", "l":
		// Apply - keep current selections and close modal
		m.SetFeatureMode(false)
		return m, nil
	case " ": // Space bar toggles current feature
		if m.Modals.featureMode.selectedIndex >= 0 && m.Modals.featureMode.selectedIndex < len(availableFeatures) {
			selectedFeature := availableFeatures[m.Modals.featureMode.selectedIndex]
			m.ToggleFeature(selectedFeature)
		}
		return m, nil
	case "a":
		// Select all features
		m.SelectAllFeatures()
		return m, nil
	case "n":
		// Select no features
		m.SelectNoFeatures()
		return m, nil
	case "q":
		// Cancel - restore previous state and close modal (same as Esc)
		m.restoreFeatureState()
		m.SetFeatureMode(false)
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	default:
		return m, nil // Ignore other keys when feature modal is open
	}
}

// handleFeatureModeNavigation handles navigation within the feature selection modal
func (m Model) handleFeatureModeNavigation(direction int) Model {
	availableFeatures := m.GetUniqueFeatures()
	if len(availableFeatures) == 0 {
		return m
	}

	newIndex := m.Modals.featureMode.selectedIndex + direction
	if newIndex < 0 {
		newIndex = 0
	} else if newIndex >= len(availableFeatures) {
		newIndex = len(availableFeatures) - 1
	}
	m.Modals.featureMode.selectedIndex = newIndex
	return m
}

// handleTaskEditModeInput handles keyboard input when task edit modal is open
func (m Model) handleTaskEditModeInput(key string) (Model, tea.Cmd) {
	// Handle text input mode for creating new feature
	if m.Modals.taskEdit.isCreatingNew {
		switch key {
		case "esc":
			// Cancel new feature creation, go back to selection mode
			m.Modals.taskEdit.isCreatingNew = false
			m.Modals.taskEdit.newFeatureName = ""
			return m, nil
		case "enter":
			// Create new feature with the entered name
			if m.Modals.taskEdit.newFeatureName != "" {
				return m.handleTaskEditConfirm(m.Modals.taskEdit.newFeatureName)
			}
			return m, nil
		case "backspace":
			// Remove last character
			if len(m.Modals.taskEdit.newFeatureName) > 0 {
				m.Modals.taskEdit.newFeatureName = m.Modals.taskEdit.newFeatureName[:len(m.Modals.taskEdit.newFeatureName)-1]
			}
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		default:
			// Add character to feature name (limit length)
			if len(key) == 1 && len(m.Modals.taskEdit.newFeatureName) < 30 {
				// Only allow alphanumeric, dash, underscore
				if (key >= "a" && key <= "z") || (key >= "A" && key <= "Z") ||
					(key >= "0" && key <= "9") || key == "-" || key == "_" {
					m.Modals.taskEdit.newFeatureName += key
				}
			}
			return m, nil
		}
	}

	// Handle feature selection mode
	availableFeatures := m.GetUniqueFeatures()

	switch key {
	case "esc":
		// Cancel and close modal
		m.SetTaskEditMode(false)
		return m, nil
	case "j", "down":
		return m.handleTaskEditNavigation(1), nil
	case "k", "up":
		return m.handleTaskEditNavigation(-1), nil
	case "enter":
		// Select current option
		if m.Modals.taskEdit.selectedIndex < len(availableFeatures) {
			// Selected an existing feature
			selectedFeature := availableFeatures[m.Modals.taskEdit.selectedIndex]
			return m.handleTaskEditConfirm(selectedFeature)
		} else {
			// Selected "Create new feature"
			m.Modals.taskEdit.isCreatingNew = true
			m.Modals.taskEdit.newFeatureName = ""
			return m, nil
		}
	case "q":
		// Cancel and close modal (same as Esc)
		m.SetTaskEditMode(false)
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	default:
		return m, nil // Ignore other keys when task edit modal is open
	}
}

// handleTaskEditNavigation handles navigation within the task edit modal
func (m Model) handleTaskEditNavigation(direction int) Model {
	availableFeatures := m.GetUniqueFeatures()
	maxIndex := len(availableFeatures) + 1 // Include "Create new feature" option

	newIndex := m.Modals.taskEdit.selectedIndex + direction
	if newIndex < 0 {
		newIndex = 0
	} else if newIndex >= maxIndex {
		newIndex = maxIndex - 1
	}
	m.Modals.taskEdit.selectedIndex = newIndex
	return m
}

// handleTaskEditConfirm applies the feature change and closes the modal
func (m Model) handleTaskEditConfirm(feature string) (Model, tea.Cmd) {
	// Get current task
	if len(m.Data.tasks) > 0 && m.Navigation.selectedIndex < len(m.GetSortedTasks()) {
		currentTask := m.GetSortedTasks()[m.Navigation.selectedIndex]

		// Close modal
		m.SetTaskEditMode(false)

		// Update task feature via API
		return m, UpdateTaskFeatureCmd(m.client, currentTask.ID, feature)
	}

	// Close modal if no valid task
	m.SetTaskEditMode(false)
	return m, nil
}
