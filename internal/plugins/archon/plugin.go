// Package archon provides a lazytasks plugin for the Archon task management API.
// This plugin adapts the Archon-specific client to the generic plugin interface.
package archon

import (
	"context"
	"fmt"
	"time"

	"github.com/yousfisaad/lazyarchon/v2/internal/plugin"
)

// PluginName is the unique identifier for this plugin
const PluginName = "archon"

// PluginVersion is the current version of this plugin
const PluginVersion = "2.0.0"

// ArchonPlugin implements the plugin.Plugin interface for Archon
type ArchonPlugin struct{}

// Name returns the plugin name
func (p *ArchonPlugin) Name() string {
	return PluginName
}

// Version returns the plugin version
func (p *ArchonPlugin) Version() string {
	return PluginVersion
}

// Description returns a human-readable description
func (p *ArchonPlugin) Description() string {
	return "Archon task management API plugin - supports projects, statuses, priorities, assignees, and task archiving"
}

// Capabilities returns what features this plugin supports
func (p *ArchonPlugin) Capabilities() plugin.Capabilities {
	return plugin.Capabilities{
		SupportsProjects:    true,
		SupportsStatuses:    true,
		SupportsPriority:    true,
		SupportsAssignees:   true,
		SupportsDueDates:    false, // Archon doesn't support due dates natively
		SupportsSubtasks:    true,  // Via parent_task_id
		SupportsTags:        true,  // Via feature field
		SupportsDescription: true,
		SupportsArchiving:   true,
		SupportsSearch:      false, // Limited search support
	}
}

// CreateClient creates a new Archon client from the plugin configuration
func (p *ArchonPlugin) CreateClient(config plugin.PluginConfig) (plugin.TaskClient, error) {
	// Determine timeout
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	// Create resilience config from plugin config
	resilienceConfig := DefaultResilienceConfig()
	if config.RetryConfig != nil {
		resilienceConfig.Retry.MaxAttempts = config.RetryConfig.MaxRetries
		if config.RetryConfig.InitialBackoff > 0 {
			resilienceConfig.Retry.BaseDelay = config.RetryConfig.InitialBackoff
		}
		if config.RetryConfig.MaxBackoff > 0 {
			resilienceConfig.Retry.MaxDelay = config.RetryConfig.MaxBackoff
		}
		if config.RetryConfig.ExponentialMultiplier > 0 {
			resilienceConfig.Retry.Multiplier = config.RetryConfig.ExponentialMultiplier
		}
	}

	// Create the resilient client
	resilientClient := NewResilientClient(
		config.BaseURL,
		config.APIKey,
		timeout,
		resilienceConfig,
	)

	// Note: config.CustomHeaders is currently ignored — the resilient client
	// does not support per-request headers yet.

	// Wrap with the adapter to implement TaskClient interface
	adapter := &ArchonClientAdapter{
		client:   resilientClient,
		mappings: config.Mappings,
	}

	return adapter, nil
}

// ArchonClientAdapter adapts the Archon-specific client to the generic plugin.TaskClient interface
type ArchonClientAdapter struct {
	client   *ResilientClient
	mappings plugin.FieldMappings
}

// ensureInitialized checks if status mappings are initialized
func (a *ArchonClientAdapter) ensureInitialized() {
	// Initialize default mappings if not set
	if a.mappings.Status == nil {
		a.mappings.Status = map[string]string{
			plugin.StatusTodo:   TaskStatusTodo,
			plugin.StatusDoing:  TaskStatusDoing,
			plugin.StatusReview: TaskStatusReview,
			plugin.StatusDone:   TaskStatusDone,
		}
	}
}

// toBackendStatus converts a generic status to Archon-specific status
func (a *ArchonClientAdapter) toBackendStatus(genericStatus string) string {
	a.ensureInitialized()
	if backendStatus, ok := a.mappings.Status[genericStatus]; ok {
		return backendStatus
	}
	// Return as-is if no mapping exists
	return genericStatus
}

// fromBackendStatus converts an Archon-specific status to generic status
func (a *ArchonClientAdapter) fromBackendStatus(backendStatus string) string {
	a.ensureInitialized()
	for generic, backend := range a.mappings.Status {
		if backend == backendStatus {
			return generic
		}
	}
	// Return as-is if no mapping exists
	return backendStatus
}

// HealthCheck implements plugin.TaskClient
func (a *ArchonClientAdapter) HealthCheck(ctx context.Context) error {
	return a.client.HealthCheck()
}

// ListTasks implements plugin.TaskClient
func (a *ArchonClientAdapter) ListTasks(ctx context.Context, filters plugin.TaskFilters) (*plugin.TaskListResult, error) {
	// Convert generic filters to Archon-specific parameters
	var projectID, status *string
	if filters.ProjectID != nil {
		projectID = filters.ProjectID
	}
	if filters.Status != nil {
		backendStatus := a.toBackendStatus(*filters.Status)
		status = &backendStatus
	}

	resp, err := a.client.ListTasks(projectID, status, filters.IncludeArchived)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	// Convert Archon tasks to generic tasks
	tasks := make([]plugin.Task, len(resp.Tasks))
	for i, t := range resp.Tasks {
		tasks[i] = a.convertTask(t)
	}

	return &plugin.TaskListResult{
		Tasks:      tasks,
		TotalCount: resp.TotalCount,
		HasMore:    false, // Archon fetches all pages
		NextOffset: 0,
	}, nil
}

