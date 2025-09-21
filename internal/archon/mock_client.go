package archon

import (
	"fmt"
	"sync"
)

// MockClient provides a test implementation of ClientInterface
// It allows recording method calls and setting up predefined responses
type MockClient struct {
	mu sync.RWMutex

	// Method call recording
	ListTasksCalls    []ListTasksCall
	GetTaskCalls      []GetTaskCall
	UpdateTaskCalls   []UpdateTaskCall
	ListProjectsCalls []ListProjectsCall
	GetProjectCalls   []GetProjectCall
	HealthCheckCalls  []HealthCheckCall

	// Response configuration
	ListTasksResponse    *TasksResponse
	ListTasksError       error
	GetTaskResponse      *TaskResponse
	GetTaskError         error
	UpdateTaskResponse   *TaskResponse
	UpdateTaskError      error
	ListProjectsResponse *ProjectsResponse
	ListProjectsError    error
	GetProjectResponse   *ProjectResponse
	GetProjectError      error
	HealthCheckError     error

	// Behavior configuration
	CallDelay map[string]int // Simulate network delays in milliseconds
}

// Call recording structures
type ListTasksCall struct {
	ProjectID     *string
	Status        *string
	IncludeClosed bool
}

type GetTaskCall struct {
	TaskID string
}

type UpdateTaskCall struct {
	TaskID  string
	Updates UpdateTaskRequest
}

type ListProjectsCall struct{}

type GetProjectCall struct {
	ProjectID string
}

type HealthCheckCall struct{}

// NewMockClient creates a new mock client with default successful responses
func NewMockClient() *MockClient {
	return &MockClient{
		ListTasksResponse: &TasksResponse{
			Tasks: []Task{},
			Count: 0,
		},
		GetTaskResponse: &TaskResponse{
			Task: Task{},
		},
		UpdateTaskResponse: &TaskResponse{
			Task: Task{},
		},
		ListProjectsResponse: &ProjectsResponse{
			Projects: []Project{},
			Count:    0,
		},
		GetProjectResponse: &ProjectResponse{
			Project: Project{},
		},
		CallDelay: make(map[string]int),
	}
}

// ListTasks mock implementation
func (m *MockClient) ListTasks(projectID *string, status *string, includeClosed bool) (*TasksResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Record the call
	m.ListTasksCalls = append(m.ListTasksCalls, ListTasksCall{
		ProjectID:     projectID,
		Status:        status,
		IncludeClosed: includeClosed,
	})

	// Return configured response/error
	if m.ListTasksError != nil {
		return nil, m.ListTasksError
	}
	return m.ListTasksResponse, nil
}

// GetTask mock implementation
func (m *MockClient) GetTask(taskID string) (*TaskResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Record the call
	m.GetTaskCalls = append(m.GetTaskCalls, GetTaskCall{
		TaskID: taskID,
	})

	// Return configured response/error
	if m.GetTaskError != nil {
		return nil, m.GetTaskError
	}
	return m.GetTaskResponse, nil
}

// UpdateTask mock implementation
func (m *MockClient) UpdateTask(taskID string, updates UpdateTaskRequest) (*TaskResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Record the call
	m.UpdateTaskCalls = append(m.UpdateTaskCalls, UpdateTaskCall{
		TaskID:  taskID,
		Updates: updates,
	})

	// Return configured response/error
	if m.UpdateTaskError != nil {
		return nil, m.UpdateTaskError
	}
	return m.UpdateTaskResponse, nil
}

// ListProjects mock implementation
func (m *MockClient) ListProjects() (*ProjectsResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Record the call
	m.ListProjectsCalls = append(m.ListProjectsCalls, ListProjectsCall{})

	// Return configured response/error
	if m.ListProjectsError != nil {
		return nil, m.ListProjectsError
	}
	return m.ListProjectsResponse, nil
}

