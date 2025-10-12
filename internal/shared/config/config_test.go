package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Test loading with defaults (no config file)
	config, err := Load()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if config == nil {
		t.Fatal("Expected config to be loaded")
	}

	// Test default values
	if config.Server.URL != "http://localhost:8181" {
		t.Errorf("Expected default URL, got %s", config.Server.URL)
	}

	if config.Server.Timeout != 30*time.Second {
		t.Errorf("Expected 30s timeout, got %v", config.Server.Timeout)
	}

	if !config.UI.Display.ShowCompletedTasks {
		t.Error("Expected ShowCompletedTasks to be true by default")
	}
}

func TestEnvironmentOverrides(t *testing.T) {
	// Set environment variables
	os.Setenv("LAZYARCHON_SERVER_URL", "http://test.example.com")
	os.Setenv("LAZYARCHON_API_KEY", "test-key-123")
	os.Setenv("LAZYARCHON_LOG_LEVEL", "debug")

	defer func() {
		os.Unsetenv("LAZYARCHON_SERVER_URL")
		os.Unsetenv("LAZYARCHON_API_KEY")
		os.Unsetenv("LAZYARCHON_LOG_LEVEL")
	}()

	config, err := Load()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if config.Server.URL != "http://test.example.com" {
		t.Errorf("Expected environment URL override, got %s", config.Server.URL)
	}

	if config.Server.APIKey != "test-key-123" {
		t.Errorf("Expected environment API key override, got %s", config.Server.APIKey)
	}

	if config.Development.LogLevel != "debug" {
		t.Errorf("Expected environment log level override, got %s", config.Development.LogLevel)
	}
}

func TestGetMethods(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			URL:    "http://example.com",
			APIKey: "test-key",
		},
		UI: UIConfig{
			Display: DisplayConfig{
				ShowCompletedTasks: false,
				DefaultSortMode:    "alphabetical",
			},
		},
	}

	if config.GetServerURL() != "http://example.com" {
		t.Errorf("Expected server URL, got %s", config.GetServerURL())
	}

	if config.GetAPIKey() != "test-key" {
		t.Errorf("Expected API key, got %s", config.GetAPIKey())
	}

	if config.ShouldShowCompletedTasks() != false {
		t.Error("Expected ShowCompletedTasks to be false")
	}

	if config.GetDefaultSortMode() != "alphabetical" {
		t.Errorf("Expected alphabetical sort mode, got %s", config.GetDefaultSortMode())
	}
}

func TestPredefinedThemes(t *testing.T) {
	tests := []struct {
		name           string
		themeName      string
		expectedBorder string
		expectedStatus string
	}{
		{"default theme", "default", "62", "205"},
		{"monokai theme", "monokai", "197", "148"},
		{"gruvbox theme", "gruvbox", "208", "142"},
		{"dracula theme", "dracula", "141", "212"},
		{"unknown theme", "unknown", "62", "205"}, // Should fall back to defaults
	}

	for _, tt := range tests { //nolint:varnamelen // tt is idiomatic for table-driven tests
		t.Run(tt.name, func(t *testing.T) {
			config := defaultConfig
			config.UI.Theme.Name = tt.themeName
			config.applyPredefinedTheme()

			if config.UI.Theme.BorderColor != tt.expectedBorder {
				t.Errorf("Expected border color %s, got %s", tt.expectedBorder, config.UI.Theme.BorderColor)
			}
			if config.UI.Theme.StatusColor != tt.expectedStatus {
				t.Errorf("Expected status color %s, got %s", tt.expectedStatus, config.UI.Theme.StatusColor)
			}
		})
	}
}

