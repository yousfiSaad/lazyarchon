package layout

const (
	// PanelBorderWidth is the horizontal/vertical space consumed by panel borders
	// Panel uses RoundedBorder which takes 1 char on each side = 2 total
	PanelBorderWidth = 2

	// ScrollbarWidth is the total width of the scrollbar column
	// Includes: gap (1) + scrollbar char (1) + padding (2) = 4 total
	ScrollbarWidth = 4

	// ModalPadding is the standard padding used in modals
	ModalPadding = 1
)

// ComponentType defines different component layout needs
type ComponentType int

const (
	// PanelComponent represents main panels (TaskList, TaskDetails, ProjectList)
	// These have borders but no padding
	PanelComponent ComponentType = iota

	// ModalComponent represents modal dialogs (Help, Feature, StatusFilter)
	// These have borders AND padding
	ModalComponent

	// ItemComponent represents list items (TaskItem, ProjectItem)
	// These render within panels
	ItemComponent
)

// DimensionCalculator provides consistent dimension calculations for components
// It handles the complexity of borders, padding, scrollbars, and reserved space
type DimensionCalculator struct {
	totalWidth    int
	totalHeight   int
	componentType ComponentType
	hasScrollbar  bool
	customPadding int
	reservedLines int
}

// NewCalculator creates a new dimension calculator
// Parameters:
//   - width: Total allocated width for the component
//   - height: Total allocated height for the component
//   - compType: Type of component (Panel, Modal, or Item)
func NewCalculator(width, height int, compType ComponentType) *DimensionCalculator {
	return &DimensionCalculator{
		totalWidth:    width,
		totalHeight:   height,
		componentType: compType,
	}
}

// WithScrollbar indicates this component will render a scrollbar
// Reduces content width by ScrollbarWidth
func (d *DimensionCalculator) WithScrollbar() *DimensionCalculator {
	d.hasScrollbar = true
	return d
}

// WithPadding sets custom padding (overrides default modal padding)
// Padding is applied on both left and right (total reduction = padding * 2)
func (d *DimensionCalculator) WithPadding(padding int) *DimensionCalculator {
	d.customPadding = padding
	return d
}

// WithReservedLines sets the number of lines reserved for headers/footers
// These lines are subtracted from viewport height
func (d *DimensionCalculator) WithReservedLines(lines int) *DimensionCalculator {
	d.reservedLines = lines
	return d
}

// PanelContentWidth returns the width available inside panel borders
// This is the width that Panel() style will use for content
func (d *DimensionCalculator) PanelContentWidth() int {
	return d.totalWidth - PanelBorderWidth
}

// PanelContentHeight returns the height available inside panel borders
// This is the height that Panel() style will use for content
func (d *DimensionCalculator) PanelContentHeight() int {
	return d.totalHeight - PanelBorderWidth
}

// ViewportWidth returns the width available for viewport content
// Accounts for borders and component-specific padding
func (d *DimensionCalculator) ViewportWidth() int {
	width := d.PanelContentWidth()

	// Apply padding based on component type
	padding := d.customPadding
	if padding == 0 && d.componentType == ModalComponent {
		padding = ModalPadding
	}
	width -= padding * 2 // Left and right padding

	return width
}

// ViewportHeight returns the height available for viewport content
// Accounts for borders and reserved lines (headers, footers, etc.)
func (d *DimensionCalculator) ViewportHeight() int {
	height := d.PanelContentHeight()

	// Subtract reserved lines (headers, footers, etc.)
	height -= d.reservedLines

	return height
}

// ContentWidth returns the actual content width after all decorations
// This is the width available for rendering actual content
// Accounts for borders, padding, AND scrollbar if present
func (d *DimensionCalculator) ContentWidth() int {
	width := d.ViewportWidth()

	if d.hasScrollbar {
		width -= ScrollbarWidth
	}

	return width
}

// Calculate returns all calculated dimensions at once
// Useful when you need multiple dimension values
func (d *DimensionCalculator) Calculate() Dimensions {
	return Dimensions{
		Total:              d.totalWidth,
		TotalHeight:        d.totalHeight,
		PanelContent:       d.PanelContentWidth(),
		Viewport:           d.ViewportWidth(),
		Content:            d.ContentWidth(),
		PanelContentHeight: d.PanelContentHeight(),
		ViewportHeight:     d.ViewportHeight(),
	}
}

// Dimensions holds all calculated dimension values
type Dimensions struct {
	Total              int // Total allocated width
	TotalHeight        int // Total allocated height
	PanelContent       int // Width inside panel borders
	Viewport           int // Width available for viewport
	Content            int // Width available for actual content (after scrollbar)
	PanelContentHeight int // Height inside panel borders
	ViewportHeight     int // Height available for viewport
}
