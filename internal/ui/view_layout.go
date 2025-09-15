package ui

import (
	"fmt"
	"strings"
	"time"

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
		// Build header with project, feature filter, and search information
		projectName := m.GetCurrentProjectName()
		featureSummary := m.GetFeatureFilterSummary()
		taskCount := len(m.GetSortedTasks()) // Use filtered task count

		// Build header parts
		var headerParts []string
		headerParts = append(headerParts, projectName)

		// Add search indicator if active
		if m.Data.searchActive && m.Data.searchQuery != "" {
			searchIndicator := fmt.Sprintf("🔍 \"%s\"", m.Data.searchQuery)
			headerParts = append(headerParts, searchIndicator)
		}

		// Add feature filter if not showing all
		if featureSummary != "All features" && featureSummary != "No features" {
			headerParts = append(headerParts, featureSummary)
		}

		// Join parts with bullets
		if len(headerParts) > 1 {
			headerText = fmt.Sprintf("LazyArchon - %s (%d)", strings.Join(headerParts, " • "), taskCount)
		} else {
			headerText = fmt.Sprintf("LazyArchon - %s (%d)", headerParts[0], taskCount)
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
		return CreateStatusBarStyle("ready").Render(statusText)
	}

	if m.Modals.projectMode.active {
		projectCount := len(m.Data.projects)
		if projectCount > 0 {
			statusText = fmt.Sprintf("[Project] %d projects available | l: select | h: back | q: quit", projectCount)
		} else {
			statusText = "Project Selection | ?: help | q: quit"
		}
		return CreateStatusBarStyle("ready").Render(statusText)
	}

	if m.IsFeatureModeActive() {
		availableFeatures := m.GetUniqueFeatures()
		featureCount := len(availableFeatures)
		if featureCount > 0 {
			statusText = fmt.Sprintf("[Features] %d features | j/k/J/K/gg/G: navigate | Space: toggle | a: all | n: none | Enter: apply | q: cancel", featureCount)
		} else {
			statusText = "Feature Selection | No features available | Enter: apply | q: cancel"
		}
		return CreateStatusBarStyle("ready").Render(statusText)
	}

	// Show loading state with spinner
	if m.Data.loading {
		spinner := m.GetLoadingSpinner()
		message := "Loading..."
		if m.Data.loadingMessage != "" {
			message = m.Data.loadingMessage
		}
		statusText = fmt.Sprintf("[Tasks] %s %s | q: quit", spinner, message)
		return CreateStatusBarStyle("loading").Render(statusText)
	}

	// Show error state with user-friendly message
	if m.Data.error != "" {
		friendlyError := m.FormatUserFriendlyError(m.Data.error)
		statusText = fmt.Sprintf("[Tasks] Error: %s | r: retry | q: quit", friendlyError)
		return CreateStatusBarStyle("error").Render(statusText)
	}

	// Show temporary status message (for copy confirmations, etc.)
	if m.Data.statusMessage != "" && time.Since(m.Data.statusMessageTime) < 3*time.Second {
		statusText = fmt.Sprintf("[Tasks] %s | ?: help | q: quit", m.Data.statusMessage)
		return CreateStatusBarStyle("info").Render(statusText)
	}

	// Handle search mode - show inline search interface
	if m.Data.searchMode {
		cursor := "_" // Simple cursor indicator
		searchText := fmt.Sprintf("[Search] %s%s", m.Data.searchInput, cursor)

		// Add match indicator if search has matches
		if m.Data.totalMatches > 0 {
			searchText += fmt.Sprintf(" (%d matches)", m.Data.totalMatches)
		}

		statusText = searchText + " | Enter: apply | Esc: cancel | Ctrl+U: clear"
		return CreateStatusBarStyle("ready").Render(statusText)
	}

	// Context-aware status based on active panel
	activePanel := m.GetActiveViewName()

	switch activePanel {
	case "Tasks":
		// Left panel active - show task counts, sort mode, connection status and position
		todo, doing, review, done := m.GetTaskStatusCounts()
		totalTasks := todo + doing + review + done
		sortMode := m.GetCurrentSortModeName()
		connectionStatus := m.GetConnectionStatusText()

		if totalTasks == 0 {
			statusText = fmt.Sprintf("[Tasks] %s No tasks found | r: refresh | q: quit", connectionStatus)
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

			// Add search match information if search is active
			if m.Data.searchActive && m.Data.searchQuery != "" {
				if m.Data.totalMatches > 0 {
					matchInfo := fmt.Sprintf("Match %d/%d", m.Data.currentMatchIndex+1, m.Data.totalMatches)
					statusParts = append(statusParts, matchInfo)
				} else {
					statusParts = append(statusParts, "No matches")
				}
			}

			statusInfo := strings.Join(statusParts, " • ")

			// Build status bar with available shortcuts
			var shortcuts []string
			if len(m.GetUniqueFeatures()) > 0 {
				shortcuts = append(shortcuts, "f: features")
			}
			shortcuts = append(shortcuts, "/: search")
			if m.Data.searchActive {
				if m.Data.totalMatches > 0 {
					shortcuts = append(shortcuts, "n/N: next/prev match")
				}
				shortcuts = append(shortcuts, "Ctrl+L: clear search")
			}
			shortcuts = append(shortcuts, "?: help")

			shortcutText := strings.Join(shortcuts, " | ")
			statusText = fmt.Sprintf("[Tasks] %s %s | %s", connectionStatus, statusInfo, shortcutText)
		}

	case "Details":
		// Right panel active - show current position, scroll info and connection status
		position := m.GetCurrentPosition()
		scrollPos := m.GetScrollPosition()
		connectionStatus := m.GetConnectionStatusText()

		if scrollPos == "Top" {
			statusText = fmt.Sprintf("[Details] %s %s | ?: help", connectionStatus, position)
		} else {
			statusText = fmt.Sprintf("[Details] %s %s • %s | ?: help", connectionStatus, position, scrollPos)
		}

	default:
		// Fallback to simple status
		statusText = fmt.Sprintf("[%s] Ready | ?: help | q: quit", activePanel)
	}

	return CreateStatusBarStyle("ready").Render(statusText)
}
