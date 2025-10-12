package projectdetails

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/internal/archon"
	"github.com/yousfisaad/lazyarchon/internal/shared/interfaces"
	"github.com/yousfisaad/lazyarchon/internal/ui/components/base"
	"github.com/yousfisaad/lazyarchon/internal/ui/components/detailspanel"
)

const ComponentID = "projectdetails"

// ProjectDetailsModel represents the project details component
// Architecture:
//   - Embeds base.BaseComponent for component identity and dimensions (single source of truth)
//   - Composes detailspanel.DetailsPanelCore for viewport infrastructure and rendering
//   - Composes ProjectContentGenerator for domain-specific content generation
type ProjectDetailsModel struct {
	base.BaseComponent

	// Viewport infrastructure and rendering (no dimensions stored)
	panelCore detailspanel.DetailsPanelCore

	// Domain-specific: Project data and content generation
	selectedProject  *archon.Project
	contentGenerator ProjectContentGenerator
}

// Options contains configuration options for creating a project details component
type Options struct {
	Width                int
	Height               int
	SelectedProject      *archon.Project
	IsRightPanelActive   bool
	ConfigProvider       interfaces.ConfigProvider
	StyleContextProvider interfaces.StyleContextProvider
	Context              *base.ComponentContext
}

// NewModel creates a new project details component using composition
func NewModel(opts Options) ProjectDetailsModel {
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

	// Create project-specific content generator
	contentGenerator := NewProjectContentGenerator(
		panelCore.GetContentWidth(),
		opts.Context,
	)

	model := ProjectDetailsModel{
		BaseComponent:    baseComponent,
		panelCore:        panelCore,
		selectedProject:  opts.SelectedProject,
		contentGenerator: contentGenerator,
	}

	// Initialize content for both empty and populated states
	model.updateContent()

	return model
}

// Init implements the base.Component interface
func (m ProjectDetailsModel) Init() tea.Cmd {
	return nil
}

// Update implements the base.Component interface
func (m *ProjectDetailsModel) Update(msg tea.Msg) tea.Cmd {
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

	// Handle project details specific messages
	switch msg := msg.(type) {
	case ProjectDetailsUpdateMsg:
		// Update selected project
		m.selectedProject = msg.SelectedProject

		// Update content generator with new project
		m.contentGenerator.SetProject(msg.SelectedProject)

		// Generate new content and update viewport
		m.updateContent()

		// Broadcast scroll position after content update
		return m.broadcastScrollPosition()

	// NOTE: ProjectDetailsSetActiveMsg handler removed - components read active state from UIState directly

	case ProjectDetailsScrollMsg:
		// Delegate scrolling to core
		m.panelCore.HandleScroll(msg.Direction)

		// Broadcast scroll position change
		return m.broadcastScrollPosition()

	case ProjectDetailsResizeMsg:
		// Update core dimensions
		m.panelCore.UpdateDimensions(msg.Width, msg.Height)

		// Update content generator dimensions (providers already set in constructor)
		m.contentGenerator.UpdateDimensions(m.panelCore.GetContentWidth())

		// Regenerate content with new dimensions (for both empty and populated states)
		m.updateContent()
		return nil
	}

	// Forward ONLY mouse messages to viewport for mouse wheel scrolling
	// Do NOT forward keyboard messages - those are handled explicitly via ProjectDetailsScrollMsg
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
func (m ProjectDetailsModel) View() string {
	// Create style context for rendering (component controls styling)
	styleContext := m.contentGenerator.CreateStyleContext(false)

	// Read active state from UIState (single source of truth)
	isActive := m.GetContext().UIState.IsRightPanelActive() && m.GetContext().UIState.IsProjectView()

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
func (m *ProjectDetailsModel) updateContent() {
	if m.selectedProject == nil {
		// Set viewport to empty message - viewport will fill the space
		m.panelCore.SetContent("No project selected")
		return
	}

	// Generate content using the ProjectContentGenerator
	m.contentGenerator.SetProject(m.selectedProject)
	contentLines := m.contentGenerator.GenerateLines()

	// Update viewport with new content via core
	m.panelCore.SetContent(strings.Join(contentLines, "\n"))
}

// broadcastScrollPosition broadcasts the current scroll position to other components
func (m ProjectDetailsModel) broadcastScrollPosition() tea.Cmd {
	position := m.panelCore.GetScrollPosition()
	return m.BroadcastMessage(ProjectDetailsScrollPositionChangedMsg{Position: position})
}

// GetSelectedProject returns the currently selected project
func (m ProjectDetailsModel) GetSelectedProject() *archon.Project {
	return m.selectedProject
}

// GetContentWidth returns the calculated content width from core
func (m ProjectDetailsModel) GetContentWidth() int {
	return m.panelCore.GetContentWidth()
}

// IsScrollable returns whether the content can be scrolled
func (m ProjectDetailsModel) IsScrollable() bool {
	return m.panelCore.IsScrollable()
}

// AtTop returns whether the scroll position is at the top
func (m ProjectDetailsModel) AtTop() bool {
	return m.panelCore.AtTop()
}

// AtBottom returns whether the scroll position is at the bottom
func (m ProjectDetailsModel) AtBottom() bool {
	return m.panelCore.AtBottom()
}

// =============================================================================
// BASE.COMPONENT INTERFACE IMPLEMENTATION
// =============================================================================

// IsActive implements base.Component interface - reads active state from UIState
func (m ProjectDetailsModel) IsActive() bool {
	return m.GetContext().UIState.IsRightPanelActive() && m.GetContext().UIState.IsProjectView()
}

// CanFocus implements base.Component interface - project details can receive focus for scrolling
func (m ProjectDetailsModel) CanFocus() bool {
	return true
}

// SetFocus implements base.Component interface - manages focus state
func (m *ProjectDetailsModel) SetFocus(focused bool) {
	m.BaseComponent.SetFocus(focused)
}

// IsFocused implements base.Component interface - returns current focus state
func (m ProjectDetailsModel) IsFocused() bool {
	return m.BaseComponent.IsFocused()
}
