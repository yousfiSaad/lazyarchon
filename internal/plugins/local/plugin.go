package local

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/yousfisaad/lazyarchon/v2/internal/plugin"
)

// LocalPlugin is the factory for the SQLite-backed local backend.
type LocalPlugin struct{}

// compile-time interface check
var _ plugin.Plugin = (*LocalPlugin)(nil)

// Name implements plugin.Plugin
func (p *LocalPlugin) Name() string { return PluginName }

// Version implements plugin.Plugin
func (p *LocalPlugin) Version() string { return "1.0.0" }

// Description implements plugin.Plugin
func (p *LocalPlugin) Description() string {
	return "Local SQLite task storage - no server required, data stays on this machine"
}

// Capabilities implements plugin.Plugin
func (p *LocalPlugin) Capabilities() plugin.Capabilities {
	return plugin.Capabilities{
		SupportsProjects:    true,
		SupportsStatuses:    true,
		SupportsPriority:    true,
		SupportsAssignees:   true,
		SupportsDueDates:    true,
		SupportsSubtasks:    true,
		SupportsTags:        true,
		SupportsDescription: true,
		SupportsArchiving:   true,
		SupportsSearch:      true,
	}
}

// CreateClient implements plugin.Plugin. The database path is read from
// config.Extra["path"]; when absent the default data directory is used.
func (p *LocalPlugin) CreateClient(config plugin.PluginConfig) (plugin.TaskClient, error) {
	path := DefaultDBPath()

	if config.Extra != nil {
		if value, ok := config.Extra["path"]; ok {
			if str, ok := value.(string); ok && str != "" {
				path = str
			}
		}
	}

	client, err := newClient(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open local database %q: %w", path, err)
	}

	return client, nil
}

// DefaultDataDir returns the default directory for local data, following
// platform conventions: XDG data dir on Linux, Application Support on macOS,
// AppData on Windows.
func DefaultDataDir() (string, error) {
	if runtime.GOOS == "linux" {
		if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
			return filepath.Join(xdg, "lazyarchon"), nil
		}

		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to determine home directory: %w", err)
		}

		return filepath.Join(home, ".local", "share", "lazyarchon"), nil
	}

	// darwin: ~/Library/Application Support/lazyarchon
	// windows: %AppData%\lazyarchon
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine config directory: %w", err)
	}

	return filepath.Join(base, "lazyarchon"), nil
}

// DefaultDBPath returns the default database file path.
func DefaultDBPath() string {
	dir, err := DefaultDataDir()
	if err != nil {
		// Fall back to the current directory if platform dirs are unavailable.
		return "lazyarchon.db"
	}

	return filepath.Join(dir, "lazyarchon.db")
}

func init() {
	plugin.Register(&LocalPlugin{})
}
