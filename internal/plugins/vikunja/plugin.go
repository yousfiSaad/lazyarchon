// Package vikunja provides a lazytasks plugin for Vikunja task management.
// This plugin uses raw HTTP requests since there's no official Go SDK.
package vikunja

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/yousfisaad/lazyarchon/v2/internal/plugin"
)

// PluginName is the unique identifier for this plugin
const PluginName = "vikunja"

// PluginVersion is the current version of this plugin
const PluginVersion = "1.0.0"

// VikunjaPlugin implements the plugin.Plugin interface for Vikunja
type VikunjaPlugin struct{}

// Name returns the plugin name
func (p *VikunjaPlugin) Name() string {
	return PluginName
}

// Version returns the plugin version
func (p *VikunjaPlugin) Version() string {
	return PluginVersion
}

// Description returns a human-readable description
func (p *VikunjaPlugin) Description() string {
	return "Vikunja task management plugin - supports projects, priorities, and subtasks"
}

// Capabilities returns what features this plugin supports
func (p *VikunjaPlugin) Capabilities() plugin.Capabilities {
	return plugin.Capabilities{
		SupportsProjects:    true,
		SupportsStatuses:    true,
		SupportsPriority:    true,
		SupportsAssignees:   true,
		SupportsDueDates:    false, // v1: skip for simplicity
		SupportsSubtasks:    false, // v1: skip for simplicity
		SupportsTags:        true,
		SupportsDescription: true,
		SupportsArchiving:   false,
		SupportsSearch:      false,
	}
}

// CreateClient creates a new Vikunja client from the plugin configuration
func (p *VikunjaPlugin) CreateClient(config plugin.PluginConfig) (plugin.TaskClient, error) {
	// Validate required configuration
	if config.BaseURL == "" {
		return nil, fmt.Errorf("base_url is required for vikunja plugin")
	}
	if config.APIKey == "" {
		return nil, fmt.Errorf("api_key is required for vikunja plugin")
	}

	// Determine timeout
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	// Create HTTP client
	httpClient := &http.Client{
		Timeout: timeout,
	}

	// Create the adapter
	adapter := &VikunjaClientAdapter{
		httpClient: httpClient,
		baseURL:    config.BaseURL,
		apiKey:     config.APIKey,
		mappings:   config.Mappings,
	}

	return adapter, nil
}

// VikunjaClientAdapter adapts the Vikunja HTTP API to the generic plugin.TaskClient interface
type VikunjaClientAdapter struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	mappings   plugin.FieldMappings
}

// doRequest makes an HTTP request with authentication
func (a *VikunjaClientAdapter) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	url := a.baseURL + "/api/v1" + path

	var bodyReader *bytes.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// toBackendStatus converts a generic status to Vikunja done boolean
func (a *VikunjaClientAdapter) toBackendStatus(genericStatus string) bool {
	switch genericStatus {
	case plugin.StatusDone:
		return true
	case plugin.StatusTodo, plugin.StatusDoing, plugin.StatusReview:
		return false
	default:
		return false
	}
}

// fromBackendStatus converts Vikunja done boolean to generic status
func (a *VikunjaClientAdapter) fromBackendStatus(done bool) string {
	if done {
		return plugin.StatusDone
	}
	return plugin.StatusTodo
}

// HealthCheck implements plugin.TaskClient
func (a *VikunjaClientAdapter) HealthCheck(ctx context.Context) error {
	// Get user info to verify connectivity
	resp, err := a.doRequest(ctx, "GET", "/user", nil)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: status %d", resp.StatusCode)
	}

	return nil
}

// ListTasks implements plugin.TaskClient
func (a *VikunjaClientAdapter) ListTasks(ctx context.Context, filters plugin.TaskFilters) (*plugin.TaskListResult, error) {
	path := "/tasks"

	// Add query parameters for filtering
	query := "?"
	if filters.ProjectID != nil {
		query += "project_id=" + *filters.ProjectID + "&"
	}
	if !filters.IncludeArchived {
		query += "done=false&"
	}
	if filters.Limit > 0 {
		query += "per_page=" + strconv.Itoa(filters.Limit) + "&"
	}

	if query != "?" {
		path += query[:len(query)-1] // Remove trailing &
	}

	resp, err := a.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list tasks: status %d", resp.StatusCode)
	}

	var vikunjaTasks []VikunjaTask
	if err := json.NewDecoder(resp.Body).Decode(&vikunjaTasks); err != nil {
		return nil, fmt.Errorf("failed to decode tasks: %w", err)
	}

	// Convert to generic tasks
	tasks := make([]plugin.Task, 0, len(vikunjaTasks))
	for _, vt := range vikunjaTasks {
		tasks = append(tasks, a.convertTask(vt))
	}

	return &plugin.TaskListResult{
		Tasks:      tasks,
		TotalCount: len(tasks),
		HasMore:    false,
		NextOffset: 0,
	}, nil
}

