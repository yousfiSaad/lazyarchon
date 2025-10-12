package taskedit

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yousfisaad/lazyarchon/v2/internal/archon"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/styling"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/utils/keys"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/base"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/messages"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const ComponentID = "task-edit-modal"

var statusOptions = []string{archon.TaskStatusTodo, archon.TaskStatusDoing, archon.TaskStatusReview, archon.TaskStatusDone}

// TaskEditModel represents the task properties edit modal component
// Architecture: Follows four-tier state pattern
// - No source data caching (receives task/feature data via ShowTaskEditModalMsg)
// - No display parameters (manages its own form state during edit session)
// - Owned state only (form fields, edit modes, working values)
// - No transient feedback (modal lifecycle managed by MainModel)
// - Modal lifecycle managed by BaseModal (active/visible state)
type TaskEditModel struct {
	base.BaseModal

	// ===================================================================
	// OWNED STATE - Component manages these directly
	// ===================================================================

	// Task context (passed via message for edit session)
	taskID string // ID of task being edited

	// Multi-field form state
	activeField FieldType // Currently focused field (0=status, 1=priority, 2=feature)

	// Field values (working state - what user is editing)
	statusValue   string // Current status selection
	priorityValue int    // Current priority value
	featureValue  string // Current feature assignment

	// Original values (for change detection)
	originalStatus   string
	originalPriority int
	originalFeature  string

	// Status field state
	statusIndex int // Index in statusOptions array

	// Priority field state
	priorityEditMode bool   // true when typing specific number
	priorityInput    string // Text being typed for priority

	// Feature field state
	availableFeatures    []string // Available features to choose from (passed via message)
	selectedFeatureIndex int      // Currently highlighted feature in selection mode
	featureSelectionMode bool     // true when viewport is expanded for selection
	isCreatingNew        bool     // true when in text input mode for new feature
	newFeatureName       string   // Text being typed for new feature
}

// NewModel creates a new task properties edit modal component
func NewModel(context *base.ComponentContext) *TaskEditModel {
	baseModal := base.NewBaseModal(
		ComponentID,
		base.TaskEditModalComponent,
		context,
	)

	model := &TaskEditModel{
		BaseModal:            baseModal,
		activeField:          FieldStatus, // Start on status field by default
		statusIndex:          0,
		priorityEditMode:     false,
		priorityInput:        "",
		selectedFeatureIndex: 0,
		featureSelectionMode: false,
		isCreatingNew:        false,
		newFeatureName:       "",
	}
	// Set dimensions using base component
	model.SetDimensions(60, 16) // Wider to accommodate all fields, taller to show all fields
	return model
}

// CanFocus overrides the base implementation to allow focus
func (m *TaskEditModel) CanFocus() bool {
	return true
}

