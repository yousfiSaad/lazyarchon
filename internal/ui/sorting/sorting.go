package sorting

import (
	"sort"
	"strings"

	"github.com/yousfisaad/lazyarchon/v2/internal/shared/interfaces"
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
func SortTasks(tasks []interfaces.Task, sortMode int) []interfaces.Task {
	if len(tasks) == 0 {
		return tasks
	}

	// Make a copy to avoid modifying the original slice
	sortedTasks := make([]interfaces.Task, len(tasks))
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
func sortByStatusPriority(tasks []interfaces.Task) {
	sort.Slice(tasks, func(i, j int) bool { //nolint:varnamelen // i, j are idiomatic for sort functions
		// First, sort by status priority
		statusI := getStatusWeight(tasks[i].Status)
		statusJ := getStatusWeight(tasks[j].Status)
		if statusI != statusJ {
			return statusI < statusJ
		}

		// Within same status:
		// - For 'done' tasks: sort by UpdatedAt (most recent first)
		// - For all other statuses: sort by priority (lower value = more urgent)
		if tasks[i].Status == interfaces.StatusDone {
			return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt)
		}
		return tasks[i].Priority < tasks[j].Priority
	})
}

// sortByPriority sorts tasks by priority only (lower value = more urgent)
func sortByPriority(tasks []interfaces.Task) {
	sort.Slice(tasks, func(i, j int) bool { //nolint:varnamelen // i, j are idiomatic for sort functions
		return tasks[i].Priority < tasks[j].Priority
	})
}

// sortByTimeCreated sorts tasks by creation time (newest first)
func sortByTimeCreated(tasks []interfaces.Task) {
	sort.Slice(tasks, func(i, j int) bool { //nolint:varnamelen // i, j are idiomatic for sort functions
		return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
	})
}

// sortByAlphabetical sorts tasks alphabetically by title
func sortByAlphabetical(tasks []interfaces.Task) {
	sort.Slice(tasks, func(i, j int) bool { //nolint:varnamelen // i, j are idiomatic for sort functions
		return strings.ToLower(tasks[i].Title) < strings.ToLower(tasks[j].Title)
	})
}

// getStatusWeight returns the priority weight for a task status
// Lower numbers = higher priority (appear first)
func getStatusWeight(status string) int {
	switch status {
	case interfaces.StatusTodo:
		return 0 // Highest priority - needs action
	case interfaces.StatusDoing:
		return 1 // Second priority - work in progress
	case interfaces.StatusReview:
		return 2 // Third priority - waiting for review
	case interfaces.StatusDone:
		return 3 // Lowest priority - completed
	default:
		return 4 // Unknown status goes to end
	}
}
