package local

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/yousfisaad/lazyarchon/v2/internal/plugin"
)

// maxParentWalk guards the cycle check against corrupt pre-existing data.
const maxParentWalk = 100

// taskSortClause sorts by status lane (todo→doing→review→done), then
// priority (1 = most urgent first), then newest first, with id as a
// deterministic tiebreaker for stable pagination.
const taskSortClause = "ORDER BY CASE status WHEN 'todo' THEN 0 WHEN 'doing' THEN 1 WHEN 'review' THEN 2 ELSE 3 END, priority ASC, created_at DESC, id ASC"

// filterBuilder incrementally builds a WHERE clause with bound arguments.
type filterBuilder struct {
	conditions []string
	args       []interface{}
}

func (b *filterBuilder) add(condition string, arg interface{}) {
	b.conditions = append(b.conditions, condition)
	b.args = append(b.args, arg)
}

func (b *filterBuilder) addRaw(condition string) {
	b.conditions = append(b.conditions, condition)
}

func (b *filterBuilder) whereClause() string {
	if len(b.conditions) == 0 {
		return ""
	}

	return "WHERE " + strings.Join(b.conditions, " AND ")
}

// updateBuilder incrementally builds a "SET a = ?, b = ?" clause with bound
// arguments, mirroring filterBuilder for UPDATE statements.
type updateBuilder struct {
	sets []string
	args []interface{}
}

// add appends a "column = ?" fragment with its bound values.
func (u *updateBuilder) add(set string, values ...interface{}) {
	u.sets = append(u.sets, set)
	u.args = append(u.args, values...)
}

// addRaw appends a fragment that binds no values (e.g. "col = NULL").
func (u *updateBuilder) addRaw(set string) {
	u.sets = append(u.sets, set)
}

func (u *updateBuilder) clause() string {
	return strings.Join(u.sets, ", ")
}

// argsWithWhere copies the bound values and appends the WHERE-clause value,
// so the builder's own slice is never aliased or grown in place.
func (u *updateBuilder) argsWithWhere(value interface{}) []interface{} {
	args := make([]interface{}, 0, len(u.args)+1)
	args = append(args, u.args...)
	return append(args, value)
}

// buildTaskFilter translates generic filters into SQL conditions.
//
//nolint:gocyclo // One guard clause per filter field; extracting them would hurt readability
func buildTaskFilter(filters plugin.TaskFilters) *filterBuilder {
	builder := &filterBuilder{}

	if filters.ProjectID != nil && *filters.ProjectID != "" {
		builder.add("project_id = ?", *filters.ProjectID)
	}

	if filters.Status != nil && *filters.Status != "" {
		builder.add("status = ?", *filters.Status)
	}

	if filters.Assignee != nil && *filters.Assignee != "" {
		builder.add("assignee = ?", *filters.Assignee)
	}

	if filters.Worktree != nil && *filters.Worktree != "" {
		builder.add("worktree = ?", *filters.Worktree)
	}

	if !filters.IncludeArchived {
		builder.addRaw("archived = 0")
	}

	if filters.SearchQuery != nil && *filters.SearchQuery != "" {
		pattern := "%" + escapeLike(*filters.SearchQuery) + "%"
		builder.conditions = append(builder.conditions, "(title LIKE ? ESCAPE '\\' OR description LIKE ? ESCAPE '\\')")
		builder.args = append(builder.args, pattern, pattern)
	}

	if len(filters.Tags) > 0 {
		placeholders := make([]string, len(filters.Tags))
		for i, tag := range filters.Tags {
			placeholders[i] = "?"
			builder.args = append(builder.args, tag)
		}
		builder.addRaw("EXISTS (SELECT 1 FROM json_each(tasks.tags) AS jt WHERE jt.value IN (" + strings.Join(placeholders, ",") + "))")
	}

	if filters.Ready {
		builder.addRaw("NOT EXISTS (SELECT 1 FROM task_dependencies d JOIN tasks dep ON dep.id = d.depends_on_id WHERE d.task_id = tasks.id AND dep.status <> 'done')")
	}

	if filters.DueBefore != nil {
		builder.add("due_date IS NOT NULL AND due_date <= ?", formatTime(*filters.DueBefore))
	}

	if filters.DueAfter != nil {
		builder.add("due_date IS NOT NULL AND due_date >= ?", formatTime(*filters.DueAfter))
	}

	return builder
}

