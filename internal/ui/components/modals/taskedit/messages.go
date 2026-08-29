package taskedit

import tea "github.com/charmbracelet/bubbletea"

// FieldType represents which field in the modal has focus
type FieldType int

const (
	FieldStatus FieldType = iota
	FieldPriority
	FieldFeature
)

// Component lifecycle messages

// ShowTaskEditModalMsg is sent to show the task properties modal
type ShowTaskEditModalMsg struct {
	TaskID            string    // ID of task being edited
	CurrentStatus     string    // Current task status (todo, doing, review, done)
	CurrentPriority   int       // Current task priority (task_order value)
	CurrentFeature    string    // Current feature assignment (can be empty)
	FocusField        FieldType // Which field to focus initially
	AvailableFeatures []string  // List of available features to choose from
}

// HideTaskEditModalMsg is sent to hide the task edit modal
type HideTaskEditModalMsg struct{}

// TaskEditModalShownMsg is broadcast when the task edit modal has been shown
type TaskEditModalShownMsg struct{}

// TaskEditModalHiddenMsg is broadcast when the task edit modal has been hidden
type TaskEditModalHiddenMsg struct{}

// Selection and interaction messages

// TaskPropertiesUpdatedMsg is sent when task properties have been updated
// Only non-nil fields are updated on the task
type TaskPropertiesUpdatedMsg struct {
	TaskID   string  // Which task was edited
	Status   *string // New status (nil if unchanged)
	Priority *int    // New priority/task_order (nil if unchanged)
	Feature  *string // New feature (nil if unchanged)
}

// FeatureSelectedMsg is sent when a feature has been selected or created
// Note: New code should use TaskPropertiesUpdatedMsg for unified updates
type FeatureSelectedMsg struct {
	TaskID       string // Which task was edited
	Feature      string // Selected or created feature name
	IsNewFeature bool   // true if user created a new feature, false if selected existing
}

// TaskEditModalScrollMsg is sent for scroll operations within the modal
type TaskEditModalScrollMsg struct {
	Direction int // Positive for down/right, negative for up/left
}

// Ensure all message types implement tea.Msg interface
var (
	_ tea.Msg = ShowTaskEditModalMsg{}
	_ tea.Msg = HideTaskEditModalMsg{}
	_ tea.Msg = TaskEditModalShownMsg{}
	_ tea.Msg = TaskEditModalHiddenMsg{}
	_ tea.Msg = TaskPropertiesUpdatedMsg{}
	_ tea.Msg = FeatureSelectedMsg{}
	_ tea.Msg = TaskEditModalScrollMsg{}
)
