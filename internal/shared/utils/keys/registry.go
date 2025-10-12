package keys

// KeyBinding represents a complete key binding with its action and help description
type KeyBinding struct {
	Key         string // The actual key (e.g., "j", "ctrl+c")
	Action      string // Semantic action (e.g., ActionMoveDown)
	Category    string // Category for organization (e.g., CategoryNavigation)
	Context     string // Context where this binding applies (e.g., "main", "modal", "help")
	Description string // Human-readable description for help
	Priority    int    // Display priority (lower = shown first)
}

// ContextualKeyBindings represents key bindings for a specific context
type ContextualKeyBindings struct {
	Context  string       // Context name (e.g., "main", "help_modal", "project_mode")
	Bindings []KeyBinding // Key bindings active in this context
}

// KeyRegistry manages all key bindings and their documentation
type KeyRegistry struct {
	contextBindings map[string][]KeyBinding // Context -> bindings
	keyToAction     map[string]string       // Key -> Action lookup
	actionToKey     map[string]string       // Action -> Key lookup
}

// NewKeyRegistry creates a new key registry with all bindings
// If keybindingsConfig is provided, it will override default bindings
func NewKeyRegistry(keybindingsConfig interface{}) *KeyRegistry {
	registry := &KeyRegistry{
		contextBindings: make(map[string][]KeyBinding),
		keyToAction:     make(map[string]string),
		actionToKey:     make(map[string]string),
	}

	// Register all key bindings with defaults
	registry.registerMainContextBindings()
	registry.registerHelpModalBindings()
	registry.registerProjectModeBindings()
	registry.registerModalBindings()

	// Apply custom keybindings if provided
	if keybindingsConfig != nil {
		registry.applyCustomKeybindings(keybindingsConfig)
	}

	return registry
}

// applyCustomKeybindings applies user-configured keybindings to override defaults
// This function is a placeholder for future implementation when config package is available
func (r *KeyRegistry) applyCustomKeybindings(keybindingsConfig interface{}) {
	// TODO: Implement custom keybinding overrides
	// This will allow users to customize their keyboard shortcuts via config.yaml
	// For now, we maintain backward compatibility by using all defaults
}

// GetContextBindings returns all key bindings for a specific context
func (r *KeyRegistry) GetContextBindings(context string) []KeyBinding {
	bindings, exists := r.contextBindings[context]
	if !exists {
		return []KeyBinding{}
	}
	return bindings
}

// GetActionForKey returns the action associated with a key in any context
func (r *KeyRegistry) GetActionForKey(key string) string {
	return r.keyToAction[key]
}

// GetKeyForAction returns the primary key for an action
func (r *KeyRegistry) GetKeyForAction(action string) string {
	return r.actionToKey[action]
}

// GetBindingsByCategory returns all bindings in a category across all contexts
func (r *KeyRegistry) GetBindingsByCategory(category string) []KeyBinding {
	var bindings []KeyBinding
	for _, contextBindings := range r.contextBindings {
		for _, binding := range contextBindings {
			if binding.Category == category {
				bindings = append(bindings, binding)
			}
		}
	}
	return bindings
}

// GetHelpSections generates organized help sections for display
func (r *KeyRegistry) GetHelpSections() []HelpSection {
	var sections []HelpSection

	// Define section order and titles
	sectionConfigs := []struct {
		Category string
		Title    string
		Context  string
		Priority int
	}{
		{CategoryNavigation, "Panel Navigation", "main", 1},
		{CategoryApplication, "Project Management", "main", 2},
		{CategoryTask, "Task Management", "main", 3},
		{CategoryApplication, "Application Controls", "main", 4},
		{CategoryNavigation, "Help Navigation", "help_modal", 5},
	}

	for _, config := range sectionConfigs {
		section := HelpSection{
			Title:    config.Title,
			Priority: config.Priority,
			Bindings: r.getFilteredBindings(config.Category, config.Context),
		}

		if len(section.Bindings) > 0 {
			sections = append(sections, section)
		}
	}

	// Add visual indicators and task status sections
	sections = append(sections, r.getVisualIndicatorsSection())
	sections = append(sections, r.getTaskStatusSection())

	return sections
}

