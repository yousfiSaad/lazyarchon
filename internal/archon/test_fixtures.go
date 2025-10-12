package archon

import (
	"time"
)

// TaskBuilder provides a fluent interface for creating test tasks
//
//nolint:varnamelen // Short variable names are idiomatic in test code
type TaskBuilder struct {
	task Task
}

// NewTaskBuilder creates a new task builder with sensible defaults
func NewTaskBuilder() *TaskBuilder {
	return &TaskBuilder{
		task: Task{
			ID:          "test-task-1",
			ProjectID:   "test-project-1",
			Title:       "Test Task",
			Description: "A test task for unit testing",
			Status:      "todo",
			TaskOrder:   10,
			Feature:     stringPtr("testing"),
			CreatedAt:   FlexibleTime{time.Now()},
			UpdatedAt:   FlexibleTime{time.Now()},
		},
	}
}

// WithID sets the task ID
func (b *TaskBuilder) WithID(id string) *TaskBuilder {
	b.task.ID = id
	return b
}

// WithProjectID sets the project ID
func (b *TaskBuilder) WithProjectID(projectID string) *TaskBuilder {
	b.task.ProjectID = projectID
	return b
}

// WithTitle sets the task title
func (b *TaskBuilder) WithTitle(title string) *TaskBuilder {
	b.task.Title = title
	return b
}

// WithDescription sets the task description
func (b *TaskBuilder) WithDescription(description string) *TaskBuilder {
	b.task.Description = description
	return b
}

// WithStatus sets the task status
func (b *TaskBuilder) WithStatus(status string) *TaskBuilder {
	b.task.Status = status
	return b
}

// WithTaskOrder sets the task order/priority
func (b *TaskBuilder) WithTaskOrder(order int) *TaskBuilder {
	b.task.TaskOrder = order
	return b
}

// WithFeature sets the task feature
func (b *TaskBuilder) WithFeature(feature string) *TaskBuilder {
	b.task.Feature = stringPtr(feature)
	return b
}

// WithoutFeature removes the feature from the task
func (b *TaskBuilder) WithoutFeature() *TaskBuilder {
	b.task.Feature = nil
	return b
}

// WithAssignee sets the task assignee
func (b *TaskBuilder) WithAssignee(assignee string) *TaskBuilder {
	b.task.Assignee = assignee
	return b
}

// Build returns the constructed task
func (b *TaskBuilder) Build() Task {
	return b.task
}

// ProjectBuilder provides a fluent interface for creating test projects
//
//nolint:varnamelen // Short variable names are idiomatic in test code
type ProjectBuilder struct {
	project Project
}

// NewProjectBuilder creates a new project builder with sensible defaults
func NewProjectBuilder() *ProjectBuilder {
	return &ProjectBuilder{
		project: Project{
			ID:          "test-project-1",
			Title:       "Test Project",
			Description: "A test project for unit testing",
			CreatedAt:   FlexibleTime{time.Now()},
			UpdatedAt:   FlexibleTime{time.Now()},
		},
	}
}

// WithID sets the project ID
func (b *ProjectBuilder) WithID(id string) *ProjectBuilder {
	b.project.ID = id
	return b
}

// WithTitle sets the project title
func (b *ProjectBuilder) WithTitle(title string) *ProjectBuilder {
	b.project.Title = title
	return b
}

// WithDescription sets the project description
func (b *ProjectBuilder) WithDescription(description string) *ProjectBuilder {
	b.project.Description = description
	return b
}

// WithGitHubRepo sets the GitHub repository URL
func (b *ProjectBuilder) WithGitHubRepo(repo string) *ProjectBuilder {
	b.project.GitHubRepo = stringPtr(repo)
	return b
}

// Build returns the constructed project
func (b *ProjectBuilder) Build() Project {
	return b.project
}

// Predefined test fixtures

// DefaultTask returns a task with default test values
func DefaultTask() Task {
	return NewTaskBuilder().Build()
}

