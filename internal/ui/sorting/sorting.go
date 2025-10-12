package sorting

import (
	"sort"
	"strings"

	"github.com/yousfisaad/lazyarchon/internal/archon"
)

// Sort mode constants
const (
	SortStatusPriority = 0 // Status + Priority (default)
	SortPriorityOnly   = 1 // Priority only
	SortTimeCreated    = 2 // Creation time (newest first)
	SortAlphabetical   = 3 // Alphabetical by title
)

// Sort mode names for UI display
var sortModeNames = []string{
	"status+priority",
	"priority",
	"time",
	"alphabetical",
}

// GetSortModeName returns the display name for a sort mode
func GetSortModeName(sortMode int) string {
	if sortMode >= 0 && sortMode < len(sortModeNames) {
		return sortModeNames[sortMode]
	}
	return "unknown"
}

// SortTasks sorts tasks based on the specified sort mode
func SortTasks(tasks []archon.Task, sortMode int) []archon.Task {
	if len(tasks) == 0 {
		return tasks
	}

	// Make a copy to avoid modifying the original slice
	sortedTasks := make([]archon.Task, len(tasks))
	copy(sortedTasks, tasks)

	switch sortMode {
	case SortStatusPriority:
		sortByStatusPriority(sortedTasks)
	case SortPriorityOnly:
		sortByPriority(sortedTasks)
	case SortTimeCreated:
		sortByTimeCreated(sortedTasks)
	case SortAlphabetical:
		sortByAlphabetical(sortedTasks)
	}

	return sortedTasks
}

// sortByStatusPriority sorts tasks by status first, then by priority or edit time
// - todo/review/doing tasks: sorted by priority (TaskOrder, higher first)
// - done tasks: sorted by edit time (UpdatedAt, most recent first)
func sortByStatusPriority(tasks []archon.Task) {
	sort.Slice(tasks, func(i, j int) bool {
		// First, sort by status priority
		statusI := getStatusWeight(tasks[i].Status)
		statusJ := getStatusWeight(tasks[j].Status)
		if statusI != statusJ {
			return statusI < statusJ
		}

		// Within same status:
		// - For 'done' tasks: sort by UpdatedAt (most recent first)
		// - For all other statuses: sort by priority (TaskOrder, higher first)
		if tasks[i].Status == archon.TaskStatusDone {
			return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt.Time)
		}
		return tasks[i].TaskOrder > tasks[j].TaskOrder
	})
}

// sortByPriority sorts tasks by priority only (TaskOrder)
func sortByPriority(tasks []archon.Task) {
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].TaskOrder > tasks[j].TaskOrder
	})
}

// sortByTimeCreated sorts tasks by creation time (newest first)
func sortByTimeCreated(tasks []archon.Task) {
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].CreatedAt.After(tasks[j].CreatedAt.Time)
	})
}

// sortByAlphabetical sorts tasks alphabetically by title
func sortByAlphabetical(tasks []archon.Task) {
	sort.Slice(tasks, func(i, j int) bool {
		return strings.ToLower(tasks[i].Title) < strings.ToLower(tasks[j].Title)
	})
}

// getStatusWeight returns the priority weight for a task status
// Lower numbers = higher priority (appear first)
func getStatusWeight(status string) int {
	switch status {
	case archon.TaskStatusTodo:
		return 0 // Highest priority - needs action
	case archon.TaskStatusDoing:
		return 1 // Second priority - work in progress
	case archon.TaskStatusReview:
		return 2 // Third priority - waiting for review
	case archon.TaskStatusDone:
		return 3 // Lowest priority - completed
	default:
		return 4 // Unknown status goes to end
	}
}
