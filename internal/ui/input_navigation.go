package ui

// handleUpNavigation handles up arrow or 'k' key - respects active panel
func (m Model) handleUpNavigation() Model {
	if m.Modals.projectMode.active {
		// Navigate projects
		if m.Modals.projectMode.selectedIndex > 0 {
			m.Modals.projectMode.selectedIndex--
		}
	} else {
		// Respect active panel for navigation
		if m.IsLeftPanelActive() {
			// Navigate tasks (existing behavior)
			if m.Navigation.selectedIndex > 0 {
				m.setSelectedTask(m.Navigation.selectedIndex - 1)
			}
		} else if m.IsRightPanelActive() {
			// Scroll content up using viewport
			m.taskDetailsViewport.LineUp(1)
		}
	}
	return m
}

// handleDownNavigation handles down arrow or 'j' key - respects active panel
func (m Model) handleDownNavigation() Model {
	if m.Modals.projectMode.active {
		// Navigate projects (including "All Tasks" option)
		maxIndex := len(m.Data.projects) // +1 for "All Tasks" option, but 0-indexed
		if m.Modals.projectMode.selectedIndex < maxIndex {
			m.Modals.projectMode.selectedIndex++
		}
	} else {
		// Respect active panel for navigation
		if m.IsLeftPanelActive() {
			// Navigate tasks (existing behavior)
			sortedTasks := m.GetSortedTasks()
			if m.Navigation.selectedIndex < len(sortedTasks)-1 {
				m.setSelectedTask(m.Navigation.selectedIndex + 1)
			}
		} else if m.IsRightPanelActive() {
			// Scroll content down using viewport
			m.taskDetailsViewport.LineDown(1)
		}
	}
	return m
}

// handleJumpToFirst handles 'gg' key - jump to first item in active panel
func (m Model) handleJumpToFirst() Model {
	if m.Modals.projectMode.active {
		m.Modals.projectMode.selectedIndex = 0
	} else {
		if m.IsLeftPanelActive() {
			// Jump to first task
			m.setSelectedTask(0)
		} else if m.IsRightPanelActive() {
			// Jump to top of task details
			m.taskDetailsViewport.GotoTop()
		}
	}
	return m
}

// handleJumpToLast handles 'G' key - jump to last item in active panel
func (m Model) handleJumpToLast() Model {
	if m.Modals.projectMode.active {
		m.Modals.projectMode.selectedIndex = len(m.Data.projects) // Last item is "All Tasks"
	} else {
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
	return m
}

// handleFastScrollUp handles 'K' key - fast scroll up (4 lines) in active panel
func (m Model) handleFastScrollUp() Model {
	if m.Modals.projectMode.active {
		return m // No scrolling in project mode
	}

	if m.IsLeftPanelActive() {
		// Fast scroll up in task list (move selection up by 4)
		fastScrollLines := 4
		newIndex := m.Navigation.selectedIndex - fastScrollLines
		if newIndex < 0 {
			newIndex = 0
		}
		m.setSelectedTask(newIndex)
	} else if m.IsRightPanelActive() {
		// Fast scroll up in task details panel (4 lines)
		m.taskDetailsViewport.LineUp(4)
	}
	return m
}

// handleFastScrollDown handles 'J' key - fast scroll down (4 lines) in active panel
func (m Model) handleFastScrollDown() Model {
	if m.Modals.projectMode.active {
		return m // No scrolling in project mode
	}

	if m.IsLeftPanelActive() {
		// Fast scroll down in task list (move selection down by 4)
		sortedTasks := m.GetSortedTasks()
		if len(sortedTasks) > 0 {
			fastScrollLines := 4
			newIndex := m.Navigation.selectedIndex + fastScrollLines
			maxIndex := len(sortedTasks) - 1
			if newIndex > maxIndex {
				newIndex = maxIndex
			}
			m.setSelectedTask(newIndex)
		}
	} else if m.IsRightPanelActive() {
		// Fast scroll down in task details panel (4 lines)
		m.taskDetailsViewport.LineDown(4)
	}
	return m
}

// handleHalfPageUp handles 'Ctrl+u' key - half-page scroll up in active panel
func (m Model) handleHalfPageUp() Model {
	if m.Modals.projectMode.active {
		return m // No scrolling in project mode
	}

	if m.IsLeftPanelActive() {
		// Half-page scroll up in task list
		halfPage := m.getTaskListHalfPageSize()
		newIndex := m.Navigation.selectedIndex - halfPage
		if newIndex < 0 {
			newIndex = 0
		}
		m.setSelectedTask(newIndex)
	} else if m.IsRightPanelActive() {
		// Half-page scroll up in task details panel
		m.taskDetailsViewport.HalfViewUp()
	}
	return m
}

// handleHalfPageDown handles 'Ctrl+d' key - half-page scroll down in active panel
func (m Model) handleHalfPageDown() Model {
	if m.Modals.projectMode.active {
		return m // No scrolling in project mode
	}

	if m.IsLeftPanelActive() {
		// Half-page scroll down in task list
		sortedTasks := m.GetSortedTasks()
		if len(sortedTasks) > 0 {
			halfPage := m.getTaskListHalfPageSize()
			newIndex := m.Navigation.selectedIndex + halfPage
			maxIndex := len(sortedTasks) - 1
			if newIndex > maxIndex {
				newIndex = maxIndex
			}
			m.setSelectedTask(newIndex)
		}
	} else if m.IsRightPanelActive() {
		// Half-page scroll down in task details panel
		m.taskDetailsViewport.HalfViewDown()
	}
	return m
}

// getTaskListHalfPageSize calculates half-page size for task list scrolling
func (m Model) getTaskListHalfPageSize() int {
	contentHeight := m.GetContentHeight()
	halfPage := (contentHeight - 6) / 2 // Account for header and padding
	if halfPage < 1 {
		halfPage = 1 // Minimum scroll amount
	}
	return halfPage
}

// getDetailHalfPageSize calculates half-page size for detail panel scrolling
func (m Model) getDetailHalfPageSize() int {
	contentHeight := m.GetContentHeight()
	halfPage := (contentHeight - 4) / 2 // Account for border and padding
	if halfPage < 1 {
		halfPage = 1 // Minimum jump amount
	}
	return halfPage
}
