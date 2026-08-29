package local

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/yousfisaad/lazyarchon/v2/internal/plugin"
)

// newTestClient opens a client over a throwaway database and returns it plus
// a seeded project ID.
func newTestClient(t *testing.T) (*LocalClient, string) {
	t.Helper()

	client, err := newClient(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("newClient() error = %v", err)
	}
	t.Cleanup(func() { client.Close() })

	projects, err := client.ListProjects(context.Background())
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	if len(projects) == 0 {
		t.Fatal("no seeded project found")
	}

	return client, projects[0].ID
}

func strPtr(s string) *string { return &s }

func intPtr(i int) *int { return &i }

func boolPtr(b bool) *bool { return &b }

// ---------- project operations ----------

func TestCreateProject(t *testing.T) {
	client, _ := newTestClient(t)
	ctx := context.Background()

	project, err := client.CreateProject(ctx, plugin.CreateProjectRequest{
		Title:       "Website Relaunch",
		Description: "Marketing site rebuild",
		Color:       "#ff0000",
		Pinned:      true,
	})
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	if project.ID == "" {
		t.Error("expected non-empty project ID")
	}

	if !project.Pinned {
		t.Error("expected pinned project")
	}

	fetched, err := client.GetProject(ctx, project.ID)
	if err != nil {
		t.Fatalf("GetProject() error = %v", err)
	}

	if fetched.Title != "Website Relaunch" || fetched.Color != "#ff0000" {
		t.Errorf("round-trip mismatch: %+v", fetched)
	}
}

func TestCreateProjectValidation(t *testing.T) {
	client, _ := newTestClient(t)

	if _, err := client.CreateProject(context.Background(), plugin.CreateProjectRequest{Title: "  "}); err == nil {
		t.Error("expected error for whitespace-only title")
	}
}

func TestUpdateProject(t *testing.T) {
	client, projectID := newTestClient(t)
	ctx := context.Background()

	updated, err := client.UpdateProject(ctx, projectID, plugin.UpdateProjectRequest{
		Description: strPtr("now with description"),
		Pinned:      boolPtr(true),
	})
	if err != nil {
		t.Fatalf("UpdateProject() error = %v", err)
	}

	if updated.Description != "now with description" || !updated.Pinned {
		t.Errorf("update not applied: %+v", updated)
	}

	if updated.Title != defaultProjectTitle {
		t.Errorf("untouched title changed: %q", updated.Title)
	}
}

func TestUpdateProjectNotFound(t *testing.T) {
	client, _ := newTestClient(t)

	_, err := client.UpdateProject(context.Background(), "missing", plugin.UpdateProjectRequest{
		Title: strPtr("x"),
	})
	if !plugin.IsNotFound(err) {
		t.Errorf("expected not-found error, got %v", err)
	}
}

func TestDeleteProjectCascadesTasks(t *testing.T) {
	client, _ := newTestClient(t)
	ctx := context.Background()

	project, err := client.CreateProject(ctx, plugin.CreateProjectRequest{Title: "Doomed"})
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	if _, err := client.CreateTask(ctx, plugin.CreateTaskRequest{
		ProjectID: project.ID, Title: "task in doomed project",
	}); err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	if err := client.DeleteProject(ctx, project.ID); err != nil {
		t.Fatalf("DeleteProject() error = %v", err)
	}

	result, err := client.ListTasks(ctx, plugin.TaskFilters{ProjectID: &project.ID, IncludeArchived: true})
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}

	if len(result.Tasks) != 0 {
		t.Errorf("tasks survived project deletion: %d left", len(result.Tasks))
	}
}

// ---------- task operations ----------

