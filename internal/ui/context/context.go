package context

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yousfisaad/lazyarchon/v2/internal/shared/config"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/interfaces"
)

// TaskState represents the state of a background task
type TaskState = int

const (
	TaskStart TaskState = iota
	TaskFinished
	TaskError
)

// ViewMode represents the current application view mode
type ViewMode int

const (
	TaskViewMode    ViewMode = iota // Viewing/managing tasks (default)
	ProjectViewMode                 // Selecting projects
)

// Canonical task status values as stored by the plugin layer; the UI filters
// and color-codes against these strings directly.
const (
	statusTodo   = "todo"
	statusDoing  = "doing"
	statusReview = "review"
	statusDone   = "done"
)

// Task represents a background operation with progress tracking
type Task struct {
	ID           string
	StartText    string
	FinishedText string
	State        TaskState
	Error        error
	StartTime    time.Time
	FinishedTime *time.Time
}

// ProgramContext holds shared application state that components need to access.
// This is the BUSINESS STATE CONTAINER - holds domain data and persistent user preferences.
//
// IMPORTANT: See docs/architecture/state-separation.md for guidelines on what belongs here.
//
// ProgramContext contains:
// 1. Environment & Configuration (runtime environment, user config)
// 2. Interface Dependencies (service interfaces for dependency injection)
// 3. Core Application Data (Tasks, Projects - SOURCE OF TRUTH)
// 4. System State (affects multiple components: Loading, Error, Connected)
// 5. User Preferences (persistent settings: SortMode, StatusFilters)
//
// ProgramContext does NOT contain:
// - Component instances (those live in MainModel)
// - Transient UI state (selectedIndex, activeView, searchMode - those live in UIState)
// - Computed/derived state (calculate on-demand instead)
// - Modal-specific state (those live in modal components)
// - Temporary feedback messages (those live in MainModel)
//
// For UI presentation state, see UIState which holds view mode, active panel, search state, etc.
type ProgramContext struct {
	// =============================================================================
	// 1. ENVIRONMENT & CONFIGURATION
	// =============================================================================
	// Runtime environment info that multiple components need to reference

	// Screen dimensions - kept for reference, components manage their own via WindowSizeMsg
	ScreenHeight int
	ScreenWidth  int

	// Configuration and user information
	Config   *config.Config // Application configuration
	User     string         // Current user
	RepoPath string         // Repository path
	Version  string         // Application version

	// =============================================================================
	// 2. INTERFACE DEPENDENCIES (Clean Architecture / Dependency Injection)
	// =============================================================================
	// Service interfaces that components use to perform operations

	TaskClient           interfaces.TaskClient           // API client for task backend
	ConfigProvider       interfaces.ConfigProvider       // Configuration access
	StyleContextProvider interfaces.StyleContextProvider // Styling and theme access
	Logger               interfaces.Logger               // Logging service

	// =============================================================================
	// 3. CORE APPLICATION DATA (Source of Truth)
	// =============================================================================
	// The actual business data that the application displays and manipulates.
	// Components should ALWAYS reference this data, never duplicate it.

	Tasks             []interfaces.Task    // All tasks from backend server (SOURCE OF TRUTH)
	Projects          []interfaces.Project // All projects from backend server (SOURCE OF TRUTH)
	SelectedProjectID *string              // Currently selected project (nil = "All Tasks", UUID = specific project)

	// =============================================================================
	// 4. GLOBAL UI STATE
	// =============================================================================
	// State that affects multiple components and represents overall application state
	// NOTE: UI presentation details (spinner animations, frame indices, etc.) are
	// component-local concerns and live in the components themselves (e.g., StatusBar)

	Connected      bool   // Connection status to Archon server (affects entire UI)
	Loading        bool   // Whether the application is loading data (affects entire UI)
	LoadingMessage string // Context-specific loading message (e.g., "Loading tasks...")
	Error          string // Current error message (displayed globally)
	LastRetryError string // Last error for retry functionality

	// =============================================================================
	// 5. USER PREFERENCES (Persistent Settings)
	// =============================================================================
	// Settings that represent user preferences and should persist across the session.
	// These are GLOBAL settings that affect how data is displayed everywhere.

	SortMode             int             // Current task sorting mode (STATUS+PRIORITY, PRIORITY, TIME, ALPHABETICAL, PLAN)
	StatusFilters        map[string]bool // Status visibility filters (todo, doing, review, done)
	StatusFilterActive   bool            // Whether custom status filtering is active (computed from StatusFilters)
	FeatureFilters       map[string]bool // Feature visibility filters (which features to show)
	FeatureFilterActive  bool            // Whether custom feature filtering is active (computed from FeatureFilters)
	WorktreeFilters      map[string]bool // Worktree visibility filters (which worktrees' tasks to show)
	WorktreeFilterActive bool            // Whether custom worktree filtering is active (computed from WorktreeFilters)
	SearchHistory        []string        // Recent search queries for history navigation (persistent across searches)
	ShowCompletedTasks   bool            // User preference for showing completed tasks (persistent setting)
	ShowArchivedTasks    bool            // User preference for showing archived tasks (hidden by default)
	ReadyOnly            bool            // User preference for showing only tasks whose blockers are all done

	// =============================================================================
	// 6. BACKGROUND TASK MANAGEMENT
	// =============================================================================
	// System for tracking long-running background operations

	StartTask       func(task Task) tea.Cmd // Function to start a background task
	BackgroundTasks []Task                  // Active background tasks
}

