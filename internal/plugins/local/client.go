package local

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/yousfisaad/lazyarchon/v2/internal/plugin"
)

// PluginName is the unique plugin registry name for the local backend.
const PluginName = "local"

// defaultProjectTitle is the project seeded into a fresh database.
const defaultProjectTitle = "Inbox"

// LocalClient implements plugin.TaskClient on top of a local SQLite store.
type LocalClient struct {
	store *store

	// enforceDependencies rejects doing/review moves while unfinished
	// blockers remain (tasks.enforce_dependencies; advisory when false).
	// Read once at construction — changing the flag restarts the process.
	enforceDependencies bool
}

// compile-time interface check
var _ plugin.TaskClient = (*LocalClient)(nil)

// newClient opens the store at path and returns a client over it, advisory
// dependencies (the default).
func newClient(path string) (*LocalClient, error) {
	return newClientWithEnforcement(path, false)
}

// newClientWithEnforcement opens the store at path with dependency
// enforcement configured.
func newClientWithEnforcement(path string, enforceDependencies bool) (*LocalClient, error) {
	st, err := openStore(path)
	if err != nil {
		return nil, err
	}

	return &LocalClient{store: st, enforceDependencies: enforceDependencies}, nil
}

// HealthCheck verifies the database is reachable.
func (c *LocalClient) HealthCheck(ctx context.Context) error {
	return c.store.ping(ctx)
}

// Close releases the database connection.
func (c *LocalClient) Close() error {
	return c.store.Close()
}

// ---------- tasks ----------

