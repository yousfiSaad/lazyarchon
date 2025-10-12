package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Version     string            `yaml:"version,omitempty" validate:"omitempty,semver"`
	Profile     string            `yaml:"profile,omitempty" validate:"omitempty,oneof=dev development staging production prod"`
	Server      ServerConfig      `yaml:"server" validate:"required"`
	UI          UIConfig          `yaml:"ui" validate:"required"`
	Development DevelopmentConfig `yaml:"development" validate:"required"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	URL            string        `yaml:"url" validate:"required,url"`
	Timeout        time.Duration `yaml:"timeout" validate:"min=1s,max=300s"`
	APIKey         string        `yaml:"api_key" validate:"omitempty,min=10"`
	EnableRealtime bool          `yaml:"enable_realtime"` // Enable HTTP polling for auto-refresh (WebSocket not supported by backend)
	PollingInterval int          `yaml:"polling_interval" validate:"min=0,max=300"` // Polling interval in seconds (0 = disabled, default: 10)
}

// UIConfig holds UI-related configuration
type UIConfig struct {
	Theme       ThemeConfig       `yaml:"theme" validate:"required"`
	Display     DisplayConfig     `yaml:"display" validate:"required"`
	Keybindings KeybindingsConfig `yaml:"keybindings"` // Keyboard shortcuts customization
}

// ThemeConfig holds theme/color configuration
type ThemeConfig struct {
	Name string `yaml:"name" validate:"oneof=default monokai gruvbox dracula"` // Predefined theme name
	// PanelBG removed - using terminal natural background
	SelectedBG  string `yaml:"selected_bg" validate:"omitempty,numeric"`
	BorderColor string `yaml:"border_color" validate:"omitempty,numeric"`
	StatusColor string `yaml:"status_color" validate:"omitempty,numeric"`
	HeaderColor string `yaml:"header_color" validate:"omitempty,numeric"`
	ErrorColor  string `yaml:"error_color" validate:"omitempty,numeric"`
}

// DisplayConfig holds display-related settings
type DisplayConfig struct {
	ShowCompletedTasks  bool   `yaml:"show_completed_tasks"`
	DefaultSortMode     string `yaml:"default_sort_mode" validate:"oneof=status+priority priority time alphabetical"`
	AutoRefreshInterval int    `yaml:"auto_refresh_interval" validate:"min=0,max=300"`

	// Color enhancement options
	FeatureColors      bool   `yaml:"feature_colors"`                                                     // Enable vibrant feature tag colors
	FeatureBackgrounds bool   `yaml:"feature_backgrounds"`                                                // Enable subtle background tints for feature groups
	PriorityIndicators bool   `yaml:"priority_indicators"`                                                // Show priority symbols and colors
	StatusColorScheme  string `yaml:"status_color_scheme" validate:"oneof=blue gray warm_gray cool_gray"` // Task status color hierarchy

	// Startup behavior
	DefaultProjectID string `yaml:"default_project_id" validate:"omitempty,uuid"` // Default project to select on startup (empty = "All Tasks")
}

// KeybindingsConfig holds customizable keyboard shortcuts
// All fields are optional - if not specified, defaults from keys package are used
type KeybindingsConfig struct {
	Application ApplicationKeybindings `yaml:"application"`
	Navigation  NavigationKeybindings  `yaml:"navigation"`
	Search      SearchKeybindings      `yaml:"search"`
	Task        TaskKeybindings        `yaml:"task"`
}

// ApplicationKeybindings defines application-level keyboard shortcuts
type ApplicationKeybindings struct {
	Quit         []string `yaml:"quit" validate:"omitempty,dive,min=1"`           // Smart quit (e.g., ["q"])
	ForceQuit    []string `yaml:"force_quit" validate:"omitempty,dive,min=1"`     // Emergency quit (e.g., ["ctrl+c"])
	Refresh      []string `yaml:"refresh" validate:"omitempty,dive,min=1"`        // Refresh data (e.g., ["r", "F5"])
	ProjectMode  []string `yaml:"project_mode" validate:"omitempty,dive,min=1"`   // Activate project selection (e.g., ["p"])
	ShowAllTasks []string `yaml:"show_all_tasks" validate:"omitempty,dive,min=1"` // Show all tasks (e.g., ["a"])
	ToggleHelp   []string `yaml:"toggle_help" validate:"omitempty,dive,min=1"`    // Toggle help modal (e.g., ["?"])
}

// NavigationKeybindings defines navigation keyboard shortcuts
type NavigationKeybindings struct {
	Up           []string `yaml:"up" validate:"omitempty,dive,min=1"`            // Move up (e.g., ["k", "up"])
	Down         []string `yaml:"down" validate:"omitempty,dive,min=1"`          // Move down (e.g., ["j", "down"])
	Left         []string `yaml:"left" validate:"omitempty,dive,min=1"`          // Move left (e.g., ["h", "left"])
	Right        []string `yaml:"right" validate:"omitempty,dive,min=1"`         // Move right (e.g., ["l", "right"])
	JumpFirst    []string `yaml:"jump_first" validate:"omitempty,dive,min=1"`    // Jump to first (e.g., ["gg", "home"])
	JumpLast     []string `yaml:"jump_last" validate:"omitempty,dive,min=1"`     // Jump to last (e.g., ["G", "end"])
	FastScrollUp []string `yaml:"fast_scroll_up" validate:"omitempty,dive,min=1"` // Fast scroll up (e.g., ["K"])
	FastScrollDown []string `yaml:"fast_scroll_down" validate:"omitempty,dive,min=1"` // Fast scroll down (e.g., ["J"])
	HalfPageUp   []string `yaml:"half_page_up" validate:"omitempty,dive,min=1"`  // Half page up (e.g., ["ctrl+u", "pgup"])
	HalfPageDown []string `yaml:"half_page_down" validate:"omitempty,dive,min=1"` // Half page down (e.g., ["ctrl+d", "pgdown"])
}

// SearchKeybindings defines search-related keyboard shortcuts
type SearchKeybindings struct {
	Activate  []string `yaml:"activate" validate:"omitempty,dive,min=1"`   // Activate search (e.g., ["/", "ctrl+f"])
	Clear     []string `yaml:"clear" validate:"omitempty,dive,min=1"`      // Clear search (e.g., ["ctrl+x", "ctrl+l"])
	NextMatch []string `yaml:"next_match" validate:"omitempty,dive,min=1"` // Next search match (e.g., ["n"])
	PrevMatch []string `yaml:"prev_match" validate:"omitempty,dive,min=1"` // Previous search match (e.g., ["N"])
}

// TaskKeybindings defines task operation keyboard shortcuts
type TaskKeybindings struct {
	ChangeStatus    []string `yaml:"change_status" validate:"omitempty,dive,min=1"`    // Change task status (e.g., ["t"])
	Edit            []string `yaml:"edit" validate:"omitempty,dive,min=1"`             // Edit task (e.g., ["e"])
	Delete          []string `yaml:"delete" validate:"omitempty,dive,min=1"`           // Delete task (e.g., ["d"])
	CopyID          []string `yaml:"copy_id" validate:"omitempty,dive,min=1"`          // Copy task ID (e.g., ["y"])
	CopyTitle       []string `yaml:"copy_title" validate:"omitempty,dive,min=1"`       // Copy task title (e.g., ["Y"])
	SelectFeature   []string `yaml:"select_feature" validate:"omitempty,dive,min=1"`   // Select feature (e.g., ["f"])
	SortForward     []string `yaml:"sort_forward" validate:"omitempty,dive,min=1"`     // Sort forward (e.g., ["s"])
	SortBackward    []string `yaml:"sort_backward" validate:"omitempty,dive,min=1"`    // Sort backward (e.g., ["S"])
}

// DevelopmentConfig holds development-related settings
type DevelopmentConfig struct {
	Debug           bool   `yaml:"debug"`
	LogLevel        string `yaml:"log_level" validate:"oneof=debug info warn error"`
	EnableProfiling bool   `yaml:"enable_profiling"`
}

// Global validator instance
var validate *validator.Validate

func init() {
	validate = validator.New()
}

// Default configuration values
var defaultConfig = Config{
	Version: "1.0.0",
	Profile: "development",
	Server: ServerConfig{
		URL:             "http://localhost:8181",
		Timeout:         30 * time.Second,
		APIKey:          "",
		EnableRealtime:  false, // Disabled by default - backend doesn't support WebSocket
		PollingInterval: 10,    // Default 10 seconds for HTTP polling
	},
	UI: UIConfig{
		Theme: ThemeConfig{
			Name: "default",
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
			FeatureColors:       true,   // Enable feature colors by default
			FeatureBackgrounds:  false,  // Disable background tints by default (subtle)
			PriorityIndicators:  true,   // Enable priority indicators by default
			StatusColorScheme:   "blue", // Default to current blue scheme
			DefaultProjectID:    "",     // Empty = "All Tasks" view on startup
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

	// Validate configuration
	if err := validate.Struct(&config); err != nil {
		return &config, fmt.Errorf("config validation failed: %w", err)
	}

	// Override with environment variables if present
	config.applyEnvironmentOverrides()

	// Apply predefined theme if specified
	config.applyPredefinedTheme()

	// Apply profile-specific configuration
	config.applyProfileDefaults()

	// Validate again after environment overrides and profile application
	if err := validate.Struct(&config); err != nil {
		return &config, fmt.Errorf("config validation failed after environment overrides and profile application: %w", err)
	}

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

	// Validate configuration
	if err := validate.Struct(&config); err != nil {
		return &config, fmt.Errorf("config validation failed: %w", err)
	}

	// Override with environment variables if present
	config.applyEnvironmentOverrides()

	// Apply predefined theme if specified
	config.applyPredefinedTheme()

	// Apply profile-specific configuration
	config.applyProfileDefaults()

	// Validate again after environment overrides and profile application
	if err := validate.Struct(&config); err != nil {
		return &config, fmt.Errorf("config validation failed after environment overrides and profile application: %w", err)
	}

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
	if projectID := os.Getenv("LAZYARCHON_DEFAULT_PROJECT_ID"); projectID != "" {
		c.UI.Display.DefaultProjectID = projectID
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

// GetDefaultProjectID returns the configured default project ID
func (c *Config) GetDefaultProjectID() string {
	return c.UI.Display.DefaultProjectID
}

// GetTheme returns the theme configuration
func (c *Config) GetTheme() *ThemeConfig {
	return &c.UI.Theme
}

// GetDisplay returns the display configuration
func (c *Config) GetDisplay() *DisplayConfig {
	return &c.UI.Display
}

// GetKeybindings returns the keybindings configuration
func (c *Config) GetKeybindings() *KeybindingsConfig {
	return &c.UI.Keybindings
}

// GetDevelopment returns the development configuration
func (c *Config) GetDevelopment() *DevelopmentConfig {
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

// IsRealtimeEnabled returns whether HTTP polling is enabled for auto-refresh
func (c *Config) IsRealtimeEnabled() bool {
	return c.Server.EnableRealtime
}

// GetPollingInterval returns the polling interval in seconds (default: 10)
func (c *Config) GetPollingInterval() int {
	if c.Server.PollingInterval == 0 {
		return 10 // Default to 10 seconds
	}
	return c.Server.PollingInterval
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

// applyProfileDefaults applies profile-specific configuration defaults
func (c *Config) applyProfileDefaults() {
	// Get profile from environment variable if not set in config
	if c.Profile == "" {
		if profile := os.Getenv("LAZYARCHON_PROFILE"); profile != "" {
			c.Profile = profile
		} else {
			c.Profile = "development" // Default profile
		}
	}

	switch c.Profile {
	case "development", "dev":
		c.Development.Debug = true
		c.Development.LogLevel = "debug"
		c.Development.EnableProfiling = true
		if c.Server.URL == "http://localhost:8181" { // Only override if using default
			c.Server.URL = "http://localhost:8181"
		}

	case "staging":
		c.Development.Debug = false
		c.Development.LogLevel = "info"
		c.Development.EnableProfiling = false
		if c.Server.URL == "http://localhost:8181" { // Only override if using default
			c.Server.URL = "http://staging.archon.example.com"
		}

	case "production", "prod":
		c.Development.Debug = false
		c.Development.LogLevel = "warn"
		c.Development.EnableProfiling = false
		if c.Server.URL == "http://localhost:8181" { // Only override if using default
			c.Server.URL = "http://archon.example.com"
		}
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	return validate.Struct(c)
}

// GetProfile returns the current configuration profile
func (c *Config) GetProfile() string {
	if c.Profile == "" {
		return "development"
	}
	return c.Profile
}

// IsProductionProfile returns true if running in production profile
func (c *Config) IsProductionProfile() bool {
	profile := c.GetProfile()
	return profile == "production" || profile == "prod"
}

// IsDevelopmentProfile returns true if running in development profile
func (c *Config) IsDevelopmentProfile() bool {
	profile := c.GetProfile()
	return profile == "development" || profile == "dev"
}

// IsStagingProfile returns true if running in staging profile
func (c *Config) IsStagingProfile() bool {
	return c.GetProfile() == "staging"
}
