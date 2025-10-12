package styling

import (
	"fmt"
	"hash/fnv"
	"strconv"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// Color management functions for LazyArchon UI styling

// colorCache provides thread-safe caching for expensive color computations
type colorCache struct {
	mu           sync.RWMutex
	featureColor map[string]string
	dimmedColor  map[string]string
}

var cache = &colorCache{
	featureColor: make(map[string]string),
	dimmedColor:  make(map[string]string),
}

// GetFeatureColor assigns consistent colors to features using a hash-based approach
func GetFeatureColor(featureName string) string {
	if featureName == "" {
		return CurrentTheme.MutedColor
	}

	// Check cache first (read lock)
	cache.mu.RLock()
	if color, exists := cache.featureColor[featureName]; exists {
		cache.mu.RUnlock()
		return color
	}
	cache.mu.RUnlock()

	// Compute color (expensive operation)
	var color string

	// Use FNV-1a hash for consistent color assignment
	h := fnv.New32a()
	h.Write([]byte(featureName))
	hashValue := h.Sum32()

	// Map to available feature colors
	if len(CurrentTheme.FeatureColors) == 0 {
		color = CurrentTheme.AccentColor // Fallback if no feature colors defined
	} else {
		colorIndex := int(hashValue) % len(CurrentTheme.FeatureColors)
		color = CurrentTheme.FeatureColors[colorIndex]
	}

	// Cache the result (write lock)
	cache.mu.Lock()
	cache.featureColor[featureName] = color
	cache.mu.Unlock()

	return color
}

// GetDimmedFeatureColor returns a dimmed version of the feature color
func GetDimmedFeatureColor(featureName string) string {
	if featureName == "" {
		return CurrentTheme.MutedColor
	}

	// Check cache first (read lock)
	cache.mu.RLock()
	if color, exists := cache.dimmedColor[featureName]; exists {
		cache.mu.RUnlock()
		return color
	}
	cache.mu.RUnlock()

	// Compute dimmed color (expensive operation)
	baseColor := GetFeatureColor(featureName)
	var dimmedColor string

	// Convert to integer for dimming calculation
	if colorNum, err := strconv.Atoi(baseColor); err == nil {
		// Reduce intensity by 50% for dimming effect
		dimmed := colorNum - 50
		if dimmed < 16 { // Ensure we don't go below standard 16 colors
			dimmed = 16
		}
		dimmedColor = fmt.Sprintf("%d", dimmed)
	} else {
		// Fallback to muted color if parsing fails
		dimmedColor = CurrentTheme.MutedColor
	}

	// Cache the result (write lock)
	cache.mu.Lock()
	cache.dimmedColor[featureName] = dimmedColor
	cache.mu.Unlock()

	return dimmedColor
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
	case StatusTodo:
		return CurrentTheme.TodoColor
	case StatusDoing:
		return CurrentTheme.DoingColor
	case StatusReview:
		return CurrentTheme.ReviewColor
	case StatusDone:
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
	case StatusTodo:
		return StatusSymbolTodo
	case StatusDoing:
		return StatusSymbolDoing
	case StatusReview:
		return StatusSymbolReview
	case StatusDone:
		return StatusSymbolDone
	default:
		return StatusSymbolTodo // Default to todo symbol
	}
}

// GetStatusSymbolMap returns a map of all status symbols for building UI elements
func GetStatusSymbolMap() map[string]string {
	return map[string]string{
		StatusTodo:   StatusSymbolTodo,
		StatusDoing:  StatusSymbolDoing,
		StatusReview: StatusSymbolReview,
		StatusDone:   StatusSymbolDone,
	}
}

// CreateStatusSymbolStyle creates a style for status symbols
func CreateStatusSymbolStyle(status string) lipgloss.Style {
	var color string
	switch status {
	case StatusTodo:
		color = CurrentTheme.TodoColor
	case StatusDoing:
		color = CurrentTheme.DoingColor
	case StatusReview:
		color = CurrentTheme.ReviewColor
	case StatusDone:
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
