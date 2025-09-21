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

// sortByStatusPriority sorts tasks by status first, then by priority
func sortByStatusPriority(tasks []archon.Task) {
	sort.Slice(tasks, func(i, j int) bool {
		// First, sort by status priority
		statusI := getStatusWeight(tasks[i].Status)
		statusJ := getStatusWeight(tasks[j].Status)
		if statusI != statusJ {
			return statusI < statusJ
		}
		// Then by task order (higher = more priority)
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

// TaskSorter provides a cached sorting interface for better performance
type TaskSorter struct {
	tasks        []archon.Task
	sortMode     int
	cachedResult []archon.Task
	cacheValid   bool
}

// NewTaskSorter creates a new TaskSorter
func NewTaskSorter(tasks []archon.Task, sortMode int) *TaskSorter {
	return &TaskSorter{
		tasks:      tasks,
		sortMode:   sortMode,
		cacheValid: false,
	}
}

// GetSorted returns the sorted tasks, using cache if valid
func (ts *TaskSorter) GetSorted() []archon.Task {
	if !ts.cacheValid {
		ts.cachedResult = SortTasks(ts.tasks, ts.sortMode)
		ts.cacheValid = true
	}
	return ts.cachedResult
}

// UpdateTasks updates the tasks and invalidates the cache
func (ts *TaskSorter) UpdateTasks(tasks []archon.Task) {
	ts.tasks = tasks
	ts.cacheValid = false
}

// UpdateSortMode updates the sort mode and invalidates the cache
func (ts *TaskSorter) UpdateSortMode(sortMode int) {
	ts.sortMode = sortMode
	ts.cacheValid = false
}

// InvalidateCache forces a re-sort on the next GetSorted call
func (ts *TaskSorter) InvalidateCache() {
	ts.cacheValid = false
}
