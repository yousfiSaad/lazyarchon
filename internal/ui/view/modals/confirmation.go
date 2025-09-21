package modals

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ConfirmationConfig holds configuration for confirmation modal
type ConfirmationConfig struct {
	Message        string
	ConfirmText    string
	CancelText     string
	SelectedOption int // 0 for confirm, 1 for cancel
}

// RenderConfirmationModal renders the confirmation modal overlay on top of the base UI
func RenderConfirmationModal(config ConfirmationConfig, factory StyleFactory, mutedColor string, windowWidth, windowHeight int) string {
	// Calculate modal dimensions (smaller than help modal)
	modalWidth := min(windowWidth-4, 50)  // Narrower modal for confirmation
	modalHeight := min(windowHeight-4, 8) // Shorter modal for confirmation

	// Get confirmation content
	confirmationContent := GetConfirmationContent(config, factory, mutedColor)

	// Create confirmation modal with border
	confirmationText := strings.Join(confirmationContent, "\n")
	confirmationModal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")). // Red border for confirmation
		Width(modalWidth).
		Height(modalHeight).
		Padding(1).
		Render(confirmationText)

	// Center the modal on screen
	centeredModal := lipgloss.Place(
		windowWidth, windowHeight,
		lipgloss.Center, lipgloss.Center,
		confirmationModal,
	)

	return centeredModal
}

// GetConfirmationContent returns the confirmation modal content
func GetConfirmationContent(config ConfirmationConfig, factory StyleFactory, mutedColor string) []string {
	var content []string

	// Title with warning style
	content = append(content, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")).Render("Confirmation"))
	content = append(content, "")

	// Message
	content = append(content, config.Message)
	content = append(content, "")

	// Options with selection indicators
	confirmOption := fmt.Sprintf("%s %s", config.ConfirmText, "(y)")
	cancelOption := fmt.Sprintf("%s %s", config.CancelText, "(n)")

	if config.SelectedOption == 0 {
		// Confirm option selected
		confirmOption = "► " + confirmOption + " ◄"
		confirmOption = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")).Render(confirmOption)
		cancelOption = "  " + cancelOption
	} else {
		// Cancel option selected
		confirmOption = "  " + confirmOption
		cancelOption = "► " + cancelOption + " ◄"
		cancelOption = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("34")).Render(cancelOption)
	}

	content = append(content, confirmOption)
	content = append(content, cancelOption)
	content = append(content, "")

	// Instructions
	content = append(content, factory.Italic(mutedColor).Render("Enter/y: confirm  Esc/n: cancel"))

	return content
}