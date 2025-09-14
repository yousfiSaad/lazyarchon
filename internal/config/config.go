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