// NewProgramContext creates a new program context with default values
func NewProgramContext(cfg *config.Config, taskClient interfaces.TaskClient, configProvider interfaces.ConfigProvider, styleContextProvider interfaces.StyleContextProvider, logger interfaces.Logger) *ProgramContext {
	return &ProgramContext{
		// Initialize with default screen size - will be updated on first WindowSizeMsg
		ScreenHeight: 24,
		ScreenWidth:  80,

		// Dependencies
		Config:               cfg,
		TaskClient:           taskClient,
		ConfigProvider:       configProvider,
		StyleContextProvider: styleContextProvider,
		Logger:               logger,

		// Initialize collections
		Tasks:           make([]interfaces.Task, 0),
		Projects:        make([]interfaces.Project, 0),
		BackgroundTasks: make([]Task, 0),

		// Initialize user preferences
		SearchHistory: make([]string, 0),
		StatusFilters: map[string]bool{
			statusTodo:   true,
			statusDoing:  true,
			statusReview: true,
			statusDone:   true,
		},
		FeatureFilters:  nil, // nil = no feature filtering active (show all). Empty map = show nothing.
		WorktreeFilters: nil, // nil = no worktree filtering active (show all). Empty map = stamped tasks only.
	}
}

// UpdateScreenDimensions updates screen dimensions for reference
// Components now manage their own dimensions through WindowSizeMsg
func (ctx *ProgramContext) UpdateScreenDimensions(width, height int) {
	ctx.ScreenWidth = width
	ctx.ScreenHeight = height
	// That's it! Components handle their own dimensions now
}

// NOTE: GetContentHeight, GetLeftPanelWidth, and GetRightPanelWidth methods removed
// Components now manage their own dimensions through WindowSizeMsg

// SetTasks updates the tasks data in the context
func (ctx *ProgramContext) SetTasks(tasks []interfaces.Task) {
	ctx.Tasks = tasks
}

// SetProjects updates the projects data in the context
func (ctx *ProgramContext) SetProjects(projects []interfaces.Project) {
	ctx.Projects = projects
}

// SetSelectedProject updates the currently selected project
func (ctx *ProgramContext) SetSelectedProject(projectID *string) {
	ctx.SelectedProjectID = projectID
}

// SetConnected updates the connection status
func (ctx *ProgramContext) SetConnected(connected bool) {
	ctx.Connected = connected
}

// GetCurrentProjectName returns the name of the currently selected project
func (ctx *ProgramContext) GetCurrentProjectName() string {
	if ctx.SelectedProjectID == nil {
		return "All Tasks"
	}

	// Find project by ID
	for _, project := range ctx.Projects {
		if project.ID == *ctx.SelectedProjectID {
			return project.Title
		}
	}

	return "Unknown Project"
}

// Task Dependency Methods
//
// Blocked/ready is computed from the loaded tasks, never stored — a stored
// flag would go stale the moment another session (TUI or MCP) finishes a
// blocker in the shared SQLite database.

// UnfinishedBlockers returns the subset of blockedIDs whose tasks are not
// done. Unknown IDs count as unfinished (conservative, matching the local
// backend where archived-but-undone blockers still block).
func (ctx *ProgramContext) UnfinishedBlockers(blockedIDs []string) []string {
	statuses := make(map[string]string, len(ctx.Tasks))
	for _, task := range ctx.Tasks {
		statuses[task.ID] = task.Status
	}

	unfinished := make([]string, 0, len(blockedIDs))
	for _, id := range blockedIDs {
		if statuses[id] != interfaces.StatusDone {
			unfinished = append(unfinished, id)
		}
	}

	return unfinished
}

