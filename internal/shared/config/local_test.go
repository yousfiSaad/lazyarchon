package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultPluginIsLocal(t *testing.T) {
	config, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got := config.GetPlugin(); got != "local" {
		t.Errorf("default plugin = %q, want local", got)
	}

	// An explicitly cleared plugin field must also fall back to local.
	config.Plugin = ""
	if got := config.GetPlugin(); got != "local" {
		t.Errorf("empty plugin = %q, want local fallback", got)
	}
}

func TestEmptyServerURLValidates(t *testing.T) {
	config := defaultConfig
	config.Plugin = "local"
	config.Server.URL = ""

	if err := config.Validate(); err != nil {
		t.Errorf("config with empty server.url failed validation: %v", err)
	}
}

func TestLocalDBPathPrecedence(t *testing.T) {
	config := defaultConfig
	config.Plugins.Local.Path = ""

	// No env, no config: empty means the plugin picks its data directory.
	if got := config.LocalDBPath(); got != "" {
		t.Errorf("LocalDBPath() = %q, want empty default", got)
	}

	// Config path wins over the default.
	config.Plugins.Local.Path = "/tmp/from-config.db"
	if got := config.LocalDBPath(); got != "/tmp/from-config.db" {
		t.Errorf("LocalDBPath() = %q, want /tmp/from-config.db", got)
	}

	// Environment variable wins over the config path.
	t.Setenv("LAZYARCHON_DB_PATH", "/tmp/from-env.db")

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got := loaded.LocalDBPath(); got != "/tmp/from-env.db" {
		t.Errorf("LocalDBPath() with env override = %q, want /tmp/from-env.db", got)
	}
}

func TestLoadLocalPluginConfigFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	content := "plugin: \"local\"\nplugins:\n  local:\n    path: /tmp/explicit.db\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	config, err := LoadFromPath(path)
	if err != nil {
		t.Fatalf("LoadFromPath() error = %v", err)
	}

	if got := config.GetPlugin(); got != "local" {
		t.Errorf("plugin = %q, want local", got)
	}

	if got := config.LocalDBPath(); got != "/tmp/explicit.db" {
		t.Errorf("LocalDBPath() = %q, want /tmp/explicit.db", got)
	}

	// The local plugin has no HTTP settings to leak through.
	pluginConfig := config.GetPluginConfig("local")
	if pluginConfig.BaseURL != "" || pluginConfig.APIKey != "" {
		t.Errorf("GetPluginConfig(local) = %+v, want zero HTTP settings", pluginConfig)
	}
}
