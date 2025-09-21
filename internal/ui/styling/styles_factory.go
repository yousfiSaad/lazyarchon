package styling

import (
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// StyleFactory generates consistent styles based on the current StyleContext
// It automatically applies selection state, search highlighting, and theme colors
type StyleFactory struct {
	context *StyleContext
}

// NewStyleFactory creates a new StyleFactory with a default context
func NewStyleFactory() *StyleFactory {
	return &StyleFactory{
		context: &StyleContext{}, // Default empty context
	}
}

// Text creates a basic text style with the specified foreground color
// Automatically applies selection background and search highlighting if active
func (f *StyleFactory) Text(color string) lipgloss.Style {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))

	// Apply selection background if selected
	if f.context.selectionState.IsSelected {
		style = style.Background(lipgloss.Color(f.context.selectionState.BackgroundColor))
	}

	return style
}

// Status creates a style for task status indicators (todo, doing, review, done)
// Uses theme-appropriate colors and applies selection state automatically
func (f *StyleFactory) Status(status string) lipgloss.Style {
	var color string
	switch status {
	case "todo":
		color = f.context.theme.TodoColor
	case "doing":
		color = f.context.theme.DoingColor
	case "review":
		color = f.context.theme.ReviewColor
	case "done":
		color = f.context.theme.DoneColor
	default:
		color = f.context.theme.TodoColor // fallback
	}

	return f.Text(color)
}

// Feature creates a style for feature tags using consistent hash-based coloring
// Automatically applies selection state and preserves feature color identity
func (f *StyleFactory) Feature(featureName string) lipgloss.Style {
	if featureName == "" {
		return f.Text(f.context.theme.MutedColor)
	}

	// Use hash-based color selection for consistent feature colors
	featureColor := f.getFeatureColor(featureName)
	return f.Text(featureColor)
}

// Priority creates a style for priority indicators
// Maps priority levels to appropriate colors with selection support
func (f *StyleFactory) Priority(priority string) lipgloss.Style {
	var color string
	switch priority {
	case "high":
		color = "#FF6B6B" // Red for high priority
	case "medium":
		color = "#FFE66D" // Yellow for medium priority
	case "low":
		color = "#4ECDC4" // Blue for low priority
	default:
		color = f.context.theme.MutedColor // Gray for no priority
	}

	return f.Text(color)
}

// Muted creates a style for secondary/muted text
func (f *StyleFactory) Muted() lipgloss.Style {
	return f.Text(f.context.theme.MutedColor)
}

// Header creates a style for header text
func (f *StyleFactory) Header() lipgloss.Style {
	style := f.Text(f.context.theme.HeaderColor)
	return style.Bold(true)
}

// Accent creates a style for accent/highlighted text
func (f *StyleFactory) Accent() lipgloss.Style {
	return f.Text(f.context.theme.AccentColor)
}

// Panel creates a style for panel containers with borders
func (f *StyleFactory) Panel(width, height int, isActive bool) lipgloss.Style {
	borderColor := "#444444" // Default inactive border
	if isActive {
		borderColor = f.context.theme.AccentColor
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Width(width).
		Height(height).
		Padding(0, 1)
}

// ApplySearchHighlighting applies search highlighting to content if search is active
// This should be called after styling but before final rendering
func (f *StyleFactory) ApplySearchHighlighting(content, originalColor string) string {
	if !f.context.searchState.IsActive || f.context.searchState.Query == "" {
		return content
	}

	return f.highlightSearchTerms(content, f.context.searchState.Query, originalColor)
}

// getFeatureColor returns a consistent color for a feature name using hash-based selection
func (f *StyleFactory) getFeatureColor(featureName string) string {
	if len(f.context.theme.FeatureColors) == 0 {
		return f.context.theme.AccentColor
	}

	// Use FNV-1a hash for consistent color selection
	h := fnv.New32a()
	h.Write([]byte(strings.ToLower(featureName)))
	hash := h.Sum32()

	// Select color from palette
	colorIndex := int(hash) % len(f.context.theme.FeatureColors)
	return f.context.theme.FeatureColors[colorIndex]
}

// highlightSearchTerms highlights search query matches in the given text
func (f *StyleFactory) highlightSearchTerms(text, query, textColor string) string {
	if query == "" {
		return text
	}

	// Case-insensitive search
	lowerText := strings.ToLower(text)
	lowerQuery := strings.ToLower(query)

	// Find all matches
	var result strings.Builder
	lastIndex := 0

	for {
		index := strings.Index(lowerText[lastIndex:], lowerQuery)
		if index == -1 {
			break
		}

		// Adjust index to absolute position
		index += lastIndex

		// Add text before match with original color
		if index > lastIndex {
			beforeMatch := text[lastIndex:index]
			beforeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(textColor))
			if f.context.selectionState.IsSelected {
				beforeStyle = beforeStyle.Background(lipgloss.Color(f.context.selectionState.BackgroundColor))
			}
			result.WriteString(beforeStyle.Render(beforeMatch))
		}

		// Add highlighted match
		match := text[index : index+len(query)]
		matchStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).          // Black text
			Background(lipgloss.Color(f.context.searchState.MatchColor)) // Yellow background
		result.WriteString(matchStyle.Render(match))

		// Move to next position
		lastIndex = index + len(query)
	}

	// Add remaining text
	if lastIndex < len(text) {
		remaining := text[lastIndex:]
		remainingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(textColor))
		if f.context.selectionState.IsSelected {
			remainingStyle = remainingStyle.Background(lipgloss.Color(f.context.selectionState.BackgroundColor))
		}
		result.WriteString(remainingStyle.Render(remaining))
	}

	return result.String()
}

