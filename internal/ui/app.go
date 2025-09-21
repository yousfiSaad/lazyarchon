package ui

import (
	"time"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/internal/ui/commands"
)

// tickMsg is sent periodically to animate the loading spinner
type tickMsg time.Time

// tick sends a tickMsg after a short delay for spinner animation
func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Init initializes the application
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		commands.LoadTasksWithProject(m.client, m.Data.selectedProjectID),
		commands.LoadProjects(m.client),
		commands.InitializeRealtimeCmd(m.wsClient),                  // Initialize WebSocket connection
		commands.ListenForRealtimeEvents(m.wsClient),               // Start listening for events
		tick(),                                            // Start spinner animation
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

		// Update help modal viewport if help is currently active
		if m.IsHelpMode() {
			modalWidth := Min(m.Window.width-4, 70)   // Maximum 70 chars wide, with margins
			modalHeight := Min(m.Window.height-4, 25) // Maximum 25 lines high, with margins
			contentHeight := modalHeight - 4          // Account for border and padding
			contentWidth := modalWidth - 4            // Account for border and padding
			
			m.helpModalViewport.Width = contentWidth
			m.helpModalViewport.Height = contentHeight
			// Refresh help content with new width constraints
			m.updateHelpModalViewport()
		}

		// Refresh task details content to reflow text at new width
		m.updateTaskDetailsViewport()

		return m, nil

	case tea.KeyMsg:
		return m.HandleKeyPress(msg.String())

	case commands.TasksLoadedMsgSimple:
		m.UpdateTasks(msg.Tasks)
		return m, nil

	case commands.ProjectsLoadedMsgSimple:
		m.UpdateProjects(msg.Projects)
		return m, nil

	case commands.TaskStatusUpdatedMsgSimple:
		// Task status was successfully updated, refresh tasks to show changes
		return m, commands.LoadTasksWithProject(m.client, m.Data.selectedProjectID)

	case commands.TaskFeatureUpdatedMsgSimple:
		// Task feature was successfully updated, refresh tasks to show changes
		return m, commands.LoadTasksWithProject(m.client, m.Data.selectedProjectID)

	case commands.ErrorMsgSimple:
		m.SetError(string(msg))
		return m, nil

	case tickMsg:
		// Advance spinner animation if loading
		if m.Data.loading {
			m.AdvanceSpinner()
		}
		return m, tick() // Continue animation

	// WebSocket real-time events
	case commands.RealtimeConnectedMsg:
		// Connection established - update UI status
		m.SetConnectionStatus(true)
		return m, commands.ListenForRealtimeEvents(m.wsClient) // Continue listening for events

	case commands.RealtimeDisconnectedMsg:
		// Connection lost - update UI status
		m.SetConnectionStatus(false)
		// Try to reconnect after a delay (the WebSocket client handles this internally)
		return m, nil

	case commands.RealtimeTaskUpdateMsg:
		// Task was updated - refresh the task list to show changes
		return m, tea.Batch(
			commands.LoadTasksWithProject(m.client, m.Data.selectedProjectID),
			commands.ListenForRealtimeEvents(m.wsClient), // Continue listening for events
		)

	case commands.RealtimeTaskCreateMsg:
		// New task was created - refresh the task list
		return m, tea.Batch(
			commands.LoadTasksWithProject(m.client, m.Data.selectedProjectID),
			commands.ListenForRealtimeEvents(m.wsClient), // Continue listening for events
		)

	case commands.RealtimeTaskDeleteMsg:
		// Task was deleted - refresh the task list
		return m, tea.Batch(
			commands.LoadTasksWithProject(m.client, m.Data.selectedProjectID),
			commands.ListenForRealtimeEvents(m.wsClient), // Continue listening for events
		)

	case commands.RealtimeProjectUpdateMsg:
		// Project was updated - refresh both tasks and projects
		return m, tea.Batch(
			commands.LoadTasksWithProject(m.client, m.Data.selectedProjectID),
			commands.LoadProjects(m.client),
			commands.ListenForRealtimeEvents(m.wsClient), // Continue listening for events
		)
	}

	return m, nil
}
