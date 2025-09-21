package input

// This file contains task operation input handling logic

// taskOperationKeys defines the keys for task operations
var taskOperationKeys = map[string]string{
	"t": "change_status",
	"e": "edit_task",
	"f": "filter_feature",
	"y": "copy_task_id",
	"Y": "copy_task_title",
}

// IsTaskOperationKey checks if a key is for task operations
func IsTaskOperationKey(key string) bool {
	_, exists := taskOperationKeys[key]
	return exists
}

// GetTaskOperationAction returns the action for a task operation key
func GetTaskOperationAction(key string) string {
	if action, exists := taskOperationKeys[key]; exists {
		return action
	}
	return ""
}

// The actual task operation functions remain in the ui package
// as methods on Model to access private fields and helper methods like:
// - SetStatusChangeMode(active bool)
// - SetTaskEditMode(active bool)
// - SetFeatureMode(active bool)
// - handleTaskIDCopy()
// - handleTaskTitleCopy()
// - GetSortedTasks()
// - GetUniqueFeatures()