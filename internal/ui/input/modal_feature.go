package input

// This file contains feature modal and task edit modal input handler logic

// featureModalInputKeys defines the keys handled in feature selection modal mode
var featureModalInputKeys = map[string]string{
	"esc":      "cancel",
	"q":        "cancel",
	"j":        "down",
	"down":     "down",
	"k":        "up",
	"up":       "up",
	"enter":    "select",
	"ctrl+c":   "quit",
	"/":        "start_search",
	"backspace": "search_backspace",
	"ctrl+x":   "clear_search",
}

// taskEditModalInputKeys defines the keys handled in task edit modal mode
var taskEditModalInputKeys = map[string]string{
	"esc":    "cancel",
	"q":      "cancel",
	"j":      "down",
	"down":   "down",
	"k":      "up",
	"up":     "up",
	"enter":  "select",
	"ctrl+c": "quit",
}

// taskEditNewFeatureInputKeys defines the keys handled when creating new feature
var taskEditNewFeatureInputKeys = map[string]string{
	"esc":       "cancel_new",
	"enter":     "confirm_new",
	"backspace": "delete_char",
	"ctrl+c":    "quit",
	// Character input is handled separately in validation function
}

// IsFeatureModalKey checks if a key is handled by feature modal
func IsFeatureModalKey(key string) bool {
	_, exists := featureModalInputKeys[key]
	return exists
}

// GetFeatureModalAction returns the action for a feature modal key
func GetFeatureModalAction(key string) string {
	if action, exists := featureModalInputKeys[key]; exists {
		return action
	}
	return ""
}

// IsTaskEditModalKey checks if a key is handled by task edit modal
func IsTaskEditModalKey(key string) bool {
	_, exists := taskEditModalInputKeys[key]
	return exists
}

// GetTaskEditModalAction returns the action for a task edit modal key
func GetTaskEditModalAction(key string) string {
	if action, exists := taskEditModalInputKeys[key]; exists {
		return action
	}
	return ""
}

// IsTaskEditNewFeatureKey checks if a key is handled in new feature creation mode
func IsTaskEditNewFeatureKey(key string) bool {
	_, exists := taskEditNewFeatureInputKeys[key]
	return exists
}

// GetTaskEditNewFeatureAction returns the action for a new feature creation key
func GetTaskEditNewFeatureAction(key string) string {
	if action, exists := taskEditNewFeatureInputKeys[key]; exists {
		return action
	}
	return ""
}

// IsValidFeatureNameChar checks if a character is valid for feature names
func IsValidFeatureNameChar(char string) bool {
	if len(char) != 1 {
		return false
	}
	c := char[0]
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') || c == '-' || c == '_'
}

// The actual handleFeatureModeInput and handleTaskEditModeInput functions remain in the ui package
// as methods on Model to access private fields and helper methods like:
// - handleFeatureNavigation(direction int)
// - handleFeatureSearch(char string)
// - handleFeatureConfirm()
// - handleTaskEditNavigation(direction int)
// - handleTaskEditConfirm(feature string)
// - SetTaskEditMode(active bool)
// - GetUniqueFeatures() []string