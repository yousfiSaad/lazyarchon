package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yousfisaad/lazyarchon/v2/internal/archon"
	"github.com/yousfisaad/lazyarchon/v2/internal/domain/projectmode"
	"github.com/yousfisaad/lazyarchon/v2/internal/domain/projects"
	"github.com/yousfisaad/lazyarchon/v2/internal/domain/tasks"
	"github.com/yousfisaad/lazyarchon/v2/internal/logging"
	configpkg "github.com/yousfisaad/lazyarchon/v2/internal/shared/config"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/interfaces"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/styling"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/utils"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/utils/keys"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/base"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/layout/header"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/layout/maincontent"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/layout/statusbar"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/modals/confirmation"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/modals/feature"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/modals/help"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/modals/status"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/modals/statusfilter"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/modals/taskedit"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/projectlist"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/taskdetails"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/tasklist"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/context"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/factories"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/helpers"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/messages"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/sorting"
	stylingprovider "github.com/yousfisaad/lazyarchon/v2/internal/ui/styling"
)

// =============================================================================
// TYPE DEFINITIONS
// =============================================================================

// ActiveView represents which panel is currently active for user input
type ActiveView int

const (
	LeftPanel  ActiveView = 0 // Task list panel
	RightPanel ActiveView = 1 // Task details panel
)

// MainModel coordinates the UI components and manages orchestration.
// This is the ORCHESTRATION LAYER - manages components and routes messages.
//
// IMPORTANT: See docs/architecture/state-separation.md for guidelines on what belongs here.
//
// MainModel contains:
// 1. Component References (UI structure: components, managers)
// 2. State References (ProgramContext, UIState - NOT owned, just referenced)
// 3. Temporary modal state (being migrated out)
//
// MainModel does NOT contain:
// - Business data (Tasks, Projects - those live in ProgramContext)
// - UI state (activeView, searchMode - those live in UIState)
// - User preferences (SortMode, StatusFilters - those live in ProgramContext)
//
// State Location Decision:
// - Business data / persistent preferences → ProgramContext
// - UI presentation state → UIState
// - Component instances / orchestration → MainModel
type MainModel struct {
	// =============================================================================
	// 1. COMPONENT REFERENCES (UI Structure)
	// =============================================================================
	// The component tree that makes up the application

	programContext *context.ProgramContext // Reference to business state (NOT owned by Model)
	uiState        *context.UIState        // Reference to UI state (NOT owned by Model)

	components factories.UIComponentSet // All UI components (layout, modals, panels)

	// =============================================================================
	// 2. TEMPORARY STATE (Modal/Dialog State)
	// =============================================================================
	// TODO: Consider moving to UIState as well

	// NOTE: featureFilters moved to ProgramContext.FeatureFilters (user preference)

	// Feature search (for feature selection modal only)
	//nolint:unused // Reserved for future feature modal search functionality
	featureSearchActive bool // Whether feature modal search is active
	//nolint:unused // Reserved for future feature modal search functionality
	featureSearchQuery string // Feature modal search query
	//nolint:unused // Reserved for future feature modal search functionality
	featureSelectedIndex int // Selected index in feature modal

	// Confirmation dialogs
	pendingDeleteTaskID string // Task ID awaiting deletion confirmation

}

// =============================================================================
// MODEL INITIALIZATION
// =============================================================================

// createServices creates the service implementations using the extracted services
//
//nolint:ireturn // Returns interfaces by design - dependency injection pattern
func createServices(cfg *configpkg.Config) (interfaces.StyleContextProvider, interfaces.Logger) {
	// Create service instances using the extracted service packages
	styleContextProvider := stylingprovider.NewProvider(cfg)
	logger := logging.NewSlogLogger(cfg.IsDebugEnabled())

	return styleContextProvider, logger
}

// NewModel creates a new application model with interface dependencies
func NewModel(cfg *configpkg.Config) MainModel {
	// Initialize theme from configuration
	styling.InitializeTheme(cfg)

	// Create service implementations using extracted services
	styleContextProvider, logger := createServices(cfg)

	// Create concrete implementations for interface dependencies
	client := archon.NewClient(cfg.GetServerURL(), cfg.GetAPIKey())
	client.SetLogger(logger) // Inject logger for HTTP request/response logging

	// Delegate to shared model creation logic
	return createModelWithDependencies(client, cfg, styleContextProvider, logger)
}

// createModelWithDependencies contains the shared model creation logic
// This eliminates duplication between NewModel and NewModelWithDependencies
func createModelWithDependencies(
	client interfaces.ArchonClient,
	config interfaces.ConfigProvider,
	styleContextProvider interfaces.StyleContextProvider,
	logger interfaces.Logger,
) MainModel {
	logger.Debug("Creating UI model with injected dependencies")

	programContext, uiState, componentContext := createContexts(client, config, styleContextProvider, logger)
	initializeContextState(programContext, config)
	applyDefaultProjectID(programContext, config)
	components := createComponents(componentContext)
	model := buildModel(programContext, uiState, components, config)

	// Wire up remaining parent-provided state accessors after model exists
	// GetSortedTasks: Still a callback since it involves complex filtering in MainModel
	componentContext.GetSortedTasks = func() []interface{} {
		tasks := model.GetSortedTasks()
		// Convert []archon.Task to []interface{} to avoid import cycle
		result := make([]interface{}, len(tasks))
		for i := range tasks {
			result[i] = tasks[i]
		}
		return result
	}

	// NOTE: Computed data accessors removed - components now call ProgramContext/UIState directly
	// This eliminates verbose lambda wiring and puts methods where the data lives

	initializeLayoutComponents(&model, componentContext)

	return model
}

// createContexts creates program, UI, and component contexts
func createContexts(
	client interfaces.ArchonClient,
	config interfaces.ConfigProvider,
	styleContextProvider interfaces.StyleContextProvider,
	logger interfaces.Logger,
) (*context.ProgramContext, *context.UIState, *base.ComponentContext) {
	programContext := context.NewProgramContext(
		config.(*configpkg.Config),
		client,
		config,
		styleContextProvider,
		logger,
	)

	// Create UI state for presentation concerns
	uiState := context.NewUIState()

	componentContext := &base.ComponentContext{
		ProgramContext:       programContext,
		UIState:              uiState,
		ConfigProvider:       config,
		StyleContextProvider: styleContextProvider,
		Logger:               logger,
		MessageChan:          make(chan tea.Msg, 100),
	}

	return programContext, uiState, componentContext
}