// Init initializes the task edit modal component
func (m *TaskEditModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the task edit modal component
func (m *TaskEditModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case ShowTaskEditModalMsg:
		m.SetActive(true)
		m.SetFocus(true)

		// Set task info
		m.taskID = msg.TaskID
		m.activeField = msg.FocusField // Start on specified field

		// Initialize status field
		m.statusValue = msg.CurrentStatus
		m.originalStatus = msg.CurrentStatus
		m.statusIndex = m.getStatusIndex(msg.CurrentStatus)

		// Initialize priority field
		m.priorityValue = msg.CurrentPriority
		m.originalPriority = msg.CurrentPriority
		m.priorityEditMode = false
		m.priorityInput = ""

		// Initialize feature field
		m.featureValue = msg.CurrentFeature
		m.originalFeature = msg.CurrentFeature
		m.availableFeatures = msg.AvailableFeatures
		m.featureSelectionMode = false
		m.isCreatingNew = false
		m.newFeatureName = ""

		// Pre-select the current feature if it exists in available features
		if m.featureValue != "" {
			if index := m.findFeatureIndex(m.featureValue, m.availableFeatures); index != -1 {
				m.selectedFeatureIndex = index
			} else {
				m.selectedFeatureIndex = 0 // Default to first item if current feature not found
			}
		} else {
			m.selectedFeatureIndex = 0 // Default to first item if no current feature
		}

		return m.BroadcastMessage(messages.ModalStateMsg{
			Type:   string(base.ModalTypeTaskEdit),
			Active: true,
		})

	case HideTaskEditModalMsg:
		m.SetActive(false)
		m.SetFocus(false)
		m.priorityEditMode = false
		m.priorityInput = ""
		m.featureSelectionMode = false
		m.isCreatingNew = false
		m.newFeatureName = ""
		return m.BroadcastMessage(messages.ModalStateMsg{
			Type:   string(base.ModalTypeTaskEdit),
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

	case TaskEditModalScrollMsg:
		if !m.IsActive() {
			return nil
		}
		return m.handleScroll(msg)

	default:
		return nil
	}
}

// View renders the task edit modal
func (m *TaskEditModel) View() string {
	if !m.IsActive() {
		return ""
	}

	return m.renderModal()
}

// handleKeyPress processes keyboard input for the task properties modal
func (m *TaskEditModel) handleKeyPress(key tea.KeyMsg) tea.Cmd {
	keyString := key.String()

	// Check if we're in a special mode that needs priority routing
	// These modes intercept keys before global handlers
	if m.priorityEditMode || m.isCreatingNew || m.featureSelectionMode {
		// Route directly to field handler for special modes
		switch m.activeField {
		case FieldPriority:
			return m.handlePriorityField(keyString)
		case FieldFeature:
			return m.handleFeatureField(keyString)
		default:
			return nil
		}
	}

	// Global keys that work when not in special mode
	switch keyString {
	case keys.KeyEscape, keys.KeyQ:
		// Cancel and close modal without saving
		return m.BroadcastMessage(HideTaskEditModalMsg{})

	case keys.KeyCtrlC:
		return tea.Quit

	case keys.KeyJ, keys.KeyArrowDown:
		// Navigate to next field (vim-style vertical navigation)
		m.activeField = (m.activeField + 1) % 3
		// Reset field-specific modes when changing fields
		m.priorityEditMode = false
		m.isCreatingNew = false
		m.featureSelectionMode = false
		return nil

	case keys.KeyK, keys.KeyArrowUp:
		// Navigate to previous field (vim-style vertical navigation)
		m.activeField = (m.activeField - 1 + 3) % 3
		// Reset field-specific modes when changing fields
		m.priorityEditMode = false
		m.isCreatingNew = false
		m.featureSelectionMode = false
		return nil
	}

	// Route to active field handler
	switch m.activeField {
	case FieldStatus:
		return m.handleStatusField(keyString)
	case FieldPriority:
		return m.handlePriorityField(keyString)
	case FieldFeature:
		return m.handleFeatureField(keyString)
	default:
		return nil
	}
}

// =============================================================================
// FIELD HANDLERS - Handle input for each field type
// =============================================================================

// handleStatusField handles input when status field is focused
func (m *TaskEditModel) handleStatusField(keyString string) tea.Cmd {
	switch keyString {
	case keys.KeyL, keys.KeyArrowRight:
		// Navigate to next status (vim-style horizontal navigation →)
		m.statusIndex = (m.statusIndex + 1) % len(statusOptions)
		m.statusValue = statusOptions[m.statusIndex]
		return nil

	case keys.KeyH, keys.KeyArrowLeft:
		// Navigate to previous status (vim-style horizontal navigation ←)
		m.statusIndex = (m.statusIndex - 1 + len(statusOptions)) % len(statusOptions)
		m.statusValue = statusOptions[m.statusIndex]
		return nil

	case "1":
		// Direct selection: Todo
		m.statusIndex = 0
		m.statusValue = statusOptions[0]
		return nil

	case "2":
		// Direct selection: Doing
		m.statusIndex = 1
		m.statusValue = statusOptions[1]
		return nil

	case "3":
		// Direct selection: Review
		m.statusIndex = 2
		m.statusValue = statusOptions[2]
		return nil

	case "4":
		// Direct selection: Done
		m.statusIndex = 3
		m.statusValue = statusOptions[3]
		return nil

	case keys.KeyEnter, keys.KeySpace:
		// Save changes and close
		return m.saveChanges()

	default:
		return nil
	}
}

// handlePriorityField handles input when priority field is focused
func (m *TaskEditModel) handlePriorityField(keyString string) tea.Cmd {
	// If in text input mode, handle numeric input
	if m.priorityEditMode {
		return m.handlePriorityTextInput(keyString)
	}

	// Navigation mode - adjust priority with h/l (vim-style horizontal navigation)
	switch keyString {
	case keys.KeyH, keys.KeyArrowLeft:
		// Decrease priority by 1 (vim-style horizontal navigation ←)
		m.priorityValue = max(0, m.priorityValue-1)
		return nil

	case keys.KeyL, keys.KeyArrowRight:
		// Increase priority by 1 (vim-style horizontal navigation →)
		m.priorityValue = min(999, m.priorityValue+1)
		return nil

	case keys.KeyHCap: // Shift+H
		// Fast decrease by 10
		m.priorityValue = max(0, m.priorityValue-10)
		return nil

	case keys.KeyLCap: // Shift+L
		// Fast increase by 10
		m.priorityValue = min(999, m.priorityValue+10)
		return nil

	case keys.KeyEnter:
		// Switch to text input mode for typing specific value
		m.priorityEditMode = true
		m.priorityInput = strconv.Itoa(m.priorityValue)
		return nil

	case keys.KeySpace:
		// Save changes and close
		return m.saveChanges()

	case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
		// Start text input mode with this digit
		m.priorityEditMode = true
		m.priorityInput = keyString
		return nil

	default:
		return nil
	}
}

// handlePriorityTextInput handles numeric input when editing priority value
func (m *TaskEditModel) handlePriorityTextInput(keyString string) tea.Cmd {
	switch keyString {
	case keys.KeyEscape:
		// Cancel text input mode
		m.priorityEditMode = false
		m.priorityInput = ""
		return nil

	case keys.KeyEnter:
		// Confirm entered value
		if value, err := strconv.Atoi(m.priorityInput); err == nil {
			// Clamp to valid range
			m.priorityValue = max(0, min(999, value))
		}
		m.priorityEditMode = false
		m.priorityInput = ""
		return nil

	case keys.KeyBackspace:
		// Remove last digit
		if len(m.priorityInput) > 0 {
			m.priorityInput = m.priorityInput[:len(m.priorityInput)-1]
		}
		return nil

	case keys.KeyCtrlU:
		// Clear entire input
		m.priorityInput = ""
		return nil

	case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
		// Add digit (max 3 digits for 999)
		if len(m.priorityInput) < 3 {
			m.priorityInput += keyString
		}
		return nil

	default:
		return nil
	}
}

// handleFeatureField handles input when feature field is focused
func (m *TaskEditModel) handleFeatureField(keyString string) tea.Cmd {
	// If creating new feature, handle text input
	if m.isCreatingNew {
		return m.handleFeatureTextInput(keyString)
	}

	// If in selection mode, handle feature list navigation
	if m.featureSelectionMode {
		return m.handleFeatureSelectionMode(keyString)
	}

	// Normal mode - feature field focused but viewport not expanded
	switch keyString {
	case keys.KeyL, keys.KeyEnter:
		// Enter selection mode - expand viewport
		m.featureSelectionMode = true
		// Set initial selection to current feature if it exists
		if m.featureValue != "" {
			if index := m.findFeatureIndex(m.featureValue, m.availableFeatures); index != -1 {
				m.selectedFeatureIndex = index
			}
		}
		return nil

	case keys.KeyN:
		// Create new feature directly
		m.isCreatingNew = true
		m.newFeatureName = ""
		return nil

	case keys.KeySpace:
		// Save changes and close modal
		return m.saveChanges()

	default:
		return nil
	}
}

// handleFeatureTextInput handles input when creating a new feature
//
//nolint:gocyclo // Complexity unavoidable - UI handler with multiple input modes and validation
func (m *TaskEditModel) handleFeatureTextInput(keyString string) tea.Cmd {
	switch keyString {
	case keys.KeyEscape:
		// Cancel new feature creation
		m.isCreatingNew = false
		m.newFeatureName = ""
		return nil

	case keys.KeyEnter:
		// Create new feature with the entered name
		if m.newFeatureName != "" {
			// Trim whitespace and set as feature value
			m.featureValue = strings.TrimSpace(m.newFeatureName)
			m.isCreatingNew = false
			m.newFeatureName = ""
		}
		return nil

	case keys.KeyBackspace:
		// Remove last character
		if len(m.newFeatureName) > 0 {
			m.newFeatureName = m.newFeatureName[:len(m.newFeatureName)-1]
		}
		return nil

	case keys.KeyCtrlU:
		// Clear entire input
		m.newFeatureName = ""
		return nil

	default:
		// Add character to feature name (basic text input with validation)
		if len(keyString) == 1 && len(m.newFeatureName) < 30 { // 30 character limit
			// Only allow alphanumeric and basic characters
			char := keyString[0]
			if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
				(char >= '0' && char <= '9') || char == '_' || char == '-' || char == ' ' {
				m.newFeatureName += keyString
			}
		}
		return nil
	}
}

// handleFeatureSelectionMode handles input when in feature selection mode (viewport expanded)
func (m *TaskEditModel) handleFeatureSelectionMode(keyString string) tea.Cmd {
	switch keyString {
	case keys.KeyJ, keys.KeyArrowDown:
		// Navigate down in feature list
		m.selectedFeatureIndex++
		if m.selectedFeatureIndex >= len(m.availableFeatures) {
			m.selectedFeatureIndex = len(m.availableFeatures) - 1
		}
		return nil

	case keys.KeyK, keys.KeyArrowUp:
		// Navigate up in feature list
		m.selectedFeatureIndex--
		if m.selectedFeatureIndex < 0 {
			m.selectedFeatureIndex = 0
		}
		return nil

	case keys.KeyEnter:
		// Confirm selection and exit selection mode
		if m.selectedFeatureIndex >= 0 && m.selectedFeatureIndex < len(m.availableFeatures) {
			m.featureValue = m.availableFeatures[m.selectedFeatureIndex]
		}
		m.featureSelectionMode = false
		return nil

	case keys.KeyH, keys.KeyEscape:
		// Cancel selection and exit selection mode
		m.featureSelectionMode = false
		return nil

	case keys.KeyN:
		// Create new feature - exit selection mode and enter create mode
		m.featureSelectionMode = false
		m.isCreatingNew = true
		m.newFeatureName = ""
		return nil

	case keys.KeySpace:
		// Save all changes and close modal
		return m.saveChanges()

	default:
		return nil
	}
}

// handleScroll processes scroll messages
func (m *TaskEditModel) handleScroll(msg TaskEditModalScrollMsg) tea.Cmd {
	// No scroll handling needed for inline fields
	// Feature selection now delegates to feature modal
	return nil
}

// =============================================================================
// NAVIGATION HELPERS
// =============================================================================
// (Feature navigation helpers removed - now delegating to feature modal)

// =============================================================================
// SAVE AND CHANGE DETECTION
// =============================================================================

// saveChanges detects what changed and broadcasts update message
func (m *TaskEditModel) saveChanges() tea.Cmd {
	// Detect changes
	var status, feature *string
	var priority *int

	if m.statusValue != m.originalStatus {
		status = &m.statusValue
	}

	if m.priorityValue != m.originalPriority {
		priority = &m.priorityValue
	}

	if m.featureValue != m.originalFeature {
		feature = &m.featureValue
	}

	// Only send update if something changed
	if status != nil || priority != nil || feature != nil {
		return tea.Batch(
			m.BroadcastMessage(TaskPropertiesUpdatedMsg{
				TaskID:   m.taskID,
				Status:   status,
				Priority: priority,
				Feature:  feature,
			}),
			m.BroadcastMessage(HideTaskEditModalMsg{}),
		)
	}

	// Nothing changed, just close
	return m.BroadcastMessage(HideTaskEditModalMsg{})
}

// getStatusIndex returns the index for a status string
func (m *TaskEditModel) getStatusIndex(status string) int {
	for i, s := range statusOptions {
		if s == status {
			return i
		}
	}
	return 0 // Default to todo
}

// updateDimensions updates the modal dimensions from WindowSizeMsg
func (m *TaskEditModel) updateDimensions(width, height int) {
	modalWidth := min(width-4, 60)
	modalHeight := min(height-4, 25)
	m.SetDimensions(modalWidth, modalHeight)
}

// renderModal renders the complete task edit modal
func (m *TaskEditModel) renderModal() string {
	// Create the content
	content := m.renderContent()

	// Use modal dimensions
	modalWidth := m.GetWidth()
	modalHeight := m.GetHeight()

	// Create the modal with border
	modal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("51")). // Bright cyan like other modals
		Width(modalWidth).
		Height(modalHeight).
		Padding(1, 2).
		Align(lipgloss.Center, lipgloss.Top). // Top align for list content
		Render(content)

	// Parent handles positioning in proper parent-child architecture
	return modal
}

// renderContent renders the modal content with all three fields
func (m *TaskEditModel) renderContent() string {
	var content strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("51"))
	title := titleStyle.Render("Edit Task Properties")
	content.WriteString(title)
	content.WriteString("\n\n")

	// Render each field
	content.WriteString(m.renderStatusField())
	content.WriteString("\n\n")
	content.WriteString(m.renderPriorityField())
	content.WriteString("\n\n")
	content.WriteString(m.renderFeatureFieldSection())

	// Instructions at bottom - context-sensitive based on mode
	content.WriteString("\n\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	var instructions string

	switch {
	case m.featureSelectionMode && m.activeField == FieldFeature:
		// In feature selection mode - show viewport navigation help
		instructions = helpStyle.Render("j/k: Navigate features • Enter: Confirm • h/Esc: Cancel • Space: Save all")
	case m.isCreatingNew && m.activeField == FieldFeature:
		// Creating new feature - show text input help
		instructions = helpStyle.Render("Type name • Enter: Confirm • Esc: Cancel")
	default:
		// Normal mode - show general navigation help
		instructions = helpStyle.Render("j/k: Change field • h/l: Adjust value • Space/Enter: Save • Esc: Cancel")
	}

	content.WriteString(instructions)

	return content.String()
}

// =============================================================================
// FIELD RENDERING - Render each field type
// =============================================================================

// renderStatusField renders the status selection field
func (m *TaskEditModel) renderStatusField() string {
	var content strings.Builder

	// Field label
	labelStyle := lipgloss.NewStyle().Bold(true)
	if m.activeField == FieldStatus {
		labelStyle = labelStyle.Foreground(lipgloss.Color("51")) // Highlight if active
	} else {
		labelStyle = labelStyle.Foreground(lipgloss.Color("240")) // Dim if inactive
	}
	content.WriteString(labelStyle.Render("Status:"))
	content.WriteString("  ")

	// Status options with symbols
	titleCaser := cases.Title(language.English)
	for i, status := range statusOptions { //nolint:varnamelen // i is idiomatic for loop index
		symbol := m.getStatusSymbol(status)
		statusText := titleCaser.String(status)

		var style lipgloss.Style
		if i == m.statusIndex {
			// Current selection
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Bold(true)
			if m.activeField == FieldStatus {
				style = style.Background(lipgloss.Color("62")) // Highlight if field active
			}
		} else {
			// Other options
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		}

		content.WriteString(style.Render(fmt.Sprintf("%s %s", symbol, statusText)))
		if i < len(statusOptions)-1 {
			content.WriteString("  ")
		}
	}

	return content.String()
}

// renderPriorityField renders the priority input/display field
func (m *TaskEditModel) renderPriorityField() string {
	var content strings.Builder

	// Field label
	labelStyle := lipgloss.NewStyle().Bold(true)
	if m.activeField == FieldPriority {
		labelStyle = labelStyle.Foreground(lipgloss.Color("51")) // Highlight if active
	} else {
		labelStyle = labelStyle.Foreground(lipgloss.Color("240")) // Dim if inactive
	}
	content.WriteString(labelStyle.Render("Priority:"))
	content.WriteString("  ")

	// Get priority category and symbol
	priority := styling.GetTaskPriority(m.priorityValue, nil)
	symbol := styling.GetPrioritySymbol(priority)
	priorityText := m.getPriorityText(priority)
	priorityColor := styling.GetPriorityColor(priority)

	// Show priority value or text input
	var valueStyle lipgloss.Style
	if m.activeField == FieldPriority {
		valueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true)
		if m.priorityEditMode {
			valueStyle = valueStyle.Background(lipgloss.Color("236")) // Input background
		} else {
			valueStyle = valueStyle.Background(lipgloss.Color("62")) // Selection background
		}
	} else {
		valueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	}

	// Render priority display
	symbolStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(priorityColor))
	content.WriteString(symbolStyle.Render(symbol))
	content.WriteString(" ")

	if m.priorityEditMode && m.activeField == FieldPriority {
		// Text input mode
		inputText := m.priorityInput + "▊"
		content.WriteString(valueStyle.Render(inputText))
	} else {
		// Display mode
		displayText := fmt.Sprintf("%s (%d)", priorityText, m.priorityValue)
		content.WriteString(valueStyle.Render(displayText))
	}

	return content.String()
}

