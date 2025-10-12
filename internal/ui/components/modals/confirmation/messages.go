package confirmation

import tea "github.com/charmbracelet/bubbletea"

// ShowConfirmationModalMsg is sent when the confirmation modal should be shown
type ShowConfirmationModalMsg struct {
	Message     string // The confirmation message to display
	ConfirmText string // Text for the confirm button (default: "Yes")
	CancelText  string // Text for the cancel button (default: "No")
}

// HideConfirmationModalMsg is sent when the confirmation modal should be hidden
type HideConfirmationModalMsg struct{}

// ConfirmationModalShownMsg is sent when the confirmation modal has been shown and is active
type ConfirmationModalShownMsg struct{}

// ConfirmationModalHiddenMsg is sent when the confirmation modal has been hidden and is inactive
type ConfirmationModalHiddenMsg struct{}

// ConfirmationSelectedMsg is sent when a confirmation choice has been made
type ConfirmationSelectedMsg struct {
	Confirmed bool   // true if confirmed, false if canceled
	Message   string // The original message that was confirmed/canceled
}

// ConfirmationModalScrollMsg is sent for internal navigation within the modal
type ConfirmationModalScrollMsg struct {
	Direction int // -1 for left, 1 for right
}

// Compile-time check to ensure our messages implement tea.Msg
var (
	_ tea.Msg = ShowConfirmationModalMsg{}
	_ tea.Msg = HideConfirmationModalMsg{}
	_ tea.Msg = ConfirmationModalShownMsg{}
	_ tea.Msg = ConfirmationModalHiddenMsg{}
	_ tea.Msg = ConfirmationSelectedMsg{}
	_ tea.Msg = ConfirmationModalScrollMsg{}
)
