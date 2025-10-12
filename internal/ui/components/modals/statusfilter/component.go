package statusfilter

import (
	"fmt"
	"maps"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yousfisaad/lazyarchon/v2/internal/archon"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/layout"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/utils/keys"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/utils/view"
	sharedviewport "github.com/yousfisaad/lazyarchon/v2/internal/shared/viewport"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/base"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/messages"
)

const ComponentID = "statusfilter-modal"

// Model represents the status filter modal component
// Architecture: Follows four-tier state pattern
// - No source data caching (receives status selection via ShowStatusFilterModalMsg)
// - No display parameters (manages its own filtering and rendering)
// - Owned state only (status selection, search state)
// - No transient feedback (modal lifecycle managed by MainModel)
// - Modal lifecycle managed by BaseModal (active/visible state)
type Model struct {
	base.BaseModal

	// Core status state
	allStatuses      []string        // All available statuses
	selectedStatuses map[string]bool // Currently selected statuses
	backupStatuses   map[string]bool // Backup for cancel functionality

	// Navigation state
	selectedIndex    int      // Currently highlighted status
	filteredStatuses []string // Statuses after search filtering

	// Search state
	searchMode        bool   // Whether actively typing search
	searchInput       string // Current search input
	searchQuery       string // Committed search query
	matchingIndices   []int  // Indices of statuses matching search
	currentMatchIndex int    // Current position in match list for n/N navigation

	viewport viewport.Model // Viewport for smooth scrolling
}

// NewModel creates a new status filter modal component
func NewModel(context *base.ComponentContext) *Model {
	baseModal := base.NewBaseModal(
		ComponentID,
		base.StatusFilterModalComponent,
		context,
	)

	// Initialize viewport for status list scrolling with reasonable defaults
	// These will be updated properly when the modal is shown and screen size is known
	vp := viewport.New(50, 10) //nolint:varnamelen // vp is idiomatic for viewport
	vp.SetContent("")          // Start with empty content

	// Default status options
	allStatuses := []string{
		archon.TaskStatusTodo,
		archon.TaskStatusDoing,
		archon.TaskStatusReview,
		archon.TaskStatusDone,
	}

	model := &Model{
		BaseModal:         baseModal,
		allStatuses:       allStatuses,
		selectedStatuses:  make(map[string]bool),
		backupStatuses:    make(map[string]bool),
		filteredStatuses:  allStatuses,
		matchingIndices:   []int{},
		selectedIndex:     0,
		searchMode:        false,
		searchInput:       "",
		searchQuery:       "",
		currentMatchIndex: 0,
		viewport:          vp,
	}
	// Set dimensions using base component
	model.SetDimensions(50, 15) // Wide enough for status list, height for status list + search
	return model
}

// CanFocus overrides the base implementation to allow focus
func (m *Model) CanFocus() bool {
	return true
}

// Init initializes the status filter modal component
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the status filter modal component
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if !m.IsActive() {
		if msg, ok := msg.(ShowStatusFilterModalMsg); ok {
			return m.handleShow(msg)
		}
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case HideStatusFilterModalMsg:
		return m.handleHide()

	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)

	case StatusFilterModalShownMsg, StatusFilterModalHiddenMsg:
		// These are notification messages, no action needed
		return nil
	}

	return nil
}

// handleShow shows the status filter modal
func (m *Model) handleShow(msg ShowStatusFilterModalMsg) tea.Cmd {
	m.SetActive(true)

	// Initialize status state
	m.selectedStatuses = make(map[string]bool)
	for status := range msg.CurrentStatuses {
		m.selectedStatuses[status] = msg.CurrentStatuses[status]
	}

	// Create backup for cancel functionality
	m.backupStatuses = make(map[string]bool)
	maps.Copy(m.backupStatuses, m.selectedStatuses)

	// Reset UI state
	m.selectedIndex = 0
	m.searchMode = false
	m.searchInput = ""
	m.searchQuery = ""
	m.filteredStatuses = m.allStatuses
	m.updateViewportContent()

	return func() tea.Msg {
		return messages.ModalStateMsg{
			Type:   string(base.ModalTypeStatusFilter),
			Active: true,
		}
	}
}

// handleHide hides the status filter modal
func (m *Model) handleHide() tea.Cmd {
	m.SetActive(false)
	return func() tea.Msg {
		return messages.ModalStateMsg{
			Type:   string(base.ModalTypeStatusFilter),
			Active: false,
		}
	}
}

