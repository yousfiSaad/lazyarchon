package ui

// SetSelectedProject sets the currently selected project
func (m *Model) SetSelectedProject(projectID *string) {
	m.Data.selectedProjectID = projectID
	m.setSelectedTask(0) // Reset task selection

	// Reset feature filters when changing projects
	// This ensures users see all features in the new project context
	m.Modals.featureMode.selectedFeatures = nil

	// Update search matches after project filter change
	m.updateSearchMatches()
}

// findAndSelectTask finds a task by ID in the current sort order and selects it
func (m *Model) findAndSelectTask(taskID string) {
	if taskID == "" {
		m.setSelectedTask(0)
		return
	}

	sortedTasks := m.GetSortedTasks()
	for i, task := range sortedTasks {
		if task.ID == taskID {
			m.setSelectedTask(i)
			return
		}
	}

	// Task not found, default to first task
	m.setSelectedTask(0)
}

// setSelectedTask sets the selected task index and updates viewport content
// This ensures viewport content is always in sync with the selected task
func (m *Model) setSelectedTask(index int) {
	sortedTasks := m.GetSortedTasks()

	// Bounds check the index
	if index < 0 {
		index = 0
	} else if index >= len(sortedTasks) {
		if len(sortedTasks) > 0 {
			index = len(sortedTasks) - 1
		} else {
			index = 0
		}
	}

	// Only update if actually changing
	if m.Navigation.selectedIndex != index {
		m.Navigation.selectedIndex = index
		m.taskDetailsViewport.GotoTop() // Reset scroll when changing tasks
		m.updateTaskDetailsViewport()   // Update viewport content
	}
}