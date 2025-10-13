package tasklist

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/v2/internal/archon"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/layout"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/styling"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/utils/view"
	sharedviewport "github.com/yousfisaad/lazyarchon/v2/internal/shared/viewport"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/base"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/taskitem"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/context"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/messages"
)

const ComponentID = "tasklist"

// fallbackStyleProvider provides minimal styling configuration for tests
type fallbackStyleProvider struct{}

func (f *fallbackStyleProvider) IsPriorityIndicatorsEnabled() bool { return false }
func (f *fallbackStyleProvider) IsFeatureColorsEnabled() bool      { return false }

// TaskListModel represents the task list component
// Architecture: Truly stateless for task data - queries parent on-demand
// - Source data: Tasks, Loading, Error (read from ctx() - ProgramContext)
// - Computed data: Sorted/filtered tasks (queried from parent via getSortedTasks())
// - Owned state: selectedIndex, searchQuery, searchActive (component-local UI state)
// - Active state: Read directly from UIState via GetContext().UIState
//
// Key principle: NO caching of task data - all queries go through parent callbacks
type TaskListModel struct {
	base.BaseComponent

	// ===================================================================
	// COMPONENT DIMENSIONS
	// ===================================================================
	maxLines int // Number of visible lines in viewport

	// ===================================================================
	// OWNED STATE - Component manages these directly
	// ===================================================================
	selectedIndex int    // Currently selected task index
	searchQuery   string // Search query for highlighting
	searchActive  bool   // Whether search highlighting is active

	// ===================================================================
	// UI STATE - Viewport for scrolling
	// ===================================================================
	viewport viewport.Model // Bubble Tea viewport for scrolling

	// Legacy fields (to be removed)
	filterFeature string
	filterStatus  string
}

// ctx returns the program context for easy access to global state
func (m *TaskListModel) ctx() *context.ProgramContext {
	return m.GetContext().ProgramContext
}

// getSortedTasks queries parent for current sorted/filtered task list
// This is the ONLY way to get task data - no caching, always current
func (m *TaskListModel) getSortedTasks() []archon.Task {
	if m.GetContext().GetSortedTasks == nil {
		return nil
	}

	// Query parent for sorted tasks
	interfaceTasks := m.GetContext().GetSortedTasks()

	// Convert []interface{} back to []archon.Task
	tasks := make([]archon.Task, len(interfaceTasks))
	for i, t := range interfaceTasks {
		if task, ok := t.(archon.Task); ok {
			tasks[i] = task
		}
	}

	return tasks
}

// Options contains configuration options for creating a task list component
type Options struct {
	Width         int
	Height        int
	Tasks         []archon.Task
	SelectedIndex int
	SearchQuery   string
	SearchActive  bool
	Context       *base.ComponentContext
}

// NewModel creates a new task list component
func NewModel(opts Options) TaskListModel {
	// Set default values
	if opts.Width == 0 {
		opts.Width = 40
	}
	if opts.Height == 0 {
		opts.Height = 20
	}

	// Create base component
	baseComponent := base.NewBaseComponent(ComponentID, base.TableComponent, opts.Context)

	// Calculate dimensions using dimension calculator
	// Always reserve scrollbar space to prevent content overflow when scrollbar appears
	calc := layout.NewCalculator(opts.Width, opts.Height, layout.PanelComponent).
		WithScrollbar().     // Reserve space for scrollbar (4 chars)
		WithReservedLines(4) // Header (2) + position info (2)
	dims := calc.Calculate()

	// Initialize viewport - use Content width (accounts for scrollbar)
	vp := viewport.New(dims.Content, dims.ViewportHeight) //nolint:varnamelen // vp is idiomatic for viewport

	model := TaskListModel{
		BaseComponent: baseComponent,
		maxLines:      dims.ViewportHeight,
		// Owned state
		selectedIndex: opts.SelectedIndex,
		searchQuery:   opts.SearchQuery,
		searchActive:  opts.SearchActive,
		// UI state
		viewport: vp,
	}
	// Set dimensions using base component
	model.SetDimensions(opts.Width, opts.Height)

	// Update viewport content with initial tasks
	model.updateViewportContent()

	return model
}

