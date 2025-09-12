package ui

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/yousfisaad/lazyarchon/internal/archon"
	"os"
	"strings"
	"time"
)

// SetSelectedProject sets the currently selected project
func (m *Model) SetSelectedProject(projectID *string) {
	m.Data.selectedProjectID = projectID
	m.Navigation.selectedIndex = 0  // Reset task selection
	m.taskDetailsViewport.GotoTop() // Reset scroll
}

// CycleSortMode cycles to the next sort mode
func (m *Model) CycleSortMode() {
	m.Data.sortMode = (m.Data.sortMode + 1) % 4
	m.Navigation.selectedIndex = 0  // Reset selection to top
	m.taskDetailsViewport.GotoTop() // Reset scroll
}

// CycleSortModePrevious cycles to the previous sort mode
func (m *Model) CycleSortModePrevious() {
	m.Data.sortMode = (m.Data.sortMode + 3) % 4 // +3 for wrap-around
	m.Navigation.selectedIndex = 0              // Reset selection to top
	m.taskDetailsViewport.GotoTop()             // Reset scroll
}

// SetError sets the error state
func (m *Model) SetError(err string) {
	m.Data.loading = false
	m.Data.error = err
}

// ClearError clears the error state
func (m *Model) ClearError() {
	m.Data.error = ""
}

// SetLoading sets the loading state
func (m *Model) SetLoading(loading bool) {
	m.Data.loading = loading
	if loading {
		m.ClearError()
	}
}

// UpdateTasks updates the task list and adjusts selection bounds
func (m *Model) UpdateTasks(tasks []archon.Task) {
	m.Data.loading = false
	m.Data.tasks = tasks
	m.ClearError()

	// Apply sorting and adjust selectedIndex if needed
	sortedTasks := m.GetSortedTasks()
	if len(sortedTasks) > 0 && m.Navigation.selectedIndex >= len(sortedTasks) {
		m.setSelectedTask(len(sortedTasks) - 1)
	} else if len(sortedTasks) == 0 {
		m.setSelectedTask(0)
	} else {
		// Tasks updated but selection still valid - update viewport content
		m.updateTaskDetailsViewport()
	}
}

// UpdateProjects updates the project list and validates current selection
func (m *Model) UpdateProjects(projects []archon.Project) {
	m.Data.projects = projects

	// Reset project selection if selected project no longer exists
	if m.Data.selectedProjectID != nil {
		projectExists := false
		for _, project := range m.Data.projects {
			if project.ID == *m.Data.selectedProjectID {
				projectExists = true
				break
			}
		}
		if !projectExists {
			m.Data.selectedProjectID = nil // Reset to "All tasks"
		}
	}
}

// HandleGenericScroll handles scrolling for different UI contexts
// NOTE: Help modal and task details now use viewport, only list scrolling remains
func (m *Model) HandleGenericScroll(context ScrollContext, direction int, scrollType string) {
	switch context {
	case ListContext:
		m.handleListContextScroll(direction, scrollType)
		// HelpContext and DetailsContext are now handled by viewport, no longer needed
	}
}

// handleListContextScroll handles scrolling for task list
func (m *Model) handleListContextScroll(direction int, scrollType string) {
	sortedTasks := m.GetSortedTasks()
	if len(sortedTasks) == 0 {
		return
	}

	switch scrollType {
	case "line":
		newIndex := m.Navigation.selectedIndex + direction
		m.setSelectedTask(newIndex)
	case "fast":
		fastScrollLines := 4
		newIndex := m.Navigation.selectedIndex + (direction * fastScrollLines)
		m.setSelectedTask(newIndex)
	case "halfpage":
		halfPage := m.getTaskListHalfPageSize()
		newIndex := m.Navigation.selectedIndex + (direction * halfPage)
		m.setSelectedTask(newIndex)
	case "top":
		m.setSelectedTask(0)
	case "bottom":
		m.setSelectedTask(len(sortedTasks) - 1)
	}
}

// Note: getTaskListHalfPageSize is defined in input_navigation.go to avoid duplication

// getHalfPageSize calculates half-page size for detail panel scrolling
func (m *Model) getHalfPageSize() int {
	contentHeight := m.GetContentHeight()
	halfPage := (contentHeight - 4) / 2 // Account for border and padding
	if halfPage < 1 {
		halfPage = 1
	}
	return halfPage
}

// validateScrollOffset ensures scroll offset is within valid bounds (0 to maxScroll)
func (m *Model) validateScrollOffset(offset, contentLines, viewportLines int) int {
	if offset < 0 {
		return 0
	}

	// Calculate maximum scroll: total lines - viewport lines
	// If content fits entirely, max scroll is 0
	if contentLines <= viewportLines {
		return 0
	}

	maxScroll := contentLines - viewportLines
	if offset > maxScroll {
		return maxScroll
	}

	return offset
}

