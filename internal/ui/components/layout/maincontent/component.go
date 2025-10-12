package maincontent

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yousfisaad/lazyarchon/v2/internal/archon"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/base"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/projectdetails"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/projectlist"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/taskdetails"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/tasklist"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/messages"
)

const ComponentID = "main_content_component"

// MainContentModel represents the main content area component that owns the panel components
type MainContentModel struct {
	base.BaseComponent

	// Panel components - owned by this component (stored as values)
	projectListComponent    projectlist.ProjectListModel
	projectDetailsComponent projectdetails.ProjectDetailsModel
	taskListComponent       tasklist.TaskListModel
	taskDetailsComponent    taskdetails.TaskdetailsModel
}

// GetSelectedTask returns the currently selected task from the TaskList component
// This is the single source of truth for what task is selected and displayed
func (m *MainContentModel) GetSelectedTask() *archon.Task {
	return m.taskListComponent.GetSelectedTask()
}

// NewModel creates a new main content component with owned panel components
func NewModel(context *base.ComponentContext) *MainContentModel {
	baseComponent := base.NewBaseComponent(ComponentID, base.MainContentComponent, context)

	// Create panel components owned by this component
	taskDetailsComponent := taskdetails.NewModel(taskdetails.Options{
		Width:   42, // Default right panel width - will be updated via messages
		Height:  20, // Default height - will be updated via messages
		Context: context,
	})

	taskListComponent := tasklist.NewModel(tasklist.Options{
		Width:         38,              // Default left panel width - will be updated via messages
		Height:        20,              // Default height - will be updated via messages
		Tasks:         []archon.Task{}, // Will be updated via messages
		SelectedIndex: 0,
		SearchQuery:   "",
		SearchActive:  false,
		Context:       context,
	})

	projectListComponent := projectlist.NewModel(projectlist.Options{
		Width:         38,                 // Default left panel width - will be updated via messages
		Height:        20,                 // Default height - will be updated via messages
		Projects:      []archon.Project{}, // Will be updated via messages
		SelectedIndex: 0,
		Loading:       false,
		ErrorMessage:  "",
		Context:       context,
	})

	projectDetailsComponent := projectdetails.NewModel(projectdetails.Options{
		Width:              42,    // Default right panel width - will be updated via messages
		Height:             20,    // Default height - will be updated via messages
		SelectedProject:    nil,   // Will be updated via messages
		IsRightPanelActive: false, // Default to inactive (left panel is active initially)
		Context:            context,
	})

	model := &MainContentModel{
		BaseComponent:           baseComponent,
		projectListComponent:    projectListComponent,
		projectDetailsComponent: projectDetailsComponent,
		taskListComponent:       taskListComponent,
		taskDetailsComponent:    taskDetailsComponent,
	}
	// Set default dimensions - will be overridden by parent
	model.SetDimensions(80, 20)
	return model
}

// Init initializes the main content component
func (m *MainContentModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the main content component
func (m *MainContentModel) Update(msg tea.Msg) tea.Cmd {
	// Handle window resize messages to update dimensions
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Store own dimensions
		m.HandleWindowResize(msg)

		// Simple 50/50 split for child panels
		leftPanelWidth := msg.Width / 2
		rightPanelWidth := msg.Width - leftPanelWidth

		// Always resize all components - ensures correct dimensions regardless of current mode
		// This is simpler and guarantees components have proper dimensions when mode switches
		var cmds []tea.Cmd

		if cmd := m.projectListComponent.Update(tea.WindowSizeMsg{
			Width:  leftPanelWidth,
			Height: msg.Height,
		}); cmd != nil {
			cmds = append(cmds, cmd)
		}

		if cmd := m.projectDetailsComponent.Update(tea.WindowSizeMsg{
			Width:  rightPanelWidth,
			Height: msg.Height,
		}); cmd != nil {
			cmds = append(cmds, cmd)
		}

		if cmd := m.taskListComponent.Update(tea.WindowSizeMsg{
			Width:  leftPanelWidth,
			Height: msg.Height,
		}); cmd != nil {
			cmds = append(cmds, cmd)
		}

		if cmd := m.taskDetailsComponent.Update(tea.WindowSizeMsg{
			Width:  rightPanelWidth,
			Height: msg.Height,
		}); cmd != nil {
			cmds = append(cmds, cmd)
		}

		return tea.Batch(cmds...)

	// Route panel-specific messages to internal components
	case projectlist.ProjectListScrollMsg, projectlist.ProjectListUpdateMsg,
		projectlist.ProjectListSelectMsg,
		projectlist.ProjectListSelectionQueryMsg,
		projectlist.ProjectListConfirmSelectionMsg:
		cmd := m.projectListComponent.Update(msg)
		return cmd

	case tasklist.TaskListScrollMsg, tasklist.TaskListUpdateMsg,
		tasklist.TaskListSelectMsg,
		tasklist.TaskListSearchMsg,
		tasklist.TaskListFilterMsg:
		cmd := m.taskListComponent.Update(msg)
		return cmd

	case projectlist.ProjectListSelectionChangedMsg:
		// In project mode, update right panel to show selected project details
		selectedProject := m.projectListComponent.GetSelectedProject()
		updateMsg := projectdetails.ProjectDetailsUpdateMsg{
			SelectedProject: selectedProject,
		}
		return m.projectDetailsComponent.Update(updateMsg)

	case tasklist.TaskListSelectionChangedMsg:
		// Intercept selection changes to update task details viewport
		// Get the newly selected task from task list component
		selectedTask := m.taskListComponent.GetSelectedTask()

		// Send update message to task details with new task
		// Note: Search highlighting state is managed separately through TaskListSearchMsg
		// which is sent when search is performed/cleared
		updateMsg := taskdetails.TaskDetailsUpdateMsg{
			SelectedTask: selectedTask,
			SearchQuery:  "", // Search state managed separately
			SearchActive: false,
		}
		return m.taskDetailsComponent.Update(updateMsg)

	case taskdetails.TaskDetailsScrollMsg, taskdetails.TaskDetailsUpdateMsg,
		taskdetails.TaskDetailsResizeMsg:
		return m.taskDetailsComponent.Update(msg)

	case projectdetails.ProjectDetailsScrollMsg, projectdetails.ProjectDetailsUpdateMsg,
		projectdetails.ProjectDetailsResizeMsg:
		return m.projectDetailsComponent.Update(msg)

	// Yank messages - route to active component based on mode
	// Smart routing: check mode once at parent level instead of broadcasting to both children
	case messages.YankIDMsg, messages.YankTitleMsg:
		if m.GetContext().UIState.IsProjectView() {
			return m.projectListComponent.Update(msg)
		}
		return m.taskListComponent.Update(msg)

		// NOTE: ProjectTaskCountsMsg handler removed - ProjectList computes task counts on-demand
	}

	return nil
}

// View renders the main content component using internally owned components
func (m *MainContentModel) View() string {
	// Get components based on current mode from internally owned components
	var leftView, rightView string

	// Query mode from UIState (shared state)
	if m.GetContext().UIState.IsProjectView() {
		// Project mode: left = project list, right = project details
		leftView = m.projectListComponent.View()
		rightView = m.projectDetailsComponent.View()
	} else {
		// Task mode: left = task list, right = task details
		leftView = m.taskListComponent.View()
		rightView = m.taskDetailsComponent.View()
	}

	// Combine horizontally using simple lipgloss layout
	return lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView)
}
