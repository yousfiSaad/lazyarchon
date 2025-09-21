package di

import (
	"go.uber.org/dig"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/internal/interfaces"
)

// Container wraps dig.Container with LazyArchon-specific functionality
type Container struct {
	*dig.Container
}

// NewContainer creates a new DI container with all providers registered
func NewContainer() (*Container, error) {
	container := dig.New()

	// Register all providers
	if err := RegisterProviders(container); err != nil {
		return nil, err
	}

	return &Container{Container: container}, nil
}

// RegisterProviders registers all service providers in the container
func RegisterProviders(container *dig.Container) error {
	// Core infrastructure providers
	if err := container.Provide(NewConfigProvider); err != nil {
		return err
	}

	if err := container.Provide(NewArchonClient); err != nil {
		return err
	}

	if err := container.Provide(NewWebSocketClient); err != nil {
		return err
	}

	if err := container.Provide(NewViewportFactory); err != nil {
		return err
	}

	if err := container.Provide(NewStyleContextProvider); err != nil {
		return err
	}

	if err := container.Provide(NewCommandExecutor); err != nil {
		return err
	}

	if err := container.Provide(NewLogger); err != nil {
		return err
	}

	if err := container.Provide(NewHealthChecker); err != nil {
		return err
	}

	// UI model provider (depends on other services) - returns tea.Model
	if err := container.Provide(NewUIModel); err != nil {
		return err
	}

	return nil
}

// GetUIModel retrieves the UI model from the container
func (c *Container) GetUIModel() (tea.Model, error) {
	var model tea.Model
	err := c.Invoke(func(m tea.Model) {
		model = m
	})
	return model, err
}

// GetArchonClient retrieves the Archon client from the container
func (c *Container) GetArchonClient() (interfaces.ArchonClient, error) {
	var client interfaces.ArchonClient
	err := c.Invoke(func(c interfaces.ArchonClient) {
		client = c
	})
	return client, err
}

// GetConfigProvider retrieves the config provider from the container
func (c *Container) GetConfigProvider() (interfaces.ConfigProvider, error) {
	var config interfaces.ConfigProvider
	err := c.Invoke(func(c interfaces.ConfigProvider) {
		config = c
	})
	return config, err
}