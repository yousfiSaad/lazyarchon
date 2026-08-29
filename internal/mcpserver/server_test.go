package mcpserver

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/yousfisaad/lazyarchon/v2/internal/plugin"
	localplugin "github.com/yousfisaad/lazyarchon/v2/internal/plugins/local"
)

// newTestSession spins up an MCP server over a temp-dir local database and
// connects a client through in-memory transports.
func newTestSession(t *testing.T) *mcp.ClientSession {
	t.Helper()

	client, err := (&localplugin.LocalPlugin{}).CreateClient(plugin.PluginConfig{
		Extra: map[string]interface{}{"path": filepath.Join(t.TempDir(), "test.db")},
	})
	if err != nil {
		t.Fatalf("creating local client: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	return connectSession(t, client)
}

// unsupportedProjectsClient forwards everything to a real local client but
// rejects project writes like the HTTP backends do.
type unsupportedProjectsClient struct {
	plugin.TaskClient
}

func (unsupportedProjectsClient) CreateProject(_ context.Context, _ plugin.CreateProjectRequest) (*plugin.Project, error) {
	return nil, plugin.Unsupported("stub", "CreateProject", "the stub backend cannot create projects")
}

func connectSession(t *testing.T, client plugin.TaskClient) *mcp.ClientSession {
	t.Helper()

	server := New(client, "local", "test-version")

	ctx := context.Background()
	t1, t2 := mcp.NewInMemoryTransports()

	serverSession, err := server.MCP().Connect(ctx, t1, nil)
	if err != nil {
		t.Fatalf("connecting server transport: %v", err)
	}
	t.Cleanup(func() { _ = serverSession.Close() })

	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
	session, err := mcpClient.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatalf("connecting client transport: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })

	return session
}

// callTool invokes a tool and decodes its structured content into out.
func callTool[T any](t *testing.T, session *mcp.ClientSession, name string, arguments map[string]interface{}) T {
	t.Helper()

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: arguments})
	if err != nil {
		t.Fatalf("CallTool(%s): %v", name, err)
	}

	if result.IsError {
		t.Fatalf("CallTool(%s) failed: %s", name, contentText(result))
	}

	if result.StructuredContent == nil {
		t.Fatalf("CallTool(%s) returned no structured content", name)
	}

	raw, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatalf("marshalling structured content: %v", err)
	}

	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshalling %s result into %T: %v", name, out, err)
	}

	return out
}

// callToolError invokes a tool that must fail and returns the message the
// caller sees. Accepts either a tool error result or a protocol-level
// rejection (e.g. input schema validation).
func callToolError(t *testing.T, session *mcp.ClientSession, name string, arguments map[string]interface{}) string {
	t.Helper()

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: arguments})
	if err != nil {
		return err.Error()
	}

	if !result.IsError {
		t.Fatalf("CallTool(%s) expected an error, got success: %+v", name, result.StructuredContent)
	}

	return contentText(result)
}

