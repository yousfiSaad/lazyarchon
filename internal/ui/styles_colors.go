package ui

import (
	"fmt"
	"hash/fnv"
	"strconv"

	"github.com/charmbracelet/lipgloss"
)

// Color management functions for LazyArchon UI styling

// GetFeatureColor assigns consistent colors to features using a hash-based approach
func GetFeatureColor(featureName string) string {
	if featureName == "" {
		return CurrentTheme.MutedColor
	}

	// Use FNV-1a hash for consistent color assignment
	h := fnv.New32a()
	h.Write([]byte(featureName))
	hashValue := h.Sum32()

	// Map to available feature colors
	if len(CurrentTheme.FeatureColors) == 0 {
		return CurrentTheme.AccentColor // Fallback if no feature colors defined
	}

	colorIndex := int(hashValue) % len(CurrentTheme.FeatureColors)
	return CurrentTheme.FeatureColors[colorIndex]
}

// GetDimmedFeatureColor returns a dimmed version of the feature color
func GetDimmedFeatureColor(featureName string) string {
	baseColor := GetFeatureColor(featureName)

	// Convert to integer for dimming calculation
	if colorNum, err := strconv.Atoi(baseColor); err == nil {
		// Reduce intensity by 50% for dimming effect
		dimmedColor := colorNum - 50
		if dimmedColor < 16 { // Ensure we don't go below standard 16 colors
			dimmedColor = 16
		}
		return fmt.Sprintf("%d", dimmedColor)
	}

	// Fallback to muted color if parsing fails
	return CurrentTheme.MutedColor
}

// GetMutedFeatureColor returns a muted version of the feature color
func GetMutedFeatureColor(featureName string) string {
	// For muted effect, just return the theme's muted color
	// This ensures consistency across all muted elements
	return CurrentTheme.MutedColor
}

// GetStatusColorHierarchy returns color arrays for different status color schemes
func GetStatusColorHierarchy(scheme string) [4]string {
	switch scheme {
	case "blue":
		// Original vibrant blue scheme (existing behavior)
		return [4]string{"153", "75", "33", "24"} // Review, Doing, Todo, Done
	case "gray":
		// Neutral gray scheme for productivity focus
		return [4]string{"250", "246", "242", "238"} // Light to dark gray progression
	case "warm_gray":
		// Warm gray scheme for comfortable viewing
		return [4]string{"180", "144", "138", "95"} // Warm gray tones
	case "cool_gray":
		// Cool gray scheme for modern professional look
		return [4]string{"152", "146", "140", "59"} // Cool gray tones
	default:
		// Default to blue scheme
		return [4]string{"153", "75", "33", "24"}
	}
}

// GetThemeStatusColor returns the appropriate status color from the current theme
func GetThemeStatusColor(status string) string {
	switch status {
	case "todo":
		return CurrentTheme.TodoColor
	case "doing":
		return CurrentTheme.DoingColor
	case "review":
		return CurrentTheme.ReviewColor
	case "done":
		return CurrentTheme.DoneColor
	default:
		return CurrentTheme.TodoColor // Default to todo color
	}
}

// GetPriorityColor returns a color for the priority level
func GetPriorityColor(priority PriorityLevel) string {
	switch priority {
	case PriorityHigh:
		return CurrentTheme.ErrorColor // High priority: red/urgent
	case PriorityMedium:
		return CurrentTheme.WarningColor // Medium priority: yellow/attention
	case PriorityLow:
		return CurrentTheme.MutedColor // Low priority: muted/subtle
	default:
		return CurrentTheme.MutedColor
	}
}

// GetStatusSymbol returns a styled status symbol for tasks - Single Source of Truth
func GetStatusSymbol(status string) string {
	switch status {
	case "todo":
		return StatusSymbolTodo
	case "doing":
		return StatusSymbolDoing
	case "review":
		return StatusSymbolReview
	case "done":
		return StatusSymbolDone
	default:
		return StatusSymbolTodo // Default to todo symbol
	}
}

// GetStatusSymbolMap returns a map of all status symbols for building UI elements
func GetStatusSymbolMap() map[string]string {
	return map[string]string{
		"todo":   StatusSymbolTodo,
		"doing":  StatusSymbolDoing,
		"review": StatusSymbolReview,
		"done":   StatusSymbolDone,
	}
}

// CreateStatusSymbolStyle creates a style for status symbols
func CreateStatusSymbolStyle(status string) lipgloss.Style {
	var color string
	switch status {
	case "todo":
		color = CurrentTheme.TodoColor
	case "doing":
		color = CurrentTheme.DoingColor
	case "review":
		color = CurrentTheme.ReviewColor
	case "done":
		color = CurrentTheme.DoneColor
	default:
		color = CurrentTheme.TodoColor
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		// Background handled by StyleContext when needed
		Bold(true)
}

// CreateHighlightStyle creates a highlight style for search matches
func CreateHighlightStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(CurrentTheme.HighlightColor)). // Highlight foreground
		Bold(true).
		Underline(true) // Use underline for highlight indication
}

// CreateErrorStyle creates an error message style
func CreateErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(CurrentTheme.ErrorColor)).
		// Background handled by StyleContext when needed
		Bold(true).
		Padding(0, 1)
}

// CreateSuccessStyle creates a success message style
func CreateSuccessStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(CurrentTheme.SuccessColor)).
		// Background handled by StyleContext when needed
		Bold(true).
		Padding(0, 1)
}

// CreateWarningStyle creates a warning message style
func CreateWarningStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(CurrentTheme.WarningColor)).
		// Background handled by StyleContext when needed
		Bold(true).
		Padding(0, 1)
}

// CreateStatusBarAccentStyle creates accent styling for status bar elements
func CreateStatusBarAccentStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(CurrentTheme.AccentColor)).
		// Background handled by StyleContext when needed
		Bold(false)
}

// CreateStatusBarInfoStyle creates informational styling for counts and stats
func CreateStatusBarInfoStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(CurrentTheme.InfoColor)).
		// Background handled by StyleContext when needed
		Bold(false)
}

// CreateTextStyle creates foreground-only text style (background handled by container)
func CreateTextStyle(foregroundColor string) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(foregroundColor))
}

// CreateBoldTextStyle creates bold foreground-only text style
func CreateBoldTextStyle(foregroundColor string) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(foregroundColor)).
		Bold(true)
}

// CreateItalicTextStyle creates italic foreground-only text style
func CreateItalicTextStyle(foregroundColor string) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(foregroundColor)).
		Italic(true)
}

// CreateThemedSelectionStyle creates selection styling without backgrounds
func CreateThemedSelectionStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")). // White text for selection
		Bold(true)
}

// GetCurrentThemeName returns the name of the current theme for debugging
func GetCurrentThemeName() string {
	// Use the theme name directly instead of trying to detect by PanelBG
	if CurrentTheme.Name != "" {
		return CurrentTheme.Name
	}
	return "unknown"
}

// Legacy functions for backward compatibility - these apply backgrounds
// Use the new functions above for clean single-layer theming

