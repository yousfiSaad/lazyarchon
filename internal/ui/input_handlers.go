package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/yousfisaad/lazyarchon/internal/domain/projectmode"
	"github.com/yousfisaad/lazyarchon/internal/domain/projects"
	"github.com/yousfisaad/lazyarchon/internal/shared/utils/keys"
	"github.com/yousfisaad/lazyarchon/internal/ui/components/modals/help"
	"github.com/yousfisaad/lazyarchon/internal/ui/components/modals/status"
	"github.com/yousfisaad/lazyarchon/internal/ui/components/projectlist"
	"github.com/yousfisaad/lazyarchon/internal/ui/messages"
)

// input_handlers.go - Consolidated input handling for the Model
// This file contains all keyboard input handling using priority-based routing.
//
// WHY ONE FILE: All keyboard handling is kept together for high cohesion.
// Splitting would distribute complexity without reducing it. The file is
// well-organized with clear sections and a method index for easy navigation.
//
// PRIORITY HIERARCHY (highest to lowest):
// 1. Emergency keys (Ctrl+C, ?) - always work
// 2. Search input mode - captures typing when active
// 3. Modal keys - capture input when modal open
// 4. Application keys (p, a, r, q, esc, enter) - work across all modes
// 5. Mode-specific keys (navigation, search, task ops) - only in specific contexts
//
// ORGANIZATION:
// - Main Input Handling - HandleKeyPress with priority routing
// - Priority-Based Routing Handlers - route by priority level
// - Application Key Handlers - cross-mode operations
// - Mode-Specific Key Handlers - navigation, search, tasks
// - Low-Level Navigation Implementation - helper functions
//
// METHOD INDEX (51 methods):
//
// Main Entry Point:
//   HandleKeyPress (line 47) - Priority-based keyboard routing
//
// Priority Routing (6 methods):
//   handleGlobalKeys (line 84) - Emergency keys (Ctrl+C, ?)
//   routeToActiveModal (line 98) - Modal key delegation
//   handleProjectModeKeys (line 108) - Project mode routing
//   handleTaskModeKeys (line 173) - Task mode routing
//   handleInlineSearchInput (line 196) - Search input capture
//   handleMultiKeySequence (line 918) - Multi-key sequences (gg)
//
// Routing Dispatchers (5 methods):
//   handleApplicationKey (line 240) - Application-level routing
//   handleNavigationKey (line 267) - Navigation key routing
//   handleSearchKey (line 295) - Search key routing
//   handleTaskKey (line 312) - Task operation routing
//   handleModalKey (line 334) - Modal key routing
//
// Application Key Handlers (8 methods):
//   HandleQuitKey (line 348) - 'q' key
//   HandleEmergencyQuitKey (line 376) - Ctrl+C
//   HandleRefreshKey (line 385) - 'r' key
//   HandleProjectModeKey (line 398) - 'p' key
//   HandleShowAllTasksKey (line 408) - 'a' key
//   HandleEscapeKey (line 420) - Escape key
//   HandleConfirmKey (line 430) - Enter key
//   HandleToggleHelpKey (line 693) - '?' key
//
// Navigation Handlers (10 methods) - See input_handlers_navigation.go
//   HandleUpNavigationKey, HandleDownNavigationKey
//   HandleLeftNavigationKey, HandleRightNavigationKey
//   HandleJumpToFirstKey, HandleJumpToLastKey
//   HandleFastScrollUpKey, HandleFastScrollDownKey
//   HandleHalfPageUpKey, HandleHalfPageDownKey
//
// Search Handlers (4 methods) - See input_handlers_search.go
//   HandleActivateSearchKey, HandleClearSearchKey
//   HandleNextSearchMatchKey, HandlePrevSearchMatchKey
//
// Task Operation Handlers (7 methods) - See input_handlers_task.go
//   HandleTaskStatusChangeKey, HandleTaskEditKey
//   HandleTaskIDCopyKey, HandleTaskTitleCopyKey
//   HandleFeatureSelectionKey, HandleSortModeKey, HandleSortModePreviousKey
//
// Low-Level Navigation Implementation (9 methods) - See input_handlers_navigation.go
//   handleUpNavigation, handleDownNavigation
//   handleJumpToFirst, handleJumpToLast
//   handleFastScrollUp, handleFastScrollDown
//   handleHalfPageUp, handleHalfPageDown
//   getTaskListHalfPageSize

