package mcpserver

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/yousfisaad/lazyarchon/v2/internal/plugin"
)

func (s *Server) registerTaskTools() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name: "list_tasks",
		Description: "List tasks as compact summaries (no descriptions). Filter by project, " +
			"status, text search, tags; archived tasks are hidden unless include_archived=true. " +
			"Paginate with limit/offset via next_offset.",
	}, s.listTasks)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_task",
		Description: "Get one task by ID, including description, extra fields and timestamps.",
	}, s.getTask)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name: "create_task",
		Description: "Create a task. Only title is required; project_id may be omitted when " +
			"exactly one project exists. Sources and code_examples are stored as extra fields.",
	}, s.createTask)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name: "update_task",
		Description: "Update a task. Omitted fields are left unchanged. Empty-string due_date " +
			"or parent_id clears the value; archived=true hides a task without deleting it.",
	}, s.updateTask)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "delete_task",
		Description: "Permanently delete a task. Subtasks survive with their parent cleared.",
	}, s.deleteTask)
}

// ---------- list_tasks ----------

type listTasksIn struct {
	ProjectID       string   `json:"project_id,omitempty" jsonschema:"Filter by project"`
	Status          string   `json:"status,omitempty" jsonschema:"Filter by status: todo, doing, review or done"`
	Search          string   `json:"search,omitempty" jsonschema:"Case-insensitive substring match on title and description"`
	Tags            []string `json:"tags,omitempty" jsonschema:"Only tasks carrying all of these tags"`
	IncludeArchived bool     `json:"include_archived,omitempty" jsonschema:"Also return archived tasks"`
	Limit           int      `json:"limit,omitempty" jsonschema:"Page size (default 500)"`
	Offset          int      `json:"offset,omitempty" jsonschema:"Skip this many results"`
}

type listTasksOut struct {
	Tasks      []taskSummary `json:"tasks"`
	Total      int           `json:"total"`
	HasMore    bool          `json:"has_more,omitempty"`
	NextOffset int           `json:"next_offset,omitempty"`
}

func (s *Server) listTasks(ctx context.Context, _ *mcp.CallToolRequest, in listTasksIn) (*mcp.CallToolResult, listTasksOut, error) {
	status, err := statusFilter(in.Status)
	if err != nil {
		return nil, listTasksOut{}, err
	}

	filters := plugin.TaskFilters{
		IncludeArchived: in.IncludeArchived,
		Tags:            in.Tags,
		Limit:           in.Limit,
		Offset:          in.Offset,
		Status:          status,
		ProjectID:       optionalString(in.ProjectID),
		SearchQuery:     optionalString(in.Search),
	}

	result, err := s.client.ListTasks(ctx, filters)
	if err != nil {
		return nil, listTasksOut{}, toolError(s.backend, err)
	}

	out := listTasksOut{
		Tasks:      make([]taskSummary, 0, len(result.Tasks)),
		Total:      result.TotalCount,
		HasMore:    result.HasMore,
		NextOffset: result.NextOffset,
	}

	for _, task := range result.Tasks {
		out.Tasks = append(out.Tasks, toTaskSummary(task))
	}

	return nil, out, nil
}

// ---------- get_task ----------

type getTaskIn struct {
	TaskID string `json:"task_id" jsonschema:"ID of the task"`
}

type getTaskOut struct {
	Task taskDetail `json:"task"`
}

func (s *Server) getTask(ctx context.Context, _ *mcp.CallToolRequest, in getTaskIn) (*mcp.CallToolResult, getTaskOut, error) {
	task, err := s.client.GetTask(ctx, in.TaskID)
	if err != nil {
		return nil, getTaskOut{}, toolError(s.backend, err)
	}

	return nil, getTaskOut{Task: toTaskDetail(*task)}, nil
}

// ---------- create_task ----------