// HelpSection represents a section in the help display
type HelpSection struct {
	Title    string       // Section title
	Priority int          // Display order
	Bindings []KeyBinding // Bindings in this section
}

// addBinding registers a key binding in the registry
func (r *KeyRegistry) addBinding(context string, binding KeyBinding) {
	// Set context if not provided
	if binding.Context == "" {
		binding.Context = context
	}

	// Add to context bindings
	r.contextBindings[context] = append(r.contextBindings[context], binding)

	// Add to lookup maps
	r.keyToAction[binding.Key] = binding.Action
	r.actionToKey[binding.Action] = binding.Key
}

// registerMainContextBindings registers bindings for the main application context
func (r *KeyRegistry) registerMainContextBindings() {
	context := "main"

	// Panel Navigation
	r.addBinding(context, KeyBinding{
		Key: KeyH + "/" + KeyL, Action: ActionMoveLeft + "/" + ActionMoveRight,
		Category: CategoryNavigation, Description: "Switch between panels", Priority: 1,
	})
	r.addBinding(context, KeyBinding{
		Key: KeyArrowUp + "/" + KeyArrowDown + " or " + KeyJ + "/" + KeyK, Action: ActionMoveUp + "/" + ActionMoveDown,
		Category: CategoryNavigation, Description: "Navigate/scroll (1 line)", Priority: 2,
	})
	r.addBinding(context, KeyBinding{
		Key: KeyJCap + "/" + KeyKCap, Action: ActionFastScrollDown + "/" + ActionFastScrollUp,
		Category: CategoryNavigation, Description: "Fast scroll (4 lines)", Priority: 3,
	})
	r.addBinding(context, KeyBinding{
		Key: KeyCtrlU + "/" + KeyCtrlD, Action: ActionHalfPageUp + "/" + ActionHalfPageDown,
		Category: CategoryNavigation, Description: "Half-page scroll", Priority: 4,
	})
	r.addBinding(context, KeyBinding{
		Key: KeyGG + "/" + KeyGCap, Action: ActionJumpFirst + "/" + ActionJumpLast,
		Category: CategoryNavigation, Description: "Jump to top/bottom", Priority: 5,
	})
	r.addBinding(context, KeyBinding{
		Key: KeyHome + "/" + KeyEnd, Action: ActionJumpFirst + "/" + ActionJumpLast,
		Category: CategoryNavigation, Description: "Jump to start/end", Priority: 6,
	})

	// Project Management
	r.addBinding(context, KeyBinding{
		Key: KeyP, Action: ActionProjectMode,
		Category: CategoryApplication, Description: "Project selection mode", Priority: 10,
	})
	r.addBinding(context, KeyBinding{
		Key: KeyA, Action: ActionShowAllTasks,
		Category: CategoryApplication, Description: "Show all tasks", Priority: 11,
	})
	r.addBinding(context, KeyBinding{
		Key: KeyEnter, Action: ActionConfirm,
		Category: CategoryApplication, Description: "Select project", Priority: 12,
	})
	r.addBinding(context, KeyBinding{
		Key: KeyEscape, Action: ActionEscape,
		Category: CategoryApplication, Description: "Exit project mode", Priority: 13,
	})

	// Task Management
	r.addBinding(context, KeyBinding{
		Key: KeyS + "/" + KeySCap, Action: ActionSortForward + "/" + ActionSortBackward,
		Category: CategoryTask, Description: "Sort tasks by different criteria", Priority: 20,
	})
	r.addBinding(context, KeyBinding{
		Key: KeyT, Action: ActionChangeStatus,
		Category: CategoryTask, Description: "Change task status (Todo/Doing/Review/Done)", Priority: 21,
	})
	r.addBinding(context, KeyBinding{
		Key: KeyE, Action: ActionEditTask,
		Category: CategoryTask, Description: "Edit task properties (status/priority/feature)", Priority: 22,
	})
	r.addBinding(context, KeyBinding{
		Key: KeyD, Action: ActionDeleteTask,
		Category: CategoryTask, Description: "Delete/archive task (with confirmation)", Priority: 23,
	})
	r.addBinding(context, KeyBinding{
		Key: KeyY, Action: ActionCopyID,
		Category: CategoryTask, Description: "Copy task ID to clipboard (yank)", Priority: 24,
	})
	r.addBinding(context, KeyBinding{
		Key: KeyYCap, Action: ActionCopyTitle,
		Category: CategoryTask, Description: "Copy task title to clipboard (yank)", Priority: 25,
	})

	// Application Controls
	r.addBinding(context, KeyBinding{
		Key: KeyR + " or " + KeyF5, Action: ActionRefresh,
		Category: CategoryApplication, Description: "Refresh data from API", Priority: 30,
	})
	r.addBinding(context, KeyBinding{
		Key: KeyQ, Action: ActionQuit,
		Category: CategoryApplication, Description: "Quit application", Priority: 31,
	})
	r.addBinding(context, KeyBinding{
		Key: KeyQuestion, Action: ActionToggleHelp,
		Category: CategoryApplication, Description: "Toggle this help", Priority: 32,
	})
}

