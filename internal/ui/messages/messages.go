package messages

import tea "github.com/charmbracelet/bubbletea"

// =============================================================================
// APPLICATION STATE MESSAGES
// =============================================================================
// Global application state messages broadcast to all interested components

// LoadingStateMsg is broadcast when loading state changes
type LoadingStateMsg struct {
	Loading bool
	Message string
}

// ErrorStateMsg is broadcast when error state changes
type ErrorStateMsg struct {
	Error string
}

// ConnectionStatusMsg is broadcast when connection status changes
type ConnectionStatusMsg struct {
	Connected bool
}

// ProjectModeMsg is broadcast when project mode is toggled
type ProjectModeMsg struct {
	Active bool
}

// ActiveViewMsg is broadcast when active view/panel changes
type ActiveViewMsg struct {
	ViewName string
}

// =============================================================================
// DATA STATE MESSAGES
// =============================================================================
// Messages about application data state (tasks, projects, features)
//
// NOTE: Display parameter messages removed - components compute display data on-demand:
// - TaskCountsMsg (removed) → Components call ctx.ProgramContext.GetTaskStatusCounts()
// - SelectionPositionMsg (removed) → Components call ctx.UIState.GetSelectedTaskIndex() and len(GetSortedTasks())
// - SortModeMsg (removed) → Components call ctx.ProgramContext.GetCurrentSortModeName()
// - FeatureCountMsg (removed) → Components call len(ctx.ProgramContext.GetUniqueFeatures())
// - ProjectCountMsg (removed) → Components call len(ctx.ProgramContext.Projects)
// - ProjectDisplayMsg (removed) → Components call ctx.ProgramContext.GetCurrentProjectName()
// - FeatureDisplayMsg (removed) → Components call ctx.ProgramContext.GetFeatureFilterSummary()
// - ProjectTaskCountsMsg (removed) → Components call ctx.GetTaskCountForProject(projectID) and ctx.GetTotalTaskCount()

// =============================================================================
// SEARCH MESSAGES
// =============================================================================
// Messages related to search functionality

// SearchStateChangedMsg is broadcast when search state changes globally
// All components interested in search state should handle this message
type SearchStateChangedMsg struct {
	Query  string // Current search query
	Active bool   // Whether search is active
}

// SearchModeMsg is broadcast when search input mode is toggled
type SearchModeMsg struct {
	Active bool
	Input  string
}

// SearchMatchInfoMsg is broadcast when search match information changes
type SearchMatchInfoMsg struct {
	CurrentMatch int
	TotalMatches int
}

// =============================================================================
// MODAL MESSAGES
// =============================================================================
// Messages related to modal lifecycle and state

// ModalStateMsg is broadcast when modal visibility changes
// Replaces component-specific XxxModalShownMsg/XxxModalHiddenMsg messages
type ModalStateMsg struct {
	Type   string // Modal type identifier (from base.ModalType)
	Active bool   // true = shown, false = hidden
}

// =============================================================================
// POLLING MESSAGES
// =============================================================================
// Messages related to HTTP polling for auto-refresh

// PollingTickMsg triggers a polling cycle to refresh tasks/projects via HTTP
// This replaces WebSocket real-time updates when backend doesn't support WebSocket
type PollingTickMsg struct{}

// =============================================================================
// USER INTERACTION MESSAGES
// =============================================================================
// Messages triggered by user actions

// YankIDMsg requests the active component to copy its ID to clipboard
// This message is sent when user presses 'y' key
type YankIDMsg struct{}

// YankTitleMsg requests the active component to copy its title to clipboard
// This message is sent when user presses 'Y' key
type YankTitleMsg struct{}

// StatusFeedbackMsg provides UI feedback from components
// Components send this message to display status/success/error messages
type StatusFeedbackMsg struct {
	Message string
}

// =============================================================================
// INTERFACE VERIFICATION
// =============================================================================
// Ensure all messages implement tea.Msg interface

var (
	// Application state messages
	_ tea.Msg = LoadingStateMsg{}
	_ tea.Msg = ErrorStateMsg{}
	_ tea.Msg = ConnectionStatusMsg{}
	_ tea.Msg = ProjectModeMsg{}
	_ tea.Msg = ActiveViewMsg{}

	// NOTE: Data state messages removed - components compute display data on-demand:
	// TaskCountsMsg, SelectionPositionMsg, SortModeMsg, FeatureCountMsg, ProjectCountMsg,
	// ProjectDisplayMsg, FeatureDisplayMsg, ProjectTaskCountsMsg

	// Search messages
	_ tea.Msg = SearchStateChangedMsg{}
	_ tea.Msg = SearchModeMsg{}
	_ tea.Msg = SearchMatchInfoMsg{}

	// Modal messages
	_ tea.Msg = ModalStateMsg{}

	// Polling messages
	_ tea.Msg = PollingTickMsg{}

	// User interaction messages
	_ tea.Msg = YankIDMsg{}
	_ tea.Msg = YankTitleMsg{}
	_ tea.Msg = StatusFeedbackMsg{}
)
