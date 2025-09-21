package interfaces

import (
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/yousfisaad/lazyarchon/internal/archon"
)

// ArchonClient defines the interface for Archon API operations
// This allows us to inject different implementations (basic, resilient, mock)
type ArchonClient interface {
	// Task operations
	ListTasks(projectID *string, status *string, includeClosed bool) (*archon.TasksResponse, error)
	GetTask(taskID string) (*archon.TaskResponse, error)
	UpdateTask(taskID string, updates archon.UpdateTaskRequest) (*archon.TaskResponse, error)

	// Project operations
	ListProjects() (*archon.ProjectsResponse, error)
	GetProject(projectID string) (*archon.ProjectResponse, error)

	// Health operations
	HealthCheck() error
}

// RealtimeClient defines the interface for real-time WebSocket operations
// This allows us to inject different implementations (WebSocket, mock, offline)
type RealtimeClient interface {
	// Connection management
	Connect() error
	Disconnect() error
	IsConnected() bool

	// Event handlers
	SetEventHandlers(
		onTaskUpdate func(archon.TaskUpdateEvent),
		onTaskCreate func(archon.TaskCreateEvent),
		onTaskDelete func(archon.TaskDeleteEvent),
		onProjectUpdate func(archon.ProjectUpdateEvent),
		onConnect func(),
		onDisconnect func(error),
	)

	// Event channel for Bubble Tea integration
	GetEventChannel() <-chan interface{}
}

// ConfigProvider defines the interface for configuration access
// This allows us to inject different config implementations or mock configs
type ConfigProvider interface {
	// Server configuration
	GetServerURL() string
	GetAPIKey() string

	// UI configuration
	GetTheme() interface{}    // Returns *config.ThemeConfig
	GetDisplay() interface{}  // Returns *config.DisplayConfig
	GetDevelopment() interface{} // Returns *config.DevelopmentConfig

	// Configuration methods
	GetDefaultSortMode() string
	IsDebugEnabled() bool
	IsDarkModeEnabled() bool
	IsCompletedTasksVisible() bool
	IsPriorityIndicatorsEnabled() bool
	IsFeatureColorsEnabled() bool
	IsFeatureBackgroundsEnabled() bool
}

// ViewportFactory defines the interface for creating viewport components
// This allows us to inject custom viewport implementations or mock viewports for testing
type ViewportFactory interface {
	CreateTaskDetailsViewport(width, height int) viewport.Model
	CreateHelpModalViewport(width, height int) viewport.Model
}

// UIModel defines the interface for the main UI model
// This allows for better testing and potential alternative UI implementations
type UIModel interface {
	// Core model interface - defined by Bubble Tea
	Init() interface{} // Returns tea.Cmd but defined as interface{} for flexibility
	Update(msg interface{}) (interface{}, interface{}) // Returns (tea.Model, tea.Cmd) but flexible
	View() string

	// Additional LazyArchon-specific methods for testing
	GetTasks() []archon.Task
	GetProjects() []archon.Project
	GetSelectedProjectID() *string
	IsLoading() bool
	GetError() string
}

// StyleContextProvider defines the interface for styling context
// This allows for theme injection and testing with different style contexts
type StyleContextProvider interface {
	CreateStyleContext(forceBackground bool) interface{} // Returns styling.StyleContext
	GetTheme() interface{} // Returns *config.ThemeConfig
}

// CommandExecutor defines the interface for executing async commands
// This allows for mocking command execution in tests
type CommandExecutor interface {
	LoadTasks(client ArchonClient, projectID *string) interface{} // Returns tea.Cmd
	LoadProjects(client ArchonClient) interface{}                 // Returns tea.Cmd
	UpdateTaskStatus(client ArchonClient, taskID string, newStatus string) interface{} // Returns tea.Cmd
	RefreshData(client ArchonClient, selectedProjectID *string) interface{} // Returns tea.Cmd
}

// Logger defines the interface for logging operations
// This allows for different logging implementations and mock logging in tests
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Fatal(msg string, args ...interface{})
}

// HealthChecker defines the interface for health checking operations
// This allows for custom health check implementations and mocking
type HealthChecker interface {
	CheckAPIConnection(client ArchonClient) error
	CheckConfiguration(config ConfigProvider) error
	GetHealthStatus() (bool, []string)
}

// Ensure that existing implementations satisfy our interfaces
// These will be validated at compile time

// Verify archon.Client implements ArchonClient
var _ ArchonClient = (*archon.Client)(nil)

// Verify archon.ResilientClient implements ArchonClient
var _ ArchonClient = (*archon.ResilientClient)(nil)

// Note: config.Config implementation verification moved to config package
// to avoid circular imports