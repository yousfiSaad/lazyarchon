package helpers

import (
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/interfaces"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/sorting"
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
func FilterAndSortTasks(tasks []interfaces.Task, sortMode int, filters TaskFilters) []interfaces.Task {
	filteredTasks := tasks
	filteredTasks = applyProjectFilter(filteredTasks, filters.ProjectID)
	filteredTasks = applyStatusFilter(filteredTasks, filters)
	filteredTasks = applyFeatureFilter(filteredTasks, filters.FeatureFilters)
	return sorting.SortTasks(filteredTasks, sortMode)
}

// applyProjectFilter filters tasks by project ID
func applyProjectFilter(tasks []interfaces.Task, projectID *string) []interfaces.Task {
	if projectID == nil {
		return tasks
	}

	filtered := make([]interfaces.Task, 0, len(tasks))
	for _, task := range tasks {
		if task.ProjectID == *projectID {
			filtered = append(filtered, task)
		}
	}
	return filtered
}

// applyStatusFilter filters tasks by status
func applyStatusFilter(tasks []interfaces.Task, filters TaskFilters) []interfaces.Task {
	// Apply custom status filters (if active)
	if filters.StatusFilterActive && filters.StatusFilters != nil {
		filtered := make([]interfaces.Task, 0, len(tasks))
		for _, task := range tasks {
			if enabled, exists := filters.StatusFilters[task.Status]; exists && enabled {
				filtered = append(filtered, task)
			}
		}
		return filtered
	}

	// Apply completed tasks filter based on configuration (only if no custom status filtering)
	if !filters.ShowCompletedTasks {
		filtered := make([]interfaces.Task, 0, len(tasks))
		for _, task := range tasks {
			if task.Status != interfaces.StatusDone {
				filtered = append(filtered, task)
			}
		}
		return filtered
	}

	return tasks
}

// applyFeatureFilter filters tasks by feature
// - nil: No filter active, show all tasks
// - empty map {}: Filter active, nothing selected, show NO tasks
// - populated map: Filter active, show selected features
func applyFeatureFilter(tasks []interfaces.Task, featureFilters map[string]bool) []interfaces.Task {
	if featureFilters == nil {
		return tasks
	}

	filtered := make([]interfaces.Task, 0, len(tasks))
	for _, task := range tasks {
		// Include task if:
		// 1. It has no tags, OR
		// 2. Any of its tags is enabled in featureFilters
		if len(task.Tags) == 0 {
			// Tasks without tags are always shown
			filtered = append(filtered, task)
		} else {
			// Check if any tag is enabled
			for _, tag := range task.Tags {
				if enabled, exists := featureFilters[tag]; exists && enabled {
					filtered = append(filtered, task)
					break
				}
			}
		}
	}
	return filtered
}
