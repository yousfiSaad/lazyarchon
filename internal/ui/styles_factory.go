package ui

import "github.com/charmbracelet/lipgloss"

// CreatePanelStyleNew creates a styled panel with given dimensions
func CreatePanelStyleNew(width, height int) lipgloss.Style {
	return BasePanelStyle.Copy().
		Width(width - BorderWidth).
		Height(height - BorderWidth).
		Padding(PanelPadding)
}

// CreateTaskItemStyleNew creates a style for task list items based on selection and status color
func CreateTaskItemStyleNew(selected bool, statusColor string) lipgloss.Style {

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(statusColor))
		// Background handled by StyleContext when needed or selection styling

	if selected {
		style = style.
			Foreground(lipgloss.Color("15")). // White text for selected items
			Bold(true)
	}
	return style
}

// CreateProjectItemStyleNew creates a style for project list items based on selection
func CreateProjectItemStyleNew(selected bool, isAllTasks bool) lipgloss.Style {
	var baseStyle lipgloss.Style
	if isAllTasks {
		baseStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ActiveTheme.InfoColor))
			// Background handled by StyleContext when needed
	} else {
		baseStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ActiveTheme.AccentColor))
			// Background handled by StyleContext when needed
	}

	if selected {
		baseStyle = baseStyle.
			Foreground(lipgloss.Color("15")). // High contrast white text
			Bold(true)
	}

	return baseStyle
}

// CreateScrollBarStyleNew creates a style for the scroll bar panel
func CreateScrollBarStyleNew(width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Foreground(lipgloss.Color("241")). // Dim color for scroll bar
		Align(lipgloss.Right).             // Align scroll bar to right
		Padding(0, 0)                      // No padding for clean alignment
}

// CreateActivePanelStyleNew creates a styled panel with active state indication
// Background is handled by individual lines via RenderLine() to avoid conflicts
func CreateActivePanelStyleNew(width, height int, isActive bool) lipgloss.Style {
	var borderColor string
	if isActive {
		borderColor = ActiveTheme.ActiveBorderColor
	} else {
		borderColor = ActiveTheme.InactiveBorderColor
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		// Background handled by StyleContext when needed for each line
		Width(width - BorderWidth).
		Height(height - BorderWidth).
		Padding(PanelPadding)
}

// CreateStatusBarStyleNew creates a status bar with contextual styling
func CreateStatusBarStyleNew(state string) lipgloss.Style {
	baseStyle := StatusBarStyle.Copy()

	switch state {
	case "loading":
		return baseStyle.Foreground(lipgloss.Color(ActiveTheme.WarningColor))
	case "error":
		return baseStyle.Foreground(lipgloss.Color(ActiveTheme.ErrorColor))
	case "connected", "ready":
		return baseStyle.Foreground(lipgloss.Color(ActiveTheme.SuccessColor))
	case "info":
		return baseStyle.Foreground(lipgloss.Color(ActiveTheme.InfoColor))
	default:
		return baseStyle.Foreground(lipgloss.Color(ActiveTheme.MutedColor))
	}
}