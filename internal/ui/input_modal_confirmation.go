package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/internal/ui/input"
)

// handleConfirmationModeInput handles keyboard input when confirmation modal is open
func (m Model) handleConfirmationModeInput(key string) (Model, tea.Cmd) {
	if !input.IsConfirmationModalKey(key) {
		return m, nil // Ignore keys not handled by confirmation modal
	}

	action := input.GetConfirmationModalAction(key)
	switch action {
	case "cancel":
		m.SetConfirmationMode(false, "", "", "")
		return m, nil
	case "confirm":
		// User pressed 'y' - always quit
		return m, tea.Quit
	case "select":
		// Enter key - act based on selected option
		if m.Modals.confirmation.selectedOption == 0 {
			// Confirm option selected - quit
			return m, tea.Quit
		} else {
			// Cancel option selected - close modal
			m.SetConfirmationMode(false, "", "", "")
			return m, nil
		}
	case "down":
		// Navigate to cancel option
		m.Modals.confirmation.selectedOption = 1
		return m, nil
	case "up":
		// Navigate to confirm option
		m.Modals.confirmation.selectedOption = 0
		return m, nil
	case "quit":
		return m, tea.Quit
	default:
		return m, nil
	}
}