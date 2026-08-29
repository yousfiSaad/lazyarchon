package styling

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/config"
)

// ThemeConfig holds configurable colors and styling options
type ThemeConfig struct {
	// Background and selection colors
	SelectedBG          string
	SecondarySelectedBG string
	// PanelBG removed - using terminal natural background

	// Border and frame colors
	BorderColor         string
	ActiveBorderColor   string
	InactiveBorderColor string

	// Text colors
	HeaderColor  string
	StatusColor  string
	ErrorColor   string
	WarningColor string
	SuccessColor string
	InfoColor    string

	// Task status colors
	TodoColor   string
	DoingColor  string
	ReviewColor string
	DoneColor   string

	// UI element colors
	AccentColor    string
	MutedColor     string
	HighlightColor string

	// Feature color palette (8 distinct colors for feature tags)
	FeatureColors []string

	// Theme metadata
	Name   string
	IsDark bool
}

// ActiveTheme holds the active theme (will be initialized from config)
var ActiveTheme ThemeConfig

// PredefinedThemes similar to lazygit/lazydocker
var PredefinedThemes = map[string]ThemeConfig{
	"default": {
		Name:                "Default",
		IsDark:              true,
		SelectedBG:          "237", // Dark gray
		SecondarySelectedBG: "235", // Lighter dark gray
		// PanelBG removed - using terminal natural background
		BorderColor:         "62",                                                            // Purple-blue
		ActiveBorderColor:   "51",                                                            // Bright cyan
		InactiveBorderColor: "240",                                                           // Dim gray
		HeaderColor:         "39",                                                            // Blue
		StatusColor:         "205",                                                           // Pink
		ErrorColor:          "196",                                                           // Bright red
		WarningColor:        "220",                                                           // Yellow
		SuccessColor:        "46",                                                            // Green
		InfoColor:           "51",                                                            // Cyan
		TodoColor:           "33",                                                            // Bright blue - "Start me now!" actionable
		DoingColor:          "75",                                                            // Medium blue - active, balanced
		ReviewColor:         "153",                                                           // Light blue - "Waiting for others" less urgent
		DoneColor:           "24",                                                            // Dim blue - completed, subtle
		AccentColor:         "75",                                                            // Light blue
		MutedColor:          "244",                                                           // Gray
		HighlightColor:      "226",                                                           // Bright yellow
		FeatureColors:       []string{"117", "213", "83", "212", "147", "204", "228", "183"}, // Cyan, pink, green, magenta, blue, purple, yellow, orange
	},
	"monokai": {
		Name:                "Monokai",
		IsDark:              true,
		SelectedBG:          "236",
		SecondarySelectedBG: "234",
		// PanelBG removed - using terminal natural background
		BorderColor:         "141", // Purple
		ActiveBorderColor:   "198", // Pink
		InactiveBorderColor: "240",
		HeaderColor:         "81",                                                            // Green
		StatusColor:         "81",                                                            // Green
		ErrorColor:          "197",                                                           // Red
		WarningColor:        "214",                                                           // Orange
		SuccessColor:        "118",                                                           // Green
		InfoColor:           "81",                                                            // Green
		TodoColor:           "33",                                                            // Bright blue - actionable items
		DoingColor:          "75",                                                            // Medium blue - active and balanced
		ReviewColor:         "153",                                                           // Light blue - waiting for validation
		DoneColor:           "24",                                                            // Dim blue - completed, understated
		AccentColor:         "198",                                                           // Pink
		MutedColor:          "59",                                                            // Dark gray
		HighlightColor:      "227",                                                           // Yellow
		FeatureColors:       []string{"81", "198", "197", "141", "214", "118", "227", "215"}, // Green, pink, red, purple, orange, light green, yellow, peach
	},
	"gruvbox": {
		Name:                "Gruvbox",
		IsDark:              true,
		SelectedBG:          "237",
		SecondarySelectedBG: "235",
		// PanelBG removed - using terminal natural background
		BorderColor:         "108", // Green
		ActiveBorderColor:   "214", // Orange
		InactiveBorderColor: "243",
		HeaderColor:         "214",                                                            // Orange
		StatusColor:         "142",                                                            // Yellow-green
		ErrorColor:          "167",                                                            // Red
		WarningColor:        "214",                                                            // Orange
		SuccessColor:        "142",                                                            // Green
		InfoColor:           "109",                                                            // Blue
		TodoColor:           "33",                                                             // Bright blue - actionable work
		DoingColor:          "75",                                                             // Medium blue - natural and active
		ReviewColor:         "153",                                                            // Light blue - awaiting feedback
		DoneColor:           "24",                                                             // Dim blue - completed, natural
		AccentColor:         "208",                                                            // Orange
		MutedColor:          "245",                                                            // Gray
		HighlightColor:      "214",                                                            // Orange
		FeatureColors:       []string{"142", "214", "167", "175", "109", "208", "172", "130"}, // Green, orange, red, purple, blue, bright orange, yellow-green, brown
	},
	"dracula": {
		Name:                "Dracula",
		IsDark:              true,
		SelectedBG:          "236",
		SecondarySelectedBG: "234",
		// PanelBG removed - using terminal natural background
		BorderColor:         "141", // Purple
		ActiveBorderColor:   "212", // Pink
		InactiveBorderColor: "240",
		HeaderColor:         "212",                                                           // Pink
		StatusColor:         "141",                                                           // Purple
		ErrorColor:          "203",                                                           // Red
		WarningColor:        "228",                                                           // Yellow
		SuccessColor:        "84",                                                            // Green
		InfoColor:           "117",                                                           // Cyan
		TodoColor:           "33",                                                            // Bright blue - immediate action items
		DoingColor:          "75",                                                            // Medium blue - active but not harsh
		ReviewColor:         "153",                                                           // Light blue - waiting for approval
		DoneColor:           "24",                                                            // Dim blue - completed, subtle
		AccentColor:         "212",                                                           // Pink
		MutedColor:          "60",                                                            // Gray
		HighlightColor:      "228",                                                           // Yellow
		FeatureColors:       []string{"212", "141", "117", "84", "228", "203", "147", "213"}, // Pink, purple, cyan, green, yellow, red, blue, magenta
	},
}

