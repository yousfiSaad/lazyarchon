package di

import (
	"go.uber.org/dig"
	"github.com/yousfisaad/lazyarchon/internal/interfaces"
	"github.com/yousfisaad/lazyarchon/internal/archon"
	"github.com/yousfisaad/lazyarchon/internal/config"
	"github.com/charmbracelet/bubbles/viewport"
)

// TestContainer creates a DI container specifically for testing
type TestContainer struct {
	*dig.Container
}

// NewTestContainer creates a new DI container with test-specific providers
func NewTestContainer() (*TestContainer, error) {
	container := dig.New()

	// Register test providers
	if err := RegisterTestProviders(container); err != nil {
		return nil, err
	}

	return &TestContainer{Container: container}, nil
}

// RegisterTestProviders registers mock/test providers in the container
func RegisterTestProviders(container *dig.Container) error {
	// Mock providers for testing
	if err := container.Provide(NewMockConfigProvider); err != nil {
		return err
	}

	if err := container.Provide(NewMockArchonClient); err != nil {
		return err
	}

	// Use real WebSocket client for testing (no need to mock for now)
	if err := container.Provide(NewWebSocketClient); err != nil {
		return err
	}

	if err := container.Provide(NewMockViewportFactory); err != nil {
		return err
	}

	if err := container.Provide(NewMockStyleContextProvider); err != nil {
		return err
	}

	if err := container.Provide(NewMockCommandExecutor); err != nil {
		return err
	}

	if err := container.Provide(NewMockLogger); err != nil {
		return err
	}

	if err := container.Provide(NewMockHealthChecker); err != nil {
		return err
	}

	return nil
}

// Mock implementations for testing

// MockConfigProvider provides test configuration
type MockConfigProvider struct {
	serverURL      string
	apiKey         string
	debugEnabled   bool
	themeConfig    *config.ThemeConfig
	displayConfig  *config.DisplayConfig
	devConfig      *config.DevelopmentConfig
}

func NewMockConfigProvider() interfaces.ConfigProvider {
	return &MockConfigProvider{
		serverURL:    "http://test-server:8080",
		apiKey:       "test-api-key",
		debugEnabled: true,
		themeConfig: &config.ThemeConfig{
			Name: "test-theme",
		},
		displayConfig: &config.DisplayConfig{
			ShowCompletedTasks: true,
			DefaultSortMode:    "priority",
		},
		devConfig: &config.DevelopmentConfig{
			Debug: true,
		},
	}
}

func (m *MockConfigProvider) GetServerURL() string                    { return m.serverURL }
func (m *MockConfigProvider) GetAPIKey() string                      { return m.apiKey }
func (m *MockConfigProvider) GetTheme() interface{}                  { return m.themeConfig }
func (m *MockConfigProvider) GetDisplay() interface{}               { return m.displayConfig }
func (m *MockConfigProvider) GetDevelopment() interface{}           { return m.devConfig }
func (m *MockConfigProvider) GetDefaultSortMode() string            { return m.displayConfig.DefaultSortMode }
func (m *MockConfigProvider) IsDebugEnabled() bool                  { return m.debugEnabled }
func (m *MockConfigProvider) IsDarkModeEnabled() bool               { return true }
func (m *MockConfigProvider) IsCompletedTasksVisible() bool         { return m.displayConfig.ShowCompletedTasks }
func (m *MockConfigProvider) IsPriorityIndicatorsEnabled() bool     { return true }
func (m *MockConfigProvider) IsFeatureColorsEnabled() bool          { return true }
func (m *MockConfigProvider) IsFeatureBackgroundsEnabled() bool     { return false }

// MockArchonClient provides test Archon client functionality
type MockArchonClient struct {
	tasks    []archon.Task
	projects []archon.Project
	errors   map[string]error // Method -> Error mapping for controlled failures
}

func NewMockArchonClient() interfaces.ArchonClient {
	return &MockArchonClient{
		tasks: []archon.Task{
			{ID: "task-1", Title: "Test Task 1", Status: "todo"},
			{ID: "task-2", Title: "Test Task 2", Status: "doing"},
		},
		projects: []archon.Project{
			{ID: "proj-1", Title: "Test Project 1"},
			{ID: "proj-2", Title: "Test Project 2"},
		},
		errors: make(map[string]error),
	}
}

func (m *MockArchonClient) ListTasks(projectID *string, status *string, includeClosed bool) (*archon.TasksResponse, error) {
	if err, exists := m.errors["ListTasks"]; exists {
		return nil, err
	}
	return &archon.TasksResponse{Tasks: m.tasks}, nil
}

func (m *MockArchonClient) GetTask(taskID string) (*archon.TaskResponse, error) {
	if err, exists := m.errors["GetTask"]; exists {
		return nil, err
	}
	for _, task := range m.tasks {
		if task.ID == taskID {
			return &archon.TaskResponse{Task: task}, nil
		}
	}
	return nil, archon.ErrTaskNotFound
}

