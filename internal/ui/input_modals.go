package ui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"time"

	"github.com/yousfisaad/lazyarchon/internal/ui/commands"
)




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
			} else if m.IsFeatureModeActive() {
				return m.handleFeatureModeJumpToFirst(), nil, true
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

		// Show loading message with context
		m.SetLoadingWithMessage(true, fmt.Sprintf("Updating task status to %s...", newStatus))

		// Update task status via API
		return m, commands.UpdateTaskStatusCmd(m.client, task.ID, newStatus)
	}

	m.SetStatusChangeMode(false)
	return m, nil
}

// handleFeatureModeInput handles keyboard input when feature selection modal is open
func (m Model) handleFeatureModeInput(key string) (Model, tea.Cmd) {
	// Handle search input mode first
	if m.Modals.featureMode.searchMode {
		return m.handleFeatureSearchInput(key)
	}

	filteredFeatures := m.GetFilteredFeatures()
	if len(filteredFeatures) == 0 && m.Modals.featureMode.searchQuery == "" {
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
		// Update search matches after filter change
		m.updateSearchMatches()
		return m, nil
	case " ": // Space bar toggles current feature
		if m.Modals.featureMode.selectedIndex >= 0 && m.Modals.featureMode.selectedIndex < len(filteredFeatures) {
			selectedFeature := filteredFeatures[m.Modals.featureMode.selectedIndex]
			m.ToggleFeature(selectedFeature)
		}
		return m, nil
	case "a":
		// Smart toggle: select all if not all selected, otherwise select none
		m.SmartToggleAllFeatures()
		return m, nil
	case "/":
		// Activate search mode
		m.activateFeatureSearch()
		return m, nil
	case "n":
		// Next search match (only when search is active)
		if m.Modals.featureMode.searchQuery != "" {
			m.nextFeatureMatch()
		}
		return m, nil
	case "N":
		// Previous search match (only when search is active)
		if m.Modals.featureMode.searchQuery != "" {
			m.previousFeatureMatch()
		}
		return m, nil
	case "ctrl+l":
		// Clear search
		if m.Modals.featureMode.searchQuery != "" {
			m.clearFeatureSearch()
		}
		return m, nil
	case "q":
		// Cancel - restore previous state and close modal (same as Esc)
		m.restoreFeatureState()
		m.SetFeatureMode(false)
		return m, nil
	case "G":
		// Jump to last feature
		return m.handleFeatureModeJumpToLast(), nil
	case "J":
		// Fast scroll down (5 features)
		return m.handleFeatureModeFastScroll(1), nil
	case "K":
		// Fast scroll up (5 features)
		return m.handleFeatureModeFastScroll(-1), nil
	case "ctrl+d", "pgdown":
		// Half-page scroll down
		return m.handleFeatureModeHalfPage(1), nil
	case "ctrl+u", "pgup":
		// Half-page scroll up
		return m.handleFeatureModeHalfPage(-1), nil
	case "home":
		// Jump to first feature
		return m.handleFeatureModeJumpToFirst(), nil
	case "end":
		// Jump to last feature
		return m.handleFeatureModeJumpToLast(), nil
	case "ctrl+c":
		return m, tea.Quit
	default:
		return m, nil // Ignore other keys when feature modal is open
	}
}

// handleFeatureModeNavigation handles navigation within the feature selection modal
func (m Model) handleFeatureModeNavigation(direction int) Model {
	filteredFeatures := m.GetFilteredFeatures()
	if len(filteredFeatures) == 0 {
		return m
	}

	newIndex := m.Modals.featureMode.selectedIndex + direction
	if newIndex < 0 {
		newIndex = 0
	} else if newIndex >= len(filteredFeatures) {
		newIndex = len(filteredFeatures) - 1
	}
	m.Modals.featureMode.selectedIndex = newIndex
	return m
}

// handleFeatureModeJumpToFirst handles 'gg' key - jump to first feature in modal
func (m Model) handleFeatureModeJumpToFirst() Model {
	filteredFeatures := m.GetFilteredFeatures()
	if len(filteredFeatures) > 0 {
		m.Modals.featureMode.selectedIndex = 0
	}
	return m
}

// handleFeatureModeJumpToLast handles 'G' key - jump to last feature in modal
func (m Model) handleFeatureModeJumpToLast() Model {
	filteredFeatures := m.GetFilteredFeatures()
	if len(filteredFeatures) > 0 {
		m.Modals.featureMode.selectedIndex = len(filteredFeatures) - 1
	}
	return m
}

// handleFeatureModeFastScroll handles 'J'/'K' keys - fast scroll in feature modal
func (m Model) handleFeatureModeFastScroll(direction int) Model {
	filteredFeatures := m.GetFilteredFeatures()
	if len(filteredFeatures) == 0 {
		return m
	}

	// Fast scroll by 5 items (similar to main interface fast scroll)
	step := 5 * direction
	newIndex := m.Modals.featureMode.selectedIndex + step

	if newIndex < 0 {
		newIndex = 0
	} else if newIndex >= len(filteredFeatures) {
		newIndex = len(filteredFeatures) - 1
	}

	m.Modals.featureMode.selectedIndex = newIndex
	return m
}

// handleFeatureModeHalfPage handles 'ctrl+u'/'ctrl+d' keys - half-page scroll in feature modal
func (m Model) handleFeatureModeHalfPage(direction int) Model {
	filteredFeatures := m.GetFilteredFeatures()
	if len(filteredFeatures) == 0 {
		return m
	}

	// Calculate half-page size based on modal height (approximately 7-8 visible features)
	halfPage := 4 // Conservative estimate for half-page in feature modal
	step := halfPage * direction
	newIndex := m.Modals.featureMode.selectedIndex + step

	if newIndex < 0 {
		newIndex = 0
	} else if newIndex >= len(filteredFeatures) {
		newIndex = len(filteredFeatures) - 1
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
		return m, commands.UpdateTaskFeatureCmd(m.client, currentTask.ID, feature)
	}

	// Close modal if no valid task
	m.SetTaskEditMode(false)
	return m, nil
}

// handleFeatureSearchInput handles keyboard input when feature search mode is active
func (m Model) handleFeatureSearchInput(key string) (Model, tea.Cmd) {
	switch key {
	case "esc":
		// Cancel search and revert to previous state
		m.cancelFeatureSearch()
		return m, nil

	case "enter":
		// Commit current search input
		m.commitFeatureSearch()
		return m, nil

	case "backspace":
		// Remove last character
		if len(m.Modals.featureMode.searchInput) > 0 {
			m.Modals.featureMode.searchInput = m.Modals.featureMode.searchInput[:len(m.Modals.featureMode.searchInput)-1]
			// Update search in real-time
			m.updateFeatureSearchMatches()
		}
		return m, nil

	case "ctrl+u":
		// Clear entire input
		m.Modals.featureMode.searchInput = ""
		m.updateFeatureSearchMatches()
		return m, nil

	case "ctrl+c":
		return m, tea.Quit

	default:
		// Handle printable characters
		if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
			m.Modals.featureMode.searchInput += key
			// Update search in real-time
			m.updateFeatureSearchMatches()
		}
		return m, nil
	}
}

