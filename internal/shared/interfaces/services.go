package interfaces

import (
	"time"

	"github.com/yousfisaad/lazyarchon/internal/archon"
	"github.com/yousfisaad/lazyarchon/internal/shared/config"
	"github.com/yousfisaad/lazyarchon/internal/shared/styling"
)

// ArchonClient defines the interface for Archon API operations
// This allows us to inject different implementations (basic, resilient, mock)
type ArchonClient interface {
	// Task operations
	ListTasks(projectID *string, status *string, includeClosed bool) (*archon.TasksResponse, error)
	GetTask(taskID string) (*archon.TaskResponse, error)
	UpdateTask(taskID string, updates archon.UpdateTaskRequest) (*archon.TaskResponse, error)
	DeleteTask(taskID string) error

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
	GetTheme() *config.ThemeConfig
	GetDisplay() *config.DisplayConfig
	GetDevelopment() *config.DevelopmentConfig

	// Configuration methods
	GetDefaultSortMode() string
	IsDebugEnabled() bool
	IsDarkModeEnabled() bool
	IsCompletedTasksVisible() bool
	IsPriorityIndicatorsEnabled() bool
	IsFeatureColorsEnabled() bool
	IsFeatureBackgroundsEnabled() bool
}

// StyleContextProvider defines the interface for styling context
// This allows for theme injection and testing with different style contexts
type StyleContextProvider interface {
	CreateStyleContext(forceBackground bool) *styling.StyleContext
	GetTheme() *config.ThemeConfig
}

// Logger defines the interface for logging operations
// This allows for different logging implementations and mock logging in tests
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Fatal(msg string, args ...interface{})

	// Structured logging methods for debug mode
	LogHTTPRequest(method, url string, args ...interface{})
	LogHTTPResponse(method, url string, statusCode int, duration time.Duration, args ...interface{})
	LogWebSocketEvent(event string, args ...interface{})
	LogStateChange(component, field string, oldValue, newValue interface{}, args ...interface{})
	LogPerformance(operation string, startTime time.Time, args ...interface{})
}

// Ensure that existing implementations satisfy our interfaces
// These will be validated at compile time

// Verify archon.Client implements ArchonClient
var _ ArchonClient = (*archon.Client)(nil)

// Verify archon.ResilientClient implements ArchonClient
var _ ArchonClient = (*archon.ResilientClient)(nil)

// Note: config.Config implementation verification moved to config package
// to avoid circular imports
