package archon

// ClientInterface defines the contract for Archon API operations
// This interface enables dependency injection and mocking for tests
type ClientInterface interface {
	// Task operations
	ListTasks(projectID *string, status *string, includeClosed bool) (*TasksResponse, error)
	GetTask(taskID string) (*TaskResponse, error)
	UpdateTask(taskID string, updates UpdateTaskRequest) (*TaskResponse, error)

	// Project operations
	ListProjects() (*ProjectsResponse, error)
	GetProject(projectID string) (*ProjectResponse, error)

	// Health operations
	HealthCheck() error
}

// Ensure Client implements ClientInterface
var _ ClientInterface = (*Client)(nil)