// =============================================================================
// 1. MAIN INPUT HANDLING - PRIORITY-BASED ROUTING
// =============================================================================

// handleKeyPress processes keyboard input using priority-based routing
// Key handling priority (highest to lowest):
// 1. Emergency keys (Ctrl+C, ?) - always work
// 2. Search input mode - captures typing when active
// 3. Modal keys - capture input when modal open
// 4. Application keys (p, a, r, q, esc, enter) - work across all modes
// 5. Mode-specific keys (navigation, task operations) - only in specific contexts
func (m *MainModel) handleKeyPress(key string) tea.Cmd {
	// 1. Global keys that work in any mode (emergency actions)
	if cmd, handled := m.handleGlobalKeys(key); handled {
		return cmd
	}

	// 2. Handle inline search input (when search mode is active)
	if m.uiState.SearchMode {
		return m.handleInlineSearchInput(key)
	}

	// 3. Modal keys (modals capture all input when active)
	if m.HasActiveModal() {
		if cmd, handled := m.handleHelpModalKey(key); handled {
			return cmd
		}
	}

	// 4. Application-level keys (work across all modes)
	if cmd, handled := m.handleApplicationKey(key); handled {
		return cmd
	}

	// 5. Mode-specific routing based on current application state
	if m.uiState.IsProjectView() {
		return m.handleProjectModeKeys(key)
	} else {
		return m.handleTaskModeKeys(key)
	}
}

// =============================================================================
// PRIORITY-BASED ROUTING HANDLERS
// =============================================================================

// handleGlobalKeys processes emergency keys that work in any mode
// These bypass all other handling for critical operations
func (m *MainModel) handleGlobalKeys(key string) (tea.Cmd, bool) {
	switch key {
	case keys.KeyCtrlC:
		// Emergency quit - always works regardless of modals or mode
		return tea.Quit, true
	case keys.KeyQuestion:
		// Help - works globally
		return func() tea.Msg { return help.ShowHelpModalMsg{} }, true
	default:
		return nil, false
	}
}

// handleProjectModeKeys processes keys when in project selection mode
func (m *MainModel) handleProjectModeKeys(key string) tea.Cmd {
	switch key {
	case "q", keys.KeyEscape:
		// Exit project mode - these are the only keys that should exit
		return func() tea.Msg { return projectmode.ProjectModeDeactivatedMsg{ShouldLoadTasks: false} }

	case keys.KeyJ, keys.KeyArrowDown:
		// Navigate down - route based on active panel
		if m.IsLeftPanelActive() {
			// Navigate down in project list - route through content component
			scrollMsg := projectlist.ProjectListScrollMsg{Direction: projectlist.ScrollDown}
			cmd := m.components.Layout.MainContent.Update(scrollMsg)
			return cmd
		} else if m.IsRightPanelActive() {
			// Future: scroll down in project details
			return nil
		}

	case keys.KeyK, keys.KeyArrowUp:
		// Navigate up - route based on active panel
		if m.IsLeftPanelActive() {
			// Navigate up in project list - route through content component
			scrollMsg := projectlist.ProjectListScrollMsg{Direction: projectlist.ScrollUp}
			cmd := m.components.Layout.MainContent.Update(scrollMsg)
			return cmd
		} else if m.IsRightPanelActive() {
			// Future: scroll up in project details
			return nil
		}

	case keys.KeyH:
		// Switch focus to left panel (project list)
		return m.setActiveView(LeftPanel)

	case keys.KeyL:
		// Switch focus to right panel (project details)
		return m.setActiveView(RightPanel)

	case keys.KeyEnter:
		// Exit project mode and load tasks for currently selected project
		// Note: Selected project is already tracked via ProjectListSelectionChangedMsg handler in app.go
		if len(m.programContext.Projects) > 0 {
			var cmds []tea.Cmd
			if cmd := m.setLoadingWithMessage(true, "Loading project tasks..."); cmd != nil {
				cmds = append(cmds, cmd)
			}
			cmds = append(cmds, func() tea.Msg { return projectmode.ProjectModeDeactivatedMsg{ShouldLoadTasks: true} })
			return tea.Batch(cmds...)
		}

	case keys.KeyGG, keys.KeyHome:
		// Jump to first project - route through content component
		scrollMsg := projectlist.ProjectListScrollMsg{Direction: projectlist.ScrollToTop}
		cmd := m.components.Layout.MainContent.Update(scrollMsg)
		return cmd

	case keys.KeyGCap, keys.KeyEnd:
		// Jump to last project - route through content component
		scrollMsg := projectlist.ProjectListScrollMsg{Direction: projectlist.ScrollToBottom}
		cmd := m.components.Layout.MainContent.Update(scrollMsg)
		return cmd

	case keys.KeyY:
		// Copy project ID - send yank message to components
		return func() tea.Msg { return messages.YankIDMsg{} }

	case keys.KeyYCap:
		// Copy project title - send yank message to components
		return func() tea.Msg { return messages.YankTitleMsg{} }
	}

	// All other keys are ignored in project mode
	return nil
}

