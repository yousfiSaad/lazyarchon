package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server      ServerConfig      `yaml:"server"`
	UI          UIConfig          `yaml:"ui"`
	Development DevelopmentConfig `yaml:"development"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	URL     string        `yaml:"url"`
	Timeout time.Duration `yaml:"timeout"`
	APIKey  string        `yaml:"api_key"`
}

// UIConfig holds UI-related configuration
type UIConfig struct {
	Theme       ThemeConfig   `yaml:"theme"`
	Display     DisplayConfig `yaml:"display"`
	Keybindings interface{}   `yaml:"keybindings"` // Future enhancement
}

// ThemeConfig holds theme/color configuration
type ThemeConfig struct {
	Name        string `yaml:"name"`         // Predefined theme name (default, monokai, gruvbox, dracula)
	// PanelBG removed - using terminal natural background
	SelectedBG  string `yaml:"selected_bg"`
	BorderColor string `yaml:"border_color"`
	StatusColor string `yaml:"status_color"`
	HeaderColor string `yaml:"header_color"`
	ErrorColor  string `yaml:"error_color"`
}

// DisplayConfig holds display-related settings
type DisplayConfig struct {
	ShowCompletedTasks  bool   `yaml:"show_completed_tasks"`
	DefaultSortMode     string `yaml:"default_sort_mode"`
	AutoRefreshInterval int    `yaml:"auto_refresh_interval"`

	// Color enhancement options
	FeatureColors       bool   `yaml:"feature_colors"`       // Enable vibrant feature tag colors
	FeatureBackgrounds  bool   `yaml:"feature_backgrounds"`  // Enable subtle background tints for feature groups
	PriorityIndicators  bool   `yaml:"priority_indicators"`  // Show priority symbols and colors
	StatusColorScheme   string `yaml:"status_color_scheme"`  // Task status color hierarchy (blue, gray, warm_gray, cool_gray)
}

// DevelopmentConfig holds development-related settings
type DevelopmentConfig struct {
	Debug           bool   `yaml:"debug"`
	LogLevel        string `yaml:"log_level"`
	EnableProfiling bool   `yaml:"enable_profiling"`
}

// Default configuration values
var defaultConfig = Config{
	Server: ServerConfig{
		URL:     "http://localhost:8181",
		Timeout: 30 * time.Second,
		APIKey:  "",
	},
	UI: UIConfig{
		Theme: ThemeConfig{
			Name:        "default",
			// PanelBG removed - using terminal natural background
			SelectedBG:  "237",
			BorderColor: "62",
			StatusColor: "205",
			HeaderColor: "39",
			ErrorColor:  "196",
		},
		Display: DisplayConfig{
			ShowCompletedTasks:  true,
			DefaultSortMode:     "status+priority",
			AutoRefreshInterval: 0,
			FeatureColors:       true,  // Enable feature colors by default
			FeatureBackgrounds:  false, // Disable background tints by default (subtle)
			PriorityIndicators:  true,  // Enable priority indicators by default
			StatusColorScheme:   "blue", // Default to current blue scheme
		},
	},
	Development: DevelopmentConfig{
		Debug:           false,
		LogLevel:        "info",
		EnableProfiling: false,
	},
}

// LoadFromPath loads configuration from a specific file path
func LoadFromPath(configPath string) (*Config, error) {
	config := defaultConfig // Start with defaults

	// Check if the specified config file exists
	if _, err := os.Stat(configPath); err != nil {
		return &config, fmt.Errorf("config file not found: %s", configPath)
	}

	// Read and parse config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return &config, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML, merging with defaults
	if err := yaml.Unmarshal(data, &config); err != nil {
		return &config, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Override with environment variables if present
	config.applyEnvironmentOverrides()

	// Apply predefined theme if specified
	config.applyPredefinedTheme()

	return &config, nil
}

// Load loads configuration from file with fallback to defaults
func Load() (*Config, error) {
	config := defaultConfig // Start with defaults

	// Try to find config file in order of preference
	configPaths := []string{
		"./config.yaml",
		"./configs/default.yaml",
		filepath.Join(os.Getenv("HOME"), ".config", "lazyarchon", "config.yaml"),
		"/etc/lazyarchon/config.yaml",
	}

	var configFile string
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			configFile = path
			break
		}
	}

	// If no config file found, use defaults but still apply environment overrides
	if configFile == "" {
		config.applyEnvironmentOverrides()
		config.applyPredefinedTheme()
		return &config, nil
	}

	// Read and parse config file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return &config, err // Return defaults even on error
	}

	// Parse YAML, merging with defaults
	if err := yaml.Unmarshal(data, &config); err != nil {
		return &config, err // Return defaults even on error
	}

	// Override with environment variables if present
	config.applyEnvironmentOverrides()

	// Apply predefined theme if specified
	config.applyPredefinedTheme()

	return &config, nil
}

// applyEnvironmentOverrides applies environment variable overrides
func (c *Config) applyEnvironmentOverrides() {
	if url := os.Getenv("LAZYARCHON_SERVER_URL"); url != "" {
		c.Server.URL = url
	}
	if apiKey := os.Getenv("LAZYARCHON_API_KEY"); apiKey != "" {
		c.Server.APIKey = apiKey
	}
	if logLevel := os.Getenv("LAZYARCHON_LOG_LEVEL"); logLevel != "" {
		c.Development.LogLevel = logLevel
	}
}

// GetServerURL returns the configured server URL
func (c *Config) GetServerURL() string {
	return c.Server.URL
}

// GetAPIKey returns the configured API key
func (c *Config) GetAPIKey() string {
	return c.Server.APIKey
}

// ShouldShowCompletedTasks returns whether to show completed tasks by default
func (c *Config) ShouldShowCompletedTasks() bool {
	return c.UI.Display.ShowCompletedTasks
}

// GetDefaultSortMode returns the default sort mode
func (c *Config) GetDefaultSortMode() string {
	return c.UI.Display.DefaultSortMode
}

// IsFeatureColorsEnabled returns whether feature tag colors are enabled
func (c *Config) IsFeatureColorsEnabled() bool {
	return c.UI.Display.FeatureColors
}

// IsFeatureBackgroundsEnabled returns whether feature background tints are enabled
func (c *Config) IsFeatureBackgroundsEnabled() bool {
	return c.UI.Display.FeatureBackgrounds
}

// IsPriorityIndicatorsEnabled returns whether priority indicators are enabled
func (c *Config) IsPriorityIndicatorsEnabled() bool {
	return c.UI.Display.PriorityIndicators
}

// GetStatusColorScheme returns the configured status color scheme
func (c *Config) GetStatusColorScheme() string {
	if c.UI.Display.StatusColorScheme == "" {
		return "blue" // Default fallback
	}
	return c.UI.Display.StatusColorScheme
}

// GetTheme returns the theme configuration
func (c *Config) GetTheme() interface{} {
	return &c.UI.Theme
}

// GetDisplay returns the display configuration
func (c *Config) GetDisplay() interface{} {
	return &c.UI.Display
}

// GetDevelopment returns the development configuration
func (c *Config) GetDevelopment() interface{} {
	return &c.Development
}

// IsDebugEnabled returns whether debug mode is enabled
func (c *Config) IsDebugEnabled() bool {
	return c.Development.Debug
}

// IsDarkModeEnabled returns whether dark mode is enabled (always true for terminal apps)
func (c *Config) IsDarkModeEnabled() bool {
	return true // Terminal applications are inherently dark mode
}

// IsCompletedTasksVisible returns whether completed tasks should be shown
func (c *Config) IsCompletedTasksVisible() bool {
	return c.UI.Display.ShowCompletedTasks
}

// applyPredefinedTheme applies a predefined theme if specified
func (c *Config) applyPredefinedTheme() {
	if c.UI.Theme.Name == "" {
		return // No theme name specified, keep current colors
	}

	// Define predefined themes (from internal/ui/styles.go)
	predefinedThemes := map[string]ThemeConfig{
		"default": {
			Name:        "default",
			SelectedBG:  "237", // Dark gray
			BorderColor: "62",  // Purple/blue
			StatusColor: "205", // Light magenta
			HeaderColor: "39",  // Bright cyan
			ErrorColor:  "196", // Bright red
		},
		"monokai": {
			Name:        "monokai",
			SelectedBG:  "235", // Dark gray
			BorderColor: "197", // Pink
			StatusColor: "148", // Green
			HeaderColor: "81",  // Cyan
			ErrorColor:  "197", // Pink/red
		},
		"gruvbox": {
			Name:        "gruvbox",
			SelectedBG:  "237", // Dark gray
			BorderColor: "208", // Orange
			StatusColor: "142", // Green
			HeaderColor: "214", // Yellow
			ErrorColor:  "167", // Red
		},
		"dracula": {
			Name:        "dracula",
			SelectedBG:  "236", // Dark gray
			BorderColor: "141", // Purple
			StatusColor: "212", // Pink
			HeaderColor: "117", // Cyan
			ErrorColor:  "203", // Red
		},
	}

	// Get predefined theme
	theme, exists := predefinedThemes[c.UI.Theme.Name]
	if !exists {
		return // Unknown theme name, keep current colors
	}

	// Apply theme colors only if not explicitly overridden in config
	// This maintains backward compatibility with manual color overrides
	if c.UI.Theme.SelectedBG == "" || c.UI.Theme.SelectedBG == defaultConfig.UI.Theme.SelectedBG {
		c.UI.Theme.SelectedBG = theme.SelectedBG
	}
	if c.UI.Theme.BorderColor == "" || c.UI.Theme.BorderColor == defaultConfig.UI.Theme.BorderColor {
		c.UI.Theme.BorderColor = theme.BorderColor
	}
	if c.UI.Theme.StatusColor == "" || c.UI.Theme.StatusColor == defaultConfig.UI.Theme.StatusColor {
		c.UI.Theme.StatusColor = theme.StatusColor
	}
	if c.UI.Theme.HeaderColor == "" || c.UI.Theme.HeaderColor == defaultConfig.UI.Theme.HeaderColor {
		c.UI.Theme.HeaderColor = theme.HeaderColor
	}
	if c.UI.Theme.ErrorColor == "" || c.UI.Theme.ErrorColor == defaultConfig.UI.Theme.ErrorColor {
		c.UI.Theme.ErrorColor = theme.ErrorColor
	}
}
