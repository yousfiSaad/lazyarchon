package ui

import (
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/yousfisaad/lazyarchon/internal/archon"
	configpkg "github.com/yousfisaad/lazyarchon/internal/config"
	"github.com/yousfisaad/lazyarchon/internal/interfaces"
	"github.com/yousfisaad/lazyarchon/internal/ui/sorting"
	"github.com/yousfisaad/lazyarchon/internal/ui/styling"
)

// NewModel creates a new application model
func NewModel(cfg *configpkg.Config) Model {
	// Initialize theme from configuration
	styling.InitializeTheme(cfg)

	// Connect to Archon server using configuration
	client := archon.NewClient(cfg.GetServerURL(), cfg.GetAPIKey())

	// Create WebSocket client for real-time updates
	wsClient := archon.NewWebSocketClient(cfg.GetServerURL(), cfg.GetAPIKey())

	// Create viewport for task details with reasonable defaults
	// Will be resized when window size is available
	vp := viewport.New(80, 20)
	vp.SetContent("") // Empty content initially

	// Create viewport for help modal with reasonable defaults
	// Will be resized when modal is opened
	helpVp := viewport.New(60, 15)
	helpVp.SetContent("") // Empty content initially

	return Model{
		client:   client,
		wsClient: wsClient,
		config:   cfg,
		Window: WindowState{
			activeView: LeftPanel, // Default to task list panel active
		},
		Navigation: NavigationState{
			selectedIndex: 0,
		},
		Data: DataState{
			loading:        true,
			loadingMessage: "Connecting to Archon server...",
			sortMode:       parseSortModeFromConfig(cfg.GetDefaultSortMode()),
		},
		Modals: ModalState{
			featureMode: FeatureModeState{
				selectedFeatures: make(map[string]bool), // Initialize empty feature selection
			},
		},
		taskDetailsViewport: vp,
		helpModalViewport:   helpVp,
	}
}

// NewModelWithDependencies creates a new application model with injected dependencies
func NewModelWithDependencies(
	client interfaces.ArchonClient,
	wsClient interfaces.RealtimeClient,
	config interfaces.ConfigProvider,
	viewportFactory interfaces.ViewportFactory,
	styleContextProvider interfaces.StyleContextProvider,
	commandExecutor interfaces.CommandExecutor,
	logger interfaces.Logger,
	healthChecker interfaces.HealthChecker,
) Model {
	// Convert interfaces to concrete types for the Model struct
	// This allows gradual migration to interfaces
	var concreteClient *archon.Client
	var concreteConfig *configpkg.Config

	// Safe type assertion with fallback
	if c, ok := client.(*archon.ResilientClient); ok {
		// Extract the underlying client from resilient wrapper
		concreteClient = c.BaseClient()
	} else if c, ok := client.(*archon.Client); ok {
		concreteClient = c
	} else {
		// Fallback: create new client from config
		concreteClient = archon.NewClient(config.GetServerURL(), config.GetAPIKey())
	}

	if c, ok := config.(*configpkg.Config); ok {
		concreteConfig = c
	} else {
		// Fallback: create default config and populate from interface
		concreteConfig = &configpkg.Config{}
		// We can't easily populate all fields, so this is a minimal fallback
		// In practice, the DI container should always provide a concrete *configpkg.Config
	}

	// Initialize theme from configuration using the concrete config
	styling.InitializeTheme(concreteConfig)

	// Create viewports using factory
	taskDetailsViewport := viewportFactory.CreateTaskDetailsViewport(80, 20)
	helpModalViewport := viewportFactory.CreateHelpModalViewport(60, 15)

	logger.Debug("Creating UI model with injected dependencies")

	// Create dependencies struct
	deps := &ModelDependencies{
		ArchonClient:        client,
		ConfigProvider:      config,
		ViewportFactory:     viewportFactory,
		StyleContextProvider: styleContextProvider,
		CommandExecutor:     commandExecutor,
		Logger:              logger,
		HealthChecker:       healthChecker,
	}

	return Model{
		client:   concreteClient,
		wsClient: wsClient,
		config:   concreteConfig,
		deps:     deps,
		Window: WindowState{
			activeView: LeftPanel, // Default to task list panel active
		},
		Navigation: NavigationState{
			selectedIndex: 0,
		},
		Data: DataState{
			loading:        true,
			loadingMessage: "Connecting to Archon server...",
			sortMode:       parseSortModeFromConfig(config.GetDefaultSortMode()),
		},
		Modals: ModalState{
			featureMode: FeatureModeState{
				selectedFeatures: make(map[string]bool), // Initialize empty feature selection
			},
		},
		taskDetailsViewport: taskDetailsViewport,
		helpModalViewport:   helpModalViewport,
	}
}

// parseSortModeFromConfig converts string sort mode to int constant
func parseSortModeFromConfig(sortModeStr string) int {
	switch sortModeStr {
	case "status+priority":
		return sorting.SortStatusPriority
	case "priority":
		return sorting.SortPriorityOnly
	case "time":
		return sorting.SortTimeCreated
	case "alphabetical":
		return sorting.SortAlphabetical
	default:
		return sorting.SortStatusPriority // Default fallback
	}
}