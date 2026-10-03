package helpers

import (
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/deps"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/interfaces"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/sorting"
)

// TaskFilters holds all filter parameters for task lists
type TaskFilters struct {
	ProjectID          *string
	StatusFilters      map[string]bool
	StatusFilterActive bool
	FeatureFilters     map[string]bool
	WorktreeFilters    map[string]bool
	ShowCompletedTasks bool
	ShowArchived       bool
	Ready              bool
}

// FilterAndSortTasks applies all filters and sorts tasks
// This is a pure function that replaces SortingCoordinator.GetSortedTasks()
func FilterAndSortTasks(tasks []interfaces.Task, sortMode int, filters TaskFilters) []interfaces.Task {
	// Dependency levels come from the unfiltered slice, for the same reason
	// the ready filter runs first: a blocker hidden by a status or project
	// filter must still hold its dependents back a wave.
	var levels map[string]int
	if sortMode == sorting.SortPlan {
		levels = deps.Build(tasks).Levels()
	}

	// The ready filter runs first: it resolves blocker statuses from the
	// task slice itself, which must still be unfiltered — a done blocker
	// hidden by the completed-tasks filter must still count as finished.
	filteredTasks := tasks
	filteredTasks = applyReadyFilter(filteredTasks, filters.Ready)
	filteredTasks = applyProjectFilter(filteredTasks, filters.ProjectID)
	filteredTasks = applyStatusFilter(filteredTasks, filters)
	filteredTasks = applyFeatureFilter(filteredTasks, filters.FeatureFilters)
	filteredTasks = applyWorktreeFilter(filteredTasks, filters.WorktreeFilters)
	filteredTasks = applyArchivedFilter(filteredTasks, filters.ShowArchived)
	return sorting.SortTasksWithLevels(filteredTasks, sortMode, levels)
}

// applyReadyFilter keeps only tasks whose blockers are all done
//   - false: No filter active, show all tasks
//
// Missing blocker IDs count as unfinished, mirroring the local backend
// where archived-but-undone blockers still block.
func applyReadyFilter(tasks []interfaces.Task, ready bool) []interfaces.Task {
	if !ready {
		return tasks
	}

	statuses := make(map[string]string, len(tasks))
	for _, task := range tasks {
		statuses[task.ID] = task.Status
	}

	filtered := make([]interfaces.Task, 0, len(tasks))
	for _, task := range tasks {
		if isTaskReady(task, statuses) {
			filtered = append(filtered, task)
		}
	}
	return filtered
}

// isTaskReady reports whether every blocker of the task is done.
func isTaskReady(task interfaces.Task, statuses map[string]string) bool {
	for _, blockerID := range task.BlockedBy {
		if statuses[blockerID] != interfaces.StatusDone {
			return false
		}
	}
	return true
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
//   - nil: No filter active, show all tasks
//   - empty map {}: Filter active, nothing selected, show untagged tasks only
//   - populated map: Filter active, show tasks with any selected feature
//     (untagged tasks always pass through)
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

// applyWorktreeFilter filters tasks by worktree attribution
//   - nil: No filter active, show all tasks
//   - empty map {}: Filter active, nothing selected, show unstamped tasks only
//   - populated map: Filter active, show tasks stamped with a selected worktree
//     (unstamped tasks always pass through)
func applyWorktreeFilter(tasks []interfaces.Task, worktreeFilters map[string]bool) []interfaces.Task {
	if worktreeFilters == nil {
		return tasks
	}

	filtered := make([]interfaces.Task, 0, len(tasks))
	for _, task := range tasks {
		// Include task if:
		// 1. It has no worktree attribution, OR
		// 2. Its worktree is enabled in worktreeFilters
		if task.Worktree == "" {
			// Tasks without a worktree are always shown
			filtered = append(filtered, task)
		} else if enabled, exists := worktreeFilters[task.Worktree]; exists && enabled {
			filtered = append(filtered, task)
		}
	}
	return filtered
}

// applyArchivedFilter hides archived tasks unless ShowArchived is set. It runs
// last, after readiness and plan levels were resolved on the unfiltered slice,
// so an archived-but-undone blocker still holds its dependents back. A plain
// bool: unlike the feature/worktree maps there is no filter-active-nothing-
// selected state to keep distinct.
func applyArchivedFilter(tasks []interfaces.Task, showArchived bool) []interfaces.Task {
	if showArchived {
		return tasks
	}

	filtered := make([]interfaces.Task, 0, len(tasks))
	for _, task := range tasks {
		if !task.Archived {
			filtered = append(filtered, task)
		}
	}
	return filtered
}
