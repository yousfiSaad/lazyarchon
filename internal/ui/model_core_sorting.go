package ui

import (
	"github.com/yousfisaad/lazyarchon/internal/ui/sorting"
	"github.com/yousfisaad/lazyarchon/internal/archon"
)

// GetSortedTasks returns the tasks sorted according to the current sort mode
// This method applies project, search, and feature filtering before sorting
func (m Model) GetSortedTasks() []archon.Task {
	filteredTasks := m.Data.tasks

	// Apply project filter first (if any)
	if m.Data.selectedProjectID != nil {
		var projectFilteredTasks []archon.Task
		for _, task := range filteredTasks {
			if task.ProjectID == *m.Data.selectedProjectID {
				projectFilteredTasks = append(projectFilteredTasks, task)
			}
		}
		filteredTasks = projectFilteredTasks
	}

	// Note: Search highlighting is now handled in rendering instead of filtering
	// This allows users to see all tasks with matches highlighted

	// Apply custom status filters (if active)
	if m.Data.statusFilterActive && m.Data.statusFilters != nil {
		var statusFilteredTasks []archon.Task
		for _, task := range filteredTasks {
			if enabled, exists := m.Data.statusFilters[task.Status]; exists && enabled {
				statusFilteredTasks = append(statusFilteredTasks, task)
			}
		}
		filteredTasks = statusFilteredTasks
	} else {
		// Apply completed tasks filter based on configuration (only if no custom status filtering)
		if !m.config.ShouldShowCompletedTasks() {
			var nonCompletedTasks []archon.Task
			for _, task := range filteredTasks {
				if task.Status != "done" {
					nonCompletedTasks = append(nonCompletedTasks, task)
				}
			}
			filteredTasks = nonCompletedTasks
		}
	}

	// Apply feature filter (if any features are explicitly disabled)
	if len(m.Modals.featureMode.selectedFeatures) > 0 {
		var featureFilteredTasks []archon.Task
		for _, task := range filteredTasks {
			// Include task if:
			// 1. It has no feature (null/empty), OR
			// 2. Its feature is enabled in selectedFeatures
			if task.Feature == nil || *task.Feature == "" {
				// Tasks without features are always shown
				featureFilteredTasks = append(featureFilteredTasks, task)
			} else if enabled, exists := m.Modals.featureMode.selectedFeatures[*task.Feature]; exists && enabled {
				featureFilteredTasks = append(featureFilteredTasks, task)
			}
		}
		filteredTasks = featureFilteredTasks
	}

	return sorting.SortTasks(filteredTasks, m.Data.sortMode)
}