// TaskTitleByID resolves a task title from the loaded tasks, falling back
// to the ID itself when the task is not currently loaded.
func (ctx *ProgramContext) TaskTitleByID(taskID string) string {
	for _, task := range ctx.Tasks {
		if task.ID == taskID {
			return task.Title
		}
	}

	return taskID
}

// Search History Management Methods
// Note: Active search state (searchMode, searchInput, searchQuery) lives in UIState
// as transient UI state. Only persistent search history belongs in ProgramContext.

// AddToSearchHistory adds a query to search history if it's new
// This is a user preference that persists across searches
func (ctx *ProgramContext) AddToSearchHistory(query string) {
	if query != "" && (len(ctx.SearchHistory) == 0 || ctx.SearchHistory[len(ctx.SearchHistory)-1] != query) {
		ctx.SearchHistory = append(ctx.SearchHistory, query)
		// Keep history to reasonable size
		if len(ctx.SearchHistory) > 10 {
			ctx.SearchHistory = ctx.SearchHistory[1:]
		}
	}
}

// Status Filtering Methods

// SetStatusFilter updates visibility for a specific status
func (ctx *ProgramContext) SetStatusFilter(status string, visible bool) {
	if ctx.StatusFilters == nil {
		ctx.StatusFilters = make(map[string]bool)
	}
	ctx.StatusFilters[status] = visible
	ctx.updateFilterActiveState()
}

// ToggleStatusFilter toggles visibility for a specific status
func (ctx *ProgramContext) ToggleStatusFilter(status string) {
	if ctx.StatusFilters == nil {
		ctx.StatusFilters = make(map[string]bool)
	}
	ctx.StatusFilters[status] = !ctx.StatusFilters[status]
	ctx.updateFilterActiveState()
}

// IsStatusVisible checks if a status should be visible
func (ctx *ProgramContext) IsStatusVisible(status string) bool {
	if ctx.StatusFilters == nil {
		return true
	}
	visible, exists := ctx.StatusFilters[status]
	return !exists || visible // Default to visible if not specified
}

// ResetStatusFilters resets all status filters to visible
func (ctx *ProgramContext) ResetStatusFilters() {
	ctx.StatusFilters = map[string]bool{
		statusTodo:   true,
		statusDoing:  true,
		statusReview: true,
		statusDone:   true,
	}
	ctx.StatusFilterActive = false
}

// updateFilterActiveState determines if any custom filtering is active
func (ctx *ProgramContext) updateFilterActiveState() {
	// Check if any status is filtered out
	ctx.StatusFilterActive = false
	for _, visible := range ctx.StatusFilters {
		if !visible {
			ctx.StatusFilterActive = true
			break
		}
	}
}

// Feature Filtering Methods

// SetFeatureFilter updates visibility for a specific feature
func (ctx *ProgramContext) SetFeatureFilter(feature string, visible bool) {
	if ctx.FeatureFilters == nil {
		ctx.FeatureFilters = make(map[string]bool)
	}
	ctx.FeatureFilters[feature] = visible
	ctx.updateFeatureFilterActiveState()
}

// ToggleFeatureFilter toggles visibility for a specific feature
func (ctx *ProgramContext) ToggleFeatureFilter(feature string) {
	if ctx.FeatureFilters == nil {
		ctx.FeatureFilters = make(map[string]bool)
	}
	ctx.FeatureFilters[feature] = !ctx.FeatureFilters[feature]
	ctx.updateFeatureFilterActiveState()
}

// IsFeatureVisible checks if a feature should be visible
func (ctx *ProgramContext) IsFeatureVisible(feature string) bool {
	if ctx.FeatureFilters == nil {
		return true
	}
	// Non-nil map: only explicitly enabled features are visible
	// (an empty map means nothing is selected, so nothing is visible)
	return ctx.FeatureFilters[feature]
}

// ResetFeatureFilters resets all feature filters (shows all features)
func (ctx *ProgramContext) ResetFeatureFilters() {
	ctx.FeatureFilters = nil // nil = no filtering, show all
	ctx.FeatureFilterActive = false
}

// updateFeatureFilterActiveState determines if any custom feature filtering is active
func (ctx *ProgramContext) updateFeatureFilterActiveState() {
	// Feature filtering is active whenever the filter map is non-nil;
	// an empty non-nil map means "nothing selected" and is still an active filter
	ctx.FeatureFilterActive = ctx.FeatureFilters != nil
}

// Worktree Filtering Methods

// SetWorktreeFilter updates visibility for a specific worktree
func (ctx *ProgramContext) SetWorktreeFilter(worktree string, visible bool) {
	if ctx.WorktreeFilters == nil {
		ctx.WorktreeFilters = make(map[string]bool)
	}
	ctx.WorktreeFilters[worktree] = visible
	ctx.updateWorktreeFilterActiveState()
}

