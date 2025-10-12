package tasks

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/v2/internal/archon"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/interfaces"
)

// =============================================================================
// TASK DOMAIN COMMANDS
// =============================================================================
// Command functions for task-related operations

// LoadTasksInterface loads all tasks using interface dependency (preferred for DI)
// Note: Always loads ALL tasks (projectID parameter is ignored) to ensure task counts
// are accurate for all projects. Filtering by project happens at the UI display layer.
func LoadTasksInterface(client interfaces.ArchonClient, projectID *string) tea.Cmd {
	return func() tea.Msg {
		// Always pass nil to load ALL tasks, regardless of selected project
		// This ensures GetTaskCountForProject() can count tasks for all projects
		resp, err := client.ListTasks(nil, nil, true) // include_closed=true for full visibility
		if err != nil {
			return TasksLoadedMsg{Error: err}
		}

		return TasksLoadedMsg{Tasks: resp.Tasks}
	}
}

// UpdateTaskStatusInterface updates a task's status using interface dependency (preferred for DI)
func UpdateTaskStatusInterface(client interfaces.ArchonClient, taskID string, newStatus string) tea.Cmd {
	return func() tea.Msg {
		// Create update request
		updateRequest := archon.UpdateTaskRequest{
			Status: &newStatus,
		}

		// Call API to update task
		resp, err := client.UpdateTask(taskID, updateRequest)
		if err != nil {
			return TaskUpdateMsg{Error: err}
		}

		return TaskUpdateMsg{Task: &resp.Task}
	}
}

// UpdateTaskFeatureInterface updates a task's feature using interface dependency (preferred for DI)
func UpdateTaskFeatureInterface(client interfaces.ArchonClient, taskID string, newFeature string) tea.Cmd {
	return func() tea.Msg {
		// Create update request
		updateRequest := archon.UpdateTaskRequest{
			Feature: &newFeature,
		}

		// Call API to update task
		resp, err := client.UpdateTask(taskID, updateRequest)
		if err != nil {
			return TaskUpdateMsg{Error: err}
		}

		return TaskUpdateMsg{Task: &resp.Task}
	}
}

// UpdateTaskWithRequest updates a task with a custom update request (for multi-field updates)
// This is the most flexible method - allows updating any combination of fields in one call
func UpdateTaskWithRequest(client interfaces.ArchonClient, taskID string, updateRequest archon.UpdateTaskRequest) tea.Cmd {
	return func() tea.Msg {
		// Call API to update task with the provided request
		resp, err := client.UpdateTask(taskID, updateRequest)
		if err != nil {
			return TaskUpdateMsg{Error: err}
		}

		return TaskUpdateMsg{Task: &resp.Task}
	}
}

// DeleteTaskInterface deletes/archives a task using interface dependency
func DeleteTaskInterface(client interfaces.ArchonClient, taskID string) tea.Cmd {
	return func() tea.Msg {
		// Call API to delete task
		err := client.DeleteTask(taskID)
		if err != nil {
			return TaskDeleteMsg{TaskID: taskID, Error: err}
		}

		return TaskDeleteMsg{TaskID: taskID, Error: nil}
	}
}
