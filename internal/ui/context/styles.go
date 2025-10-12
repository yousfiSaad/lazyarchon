package context

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/yousfisaad/lazyarchon/internal/shared/config"
)

// Styles holds centralized styling configuration for components
// This provides consistent theming across all UI components
type Styles struct {
	// Colors for different task states and UI elements
	Colors struct {
		TodoTask   lipgloss.AdaptiveColor
		DoingTask  lipgloss.AdaptiveColor
		ReviewTask lipgloss.AdaptiveColor
		DoneTask   lipgloss.AdaptiveColor
		Success    lipgloss.AdaptiveColor
		Error      lipgloss.AdaptiveColor
		Warning    lipgloss.AdaptiveColor
		Info       lipgloss.AdaptiveColor
	}

	// Common styles used across components
	Common struct {
		HeaderStyle    lipgloss.Style
		StatusBarStyle lipgloss.Style
		PanelStyle     lipgloss.Style
		BorderStyle    lipgloss.Style
		ErrorStyle     lipgloss.Style
		SuccessStyle   lipgloss.Style
	}

	// Section-specific styles
	Section struct {
		ContainerPadding int
		ContainerStyle   lipgloss.Style
		SpinnerStyle     lipgloss.Style
		EmptyStateStyle  lipgloss.Style
		TitleStyle       lipgloss.Style
		KeyStyle         lipgloss.Style
	}

	// Table styles for task lists
	Table struct {
		CellStyle         lipgloss.Style
		SelectedCellStyle lipgloss.Style
		HeaderStyle       lipgloss.Style
		RowStyle          lipgloss.Style
	}

	// Task detail panel styles
	TaskDetails struct {
		PanelStyle    lipgloss.Style
		HeaderStyle   lipgloss.Style
		ContentStyle  lipgloss.Style
		MetadataStyle lipgloss.Style
		TagStyle      lipgloss.Style
	}

	// Modal styles
	Modal struct {
		OverlayStyle   lipgloss.Style
		ContentStyle   lipgloss.Style
		TitleStyle     lipgloss.Style
		ButtonStyle    lipgloss.Style
		SelectedButton lipgloss.Style
	}

	// Search styles
	Search struct {
		InputStyle       lipgloss.Style
		ActiveStyle      lipgloss.Style
		PlaceholderStyle lipgloss.Style
	}
}

// InitStyles initializes the styles based on the current theme
func InitStyles(cfg *config.Config) Styles {
	var s Styles

	// Initialize color scheme
	s.initColors(cfg)

	// Initialize common styles
	s.initCommonStyles()

	// Initialize component-specific styles
	s.initSectionStyles()
	s.initTableStyles()
	s.initTaskDetailsStyles()
	s.initModalStyles()
	s.initSearchStyles()

	return s
}

// initColors sets up the color scheme based on configuration
func (s *Styles) initColors(cfg *config.Config) {
	// Get configured color scheme and apply it
	scheme := cfg.GetStatusColorScheme()

	switch scheme {
	case "gray":
		s.setGrayScheme()
	case "warm_gray":
		s.setWarmGrayScheme()
	case "cool_gray":
		s.setCoolGrayScheme()
	default: // "blue" or any invalid value
		s.setBlueScheme()
	}

	// Set other UI colors based on task status colors
	s.Colors.Success = s.Colors.DoneTask
	s.Colors.Error = lipgloss.AdaptiveColor{
		Light: "#EF4444",
		Dark:  "#F87171",
	}
	s.Colors.Warning = s.Colors.DoingTask
	s.Colors.Info = lipgloss.AdaptiveColor{
		Light: "#3B82F6",
		Dark:  "#60A5FA",
	}
}

// setBlueScheme applies the vibrant blue color scheme (default)
// This maintains the original color scheme for backward compatibility
func (s *Styles) setBlueScheme() {
	s.Colors.TodoTask = lipgloss.AdaptiveColor{
		Light: "#6B7280", // Gray for todo items
		Dark:  "#9CA3AF",
	}

	s.Colors.DoingTask = lipgloss.AdaptiveColor{
		Light: "#F59E0B", // Amber for active work
		Dark:  "#FCD34D",
	}

	s.Colors.ReviewTask = lipgloss.AdaptiveColor{
		Light: "#8B5CF6", // Purple for items under review
		Dark:  "#A78BFA",
	}

	s.Colors.DoneTask = lipgloss.AdaptiveColor{
		Light: "#10B981", // Green for completed items
		Dark:  "#34D399",
	}
}

// setGrayScheme applies a neutral gray color scheme
// Hierarchy: Review (brightest) > Doing > Todo > Done (dimmest)
// Best for productivity-focused work and reduced eye strain
func (s *Styles) setGrayScheme() {
	s.Colors.ReviewTask = lipgloss.AdaptiveColor{
		Light: "#374151", // Very dark gray (highest attention)
		Dark:  "#E5E7EB", // Very light gray
	}

	s.Colors.DoingTask = lipgloss.AdaptiveColor{
		Light: "#4B5563", // Dark gray
		Dark:  "#D1D5DB", // Light gray
	}

	s.Colors.TodoTask = lipgloss.AdaptiveColor{
		Light: "#6B7280", // Medium gray
		Dark:  "#9CA3AF", // Medium-light gray
	}

	s.Colors.DoneTask = lipgloss.AdaptiveColor{
		Light: "#9CA3AF", // Light gray (lowest attention)
		Dark:  "#6B7280", // Medium-dark gray
	}
}

