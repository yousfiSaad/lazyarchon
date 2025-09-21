package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/internal/ui/input"
)

// handleStatusChangeModeInput handles keyboard input when status change modal is open
func (m Model) handleStatusChangeModeInput(key string) (Model, tea.Cmd) {
	if !input.IsStatusChangeModalKey(key) {
		return m, nil // Ignore keys not handled by status change modal
	}

	action := input.GetStatusChangeModalAction(key)
	switch action {
	case "close":
		m.SetStatusChangeMode(false)
		return m, nil
	case "down":
		return m.handleStatusChangeNavigation(1), nil
	case "up":
		return m.handleStatusChangeNavigation(-1), nil
	case "confirm":
		// Apply status change and close modal
		return m.handleStatusChangeConfirm()
	case "quit":
		return m, tea.Quit
	default:
		return m, nil
	}
}