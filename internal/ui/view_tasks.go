package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yousfisaad/lazyarchon/internal/ui/styling"
)


// renderTaskList renders the left panel with the list of tasks
func (m Model) renderTaskList(width, height int) string {
	// Create style context for panel styling
	styleContext := m.CreateStyleContext(false)
	factory := styleContext.Factory()
	listStyle := factory.Panel(width, height, m.IsLeftPanelActive())

	if m.Data.loading {
		return listStyle.Render("Loading tasks...")
	}

	if m.Data.error != "" {
		return listStyle.Render(fmt.Sprintf("Error: %s\n\nPress 'r' to retry", m.Data.error))
	}

	if len(m.Data.tasks) == 0 {
		return listStyle.Render("No tasks found")
	}

	// Apply current sorting to tasks
	sortedTasks := m.GetSortedTasks()

	maxLines := height - 6 // Account for border and padding
	start, end := calculateScrollWindow(len(sortedTasks), m.Navigation.selectedIndex, maxLines)

	// Generate scroll bar if needed
	scrollBar := renderScrollBar(start, len(sortedTasks), maxLines)
	hasScrollBar := scrollBar != nil

	// Calculate available width for task content more accurately
	// Only subtract actual panel chrome (borders and padding)
	contentWidth := width - 8 // Panel borders and padding
	if hasScrollBar {
		contentWidth -= 2 // Scroll bar width
	}

	var lines []string
	lines = append(lines, styling.RenderLine("Tasks:", contentWidth))
	lines = append(lines, styling.RenderLine("", contentWidth))

	// Available width for task line content (before selection indicator)
	// Selection indicator is added outside the content width
	taskContentWidth := contentWidth - 2 // Selection indicator space ("► ")

	for i := start; i < end; i++ {
		task := sortedTasks[i]


		// Build task line using the new TaskLineBuilder with styling context
		isSelected := i == m.Navigation.selectedIndex
		styleContext := m.CreateStyleContext(isSelected)
		builder := styling.NewTaskLineBuilder(taskContentWidth, styleContext)

		// Assemble line components in logical order
		line := builder.
			AddPriorityIndicator(task).
			AddStatusIndicator(task).
			AddTitle(task, m.Data.searchQuery, m.Data.searchActive).
			AddFeatureTag(task).
			Build(m.Data.searchQuery, m.Data.searchActive)

		// Add selection indicator and ensure consistent width with headers
		if isSelected {
			line = styling.SelectionIndicator + line
		} else {
			line = styling.NoSelection + line
		}

		// Components now handle their own backgrounds for selected items
		line = styling.RenderLine(line, contentWidth)

		lines = append(lines, line)
	}

	// Show enhanced scrolling indicator if needed
	if len(sortedTasks) > maxLines {
		lines = append(lines, styling.RenderLine("", contentWidth))

		// Enhanced position feedback with percentage
		percentage := ((end * 100) / len(sortedTasks))
		if percentage > 100 {
			percentage = 100
		}

		positionInfo := fmt.Sprintf("Showing %d-%d of %d tasks (%d%%)",
			start+1, end, len(sortedTasks), percentage)

		// Add selected task position indicator
		selectedPos := m.Navigation.selectedIndex + 1
		positionInfo += fmt.Sprintf(" | Task %d selected", selectedPos)

		// Create style context for styling position info
		styleContext := m.CreateStyleContext(false)
		factory := styleContext.Factory()
		styledPositionInfo := factory.Text(styling.CurrentTheme.MutedColor).Render(positionInfo)
		lines = append(lines, styling.RenderLine(styledPositionInfo, contentWidth))
	}

	// Combine task list with scroll bar if present
	taskContent := strings.Join(lines, "\n")
	if hasScrollBar {
		// Create task list without border for horizontal joining
		taskListStyle := factory.Panel(width-2, height, m.IsLeftPanelActive())
		taskPanel := taskListStyle.Render(taskContent)

		// Create scroll bar panel
		scrollContent := strings.Join(scrollBar, "\n")
		scrollStyle := styling.CreateScrollBarStyle(2, len(scrollBar))
		scrollPanel := scrollStyle.Render(scrollContent)

		// Join horizontally
		return lipgloss.JoinHorizontal(lipgloss.Top, taskPanel, scrollPanel)
	}

	// Since lines are already fully styled with backgrounds, just add border/padding
	return listStyle.Render(taskContent)
}

// renderTaskDetails renders the right panel with detailed task information and scrolling support
func (m Model) renderTaskDetails(width, height int) string {
	// Create style context for panel styling - reuse from renderTaskList
	styleContext := m.CreateStyleContext(false)
	factory := styleContext.Factory()
	detailStyle := factory.Panel(width, height, m.IsRightPanelActive())

	// The viewport content is managed by the model and updated when tasks change
	// We just need to render the viewport view
	viewportContent := m.taskDetailsViewport.View()

	return detailStyle.Render(viewportContent)
}

// Note: truncateTaskLine function removed - replaced by TaskLineBuilder for cleaner, more efficient truncation
