package modals

import (
	"github.com/charmbracelet/lipgloss"
)

// RenderHelpModal renders the help modal overlay on top of the base UI
func RenderHelpModal(viewportContent string, windowWidth, windowHeight int) string {
	// Calculate modal dimensions
	modalWidth := min(windowWidth-4, 70)   // Maximum 70 chars wide, with margins
	modalHeight := min(windowHeight-4, 25) // Maximum 25 lines high, with margins

	// Create help modal with border
	helpModal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("51")). // Bright cyan like active panels
		Width(modalWidth).
		Height(modalHeight).
		Padding(1).
		Render(viewportContent)

	// Center the modal on screen
	centeredModal := lipgloss.Place(
		windowWidth, windowHeight,
		lipgloss.Center, lipgloss.Center,
		helpModal,
	)

	return centeredModal
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}