package input

// This file contains status change modal input handler logic

// statusChangeModalInputKeys defines the keys handled in status change modal mode
var statusChangeModalInputKeys = map[string]string{
	"esc":   "close",
	"j":     "down",
	"down":  "down",
	"k":     "up",
	"up":    "up",
	"enter": "confirm",
	"q":     "close",
	"ctrl+c": "quit",
}

// IsStatusChangeModalKey checks if a key is handled by status change modal
func IsStatusChangeModalKey(key string) bool {
	_, exists := statusChangeModalInputKeys[key]
	return exists
}

// GetStatusChangeModalAction returns the action for a status change modal key
func GetStatusChangeModalAction(key string) string {
	if action, exists := statusChangeModalInputKeys[key]; exists {
		return action
	}
	return ""
}

// The actual handleStatusChangeModeInput function remains in the ui package
// as a method on Model to access private fields and helper methods like:
// - handleStatusChangeNavigation(direction int)
// - handleStatusChangeConfirm()
// - SetStatusChangeMode(active bool)