// calculateDetailContentLength estimates the content length for task details
// Uses conservative estimates to account for markdown rendering and formatting differences
func (m *Model) calculateDetailContentLength() int {
	sortedTasks := m.GetSortedTasks()
	if len(sortedTasks) == 0 || m.Navigation.selectedIndex >= len(sortedTasks) {
		return 20 // Default minimum for empty content
	}

	task := sortedTasks[m.Navigation.selectedIndex]
	width := m.GetRightPanelWidth() // Use same width as view for word wrapping

	// Conservative estimate with buffers for formatting differences
	var estimatedLines int

	// Fixed elements (headers, metadata, timestamps)
	estimatedLines += 8 // Task Details header, title header, status/assignee, timestamps, spacing

	// Title (word wrapped)
	titleLines := len(strings.Split(m.wordWrap(task.Title, width-6), "\n"))
	estimatedLines += titleLines

	// Feature tag if present
	if task.Feature != nil && *task.Feature != "" {
		estimatedLines += 1
	}

	// Description (with generous buffer for markdown)
	if task.Description != "" {
		descriptionLength := len(task.Description)
		// Rough estimate: chars per line with buffer for markdown formatting
		charsPerLine := width - 10 // Conservative line width
		if charsPerLine < 40 {
			charsPerLine = 40 // Minimum reasonable line width
		}

		basicLines := (descriptionLength / charsPerLine) + 1
		// Add 50% buffer for markdown formatting, code blocks, lists, etc.
		estimatedLines += int(float64(basicLines)*1.5) + 2 // +2 for header and spacing
	}

	// Sources
	if len(task.Sources) > 0 {
		estimatedLines += len(task.Sources) + 2 // +2 for header and spacing
	}

	// Code Examples
	if len(task.CodeExamples) > 0 {
		estimatedLines += len(task.CodeExamples) + 2 // +2 for header and spacing
	}

	// Add conservative buffer for any missing elements
	estimatedLines = int(float64(estimatedLines) * 1.1) // 10% additional buffer

	// Ensure minimum reasonable content length
	if estimatedLines < 10 {
		estimatedLines = 10
	}

	return estimatedLines
}

// wordWrap wraps text to fit within specified width (helper for content calculation)
func (m *Model) wordWrap(text string, width int) string {
	if len(text) <= width {
		return text
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	currentLine := words[0]

	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	lines = append(lines, currentLine)

	return strings.Join(lines, "\n")
}

// debugLogScroll logs scroll operations for debugging overshoot issues with enhanced content info
func (m *Model) debugLogScroll(context, operation string, before, after, maxScroll int, prevented bool) {
	logFile, err := os.OpenFile("/tmp/lazyarchon_scroll.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return // Fail silently if can't log
	}
	defer logFile.Close()

	timestamp := time.Now().Format("15:04:05.000")
	preventedStr := ""
	if prevented {
		preventedStr = " [PREVENTED]"
	}

	// Add enhanced content length information for debugging mismatches
	var contentInfo string
	if context == "DETAILS" {
		contentLength := m.calculateDetailContentLength()
		viewportLines := m.GetContentHeight() - 4
		actualMaxScroll := contentLength - viewportLines
		if actualMaxScroll < 0 {
			actualMaxScroll = 0
		}

		// Flag potential content length mismatches
		mismatchFlag := ""
		if actualMaxScroll != maxScroll {
			mismatchFlag = " [CONTENT_MISMATCH]"
		}

		contentInfo = fmt.Sprintf(" | content:%d viewport:%d actualMax:%d%s",
			contentLength, viewportLines, actualMaxScroll, mismatchFlag)
	}

	logMsg := fmt.Sprintf("[%s] %s %s: %d → %d (max: %d)%s%s\n",
		timestamp, context, operation, before, after, maxScroll, preventedStr, contentInfo)
	logFile.WriteString(logMsg)
}

// validateDetailScrollAfterContentChange - REMOVED: Task details now use viewport component

// updateTaskDetailsViewport updates the viewport content with the currently selected task
// This builds the same rich content as the original renderTaskDetails with full styling
func (m *Model) updateTaskDetailsViewport() {
	sortedTasks := m.GetSortedTasks()
	if len(sortedTasks) == 0 || m.Navigation.selectedIndex >= len(sortedTasks) {
		m.taskDetailsViewport.SetContent("No task selected")
		return
	}

	task := sortedTasks[m.Navigation.selectedIndex]
	width := m.GetRightPanelWidth()

	// Build all content with original styling - exactly like renderTaskDetails
	var allContent []string
	allContent = append(allContent, DetailHeaderStyle.Render("Task Details"))
	allContent = append(allContent, "")

	// Title with proper styling
	allContent = append(allContent, DetailHeaderStyle.Render("Title:"))
	titleLines := strings.Split(wordWrap(task.Title, width-6), "\n")
	allContent = append(allContent, titleLines...)
	allContent = append(allContent, "")

	// Status and assignee with colors
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(task.GetStatusColor()))
	allContent = append(allContent, fmt.Sprintf("Status: %s %s",
		task.GetStatusSymbol(),
		statusStyle.Render(strings.ToUpper(task.Status))))
	allContent = append(allContent, fmt.Sprintf("Assignee: %s", task.Assignee))

	if task.Feature != nil && *task.Feature != "" {
		tagText := TagStyle.Render(fmt.Sprintf("#%s", *task.Feature))
		allContent = append(allContent, fmt.Sprintf("Tags: %s", tagText))
	}
	allContent = append(allContent, "")

	// Description with markdown rendering
	if task.Description != "" {
		allContent = append(allContent, DetailHeaderStyle.Render("Description:"))
		descriptionContent := renderMarkdown(task.Description, width-6)
		descriptionLines := strings.Split(descriptionContent, "\n")
		allContent = append(allContent, descriptionLines...)
		allContent = append(allContent, "")
	}

	// Timestamps
	allContent = append(allContent, fmt.Sprintf("Created: %s", task.CreatedAt.Format("2006-01-02 15:04")))
	allContent = append(allContent, fmt.Sprintf("Updated: %s", task.UpdatedAt.Format("2006-01-02 15:04")))

	// Sources (if any)
	if len(task.Sources) > 0 {
		allContent = append(allContent, "")
		allContent = append(allContent, DetailHeaderStyle.Render("Sources:"))
		for _, source := range task.Sources {
			allContent = append(allContent, fmt.Sprintf("• %s (%s)", source.URL, source.Type))
		}
	}

	// Code Examples (if any) - this was missing from the simplified version
	if len(task.CodeExamples) > 0 {
		allContent = append(allContent, "")
		allContent = append(allContent, DetailHeaderStyle.Render("Code Examples:"))
		for _, example := range task.CodeExamples {
			allContent = append(allContent, fmt.Sprintf("• %s - %s", example.File, example.Purpose))
		}
	}

	// Set the content in the viewport
	content := strings.Join(allContent, "\n")
	m.taskDetailsViewport.SetContent(content)
}

