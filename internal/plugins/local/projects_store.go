package local

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/yousfisaad/lazyarchon/v2/internal/plugin"
)

// listProjects returns all projects, pinned first then alphabetical.
func (s *store) listProjects(ctx context.Context) ([]plugin.Project, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+projectColumns+" FROM projects ORDER BY pinned DESC, title ASC")
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	// Collect rows fully before any further query (single connection).
	var projects []plugin.Project
	for rows.Next() {
		p, err := s.scanProject(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate projects: %w", err)
	}

	if projects == nil {
		projects = []plugin.Project{}
	}

	return projects, nil
}

// getProject returns a single project by ID. Returns sql.ErrNoRows (wrapped)
// when the project does not exist.
func (s *store) getProject(ctx context.Context, projectID string) (*plugin.Project, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+projectColumns+" FROM projects WHERE id = ?", projectID)

	project, err := s.scanProject(row)
	if err != nil {
		return nil, fmt.Errorf("failed to get project %s: %w", projectID, err)
	}

	return &project, nil
}

// insertProject persists a new project row.
func (s *store) insertProject(ctx context.Context, proj *plugin.Project) error {
	_, _, extraJSON := splitExtra(proj.Extra)

	now := formatTime(proj.CreatedAt)
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO projects (id, title, description, color, pinned, extra, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		proj.ID, proj.Title, proj.Description, proj.Color, proj.Pinned, extraJSON, now, now,
	)
	if err != nil {
		return fmt.Errorf("failed to insert project: %w", err)
	}

	return nil
}

// updateProject applies non-nil fields and bumps updated_at.
func (s *store) updateProject(ctx context.Context, projectID string, request plugin.UpdateProjectRequest) (*plugin.Project, error) {
	builder := &updateBuilder{}
	builder.add("updated_at = ?", formatTime(nowFunc()))

	if request.Title != nil {
		builder.add("title = ?", *request.Title)
	}

	if request.Description != nil {
		builder.add("description = ?", *request.Description)
	}

	if request.Color != nil {
		builder.add("color = ?", *request.Color)
	}

	if request.Pinned != nil {
		builder.add("pinned = ?", *request.Pinned)
	}

	if request.Extra != nil {
		_, _, extraJSON := splitExtra(*request.Extra)
		builder.add("extra = ?", extraJSON)
	}

	args := builder.argsWithWhere(projectID)

	result, err := s.db.ExecContext(ctx,
		//nolint:gosec // G202: the builder holds only fixed "column = ?" literals
		// chosen per non-nil request field; every value is ?-bound in args.
		"UPDATE projects SET "+builder.clause()+" WHERE id = ?", args...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update project %s: %w", projectID, err)
	}

	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return nil, fmt.Errorf("failed to update project %s: %w", projectID, sql.ErrNoRows)
	}

	return s.getProject(ctx, projectID)
}

// deleteProject removes a project; its tasks cascade at the database level.
func (s *store) deleteProject(ctx context.Context, projectID string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM projects WHERE id = ?", projectID)
	if err != nil {
		return fmt.Errorf("failed to delete project %s: %w", projectID, err)
	}

	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return fmt.Errorf("failed to delete project %s: %w", projectID, sql.ErrNoRows)
	}

	return nil
}

// projectExists reports whether the project ID is present.
func (s *store) projectExists(ctx context.Context, projectID string) (bool, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM projects WHERE id = ?", projectID).Scan(&count); err != nil {
		return false, fmt.Errorf("failed to check project %s: %w", projectID, err)
	}

	return count > 0, nil
}
