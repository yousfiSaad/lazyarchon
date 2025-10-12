package base

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/internal/shared/interfaces"
	"github.com/yousfisaad/lazyarchon/internal/ui/context"
)

// ComponentType identifies the type of component for message routing
type ComponentType string

const (
	HelpModalComponent             ComponentType = "help_modal"
	StatusModalComponent           ComponentType = "status_modal"
	StatusFilterModalComponent     ComponentType = "status_filter_modal"
	FeatureModalComponent          ComponentType = "feature_modal"
	TaskEditModalComponent         ComponentType = "task_edit_modal"
	ConfirmationModalComponent     ComponentType = "confirmation_modal"
	SearchComponent                ComponentType = "search"
	TableComponent                 ComponentType = "table"
	SidebarComponent               ComponentType = "sidebar"
	ScrollbarComponent             ComponentType = "scrollbar"
	ItemComponent                  ComponentType = "item" // For individual items within lists
	NavigationCoordinatorComponent ComponentType = "navigation_coordinator"
	HeaderComponent                ComponentType = "header"
	StatusBarComponent             ComponentType = "statusbar"
	MainContentComponent           ComponentType = "main_content"
	MessageHandlerComponent        ComponentType = "message_handler"
)

// ModalType identifies which modal is currently active
type ModalType string

const (
	ModalTypeNone         ModalType = ""              // No modal active
	ModalTypeHelp         ModalType = "help"          // Help modal
	ModalTypeFeature      ModalType = "feature"       // Feature selection modal
	ModalTypeStatus       ModalType = "status"        // Status change modal
	ModalTypeStatusFilter ModalType = "status_filter" // Status filter modal
	ModalTypeTaskEdit     ModalType = "task_edit"     // Task edit modal
	ModalTypeConfirmation ModalType = "confirmation"  // Confirmation modal
)

// Layout constants for component rendering
const (
	// PanelBorderLines is the vertical space used by Panel borders.
	// Panel(width, height) renders borders INSIDE the height, not outside.
	// Content area = height - PanelBorderLines
	PanelBorderLines = 2 // Top border (1) + Bottom border (1)
)

// ComponentContext provides shared context that components need
// Following clean architecture principles with single source of truth
type ComponentContext struct {
	// Reference to application state (single source of truth)
	ProgramContext *context.ProgramContext

	// Reference to UI presentation state (single source of truth)
	UIState *context.UIState

	// Component-specific service dependencies
	ConfigProvider       interfaces.ConfigProvider
	StyleContextProvider interfaces.StyleContextProvider
	Logger               interfaces.Logger

	// Component communication
	MessageChan chan tea.Msg

	// Parent-provided callbacks for accessing computed data
	// GetSortedTasks provides access to sorted/filtered tasks from MainModel
	// This stays as a callback because it involves complex filtering logic in MainModel
	GetSortedTasks func() []interface{} // []archon.Task but using interface{} to avoid import cycle

	// NOTE: Computed data accessors removed - components now call ProgramContext/UIState methods directly:
	// - GetTaskStatusCounts() → ctx.ProgramContext.GetTaskStatusCounts()
	// - GetCurrentSortModeName() → ctx.ProgramContext.GetCurrentSortModeName()
	// - GetSelectedTaskIndex() → ctx.UIState.GetSelectedTaskIndex()
	// - GetTaskSearchState() → ctx.UIState.GetTaskSearchState(selectedIndex)
	// - GetUniqueFeatures() → ctx.ProgramContext.GetUniqueFeatures()
	// - GetActiveViewName() → ctx.UIState.GetActiveViewName()
	// - GetCurrentProjectName() → ctx.ProgramContext.GetCurrentProjectName()
	// - GetFeatureFilterSummary() → ctx.ProgramContext.GetFeatureFilterSummary(featureFilters)
}

// GetScreenWidth returns the current screen width from ProgramContext
func (ctx *ComponentContext) GetScreenWidth() int {
	if ctx.ProgramContext == nil {
		return 80 // Default fallback
	}
	return ctx.ProgramContext.ScreenWidth
}

// GetScreenHeight returns the current screen height from ProgramContext
func (ctx *ComponentContext) GetScreenHeight() int {
	if ctx.ProgramContext == nil {
		return 24 // Default fallback
	}
	return ctx.ProgramContext.ScreenHeight
}

