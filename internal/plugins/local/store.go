package local

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite" // pure-Go SQLite driver, registers as "sqlite"

	"github.com/yousfisaad/lazyarchon/v2/internal/plugin"
)

// timeStampLayout is a fixed-width UTC format (always 6 fractional digits) so
// lexicographic ordering matches chronological ordering.
const timeStampLayout = "2006-01-02T15:04:05.000000Z"

// defaultListLimit caps unbounded ListTasks calls, mirroring Archon's page size.
const defaultListLimit = 500

// store owns the SQLite connection and all SQL for the local backend.
type store struct {
	db *sql.DB
}

// openStore opens (creating if needed) the database at path, applies pending
// migrations, and seeds a default project on first creation.
func openStore(path string) (*store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// WAL + busy_timeout allow a TUI and an MCP server process to share the
	// database; foreign_keys enforces referential integrity; synchronous
	// NORMAL is the recommended trade-off for WAL mode.
	dsn := "file:" + filepath.ToSlash(path) +
		"?_pragma=journal_mode(WAL)" +
		"&_pragma=busy_timeout(5000)" +
		"&_pragma=foreign_keys(1)" +
		"&_pragma=synchronous(NORMAL)"

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// A single connection removes intra-process SQLITE_BUSY entirely;
	// cross-process contention is handled by WAL + busy_timeout. This also
	// means we must never run a query while scanning rows from another.
	db.SetMaxOpenConns(1)

	st := &store{db: db}

	if err := st.migrate(); err != nil {
		db.Close()
		return nil, err
	}

	if err := st.seedIfEmpty(); err != nil {
		db.Close()
		return nil, err
	}

	return st, nil
}

// Close releases the database connection.
func (s *store) Close() error {
	return s.db.Close()
}

// ping verifies the database is reachable (used by HealthCheck).
func (s *store) ping(ctx context.Context) error {
	var one int
	if err := s.db.QueryRowContext(ctx, "SELECT 1").Scan(&one); err != nil || one != 1 {
		if err != nil {
			return fmt.Errorf("database ping failed: %w", err)
		}
		return fmt.Errorf("database ping returned %d, want 1", one)
	}

	return nil
}

// migrate applies pending migrations in order, tracking progress with
// PRAGMA user_version.
func (s *store) migrate() error {
	var version int
	if err := s.db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return fmt.Errorf("failed to read schema version: %w", err)
	}

	for current := version; current < len(migrations); current++ {
		tx, err := s.db.Begin()
		if err != nil {
			return fmt.Errorf("failed to start migration %d: %w", current+1, err)
		}

		if _, err := tx.Exec(migrations[current]); err != nil {
			// The transaction already failed; rollback is best-effort.
			_ = tx.Rollback()
			return fmt.Errorf("migration %d failed: %w", current+1, err)
		}

		if _, err := tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", current+1)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to record migration %d: %w", current+1, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", current+1, err)
		}
	}

	return nil
}

// seedIfEmpty creates a default "Inbox" project when the database was just
// created, so the TUI (which cannot create projects) is usable immediately.
func (s *store) seedIfEmpty() error {
	var count int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM projects").Scan(&count); err != nil {
		return fmt.Errorf("failed to count projects: %w", err)
	}

	if count > 0 {
		return nil
	}

	now := formatTime(time.Now())
	_, err := s.db.Exec(
		"INSERT INTO projects (id, title, description, color, pinned, extra, created_at, updated_at) VALUES (?, ?, ?, '', 0, '{}', ?, ?)",
		newID(), "Inbox", "Default project", now, now,
	)
	if err != nil {
		return fmt.Errorf("failed to seed default project: %w", err)
	}

	return nil
}

// formatTime stores a timestamp in the canonical UTC layout.
func formatTime(t time.Time) string {
	return t.UTC().Format(timeStampLayout)
}

// parseTime reads a timestamp from the canonical layout.
func parseTime(value string) (time.Time, error) {
	return time.Parse(timeStampLayout, value)
}

// ---------- row types and conversion ----------

// projectRow mirrors the projects table columns.
type projectRow struct {
	ID, Title, Description, Color string
	Pinned                        bool
	Extra                         string
	CreatedAt, UpdatedAt          string
}

func (r projectRow) toPlugin() (plugin.Project, error) {
	createdAt, err := parseTime(r.CreatedAt)
	if err != nil {
		return plugin.Project{}, fmt.Errorf("invalid created_at for project %s: %w", r.ID, err)
	}

	updatedAt, err := parseTime(r.UpdatedAt)
	if err != nil {
		return plugin.Project{}, fmt.Errorf("invalid updated_at for project %s: %w", r.ID, err)
	}

	extra, err := decodeMap(r.Extra)
	if err != nil {
		return plugin.Project{}, fmt.Errorf("invalid extra for project %s: %w", r.ID, err)
	}

	return plugin.Project{
		ID:          r.ID,
		Title:       r.Title,
		Description: r.Description,
		Color:       r.Color,
		Pinned:      r.Pinned,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		Extra:       extra,
	}, nil
}

const projectColumns = "id, title, description, color, pinned, extra, created_at, updated_at"

