package modals

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/yousfisaad/lazyarchon/internal/ui/view"
)

// Base modal utilities for common modal functionality
// This module provides shared utilities for modal sizing, centering, and styling

// ModalConfig holds configuration for modal appearance
type ModalConfig struct {
	Width  int
	Height int
	BorderColor string
	Padding int
}

// DefaultModalConfig returns default configuration for modals
func DefaultModalConfig() ModalConfig {
	return ModalConfig{
		Width:       70,
		Height:      25,
		BorderColor: "51", // Bright cyan like active panels
		Padding:     1,
	}
}

// CreateModalStyle creates a lipgloss style for modal rendering
func CreateModalStyle(config ModalConfig, windowWidth, windowHeight int) lipgloss.Style {
	// Calculate actual modal dimensions with bounds checking
	modalWidth := view.Min(windowWidth-4, config.Width)   // Margins
	modalHeight := view.Min(windowHeight-4, config.Height) // Margins

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(config.BorderColor)).
		Width(modalWidth).
		Height(modalHeight).
		Padding(config.Padding)
}

// CenterModal centers a modal on screen
func CenterModal(content string, windowWidth, windowHeight int) string {
	return lipgloss.Place(
		windowWidth, windowHeight,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

// RenderModal combines modal styling and centering
func RenderModal(content string, config ModalConfig, windowWidth, windowHeight int) string {
	modalStyle := CreateModalStyle(config, windowWidth, windowHeight)
	styledModal := modalStyle.Render(content)
	return CenterModal(styledModal, windowWidth, windowHeight)
}