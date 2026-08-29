package plugin

import (
	"fmt"
	"sync"
)

// Registry manages the collection of available plugins.
// It provides a central place to register and retrieve plugins.
type Registry struct {
	mu      sync.RWMutex
	plugins map[string]Plugin
}

// NewRegistry creates a new empty plugin registry
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]Plugin),
	}
}

// Register adds a plugin to the registry.
// Returns an error if a plugin with the same name is already registered.
func (r *Registry) Register(plugin Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := plugin.Name()
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin '%s' is already registered", name)
	}

	r.plugins[name] = plugin
	return nil
}

// Unregister removes a plugin from the registry
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.plugins, name)
}

// Get retrieves a plugin by name
func (r *Registry) Get(name string) (Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, exists := r.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin '%s' not found", name)
	}

	return plugin, nil
}

// Has checks if a plugin is registered
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.plugins[name]
	return exists
}

// List returns all registered plugin names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.plugins))
	for name := range r.plugins {
		names = append(names, name)
	}
	return names
}

// GetAll returns all registered plugins
func (r *Registry) GetAll() []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := make([]Plugin, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		plugins = append(plugins, plugin)
	}
	return plugins
}

// Count returns the number of registered plugins
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.plugins)
}

// globalRegistry is the default global registry
var globalRegistry = NewRegistry()

// Register adds a plugin to the global registry
func Register(plugin Plugin) error {
	return globalRegistry.Register(plugin)
}

// Unregister removes a plugin from the global registry
func Unregister(name string) {
	globalRegistry.Unregister(name)
}

// Get retrieves a plugin from the global registry
func Get(name string) (Plugin, error) {
	return globalRegistry.Get(name)
}

// Has checks if a plugin exists in the global registry
func Has(name string) bool {
	return globalRegistry.Has(name)
}

// List returns all plugin names from the global registry
func List() []string {
	return globalRegistry.List()
}

// GetAll returns all plugins from the global registry
func GetAll() []Plugin {
	return globalRegistry.GetAll()
}

// PluginFactory creates plugins on demand
type PluginFactory func() (Plugin, error)

// FactoryRegistry is a registry that creates plugins lazily
type FactoryRegistry struct {
	mu        sync.RWMutex
	factories map[string]PluginFactory
	plugins   map[string]Plugin
}

// NewFactoryRegistry creates a new factory registry
func NewFactoryRegistry() *FactoryRegistry {
	return &FactoryRegistry{
		factories: make(map[string]PluginFactory),
		plugins:   make(map[string]Plugin),
	}
}

// RegisterFactory adds a plugin factory to the registry
func (r *FactoryRegistry) RegisterFactory(name string, factory PluginFactory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("plugin factory '%s' is already registered", name)
	}

	r.factories[name] = factory
	return nil
}

// Get retrieves or creates a plugin by name
func (r *FactoryRegistry) Get(name string) (Plugin, error) {
	// First check if plugin is already created
	r.mu.RLock()
	plugin, exists := r.plugins[name]
	r.mu.RUnlock()

	if exists {
		return plugin, nil
	}

	// Need to create it
	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if plugin, exists := r.plugins[name]; exists {
		return plugin, nil
	}

	factory, exists := r.factories[name]
	if !exists {
		return nil, fmt.Errorf("plugin factory '%s' not found", name)
	}

	plugin, err := factory()
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin '%s': %w", name, err)
	}

	r.plugins[name] = plugin
	return plugin, nil
}

// List returns all available plugin names (including factories not yet instantiated)
func (r *FactoryRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	return names
}

// globalFactoryRegistry is the default global factory registry
var globalFactoryRegistry = NewFactoryRegistry()

// RegisterFactory adds a factory to the global registry
func RegisterFactory(name string, factory PluginFactory) error {
	return globalFactoryRegistry.RegisterFactory(name, factory)
}

// GetFromFactory retrieves or creates a plugin from the global factory registry
func GetFromFactory(name string) (Plugin, error) {
	return globalFactoryRegistry.Get(name)
}
