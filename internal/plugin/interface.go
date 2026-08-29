// Package plugin provides a generic plugin system for lazytasks.
// This allows the TUI to work with multiple task backends (Archon, Jira, GitHub Issues, etc.)
// through a unified interface.
package plugin

import (
	"context"
	"time"
)

// Plugin is the main interface that all task backend plugins must implement.
// It provides metadata about the plugin and acts as a factory for the client.
type Plugin interface {
	// Name returns the unique name of the plugin (e.g., "archon", "github", "jira")
	Name() string

	// Version returns the plugin version
	Version() string

	// Description returns a human-readable description of the plugin
	Description() string

	// Capabilities returns what features this plugin supports
	Capabilities() Capabilities

	// CreateClient creates a new client instance with the provided configuration
	CreateClient(config PluginConfig) (TaskClient, error)
}

// Capabilities describes what features a plugin supports
type Capabilities struct {
	// SupportsProjects indicates if the backend has a concept of projects
	SupportsProjects bool

	// SupportsStatuses indicates if tasks have configurable statuses
	SupportsStatuses bool

	// SupportsPriority indicates if tasks have priority/order
	SupportsPriority bool

	// SupportsAssignees indicates if tasks can be assigned to users
	SupportsAssignees bool

	// SupportsDueDates indicates if tasks can have due dates
	SupportsDueDates bool

	// SupportsSubtasks indicates if tasks can have parent/child relationships
	SupportsSubtasks bool

	// SupportsTags indicates if tasks can have tags/labels
	SupportsTags bool

	// SupportsDescription indicates if tasks have descriptions
	SupportsDescription bool

	// SupportsArchiving indicates if tasks can be archived/hidden
	SupportsArchiving bool

	// SupportsSearch indicates if the backend supports searching/filtering tasks
	SupportsSearch bool
}

// PluginConfig holds configuration for a plugin instance
type PluginConfig struct {
	// BaseURL is the API base URL (e.g., "https://api.archon.local")
	BaseURL string

	// APIKey for authentication (if required)
	APIKey string

	// AuthToken for OAuth/token-based authentication (if required)
	AuthToken string

	// Username for basic auth (if required)
	Username string

	// Password for basic auth (if required)
	Password string

	// CustomHeaders for additional headers (e.g., custom auth)
	CustomHeaders map[string]string

	// Mappings for field/status mappings between generic and backend-specific values
	Mappings FieldMappings

	// Timeout for API requests
	Timeout time.Duration

	// RetryConfig for request retries
	RetryConfig *RetryConfig

	// Extra holds any plugin-specific configuration
	Extra map[string]interface{}
}

// FieldMappings defines how generic fields map to backend-specific fields
type FieldMappings struct {
	// Status maps generic statuses to backend statuses
	// e.g., {"todo": "open", "doing": "in_progress", "done": "closed"}
	Status map[string]string

	// Priority maps generic priority levels to backend values
	// e.g., {"high": 1, "medium": 2, "low": 3}
	Priority map[string]interface{}

	// FieldName maps generic field names to backend field names
	// e.g., {"title": "summary", "description": "body"}
	FieldName map[string]string

	// Tags maps tag names between generic and backend
	Tags map[string]string
}

// RetryConfig defines retry behavior for API requests
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int

	// InitialBackoff is the initial retry delay
	InitialBackoff time.Duration

	// MaxBackoff is the maximum retry delay
	MaxBackoff time.Duration

	// ExponentialMultiplier for backoff calculation
	ExponentialMultiplier float64
}