// Init implements the base.Component interface
func (m *TaskListModel) Init() tea.Cmd {
	return nil
}

// Update implements the base.Component interface Update method
func (m *TaskListModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)
	case TaskListUpdateMsg, TaskListSelectMsg, TaskListSearchMsg, TaskListFilterMsg:
		return m.handleDataMessages(msg)
	case TaskListScrollMsg:
		return m.handleScrollMessages(msg)
	case messages.YankIDMsg, messages.YankTitleMsg:
		return m.handleYankMessages(msg)
	}
	return nil
}

// handleWindowSize processes window resize events
func (m *TaskListModel) handleWindowSize(msg tea.WindowSizeMsg) tea.Cmd {
	// Store new dimensions using base component
	m.HandleWindowResize(msg)

	// Recalculate dimensions using dimension calculator
	m.updateDimensions()
	m.updateViewportContent()
	return nil
}

// handleDataMessages processes task data update messages
func (m *TaskListModel) handleDataMessages(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case TaskListUpdateMsg:
		// Notification only - task data is queried on-demand via getSortedTasks()
		// Message signals that parent state has changed and viewport should refresh
		// Note: Loading and Error are read from ctx(), not cached
		m.updateViewportContent()
		return nil

	case TaskListSelectMsg:
		// Use helper to ensure viewport regeneration and scroll adjustment
		m.setSelectedIndex(msg.Index)
		return func() tea.Msg { return TaskListSelectionChangedMsg{Index: m.selectedIndex} }

	case TaskListSearchMsg:
		// Update owned state for search highlighting
		m.searchQuery = msg.Query
		m.searchActive = msg.Active
		m.updateViewportContent()
		return nil

	case TaskListFilterMsg:
		// Legacy message - filtering is now handled by MainModel
		// Keep handler for backward compatibility but don't act on it
		m.filterFeature = msg.Feature
		m.filterStatus = msg.Status
		return nil
	}
	return nil
}

// handleScrollMessages processes all scroll direction messages
func (m *TaskListModel) handleScrollMessages(msg TaskListScrollMsg) tea.Cmd {
	// Query current task list for bounds checking
	sortedTasks := m.getSortedTasks()
	taskCount := len(sortedTasks)

	// Calculate new index based on scroll direction, then use helper to update
	switch msg.Direction {
	case ScrollUp:
		m.setSelectedIndex(m.selectedIndex - 1)
	case ScrollDown:
		m.setSelectedIndex(m.selectedIndex + 1)
	case ScrollToTop:
		m.setSelectedIndex(0)
	case ScrollToBottom:
		m.setSelectedIndex(taskCount - 1)
	case ScrollFastUp:
		// Fast scroll up by 4 lines
		m.setSelectedIndex(max(0, m.selectedIndex-4))
	case ScrollFastDown:
		// Fast scroll down by 4 lines
		m.setSelectedIndex(min(taskCount-1, m.selectedIndex+4))
	case ScrollPageUp:
		m.setSelectedIndex(max(0, m.selectedIndex-m.maxLines))
	case ScrollPageDown:
		m.setSelectedIndex(min(taskCount-1, m.selectedIndex+m.maxLines))
	}

	// Note: viewport regeneration and scroll adjustment now handled by setSelectedIndex()

	return func() tea.Msg { return TaskListSelectionChangedMsg{Index: m.selectedIndex} }
}

// handleYankMessages processes ID and title copy operations
// Note: Parent (MainContent) routes yank messages based on mode, so this component
// only receives yank messages when in task mode
func (m *TaskListModel) handleYankMessages(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case messages.YankIDMsg:
		return m.handleYankID()
	case messages.YankTitleMsg:
		return m.handleYankTitle()
	}
	return nil
}

