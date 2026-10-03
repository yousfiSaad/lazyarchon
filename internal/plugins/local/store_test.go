package local

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"
)

func TestOpenStoreCreatesAndSeeds(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")

	st, err := openStore(path)
	if err != nil {
		t.Fatalf("openStore() error = %v", err)
	}
	defer st.Close()

	// Fresh database must be at schema version 1 with a seeded Inbox project.
	var version int
	if err := st.db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		t.Fatalf("reading user_version: %v", err)
	}

	if version != schemaVersion() {
		t.Errorf("user_version = %d, want %d", version, schemaVersion())
	}

	var count int
	if err := st.db.QueryRow("SELECT COUNT(*) FROM projects").Scan(&count); err != nil {
		t.Fatalf("counting projects: %v", err)
	}

	if count != 1 {
		t.Fatalf("seeded project count = %d, want 1", count)
	}

	var title string
	if err := st.db.QueryRow("SELECT title FROM projects").Scan(&title); err != nil {
		t.Fatalf("reading seeded project: %v", err)
	}

	if title != defaultProjectTitle {
		t.Errorf("seeded project title = %q, want %q", title, defaultProjectTitle)
	}
}

func TestOpenStoreReopenDoesNotReseed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")

	st, err := openStore(path)
	if err != nil {
		t.Fatalf("first openStore() error = %v", err)
	}
	st.Close()

	st, err = openStore(path)
	if err != nil {
		t.Fatalf("second openStore() error = %v", err)
	}
	defer st.Close()

	var count int
	if err := st.db.QueryRow("SELECT COUNT(*) FROM projects").Scan(&count); err != nil {
		t.Fatalf("counting projects: %v", err)
	}

	if count != 1 {
		t.Errorf("project count after reopen = %d, want 1 (no reseed)", count)
	}
}

func TestTimestampFormatIsFixedWidth(t *testing.T) {
	reference := time.Date(2026, 8, 28, 14, 3, 9, 123456789, time.UTC)

	formatted := formatTime(reference)

	// Fixed width guarantees lexicographic == chronological ordering.
	if len(formatted) != len(timeStampLayout) {
		t.Errorf("formatTime length = %d, want %d (%q)", len(formatted), len(timeStampLayout), formatted)
	}

	parsed, err := parseTime(formatted)
	if err != nil {
		t.Fatalf("parseTime(%q) error = %v", formatted, err)
	}

	// Round-trip loses nothing beyond microsecond precision.
	want := reference.Truncate(time.Microsecond)
	if !parsed.Equal(want) {
		t.Errorf("round-trip = %v, want %v", parsed, want)
	}

	// Lexicographic ordering must match chronological ordering.
	earlier := formatTime(reference.Add(-time.Second))
	if earlier >= formatted {
		t.Errorf("earlier timestamp %q not lexically before %q", earlier, formatted)
	}
}

func TestMigrateAddsWorktreeColumn(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")

	// Build a version-1 database by hand: apply only the initial schema,
	// mark it applied, and seed one task row predating the worktree column.
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path))
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}

	if _, err := db.Exec(migrationV1); err != nil {
		t.Fatalf("applying migrationV1: %v", err)
	}
	if _, err := db.Exec("PRAGMA user_version = 1"); err != nil {
		t.Fatalf("setting user_version: %v", err)
	}

	now := formatTime(time.Now())
	if _, err := db.Exec("INSERT INTO projects (id, title, created_at, updated_at) VALUES ('p1', 'Legacy', ?, ?)", now, now); err != nil {
		t.Fatalf("seeding project: %v", err)
	}
	if _, err := db.Exec("INSERT INTO tasks (id, project_id, title, created_at, updated_at) VALUES ('t1', 'p1', 'legacy task', ?, ?)", now, now); err != nil {
		t.Fatalf("seeding task: %v", err)
	}
	db.Close()

	st, err := openStore(path)
	if err != nil {
		t.Fatalf("openStore() on v1 database: %v", err)
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

	if task.Title != "legacy task" {
		t.Errorf("migrated task title = %q, want %q", task.Title, "legacy task")
	}

	if task.Worktree != "" {
		t.Errorf("migrated task Worktree = %q, want empty", task.Worktree)
	}
}

