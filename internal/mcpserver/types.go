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

func toProjectSummary(p plugin.Project) projectSummary {
	return projectSummary{ID: p.ID, Title: p.Title, Pinned: p.Pinned}
}

func toProjectDetail(p plugin.Project) projectDetail {
	return projectDetail{
		ID:          p.ID,
		Title:       p.Title,
		Description: p.Description,
		Color:       p.Color,
		Pinned:      p.Pinned,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
		Extra:       p.Extra,
	}
}

func toTaskSummary(t plugin.Task) taskSummary {
	summary := taskSummary{
		ID:        t.ID,
		Title:     t.Title,
		Status:    t.Status,
		Priority:  t.Priority,
		ProjectID: t.ProjectID,
		ParentID:  t.ParentID,
		Tags:      t.Tags,
		Archived:  t.Archived,
	}

	if t.DueDate != nil {
		formatted := t.DueDate.Format(time.RFC3339)
		summary.DueDate = &formatted
	}

	return summary
}

func toTaskDetail(t plugin.Task) taskDetail {
	detail := taskDetail{
		ID:          t.ID,
		ProjectID:   t.ProjectID,
		ParentID:    t.ParentID,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		Priority:    t.Priority,
		Assignee:    t.Assignee,
		Tags:        t.Tags,
		Archived:    t.Archived,
		CreatedAt:   t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   t.UpdatedAt.Format(time.RFC3339),
		Extra:       t.Extra,
	}

	if t.DueDate != nil {
		formatted := t.DueDate.Format(time.RFC3339)
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
