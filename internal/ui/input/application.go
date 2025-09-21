package input

// This file contains application-level command input handling logic

// applicationKeys defines the keys for application-level commands
var applicationKeys = map[string]string{
	"q":      "quit",
	"ctrl+c": "force_quit",
	"r":      "refresh",
	"F5":     "refresh",
	"p":      "project_mode",
	"a":      "show_all_tasks",
	"esc":    "escape",
	"enter":  "confirm",
	"?":      "toggle_help",
}

// smartQuitKeys defines keys that have smart quit behavior (close modal instead of quit)
var smartQuitKeys = map[string]string{
	"q": "smart_quit",
}

// projectModeKeys defines keys specific to project mode
var projectModeKeys = map[string]string{
	"p":     "activate_project_mode",
	"esc":   "exit_project_mode",
	"enter": "select_project",
}

// IsApplicationKey checks if a key is for application commands
func IsApplicationKey(key string) bool {
	_, exists := applicationKeys[key]
	return exists
}

// GetApplicationAction returns the action for an application key
func GetApplicationAction(key string) string {
	if action, exists := applicationKeys[key]; exists {
		return action
	}
	return ""
}

// IsSmartQuitKey checks if a key has smart quit behavior
func IsSmartQuitKey(key string) bool {
	_, exists := smartQuitKeys[key]
	return exists
}

// GetSmartQuitAction returns the action for a smart quit key
func GetSmartQuitAction(key string) string {
	if action, exists := smartQuitKeys[key]; exists {
		return action
	}
	return ""
}

// IsProjectModeKey checks if a key is for project mode operations
func IsProjectModeKey(key string) bool {
	_, exists := projectModeKeys[key]
	return exists
}

// GetProjectModeAction returns the action for a project mode key
func GetProjectModeAction(key string) string {
	if action, exists := projectModeKeys[key]; exists {
		return action
	}
	return ""
}

// The actual application command functions remain in the ui package
// as methods on Model to access private fields and helper methods like:
// - SetHelpMode(active bool)
// - IsHelpMode()
// - HasActiveModal()
// - SetSelectedProject(projectID *string)
// - SetLoadingWithMessage(loading bool, message string)
// - ClearError()
// - LoadTasksWithProject(client, projectID)
// - IsLeftPanelActive()
// - IsRightPanelActive()