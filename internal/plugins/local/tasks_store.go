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

// taskOrderClause sorts by status lane (todo→doing→review→done), then
// priority (1 = most urgent first), then newest first, with id as a
// deterministic tiebreaker for stable pagination.
const taskOrderClause = "ORDER BY CASE status WHEN 'todo' THEN 0 WHEN 'doing' THEN 1 WHEN 'review' THEN 2 ELSE 3 END, priority ASC, created_at DESC, id ASC"

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

// buildTaskFilter translates generic filters into SQL conditions.
func buildTaskFilter(filters plugin.TaskFilters) *filterBuilder {
	b := &filterBuilder{}

	if filters.ProjectID != nil && *filters.ProjectID != "" {
		b.add("project_id = ?", *filters.ProjectID)
	}

	if filters.Status != nil && *filters.Status != "" {
		b.add("status = ?", *filters.Status)
	}

	if filters.Assignee != nil && *filters.Assignee != "" {
		b.add("assignee = ?", *filters.Assignee)
	}

	if !filters.IncludeArchived {
		b.addRaw("archived = 0")
	}

	if filters.SearchQuery != nil && *filters.SearchQuery != "" {
		pattern := "%" + escapeLike(*filters.SearchQuery) + "%"
		b.conditions = append(b.conditions, "(title LIKE ? ESCAPE '\\' OR description LIKE ? ESCAPE '\\')")
		b.args = append(b.args, pattern, pattern)
	}

	if len(filters.Tags) > 0 {
		placeholders := make([]string, len(filters.Tags))
		for i, tag := range filters.Tags {
			placeholders[i] = "?"
			b.args = append(b.args, tag)
		}
		b.addRaw("EXISTS (SELECT 1 FROM json_each(tasks.tags) AS jt WHERE jt.value IN (" + strings.Join(placeholders, ",") + "))")
	}

	if filters.DueBefore != nil {
		b.add("due_date IS NOT NULL AND due_date <= ?", formatTime(*filters.DueBefore))
	}

	if filters.DueAfter != nil {
		b.add("due_date IS NOT NULL AND due_date >= ?", formatTime(*filters.DueAfter))
	}

	return b
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
		"SELECT "+taskColumns+" FROM tasks "+where.whereClause()+" "+taskOrderClause+" LIMIT ? OFFSET ?",
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

// getTask returns a single task by ID. Returns sql.ErrNoRows (wrapped) when
// the task does not exist.
func (s *store) getTask(ctx context.Context, taskID string) (*plugin.Task, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+taskColumns+" FROM tasks WHERE id = ?", taskID)

	task, err := scanTask(row)
	if err != nil {
		return nil, fmt.Errorf("failed to get task %s: %w", taskID, err)
	}

	return &task, nil
}

// insertTask persists a new task row.
func (s *store) insertTask(ctx context.Context, t *plugin.Task) error {
	tagsJSON, err := json.Marshal(t.Tags)
	if err != nil {
		return fmt.Errorf("failed to encode tags: %w", err)
	}

	sourcesJSON, codeExamplesJSON, extraJSON := splitExtra(t.Extra)

	var parentID interface{}
	if t.ParentID != nil {
		parentID = *t.ParentID
	}

	var dueDate interface{}
	if t.DueDate != nil {
		dueDate = formatTime(*t.DueDate)
	}

	now := formatTime(t.CreatedAt)

	_, err = s.db.ExecContext(ctx,
		"INSERT INTO tasks (id, project_id, parent_id, title, description, status, priority, assignee, tags, due_date, sources, code_examples, extra, archived, archived_at, created_at, updated_at) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, NULL, ?, ?)",
		t.ID, t.ProjectID, parentID, t.Title, t.Description, t.Status, t.Priority, t.Assignee,
		string(tagsJSON), dueDate, sourcesJSON, codeExamplesJSON, extraJSON, now, now,
	)
	if err != nil {
		return fmt.Errorf("failed to insert task: %w", err)
	}

	return nil
}

// updateTask applies non-nil request fields and bumps updated_at, then
// returns the re-read task.
func (s *store) updateTask(ctx context.Context, taskID string, request plugin.UpdateTaskRequest) (*plugin.Task, error) {
	sets := []string{"updated_at = ?"}
	args := []interface{}{formatTime(nowFunc())}

	if request.Title != nil {
		sets = append(sets, "title = ?")
		args = append(args, *request.Title)
	}

	if request.Description != nil {
		sets = append(sets, "description = ?")
		args = append(args, *request.Description)
	}

	if request.Status != nil {
		sets = append(sets, "status = ?")
		args = append(args, *request.Status)
	}

	if request.Priority != nil {
		sets = append(sets, "priority = ?")
		args = append(args, *request.Priority)
	}

	if request.Assignee != nil {
		sets = append(sets, "assignee = ?")
		args = append(args, *request.Assignee)
	}

	if request.Tags != nil {
		tagsJSON, err := json.Marshal(*request.Tags)
		if err != nil {
			return nil, fmt.Errorf("failed to encode tags: %w", err)
		}
		sets = append(sets, "tags = ?")
		args = append(args, string(tagsJSON))
	}

	if request.DueDate != nil {
		if *request.DueDate == nil {
			sets = append(sets, "due_date = NULL")
		} else {
			sets = append(sets, "due_date = ?")
			args = append(args, formatTime(**request.DueDate))
		}
	}

	if request.ProjectID != nil {
		sets = append(sets, "project_id = ?")
		args = append(args, *request.ProjectID)
	}

	if request.ParentID != nil {
		if *request.ParentID == "" {
			sets = append(sets, "parent_id = NULL")
		} else {
			sets = append(sets, "parent_id = ?")
			args = append(args, *request.ParentID)
		}
	}

	if request.Archived != nil {
		sets = append(sets, "archived = ?")
		args = append(args, *request.Archived)

		if *request.Archived {
			sets = append(sets, "archived_at = ?")
			args = append(args, formatTime(nowFunc()))
		} else {
			sets = append(sets, "archived_at = NULL")
		}
	}

	if request.Extra != nil {
		sourcesJSON, codeExamplesJSON, extraJSON := splitExtra(*request.Extra)
		sets = append(sets, "sources = ?", "code_examples = ?", "extra = ?")
		args = append(args, sourcesJSON, codeExamplesJSON, extraJSON)
	}

	args = append(args, taskID)

	result, err := s.db.ExecContext(ctx,
		"UPDATE tasks SET "+strings.Join(sets, ", ")+" WHERE id = ?", args...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update task %s: %w", taskID, err)
	}

	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return nil, fmt.Errorf("failed to update task %s: %w", taskID, sql.ErrNoRows)
	}

	return s.getTask(ctx, taskID)
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
