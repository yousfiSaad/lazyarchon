package input

// This file contains all search input handling logic

// inlineSearchInputKeys defines the keys handled when inline search mode is active
var inlineSearchInputKeys = map[string]string{
	"esc":      "cancel",
	"enter":    "commit",
	"backspace": "delete_char",
	"ctrl+u":   "clear_all",
	"ctrl+c":   "quit",
	// Character input is handled separately in validation function
}

// searchActivationKeys defines the keys that activate search mode
var searchActivationKeys = map[string]string{
	"/":      "activate_search",
	"ctrl+f": "activate_search",
}

// searchNavigationKeys defines the keys for search navigation
var searchNavigationKeys = map[string]string{
	"n": "next_match",
	"N": "previous_match",
}

// searchClearKeys defines the keys that clear search
var searchClearKeys = map[string]string{
	"ctrl+x": "clear_search",
	"ctrl+l": "clear_search",
}

// featureModalSearchKeys defines search keys specific to feature modal
var featureModalSearchKeys = map[string]string{
	"/":        "start_search",
	"backspace": "search_backspace",
	"ctrl+l":   "clear_search",
}

// IsInlineSearchKey checks if a key is handled in inline search mode
func IsInlineSearchKey(key string) bool {
	_, exists := inlineSearchInputKeys[key]
	return exists
}

// GetInlineSearchAction returns the action for an inline search key
func GetInlineSearchAction(key string) string {
	if action, exists := inlineSearchInputKeys[key]; exists {
		return action
	}
	return ""
}

// IsSearchActivationKey checks if a key activates search mode
func IsSearchActivationKey(key string) bool {
	_, exists := searchActivationKeys[key]
	return exists
}

// GetSearchActivationAction returns the action for a search activation key
func GetSearchActivationAction(key string) string {
	if action, exists := searchActivationKeys[key]; exists {
		return action
	}
	return ""
}

// IsSearchNavigationKey checks if a key is for search navigation
func IsSearchNavigationKey(key string) bool {
	_, exists := searchNavigationKeys[key]
	return exists
}

// GetSearchNavigationAction returns the action for a search navigation key
func GetSearchNavigationAction(key string) string {
	if action, exists := searchNavigationKeys[key]; exists {
		return action
	}
	return ""
}

// IsSearchClearKey checks if a key clears search
func IsSearchClearKey(key string) bool {
	_, exists := searchClearKeys[key]
	return exists
}

// GetSearchClearAction returns the action for a search clear key
func GetSearchClearAction(key string) string {
	if action, exists := searchClearKeys[key]; exists {
		return action
	}
	return ""
}

// IsFeatureModalSearchKey checks if a key is handled in feature modal search
func IsFeatureModalSearchKey(key string) bool {
	_, exists := featureModalSearchKeys[key]
	return exists
}

// GetFeatureModalSearchAction returns the action for a feature modal search key
func GetFeatureModalSearchAction(key string) string {
	if action, exists := featureModalSearchKeys[key]; exists {
		return action
	}
	return ""
}

// IsValidSearchChar checks if a character is valid for search input
func IsValidSearchChar(char string) bool {
	if len(char) != 1 {
		return false
	}
	c := char[0]
	// Allow printable ASCII characters
	return c >= 32 && c <= 126
}

// The actual search input functions remain in the ui package
// as methods on Model to access private fields and helper methods like:
// - handleInlineSearchInput(key string)
// - ActivateInlineSearch()
// - CancelInlineSearch()
// - CommitInlineSearch()
// - ClearSearch()
// - UpdateRealTimeSearch()
// - nextSearchMatch()
// - previousSearchMatch()
// - activateFeatureSearch()
// - clearFeatureSearch()
// - handleMultiKeySequence(key string)