package mcpserver

import (
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/yousfisaad/lazyarchon/v2/internal/plugin"
	localplugin "github.com/yousfisaad/lazyarchon/v2/internal/plugins/local"
)

// newLocalClient opens a local client over a throwaway database, for tests
// that drive several sessions against the same data.
func newLocalClient(t *testing.T) plugin.TaskClient {
	t.Helper()

	client, err := (&localplugin.LocalPlugin{}).CreateClient(plugin.PluginConfig{
		Extra: map[string]interface{}{"path": filepath.Join(t.TempDir(), "test.db")},
	})
	if err != nil {
		t.Fatalf("creating local client: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	return client
}

func TestCreateTaskAutoStampsWorktree(t *testing.T) {
	session := connectSessionInWorktree(t, newLocalClient(t), "feature-x")

	out := callTool[createTaskOut](t, session, "create_task", map[string]interface{}{
		"title": "stamped on create",
	})

	if out.Task.Worktree != "feature-x" {
		t.Errorf("created task Worktree = %q, want feature-x", out.Task.Worktree)
	}

	list := callTool[listTasksOut](t, session, "list_tasks", map[string]interface{}{})

	if len(list.Tasks) != 1 || list.Tasks[0].Worktree != "feature-x" {
		t.Errorf("list summary worktree = %+v, want feature-x on the only task", list.Tasks)
	}
}

func TestCreateTaskExplicitWorktreeOverrides(t *testing.T) {
	session := connectSessionInWorktree(t, newLocalClient(t), "feature-x")

	out := callTool[createTaskOut](t, session, "create_task", map[string]interface{}{
		"title": "manual attribution", "worktree": "manual-wt",
	})

	if out.Task.Worktree != "manual-wt" {
		t.Errorf("created task Worktree = %q, want manual-wt", out.Task.Worktree)
	}
}

func TestUpdateTaskRestampsOnDoingAndReview(t *testing.T) {
	session := connectSessionInWorktree(t, newLocalClient(t), "wt-a")

	created := callTool[createTaskOut](t, session, "create_task", map[string]interface{}{
		"title": "restamp me", "worktree": "wt-old",
	})

	for _, status := range []string{"doing", "review"} {
		out := callTool[updateTaskOut](t, session, "update_task", map[string]interface{}{
			"task_id": created.Task.ID, "status": status,
		})

		if out.Task.Worktree != "wt-a" {
			t.Errorf("after status=%s Worktree = %q, want wt-a", status, out.Task.Worktree)
		}
	}
}

func TestUpdateTaskPlainEditsDoNotStealWorktree(t *testing.T) {
	client := newLocalClient(t)
	plain := connectSession(t, client)
	wtSession := connectSessionInWorktree(t, client, "wt-a")

	created := callTool[createTaskOut](t, plain, "create_task", map[string]interface{}{
		"title": "owned elsewhere", "worktree": "wt-old",
	})

	scenarios := []struct {
		name    string
		session *mcp.ClientSession
		args    map[string]interface{}
	}{
		{
			name:    "title edit from another worktree",
			session: wtSession,
			args:    map[string]interface{}{"task_id": created.Task.ID, "title": "renamed"},
		},
		{
			name:    "status done from another worktree",
			session: wtSession,
			args:    map[string]interface{}{"task_id": created.Task.ID, "status": "done"},
		},
		{
			name:    "status todo from another worktree",
			session: wtSession,
			args:    map[string]interface{}{"task_id": created.Task.ID, "status": "todo"},
		},
		{
			name:    "status doing from a main-checkout session",
			session: plain,
			args:    map[string]interface{}{"task_id": created.Task.ID, "status": "doing"},
		},
		{
			name:    "status review from a main-checkout session",
			session: plain,
			args:    map[string]interface{}{"task_id": created.Task.ID, "status": "review"},
		},
	}

	for _, tt := range scenarios {
		t.Run(tt.name, func(t *testing.T) {
			out := callTool[updateTaskOut](t, tt.session, "update_task", tt.args)

			if out.Task.Worktree != "wt-old" {
				t.Errorf("Worktree = %q, want wt-old (ownership must not change)", out.Task.Worktree)
			}
		})
	}
}

func TestUpdateTaskExplicitWorktreeSetAndClear(t *testing.T) {
	session := connectSession(t, newLocalClient(t))

	created := callTool[createTaskOut](t, session, "create_task", map[string]interface{}{
		"title": "manual lifecycle",
	})

	if created.Task.Worktree != "" {
		t.Fatalf("created task Worktree = %q, want empty (main-checkout session)", created.Task.Worktree)
	}

	set := callTool[updateTaskOut](t, session, "update_task", map[string]interface{}{
		"task_id": created.Task.ID, "worktree": "manual-wt",
	})

	if set.Task.Worktree != "manual-wt" {
		t.Errorf("after explicit set Worktree = %q, want manual-wt", set.Task.Worktree)
	}

	cleared := callTool[updateTaskOut](t, session, "update_task", map[string]interface{}{
		"task_id": created.Task.ID, "worktree": "",
	})

	if cleared.Task.Worktree != "" {
		t.Errorf("after explicit clear Worktree = %q, want empty", cleared.Task.Worktree)
	}
}

func TestListTasksWorktreeFilter(t *testing.T) {
	client := newLocalClient(t)
	wtSession := connectSessionInWorktree(t, client, "wt-a")
	plain := connectSession(t, client)

	callTool[createTaskOut](t, wtSession, "create_task", map[string]interface{}{
		"title": "in wt-a",
	})

	callTool[createTaskOut](t, plain, "create_task", map[string]interface{}{
		"title": "in wt-b", "worktree": "wt-b",
	})

	callTool[createTaskOut](t, plain, "create_task", map[string]interface{}{
		"title": "unstamped",
	})

	filtered := callTool[listTasksOut](t, plain, "list_tasks", map[string]interface{}{
		"worktree": "wt-a",
	})

	if len(filtered.Tasks) != 1 || filtered.Tasks[0].Worktree != "wt-a" || filtered.Total != 1 {
		t.Errorf("filtered = %+v (total %d), want the single wt-a task", filtered.Tasks, filtered.Total)
	}

	all := callTool[listTasksOut](t, plain, "list_tasks", map[string]interface{}{})

	if all.Total != 3 {
		t.Errorf("unfiltered total = %d, want 3", all.Total)
	}
}
