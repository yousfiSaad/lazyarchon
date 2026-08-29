package projects

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/v2/internal/domain/tasks"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/interfaces"
)

// =============================================================================
// PROJECT DOMAIN COMMANDS
// =============================================================================
// Command functions for project-related operations

// LoadProjectsInterface loads projects using interface dependency (preferred for DI)
func LoadProjectsInterface(client interfaces.TaskClient) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		projects, err := client.ListProjects(ctx)
		if err != nil {
			return ProjectsLoadedMsg{Error: err}
		}

		return ProjectsLoadedMsg{Projects: projects}
	}
}

// RefreshDataInterface refreshes both tasks and projects using interface dependency (preferred for DI)
func RefreshDataInterface(client interfaces.TaskClient, selectedProjectID *string) tea.Cmd {
	return tea.Batch(
		tasks.LoadTasksInterface(client, selectedProjectID),
		LoadProjectsInterface(client),
	)
}
