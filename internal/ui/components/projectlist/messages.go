package projectlist

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/v2/internal/archon"
)

// ProjectListUpdateMsg is sent to update the project list data
type ProjectListUpdateMsg struct {
	Projects []archon.Project // Updated project data
	Loading  bool             // Whether loading is in progress
	Error    string           // Error message if any
}

// ProjectListSelectMsg is sent to select a specific project by index
type ProjectListSelectMsg struct {
	Index int // Index to select (can be len(projects) for "All Tasks")
}

// ProjectListSelectionChangedMsg is sent when the selection changes
type ProjectListSelectionChangedMsg struct {
	Index int // New selected index
}

// ScrollDirection represents the direction of scrolling
type ScrollDirection int

const (
	ScrollUp ScrollDirection = iota
	ScrollDown
	ScrollToTop
	ScrollToBottom
)

// ProjectListScrollMsg is sent to handle scrolling operations
type ProjectListScrollMsg struct {
	Direction ScrollDirection
}

// NOTE: ProjectListSetActiveMsg removed - components read active state from UIState directly

// ProjectListSetPanelMsg removed - help functionality moved to global help modal

// Helper functions to create message commands

// UpdateProjectList creates a command to update the project list
func UpdateProjectList(projects []archon.Project, loading bool, error string) tea.Cmd {
	return func() tea.Msg {
		return ProjectListUpdateMsg{
			Projects: projects,
			Loading:  loading,
			Error:    error,
		}
	}
}

// SelectProject creates a command to select a project by index
func SelectProject(index int) tea.Cmd {
	return func() tea.Msg {
		return ProjectListSelectMsg{Index: index}
	}
}

// ScrollProjectList creates a command to scroll the project list
func ScrollProjectList(direction ScrollDirection) tea.Cmd {
	return func() tea.Msg {
		return ProjectListScrollMsg{Direction: direction}
	}
}

// NOTE: SetProjectListActive helper removed - components read active state from UIState directly

// SetProjectListPanel function removed - help functionality moved to global help modal

// ProjectListSelectionQueryMsg is sent to query the current selection index
type ProjectListSelectionQueryMsg struct{}

// ProjectListSelectionResponseMsg is returned in response to a selection query
type ProjectListSelectionResponseMsg struct {
	Index int // Current selected index
}

// ProjectListConfirmSelectionMsg is sent when user confirms selection (e.g., presses Enter)
type ProjectListConfirmSelectionMsg struct {
	Index int // Confirmed selection index
}

// Ensure all message types implement tea.Msg
var (
	_ tea.Msg = ProjectListUpdateMsg{}
	_ tea.Msg = ProjectListSelectMsg{}
	_ tea.Msg = ProjectListSelectionChangedMsg{}
	_ tea.Msg = ProjectListScrollMsg{}
	// NOTE: ProjectListSetActiveMsg interface check removed - message type deleted
	_ tea.Msg = ProjectListSelectionQueryMsg{}
	_ tea.Msg = ProjectListSelectionResponseMsg{}
	_ tea.Msg = ProjectListConfirmSelectionMsg{}
)
