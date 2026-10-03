package interfaces

import (
	"time"

	"github.com/yousfisaad/lazyarchon/v2/internal/plugin"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/config"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/styling"
)

// TaskClient defines the interface for task backend operations
// This is a type alias for plugin.TaskClient for use in the UI layer
//
//nolint:interfacebloat // This interface intentionally includes all API operations for complete client abstraction
type TaskClient = plugin.TaskClient

// Task represents a generic task
// This is a type alias for plugin.Task for use in the UI layer
type Task = plugin.Task

// Project represents a generic project
// This is a type alias for plugin.Project for use in the UI layer
type Project = plugin.Project

// TaskFilters defines filters for listing tasks
// This is a type alias for plugin.TaskFilters for use in the UI layer
type TaskFilters = plugin.TaskFilters

// CreateTaskRequest is used to create a new task
// This is a type alias for plugin.CreateTaskRequest for use in the UI layer
type CreateTaskRequest = plugin.CreateTaskRequest

// UpdateTaskRequest is used to update an existing task
// This is a type alias for plugin.UpdateTaskRequest for use in the UI layer
type UpdateTaskRequest = plugin.UpdateTaskRequest

// TaskListResult holds the results of a ListTasks call
// This is a type alias for plugin.TaskListResult for use in the UI layer
type TaskListResult = plugin.TaskListResult

// CreateProjectRequest is used to create a new project
// This is a type alias for plugin.CreateProjectRequest for use in the UI layer
type CreateProjectRequest = plugin.CreateProjectRequest

// UpdateProjectRequest is used to update an existing project
// This is a type alias for plugin.UpdateProjectRequest for use in the UI layer
type UpdateProjectRequest = plugin.UpdateProjectRequest

// Status constants
const (
	StatusTodo   = plugin.StatusTodo
	StatusDoing  = plugin.StatusDoing
	StatusReview = plugin.StatusReview
	StatusDone   = plugin.StatusDone
)

// GetStatusColor returns a color code for a status
var GetStatusColor = plugin.GetStatusColor

// GetStatusSymbol returns a symbol for a status
var GetStatusSymbol = plugin.GetStatusSymbol

// ConfigProvider defines the interface for configuration access
// This allows us to inject different config implementations or mock configs
//
//nolint:interfacebloat // Config interface requires 12 methods for comprehensive configuration access
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
//
//nolint:interfacebloat // Logger interface requires 9 methods for comprehensive logging (5 levels + 4 structured)
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Fatal(msg string, args ...interface{})

	// Structured logging methods for debug mode
	LogHTTPRequest(method, url string, args ...interface{})
	LogHTTPResponse(method, url string, statusCode int, duration time.Duration, args ...interface{})
	LogStateChange(component, field string, oldValue, newValue interface{}, args ...interface{})
	LogPerformance(operation string, startTime time.Time, args ...interface{})
}

// Ensure that existing implementations satisfy our interfaces
// These will be validated at compile time

// Note: config.Config implementation verification moved to config package
// to avoid circular imports
