package local

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yousfisaad/lazyarchon/v2/internal/plugin"
)

// newEnforcingClient opens a client with dependency enforcement on, plus a
// seeded project ID.
func newEnforcingClient(t *testing.T) (*LocalClient, string) {
	t.Helper()

	client, err := newClientWithEnforcement(filepath.Join(t.TempDir(), "test.db"), true)
	if err != nil {
		t.Fatalf("newClientWithEnforcement() error = %v", err)
	}
	t.Cleanup(func() { client.Close() })

	projects, err := client.ListProjects(context.Background())
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}

	return client, projects[0].ID
}

// seedPair creates a blocker task and a dependent task blocked by it.
func seedPair(t *testing.T, client *LocalClient, projectID string) (blocker, dependent *plugin.Task) {
	t.Helper()
	ctx := context.Background()

	blocker, err := client.CreateTask(ctx, plugin.CreateTaskRequest{
		ProjectID: projectID,
		Title:     "Fix auth",
	})
	if err != nil {
		t.Fatalf("CreateTask(blocker) error = %v", err)
	}

	dependent, err = client.CreateTask(ctx, plugin.CreateTaskRequest{
		ProjectID: projectID,
		Title:     "Ship release",
		BlockedBy: []string{blocker.ID},
	})
	if err != nil {
		t.Fatalf("CreateTask(dependent) error = %v", err)
	}

	return blocker, dependent
}

func TestCreateTaskPersistsDependencies(t *testing.T) {
	client, projectID := newTestClient(t)
	ctx := context.Background()

	blocker, dependent := seedPair(t, client, projectID)

	if len(dependent.BlockedBy) != 1 || dependent.BlockedBy[0] != blocker.ID {
		t.Errorf("dependent.BlockedBy = %v, want [%s]", dependent.BlockedBy, blocker.ID)
	}

	if len(dependent.Blocks) != 0 {
		t.Errorf("dependent.Blocks = %v, want empty on a fresh task", dependent.Blocks)
	}

	// The reverse edge is visible on the blocker.
	fetched, err := client.GetTask(ctx, blocker.ID)
	if err != nil {
		t.Fatalf("GetTask(blocker) error = %v", err)
	}

	if len(fetched.Blocks) != 1 || fetched.Blocks[0] != dependent.ID {
		t.Errorf("blocker.Blocks = %v, want [%s]", fetched.Blocks, dependent.ID)
	}
}

func TestCreateTaskDependencyValidation(t *testing.T) {
	client, projectID := newTestClient(t)
	ctx := context.Background()

	blocker, _ := seedPair(t, client, projectID)

	// Unknown blocker IDs are not-found, mirroring the parent-task wording.
	if _, err := client.CreateTask(ctx, plugin.CreateTaskRequest{
		ProjectID: projectID,
		Title:     "bad refs",
		BlockedBy: []string{blocker.ID, "missing"},
	}); !plugin.IsNotFound(err) {
		t.Errorf("unknown blocker error = %v, want not-found", err)
	}

	// Duplicates collapse to one edge rather than tripping the PK.
	task, err := client.CreateTask(ctx, plugin.CreateTaskRequest{
		ProjectID: projectID,
		Title:     "duplicated refs",
		BlockedBy: []string{blocker.ID, blocker.ID},
	})
	if err != nil {
		t.Fatalf("CreateTask(duplicates) error = %v", err)
	}

	if len(task.BlockedBy) != 1 || task.BlockedBy[0] != blocker.ID {
		t.Errorf("task.BlockedBy = %v, want one entry", task.BlockedBy)
	}
}

func TestUpdateTaskDependencyValidation(t *testing.T) {
	client, projectID := newTestClient(t)
	ctx := context.Background()

	blocker, dependent := seedPair(t, client, projectID)

	// Self-dependency gets the friendly error (the CHECK is only a backstop).
	if _, err := client.UpdateTask(ctx, dependent.ID, plugin.UpdateTaskRequest{
		BlockedBy: &[]string{dependent.ID},
	}); err == nil || !strings.Contains(err.Error(), "cannot depend on itself") {
		t.Errorf("self-dependency error = %v, want friendly message", err)
	}

	if _, err := client.UpdateTask(ctx, dependent.ID, plugin.UpdateTaskRequest{
		BlockedBy: &[]string{"missing"},
	}); !plugin.IsNotFound(err) {
		t.Errorf("unknown blocker error = %v, want not-found", err)
	}

	// A two-node cycle is rejected.
	if _, err := client.UpdateTask(ctx, blocker.ID, plugin.UpdateTaskRequest{
		BlockedBy: &[]string{dependent.ID},
	}); err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Errorf("cycle error = %v, want cycle message", err)
	}

	// Longer cycles are caught too: storing C -> A and then proposing
	// B -> C closes A -> B, C -> A, B -> C into a loop.
	third, err := client.CreateTask(ctx, plugin.CreateTaskRequest{
		ProjectID: projectID,
		Title:     "third task",
		BlockedBy: []string{dependent.ID},
	})
	if err != nil {
		t.Fatalf("CreateTask(third) error = %v", err)
	}

	if _, err := client.UpdateTask(ctx, blocker.ID, plugin.UpdateTaskRequest{
		BlockedBy: &[]string{third.ID},
	}); err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Errorf("three-node cycle error = %v, want cycle message", err)
	}

	// Replacing with a non-cyclic set works, and the old edge is gone.
	fourth, err := client.CreateTask(ctx, plugin.CreateTaskRequest{
		ProjectID: projectID,
		Title:     "fourth task",
	})
	if err != nil {
		t.Fatalf("CreateTask(fourth) error = %v", err)
	}

	updated, err := client.UpdateTask(ctx, dependent.ID, plugin.UpdateTaskRequest{
		BlockedBy: &[]string{fourth.ID},
	})
	if err != nil {
		t.Fatalf("UpdateTask(replace deps) error = %v", err)
	}

	if len(updated.BlockedBy) != 1 || updated.BlockedBy[0] != fourth.ID {
		t.Errorf("updated.BlockedBy = %v, want [%s]", updated.BlockedBy, fourth.ID)
	}

	oldBlocker, err := client.GetTask(ctx, blocker.ID)
	if err != nil {
		t.Fatalf("GetTask(blocker) error = %v", err)
	}

	// Nobody waits on the old blocker anymore: the dependent moved to the
	// fourth task, and third is blocked by the dependent, not by it.
	if len(oldBlocker.Blocks) != 0 {
		t.Errorf("blocker.Blocks after replace = %v, want empty", oldBlocker.Blocks)
	}

	newBlocker, err := client.GetTask(ctx, fourth.ID)
	if err != nil {
		t.Fatalf("GetTask(fourth) error = %v", err)
	}

	if len(newBlocker.Blocks) != 1 || newBlocker.Blocks[0] != dependent.ID {
		t.Errorf("fourth.Blocks after replace = %v, want [%s]", newBlocker.Blocks, dependent.ID)
	}
}

