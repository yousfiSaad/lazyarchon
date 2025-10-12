package keys

// Key constants provide type-safe, maintainable definitions for all keyboard input
// in the LazyArchon application. This replaces string literals throughout the codebase
// to improve type safety, maintainability, and documentation.

// Application Control Keys
// These keys control application-level operations like quit, refresh, and mode switching
const (
	// Quit and Exit Operations
	KeyQ      = "q"      // Smart quit (close modal or show quit confirmation)
	KeyCtrlC  = "ctrl+c" // Emergency quit (bypass all modals)
	KeyEscape = "esc"    // General escape/cancel operation

	// Refresh and Retry Operations
	KeyR  = "r"  // Refresh data or retry failed operations
	KeyF5 = "F5" // Alternative refresh key

	// Mode Control Keys
	KeyP     = "p"     // Activate project selection mode
	KeyA     = "a"     // Show all tasks (exit project filtering)
	KeyEnter = "enter" // General confirmation/selection

	// Help and Information
	KeyQuestion = "?" // Toggle help modal
)

// Navigation Keys
// These keys control movement and scrolling within the interface
const (
	// Basic Movement (Vim-style)
	KeyH = "h" // Left/back navigation
	KeyJ = "j" // Down navigation
	KeyK = "k" // Up navigation
	KeyL = "l" // Right/forward navigation
	KeyG = "g" // Used in multi-key sequences like 'gg'

	// Arrow Keys (Alternative navigation)
	KeyArrowUp    = "up"    // Up navigation (alternative to k)
	KeyArrowDown  = "down"  // Down navigation (alternative to j)
	KeyArrowLeft  = "left"  // Left navigation (alternative to h)
	KeyArrowRight = "right" // Right navigation (alternative to l)

	// Jump Navigation
	KeyGG   = "gg"   // Jump to first item (vim-style)
	KeyGCap = "G"    // Jump to last item
	KeyHome = "home" // Jump to first item (alternative)
	KeyEnd  = "end"  // Jump to last item (alternative)

	// Fast Scrolling
	KeyJCap = "J" // Fast scroll down (4 lines)
	KeyKCap = "K" // Fast scroll up (4 lines)
	KeyHCap = "H" // Fast adjustment left/decrease (modal context)
	KeyLCap = "L" // Fast adjustment right/increase (modal context)

	// Page Navigation
	KeyCtrlU = "ctrl+u" // Half-page up
	KeyCtrlD = "ctrl+d" // Half-page down
	KeyPgUp  = "pgup"   // Page up (alternative)
	KeyPgDn  = "pgdown" // Page down (alternative)
)

// Search and Filter Keys
// These keys control search functionality and content filtering
const (
	// Search Activation
	KeySlash = "/"      // Activate inline search
	KeyCtrlF = "ctrl+f" // Alternative search activation

	// Search Navigation
	KeyN    = "n" // Next search match
	KeyNCap = "N" // Previous search match

	// Search Control
	KeyCtrlX = "ctrl+x" // Clear current search
	KeyCtrlL = "ctrl+l" // Alternative clear search
	// Note: KeyCtrlU is defined in Navigation section as it's primarily used for half-page up
)

// Task Operation Keys
// These keys control task-specific operations
const (
	// Task Status and Editing
	KeyT = "t" // Open task status change modal
	KeyE = "e" // Open task edit modal
	KeyD = "d" // Delete/archive task

	// Copy Operations (Yank in vim terminology)
	KeyY    = "y" // Copy task ID (yank)
	KeyYCap = "Y" // Copy task title (yank title)

	// Task Organization
	KeyF    = "f" // Open feature selection modal
	KeyS    = "s" // Cycle sort mode forward
	KeySCap = "S" // Cycle sort mode backward
)

// Modal and Special Input Keys
// These keys are used within modals and special input modes
const (
	// Text Input Operations
	KeyBackspace = "backspace" // Remove character in input fields
	KeySpace     = " "         // Space bar (toggle in feature modal)
	KeyTab       = "tab"       // Navigate to next field in modal
	KeyShiftTab  = "shift+tab" // Navigate to previous field in modal

	// Modal Navigation and Control
	// (Uses same navigation keys as above, but in modal context)
)

// Key Categories for Registry Classification
// These help the key registry categorize and prioritize key handling
const (
	CategoryApplication = "application"
	CategoryNavigation  = "navigation"
	CategorySearch      = "search"
	CategoryTask        = "task"
	CategoryModal       = "modal"
)

// Key Actions for Semantic Mapping
// These provide semantic meaning for key operations
const (
	// Application Actions
	ActionQuit         = "quit"
	ActionForceQuit    = "force_quit"
	ActionRefresh      = "refresh"
	ActionProjectMode  = "project_mode"
	ActionShowAllTasks = "show_all_tasks"
	ActionEscape       = "escape"
	ActionConfirm      = "confirm"
	ActionToggleHelp   = "toggle_help"

	// Navigation Actions
	ActionMoveUp         = "move_up"
	ActionMoveDown       = "move_down"
	ActionMoveLeft       = "move_left"
	ActionMoveRight      = "move_right"
	ActionJumpFirst      = "jump_first"
	ActionJumpLast       = "jump_last"
	ActionFastScrollDown = "fast_scroll_down"
	ActionFastScrollUp   = "fast_scroll_up"
	ActionHalfPageUp     = "half_page_up"
	ActionHalfPageDown   = "half_page_down"

	// Search Actions
	ActionActivateSearch = "activate_search"
	ActionClearSearch    = "clear_search"
	ActionNextMatch      = "next_match"
	ActionPrevMatch      = "prev_match"

	// Task Actions
	ActionChangeStatus   = "change_status"
	ActionEditTask       = "edit_task"
	ActionDeleteTask     = "delete_task"
	ActionCopyID         = "copy_id"
	ActionCopyTitle      = "copy_title"
	ActionSelectFeatures = "select_features"
	ActionSortForward    = "sort_forward"
	ActionSortBackward   = "sort_backward"

	// Modal Actions
	ActionToggle = "toggle"
	ActionClose  = "close"
	ActionCancel = "cancel"
	ActionSelect = "select"
	ActionDown   = "down"
	ActionUp     = "up"

	// Help Modal Navigation Actions
	ActionDown1    = "down1"
	ActionUp1      = "up1"
	ActionDown4    = "down4"
	ActionUp4      = "up4"
	ActionHalfUp   = "halfup"
	ActionHalfDown = "halfdown"
	ActionTop      = "top"
	ActionBottom   = "bottom"

	// Feature Modal Actions
	ActionStartSearch     = "start_search"
	ActionSearchBackspace = "search_backspace"
	ActionCancelNew       = "cancel_new"
	ActionConfirmNew      = "confirm_new"
	ActionDeleteChar      = "delete_char"
	ActionToggleAll       = "toggle_all"
	ActionSearch          = "search"
	ActionFastDown        = "fast_down"
	ActionFastUp          = "fast_up"
)
