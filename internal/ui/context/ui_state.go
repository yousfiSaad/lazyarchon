package context

// UIState holds transient UI presentation state.
// This is separate from ProgramContext which holds business logic and persistent data.
//
// UIState contains:
// - View/navigation state (which view, which panel is active)
// - Search interaction state (typing, filtering)
// - Selection state (which item is selected)
// - Modal state (what's open)
//
// This state is ephemeral and resets on application restart.
// Components read this state directly - no callbacks needed.
type UIState struct {
	// =============================================================================
	// VIEW & NAVIGATION STATE
	// =============================================================================
	// Controls which view and panel the user is currently interacting with

	// CurrentViewMode determines which major view is displayed (tasks or projects)
	CurrentViewMode ViewMode // TaskViewMode or ProjectViewMode

	// ActivePanel determines which panel has focus (left or right)
	ActivePanel ActivePanel // LeftPanel or RightPanel

	// =============================================================================
	// SEARCH INTERACTION STATE
	// =============================================================================
	// Current search session state - transient, not persisted
	// Note: SearchHistory (persistent) lives in ProgramContext

	// SearchMode indicates whether user is actively typing in search input
	SearchMode bool

	// SearchInput is the text currently being typed in search bar
	SearchInput string

	// SearchActive indicates whether search filtering is applied to results
	SearchActive bool

	// SearchQuery is the active search query string used for filtering
	SearchQuery string

	// =============================================================================
	// SELECTION STATE
	// =============================================================================
	// Tracks which items are currently selected in lists

	// SelectedTaskIndex is the currently selected task index in task list
	SelectedTaskIndex int

	// SelectedProjectIndex is the currently selected project index in project list
	SelectedProjectIndex int

	// =============================================================================
	// COMPUTED SEARCH STATE
	// =============================================================================
	// Derived from search query - recalculated when search changes

	// TaskMatchingIndices holds indices of tasks matching current search
	TaskMatchingIndices []int

	// TaskTotalMatches is the total number of tasks matching search
	TaskTotalMatches int

	// =============================================================================
	// MODAL STATE
	// =============================================================================
	// Tracks temporary UI overlays (future: could be a modal stack)

	// Future: ModalStack []ModalType for managing multiple modals
}

// ActivePanel represents which panel is currently focused for user input
type ActivePanel int

const (
	LeftPanel  ActivePanel = 0 // Left panel (list view)
	RightPanel ActivePanel = 1 // Right panel (details view)
)

// NewUIState creates a new UIState with default values
func NewUIState() *UIState {
	return &UIState{
		CurrentViewMode:      TaskViewMode,
		ActivePanel:          LeftPanel,
		SearchMode:           false,
		SearchInput:          "",
		SearchActive:         false,
		SearchQuery:          "",
		SelectedTaskIndex:    0,
		SelectedProjectIndex: 0,
		TaskMatchingIndices:  make([]int, 0),
		TaskTotalMatches:     0,
	}
}

// SetActivePanel updates which panel is currently focused
func (s *UIState) SetActivePanel(panel ActivePanel) {
	s.ActivePanel = panel
}

// GetActivePanel returns the currently focused panel
func (s *UIState) GetActivePanel() ActivePanel {
	return s.ActivePanel
}

// IsLeftPanelActive returns true if left panel is currently active
func (s *UIState) IsLeftPanelActive() bool {
	return s.ActivePanel == LeftPanel
}

// IsRightPanelActive returns true if right panel is currently active
func (s *UIState) IsRightPanelActive() bool {
	return s.ActivePanel == RightPanel
}

// SetViewMode updates the current view mode
func (s *UIState) SetViewMode(mode ViewMode) {
	s.CurrentViewMode = mode
}

// GetViewMode returns the current view mode
func (s *UIState) GetViewMode() ViewMode {
	return s.CurrentViewMode
}

// IsProjectView returns true if in project selection mode
func (s *UIState) IsProjectView() bool {
	return s.CurrentViewMode == ProjectViewMode
}

// IsTaskView returns true if in task view mode
func (s *UIState) IsTaskView() bool {
	return s.CurrentViewMode == TaskViewMode
}

// ActivateSearch enters search input mode
func (s *UIState) ActivateSearch() {
	s.SearchMode = true
	s.SearchInput = s.SearchQuery // Start with current query
}

// CancelSearch exits search input mode without applying changes
func (s *UIState) CancelSearch() {
	s.SearchMode = false
	s.SearchInput = ""
}

// CommitSearch applies the search input and exits search mode
func (s *UIState) CommitSearch() string {
	query := s.SearchInput
	s.SearchMode = false
	s.SearchInput = ""
	return query
}

// SetSearchQuery updates the active search query and state
func (s *UIState) SetSearchQuery(query string) {
	s.SearchQuery = query
	s.SearchActive = (query != "")
}

// ClearSearch clears the active search
func (s *UIState) ClearSearch() {
	s.SearchQuery = ""
	s.SearchActive = false
	s.TaskMatchingIndices = nil
	s.TaskTotalMatches = 0
}

// UpdateSearchMatches updates the computed search match state
func (s *UIState) UpdateSearchMatches(indices []int, total int) {
	s.TaskMatchingIndices = indices
	s.TaskTotalMatches = total
}

// =============================================================================
// COMPUTED DATA METHODS
// =============================================================================
// These methods compute derived data from UIState.
// Previously these lived in MainModel, but they logically belong here since
// they operate on UIState data.

// GetTaskSearchState returns the current search state with computed match position
// Takes selectedIndex as parameter since that's needed for computing current match
func (s *UIState) GetTaskSearchState(selectedIndex int) (active bool, query string, indices []int, currentMatch int, totalMatches int) {
	active = s.SearchActive
	query = s.SearchQuery
	indices = s.TaskMatchingIndices
	totalMatches = s.TaskTotalMatches
	currentMatch = -1

	// Calculate current match position if search is active
	if active && totalMatches > 0 {
		// Find position of selectedIndex in matching indices
		for i, matchIdx := range indices {
			if matchIdx == selectedIndex {
				currentMatch = i
				break
			}
		}
	}

	return
}

// GetSelectedTaskIndex returns the currently selected task index
func (s *UIState) GetSelectedTaskIndex() int {
	return s.SelectedTaskIndex
}

// GetActiveViewName returns a human-readable name of the currently active view
func (s *UIState) GetActiveViewName() string {
	switch s.ActivePanel {
	case LeftPanel:
		if s.IsProjectView() {
			return "Project List"
		}
		return "Task List"
	case RightPanel:
		if s.IsProjectView() {
			return "Project Details"
		}
		return "Task Details"
	default:
		return "Unknown"
	}
}
