package archon

import (
	"fmt"
	"strings"
	"time"
)

// FlexibleTime handles multiple timestamp formats from the API
type FlexibleTime struct {
	time.Time
}

// UnmarshalJSON implements custom JSON unmarshaling for flexible timestamp parsing
func (ft *FlexibleTime) UnmarshalJSON(data []byte) error {
	// Remove quotes from the JSON string
	s := strings.Trim(string(data), `"`)

	// Try multiple timestamp formats that the API might return
	formats := []string{
		"2006-01-02T15:04:05.999999Z07:00", // RFC3339 with timezone and microseconds
		"2006-01-02T15:04:05.999999",       // Without timezone, with microseconds
		"2006-01-02T15:04:05Z07:00",        // RFC3339 with timezone, no microseconds
		"2006-01-02T15:04:05",              // Simple format without timezone
		"2006-01-02 15:04:05.999999",       // Alternative format with space
		"2006-01-02 15:04:05",              // Alternative format with space, no microseconds
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			ft.Time = t
			return nil
		}
	}

	return fmt.Errorf("unable to parse timestamp: %s", s)
}

// MarshalJSON implements custom JSON marshaling (returns RFC3339 format)
func (ft FlexibleTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + ft.Time.Format(time.RFC3339) + `"`), nil
}

// Project represents an Archon project
type Project struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	GitHubRepo  *string                `json:"github_repo"`
	CreatedAt   FlexibleTime           `json:"created_at"`
	UpdatedAt   FlexibleTime           `json:"updated_at"`
	Docs        []Document             `json:"docs"`
	Features    map[string]interface{} `json:"features"`
	Data        map[string]interface{} `json:"data"`
	Pinned      bool                   `json:"pinned"`
}

// Task represents an Archon task
type Task struct {
	ID           string        `json:"id"`
	ProjectID    string        `json:"project_id"`
	ParentTaskID *string       `json:"parent_task_id"`
	Title        string        `json:"title"`
	Description  string        `json:"description"`
	Status       string        `json:"status"` // todo, doing, review, done
	Assignee     string        `json:"assignee"`
	TaskOrder    int           `json:"task_order"`
	Feature      *string       `json:"feature"`
	Sources      []Source      `json:"sources"`
	CodeExamples []CodeExample `json:"code_examples"`
	Archived     bool          `json:"archived"`
	ArchivedAt   *FlexibleTime `json:"archived_at"`
	ArchivedBy   *string       `json:"archived_by"`
	CreatedAt    FlexibleTime  `json:"created_at"`
	UpdatedAt    FlexibleTime  `json:"updated_at"`
}

// Document represents an Archon document
type Document struct {
	ID        string                 `json:"id"`
	ProjectID string                 `json:"project_id"`
	Title     string                 `json:"title"`
	Type      string                 `json:"document_type"`
	Content   map[string]interface{} `json:"content"`
	Tags      []string               `json:"tags"`
	Author    *string                `json:"author"`
	CreatedAt FlexibleTime           `json:"created_at"`
	UpdatedAt FlexibleTime           `json:"updated_at"`
}

// Source represents a source reference for tasks
type Source struct {
	URL       string `json:"url"`
	Type      string `json:"type"`
	Relevance string `json:"relevance"`
}

// CodeExample represents a code example reference for tasks
type CodeExample struct {
	File     string `json:"file"`
	Function string `json:"function"`
	Purpose  string `json:"purpose"`
}

// TasksResponse represents the API response for listing tasks
type TasksResponse struct {
	Success bool   `json:"success"`
	Tasks   []Task `json:"tasks"`
	Count   int    `json:"count"`
	Page    int    `json:"page"`
	PerPage int    `json:"per_page"`
	Error   string `json:"error,omitempty"`
}

// TaskResponse represents the API response for a single task
type TaskResponse struct {
	Success bool   `json:"success"`
	Task    Task   `json:"task"`
	TaskID  string `json:"task_id"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// ProjectsResponse represents the API response for listing projects
type ProjectsResponse struct {
	Success  bool      `json:"success"`
	Projects []Project `json:"projects"`
	Count    int       `json:"count"`
	Error    string    `json:"error,omitempty"`
}

// ProjectResponse represents the API response for a single project
type ProjectResponse struct {
	Success   bool    `json:"success"`
	Project   Project `json:"project"`
	ProjectID string  `json:"project_id"`
	Message   string  `json:"message"`
	Error     string  `json:"error,omitempty"`
}

// UpdateTaskRequest represents a request to update a task
type UpdateTaskRequest struct {
	Title        *string        `json:"title,omitempty"`
	Description  *string        `json:"description,omitempty"`
	Status       *string        `json:"status,omitempty"`
	Assignee     *string        `json:"assignee,omitempty"`
	TaskOrder    *int           `json:"task_order,omitempty"`
	Feature      *string        `json:"feature,omitempty"`
	Sources      *[]Source      `json:"sources,omitempty"`
	CodeExamples *[]CodeExample `json:"code_examples,omitempty"`
}

// TaskStatus constants
const (
	TaskStatusTodo   = "todo"
	TaskStatusDoing  = "doing"
	TaskStatusReview = "review"
	TaskStatusDone   = "done"
)

// GetStatusColor returns a color code for the task status
func (t Task) GetStatusColor() string {
	switch t.Status {
	case TaskStatusTodo:
		return "240" // gray
	case TaskStatusDoing:
		return "33" // yellow
	case TaskStatusReview:
		return "34" // blue
	case TaskStatusDone:
		return "32" // green
	default:
		return "37" // white
	}
}

// GetStatusSymbol returns a symbol for the task status
func (t Task) GetStatusSymbol() string {
	switch t.Status {
	case TaskStatusTodo:
		return "○"
	case TaskStatusDoing:
		return "◐"
	case TaskStatusReview:
		return "◉"
	case TaskStatusDone:
		return "●"
	default:
		return "?"
	}
}
