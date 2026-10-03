// Package gitea provides a lazytasks plugin for Gitea issues.
// This plugin uses the official Gitea Go SDK (code.gitea.io/sdk/gitea).
package gitea

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/yousfisaad/lazyarchon/v2/internal/plugin"
)

// PluginName is the unique identifier for this plugin
const PluginName = "gitea"

// PluginVersion is the current version of this plugin
const PluginVersion = "1.0.0"

// GiteaPlugin implements the plugin.Plugin interface for Gitea
type GiteaPlugin struct{}

// Name returns the plugin name
func (p *GiteaPlugin) Name() string {
	return PluginName
}

// Version returns the plugin version
func (p *GiteaPlugin) Version() string {
	return PluginVersion
}

// Description returns a human-readable description
func (p *GiteaPlugin) Description() string {
	return "Gitea Issues plugin - supports repository-based projects and issue tracking"
}

// Capabilities returns what features this plugin supports
func (p *GiteaPlugin) Capabilities() plugin.Capabilities {
	return plugin.Capabilities{
		SupportsProjects:    true,
		SupportsStatuses:    true,
		SupportsPriority:    false,
		SupportsAssignees:   true,
		SupportsDueDates:    false,
		SupportsSubtasks:    false,
		SupportsTags:        true,
		SupportsDescription: true,
		SupportsArchiving:   true,
		SupportsSearch:      false,
	}
}

// CreateClient creates a new Gitea client from the plugin configuration
func (p *GiteaPlugin) CreateClient(config plugin.PluginConfig) (plugin.TaskClient, error) {
	// Validate required configuration
	if config.BaseURL == "" {
		return nil, fmt.Errorf("base_url is required for gitea plugin")
	}
	if config.APIKey == "" {
		return nil, fmt.Errorf("api_key is required for gitea plugin")
	}

	// Determine timeout
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: timeout,
	}

	// Create Gitea SDK client
	client, err := gitea.NewClient(config.BaseURL,
		gitea.SetToken(config.APIKey),
		gitea.SetHTTPClient(httpClient),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gitea client: %w", err)
	}

	// Create the adapter
	adapter := &GiteaClientAdapter{
		client:   client,
		mappings: config.Mappings,
		baseURL:  config.BaseURL,
	}

	return adapter, nil
}

// GiteaClientAdapter adapts the Gitea SDK client to the generic plugin.TaskClient interface
type GiteaClientAdapter struct {
	client   *gitea.Client
	mappings plugin.FieldMappings
	baseURL  string
}

// toBackendStatus converts a generic status to Gitea state
func (a *GiteaClientAdapter) toBackendStatus(genericStatus string) gitea.StateType {
	switch genericStatus {
	case plugin.StatusTodo, plugin.StatusDoing, plugin.StatusReview:
		return gitea.StateOpen
	case plugin.StatusDone:
		return gitea.StateClosed
	default:
		return gitea.StateOpen
	}
}

// fromBackendStatus converts a Gitea state to generic status
func (a *GiteaClientAdapter) fromBackendStatus(state gitea.StateType) string {
	switch state {
	case gitea.StateOpen:
		return plugin.StatusTodo
	case gitea.StateClosed:
		return plugin.StatusDone
	default:
		return plugin.StatusTodo
	}
}

// HealthCheck implements plugin.TaskClient
func (a *GiteaClientAdapter) HealthCheck(ctx context.Context) error {
	// Get version to verify connectivity
	_, _, err := a.client.ServerVersion()
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	return nil
}

// ListTasks implements plugin.TaskClient
func (a *GiteaClientAdapter) ListTasks(ctx context.Context, filters plugin.TaskFilters) (*plugin.TaskListResult, error) {
	// For Gitea, we need to know which repository to query
	// If ProjectID is set, use it as "owner/repo"
	var owner, repo string
	if filters.ProjectID != nil {
		owner, repo = parseRepoID(*filters.ProjectID)
	}

	if owner == "" || repo == "" {
		// No specific repo selected, list all issues from all repos
		return a.listAllIssues(ctx, filters)
	}

	// List issues from specific repository
	return a.listRepoIssues(ctx, owner, repo, filters)
}