// handleWindowSize updates dimensions when window is resized
func (m *Model) handleWindowSize(msg tea.WindowSizeMsg) tea.Cmd {
	// Calculate modal dimensions (centered with some padding)
	modalWidth := min(60, msg.Width-4)
	modalHeight := min(20, msg.Height-4)
	m.SetDimensions(modalWidth, modalHeight)

	// Calculate viewport dimensions using dimension calculator
	// Always reserve scrollbar space to prevent content overflow when scrollbar appears
	calc := layout.NewCalculator(modalWidth, modalHeight, layout.ModalComponent).
		WithScrollbar().     // Reserve space for scrollbar (4 chars)
		WithPadding(2).      // Modal padding
		WithReservedLines(8) // Title, search bar, and buttons

	dims := calc.Calculate()

	// Update viewport dimensions with minimum height check
	viewportHeight := dims.ViewportHeight
	if viewportHeight < 3 {
		viewportHeight = 3
	}
	// Use Content width (accounts for scrollbar)
	m.viewport.Width = dims.Content
	m.viewport.Height = viewportHeight

	m.updateViewportContent()
	return nil
}

// handleKeyPress processes keyboard input
func (m *Model) handleKeyPress(key tea.KeyMsg) tea.Cmd {
	keyString := key.String()

	if m.searchMode {
		return m.handleSearchKeyPress(key)
	}

	switch keyString {
	case keys.KeyEscape, keys.KeyQ:
		// Cancel - restore backup and close
		maps.Copy(m.selectedStatuses, m.backupStatuses)
		return m.handleHide()

	case keys.KeyEnter:
		// Apply selection and close
		return m.handleApply()

	case keys.KeyK, keys.KeyArrowUp:
		if m.selectedIndex > 0 {
			m.selectedIndex--
			m.updateViewportContent()
		}
		return nil

	case keys.KeyJ, keys.KeyArrowDown:
		if m.selectedIndex < len(m.filteredStatuses)-1 {
			m.selectedIndex++
			m.updateViewportContent()
		}
		return nil

	case keys.KeySpace:
		// Toggle status selection
		if m.selectedIndex < len(m.filteredStatuses) {
			status := m.filteredStatuses[m.selectedIndex]
			m.selectedStatuses[status] = !m.selectedStatuses[status]
			m.updateViewportContent()
		}
		return nil

	case keys.KeySlash:
		// Enter search mode
		m.searchMode = true
		m.searchInput = ""
		return nil

	case keys.KeyA:
		// Select all filtered statuses
		for _, status := range m.filteredStatuses {
			m.selectedStatuses[status] = true
		}
		m.updateViewportContent()
		return nil

	case "n":
		// Deselect all filtered statuses
		for _, status := range m.filteredStatuses {
			m.selectedStatuses[status] = false
		}
		m.updateViewportContent()
		return nil
	}

	return nil
}

// handleSearchKeyPress processes keyboard input in search mode
func (m *Model) handleSearchKeyPress(key tea.KeyMsg) tea.Cmd {
	keyString := key.String()

	switch keyString {
	case keys.KeyEscape:
		// Exit search mode
		m.searchMode = false
		m.searchInput = ""
		m.searchQuery = ""
		m.filteredStatuses = m.allStatuses
		m.selectedIndex = 0
		m.updateViewportContent()
		return nil

	case keys.KeyEnter:
		// Commit search query
		m.searchMode = false
		m.searchQuery = strings.TrimSpace(m.searchInput)
		m.applySearchFilter()
		m.updateViewportContent()
		return nil

	case keys.KeyBackspace:
		// Remove character from search input
		if len(m.searchInput) > 0 {
			m.searchInput = m.searchInput[:len(m.searchInput)-1]
			m.applySearchFilter()
			m.updateViewportContent()
		}
		return nil

	default:
		// Add character to search input
		if key.Type == tea.KeyRunes && len(m.searchInput) < 50 {
			m.searchInput += string(key.Runes)
			m.applySearchFilter()
			m.updateViewportContent()
		}
		return nil
	}
}

// handleApply applies the current status selection
func (m *Model) handleApply() tea.Cmd {
	cmd := func() tea.Msg {
		return StatusFilterAppliedMsg{
			SelectedStatuses: maps.Clone(m.selectedStatuses),
		}
	}

	m.SetActive(false)
	return tea.Batch(cmd, func() tea.Msg {
		return messages.ModalStateMsg{
			Type:   string(base.ModalTypeStatusFilter),
			Active: false,
		}
	})
}

// applySearchFilter filters statuses based on search query
func (m *Model) applySearchFilter() {
	if m.searchInput == "" {
		m.filteredStatuses = m.allStatuses
		m.selectedIndex = 0
		return
	}

	query := strings.ToLower(m.searchInput)
	m.filteredStatuses = []string{}
	for _, status := range m.allStatuses {
		if strings.Contains(strings.ToLower(status), query) {
			m.filteredStatuses = append(m.filteredStatuses, status)
		}
	}

	// Reset selection to first item
	m.selectedIndex = 0
}