func TestMigrateWidensPriorityCheck(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")

	// Build a version-2 database by hand, with a parent/child pair (the
	// rebuild must not trip the self-FK) and pre-widening priorities.
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path))
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}

	if _, err := db.Exec(migrationV1); err != nil {
		t.Fatalf("applying migrationV1: %v", err)
	}
	if _, err := db.Exec(migrationV2); err != nil {
		t.Fatalf("applying migrationV2: %v", err)
	}
	if _, err := db.Exec("PRAGMA user_version = 2"); err != nil {
		t.Fatalf("setting user_version: %v", err)
	}

	now := formatTime(time.Now())
	if _, err := db.Exec("INSERT INTO projects (id, title, created_at, updated_at) VALUES ('p1', 'Legacy', ?, ?)", now, now); err != nil {
		t.Fatalf("seeding project: %v", err)
	}
	if _, err := db.Exec("INSERT INTO tasks (id, project_id, title, priority, worktree, created_at, updated_at) VALUES ('parent', 'p1', 'parent task', 4, 'wt-a', ?, ?)", now, now); err != nil {
		t.Fatalf("seeding parent task: %v", err)
	}
	if _, err := db.Exec("INSERT INTO tasks (id, project_id, parent_id, title, priority, created_at, updated_at) VALUES ('child', 'p1', 'parent', 'child task', 1, ?, ?)", now, now); err != nil {
		t.Fatalf("seeding child task: %v", err)
	}
	db.Close()

	st, err := openStore(path)
	if err != nil {
		t.Fatalf("openStore() on v2 database: %v", err)
	}
	defer st.Close()

	var version int
	if err := st.db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		t.Fatalf("reading user_version: %v", err)
	}

	if version != schemaVersion() {
		t.Errorf("user_version = %d, want %d", version, schemaVersion())
	}

	child, err := st.getTask(context.Background(), "child")
	if err != nil {
		t.Fatalf("getTask() on migrated child: %v", err)
	}

	if child.ParentID == nil || *child.ParentID != "parent" {
		t.Errorf("migrated child parent = %v, want parent link intact", child.ParentID)
	}

	parent, err := st.getTask(context.Background(), "parent")
	if err != nil {
		t.Fatalf("getTask() on migrated parent: %v", err)
	}

	if parent.Priority != 4 || parent.Worktree != "wt-a" {
		t.Errorf("migrated parent priority/worktree = %d/%q, want 4/wt-a", parent.Priority, parent.Worktree)
	}

	// The widened CHECK now accepts backlog-tier priorities.
	if _, err := st.db.Exec("UPDATE tasks SET priority = 5 WHERE id = 'parent'"); err != nil {
		t.Errorf("setting priority 5 after migration: %v", err)
	}
}

func TestSplitExtraSeparatesReservedKeys(t *testing.T) {
	extra := map[string]interface{}{
		"sources":       []interface{}{map[string]interface{}{"url": "https://example.com"}},
		"code_examples": []interface{}{},
		"github_repo":   "owner/repo",
	}

	sources, codeExamples, rest := splitExtra(extra)

	if sources == "" || codeExamples == "" || rest == "" {
		t.Fatalf("splitExtra returned empty part: %q %q %q", sources, codeExamples, rest)
	}

	if sources == "null" {
		t.Errorf("sources = %q, want JSON array", sources)
	}

	if codeExamples != "[]" {
		t.Errorf("codeExamples = %q, want []", codeExamples)
	}

	decodedRest, err := decodeMap(rest)
	if err != nil {
		t.Fatalf("decodeMap(rest) error = %v", err)
	}

	if _, ok := decodedRest["sources"]; ok {
		t.Error("rest should not contain sources")
	}

	if decodedRest["github_repo"] != "owner/repo" {
		t.Errorf("rest[github_repo] = %v, want owner/repo", decodedRest["github_repo"])
	}
}
