package utils

import "github.com/yousfisaad/lazyarchon/internal/archon"

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
	case archon.TaskStatusTodo:
		return 0
	case archon.TaskStatusDoing:
		return 1
	case archon.TaskStatusReview:
		return 2
	case archon.TaskStatusDone:
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
		return archon.TaskStatusTodo
	case 1:
		return archon.TaskStatusDoing
	case 2:
		return archon.TaskStatusReview
	case 3:
		return archon.TaskStatusDone
	default:
		return archon.TaskStatusTodo
	}
}

// GetAllStatuses returns all valid task statuses in order
func (u *TaskStatusUtils) GetAllStatuses() []string {
	return []string{
		archon.TaskStatusTodo,
		archon.TaskStatusDoing,
		archon.TaskStatusReview,
		archon.TaskStatusDone,
	}
}

// IsValidStatus checks if a status string is valid
func (u *TaskStatusUtils) IsValidStatus(status string) bool {
	switch status {
	case archon.TaskStatusTodo, archon.TaskStatusDoing, archon.TaskStatusReview, archon.TaskStatusDone:
		return true
	default:
		return false
	}
}