// TaskClient is the interface for interacting with the task backend.
// All operations accept a context for cancellation/timeouts.
//
//nolint:interfacebloat // This interface intentionally includes all API operations for complete client abstraction
type TaskClient interface {
	// Health operations
	// HealthCheck verifies the backend is accessible and authenticated
	HealthCheck(ctx context.Context) error

	// Task operations

	// ListTasks retrieves tasks matching the given filters
	ListTasks(ctx context.Context, filters TaskFilters) (*TaskListResult, error)

	// GetTask retrieves a single task by ID
	GetTask(ctx context.Context, taskID string) (*Task, error)

	// CreateTask creates a new task
	CreateTask(ctx context.Context, request CreateTaskRequest) (*Task, error)

	// UpdateTask updates an existing task
	UpdateTask(ctx context.Context, taskID string, request UpdateTaskRequest) (*Task, error)

	// DeleteTask deletes a task (or archives it if the backend doesn't support deletion)
	DeleteTask(ctx context.Context, taskID string) error

	// Project operations (optional - check Capabilities.SupportsProjects first)

	// ListProjects retrieves all available projects
	ListProjects(ctx context.Context) ([]Project, error)

	// GetProject retrieves a single project by ID
	GetProject(ctx context.Context, projectID string) (*Project, error)

	// Project write operations.
	// Backends that cannot create/update/delete projects must return an error
	// built with Unsupported() so callers can detect it via IsUnsupported.

	// CreateProject creates a new project
	CreateProject(ctx context.Context, request CreateProjectRequest) (*Project, error)

	// UpdateProject updates an existing project
	UpdateProject(ctx context.Context, projectID string, request UpdateProjectRequest) (*Project, error)

	// DeleteProject deletes a project (backends with task-referential integrity
	// may refuse until the project is empty or an explicit cascade is requested)
	DeleteProject(ctx context.Context, projectID string) error

	// Close releases any resources held by the client
	Close() error
}

// TaskFilters defines filters for listing tasks
type TaskFilters struct {
	// ProjectID filters by project (if supported)
	ProjectID *string

	// Status filters by status (use generic status names)
	Status *string

	// Assignee filters by assignee
	Assignee *string

	// IncludeArchived includes archived/closed tasks
	IncludeArchived bool

	// SearchQuery filters by text search
	SearchQuery *string

	// Tags filters by tags/labels
	Tags []string

	// DueBefore filters tasks due before a date
	DueBefore *time.Time

	// DueAfter filters tasks due after a date
	DueAfter *time.Time

	// Limit maximum number of results
	Limit int

	// Offset for pagination
	Offset int
}

// Task represents a generic task that works across all backends
type Task struct {
	// ID is the unique identifier for the task
	ID string

	// ProjectID links the task to a project (if supported)
	ProjectID string

	// ParentID for subtasks (if supported)
	ParentID *string

	// Title is the task title/summary
	Title string

	// Description is the task body/details (may contain markdown)
	Description string

	// Status is the generic status (todo, doing, review, done)
	Status string

	// Priority is a numeric priority (lower = higher priority)
	Priority int

	// Assignee is the user assigned to the task
	Assignee string

	// Tags/Labels for the task
	Tags []string

	// DueDate for the task
	DueDate *time.Time

	// CreatedAt timestamp
	CreatedAt time.Time

	// UpdatedAt timestamp
	UpdatedAt time.Time

	// Archived indicates if the task is archived/hidden
	Archived bool

	// ArchivedAt timestamp
	ArchivedAt *time.Time

	// Extra holds backend-specific fields not covered by the generic model
	Extra map[string]interface{}
}

// Project represents a generic project that works across all backends
type Project struct {
	// ID is the unique identifier
	ID string

	// Title/Name of the project
	Title string

	// Description of the project
	Description string

	// Color or theme for the project
	Color string

	// Pinned indicates if the project should be shown first
	Pinned bool

	// CreatedAt timestamp
	CreatedAt time.Time

	// UpdatedAt timestamp
	UpdatedAt time.Time

	// Extra holds backend-specific fields
	Extra map[string]interface{}
}