// renderFeatureFieldSection renders the feature field with optional viewport expansion
func (m *TaskEditModel) renderFeatureFieldSection() string {
	var content strings.Builder

	// Field label
	labelStyle := lipgloss.NewStyle().Bold(true)
	if m.activeField == FieldFeature {
		labelStyle = labelStyle.Foreground(lipgloss.Color("51")) // Highlight if active
	} else {
		labelStyle = labelStyle.Foreground(lipgloss.Color("240")) // Dim if inactive
	}
	content.WriteString(labelStyle.Render("Feature:"))

	// If creating new feature, show text input
	if m.isCreatingNew {
		content.WriteString("  ")
		inputStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("236")).
			Bold(true)
		inputText := m.newFeatureName + "▊"
		content.WriteString(inputStyle.Render(inputText))
		return content.String()
	}

	// If in selection mode, show expanded viewport with feature list
	if m.featureSelectionMode && m.activeField == FieldFeature {
		content.WriteString("\n")
		content.WriteString(m.renderFeatureViewport())
		return content.String()
	}

	// Normal mode - collapsed, just show current value
	content.WriteString("  ")
	var valueStyle lipgloss.Style
	if m.activeField == FieldFeature {
		valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Bold(true).
			Background(lipgloss.Color("62")) // Selection background
	} else {
		valueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	}

	if m.featureValue != "" {
		content.WriteString(valueStyle.Render(m.featureValue))
	} else {
		content.WriteString(valueStyle.Render("(none)"))
	}

	// Hint for feature field
	if m.activeField == FieldFeature {
		content.WriteString("  ")
		hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
		content.WriteString(hintStyle.Render("[l/Enter: select | n: new]"))
	}

	return content.String()
}