// updateHelpModalViewport updates the help modal viewport with current help content
// This builds the same rich content as the original getHelpContent with full styling
func (m *Model) updateHelpModalViewport() {
	// Build all content with original styling - exactly like getHelpContent
	var help []string

	// Title
	help = append(help, lipgloss.NewStyle().Bold(true).Render("LazyArchon Help"))
	help = append(help, "")

	// Panel Navigation section
	help = append(help, lipgloss.NewStyle().Bold(true).Render("Panel Navigation:"))
	help = append(help, "  h/l          Switch between panels")
	help = append(help, "  ↑/↓ or j/k   Navigate/scroll (1 line)")
	help = append(help, "  J/K          Fast scroll (4 lines)")
	help = append(help, "  Ctrl+u/d     Half-page scroll")
	help = append(help, "  gg/G         Jump to top/bottom")
	help = append(help, "  Home/End     Jump to start/end")
	help = append(help, "")

	// Project Management section
	help = append(help, lipgloss.NewStyle().Bold(true).Render("Project Management:"))
	help = append(help, "  p            Project selection mode")
	help = append(help, "  a            Show all tasks")
	help = append(help, "  Enter        Select project")
	help = append(help, "  Esc          Exit project mode")
	help = append(help, "")

	// Task List section (when left panel is active)
	help = append(help, lipgloss.NewStyle().Bold(true).Render("Task List (when Tasks panel active):"))
	help = append(help, "  s/S          Sort tasks by different criteria")
	help = append(help, "")

	// Task Management section
	help = append(help, lipgloss.NewStyle().Bold(true).Render("Task Management:"))
	help = append(help, "  t            Change task status (Todo/Doing/Review/Done)")
	help = append(help, "")

	// Application Controls section
	help = append(help, lipgloss.NewStyle().Bold(true).Render("Application Controls:"))
	help = append(help, "  r or F5      Refresh data from API")
	help = append(help, "  q            Quit application")
	help = append(help, "  ?            Toggle this help")
	help = append(help, "")

	// Visual Indicators section
	help = append(help, lipgloss.NewStyle().Bold(true).Render("Visual Indicators:"))
	help = append(help, "  Bright cyan border    Active panel")
	help = append(help, "  Dim gray border       Inactive panel")
	help = append(help, "  [Tasks]/[Details]     Active panel in status bar")
	help = append(help, "  ▓░ scroll bar         Position indicator")
	help = append(help, "")

	// Task Status Symbols section
	help = append(help, lipgloss.NewStyle().Bold(true).Render("Task Status Symbols:"))
	help = append(help, "  ○  Todo       Not started")
	help = append(help, "  ◐  Doing      In progress")
	help = append(help, "  ◉  Review     Under review")
	help = append(help, "  ●  Done       Completed")
	help = append(help, "")

	// Help Navigation section
	help = append(help, lipgloss.NewStyle().Bold(true).Render("Help Navigation:"))
	help = append(help, "  j/k          Scroll help (1 line)")
	help = append(help, "  J/K          Fast scroll help (4 lines)")
	help = append(help, "  Ctrl+u/d     Half-page scroll help")
	help = append(help, "  gg/G         Jump to help top/bottom")
	help = append(help, "")

	// Footer
	help = append(help, lipgloss.NewStyle().Italic(true).Render("Press ? or ESC to close this help"))

	// Set the content in the viewport
	content := strings.Join(help, "\n")
	m.helpModalViewport.SetContent(content)
}