// GetTask implements plugin.TaskClient
func (a *VikunjaClientAdapter) GetTask(ctx context.Context, taskID string) (*plugin.Task, error) {
	path := "/tasks/" + taskID

	resp, err := a.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get task: status %d", resp.StatusCode)
	}

	var vikunjaTask VikunjaTask
	if err := json.NewDecoder(resp.Body).Decode(&vikunjaTask); err != nil {
		return nil, fmt.Errorf("failed to decode task: %w", err)
	}

	task := a.convertTask(vikunjaTask)
	return &task, nil
}

// CreateTask implements plugin.TaskClient
func (a *VikunjaClientAdapter) CreateTask(ctx context.Context, request plugin.CreateTaskRequest) (*plugin.Task, error) {
	// Build create request
	createReq := CreateTaskRequest{
		Title:       request.Title,
		Description: request.Description,
		ProjectID:   parseInt64(request.ProjectID),
		Done:        a.toBackendStatus(request.Status),
		Priority:    request.Priority,
	}

	if request.ParentID != nil {
		createReq.ParentTaskID = parseInt64(*request.ParentID)
	}

	// Add labels as tags
	if len(request.Tags) > 0 {
		createReq.Labels = make([]int64, 0, len(request.Tags))
		// Note: v1 doesn't handle label ID mapping - would need to fetch labels first
	}

	resp, err := a.doRequest(ctx, "PUT", "/tasks", createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create task: status %d", resp.StatusCode)
	}

	var vikunjaTask VikunjaTask
	if err := json.NewDecoder(resp.Body).Decode(&vikunjaTask); err != nil {
		return nil, fmt.Errorf("failed to decode created task: %w", err)
	}

	task := a.convertTask(vikunjaTask)
	return &task, nil
}

// UpdateTask implements plugin.TaskClient
func (a *VikunjaClientAdapter) UpdateTask(ctx context.Context, taskID string, request plugin.UpdateTaskRequest) (*plugin.Task, error) {
	// First, get the current task
	currentTask, err := a.GetTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current task: %w", err)
	}

	// Build update request
	updateReq := UpdateTaskRequest{
		ID: parseInt64(taskID),
	}

	// Apply updates
	if request.Title != nil {
		updateReq.Title = *request.Title
	} else {
		updateReq.Title = currentTask.Title
	}

	if request.Description != nil {
		updateReq.Description = *request.Description
	} else {
		updateReq.Description = currentTask.Description
	}

	if request.Status != nil {
		updateReq.Done = a.toBackendStatus(*request.Status)
	} else {
		updateReq.Done = a.toBackendStatus(currentTask.Status)
	}

	if request.Priority != nil {
		updateReq.Priority = *request.Priority
	} else {
		updateReq.Priority = currentTask.Priority
	}

	updateReq.ProjectID = parseInt64(currentTask.ProjectID)

	resp, err := a.doRequest(ctx, "POST", "/tasks/"+taskID, updateReq)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to update task: status %d", resp.StatusCode)
	}

	var vikunjaTask VikunjaTask
	if err := json.NewDecoder(resp.Body).Decode(&vikunjaTask); err != nil {
		return nil, fmt.Errorf("failed to decode updated task: %w", err)
	}

	task := a.convertTask(vikunjaTask)
	return &task, nil
}

// DeleteTask implements plugin.TaskClient
// In Vikunja, we mark the task as done instead of deleting
func (a *VikunjaClientAdapter) DeleteTask(ctx context.Context, taskID string) error {
	// Mark as done (Vikunja doesn't support true deletion via API)
	_, err := a.UpdateTask(ctx, taskID, plugin.UpdateTaskRequest{
		Status: strPtr(plugin.StatusDone),
	})
	return err
}

// ListProjects implements plugin.TaskClient
func (a *VikunjaClientAdapter) ListProjects(ctx context.Context) ([]plugin.Project, error) {
	resp, err := a.doRequest(ctx, "GET", "/projects", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list projects: status %d", resp.StatusCode)
	}

	var vikunjaProjects []VikunjaProject
	if err := json.NewDecoder(resp.Body).Decode(&vikunjaProjects); err != nil {
		return nil, fmt.Errorf("failed to decode projects: %w", err)
	}

	projects := make([]plugin.Project, 0, len(vikunjaProjects))
	for _, vp := range vikunjaProjects {
		projects = append(projects, a.convertProject(vp))
	}

	return projects, nil
}

// GetProject implements plugin.TaskClient
func (a *VikunjaClientAdapter) GetProject(ctx context.Context, projectID string) (*plugin.Project, error) {
	path := "/projects/" + projectID

	resp, err := a.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get project: status %d", resp.StatusCode)
	}

	var vikunjaProject VikunjaProject
	if err := json.NewDecoder(resp.Body).Decode(&vikunjaProject); err != nil {
		return nil, fmt.Errorf("failed to decode project: %w", err)
	}

	project := a.convertProject(vikunjaProject)
	return &project, nil
}

// CreateProject implements plugin.TaskClient.
// Vikunja's API does support project creation (PUT /projects), but wiring it
// up is not implemented yet.
func (a *VikunjaClientAdapter) CreateProject(ctx context.Context, request plugin.CreateProjectRequest) (*plugin.Project, error) {
	return nil, plugin.Unsupported("vikunja", "CreateProject", "not implemented yet; create the project in Vikunja directly and it will appear on refresh")
}

