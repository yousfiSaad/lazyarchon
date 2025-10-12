package status

import (
	"fmt"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yousfisaad/lazyarchon/internal/archon"
	"github.com/yousfisaad/lazyarchon/internal/shared/utils/keys"
	"github.com/yousfisaad/lazyarchon/internal/ui/components/base"
	"github.com/yousfisaad/lazyarchon/internal/ui/messages"
)

const ComponentID = "status-modal"

var statusOptions = []string{archon.TaskStatusTodo, archon.TaskStatusDoing, archon.TaskStatusReview, archon.TaskStatusDone}

// StatusModel represents the status change modal component
// Architecture: Follows four-tier state pattern
// - No source data caching (receives task info via ShowStatusModalMsg)
// - No display parameters (simple selection modal)
// - Owned state only (selection, task context)
// - No transient feedback (modal lifecycle managed by MainModel)
// - Modal lifecycle managed by BaseModal (active/visible state)
type StatusModel struct {
	base.BaseModal

	// ===================================================================
	// OWNED STATE - Component manages these directly
	// ===================================================================
	selectedIndex int    // Currently selected status option (0-3)
	taskID        string // ID of the task being updated (passed via message)
	currentStatus string // Current status of the task (passed via message)
}

// NewModel creates a new status modal component
func NewModel(context *base.ComponentContext) *StatusModel {
	baseModal := base.NewBaseModal(
		ComponentID,
		base.StatusModalComponent,
		context,
	)

	model := &StatusModel{
		BaseModal:     baseModal,
		selectedIndex: 0,
	}
	// Set dimensions using base component
	model.SetDimensions(40, 8)
	return model
}

// CanFocus overrides the base implementation to allow focus
func (m *StatusModel) CanFocus() bool {
	return true
}

// Init initializes the status modal component
func (m *StatusModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the status modal component
func (m *StatusModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case ShowStatusModalMsg:
		m.SetActive(true)
		m.SetFocus(true)
		m.selectedIndex = m.getInitialSelectedIndex()
		return m.BroadcastMessage(messages.ModalStateMsg{
			Type:   string(base.ModalTypeStatus),
			Active: true,
		})

	case HideStatusModalMsg:
		m.SetActive(false)
		m.SetFocus(false)
		return m.BroadcastMessage(messages.ModalStateMsg{
			Type:   string(base.ModalTypeStatus),
			Active: false,
		})

	case tea.WindowSizeMsg:
		m.updateDimensions(msg.Width, msg.Height)
		return nil

	case tea.KeyMsg:
		if !m.IsActive() || !m.IsFocused() {
			return nil
		}
		return m.handleKeyPress(msg)

	case StatusModalScrollMsg:
		if !m.IsActive() {
			return nil
		}
		return m.handleScroll(msg)

	default:
		return nil
	}
}

// View renders the status modal
func (m *StatusModel) View() string {
	if !m.IsActive() {
		return ""
	}

	return m.renderModal()
}

// SetTaskInfo sets the task information for the modal
func (m *StatusModel) SetTaskInfo(taskID, currentStatus string) {
	m.taskID = taskID
	m.currentStatus = currentStatus
	m.selectedIndex = m.getInitialSelectedIndex()
}

// getInitialSelectedIndex returns the index of the current status
func (m *StatusModel) getInitialSelectedIndex() int {
	for i, status := range statusOptions {
		if status == m.currentStatus {
			return i
		}
	}
	return 0 // Default to "todo" if current status not found
}

