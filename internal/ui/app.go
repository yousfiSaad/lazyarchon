package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Init initializes the application
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		LoadTasksWithProject(m.client, m.Data.selectedProjectID),
		LoadProjects(m.client),
	)
}

// Update handles incoming events and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Window.width = msg.Width
		m.Window.height = msg.Height
		m.Window.ready = true

		// Update viewport size for task details panel
		// Right panel gets roughly half the width, minus borders/padding
		rightPanelWidth := m.GetRightPanelWidth() - 4 // Account for borders/padding
		rightPanelHeight := m.GetContentHeight() - 4  // Account for borders/padding
		m.taskDetailsViewport.Width = rightPanelWidth
		m.taskDetailsViewport.Height = rightPanelHeight

		return m, nil

	case tea.KeyMsg:
		return m.HandleKeyPress(msg.String())

	case tasksLoadedMsg:
		m.UpdateTasks(msg.tasks)
		return m, nil

	case projectsLoadedMsg:
		m.UpdateProjects(msg.projects)
		return m, nil

	case taskStatusUpdatedMsg:
		// Task status was successfully updated, refresh tasks to show changes
		return m, LoadTasksWithProject(m.client, m.Data.selectedProjectID)

	case taskFeatureUpdatedMsg:
		// Task feature was successfully updated, refresh tasks to show changes
		return m, LoadTasksWithProject(m.client, m.Data.selectedProjectID)

	case errorMsg:
		m.SetError(string(msg))
		return m, nil
	}

	return m, nil
}
