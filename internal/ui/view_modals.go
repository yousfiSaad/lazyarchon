package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderHelpModal renders the help modal overlay on top of the base UI
func (m Model) renderHelpModal(baseUI string) string {
	// Calculate modal dimensions
	modalWidth := Min(m.Window.width-4, 70)   // Maximum 70 chars wide, with margins
	modalHeight := Min(m.Window.height-4, 25) // Maximum 25 lines high, with margins

	// The viewport content is managed by the model and updated when help is opened
	// We just need to render the viewport view
	viewportContent := m.helpModalViewport.View()

	// Create help modal with border
	helpModal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("51")). // Bright cyan like active panels
		Width(modalWidth).
		Height(modalHeight).
		Padding(1).
		Render(viewportContent)

	// Center the modal on screen
	centeredModal := lipgloss.Place(
		m.Window.width, m.Window.height,
		lipgloss.Center, lipgloss.Center,
		helpModal,
	)

	return centeredModal
}

// renderStatusChangeModal renders the status change modal overlay on top of the base UI
func (m Model) renderStatusChangeModal(baseUI string) string {
	// Calculate modal dimensions (smaller than help modal)
	modalWidth := Min(m.Window.width-4, 40)   // Narrower modal for status selection
	modalHeight := Min(m.Window.height-4, 10) // Shorter modal for 4 status options

	// Get status change content
	statusContent := m.getStatusChangeContent()

	// Create status modal with border
	statusText := strings.Join(statusContent, "\n")
	statusModal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("51")). // Bright cyan like active panels
		Width(modalWidth).
		Height(modalHeight).
		Padding(1).
		Render(statusText)

	// Center the modal on screen
	centeredModal := lipgloss.Place(
		m.Window.width, m.Window.height,
		lipgloss.Center, lipgloss.Center,
		statusModal,
	)

	return centeredModal
}

// getStatusChangeContent returns the status selection content
func (m Model) getStatusChangeContent() []string {
	var content []string

	// Title
	content = append(content, lipgloss.NewStyle().Bold(true).Render("Change Status"))
	content = append(content, "")

	// Status options with symbols and colors
	statuses := []struct {
		name   string
		symbol string
		color  string
	}{
		{"Todo", "○", "240"},  // gray
		{"Doing", "◐", "33"},  // yellow
		{"Review", "◉", "34"}, // blue
		{"Done", "●", "32"},   // green
	}

	for i, status := range statuses {
		// Style for status line
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(status.color))
		line := fmt.Sprintf("%s %s", status.symbol, status.name)

		// Highlight selected status
		if i == m.Modals.statusChange.selectedIndex {
			line = "► " + line + " ◄" // Selection indicator
			line = lipgloss.NewStyle().Bold(true).Render(line)
		} else {
			line = "  " + line
		}

		content = append(content, statusStyle.Render(line))
	}

	content = append(content, "")
	content = append(content, lipgloss.NewStyle().Italic(true).Render("Enter: select  Esc: cancel"))

	return content
}

// renderConfirmationModal renders the confirmation modal overlay on top of the base UI
func (m Model) renderConfirmationModal(baseUI string) string {
	// Calculate modal dimensions (smaller than help modal)
	modalWidth := Min(m.Window.width-4, 50)  // Narrower modal for confirmation
	modalHeight := Min(m.Window.height-4, 8) // Shorter modal for confirmation

	// Get confirmation content
	confirmationContent := m.getConfirmationContent()

	// Create confirmation modal with border
	confirmationText := strings.Join(confirmationContent, "\n")
	confirmationModal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")). // Red border for confirmation
		Width(modalWidth).
		Height(modalHeight).
		Padding(1).
		Render(confirmationText)

	// Center the modal on screen
	centeredModal := lipgloss.Place(
		m.Window.width, m.Window.height,
		lipgloss.Center, lipgloss.Center,
		confirmationModal,
	)

	return centeredModal
}

// getConfirmationContent returns the confirmation modal content
func (m Model) getConfirmationContent() []string {
	var content []string

	// Title with warning style
	content = append(content, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")).Render("Confirmation"))
	content = append(content, "")

	// Message
	content = append(content, m.Modals.confirmation.message)
	content = append(content, "")

	// Options with selection indicators
	confirmOption := fmt.Sprintf("%s %s", m.Modals.confirmation.confirmText, "(y)")
	cancelOption := fmt.Sprintf("%s %s", m.Modals.confirmation.cancelText, "(n)")

	if m.Modals.confirmation.selectedOption == 0 {
		// Confirm option selected
		confirmOption = "► " + confirmOption + " ◄"
		confirmOption = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")).Render(confirmOption)
		cancelOption = "  " + cancelOption
	} else {
		// Cancel option selected
		confirmOption = "  " + confirmOption
		cancelOption = "► " + cancelOption + " ◄"
		cancelOption = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("34")).Render(cancelOption)
	}

	content = append(content, confirmOption)
	content = append(content, cancelOption)
	content = append(content, "")

	// Instructions
	content = append(content, lipgloss.NewStyle().Italic(true).Render("Enter/y: confirm  Esc/n: cancel"))

	return content
}

