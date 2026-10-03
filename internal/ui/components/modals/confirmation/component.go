package confirmation

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/utils/keys"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/base"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/messages"
)

const ComponentID = "confirmation-modal"

// confirmationOptions represents the available choices
var confirmationOptions = []string{"confirm", "cancel"}

// ConfirmationModel represents the confirmation modal component
// Architecture: Follows four-tier state pattern
// - No source data caching (self-contained modal, no ProgramContext dependencies)
// - No display parameters (simple modal with direct user interaction)
// - Owned state only (selection, modal content)
// - No transient feedback (modal lifecycle managed by MainModel)
// - Modal lifecycle managed by BaseModal (active/visible state)
type ConfirmationModel struct {
	base.BaseModal

	// ===================================================================
	// OWNED STATE - Component manages these directly
	// ===================================================================
	selectedIndex int    // Currently selected option (0=confirm, 1=cancel)
	message       string // The confirmation message to display
	confirmText   string // Text for confirm button
	cancelText    string // Text for cancel button
}

// NewModel creates a new confirmation modal component
func NewModel(context *base.ComponentContext) *ConfirmationModel {
	baseModal := base.NewBaseModal(
		ComponentID,
		base.ConfirmationModalComponent,
		context,
	)

	model := &ConfirmationModel{
		BaseModal:     baseModal,
		selectedIndex: 0, // Default to confirm option
		confirmText:   "Yes",
		cancelText:    "No",
	}
	// Set dimensions using base component
	model.SetDimensions(45, 9) // Slightly smaller, more appropriate for confirmations
	return model
}