// listRepoIssues lists issues from a specific repository
func (a *GiteaClientAdapter) listRepoIssues(ctx context.Context, owner, repo string, filters plugin.TaskFilters) (*plugin.TaskListResult, error) {
	// Build issue list options
	opts := gitea.ListIssueOption{
		ListOptions: gitea.ListOptions{
			PageSize: 100, // Fetch up to 100 issues per page
		},
	}

	// Apply state filter
	if filters.Status != nil {
		opts.State = a.toBackendStatus(*filters.Status)
	} else if !filters.IncludeArchived {
		// By default, only show open issues
		opts.State = gitea.StateOpen
	}

	issues, resp, err := a.client.ListRepoIssues(owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list issues: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list issues: status %d", resp.StatusCode)
	}

	// Convert Gitea issues to generic tasks
	tasks := make([]plugin.Task, 0, len(issues))
	for _, issue := range issues {
		// Skip pull requests (Gitea returns PRs as issues)
		if issue.PullRequest != nil {
			continue
		}
		tasks = append(tasks, a.convertIssue(issue, owner, repo))
	}

	return &plugin.TaskListResult{
		Tasks:      tasks,
		TotalCount: len(tasks),
		HasMore:    false,
		NextOffset: 0,
	}, nil
}

// listAllIssues lists issues from all accessible repositories
func (a *GiteaClientAdapter) listAllIssues(ctx context.Context, filters plugin.TaskFilters) (*plugin.TaskListResult, error) {
	// Get list of repos
	repos, resp, err := a.client.ListMyRepos(gitea.ListReposOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list repositories: status %d", resp.StatusCode)
	}

	// Collect issues from all repos
	allTasks := make([]plugin.Task, 0)

	for _, repo := range repos {
		// Skip forks and mirrors for cleaner task list
		if repo.Fork || repo.Mirror {
			continue
		}

		result, err := a.listRepoIssues(ctx, repo.Owner.UserName, repo.Name, filters)
		if err != nil {
			// Log error but continue with other repos
			continue
		}

		allTasks = append(allTasks, result.Tasks...)
	}

	return &plugin.TaskListResult{
		Tasks:      allTasks,
		TotalCount: len(allTasks),
		HasMore:    false,
		NextOffset: 0,
	}, nil
}

// GetTask implements plugin.TaskClient
func (a *GiteaClientAdapter) GetTask(ctx context.Context, taskID string) (*plugin.Task, error) {
	// Parse task ID: "owner/repo#number" or just "number" (requires ProjectID context)
	owner, repo, number, err := parseTaskID(taskID)
	if err != nil {
		return nil, fmt.Errorf("invalid task id: %w", err)
	}

	issue, resp, err := a.client.GetIssue(owner, repo, number)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get issue: status %d", resp.StatusCode)
	}

	task := a.convertIssue(issue, owner, repo)
	return &task, nil
}

// CreateTask implements plugin.TaskClient
func (a *GiteaClientAdapter) CreateTask(ctx context.Context, request plugin.CreateTaskRequest) (*plugin.Task, error) {
	// Parse project ID to get owner/repo
	owner, repo := parseRepoID(request.ProjectID)
	if owner == "" || repo == "" {
		return nil, fmt.Errorf("valid project_id (owner/repo format) is required")
	}

	// Map generic status to Gitea state
	state := a.toBackendStatus(request.Status)
	isClosed := state == gitea.StateClosed

	// Create issue options
	opts := gitea.CreateIssueOption{
		Title:  request.Title,
		Body:   request.Description,
		Closed: isClosed,
	}

	// Add assignee if specified
	if request.Assignee != "" {
		opts.Assignees = []string{request.Assignee}
	}

	issue, resp, err := a.client.CreateIssue(owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create issue: status %d", resp.StatusCode)
	}

	task := a.convertIssue(issue, owner, repo)
	return &task, nil
}

