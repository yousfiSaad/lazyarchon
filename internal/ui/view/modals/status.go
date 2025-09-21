package modals

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// StatusOption represents a status option in the modal
type StatusOption struct {
	Name   string
	Symbol string
	Color  string
}

// StatusChangeConfig holds configuration for status change modal
type StatusChangeConfig struct {
	SelectedIndex int
	StatusOptions []StatusOption
}

// StyleFactory interface for creating styled text
type StyleFactory interface {
	Header() lipgloss.Style
	Text(color string) lipgloss.Style
	Bold(color string) lipgloss.Style
	Italic(color string) lipgloss.Style
}

// RenderStatusChangeModal renders the status change modal overlay on top of the base UI
func RenderStatusChangeModal(config StatusChangeConfig, factory StyleFactory, mutedColor string, windowWidth, windowHeight int) string {
	// Calculate modal dimensions (smaller than help modal)
	modalWidth := min(windowWidth-4, 40)   // Narrower modal for status selection
	modalHeight := min(windowHeight-4, 10) // Shorter modal for 4 status options

	// Get status change content
	statusContent := GetStatusChangeContent(config, factory, mutedColor)

	// Create status modal with border
	statusText := strings.Join(statusContent, "\n")
	statusModal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("51")). // Bright cyan like active panels
		Width(modalWidth).
		Height(modalHeight).
		Padding(1).
		Render(statusText)

	// Center the modal on screen
	centeredModal := lipgloss.Place(
		windowWidth, windowHeight,
		lipgloss.Center, lipgloss.Center,
		statusModal,
	)

	return centeredModal
}

// GetStatusChangeContent returns the status selection content
func GetStatusChangeContent(config StatusChangeConfig, factory StyleFactory, mutedColor string) []string {
	var content []string

	// Title
	content = append(content, factory.Header().Render("Change Status"))
	content = append(content, "")

	// Status options with selection highlighting
	for i, status := range config.StatusOptions {
		// Style for status line
		statusStyle := factory.Text(status.Color)
		line := fmt.Sprintf("%s %s", status.Symbol, status.Name)

		// Highlight selected status
		if i == config.SelectedIndex {
			line = "► " + line + " ◄" // Selection indicator
			line = factory.Bold(status.Color).Render(line)
		} else {
			line = "  " + line
		}

		content = append(content, statusStyle.Render(line))
	}

	content = append(content, "")
	content = append(content, factory.Italic(mutedColor).Render("Enter: select  Esc: cancel"))

	return content
}