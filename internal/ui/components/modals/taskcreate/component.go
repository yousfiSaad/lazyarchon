package taskcreate

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yousfisaad/lazyarchon/v2/internal/archon"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/styling"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/utils/keys"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/base"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/messages"
)

const ComponentID = "task-create-modal"

// TaskCreateModel represents the task creation modal component
// Architecture: Follows four-tier state pattern
// - No source data caching (receives project/feature data via ShowTaskCreateModalMsg)
// - No display parameters (manages its own form state during creation)
// - Owned state only (form fields, edit modes, working values)
// - No transient feedback (modal lifecycle managed by MainModel)
// - Modal lifecycle managed by BaseModal (active/visible state)
type TaskCreateModel struct {
	base.BaseModal

	// ===================================================================
	// OWNED STATE - Component manages these directly
	// ===================================================================

	// Project context (passed via message)
	projectID string // Project to create task in

	// Multi-field form state
	activeField FieldType // Currently focused field

	// Field values (working state - what user is entering)
	titleValue       string // Task title (required)
	descriptionValue string // Task description (optional)
	featureValue     string // Feature assignment (optional)
	priorityValue    int    // Task priority (defaults to 50)
	statusValue      string // Task status (defaults to "todo")

	// Title field state (text input)
	titleCursorPos int // Cursor position in title field

	// Description field state (text input)
	descriptionCursorPos int // Cursor position in description field

	// Priority field state
	priorityEditMode bool   // true when typing specific number
	priorityInput    string // Text being typed for priority

	// Feature field state
	availableFeatures    []string // Available features to choose from
	selectedFeatureIndex int      // Currently highlighted feature in selection mode
	featureSelectionMode bool     // true when viewport is expanded for selection
	isCreatingNew        bool     // true when in text input mode for new feature
	newFeatureName       string   // Text being typed for new feature
}

// NewModel creates a new task creation modal component
func NewModel(context *base.ComponentContext) *TaskCreateModel {
	baseModal := base.NewBaseModal(
		ComponentID,
		base.TaskCreateModalComponent,
		context,
	)

	model := &TaskCreateModel{
		BaseModal:            baseModal,
		activeField:          FieldTitle,            // Start on title field
		priorityValue:        50,                    // Default medium priority
		statusValue:          archon.TaskStatusTodo, // Default to todo
		titleCursorPos:       0,
		descriptionCursorPos: 0,
		priorityEditMode:     false,
		priorityInput:        "",
		selectedFeatureIndex: 0,
		featureSelectionMode: false,
		isCreatingNew:        false,
		newFeatureName:       "",
	}
	// Set dimensions using base component
	model.SetDimensions(70, 20) // Wider for description, taller for all fields
	return model
}

// CanFocus overrides the base implementation to allow focus
func (m *TaskCreateModel) CanFocus() bool {
	return true
}