// GetProject mock implementation
func (m *MockClient) GetProject(projectID string) (*ProjectResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Record the call
	m.GetProjectCalls = append(m.GetProjectCalls, GetProjectCall{
		ProjectID: projectID,
	})

	// Return configured response/error
	if m.GetProjectError != nil {
		return nil, m.GetProjectError
	}
	return m.GetProjectResponse, nil
}

// HealthCheck mock implementation
func (m *MockClient) HealthCheck() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Record the call
	m.HealthCheckCalls = append(m.HealthCheckCalls, HealthCheckCall{})

	// Return configured error
	return m.HealthCheckError
}

// Helper methods for test setup

// SetListTasksResponse configures the response for ListTasks calls
func (m *MockClient) SetListTasksResponse(response *TasksResponse, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ListTasksResponse = response
	m.ListTasksError = err
}

// SetGetTaskResponse configures the response for GetTask calls
func (m *MockClient) SetGetTaskResponse(response *TaskResponse, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.GetTaskResponse = response
	m.GetTaskError = err
}

// SetUpdateTaskResponse configures the response for UpdateTask calls
func (m *MockClient) SetUpdateTaskResponse(response *TaskResponse, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.UpdateTaskResponse = response
	m.UpdateTaskError = err
}

// SetListProjectsResponse configures the response for ListProjects calls
func (m *MockClient) SetListProjectsResponse(response *ProjectsResponse, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ListProjectsResponse = response
	m.ListProjectsError = err
}

// SetGetProjectResponse configures the response for GetProject calls
func (m *MockClient) SetGetProjectResponse(response *ProjectResponse, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.GetProjectResponse = response
	m.GetProjectError = err
}

// SetHealthCheckError configures the error for HealthCheck calls
func (m *MockClient) SetHealthCheckError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.HealthCheckError = err
}

// Call count helpers

// GetListTasksCallCount returns the number of ListTasks calls made
func (m *MockClient) GetListTasksCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.ListTasksCalls)
}

// GetGetTaskCallCount returns the number of GetTask calls made
func (m *MockClient) GetGetTaskCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.GetTaskCalls)
}

// GetUpdateTaskCallCount returns the number of UpdateTask calls made
func (m *MockClient) GetUpdateTaskCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.UpdateTaskCalls)
}

// GetListProjectsCallCount returns the number of ListProjects calls made
func (m *MockClient) GetListProjectsCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.ListProjectsCalls)
}

// GetGetProjectCallCount returns the number of GetProject calls made
func (m *MockClient) GetGetProjectCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.GetProjectCalls)
}

// GetHealthCheckCallCount returns the number of HealthCheck calls made
func (m *MockClient) GetHealthCheckCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.HealthCheckCalls)
}

// Reset clears all recorded calls and resets responses to defaults
func (m *MockClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear all call recordings
	m.ListTasksCalls = nil
	m.GetTaskCalls = nil
	m.UpdateTaskCalls = nil
	m.ListProjectsCalls = nil
	m.GetProjectCalls = nil
	m.HealthCheckCalls = nil

	// Reset responses to defaults
	m.ListTasksResponse = &TasksResponse{Tasks: []Task{}, Count: 0}
	m.ListTasksError = nil
	m.GetTaskResponse = &TaskResponse{Task: Task{}}
	m.GetTaskError = nil
	m.UpdateTaskResponse = &TaskResponse{Task: Task{}}
	m.UpdateTaskError = nil
	m.ListProjectsResponse = &ProjectsResponse{Projects: []Project{}, Count: 0}
	m.ListProjectsError = nil
	m.GetProjectResponse = &ProjectResponse{Project: Project{}}
	m.GetProjectError = nil
	m.HealthCheckError = nil
}

// SimulateNetworkError returns a common network error for testing
func SimulateNetworkError() error {
	return fmt.Errorf("network error: connection refused")
}

// SimulateTimeoutError returns a timeout error for testing
func SimulateTimeoutError() error {
	return fmt.Errorf("request timeout: deadline exceeded")
}

// SimulateAPIError returns an API error for testing
func SimulateAPIError(statusCode int, message string) error {
	return fmt.Errorf("API error (status %d): %s", statusCode, message)
}