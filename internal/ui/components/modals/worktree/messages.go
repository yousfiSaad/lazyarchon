package worktree

import tea "github.com/charmbracelet/bubbletea"

// ShowWorktreeModalMsg is sent to show the worktree selection modal
type ShowWorktreeModalMsg struct {
	AllWorktrees      []string        // All available worktrees
	SelectedWorktrees map[string]bool // Currently selected worktrees
}

// HideWorktreeModalMsg is sent to hide the worktree selection modal
type HideWorktreeModalMsg struct{}

// WorktreeModalShownMsg is sent when the worktree modal has been shown
type WorktreeModalShownMsg struct{}

// WorktreeModalHiddenMsg is sent when the worktree modal has been hidden
type WorktreeModalHiddenMsg struct{}

// WorktreeSelectionAppliedMsg is sent when worktree selection is applied
type WorktreeSelectionAppliedMsg struct {
	SelectedWorktrees map[string]bool // Final selected worktrees
}

// WorktreeModalSearchMsg is sent when search query changes
type WorktreeModalSearchMsg struct {
	Query string // Search query
}

// WorktreeModalScrollMsg is sent for scrolling within the modal
type WorktreeModalScrollMsg struct {
	Direction int // -1 for up, 1 for down
}

// WorktreeModalToggleMsg is sent to toggle a specific worktree
type WorktreeModalToggleMsg struct {
	Worktree string // Worktree to toggle
}

// WorktreeModalClearSearchMsg is sent to clear the search
type WorktreeModalClearSearchMsg struct{}

// WorktreeModalSelectAllMsg is sent to select all visible worktrees
type WorktreeModalSelectAllMsg struct{}

// WorktreeModalDeselectAllMsg is sent to deselect all worktrees
type WorktreeModalDeselectAllMsg struct{}

// Ensure all message types implement tea.Msg
var (
	_ tea.Msg = ShowWorktreeModalMsg{}
	_ tea.Msg = HideWorktreeModalMsg{}
	_ tea.Msg = WorktreeModalShownMsg{}
	_ tea.Msg = WorktreeModalHiddenMsg{}
	_ tea.Msg = WorktreeSelectionAppliedMsg{}
	_ tea.Msg = WorktreeModalSearchMsg{}
	_ tea.Msg = WorktreeModalScrollMsg{}
	_ tea.Msg = WorktreeModalToggleMsg{}
	_ tea.Msg = WorktreeModalClearSearchMsg{}
	_ tea.Msg = WorktreeModalSelectAllMsg{}
	_ tea.Msg = WorktreeModalDeselectAllMsg{}
)
