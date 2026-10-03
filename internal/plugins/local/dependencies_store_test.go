package local

import (
	"context"
	"database/sql"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/yousfisaad/lazyarchon/v2/internal/plugin"
)

// seedStoreTask inserts a minimal todo task directly through the store.
func seedStoreTask(t *testing.T, st *store, id, title string) {
	t.Helper()

	var projectID string
	if err := st.db.QueryRow("SELECT id FROM projects LIMIT 1").Scan(&projectID); err != nil {
		t.Fatalf("reading seeded project: %v", err)
	}

	now := time.Now()
	task := &plugin.Task{
		ID:        id,
		ProjectID: projectID,
		Title:     title,
		Status:    plugin.StatusTodo,
		Priority:  plugin.PriorityMedium,
		Tags:      []string{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := st.insertTask(context.Background(), task); err != nil {
		t.Fatalf("insertTask(%s) error = %v", id, err)
	}
}

// setBlockedBy replaces a task's blocker set through the store update path.
func setBlockedBy(t *testing.T, st *store, taskID string, ids []string) {
	t.Helper()

	if _, err := st.updateTask(context.Background(), taskID, plugin.UpdateTaskRequest{BlockedBy: &ids}); err != nil {
		t.Fatalf("updateTask(%s, BlockedBy=%v) error = %v", taskID, ids, err)
	}
}

// listTitles returns the titles of a plain listTasks call, in page order.
func listTitles(t *testing.T, st *store, filters plugin.TaskFilters) []string {
	t.Helper()

	result, err := st.listTasks(context.Background(), filters)
	if err != nil {
		t.Fatalf("listTasks(%+v) error = %v", filters, err)
	}

	titles := make([]string, 0, len(result.Tasks))
	for _, task := range result.Tasks {
		titles = append(titles, task.Title)
	}

	return titles
}

func TestMigrateAddsDependenciesTable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")

	// Build a version-3 database by hand: all three applied migrations, one
	// surviving task row in the post-widening shape.
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path))
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}

	for _, migration := range []string{migrationV1, migrationV2, migrationV3} {
		if _, err := db.Exec(migration); err != nil {
			t.Fatalf("applying migration: %v", err)
		}
	}
	if _, err := db.Exec("PRAGMA user_version = 3"); err != nil {
		t.Fatalf("setting user_version: %v", err)
	}

	now := formatTime(time.Now())
	if _, err := db.Exec("INSERT INTO projects (id, title, created_at, updated_at) VALUES ('p1', 'Legacy', ?, ?)", now, now); err != nil {
		t.Fatalf("seeding project: %v", err)
	}
	if _, err := db.Exec("INSERT INTO tasks (id, project_id, title, priority, worktree, created_at, updated_at) VALUES ('t1', 'p1', 'legacy task', 5, 'wt-a', ?, ?)", now, now); err != nil {
		t.Fatalf("seeding task: %v", err)
	}
	db.Close()

	st, err := openStore(path)
	if err != nil {
		t.Fatalf("openStore() on v3 database: %v", err)
	}
	defer st.Close()

	var version int
	if err := st.db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		t.Fatalf("reading user_version: %v", err)
	}

	if version != schemaVersion() {
		t.Errorf("user_version = %d, want %d", version, schemaVersion())
	}

	task, err := st.getTask(context.Background(), "t1")
	if err != nil {
		t.Fatalf("getTask() on migrated row: %v", err)
	}

	if task.Title != "legacy task" || task.Priority != 5 || task.Worktree != "wt-a" {
		t.Errorf("migrated task = %q/%d/%q, want legacy task/5/wt-a", task.Title, task.Priority, task.Worktree)
	}

	// The new table exists and enforces its shape: a self-dependency is
	// unrepresentable, and a real edge round-trips.
	if _, err := st.db.Exec("INSERT INTO task_dependencies (task_id, depends_on_id, created_at) VALUES ('t1', 't1', ?)", now); err == nil {
		t.Error("self-dependency insert should fail the CHECK constraint")
	}
}

func TestUpdateTaskReplacesDependencies(t *testing.T) {
	st, err := openStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("openStore() error = %v", err)
	}
	defer st.Close()

	seedStoreTask(t, st, "first", "first task")
	seedStoreTask(t, st, "second", "second task")
	seedStoreTask(t, st, "third", "third task")

	// Set, via the one-transaction update path.
	setBlockedBy(t, st, "first", []string{"second", "third"})

	task, err := st.getTask(context.Background(), "first")
	if err != nil {
		t.Fatalf("getTask(first) error = %v", err)
	}

	if len(task.BlockedBy) != 2 || !slices.Contains(task.BlockedBy, "second") || !slices.Contains(task.BlockedBy, "third") {
		t.Errorf("first.BlockedBy = %v, want [second third]", task.BlockedBy)
	}

	if len(task.Blocks) != 0 {
		t.Errorf("first.Blocks = %v, want empty (no reverse edges)", task.Blocks)
	}

	// The reverse edge is visible from the blocker.
	blocker, err := st.getTask(context.Background(), "second")
	if err != nil {
		t.Fatalf("getTask(second) error = %v", err)
	}

	if len(blocker.Blocks) != 1 || blocker.Blocks[0] != "first" {
		t.Errorf("second.Blocks = %v, want [first]", blocker.Blocks)
	}

	if len(blocker.BlockedBy) != 0 {
		t.Errorf("second.BlockedBy = %v, want empty", blocker.BlockedBy)
	}

	// Replace with a different set.
	setBlockedBy(t, st, "first", []string{"third"})

	task, err = st.getTask(context.Background(), "first")
	if err != nil {
		t.Fatalf("getTask(first) after replace error = %v", err)
	}

	if len(task.BlockedBy) != 1 || task.BlockedBy[0] != "third" {
		t.Errorf("first.BlockedBy after replace = %v, want [third]", task.BlockedBy)
	}

	// Clear-all with an empty (but non-nil) slice.
	setBlockedBy(t, st, "first", []string{})

	task, err = st.getTask(context.Background(), "first")
	if err != nil {
		t.Fatalf("getTask(first) after clear error = %v", err)
	}

	if len(task.BlockedBy) != 0 || len(task.Blocks) != 0 {
		t.Errorf("first edges after clear = %v/%v, want empty/empty", task.BlockedBy, task.Blocks)
	}

	blocker, err = st.getTask(context.Background(), "third")
	if err != nil {
		t.Fatalf("getTask(third) after clear error = %v", err)
	}

	if len(blocker.Blocks) != 0 {
		t.Errorf("third.Blocks after clear = %v, want empty", blocker.Blocks)
	}
}