func TestEnforcementAdvisoryByDefault(t *testing.T) {
	client, projectID := newTestClient(t)
	ctx := context.Background()

	_, dependent := seedPair(t, client, projectID)

	// Advisory mode (the default): moving to doing succeeds despite the
	// unfinished blocker.
	doing := plugin.StatusDoing
	updated, err := client.UpdateTask(ctx, dependent.ID, plugin.UpdateTaskRequest{Status: &doing})
	if err != nil {
		t.Fatalf("UpdateTask(doing) in advisory mode error = %v", err)
	}

	if updated.Status != plugin.StatusDoing {
		t.Errorf("updated.Status = %q, want doing", updated.Status)
	}
}

func TestEnforcementRejectsBlockedTransitions(t *testing.T) {
	client, projectID := newEnforcingClient(t)
	ctx := context.Background()

	blocker, dependent := seedPair(t, client, projectID)

	doing := plugin.StatusDoing
	if _, err := client.UpdateTask(ctx, dependent.ID, plugin.UpdateTaskRequest{Status: &doing}); err == nil ||
		!strings.Contains(err.Error(), "task is blocked by unfinished tasks: Fix auth (") {
		t.Errorf("doing while blocked error = %v, want rejection naming blocker", err)
	}

	review := plugin.StatusReview
	if _, err := client.UpdateTask(ctx, dependent.ID, plugin.UpdateTaskRequest{Status: &review}); err == nil ||
		!strings.Contains(err.Error(), "task is blocked by unfinished tasks") {
		t.Errorf("review while blocked error = %v, want rejection", err)
	}

	// Creating straight into doing with blockers is gated the same way.
	if _, err := client.CreateTask(ctx, plugin.CreateTaskRequest{
		ProjectID: projectID,
		Title:     "born doing",
		Status:    plugin.StatusDoing,
		BlockedBy: []string{blocker.ID},
	}); err == nil || !strings.Contains(err.Error(), "task is blocked by unfinished tasks") {
		t.Errorf("create-with-doing error = %v, want rejection", err)
	}

	// todo and done remain reachable.
	todo := plugin.StatusTodo
	if _, err := client.UpdateTask(ctx, dependent.ID, plugin.UpdateTaskRequest{Status: &todo}); err != nil {
		t.Fatalf("todo while blocked error = %v, want allowed", err)
	}

	done := plugin.StatusDone
	if _, err := client.UpdateTask(ctx, dependent.ID, plugin.UpdateTaskRequest{Status: &done}); err != nil {
		t.Fatalf("done while blocked error = %v, want allowed", err)
	}

	// Finishing the blocker unblocks the dependent.
	if _, err := client.UpdateTask(ctx, blocker.ID, plugin.UpdateTaskRequest{Status: &done}); err != nil {
		t.Fatalf("finishing blocker error = %v", err)
	}

	if _, err := client.UpdateTask(ctx, dependent.ID, plugin.UpdateTaskRequest{Status: &doing}); err != nil {
		t.Fatalf("doing after unblock error = %v, want allowed", err)
	}
}

func TestEnforcementJudgesIncomingDependencySet(t *testing.T) {
	client, projectID := newEnforcingClient(t)
	ctx := context.Background()

	blocker, dependent := seedPair(t, client, projectID)

	doing := plugin.StatusDoing

	// A single request that clears the deps and moves to doing is judged on
	// the incoming set: allowed.
	empty := []string{}
	updated, err := client.UpdateTask(ctx, dependent.ID, plugin.UpdateTaskRequest{
		Status:    &doing,
		BlockedBy: &empty,
	})
	if err != nil {
		t.Fatalf("clearing deps + doing error = %v, want allowed", err)
	}

	if updated.Status != plugin.StatusDoing || len(updated.BlockedBy) != 0 {
		t.Errorf("updated = %q/%v, want doing with no blockers", updated.Status, updated.BlockedBy)
	}

	// The mirror image: re-adding a blocker while starting work is rejected
	// on the incoming set, even though the stored set was empty.
	blocked := []string{blocker.ID}
	if _, err := client.UpdateTask(ctx, dependent.ID, plugin.UpdateTaskRequest{
		Status:    &doing,
		BlockedBy: &blocked,
	}); err == nil || !strings.Contains(err.Error(), "task is blocked by unfinished tasks") {
		t.Errorf("adding deps + doing error = %v, want rejection", err)
	}
}