// setWarmGrayScheme applies a warm gray color scheme
// Hierarchy: Review (brightest) > Doing > Todo > Done (dimmest)
// Gentle warmth reduces eye fatigue during long coding sessions
func (s *Styles) setWarmGrayScheme() {
	s.Colors.ReviewTask = lipgloss.AdaptiveColor{
		Light: "#44403C", // Very dark warm gray (highest attention)
		Dark:  "#E7E5E4", // Very light warm gray
	}

	s.Colors.DoingTask = lipgloss.AdaptiveColor{
		Light: "#57534E", // Dark warm gray
		Dark:  "#D6D3D1", // Light warm gray
	}

	s.Colors.TodoTask = lipgloss.AdaptiveColor{
		Light: "#78716C", // Medium warm gray
		Dark:  "#A8A29E", // Medium-light warm gray
	}

	s.Colors.DoneTask = lipgloss.AdaptiveColor{
		Light: "#A8A29E", // Light warm gray (lowest attention)
		Dark:  "#78716C", // Medium-dark warm gray
	}
}

// setCoolGrayScheme applies a cool gray color scheme
// Hierarchy: Review (brightest) > Doing > Todo > Done (dimmest)
// Modern, professional appearance with subtle blue undertones
func (s *Styles) setCoolGrayScheme() {
	s.Colors.ReviewTask = lipgloss.AdaptiveColor{
		Light: "#1E293B", // Very dark cool gray (highest attention)
		Dark:  "#F1F5F9", // Very light cool gray
	}

	s.Colors.DoingTask = lipgloss.AdaptiveColor{
		Light: "#334155", // Dark cool gray
		Dark:  "#E2E8F0", // Light cool gray
	}

	s.Colors.TodoTask = lipgloss.AdaptiveColor{
		Light: "#64748B", // Medium cool gray
		Dark:  "#94A3B8", // Medium-light cool gray
	}

	s.Colors.DoneTask = lipgloss.AdaptiveColor{
		Light: "#94A3B8", // Light cool gray (lowest attention)
		Dark:  "#64748B", // Medium-dark cool gray
	}
}

// initCommonStyles sets up common styles used across components
func (s *Styles) initCommonStyles() {
	s.Common.HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#007ACC")).
		Padding(0, 1).
		MarginBottom(1)

	s.Common.StatusBarStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#404040")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1)

	s.Common.PanelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#404040")).
		Padding(1)

	s.Common.BorderStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#404040"))

	s.Common.ErrorStyle = lipgloss.NewStyle().
		Foreground(s.Colors.Error)

	s.Common.SuccessStyle = lipgloss.NewStyle().
		Foreground(s.Colors.Success)
}

// initSectionStyles sets up section-specific styles
func (s *Styles) initSectionStyles() {
	s.Section.ContainerPadding = 1
	s.Section.ContainerStyle = lipgloss.NewStyle().
		Padding(0, s.Section.ContainerPadding)

	s.Section.SpinnerStyle = lipgloss.NewStyle().
		Padding(0, 1)

	s.Section.EmptyStateStyle = lipgloss.NewStyle().
		Faint(true).
		PaddingLeft(1).
		MarginBottom(1)

	s.Section.TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF"))

	s.Section.KeyStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#007ACC")).
		Padding(0, 1)
}

// initTableStyles sets up table-specific styles
func (s *Styles) initTableStyles() {
	s.Table.CellStyle = lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(1).
		MaxHeight(1)

	s.Table.SelectedCellStyle = s.Table.CellStyle.
		Background(lipgloss.Color("#404040")).
		Foreground(lipgloss.Color("#FFFFFF"))

	s.Table.HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(lipgloss.Color("#404040"))

	s.Table.RowStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#2A2A2A"))
}

// initTaskDetailsStyles sets up task detail panel styles
func (s *Styles) initTaskDetailsStyles() {
	s.TaskDetails.PanelStyle = s.Common.PanelStyle

	s.TaskDetails.HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		BorderBottom(true).
		BorderForeground(lipgloss.Color("#404040")).
		MarginBottom(1)

	s.TaskDetails.ContentStyle = lipgloss.NewStyle().
		Padding(0, 1)

	s.TaskDetails.MetadataStyle = lipgloss.NewStyle().
		Faint(true).
		Foreground(lipgloss.Color("#888888"))

	s.TaskDetails.TagStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#007ACC")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1).
		Margin(0, 1, 0, 0)
}

// initModalStyles sets up modal-specific styles
func (s *Styles) initModalStyles() {
	s.Modal.OverlayStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("0")). // Semi-transparent overlay
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	s.Modal.ContentStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#2A2A2A")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#404040")).
		Padding(1, 2)

	s.Modal.TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		MarginBottom(1)

	s.Modal.ButtonStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#404040")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 2).
		Margin(0, 1)

	s.Modal.SelectedButton = s.Modal.ButtonStyle.
		Background(lipgloss.Color("#007ACC")).
		Bold(true)
}

// initSearchStyles sets up search-specific styles
func (s *Styles) initSearchStyles() {
	s.Search.InputStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#2A2A2A")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#404040")).
		Padding(0, 1)

	s.Search.ActiveStyle = s.Search.InputStyle.
		BorderForeground(lipgloss.Color("#007ACC"))

	s.Search.PlaceholderStyle = lipgloss.NewStyle().
		Faint(true).
		Foreground(lipgloss.Color("#888888"))
}

// GetStatusColor returns the appropriate color for a task status
func (s *Styles) GetStatusColor(status string) lipgloss.AdaptiveColor {
	switch status {
	case "todo":
		return s.Colors.TodoTask
	case "doing":
		return s.Colors.DoingTask
	case "review":
		return s.Colors.ReviewTask
	case "done":
		return s.Colors.DoneTask
	default:
		return s.Colors.TodoTask
	}
}

// GetStatusStyle returns a complete style for a task status
func (s *Styles) GetStatusStyle(status string) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(s.GetStatusColor(status))
}