func TestCreateTaskDefaultsAndRoundTrip(t *testing.T) {
	client, projectID := newTestClient(t)
	ctx := context.Background()

	due := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	created, err := client.CreateTask(ctx, plugin.CreateTaskRequest{
		ProjectID:   projectID,
		Title:       "  Write migration guide  ",
		Description: "body text",
		Tags:        []string{"docs", "release"},
		DueDate:     &due,
		Extra: map[string]interface{}{
			"sources": []interface{}{map[string]interface{}{"url": "https://example.com", "type": "docs"}},
		},
	})
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	if created.Status != plugin.StatusTodo {
		t.Errorf("default status = %q, want todo", created.Status)
	}

	if created.Priority != plugin.PriorityMedium {
		t.Errorf("default priority = %d, want %d", created.Priority, plugin.PriorityMedium)
	}

	if created.Title != "Write migration guide" {
		t.Errorf("title not trimmed: %q", created.Title)
	}

	if created.DueDate == nil || !created.DueDate.Equal(due) {
		t.Errorf("due date not preserved: %v", created.DueDate)
	}

	if len(created.Tags) != 2 {
		t.Errorf("tags = %v, want 2 entries", created.Tags)
	}

	sources, ok := created.Extra["sources"].([]interface{})
	if !ok || len(sources) != 1 {
		t.Errorf("Extra[sources] = %v, want 1 entry", created.Extra["sources"])
	}
}

func TestCreateTaskValidation(t *testing.T) {
	client, projectID := newTestClient(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		request plugin.CreateTaskRequest
	}{
		{name: "empty title", request: plugin.CreateTaskRequest{ProjectID: projectID, Title: ""}},
		{name: "missing project", request: plugin.CreateTaskRequest{Title: "x"}},
		{
			name:    "unknown project",
			request: plugin.CreateTaskRequest{ProjectID: "missing", Title: "x"},
		},
		{
			name:    "bad status",
			request: plugin.CreateTaskRequest{ProjectID: projectID, Title: "x", Status: "blocked"},
		},
		{
			name:    "bad priority",
			request: plugin.CreateTaskRequest{ProjectID: projectID, Title: "x", Priority: 9},
		},
		{
			name:    "unknown parent",
			request: plugin.CreateTaskRequest{ProjectID: projectID, Title: "x", ParentID: strPtr("missing")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := client.CreateTask(ctx, tt.request); err == nil {
				t.Error("expected an error")
			}
		})
	}
}

func TestGetTaskNotFound(t *testing.T) {
	client, _ := newTestClient(t)

	if _, err := client.GetTask(context.Background(), "missing"); !plugin.IsNotFound(err) {
		t.Errorf("expected not-found error, got %v", err)
	}
}

func TestUpdateTaskFields(t *testing.T) {
	client, projectID := newTestClient(t)
	ctx := context.Background()

	created, err := client.CreateTask(ctx, plugin.CreateTaskRequest{ProjectID: projectID, Title: "original"})
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	due := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	duePtr := &due
	updated, err := client.UpdateTask(ctx, created.ID, plugin.UpdateTaskRequest{
		Status:   strPtr(plugin.StatusDoing),
		Priority: intPtr(plugin.PriorityCritical),
		Tags:     &[]string{"urgent"},
		DueDate:  &duePtr,
		Assignee: strPtr("ysaad"),
	})
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}

	if updated.Status != plugin.StatusDoing || updated.Priority != plugin.PriorityCritical {
		t.Errorf("status/priority not applied: %q %d", updated.Status, updated.Priority)
	}

	if len(updated.Tags) != 1 || updated.Tags[0] != "urgent" {
		t.Errorf("tags not applied: %v", updated.Tags)
	}

	if updated.DueDate == nil || !updated.DueDate.Equal(due) {
		t.Errorf("due date not applied: %v", updated.DueDate)
	}

	if updated.Title != "original" {
		t.Errorf("untouched title changed: %q", updated.Title)
	}

	// Clearing the due date: pointer to a nil *time.Time.
	var nilTime *time.Time
	updated, err = client.UpdateTask(ctx, created.ID, plugin.UpdateTaskRequest{DueDate: &nilTime})
	if err != nil {
		t.Fatalf("UpdateTask(clear due) error = %v", err)
	}

	if updated.DueDate != nil {
		t.Errorf("due date not cleared: %v", updated.DueDate)
	}
}

