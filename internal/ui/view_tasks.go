package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderTaskList renders the left panel with the list of tasks
func (m Model) renderTaskList(width, height int) string {
	listStyle := CreateActivePanelStyle(width, height, m.IsLeftPanelActive())

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

	var lines []string
	lines = append(lines, "Tasks:")
	lines = append(lines, "")

	maxLines := height - 6 // Account for border and padding
	start, end := calculateScrollWindow(len(sortedTasks), m.Navigation.selectedIndex, maxLines)

	// Generate scroll bar if needed
	scrollBar := renderScrollBar(start, len(sortedTasks), maxLines)
	hasScrollBar := scrollBar != nil

	// Adjust available width for task text if scroll bar is present
	textWidth := width - 8 // Base padding
	if hasScrollBar {
		textWidth -= 2 // Make room for scroll bar
	}

	for i := start; i < end; i++ {
		task := sortedTasks[i]

		// Format task line
		symbol := task.GetStatusSymbol()
		statusColor := task.GetStatusColor()

		// Apply search highlighting to title if search is active
		title := task.Title
		if m.Data.searchActive && m.Data.searchQuery != "" {
			title = highlightSearchTerms(task.Title, m.Data.searchQuery)
		}

		baseLine := fmt.Sprintf("%s %s", symbol, title)

		// Add feature/tag if present and space allows
		var line string
		if task.Feature != nil && *task.Feature != "" {
			tagText := fmt.Sprintf(" #%s", *task.Feature)
			// Check if we have enough space for the tag (minimum 15 chars for readability)
			if len(baseLine)+len(tagText) <= textWidth {
				// We have space - render the tag with styling
				styledTag := TagStyle.Render(tagText)
				line = baseLine + styledTag
			} else {
				// Not enough space - show just the base line
				line = baseLine
			}
		} else {
			line = baseLine
		}

		// Truncate if still too long
		if len(line) > textWidth {
			line = line[:textWidth-3] + "..."
		}

		// Style based on selection
		style := CreateTaskItemStyle(i == m.Navigation.selectedIndex, statusColor)

		if i == m.Navigation.selectedIndex {
			line = SelectionIndicator + line
		} else {
			line = NoSelection + line
		}

		lines = append(lines, style.Render(line))
	}

	// Show enhanced scrolling indicator if needed
	if len(sortedTasks) > maxLines {
		lines = append(lines, "")

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

		lines = append(lines, positionInfo)
	}

	// Combine task list with scroll bar if present
	taskContent := strings.Join(lines, "\n")
	if hasScrollBar {
		// Create task list without border for horizontal joining
		taskListStyle := CreateActivePanelStyle(width-2, height, m.IsLeftPanelActive())
		taskPanel := taskListStyle.Render(taskContent)

		// Create scroll bar panel
		scrollContent := strings.Join(scrollBar, "\n")
		scrollStyle := CreateScrollBarStyle(2, len(scrollBar))
		scrollPanel := scrollStyle.Render(scrollContent)

		// Join horizontally
		return lipgloss.JoinHorizontal(lipgloss.Top, taskPanel, scrollPanel)
	}

	return listStyle.Render(taskContent)
}

// renderTaskDetails renders the right panel with detailed task information and scrolling support
func (m Model) renderTaskDetails(width, height int) string {
	detailStyle := CreateActivePanelStyle(width, height, m.IsRightPanelActive())

	// The viewport content is managed by the model and updated when tasks change
	// We just need to render the viewport view
	viewportContent := m.taskDetailsViewport.View()

	return detailStyle.Render(viewportContent)
}
