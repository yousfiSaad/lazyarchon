package taskitem

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/v2/internal/archon"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/styling"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/base"
)

const ComponentID = "taskitem"

// Model represents a single task item component
type Model struct {
	base.BaseComponent

	// Task data
	task  archon.Task
	index int // Position in the overall list

	// Display state
	isSelected    bool   // Whether this task is currently selected
	isHighlighted bool   // Whether this task matches search criteria
	searchQuery   string // Current search query for highlighting
}

// Options contains configuration for creating a task item component
type Options struct {
	Task          archon.Task
	Index         int
	Width         int
	IsSelected    bool
	IsHighlighted bool
	SearchQuery   string
	Context       *base.ComponentContext
}

// NewModel creates a new task item component
func NewModel(opts Options) Model {
	baseComponent := base.NewBaseComponent(ComponentID, base.ItemComponent, opts.Context)

	model := Model{
		BaseComponent: baseComponent,
		task:          opts.Task,
		index:         opts.Index,
		isSelected:    opts.IsSelected,
		isHighlighted: opts.IsHighlighted,
		searchQuery:   opts.SearchQuery,
	}
	// Set dimensions using base component
	model.SetDimensions(opts.Width, 1) // Task items are always single line
	return model
}

// Init initializes the task item component
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the task item
//
//nolint:ireturn // Required by Bubble Tea framework - must return tea.Model interface
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TaskItemUpdateMsg:
		if msg.Index == m.index {
			m.task = msg.Task
			m.isSelected = msg.IsSelected
			m.isHighlighted = msg.IsHighlighted
			m.searchQuery = msg.SearchQuery
		}

	case TaskItemResizeMsg:
		if msg.Index == m.index {
			m.SetDimensions(msg.Width, m.GetHeight())
		}
	}

	return m, nil
}

// View renders the task item with selection indicator
func (m *Model) View() string {
	// Use styling if available, otherwise fallback
	if m.GetContext() == nil || m.GetContext().StyleContextProvider == nil {
		return m.renderFallback()
	}

	// Create style context using the provider (similar to existing TaskList pattern)
	styleContext := m.GetContext().StyleContextProvider.CreateStyleContext(m.isSelected).
		WithSearch(m.searchQuery, m.isHighlighted)

	// Build the task line using the existing TaskLineBuilder
	// Width minus 2 for selection indicator space
	contentWidth := m.GetWidth() - 2
	builder := styling.NewTaskLineBuilder(contentWidth, styleContext)

	// Add components in order (following existing pattern from TaskList)
	taskContent := builder.AddPriorityIndicator(m.task).
		AddStatusIndicator(m.task).
		AddTitle(m.task, m.searchQuery, m.isHighlighted).
		AddFeatureTag(m.task).
		Build(m.searchQuery, m.isHighlighted)

	// Add selection indicator (TaskItem owns this responsibility)
	if m.isSelected {
		return styling.SelectionIndicator + taskContent
	}
	return styling.NoSelection + taskContent
}

// renderFallback provides a basic rendering when dependencies are not available
func (m *Model) renderFallback() string {
	status := m.task.Status
	title := m.task.Title

	// Truncate title if needed
	maxTitleWidth := m.GetWidth() - 10 // Leave space for status and padding
	if maxTitleWidth > 0 && len(title) > maxTitleWidth {
		title = title[:maxTitleWidth-3] + "..."
	}

	line := status + " " + title

	// Apply basic selection styling
	if m.isSelected {
		line = "> " + line
	} else {
		line = "  " + line
	}

	// Pad or truncate to fit width
	if len(line) > m.GetWidth() {
		line = line[:m.GetWidth()]
	} else if len(line) < m.GetWidth() {
		line += strings.Repeat(" ", m.GetWidth()-len(line))
	}

	return line
}

// GetTask returns the current task data
func (m *Model) GetTask() archon.Task {
	return m.task
}

// GetIndex returns the task's position in the list
func (m *Model) GetIndex() int {
	return m.index
}

// IsSelected returns whether this task is currently selected
func (m *Model) IsSelected() bool {
	return m.isSelected
}

// IsHighlighted returns whether this task matches search criteria
func (m *Model) IsHighlighted() bool {
	return m.isHighlighted
}

// SetSelected updates the selection state
func (m *Model) SetSelected(selected bool) {
	m.isSelected = selected
}

// SetHighlighted updates the highlight state
func (m *Model) SetHighlighted(highlighted bool, searchQuery string) {
	m.isHighlighted = highlighted
	m.searchQuery = searchQuery
}

// UpdateTask updates the underlying task data
func (m *Model) UpdateTask(task archon.Task) {
	m.task = task
}

// Resize updates the available width for rendering
func (m *Model) Resize(width int) {
	m.SetDimensions(width, m.GetHeight())
}