// NOTE: GetMainContentWidth and GetMainContentHeight methods removed
// Components now store their own dimensions received via WindowSizeMsg

// ComponentMessage is a wrapper for messages sent between components
type ComponentMessage struct {
	SenderID   string
	SenderType ComponentType
	TargetID   string        // Empty string means broadcast to all components
	TargetType ComponentType // Can be used for type-specific routing
	Payload    tea.Msg       // The actual message content
}

// Common component messages
type (
	// ShowComponentMsg is sent when a component should be shown
	ShowComponentMsg struct {
		ComponentID string
		Data        interface{} // Optional data to pass to the component
	}

	// HideComponentMsg is sent when a component should be hidden
	HideComponentMsg struct {
		ComponentID string
	}

	// ComponentShownMsg is sent when a component has been shown
	ComponentShownMsg struct {
		ComponentID   string
		ComponentType ComponentType
	}

	// ComponentHiddenMsg is sent when a component has been hidden
	ComponentHiddenMsg struct {
		ComponentID   string
		ComponentType ComponentType
	}

	// FocusComponentMsg is sent when a component should receive focus
	FocusComponentMsg struct {
		ComponentID string
	}

	// BlurComponentMsg is sent when a component should lose focus
	BlurComponentMsg struct {
		ComponentID string
	}
)

// BaseComponent provides common functionality for all components.
//
// BaseComponent contains state that ALL components need:
// - Identity (id, compType)
// - Focus management (focused) - for keyboard input handling
// - Dimensions (width, height) - for rendering
// - Context reference - for accessing global state and services
//
// Note: Modal-specific state (active/visible) has been moved to BaseModal.
// Panels should NOT have an active field - they read active state from UIState directly.
type BaseComponent struct {
	id       string
	compType ComponentType
	focused  bool
	context  *ComponentContext

	// Dimensions - all components need to track their allocated space
	width  int
	height int
}

// NewBaseComponent creates a new base component
func NewBaseComponent(id string, compType ComponentType, context *ComponentContext) BaseComponent {
	return BaseComponent{
		id:       id,
		compType: compType,
		focused:  false,
		context:  context,
	}
}

// GetID returns the component's unique identifier
func (c *BaseComponent) GetID() string {
	return c.id
}

// GetType returns the component's type
func (c *BaseComponent) GetType() ComponentType {
	return c.compType
}

// CanFocus returns whether this component can receive focus
// Base implementation returns false, components that can receive input should override
func (c *BaseComponent) CanFocus() bool {
	return false
}

// SetFocus sets the component's focus state
func (c *BaseComponent) SetFocus(focused bool) {
	c.focused = focused
}

// IsFocused returns whether the component currently has focus
func (c *BaseComponent) IsFocused() bool {
	return c.focused
}

// GetContext returns the component's context
func (c *BaseComponent) GetContext() *ComponentContext {
	return c.context
}

// =============================================================================
// DIMENSION MANAGEMENT
// =============================================================================

// SetDimensions sets the component's width and height
func (c *BaseComponent) SetDimensions(width, height int) {
	c.width = width
	c.height = height
}

// GetWidth returns the component's width
func (c *BaseComponent) GetWidth() int {
	return c.width
}

// GetHeight returns the component's height
func (c *BaseComponent) GetHeight() int {
	return c.height
}

// HandleWindowResize is a convenience method for handling WindowSizeMsg
// Components can call this from their Update() method for standard resize handling
func (c *BaseComponent) HandleWindowResize(msg tea.WindowSizeMsg) {
	c.width = msg.Width
	c.height = msg.Height
}

// =============================================================================
// MESSAGE SENDING
// =============================================================================

// SendMessage sends a message to other components via the context
func (c *BaseComponent) SendMessage(targetID string, targetType ComponentType, payload tea.Msg) tea.Cmd {
	if c.context == nil || c.context.MessageChan == nil {
		return nil
	}

	return func() tea.Msg {
		return ComponentMessage{
			SenderID:   c.id,
			SenderType: c.compType,
			TargetID:   targetID,
			TargetType: targetType,
			Payload:    payload,
		}
	}
}

// BroadcastMessage sends a message to all components
func (c *BaseComponent) BroadcastMessage(payload tea.Msg) tea.Cmd {
	return c.SendMessage("", "", payload)
}