// TodoTask returns a task in todo status
func TodoTask() Task {
	return NewTaskBuilder().
		WithID("todo-task-1").
		WithTitle("Todo Task").
		WithStatus("todo").
		Build()
}

// DoingTask returns a task in doing status
func DoingTask() Task {
	return NewTaskBuilder().
		WithID("doing-task-1").
		WithTitle("Doing Task").
		WithStatus("doing").
		Build()
}

// ReviewTask returns a task in review status
func ReviewTask() Task {
	return NewTaskBuilder().
		WithID("review-task-1").
		WithTitle("Review Task").
		WithStatus("review").
		Build()
}

// DoneTask returns a task in done status
func DoneTask() Task {
	return NewTaskBuilder().
		WithID("done-task-1").
		WithTitle("Done Task").
		WithStatus("done").
		Build()
}

// HighPriorityTask returns a task with high priority
func HighPriorityTask() Task {
	return NewTaskBuilder().
		WithID("high-priority-task").
		WithTitle("High Priority Task").
		WithTaskOrder(100).
		Build()
}

// LowPriorityTask returns a task with low priority
func LowPriorityTask() Task {
	return NewTaskBuilder().
		WithID("low-priority-task").
		WithTitle("Low Priority Task").
		WithTaskOrder(1).
		Build()
}

// AuthenticationTask returns a task related to authentication feature
func AuthenticationTask() Task {
	return NewTaskBuilder().
		WithID("auth-task-1").
		WithTitle("Implement JWT authentication").
		WithFeature("authentication").
		WithStatus("todo").
		Build()
}

// UITask returns a task related to UI feature
func UITask() Task {
	return NewTaskBuilder().
		WithID("ui-task-1").
		WithTitle("Update task list styling").
		WithFeature("ui").
		WithStatus("doing").
		Build()
}

// TaskWithoutFeature returns a task without a feature assigned
func TaskWithoutFeature() Task {
	return NewTaskBuilder().
		WithID("no-feature-task").
		WithTitle("Task without feature").
		WithoutFeature().
		Build()
}

// DefaultProject returns a project with default test values
func DefaultProject() Project {
	return NewProjectBuilder().Build()
}

// WebProject returns a project for web development
func WebProject() Project {
	return NewProjectBuilder().
		WithID("web-project-1").
		WithTitle("Web Application").
		WithDescription("A web application project").
		WithGitHubRepo("https://github.com/test/web-app").
		Build()
}

// APIProject returns a project for API development
func APIProject() Project {
	return NewProjectBuilder().
		WithID("api-project-1").
		WithTitle("REST API").
		WithDescription("A REST API project").
		Build()
}

// Multiple item fixtures

// SampleTasks returns a diverse set of tasks for testing
func SampleTasks() []Task {
	return []Task{
		TodoTask(),
		DoingTask(),
		ReviewTask(),
		DoneTask(),
		HighPriorityTask(),
		LowPriorityTask(),
		AuthenticationTask(),
		UITask(),
		TaskWithoutFeature(),
	}
}

// SampleProjects returns a diverse set of projects for testing
func SampleProjects() []Project {
	return []Project{
		DefaultProject(),
		WebProject(),
		APIProject(),
	}
}

// Response builders for API testing

// TasksResponseFixture creates a tasks response with the given tasks
func TasksResponseFixture(tasks []Task) *TasksResponse {
	return &TasksResponse{
		Tasks: tasks,
		Count: len(tasks),
	}
}

// TaskResponseFixture creates a task response with the given task
func TaskResponseFixture(task Task) *TaskResponse {
	return &TaskResponse{
		Task: task,
	}
}

// ProjectsResponseFixture creates a projects response with the given projects
func ProjectsResponseFixture(projects []Project) *ProjectsResponse {
	return &ProjectsResponse{
		Projects: projects,
		Count:    len(projects),
	}
}

