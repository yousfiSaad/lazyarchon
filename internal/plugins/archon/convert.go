package archon

import "github.com/yousfisaad/lazyarchon/v2/internal/plugin"

// The generic plugin scale and Archon's legacy scale run in opposite
// directions: generic priority counts down (1 = critical, 5 = backlog) while
// Archon's task_order counts up (0-100, higher = more urgent, with >=80
// counting as high and >=50 as medium on the server). These helpers are the
// only place the two scales meet — values cross the adapter edge exclusively
// through them.

// priorityToTaskOrder maps the generic 1-5 scale onto Archon's legacy
// task_order. Monotonic, and each value lands inside the server tier that
// matches its label (1-2 high, 3 medium, 4-5 low).
func priorityToTaskOrder(priority int) int {
	switch priority {
	case plugin.PriorityCritical:
		return 95
	case plugin.PriorityHigh:
		return 85
	case plugin.PriorityMedium:
		return 60
	case plugin.PriorityLow:
		return 40
	case plugin.PriorityBacklog:
		return 10
	default:
		// Out-of-range values (including the 0 "unset" default) fall back
		// to the server's own default band.
		return 50
	}
}

// taskOrderToPriority is the lossy inverse: Archon only distinguishes three
// tiers, so legacy orders collapse onto high/medium/low. The critical and
// backlog tiers do not survive a round-trip.
func taskOrderToPriority(taskOrder int) int {
	switch {
	case taskOrder >= 80:
		return plugin.PriorityHigh
	case taskOrder >= 50:
		return plugin.PriorityMedium
	default:
		return plugin.PriorityLow
	}
}
