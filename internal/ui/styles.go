package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color constants
const (
	ColorSelected       = "237" // Background color for selected items
	ColorBorder         = "62"  // Border color for inactive panels
	ColorActiveBorder   = "51"  // Border color for active panel (bright cyan)
	ColorInactiveBorder = "240" // Border color for inactive panel (dim)
	ColorStatus         = "205" // Header text color
	ColorStatusBar      = "241" // Status bar text color
	ColorProject        = "39"  // Blue for projects
	ColorAllTasks       = "208" // Orange for "All Tasks" option

	// Status bar state colors
	ColorStatusConnected = "46"  // Green for connected/ready states
	ColorStatusWarning   = "220" // Yellow for warnings/loading
	ColorStatusError     = "196" // Red for errors/disconnected
	ColorStatusInfo      = "51"  // Cyan for informational elements
	ColorStatusAccent    = "75"  // Light blue for accent text
)

// Layout constants
const (
	HeaderHeight       = 1
	StatusBarHeight    = 1
	PanelPadding       = 1
	BorderWidth        = 2
	SelectionIndicator = "> "
	NoSelection        = "  "
	MaxTasksPerPage    = 100
)

// Styles for different UI components
var (
	// Header style for the top bar
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ColorStatus)).
			Padding(0, 2)

	// Status bar style for the bottom bar
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorStatusBar)).
			Padding(0, 1).
			Width(0) // Let it expand naturally

	// Base panel style (can be customized for specific panels)
	BasePanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorBorder))

	// Style for selected list items
	SelectedItemStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(ColorSelected)).
				Bold(true)

	// Style for project items in project list
	ProjectStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorProject))

	// Style for "All Tasks" option
	AllTasksStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorAllTasks))

	// Style for task detail headers
	DetailHeaderStyle = lipgloss.NewStyle().
				Bold(true)

	// Style for scroll indicators
	ScrollIndicatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorStatusBar))

	// Style for feature/tag display
	TagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")) // Subtle gray - good contrast without being jarring
)

// CreatePanelStyle creates a styled panel with given dimensions
func CreatePanelStyle(width, height int) lipgloss.Style {
	return BasePanelStyle.Copy().
		Width(width - BorderWidth).
		Height(height - BorderWidth).
		Padding(PanelPadding)
}

// CreateTaskItemStyle creates a style for task list items based on selection and status
func CreateTaskItemStyle(selected bool, statusColor string) lipgloss.Style {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor))
	if selected {
		style = style.Background(lipgloss.Color(ColorSelected)).Bold(true)
	}
	return style
}

// CreateProjectItemStyle creates a style for project list items based on selection
func CreateProjectItemStyle(selected bool, isAllTasks bool) lipgloss.Style {
	var baseStyle lipgloss.Style
	if isAllTasks {
		baseStyle = AllTasksStyle.Copy()
	} else {
		baseStyle = ProjectStyle.Copy()
	}

	if selected {
		baseStyle = baseStyle.Background(lipgloss.Color(ColorSelected)).Bold(true)
	}

	return baseStyle
}

// CreateScrollBarStyle creates a style for the scroll bar panel
func CreateScrollBarStyle(width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Foreground(lipgloss.Color("241")). // Dim color for scroll bar
		Align(lipgloss.Right).             // Align scroll bar to right
		Padding(0, 0)                      // No padding for clean alignment
}

// CreateActivePanelStyle creates a styled panel with active state indication
func CreateActivePanelStyle(width, height int, isActive bool) lipgloss.Style {
	var borderColor string
	if isActive {
		borderColor = ColorActiveBorder
	} else {
		borderColor = ColorInactiveBorder
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Width(width - BorderWidth).
		Height(height - BorderWidth).
		Padding(PanelPadding)
}

// Status Bar Contextual Styles

// CreateStatusBarStyle creates a status bar with contextual styling
func CreateStatusBarStyle(state string) lipgloss.Style {
	baseStyle := StatusBarStyle.Copy()

	switch state {
	case "loading":
		return baseStyle.Foreground(lipgloss.Color(ColorStatusWarning))
	case "error":
		return baseStyle.Foreground(lipgloss.Color(ColorStatusError))
	case "connected", "ready":
		return baseStyle.Foreground(lipgloss.Color(ColorStatusBar))
	case "info":
		return baseStyle.Foreground(lipgloss.Color(ColorStatusInfo))
	default:
		return baseStyle
	}
}

// CreateStatusBarAccentStyle creates accent styling for status bar elements
func CreateStatusBarAccentStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorStatusAccent)).
		Bold(false)
}

// CreateStatusBarInfoStyle creates informational styling for counts and stats
func CreateStatusBarInfoStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorStatusInfo)).
		Bold(false)
}