// handleTaskModeKeys processes keys when in normal task view mode
// Note: Application keys (p, a, r, q, etc.) are handled before this function is called
func (m *MainModel) handleTaskModeKeys(key string) tea.Cmd {
	// Handle multi-key sequences (like 'gg')
	if cmd, handled := m.handleMultiKeySequence(key); handled {
		return cmd
	}

	// Route to mode-specific handlers (navigation, search, task operations)
	// Application keys are no longer checked here - handled at higher priority level
	if cmd, handled := m.handleNavigationKey(key); handled {
		return cmd
	}
	if cmd, handled := m.handleSearchKey(key); handled {
		return cmd
	}
	if cmd, handled := m.handleTaskKey(key); handled {
		return cmd
	}

	// Key not handled
	return nil
}

// handleInlineSearchInput processes input when inline search mode is active
func (m *MainModel) handleInlineSearchInput(key string) tea.Cmd {
	switch key {
	case "esc":
		// Cancel search and revert to previous state
		m.cancelInlineSearch()
		return nil

	case "enter":
		// Commit current search input
		return m.commitInlineSearch()

	case "backspace":
		// Remove last character
		if len(m.uiState.SearchInput) > 0 {
			m.uiState.SearchInput = m.uiState.SearchInput[:len(m.uiState.SearchInput)-1]
			// Update search in real-time
			return m.updateRealTimeSearch()
		}
		return nil

	case "ctrl+u":
		// Clear entire input
		m.uiState.SearchInput = ""
		return m.updateRealTimeSearch()

	default:
		// Handle printable characters
		if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
			m.uiState.SearchInput = m.uiState.SearchInput + key
			// Update search in real-time
			return m.updateRealTimeSearch()
		}
		return nil
	}
}

// =============================================================================
// 2. KEY ROUTING - APPLICATION LEVEL
// =============================================================================

// handleApplicationKey routes application-level keys to their specific handlers
// These keys work across all modes (task mode and project mode) for consistent UX
func (m *MainModel) handleApplicationKey(key string) (tea.Cmd, bool) {
	switch key {
	case keys.KeyQ:
		// Don't handle 'q' in project mode - let project mode handler deal with it
		// This allows 'q' to behave like Escape (exit project mode, not show quit modal)
		if m.uiState.IsTaskView() {
			return m.handleQuitKey(key)
		}
		return nil, false // Not handled, pass to project mode handler
	case keys.KeyCtrlC:
		return m.handleEmergencyQuitKey(key)
	case keys.KeyR, keys.KeyF5:
		return m.handleRefreshKey(key)
	case keys.KeyP:
		return m.handleProjectModeKey(key)
	case keys.KeyA:
		return m.handleShowAllTasksKey(key)
	case keys.KeyEscape:
		return m.handleEscapeKey(key)
	case keys.KeyEnter:
		return m.handleConfirmKey(key)
	default:
		return nil, false
	}
}

// =============================================================================
// 3. KEY ROUTING - MODE SPECIFIC
// =============================================================================

// handleNavigationKey routes navigation keys to their specific handlers
// These are mode-specific and behavior depends on current mode (task/project)
func (m *MainModel) handleNavigationKey(key string) (tea.Cmd, bool) {
	switch key {
	case keys.KeyArrowUp, keys.KeyK:
		return m.handleUpNavigationKey(key)
	case keys.KeyArrowDown, keys.KeyJ:
		return m.handleDownNavigationKey(key)
	case keys.KeyH:
		return m.handleLeftNavigationKey(key)
	case keys.KeyL:
		return m.handleRightNavigationKey(key)
	case keys.KeyGG, keys.KeyHome:
		return m.handleJumpToFirstKey(key)
	case keys.KeyGCap, keys.KeyEnd:
		return m.handleJumpToLastKey(key)
	case keys.KeyJCap:
		return m.handleFastScrollDownKey(key)
	case keys.KeyKCap:
		return m.handleFastScrollUpKey(key)
	case keys.KeyCtrlU, keys.KeyPgUp:
		return m.handleHalfPageUpKey(key)
	case keys.KeyCtrlD, keys.KeyPgDn:
		return m.handleHalfPageDownKey(key)
	default:
		return nil, false
	}
}