func TestThemeOverrides(t *testing.T) {
	config := defaultConfig
	config.UI.Theme.Name = "gruvbox"
	config.UI.Theme.BorderColor = "999" // Custom override

	config.applyPredefinedTheme()

	// Custom override should be preserved
	if config.UI.Theme.BorderColor != "999" {
		t.Errorf("Expected custom border color to be preserved, got %s", config.UI.Theme.BorderColor)
	}

	// Other gruvbox colors should be applied
	if config.UI.Theme.StatusColor != "142" {
		t.Errorf("Expected gruvbox status color 142, got %s", config.UI.Theme.StatusColor)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		shouldErr bool
		errMsg    string
	}{
		{
			name:      "valid config",
			config:    defaultConfig,
			shouldErr: false,
		},
		{
			name: "invalid theme name",
			config: Config{
				Version: "1.0.0",
				Profile: "dev",
				Server: ServerConfig{
					URL:     "http://localhost:8181",
					Timeout: 30 * time.Second,
				},
				UI: UIConfig{
					Theme: ThemeConfig{
						Name: "invalid-theme",
					},
					Display: DisplayConfig{
						DefaultSortMode:   "status+priority",
						StatusColorScheme: "blue",
					},
				},
				Development: DevelopmentConfig{
					LogLevel: "info",
				},
			},
			shouldErr: true,
			errMsg:    "Theme.Name",
		},
		{
			name: "invalid log level",
			config: Config{
				Version: "1.0.0",
				Profile: "dev",
				Server: ServerConfig{
					URL:     "http://localhost:8181",
					Timeout: 30 * time.Second,
				},
				UI: UIConfig{
					Theme: ThemeConfig{
						Name: "default",
					},
					Display: DisplayConfig{
						DefaultSortMode:   "status+priority",
						StatusColorScheme: "blue",
					},
				},
				Development: DevelopmentConfig{
					LogLevel: "invalid",
				},
			},
			shouldErr: true,
			errMsg:    "Development.LogLevel",
		},
	}

	for _, tt := range tests { //nolint:varnamelen // tt is idiomatic for table-driven tests
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected validation error but got none")
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error to contain '%s', got: %s", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error, got: %s", err.Error())
				}
			}
		})
	}
}

func TestProfileDefaults(t *testing.T) {
	tests := []struct {
		name     string
		profile  string
		expected struct {
			debug    bool
			logLevel string
		}
	}{
		{
			name:    "development profile",
			profile: "development",
			expected: struct {
				debug    bool
				logLevel string
			}{true, "debug"},
		},
		{
			name:    "production profile",
			profile: "production",
			expected: struct {
				debug    bool
				logLevel string
			}{false, "warn"},
		},
		{
			name:    "staging profile",
			profile: "staging",
			expected: struct {
				debug    bool
				logLevel string
			}{false, "info"},
		},
	}

	for _, tt := range tests { //nolint:varnamelen // tt is idiomatic for table-driven tests
		t.Run(tt.name, func(t *testing.T) {
			config := defaultConfig
			config.Profile = tt.profile
			config.applyProfileDefaults()

			if config.Development.Debug != tt.expected.debug {
				t.Errorf("Expected debug=%v for profile %s, got %v", tt.expected.debug, tt.profile, config.Development.Debug)
			}
			if config.Development.LogLevel != tt.expected.logLevel {
				t.Errorf("Expected logLevel=%s for profile %s, got %s", tt.expected.logLevel, tt.profile, config.Development.LogLevel)
			}
		})
	}
}

func TestProfileHelpers(t *testing.T) {
	tests := []struct {
		profile      string
		isDev        bool
		isStaging    bool
		isProduction bool
	}{
		{"development", true, false, false},
		{"dev", true, false, false},
		{"staging", false, true, false},
		{"production", false, false, true},
		{"prod", false, false, true},
		{"", true, false, false}, // Default to development
	}

	for _, tt := range tests { //nolint:varnamelen // tt is idiomatic for table-driven tests
		config := &Config{Profile: tt.profile}

		if config.IsDevelopmentProfile() != tt.isDev {
			t.Errorf("Profile %s: expected IsDevelopmentProfile()=%v, got %v",
				tt.profile, tt.isDev, config.IsDevelopmentProfile())
		}
		if config.IsStagingProfile() != tt.isStaging {
			t.Errorf("Profile %s: expected IsStagingProfile()=%v, got %v",
				tt.profile, tt.isStaging, config.IsStagingProfile())
		}
		if config.IsProductionProfile() != tt.isProduction {
			t.Errorf("Profile %s: expected IsProductionProfile()=%v, got %v",
				tt.profile, tt.isProduction, config.IsProductionProfile())
		}
	}
}

