package taskdetails

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/v2/internal/archon"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/interfaces"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/base"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/detailspanel"
)

const ComponentID = "taskdetails"

// TaskdetailsModel represents the task details component
// Architecture:
//   - Embeds base.BaseComponent for component identity and dimensions (single source of truth)
//   - Composes detailspanel.DetailsPanelCore for viewport infrastructure and rendering
//   - Composes TaskContentGenerator for domain-specific content generation
type TaskdetailsModel struct {
	base.BaseComponent

	// Viewport infrastructure and rendering (no dimensions stored)
	panelCore detailspanel.DetailsPanelCore

	// Domain-specific: Task data and content generation
	selectedTask     *archon.Task
	contentGenerator TaskContentGenerator
}

// Options contains configuration options for creating a task details component
type Options struct {
	Width                int
	Height               int
	SelectedTask         *archon.Task
	SearchQuery          string
	SearchActive         bool
	IsRightPanelActive   bool
	ConfigProvider       interfaces.ConfigProvider
	StyleContextProvider interfaces.StyleContextProvider
	Context              *base.ComponentContext
	ProgramContext       interface{} // Generic context interface for gradual migration
}

// NewModel creates a new task details component using composition
func NewModel(opts Options) TaskdetailsModel {
	// Set default values
	if opts.Width == 0 {
		opts.Width = 40
	}
	if opts.Height == 0 {
		opts.Height = 20
	}

	// Create base component
	baseComponent := base.NewBaseComponent(ComponentID, base.TableComponent, opts.Context)

	// Create shared panel core for viewport infrastructure and rendering
	panelCore := detailspanel.NewCore(detailspanel.CoreOptions{
		Width:  opts.Width,
		Height: opts.Height,
	})

	// Create task-specific content generator
	contentGenerator := NewTaskContentGenerator(
		panelCore.GetContentWidth(),
		opts.Context,
	)

	model := TaskdetailsModel{
		BaseComponent:    baseComponent,
		panelCore:        panelCore,
		selectedTask:     opts.SelectedTask,
		contentGenerator: contentGenerator,
	}

	// Set search parameters if provided
	if opts.SearchQuery != "" || opts.SearchActive {
		model.contentGenerator.SetSearch(opts.SearchQuery, opts.SearchActive)
	}

	// Initialize content for both empty and populated states
	model.updateContent()

	return model
}

// Init implements the base.Component interface
func (m TaskdetailsModel) Init() tea.Cmd {
	return nil
}

// Update implements the base.Component interface Update method
func (m *TaskdetailsModel) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	// Handle window resize - parent calculates exact dimensions for us
	if windowMsg, ok := msg.(tea.WindowSizeMsg); ok {
		// Update core dimensions
		m.panelCore.UpdateDimensions(windowMsg.Width, windowMsg.Height)

		// Update content generator dimensions (providers already set in constructor)
		m.contentGenerator.UpdateDimensions(m.panelCore.GetContentWidth())

		// Regenerate content with new dimensions (for both empty and populated states)
		m.updateContent()
		return nil
	}

	// Handle task details specific messages
	switch msg := msg.(type) {
	case TaskDetailsUpdateMsg:
		// Update selected task
		m.selectedTask = msg.SelectedTask

		// Update content generator with new task and search parameters
		m.contentGenerator.SetTask(msg.SelectedTask)
		m.contentGenerator.SetSearch(msg.SearchQuery, msg.SearchActive)

		// Generate new content and update viewport
		m.updateContent()

		// Broadcast scroll position after content update
		return m.broadcastScrollPosition()

	// NOTE: TaskDetailsSetActiveMsg handler removed - components read active state from UIState directly

	case TaskDetailsScrollMsg:
		// Delegate scrolling to core
		m.panelCore.HandleScroll(msg.Direction)

		// Broadcast scroll position change
		return m.broadcastScrollPosition()

	case TaskDetailsResizeMsg:
		// Update core dimensions
		m.panelCore.UpdateDimensions(msg.Width, msg.Height)

		// Update content generator dimensions (providers already set in constructor)
		m.contentGenerator.UpdateDimensions(m.panelCore.GetContentWidth())

		// Regenerate content with new dimensions (for both empty and populated states)
		m.updateContent()
		return nil
	}

	// Forward ONLY mouse messages to viewport for mouse wheel scrolling
	// Do NOT forward keyboard messages - those are handled explicitly via TaskDetailsScrollMsg
	if _, isMouseMsg := msg.(tea.MouseMsg); isMouseMsg {
		viewport := m.panelCore.GetViewport()
		var cmd tea.Cmd
		*viewport, cmd = viewport.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return tea.Batch(cmds...)
}

// View implements the base.Component interface
func (m TaskdetailsModel) View() string {
	// Create style context for rendering (component controls styling)
	styleContext := m.contentGenerator.CreateStyleContext(false)

	// Read active state from UIState (single source of truth)
	isActive := m.GetContext().UIState.IsRightPanelActive() && m.GetContext().UIState.IsTaskView()

	// Always use unified rendering - viewport fills space for both empty and populated states
	// Content is set via updateContent() for both empty and populated states
	return m.panelCore.RenderPanelWithScrollbar(
		m.GetWidth(),
		m.GetHeight(),
		isActive,
		styleContext,
	)
}

// updateContent generates new content and updates the viewport via core
func (m *TaskdetailsModel) updateContent() {
	if m.selectedTask == nil {
		// Set viewport to empty message - viewport will fill the space
		m.panelCore.SetContent("No task selected")
		return
	}

	// Generate content using the TaskContentGenerator
	m.contentGenerator.SetTask(m.selectedTask)
	contentLines := m.contentGenerator.GenerateLines()

	// Update viewport with new content via core
	m.panelCore.SetContent(strings.Join(contentLines, "\n"))
}

// broadcastScrollPosition broadcasts the current scroll position to other components
func (m TaskdetailsModel) broadcastScrollPosition() tea.Cmd {
	position := m.panelCore.GetScrollPosition()
	return m.BroadcastMessage(TaskDetailsScrollPositionChangedMsg{Position: position})
}

// GetSelectedTask returns the currently selected task
func (m TaskdetailsModel) GetSelectedTask() *archon.Task {
	return m.selectedTask
}

// GetContentWidth returns the calculated content width from core
func (m TaskdetailsModel) GetContentWidth() int {
	return m.panelCore.GetContentWidth()
}

// IsScrollable returns whether the content can be scrolled
func (m TaskdetailsModel) IsScrollable() bool {
	return m.panelCore.IsScrollable()
}

// AtTop returns whether the scroll position is at the top
func (m TaskdetailsModel) AtTop() bool {
	return m.panelCore.AtTop()
}

// AtBottom returns whether the scroll position is at the bottom
func (m TaskdetailsModel) AtBottom() bool {
	return m.panelCore.AtBottom()
}

// =============================================================================
// BASE.COMPONENT INTERFACE IMPLEMENTATION
// =============================================================================

// IsActive implements base.Component interface - reads active state from UIState
func (m TaskdetailsModel) IsActive() bool {
	return m.GetContext().UIState.IsRightPanelActive() && m.GetContext().UIState.IsTaskView()
}

// CanFocus implements base.Component interface - task details can receive focus for scrolling
func (m TaskdetailsModel) CanFocus() bool {
	return true
}

// SetFocus implements base.Component interface - manages focus state
func (m *TaskdetailsModel) SetFocus(focused bool) {
	m.BaseComponent.SetFocus(focused)
}

// IsFocused implements base.Component interface - returns current focus state
func (m TaskdetailsModel) IsFocused() bool {
	return m.BaseComponent.IsFocused()
}