// handleSearchKey routes search keys to their specific handlers
func (m *MainModel) handleSearchKey(key string) (tea.Cmd, bool) {
	// Debug output for testing
	switch key {
	case keys.KeySlash, keys.KeyCtrlF:
		return m.handleActivateSearchKey(key)
	case keys.KeyCtrlX, keys.KeyCtrlL:
		return m.handleClearSearchKey(key)
	case keys.KeyN:
		return m.handleNextSearchMatchKey(key)
	case keys.KeyNCap:
		return m.handlePrevSearchMatchKey(key)
	default:
		return nil, false
	}
}

// handleTaskKey routes task operation keys to their specific handlers
func (m *MainModel) handleTaskKey(key string) (tea.Cmd, bool) {
	switch key {
	case keys.KeyT:
		return m.handleTaskStatusChangeKey(key)
	case keys.KeyE:
		return m.handleTaskEditKey(key)
	case keys.KeyD:
		return m.handleTaskDeleteKey(key)
	case keys.KeyY:
		return m.handleTaskIDCopyKey(key)
	case keys.KeyYCap:
		return m.handleTaskTitleCopyKey(key)
	case keys.KeyF:
		return m.handleFeatureSelectionKey(key)
	case keys.KeyS:
		return m.handleSortModeKey(key)
	case keys.KeySCap:
		return m.handleSortModePreviousKey(key)
	default:
		return nil, false
	}
}

// handleHelpModalKey routes modal activation keys to their specific handlers
func (m *MainModel) handleHelpModalKey(key string) (tea.Cmd, bool) {
	switch key {
	case keys.KeyQuestion:
		return m.handleToggleHelpKey(key)
	default:
		return nil, false
	}
}

// =============================================================================
// APPLICATION KEY HANDLERS (Work Across All Modes)
// =============================================================================

// HandleQuitKey handles the 'q' key - smart quit behavior with modal awareness
func (m *MainModel) handleQuitKey(key string) (tea.Cmd, bool) {
	// Smart quit behavior: if search mode is active, close it first
	if m.uiState.SearchMode {
		m.uiState.CancelSearch()
		return nil, true
	}

	// Then check for other active modals
	if m.HasActiveModal() {
		// Close the active modal instead of quitting
		if m.components.Modals.HelpModel.IsActive() {
			return func() tea.Msg { return help.HideHelpModalMsg{} }, true
		} else if m.components.Modals.StatusModel.IsActive() {
			return func() tea.Msg { return status.HideStatusModalMsg{} }, true
		} else if m.uiState.IsProjectView() {
			// Use message-based approach to deactivate project mode (no task loading needed)
			return func() tea.Msg { return projectmode.ProjectModeDeactivatedMsg{ShouldLoadTasks: false} }, true
		} else if m.components.Modals.TaskEditModel.IsActive() {
			m.components.Modals.TaskEditModel.SetActive(false)
		}
		return nil, true
	} else {
		// No modal active, show quit confirmation
		return m.showQuitConfirmation(), true
	}
}

// HandleEmergencyQuitKey handles 'ctrl+c' key - emergency quit bypassing modals
func (m *MainModel) handleEmergencyQuitKey(key string) (tea.Cmd, bool) {
	if key == keys.KeyCtrlC {
		// Emergency quit - always works regardless of modals
		return tea.Quit, true
	}
	return nil, true
}