// initializeContextState sets initial loading and sort state
func initializeContextState(programContext *context.ProgramContext, config interfaces.ConfigProvider) {
	programContext.SetLoading(true, "Connecting to Archon server...")
	setSortMode(programContext, config)
}

// createComponents creates modal components
func createComponents(
	componentContext *base.ComponentContext,
) *factories.UIComponentSet {
	return factories.CreateComponents(factories.ComponentConfig{
		ComponentContext: componentContext,
	})
}

// buildModel constructs the Model struct with components
func buildModel(
	programContext *context.ProgramContext,
	uiState *context.UIState,
	components *factories.UIComponentSet,
	config interfaces.ConfigProvider,
) MainModel {
	model := MainModel{
		programContext: programContext,
		uiState:        uiState,
		components:     *components,
	}

	// Initialize ShowCompletedTasks in ProgramContext from config
	if concreteConfig, ok := config.(*configpkg.Config); ok {
		programContext.SetShowCompletedTasks(concreteConfig.ShouldShowCompletedTasks())
	}

	return model
}

// initializeLayoutComponents creates and wires layout components
func initializeLayoutComponents(
	model *MainModel,
	componentContext *base.ComponentContext,
) {
	model.components.Layout.Header = header.NewModel(componentContext)

	model.components.Layout.MainContent = maincontent.NewModel(componentContext)

	model.components.Layout.StatusBar = statusbar.NewModel(componentContext)
}

// setSortMode sets the initial sort mode from configuration
func setSortMode(programContext *context.ProgramContext, config interfaces.ConfigProvider) {
	defaultSortMode := "status+priority" // Default fallback
	if configProvider, ok := config.(*configpkg.Config); ok {
		defaultSortMode = configProvider.GetDefaultSortMode()
	}

	// Set sort mode - use the same logic as factory
	sortMode := sorting.SortStatusPriority // default
	switch defaultSortMode {
	case "status+priority":
		sortMode = sorting.SortStatusPriority
	case "priority":
		sortMode = sorting.SortPriorityOnly
	case "time":
		sortMode = sorting.SortTimeCreated
	case "alphabetical":
		sortMode = sorting.SortAlphabetical
	}
	programContext.SetSortMode(sortMode)
}

// applyDefaultProjectID applies the default project ID from configuration
func applyDefaultProjectID(programContext *context.ProgramContext, config interfaces.ConfigProvider) {
	if concreteConfig, ok := config.(*configpkg.Config); ok {
		if defaultProjectID := concreteConfig.GetDefaultProjectID(); defaultProjectID != "" {
			programContext.SetSelectedProject(&defaultProjectID)
		}
	}
}

// =============================================================================
// BUBBLE TEA INTERFACE
// =============================================================================

// Init initializes the application
func (m MainModel) Init() tea.Cmd {
	cmds := []tea.Cmd{
		tasks.LoadTasksInterface(m.programContext.ArchonClient, m.programContext.SelectedProjectID),
		projects.LoadProjectsInterface(m.programContext.ArchonClient),
		m.components.Layout.StatusBar.Init(), // Initialize StatusBar (starts spinner)
		m.startPolling(),                     // Use HTTP polling for auto-refresh
	}

	return tea.Batch(cmds...)
}

// Update handles incoming events and updates the model
// Uses pointer receiver to maintain component reference validity across updates
//
//nolint:ireturn // Required by Bubble Tea framework - must return tea.Model interface
func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowResize(msg)
	case tea.KeyMsg:
		return m.handleKeyInput(msg)
	case tasks.TasksLoadedMsg, tasks.TaskUpdateMsg, tasks.TaskDeleteMsg:
		return m.handleTaskMessages(msg)
	case projects.ProjectsLoadedMsg:
		return m.handleProjectMessages(msg)
	case messages.PollingTickMsg:
		return m.handlePollingTick()
	case help.ShowHelpModalMsg, help.HideHelpModalMsg, help.HelpModalShownMsg, help.HelpModalHiddenMsg,
		status.ShowStatusModalMsg, status.HideStatusModalMsg, status.StatusModalShownMsg, status.StatusModalHiddenMsg,
		confirmation.ShowConfirmationModalMsg, confirmation.HideConfirmationModalMsg, confirmation.ConfirmationModalShownMsg, confirmation.ConfirmationModalHiddenMsg,
		taskedit.ShowTaskEditModalMsg, taskedit.HideTaskEditModalMsg, taskedit.TaskEditModalShownMsg, taskedit.TaskEditModalHiddenMsg,
		feature.ShowFeatureModalMsg, feature.HideFeatureModalMsg, feature.FeatureModalShownMsg, feature.FeatureModalHiddenMsg:
		return m.handleModalLifecycle(msg)
	case status.StatusSelectedMsg, taskedit.TaskPropertiesUpdatedMsg, confirmation.ConfirmationSelectedMsg,
		taskedit.FeatureSelectedMsg, feature.FeatureSelectionAppliedMsg, statusfilter.StatusFilterAppliedMsg:
		return m.handleModalActions(msg)
	case projectlist.ProjectListUpdateMsg, projectlist.ProjectListSelectMsg, projectlist.ProjectListScrollMsg,
		projectlist.ProjectListSelectionChangedMsg, tasklist.TaskListSelectionChangedMsg,
		messages.YankIDMsg, messages.YankTitleMsg, messages.StatusFeedbackMsg, messages.SearchStateChangedMsg:
		return m.handleComponentMessages(msg)
	case projectmode.ProjectModeActivatedMsg, projectmode.ProjectModeDeactivatedMsg:
		return m.handleProjectModeMessages(msg)
	case base.ComponentMessage:
		// Process the payload message that was wrapped by the component
		return m.Update(msg.Payload)
	}

	// Fallback: broadcast all other messages to component tree
	return m, m.components.Update(msg)
}

// =============================================================================
// MESSAGE HANDLERS - Extracted from Update() for better organization
// =============================================================================