// escapeLike escapes LIKE wildcards and the escape character itself.
func escapeLike(input string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(input)
}

// listTasks returns a page of tasks matching the filters plus total counts.
func (s *store) listTasks(ctx context.Context, filters plugin.TaskFilters) (*plugin.TaskListResult, error) {
	where := buildTaskFilter(filters)

	limit := filters.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}

	offset := max(filters.Offset, 0)

	// Collect rows fully before running the count query (single connection).
	rows, err := s.db.QueryContext(ctx,
		//nolint:gosec // G202: every fragment is a compile-time literal (taskColumns,
		// taskSortClause consts, fixed clause strings); all values are ?-bound in args.
		"SELECT "+taskColumns+" FROM tasks "+where.whereClause()+" "+taskSortClause+" LIMIT ? OFFSET ?",
		append(where.args, limit, offset)...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	tasks := []plugin.Task{}
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("failed to iterate tasks: %w", err)
	}
	rows.Close()

	countArgs := make([]interface{}, len(where.args))
	copy(countArgs, where.args)

	var total int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks "+where.whereClause(), countArgs...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count tasks: %w", err)
	}

	if err := s.annotateBlockedBy(ctx, tasks); err != nil {
		return nil, err
	}

	result := &plugin.TaskListResult{
		Tasks:      tasks,
		TotalCount: total,
	}

	result.HasMore = total > offset+len(tasks)
	if result.HasMore {
		result.NextOffset = offset + len(tasks)
	}

	return result, nil
}

// getTask returns a single task by ID, with its dependency edges in both
// directions. Returns sql.ErrNoRows (wrapped) when the task does not exist.
func (s *store) getTask(ctx context.Context, taskID string) (*plugin.Task, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+taskColumns+" FROM tasks WHERE id = ?", taskID)

	task, err := scanTask(row)
	if err != nil {
		return nil, fmt.Errorf("failed to get task %s: %w", taskID, err)
	}

	blockedBy, blocks, err := s.taskEdges(ctx, taskID)
	if err != nil {
		return nil, err
	}
	task.BlockedBy = blockedBy
	task.Blocks = blocks

	return &task, nil
}

// insertTask persists a new task row with a strictly monotonic created_at:
// whenever the microsecond-truncated clock does not strictly exceed
// MAX(created_at) — a tie, a sub-microsecond gap that would truncate to a
// tie, or a clock step-back — the stamp is bumped one microsecond past the
// maximum. Ordering therefore never falls through to the id tiebreak, at
// the cost of stamps running a few microseconds ahead of wall clock after
// bursts of tied creates (hybrid-clock-lite). The MAX read and the INSERT
// share one transaction so they are atomic across the TUI and MCP server
// processes that deliberately share this file (_txlock=immediate takes the
// write lock up front; busy_timeout makes it wait instead of racing).
func (s *store) insertTask(ctx context.Context, task *plugin.Task) error {
	tagsJSON, err := json.Marshal(task.Tags)
	if err != nil {
		return fmt.Errorf("failed to encode tags: %w", err)
	}

	sourcesJSON, codeExamplesJSON, extraJSON := splitExtra(task.Extra)

	var parentID interface{}
	if task.ParentID != nil {
		parentID = *task.ParentID
	}

	var dueDate interface{}
	if task.DueDate != nil {
		dueDate = formatTime(*task.DueDate)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin task insert transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			// The transaction already failed; rollback is best-effort.
			_ = tx.Rollback()
		}
	}()

	// Truncate BEFORE comparing: a raw +500ns offset would pass a naive
	// After check yet format to the same microsecond string — a tie.
	createdAt := task.CreatedAt.Truncate(time.Microsecond)

	// MAX over the fixed-width TEXT layout is chronological; NULL on an
	// empty table means no guard applies.
	var last sql.NullString
	if err := tx.QueryRowContext(ctx, "SELECT MAX(created_at) FROM tasks").Scan(&last); err != nil {
		return fmt.Errorf("failed to read max created_at: %w", err)
	}
	if last.Valid {
		if prev, perr := parseTime(last.String); perr == nil && !createdAt.After(prev) {
			createdAt = prev.Add(time.Microsecond)
		}
	}

	// The caller returns this struct (create_task responses included), so it
	// must show exactly what was stored.
	task.CreatedAt, task.UpdatedAt = createdAt, createdAt
	now := formatTime(createdAt)

	_, err = tx.ExecContext(ctx,
		"INSERT INTO tasks (id, project_id, parent_id, title, description, status, priority, assignee, tags, due_date, sources, code_examples, extra, archived, archived_at, created_at, updated_at, worktree) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, NULL, ?, ?, ?)",
		task.ID, task.ProjectID, parentID, task.Title, task.Description, task.Status, task.Priority, task.Assignee,
		string(tagsJSON), dueDate, sourcesJSON, codeExamplesJSON, extraJSON, now, now, task.Worktree,
	)
	if err != nil {
		return fmt.Errorf("failed to insert task: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit task insert: %w", err)
	}
	committed = true

	return nil
}