// setSelectedTask sets the selected task index and updates viewport content
// This ensures viewport content is always in sync with the selected task
func (m *Model) setSelectedTask(index int) {
	sortedTasks := m.GetSortedTasks()

	// Bounds check the index
	if index < 0 {
		index = 0
	} else if index >= len(sortedTasks) {
		if len(sortedTasks) > 0 {
			index = len(sortedTasks) - 1
		} else {
			index = 0
		}
	}

	// Only update if actually changing
	if m.Navigation.selectedIndex != index {
		m.Navigation.selectedIndex = index
		m.taskDetailsViewport.GotoTop() // Reset scroll when changing tasks
		m.updateTaskDetailsViewport()   // Update viewport content
	}
}

// ToggleFeature toggles a feature on/off in the selection
func (m *Model) ToggleFeature(featureName string) {
	if m.Modals.featureMode.selectedFeatures == nil {
		m.Modals.featureMode.selectedFeatures = make(map[string]bool)
	}

	// Toggle the feature
	currentState, exists := m.Modals.featureMode.selectedFeatures[featureName]
	if !exists {
		// If feature doesn't exist in map, it was implicitly enabled, so disable it
		m.Modals.featureMode.selectedFeatures[featureName] = false
	} else {
		m.Modals.featureMode.selectedFeatures[featureName] = !currentState
	}

	// Reset task selection since filtering changed
	m.Navigation.selectedIndex = 0
	m.taskDetailsViewport.GotoTop()
	m.updateTaskDetailsViewport()
}

// SelectAllFeatures enables all available features
func (m *Model) SelectAllFeatures() {
	availableFeatures := m.GetUniqueFeatures()
	if m.Modals.featureMode.selectedFeatures == nil {
		m.Modals.featureMode.selectedFeatures = make(map[string]bool)
	}

	for _, feature := range availableFeatures {
		m.Modals.featureMode.selectedFeatures[feature] = true
	}

	// Reset task selection since filtering changed
	m.Navigation.selectedIndex = 0
	m.taskDetailsViewport.GotoTop()
	m.updateTaskDetailsViewport()
}

// SelectNoFeatures disables all features
func (m *Model) SelectNoFeatures() {
	availableFeatures := m.GetUniqueFeatures()
	if m.Modals.featureMode.selectedFeatures == nil {
		m.Modals.featureMode.selectedFeatures = make(map[string]bool)
	}

	for _, feature := range availableFeatures {
		m.Modals.featureMode.selectedFeatures[feature] = false
	}

	// Reset task selection since filtering changed
	m.Navigation.selectedIndex = 0
	m.taskDetailsViewport.GotoTop()
	m.updateTaskDetailsViewport()
}

// backupFeatureState saves the current feature selection state for cancel functionality
func (m *Model) backupFeatureState() {
	if m.Modals.featureMode.selectedFeatures == nil {
		m.Modals.featureMode.backupFeatures = nil
		return
	}

	// Deep copy the current state
	m.Modals.featureMode.backupFeatures = make(map[string]bool, len(m.Modals.featureMode.selectedFeatures))
	for feature, enabled := range m.Modals.featureMode.selectedFeatures {
		m.Modals.featureMode.backupFeatures[feature] = enabled
	}
}

// restoreFeatureState restores the backup feature selection state (for cancel)
func (m *Model) restoreFeatureState() {
	if m.Modals.featureMode.backupFeatures == nil {
		m.Modals.featureMode.selectedFeatures = nil
	} else {
		// Deep copy the backup state back
		m.Modals.featureMode.selectedFeatures = make(map[string]bool, len(m.Modals.featureMode.backupFeatures))
		for feature, enabled := range m.Modals.featureMode.backupFeatures {
			m.Modals.featureMode.selectedFeatures[feature] = enabled
		}
	}

	// Reset task selection since filtering changed
	m.Navigation.selectedIndex = 0
	m.taskDetailsViewport.GotoTop()
	m.updateTaskDetailsViewport()
}
