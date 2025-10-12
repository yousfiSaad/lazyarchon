package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/v2/internal/domain/projectmode"
	"github.com/yousfisaad/lazyarchon/v2/internal/domain/projects"
	"github.com/yousfisaad/lazyarchon/v2/internal/domain/tasks"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/projectdetails"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/projectlist"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/context"
)

// =============================================================================
// TASK & PROJECT MESSAGE HANDLERS
// =============================================================================
// This file contains handlers for task-related and project-related messages

// handleTaskMessages processes task-related messages (loaded, updated, deleted)
//
//nolint:ireturn // Required by Bubble Tea framework - must return tea.Model interface
func (m *MainModel) handleTaskMessages(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tasks.TasksLoadedMsg:
		if msg.Error != nil {
			m.setError(msg.Error.Error())
			m.setLoading(false)
			return m, nil
		}
		m.updateTasks(msg.Tasks)
		return m, nil

	case tasks.TaskUpdateMsg:
		if msg.Error != nil {
			m.setError(msg.Error.Error())
			m.setLoading(false)
			return m, nil
		}
		// Task updated successfully, refresh tasks to show changes
		return m, tasks.LoadTasksInterface(m.programContext.ArchonClient, m.programContext.SelectedProjectID)

	case tasks.TaskDeleteMsg:
		if msg.Error != nil {
			m.setError(msg.Error.Error())
			m.setLoading(false)
			return m, nil
		}
		// Task deleted successfully, refresh tasks to reflect deletion
		m.setLoadingWithMessage(true, "Refreshing tasks...")
		return m, tasks.LoadTasksInterface(m.programContext.ArchonClient, m.programContext.SelectedProjectID)
	}
	return m, nil
}

// handleProjectMessages processes project-related messages
//
//nolint:ireturn // Required by Bubble Tea framework - must return tea.Model interface
func (m *MainModel) handleProjectMessages(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(projects.ProjectsLoadedMsg); ok {
		if msg.Error != nil {
			m.setError(msg.Error.Error())
			return m, nil
		}
		m.updateProjects(msg.Projects)
		return m, nil
	}
	return m, nil
}

// handleProjectModeMessages processes project mode activation/deactivation
//
//nolint:ireturn // Required by Bubble Tea framework - must return tea.Model interface
func (m *MainModel) handleProjectModeMessages(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case projectmode.ProjectModeActivatedMsg:
		// Activate project mode by setting ProjectList component active
		m.uiState.SetActivePanel(context.LeftPanel)
		m.uiState.SetViewMode(context.ProjectViewMode)

		var cmds []tea.Cmd

		// Synchronize project list cursor with selected project filter state
		// This ensures the cursor position matches the actual filtered project
		cursorIndex := m.findProjectIndexForCursor()
		selectMsg := projectlist.ProjectListSelectMsg{Index: cursorIndex}
		cmds = append(cmds, m.components.Layout.MainContent.Update(selectMsg))

		// Note: No need to send ProjectListSetActiveMsg - component reads active state via IsComponentActive() callback

		// Initialize ProjectDetails with currently selected project
		if m.programContext.SelectedProjectID != nil {
			for _, project := range m.programContext.Projects {
				if project.ID == *m.programContext.SelectedProjectID {
					updateMsg := projectdetails.ProjectDetailsUpdateMsg{
						SelectedProject: &project,
					}
					cmds = append(cmds, m.components.Layout.MainContent.Update(updateMsg))
					break
				}
			}
		}

		// Broadcast updated state to StatusBar
		if cmd := m.broadcastStatusBarState(); cmd != nil {
			cmds = append(cmds, cmd)
		}

		return m, tea.Batch(cmds...)

	case projectmode.ProjectModeDeactivatedMsg:
		// Deactivate project mode
		m.uiState.SetViewMode(context.TaskViewMode)

		// Note: No need to send ProjectListSetActiveMsg - component reads active state via IsComponentActive() callback

		// Broadcast updated state to StatusBar
		statusBarCmd := m.broadcastStatusBarState()

		// If task loading is requested, do it after deactivation
		if msg.ShouldLoadTasks {
			loadCmd := tasks.LoadTasksInterface(m.programContext.ArchonClient, m.programContext.SelectedProjectID)
			return m, tea.Batch(statusBarCmd, loadCmd)
		}
		return m, statusBarCmd
	}
	return m, nil
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// findProjectIndexForCursor returns the cursor index that matches the current project filter state
// Returns the project's index if a specific project is selected, or len(projects) for "All Tasks"
func (m *MainModel) findProjectIndexForCursor() int {
	// If no project is selected (All Tasks filter), cursor points to "All Tasks" option
	if m.programContext.SelectedProjectID == nil {
		return len(m.programContext.Projects) // Last index is "All Tasks"
	}

	// Find the index of the selected project in the projects list
	for i, project := range m.programContext.Projects {
		if project.ID == *m.programContext.SelectedProjectID {
			return i
		}
	}

	// Fallback: if selected project not found, point to "All Tasks"
	return len(m.programContext.Projects)
}
