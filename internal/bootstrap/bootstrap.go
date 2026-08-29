// Package bootstrap wires configuration, logging, and backend plugins for
// both lazyarchon entrypoints: the TUI and the stdio MCP server.
package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/yousfisaad/lazyarchon/v2/internal/logging"
	"github.com/yousfisaad/lazyarchon/v2/internal/plugin"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/config"

	// Register built-in plugins.
	_ "github.com/yousfisaad/lazyarchon/v2/internal/plugins/archon"
	_ "github.com/yousfisaad/lazyarchon/v2/internal/plugins/gitea"
	_ "github.com/yousfisaad/lazyarchon/v2/internal/plugins/local"
	_ "github.com/yousfisaad/lazyarchon/v2/internal/plugins/vikunja"
)

// LoadConfig loads configuration, tolerating errors with defaults (the TUI
// stays usable when no config file exists).
func LoadConfig() *config.Config {
	cfg, err := config.Load()
	if err != nil {
		slog.Warn("error while loading configs", "error", err.Error(), "fallback", "using default configs")
	}

	return cfg
}

// SetupLogging points the application logger and the global slog default
// at the log file. Must run after CLI flags have been applied (the log file
// path arrives via LAZYARCHON_LOG_FILE).
//
//nolint:ireturn // Returns the interfaces.Logger used throughout the UI
func SetupLogging(cfg *config.Config) *logging.SlogLogger {
	logger := logging.NewSlogLogger(cfg.IsDebugEnabled())
	logger.SetDefault()

	return logger
}

// InitPlugin loads the configured backend plugin and verifies connectivity.
// The local plugin needs no HTTP settings — only the database path.
func InitPlugin(ctx context.Context, cfg *config.Config) (*plugin.Manager, error) {
	pluginName := cfg.GetPlugin()

	slog.Info("initializing plugin", "plugin", pluginName)

	manager := plugin.NewManagerWithGlobalRegistry()

	if err := manager.LoadPlugin(pluginName, BuildPluginConfig(cfg, pluginName)); err != nil {
		return nil, fmt.Errorf("failed to load plugin '%s': %w", pluginName, err)
	}

	if err := manager.HealthCheck(ctx); err != nil {
		return nil, fmt.Errorf("health check failed for plugin '%s': %w", pluginName, err)
	}

	slog.Info("plugin initialized", "plugin", pluginName)

	return manager, nil
}

// BuildPluginConfig assembles the plugin.PluginConfig for a backend: the
// local plugin receives the database path via Extra; HTTP plugins fall
// back from their plugin-specific settings to the global server block.
func BuildPluginConfig(cfg *config.Config, pluginName string) plugin.PluginConfig {
	if pluginName == "local" {
		return plugin.PluginConfig{
			Extra: map[string]interface{}{"path": cfg.LocalDBPath()},
		}
	}

	pluginCfg := cfg.GetPluginConfig(pluginName)

	baseURL := pluginCfg.BaseURL
	if baseURL == "" {
		baseURL = cfg.GetServerURL()
	}

	apiKey := pluginCfg.APIKey
	if apiKey == "" {
		apiKey = cfg.GetAPIKey()
	}

	timeout := pluginCfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return plugin.PluginConfig{
		BaseURL:       baseURL,
		APIKey:        apiKey,
		AuthToken:     pluginCfg.AuthToken,
		Username:      pluginCfg.Username,
		Password:      pluginCfg.Password,
		Timeout:       timeout,
		CustomHeaders: pluginCfg.CustomHeaders,
		Mappings: plugin.FieldMappings{
			Status:    pluginCfg.Mappings.Status,
			Priority:  pluginCfg.Mappings.Priority,
			FieldName: pluginCfg.Mappings.FieldName,
			Tags:      pluginCfg.Mappings.Tags,
		},
	}
}