func TestDefaultProjectID(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		shouldErr bool
		errMsg    string
	}{
		{
			name: "valid UUID is accepted",
			config: Config{
				Version: "1.0.0",
				Profile: "dev",
				Server: ServerConfig{
					URL:     "http://localhost:8181",
					Timeout: 30 * time.Second,
				},
				UI: UIConfig{
					Theme: ThemeConfig{
						Name: "default",
					},
					Display: DisplayConfig{
						DefaultSortMode:   "status+priority",
						StatusColorScheme: "blue",
						DefaultProjectID:  "550e8400-e29b-41d4-a716-446655440000",
					},
				},
				Development: DevelopmentConfig{
					LogLevel: "info",
				},
			},
			shouldErr: false,
		},
		{
			name: "empty string is accepted",
			config: Config{
				Version: "1.0.0",
				Profile: "dev",
				Server: ServerConfig{
					URL:     "http://localhost:8181",
					Timeout: 30 * time.Second,
				},
				UI: UIConfig{
					Theme: ThemeConfig{
						Name: "default",
					},
					Display: DisplayConfig{
						DefaultSortMode:   "status+priority",
						StatusColorScheme: "blue",
						DefaultProjectID:  "",
					},
				},
				Development: DevelopmentConfig{
					LogLevel: "info",
				},
			},
			shouldErr: false,
		},
		{
			name: "invalid UUID format is rejected",
			config: Config{
				Version: "1.0.0",
				Profile: "dev",
				Server: ServerConfig{
					URL:     "http://localhost:8181",
					Timeout: 30 * time.Second,
				},
				UI: UIConfig{
					Theme: ThemeConfig{
						Name: "default",
					},
					Display: DisplayConfig{
						DefaultSortMode:   "status+priority",
						StatusColorScheme: "blue",
						DefaultProjectID:  "not-a-valid-uuid",
					},
				},
				Development: DevelopmentConfig{
					LogLevel: "info",
				},
			},
			shouldErr: true,
			errMsg:    "DefaultProjectID",
		},
	}

	for _, tt := range tests { //nolint:varnamelen // tt is idiomatic for table-driven tests
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected validation error but got none")
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error to contain '%s', got: %s", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error, got: %s", err.Error())
				}
			}
		})
	}
}

func TestDefaultProjectIDEnvironmentOverride(t *testing.T) {
	// Set environment variable
	os.Setenv("LAZYARCHON_DEFAULT_PROJECT_ID", "123e4567-e89b-12d3-a456-426614174000")
	defer os.Unsetenv("LAZYARCHON_DEFAULT_PROJECT_ID")

	config, err := Load()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if config.UI.Display.DefaultProjectID != "123e4567-e89b-12d3-a456-426614174000" {
		t.Errorf("Expected environment override for default project ID, got %s", config.UI.Display.DefaultProjectID)
	}
}

func TestGetDefaultProjectID(t *testing.T) {
	config := &Config{
		UI: UIConfig{
			Display: DisplayConfig{
				DefaultProjectID: "550e8400-e29b-41d4-a716-446655440000",
			},
		},
	}

	if config.GetDefaultProjectID() != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("Expected default project ID, got %s", config.GetDefaultProjectID())
	}

	// Test empty string
	config.UI.Display.DefaultProjectID = ""
	if config.GetDefaultProjectID() != "" {
		t.Errorf("Expected empty string, got %s", config.GetDefaultProjectID())
	}
}

func TestGetKeybindings(t *testing.T) {
	config := &Config{
		UI: UIConfig{
			Keybindings: KeybindingsConfig{
				Application: ApplicationKeybindings{
					Quit:    []string{"q"},
					Refresh: []string{"r", "F5"},
				},
			},
		},
	}

	keybindings := config.GetKeybindings()
	if keybindings == nil {
		t.Fatal("Expected keybindings to be returned")
	}

	if len(keybindings.Application.Quit) != 1 || keybindings.Application.Quit[0] != "q" {
		t.Errorf("Expected quit keybinding 'q', got %v", keybindings.Application.Quit)
	}

	if len(keybindings.Application.Refresh) != 2 {
		t.Errorf("Expected 2 refresh keybindings, got %d", len(keybindings.Application.Refresh))
	}
}

func TestKeybindingsValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		shouldErr bool
		errMsg    string
	}{
		{
			name: "valid keybindings with multiple keys",
			config: Config{
				Version: "1.0.0",
				Profile: "dev",
				Server: ServerConfig{
					URL:     "http://localhost:8181",
					Timeout: 30 * time.Second,
				},
				UI: UIConfig{
					Theme: ThemeConfig{
						Name: "default",
					},
					Display: DisplayConfig{
						DefaultSortMode:   "status+priority",
						StatusColorScheme: "blue",
					},
					Keybindings: KeybindingsConfig{
						Application: ApplicationKeybindings{
							Quit:    []string{"q"},
							Refresh: []string{"r", "F5"},
						},
						Navigation: NavigationKeybindings{
							Up:   []string{"k", "up"},
							Down: []string{"j", "down"},
						},
					},
				},
				Development: DevelopmentConfig{
					LogLevel: "info",
				},
			},
			shouldErr: false,
		},
		{
			name: "empty keybindings are valid (will use defaults)",
			config: Config{
				Version: "1.0.0",
				Profile: "dev",
				Server: ServerConfig{
					URL:     "http://localhost:8181",
					Timeout: 30 * time.Second,
				},
				UI: UIConfig{
					Theme: ThemeConfig{
						Name: "default",
					},
					Display: DisplayConfig{
						DefaultSortMode:   "status+priority",
						StatusColorScheme: "blue",
					},
					Keybindings: KeybindingsConfig{},
				},
				Development: DevelopmentConfig{
					LogLevel: "info",
				},
			},
			shouldErr: false,
		},
		{
			name: "invalid keybinding - empty string in array",
			config: Config{
				Version: "1.0.0",
				Profile: "dev",
				Server: ServerConfig{
					URL:     "http://localhost:8181",
					Timeout: 30 * time.Second,
				},
				UI: UIConfig{
					Theme: ThemeConfig{
						Name: "default",
					},
					Display: DisplayConfig{
						DefaultSortMode:   "status+priority",
						StatusColorScheme: "blue",
					},
					Keybindings: KeybindingsConfig{
						Application: ApplicationKeybindings{
							Quit: []string{""}, // Invalid: empty string
						},
					},
				},
				Development: DevelopmentConfig{
					LogLevel: "info",
				},
			},
			shouldErr: true,
			errMsg:    "Quit",
		},
	}

	for _, tt := range tests { //nolint:varnamelen // tt is idiomatic for table-driven tests
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected validation error but got none")
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error to contain '%s', got: %s", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error, got: %s", err.Error())
				}
			}
		})
	}
}

func TestKeybindingsStructure(t *testing.T) {
	// Test that keybindings config has proper structure
	keybindings := KeybindingsConfig{
		Application: ApplicationKeybindings{
			Quit:         []string{"q"},
			ForceQuit:    []string{"ctrl+c"},
			Refresh:      []string{"r", "F5"},
			ProjectMode:  []string{"p"},
			ShowAllTasks: []string{"a"},
			ToggleHelp:   []string{"?"},
		},
		Navigation: NavigationKeybindings{
			Up:             []string{"k", "up"},
			Down:           []string{"j", "down"},
			Left:           []string{"h", "left"},
			Right:          []string{"l", "right"},
			JumpFirst:      []string{"gg", "home"},
			JumpLast:       []string{"G", "end"},
			FastScrollUp:   []string{"K"},
			FastScrollDown: []string{"J"},
			HalfPageUp:     []string{"ctrl+u", "pgup"},
			HalfPageDown:   []string{"ctrl+d", "pgdown"},
		},
		Search: SearchKeybindings{
			Activate:  []string{"/", "ctrl+f"},
			Clear:     []string{"ctrl+x", "ctrl+l"},
			NextMatch: []string{"n"},
			PrevMatch: []string{"N"},
		},
		Task: TaskKeybindings{
			ChangeStatus:  []string{"t"},
			Edit:          []string{"e"},
			Delete:        []string{"d"},
			CopyID:        []string{"y"},
			CopyTitle:     []string{"Y"},
			SelectFeature: []string{"f"},
			SortForward:   []string{"s"},
			SortBackward:  []string{"S"},
		},
	}

	// Verify all fields are accessible and have expected types
	if len(keybindings.Application.Quit) == 0 {
		t.Error("Expected quit keybinding to be set")
	}
	if len(keybindings.Navigation.Up) == 0 {
		t.Error("Expected up navigation keybinding to be set")
	}
	if len(keybindings.Search.Activate) == 0 {
		t.Error("Expected search activate keybinding to be set")
	}
	if len(keybindings.Task.ChangeStatus) == 0 {
		t.Error("Expected task change status keybinding to be set")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