// renderFeatureModal renders the feature selection modal overlay on top of the base UI
func (m Model) renderFeatureModal(baseUI string) string {
	// Calculate modal dimensions (similar to status change modal but taller)
	modalWidth := Min(m.Window.width-4, 45)   // Slightly wider for feature names + task counts
	modalHeight := Min(m.Window.height-4, 15) // Taller to accommodate multiple features

	// Get feature selection content
	featureContent := m.getFeatureSelectionContent()

	// Create feature modal with border
	featureText := strings.Join(featureContent, "\n")
	featureModal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("51")). // Bright cyan like active panels
		Width(modalWidth).
		Height(modalHeight).
		Padding(1).
		Render(featureText)

	// Center the modal on screen
	centeredModal := lipgloss.Place(
		m.Window.width, m.Window.height,
		lipgloss.Center, lipgloss.Center,
		featureModal,
	)

	return centeredModal
}

// getFeatureSelectionContent returns the feature selection modal content
func (m Model) getFeatureSelectionContent() []string {
	var content []string
	availableFeatures := m.GetUniqueFeatures()

	// Title
	content = append(content, lipgloss.NewStyle().Bold(true).Render("Select Features"))
	content = append(content, "")

	if len(availableFeatures) == 0 {
		content = append(content, "No features available")
		content = append(content, "")
		content = append(content, lipgloss.NewStyle().Italic(true).Render("Enter: close"))
		return content
	}

	// Feature list with checkboxes
	for i, feature := range availableFeatures {
		// Determine if feature is enabled
		enabled := true // Default to enabled
		if len(m.Modals.featureMode.selectedFeatures) > 0 {
			if state, exists := m.Modals.featureMode.selectedFeatures[feature]; exists {
				enabled = state
			}
		}

		// Checkbox symbol
		checkbox := "☐" // Unchecked
		if enabled {
			checkbox = "☑" // Checked
		}

		// Task count for this feature
		taskCount := m.GetFeatureTaskCount(feature)
		taskText := fmt.Sprintf("(%d task", taskCount)
		if taskCount != 1 {
			taskText += "s"
		}
		taskText += ")"

		// Build the line
		line := fmt.Sprintf("%s %s %s", checkbox, feature,
			lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render(taskText))

		// Highlight selected feature
		if i == m.Modals.featureMode.selectedIndex {
			line = "► " + line + " ◄" // Selection indicator
			line = lipgloss.NewStyle().Bold(true).Render(line)
		} else {
			line = "  " + line
		}

		content = append(content, line)
	}

	content = append(content, "")
	content = append(content, "──────────────────────────────────────")
	content = append(content, "")

	// Instructions
	content = append(content, lipgloss.NewStyle().Italic(true).Render("Space: toggle  a: all  n: none"))
	content = append(content, lipgloss.NewStyle().Italic(true).Render("Enter: apply   Esc/q: cancel"))

	return content
}

// renderTaskEditModal renders the task edit modal overlay on top of the base UI
func (m Model) renderTaskEditModal(baseUI string) string {
	// Calculate modal dimensions (similar to feature modal)
	modalWidth := Min(m.Window.width-4, 50)   // Wide enough for feature names + task counts
	modalHeight := Min(m.Window.height-4, 18) // Tall enough for feature list + create new option

	// Get task edit content
	editContent := m.getTaskEditContent()

	// Create edit modal with border
	editText := strings.Join(editContent, "\n")
	editModal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("51")). // Bright cyan like active panels
		Width(modalWidth).
		Height(modalHeight).
		Padding(1).
		Render(editText)

	// Center the modal on screen
	centeredModal := lipgloss.Place(
		m.Window.width, m.Window.height,
		lipgloss.Center, lipgloss.Center,
		editModal,
	)

	return centeredModal
}

// getTaskEditContent returns the task edit modal content
func (m Model) getTaskEditContent() []string {
	var content []string

	// Title
	content = append(content, lipgloss.NewStyle().Bold(true).Render("Edit Task"))
	content = append(content, "")

	// If in text input mode for creating new feature
	if m.Modals.taskEdit.isCreatingNew {
		content = append(content, "Feature name:")
		featureInput := m.Modals.taskEdit.newFeatureName + "_"
		content = append(content, lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Render(featureInput))
		content = append(content, "")
		content = append(content, lipgloss.NewStyle().Italic(true).Render("Enter: create  Esc: cancel"))
		return content
	}

	// Show available features
	availableFeatures := m.GetUniqueFeatures()

	if len(availableFeatures) == 0 {
		content = append(content, "No existing features")
		content = append(content, "")
	} else {
		content = append(content, "Feature:")
		content = append(content, "")

		// Feature list
		for i, feature := range availableFeatures {
			// Task count for this feature
			taskCount := m.GetFeatureTaskCount(feature)
			taskText := fmt.Sprintf("(%d task", taskCount)
			if taskCount != 1 {
				taskText += "s"
			}
			taskText += ")"

			// Build the line
			line := fmt.Sprintf("%s %s", feature,
				lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render(taskText))

			// Highlight selected feature
			if i == m.Modals.taskEdit.selectedIndex {
				line = "► " + line + " ◄" // Selection indicator
				line = lipgloss.NewStyle().Bold(true).Render(line)
			} else {
				line = "  " + line
			}

			content = append(content, line)
		}

		content = append(content, "")
	}

	// "Create new feature" option
	createNewLine := "[Create new feature]"
	createNewIndex := len(availableFeatures)

	if m.Modals.taskEdit.selectedIndex == createNewIndex {
		createNewLine = "► " + createNewLine + " ◄"
		createNewLine = lipgloss.NewStyle().Bold(true).Render(createNewLine)
	} else {
		createNewLine = "  " + createNewLine
	}

	content = append(content, lipgloss.NewStyle().Foreground(lipgloss.Color("34")).Render(createNewLine))
	content = append(content, "")

	// Instructions
	content = append(content, lipgloss.NewStyle().Italic(true).Render("Enter: select  Esc: cancel"))

	return content
}