// addTaskScalarSets maps the plain scalar fields of an update request.
func addTaskScalarSets(builder *updateBuilder, request plugin.UpdateTaskRequest) {
	if request.Title != nil {
		builder.add("title = ?", *request.Title)
	}

	if request.Description != nil {
		builder.add("description = ?", *request.Description)
	}

	if request.Status != nil {
		builder.add("status = ?", *request.Status)
	}

	if request.Priority != nil {
		builder.add("priority = ?", *request.Priority)
	}

	if request.Assignee != nil {
		builder.add("assignee = ?", *request.Assignee)
	}

	if request.Worktree != nil {
		builder.add("worktree = ?", *request.Worktree)
	}
}

// taskUpdateSets translates non-nil update fields into SET fragments,
// starting from the mandatory updated_at bump.
func taskUpdateSets(request plugin.UpdateTaskRequest) (*updateBuilder, error) {
	builder := &updateBuilder{}
	builder.add("updated_at = ?", formatTime(nowFunc()))

	addTaskScalarSets(builder, request)

	if request.Tags != nil {
		tagsJSON, err := json.Marshal(*request.Tags)
		if err != nil {
			return nil, fmt.Errorf("failed to encode tags: %w", err)
		}
		builder.add("tags = ?", string(tagsJSON))
	}

	if request.DueDate != nil {
		if *request.DueDate == nil {
			builder.addRaw("due_date = NULL")
		} else {
			builder.add("due_date = ?", formatTime(**request.DueDate))
		}
	}

	if request.ProjectID != nil {
		builder.add("project_id = ?", *request.ProjectID)
	}

	if request.ParentID != nil {
		if *request.ParentID == "" {
			builder.addRaw("parent_id = NULL")
		} else {
			builder.add("parent_id = ?", *request.ParentID)
		}
	}

	if request.Archived != nil {
		builder.add("archived = ?", *request.Archived)

		if *request.Archived {
			builder.add("archived_at = ?", formatTime(nowFunc()))
		} else {
			builder.addRaw("archived_at = NULL")
		}
	}

	if request.Extra != nil {
		sourcesJSON, codeExamplesJSON, extraJSON := splitExtra(*request.Extra)
		builder.add("sources = ?", sourcesJSON)
		builder.add("code_examples = ?", codeExamplesJSON)
		builder.add("extra = ?", extraJSON)
	}

	return builder, nil
}

