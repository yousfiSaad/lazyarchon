package helpers

import (
	"github.com/yousfisaad/lazyarchon/v2/internal/archon"
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
func FilterAndSortTasks(tasks []archon.Task, sortMode int, filters TaskFilters) []archon.Task {
	filteredTasks := tasks
	filteredTasks = applyProjectFilter(filteredTasks, filters.ProjectID)
	filteredTasks = applyStatusFilter(filteredTasks, filters)
	filteredTasks = applyFeatureFilter(filteredTasks, filters.FeatureFilters)
	return sorting.SortTasks(filteredTasks, sortMode)
}

// applyProjectFilter filters tasks by project ID
func applyProjectFilter(tasks []archon.Task, projectID *string) []archon.Task {
	if projectID == nil {
		return tasks
	}

	filtered := make([]archon.Task, 0, len(tasks))
	for _, task := range tasks {
		if task.ProjectID == *projectID {
			filtered = append(filtered, task)
		}
	}
	return filtered
}

// applyStatusFilter filters tasks by status
func applyStatusFilter(tasks []archon.Task, filters TaskFilters) []archon.Task {
	// Apply custom status filters (if active)
	if filters.StatusFilterActive && filters.StatusFilters != nil {
		filtered := make([]archon.Task, 0, len(tasks))
		for _, task := range tasks {
			if enabled, exists := filters.StatusFilters[task.Status]; exists && enabled {
				filtered = append(filtered, task)
			}
		}
		return filtered
	}

	// Apply completed tasks filter based on configuration (only if no custom status filtering)
	if !filters.ShowCompletedTasks {
		filtered := make([]archon.Task, 0, len(tasks))
		for _, task := range tasks {
			if task.Status != archon.TaskStatusDone {
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
func applyFeatureFilter(tasks []archon.Task, featureFilters map[string]bool) []archon.Task {
	if featureFilters == nil {
		return tasks
	}

	filtered := make([]archon.Task, 0, len(tasks))
	for _, task := range tasks {
		// Include task if:
		// 1. It has no feature (null/empty), OR
		// 2. Its feature is enabled in featureFilters
		if task.Feature == nil || *task.Feature == "" {
			// Tasks without features are always shown
			filtered = append(filtered, task)
		} else if enabled, exists := featureFilters[*task.Feature]; exists && enabled {
			filtered = append(filtered, task)
		}
	}
	return filtered
}
