package base

// BaseModal provides common functionality for modal components.
// Modals are temporary overlays that need lifecycle management (show/hide).
//
// BaseModal embeds BaseComponent and adds modal-specific state:
// - active: Whether the modal is currently visible/shown
//
// All modal components should embed BaseModal instead of BaseComponent directly.
// This provides shared modal lifecycle management while keeping BaseComponent clean.
type BaseModal struct {
	BaseComponent

	// ===================================================================
	// MODAL LIFECYCLE STATE
	// ===================================================================
	// Whether modal is currently visible (shown to user)
	// This is distinct from 'focused' which is about keyboard input handling
	active bool
}

// NewBaseModal creates a new base modal with the given ID, type, and context.
// The modal is initially inactive (hidden).
func NewBaseModal(id string, modalType ComponentType, context *ComponentContext) BaseModal {
	return BaseModal{
		BaseComponent: NewBaseComponent(id, modalType, context),
		active:        false,
	}
}

// SetActive sets the modal's visibility state.
// When active=true, the modal is shown and can receive input.
// When active=false, the modal is hidden and should return empty string from View().
func (m *BaseModal) SetActive(active bool) {
	m.active = active
}

// IsActive returns whether the modal is currently visible.
// Modals typically check this in View() to return empty string when inactive,
// and in Update() to ignore input when inactive.
func (m *BaseModal) IsActive() bool {
	return m.active
}
