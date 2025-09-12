package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/internal/archon"
)

// Messages for async operations
type tasksLoadedMsg struct {
	tasks []archon.Task
}

type projectsLoadedMsg struct {
	projects []archon.Project
}

type errorMsg string

type taskStatusUpdatedMsg struct {
	taskID    string
	newStatus string
}

type taskFeatureUpdatedMsg struct {
	taskID     string
	newFeature string
}

// LoadTasks loads all tasks from the API (deprecated - use LoadTasksWithProject)
func LoadTasks(client *archon.Client) tea.Cmd {
	return LoadTasksWithProject(client, nil)
}

// LoadTasksWithProject loads tasks from the API, optionally filtered by project
// Now includes completed tasks by default to show full project progress
func LoadTasksWithProject(client *archon.Client, projectID *string) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.ListTasks(projectID, nil, true) // include_closed=true for full visibility
		if err != nil {
			return errorMsg("Failed to load tasks: " + err.Error())
		}

		return tasksLoadedMsg{tasks: resp.Tasks}
	}
}

// LoadProjects loads projects from the API
func LoadProjects(client *archon.Client) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.ListProjects()
		if err != nil {
			// Return error if projects can't be loaded
			// App will still work with task list only
			return errorMsg("Failed to load projects: " + err.Error())
		}

		return projectsLoadedMsg{projects: resp.Projects}
	}
}

// RefreshData refreshes both tasks and projects
func RefreshData(client *archon.Client, selectedProjectID *string) tea.Cmd {
	return tea.Batch(
		LoadTasksWithProject(client, selectedProjectID),
		LoadProjects(client),
	)
}

// UpdateTaskStatusCmd updates a task's status via API
func UpdateTaskStatusCmd(client *archon.Client, taskID string, newStatus string) tea.Cmd {
	return func() tea.Msg {
		// Create update request
		updateRequest := archon.UpdateTaskRequest{
			Status: &newStatus,
		}

		// Call API to update task
		_, err := client.UpdateTask(taskID, updateRequest)
		if err != nil {
			return errorMsg("Failed to update task status: " + err.Error())
		}

		return taskStatusUpdatedMsg{
			taskID:    taskID,
			newStatus: newStatus,
		}
	}
}

// UpdateTaskFeatureCmd updates a task's feature via API
func UpdateTaskFeatureCmd(client *archon.Client, taskID string, newFeature string) tea.Cmd {
	return func() tea.Msg {
		// Create update request
		updateRequest := archon.UpdateTaskRequest{
			Feature: &newFeature,
		}

		// Call API to update task
		_, err := client.UpdateTask(taskID, updateRequest)
		if err != nil {
			return errorMsg("Failed to update task feature: " + err.Error())
		}

		return taskFeatureUpdatedMsg{
			taskID:     taskID,
			newFeature: newFeature,
		}
	}
}
