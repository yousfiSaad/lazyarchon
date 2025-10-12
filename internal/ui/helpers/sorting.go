package helpers

import (
	"github.com/yousfisaad/lazyarchon/internal/archon"
	"github.com/yousfisaad/lazyarchon/internal/ui/sorting"
)

// TaskFilters holds all filter parameters for task lists
type TaskFilters struct {
	ProjectID          *string
	StatusFilters      map[string]bool
	StatusFilterActive bool
	FeatureFilters     map[string]bool
	ShowCompletedTasks bool
}

// FilterAndSortTasks applies all filters and sorts tasks
// This is a pure function that replaces SortingCoordinator.GetSortedTasks()
func FilterAndSortTasks(tasks []archon.Task, sortMode int, filters TaskFilters) []archon.Task {
	filteredTasks := tasks

	// Apply project filter first (if any)
	if filters.ProjectID != nil {
		var projectFilteredTasks []archon.Task
		for _, task := range filteredTasks {
			if task.ProjectID == *filters.ProjectID {
				projectFilteredTasks = append(projectFilteredTasks, task)
			}
		}
		filteredTasks = projectFilteredTasks
	}

	// Apply custom status filters (if active)
	if filters.StatusFilterActive && filters.StatusFilters != nil {
		var statusFilteredTasks []archon.Task
		for _, task := range filteredTasks {
			if enabled, exists := filters.StatusFilters[task.Status]; exists && enabled {
				statusFilteredTasks = append(statusFilteredTasks, task)
			}
		}
		filteredTasks = statusFilteredTasks
	} else if !filters.ShowCompletedTasks {
		// Apply completed tasks filter based on configuration (only if no custom status filtering)
		var nonCompletedTasks []archon.Task
		for _, task := range filteredTasks {
			if task.Status != archon.TaskStatusDone {
				nonCompletedTasks = append(nonCompletedTasks, task)
			}
		}
		filteredTasks = nonCompletedTasks
	}

	// Apply feature filter (three-state logic):
	// - nil: No filter active, show all tasks
	// - empty map {}: Filter active, nothing selected, show NO tasks
	// - populated map: Filter active, show selected features
	if filters.FeatureFilters != nil {
		var featureFilteredTasks []archon.Task
		for _, task := range filteredTasks {
			// Include task if:
			// 1. It has no feature (null/empty), OR
			// 2. Its feature is enabled in featureFilters
			if task.Feature == nil || *task.Feature == "" {
				// Tasks without features are always shown
				featureFilteredTasks = append(featureFilteredTasks, task)
			} else if enabled, exists := filters.FeatureFilters[*task.Feature]; exists && enabled {
				featureFilteredTasks = append(featureFilteredTasks, task)
			}
		}
		filteredTasks = featureFilteredTasks
	}

	return sorting.SortTasks(filteredTasks, sortMode)
}
