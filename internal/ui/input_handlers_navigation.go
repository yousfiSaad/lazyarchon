package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/v2/internal/domain/projectmode"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/viewport"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/projectlist"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/taskdetails"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/tasklist"
)

// =============================================================================
// NAVIGATION KEY HANDLERS
// =============================================================================
// This file contains all navigation-related keyboard handlers

// HandleUpNavigationKey handles 'up' and 'k' keys - move up/scroll up
//
//nolint:unparam // key parameter intentionally unused - handler is dispatched by routing layer
func (m *MainModel) handleUpNavigationKey(key string) (tea.Cmd, bool) {
	cmd := m.handleUpNavigation()
	return cmd, true
}

// HandleDownNavigationKey handles 'down' and 'j' keys - move down/scroll down
//
//nolint:unparam // key parameter intentionally unused - handler is dispatched by routing layer
func (m *MainModel) handleDownNavigationKey(key string) (tea.Cmd, bool) {
	cmd := m.handleDownNavigation()
	return cmd, true
}

// HandleLeftNavigationKey handles 'h' key - switch to left panel or go back
//
//nolint:unparam // key parameter intentionally unused - handler is dispatched by routing layer
func (m *MainModel) handleLeftNavigationKey(key string) (tea.Cmd, bool) {
	if m.uiState.IsProjectView() {
		// In project mode, h goes back - use message-based approach (no task loading needed)
		cmd := func() tea.Msg { return projectmode.ProjectModeDeactivatedMsg{ShouldLoadTasks: false} }
		return cmd, true
	} else {
		// In task view mode, h switches to left panel
		cmd := m.setActiveView(LeftPanel)
		return cmd, true
	}
}

// HandleRightNavigationKey handles 'l' key - switch to right panel or go forward
//
//nolint:unparam // key parameter intentionally unused - handler is dispatched by routing layer
func (m *MainModel) handleRightNavigationKey(key string) (tea.Cmd, bool) {
	if m.uiState.IsProjectView() && len(m.programContext.Projects) > 0 {
		// In project mode, l selects project using component state through content component
		// Exit project mode and load tasks for currently selected project
		// Note: Selected project is already tracked via ProjectListSelectionChangedMsg handler in app.go
		m.setLoadingWithMessage(true, "Loading project tasks...")

		// Use message-based approach to deactivate project mode and load tasks sequentially
		deactivateCmd := func() tea.Msg { return projectmode.ProjectModeDeactivatedMsg{ShouldLoadTasks: true} }

		return deactivateCmd, true
	} else if m.uiState.IsTaskView() {
		// In task view mode, l switches to right panel
		cmd := m.setActiveView(RightPanel)
		return cmd, true
	}
	return nil, true
}

// HandleJumpToFirstKey handles 'gg' and 'home' keys - jump to first item
//
//nolint:unparam // key parameter intentionally unused - handler is dispatched by routing layer
func (m *MainModel) handleJumpToFirstKey(key string) (tea.Cmd, bool) {
	cmd := m.handleJumpToFirst()
	return cmd, true
}

// HandleJumpToLastKey handles 'G' and 'end' keys - jump to last item
func (m *MainModel) handleJumpToLastKey(key string) (tea.Cmd, bool) {
	if key == "end" {
		// Special handling for end key - delegate to handleJumpToLast for consistent behavior
		cmd := m.handleJumpToLast()
		return cmd, true
	}
	cmd := m.handleJumpToLast()
	return cmd, true
}

// HandleFastScrollDownKey handles 'J' key - fast scroll down
//
//nolint:unparam // key parameter intentionally unused - handler is dispatched by routing layer
func (m *MainModel) handleFastScrollDownKey(key string) (tea.Cmd, bool) {
	return m.handleFastScrollDown(), true
}

// HandleFastScrollUpKey handles 'K' key - fast scroll up
//
//nolint:unparam // key parameter intentionally unused - handler is dispatched by routing layer
func (m *MainModel) handleFastScrollUpKey(key string) (tea.Cmd, bool) {
	return m.handleFastScrollUp(), true
}

// HandleHalfPageUpKey handles 'ctrl+u' and 'pgup' keys - half-page scroll up
func (m *MainModel) handleHalfPageUpKey(key string) (tea.Cmd, bool) {
	if key == "home" {
		// Special handling for home key - delegate to handleJumpToFirst for consistent behavior
		cmd := m.handleJumpToFirst()
		return cmd, true
	}
	return m.handleHalfPageUp(), true
}

