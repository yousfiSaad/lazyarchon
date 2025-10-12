package taskitem

import "github.com/yousfisaad/lazyarchon/v2/internal/archon"

// TaskItemUpdateMsg updates a specific task item's state
type TaskItemUpdateMsg struct {
	Index         int         // Index of the task item to update
	Task          archon.Task // Updated task data
	IsSelected    bool        // Whether this task is selected
	IsHighlighted bool        // Whether this task matches search
	SearchQuery   string      // Current search query for highlighting
}

// TaskItemResizeMsg updates a task item's available width
type TaskItemResizeMsg struct {
	Index int // Index of the task item to resize
	Width int // New available width
}

// TaskItemSelectionChangedMsg indicates a task item's selection state changed
type TaskItemSelectionChangedMsg struct {
	Index      int  // Index of the task item
	IsSelected bool // New selection state
}

// TaskItemClickedMsg indicates a task item was clicked/activated
type TaskItemClickedMsg struct {
	Index int         // Index of the clicked task item
	Task  archon.Task // The clicked task
}

// TaskItemHighlightChangedMsg indicates a task item's highlight state changed
type TaskItemHighlightChangedMsg struct {
	Index         int    // Index of the task item
	IsHighlighted bool   // New highlight state
	SearchQuery   string // Search query causing the highlight
}

// TaskItemDataChangedMsg indicates the underlying task data changed
type TaskItemDataChangedMsg struct {
	Index   int         // Index of the task item
	OldTask archon.Task // Previous task data
	NewTask archon.Task // Updated task data
}

// TaskItemFocusChangedMsg indicates focus state change for accessibility
type TaskItemFocusChangedMsg struct {
	Index    int  // Index of the task item
	HasFocus bool // Whether this item has focus
}

// TaskItemActionRequestMsg requests an action to be performed on a task
type TaskItemActionRequestMsg struct {
	Index  int         // Index of the task item
	Task   archon.Task // Task to perform action on
	Action string      // Action to perform ("edit", "delete", "toggle_status", etc.)
}

// Helper functions for creating common messages

// NewTaskItemUpdate creates a TaskItemUpdateMsg
func NewTaskItemUpdate(index int, task archon.Task, isSelected, isHighlighted bool, searchQuery string) TaskItemUpdateMsg {
	return TaskItemUpdateMsg{
		Index:         index,
		Task:          task,
		IsSelected:    isSelected,
		IsHighlighted: isHighlighted,
		SearchQuery:   searchQuery,
	}
}

// NewTaskItemResize creates a TaskItemResizeMsg
func NewTaskItemResize(index, width int) TaskItemResizeMsg {
	return TaskItemResizeMsg{
		Index: index,
		Width: width,
	}
}

// NewTaskItemSelectionChanged creates a TaskItemSelectionChangedMsg
func NewTaskItemSelectionChanged(index int, isSelected bool) TaskItemSelectionChangedMsg {
	return TaskItemSelectionChangedMsg{
		Index:      index,
		IsSelected: isSelected,
	}
}

// NewTaskItemClicked creates a TaskItemClickedMsg
func NewTaskItemClicked(index int, task archon.Task) TaskItemClickedMsg {
	return TaskItemClickedMsg{
		Index: index,
		Task:  task,
	}
}

// NewTaskItemHighlightChanged creates a TaskItemHighlightChangedMsg
func NewTaskItemHighlightChanged(index int, isHighlighted bool, searchQuery string) TaskItemHighlightChangedMsg {
	return TaskItemHighlightChangedMsg{
		Index:         index,
		IsHighlighted: isHighlighted,
		SearchQuery:   searchQuery,
	}
}

// NewTaskItemDataChanged creates a TaskItemDataChangedMsg
func NewTaskItemDataChanged(index int, oldTask, newTask archon.Task) TaskItemDataChangedMsg {
	return TaskItemDataChangedMsg{
		Index:   index,
		OldTask: oldTask,
		NewTask: newTask,
	}
}

// NewTaskItemActionRequest creates a TaskItemActionRequestMsg
func NewTaskItemActionRequest(index int, task archon.Task, action string) TaskItemActionRequestMsg {
	return TaskItemActionRequestMsg{
		Index:  index,
		Task:   task,
		Action: action,
	}
}