// ListTasks retrieves tasks matching the filters.
func (c *LocalClient) ListTasks(ctx context.Context, filters plugin.TaskFilters) (*plugin.TaskListResult, error) {
	if filters.Status != nil && *filters.Status != "" && !isValidStatus(*filters.Status) {
		return nil, validationError("ListTasks", fmt.Sprintf("invalid status %q (want todo, doing, review or done)", *filters.Status))
	}

	result, err := c.store.listTasks(ctx, filters)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetTask retrieves a single task by ID.
func (c *LocalClient) GetTask(ctx context.Context, taskID string) (*plugin.Task, error) {
	task, err := c.store.getTask(ctx, taskID)
	if err != nil {
		return nil, mapRowError(err, "GetTask", "task "+taskID)
	}

	return task, nil
}

// CreateTask creates a new task with defaults and validation applied.
func (c *LocalClient) CreateTask(ctx context.Context, request plugin.CreateTaskRequest) (*plugin.Task, error) {
	if err := applyCreateDefaults(&request); err != nil {
		return nil, err
	}

	if err := c.validateCreateRefs(ctx, request); err != nil {
		return nil, err
	}

	// Self-reference is impossible here (the new ID does not exist yet), so
	// validation only checks that each blocker exists.
	blockedBy, err := c.prepareCreateDependencies(ctx, request.Status, request.BlockedBy)
	if err != nil {
		return nil, err
	}

	// One clock reading for both stamps; insertTask makes the stored pair
	// strictly monotonic against existing rows anyway.
	now := nowFunc()

	task := &plugin.Task{
		ID:          newID(),
		ProjectID:   request.ProjectID,
		ParentID:    request.ParentID,
		Title:       request.Title,
		Description: request.Description,
		Status:      request.Status,
		Priority:    request.Priority,
		Assignee:    request.Assignee,
		Worktree:    request.Worktree,
		Tags:        normalizeTags(request.Tags),
		DueDate:     request.DueDate,
		Extra:       request.Extra,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := c.store.insertTask(ctx, task); err != nil {
		// FK violations surface here under concurrency; translate them.
		if isForeignKeyError(err) {
			return nil, notFound("CreateTask", "referenced project or parent task")
		}
		return nil, err
	}

	if len(blockedBy) > 0 {
		if err := c.store.setTaskDependencies(ctx, task.ID, blockedBy); err != nil {
			if isForeignKeyError(err) {
				return nil, notFound("CreateTask", "referenced blocker task")
			}
			return nil, err
		}
	}

	return c.store.getTask(ctx, task.ID)
}

// applyCreateDefaults trims the title and fills default status/priority,
// rejecting invalid scalar fields.
func applyCreateDefaults(request *plugin.CreateTaskRequest) error {
	title := strings.TrimSpace(request.Title)
	if title == "" {
		return validationError("CreateTask", "title is required")
	}
	request.Title = title

	if request.Status == "" {
		request.Status = plugin.StatusTodo
	} else if !isValidStatus(request.Status) {
		return validationError("CreateTask", fmt.Sprintf("invalid status %q (want todo, doing, review or done)", request.Status))
	}

	if request.Priority == 0 {
		request.Priority = plugin.PriorityMedium
	} else if !isValidPriority(request.Priority) {
		return validationError("CreateTask", fmt.Sprintf("invalid priority %d (want 1-5, lower is more urgent)", request.Priority))
	}

	return nil
}

// validateCreateRefs checks that the referenced project and parent task
// exist.
func (c *LocalClient) validateCreateRefs(ctx context.Context, request plugin.CreateTaskRequest) error {
	if request.ProjectID == "" {
		return validationError("CreateTask", "project_id is required")
	}

	projectExists, err := c.store.projectExists(ctx, request.ProjectID)
	if err != nil {
		return err
	}
	if !projectExists {
		return notFound("CreateTask", "project "+request.ProjectID)
	}

	if request.ParentID != nil && *request.ParentID != "" {
		parentExists, err := c.store.taskExists(ctx, *request.ParentID)
		if err != nil {
			return err
		}
		if !parentExists {
			return notFound("CreateTask", "parent task "+*request.ParentID)
		}
	}

	return nil
}

// prepareDependencyUpdate validates a proposed blocker-set replacement —
// every ID exists, no self-reference, no cycle — and normalizes it in place.
func (c *LocalClient) prepareDependencyUpdate(ctx context.Context, taskID string, request *plugin.UpdateTaskRequest) error {
	if request.BlockedBy == nil {
		return nil
	}

	blockedBy, err := c.validateBlockedBy(ctx, "UpdateTask", taskID, *request.BlockedBy)
	if err != nil {
		return err
	}

	edges, err := c.store.dependencyEdges(ctx)
	if err != nil {
		return err
	}
	if wouldCreateDependencyCycle(edges, taskID, blockedBy) {
		return validationError("UpdateTask", "setting these dependencies would create a cycle")
	}

	request.BlockedBy = &blockedBy

	return nil
}

// enforceUpdateReadiness rejects doing/review moves while unfinished
// blockers remain, judged on the effective blocker set: the incoming
// replacement when the same request carries it, else the stored edges — a
// one-request status+deps change is judged on the new deps.
func (c *LocalClient) enforceUpdateReadiness(ctx context.Context, taskID string, request plugin.UpdateTaskRequest) error {
	if !c.enforceDependencies || request.Status == nil || !requiresReadiness(*request.Status) {
		return nil
	}

	var blockers []unfinishedBlocker
	var err error
	if request.BlockedBy != nil {
		blockers, err = c.store.unfinishedAmong(ctx, *request.BlockedBy)
	} else {
		blockers, err = c.store.unfinishedBlockers(ctx, taskID)
	}
	if err != nil {
		return err
	}

	return rejectIfBlocked("UpdateTask", blockers)
}

// prepareCreateDependencies validates the blocker set of a create request
// and enforces readiness when the task is born into doing/review. Returns
// the normalized set to persist.
func (c *LocalClient) prepareCreateDependencies(ctx context.Context, status string, ids []string) ([]string, error) {
	blockedBy, err := c.validateBlockedBy(ctx, "CreateTask", "", ids)
	if err != nil {
		return nil, err
	}

	if c.enforceDependencies && requiresReadiness(status) {
		blockers, err := c.store.unfinishedAmong(ctx, blockedBy)
		if err != nil {
			return nil, err
		}
		if err := rejectIfBlocked("CreateTask", blockers); err != nil {
			return nil, err
		}
	}

	return blockedBy, nil
}

// validateBlockedBy normalizes a proposed blocker set (deduplicated,
// order-preserving) and checks that every ID refers to an existing task.
// taskID is "" on create, where self-reference is impossible.
func (c *LocalClient) validateBlockedBy(ctx context.Context, operation, taskID string, ids []string) ([]string, error) {
	normalized := make([]string, 0, len(ids))
	seen := map[string]bool{}

	for _, id := range ids {
		if id == taskID {
			return nil, validationError(operation, "a task cannot depend on itself")
		}

		if seen[id] {
			continue
		}
		seen[id] = true

		exists, err := c.store.taskExists(ctx, id)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, notFound(operation, "blocker task "+id)
		}

		normalized = append(normalized, id)
	}

	return normalized, nil
}

// requiresReadiness reports whether a status means work on the task has
// started, so its blockers must be finished first.
func requiresReadiness(status string) bool {
	return status == plugin.StatusDoing || status == plugin.StatusReview
}

// rejectIfBlocked builds the enforcement rejection naming the unfinished
// blockers, or returns nil when they are all done.
func rejectIfBlocked(operation string, blockers []unfinishedBlocker) error {
	if len(blockers) == 0 {
		return nil
	}

	parts := make([]string, 0, len(blockers))
	for _, blocker := range blockers {
		parts = append(parts, fmt.Sprintf("%s (%s)", blocker.Title, blocker.ID))
	}

	return validationError(operation, "task is blocked by unfinished tasks: "+strings.Join(parts, ", "))
}

// validateTaskUpdateFields checks scalar fields that need no store access.
func validateTaskUpdateFields(request plugin.UpdateTaskRequest) error {
	if request.Status != nil && !isValidStatus(*request.Status) {
		return validationError("UpdateTask", fmt.Sprintf("invalid status %q (want todo, doing, review or done)", *request.Status))
	}

	if request.Priority != nil && !isValidPriority(*request.Priority) {
		return validationError("UpdateTask", fmt.Sprintf("invalid priority %d (want 1-5, lower is more urgent)", *request.Priority))
	}

	if request.Title != nil && strings.TrimSpace(*request.Title) == "" {
		return validationError("UpdateTask", "title cannot be empty")
	}

	return nil
}

// validateTaskUpdateRefs checks that a referenced project and parent task
// exist and that reparenting would not create a cycle.
func (c *LocalClient) validateTaskUpdateRefs(ctx context.Context, taskID string, request plugin.UpdateTaskRequest) error {
	if request.ProjectID != nil {
		projectExists, err := c.store.projectExists(ctx, *request.ProjectID)
		if err != nil {
			return err
		}
		if !projectExists {
			return notFound("UpdateTask", "project "+*request.ProjectID)
		}
	}

	if request.ParentID != nil && *request.ParentID != "" {
		if *request.ParentID == taskID {
			return validationError("UpdateTask", "a task cannot be its own parent")
		}

		parentExists, err := c.store.taskExists(ctx, *request.ParentID)
		if err != nil {
			return err
		}
		if !parentExists {
			return notFound("UpdateTask", "parent task "+*request.ParentID)
		}

		cycle, err := c.store.wouldCreateCycle(ctx, taskID, *request.ParentID)
		if err != nil {
			return err
		}
		if cycle {
			return validationError("UpdateTask", "setting this parent would create a task cycle")
		}
	}

	return nil
}

// UpdateTask updates an existing task with the provided (non-nil) fields.
func (c *LocalClient) UpdateTask(ctx context.Context, taskID string, request plugin.UpdateTaskRequest) (*plugin.Task, error) {
	exists, err := c.store.taskExists(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, notFound("UpdateTask", "task "+taskID)
	}

	if err := validateTaskUpdateFields(request); err != nil {
		return nil, err
	}

	if err := c.validateTaskUpdateRefs(ctx, taskID, request); err != nil {
		return nil, err
	}

	if err := c.prepareDependencyUpdate(ctx, taskID, &request); err != nil {
		return nil, err
	}

	if err := c.enforceUpdateReadiness(ctx, taskID, request); err != nil {
		return nil, err
	}

	task, err := c.store.updateTask(ctx, taskID, request)
	if err != nil {
		if isForeignKeyError(err) {
			return nil, notFound("UpdateTask", "referenced project or parent task")
		}
		return nil, mapRowError(err, "UpdateTask", "task "+taskID)
	}

	return task, nil
}

// DeleteTask permanently deletes a task. Child tasks survive with their
// parent cleared; use UpdateTask with Archived=true to keep the task around.
func (c *LocalClient) DeleteTask(ctx context.Context, taskID string) error {
	if err := c.store.deleteTask(ctx, taskID); err != nil {
		return mapRowError(err, "DeleteTask", "task "+taskID)
	}

	return nil
}

// ---------- projects ----------

// ListProjects retrieves all projects.
func (c *LocalClient) ListProjects(ctx context.Context) ([]plugin.Project, error) {
	return c.store.listProjects(ctx)
}

// GetProject retrieves a single project by ID.
func (c *LocalClient) GetProject(ctx context.Context, projectID string) (*plugin.Project, error) {
	project, err := c.store.getProject(ctx, projectID)
	if err != nil {
		return nil, mapRowError(err, "GetProject", "project "+projectID)
	}

	return project, nil
}

// CreateProject creates a new project.
func (c *LocalClient) CreateProject(ctx context.Context, request plugin.CreateProjectRequest) (*plugin.Project, error) {
	title := strings.TrimSpace(request.Title)
	if title == "" {
		return nil, validationError("CreateProject", "title is required")
	}

	now := nowFunc()
	project := &plugin.Project{
		ID:          newID(),
		Title:       title,
		Description: request.Description,
		Color:       request.Color,
		Pinned:      request.Pinned,
		Extra:       request.Extra,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := c.store.insertProject(ctx, project); err != nil {
		return nil, err
	}

	return c.store.getProject(ctx, project.ID)
}

// UpdateProject updates an existing project with the provided (non-nil) fields.
func (c *LocalClient) UpdateProject(ctx context.Context, projectID string, request plugin.UpdateProjectRequest) (*plugin.Project, error) {
	if request.Title != nil && strings.TrimSpace(*request.Title) == "" {
		return nil, validationError("UpdateProject", "title cannot be empty")
	}

	project, err := c.store.updateProject(ctx, projectID, request)
	if err != nil {
		return nil, mapRowError(err, "UpdateProject", "project "+projectID)
	}

	return project, nil
}

// DeleteProject permanently deletes a project and all tasks in it.
func (c *LocalClient) DeleteProject(ctx context.Context, projectID string) error {
	if err := c.store.deleteProject(ctx, projectID); err != nil {
		return mapRowError(err, "DeleteProject", "project "+projectID)
	}

	return nil
}

// ---------- helpers ----------

// taskExists reports whether the task ID is present.
func (s *store) taskExists(ctx context.Context, taskID string) (bool, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks WHERE id = ?", taskID).Scan(&count); err != nil {
		return false, fmt.Errorf("failed to check task %s: %w", taskID, err)
	}

	return count > 0, nil
}

func isValidStatus(status string) bool {
	switch status {
	case plugin.StatusTodo, plugin.StatusDoing, plugin.StatusReview, plugin.StatusDone:
		return true
	default:
		return false
	}
}

func isValidPriority(priority int) bool {
	return priority >= plugin.PriorityCritical && priority <= plugin.PriorityBacklog
}

// normalizeTags ensures a nil slice becomes an empty slice (JSON "[]").
func normalizeTags(tags []string) []string {
	if tags == nil {
		return []string{}
	}

	return tags
}

// mapRowError translates sql.ErrNoRows into a plugin not-found error.
func mapRowError(err error, operation, entity string) error {
	if errors.Is(err, sql.ErrNoRows) {
		return notFound(operation, entity)
	}

	return err
}

// isForeignKeyError detects SQLite FOREIGN KEY constraint failures.
func isForeignKeyError(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "FOREIGN KEY constraint failed")
}

func notFound(operation, entity string) error {
	return plugin.NotFound(PluginName, operation, entity)
}

func validationError(operation, message string) error {
	return &plugin.PluginError{
		Plugin:      PluginName,
		Operation:   operation,
		Message:     message,
		Code:        "VALIDATION",
		Recoverable: false,
	}
}