// UpdateTask implements plugin.TaskClient
func (a *GiteaClientAdapter) UpdateTask(ctx context.Context, taskID string, request plugin.UpdateTaskRequest) (*plugin.Task, error) {
	// Parse task ID
	owner, repo, number, err := parseTaskID(taskID)
	if err != nil {
		return nil, fmt.Errorf("invalid task id: %w", err)
	}

	// Build edit options
	opts := gitea.EditIssueOption{}

	if request.Title != nil {
		opts.Title = *request.Title
	}
	if request.Description != nil {
		desc := *request.Description
		opts.Body = &desc
	}
	if request.Status != nil {
		state := a.toBackendStatus(*request.Status)
		opts.State = &state
	}
	if request.Assignee != nil {
		if *request.Assignee == "" {
			opts.Assignees = []string{}
		} else {
			opts.Assignees = []string{*request.Assignee}
		}
	}

	issue, resp, err := a.client.EditIssue(owner, repo, number, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to update issue: %w", err)
	}
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to update issue: status %d", resp.StatusCode)
	}

	task := a.convertIssue(issue, owner, repo)
	return &task, nil
}

// DeleteTask implements plugin.TaskClient
// In Gitea, we close the issue instead of deleting it
func (a *GiteaClientAdapter) DeleteTask(ctx context.Context, taskID string) error {
	// Parse task ID
	owner, repo, number, err := parseTaskID(taskID)
	if err != nil {
		return fmt.Errorf("invalid task id: %w", err)
	}

	// Close the issue (Gitea doesn't support true deletion via API)
	closed := gitea.StateClosed
	opts := gitea.EditIssueOption{
		State: &closed,
	}

	_, resp, err := a.client.EditIssue(owner, repo, number, opts)
	if err != nil {
		return fmt.Errorf("failed to close issue: %w", err)
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to close issue: status %d", resp.StatusCode)
	}

	return nil
}

// ListProjects implements plugin.TaskClient
// For Gitea, repositories act as projects
func (a *GiteaClientAdapter) ListProjects(ctx context.Context) ([]plugin.Project, error) {
	repos, resp, err := a.client.ListMyRepos(gitea.ListReposOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list repositories: status %d", resp.StatusCode)
	}

	projects := make([]plugin.Project, 0, len(repos))
	for _, repo := range repos {
		// Skip forks and mirrors for cleaner project list
		if repo.Fork || repo.Mirror {
			continue
		}

		projects = append(projects, plugin.Project{
			ID:          fmt.Sprintf("%s/%s", repo.Owner.UserName, repo.Name),
			Title:       repo.Name,
			Description: repo.Description,
			CreatedAt:   repo.Created,
			UpdatedAt:   repo.Updated,
			Extra: map[string]interface{}{
				"owner":     repo.Owner.UserName,
				"repo":      repo.Name,
				"html_url":  repo.HTMLURL,
				"is_fork":   repo.Fork,
				"is_mirror": repo.Mirror,
			},
		})
	}

	return projects, nil
}

// GetProject implements plugin.TaskClient
func (a *GiteaClientAdapter) GetProject(ctx context.Context, projectID string) (*plugin.Project, error) {
	owner, repo := parseRepoID(projectID)
	if owner == "" || repo == "" {
		return nil, fmt.Errorf("invalid project id: expected owner/repo format")
	}

	repository, resp, err := a.client.GetRepo(owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get repository: status %d", resp.StatusCode)
	}

	project := plugin.Project{
		ID:          fmt.Sprintf("%s/%s", repository.Owner.UserName, repository.Name),
		Title:       repository.Name,
		Description: repository.Description,
		CreatedAt:   repository.Created,
		UpdatedAt:   repository.Updated,
		Extra: map[string]interface{}{
			"owner":     repository.Owner.UserName,
			"repo":      repository.Name,
			"html_url":  repository.HTMLURL,
			"is_fork":   repository.Fork,
			"is_mirror": repository.Mirror,
		},
	}

	return &project, nil
}

// CreateProject implements plugin.TaskClient
func (a *GiteaClientAdapter) CreateProject(ctx context.Context, request plugin.CreateProjectRequest) (*plugin.Project, error) {
	return nil, plugin.Unsupported("gitea", "CreateProject", "projects map to repositories and creating repositories is out of scope; create tasks in an existing project instead")
}

