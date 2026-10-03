package help

import tea "github.com/charmbracelet/bubbletea"

// ShowHelpModalMsg is sent when the help modal should be shown
type ShowHelpModalMsg struct{}

// HideHelpModalMsg is sent when the help modal should be hidden
type HideHelpModalMsg struct{}

// HelpModalShownMsg is sent when the help modal has been shown
type HelpModalShownMsg struct{}

// HelpModalHiddenMsg is sent when the help modal has been hidden
type HelpModalHiddenMsg struct{}

// HelpModalScrollMsg handles scrolling within the help modal
type HelpModalScrollMsg struct {
	Direction ScrollDirection
	Amount    int
}

type ScrollDirection int

const (
	ScrollUp ScrollDirection = iota
	ScrollDown
	ScrollToTop
	ScrollToBottom
	ScrollHalfUp
	ScrollHalfDown
)

// HelpModalKeyMsg represents a key press specifically for the help modal
type HelpModalKeyMsg struct {
	Key tea.KeyMsg
}