func TestUpdateTaskParentLifecycle(t *testing.T) {
	client, projectID := newTestClient(t)
	ctx := context.Background()

	parent, err := client.CreateTask(ctx, plugin.CreateTaskRequest{ProjectID: projectID, Title: "parent"})
	if err != nil {
		t.Fatalf("CreateTask(parent) error = %v", err)
	}

	child, err := client.CreateTask(ctx, plugin.CreateTaskRequest{ProjectID: projectID, Title: "child"})
	if err != nil {
		t.Fatalf("CreateTask(child) error = %v", err)
	}

	updated, err := client.UpdateTask(ctx, child.ID, plugin.UpdateTaskRequest{ParentID: &parent.ID})
	if err != nil {
		t.Fatalf("UpdateTask(set parent) error = %v", err)
	}

	if updated.ParentID == nil || *updated.ParentID != parent.ID {
		t.Errorf("parent not set: %v", updated.ParentID)
	}

	if _, err := client.UpdateTask(ctx, parent.ID, plugin.UpdateTaskRequest{ParentID: &child.ID}); err == nil {
		t.Error("expected cycle error when making parent a child of its child")
	}

	// Clearing the parent: pointer to empty string.
	updated, err = client.UpdateTask(ctx, child.ID, plugin.UpdateTaskRequest{ParentID: strPtr("")})
	if err != nil {
		t.Fatalf("UpdateTask(clear parent) error = %v", err)
	}

	if updated.ParentID != nil {
		t.Errorf("parent not cleared: %v", updated.ParentID)
	}
}

func TestArchiveFlow(t *testing.T) {
	client, projectID := newTestClient(t)
	ctx := context.Background()

	created, err := client.CreateTask(ctx, plugin.CreateTaskRequest{ProjectID: projectID, Title: "to archive"})
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	updated, err := client.UpdateTask(ctx, created.ID, plugin.UpdateTaskRequest{Archived: boolPtr(true)})
	if err != nil {
		t.Fatalf("UpdateTask(archive) error = %v", err)
	}

	if !updated.Archived || updated.ArchivedAt == nil {
		t.Errorf("archive not applied: %+v", updated)
	}

	// Default listing hides archived tasks.
	visible, err := client.ListTasks(ctx, plugin.TaskFilters{})
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}

	for _, task := range visible.Tasks {
		if task.ID == created.ID {
			t.Error("archived task still visible without IncludeArchived")
		}
	}

	// IncludeArchived shows it.
	all, err := client.ListTasks(ctx, plugin.TaskFilters{IncludeArchived: true})
	if err != nil {
		t.Fatalf("ListTasks(includeArchived) error = %v", err)
	}

	found := false
	for _, task := range all.Tasks {
		if task.ID == created.ID {
			found = true
		}
	}

	if !found {
		t.Error("archived task missing with IncludeArchived")
	}

	// Unarchive clears the timestamp.
	updated, err = client.UpdateTask(ctx, created.ID, plugin.UpdateTaskRequest{Archived: boolPtr(false)})
	if err != nil {
		t.Fatalf("UpdateTask(unarchive) error = %v", err)
	}

	if updated.Archived || updated.ArchivedAt != nil {
		t.Errorf("unarchive not applied: %+v", updated)
	}
}

func TestDeleteTaskOrphansChildren(t *testing.T) {
	client, projectID := newTestClient(t)
	ctx := context.Background()

	parent, err := client.CreateTask(ctx, plugin.CreateTaskRequest{ProjectID: projectID, Title: "parent"})
	if err != nil {
		t.Fatalf("CreateTask(parent) error = %v", err)
	}

	child, err := client.CreateTask(ctx, plugin.CreateTaskRequest{
		ProjectID: projectID, Title: "child", ParentID: &parent.ID,
	})
	if err != nil {
		t.Fatalf("CreateTask(child) error = %v", err)
	}

	if err := client.DeleteTask(ctx, parent.ID); err != nil {
		t.Fatalf("DeleteTask() error = %v", err)
	}

	survivor, err := client.GetTask(ctx, child.ID)
	if err != nil {
		t.Fatalf("GetTask(child) error = %v", err)
	}

	if survivor.ParentID != nil {
		t.Errorf("child parent not cleared on parent delete: %v", survivor.ParentID)
	}
}

// ---------- filters ----------

