package ui

import (
	tea "github.com/charmbracelet/bubbletea"
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
			}
			return m, nil
		} else {
			// No modal active, show quit confirmation
			m.ShowQuitConfirmation()
			return m, nil
		}
	case "ctrl+c":
		// Emergency quit - always works regardless of modals
		return m, tea.Quit

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
		m.SetLoading(true)
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
			m.SetLoading(true)
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
			m.SetLoading(true)
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
				m.Navigation.selectedIndex = 0
				m.taskDetailsViewport.GotoTop()
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
					m.Navigation.selectedIndex = len(sortedTasks) - 1
					m.taskDetailsViewport.GotoTop() // Reset scroll for new task
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
		if !m.Modals.projectMode.active && len(m.Data.tasks) > 0 && m.Navigation.selectedIndex < len(m.GetSortedTasks()) {
			m.SetStatusChangeMode(true)
		}
		return m, nil

	// Task edit modal
	case "e":
		if !m.Modals.projectMode.active && len(m.Data.tasks) > 0 && m.Navigation.selectedIndex < len(m.GetSortedTasks()) {
			m.SetTaskEditMode(true)
		}
		return m, nil

	// Feature selection modal
	case "f":
		if !m.Modals.projectMode.active && len(m.GetUniqueFeatures()) > 0 {
			m.SetFeatureMode(true)
		}
		return m, nil

	// Refresh
	case "r", "F5":
		m.SetLoading(true)
		return m, RefreshData(m.client, m.Data.selectedProjectID)

	default:
		return m, nil
	}
}