// HandleRefreshKey handles 'r' and 'F5' keys - refresh/retry operation
func (m *MainModel) handleRefreshKey(key string) (tea.Cmd, bool) {
	var cmds []tea.Cmd
	if m.programContext.Error != "" {
		// Retry last failed operation
		if cmd := m.clearError(); cmd != nil {
			cmds = append(cmds, cmd)
		}
		if cmd := m.setLoadingWithMessage(true, "Retrying..."); cmd != nil {
			cmds = append(cmds, cmd)
		}
	} else {
		// Regular refresh
		if cmd := m.setLoadingWithMessage(true, "Refreshing data..."); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	cmds = append(cmds, projects.RefreshDataInterface(m.programContext.ArchonClient, m.programContext.SelectedProjectID))
	return tea.Batch(cmds...), true
}

// HandleProjectModeKey handles 'p' key - activate project selection
func (m *MainModel) handleProjectModeKey(key string) (tea.Cmd, bool) {
	if m.uiState.IsTaskView() {
		// Use message-based approach instead of direct state mutation
		cmd := func() tea.Msg { return projectmode.ProjectModeActivatedMsg{} }
		return cmd, true
	}
	return nil, true
}

// HandleShowAllTasksKey handles 'a' key - show all tasks
func (m *MainModel) handleShowAllTasksKey(key string) (tea.Cmd, bool) {
	// Show all tasks - works from any mode
	m.setSelectedProject(nil)

	var cmds []tea.Cmd
	if cmd := m.setLoadingWithMessage(true, "Loading all tasks..."); cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Use message-based approach to deactivate project mode and load tasks sequentially
	deactivateCmd := func() tea.Msg { return projectmode.ProjectModeDeactivatedMsg{ShouldLoadTasks: true} }
	cmds = append(cmds, deactivateCmd)

	return tea.Batch(cmds...), true
}

// HandleEscapeKey handles 'esc' key - general escape/cancel
func (m *MainModel) handleEscapeKey(key string) (tea.Cmd, bool) {
	if m.uiState.IsProjectView() {
		// Use message-based approach to deactivate project mode (no task loading needed)
		cmd := func() tea.Msg { return projectmode.ProjectModeDeactivatedMsg{ShouldLoadTasks: false} }
		return cmd, true
	}
	return nil, false // Not handled in other contexts
}

// HandleConfirmKey handles 'enter' key - general confirmation
func (m *MainModel) handleConfirmKey(key string) (tea.Cmd, bool) {
	if m.uiState.IsProjectView() && len(m.programContext.Projects) > 0 {
		// Exit project mode and load tasks for currently selected project
		// Note: Selected project is already tracked via ProjectListSelectionChangedMsg handler in app.go
		var cmds []tea.Cmd
		if cmd := m.setLoadingWithMessage(true, "Loading project tasks..."); cmd != nil {
			cmds = append(cmds, cmd)
		}

		// Use message-based approach to deactivate project mode and load tasks sequentially
		deactivateCmd := func() tea.Msg { return projectmode.ProjectModeDeactivatedMsg{ShouldLoadTasks: true} }
		cmds = append(cmds, deactivateCmd)

		return tea.Batch(cmds...), true
	}
	return nil, false // Not handled in other contexts
}

// =============================================================================
// MODAL KEY HANDLERS
// =============================================================================

// HandleToggleHelpKey handles '?' key - toggle help modal
func (m *MainModel) handleToggleHelpKey(key string) (tea.Cmd, bool) {
	// Use the component-based approach
	if m.components.Modals.HelpModel.IsActive() {
		return func() tea.Msg { return help.HideHelpModalMsg{} }, true
	} else {
		return func() tea.Msg { return help.ShowHelpModalMsg{} }, true
	}
}

// =============================================================================
// MULTI-KEY SEQUENCES
// =============================================================================

// handleMultiKeySequence handles multi-key sequences like 'gg'
// NOTE: This is legacy implementation. NavigationCoordinator should handle this in the future.
func (m *MainModel) handleMultiKeySequence(key string) (tea.Cmd, bool) {
	// TODO: Migrate to NavigationCoordinator when it's fully implemented
	// For now, we'll disable multi-key sequences since the state was removed
	// This simplifies the cleanup while maintaining basic functionality

	if key == keys.KeyG {
		// Since we removed the key sequence state, we'll just handle single 'g' as jump to first
		// Users can press 'g' twice quickly for the same effect
		if m.components.Modals.HelpModel.IsActive() {
			// Help modal navigation is handled by the component
			return nil, true
		} else if m.components.Modals.FeatureModel.IsActive() {
			// Feature modal - handled by component system
			return nil, true
		} else {
			cmd := m.handleJumpToFirst()
			return cmd, true
		}
	}

	return nil, false
}