// Init initializes the task create modal component
func (m *TaskCreateModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the task create modal component
func (m *TaskCreateModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case ShowTaskCreateModalMsg:
		m.SetActive(true)
		m.SetFocus(true)

		// Set project info
		m.projectID = msg.DefaultProjectID
		m.availableFeatures = msg.AvailableFeatures

		// Reset form to defaults
		m.activeField = FieldTitle
		m.titleValue = ""
		m.descriptionValue = ""
		m.featureValue = msg.DefaultFeature
		m.priorityValue = 50
		m.statusValue = archon.TaskStatusTodo
		m.titleCursorPos = 0
		m.descriptionCursorPos = 0
		m.priorityEditMode = false
		m.priorityInput = ""
		m.featureSelectionMode = false
		m.isCreatingNew = false
		m.newFeatureName = ""

		// Pre-select the default feature if provided
		if m.featureValue != "" {
			if index := m.findFeatureIndex(m.featureValue, m.availableFeatures); index != -1 {
				m.selectedFeatureIndex = index
			} else {
				m.selectedFeatureIndex = 0
			}
		} else {
			m.selectedFeatureIndex = 0
		}

		return m.BroadcastMessage(messages.ModalStateMsg{
			Type:   string(base.ModalTypeTaskCreate),
			Active: true,
		})

	case HideTaskCreateModalMsg:
		m.SetActive(false)
		m.SetFocus(false)
		m.priorityEditMode = false
		m.priorityInput = ""
		m.featureSelectionMode = false
		m.isCreatingNew = false
		m.newFeatureName = ""
		return m.BroadcastMessage(messages.ModalStateMsg{
			Type:   string(base.ModalTypeTaskCreate),
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

	default:
		return nil
	}
}

// View renders the task create modal
func (m *TaskCreateModel) View() string {
	if !m.IsActive() {
		return ""
	}

	return m.renderModal()
}

// handleKeyPress processes keyboard input for the task creation modal
//
//nolint:gocyclo // Complexity unavoidable - modal has multiple input modes and field types
func (m *TaskCreateModel) handleKeyPress(key tea.KeyMsg) tea.Cmd {
	keyString := key.String()

	// Check if we're in a special mode that needs priority routing
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
		// Cancel and close modal without creating
		return m.BroadcastMessage(HideTaskCreateModalMsg{})

	case keys.KeyCtrlC:
		return tea.Quit

	case keys.KeyArrowDown, keys.KeyTab:
		// Navigate to next field (Tab or Arrow Down)
		m.activeField = (m.activeField + 1) % 4
		// Reset field-specific modes when changing fields
		m.priorityEditMode = false
		m.isCreatingNew = false
		m.featureSelectionMode = false
		return nil

	case keys.KeyArrowUp, keys.KeyShiftTab:
		// Navigate to previous field (Shift-Tab or Arrow Up)
		m.activeField = (m.activeField - 1 + 4) % 4
		// Reset field-specific modes when changing fields
		m.priorityEditMode = false
		m.isCreatingNew = false
		m.featureSelectionMode = false
		return nil
	}

	// Route to active field handler
	switch m.activeField {
	case FieldTitle:
		return m.handleTitleField(keyString)
	case FieldDescription:
		return m.handleDescriptionField(keyString)
	case FieldFeature:
		return m.handleFeatureField(keyString)
	case FieldPriority:
		return m.handlePriorityField(keyString)
	default:
		return nil
	}
}

// =============================================================================
// FIELD HANDLERS - Handle input for each field type
// =============================================================================

// handleTitleField handles input when title field is focused
//
//nolint:dupl // Similar structure to handleDescriptionField but operates on different fields with different limits
func (m *TaskCreateModel) handleTitleField(keyString string) tea.Cmd {
	switch keyString {
	case keys.KeyEnter:
		// Create task if title is not empty
		if strings.TrimSpace(m.titleValue) != "" {
			return m.createTask()
		}
		return nil

	case keys.KeyBackspace:
		// Remove character before cursor
		if m.titleCursorPos > 0 && len(m.titleValue) > 0 {
			m.titleValue = m.titleValue[:m.titleCursorPos-1] + m.titleValue[m.titleCursorPos:]
			m.titleCursorPos--
		}
		return nil

	case keys.KeyCtrlU:
		// Clear entire title
		m.titleValue = ""
		m.titleCursorPos = 0
		return nil

	case keys.KeyArrowLeft:
		// Move cursor left
		if m.titleCursorPos > 0 {
			m.titleCursorPos--
		}
		return nil

	case keys.KeyArrowRight:
		// Move cursor right
		if m.titleCursorPos < len(m.titleValue) {
			m.titleCursorPos++
		}
		return nil

	default:
		// Add character to title (limit to 100 chars)
		if len(keyString) == 1 && len(m.titleValue) < 100 {
			m.titleValue = m.titleValue[:m.titleCursorPos] + keyString + m.titleValue[m.titleCursorPos:]
			m.titleCursorPos++
		}
		return nil
	}
}

// handleDescriptionField handles input when description field is focused
//
//nolint:dupl // Similar structure to handleTitleField but operates on different fields with different limits
func (m *TaskCreateModel) handleDescriptionField(keyString string) tea.Cmd {
	switch keyString {
	case keys.KeyEnter:
		// Create task
		if strings.TrimSpace(m.titleValue) != "" {
			return m.createTask()
		}
		return nil

	case keys.KeyBackspace:
		// Remove character before cursor
		if m.descriptionCursorPos > 0 && len(m.descriptionValue) > 0 {
			m.descriptionValue = m.descriptionValue[:m.descriptionCursorPos-1] + m.descriptionValue[m.descriptionCursorPos:]
			m.descriptionCursorPos--
		}
		return nil

	case keys.KeyCtrlU:
		// Clear entire description
		m.descriptionValue = ""
		m.descriptionCursorPos = 0
		return nil

	case keys.KeyArrowLeft:
		// Move cursor left
		if m.descriptionCursorPos > 0 {
			m.descriptionCursorPos--
		}
		return nil

	case keys.KeyArrowRight:
		// Move cursor right
		if m.descriptionCursorPos < len(m.descriptionValue) {
			m.descriptionCursorPos++
		}
		return nil

	default:
		// Add character to description (limit to 500 chars)
		if len(keyString) == 1 && len(m.descriptionValue) < 500 {
			m.descriptionValue = m.descriptionValue[:m.descriptionCursorPos] + keyString + m.descriptionValue[m.descriptionCursorPos:]
			m.descriptionCursorPos++
		}
		return nil
	}
}

// handlePriorityField handles input when priority field is focused
//
//nolint:gocyclo // Complexity unavoidable - handles multiple input modes and validation
func (m *TaskCreateModel) handlePriorityField(keyString string) tea.Cmd {
	// If in text input mode, handle numeric input
	if m.priorityEditMode {
		return m.handlePriorityTextInput(keyString)
	}

	// Navigation mode - adjust priority with h/l (vim-style horizontal navigation)
	switch keyString {
	case keys.KeyH, keys.KeyArrowLeft:
		// Decrease priority by 1
		m.priorityValue = max(0, m.priorityValue-1)
		return nil

	case keys.KeyL, keys.KeyArrowRight:
		// Increase priority by 1
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
		m.priorityInput = ""
		return nil

	case keys.KeySpace:
		// Create task
		if strings.TrimSpace(m.titleValue) != "" {
			return m.createTask()
		}
		return nil

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
func (m *TaskCreateModel) handlePriorityTextInput(keyString string) tea.Cmd {
	switch keyString {
	case keys.KeyEscape:
		// Cancel text input mode
		m.priorityEditMode = false
		m.priorityInput = ""
		return nil

	case keys.KeyEnter:
		// Confirm entered value
		if m.priorityInput != "" {
			// Parse the input
			value := 0
			for _, char := range m.priorityInput {
				value = value*10 + int(char-'0')
			}
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
//
//nolint:gocyclo // Complexity unavoidable - handles multiple input modes and navigation
func (m *TaskCreateModel) handleFeatureField(keyString string) tea.Cmd {
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
		// Create task
		if strings.TrimSpace(m.titleValue) != "" {
			return m.createTask()
		}
		return nil

	default:
		return nil
	}
}

// handleFeatureTextInput handles input when creating a new feature
//
//nolint:gocyclo // Complexity unavoidable - handles multiple key types and character validation
func (m *TaskCreateModel) handleFeatureTextInput(keyString string) tea.Cmd {
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
//
//nolint:gocyclo // Complexity unavoidable - handles navigation and selection in feature viewport
func (m *TaskCreateModel) handleFeatureSelectionMode(keyString string) tea.Cmd {
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
		// Create task
		if strings.TrimSpace(m.titleValue) != "" {
			return m.createTask()
		}
		return nil

	default:
		return nil
	}
}

// =============================================================================
// TASK CREATION
// =============================================================================

// createTask creates a new task and broadcasts the creation message
func (m *TaskCreateModel) createTask() tea.Cmd {
	// Validate title
	title := strings.TrimSpace(m.titleValue)
	if title == "" {
		// Don't create if title is empty
		return nil
	}

	// Prepare feature value
	var feature *string
	if m.featureValue != "" {
		feature = &m.featureValue
	}

	// Broadcast task created message
	return tea.Batch(
		m.BroadcastMessage(TaskCreatedMsg{
			Title:       title,
			Description: strings.TrimSpace(m.descriptionValue),
			ProjectID:   m.projectID,
			Feature:     feature,
			Priority:    m.priorityValue,
			Status:      m.statusValue,
		}),
		m.BroadcastMessage(HideTaskCreateModalMsg{}),
	)
}

// updateDimensions updates the modal dimensions from WindowSizeMsg
func (m *TaskCreateModel) updateDimensions(width, height int) {
	modalWidth := min(width-4, 70)
	modalHeight := min(height-4, 25)
	m.SetDimensions(modalWidth, modalHeight)
}

// renderModal renders the complete task create modal
func (m *TaskCreateModel) renderModal() string {
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
		Align(lipgloss.Center, lipgloss.Top). // Top align for form content
		Render(content)

	// Parent handles positioning in proper parent-child architecture
	return modal
}

// renderContent renders the modal content with all fields
func (m *TaskCreateModel) renderContent() string {
	var content strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("51"))
	title := titleStyle.Render("Create New Task")
	content.WriteString(title)
	content.WriteString("\n\n")

	// Render each field
	content.WriteString(m.renderTitleField())
	content.WriteString("\n\n")
	content.WriteString(m.renderDescriptionField())
	content.WriteString("\n\n")
	content.WriteString(m.renderFeatureFieldSection())
	content.WriteString("\n\n")
	content.WriteString(m.renderPriorityField())

	// Instructions at bottom - context-sensitive based on mode
	content.WriteString("\n\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	var instructions string

	switch {
	case m.featureSelectionMode && m.activeField == FieldFeature:
		// In feature selection mode
		instructions = helpStyle.Render("j/k: Navigate • Enter: Confirm • h/Esc: Cancel • Space: Create")
	case m.isCreatingNew && m.activeField == FieldFeature:
		// Creating new feature
		instructions = helpStyle.Render("Type name • Enter: Confirm • Esc: Cancel")
	case m.priorityEditMode && m.activeField == FieldPriority:
		// Editing priority
		instructions = helpStyle.Render("Type number • Enter: Confirm • Esc: Cancel")
	default:
		// Normal mode
		instructions = helpStyle.Render("Tab/↑↓: Change field • Type: Enter text • Enter/Space: Create • Esc: Cancel")
	}

	content.WriteString(instructions)

	return content.String()
}

// =============================================================================
// FIELD RENDERING
// =============================================================================

// renderTitleField renders the title input field
//
//nolint:dupl // Similar structure to renderDescriptionField but renders different field with different placeholder
func (m *TaskCreateModel) renderTitleField() string {
	var content strings.Builder

	// Field label
	labelStyle := lipgloss.NewStyle().Bold(true)
	if m.activeField == FieldTitle {
		labelStyle = labelStyle.Foreground(lipgloss.Color("51")) // Highlight if active
	} else {
		labelStyle = labelStyle.Foreground(lipgloss.Color("240")) // Dim if inactive
	}
	content.WriteString(labelStyle.Render("Title *"))
	content.WriteString(" ")

	// Required indicator
	requiredStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Italic(true)
	content.WriteString(requiredStyle.Render("(required)"))
	content.WriteString("\n")

	// Input field
	var valueStyle lipgloss.Style
	if m.activeField == FieldTitle {
		valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("236")) // Input background
	} else {
		valueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	}

	// Show cursor if active
	displayValue := m.titleValue
	if m.activeField == FieldTitle {
		// Insert cursor
		if m.titleCursorPos <= len(displayValue) {
			displayValue = displayValue[:m.titleCursorPos] + "▊" + displayValue[m.titleCursorPos:]
		}
	}

	// Show placeholder if empty
	if displayValue == "" {
		placeholderStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)
		if m.activeField == FieldTitle {
			placeholderStyle = placeholderStyle.Background(lipgloss.Color("236"))
			displayValue = "Enter task title▊"
		} else {
			displayValue = "Enter task title"
		}
		content.WriteString(placeholderStyle.Render(displayValue))
	} else {
		content.WriteString(valueStyle.Render(displayValue))
	}

	return content.String()
}

// renderDescriptionField renders the description input field
//
//nolint:dupl // Similar structure to renderTitleField but renders different field with different placeholder
func (m *TaskCreateModel) renderDescriptionField() string {
	var content strings.Builder

	// Field label
	labelStyle := lipgloss.NewStyle().Bold(true)
	if m.activeField == FieldDescription {
		labelStyle = labelStyle.Foreground(lipgloss.Color("51")) // Highlight if active
	} else {
		labelStyle = labelStyle.Foreground(lipgloss.Color("240")) // Dim if inactive
	}
	content.WriteString(labelStyle.Render("Description"))
	content.WriteString(" ")

	// Optional indicator
	optionalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	content.WriteString(optionalStyle.Render("(optional)"))
	content.WriteString("\n")

	// Input field
	var valueStyle lipgloss.Style
	if m.activeField == FieldDescription {
		valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("236")) // Input background
	} else {
		valueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	}

	// Show cursor if active
	displayValue := m.descriptionValue
	if m.activeField == FieldDescription {
		// Insert cursor
		if m.descriptionCursorPos <= len(displayValue) {
			displayValue = displayValue[:m.descriptionCursorPos] + "▊" + displayValue[m.descriptionCursorPos:]
		}
	}

	// Show placeholder if empty
	if displayValue == "" {
		placeholderStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)
		if m.activeField == FieldDescription {
			placeholderStyle = placeholderStyle.Background(lipgloss.Color("236"))
			displayValue = "Enter task description (optional)▊"
		} else {
			displayValue = "Enter task description (optional)"
		}
		content.WriteString(placeholderStyle.Render(displayValue))
	} else {
		content.WriteString(valueStyle.Render(displayValue))
	}

	return content.String()
}

// renderPriorityField renders the priority input/display field
func (m *TaskCreateModel) renderPriorityField() string {
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
	priorityText := getPriorityText(priority)
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
		displayText := priorityText + " (" + itoa(m.priorityValue) + ")"
		content.WriteString(valueStyle.Render(displayText))
	}

	return content.String()
}

// renderFeatureFieldSection renders the feature field with optional viewport expansion
func (m *TaskCreateModel) renderFeatureFieldSection() string {
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
func (m *TaskCreateModel) renderFeatureViewport() string {
	if len(m.availableFeatures) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true).
			Padding(0, 2)
		return emptyStyle.Render("(no features available)")
	}

	items := make([]string, 0, len(m.availableFeatures))
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

	// Render viewport with border
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
// HELPER FUNCTIONS
// =============================================================================

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

//nolint:unparam // Used with various values throughout the file, not just constant 0
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// findFeatureIndex finds the index of a feature in the available features slice
// Returns the index if found, -1 if not found
func (m *TaskCreateModel) findFeatureIndex(feature string, features []string) int {
	for i, f := range features {
		if f == feature {
			return i
		}
	}
	return -1
}

// getPriorityText returns human-readable priority text
func getPriorityText(priority styling.PriorityLevel) string {
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

// itoa converts an integer to a string (simple implementation for priority display)
func itoa(num int) string {
	if num == 0 {
		return "0"
	}
	if num < 0 {
		return "-" + itoa(-num)
	}
	digits := ""
	for num > 0 {
		digits = string('0'+byte(num%10)) + digits
		num /= 10
	}
	return digits
}