func TestListTasksAnnotatesBlockedBy(t *testing.T) {
	st, err := openStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("openStore() error = %v", err)
	}
	defer st.Close()

	seedStoreTask(t, st, "first", "first task")
	seedStoreTask(t, st, "second", "second task")
	setBlockedBy(t, st, "first", []string{"second"})

	result, err := st.listTasks(context.Background(), plugin.TaskFilters{})
	if err != nil {
		t.Fatalf("listTasks() error = %v", err)
	}

	blockedBy := map[string][]string{}
	for _, task := range result.Tasks {
		blockedBy[task.ID] = task.BlockedBy
	}

	if len(blockedBy["first"]) != 1 || blockedBy["first"][0] != "second" {
		t.Errorf("listed first.BlockedBy = %v, want [second]", blockedBy["first"])
	}

	if len(blockedBy["second"]) != 0 {
		t.Errorf("listed second.BlockedBy = %v, want empty", blockedBy["second"])
	}
}

func TestReadyFilterHidesBlockedTasks(t *testing.T) {
	st, err := openStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("openStore() error = %v", err)
	}
	defer st.Close()

	seedStoreTask(t, st, "blocked", "blocked task")
	seedStoreTask(t, st, "blocker", "blocker task")
	seedStoreTask(t, st, "free", "free task")
	setBlockedBy(t, st, "blocked", []string{"blocker"})

	// While the blocker is unfinished, only it and the unblocked task show.
	got := listTitles(t, st, plugin.TaskFilters{Ready: true})
	if len(got) != 2 || !slices.Contains(got, "blocker task") || !slices.Contains(got, "free task") {
		t.Errorf("ready list while blocked = %v, want [blocker task free task]", got)
	}

	// The COUNT follows the same filter, keeping pagination totals honest.
	result, err := st.listTasks(context.Background(), plugin.TaskFilters{Ready: true})
	if err != nil {
		t.Fatalf("listTasks(ready) error = %v", err)
	}

	if result.TotalCount != 2 {
		t.Errorf("ready TotalCount = %d, want 2", result.TotalCount)
	}

	// Finishing the blocker unblocks the dependent.
	done := plugin.StatusDone
	if _, err := st.updateTask(context.Background(), "blocker", plugin.UpdateTaskRequest{Status: &done}); err != nil {
		t.Fatalf("completing blocker: %v", err)
	}

	got = listTitles(t, st, plugin.TaskFilters{Ready: true})
	if len(got) != 3 || !slices.Contains(got, "blocked task") {
		t.Errorf("ready list after unblock = %v, want all three tasks", got)
	}

	// Without the filter nothing is hidden, blocked or not.
	got = listTitles(t, st, plugin.TaskFilters{})
	if len(got) != 3 {
		t.Errorf("unfiltered list = %v, want all three tasks", got)
	}
}

func TestDeleteTaskCascadesDependencyEdges(t *testing.T) {
	st, err := openStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("openStore() error = %v", err)
	}
	defer st.Close()

	seedStoreTask(t, st, "dependent", "dependent task")
	seedStoreTask(t, st, "blocker", "blocker task")
	seedStoreTask(t, st, "spare", "spare blocker task")
	setBlockedBy(t, st, "dependent", []string{"blocker", "spare"})

	// Deleting one blocker removes its edge; the other survives.
	if err := st.deleteTask(context.Background(), "blocker"); err != nil {
		t.Fatalf("deleteTask(blocker) error = %v", err)
	}

	task, err := st.getTask(context.Background(), "dependent")
	if err != nil {
		t.Fatalf("getTask(dependent) error = %v", err)
	}

	if len(task.BlockedBy) != 1 || task.BlockedBy[0] != "spare" {
		t.Errorf("dependent.BlockedBy after blocker deleted = %v, want [spare]", task.BlockedBy)
	}

	// Deleting the dependent removes the surviving edge in the other
	// direction too.
	if err := st.deleteTask(context.Background(), "dependent"); err != nil {
		t.Fatalf("deleteTask(dependent) error = %v", err)
	}

	var count int
	if err := st.db.QueryRow("SELECT COUNT(*) FROM task_dependencies").Scan(&count); err != nil {
		t.Fatalf("counting edges: %v", err)
	}

	if count != 0 {
		t.Errorf("edge count after deletes = %d, want 0", count)
	}
}
