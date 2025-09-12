package ui

import (
	"fmt"
	"strings"
)

// renderProjectList renders the left panel with the list of projects
func (m Model) renderProjectList(width, height int) string {
	listStyle := CreateActivePanelStyle(width, height, true) // Always active in project mode

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
		// TODO: Get actual task count for each project
		// For now, we'll show a placeholder
		line := fmt.Sprintf("%s", project.Title)
		if len(line) > width-8 {
			line = line[:width-11] + "..."
		}

		// Style based on selection
		style := CreateProjectItemStyle(i == m.Modals.projectMode.selectedIndex, false)

		if i == m.Modals.projectMode.selectedIndex {
			line = SelectionIndicator + line
		} else {
			line = NoSelection + line
		}

		lines = append(lines, style.Render(line))
	}

	// Add "All Tasks" option at the end
	allTasksLine := "[All Tasks]"
	allTasksStyle := CreateProjectItemStyle(m.Modals.projectMode.selectedIndex == len(m.Data.projects), true)

	if m.Modals.projectMode.selectedIndex == len(m.Data.projects) {
		allTasksLine = SelectionIndicator + allTasksLine
	} else {
		allTasksLine = NoSelection + allTasksLine
	}

	lines = append(lines, "")
	lines = append(lines, allTasksStyle.Render(allTasksLine))

	return listStyle.Render(strings.Join(lines, "\n"))
}

// renderProjectModeHelp renders the right panel with project mode instructions
func (m Model) renderProjectModeHelp(width, height int) string {
	helpStyle := CreateActivePanelStyle(width, height, false) // Inactive in project mode

	var helpLines []string
	helpLines = append(helpLines, DetailHeaderStyle.Render("Project Selection"))
	helpLines = append(helpLines, "")
	helpLines = append(helpLines, "Select a project to filter tasks, or choose")
	helpLines = append(helpLines, "[All Tasks] to view all tasks.")
	helpLines = append(helpLines, "")
	helpLines = append(helpLines, DetailHeaderStyle.Render("Controls:"))
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
		helpLines = append(helpLines, DetailHeaderStyle.Render("Selected Project:"))
		helpLines = append(helpLines, "")
		helpLines = append(helpLines, fmt.Sprintf("Name: %s", project.Title))
		if project.Description != "" {
			helpLines = append(helpLines, fmt.Sprintf("Description: %s", project.Description))
		}
	}

	return helpStyle.Render(strings.Join(helpLines, "\n"))
}