// updateViewportContent updates the viewport with current status list
func (m *Model) updateViewportContent() {
	lines := make([]string, 0, len(m.filteredStatuses)) // Preallocate for all statuses

	for i, status := range m.filteredStatuses {
		selected := m.selectedStatuses[status]
		cursor := " "
		if i == m.selectedIndex {
			cursor = ">"
		}

		checkbox := "☐"
		if selected {
			checkbox = "☑"
		}

		line := cursor + " " + checkbox + " " + status
		lines = append(lines, line)
	}

	// Use lipgloss.JoinVertical for proper vertical composition
	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	m.viewport.SetContent(content)

	// Ensure selected item is visible
	if m.selectedIndex < len(m.filteredStatuses) {
		// Simple approach: scroll to make the selected item visible
		totalLines := len(m.filteredStatuses)
		if totalLines > m.viewport.Height {
			// Calculate desired viewport top to center the selected item
			desiredTop := m.selectedIndex - m.viewport.Height/2
			if desiredTop < 0 {
				desiredTop = 0
			}
			if desiredTop > totalLines-m.viewport.Height {
				desiredTop = totalLines - m.viewport.Height
			}
			m.viewport.SetYOffset(desiredTop)
		}
	}
}

// View renders the status filter modal
func (m *Model) View() string {
	if !m.IsActive() {
		return ""
	}

	// Create the content
	content := m.renderContent()

	// Calculate modal dimensions
	context := m.GetContext()
	modalWidth := m.GetWidth()
	modalHeight := m.GetHeight()

	if context != nil && context.ProgramContext != nil {
		screenWidth := context.ProgramContext.ScreenWidth
		screenHeight := context.ProgramContext.ScreenHeight
		modalWidth = min(m.GetWidth(), screenWidth-4)
		modalHeight = min(m.GetHeight(), screenHeight-4)
	}

	// Create the modal with border
	modal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("51")). // Bright cyan like other modals
		Width(modalWidth).
		Height(modalHeight).
		Padding(1, 2).
		Align(lipgloss.Center, lipgloss.Top). // Top align for list content
		Render(content)

	// Parent handles positioning - modal just returns its content
	return modal
}

// renderContent renders the modal content
func (m *Model) renderContent() string {
	var content strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("51")).
		Align(lipgloss.Center).
		MarginBottom(1)
	title := titleStyle.Render("Status Filter")
	content.WriteString(title)
	content.WriteString("\n\n")

	// Search bar
	searchPrompt := "Search: "
	switch {
	case m.searchMode:
		searchPrompt += m.searchInput + "█" // Show cursor
	case m.searchQuery != "":
		searchPrompt += m.searchQuery
	default:
		searchPrompt += "(Press / to search)"
	}

	searchStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		MarginBottom(1)
	content.WriteString(searchStyle.Render(searchPrompt))
	content.WriteString("\n\n")

	// Status list viewport with scrollbar
	viewportContent := m.viewport.View()

	// Add scrollbar if content is scrollable
	totalLines := m.viewport.TotalLineCount()
	viewportHeight := m.viewport.Height
	if totalLines > viewportHeight {
		// Generate scrollbar
		scrollbar := view.RenderScrollBarExact(m.viewport.YOffset, totalLines, viewportHeight)

		// Compose content with scrollbar
		// Calculate width: viewport width accounts for modal padding
		viewportContent = sharedviewport.ComposeWithScrollbar(viewportContent, scrollbar, m.viewport.Width+2, 0)
	}

	content.WriteString(viewportContent)
	content.WriteString("\n\n")

	// Summary
	selectedCount := 0
	for _, selected := range m.selectedStatuses {
		if selected {
			selectedCount++
		}
	}
	totalCount := len(m.allStatuses)

	summaryStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Align(lipgloss.Center)
	summary := fmt.Sprintf("Selected: %d/%d statuses", selectedCount, totalCount)
	content.WriteString(summaryStyle.Render(summary))
	content.WriteString("\n\n")

	// Instructions
	var instructions []string
	if m.searchMode {
		instructions = []string{"Enter: commit search", "Esc: cancel search"}
	} else {
		instructions = []string{
			"Space: toggle", "/: search", "a: select all", "n: select none",
			"Enter: apply", "Esc: cancel",
		}
	}

	instructionsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Align(lipgloss.Center)
	content.WriteString(instructionsStyle.Render(strings.Join(instructions, " • ")))

	return content.String()
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
