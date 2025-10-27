package taskcreate

import tea "github.com/charmbracelet/bubbletea"

// ShowTaskCreateModalMsg triggers the task creation modal to be shown
type ShowTaskCreateModalMsg struct {
	DefaultProjectID  string   // Project ID to create task in
	AvailableFeatures []string // Available features for selection
	DefaultFeature    string   // Pre-selected feature (if any)
}

// HideTaskCreateModalMsg triggers the task creation modal to be hidden
type HideTaskCreateModalMsg struct{}

// TaskCreateModalShownMsg is sent when the modal becomes active
type TaskCreateModalShownMsg struct{}

// TaskCreateModalHiddenMsg is sent when the modal becomes inactive
type TaskCreateModalHiddenMsg struct{}

// TaskCreatedMsg is sent when a new task has been created successfully
type TaskCreatedMsg struct {
	Title       string  // Task title
	Description string  // Task description
	ProjectID   string  // Project ID
	Feature     *string // Feature tag (optional)
	Priority    int     // Task priority
	Status      string  // Task status (defaults to "todo")
}

// Ensure all message types implement tea.Msg
var (
	_ tea.Msg = ShowTaskCreateModalMsg{}
	_ tea.Msg = HideTaskCreateModalMsg{}
	_ tea.Msg = TaskCreateModalShownMsg{}
	_ tea.Msg = TaskCreateModalHiddenMsg{}
	_ tea.Msg = TaskCreatedMsg{}
)