// handleWindowResize processes window resize events and updates layout
//
//nolint:ireturn // Required by Bubble Tea framework - must return tea.Model interface
func (m *MainModel) handleWindowResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	// Update global screen tracking (keep for reference)
	m.programContext.UpdateScreenDimensions(msg.Width, msg.Height)

	// Simple hardcoded layout calculation
	headerHeight := 1
	footerHeight := 1
	mainContentHeight := msg.Height - headerHeight - footerHeight

	// Ensure minimum height
	if mainContentHeight < 5 {
		mainContentHeight = 5
		footerHeight = max(0, msg.Height-headerHeight-mainContentHeight)
	}

	var cmds []tea.Cmd

	// Send dimensions to header component
	if m.components.Layout.Header != nil {
		cmd := m.components.Layout.Header.Update(tea.WindowSizeMsg{
			Width:  msg.Width,
			Height: headerHeight,
		})
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Send dimensions to main content component
	if m.components.Layout.MainContent != nil {
		cmd := m.components.Layout.MainContent.Update(tea.WindowSizeMsg{
			Width:  msg.Width,
			Height: mainContentHeight,
		})
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Send dimensions to status bar component
	if m.components.Layout.StatusBar != nil {
		cmd := m.components.Layout.StatusBar.Update(tea.WindowSizeMsg{
			Width:  msg.Width,
			Height: footerHeight,
		})
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// handleKeyInput processes keyboard input with modal awareness
//
//nolint:ireturn // Required by Bubble Tea framework - must return tea.Model interface
func (m *MainModel) handleKeyInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Broadcast to all components first (hierarchical pattern)
	// Active modals will handle keys, inactive ones will ignore
	componentCmd := m.components.Update(msg)

	// If a modal is active, only allow emergency keys through to HandleKeyPress
	// All other keys should be handled exclusively by the modal
	var modelCmd tea.Cmd
	if m.HasActiveModal() {
		// Only process global emergency keys when modal is active
		// This prevents navigation/task keys from leaking to underlying view
		keyStr := msg.String()
		if keyStr == keys.KeyCtrlC || keyStr == keys.KeyQuestion {
			modelCmd = m.handleKeyPress(keyStr)
		}
		// All other keys are handled only by the modal (via componentCmd)
	} else {
		// No modal active - process all keys normally
		// handleKeyPress updates model in-place with pointer receiver
		modelCmd = m.handleKeyPress(msg.String())
	}
	return m, tea.Batch(componentCmd, modelCmd)
}

// handleComponentMessages processes component-specific messages
// Note: Task, project, modal, and realtime handlers moved to separate files:
// - model_handlers_task.go: handleTaskMessages, handleProjectMessages, handleProjectModeMessages
// - model_handlers_modal.go: handleModalLifecycle, handleModalActions
// - model_handlers_realtime.go: handleRealtimeMessages, handleTickAnimation
//
//nolint:ireturn // Required by Bubble Tea framework - must return tea.Model interface
func (m *MainModel) handleComponentMessages(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case projectlist.ProjectListUpdateMsg, projectlist.ProjectListSelectMsg,
		projectlist.ProjectListScrollMsg,
		messages.YankIDMsg, messages.YankTitleMsg, messages.StatusFeedbackMsg:
		// Broadcast to components only (coordinators removed - state now in Model)
		return m, m.components.Update(msg)

	case messages.SearchStateChangedMsg:
		// Update UIState's search state from broadcast (SINGLE SOURCE OF TRUTH)
		m.uiState.SetSearchQuery(msg.Query)

		// Forward search state change to all interested components
		searchMsg := tasklist.TaskListSearchMsg{
			Query:  msg.Query,
			Active: msg.Active,
		}
		var cmds []tea.Cmd
		if cmd := m.components.Layout.MainContent.Update(searchMsg); cmd != nil {
			cmds = append(cmds, cmd)
		}
		if cmd := m.updateTaskDetailsComponent(); cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case projectlist.ProjectListSelectionChangedMsg:
		// Convert selection index to project ID and update context
		projects := m.programContext.Projects
		if msg.Index < len(projects) {
			selectedProjectID := projects[msg.Index].ID
			m.programContext.SetSelectedProject(&selectedProjectID)
		} else {
			m.programContext.SetSelectedProject(nil)
		}

		// Forward to MainContent to update right panel display
		if m.components.Layout.MainContent != nil {
			return m, m.components.Layout.MainContent.Update(msg)
		}
		return m, nil

	case tasklist.TaskListSelectionChangedMsg:
		// Sync UIState's selectedIndex with TaskList (SINGLE SOURCE OF TRUTH)
		m.uiState.SelectedTaskIndex = msg.Index

		// Forward to MainContent component which will intercept and update TaskDetails
		if m.components.Layout.MainContent != nil {
			return m, m.components.Layout.MainContent.Update(msg)
		}
		return m, nil
	}
	return m, nil
}

// View renders the complete UI using simple direct component rendering
// High complexity (19) due to comprehensive modal overlay logic for 5+ modal types
//
//nolint:gocyclo // View requires checking all modal states for proper overlay rendering
func (m MainModel) View() string {
	// Simple three-part layout: header + main + footer
	// Components manage their own dimensions from WindowSizeMsg
	var parts []string

	// Render header component
	if m.components.Layout.Header != nil {
		parts = append(parts, m.components.Layout.Header.View())
	}

	// Render main content component
	if m.components.Layout.MainContent != nil {
		parts = append(parts, m.components.Layout.MainContent.View())
	}

	// Render status bar component
	if m.components.Layout.StatusBar != nil {
		parts = append(parts, m.components.Layout.StatusBar.View())
	}

	// Combine vertically using lipgloss
	baseUI := lipgloss.JoinVertical(lipgloss.Left, parts...)

	// Get screen dimensions for modal positioning
	screenWidth := m.programContext.ScreenWidth
	screenHeight := m.programContext.ScreenHeight

	// Overlay modals on top of base UI using proper parent-child architecture
	// Each modal is centered and rendered as an overlay, not a replacement

	// Check if any modal is active and render it as an overlay
	var activeModal string

	// Help modal takes priority (usually opened with ?)
	if m.components.Modals.HelpModel.IsActive() {
		helpModalView := m.components.Modals.HelpModel.View()
		if helpModalView != "" {
			activeModal = helpModalView
		}
	}

	// Status change modal
	if activeModal == "" && m.components.Modals.StatusModel.IsActive() {
		statusModalView := m.components.Modals.StatusModel.View()
		if statusModalView != "" {
			activeModal = statusModalView
		}
	}

	// Confirmation modal
	if activeModal == "" && m.components.Modals.ConfirmationModel.IsActive() {
		confirmationModalView := m.components.Modals.ConfirmationModel.View()
		if confirmationModalView != "" {
			activeModal = confirmationModalView
		}
	}

	// Task edit modal
	if activeModal == "" && m.components.Modals.TaskEditModel.IsActive() {
		taskEditModalView := m.components.Modals.TaskEditModel.View()
		if taskEditModalView != "" {
			activeModal = taskEditModalView
		}
	}

	// Feature modal
	if activeModal == "" && m.components.Modals.FeatureModel.IsActive() {
		featureModalView := m.components.Modals.FeatureModel.View()
		if featureModalView != "" {
			activeModal = featureModalView
		}
	}

	// If a modal is active, overlay it on top of baseUI
	if activeModal != "" {
		// Place the modal centered over the base UI
		// This properly overlays the modal while keeping base content visible
		return lipgloss.Place(
			screenWidth, screenHeight,
			lipgloss.Center, lipgloss.Center,
			activeModal,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
		)
	}

	// Return base UI without background - let terminal handle backgrounds naturally
	return baseUI
}

// =============================================================================
// CORE DATA STATE MANAGEMENT
// =============================================================================

// setError sets the error state and clears loading
//
//nolint:unparam // return value intentionally unused in some call sites - kept for consistency
func (m *MainModel) setError(err string) tea.Cmd {
	m.programContext.SetLoading(false, "")
	m.programContext.SetError(err)
	m.programContext.SetLastRetryError(err)
	return m.broadcastStatusBarState()
}

// clearError clears the error state
func (m *MainModel) clearError() tea.Cmd {
	m.programContext.ClearError()
	m.programContext.SetLastRetryError("")
	return m.broadcastStatusBarState()
}

// setLoading sets the loading state with optional context message
//
//nolint:unparam // return value intentionally unused in some call sites - kept for consistency
func (m *MainModel) setLoading(loading bool) tea.Cmd {
	if loading {
		m.programContext.ClearError()
		m.programContext.SetLoading(true, "")
	} else {
		m.programContext.SetLoading(false, "")
	}
	return m.broadcastStatusBarState()
}

// setLoadingWithMessage sets loading state with specific context message
//
//nolint:unparam // loading parameter kept for API consistency - may be used with false in future
func (m *MainModel) setLoadingWithMessage(loading bool, message string) tea.Cmd {
	if loading {
		m.programContext.ClearError()
	}
	m.programContext.SetLoading(loading, message)
	return m.broadcastStatusBarState()
}

// formatUserFriendlyError converts technical errors to user-friendly messages
//
//nolint:unused // Reserved for future user-friendly error formatting
func (m *MainModel) formatUserFriendlyError(err string) string {
	return utils.FormatUserFriendlyError(err)
}

// setConnectionStatus sets the connection status
//
//nolint:unused // Reserved for future connection status management
func (m *MainModel) setConnectionStatus(connected bool) tea.Cmd {
	m.programContext.Connected = connected
	return m.broadcastStatusBarState()
}

// getConnectionStatusText returns a text indicator for connection status
//
//nolint:unused // Reserved for future connection status display
func (m *MainModel) getConnectionStatusText() string {
	if m.programContext.Connected {
		return "●" // Connected
	}
	return "○" // Disconnected
}

// =============================================================================
// TASK AND PROJECT DATA MANAGEMENT
// =============================================================================

// updateTasks updates the task list and adjusts selection bounds
func (m *MainModel) updateTasks(tasks []archon.Task) {
	startTime := time.Now()
	oldTaskCount := len(m.programContext.Tasks)

	m.programContext.SetLoading(false, "")

	// Preserve selected task ID from sorted list before updating
	var selectedTaskID string
	if selectedTask := m.GetSelectedTask(); selectedTask != nil {
		selectedTaskID = selectedTask.ID
	}

	m.programContext.SetTasks(tasks)
	m.programContext.SetConnected(true)
	m.clearError()

	// Log state change
	m.programContext.Logger.LogStateChange("Model", "Tasks", oldTaskCount, len(tasks),
		"selected_task_id", selectedTaskID)

	// TaskManager is now stateless - operates on tasks directly from ProgramContext
	sortedTasks := m.GetSortedTasks()

	// Restore selection by finding task ID in new sorted list
	if selectedTaskID != "" {
		m.findAndSelectTask(selectedTaskID)
	}

	m.adjustTaskSelection(sortedTasks)
	_ = m.updateTaskListComponents(sortedTasks)
	_ = m.updateTaskDetailsComponent()
	m.updateSearchMatches()

	// Broadcast all state to StatusBar
	_ = m.broadcastStatusBarState()

	// Log performance
	m.programContext.Logger.LogPerformance("UpdateTasks", startTime, "task_count", len(tasks))
}

// broadcastStatusBarState is deprecated and no longer needed.
// StatusBar now reads all state directly from ProgramContext and UIState via ctx() helper.
// This method remains as a no-op stub to avoid breaking existing call sites during migration.
//
// Previous behavior (now obsolete):
// - Broadcast LoadingStateMsg, ErrorStateMsg, ConnectionStatusMsg → StatusBar reads ctx().Loading, ctx().Error, ctx().Connected
// - Broadcast ProjectModeMsg, ActiveViewMsg → StatusBar reads UIState.IsProjectView(), UIState.GetActiveViewName()
// - Broadcast SearchModeMsg, SearchMatchInfoMsg → StatusBar reads UIState search state
// - Broadcast TaskCountsMsg, SelectionPositionMsg → StatusBar calls ctx().GetTaskStatusCounts(), UIState.GetSelectedTaskIndex()
// - Broadcast SortModeMsg, FeatureCountMsg, ProjectCountMsg → StatusBar calls ctx().GetCurrentSortModeName(), ctx().GetUniqueFeatures(), len(ctx().Projects)
//
// Result: Eliminated ~80 lines of message broadcasting overhead. StatusBar reactively renders on any state change.
func (m *MainModel) broadcastStatusBarState() tea.Cmd {
	// No-op: StatusBar reads all state directly from context
	// BubbleTea's reactive pattern ensures StatusBar.View() is called on any model update
	return nil
}

// adjustTaskSelection validates and adjusts selectedIndex after task updates
func (m *MainModel) adjustTaskSelection(sortedTasks []archon.Task) {
	if len(sortedTasks) > 0 && m.uiState.SelectedTaskIndex >= len(sortedTasks) {
		_ = m.setSelectedTask(len(sortedTasks) - 1) // Command handled by updateTaskListComponents below
	} else if len(sortedTasks) == 0 {
		_ = m.setSelectedTask(0) // Command handled by updateTaskListComponents below
	}
}

// updateTaskListComponents sends update messages to TaskList component
func (m *MainModel) updateTaskListComponents(sortedTasks []archon.Task) tea.Cmd {
	updateMsg := tasklist.TaskListUpdateMsg{
		Tasks:   sortedTasks,
		Loading: false,
		Error:   "",
	}
	var cmds []tea.Cmd
	if cmd := m.components.Layout.MainContent.Update(updateMsg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	selectMsg := tasklist.TaskListSelectMsg{Index: m.uiState.SelectedTaskIndex}
	if cmd := m.components.Layout.MainContent.Update(selectMsg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

// updateTaskDetailsComponent sends update message to TaskDetails component
func (m *MainModel) updateTaskDetailsComponent() tea.Cmd {
	selectedTask, _ := m.validateTaskSelection()
	updateMsg := taskdetails.TaskDetailsUpdateMsg{
		SelectedTask: selectedTask,
		SearchQuery:  m.uiState.SearchQuery,
		SearchActive: m.uiState.SearchActive,
	}
	return m.components.Layout.MainContent.Update(updateMsg)
}

// updateProjects updates the project list and validates current selection
func (m *MainModel) updateProjects(projects []archon.Project) {
	m.programContext.SetProjects(projects)
	m.programContext.SetConnected(true)

	// Validate project selection inline
	selectedProjectID := m.programContext.SelectedProjectID
	if selectedProjectID != nil {
		// Check if selected project still exists
		projectExists := false
		for _, project := range projects {
			if project.ID == *selectedProjectID {
				projectExists = true
				break
			}
		}
		if !projectExists {
			// Selected project no longer exists, clear selection
			m.programContext.SetSelectedProject(nil)
		}
	}

	_ = m.updateProjectListComponent(projects)
	_ = m.broadcastStatusBarState()
}

// updateProjectListComponent sends update message to ProjectList component
func (m *MainModel) updateProjectListComponent(projects []archon.Project) tea.Cmd {
	updateMsg := projectlist.ProjectListUpdateMsg{
		Projects: projects,
		Loading:  false,
		Error:    "",
	}
	var cmds []tea.Cmd
	if cmd := m.components.Layout.MainContent.Update(updateMsg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	// If there's a default/selected project, update ProjectList selection to match
	if m.programContext.SelectedProjectID != nil {
		for i, project := range projects {
			if project.ID == *m.programContext.SelectedProjectID {
				selectMsg := projectlist.ProjectListSelectMsg{Index: i}
				if cmd := m.components.Layout.MainContent.Update(selectMsg); cmd != nil {
					cmds = append(cmds, cmd)
				}
				break
			}
		}
	}

	return tea.Batch(cmds...)
}

// =============================================================================
// NAVIGATION STATE MANAGEMENT
// =============================================================================

// setSelectedProject sets the currently selected project
func (m *MainModel) setSelectedProject(projectID *string) {
	// ProjectManager is now stateless - just update ProgramContext directly
	m.programContext.SetSelectedProject(projectID)

	_ = m.setSelectedTask(0) // Reset task selection (command handled by subsequent updates)

	// Update search matches after project filter change
	m.updateSearchMatches()
}

// findAndSelectTask finds a task by ID in the current sort order and selects it
func (m *MainModel) findAndSelectTask(taskID string) {
	if taskID == "" {
		_ = m.setSelectedTask(0) // Command handled by context caller
		return
	}

	sortedTasks := m.GetSortedTasks()
	for i, task := range sortedTasks {
		if task.ID == taskID {
			_ = m.setSelectedTask(i) // Command handled by context caller
			return
		}
	}

	// Task not found, default to first task
	_ = m.setSelectedTask(0) // Command handled by context caller
}

// setSelectedTask sets the selected task index and updates viewport content
func (m *MainModel) setSelectedTask(index int) tea.Cmd {
	sortedTasks := m.GetSortedTasks()
	index = m.boundsCheckTaskIndex(index, len(sortedTasks))

	if m.uiState.SelectedTaskIndex != index {
		m.uiState.SelectedTaskIndex = index
		return m.updateTaskSelectionComponents(index)
	}
	return nil
}

// boundsCheckTaskIndex ensures task index is within valid bounds
func (m *MainModel) boundsCheckTaskIndex(index, taskCount int) int {
	if index < 0 {
		return 0
	}
	if index >= taskCount {
		if taskCount > 0 {
			return taskCount - 1
		}
		return 0
	}
	return index
}

// updateTaskSelectionComponents updates components after task selection change
func (m *MainModel) updateTaskSelectionComponents(index int) tea.Cmd {
	selectedTask, _ := m.validateTaskSelection()
	updateMsg := taskdetails.TaskDetailsUpdateMsg{
		SelectedTask: selectedTask,
		SearchQuery:  m.uiState.SearchQuery,
		SearchActive: m.uiState.SearchActive,
	}
	var cmds []tea.Cmd
	if cmd := m.components.Layout.MainContent.Update(updateMsg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	selectMsg := tasklist.TaskListSelectMsg{Index: index}
	if cmd := m.components.Layout.MainContent.Update(selectMsg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

// validateTaskSelection validates task selection and returns the selected task
//
//nolint:unparam // second return value (bool) intentionally unused in some call sites - kept for consistency
func (m *MainModel) validateTaskSelection() (*archon.Task, bool) {
	sortedTasks := m.GetSortedTasks()
	if len(sortedTasks) == 0 || m.uiState.SelectedTaskIndex >= len(sortedTasks) {
		return nil, false
	}
	return &sortedTasks[m.uiState.SelectedTaskIndex], true
}

// setNavigationSelectedIndex sets the navigation selected index
//
//nolint:unused // Reserved for future navigation index management
func (m *MainModel) setNavigationSelectedIndex(index int) {
	m.uiState.SelectedTaskIndex = index
}

// =============================================================================
// MODAL STATE MANAGEMENT
// =============================================================================

// showQuitConfirmation shows the quit confirmation modal
func (m *MainModel) showQuitConfirmation() tea.Cmd {
	return func() tea.Msg {
		return confirmation.ShowConfirmationModalMsg{
			Message:     "Are you sure you want to quit LazyArchon?",
			ConfirmText: "Quit",
			CancelText:  "Stay",
		}
	}
}

// HasActiveModal returns true if any modal overlay is currently active.
// Note: View modes (like Project Mode) are NOT modals - they are full-screen
// views that need normal key routing through HandleKeyPress().
func (m MainModel) HasActiveModal() bool {
	return m.components.Modals.HelpModel.IsActive() ||
		m.components.Modals.StatusModel.IsActive() ||
		m.components.Modals.ConfirmationModel.IsActive() ||
		m.components.Modals.FeatureModel.IsActive() ||
		m.components.Modals.TaskEditModel.IsActive()
}

// =============================================================================
// PANEL AND VIEW MANAGEMENT
// =============================================================================

// IsLeftPanelActive returns true if the left panel (task list) is currently active
func (m MainModel) IsLeftPanelActive() bool {
	return m.uiState.IsLeftPanelActive()
}

// IsRightPanelActive returns true if the right panel (task details) is currently active
func (m MainModel) IsRightPanelActive() bool {
	return m.uiState.IsRightPanelActive()
}

// setActiveView sets the currently active panel
// Components now read active state from UIState directly - no callbacks or messages needed
func (m *MainModel) setActiveView(view ActiveView) tea.Cmd {
	// Update UIState (single source of truth)
	m.uiState.SetActivePanel(context.ActivePanel(view))

	// Broadcast updated state to StatusBar (for active view indicator)
	return m.broadcastStatusBarState()
}

// GetActiveViewName returns a human-readable name of the currently active view
func (m MainModel) GetActiveViewName() string {
	return m.uiState.GetActiveViewName()
}

// =============================================================================
// SEARCH MANAGEMENT
// =============================================================================

// ActivateInlineSearch activates inline search mode
func (m *MainModel) activateInlineSearch() tea.Cmd {
	m.uiState.ActivateSearch()
	return m.broadcastStatusBarState()
}

// CancelInlineSearch cancels inline search mode
func (m *MainModel) cancelInlineSearch() tea.Cmd {
	m.uiState.CancelSearch()
	// Restore search state to what it was before inline search
	// This doesn't change the actual search query, just exits inline mode
	return m.broadcastStatusBarState()
}

// CommitInlineSearch applies the current search input and exits search mode
func (m *MainModel) commitInlineSearch() tea.Cmd {
	// Capture search input before state change clears it
	searchQuery := m.uiState.CommitSearch()
	return m.setSearchQuery(searchQuery) // Commit captured value to search query (app state)
}

// updateRealTimeSearch applies search filtering as user types
func (m *MainModel) updateRealTimeSearch() tea.Cmd {
	// Temporarily update search query for real-time filtering
	return m.setSearchQuery(m.uiState.SearchInput)
}

// SetSearchQuery sets the current search query and broadcasts the change
func (m *MainModel) setSearchQuery(query string) tea.Cmd {
	query = strings.TrimSpace(query)
	selectedTaskID := m.getSelectedTaskID()
	oldQuery := m.uiState.SearchQuery

	m.updateSearchState(query)
	m.findAndSelectTask(selectedTaskID)

	// Log search state change
	if oldQuery != query {
		m.programContext.Logger.LogStateChange("Model", "SearchQuery", oldQuery, query,
			"matches", m.uiState.TaskTotalMatches)
	}

	// Broadcast the search state change - components will handle it themselves
	return func() tea.Msg {
		return messages.SearchStateChangedMsg{
			Query:  m.uiState.SearchQuery,
			Active: m.uiState.SearchActive,
		}
	}
}

// getSelectedTaskID returns the ID of the currently selected task
func (m *MainModel) getSelectedTaskID() string {
	sortedTasks := m.GetSortedTasks()
	if len(sortedTasks) > 0 && m.uiState.SelectedTaskIndex < len(sortedTasks) {
		return sortedTasks[m.uiState.SelectedTaskIndex].ID
	}
	return ""
}

// updateSearchState updates search query, active state, history, and matches
func (m *MainModel) updateSearchState(query string) {
	m.uiState.SetSearchQuery(query)

	if query != "" {
		m.addToSearchHistory(query)
	}
	m.updateSearchMatches()
}

// updateTaskListSearch sends search state update to TaskList component
//
//nolint:unused // Reserved for future task list search coordination
func (m *MainModel) updateTaskListSearch() tea.Cmd {
	searchMsg := tasklist.TaskListSearchMsg{
		Query:  m.uiState.SearchQuery,
		Active: m.uiState.SearchActive,
	}
	return m.components.Layout.MainContent.Update(searchMsg)
}

// addToSearchHistory adds a query to search history, avoiding duplicates
func (m *MainModel) addToSearchHistory(query string) {
	// Delegate to ProgramContext for search history management
	m.programContext.AddToSearchHistory(query)
}

// updateSearchMatches updates the search matches for task search
func (m *MainModel) updateSearchMatches() {
	if !m.uiState.SearchActive || m.uiState.SearchQuery == "" {
		m.uiState.UpdateSearchMatches(nil, 0)
		return
	}

	// Use helper to find matching tasks
	sortedTasks := m.GetSortedTasks()
	indices, total := helpers.SearchTasks(
		sortedTasks,
		m.uiState.SearchQuery,
	)
	m.uiState.UpdateSearchMatches(indices, total)
}

// nextSearchMatch navigates to the next search match (n command)
func (m *MainModel) nextSearchMatch() tea.Cmd {
	if m.uiState.TaskTotalMatches == 0 {
		return nil
	}

	nextIndex := helpers.GetNextMatch(m.uiState.TaskMatchingIndices, m.uiState.SelectedTaskIndex)
	return m.setSelectedTask(nextIndex)
}

// previousSearchMatch navigates to the previous search match (N command)
func (m *MainModel) previousSearchMatch() tea.Cmd {
	if m.uiState.TaskTotalMatches == 0 {
		return nil
	}

	prevIndex := helpers.GetPreviousMatch(m.uiState.TaskMatchingIndices, m.uiState.SelectedTaskIndex)
	return m.setSelectedTask(prevIndex)
}

// ClearSearch clears the current search and broadcasts the change
func (m *MainModel) clearSearch() tea.Cmd {
	m.uiState.ClearSearch()

	// Broadcast the search state change - components will handle it themselves
	return func() tea.Msg {
		return messages.SearchStateChangedMsg{
			Query:  "",
			Active: false,
		}
	}
}

// =============================================================================
// UI UTILITIES - SORTING, SCROLLING, CLIPBOARD
// =============================================================================

// CycleSortMode cycles to the next sort mode
func (m *MainModel) cycleSortMode() tea.Cmd {
	// Remember currently selected task
	var selectedTaskID string
	sortedTasks := m.GetSortedTasks()
	if len(sortedTasks) > 0 && m.uiState.SelectedTaskIndex < len(sortedTasks) {
		selectedTaskID = sortedTasks[m.uiState.SelectedTaskIndex].ID
	}

	// Cycle to next sort mode - ProgramContext.SortMode is the single source of truth
	currentMode := m.programContext.SortMode
	newMode := (currentMode + 1) % 4 // 4 sort modes: Status+Priority, Priority, Time, Alphabetical

	// Log state change
	m.programContext.Logger.LogStateChange("Model", "SortMode",
		m.GetSortModeName(currentMode),
		m.GetSortModeName(newMode))

	m.programContext.SetSortMode(newMode)

	// Find the same task in new sort order and select it
	m.findAndSelectTask(selectedTaskID)

	// Update search matches for new sort order
	m.updateSearchMatches()

	// Broadcast updated state to StatusBar
	return m.broadcastStatusBarState()
}

// cycleSortModePrevious cycles to the previous sort mode
func (m *MainModel) cycleSortModePrevious() tea.Cmd {
	// Remember currently selected task
	var selectedTaskID string
	sortedTasks := m.GetSortedTasks()
	if len(sortedTasks) > 0 && m.uiState.SelectedTaskIndex < len(sortedTasks) {
		selectedTaskID = sortedTasks[m.uiState.SelectedTaskIndex].ID
	}

	// Cycle to previous sort mode - ProgramContext.SortMode is the single source of truth
	currentMode := m.programContext.SortMode
	newMode := (currentMode - 1 + 4) % 4 // 4 sort modes: wrap around
	m.programContext.SetSortMode(newMode)

	// Find the same task in new sort order and select it
	m.findAndSelectTask(selectedTaskID)

	// Update search matches for new sort order
	m.updateSearchMatches()

	// Broadcast updated state to StatusBar
	return m.broadcastStatusBarState()
}

// =============================================================================
// DATA PROVIDER METHODS
// =============================================================================

// GetTasks returns the current tasks
func (m MainModel) GetTasks() []archon.Task {
	return m.programContext.Tasks
}

// GetProjects returns the current projects
func (m MainModel) GetProjects() []archon.Project {
	return m.programContext.Projects
}

// GetSortedTasks returns the tasks sorted according to the current sort mode
// Note: Uses programContext.SelectedProjectID for filtering. When a project is selected,
// only that project's tasks are displayed. When nil (All Tasks), all tasks are shown.
func (m MainModel) GetSortedTasks() []archon.Task {
	filters := helpers.TaskFilters{
		ProjectID:          m.programContext.SelectedProjectID,  // Global state (ProgramContext)
		StatusFilters:      m.programContext.StatusFilters,      // User preference (ProgramContext)
		StatusFilterActive: m.programContext.StatusFilterActive, // Computed from StatusFilters (ProgramContext)
		FeatureFilters:     m.programContext.FeatureFilters,     // User preference (ProgramContext)
		ShowCompletedTasks: m.programContext.ShowCompletedTasks, // User preference (ProgramContext)
	}
	// ProgramContext.SortMode is the single source of truth for sort mode
	return helpers.FilterAndSortTasks(m.programContext.Tasks, m.programContext.SortMode, filters)
}

// GetSelectedTask returns the currently selected task or nil if none selected
// CRITICAL: Delegates to MainContent/TaskList to get the actual displayed task
// This ensures we always get the task the user sees on screen, not a recomputed version
func (m MainModel) GetSelectedTask() *archon.Task {
	if m.components.Layout.MainContent == nil {
		return nil
	}
	return m.components.Layout.MainContent.GetSelectedTask()
}

// GetSelectedTaskIndex returns the currently selected task index
func (m MainModel) GetSelectedTaskIndex() int {
	return m.uiState.SelectedTaskIndex
}

// GetTaskCount returns the total number of tasks
func (m MainModel) GetTaskCount() int {
	return len(m.GetTasks())
}

// GetSortedTaskCount returns the number of sorted/filtered tasks
func (m MainModel) GetSortedTaskCount() int {
	return len(m.GetSortedTasks())
}

// GetSelectedProjectID returns the currently selected project ID
func (m MainModel) GetSelectedProjectID() *string {
	return m.programContext.SelectedProjectID
}

// IsLoading returns whether the model is currently loading data
func (m MainModel) IsLoading() bool {
	return m.programContext.Loading
}

// GetError returns the current error message
func (m MainModel) GetError() string {
	return m.programContext.Error
}

// getTaskStatusCounts returns counts of tasks by status (internal helper)
func (m *MainModel) getTaskStatusCounts() (todo, doing, review, done int) {
	for _, task := range m.programContext.Tasks {
		switch task.Status {
		case archon.TaskStatusTodo:
			todo++
		case archon.TaskStatusDoing:
			doing++
		case archon.TaskStatusReview:
			review++
		case archon.TaskStatusDone:
			done++
		}
	}
	return
}

// GetTaskStatusCounts returns counts of tasks by status
func (m MainModel) GetTaskStatusCounts() (int, int, int, int) {
	return m.getTaskStatusCounts()
}

// IsProjectSelected returns true if a specific project is currently selected
func (m MainModel) IsProjectSelected() bool {
	return m.programContext.SelectedProjectID != nil
}

// GetSelectedProject returns the currently selected project, if any
func (m MainModel) GetSelectedProject() *archon.Project {
	if m.programContext.SelectedProjectID == nil {
		return nil
	}

	for _, project := range m.programContext.Projects {
		if project.ID == *m.programContext.SelectedProjectID {
			return &project
		}
	}
	return nil
}

// GetCurrentProjectName returns the name of the current project or "All Tasks"
func (m MainModel) GetCurrentProjectName() string {
	selectedProject := m.GetSelectedProject()
	if selectedProject != nil {
		return selectedProject.Title
	}
	return "All Tasks"
}

// GetTaskCountForProject returns the number of tasks for a specific project
func (m MainModel) GetTaskCountForProject(projectID string) int {
	count := 0
	for _, task := range m.programContext.Tasks {
		if task.ProjectID == projectID {
			count++
		}
	}
	return count
}

// GetTaskSearchState returns the current search state
func (m MainModel) GetTaskSearchState() (bool, string, []int, int, int) {
	return m.uiState.GetTaskSearchState(m.uiState.SelectedTaskIndex)
}

// GetScreenWidth returns the current screen width
func (m MainModel) GetScreenWidth() int {
	return m.programContext.ScreenWidth
}

// IsSearchMode returns whether search input mode is active
func (m MainModel) IsSearchMode() bool {
	return m.uiState.SearchMode
}

// GetLoadingMessage returns the current loading message
func (m MainModel) GetLoadingMessage() string {
	return m.programContext.LoadingMessage
}

// GetSearchInput returns the current search input text
func (m MainModel) GetSearchInput() string {
	return m.uiState.SearchInput
}

// GetCurrentPosition returns position info for the current selection
func (m MainModel) GetCurrentPosition() string {
	sortedTasks := m.GetSortedTasks()
	if len(sortedTasks) == 0 {
		return "No tasks"
	}

	if m.uiState.SelectedTaskIndex >= len(sortedTasks) {
		return "No selection"
	}

	return fmt.Sprintf("Task %d of %d", m.uiState.SelectedTaskIndex+1, len(sortedTasks))
}

// GetCurrentSortModeName returns the human-readable name of the current sort mode
func (m MainModel) GetCurrentSortModeName() string {
	return m.GetSortModeName(m.programContext.SortMode)
}

// GetSortModeName returns the human-readable name for a given sort mode
func (m MainModel) GetSortModeName(mode int) string {
	switch mode {
	case sorting.SortStatusPriority:
		return "Status"
	case sorting.SortPriorityOnly:
		return "Priority"
	case sorting.SortTimeCreated:
		return "Created"
	case sorting.SortAlphabetical:
		return "Alpha"
	default:
		return "Unknown"
	}
}

// TruncateStatusText intelligently truncates status text to fit the available width
func (m MainModel) TruncateStatusText(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}

	// Try to preserve the core status part by truncating from the shortcuts section
	parts := strings.Split(text, " | ")
	if len(parts) <= 1 {
		// No separators, just truncate with ellipsis
		if maxWidth <= 3 {
			return text[:maxWidth]
		}
		return text[:maxWidth-3] + "..."
	}

	// Keep the first part (core status) and as many shortcuts as possible
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		candidate := result + " | " + parts[i]
		if len(candidate) <= maxWidth {
			result = candidate
		} else {
			// Add ellipsis if we had to cut shortcuts
			if len(result+" | ...") <= maxWidth {
				result += " | ..."
			}
			break
		}
	}

	// Final fallback: if even the core part is too long
	if len(result) > maxWidth {
		if maxWidth <= 3 {
			return result[:maxWidth]
		}
		return result[:maxWidth-3] + "..."
	}

	return result
}

// CreateStyleContext creates a StyleContext for UI components with current model state
func (m MainModel) CreateStyleContext(isSelected bool) *styling.StyleContext {
	themeAdapter := &styling.ThemeAdapter{
		TodoColor:     styling.CurrentTheme.TodoColor,
		DoingColor:    styling.CurrentTheme.DoingColor,
		ReviewColor:   styling.CurrentTheme.ReviewColor,
		DoneColor:     styling.CurrentTheme.DoneColor,
		HeaderColor:   styling.CurrentTheme.HeaderColor,
		MutedColor:    styling.CurrentTheme.MutedColor,
		AccentColor:   styling.CurrentTheme.AccentColor,
		StatusColor:   styling.CurrentTheme.StatusColor,
		FeatureColors: styling.CurrentTheme.FeatureColors,
		Name:          styling.CurrentTheme.Name,
	}

	return styling.NewStyleContext(themeAdapter, m.programContext.ConfigProvider).
		WithSelection(isSelected).
		WithSearch(m.uiState.SearchQuery, m.uiState.SearchActive)
}

// GetContentHeight returns the content height for panels
func (m MainModel) GetContentHeight() int {
	return 20 // Default content height - components now manage their own dimensions
}

// GetLeftPanelWidth returns the left panel width
func (m MainModel) GetLeftPanelWidth() int {
	return 38 // Default left panel width - components now manage their own dimensions
}

// GetRightPanelWidth returns the right panel width
func (m MainModel) GetRightPanelWidth() int {
	return 42 // Default right panel width - components now manage their own dimensions
}

// GetProgramContext returns the program context for command execution
func (m MainModel) GetProgramContext() interface{} {
	return m.programContext
}

// =============================================================================
// FEATURE COORDINATION
// =============================================================================

// GetUniqueFeatures returns a sorted list of unique features from current tasks
func (m MainModel) GetUniqueFeatures() []string {
	return helpers.GetUniqueFeatures(m.programContext.Tasks)
}

// GetFeatureFilterSummary returns a summary of active feature filters
func (m MainModel) GetFeatureFilterSummary() string {
	// Delegate to ProgramContext which now owns FeatureFilters
	return m.programContext.GetFeatureFilterSummary()
}

// max returns the larger of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// =============================================================================
// CONFIGURATION HELPERS
// =============================================================================

// isRealtimeEnabled returns whether WebSocket realtime updates are enabled
// Note: Currently WebSocket is disabled as backend doesn't support it
//
//nolint:unused // Reserved for future WebSocket realtime functionality
func (m MainModel) isRealtimeEnabled() bool {
	if cfg, ok := m.programContext.ConfigProvider.(*configpkg.Config); ok {
		return cfg.IsRealtimeEnabled()
	}
	return false
}

// =============================================================================
// HTTP POLLING FOR AUTO-REFRESH
// =============================================================================

// startPolling starts the HTTP polling loop for auto-refresh
// This is used when WebSocket is disabled (backend doesn't support it)
func (m MainModel) startPolling() tea.Cmd {
	// Get polling interval from config (default: 10 seconds)
	interval := 10 * time.Second
	if cfg, ok := m.programContext.ConfigProvider.(*configpkg.Config); ok {
		interval = time.Duration(cfg.GetPollingInterval()) * time.Second
	}

	// Use tea.Tick for non-blocking, efficient timer
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return messages.PollingTickMsg{}
	})
}

// handlePollingTick handles a polling tick by refreshing tasks and scheduling the next tick
//
//nolint:ireturn // Required by Bubble Tea framework - must return tea.Model interface
func (m *MainModel) handlePollingTick() (tea.Model, tea.Cmd) {
	// Refresh tasks and projects via HTTP
	return m, tea.Batch(
		tasks.LoadTasksInterface(m.programContext.ArchonClient, m.programContext.SelectedProjectID),
		projects.LoadProjectsInterface(m.programContext.ArchonClient),
		m.startPolling(), // Schedule next polling tick
	)
}
