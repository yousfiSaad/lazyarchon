package plugin

import (
	"context"
	"fmt"
	"sync"
)

// Manager handles the lifecycle of plugins and their clients.
// It provides a high-level interface for working with plugins.
type Manager struct {
	mu           sync.RWMutex
	registry     *Registry
	activePlugin Plugin
	activeClient TaskClient
	activeConfig *PluginConfig
}

// NewManager creates a new plugin manager with the given registry
func NewManager(registry *Registry) *Manager {
	return &Manager{
		registry: registry,
	}
}

// NewManagerWithGlobalRegistry creates a new plugin manager using the global registry
func NewManagerWithGlobalRegistry() *Manager {
	return NewManager(globalRegistry)
}

// LoadPlugin loads and initializes a plugin with the given configuration.
// This replaces any currently active plugin.
func (m *Manager) LoadPlugin(name string, config PluginConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Close existing client if any
	if m.activeClient != nil {
		if err := m.activeClient.Close(); err != nil {
			// Log error but continue
			_ = err
		}
	}

	// Get plugin from registry
	plugin, err := m.registry.Get(name)
	if err != nil {
		return fmt.Errorf("failed to load plugin '%s': %w", name, err)
	}

	// Create client
	client, err := plugin.CreateClient(config)
	if err != nil {
		return fmt.Errorf("failed to create client for plugin '%s': %w", name, err)
	}

	// Verify connection with health check
	ctx := context.Background()
	if err := client.HealthCheck(ctx); err != nil {
		client.Close()
		return fmt.Errorf("health check failed for plugin '%s': %w", name, err)
	}

	m.activePlugin = plugin
	m.activeClient = client
	m.activeConfig = &config

	return nil
}

// Unload unloads the current plugin and closes its client
func (m *Manager) Unload() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.activeClient != nil {
		if err := m.activeClient.Close(); err != nil {
			return err
		}
		m.activeClient = nil
	}

	m.activePlugin = nil
	m.activeConfig = nil

	return nil
}

// GetActivePlugin returns the currently loaded plugin
func (m *Manager) GetActivePlugin() (Plugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.activePlugin == nil {
		return nil, fmt.Errorf("no plugin is currently loaded")
	}

	return m.activePlugin, nil
}

// GetActiveClient returns the currently active client
func (m *Manager) GetActiveClient() (TaskClient, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.activeClient == nil {
		return nil, fmt.Errorf("no plugin is currently loaded")
	}

	return m.activeClient, nil
}

// IsLoaded returns true if a plugin is currently loaded
func (m *Manager) IsLoaded() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeClient != nil
}

// GetActivePluginName returns the name of the currently loaded plugin
func (m *Manager) GetActivePluginName() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.activePlugin == nil {
		return ""
	}

	return m.activePlugin.Name()
}

// ListAvailablePlugins returns all available plugin names from the registry
func (m *Manager) ListAvailablePlugins() []string {
	return m.registry.List()
}

// GetPluginInfo returns information about a specific plugin
func (m *Manager) GetPluginInfo(name string) (*PluginInfo, error) {
	plugin, err := m.registry.Get(name)
	if err != nil {
		return nil, err
	}

	return &PluginInfo{
		Name:         plugin.Name(),
		Version:      plugin.Version(),
		Description:  plugin.Description(),
		Capabilities: plugin.Capabilities(),
	}, nil
}

// PluginInfo holds metadata about a plugin
type PluginInfo struct {
	Name         string
	Version      string
	Description  string
	Capabilities Capabilities
}

// Close unloads the current plugin and cleans up resources
func (m *Manager) Close() error {
	return m.Unload()
}

// WithClient executes a function with the active client, handling locking
func (m *Manager) WithClient(fn func(TaskClient) error) error {
	client, err := m.GetActiveClient()
	if err != nil {
		return err
	}

	return fn(client)
}

// Helper methods that delegate to the active client

// ListTasks retrieves tasks using the active plugin
func (m *Manager) ListTasks(ctx context.Context, filters TaskFilters) (*TaskListResult, error) {
	client, err := m.GetActiveClient()
	if err != nil {
		return nil, err
	}

	return client.ListTasks(ctx, filters)
}

// GetTask retrieves a task using the active plugin
func (m *Manager) GetTask(ctx context.Context, taskID string) (*Task, error) {
	client, err := m.GetActiveClient()
	if err != nil {
		return nil, err
	}

	return client.GetTask(ctx, taskID)
}

// CreateTask creates a task using the active plugin
func (m *Manager) CreateTask(ctx context.Context, request CreateTaskRequest) (*Task, error) {
	client, err := m.GetActiveClient()
	if err != nil {
		return nil, err
	}

	return client.CreateTask(ctx, request)
}

// UpdateTask updates a task using the active plugin
func (m *Manager) UpdateTask(ctx context.Context, taskID string, request UpdateTaskRequest) (*Task, error) {
	client, err := m.GetActiveClient()
	if err != nil {
		return nil, err
	}

	return client.UpdateTask(ctx, taskID, request)
}

// DeleteTask deletes a task using the active plugin
func (m *Manager) DeleteTask(ctx context.Context, taskID string) error {
	client, err := m.GetActiveClient()
	if err != nil {
		return err
	}

	return client.DeleteTask(ctx, taskID)
}

// ListProjects retrieves projects using the active plugin
func (m *Manager) ListProjects(ctx context.Context) ([]Project, error) {
	client, err := m.GetActiveClient()
	if err != nil {
		return nil, err
	}

	return client.ListProjects(ctx)
}

// GetProject retrieves a project using the active plugin
func (m *Manager) GetProject(ctx context.Context, projectID string) (*Project, error) {
	client, err := m.GetActiveClient()
	if err != nil {
		return nil, err
	}

	return client.GetProject(ctx, projectID)
}

// CreateProject creates a project using the active plugin
func (m *Manager) CreateProject(ctx context.Context, request CreateProjectRequest) (*Project, error) {
	client, err := m.GetActiveClient()
	if err != nil {
		return nil, err
	}

	return client.CreateProject(ctx, request)
}

// UpdateProject updates a project using the active plugin
func (m *Manager) UpdateProject(ctx context.Context, projectID string, request UpdateProjectRequest) (*Project, error) {
	client, err := m.GetActiveClient()
	if err != nil {
		return nil, err
	}

	return client.UpdateProject(ctx, projectID, request)
}

// DeleteProject deletes a project using the active plugin
func (m *Manager) DeleteProject(ctx context.Context, projectID string) error {
	client, err := m.GetActiveClient()
	if err != nil {
		return err
	}

	return client.DeleteProject(ctx, projectID)
}

// HealthCheck verifies the active plugin connection
func (m *Manager) HealthCheck(ctx context.Context) error {
	client, err := m.GetActiveClient()
	if err != nil {
		return err
	}

	return client.HealthCheck(ctx)
}
