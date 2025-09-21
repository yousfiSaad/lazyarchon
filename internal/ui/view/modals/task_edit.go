package modals

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// TaskEditConfig holds configuration for task edit modal
type TaskEditConfig struct {
	IsCreatingNew    bool
	NewFeatureName   string
	SelectedIndex    int
	AvailableFeatures []string
}

// TaskEditHelpers interface for task edit related functions
type TaskEditHelpers interface {
	GetUniqueFeatures() []string
	GetFeatureTaskCount(feature string) int
}

// RenderTaskEditModal renders the task edit modal overlay on top of the base UI
func RenderTaskEditModal(config TaskEditConfig, factory StyleFactory, helpers TaskEditHelpers, headerColor, mutedColor string, windowWidth, windowHeight int) string {
	// Calculate modal dimensions (similar to feature modal)
	modalWidth := min(windowWidth-4, 50)   // Wide enough for feature names + task counts
	modalHeight := min(windowHeight-4, 18) // Tall enough for feature list + create new option

	// Get task edit content
	editContent := GetTaskEditContent(config, factory, helpers, headerColor, mutedColor)

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
		windowWidth, windowHeight,
		lipgloss.Center, lipgloss.Center,
		editModal,
	)

	return centeredModal
}

// GetTaskEditContent returns the task edit modal content
func GetTaskEditContent(config TaskEditConfig, factory StyleFactory, helpers TaskEditHelpers, headerColor, mutedColor string) []string {
	var content []string

	// Title
	content = append(content, factory.Header().Render("Edit Task"))
	content = append(content, "")

	// If in text input mode for creating new feature
	if config.IsCreatingNew {
		content = append(content, "Feature name:")
		featureInput := config.NewFeatureName + "_"
		content = append(content, factory.Text("33").Render(featureInput))
		content = append(content, "")
		content = append(content, factory.Italic(mutedColor).Render("Enter: create  Esc: cancel"))
		return content
	}

	// Show available features
	availableFeatures := config.AvailableFeatures
	if len(availableFeatures) == 0 {
		availableFeatures = helpers.GetUniqueFeatures()
	}

	if len(availableFeatures) == 0 {
		content = append(content, "No existing features")
		content = append(content, "")
	} else {
		content = append(content, "Feature:")
		content = append(content, "")

		// Feature list
		for i, feature := range availableFeatures {
			// Task count for this feature
			taskCount := helpers.GetFeatureTaskCount(feature)
			taskText := fmt.Sprintf("(%d task", taskCount)
			if taskCount != 1 {
				taskText += "s"
			}
			taskText += ")"

			// Build the line
			line := fmt.Sprintf("%s %s", feature,
				factory.Text("244").Render(taskText))

			// Highlight selected feature
			if i == config.SelectedIndex {
				line = "► " + line + " ◄" // Selection indicator
				line = factory.Bold(headerColor).Render(line)
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

	if config.SelectedIndex == createNewIndex {
		createNewLine = "► " + createNewLine + " ◄"
		createNewLine = factory.Bold(headerColor).Render(createNewLine)
	} else {
		createNewLine = "  " + createNewLine
	}

	content = append(content, factory.Text("34").Render(createNewLine))
	content = append(content, "")

	// Instructions
	content = append(content, factory.Italic(mutedColor).Render("Enter: select  Esc: cancel"))

	return content
}