// GetTask implements plugin.TaskClient
func (a *ArchonClientAdapter) GetTask(ctx context.Context, taskID string) (*plugin.Task, error) {
	resp, err := a.client.GetTask(taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	task := a.convertTask(resp.Task)
	return &task, nil
}

// CreateTask implements plugin.TaskClient
func (a *ArchonClientAdapter) CreateTask(ctx context.Context, request plugin.CreateTaskRequest) (*plugin.Task, error) {
	// Convert generic request to Archon-specific request
	archonReq := CreateTaskRequest{
		ProjectID:   request.ProjectID,
		Title:       request.Title,
		Description: request.Description,
		Status:      a.toBackendStatus(request.Status),
		Assignee:    request.Assignee,
		TaskOrder:   request.Priority,
	}

	// Note: request.ParentID is dropped — the Archon API has no parent-task
	// field to map it to.

	resp, err := a.client.CreateTask(archonReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	task := a.convertTask(resp.Task)
	return &task, nil
}

// UpdateTask implements plugin.TaskClient
func (a *ArchonClientAdapter) UpdateTask(ctx context.Context, taskID string, request plugin.UpdateTaskRequest) (*plugin.Task, error) {
	// Convert generic request to Archon-specific request
	archonReq := UpdateTaskRequest{}

	if request.Title != nil {
		archonReq.Title = request.Title
	}
	if request.Description != nil {
		archonReq.Description = request.Description
	}
	if request.Status != nil {
		backendStatus := a.toBackendStatus(*request.Status)
		archonReq.Status = &backendStatus
	}
	if request.Assignee != nil {
		archonReq.Assignee = request.Assignee
	}
	if request.Priority != nil {
		archonReq.TaskOrder = request.Priority
	}
	if request.Tags != nil {
		// Archon uses "feature" field for tags
		// For simplicity, we'll use the first tag as the feature
		if len(*request.Tags) > 0 {
			feature := (*request.Tags)[0]
			archonReq.Feature = &feature
		}
	}

	resp, err := a.client.UpdateTask(taskID, archonReq)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	task := a.convertTask(resp.Task)
	return &task, nil
}

// DeleteTask implements plugin.TaskClient
func (a *ArchonClientAdapter) DeleteTask(ctx context.Context, taskID string) error {
	err := a.client.DeleteTask(taskID)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

// ListProjects implements plugin.TaskClient
func (a *ArchonClientAdapter) ListProjects(ctx context.Context) ([]plugin.Project, error) {
	resp, err := a.client.ListProjects()
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	projects := make([]plugin.Project, len(resp.Projects))
	for i, p := range resp.Projects {
		projects[i] = a.convertProject(p)
	}

	return projects, nil
}

// GetProject implements plugin.TaskClient
func (a *ArchonClientAdapter) GetProject(ctx context.Context, projectID string) (*plugin.Project, error) {
	resp, err := a.client.GetProject(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	project := a.convertProject(resp.Project)
	return &project, nil
}

// CreateProject implements plugin.TaskClient
func (a *ArchonClientAdapter) CreateProject(ctx context.Context, request plugin.CreateProjectRequest) (*plugin.Project, error) {
	return nil, plugin.Unsupported("archon", "CreateProject", "the Archon client does not expose project write endpoints")
}

// UpdateProject implements plugin.TaskClient
func (a *ArchonClientAdapter) UpdateProject(ctx context.Context, projectID string, request plugin.UpdateProjectRequest) (*plugin.Project, error) {
	return nil, plugin.Unsupported("archon", "UpdateProject", "the Archon client does not expose project write endpoints")
}

// DeleteProject implements plugin.TaskClient
func (a *ArchonClientAdapter) DeleteProject(ctx context.Context, projectID string) error {
	return plugin.Unsupported("archon", "DeleteProject", "the Archon client does not expose project write endpoints")
}

// Close implements plugin.TaskClient
func (a *ArchonClientAdapter) Close() error {
	// No resources to close for the Archon client
	return nil
}

// convertTask converts an Archon Task to a generic plugin.Task
func (a *ArchonClientAdapter) convertTask(task Task) plugin.Task {
	// Extract tags from the feature field
	tags := []string{}
	if task.Feature != nil && *task.Feature != "" {
		tags = append(tags, *task.Feature)
	}

	// Convert timestamps
	createdAt := task.CreatedAt.Time
	updatedAt := task.UpdatedAt.Time

	var archivedAt *time.Time
	if task.ArchivedAt != nil {
		archivedTime := task.ArchivedAt.Time
		archivedAt = &archivedTime
	}

	// Get parent ID
	var parentID *string
	if task.ParentTaskID != nil && *task.ParentTaskID != "" {
		parentID = task.ParentTaskID
	}

	return plugin.Task{
		ID:          task.ID,
		ProjectID:   task.ProjectID,
		ParentID:    parentID,
		Title:       task.Title,
		Description: task.Description,
		Status:      a.fromBackendStatus(task.Status),
		Priority:    task.TaskOrder,
		Assignee:    task.Assignee,
		Tags:        tags,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		Archived:    task.Archived,
		ArchivedAt:  archivedAt,
		Extra: map[string]interface{}{
			"sources":       task.Sources,
			"code_examples": task.CodeExamples,
		},
	}
}

// convertProject converts an Archon Project to a generic plugin.Project
func (a *ArchonClientAdapter) convertProject(project Project) plugin.Project {
	return plugin.Project{
		ID:          project.ID,
		Title:       project.Title,
		Description: project.Description,
		Pinned:      project.Pinned,
		CreatedAt:   project.CreatedAt.Time,
		UpdatedAt:   project.UpdatedAt.Time,
		Extra: map[string]interface{}{
			"github_repo": project.GitHubRepo,
			"docs":        project.Docs,
			"features":    project.Features,
			"data":        project.Data,
		},
	}
}

// init registers the Archon plugin with the global registry. A failure here
// means duplicate registration — a programming error, so fail loudly.
func init() {
	if err := plugin.Register(&ArchonPlugin{}); err != nil {
		panic(err)
	}
}