// Init initializes the confirmation modal component
func (m *ConfirmationModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the confirmation modal component
func (m *ConfirmationModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case ShowConfirmationModalMsg:
		m.SetActive(true)
		m.SetFocus(true)
		m.message = msg.Message
		if msg.ConfirmText != "" {
			m.confirmText = msg.ConfirmText
		}
		if msg.CancelText != "" {
			m.cancelText = msg.CancelText
		}
		m.selectedIndex = 0 // Reset to confirm option
		return m.BroadcastMessage(messages.ModalStateMsg{
			Type:   string(base.ModalTypeConfirmation),
			Active: true,
		})

	case HideConfirmationModalMsg:
		m.SetActive(false)
		m.SetFocus(false)
		return m.BroadcastMessage(messages.ModalStateMsg{
			Type:   string(base.ModalTypeConfirmation),
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

	case ConfirmationModalScrollMsg:
		if !m.IsActive() {
			return nil
		}
		return m.handleScroll(msg)

	default:
		// Base component doesn't have Update method, just return unchanged
		return nil
	}
}

// View renders the confirmation modal
func (m *ConfirmationModel) View() string {
	if !m.IsActive() {
		return ""
	}

	return m.renderModal()
}

// CanFocus implements base.Component interface - confirmation modal can receive focus for keyboard input
func (m *ConfirmationModel) CanFocus() bool {
	return true
}

// SetConfirmationInfo sets the confirmation information for the modal
func (m *ConfirmationModel) SetConfirmationInfo(message, confirmText, cancelText string) {
	m.message = message
	if confirmText != "" {
		m.confirmText = confirmText
	}
	if cancelText != "" {
		m.cancelText = cancelText
	}
}

// handleKeyPress processes keyboard input for the confirmation modal
func (m *ConfirmationModel) handleKeyPress(key tea.KeyMsg) tea.Cmd {
	keyString := key.String()

	switch keyString {
	case keys.KeyQuestion, keys.KeyEscape, keys.KeyQ:
		// Cancel action
		return tea.Batch(
			m.BroadcastMessage(ConfirmationSelectedMsg{
				Confirmed: false,
				Message:   m.message,
			}),
			m.BroadcastMessage(HideConfirmationModalMsg{}),
		)

	case keys.KeyH, keys.KeyArrowLeft:
		m.navigateLeft()
		return nil

	case keys.KeyL, keys.KeyArrowRight:
		m.navigateRight()
		return nil

	case "tab":
		// Tab cycles between options
		m.selectedIndex = (m.selectedIndex + 1) % len(confirmationOptions)
		return nil

	case keys.KeyEnter, keys.KeySpace:
		// Confirm current selection
		confirmed := m.selectedIndex == 0 // 0 = confirm, 1 = cancel
		return tea.Batch(
			m.BroadcastMessage(ConfirmationSelectedMsg{
				Confirmed: confirmed,
				Message:   m.message,
			}),
			m.BroadcastMessage(HideConfirmationModalMsg{}),
		)

	case keys.KeyCtrlC:
		return tea.Quit

	// Direct selection keys
	case "y", "Y":
		// Yes/confirm
		return tea.Batch(
			m.BroadcastMessage(ConfirmationSelectedMsg{
				Confirmed: true,
				Message:   m.message,
			}),
			m.BroadcastMessage(HideConfirmationModalMsg{}),
		)

	case "n", "N":
		// No/cancel
		return tea.Batch(
			m.BroadcastMessage(ConfirmationSelectedMsg{
				Confirmed: false,
				Message:   m.message,
			}),
			m.BroadcastMessage(HideConfirmationModalMsg{}),
		)

	default:
		return nil
	}
}

// handleScroll processes scroll messages
func (m *ConfirmationModel) handleScroll(msg ConfirmationModalScrollMsg) tea.Cmd {
	if msg.Direction > 0 {
		m.navigateRight()
	} else {
		m.navigateLeft()
	}
	return nil
}

// navigateLeft moves selection left (to confirm)
func (m *ConfirmationModel) navigateLeft() {
	m.selectedIndex--
	if m.selectedIndex < 0 {
		m.selectedIndex = 0
	}
}

// navigateRight moves selection right (to cancel)
func (m *ConfirmationModel) navigateRight() {
	m.selectedIndex++
	if m.selectedIndex >= len(confirmationOptions) {
		m.selectedIndex = len(confirmationOptions) - 1
	}
}

// updateDimensions updates the modal dimensions based on screen size
func (m *ConfirmationModel) updateDimensions(screenWidth, screenHeight int) {
	// Modal should be compact but readable - confirmation modals should be small
	width := min(45, screenWidth-6)  // Leave more margin for better centering
	height := min(9, screenHeight-6) // Height to accommodate proper spacing
	m.SetDimensions(width, height)
}

// renderModal renders the complete confirmation modal
func (m *ConfirmationModel) renderModal() string {
	// Create the content
	content := m.renderContent()

	// Use modal dimensions already calculated by parent-child architecture
	// No direct screen access - dimensions flow from parent through ViewWithDimensions
	modalWidth := m.GetWidth()
	modalHeight := m.GetHeight()

	// Create the modal with border (similar to other modal components)
	modal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("51")). // Bright cyan like active panels
		Width(modalWidth).
		Height(modalHeight).
		Padding(1, 2).                           // More horizontal padding for better centering
		Align(lipgloss.Center, lipgloss.Center). // Ensure content is centered within the modal
		Render(content)

	// Parent handles positioning in proper parent-child architecture
	return modal
}

// renderContent renders the modal content
func (m *ConfirmationModel) renderContent() string {
	var content strings.Builder

	// Add some top spacing
	content.WriteString("\n")

	// Title - centered for consistent layout
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("51")).Align(lipgloss.Center)
	title := titleStyle.Render("Confirmation")
	content.WriteString(title)
	content.WriteString("\n\n")

	// Message - centered for better visual appeal
	messageStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Align(lipgloss.Center)
	message := messageStyle.Render(m.message)
	content.WriteString(message)
	content.WriteString("\n\n")

	// Options - centered for better visual appeal
	optionsLine := m.renderOptions()
	centeredOptions := lipgloss.NewStyle().Align(lipgloss.Center).Render(optionsLine)
	content.WriteString(centeredOptions)
	content.WriteString("\n\n")

	// Instructions - centered and more compact
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Align(lipgloss.Center)
	instructions := helpStyle.Render("←/→ • Enter • Y/N • Esc")
	content.WriteString(instructions)

	// Add some bottom spacing
	content.WriteString("\n")

	return content.String()
}

// renderOptions renders the confirmation options
func (m *ConfirmationModel) renderOptions() string {
	var options strings.Builder

	// Render confirm option
	confirmText := m.renderOption(0, m.confirmText)
	options.WriteString(confirmText)

	options.WriteString("      ") // More spacing between options for better visual separation

	// Render cancel option
	cancelText := m.renderOption(1, m.cancelText)
	options.WriteString(cancelText)

	return options.String()
}

// renderOption renders a single option button
func (m *ConfirmationModel) renderOption(index int, text string) string {
	isSelected := index == m.selectedIndex

	// Create the option button with consistent styling
	var buttonStyle lipgloss.Style
	if isSelected {
		// Selected option styling - matches status modal selection style
		buttonStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("15")).
			Padding(0, 2).
			Bold(true)
	} else {
		// Unselected option styling - clean, subtle appearance
		buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(0, 2)
	}

	return buttonStyle.Render(" " + text + " ")
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
