package helpers

import (
	"fmt"
	"sort"

	"github.com/yousfisaad/lazyarchon/v2/internal/shared/interfaces"
)

// GetUniqueWorktrees returns a sorted list of unique worktrees from tasks
func GetUniqueWorktrees(tasks []interfaces.Task) []string {
	worktreeSet := make(map[string]bool)

	// Collect unique worktrees from task attribution
	for _, task := range tasks {
		if task.Worktree != "" {
			worktreeSet[task.Worktree] = true
		}
	}

	// Convert to sorted slice
	worktrees := make([]string, 0, len(worktreeSet))
	for worktree := range worktreeSet {
		worktrees = append(worktrees, worktree)
	}
	sort.Strings(worktrees)

	return worktrees
}

// GetWorktreeTaskCount returns the count of tasks for a specific worktree
func GetWorktreeTaskCount(tasks []interfaces.Task, worktree string) int {
	count := 0
	for _, task := range tasks {
		if task.Worktree == worktree {
			count++
		}
	}
	return count
}

// GetWorktreeFilterSummary returns a summary of active worktree filters
// Three-state logic:
// - nil map: No filter active (show all)
// - empty map {}: Filter active, nothing selected (unstamped tasks only)
// - populated map: Filter active with selections
func GetWorktreeFilterSummary(availableWorktrees []string, enabledWorktrees map[string]bool) string {
	if len(availableWorktrees) == 0 {
		return "No worktrees"
	}

	// nil = no filter active, show all
	if enabledWorktrees == nil {
		return "All worktrees"
	}

	enabledCount := 0
	var enabledList []string

	for _, worktree := range availableWorktrees {
		if enabled, exists := enabledWorktrees[worktree]; exists && enabled {
			enabledCount++
			enabledList = append(enabledList, worktree)
		}
	}

	totalWorktrees := len(availableWorktrees)

	switch enabledCount {
	case 0:
		return "No worktrees" // Empty map = explicitly deselected all
	case totalWorktrees:
		return "All worktrees"
	case 1:
		return fmt.Sprintf("@%s only", enabledList[0])
	default:
		return fmt.Sprintf("%d/%d worktrees", enabledCount, totalWorktrees)
	}
}
