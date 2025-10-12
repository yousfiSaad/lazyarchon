package projectlist

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/internal/archon"
	"github.com/yousfisaad/lazyarchon/internal/shared/styling"
	"github.com/yousfisaad/lazyarchon/internal/ui/components/base"
	"github.com/yousfisaad/lazyarchon/internal/ui/context"
	"github.com/yousfisaad/lazyarchon/internal/ui/messages"
)

const ComponentID = "projectlist"

// ProjectList component now focuses solely on project selection
// Help functionality has been moved to the global help modal

// ProjectListModel represents the project list component
// Architecture: Follows four-tier state pattern (Display Parameters eliminated)
// - Source data: Projects (read from ProgramContext via ctx())
// - UI Presentation State: Read from UIState (view mode, active panel)
// - Owned state: selectedIndex only
// - Transient feedback: None (feedback handled by StatusBar)
//
// Task counts computed on-demand via ctx().GetTaskCountForProject()
type ProjectListModel struct {
	base.BaseComponent

	// ===================================================================
	// OWNED STATE - Component manages these directly
	// ===================================================================
	selectedIndex int // Currently selected project index

	// NOTE: Display parameters removed - compute on-demand from context:
	// - displayProjectTaskCounts → ctx().GetTaskCountForProject(projectID)
	// - displayTotalTaskCount → ctx().GetTotalTaskCount()
}

// ctx returns the program context for easy access to global state
func (m *ProjectListModel) ctx() *context.ProgramContext {
	return m.GetContext().ProgramContext
}

// fallbackStyleProvider provides minimal styling configuration for tests
type fallbackStyleProvider struct{}

func (f *fallbackStyleProvider) IsPriorityIndicatorsEnabled() bool { return false }
func (f *fallbackStyleProvider) IsFeatureColorsEnabled() bool      { return false }

// Options contains configuration options for creating a project list component
type Options struct {
	Width         int
	Height        int
	Projects      []archon.Project
	SelectedIndex int
	Loading       bool
	ErrorMessage  string
	Context       *base.ComponentContext
}

// NewModel creates a new project list component
func NewModel(opts Options) ProjectListModel {
	// Set default values
	if opts.Width == 0 {
		opts.Width = 40
	}
	if opts.Height == 0 {
		opts.Height = 20
	}

	// Create base component
	baseComponent := base.NewBaseComponent(ComponentID, base.TableComponent, opts.Context)

	model := ProjectListModel{
		BaseComponent: baseComponent,
		// Owned state
		selectedIndex: opts.SelectedIndex,
		// NOTE: Display parameters removed - compute on-demand from context
	}
	// Set dimensions using base component
	model.SetDimensions(opts.Width, opts.Height)
	return model
}

// Init implements the base.Component interface
func (m *ProjectListModel) Init() tea.Cmd {
	return nil
}

// View implements the base.Component interface
func (m *ProjectListModel) View() string {
	// Handle special states first (reads from ctx())
	if specialContent := m.renderSpecialStates(); specialContent != "" {
		return specialContent
	}

	var lines []string
	lines = append(lines, "Projects:")
	lines = append(lines, "")

	// Add projects with task counts (reads from ctx())
	projects := m.ctx().Projects
	for i, project := range projects {
		line := m.renderProjectLine(project, i)
		lines = append(lines, line)
	}

	// Add "All Tasks" option at the end
	allTasksLine := m.renderAllTasksLine()
	lines = append(lines, "")
	lines = append(lines, allTasksLine)

	// Create final panel
	content := strings.Join(lines, "\n")
	styleContext := m.createStyleContext(false)
	factory := styleContext.Factory()
	// Read active state from UIState (single source of truth)
	isActive := m.GetContext().UIState.IsLeftPanelActive() && m.GetContext().UIState.IsProjectView()
	listStyle := factory.Panel(m.GetWidth(), m.GetHeight(), isActive)

	return listStyle.Render(content)
}

// renderProjectModeHelp method removed - help functionality moved to global help modal

// Helper methods

