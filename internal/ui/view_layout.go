package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the complete UI
func (m Model) View() string {
	if !m.Window.ready {
		return "Loading..."
	}

	// Header with project state
	header := m.renderHeader()

	// Status bar with mode-specific controls
	status := m.renderStatusBar()

	// Calculate dimensions for panels
	contentHeight := m.GetContentHeight()
	leftWidth := m.GetLeftPanelWidth()
	rightWidth := m.GetRightPanelWidth()

	// Left panel (task list or project list depending on mode)
	var leftPanel string
	if m.Modals.projectMode.active {
		leftPanel = m.renderProjectList(leftWidth, contentHeight)
	} else {
		leftPanel = m.renderTaskList(leftWidth, contentHeight)
	}

	// Right panel (task details or project mode instructions)
	var rightPanel string
	if m.Modals.projectMode.active {
		rightPanel = m.renderProjectModeHelp(rightWidth, contentHeight)
	} else {
		rightPanel = m.renderTaskDetails(rightWidth, contentHeight)
	}

	// Combine panels horizontally
	content := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	// Base UI layout
	baseUI := lipgloss.JoinVertical(lipgloss.Left, header, content, status)

	// Overlay help modal if open
	if m.IsHelpMode() {
		return m.renderHelpModal(baseUI)
	}

	// Overlay status change modal if open
	if m.IsStatusChangeMode() {
		return m.renderStatusChangeModal(baseUI)
	}

	// Overlay confirmation modal if open
	if m.IsConfirmationMode() {
		return m.renderConfirmationModal(baseUI)
	}

	// Overlay feature selection modal if open
	if m.IsFeatureModeActive() {
		return m.renderFeatureModal(baseUI)
	}

	// Overlay task edit modal if open
	if m.IsTaskEditModeActive() {
		return m.renderTaskEditModal(baseUI)
	}

	return baseUI
}

// renderHeader renders the top header bar
func (m Model) renderHeader() string {
	var headerText string

	if m.Modals.projectMode.active {
		headerText = "LazyArchon - Select Project"
	} else if m.IsFeatureModeActive() {
		headerText = "LazyArchon - Select Features"
	} else {
		// Build header with project and feature filter information
		projectName := m.GetCurrentProjectName()
		featureSummary := m.GetFeatureFilterSummary()
		taskCount := len(m.GetSortedTasks()) // Use filtered task count

		if featureSummary == "All features" || featureSummary == "No features" {
			// No explicit feature filtering, show simple format
			headerText = fmt.Sprintf("LazyArchon - %s (%d)", projectName, taskCount)
		} else {
			// Show feature filter status
			headerText = fmt.Sprintf("LazyArchon - %s • %s (%d)", projectName, featureSummary, taskCount)
		}
	}

	return HeaderStyle.Render(headerText)
}

// renderStatusBar renders the bottom status bar
func (m Model) renderStatusBar() string {
	var statusText string

	// Handle special states first
	if m.IsHelpMode() {
		statusText = "[Help] j/k to scroll • ESC: close | q: quit"
		return StatusBarStyle.Render(statusText)
	}

	if m.Modals.projectMode.active {
		projectCount := len(m.Data.projects)
		if projectCount > 0 {
			statusText = fmt.Sprintf("[Project] %d projects available | l: select | h: back | q: quit", projectCount)
		} else {
			statusText = "Project Selection | ?: help | q: quit"
		}
		return StatusBarStyle.Render(statusText)
	}

	if m.IsFeatureModeActive() {
		availableFeatures := m.GetUniqueFeatures()
		featureCount := len(availableFeatures)
		if featureCount > 0 {
			statusText = fmt.Sprintf("[Features] %d features available | Space: toggle | a: all | n: none | Enter: apply | q: cancel", featureCount)
		} else {
			statusText = "Feature Selection | No features available | Enter: apply | q: cancel"
		}
		return StatusBarStyle.Render(statusText)
	}

	// Show loading state
	if m.Data.loading {
		statusText = "[Tasks] Loading... | q: quit"
		return StatusBarStyle.Render(statusText)
	}

	// Show error state
	if m.Data.error != "" {
		statusText = "[Tasks] Error • Check connection | r: retry | q: quit"
		return StatusBarStyle.Render(statusText)
	}

	// Context-aware status based on active panel
	activePanel := m.GetActiveViewName()

	switch activePanel {
	case "Tasks":
		// Left panel active - show task counts, sort mode, and position
		todo, doing, review, done := m.GetTaskStatusCounts()
		totalTasks := todo + doing + review + done
		sortMode := m.GetCurrentSortModeName()

		if totalTasks == 0 {
			statusText = "[Tasks] No tasks found | r: refresh | q: quit"
		} else {
			// Build status with task counts and sort mode
			var statusParts []string
			statusParts = append(statusParts, fmt.Sprintf("%d items", totalTasks))

			// Add status distribution if there are active tasks
			if doing > 0 || review > 0 {
				if doing > 0 {
					statusParts = append(statusParts, fmt.Sprintf("%d doing", doing))
				}
				if review > 0 {
					statusParts = append(statusParts, fmt.Sprintf("%d review", review))
				}
			}

			// Add todo count if significant
			if todo > 0 {
				statusParts = append(statusParts, fmt.Sprintf("%d todo", todo))
			}

			// Add sort mode
			statusParts = append(statusParts, fmt.Sprintf("Sort: %s", sortMode))

			statusInfo := strings.Join(statusParts, " • ")

			// Add feature shortcut if features are available
			if len(m.GetUniqueFeatures()) > 0 {
				statusText = fmt.Sprintf("[Tasks] %s | f: features | ?: help", statusInfo)
			} else {
				statusText = fmt.Sprintf("[Tasks] %s | ?: help", statusInfo)
			}
		}

	case "Details":
		// Right panel active - show current position and scroll info
		position := m.GetCurrentPosition()
		scrollPos := m.GetScrollPosition()

		if scrollPos == "Top" {
			statusText = fmt.Sprintf("[Details] %s | ?: help", position)
		} else {
			statusText = fmt.Sprintf("[Details] %s • %s | ?: help", position, scrollPos)
		}

	default:
		// Fallback to simple status
		statusText = fmt.Sprintf("[%s] Ready | ?: help | q: quit", activePanel)
	}

	return StatusBarStyle.Render(statusText)
}
