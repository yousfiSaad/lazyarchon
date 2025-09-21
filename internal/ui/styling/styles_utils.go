package styling

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Utility functions for LazyArchon UI styling

// RenderLine creates a complete line with proper padding (background handled globally)
// This provides consistent line width without applying backgrounds
func RenderLine(content string, width int) string {
	if content == "" {
		// For empty lines, return spaces (background applied globally)
		return strings.Repeat(" ", width)
	}

	// Calculate the visual width of the content (accounting for ANSI codes)
	visualWidth := lipgloss.Width(content)

	// If content is already wider than available width, return as-is
	if visualWidth >= width {
		return content
	}

	// Pad to full width (background will be applied globally)
	paddingNeeded := width - visualWidth
	return content + strings.Repeat(" ", paddingNeeded)
}

// PadLineToWidth pads content to specified width (legacy function for backward compatibility)
func PadLineToWidth(content string, width int) string {
	if content == "" {
		return strings.Repeat(" ", width)
	}

	visualWidth := lipgloss.Width(content)
	if visualWidth >= width {
		return content
	}

	paddingNeeded := width - visualWidth
	return content + strings.Repeat(" ", paddingNeeded)
}

// RenderFullWidthLine creates a full-width line with background applied (legacy function)
// Use PadLineToWidth for clean single-layer theming
func RenderFullWidthLine(content string, width int) string {
	if content == "" {
		return ApplyThemeBackground(strings.Repeat(" ", width))
	}

	visualWidth := lipgloss.Width(content)
	if visualWidth >= width {
		return content
	}

	paddingNeeded := width - visualWidth
	paddedContent := content + strings.Repeat(" ", paddingNeeded)
	return ApplyThemeBackground(paddedContent)
}

// ApplyThemeBackground applies only theme background to text, preserving existing foreground styling
// This is useful for content that already has styling (markdown, search highlights, etc.)
// NOTE: Background styling now managed by StyleContext system
func ApplyThemeBackground(text string) string {
	// Return text as-is, background handled by global system
	return text
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

// CreatePriorityIndicatorStyle creates a styled priority indicator
// Deprecated: Use StyleContext.Factory().Priority() for new code
func CreatePriorityIndicatorStyle(priority PriorityLevel) lipgloss.Style {
	priorityColor := GetPriorityColor(priority)

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(priorityColor))
		// Background handled by StyleContext when needed
}