// HandleHalfPageDownKey handles 'ctrl+d' and 'pgdown' keys - half-page scroll down
//
//nolint:unparam // key parameter intentionally unused - handler is dispatched by routing layer
func (m *MainModel) handleHalfPageDownKey(key string) (tea.Cmd, bool) {
	return m.handleHalfPageDown(), true
}

// =============================================================================
// LOW-LEVEL NAVIGATION IMPLEMENTATION
// =============================================================================

// handleUpNavigation handles up arrow or 'k' key - respects active panel
func (m *MainModel) handleUpNavigation() tea.Cmd {
	if m.uiState.IsProjectView() {
		// Navigate projects using component message - route through content component
		scrollMsg := projectlist.ProjectListScrollMsg{Direction: projectlist.ScrollUp}
		cmd := m.components.Layout.MainContent.Update(scrollMsg)
		return cmd
	} else {
		// Respect active panel for navigation
		if m.IsLeftPanelActive() {
			// Navigate tasks using component message - route through content component
			scrollMsg := tasklist.TaskListScrollMsg{Direction: tasklist.ScrollUp}
			cmd := m.components.Layout.MainContent.Update(scrollMsg)
			return cmd
		} else if m.IsRightPanelActive() {
			// Send scroll up message to task details component - route through content component
			scrollMsg := taskdetails.TaskDetailsScrollMsg{Direction: viewport.ScrollUp}
			cmd := m.components.Layout.MainContent.Update(scrollMsg)
			return cmd
		}
	}
	return nil
}

// handleDownNavigation handles down arrow or 'j' key - respects active panel
func (m *MainModel) handleDownNavigation() tea.Cmd {
	if m.uiState.IsProjectView() {
		// Navigate projects using component message - route through content component
		scrollMsg := projectlist.ProjectListScrollMsg{Direction: projectlist.ScrollDown}
		cmd := m.components.Layout.MainContent.Update(scrollMsg)
		return cmd
	} else {
		// Respect active panel for navigation
		if m.IsLeftPanelActive() {
			// Navigate tasks using component message - route through content component
			scrollMsg := tasklist.TaskListScrollMsg{Direction: tasklist.ScrollDown}
			cmd := m.components.Layout.MainContent.Update(scrollMsg)
			return cmd
		} else if m.IsRightPanelActive() {
			// Send scroll down message to task details component - route through content component
			scrollMsg := taskdetails.TaskDetailsScrollMsg{Direction: viewport.ScrollDown}
			cmd := m.components.Layout.MainContent.Update(scrollMsg)
			return cmd
		}
	}
	return nil
}

// handleJumpToFirst handles 'gg' key - jump to first item in active panel
func (m *MainModel) handleJumpToFirst() tea.Cmd {
	if m.uiState.IsProjectView() {
		// Jump to first project using component message - route through content component
		scrollMsg := projectlist.ProjectListScrollMsg{Direction: projectlist.ScrollToTop}
		cmd := m.components.Layout.MainContent.Update(scrollMsg)
		return cmd
	} else {
		if m.IsLeftPanelActive() {
			// Jump to first task using component message - route through content component
			scrollMsg := tasklist.TaskListScrollMsg{Direction: tasklist.ScrollToTop}
			cmd := m.components.Layout.MainContent.Update(scrollMsg)
			return cmd
		} else if m.IsRightPanelActive() {
			// Send scroll to top message to task details component - route through content component
			scrollMsg := taskdetails.TaskDetailsScrollMsg{Direction: viewport.ScrollToTop}
			cmd := m.components.Layout.MainContent.Update(scrollMsg)
			return cmd
		}
	}
	return nil
}

// handleJumpToLast handles 'G' key - jump to last item in active panel
func (m *MainModel) handleJumpToLast() tea.Cmd {
	if m.uiState.IsProjectView() {
		// Jump to last project using component message - route through content component
		scrollMsg := projectlist.ProjectListScrollMsg{Direction: projectlist.ScrollToBottom}
		cmd := m.components.Layout.MainContent.Update(scrollMsg)
		return cmd
	} else {
		if m.IsLeftPanelActive() {
			// Jump to last task using component message - route through content component
			scrollMsg := tasklist.TaskListScrollMsg{Direction: tasklist.ScrollToBottom}
			cmd := m.components.Layout.MainContent.Update(scrollMsg)
			return cmd
		} else if m.IsRightPanelActive() {
			// Send scroll to bottom message to task details component - route through content component
			scrollMsg := taskdetails.TaskDetailsScrollMsg{Direction: viewport.ScrollToBottom}
			cmd := m.components.Layout.MainContent.Update(scrollMsg)
			return cmd
		}
	}
	return nil
}

