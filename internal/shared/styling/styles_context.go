package styling

// StyleProvider defines the interface for styling configuration
// This allows the styling system to work with any configuration implementation
type StyleProvider interface {
	IsPriorityIndicatorsEnabled() bool
	IsFeatureColorsEnabled() bool
}

// ThemeAdapter provides access to theme colors without circular imports
// This will be implemented by a wrapper around the actual Theme struct
type ThemeAdapter struct {
	TodoColor     string
	DoingColor    string
	ReviewColor   string
	DoneColor     string
	HeaderColor   string
	MutedColor    string
	AccentColor   string
	StatusColor   string
	FeatureColors []string
	Name          string
}

// StyleContext provides a centralized container for all styling-related state
// It serves as the single source of truth for styling decisions
type StyleContext struct {
	theme          *ThemeAdapter
	selectionState SelectionState
	searchState    SearchState
	styleProvider  StyleProvider
	factory        *StyleFactory // Cached factory instance
}

// NewStyleContext creates a new styling context with default states
func NewStyleContext(theme *ThemeAdapter, styleProvider StyleProvider) *StyleContext {
	ctx := &StyleContext{
		theme:          theme,
		selectionState: NewSelectionState(),
		searchState:    NewSearchState(),
		styleProvider:  styleProvider,
	}
	// Create factory instance that references this context
	ctx.factory = &StyleFactory{context: ctx}
	return ctx
}

// WithSelection returns a new context with updated selection state
func (c *StyleContext) WithSelection(selected bool) *StyleContext {
	newCtx := *c // Copy the context
	newCtx.selectionState = c.selectionState.WithSelected(selected)
	newCtx.factory = &StyleFactory{context: &newCtx} // Update factory reference
	return &newCtx
}

// WithCustomSelection returns a new context with fully custom selection state
func (c *StyleContext) WithCustomSelection(state SelectionState) *StyleContext {
	newCtx := *c
	newCtx.selectionState = state
	newCtx.factory = &StyleFactory{context: &newCtx}
	return &newCtx
}

// WithSearch returns a new context with updated search state
func (c *StyleContext) WithSearch(query string, active bool) *StyleContext {
	newCtx := *c
	newCtx.searchState = c.searchState.WithQuery(query)
	newCtx.searchState.IsActive = active // Override with explicit active flag
	newCtx.factory = &StyleFactory{context: &newCtx}
	return &newCtx
}

// WithCustomSearch returns a new context with fully custom search state
func (c *StyleContext) WithCustomSearch(state SearchState) *StyleContext {
	newCtx := *c
	newCtx.searchState = state
	newCtx.factory = &StyleFactory{context: &newCtx}
	return &newCtx
}

// Factory returns the StyleFactory for this context
func (c *StyleContext) Factory() *StyleFactory {
	return c.factory
}

// Getters for accessing state (read-only)

// Theme returns the current theme
func (c *StyleContext) Theme() *ThemeAdapter {
	return c.theme
}

// SelectionState returns the current selection state
func (c *StyleContext) SelectionState() SelectionState {
	return c.selectionState
}

// SearchState returns the current search state
func (c *StyleContext) SearchState() SearchState {
	return c.searchState
}

// StyleProvider returns the style configuration provider
//
//nolint:ireturn // Returns interface by design - provides abstraction for styling system
func (c *StyleContext) StyleProvider() StyleProvider {
	return c.styleProvider
}

// IsSelected is a convenience method to check if currently selected
func (c *StyleContext) IsSelected() bool {
	return c.selectionState.IsSelected
}

// IsSearchActive is a convenience method to check if search is active
func (c *StyleContext) IsSearchActive() bool {
	return c.searchState.IsActive
}
