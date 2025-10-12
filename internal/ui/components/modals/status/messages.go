package status

import tea "github.com/charmbracelet/bubbletea"

// ShowStatusModalMsg is sent when the status modal should be shown
type ShowStatusModalMsg struct{}

// HideStatusModalMsg is sent when the status modal should be hidden
type HideStatusModalMsg struct{}

// StatusModalShownMsg is sent when the status modal has been shown and is active
type StatusModalShownMsg struct{}

// StatusModalHiddenMsg is sent when the status modal has been hidden and is inactive
type StatusModalHiddenMsg struct{}

// StatusSelectedMsg is sent when a status has been selected and confirmed
type StatusSelectedMsg struct {
	Status string // The selected status: "todo", "doing", "review", "done"
	TaskID string // The ID of the task to update
}

// StatusModalScrollMsg is sent for internal scrolling within the modal
type StatusModalScrollMsg struct {
	Direction int // -1 for up, 1 for down
}

// Compile-time check to ensure our messages implement tea.Msg
var (
	_ tea.Msg = ShowStatusModalMsg{}
	_ tea.Msg = HideStatusModalMsg{}
	_ tea.Msg = StatusModalShownMsg{}
	_ tea.Msg = StatusModalHiddenMsg{}
	_ tea.Msg = StatusSelectedMsg{}
	_ tea.Msg = StatusModalScrollMsg{}
)