// UpdateProject implements plugin.TaskClient
func (a *VikunjaClientAdapter) UpdateProject(ctx context.Context, projectID string, request plugin.UpdateProjectRequest) (*plugin.Project, error) {
	return nil, plugin.Unsupported("vikunja", "UpdateProject", "not implemented yet; update the project in Vikunja directly")
}

// DeleteProject implements plugin.TaskClient
func (a *VikunjaClientAdapter) DeleteProject(ctx context.Context, projectID string) error {
	return plugin.Unsupported("vikunja", "DeleteProject", "not implemented yet; delete the project in Vikunja directly")
}

// Close implements plugin.TaskClient
func (a *VikunjaClientAdapter) Close() error {
	// No resources to close
	return nil
}

// convertTask converts a VikunjaTask to a generic plugin.Task
func (a *VikunjaClientAdapter) convertTask(vt VikunjaTask) plugin.Task {
	// Extract labels as tags
	tags := make([]string, 0, len(vt.Labels))
	for _, label := range vt.Labels {
		tags = append(tags, label.Title)
	}

	// Get first assignee
	assignee := ""
	if len(vt.Assignees) > 0 {
		assignee = vt.Assignees[0].Username
	}

	return plugin.Task{
		ID:          strconv.FormatInt(vt.ID, 10),
		ProjectID:   strconv.FormatInt(vt.ProjectID, 10),
		ParentID:    int64PtrToStringPtr(vt.ParentTaskID),
		Title:       vt.Title,
		Description: vt.Description,
		Status:      a.fromBackendStatus(vt.Done),
		Priority:    vt.Priority,
		Assignee:    assignee,
		Tags:        tags,
		DueDate:     nil, // v1: skip due dates
		CreatedAt:   vt.Created,
		UpdatedAt:   vt.Updated,
		Archived:    vt.Done,
		ArchivedAt:  nil,
		Extra: map[string]interface{}{
			"identifier":   vt.Identifier,
			"index":        vt.Index,
			"hex_color":    vt.HexColor,
			"percent_done": vt.PercentDone,
		},
	}
}

// convertProject converts a VikunjaProject to a generic plugin.Project
func (a *VikunjaClientAdapter) convertProject(vp VikunjaProject) plugin.Project {
	return plugin.Project{
		ID:          strconv.FormatInt(vp.ID, 10),
		Title:       vp.Title,
		Description: vp.Description,
		Color:       vp.HexColor,
		Pinned:      false, // v1: not implemented
		CreatedAt:   vp.Created,
		UpdatedAt:   vp.Updated,
		Extra: map[string]interface{}{
			"owner_id":    vp.OwnerID,
			"identifier":  vp.Identifier,
			"hex_color":   vp.HexColor,
			"is_archived": vp.IsArchived,
		},
	}
}

// Helper functions

func parseInt64(s string) int64 {
	if s == "" {
		return 0
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

func int64PtrToStringPtr(n *int64) *string {
	if n == nil {
		return nil
	}
	s := strconv.FormatInt(*n, 10)
	return &s
}

func strPtr(s string) *string {
	return &s
}

// VikunjaTask represents a task in Vikunja
type VikunjaTask struct {
	ID           int64          `json:"id"`
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	Done         bool           `json:"done"`
	ProjectID    int64          `json:"project_id"`
	Priority     int            `json:"priority"`
	Labels       []VikunjaLabel `json:"labels"`
	Assignees    []VikunjaUser  `json:"assignees"`
	ParentTaskID *int64         `json:"parent_task_id"`
	Identifier   string         `json:"identifier"`
	Index        int            `json:"index"`
	HexColor     string         `json:"hex_color"`
	PercentDone  float64        `json:"percent_done"`
	Created      time.Time      `json:"created"`
	Updated      time.Time      `json:"updated"`
}

// VikunjaProject represents a project in Vikunja
type VikunjaProject struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	OwnerID     int64     `json:"owner_id"`
	Identifier  string    `json:"identifier"`
	HexColor    string    `json:"hex_color"`
	IsArchived  bool      `json:"is_archived"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
}

// VikunjaLabel represents a label in Vikunja
type VikunjaLabel struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	HexColor string `json:"hex_color"`
}

// VikunjaUser represents a user in Vikunja
type VikunjaUser struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// CreateTaskRequest represents the request body for creating a task
type CreateTaskRequest struct {
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	ProjectID    int64   `json:"project_id"`
	Done         bool    `json:"done"`
	Priority     int     `json:"priority"`
	Labels       []int64 `json:"labels"`
	ParentTaskID int64   `json:"parent_task_id,omitempty"`
}

// UpdateTaskRequest represents the request body for updating a task
type UpdateTaskRequest struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ProjectID   int64  `json:"project_id"`
	Done        bool   `json:"done"`
	Priority    int    `json:"priority"`
}

// init registers the Vikunja plugin with the global registry. A failure here
// means duplicate registration — a programming error, so fail loudly.
func init() {
	if err := plugin.Register(&VikunjaPlugin{}); err != nil {
		panic(err)
	}
}