func (m *MockArchonClient) UpdateTask(taskID string, updates archon.UpdateTaskRequest) (*archon.TaskResponse, error) {
	if err, exists := m.errors["UpdateTask"]; exists {
		return nil, err
	}
	for i, task := range m.tasks {
		if task.ID == taskID {
			if updates.Status != nil {
				m.tasks[i].Status = *updates.Status
			}
			return &archon.TaskResponse{Task: m.tasks[i]}, nil
		}
	}
	return nil, archon.ErrTaskNotFound
}

func (m *MockArchonClient) ListProjects() (*archon.ProjectsResponse, error) {
	if err, exists := m.errors["ListProjects"]; exists {
		return nil, err
	}
	return &archon.ProjectsResponse{Projects: m.projects}, nil
}

func (m *MockArchonClient) GetProject(projectID string) (*archon.ProjectResponse, error) {
	if err, exists := m.errors["GetProject"]; exists {
		return nil, err
	}
	for _, project := range m.projects {
		if project.ID == projectID {
			return &archon.ProjectResponse{Project: project}, nil
		}
	}
	return nil, archon.ErrProjectNotFound
}

func (m *MockArchonClient) HealthCheck() error {
	if err, exists := m.errors["HealthCheck"]; exists {
		return err
	}
	return nil
}

// SetError allows tests to inject specific errors for controlled failure scenarios
func (m *MockArchonClient) SetError(method string, err error) {
	if err == nil {
		delete(m.errors, method)
	} else {
		m.errors[method] = err
	}
}

// MockViewportFactory creates mock viewports for testing
type MockViewportFactory struct{}

func NewMockViewportFactory() interfaces.ViewportFactory {
	return &MockViewportFactory{}
}

func (f *MockViewportFactory) CreateTaskDetailsViewport(width, height int) viewport.Model {
	return viewport.New(width, height)
}

func (f *MockViewportFactory) CreateHelpModalViewport(width, height int) viewport.Model {
	return viewport.New(width, height)
}

// MockStyleContextProvider provides mock styling context
type MockStyleContextProvider struct{}

func NewMockStyleContextProvider() interfaces.StyleContextProvider {
	return &MockStyleContextProvider{}
}

func (s *MockStyleContextProvider) CreateStyleContext(forceBackground bool) interface{} {
	return struct{}{}
}

func (s *MockStyleContextProvider) GetTheme() interface{} {
	return &config.ThemeConfig{Name: "test-theme"}
}

// MockCommandExecutor provides mock command execution
type MockCommandExecutor struct{}

func NewMockCommandExecutor() interfaces.CommandExecutor {
	return &MockCommandExecutor{}
}

func (c *MockCommandExecutor) LoadTasks(client interfaces.ArchonClient, projectID *string) interface{} {
	return func() interface{} { return struct{}{} }
}

func (c *MockCommandExecutor) LoadProjects(client interfaces.ArchonClient) interface{} {
	return func() interface{} { return struct{}{} }
}

func (c *MockCommandExecutor) UpdateTaskStatus(client interfaces.ArchonClient, taskID string, newStatus string) interface{} {
	return func() interface{} { return struct{}{} }
}

func (c *MockCommandExecutor) RefreshData(client interfaces.ArchonClient, selectedProjectID *string) interface{} {
	return func() interface{} { return struct{}{} }
}

// MockLogger provides test logging
type MockLogger struct {
	logs []LogEntry
}

type LogEntry struct {
	Level   string
	Message string
	Args    []interface{}
}

func NewMockLogger() interfaces.Logger {
	return &MockLogger{
		logs: make([]LogEntry, 0),
	}
}

func (l *MockLogger) Debug(msg string, args ...interface{}) {
	l.logs = append(l.logs, LogEntry{Level: "DEBUG", Message: msg, Args: args})
}

func (l *MockLogger) Info(msg string, args ...interface{}) {
	l.logs = append(l.logs, LogEntry{Level: "INFO", Message: msg, Args: args})
}

func (l *MockLogger) Warn(msg string, args ...interface{}) {
	l.logs = append(l.logs, LogEntry{Level: "WARN", Message: msg, Args: args})
}

func (l *MockLogger) Error(msg string, args ...interface{}) {
	l.logs = append(l.logs, LogEntry{Level: "ERROR", Message: msg, Args: args})
}

func (l *MockLogger) Fatal(msg string, args ...interface{}) {
	l.logs = append(l.logs, LogEntry{Level: "FATAL", Message: msg, Args: args})
}

// GetLogs returns all logged entries for test verification
func (l *MockLogger) GetLogs() []LogEntry {
	return l.logs
}

// MockHealthChecker provides mock health checking
type MockHealthChecker struct {
	healthy bool
	issues  []string
}

func NewMockHealthChecker() interfaces.HealthChecker {
	return &MockHealthChecker{
		healthy: true,
		issues:  make([]string, 0),
	}
}

func (h *MockHealthChecker) CheckAPIConnection(client interfaces.ArchonClient) error {
	return client.HealthCheck()
}

func (h *MockHealthChecker) CheckConfiguration(config interfaces.ConfigProvider) error {
	return nil
}

func (h *MockHealthChecker) GetHealthStatus() (bool, []string) {
	return h.healthy, h.issues
}

// SetHealthy allows tests to control health status
func (h *MockHealthChecker) SetHealthy(healthy bool, issues []string) {
	h.healthy = healthy
	h.issues = issues
}