// CreateTaskRequest is used to create a new task
type CreateTaskRequest struct {
	// ProjectID is required if the backend supports projects
	ProjectID string

	// ParentID for creating subtasks
	ParentID *string

	// Title is required
	Title string

	// Description is optional
	Description string

	// Status defaults to "todo" if not specified
	Status string

	// Priority defaults to medium
	Priority int

	// Assignee is optional
	Assignee string

	// Tags are optional
	Tags []string

	// DueDate is optional
	DueDate *time.Time

	// Extra holds backend-specific fields to persist (e.g. "sources",
	// "code_examples" for the local backend). HTTP plugins may ignore it.
	Extra map[string]interface{}
}

// UpdateTaskRequest is used to update an existing task.
// All fields are pointers to distinguish between "not set" and "set to empty/zero".
type UpdateTaskRequest struct {
	Title       *string
	Description *string
	Status      *string
	Priority    *int
	Assignee    *string
	Tags        *[]string
	DueDate     **time.Time
	ProjectID   *string

	// ParentID reparents the task (nil pointer = not set).
	// To clear the parent, set it to a pointer to an empty string ("").
	ParentID *string

	// Archived archives or unarchives the task without deleting it
	Archived *bool

	// Extra replaces the backend-specific Extra map when set
	Extra *map[string]interface{}
}

// CreateProjectRequest is used to create a new project
type CreateProjectRequest struct {
	// Title is required
	Title string

	// Description is optional
	Description string

	// Color is optional
	Color string

	// Pinned is optional
	Pinned bool

	// Extra holds backend-specific fields to persist
	Extra map[string]interface{}
}

// UpdateProjectRequest is used to update an existing project.
// All fields are pointers to distinguish between "not set" and "set to empty/zero".
type UpdateProjectRequest struct {
	Title       *string
	Description *string
	Color       *string
	Pinned      *bool

	// Extra replaces the backend-specific Extra map when set
	Extra *map[string]interface{}
}

// TaskListResult holds the results of a ListTasks call
type TaskListResult struct {
	// Tasks is the list of tasks
	Tasks []Task

	// TotalCount is the total number of tasks matching the query
	TotalCount int

	// HasMore indicates if there are more results available
	HasMore bool

	// NextOffset for pagination
	NextOffset int
}

// Standard status values that all plugins should support
const (
	StatusTodo   = "todo"
	StatusDoing  = "doing"
	StatusReview = "review"
	StatusDone   = "done"
)

// Standard priority values
const (
	PriorityCritical = 1
	PriorityHigh     = 2
	PriorityMedium   = 3
	PriorityLow      = 4
)

// GetStatusColor returns a lipgloss-compatible color code for a status
func GetStatusColor(status string) string {
	switch status {
	case StatusTodo:
		return "240" // gray
	case StatusDoing:
		return "33" // yellow
	case StatusReview:
		return "34" // blue
	case StatusDone:
		return "32" // green
	default:
		return "37" // white
	}
}

// GetStatusSymbol returns a Unicode symbol for a status
func GetStatusSymbol(status string) string {
	switch status {
	case StatusTodo:
		return "○" // Empty circle
	case StatusDoing:
		return "◐" // Half-filled circle
	case StatusReview:
		return "◈" // Diamond
	case StatusDone:
		return "✓" // Checkmark
	default:
		return "○"
	}
}

// PluginError represents an error from a plugin operation
type PluginError struct {
	// Plugin is the name of the plugin that returned the error
	Plugin string

	// Operation is what was being attempted
	Operation string

	// Message is the error message
	Message string

	// Code is an optional error code
	Code string

	// Recoverable indicates if the operation can be retried
	Recoverable bool

	// Cause is the underlying error
	Cause error
}

func (e *PluginError) Error() string {
	if e.Cause != nil {
		return e.Plugin + ": " + e.Operation + " - " + e.Message + ": " + e.Cause.Error()
	}
	return e.Plugin + ": " + e.Operation + " - " + e.Message
}

func (e *PluginError) Unwrap() error {
	return e.Cause
}
