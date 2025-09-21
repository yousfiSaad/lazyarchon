package input

// This file contains all navigation input handling logic

// basicNavigationKeys defines the basic up/down navigation keys
var basicNavigationKeys = map[string]string{
	"up":   "nav_up",
	"k":    "nav_up",
	"down": "nav_down",
	"j":    "nav_down",
}

// jumpNavigationKeys defines the jump navigation keys (gg/G handled separately)
var jumpNavigationKeys = map[string]string{
	"g":       "jump_first",  // Only valid as part of "gg" sequence
	"G":       "jump_last",
	"home":    "jump_first",
	"end":     "jump_last",
}

// fastScrollKeys defines the fast scroll navigation keys
var fastScrollKeys = map[string]string{
	"J": "fast_scroll_down",
	"K": "fast_scroll_up",
}

// halfPageScrollKeys defines the half-page scroll navigation keys
var halfPageScrollKeys = map[string]string{
	"ctrl+u": "half_page_up",
	"pgup":   "half_page_up",
	"ctrl+d": "half_page_down",
	"pgdown": "half_page_down",
}

// horizontalNavigationKeys defines the left/right panel navigation keys
var horizontalNavigationKeys = map[string]string{
	"h": "nav_left",
	"l": "nav_right",
}

// projectModeNavigationKeys defines navigation keys specific to project mode
var projectModeNavigationKeys = map[string]string{
	"h": "project_back",
	"l": "project_select",
}

// IsBasicNavigationKey checks if a key is for basic up/down navigation
func IsBasicNavigationKey(key string) bool {
	_, exists := basicNavigationKeys[key]
	return exists
}

// GetBasicNavigationAction returns the action for a basic navigation key
func GetBasicNavigationAction(key string) string {
	if action, exists := basicNavigationKeys[key]; exists {
		return action
	}
	return ""
}

// IsJumpNavigationKey checks if a key is for jump navigation
func IsJumpNavigationKey(key string) bool {
	_, exists := jumpNavigationKeys[key]
	return exists
}

// GetJumpNavigationAction returns the action for a jump navigation key
func GetJumpNavigationAction(key string) string {
	if action, exists := jumpNavigationKeys[key]; exists {
		return action
	}
	return ""
}

// IsFastScrollKey checks if a key is for fast scrolling
func IsFastScrollKey(key string) bool {
	_, exists := fastScrollKeys[key]
	return exists
}

// GetFastScrollAction returns the action for a fast scroll key
func GetFastScrollAction(key string) string {
	if action, exists := fastScrollKeys[key]; exists {
		return action
	}
	return ""
}

// IsHalfPageScrollKey checks if a key is for half-page scrolling
func IsHalfPageScrollKey(key string) bool {
	_, exists := halfPageScrollKeys[key]
	return exists
}

// GetHalfPageScrollAction returns the action for a half-page scroll key
func GetHalfPageScrollAction(key string) string {
	if action, exists := halfPageScrollKeys[key]; exists {
		return action
	}
	return ""
}

// IsHorizontalNavigationKey checks if a key is for horizontal navigation
func IsHorizontalNavigationKey(key string) bool {
	_, exists := horizontalNavigationKeys[key]
	return exists
}

// GetHorizontalNavigationAction returns the action for a horizontal navigation key
func GetHorizontalNavigationAction(key string) string {
	if action, exists := horizontalNavigationKeys[key]; exists {
		return action
	}
	return ""
}

// IsProjectModeNavigationKey checks if a key is for project mode navigation
func IsProjectModeNavigationKey(key string) bool {
	_, exists := projectModeNavigationKeys[key]
	return exists
}

// GetProjectModeNavigationAction returns the action for a project mode navigation key
func GetProjectModeNavigationAction(key string) string {
	if action, exists := projectModeNavigationKeys[key]; exists {
		return action
	}
	return ""
}

// IsAnyNavigationKey checks if a key is any type of navigation key
func IsAnyNavigationKey(key string) bool {
	return IsBasicNavigationKey(key) ||
		IsJumpNavigationKey(key) ||
		IsFastScrollKey(key) ||
		IsHalfPageScrollKey(key) ||
		IsHorizontalNavigationKey(key) ||
		IsProjectModeNavigationKey(key)
}

// The actual navigation functions remain in the ui package
// as methods on Model to access private fields and helper methods like:
// - handleUpNavigation()
// - handleDownNavigation()
// - handleJumpToFirst()
// - handleJumpToLast()
// - handleFastScrollUp()
// - handleFastScrollDown()
// - handleHalfPageUp()
// - handleHalfPageDown()
// - handleMultiKeySequence(key string) // for "gg" sequence
// - getTaskListHalfPageSize()
// - getDetailHalfPageSize()
// - setSelectedTask(index int)
// - IsLeftPanelActive()
// - IsRightPanelActive()
// - GetSortedTasks()
// - GetContentHeight()