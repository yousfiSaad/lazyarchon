package local

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/yousfisaad/lazyarchon/v2/internal/plugin"
)

// dependencies_store.go - persistence for blocked_by edges. Blocked/ready is
// computed at read time from these edges plus task statuses; nothing derived
// is stored, so it cannot go stale across the processes sharing the database.

// replaceTaskDependencies swaps the whole blocker set of a task inside the
// caller's transaction (DELETE + plain INSERTs — a typo'd blocker ID must
// error, not silently vanish through OR IGNORE).
func replaceTaskDependencies(ctx context.Context, tx *sql.Tx, taskID string, ids []string) error {
	if _, err := tx.ExecContext(ctx, "DELETE FROM task_dependencies WHERE task_id = ?", taskID); err != nil {
		return fmt.Errorf("failed to clear dependencies of task %s: %w", taskID, err)
	}

	now := formatTime(nowFunc())
	for _, id := range ids {
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO task_dependencies (task_id, depends_on_id, created_at) VALUES (?, ?, ?)",
			taskID, id, now,
		); err != nil {
			return fmt.Errorf("failed to record dependency %s -> %s: %w", taskID, id, err)
		}
	}

	return nil
}

// dependencyEdges loads the full edge map (dependent -> blockers). The
// dataset is personal-scale, so one query feeds the cycle guard comfortably.
func (s *store) dependencyEdges(ctx context.Context) (map[string][]string, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT task_id, depends_on_id FROM task_dependencies")
	if err != nil {
		return nil, fmt.Errorf("failed to load dependencies: %w", err)
	}
	defer rows.Close()

	edges := make(map[string][]string)
	for rows.Next() {
		var taskID, dependsOnID string
		if err := rows.Scan(&taskID, &dependsOnID); err != nil {
			return nil, fmt.Errorf("failed to scan dependency: %w", err)
		}
		edges[taskID] = append(edges[taskID], dependsOnID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate dependencies: %w", err)
	}

	return edges, nil
}

// setTaskDependencies installs a task's blocker set outside an update
// (the create path).
func (s *store) setTaskDependencies(ctx context.Context, taskID string, ids []string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin dependency transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			// The transaction already failed; rollback is best-effort.
			_ = tx.Rollback()
		}
	}()

	if err := replaceTaskDependencies(ctx, tx, taskID, ids); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit dependencies of task %s: %w", taskID, err)
	}
	committed = true

	return nil
}

// dbQuerier abstracts QueryContext over the connection pool and a
// transaction. The pool holds a single connection (see openStore), and an
// open transaction occupies it — so transaction-time helpers must query
// through the tx itself, never the pool, or they deadlock.
type dbQuerier interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// taskEdges returns both directions of a task's dependency edges, for
// GetTask and the post-update re-read.
func (s *store) taskEdges(ctx context.Context, taskID string) (blockedBy, blocks []string, err error) {
	return taskEdgesFrom(ctx, s.db, taskID)
}

// taskEdgesFrom is taskEdges over either the pool or an open transaction.
func taskEdgesFrom(ctx context.Context, querier dbQuerier, taskID string) (blockedBy, blocks []string, err error) {
	blockedBy, err = edgeIDs(ctx, querier, "SELECT depends_on_id FROM task_dependencies WHERE task_id = ?", taskID)
	if err != nil {
		return nil, nil, err
	}

	blocks, err = edgeIDs(ctx, querier, "SELECT task_id FROM task_dependencies WHERE depends_on_id = ?", taskID)
	if err != nil {
		return nil, nil, err
	}

	return blockedBy, blocks, nil
}

// edgeIDs runs a single-column ID query.
func edgeIDs(ctx context.Context, querier dbQuerier, query, arg string) ([]string, error) {
	rows, err := querier.QueryContext(ctx, query, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to query dependency edges: %w", err)
	}
	defer rows.Close()

	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan dependency edge: %w", err)
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate dependency edges: %w", err)
	}

	return ids, nil
}

