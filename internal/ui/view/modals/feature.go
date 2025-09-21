package modals

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// FeatureConfig holds configuration for feature modal
type FeatureConfig struct {
	SearchQuery       string
	SearchMode        bool
	SearchInput       string
	SelectedIndex     int
	SelectedFeatures  map[string]bool
	FilteredFeatures  []string
	AllFeatures       []string
	FeatureColorsEnabled bool
}

// FeatureHelpers interface for feature-related functions
type FeatureHelpers interface {
	GetFeatureTaskCount(feature string) int
	GetFeatureColor(feature string) string
	GetMutedFeatureColor(feature string) string
	HighlightSearchTermsWithColor(text, query, textColor string) string
}

// RenderFeatureModal renders the feature selection modal overlay on top of the base UI
func RenderFeatureModal(config FeatureConfig, factory StyleFactory, helpers FeatureHelpers, headerColor, mutedColor string, windowWidth, windowHeight int) string {
	// Calculate modal dimensions (similar to status change modal but taller)
	modalWidth := min(windowWidth-4, 45)   // Slightly wider for feature names + task counts
	modalHeight := min(windowHeight-4, 15) // Taller to accommodate multiple features

	// Get feature selection content
	featureContent := GetFeatureSelectionContent(config, factory, helpers, headerColor, mutedColor)

	// Create feature modal with border
	featureText := strings.Join(featureContent, "\n")
	featureModal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("51")). // Bright cyan like active panels
		Width(modalWidth).
		Height(modalHeight).
		Padding(1).
		Render(featureText)

	// Center the modal on screen
	centeredModal := lipgloss.Place(
		windowWidth, windowHeight,
		lipgloss.Center, lipgloss.Center,
		featureModal,
	)

	return centeredModal
}

// GetFeatureSelectionContent returns the feature selection modal content
func GetFeatureSelectionContent(config FeatureConfig, factory StyleFactory, helpers FeatureHelpers, headerColor, mutedColor string) []string {
	var content []string

	// Title with search indicator
	title := "Select Features"
	if config.SearchQuery != "" {
		title += fmt.Sprintf(" (search: \"%s\")", config.SearchQuery)
	}
	content = append(content, factory.Header().Render(title))
	content = append(content, "")

	// Handle search input mode
	if config.SearchMode {
		cursor := "_" // Simple cursor indicator
		searchText := fmt.Sprintf("[Search] %s%s", config.SearchInput, cursor)

		// Add match indicator if search has results
		if len(config.FilteredFeatures) > 0 {
			searchText += fmt.Sprintf(" (%d matches)", len(config.FilteredFeatures))
		} else if config.SearchInput != "" {
			searchText += " (no matches)"
		}

		content = append(content, searchText)
		content = append(content, "")
		content = append(content, factory.Italic(mutedColor).Render("Enter: apply search  Esc: cancel  Ctrl+U: clear"))
		return content
	}

	if len(config.AllFeatures) == 0 {
		content = append(content, "No features available")
		content = append(content, "")
		content = append(content, factory.Italic(mutedColor).Render("Enter: close"))
		return content
	}

	if len(config.FilteredFeatures) == 0 && config.SearchQuery != "" {
		content = append(content, "No features match your search")
		content = append(content, "")
		content = append(content, factory.Italic(mutedColor).Render("Ctrl+L: clear search  Esc: cancel"))
		return content
	}

	// Feature list with checkboxes (showing filtered features)
	for i, feature := range config.FilteredFeatures {
		// Determine if feature is enabled
		enabled := true // Default to enabled
		if len(config.SelectedFeatures) > 0 {
			if state, exists := config.SelectedFeatures[feature]; exists {
				enabled = state
			}
		}

		// Checkbox symbol with feature-specific color styling
		var checkbox string
		if config.FeatureColorsEnabled {
			// Use feature-specific colors when enabled
			if enabled {
				// Checked: filled square with feature-specific color
				featureColor := helpers.GetFeatureColor(feature)
				checkbox = factory.Text(featureColor).Render("■")
			} else {
				// Unchecked: empty square with muted feature color
				mutedColor := helpers.GetMutedFeatureColor(feature)
				checkbox = factory.Text(mutedColor).Render("□")
			}
		} else {
			// Use default colors when feature colors are disabled
			if enabled {
				// Checked: filled square with green color
				checkbox = factory.Text("46").Render("■")
			} else {
				// Unchecked: empty square with gray color
				checkbox = factory.Text("244").Render("□")
			}
		}

		// Task count for this feature
		taskCount := helpers.GetFeatureTaskCount(feature)
		taskText := fmt.Sprintf("(%d task", taskCount)
		if taskCount != 1 {
			taskText += "s"
		}
		taskText += ")"

		// Style feature name with subtle feature color tinting
		var styledFeatureName string
		if config.FeatureColorsEnabled {
			// Apply subtle feature color to the name
			featureColor := helpers.GetFeatureColor(feature)
			nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(featureColor))

			// Apply search highlighting if active, then apply color
			if config.SearchQuery != "" {
				// Use feature color for both highlighted and non-highlighted text
				styledFeatureName = helpers.HighlightSearchTermsWithColor(feature, config.SearchQuery, featureColor)
			} else {
				styledFeatureName = nameStyle.Render(feature)
			}
		} else {
			// Use default styling when feature colors are disabled
			if config.SearchQuery != "" {
				// Use theme header color instead of black for better visibility
				styledFeatureName = helpers.HighlightSearchTermsWithColor(feature, config.SearchQuery, headerColor)
			} else {
				styledFeatureName = feature
			}
		}

		// Build the line
		line := fmt.Sprintf("%s %s %s", checkbox, styledFeatureName,
			factory.Text("244").Render(taskText))

		// Highlight selected feature
		if i == config.SelectedIndex {
			line = "► " + line + " ◄" // Selection indicator
			line = factory.Bold(headerColor).Render(line)
		} else {
			line = "  " + line
		}

		content = append(content, line)
	}

	content = append(content, "")
	content = append(content, "──────────────────────────────────────")
	content = append(content, "")

	// Instructions (dynamic based on search state)
	if config.SearchQuery != "" {
		// Show search-specific instructions
		content = append(content, factory.Italic(mutedColor).Render("j/k: navigate  gg/G: top/bottom  J/K: fast scroll"))
		content = append(content, factory.Italic(mutedColor).Render("Space: toggle  a: toggle all/none  /: search"))
		content = append(content, factory.Italic(mutedColor).Render("n/N: next/prev match  Ctrl+L: clear search"))
		content = append(content, factory.Italic(mutedColor).Render("Enter: apply   Esc/q: cancel"))
	} else {
		// Show normal instructions
		content = append(content, factory.Italic(mutedColor).Render("j/k: navigate  gg/G: top/bottom  J/K: fast scroll"))
		content = append(content, factory.Italic(mutedColor).Render("Space: toggle  a: toggle all/none  /: search"))
		content = append(content, factory.Italic(mutedColor).Render("Enter: apply   Esc/q: cancel"))
	}

	return content
}