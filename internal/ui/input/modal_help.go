package input

// This file contains the help modal input handler logic
// The actual implementation will be added to the ui package as a method on Model
// to access private fields properly.

// helpModalInputKeys defines the keys handled in help modal mode
var helpModalInputKeys = map[string]string{
	"?":         "toggle",
	"esc":       "close",
	"j":         "down1",
	"down":      "down1",
	"k":         "up1",
	"up":        "up1",
	"J":         "down4",
	"K":         "up4",
	"ctrl+u":    "halfup",
	"pgup":      "halfup",
	"ctrl+d":    "halfdown",
	"pgdown":    "halfdown",
	"gg":        "top",
	"G":         "bottom",
	"home":      "top",
	"end":       "bottom",
	"q":         "close",
	"ctrl+c":    "quit",
}

// IsHelpModalKey checks if a key is handled by help modal
func IsHelpModalKey(key string) bool {
	_, exists := helpModalInputKeys[key]
	return exists
}

// GetHelpModalAction returns the action for a help modal key
func GetHelpModalAction(key string) string {
	if action, exists := helpModalInputKeys[key]; exists {
		return action
	}
	return ""
}

// The actual handleHelpModeInput function will be implemented in the ui package
// as a method on Model to access private viewport fields.