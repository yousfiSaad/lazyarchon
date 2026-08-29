package mcpserver

import (
	"fmt"
	"time"

	"github.com/yousfisaad/lazyarchon/v2/internal/plugin"
)

// ---------- output shapes ----------
//
// List tools return compact summaries (no long text); get/mutation tools
// return full objects. Everything optional is omitempty so backends that do
// not populate a field simply omit it.

type projectSummary struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Pinned bool   `json:"pinned,omitempty"`
}

type projectDetail struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	Color       string                 `json:"color,omitempty"`
	Pinned      bool                   `json:"pinned,omitempty"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

type taskSummary struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Status    string   `json:"status"`
	Priority  int      `json:"priority"`
	ProjectID string   `json:"project_id,omitempty"`
	ParentID  *string  `json:"parent_id,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	DueDate   *string  `json:"due_date,omitempty"`
	Archived  bool     `json:"archived,omitempty"`
}

type taskDetail struct {
	ID          string                 `json:"id"`
	ProjectID   string                 `json:"project_id"`
	ParentID    *string                `json:"parent_id,omitempty"`
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	Status      string                 `json:"status"`
	Priority    int                    `json:"priority"`
	Assignee    string                 `json:"assignee,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	DueDate     *string                `json:"due_date,omitempty"`
	Archived    bool                   `json:"archived,omitempty"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

// ---------- conversions ----------

func toProjectSummary(project plugin.Project) projectSummary {
	return projectSummary{ID: project.ID, Title: project.Title, Pinned: project.Pinned}
}

func toProjectDetail(project plugin.Project) projectDetail {
	return projectDetail{
		ID:          project.ID,
		Title:       project.Title,
		Description: project.Description,
		Color:       project.Color,
		Pinned:      project.Pinned,
		CreatedAt:   project.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   project.UpdatedAt.Format(time.RFC3339),
		Extra:       project.Extra,
	}
}

func toTaskSummary(task plugin.Task) taskSummary {
	summary := taskSummary{
		ID:        task.ID,
		Title:     task.Title,
		Status:    task.Status,
		Priority:  task.Priority,
		ProjectID: task.ProjectID,
		ParentID:  task.ParentID,
		Tags:      task.Tags,
		Archived:  task.Archived,
	}

	if task.DueDate != nil {
		formatted := task.DueDate.Format(time.RFC3339)
		summary.DueDate = &formatted
	}

	return summary
}

func toTaskDetail(task plugin.Task) taskDetail {
	detail := taskDetail{
		ID:          task.ID,
		ProjectID:   task.ProjectID,
		ParentID:    task.ParentID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Priority:    task.Priority,
		Assignee:    task.Assignee,
		Tags:        task.Tags,
		Archived:    task.Archived,
		CreatedAt:   task.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   task.UpdatedAt.Format(time.RFC3339),
		Extra:       task.Extra,
	}

	if task.DueDate != nil {
		formatted := task.DueDate.Format(time.RFC3339)
		detail.DueDate = &formatted
	}

	return detail
}

// ---------- date handling ----------

// parseDueDate accepts "YYYY-MM-DD" (treated as UTC midnight) or a full
// RFC3339 timestamp.
func parseDueDate(value string) (time.Time, error) {
	if date, err := time.Parse("2006-01-02", value); err == nil {
		return date, nil
	}

	if stamp, err := time.Parse(time.RFC3339, value); err == nil {
		return stamp, nil
	}

	return time.Time{}, fmt.Errorf("invalid date %q: use YYYY-MM-DD or RFC3339", value)
}
