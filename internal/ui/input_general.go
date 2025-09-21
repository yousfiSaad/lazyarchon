package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/yousfisaad/lazyarchon/internal/ui/input"
)

// HandleKeyPress processes keyboard input and returns updated model and commands
func (m Model) HandleKeyPress(key string) (Model, tea.Cmd) {
	// Handle help modal keys first (when help is open)
	if m.IsHelpMode() {
		return m.handleHelpModeInput(key)
	}

	// Handle status change modal keys (when status change modal is open)
	if m.IsStatusChangeMode() {
		return m.handleStatusChangeModeInput(key)
	}

	// Handle confirmation modal keys (when confirmation modal is open)
	if m.IsConfirmationMode() {
		return m.handleConfirmationModeInput(key)
	}

	// Handle feature modal keys (when feature modal is open)
	if m.IsFeatureModeActive() {
		return m.handleFeatureModeInput(key)
	}

	// Handle task edit modal keys (when task edit modal is open)
	if m.IsTaskEditModeActive() {
		return m.handleTaskEditModeInput(key)
	}

	// Handle inline search input (when search mode is active)
	if m.Data.searchMode {
		return m.handleInlineSearchInput(key)
	}

	// Handle multi-key sequences (like 'gg')
	if newM, cmd, handled := m.handleMultiKeySequence(key); handled {
		return newM, cmd
	}

	switch key {
	// Application controls
	case "q":
		// Smart quit behavior: if any modal is active, close it instead of quitting
		if m.HasActiveModal() {
			// Close the active modal instead of quitting
			if m.IsHelpMode() {
				m.SetHelpMode(false)
			} else if m.IsStatusChangeMode() {
				m.SetStatusChangeMode(false)
			} else if m.Modals.projectMode.active {
				m.Modals.projectMode.active = false
			} else if m.IsTaskEditModeActive() {
				m.SetTaskEditMode(false)
			} else if m.Data.searchMode {
				m.CancelInlineSearch()
			}
			return m, nil
		} else {
			// No modal active, show quit confirmation
			m.ShowQuitConfirmation()
			return m, nil
		}
	case "ctrl+c":
		if input.IsApplicationKey(key) {
			// Emergency quit - always works regardless of modals
			return m, tea.Quit
		}
		return m, nil

	// Project mode controls
	case "p":
		if !m.Modals.projectMode.active {
			m.Modals.projectMode.active = true
			m.Modals.projectMode.selectedIndex = 0
		}
		return m, nil

	case "a":
		// Show all tasks - works from any mode
		m.SetSelectedProject(nil)
		m.Modals.projectMode.active = false
		m.SetLoadingWithMessage(true, "Loading all tasks...")
		return m, LoadTasksWithProject(m.client, m.Data.selectedProjectID)

	case "h":
		if m.Modals.projectMode.active {
			// In project mode, h goes back (existing behavior)
			m.Modals.projectMode.active = false
		} else {
			// In task view mode, h switches to left panel
			m.SetActiveView(LeftPanel)
		}
		return m, nil

	case "esc":
		if m.Modals.projectMode.active {
			m.Modals.projectMode.active = false
		}
		return m, nil

	case "l":
		if m.Modals.projectMode.active && len(m.Data.projects) > 0 {
			// In project mode, l selects project (existing behavior)
			if m.Modals.projectMode.selectedIndex < len(m.Data.projects) {
				// Select specific project
				m.SetSelectedProject(&m.Data.projects[m.Modals.projectMode.selectedIndex].ID)
			} else {
				// Select "All Tasks" option
				m.SetSelectedProject(nil)
			}
			m.Modals.projectMode.active = false
			m.SetLoadingWithMessage(true, "Loading project tasks...")
			return m, LoadTasksWithProject(m.client, m.Data.selectedProjectID)
		} else if !m.Modals.projectMode.active {
			// In task view mode, l switches to right panel
			m.SetActiveView(RightPanel)
		}
		return m, nil

	case "enter":
		if m.Modals.projectMode.active && len(m.Data.projects) > 0 {
			// In project mode, enter selects project (existing behavior)
			if m.Modals.projectMode.selectedIndex < len(m.Data.projects) {
				// Select specific project
				m.SetSelectedProject(&m.Data.projects[m.Modals.projectMode.selectedIndex].ID)
			} else {
				// Select "All Tasks" option
				m.SetSelectedProject(nil)
			}
			m.Modals.projectMode.active = false
			m.SetLoadingWithMessage(true, "Loading project tasks...")
			return m, LoadTasksWithProject(m.client, m.Data.selectedProjectID)
		}
		return m, nil

	// Navigation controls
	case "up", "k":
		return m.handleUpNavigation(), nil

	case "down", "j":
		return m.handleDownNavigation(), nil

	case "gg":
		return m.handleJumpToFirst(), nil

	case "G":
		return m.handleJumpToLast(), nil

	// Fast scrolling controls (work on active panel)
	case "J":
		return m.handleFastScrollDown(), nil

	case "K":
		return m.handleFastScrollUp(), nil

	case "ctrl+u", "pgup":
		return m.handleHalfPageUp(), nil

	case "ctrl+d", "pgdown":
		return m.handleHalfPageDown(), nil

	case "home":
		if !m.Modals.projectMode.active {
			if m.IsLeftPanelActive() {
				// Jump to first task
				m.setSelectedTask(0)
			} else if m.IsRightPanelActive() {
				// Jump to top of task details
				m.taskDetailsViewport.GotoTop()
			}
		}
		return m, nil

	case "end":
		if !m.Modals.projectMode.active {
			if m.IsLeftPanelActive() {
				// Jump to last task
				sortedTasks := m.GetSortedTasks()
				if len(sortedTasks) > 0 {
					m.setSelectedTask(len(sortedTasks) - 1)
				}
			} else if m.IsRightPanelActive() {
				// Jump to bottom of task details
				m.taskDetailsViewport.GotoBottom()
			}
		}
		return m, nil

	// Sorting controls (only works in task view mode)
	case "s":
		if !m.Modals.projectMode.active {
			m.CycleSortMode()
		}
		return m, nil

	case "S":
		if !m.Modals.projectMode.active {
			m.CycleSortModePrevious()
		}
		return m, nil

	// Help modal
	case "?":
		m.SetHelpMode(!m.IsHelpMode())
		return m, nil

	// Status change modal
	case "t":
		if input.IsTaskOperationKey(key) && !m.Modals.projectMode.active && len(m.Data.tasks) > 0 && m.Navigation.selectedIndex < len(m.GetSortedTasks()) {
			m.SetStatusChangeMode(true)
		}
		return m, nil

	// Task edit modal
	case "e":
		if input.IsTaskOperationKey(key) && !m.Modals.projectMode.active && len(m.Data.tasks) > 0 && m.Navigation.selectedIndex < len(m.GetSortedTasks()) {
			m.SetTaskEditMode(true)
		}
		return m, nil

	// Feature selection modal
	case "f":
		if input.IsTaskOperationKey(key) && !m.Modals.projectMode.active && len(m.GetUniqueFeatures()) > 0 {
			m.SetFeatureMode(true)
		}
		return m, nil

	// Yank (copy) functionality
	case "y":
		if input.IsTaskOperationKey(key) && !m.Modals.projectMode.active && len(m.Data.tasks) > 0 && m.Navigation.selectedIndex < len(m.GetSortedTasks()) {
			return m.handleTaskIDCopy()
		}
		return m, nil

	case "Y":
		if input.IsTaskOperationKey(key) && !m.Modals.projectMode.active && len(m.Data.tasks) > 0 && m.Navigation.selectedIndex < len(m.GetSortedTasks()) {
			return m.handleTaskTitleCopy()
		}
		return m, nil

	// Inline search activation
	case "/", "ctrl+f":
		if !m.Modals.projectMode.active && !m.Data.searchMode {
			m.ActivateInlineSearch()
		}
		return m, nil

	// Clear search
	case "ctrl+x", "ctrl+l":
		if !m.Modals.projectMode.active && m.Data.searchActive {
			m.ClearSearch()
		}
		return m, nil

	// Refresh/Retry
	case "r", "F5":
		if m.Data.error != "" {
			// Retry last failed operation
			m.ClearError()
			m.SetLoadingWithMessage(true, "Retrying...")
		} else {
			// Regular refresh
			m.SetLoadingWithMessage(true, "Refreshing data...")
		}
		return m, RefreshData(m.client, m.Data.selectedProjectID)

	// Search navigation (n/N keys for next/previous match)
	case "n":
		if !m.Modals.projectMode.active && m.Data.searchActive && m.Data.totalMatches > 0 {
			m.nextSearchMatch()
		}
		return m, nil

	case "N":
		if !m.Modals.projectMode.active && m.Data.searchActive && m.Data.totalMatches > 0 {
			m.previousSearchMatch()
		}
		return m, nil

	default:
		return m, nil
	}
}

// handleInlineSearchInput processes input when inline search mode is active
func (m Model) handleInlineSearchInput(key string) (Model, tea.Cmd) {
	switch key {
	case "esc":
		// Cancel search and revert to previous state
		m.CancelInlineSearch()
		return m, nil

	case "enter":
		// Commit current search input
		m.CommitInlineSearch()
		return m, nil

	case "backspace":
		// Remove last character
		if len(m.Data.searchInput) > 0 {
			m.Data.searchInput = m.Data.searchInput[:len(m.Data.searchInput)-1]
			// Update search in real-time
			m.UpdateRealTimeSearch()
		}
		return m, nil

	case "ctrl+u":
		// Clear entire input
		m.Data.searchInput = ""
		m.UpdateRealTimeSearch()
		return m, nil

	default:
		// Handle printable characters
		if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
			m.Data.searchInput += key
			// Update search in real-time
			m.UpdateRealTimeSearch()
		}
		return m, nil
	}
}