// handleKeyPress processes keyboard input for the status modal
func (m *StatusModel) handleKeyPress(key tea.KeyMsg) tea.Cmd {
	keyString := key.String()

	switch keyString {
	case keys.KeyQuestion, keys.KeyEscape, keys.KeyQ:
		return m.BroadcastMessage(HideStatusModalMsg{})

	case keys.KeyJ, keys.KeyArrowDown:
		m.navigateDown()
		return nil

	case keys.KeyK, keys.KeyArrowUp:
		m.navigateUp()
		return nil

	case keys.KeyEnter, keys.KeyL:
		// Confirm selection and apply status change
		selectedStatus := statusOptions[m.selectedIndex]
		return tea.Batch(
			m.BroadcastMessage(StatusSelectedMsg{
				Status: selectedStatus,
				TaskID: m.taskID,
			}),
			m.BroadcastMessage(HideStatusModalMsg{}),
		)

	case keys.KeyCtrlC:
		return tea.Quit

	// Number keys for direct selection
	case "1":
		m.selectedIndex = 0
		return nil
	case "2":
		m.selectedIndex = 1
		return nil
	case "3":
		m.selectedIndex = 2
		return nil
	case "4":
		m.selectedIndex = 3
		return nil

	default:
		return nil
	}
}

// handleScroll processes scroll messages
func (m *StatusModel) handleScroll(msg StatusModalScrollMsg) tea.Cmd {
	if msg.Direction > 0 {
		m.navigateDown()
	} else {
		m.navigateUp()
	}
	return nil
}

// navigateDown moves selection down
func (m *StatusModel) navigateDown() {
	m.selectedIndex++
	if m.selectedIndex >= len(statusOptions) {
		m.selectedIndex = len(statusOptions) - 1
	}
}

// navigateUp moves selection up
func (m *StatusModel) navigateUp() {
	m.selectedIndex--
	if m.selectedIndex < 0 {
		m.selectedIndex = 0
	}
}

// updateDimensions updates the modal dimensions based on screen size
func (m *StatusModel) updateDimensions(screenWidth, screenHeight int) {
	// Modal should be centered and reasonably sized
	width := min(40, screenWidth-4)
	height := min(8, screenHeight-4)
	m.SetDimensions(width, height)
}

// renderModal renders the complete status modal
func (m *StatusModel) renderModal() string {
	// Create the content
	content := m.renderContent()

	// Use modal dimensions already calculated by parent-child architecture
	// No direct screen access - dimensions flow from parent through ViewWithDimensions
	modalWidth := m.GetWidth()
	modalHeight := m.GetHeight()

	// Create the modal with border (similar to help modal style)
	modal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("51")). // Bright cyan like active panels
		Width(modalWidth).
		Height(modalHeight).
		Padding(1).
		Render(content)

	// Parent handles positioning in proper parent-child architecture
	return modal
}

// renderContent renders the modal content
func (m *StatusModel) renderContent() string {
	var content strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("51"))
	title := titleStyle.Render("Change Task Status")
	content.WriteString(title)
	content.WriteString("\n\n")

	// Status options
	for i, status := range statusOptions {
		line := m.renderStatusOption(i, status)
		content.WriteString(line)
		content.WriteString("\n")
	}

	// Instructions
	content.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	instructions := helpStyle.Render("↑/↓ navigate • Enter confirm • Esc cancel")
	content.WriteString(instructions)

	return content.String()
}

// renderStatusOption renders a single status option
func (m *StatusModel) renderStatusOption(index int, status string) string {
	// Determine if this option is selected
	isSelected := index == m.selectedIndex
	isCurrent := status == m.currentStatus

	// Create the option text
	var prefix string
	if isSelected {
		prefix = "▶ "
	} else {
		prefix = "  "
	}

	// Add current status indicator
	statusText := m.formatStatusText(status)
	if isCurrent {
		statusText += " (current)"
	}

	// Apply styling
	line := prefix + statusText

	if isSelected {
		selectedStyle := lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("15"))
		line = selectedStyle.Render(line)
	} else if isCurrent {
		currentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
		line = currentStyle.Render(line)
	}

	return line
}

// formatStatusText formats a status string for display
func (m *StatusModel) formatStatusText(status string) string {
	switch status {
	case "todo":
		return "📝 Todo"
	case "doing":
		return "🔄 Doing"
	case "review":
		return "👀 Review"
	case "done":
		return "✅ Done"
	default:
		return fmt.Sprintf("❓ %s", titleCase(status))
	}
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// titleCase capitalizes the first letter of a string (replacement for deprecated strings.Title)
func titleCase(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