func (m *ProjectListModel) renderSpecialStates() string {
	styleContext := m.createStyleContext(false)
	factory := styleContext.Factory()
	// Read active state from UIState (single source of truth)
	isActive := m.GetContext().UIState.IsLeftPanelActive() && m.GetContext().UIState.IsProjectView()
	listStyle := factory.Panel(m.GetWidth(), m.GetHeight(), isActive)

	// Read state from ProgramContext (single source of truth)
	ctx := m.ctx()

	if ctx.Loading {
		return listStyle.Render("Loading projects...")
	}

	if ctx.Error != "" {
		return listStyle.Render(fmt.Sprintf("Error: %s\n\nPress 'r' to retry", ctx.Error))
	}

	if len(ctx.Projects) == 0 {
		return listStyle.Render("No projects found")
	}

	return ""
}

func (m *ProjectListModel) renderProjectLine(project archon.Project, index int) string {
	// Compute task count on-demand from ProgramContext
	taskCount := m.ctx().GetTaskCountForProject(project.ID)

	line := fmt.Sprintf("%s (%d)", project.Title, taskCount)
	if len(line) > m.GetWidth()-8 {
		line = line[:m.GetWidth()-11] + "..."
	}

	// Apply selection styling
	isSelected := index == m.selectedIndex
	itemStyleContext := m.createStyleContext(isSelected)
	itemFactory := itemStyleContext.Factory()
	style := itemFactory.ProjectItem(isSelected, false)

	if isSelected {
		line = styling.SelectionIndicator + line
	} else {
		line = styling.NoSelection + line
	}

	return style.Render(line)
}

func (m *ProjectListModel) renderAllTasksLine() string {
	// Compute total task count on-demand from ProgramContext
	totalCount := m.ctx().GetTotalTaskCount()

	// Read project count from ctx()
	projectCount := len(m.ctx().Projects)
	allTasksLine := fmt.Sprintf("[All Tasks] (%d)", totalCount)
	isAllTasksSelected := m.selectedIndex == projectCount

	allTasksStyleContext := m.createStyleContext(isAllTasksSelected)
	allTasksFactory := allTasksStyleContext.Factory()
	allTasksStyle := allTasksFactory.ProjectItem(isAllTasksSelected, true)

	if isAllTasksSelected {
		allTasksLine = styling.SelectionIndicator + allTasksLine
	} else {
		allTasksLine = styling.NoSelection + allTasksLine
	}

	return allTasksStyle.Render(allTasksLine)
}

func (m *ProjectListModel) createStyleContext(selected bool) *styling.StyleContext {
	if m.GetContext() != nil && m.GetContext().StyleContextProvider != nil {
		return m.GetContext().StyleContextProvider.CreateStyleContext(selected)
	}
	// Fallback to a basic style context with minimal theme
	theme := &styling.ThemeAdapter{
		TodoColor:   "yellow",
		DoingColor:  "blue",
		ReviewColor: "orange",
		DoneColor:   "green",
		HeaderColor: "cyan",
		MutedColor:  "gray",
		Name:        "fallback",
	}
	// Create a minimal style provider for the fallback
	styleProvider := &fallbackStyleProvider{}
	return styling.NewStyleContext(theme, styleProvider)
}

// GetSelectedProject returns the currently selected project
func (m *ProjectListModel) GetSelectedProject() *archon.Project {
	// Read projects from ProgramContext (single source of truth)
	projects := m.ctx().Projects
	if m.selectedIndex >= 0 && m.selectedIndex < len(projects) {
		return &projects[m.selectedIndex]
	}
	return nil
}

// GetSelectedIndex returns the currently selected index
func (m *ProjectListModel) GetSelectedIndex() int {
	return m.selectedIndex
}

// IsAllTasksSelected returns true if "All Tasks" option is selected
func (m *ProjectListModel) IsAllTasksSelected() bool {
	// Read project count from ProgramContext
	return m.selectedIndex == len(m.ctx().Projects)
}

// GetProjectCount returns the total number of projects
func (m *ProjectListModel) GetProjectCount() int {
	// Read project count from ProgramContext (single source of truth)
	return len(m.ctx().Projects)
}

// IsActive returns whether the project list is currently active (project mode)
// Reads from UIState (single source of truth)
func (m *ProjectListModel) IsActive() bool {
	return m.GetContext().UIState.IsLeftPanelActive() && m.GetContext().UIState.IsProjectView()
}

// =============================================================================
// BASE.COMPONENT INTERFACE IMPLEMENTATION
// =============================================================================

