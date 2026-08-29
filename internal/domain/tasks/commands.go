package tasks

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/interfaces"
)

// =============================================================================
// TASK DOMAIN COMMANDS
// =============================================================================
// Command functions for task-related operations

// LoadTasksInterface loads all tasks using interface dependency (preferred for DI)
// Note: Always loads ALL tasks (projectID parameter is ignored) to ensure task counts
// are accurate for all projects. Filtering by project happens at the UI display layer.
func LoadTasksInterface(client interfaces.TaskClient, projectID *string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Always pass nil to load ALL tasks, regardless of selected project
		// This ensures GetTaskCountForProject() can count tasks for all projects
		filters := interfaces.TaskFilters{
			IncludeArchived: true, // include_closed=true for full visibility
		}

		resp, err := client.ListTasks(ctx, filters)
		if err != nil {
			return TasksLoadedMsg{Error: err}
		}

		return TasksLoadedMsg{Tasks: resp.Tasks}
	}
}

// CreateTaskInterface creates a new task using interface dependency (preferred for DI)
func CreateTaskInterface(client interfaces.TaskClient, request interfaces.CreateTaskRequest) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Call API to create task
		task, err := client.CreateTask(ctx, request)
		if err != nil {
			return TaskCreateMsg{Error: err}
		}

		return TaskCreateMsg{Task: task}
	}
}

// UpdateTaskStatusInterface updates a task's status using interface dependency (preferred for DI)
func UpdateTaskStatusInterface(client interfaces.TaskClient, taskID string, newStatus string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Create update request
		updateRequest := interfaces.UpdateTaskRequest{
			Status: &newStatus,
		}

		// Call API to update task
		task, err := client.UpdateTask(ctx, taskID, updateRequest)
		if err != nil {
			return TaskUpdateMsg{Error: err}
		}

		return TaskUpdateMsg{Task: task}
	}
}

// UpdateTaskFeatureInterface updates a task's feature/tag using interface dependency (preferred for DI)
// Note: In the generic interface, features are stored as Tags. This updates the first tag.
func UpdateTaskFeatureInterface(client interfaces.TaskClient, taskID string, newFeature string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Create update request - features are now tags in the generic interface
		tags := []string{newFeature}
		updateRequest := interfaces.UpdateTaskRequest{
			Tags: &tags,
		}

		// Call API to update task
		task, err := client.UpdateTask(ctx, taskID, updateRequest)
		if err != nil {
			return TaskUpdateMsg{Error: err}
		}

		return TaskUpdateMsg{Task: task}
	}
}

// UpdateTaskWithRequest updates a task with a custom update request (for multi-field updates)
// This is the most flexible method - allows updating any combination of fields in one call
func UpdateTaskWithRequest(client interfaces.TaskClient, taskID string, updateRequest interfaces.UpdateTaskRequest) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Call API to update task with the provided request
		task, err := client.UpdateTask(ctx, taskID, updateRequest)
		if err != nil {
			return TaskUpdateMsg{Error: err}
		}

		return TaskUpdateMsg{Task: task}
	}
}

// DeleteTaskInterface deletes/archives a task using interface dependency
func DeleteTaskInterface(client interfaces.TaskClient, taskID string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Call API to delete task
		err := client.DeleteTask(ctx, taskID)
		if err != nil {
			return TaskDeleteMsg{TaskID: taskID, Error: err}
		}

		return TaskDeleteMsg{TaskID: taskID, Error: nil}
	}
}
