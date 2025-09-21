package ui

import (
	"fmt"
	"github.com/yousfisaad/lazyarchon/internal/archon"
	"github.com/yousfisaad/lazyarchon/internal/ui/sorting"
)

// GetTasks returns the current tasks (implements interfaces.UIModel)
func (m Model) GetTasks() []archon.Task {
	return m.Data.tasks
}

// GetProjects returns the current projects (implements interfaces.UIModel)
func (m Model) GetProjects() []archon.Project {
	return m.Data.projects
}

// GetSelectedProjectID returns the currently selected project ID (implements interfaces.UIModel)
func (m Model) GetSelectedProjectID() *string {
	return m.Data.selectedProjectID
}

// IsLoading returns whether the model is currently loading data (implements interfaces.UIModel)
func (m Model) IsLoading() bool {
	return m.Data.loading
}

// GetError returns the current error message (implements interfaces.UIModel)
func (m Model) GetError() string {
	return m.Data.error
}

// IsProjectSelected returns true if a specific project is currently selected
func (m Model) IsProjectSelected() bool {
	return m.Data.selectedProjectID != nil
}

// GetSelectedProject returns the currently selected project, if any
func (m Model) GetSelectedProject() *archon.Project {
	if !m.IsProjectSelected() {
		return nil
	}

	for _, project := range m.Data.projects {
		if project.ID == *m.Data.selectedProjectID {
			return &project
		}
	}
	return nil
}

// GetCurrentProjectName returns the name of the current project or "All Tasks"
func (m Model) GetCurrentProjectName() string {
	if selectedProject := m.GetSelectedProject(); selectedProject != nil {
		return selectedProject.Title
	}
	return "All Tasks"
}

// GetTaskStatusCounts returns counts of tasks by status
func (m Model) GetTaskStatusCounts() (int, int, int, int) {
	var todo, doing, review, done int

	for _, task := range m.Data.tasks {
		switch task.Status {
		case "todo":
			todo++
		case "doing":
			doing++
		case "review":
			review++
		case "done":
			done++
		}
	}

	return todo, doing, review, done
}

// GetCurrentPosition returns position info for the current selection
func (m Model) GetCurrentPosition() string {
	sortedTasks := m.GetSortedTasks()
	if len(sortedTasks) == 0 {
		return "No tasks"
	}

	if m.Navigation.selectedIndex >= len(sortedTasks) {
		return "No selection"
	}

	return fmt.Sprintf("Task %d of %d", m.Navigation.selectedIndex+1, len(sortedTasks))
}

// GetCurrentSortModeName returns the human-readable name of the current sort mode
func (m Model) GetCurrentSortModeName() string {
	switch m.Data.sortMode {
	case sorting.SortStatusPriority:
		return "Status"
	case sorting.SortPriorityOnly:
		return "Priority"
	case sorting.SortTimeCreated:
		return "Created"
	case sorting.SortAlphabetical:
		return "Alpha"
	default:
		return "Unknown"
	}
}

// GetScrollPosition returns scroll position as percentage for right panel
func (m Model) GetScrollPosition() string {
	// Use viewport scroll information
	if m.taskDetailsViewport.AtTop() {
		return "Top"
	} else if m.taskDetailsViewport.AtBottom() {
		return "Bottom"
	}
	return "Scrolled"
}

// GetTaskCountForProject returns the number of tasks for a specific project
func (m Model) GetTaskCountForProject(projectID string) int {
	count := 0
	for _, task := range m.Data.tasks {
		if task.ProjectID == projectID {
			count++
		}
	}
	return count
}