// setSelectedIndex changes the selected task index and ensures UI consistency
// This is the ONLY place where selectedIndex should be modified to maintain invariants:
// - Viewport content is regenerated to show cursor at new position
// - Viewport scroll is adjusted to keep selection visible
func (m *TaskListModel) setSelectedIndex(newIndex int) {
	// Query parent for current task count (bounds check)
	taskCount := len(m.getSortedTasks())
	if newIndex < 0 || newIndex >= taskCount {
		return // Invalid index, don't change
	}

	m.selectedIndex = newIndex
	m.updateViewportContent() // Regenerate content with cursor at new position
	m.followSelection()       // Adjust scroll to keep selection visible
}

// View implements the base.Component interface
func (m *TaskListModel) View() string {
	// Handle special states first
	if specialContent := m.renderSpecialStates(); specialContent != "" {
		return specialContent
	}

	// Render static headers (never scroll)
	effectiveWidth := m.getEffectiveContentWidth()
	headers := styling.RenderLine("Tasks:", effectiveWidth) + "\n" +
		styling.RenderLine("", effectiveWidth)

	// Get viewport content (scrollable tasks)
	viewportContent := m.viewport.View()

	// Add position info if needed
	taskCount := len(m.getSortedTasks())
	if taskCount > m.maxLines {
		positionInfo := m.buildPositionInfoFromViewport()
		viewportContent += "\n\n" + positionInfo
	}

	// Add scrollbar if content is scrollable
	totalLines := m.viewport.TotalLineCount()
	viewportHeight := m.viewport.Height
	if totalLines > viewportHeight {
		// Generate scrollbar
		scrollbar := view.RenderScrollBarExact(m.viewport.YOffset, totalLines, viewportHeight)

		// Compose viewport content with scrollbar
		// Note: Headers are outside viewport, so no header offset needed
		viewportContent = sharedviewport.ComposeWithScrollbar(viewportContent, scrollbar, m.GetWidth(), 0)
	}

	// Combine static headers with scrollable viewport content
	fullContent := headers + "\n" + viewportContent

	// Wrap in panel
	styleContext := m.createStyleContext(false)
	factory := styleContext.Factory()
	// Read active state from UIState (single source of truth)
	isActive := m.GetContext().UIState.IsLeftPanelActive() && m.GetContext().UIState.IsTaskView()
	panelStyle := factory.Panel(m.GetWidth(), m.GetHeight(), isActive)
	return panelStyle.Render(fullContent)
}

// Helper methods

// updateViewportContent renders all tasks to viewport content
func (m *TaskListModel) updateViewportContent() {
	// Query parent for current sorted task list
	sortedTasks := m.getSortedTasks()

	if len(sortedTasks) == 0 {
		m.viewport.SetContent("")
		return
	}

	lines := make([]string, 0, len(sortedTasks)) // Preallocate for all tasks
	effectiveWidth := m.getEffectiveContentWidth()

	// Render all tasks (headers are now rendered statically in View())
	for i, task := range sortedTasks { //nolint:varnamelen // i is idiomatic for loop index
		isSelected := i == m.selectedIndex
		isHighlighted := m.searchActive && m.matchesSearch(task)

		// Create TaskItem for rendering
		item := taskitem.NewModel(taskitem.Options{
			Task:          task,
			Index:         i,
			Width:         effectiveWidth,
			IsSelected:    isSelected,
			IsHighlighted: isHighlighted,
			SearchQuery:   m.searchQuery,
			Context:       m.GetContext(),
		})

		lines = append(lines, item.View())
	}

	// Set viewport content
	m.viewport.SetContent(strings.Join(lines, "\n"))
}

