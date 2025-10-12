package projects

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/v2/internal/domain/tasks"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/interfaces"
)

// =============================================================================
// PROJECT DOMAIN COMMANDS
// =============================================================================
// Command functions for project-related operations

// LoadProjectsInterface loads projects using interface dependency (preferred for DI)
func LoadProjectsInterface(client interfaces.ArchonClient) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.ListProjects()
		if err != nil {
			return ProjectsLoadedMsg{Error: err}
		}

		return ProjectsLoadedMsg{Projects: resp.Projects}
	}
}

// RefreshDataInterface refreshes both tasks and projects using interface dependency (preferred for DI)
func RefreshDataInterface(client interfaces.ArchonClient, selectedProjectID *string) tea.Cmd {
	return tea.Batch(
		tasks.LoadTasksInterface(client, selectedProjectID),
		LoadProjectsInterface(client),
	)
}
