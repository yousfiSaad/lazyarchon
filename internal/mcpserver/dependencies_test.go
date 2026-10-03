package mcpserver

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/yousfisaad/lazyarchon/v2/internal/plugin"
	localplugin "github.com/yousfisaad/lazyarchon/v2/internal/plugins/local"
)

// newEnforcingSession is newTestSession with dependency enforcement on.
func newEnforcingSession(t *testing.T) *mcp.ClientSession {
	t.Helper()

	client, err := (&localplugin.LocalPlugin{}).CreateClient(plugin.PluginConfig{
		Extra: map[string]interface{}{
			"path":                 filepath.Join(t.TempDir(), "test.db"),
			"enforce_dependencies": true,
		},
	})
	if err != nil {
		t.Fatalf("creating local client: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	return connectSession(t, client)
}

func TestCreateTaskEchoesDependencies(t *testing.T) {
	session := newTestSession(t)

	blocker := callTool[createTaskOut](t, session, "create_task", map[string]interface{}{
		"title": "Fix auth",
	})

	created := callTool[createTaskOut](t, session, "create_task", map[string]interface{}{
		"title":      "Ship release",
		"blocked_by": []string{blocker.Task.ID},
	})

	if len(created.Task.BlockedBy) != 1 || created.Task.BlockedBy[0] != blocker.Task.ID {
		t.Errorf("created blocked_by = %v, want [%s]", created.Task.BlockedBy, blocker.Task.ID)
	}

	// The reverse edge shows up on the blocker's detail.
	fetched := callTool[getTaskOut](t, session, "get_task", map[string]interface{}{
		"task_id": blocker.Task.ID,
	})

	if len(fetched.Task.Blocks) != 1 || fetched.Task.Blocks[0] != created.Task.ID {
		t.Errorf("blocker blocks = %v, want [%s]", fetched.Task.Blocks, created.Task.ID)
	}

	// An empty array clears the set.
	updated := callTool[updateTaskOut](t, session, "update_task", map[string]interface{}{
		"task_id":    created.Task.ID,
		"blocked_by": []string{},
	})

	if len(updated.Task.BlockedBy) != 0 {
		t.Errorf("updated blocked_by after clear = %v, want empty", updated.Task.BlockedBy)
	}

	// Omitting the field leaves the (re-set) dependencies alone.
	reSet := callTool[updateTaskOut](t, session, "update_task", map[string]interface{}{
		"task_id": created.Task.ID,
		"status":  "todo",
		"blocked_by": []string{
			blocker.Task.ID,
		},
	})
	if len(reSet.Task.BlockedBy) != 1 {
		t.Fatalf("re-set blocked_by = %v, want one entry", reSet.Task.BlockedBy)
	}

	untouched := callTool[updateTaskOut](t, session, "update_task", map[string]interface{}{
		"task_id": created.Task.ID,
		"title":   "Ship release v2",
	})

	if len(untouched.Task.BlockedBy) != 1 || untouched.Task.Title != "Ship release v2" {
		t.Errorf("untouched = %q/%v, want title change with dependencies intact", untouched.Task.Title, untouched.Task.BlockedBy)
	}
}

func TestListTasksReadyFilter(t *testing.T) {
	session := newTestSession(t)

	blocker := callTool[createTaskOut](t, session, "create_task", map[string]interface{}{
		"title": "Fix auth",
	})
	callTool[createTaskOut](t, session, "create_task", map[string]interface{}{
		"title":      "Ship release",
		"blocked_by": []string{blocker.Task.ID},
	})

	blocked := callTool[listTasksOut](t, session, "list_tasks", map[string]interface{}{
		"ready": true,
	})

	if blocked.Total != 1 || blocked.Tasks[0].Title != "Fix auth" {
		t.Errorf("ready list = %d tasks (first %q), want only the unblocked task", blocked.Total, blocked.Tasks[0].Title)
	}

	// Finishing the blocker makes the dependent ready.
	callTool[updateTaskOut](t, session, "update_task", map[string]interface{}{
		"task_id": blocker.Task.ID,
		"status":  "done",
	})

	unblocked := callTool[listTasksOut](t, session, "list_tasks", map[string]interface{}{
		"ready": true,
	})

	if unblocked.Total != 2 {
		t.Errorf("ready list after unblock = %d tasks, want 2", unblocked.Total)
	}
}

func TestEnforcedDependenciesRejectStatusMove(t *testing.T) {
	session := newEnforcingSession(t)

	blocker := callTool[createTaskOut](t, session, "create_task", map[string]interface{}{
		"title": "Fix auth",
	})
	dependent := callTool[createTaskOut](t, session, "create_task", map[string]interface{}{
		"title":      "Ship release",
		"blocked_by": []string{blocker.Task.ID},
	})

	// The rejection surfaces as a tool error naming the blocker.
	message := callToolError(t, session, "update_task", map[string]interface{}{
		"task_id": dependent.Task.ID,
		"status":  "doing",
	})

	if !strings.Contains(message, "task is blocked by unfinished tasks: Fix auth (") {
		t.Errorf("enforcement error = %q, want blockers named", message)
	}

	// Advisory data is still queryable: ready=true lists only the blocker.
	ready := callTool[listTasksOut](t, session, "list_tasks", map[string]interface{}{
		"ready": true,
	})

	if ready.Total != 1 || ready.Tasks[0].Title != "Fix auth" {
		t.Errorf("ready list under enforcement = %d tasks, want only the blocker", ready.Total)
	}
}
