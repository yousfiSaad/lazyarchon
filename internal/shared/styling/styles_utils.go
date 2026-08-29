package styling

import (
	"github.com/charmbracelet/lipgloss"
)

// Utility functions for LazyArchon UI styling

// RenderLine creates a complete line with proper padding (background handled globally)
// This provides consistent line width without applying backgrounds
func RenderLine(content string, width int) string {
	if content == "" {
		// For empty lines, use lipgloss to create properly sized content
		return lipgloss.NewStyle().Width(width).Render("")
	}

	// Use lipgloss Width method to handle padding automatically
	// This accounts for ANSI codes and ensures consistent width
	return lipgloss.NewStyle().Width(width).Render(content)
}

// CreateModalStyle creates a modal overlay style
func CreateModalStyle(width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color(CurrentTheme.ActiveBorderColor)).
		// Background handled by global background system
		Width(width).
		Height(height).
		Padding(1, 2).
		Margin(1, 2)
}

// CreateKeyBindingStyle creates a style for keyboard shortcuts
func CreateKeyBindingStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(CurrentTheme.AccentColor)).
		Bold(true).
		Underline(true). // Use underline instead of background
		Padding(0, 1).
		Margin(0, 1)
}

// CreateButtonStyle creates a button-like style
func CreateButtonStyle(isSelected bool) lipgloss.Style {
	baseStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 2)

	if isSelected {
		return baseStyle.
			BorderForeground(lipgloss.Color(CurrentTheme.ActiveBorderColor)).
			Foreground(lipgloss.Color("15")).
			Bold(true)
	}

	return baseStyle.
		BorderForeground(lipgloss.Color(CurrentTheme.BorderColor)).
		// Background handled by global background system
		Foreground(lipgloss.Color(CurrentTheme.AccentColor))
}

// CreateProgressBarStyle creates a progress bar style without backgrounds
func CreateProgressBarStyle(filled bool) lipgloss.Style {
	if filled {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(CurrentTheme.SuccessColor)).
			Bold(true)
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(CurrentTheme.MutedColor))
}