// followSelection updates viewport offset to keep selected item visible
// Uses dynamic scroll margins (25% of viewport height) for better UX with lookahead
func (m *TaskListModel) followSelection() {
	// Query parent for current task count
	if len(m.getSortedTasks()) == 0 {
		return
	}

	// Dynamic scroll margin: 25% of viewport height
	// Provides proportional lookahead/context regardless of terminal size
	scrollMargin := m.viewport.Height / 4
	if scrollMargin < 1 {
		scrollMargin = 1 // Minimum 1 line on very small viewports
	}

	// Calculate line position of selected task in viewport content
	// Headers are now outside viewport, so task index maps directly to line number
	selectedLine := m.selectedIndex

	// Current viewport bounds
	viewportTop := m.viewport.YOffset
	viewportBottom := m.viewport.YOffset + m.viewport.Height - 1

	// Calculate margin boundaries (safe zone)
	marginTop := viewportTop + scrollMargin
	marginBottom := viewportBottom - scrollMargin

	if selectedLine < marginTop {
		// Too close to top margin - scroll up to maintain context above
		newOffset := max(0, selectedLine-scrollMargin)
		m.viewport.SetYOffset(newOffset)
	} else if selectedLine > marginBottom {
		// Too close to bottom margin - scroll down to maintain lookahead below
		newOffset := selectedLine - m.viewport.Height + scrollMargin + 1
		m.viewport.SetYOffset(newOffset)
	}
	// If in safe zone (between margins), don't scroll
}

func (m *TaskListModel) renderSpecialStates() string {
	styleContext := m.createStyleContext(false)
	factory := styleContext.Factory()
	// Read active state from UIState (single source of truth)
	isActive := m.GetContext().UIState.IsLeftPanelActive() && m.GetContext().UIState.IsTaskView()
	listStyle := factory.Panel(m.GetWidth(), m.GetHeight(), isActive)

	// Read state from ProgramContext (single source of truth)
	ctx := m.ctx()

	if ctx.Loading {
		return listStyle.Render("Loading tasks...")
	}

	if ctx.Error != "" {
		return listStyle.Render(fmt.Sprintf("Error: %s\n\nPress 'r' to retry", ctx.Error))
	}

	if len(ctx.Tasks) == 0 {
		return listStyle.Render("No tasks found")
	}

	return ""
}

// getEffectiveContentWidth returns the actual usable content width accounting for scrollbar
func (m *TaskListModel) getEffectiveContentWidth() int {
	// Panel content area: width - borders(2)
	panelContentWidth := m.GetWidth() - 2

	// No reservation here - composeWithScrollBar handles padding and scrollbar
	// TaskItem gets full width, uses (width - 2) for content after selection indicator
	return panelContentWidth
}

// buildPositionInfoFromViewport creates position info based on viewport state
func (m *TaskListModel) buildPositionInfoFromViewport() string {
	// Query parent for current task count
	taskCount := len(m.getSortedTasks())

	// Calculate visible range from viewport offset
	// Headers are now outside viewport, so YOffset maps directly to task index
	firstVisibleTask := max(0, m.viewport.YOffset)
	lastVisibleTask := min(taskCount-1, firstVisibleTask+m.maxLines-1)

	// Calculate percentage
	percentage := ((lastVisibleTask + 1) * 100) / taskCount
	if percentage > 100 {
		percentage = 100
	}

	positionText := fmt.Sprintf("Showing %d-%d of %d tasks (%d%%)",
		firstVisibleTask+1, lastVisibleTask+1, taskCount, percentage)

	// Add selected task position indicator
	selectedPos := m.selectedIndex + 1
	positionText += fmt.Sprintf(" | Task %d selected", selectedPos)

	// Style the position info
	styleContext := m.createStyleContext(false)
	factory := styleContext.Factory()
	styledPositionInfo := factory.Text(styling.CurrentTheme.MutedColor).Render(positionText)

	return styling.RenderLine(styledPositionInfo, m.getEffectiveContentWidth())
}

