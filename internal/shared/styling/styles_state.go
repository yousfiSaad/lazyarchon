package styling

// SelectionState encapsulates all selection-related styling state
type SelectionState struct {
	IsSelected      bool    // Whether this item is currently selected
	BackgroundColor string  // Background color for selected items (e.g., "238")
	ForegroundBoost float32 // Brightness multiplier for selected items (1.0 = no change)
}

// SearchState encapsulates search highlighting configuration
type SearchState struct {
	IsActive   bool   // Whether search highlighting is currently active
	Query      string // The search query string
	MatchColor string // Color for highlighted search matches
}

// NewSelectionState creates a default selection state
func NewSelectionState() SelectionState {
	return SelectionState{
		IsSelected:      false,
		BackgroundColor: "238", // Subtle gray background
		ForegroundBoost: 1.0,   // No brightness change by default
	}
}

// NewSearchState creates a default search state
func NewSearchState() SearchState {
	return SearchState{
		IsActive:   false,
		Query:      "",
		MatchColor: "226", // Bright yellow for search matches
	}
}

// WithSelected returns a new SelectionState with IsSelected set
func (s SelectionState) WithSelected(selected bool) SelectionState {
	s.IsSelected = selected
	return s
}

// WithBackgroundColor returns a new SelectionState with custom background color
func (s SelectionState) WithBackgroundColor(color string) SelectionState {
	s.BackgroundColor = color
	return s
}

// WithForegroundBoost returns a new SelectionState with custom brightness boost
func (s SelectionState) WithForegroundBoost(boost float32) SelectionState {
	s.ForegroundBoost = boost
	return s
}

// WithQuery returns a new SearchState with the specified query and activates search
func (s SearchState) WithQuery(query string) SearchState {
	s.IsActive = query != ""
	s.Query = query
	return s
}

// WithMatchColor returns a new SearchState with custom match highlighting color
func (s SearchState) WithMatchColor(color string) SearchState {
	s.MatchColor = color
	return s
}
