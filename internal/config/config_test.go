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
