package detailspanel

import (
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/layout"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/styling"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/utils/view"
	sharedviewport "github.com/yousfisaad/lazyarchon/v2/internal/shared/viewport"
)

// DetailsPanelCore provides shared infrastructure for details panels (viewport, scrolling, rendering)
// This is domain-agnostic - it doesn't know about Tasks, Projects, or any specific content
// Core owns the viewport and renders it - styling is passed as parameters
// Note: Active state is NOT cached - it's passed as a parameter during rendering
type DetailsPanelCore struct {
	// Viewport for scrollable content
	viewport viewport.Model

	// Calculated content width (accounting for scrollbar)
	contentWidth int
}

// CoreOptions contains configuration for creating a details panel core
type CoreOptions struct {
	Width  int
	Height int
}

// NewCore creates a new details panel core with viewport infrastructure
func NewCore(opts CoreOptions) DetailsPanelCore {
	// Set default values
	if opts.Width == 0 {
		opts.Width = 40
	}
	if opts.Height == 0 {
		opts.Height = 20
	}

	// Calculate dimensions using dimension calculator
	// Always reserve scrollbar space to prevent content overflow when scrollbar appears
	// Details panels have no static headers/footers, so no reserved lines needed
	calc := layout.NewCalculator(opts.Width, opts.Height, layout.PanelComponent).
		WithScrollbar() // Reserve space for scrollbar (4 chars)
	dims := calc.Calculate()

	// Create viewport with calculated dimensions
	viewportInstance := viewport.New(dims.Content, dims.ViewportHeight)

	return DetailsPanelCore{
		viewport:     viewportInstance,
		contentWidth: dims.Content,
	}
}

// UpdateDimensions recalculates viewport dimensions when panel is resized
// Width and height are passed as parameters (not stored) to avoid duplication with BaseComponent
func (c *DetailsPanelCore) UpdateDimensions(width, height int) {
	// Recalculate using dimension calculator
	// Details panels have no static headers/footers, so no reserved lines needed
	calc := layout.NewCalculator(width, height, layout.PanelComponent).
		WithScrollbar()
	dims := calc.Calculate()

	// Update calculated dimensions
	c.contentWidth = dims.Content

	// Update viewport dimensions
	c.viewport.Width = dims.Content
	c.viewport.Height = dims.ViewportHeight
}

// SetContent updates the viewport content
func (c *DetailsPanelCore) SetContent(content string) {
	c.viewport.SetContent(content)
}

// GetContentWidth returns the calculated content width (accounting for scrollbar)
func (c *DetailsPanelCore) GetContentWidth() int {
	return c.contentWidth
}

// HandleScroll performs the specified scroll operation
func (c *DetailsPanelCore) HandleScroll(direction sharedviewport.ScrollDirection) {
	switch direction {
	case sharedviewport.ScrollUp:
		c.viewport.ScrollUp(1)
	case sharedviewport.ScrollDown:
		c.viewport.ScrollDown(1)
	case sharedviewport.ScrollToTop:
		c.viewport.GotoTop()
	case sharedviewport.ScrollToBottom:
		c.viewport.GotoBottom()
	case sharedviewport.ScrollFastUp:
		c.viewport.ScrollUp(4)
	case sharedviewport.ScrollFastDown:
		c.viewport.ScrollDown(4)
	case sharedviewport.ScrollHalfPageUp:
		c.viewport.HalfPageUp()
	case sharedviewport.ScrollHalfPageDown:
		c.viewport.HalfPageDown()
	}
}

// IsScrollable returns whether the content can be scrolled
func (c *DetailsPanelCore) IsScrollable() bool {
	return c.viewport.TotalLineCount() > c.viewport.Height
}

// AtTop returns whether the scroll position is at the top
func (c *DetailsPanelCore) AtTop() bool {
	return c.viewport.AtTop()
}

// AtBottom returns whether the scroll position is at the bottom
func (c *DetailsPanelCore) AtBottom() bool {
	return c.viewport.AtBottom()
}

// GetScrollPosition returns the current scroll position as a string
func (c *DetailsPanelCore) GetScrollPosition() string {
	if c.AtTop() {
		return ScrollPositionTop
	} else if c.AtBottom() {
		return ScrollPositionBottom
	}
	return ScrollPositionScrolled
}

// GetViewport returns the underlying viewport for direct access (e.g., mouse events)
func (c *DetailsPanelCore) GetViewport() *viewport.Model {
	return &c.viewport
}

// =============================================================================
// RENDERING METHODS
// =============================================================================

// RenderPanelWithScrollbar renders the viewport with panel borders and scrollbar
// isActive parameter determines border styling - no cached state
// Works for both populated content and empty states - viewport content fills the space
func (c DetailsPanelCore) RenderPanelWithScrollbar(
	width int,
	height int,
	isActive bool,
	styleContext *styling.StyleContext,
) string {
	// Get the visible portion of content from viewport
	viewportContent := c.viewport.View()

	// Generate scrollbar if content is scrollable, otherwise nil
	totalLines := c.viewport.TotalLineCount()
	viewportHeight := c.viewport.Height
	var scrollbar []string
	if totalLines > viewportHeight {
		// Generate scrollbar matching viewport height
		// The scrollbar height must match the viewport content height for proper alignment
		scrollbar = view.RenderScrollBarExact(c.viewport.YOffset, totalLines, viewportHeight)
	}

	// Always compose with scrollbar column (even if nil) to fill reserved scrollbar space
	// When scrollbar is nil, ComposeWithScrollbar fills the space with empty characters
	viewportContent = sharedviewport.ComposeWithScrollbar(viewportContent, scrollbar, width, viewportHeight)

	// Render with panel styling (using isActive parameter, not cached state)
	panelFactory := styleContext.Factory()
	detailStyle := panelFactory.Panel(width, height, isActive)

	return detailStyle.Render(viewportContent)
}