// registerHelpModalBindings registers bindings specific to the help modal
func (r *KeyRegistry) registerHelpModalBindings() {
	context := "help_modal"

	r.addBinding(context, KeyBinding{
		Key: KeyJ + "/" + KeyK, Action: ActionDown1 + "/" + ActionUp1,
		Category: CategoryNavigation, Description: "Scroll help (1 line)", Priority: 1,
	})
	r.addBinding(context, KeyBinding{
		Key: KeyJCap + "/" + KeyKCap, Action: ActionDown4 + "/" + ActionUp4,
		Category: CategoryNavigation, Description: "Fast scroll help (4 lines)", Priority: 2,
	})
	r.addBinding(context, KeyBinding{
		Key: KeyCtrlU + "/" + KeyCtrlD, Action: ActionHalfUp + "/" + ActionHalfDown,
		Category: CategoryNavigation, Description: "Half-page scroll help", Priority: 3,
	})
	r.addBinding(context, KeyBinding{
		Key: KeyGG + "/" + KeyGCap, Action: ActionTop + "/" + ActionBottom,
		Category: CategoryNavigation, Description: "Jump to help top/bottom", Priority: 4,
	})
}

// registerProjectModeBindings registers bindings for project selection mode
func (r *KeyRegistry) registerProjectModeBindings() {
	// Project mode inherits most navigation bindings from main context
	// We can add project-specific ones here if needed
}

// registerModalBindings registers common modal bindings
func (r *KeyRegistry) registerModalBindings() {
	context := "modal"

	r.addBinding(context, KeyBinding{
		Key: KeyQuestion + "/" + KeyEscape + "/" + KeyQ, Action: ActionClose,
		Category: CategoryModal, Description: "Close modal", Priority: 1,
	})
}

// getFilteredBindings returns bindings matching category and context
func (r *KeyRegistry) getFilteredBindings(category, context string) []KeyBinding {
	var bindings []KeyBinding
	contextBindings := r.GetContextBindings(context)

	for _, binding := range contextBindings {
		if binding.Category == category {
			bindings = append(bindings, binding)
		}
	}

	return bindings
}

// getVisualIndicatorsSection creates the visual indicators help section
func (r *KeyRegistry) getVisualIndicatorsSection() HelpSection {
	return HelpSection{
		Title:    "Visual Indicators",
		Priority: 100,
		Bindings: []KeyBinding{
			{Description: "Bright cyan border    Active panel"},
			{Description: "Dim gray border       Inactive panel"},
			{Description: "[Tasks]/[Details]     Active panel in status bar"},
			{Description: "▓░ scroll bar         Position indicator"},
		},
	}
}

// getTaskStatusSection creates the task status symbols help section
func (r *KeyRegistry) getTaskStatusSection() HelpSection {
	return HelpSection{
		Title:    "Task Status Symbols",
		Priority: 101,
		Bindings: []KeyBinding{
			{Description: "○  Todo       Not started"},
			{Description: "◐  Doing      In progress"},
			{Description: "◈  Review     Under review"},
			{Description: "✓  Done       Completed"},
		},
	}
}