// ToggleWorktreeFilter toggles visibility for a specific worktree
func (ctx *ProgramContext) ToggleWorktreeFilter(worktree string) {
	if ctx.WorktreeFilters == nil {
		ctx.WorktreeFilters = make(map[string]bool)
	}
	ctx.WorktreeFilters[worktree] = !ctx.WorktreeFilters[worktree]
	ctx.updateWorktreeFilterActiveState()
}

// IsWorktreeVisible checks if a worktree's tasks should be visible
func (ctx *ProgramContext) IsWorktreeVisible(worktree string) bool {
	if ctx.WorktreeFilters == nil {
		return true
	}
	// Non-nil map: only explicitly enabled worktrees are visible.
	// Unstamped tasks are handled separately (they always pass).
	return ctx.WorktreeFilters[worktree]
}

// ResetWorktreeFilters resets all worktree filters (shows all)
func (ctx *ProgramContext) ResetWorktreeFilters() {
	ctx.WorktreeFilters = nil // nil = no filtering, show all
	ctx.WorktreeFilterActive = false
}

// updateWorktreeFilterActiveState determines if any custom worktree filtering is active
func (ctx *ProgramContext) updateWorktreeFilterActiveState() {
	// Worktree filtering is active whenever the filter map is non-nil;
	// an empty non-nil map hides every stamped task (unstamped always pass)
	ctx.WorktreeFilterActive = ctx.WorktreeFilters != nil
}

// UI State Management Methods

// SetLoading updates the loading state and message
func (ctx *ProgramContext) SetLoading(loading bool, message string) {
	ctx.Loading = loading
	ctx.LoadingMessage = message
}

// SetError updates the current error message
func (ctx *ProgramContext) SetError(err string) {
	ctx.Error = err
}

// ClearError clears the current error message
func (ctx *ProgramContext) ClearError() {
	ctx.Error = ""
}

// SetLastRetryError updates the last retry error for tracking
func (ctx *ProgramContext) SetLastRetryError(err string) {
	ctx.LastRetryError = err
}

// Sorting Management Methods

// SetSortMode updates the current sorting mode
func (ctx *ProgramContext) SetSortMode(mode int) {
	ctx.SortMode = mode
}

// GetSortMode returns the current sorting mode
func (ctx *ProgramContext) GetSortMode() int {
	return ctx.SortMode
}

// SetShowCompletedTasks updates the show completed tasks preference
func (ctx *ProgramContext) SetShowCompletedTasks(show bool) {
	ctx.ShowCompletedTasks = show
}

// ToggleShowCompletedTasks toggles the show completed tasks preference
func (ctx *ProgramContext) ToggleShowCompletedTasks() {
	ctx.ShowCompletedTasks = !ctx.ShowCompletedTasks
}

// SetShowArchivedTasks updates the show archived tasks preference
func (ctx *ProgramContext) SetShowArchivedTasks(show bool) {
	ctx.ShowArchivedTasks = show
}

// ToggleShowArchivedTasks toggles the show archived tasks preference
func (ctx *ProgramContext) ToggleShowArchivedTasks() {
	ctx.ShowArchivedTasks = !ctx.ShowArchivedTasks
}

// =============================================================================
// COMPUTED DATA METHODS
// =============================================================================
// These methods compute derived data from the core application state.
// Previously these lived in MainModel, but they logically belong here since
// they operate on ProgramContext data.

// GetTaskStatusCounts returns counts of tasks by status. Archived tasks are
// excluded: the totals summarize active work, independent of the visibility
// toggle.
func (ctx *ProgramContext) GetTaskStatusCounts() (todo, doing, review, done int) {
	for _, task := range ctx.Tasks {
		if task.Archived {
			continue
		}
		switch task.Status {
		case interfaces.StatusTodo:
			todo++
		case interfaces.StatusDoing:
			doing++
		case interfaces.StatusReview:
			review++
		case interfaces.StatusDone:
			done++
		}
	}
	return
}

// GetCurrentSortModeName returns the human-readable name of the current sort mode
func (ctx *ProgramContext) GetCurrentSortModeName() string {
	// Import cycle prevention: We can't import sorting package here.
	// Instead, use magic numbers that match sorting.Sort* constants.
	// This is acceptable since sort modes are stable and well-defined.
	switch ctx.SortMode {
	case 0: // sorting.SortStatusPriority
		return "Status"
	case 1: // sorting.SortPriorityOnly
		return "Priority"
	case 2: // sorting.SortTimeCreated
		return "Created"
	case 3: // sorting.SortAlphabetical
		return "Alpha"
	case 4: // sorting.SortPlan
		return "Plan"
	default:
		return "Unknown"
	}
}