func seedFilterFixture(t *testing.T, client *LocalClient, projectID string) (todoID, doneID string) {
	t.Helper()
	ctx := context.Background()

	todo, err := client.CreateTask(ctx, plugin.CreateTaskRequest{
		ProjectID: projectID, Title: "Searchable alpha task", Status: plugin.StatusTodo,
	})
	if err != nil {
		t.Fatalf("CreateTask(todo) error = %v", err)
	}

	done, err := client.CreateTask(ctx, plugin.CreateTaskRequest{
		ProjectID: projectID, Title: "beta task", Status: plugin.StatusDone, Tags: []string{"infra"},
	})
	if err != nil {
		t.Fatalf("CreateTask(done) error = %v", err)
	}

	// The priority-1 fixture matters for filters, but its ID is not needed.
	if _, err := client.CreateTask(ctx, plugin.CreateTaskRequest{
		ProjectID: projectID, Title: "gamma", Status: plugin.StatusDoing, Priority: plugin.PriorityCritical,
	}); err != nil {
		t.Fatalf("CreateTask(urgent) error = %v", err)
	}

	return todo.ID, done.ID
}

func TestListTasksFilters(t *testing.T) {
	client, projectID := newTestClient(t)
	ctx := context.Background()

	todoID, doneID := seedFilterFixture(t, client, projectID)

	tests := []struct {
		name    string
		filter  plugin.TaskFilters
		wantIDs []string
	}{
		{
			name:    "by status",
			filter:  plugin.TaskFilters{Status: strPtr(plugin.StatusDone)},
			wantIDs: []string{doneID},
		},
		{
			name:    "by search",
			filter:  plugin.TaskFilters{SearchQuery: strPtr("alpha")},
			wantIDs: []string{todoID},
		},
		{
			name:    "by tag",
			filter:  plugin.TaskFilters{Tags: []string{"infra"}},
			wantIDs: []string{doneID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.ListTasks(ctx, tt.filter)
			if err != nil {
				t.Fatalf("ListTasks() error = %v", err)
			}

			if len(result.Tasks) != len(tt.wantIDs) {
				t.Fatalf("got %d tasks, want %d", len(result.Tasks), len(tt.wantIDs))
			}

			for i, want := range tt.wantIDs {
				if result.Tasks[i].ID != want {
					t.Errorf("task[%d].ID = %s, want %s", i, result.Tasks[i].ID, want)
				}
			}
		})
	}
}

func TestListTasksStatusOrdering(t *testing.T) {
	client, projectID := newTestClient(t)
	ctx := context.Background()

	seedFilterFixture(t, client, projectID)

	result, err := client.ListTasks(ctx, plugin.TaskFilters{})
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}

	// Expect workflow order: todo, then doing, then done.
	wantStatuses := []string{plugin.StatusTodo, plugin.StatusDoing, plugin.StatusDone}
	for i, want := range wantStatuses {
		if result.Tasks[i].Status != want {
			t.Errorf("position %d status = %q, want %q", i, result.Tasks[i].Status, want)
		}
	}
}

func TestListTasksPagination(t *testing.T) {
	client, projectID := newTestClient(t)
	ctx := context.Background()

	for range 5 {
		if _, err := client.CreateTask(ctx, plugin.CreateTaskRequest{
			ProjectID: projectID, Title: "task", Priority: plugin.PriorityLow,
		}); err != nil {
			t.Fatalf("CreateTask() error = %v", err)
		}
	}

	page1, err := client.ListTasks(ctx, plugin.TaskFilters{Limit: 2, Offset: 0})
	if err != nil {
		t.Fatalf("ListTasks(page1) error = %v", err)
	}

	if len(page1.Tasks) != 2 || page1.TotalCount != 5 || !page1.HasMore || page1.NextOffset != 2 {
		t.Errorf("page1 = len:%d total:%d hasMore:%v next:%d", len(page1.Tasks), page1.TotalCount, page1.HasMore, page1.NextOffset)
	}

	page3, err := client.ListTasks(ctx, plugin.TaskFilters{Limit: 2, Offset: 4})
	if err != nil {
		t.Fatalf("ListTasks(page3) error = %v", err)
	}

	if len(page3.Tasks) != 1 || page3.HasMore {
		t.Errorf("page3 = len:%d hasMore:%v, want len:1 hasMore:false", len(page3.Tasks), page3.HasMore)
	}
}