// Helper methods for backward compatibility and special cases

// TextWithBackground creates a text style with explicit background (for migration)
func (f *StyleFactory) TextWithBackground(foregroundColor, backgroundColor string) lipgloss.Style {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(foregroundColor))
	if backgroundColor != "" {
		style = style.Background(lipgloss.Color(backgroundColor))
	}
	return style
}

// Bold creates a bold text style with the specified color
func (f *StyleFactory) Bold(color string) lipgloss.Style {
	return f.Text(color).Bold(true)
}

// StatusBar creates a status bar style with contextual state coloring
func (f *StyleFactory) StatusBar(state string) lipgloss.Style {
	baseStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1).
		Height(1)

	switch state {
	case "ready":
		// Use theme-appropriate background colors that work well as backgrounds
		return baseStyle.Background(lipgloss.Color(f.getStatusBarBackgroundColor()))
	case "loading":
		return baseStyle.Background(lipgloss.Color("#1976D2")) // Blue for loading
	case "error":
		return baseStyle.Background(lipgloss.Color("#D32F2F")) // Red for error
	case "info":
		return baseStyle.Background(lipgloss.Color("#F57C00")) // Orange for info
	default:
		return baseStyle.Background(lipgloss.Color("#444444")) // Default gray
	}
}

// getStatusBarBackgroundColor returns a theme-appropriate background color for the status bar
func (f *StyleFactory) getStatusBarBackgroundColor() string {
	// Use theme-appropriate, muted background colors that provide good contrast
	switch f.context.theme.Name {
	case "Monokai":
		return "#2D5016" // Dark green background (muted version of monokai green)
	case "Gruvbox":
		return "#665C54" // Warm gray-brown background (gruvbox style)
	case "Dracula":
		return "#44475A" // Dark purple-gray background (dracula style)
	default:
		// Default theme - use a sophisticated blue-gray that works with the pink accent
		return "#2C3E50" // Dark blue-gray background
	}
}

// Italic creates an italic text style with the specified color
func (f *StyleFactory) Italic(color string) lipgloss.Style {
	return f.Text(color).Italic(true)
}

// ProjectItem creates a style for project list items with selection and type context
func (f *StyleFactory) ProjectItem(isSelected bool, isAllTasks bool) lipgloss.Style {
	var baseStyle lipgloss.Style
	if isAllTasks {
		baseStyle = f.Text(f.context.theme.AccentColor).Bold(true)
	} else {
		baseStyle = f.Text(f.context.theme.HeaderColor)
	}

	// Apply selection highlighting if this item is selected
	if isSelected {
		return baseStyle.Background(lipgloss.Color("#444444"))
	}

	return baseStyle
}

// BrightenColor brightens a color for selection (placeholder - would need proper color math)
func (f *StyleFactory) BrightenColor(color string, boost float32) string {
	if boost == 1.0 {
		return color
	}

	// Simple brightness boost by converting to higher color codes
	// This is a placeholder - real implementation would do proper color math
	if colorCode, err := strconv.Atoi(color); err == nil {
		boosted := int(float32(colorCode) * boost)
		if boosted > 255 {
			boosted = 255
		}
		return fmt.Sprintf("%d", boosted)
	}

	return color // Return original if can't parse
}