// GetUniqueFeatures returns a sorted list of unique tags from current tasks
// (Previously called "features" in Archon, now "tags" in generic interface)
func (ctx *ProgramContext) GetUniqueFeatures() []string {
	tagSet := make(map[string]bool)
	for _, task := range ctx.Tasks {
		for _, tag := range task.Tags {
			if tag != "" {
				tagSet[tag] = true
			}
		}
	}

	// Convert map to sorted slice
	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	// Sort alphabetically (simple string sort)
	// Note: We can't import helpers package here due to import cycles
	// Use a simple bubble sort for small lists
	for i := 0; i < len(tags); i++ {
		for j := i + 1; j < len(tags); j++ {
			if tags[i] > tags[j] {
				tags[i], tags[j] = tags[j], tags[i]
			}
		}
	}

	return tags
}

// GetFeatureFilterSummary returns a summary of active feature filters
// Reads from ctx.FeatureFilters (user preference)
func (ctx *ProgramContext) GetFeatureFilterSummary() string {
	allFeatures := ctx.GetUniqueFeatures()

	if len(allFeatures) == 0 {
		return "No features"
	}

	if ctx.FeatureFilters == nil {
		return "All features"
	}

	// Count active filters
	activeCount := 0
	for _, feature := range allFeatures {
		if ctx.FeatureFilters[feature] {
			activeCount++
		}
	}

	if activeCount == 0 {
		return "No features"
	}

	if activeCount == len(allFeatures) {
		return "All features"
	}

	// Show count of active filters
	if activeCount == 1 {
		// Find and show the single active feature
		for _, feature := range allFeatures {
			if ctx.FeatureFilters[feature] {
				return feature
			}
		}
	}

	return fmt.Sprintf("%d features", activeCount)
}

// GetUniqueWorktrees returns a sorted list of unique worktrees from current tasks
func (ctx *ProgramContext) GetUniqueWorktrees() []string {
	worktreeSet := make(map[string]bool)
	for _, task := range ctx.Tasks {
		if task.Worktree != "" {
			worktreeSet[task.Worktree] = true
		}
	}

	// Convert map to sorted slice
	worktrees := make([]string, 0, len(worktreeSet))
	for worktree := range worktreeSet {
		worktrees = append(worktrees, worktree)
	}

	// Sort alphabetically (simple string sort)
	// Note: We can't import helpers package here due to import cycles
	// Use a simple bubble sort for small lists
	for i := 0; i < len(worktrees); i++ {
		for j := i + 1; j < len(worktrees); j++ {
			if worktrees[i] > worktrees[j] {
				worktrees[i], worktrees[j] = worktrees[j], worktrees[i]
			}
		}
	}

	return worktrees
}

// GetWorktreeFilterSummary returns a summary of active worktree filters
// Reads from ctx.WorktreeFilters (user preference)
func (ctx *ProgramContext) GetWorktreeFilterSummary() string {
	allWorktrees := ctx.GetUniqueWorktrees()

	if len(allWorktrees) == 0 {
		return "No worktrees"
	}

	if ctx.WorktreeFilters == nil {
		return "All worktrees"
	}

	// Count active filters
	activeCount := 0
	for _, worktree := range allWorktrees {
		if ctx.WorktreeFilters[worktree] {
			activeCount++
		}
	}

	if activeCount == 0 {
		return "No worktrees"
	}

	if activeCount == len(allWorktrees) {
		return "All worktrees"
	}

	// Show count of active filters
	if activeCount == 1 {
		// Find and show the single active worktree
		for _, worktree := range allWorktrees {
			if ctx.WorktreeFilters[worktree] {
				return "@" + worktree
			}
		}
	}

	return fmt.Sprintf("%d worktrees", activeCount)
}

// GetTaskCountForProject returns the number of non-archived tasks for a
// specific project
func (ctx *ProgramContext) GetTaskCountForProject(projectID string) int {
	count := 0
	for _, task := range ctx.Tasks {
		if task.Archived {
			continue
		}
		if task.ProjectID == projectID {
			count++
		}
	}
	return count
}

// GetTotalTaskCount returns the total number of non-archived tasks
func (ctx *ProgramContext) GetTotalTaskCount() int {
	count := 0
	for _, task := range ctx.Tasks {
		if !task.Archived {
			count++
		}
	}
	return count
}
