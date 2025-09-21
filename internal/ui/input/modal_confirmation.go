package input

// This file contains confirmation modal input handler logic

// confirmationModalInputKeys defines the keys handled in confirmation modal mode
var confirmationModalInputKeys = map[string]string{
	"esc":   "cancel",
	"n":     "cancel",
	"y":     "confirm",
	"enter": "select",  // Act based on selected option
	"j":     "down",    // Navigate to cancel option
	"down":  "down",
	"k":     "up",      // Navigate to confirm option
	"up":    "up",
	"q":     "cancel",  // Close modal instead of recursive confirmation
	"ctrl+c": "quit",
}

// IsConfirmationModalKey checks if a key is handled by confirmation modal
func IsConfirmationModalKey(key string) bool {
	_, exists := confirmationModalInputKeys[key]
	return exists
}

// GetConfirmationModalAction returns the action for a confirmation modal key
func GetConfirmationModalAction(key string) string {
	if action, exists := confirmationModalInputKeys[key]; exists {
		return action
	}
	return ""
}

// The actual handleConfirmationModeInput function remains in the ui package
// as a method on Model to access private modal state and methods like:
// - SetConfirmationMode(active bool, title, message, confirmText string)
// - m.Modals.confirmation.selectedOption (for navigation)