func TestExtraReplaceSemantics(t *testing.T) {
	client, projectID := newTestClient(t)
	ctx := context.Background()

	created, err := client.CreateTask(ctx, plugin.CreateTaskRequest{
		ProjectID: projectID,
		Title:     "with extra",
		Extra: map[string]interface{}{
			"github_repo": "owner/repo",
			"sources":     []interface{}{map[string]interface{}{"url": "https://a.example"}},
		},
	})
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	updated, err := client.UpdateTask(ctx, created.ID, plugin.UpdateTaskRequest{
		Extra: &map[string]interface{}{
			"github_repo": "other/repo",
			"sources":     []interface{}{},
		},
	})
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}

	if updated.Extra["github_repo"] != "other/repo" {
		t.Errorf("Extra[github_repo] = %v, want other/repo (replace semantics)", updated.Extra["github_repo"])
	}

	if sources, ok := updated.Extra["sources"].([]interface{}); !ok || len(sources) != 0 {
		t.Errorf("Extra[sources] = %v, want empty array", updated.Extra["sources"])
	}
}

func TestConcurrentClientsOnSameFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "shared.db")

	clientA, err := newClient(path)
	if err != nil {
		t.Fatalf("newClient(A) error = %v", err)
	}
	defer clientA.Close()

	clientB, err := newClient(path)
	if err != nil {
		t.Fatalf("newClient(B) error = %v", err)
	}
	defer clientB.Close()

	ctx := context.Background()
	project, err := clientA.CreateProject(ctx, plugin.CreateProjectRequest{Title: "shared"})
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	const writers = 4
	const perWriter = 10

	var wg sync.WaitGroup
	errs := make(chan error, writers*perWriter)

	for w := range writers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range perWriter {
				client := clientA
				if w%2 == 1 {
					client = clientB
				}
				if _, err := client.CreateTask(ctx, plugin.CreateTaskRequest{
					ProjectID: project.ID,
					Title:     "concurrent task",
				}); err != nil {
					errs <- err
				}
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent CreateTask error: %v", err)
	}

	result, err := clientB.ListTasks(ctx, plugin.TaskFilters{ProjectID: &project.ID})
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}

	if result.TotalCount != writers*perWriter {
		t.Errorf("total = %d, want %d", result.TotalCount, writers*perWriter)
	}
}

func TestCreateTaskViaPluginFactory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "factory.db")

	p := &LocalPlugin{}
	client, err := p.CreateClient(plugin.PluginConfig{
		Extra: map[string]interface{}{"path": path},
	})
	if err != nil {
		t.Fatalf("CreateClient() error = %v", err)
	}
	defer client.Close()

	if err := client.HealthCheck(context.Background()); err != nil {
		t.Errorf("HealthCheck() error = %v", err)
	}

	projects, err := client.ListProjects(context.Background())
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}

	if len(projects) != 1 || projects[0].Title != defaultProjectTitle {
		t.Errorf("expected seeded Inbox project, got %+v", projects)
	}
}

func TestSearchEscapesWildcards(t *testing.T) {
	client, projectID := newTestClient(t)
	ctx := context.Background()

	// One title contains literal LIKE wildcards, one does not.
	if _, err := client.CreateTask(ctx, plugin.CreateTaskRequest{
		ProjectID: projectID, Title: "100% done_share",
	}); err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	if _, err := client.CreateTask(ctx, plugin.CreateTaskRequest{
		ProjectID: projectID, Title: "plain title",
	}); err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// A query of "%" must match only the title containing a literal percent
	// sign, not act as a match-all wildcard.
	result, err := client.ListTasks(ctx, plugin.TaskFilters{SearchQuery: strPtr("%")})
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}

	if len(result.Tasks) != 1 || !strings.Contains(result.Tasks[0].Title, "%") {
		t.Errorf("LIKE wildcard not escaped, matched: %+v", result.Tasks)
	}

	// Same for underscore (matches any single character unescaped).
	result, err = client.ListTasks(ctx, plugin.TaskFilters{SearchQuery: strPtr("done_share")})
	if err != nil {
		t.Fatalf("ListTasks(underscore) error = %v", err)
	}

	if len(result.Tasks) != 1 {
		t.Errorf("expected exact substring match, matched: %+v", result.Tasks)
	}
}