// ProjectResponseFixture creates a project response with the given project
func ProjectResponseFixture(project Project) *ProjectResponse {
	return &ProjectResponse{
		Project: project,
	}
}

// UpdateTaskRequestFixture creates an update request for testing
func UpdateTaskRequestFixture(status *string, feature *string) UpdateTaskRequest {
	return UpdateTaskRequest{
		Status:  status,
		Feature: feature,
	}
}

// Error simulation helpers (use the ones from mock_client.go)

// Test error types are defined in mock_client.go

// SetupMockServerWithData creates a mock server pre-populated with test data
func SetupMockServerWithData() *MockServer {
	server := NewMockServer()

	// Add sample data
	for _, task := range SampleTasks() {
		server.AddTask(task)
	}

	for _, project := range SampleProjects() {
		server.AddProject(project)
	}

	return server
}

// Test assertion helpers

// AssertNoError fails the test if err is not nil
//
//nolint:varnamelen // Short variable names (t) are idiomatic in test code
func AssertNoError(t interface {
	Fatalf(format string, args ...interface{})
}, err error) {
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// AssertError fails the test if err is nil
//
//nolint:varnamelen // Short variable names (t) are idiomatic in test code
func AssertError(t interface{ Fatal(args ...interface{}) }, err error) {
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}
}

// AssertErrorContains fails the test if err is nil or doesn't contain the expected text
//
//nolint:varnamelen // Short variable names (t) are idiomatic in test code
func AssertErrorContains(t interface {
	Errorf(format string, args ...interface{})
}, err error, expected string) {
	if err == nil {
		t.Errorf("Expected an error containing '%s', got nil", expected)
		return
	}
	if !contains(err.Error(), expected) {
		t.Errorf("Expected error to contain '%s', got: %v", expected, err)
	}
}

// AssertTaskEqual compares two tasks for equality
//
//nolint:varnamelen // Short variable names (t) are idiomatic in test code
func AssertTaskEqual(t interface {
	Errorf(format string, args ...interface{})
}, expected, actual Task) {
	if expected.ID != actual.ID {
		t.Errorf("Task ID: expected %s, got %s", expected.ID, actual.ID)
	}
	if expected.Title != actual.Title {
		t.Errorf("Task Title: expected %s, got %s", expected.Title, actual.Title)
	}
	if expected.Status != actual.Status {
		t.Errorf("Task Status: expected %s, got %s", expected.Status, actual.Status)
	}
	if expected.ProjectID != actual.ProjectID {
		t.Errorf("Task ProjectID: expected %s, got %s", expected.ProjectID, actual.ProjectID)
	}

	// Compare feature pointers
	if (expected.Feature == nil) != (actual.Feature == nil) {
		t.Errorf("Task Feature presence mismatch: expected %v, got %v", expected.Feature, actual.Feature)
	} else if expected.Feature != nil && actual.Feature != nil && *expected.Feature != *actual.Feature {
		t.Errorf("Task Feature: expected %s, got %s", *expected.Feature, *actual.Feature)
	}
}

// AssertProjectEqual compares two projects for equality
//
//nolint:varnamelen // Short variable names (t) are idiomatic in test code
func AssertProjectEqual(t interface {
	Errorf(format string, args ...interface{})
}, expected, actual Project) {
	if expected.ID != actual.ID {
		t.Errorf("Project ID: expected %s, got %s", expected.ID, actual.ID)
	}
	if expected.Title != actual.Title {
		t.Errorf("Project Title: expected %s, got %s", expected.Title, actual.Title)
	}
	if expected.Description != actual.Description {
		t.Errorf("Project Description: expected %s, got %s", expected.Description, actual.Description)
	}
}

// contains checks if a string contains a substring
//
//nolint:varnamelen // Short variable names are acceptable for simple utility functions
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Helper function
//
//nolint:varnamelen // Short variable names are acceptable for simple utility functions
func stringPtr(s string) *string {
	return &s
}