// UpdateProject implements plugin.TaskClient
func (a *GiteaClientAdapter) UpdateProject(ctx context.Context, projectID string, request plugin.UpdateProjectRequest) (*plugin.Project, error) {
	return nil, plugin.Unsupported("gitea", "UpdateProject", "projects map to repositories and editing repository metadata is out of scope")
}

// DeleteProject implements plugin.TaskClient
func (a *GiteaClientAdapter) DeleteProject(ctx context.Context, projectID string) error {
	return plugin.Unsupported("gitea", "DeleteProject", "projects map to repositories and deleting repositories is out of scope")
}

// Close implements plugin.TaskClient
func (a *GiteaClientAdapter) Close() error {
	// No resources to close
	return nil
}

// convertIssue converts a Gitea Issue to a generic plugin.Task
func (a *GiteaClientAdapter) convertIssue(issue *gitea.Issue, owner, repo string) plugin.Task {
	// Extract labels as tags
	tags := make([]string, 0, len(issue.Labels))
	for _, label := range issue.Labels {
		tags = append(tags, label.Name)
	}

	// Get first assignee (Gitea supports multiple)
	assignee := ""
	if len(issue.Assignees) > 0 {
		assignee = issue.Assignees[0].UserName
	}

	return plugin.Task{
		ID:          fmt.Sprintf("%s/%s#%d", owner, repo, issue.Index),
		ProjectID:   fmt.Sprintf("%s/%s", owner, repo),
		Title:       issue.Title,
		Description: issue.Body,
		Status:      a.fromBackendStatus(issue.State),
		Priority:    0, // Gitea doesn't have native priority
		Assignee:    assignee,
		Tags:        tags,
		CreatedAt:   issue.Created,
		UpdatedAt:   issue.Updated,
		Archived:    issue.State == gitea.StateClosed,
		ArchivedAt:  nil,
		Extra: map[string]interface{}{
			"number":   issue.Index,
			"owner":    owner,
			"repo":     repo,
			"html_url": issue.URL,
		},
	}
}

// parseRepoID parses a project ID in format "owner/repo"
func parseRepoID(projectID string) (owner, repo string) {
	// Split by "/"
	parts := splitN(projectID, "/", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

// parseTaskID parses a task ID in format "owner/repo#number" or "number"
// If only number is provided, we need to look it up (not supported in v1)
func parseTaskID(taskID string) (owner, repo string, number int64, err error) {
	// Try "owner/repo#number" format first
	parts := splitN(taskID, "#", 2)
	if len(parts) == 2 {
		// Extract owner/repo
		repoParts := splitN(parts[0], "/", 2)
		if len(repoParts) != 2 {
			return "", "", 0, fmt.Errorf("invalid task id format")
		}
		owner = repoParts[0]
		repo = repoParts[1]

		// Parse number
		n, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return "", "", 0, fmt.Errorf("invalid issue number: %w", err)
		}
		number = n
		return owner, repo, number, nil
	}

	// If no "#" found, try parsing as just a number (requires context from ProjectID)
	n, err := strconv.ParseInt(taskID, 10, 64)
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid task id format: expected owner/repo#number")
	}

	// Return number only - caller must provide owner/repo context
	return "", "", n, nil
}

// splitN splits a string by separator, max n parts
func splitN(s, sep string, n int) []string {
	if n <= 0 {
		return []string{s}
	}

	result := make([]string, 0, n)
	start := 0

	for i := 0; i < n-1 && start < len(s); i++ {
		idx := findSep(s, sep, start)
		if idx == -1 {
			break
		}
		result = append(result, s[start:idx])
		start = idx + len(sep)
	}

	if start < len(s) {
		result = append(result, s[start:])
	}

	return result
}

// findSep finds the index of separator in string starting from pos
func findSep(s, sep string, pos int) int {
	if pos >= len(s) {
		return -1
	}
	for i := pos; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			return i
		}
	}
	return -1
}

// init registers the Gitea plugin with the global registry
func init() {
	plugin.Register(&GiteaPlugin{})
}