func contentText(result *mcp.CallToolResult) string {
	var sb strings.Builder
	for _, content := range result.Content {
		if text, ok := content.(*mcp.TextContent); ok {
			sb.WriteString(text.Text)
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func TestListProjectsReturnsSeededInbox(t *testing.T) {
	session := newTestSession(t)

	out := callTool[listProjectsOut](t, session, "list_projects", map[string]interface{}{})

	if out.Total != 1 || len(out.Projects) != 1 {
		t.Fatalf("projects = %d (total %d), want 1 seeded project", len(out.Projects), out.Total)
	}

	if out.Projects[0].Title != "Inbox" {
		t.Errorf("seeded project title = %q, want %q", out.Projects[0].Title, "Inbox")
	}
}

func TestProjectLifecycle(t *testing.T) {
	session := newTestSession(t)

	created := callTool[createProjectOut](t, session, "create_project", map[string]interface{}{
		"title":       "Side Project",
		"description": "things that pile up",
		"pinned":      true,
	})

	if created.Project.ID == "" {
		t.Fatal("created project has no ID")
	}

	if created.Project.Title != "Side Project" || !created.Project.Pinned {
		t.Errorf("created project = %+v, want title Side Project pinned", created.Project)
	}

	fetched := callTool[getProjectOut](t, session, "get_project", map[string]interface{}{
		"project_id": created.Project.ID,
	})

	if fetched.Project.ID != created.Project.ID || fetched.Project.Description != "things that pile up" {
		t.Errorf("fetched project = %+v", fetched.Project)
	}

	updated := callTool[updateProjectOut](t, session, "update_project", map[string]interface{}{
		"project_id": created.Project.ID,
		"title":      "Renamed",
	})

	if updated.Project.Title != "Renamed" || !updated.Project.Pinned {
		t.Errorf("updated project = %+v, want renamed and still pinned", updated.Project)
	}

	deleted := callTool[deleteProjectOut](t, session, "delete_project", map[string]interface{}{
		"project_id": created.Project.ID,
	})

	if !deleted.Deleted || deleted.TasksDeleted != 0 {
		t.Errorf("delete result = %+v, want deleted with 0 tasks", deleted)
	}

	notFound := callToolError(t, session, "get_project", map[string]interface{}{
		"project_id": created.Project.ID,
	})

	if !strings.Contains(notFound, "not found") {
		t.Errorf("error after delete = %q, want not found", notFound)
	}
}

func TestDeleteProjectGuardsAgainstLiveTasks(t *testing.T) {
	session := newTestSession(t)

	project := callTool[createProjectOut](t, session, "create_project", map[string]interface{}{"title": "Guarded"})
	task := callTool[createTaskOut](t, session, "create_task", map[string]interface{}{
		"title":      "blocking task",
		"project_id": project.Project.ID,
	})

	guard := callToolError(t, session, "delete_project", map[string]interface{}{
		"project_id": project.Project.ID,
	})

	if !strings.Contains(guard, "1 task") || !strings.Contains(guard, "delete_tasks") {
		t.Errorf("guard message = %q, want task count and delete_tasks hint", guard)
	}

	deleted := callTool[deleteProjectOut](t, session, "delete_project", map[string]interface{}{
		"project_id":   project.Project.ID,
		"delete_tasks": true,
	})

	if !deleted.Deleted || deleted.TasksDeleted != 1 {
		t.Errorf("delete result = %+v, want deleted with 1 task", deleted)
	}

	// The cascaded task must be gone.
	post := callToolError(t, session, "get_task", map[string]interface{}{"task_id": task.Task.ID})
	if !strings.Contains(post, "not found") {
		t.Errorf("task after project delete = %q, want not found", post)
	}
}

func TestTaskLifecycle(t *testing.T) {
	session := newTestSession(t)

	// No project_id: must resolve to the single seeded project.
	created := callTool[createTaskOut](t, session, "create_task", map[string]interface{}{
		"title": "Write the docs",
	})

	if created.Task.ID == "" || created.Task.ProjectID == "" {
		t.Fatalf("created task = %+v, want ID and resolved project", created.Task)
	}

	if created.Task.Status != plugin.StatusTodo || created.Task.Priority != plugin.PriorityMedium {
		t.Errorf("defaults: status = %q priority = %d, want todo/3", created.Task.Status, created.Task.Priority)
	}

	updated := callTool[updateTaskOut](t, session, "update_task", map[string]interface{}{
		"task_id":  created.Task.ID,
		"status":   "doing",
		"priority": 2,
		"tags":     []interface{}{"docs", "quality"},
		"due_date": "2026-09-01",
		"assignee": "saad",
	})

	if updated.Task.Status != "doing" || updated.Task.Priority != 2 {
		t.Errorf("updated status/priority = %q/%d", updated.Task.Status, updated.Task.Priority)
	}

	if updated.Task.DueDate == nil || !strings.HasPrefix(*updated.Task.DueDate, "2026-09-01") {
		t.Errorf("due date = %v, want 2026-09-01", updated.Task.DueDate)
	}

	if len(updated.Task.Tags) != 2 || updated.Task.Tags[0] != "docs" {
		t.Errorf("tags = %v", updated.Task.Tags)
	}

	// Empty string clears the due date.
	cleared := callTool[updateTaskOut](t, session, "update_task", map[string]interface{}{
		"task_id":  created.Task.ID,
		"due_date": "",
	})

	if cleared.Task.DueDate != nil {
		t.Errorf("due date after clear = %v, want nil", cleared.Task.DueDate)
	}

	// Subtask, then delete the parent: the child survives parentless.
	child := callTool[createTaskOut](t, session, "create_task", map[string]interface{}{
		"title":     "child task",
		"parent_id": created.Task.ID,
	})

	if child.Task.ParentID == nil || *child.Task.ParentID != created.Task.ID {
		t.Fatalf("child parent = %v, want %s", child.Task.ParentID, created.Task.ID)
	}

	callTool[deleteTaskOut](t, session, "delete_task", map[string]interface{}{"task_id": created.Task.ID})

	orphan := callTool[getTaskOut](t, session, "get_task", map[string]interface{}{"task_id": child.Task.ID})
	if orphan.Task.ParentID != nil {
		t.Errorf("orphan parent = %v, want nil after parent delete", orphan.Task.ParentID)
	}

	deleted := callTool[deleteTaskOut](t, session, "delete_task", map[string]interface{}{"task_id": child.Task.ID})
	if !deleted.Deleted {
		t.Error("delete_task reported deleted=false")
	}

	list := callTool[listTasksOut](t, session, "list_tasks", map[string]interface{}{})
	if list.Total != 0 || len(list.Tasks) != 0 {
		t.Errorf("task count after deletes = %d, want 0", list.Total)
	}
}

func TestListTasksFiltersAndPagination(t *testing.T) {
	session := newTestSession(t)

	mustCreate := func(title, status string, tags ...string) string {
		out := callTool[createTaskOut](t, session, "create_task", map[string]interface{}{
			"title":  title,
			"status": status,
			"tags":   toAnySlice(tags),
		})
		return out.Task.ID
	}

	mustCreate("alpha needle", "todo", "research")
	mustCreate("beta needle", "doing", "research")
	mustCreate("gamma plain", "done")

	byStatus := callTool[listTasksOut](t, session, "list_tasks", map[string]interface{}{"status": "todo"})
	if byStatus.Total != 1 || byStatus.Tasks[0].Title != "alpha needle" {
		t.Errorf("status filter: total = %d", byStatus.Total)
	}

	bySearch := callTool[listTasksOut](t, session, "list_tasks", map[string]interface{}{"search": "needle"})
	if bySearch.Total != 2 {
		t.Errorf("search filter: total = %d, want 2", bySearch.Total)
	}

	byTag := callTool[listTasksOut](t, session, "list_tasks", map[string]interface{}{"tags": []interface{}{"research"}})
	if byTag.Total != 2 {
		t.Errorf("tag filter: total = %d, want 2", byTag.Total)
	}

	page1 := callTool[listTasksOut](t, session, "list_tasks", map[string]interface{}{"limit": 2})
	if len(page1.Tasks) != 2 || page1.Total != 3 || !page1.HasMore || page1.NextOffset != 2 {
		t.Errorf("page 1 = %d tasks (total %d, more %v, next %d), want 2/3/true/2",
			len(page1.Tasks), page1.Total, page1.HasMore, page1.NextOffset)
	}

	page2 := callTool[listTasksOut](t, session, "list_tasks", map[string]interface{}{"limit": 2, "offset": 2})
	if len(page2.Tasks) != 1 || page2.HasMore {
		t.Errorf("page 2 = %d tasks (more %v), want 1/false", len(page2.Tasks), page2.HasMore)
	}

	// Archiving hides a task unless explicitly included.
	callTool[updateTaskOut](t, session, "update_task", map[string]interface{}{
		"task_id":  bySearch.Tasks[0].ID,
		"archived": true,
	})

	defaultList := callTool[listTasksOut](t, session, "list_tasks", map[string]interface{}{})
	if defaultList.Total != 2 {
		t.Errorf("total after archive = %d, want 2", defaultList.Total)
	}

	withArchived := callTool[listTasksOut](t, session, "list_tasks", map[string]interface{}{"include_archived": true})
	if withArchived.Total != 3 {
		t.Errorf("include_archived total = %d, want 3", withArchived.Total)
	}
}

func TestSourcesAndCodeExamplesRoundTrip(t *testing.T) {
	session := newTestSession(t)

	created := callTool[createTaskOut](t, session, "create_task", map[string]interface{}{
		"title":   "research task",
		"sources": []interface{}{map[string]interface{}{"url": "https://example.com"}},
	})

	sources, ok := created.Task.Extra["sources"].([]interface{})
	if !ok || len(sources) != 1 {
		t.Fatalf("sources after create = %#v", created.Task.Extra["sources"])
	}

	// Updating only code_examples must preserve sources.
	updated := callTool[updateTaskOut](t, session, "update_task", map[string]interface{}{
		"task_id":       created.Task.ID,
		"code_examples": []interface{}{map[string]interface{}{"lang": "go", "code": "fmt.Println()"}},
	})

	if _, ok := updated.Task.Extra["sources"].([]interface{}); !ok {
		t.Errorf("sources lost after code_examples update: %#v", updated.Task.Extra)
	}

	if _, ok := updated.Task.Extra["code_examples"].([]interface{}); !ok {
		t.Errorf("code_examples missing after update: %#v", updated.Task.Extra)
	}
}

func TestValidationErrorsSurfaceAsToolErrors(t *testing.T) {
	session := newTestSession(t)

	cases := []struct {
		name      string
		tool      string
		arguments map[string]interface{}
		wantIn    string
	}{
		{
			name:      "bad status filter",
			tool:      "list_tasks",
			arguments: map[string]interface{}{"status": "bogus"},
			wantIn:    "invalid status",
		},
		{
			name:      "bad priority",
			tool:      "create_task",
			arguments: map[string]interface{}{"title": "x", "priority": 9},
			wantIn:    "invalid priority",
		},
		{
			name:      "bad due date",
			tool:      "create_task",
			arguments: map[string]interface{}{"title": "x", "due_date": "not-a-date"},
			wantIn:    "invalid date",
		},
		{
			name:      "missing task",
			tool:      "get_task",
			arguments: map[string]interface{}{"task_id": "nope"},
			wantIn:    "not found",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			message := callToolError(t, session, testCase.tool, testCase.arguments)
			if !strings.Contains(message, testCase.wantIn) {
				t.Errorf("error = %q, want it to contain %q", message, testCase.wantIn)
			}
		})
	}
}

func TestCreateTaskRequiresProjectWhenSeveralExist(t *testing.T) {
	session := newTestSession(t)

	callTool[createProjectOut](t, session, "create_project", map[string]interface{}{"title": "Second project"})

	message := callToolError(t, session, "create_task", map[string]interface{}{"title": "orphan"})
	if !strings.Contains(message, "project_id is required") || !strings.Contains(message, "list_projects") {
		t.Errorf("error = %q, want project_id guidance mentioning list_projects", message)
	}
}

func TestUnsupportedBackendSurfacesHint(t *testing.T) {
	realClient, err := (&localplugin.LocalPlugin{}).CreateClient(plugin.PluginConfig{
		Extra: map[string]interface{}{"path": filepath.Join(t.TempDir(), "stub.db")},
	})
	if err != nil {
		t.Fatalf("creating local client: %v", err)
	}
	t.Cleanup(func() { _ = realClient.Close() })

	session := connectSession(t, unsupportedProjectsClient{TaskClient: realClient})

	message := callToolError(t, session, "create_project", map[string]interface{}{"title": "Nope"})

	for _, want := range []string{"stub backend", "not supported"} {
		if !strings.Contains(message, want) {
			t.Errorf("error = %q, want it to contain %q", message, want)
		}
	}

	// Task tools must still work through the wrapped client.
	created := callTool[createTaskOut](t, session, "create_task", map[string]interface{}{"title": "still works"})
	if created.Task.ID == "" {
		t.Error("task creation through wrapped client failed")
	}
}

func TestToolResultCarriesTextAndStructuredContent(t *testing.T) {
	session := newTestSession(t)

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "list_projects",
		Arguments: map[string]interface{}{},
	})
	if err != nil {
		t.Fatalf("CallTool(list_projects): %v", err)
	}

	if result.IsError {
		t.Fatalf("list_projects failed: %s", contentText(result))
	}

	structured, ok := result.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("structured content = %#v, want an object", result.StructuredContent)
	}

	if _, ok := structured["projects"]; !ok {
		t.Errorf("structured content = %#v, want a projects key", structured)
	}

	// Clients without structured-output support fall back to text content.
	if !strings.Contains(contentText(result), "projects") {
		t.Errorf("text content = %q, want JSON including projects", contentText(result))
	}
}

func toAnySlice(values []string) []interface{} {
	out := make([]interface{}, len(values))
	for i, value := range values {
		out[i] = value
	}

	return out
}