// updateTask applies non-nil request fields and bumps updated_at, then
// returns the re-read task.
func (s *store) updateTask(ctx context.Context, taskID string, request plugin.UpdateTaskRequest) (*plugin.Task, error) {
	builder, err := taskUpdateSets(request)
	if err != nil {
		return nil, err
	}

	args := builder.argsWithWhere(taskID)

	// A dependency replacement shares one transaction with the row update so
	// a late FK failure can't leave new edges unwritten next to an updated
	// row (the migrate() precedent).
	if request.BlockedBy != nil {
		return s.updateTaskWithDependencies(ctx, taskID, *request.BlockedBy, builder.clause(), args)
	}

	result, err := s.db.ExecContext(ctx,
		//nolint:gosec // G202: the builder holds only fixed "column = ?" literals
		// chosen per non-nil request field; every value is ?-bound in args.
		"UPDATE tasks SET "+builder.clause()+" WHERE id = ?", args...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update task %s: %w", taskID, err)
	}

	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return nil, fmt.Errorf("failed to update task %s: %w", taskID, sql.ErrNoRows)
	}

	return s.getTask(ctx, taskID)
}

// updateTaskWithDependencies applies the row update and swaps the dependency
// set inside one transaction, re-reading the task (with its new edges) from
// that same transaction before commit.
func (s *store) updateTaskWithDependencies(ctx context.Context, taskID string, blockedBy []string, setClause string, args []interface{}) (*plugin.Task, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin task update transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			// The transaction already failed; rollback is best-effort.
			_ = tx.Rollback()
		}
	}()

	result, err := tx.ExecContext(ctx,
		//nolint:gosec // G202: setClause holds only fixed "column = ?" literals
		// from taskUpdateSets; every value is ?-bound in args.
		"UPDATE tasks SET "+setClause+" WHERE id = ?", args...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update task %s: %w", taskID, err)
	}

	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return nil, fmt.Errorf("failed to update task %s: %w", taskID, sql.ErrNoRows)
	}

	if err := replaceTaskDependencies(ctx, tx, taskID, blockedBy); err != nil {
		return nil, err
	}

	// scanTask's row parameter is satisfied by transaction rows too.
	row := tx.QueryRowContext(ctx, "SELECT "+taskColumns+" FROM tasks WHERE id = ?", taskID)
	task, err := scanTask(row)
	if err != nil {
		return nil, fmt.Errorf("failed to get task %s: %w", taskID, err)
	}

	// Edge queries go through the tx — the pool's single connection is held
	// by this transaction until commit.
	newBlockedBy, blocks, err := taskEdgesFrom(ctx, tx, taskID)
	if err != nil {
		return nil, err
	}
	task.BlockedBy = newBlockedBy
	task.Blocks = blocks

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit task update %s: %w", taskID, err)
	}
	committed = true

	return &task, nil
}

// deleteTask hard-deletes a task; child tasks keep existing with their
// parent cleared (ON DELETE SET NULL).
func (s *store) deleteTask(ctx context.Context, taskID string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = ?", taskID)
	if err != nil {
		return fmt.Errorf("failed to delete task %s: %w", taskID, err)
	}

	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return fmt.Errorf("failed to delete task %s: %w", taskID, sql.ErrNoRows)
	}

	return nil
}

// taskParent returns the parent ID of a task ("" when it has none).
func (s *store) taskParent(ctx context.Context, taskID string) (string, error) {
	var parent sql.NullString
	if err := s.db.QueryRowContext(ctx, "SELECT parent_id FROM tasks WHERE id = ?", taskID).Scan(&parent); err != nil {
		return "", fmt.Errorf("failed to read parent of task %s: %w", taskID, err)
	}

	if !parent.Valid {
		return "", nil
	}

	return parent.String, nil
}

// wouldCreateCycle reports whether setting newParentID as the parent of
// taskID would introduce a parent cycle.
func (s *store) wouldCreateCycle(ctx context.Context, taskID, newParentID string) (bool, error) {
	current := newParentID
	for range maxParentWalk {
		if current == "" {
			return false, nil
		}

		if current == taskID {
			return true, nil
		}

		next, err := s.taskParent(ctx, current)
		if err != nil {
			return false, err
		}

		current = next
	}

	return true, nil
}

// nowFunc is a variable so tests can pin timestamps.
var nowFunc = time.Now