func (m *TaskListModel) createStyleContext(selected bool) *styling.StyleContext {
	if m.GetContext() != nil && m.GetContext().StyleContextProvider != nil {
		return m.GetContext().StyleContextProvider.CreateStyleContext(selected).
			WithSearch(m.searchQuery, m.searchActive)
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
	return styling.NewStyleContext(theme, styleProvider).
		WithSearch(m.searchQuery, m.searchActive)
}

// GetSelectedTask returns the currently selected task
func (m *TaskListModel) GetSelectedTask() *archon.Task {
	// Query parent for current sorted tasks
	sortedTasks := m.getSortedTasks()
	if m.selectedIndex >= 0 && m.selectedIndex < len(sortedTasks) {
		return &sortedTasks[m.selectedIndex]
	}
	return nil
}

// GetSelectedIndex returns the currently selected index
func (m *TaskListModel) GetSelectedIndex() int {
	return m.selectedIndex
}

// GetTaskCount returns the total number of tasks
func (m *TaskListModel) GetTaskCount() int {
	return len(m.getSortedTasks())
}

// Utility functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// matchesSearch checks if a task matches the current search query
func (m *TaskListModel) matchesSearch(task archon.Task) bool {
	if m.searchQuery == "" {
		return false
	}

	query := strings.ToLower(m.searchQuery)
	return strings.Contains(strings.ToLower(task.Title), query) ||
		strings.Contains(strings.ToLower(task.Status), query) ||
		(task.Feature != nil && strings.Contains(strings.ToLower(*task.Feature), query)) ||
		strings.Contains(strings.ToLower(task.ID), query)
}

// updateDimensions recalculates all dimensions using the dimension calculator
// This ensures consistent calculations across resize events
func (m *TaskListModel) updateDimensions() {
	// Always reserve scrollbar space to prevent content overflow when scrollbar appears
	calc := layout.NewCalculator(m.GetWidth(), m.GetHeight(), layout.PanelComponent).
		WithScrollbar().     // Reserve space for scrollbar (4 chars)
		WithReservedLines(4) // Header (2) + position info (2)
	dims := calc.Calculate()

	// Update stored dimensions
	m.maxLines = dims.ViewportHeight

	// Update viewport dimensions - use Content width (accounts for scrollbar)
	m.viewport.Width = dims.Content
	m.viewport.Height = dims.ViewportHeight
}

// =============================================================================
// BASE.COMPONENT INTERFACE IMPLEMENTATION
// =============================================================================

// IsActive implements base.Component interface - returns true if component is active
// Reads from UIState (single source of truth)
func (m *TaskListModel) IsActive() bool {
	return m.GetContext().UIState.IsLeftPanelActive() && m.GetContext().UIState.IsTaskView()
}

// CanFocus implements base.Component interface - task list can receive focus for navigation
func (m *TaskListModel) CanFocus() bool {
	return true
}

// SetFocus implements base.Component interface - manages focus state
func (m *TaskListModel) SetFocus(focused bool) {
	m.BaseComponent.SetFocus(focused)
}

// IsFocused implements base.Component interface - returns current focus state
func (m *TaskListModel) IsFocused() bool {
	return m.BaseComponent.IsFocused()
}

// =============================================================================
// YANK (COPY) OPERATIONS
// =============================================================================

// handleYankID copies the selected task ID to clipboard
func (m *TaskListModel) handleYankID() tea.Cmd {
	task := m.GetSelectedTask()
	if task == nil {
		return func() tea.Msg {
			return messages.StatusFeedbackMsg{Message: "No task selected"}
		}
	}

	err := clipboard.WriteAll(task.ID)
	if err != nil {
		return func() tea.Msg {
			return messages.StatusFeedbackMsg{Message: "Failed to copy task ID"}
		}
	}

	return func() tea.Msg {
		return messages.StatusFeedbackMsg{
			Message: fmt.Sprintf("Copied task ID: %s", task.ID),
		}
	}
}

// handleYankTitle copies the selected task title to clipboard
func (m *TaskListModel) handleYankTitle() tea.Cmd {
	task := m.GetSelectedTask()
	if task == nil {
		return func() tea.Msg {
			return messages.StatusFeedbackMsg{Message: "No task selected"}
		}
	}

	err := clipboard.WriteAll(task.Title)
	if err != nil {
		return func() tea.Msg {
			return messages.StatusFeedbackMsg{Message: "Failed to copy task title"}
		}
	}

	return func() tea.Msg {
		return messages.StatusFeedbackMsg{
			Message: fmt.Sprintf("Copied task title: %s", task.Title),
		}
	}
}