// InitializeThemeNew sets up the theme from configuration
func InitializeThemeNew(cfg *config.Config) {
	// Try to load the specified theme first, fall back to default
	themeName := cfg.UI.Theme.Name
	if themeName == "" {
		themeName = "default"
	}

	if predefinedTheme, exists := PredefinedThemes[themeName]; exists {
		ActiveTheme = predefinedTheme

		// Apply configurable status color scheme
		statusColorScheme := cfg.GetStatusColorScheme()
		colors := GetStatusColorHierarchy(statusColorScheme)
		ActiveTheme.ReviewColor = colors[0] // Highest attention
		ActiveTheme.DoingColor = colors[1]  // Second priority
		ActiveTheme.TodoColor = colors[2]   // Third priority
		ActiveTheme.DoneColor = colors[3]   // Lowest attention

		// Override with config values if specified
		// PanelBG override removed - using terminal natural background
		if cfg.UI.Theme.SelectedBG != "" {
			ActiveTheme.SelectedBG = cfg.UI.Theme.SelectedBG
		}
		if cfg.UI.Theme.BorderColor != "" {
			ActiveTheme.BorderColor = cfg.UI.Theme.BorderColor
		}
		if cfg.UI.Theme.StatusColor != "" {
			ActiveTheme.StatusColor = cfg.UI.Theme.StatusColor
		}
		if cfg.UI.Theme.HeaderColor != "" {
			ActiveTheme.HeaderColor = cfg.UI.Theme.HeaderColor
		}
		if cfg.UI.Theme.ErrorColor != "" {
			ActiveTheme.ErrorColor = cfg.UI.Theme.ErrorColor
		}

	} else {
		// Apply configurable status color scheme to default colors
		statusColorScheme := cfg.GetStatusColorScheme()
		colors := GetStatusColorHierarchy(statusColorScheme)

		// Fallback to manual configuration
		ActiveTheme = ThemeConfig{
			Name:                "Custom",
			IsDark:              true,
			SelectedBG:          cfg.UI.Theme.SelectedBG,
			SecondarySelectedBG: "235",
			// PanelBG removed - using terminal natural background
			BorderColor:         cfg.UI.Theme.BorderColor,
			ActiveBorderColor:   "51",
			InactiveBorderColor: "240",
			HeaderColor:         cfg.UI.Theme.HeaderColor,
			StatusColor:         cfg.UI.Theme.StatusColor,
			ErrorColor:          cfg.UI.Theme.ErrorColor,
			WarningColor:        "220",
			SuccessColor:        "46",
			InfoColor:           "51",
			ReviewColor:         colors[0], // Highest attention
			DoingColor:          colors[1], // Second priority
			TodoColor:           colors[2], // Third priority
			DoneColor:           colors[3], // Lowest attention
			AccentColor:         "75",
			MutedColor:          "244",
			HighlightColor:      "226",
		}
	}

	// Update styles with new theme
	updateStylesFromThemeNew()
}

// updateStylesFromThemeNew applies the current theme to all styles
func updateStylesFromThemeNew() {
	// Update header style with enhanced styling
	HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ActiveTheme.HeaderColor)).
		// Background handled by StyleContext when needed
		Padding(0, 2).
		MarginBottom(0)

	// Update status bar style
	StatusBarStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ActiveTheme.MutedColor)).
		// Background handled by StyleContext when needed
		Padding(0, 1).
		Width(0)

	// Update base panel style with enhanced borders
	BasePanelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ActiveTheme.BorderColor)).
		// Background handled by StyleContext when needed
		Padding(0, 1)

	// Update task detail headers style
	DetailHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ActiveTheme.HeaderColor))
		// Background handled by StyleContext when needed

	// Update scroll indicator style
	ScrollIndicatorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ActiveTheme.MutedColor))
		// Background handled by StyleContext when needed

	// Update tag style
	TagStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ActiveTheme.MutedColor))
	// Background handled by StyleContext when needed
}
