package commands

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/internal/archon"
	"github.com/yousfisaad/lazyarchon/internal/interfaces"
)

// Messages for async operations
type TasksLoadedMsgSimple struct {
	Tasks []archon.Task
}

type ProjectsLoadedMsgSimple struct {
	Projects []archon.Project
}

type ErrorMsgSimple string

type TaskStatusUpdatedMsgSimple struct {
	TaskID    string
	NewStatus string
}

type TaskFeatureUpdatedMsgSimple struct {
	TaskID     string
	NewFeature string
}

// Additional message types for dependency injection
type TasksLoadedMsg struct {
	Tasks []archon.Task
	Error error
}

type ProjectsLoadedMsg struct {
	Projects []archon.Project
	Error    error
}

type TaskUpdateMsg struct {
	Task  *archon.Task
	Error error
}

// Real-time WebSocket message types
type RealtimeConnectedMsg struct{}

type RealtimeDisconnectedMsg struct {
	Error error
}

type RealtimeTaskUpdateMsg struct {
	TaskID string
	Task   archon.Task
	Old    *archon.Task
}

type RealtimeTaskCreateMsg struct {
	Task archon.Task
}

type RealtimeTaskDeleteMsg struct {
	TaskID string
	Task   archon.Task
}

type RealtimeProjectUpdateMsg struct {
	ProjectID string
	Project   archon.Project
	Old       *archon.Project
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
			return ErrorMsgSimple("Failed to load tasks: " + err.Error())
		}

		return TasksLoadedMsgSimple{Tasks: resp.Tasks}
	}
}

// LoadProjects loads projects from the API
func LoadProjects(client *archon.Client) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.ListProjects()
		if err != nil {
			// Return error if projects can't be loaded
			// App will still work with task list only
			return ErrorMsgSimple("Failed to load projects: " + err.Error())
		}

		return ProjectsLoadedMsgSimple{Projects: resp.Projects}
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
			return ErrorMsgSimple("Failed to update task status: " + err.Error())
		}

		return TaskStatusUpdatedMsgSimple{
			TaskID:    taskID,
			NewStatus: newStatus,
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
			return ErrorMsgSimple("Failed to update task feature: " + err.Error())
		}

		return TaskFeatureUpdatedMsgSimple{
			TaskID:     taskID,
			NewFeature: newFeature,
		}
	}
}

// InitializeRealtimeCmd sets up the WebSocket connection and event handlers
func InitializeRealtimeCmd(wsClient interfaces.RealtimeClient) tea.Cmd {
	return func() tea.Msg {
		// Set up legacy event handlers (optional, mainly for backward compatibility)
		wsClient.SetEventHandlers(
			// Task update handler
			func(event archon.TaskUpdateEvent) {
				// Legacy handler - events are now handled via channel
			},
			// Task create handler
			func(event archon.TaskCreateEvent) {
				// Legacy handler - events are now handled via channel
			},
			// Task delete handler
			func(event archon.TaskDeleteEvent) {
				// Legacy handler - events are now handled via channel
			},
			// Project update handler
			func(event archon.ProjectUpdateEvent) {
				// Legacy handler - events are now handled via channel
			},
			// Connect handler
			func() {
				// Legacy handler - events are now handled via channel
			},
			// Disconnect handler
			func(err error) {
				// Legacy handler - events are now handled via channel
			},
		)

		// Attempt to connect
		if err := wsClient.Connect(); err != nil {
			return RealtimeDisconnectedMsg{Error: err}
		}

		return RealtimeConnectedMsg{}
	}
}

// ListenForRealtimeEvents creates a command that listens for WebSocket events
// and converts them to Bubble Tea messages
func ListenForRealtimeEvents(wsClient interfaces.RealtimeClient) tea.Cmd {
	return func() tea.Msg {
		// Block and wait for the next event from the WebSocket client
		eventCh := wsClient.GetEventChannel()

		select {
		case event := <-eventCh:
			// Convert archon message types to ui message types
			switch e := event.(type) {
			case archon.RealtimeTaskCreateMsg:
				return RealtimeTaskCreateMsg{Task: e.Task}
			case archon.RealtimeTaskUpdateMsg:
				return RealtimeTaskUpdateMsg{
					TaskID: e.TaskID,
					Task:   e.Task,
					Old:    e.Old,
				}
			case archon.RealtimeTaskDeleteMsg:
				return RealtimeTaskDeleteMsg{
					TaskID: e.TaskID,
					Task:   e.Task,
				}
			case archon.RealtimeProjectUpdateMsg:
				return RealtimeProjectUpdateMsg{
					ProjectID: e.ProjectID,
					Project:   e.Project,
					Old:       e.Old,
				}
			case archon.RealtimeConnectedMsg:
				return RealtimeConnectedMsg{}
			case archon.RealtimeDisconnectedMsg:
				return RealtimeDisconnectedMsg{Error: e.Error}
			default:
				// Unknown event type, return nil to ignore
				return nil
			}
		}
	}
}
