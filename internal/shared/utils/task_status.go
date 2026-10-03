package utils

import (
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/interfaces"
)

// TaskStatusUtils provides utility functions for task status operations
type TaskStatusUtils struct{}

// NewTaskStatusUtils creates a new instance of TaskStatusUtils
func NewTaskStatusUtils() *TaskStatusUtils {
	return &TaskStatusUtils{}
}

// GetStatusIndex returns the index (0-3) for a given status string
// This is useful for UI components that need to map status to array indices
func (u *TaskStatusUtils) GetStatusIndex(status string) int {
	switch status {
	case interfaces.StatusTodo:
		return 0
	case interfaces.StatusDoing:
		return 1
	case interfaces.StatusReview:
		return 2
	case interfaces.StatusDone:
		return 3
	default:
		return 0 // Default to todo
	}
}

// GetStatusFromIndex returns the status string for a given index (0-3)
// This is useful for UI components that work with indexed status arrays
func (u *TaskStatusUtils) GetStatusFromIndex(index int) string {
	switch index {
	case 0:
		return interfaces.StatusTodo
	case 1:
		return interfaces.StatusDoing
	case 2:
		return interfaces.StatusReview
	case 3:
		return interfaces.StatusDone
	default:
		return interfaces.StatusTodo
	}
}

// GetAllStatuses returns all valid task statuses in order
func (u *TaskStatusUtils) GetAllStatuses() []string {
	return []string{
		interfaces.StatusTodo,
		interfaces.StatusDoing,
		interfaces.StatusReview,
		interfaces.StatusDone,
	}
}

// IsValidStatus checks if a status string is valid
func (u *TaskStatusUtils) IsValidStatus(status string) bool {
	switch status {
	case interfaces.StatusTodo, interfaces.StatusDoing, interfaces.StatusReview, interfaces.StatusDone:
		return true
	default:
		return false
	}
}