// Update implements the base.Component interface Update method
func (m *ProjectListModel) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	// Handle window resize
	if windowMsg, ok := msg.(tea.WindowSizeMsg); ok {
		m.HandleWindowResize(windowMsg)
		// When calculating content for scrolling, account for Panel borders:
		// contentArea := m.GetHeight() - base.PanelBorderLines
		return nil
	}

	// Handle project list specific messages
	switch msg := msg.(type) {
	case ProjectListUpdateMsg:
		// No longer caching projects, loading, or error
		// These are read directly from ProgramContext via ctx()
		return nil

	case ProjectListSelectMsg:
		// Read project count from ProgramContext
		projectCount := len(m.ctx().Projects)
		if msg.Index >= 0 && msg.Index <= projectCount { // Allow selecting "All Tasks" option
			m.selectedIndex = msg.Index
		}
		return func() tea.Msg { return ProjectListSelectionChangedMsg{Index: m.selectedIndex} }

	case ProjectListScrollMsg:
		// Read project count from ProgramContext
		projectCount := len(m.ctx().Projects)
		switch msg.Direction {
		case ScrollUp:
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}
		case ScrollDown:
			if m.selectedIndex < projectCount { // Allow selecting "All Tasks" option
				m.selectedIndex++
			}
		case ScrollToTop:
			m.selectedIndex = 0
		case ScrollToBottom:
			m.selectedIndex = projectCount // "All Tasks" option
		}
		return func() tea.Msg { return ProjectListSelectionChangedMsg{Index: m.selectedIndex} }

	// NOTE: ProjectListSetActiveMsg handler removed - components read active state from UIState directly

	case ProjectListSelectionQueryMsg:
		// Return current selection index as a response message
		return func() tea.Msg {
			return ProjectListSelectionResponseMsg{Index: m.selectedIndex}
		}

	case messages.YankIDMsg:
		// Note: Parent (MainContent) routes yank messages based on mode, so this component
		// only receives yank messages when in project mode
		return m.handleYankID()

	case messages.YankTitleMsg:
		// Note: Parent (MainContent) routes yank messages based on mode, so this component
		// only receives yank messages when in project mode
		return m.handleYankTitle()

	// NOTE: ProjectTaskCountsMsg handler removed - task counts computed on-demand from context

		// Note: ProjectListConfirmSelectionMsg is outgoing only - sent by this component
		// It's not handled here, but by parent components (app.go)

		// ProjectListSetPanelMsg removed - help functionality moved to global help modal
	}

	return tea.Batch(cmds...)
}

// CanFocus implements base.Component interface - project list can receive focus for navigation
func (m *ProjectListModel) CanFocus() bool {
	return true
}

// SetFocus implements base.Component interface - manages focus state
func (m *ProjectListModel) SetFocus(focused bool) {
	m.BaseComponent.SetFocus(focused)
}

// IsFocused implements base.Component interface - returns current focus state
func (m *ProjectListModel) IsFocused() bool {
	return m.BaseComponent.IsFocused()
}

// =============================================================================
// YANK (COPY) OPERATIONS
// =============================================================================

// handleYankID copies the selected project ID to clipboard
func (m *ProjectListModel) handleYankID() tea.Cmd {
	project := m.GetSelectedProject()
	if project == nil {
		return func() tea.Msg {
			return messages.StatusFeedbackMsg{Message: "No project selected"}
		}
	}

	err := clipboard.WriteAll(project.ID)
	if err != nil {
		return func() tea.Msg {
			return messages.StatusFeedbackMsg{Message: "Failed to copy project ID"}
		}
	}

	return func() tea.Msg {
		return messages.StatusFeedbackMsg{
			Message: fmt.Sprintf("Copied project ID: %s", project.ID),
		}
	}
}

// handleYankTitle copies the selected project title to clipboard
func (m *ProjectListModel) handleYankTitle() tea.Cmd {
	project := m.GetSelectedProject()
	if project == nil {
		return func() tea.Msg {
			return messages.StatusFeedbackMsg{Message: "No project selected"}
		}
	}

	err := clipboard.WriteAll(project.Title)
	if err != nil {
		return func() tea.Msg {
			return messages.StatusFeedbackMsg{Message: "Failed to copy project title"}
		}
	}

	return func() tea.Msg {
		return messages.StatusFeedbackMsg{
			Message: fmt.Sprintf("Copied project title: %s", project.Title),
		}
	}
}