type createTaskIn struct {
	Title        string        `json:"title" jsonschema:"Task title"`
	ProjectID    string        `json:"project_id,omitempty" jsonschema:"Target project; omit when exactly one project exists"`
	ParentID     string        `json:"parent_id,omitempty" jsonschema:"Existing task ID to nest under"`
	Description  string        `json:"description,omitempty" jsonschema:"Task details (markdown supported)"`
	Status       string        `json:"status,omitempty" jsonschema:"One of todo (default), doing, review, done"`
	Priority     int           `json:"priority,omitempty" jsonschema:"1 critical to 4 low (default 3)"`
	Assignee     string        `json:"assignee,omitempty" jsonschema:"User responsible for the task"`
	Tags         []string      `json:"tags,omitempty" jsonschema:"Labels for the task"`
	DueDate      string        `json:"due_date,omitempty" jsonschema:"YYYY-MM-DD or RFC3339"`
	Sources      []interface{} `json:"sources,omitempty" jsonschema:"Reference links stored as an extra field"`
	CodeExamples []interface{} `json:"code_examples,omitempty" jsonschema:"Code snippets stored as an extra field"`
}

type createTaskOut struct {
	Task taskDetail `json:"task"`
}

func (s *Server) createTask(ctx context.Context, _ *mcp.CallToolRequest, in createTaskIn) (*mcp.CallToolResult, createTaskOut, error) {
	if _, err := statusFilter(in.Status); err != nil {
		return nil, createTaskOut{}, err
	}

	if err := validatePriority(in.Priority); err != nil {
		return nil, createTaskOut{}, err
	}

	projectID, err := s.resolveProjectID(ctx, in.ProjectID)
	if err != nil {
		return nil, createTaskOut{}, err
	}

	request := plugin.CreateTaskRequest{
		ProjectID:   projectID,
		Title:       in.Title,
		Description: in.Description,
		Status:      in.Status,
		Priority:    in.Priority,
		Assignee:    in.Assignee,
		Tags:        in.Tags,
	}

	if in.ParentID != "" {
		parent := in.ParentID
		request.ParentID = &parent
	}

	if in.DueDate != "" {
		due, err := parseDueDate(in.DueDate)
		if err != nil {
			return nil, createTaskOut{}, err
		}
		request.DueDate = &due
	}

	request.Extra = buildExtra(in.Sources, in.CodeExamples, nil)

	task, err := s.client.CreateTask(ctx, request)
	if err != nil {
		return nil, createTaskOut{}, toolError(s.backend, err)
	}

	return nil, createTaskOut{Task: toTaskDetail(*task)}, nil
}

// ---------- update_task ----------

type updateTaskIn struct {
	TaskID       string        `json:"task_id" jsonschema:"ID of the task to update"`
	Title        *string       `json:"title,omitempty" jsonschema:"New title"`
	Description  *string       `json:"description,omitempty" jsonschema:"New description"`
	Status       *string       `json:"status,omitempty" jsonschema:"New status: todo, doing, review or done"`
	Priority     *int          `json:"priority,omitempty" jsonschema:"New priority: 1 critical to 4 low"`
	Assignee     *string       `json:"assignee,omitempty" jsonschema:"New assignee"`
	Tags         *[]string     `json:"tags,omitempty" jsonschema:"Replacement tag set"`
	DueDate      *string       `json:"due_date,omitempty" jsonschema:"New due date; empty string clears it"`
	ProjectID    *string       `json:"project_id,omitempty" jsonschema:"Move the task to another project"`
	ParentID     *string       `json:"parent_id,omitempty" jsonschema:"New parent task; empty string clears it"`
	Archived     *bool         `json:"archived,omitempty" jsonschema:"Archive or restore the task"`
	Sources      []interface{} `json:"sources,omitempty" jsonschema:"Replacement reference links (extra field)"`
	CodeExamples []interface{} `json:"code_examples,omitempty" jsonschema:"Replacement code snippets (extra field)"`
}

type updateTaskOut struct {
	Task taskDetail `json:"task"`
}

