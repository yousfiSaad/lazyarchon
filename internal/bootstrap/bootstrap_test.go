package bootstrap

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/yousfisaad/lazyarchon/v2/internal/shared/config"
)

func TestBuildPluginConfigLocalUsesPath(t *testing.T) {
	cfg := &config.Config{
		Plugin: "local",
		Plugins: config.PluginsConfig{
			Local: config.LocalConfig{Path: "/tmp/lazyarchon-test.db"},
		},
	}

	built := BuildPluginConfig(cfg, "local")

	if built.BaseURL != "" || built.APIKey != "" {
		t.Errorf("local plugin config = %+v, want no HTTP settings", built)
	}

	if got, ok := built.Extra["path"].(string); !ok || got != "/tmp/lazyarchon-test.db" {
		t.Errorf("Extra[path] = %#v, want /tmp/lazyarchon-test.db", built.Extra["path"])
	}
}

func TestBuildPluginConfigHTTPFallsBackToServer(t *testing.T) {
	cfg := &config.Config{
		Plugin: "archon",
		Server: config.ServerConfig{URL: "http://localhost:8181", APIKey: "global-key-123"},
	}

	built := BuildPluginConfig(cfg, "archon")

	if built.BaseURL != "http://localhost:8181" {
		t.Errorf("BaseURL = %q, want global server URL fallback", built.BaseURL)
	}

	if built.APIKey != "global-key-123" {
		t.Errorf("APIKey = %q, want global API key fallback", built.APIKey)
	}

	if built.Timeout == 0 {
		t.Error("Timeout = 0, want a default timeout")
	}
}

func TestInitPluginLocalSeedsDatabase(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "bootstrap.db")

	cfg := &config.Config{
		Plugin: "local",
		Plugins: config.PluginsConfig{
			Local: config.LocalConfig{Path: dbPath},
		},
	}

	manager, err := InitPlugin(context.Background(), cfg)
	if err != nil {
		t.Fatalf("InitPlugin() error = %v", err)
	}
	defer manager.Close()

	client, err := manager.GetActiveClient()
	if err != nil {
		t.Fatalf("GetActiveClient() error = %v", err)
	}

	projects, err := client.ListProjects(context.Background())
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}

	if len(projects) != 1 || projects[0].Title != "Inbox" {
		t.Errorf("projects = %+v, want the seeded Inbox project", projects)
	}
}
