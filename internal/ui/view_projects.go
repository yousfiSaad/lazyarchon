package ui

import (
	"fmt"
	"strings"

	"github.com/yousfisaad/lazyarchon/internal/ui/styling"
)

// renderProjectList renders the left panel with the list of projects
func (m Model) renderProjectList(width, height int) string {
	// Create style context for panel styling - always active in project mode
	styleContext := m.CreateStyleContext(false) // No selection highlighting for panel itself
	factory := styleContext.Factory()
	listStyle := factory.Panel(width, height, true) // Always active in project mode

	if m.Data.loading {
		return listStyle.Render("Loading projects...")
	}

	if len(m.Data.projects) == 0 {
		return listStyle.Render("No projects found")
	}

	var lines []string
	lines = append(lines, "Projects:")
	lines = append(lines, "")

	// Add projects with task counts
	for i, project := range m.Data.projects {
		taskCount := m.GetTaskCountForProject(project.ID)
		line := fmt.Sprintf("%s (%d)", project.Title, taskCount)
		if len(line) > width-8 {
			line = line[:width-11] + "..."
		}

		// Create style context with selection state for this project item
		isSelected := i == m.Modals.projectMode.selectedIndex
		itemStyleContext := m.CreateStyleContext(isSelected)
		itemFactory := itemStyleContext.Factory()

		// Style based on selection using factory
		style := itemFactory.ProjectItem(isSelected, false)

		if isSelected {
			line = styling.SelectionIndicator + line
		} else {
			line = styling.NoSelection + line
		}

		lines = append(lines, style.Render(line))
	}

	// Add "All Tasks" option at the end
	allTasksLine := "[All Tasks]"
	isAllTasksSelected := m.Modals.projectMode.selectedIndex == len(m.Data.projects)
	allTasksStyleContext := m.CreateStyleContext(isAllTasksSelected)
	allTasksFactory := allTasksStyleContext.Factory()
	allTasksStyle := allTasksFactory.ProjectItem(isAllTasksSelected, true)

	if isAllTasksSelected {
		allTasksLine = styling.SelectionIndicator + allTasksLine
	} else {
		allTasksLine = styling.NoSelection + allTasksLine
	}

	lines = append(lines, "")
	lines = append(lines, allTasksStyle.Render(allTasksLine))

	return listStyle.Render(strings.Join(lines, "\n"))
}

// renderProjectModeHelp renders the right panel with project mode instructions
func (m Model) renderProjectModeHelp(width, height int) string {
	// Create style context for panel styling - inactive in project mode
	styleContext := m.CreateStyleContext(false)
	factory := styleContext.Factory()
	helpStyle := factory.Panel(width, height, false) // Inactive in project mode

	var helpLines []string
	helpLines = append(helpLines, factory.Header().Render("Project Selection"))
	helpLines = append(helpLines, "")
	helpLines = append(helpLines, "Select a project to filter tasks, or choose")
	helpLines = append(helpLines, "[All Tasks] to view all tasks.")
	helpLines = append(helpLines, "")
	helpLines = append(helpLines, factory.Header().Render("Controls:"))
	helpLines = append(helpLines, "")
	helpLines = append(helpLines, "↑↓ or j/k    Navigate projects")
	helpLines = append(helpLines, "l or Enter   Select project")
	helpLines = append(helpLines, "h or Esc     Back to task view")
	helpLines = append(helpLines, "a            Show all tasks")
	helpLines = append(helpLines, "r            Refresh")
	helpLines = append(helpLines, "q            Quit")

	if len(m.Data.projects) > 0 && m.Modals.projectMode.selectedIndex < len(m.Data.projects) {
		project := m.Data.projects[m.Modals.projectMode.selectedIndex]
		helpLines = append(helpLines, "")
		helpLines = append(helpLines, factory.Header().Render("Selected Project:"))
		helpLines = append(helpLines, "")
		helpLines = append(helpLines, fmt.Sprintf("Name: %s", project.Title))
		if project.Description != "" {
			helpLines = append(helpLines, fmt.Sprintf("Description: %s", project.Description))
		}
	}

	return helpStyle.Render(strings.Join(helpLines, "\n"))
}