func (s *store) scanProject(row interface{ Scan(...any) error }) (plugin.Project, error) {
	var r projectRow
	if err := row.Scan(&r.ID, &r.Title, &r.Description, &r.Color, &r.Pinned, &r.Extra, &r.CreatedAt, &r.UpdatedAt); err != nil {
		return plugin.Project{}, err
	}

	return r.toPlugin()
}

// taskRow mirrors the tasks table columns.
type taskRow struct {
	ID, ProjectID, Title, Description, Status, Assignee string
	Priority                                            int
	ParentID, DueDate, ArchivedAt                       sql.NullString
	Tags, Sources, CodeExamples, Extra                  string
	Archived                                            bool
	CreatedAt, UpdatedAt                                string
}

func (r taskRow) toPlugin() (plugin.Task, error) {
	createdAt, err := parseTime(r.CreatedAt)
	if err != nil {
		return plugin.Task{}, fmt.Errorf("invalid created_at for task %s: %w", r.ID, err)
	}

	updatedAt, err := parseTime(r.UpdatedAt)
	if err != nil {
		return plugin.Task{}, fmt.Errorf("invalid updated_at for task %s: %w", r.ID, err)
	}

	tags, err := decodeStrings(r.Tags)
	if err != nil {
		return plugin.Task{}, fmt.Errorf("invalid tags for task %s: %w", r.ID, err)
	}

	extra, err := decodeMap(r.Extra)
	if err != nil {
		return plugin.Task{}, fmt.Errorf("invalid extra for task %s: %w", r.ID, err)
	}

	// Surface the dedicated columns under the same Extra keys the archon
	// adapter uses, so consumers see identical shapes across backends.
	sources, err := decodeAny(r.Sources)
	if err != nil {
		return plugin.Task{}, fmt.Errorf("invalid sources for task %s: %w", r.ID, err)
	}
	extra["sources"] = sources

	codeExamples, err := decodeAny(r.CodeExamples)
	if err != nil {
		return plugin.Task{}, fmt.Errorf("invalid code_examples for task %s: %w", r.ID, err)
	}
	extra["code_examples"] = codeExamples

	task := plugin.Task{
		ID:          r.ID,
		ProjectID:   r.ProjectID,
		Title:       r.Title,
		Description: r.Description,
		Status:      r.Status,
		Priority:    r.Priority,
		Assignee:    r.Assignee,
		Tags:        tags,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		Archived:    r.Archived,
		Extra:       extra,
	}

	if r.ParentID.Valid {
		parentID := r.ParentID.String
		task.ParentID = &parentID
	}

	if r.DueDate.Valid {
		due, err := parseTime(r.DueDate.String)
		if err != nil {
			return plugin.Task{}, fmt.Errorf("invalid due_date for task %s: %w", r.ID, err)
		}
		task.DueDate = &due
	}

	if r.ArchivedAt.Valid {
		archivedAt, err := parseTime(r.ArchivedAt.String)
		if err != nil {
			return plugin.Task{}, fmt.Errorf("invalid archived_at for task %s: %w", r.ID, err)
		}
		task.ArchivedAt = &archivedAt
	}

	return task, nil
}

const taskColumns = "id, project_id, parent_id, title, description, status, priority, assignee, tags, due_date, sources, code_examples, extra, archived, archived_at, created_at, updated_at"

func scanTask(row interface{ Scan(...any) error }) (plugin.Task, error) {
	var record taskRow
	if err := row.Scan(
		&record.ID, &record.ProjectID, &record.ParentID, &record.Title, &record.Description, &record.Status,
		&record.Priority, &record.Assignee, &record.Tags, &record.DueDate, &record.Sources, &record.CodeExamples,
		&record.Extra, &record.Archived, &record.ArchivedAt, &record.CreatedAt, &record.UpdatedAt,
	); err != nil {
		return plugin.Task{}, err
	}

	return record.toPlugin()
}

// ---------- JSON column helpers ----------

func decodeMap(raw string) (map[string]interface{}, error) {
	out := make(map[string]interface{})
	if strings.TrimSpace(raw) == "" {
		return out, nil
	}

	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}

	return out, nil
}

func decodeStrings(raw string) ([]string, error) {
	out := []string{}
	if strings.TrimSpace(raw) == "" {
		return out, nil
	}

	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}

	return out, nil
}

// decodeAny decodes a JSON column into a generic value, returning an empty
// slice for empty input.
func decodeAny(raw string) (interface{}, error) {
	if strings.TrimSpace(raw) == "" {
		return []interface{}{}, nil
	}

	var out interface{}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}

	return out, nil
}

// splitExtra separates the reserved Extra keys backed by dedicated columns
// from the remaining free-form keys. Returns (sources, codeExamples, rest).
func splitExtra(extra map[string]interface{}) (string, string, string) {
	rest := make(map[string]interface{})
	for k, v := range extra {
		if k != "sources" && k != "code_examples" {
			rest[k] = v
		}
	}

	sources, _ := json.Marshal(extra["sources"])
	if extra["sources"] == nil {
		sources = []byte("[]")
	}

	codeExamples, _ := json.Marshal(extra["code_examples"])
	if extra["code_examples"] == nil {
		codeExamples = []byte("[]")
	}

	restJSON, _ := json.Marshal(rest)

	return string(sources), string(codeExamples), string(restJSON)
}