// handleFastScrollUp handles 'K' key - fast scroll up (4 lines) in active panel
func (m *MainModel) handleFastScrollUp() tea.Cmd {
	if m.uiState.IsProjectView() {
		return nil // No scrolling in project mode
	}

	if m.IsLeftPanelActive() {
		// Fast scroll up in task list using component message - route through content component
		scrollMsg := tasklist.TaskListScrollMsg{Direction: tasklist.ScrollFastUp}
		cmd := m.components.Layout.MainContent.Update(scrollMsg)
		return cmd
	} else if m.IsRightPanelActive() {
		// Send fast scroll up message to task details component - route through content component
		scrollMsg := taskdetails.TaskDetailsScrollMsg{Direction: viewport.ScrollFastUp}
		cmd := m.components.Layout.MainContent.Update(scrollMsg)
		return cmd
	}
	return nil
}

// handleFastScrollDown handles 'J' key - fast scroll down (4 lines) in active panel
func (m *MainModel) handleFastScrollDown() tea.Cmd {
	if m.uiState.IsProjectView() {
		return nil // No scrolling in project mode
	}

	if m.IsLeftPanelActive() {
		// Fast scroll down in task list using component message - route through content component
		scrollMsg := tasklist.TaskListScrollMsg{Direction: tasklist.ScrollFastDown}
		cmd := m.components.Layout.MainContent.Update(scrollMsg)
		return cmd
	} else if m.IsRightPanelActive() {
		// Send fast scroll down message to task details component - route through content component
		scrollMsg := taskdetails.TaskDetailsScrollMsg{Direction: viewport.ScrollFastDown}
		cmd := m.components.Layout.MainContent.Update(scrollMsg)
		return cmd
	}
	return nil
}

// handleHalfPageUp handles 'Ctrl+u' key - half-page scroll up in active panel
func (m *MainModel) handleHalfPageUp() tea.Cmd {
	if m.uiState.IsProjectView() {
		return nil // No scrolling in project mode
	}

	if m.IsLeftPanelActive() {
		// Half-page scroll up in task list using component message - route through content component
		scrollMsg := tasklist.TaskListScrollMsg{Direction: tasklist.ScrollPageUp}
		cmd := m.components.Layout.MainContent.Update(scrollMsg)
		return cmd
	} else if m.IsRightPanelActive() {
		// Send half-page scroll up message to task details component - route through content component
		scrollMsg := taskdetails.TaskDetailsScrollMsg{Direction: viewport.ScrollHalfPageUp}
		cmd := m.components.Layout.MainContent.Update(scrollMsg)
		return cmd
	}
	return nil
}

// handleHalfPageDown handles 'Ctrl+d' key - half-page scroll down in active panel
func (m *MainModel) handleHalfPageDown() tea.Cmd {
	if m.uiState.IsProjectView() {
		return nil // No scrolling in project mode
	}

	if m.IsLeftPanelActive() {
		// Half-page scroll down in task list using component message - route through content component
		scrollMsg := tasklist.TaskListScrollMsg{Direction: tasklist.ScrollPageDown}
		cmd := m.components.Layout.MainContent.Update(scrollMsg)
		return cmd
	} else if m.IsRightPanelActive() {
		// Send half-page scroll down message to task details component - route through content component
		scrollMsg := taskdetails.TaskDetailsScrollMsg{Direction: viewport.ScrollHalfPageDown}
		cmd := m.components.Layout.MainContent.Update(scrollMsg)
		return cmd
	}
	return nil
}

// getTaskListHalfPageSize calculates half-page size for task list scrolling
//
//nolint:unused // Reserved for future half-page scrolling calculations
func (m MainModel) getTaskListHalfPageSize() int {
	contentHeight := 20                 // Default content height - components now manage their own dimensions
	halfPage := (contentHeight - 6) / 2 // Account for header and padding
	if halfPage < 1 {
		halfPage = 1 // Minimum scroll amount
	}
	return halfPage
}
