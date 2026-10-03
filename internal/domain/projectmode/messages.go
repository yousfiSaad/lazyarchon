package projectmode

import tea "github.com/charmbracelet/bubbletea"

// =============================================================================
// PROJECT MODE DOMAIN MESSAGES
// =============================================================================
// Messages representing project mode feature state transitions

// ProjectModeActivatedMsg is sent when project mode should be activated
type ProjectModeActivatedMsg struct{}

// ProjectModeDeactivatedMsg is sent when project mode should be deactivated
type ProjectModeDeactivatedMsg struct {
	ShouldLoadTasks bool // Whether to load tasks after deactivation
}

// Ensure all message types implement tea.Msg
var (
	_ tea.Msg = ProjectModeActivatedMsg{}
	_ tea.Msg = ProjectModeDeactivatedMsg{}
)