func (s *Server) updateTask(ctx context.Context, _ *mcp.CallToolRequest, in updateTaskIn) (*mcp.CallToolResult, updateTaskOut, error) {
	if in.Status != nil {
		if _, err := statusFilter(*in.Status); err != nil {
			return nil, updateTaskOut{}, err
		}
	}

	if in.Priority != nil {
		if err := validatePriority(*in.Priority); err != nil {
			return nil, updateTaskOut{}, err
		}
	}

	request := plugin.UpdateTaskRequest{
		Title:       in.Title,
		Description: in.Description,
		Status:      in.Status,
		Priority:    in.Priority,
		Assignee:    in.Assignee,
		Tags:        in.Tags,
		ProjectID:   in.ProjectID,
		ParentID:    in.ParentID,
		Archived:    in.Archived,
	}

	if in.DueDate != nil {
		if *in.DueDate == "" {
			// Inner nil pointer = clear the due date.
			request.DueDate = new(*time.Time)
		} else {
			due, err := parseDueDate(*in.DueDate)
			if err != nil {
				return nil, updateTaskOut{}, err
			}
			duePtr := &due
			request.DueDate = &duePtr
		}
	}

	// sources/code_examples live inside Extra, which the backend replaces
	// wholesale: merge with the task's current Extra so unspecified keys
	// survive.
	if in.Sources != nil || in.CodeExamples != nil {
		current, err := s.client.GetTask(ctx, in.TaskID)
		if err != nil {
			return nil, updateTaskOut{}, toolError(s.backend, err)
		}

		request.Extra = new(map[string]interface{})
		*request.Extra = buildExtra(in.Sources, in.CodeExamples, current.Extra)
	}

	task, err := s.client.UpdateTask(ctx, in.TaskID, request)
	if err != nil {
		return nil, updateTaskOut{}, toolError(s.backend, err)
	}

	return nil, updateTaskOut{Task: toTaskDetail(*task)}, nil
}

// ---------- delete_task ----------

type deleteTaskIn struct {
	TaskID string `json:"task_id" jsonschema:"ID of the task to delete"`
}

type deleteTaskOut struct {
	Deleted bool `json:"deleted"`
}

func (s *Server) deleteTask(ctx context.Context, _ *mcp.CallToolRequest, in deleteTaskIn) (*mcp.CallToolResult, deleteTaskOut, error) {
	if err := s.client.DeleteTask(ctx, in.TaskID); err != nil {
		return nil, deleteTaskOut{}, toolError(s.backend, err)
	}

	return nil, deleteTaskOut{Deleted: true}, nil
}

// ---------- shared helpers ----------

// optionalString wraps a value in a pointer only when set, so empty inputs
// do not become empty-string filters.
func optionalString(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}

// statusFilter validates an optional status filter so a typo produces an
// error rather than a silently unfiltered (or empty) result set.
func statusFilter(status string) (*string, error) {
	if status == "" {
		return nil, nil
	}

	switch status {
	case plugin.StatusTodo, plugin.StatusDoing, plugin.StatusReview, plugin.StatusDone:
		return &status, nil
	default:
		return nil, fmt.Errorf("invalid status %q: must be one of todo, doing, review, done", status)
	}
}

// validatePriority checks an optional priority (0 = not set).
func validatePriority(priority int) error {
	if priority != 0 && (priority < plugin.PriorityCritical || priority > plugin.PriorityLow) {
		return fmt.Errorf("invalid priority %d: must be 1 (critical) to 4 (low)", priority)
	}

	return nil
}

// resolveProjectID defaults an omitted project to the only existing project,
// which covers a freshly seeded database.
func (s *Server) resolveProjectID(ctx context.Context, projectID string) (string, error) {
	if projectID != "" {
		return projectID, nil
	}

	projects, err := s.client.ListProjects(ctx)
	if err != nil {
		return "", toolError(s.backend, err)
	}

	switch len(projects) {
	case 1:
		return projects[0].ID, nil
	case 0:
		return "", fmt.Errorf("no projects exist yet; create one with create_project first")
	default:
		return "", fmt.Errorf("project_id is required when several projects exist; call list_projects to see them")
	}
}

// buildExtra assembles the backend Extra map from sources/code_examples,
// preserving any other keys already present in current.
func buildExtra(sources, codeExamples []interface{}, current map[string]interface{}) map[string]interface{} {
	if sources == nil && codeExamples == nil && current == nil {
		return nil
	}

	extra := make(map[string]interface{}, len(current)+2)
	for key, value := range current {
		if key != "sources" && key != "code_examples" {
			extra[key] = value
		}
	}

	if sources != nil {
		extra["sources"] = sources
	}

	if codeExamples != nil {
		extra["code_examples"] = codeExamples
	}

	return extra
}