// annotateBlockedBy fills BlockedBy on a page of listed tasks with one
// query. Must run after the listing rows are closed — single-connection
// discipline, same as the COUNT query in listTasks.
func (s *store) annotateBlockedBy(ctx context.Context, tasks []plugin.Task) error {
	if len(tasks) == 0 {
		return nil
	}

	placeholders := make([]string, len(tasks))
	args := make([]interface{}, 0, len(tasks))
	for i, task := range tasks {
		placeholders[i] = "?"
		args = append(args, task.ID)
	}

	rows, err := s.db.QueryContext(ctx,
		//nolint:gosec // G202: placeholders are one "?" per listed task; values are ?-bound.
		"SELECT task_id, depends_on_id FROM task_dependencies WHERE task_id IN ("+strings.Join(placeholders, ",")+")",
		args...,
	)
	if err != nil {
		return fmt.Errorf("failed to load page dependencies: %w", err)
	}
	defer rows.Close()

	blockedBy := make(map[string][]string, len(tasks))
	for rows.Next() {
		var taskID, dependsOnID string
		if err := rows.Scan(&taskID, &dependsOnID); err != nil {
			return fmt.Errorf("failed to scan page dependency: %w", err)
		}
		blockedBy[taskID] = append(blockedBy[taskID], dependsOnID)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed to iterate page dependencies: %w", err)
	}

	for i := range tasks {
		if ids, ok := blockedBy[tasks[i].ID]; ok {
			tasks[i].BlockedBy = ids
		}
	}

	return nil
}

// unfinishedBlocker is a blocker that has not reached done. Only
// status = done unblocks; an archived-but-undone blocker still blocks.
type unfinishedBlocker struct {
	ID    string
	Title string
}

// unfinishedBlockers lists the unfinished blockers recorded against a task —
// the enforcement query for updates that do not replace the blocker set.
func (s *store) unfinishedBlockers(ctx context.Context, taskID string) ([]unfinishedBlocker, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT dep.id, dep.title FROM task_dependencies d JOIN tasks dep ON dep.id = d.depends_on_id "+
			"WHERE d.task_id = ? AND dep.status <> 'done'",
		taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load unfinished blockers of task %s: %w", taskID, err)
	}
	defer rows.Close()

	blockers := []unfinishedBlocker{}
	for rows.Next() {
		var blocker unfinishedBlocker
		if err := rows.Scan(&blocker.ID, &blocker.Title); err != nil {
			return nil, fmt.Errorf("failed to scan unfinished blocker: %w", err)
		}
		blockers = append(blockers, blocker)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate unfinished blockers: %w", err)
	}

	return blockers, nil
}

// unfinishedAmong lists the members of an arbitrary blocker-ID set that have
// not reached done — the enforcement query for creates and for updates whose
// request replaces the set (judged on the incoming IDs, before they are
// stored).
func (s *store) unfinishedAmong(ctx context.Context, ids []string) ([]unfinishedBlocker, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, 0, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args = append(args, id)
	}

	rows, err := s.db.QueryContext(ctx,
		//nolint:gosec // G202: placeholders are one "?" per blocker ID; values are ?-bound.
		"SELECT id, title FROM tasks WHERE id IN ("+strings.Join(placeholders, ",")+") AND status <> 'done'",
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load blocker statuses: %w", err)
	}
	defer rows.Close()

	blockers := []unfinishedBlocker{}
	for rows.Next() {
		var blocker unfinishedBlocker
		if err := rows.Scan(&blocker.ID, &blocker.Title); err != nil {
			return nil, fmt.Errorf("failed to scan blocker status: %w", err)
		}
		blockers = append(blockers, blocker)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate blocker statuses: %w", err)
	}

	return blockers, nil
}

// wouldCreateDependencyCycle reports whether giving taskID the proposed
// blocker set would introduce a dependency cycle: walk from taskID along
// dependent -> dependency edges (with the proposal overlaid) and see whether
// the walk leads back to taskID. The visited set terminates the walk even on
// corrupt pre-existing cyclic data, so no depth cap is needed.
func wouldCreateDependencyCycle(edges map[string][]string, taskID string, proposed []string) bool {
	overlay := make(map[string][]string, len(edges)+1)
	for key, value := range edges {
		overlay[key] = value
	}
	overlay[taskID] = proposed

	visited := map[string]bool{taskID: true}
	queue := append([]string{}, proposed...)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == taskID {
			return true
		}

		if visited[current] {
			continue
		}
		visited[current] = true

		queue = append(queue, overlay[current]...)
	}

	return false
}