// renderFeatureViewport renders the expanded feature list viewport
func (m *TaskEditModel) renderFeatureViewport() string {
	if len(m.availableFeatures) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true).
			Padding(0, 2)
		return emptyStyle.Render("(no features available)")
	}

	items := make([]string, 0, len(m.availableFeatures)) // Preallocate for all features
	for i, feature := range m.availableFeatures {
		isSelected := i == m.selectedFeatureIndex

		var itemStyle lipgloss.Style
		var prefix string

		if isSelected {
			// Highlighted selection
			itemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Bold(true)
			prefix = "► "
		} else {
			// Normal item
			itemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))
			prefix = "  "
		}

		items = append(items, itemStyle.Render(prefix+feature))
	}

	// Limit to 7 visible items with scrolling
	startIndex := m.selectedFeatureIndex - 3
	if startIndex < 0 {
		startIndex = 0
	}
	endIndex := startIndex + 7
	if endIndex > len(items) {
		endIndex = len(items)
		startIndex = max(0, endIndex-7)
	}

	visibleItems := items[startIndex:endIndex]

	// Render viewport with border using lipgloss.JoinVertical
	viewportContent := lipgloss.JoinVertical(lipgloss.Left, visibleItems...)
	viewport := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1).
		Width(40).
		Render(viewportContent)

	return viewport
}

// =============================================================================
// RENDERING HELPERS
// =============================================================================

// getStatusSymbol returns the visual symbol for a status
func (m *TaskEditModel) getStatusSymbol(status string) string {
	switch status {
	case archon.TaskStatusTodo:
		return "○"
	case archon.TaskStatusDoing:
		return "◐"
	case archon.TaskStatusReview:
		return "◈"
	case archon.TaskStatusDone:
		return "✓"
	default:
		return "○"
	}
}

// getPriorityText returns human-readable priority text
func (m *TaskEditModel) getPriorityText(priority styling.PriorityLevel) string {
	switch priority {
	case styling.PriorityHigh:
		return "High"
	case styling.PriorityMedium:
		return "Medium"
	case styling.PriorityLow:
		return "Low"
	default:
		return "Unknown"
	}
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

//nolint:unparam // a always 0 in current usage - helper kept for symmetry with min/max pattern
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// findFeatureIndex finds the index of a feature in the available features slice
// Returns the index if found, -1 if not found
func (m *TaskEditModel) findFeatureIndex(feature string, features []string) int {
	for i, f := range features {
		if f == feature {
			return i
		}
	}
	return -1
}
