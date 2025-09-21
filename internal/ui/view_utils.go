package ui

import (
	"github.com/yousfisaad/lazyarchon/internal/ui/view"
	"github.com/yousfisaad/lazyarchon/internal/ui/styling"
)

// Min returns the minimum of two integers
func Min(a, b int) int {
	return view.Min(a, b)
}

// Max returns the maximum of two integers
func Max(a, b int) int {
	return view.Max(a, b)
}

// calculateScrollWindow calculates the start and end indices for center-focus scrolling
func calculateScrollWindow(totalItems, selectedIndex, maxItems int) (int, int) {
	return view.CalculateScrollWindow(totalItems, selectedIndex, maxItems)
}

// applyScrolling applies scrolling to content and adds scroll indicators
func (m Model) applyScrolling(content []string, scrollOffset, maxLines int) []string {
	return view.ApplyScrolling(content, scrollOffset, maxLines, styling.ScrollIndicatorStyle)
}

// applyScrollingWithScrollBar applies scrolling to content with visual scroll bar and enhanced feedback
func (m Model) applyScrollingWithScrollBar(content []string, scrollOffset, maxLines int) ([]string, []string) {
	return view.ApplyScrollingWithScrollBar(content, scrollOffset, maxLines, styling.ScrollIndicatorStyle)
}

// renderDetailScrollBar generates ASCII scroll bar for task details panel
func renderDetailScrollBar(currentPos, totalLines, viewportHeight int) []string {
	return view.RenderDetailScrollBar(currentPos, totalLines, viewportHeight)
}

// renderMarkdown renders markdown text using glamour for terminal display
func renderMarkdown(text string, width int) string {
	return view.RenderMarkdown(text, width)
}

// wordWrap wraps text to fit within the specified width (fallback function)
func wordWrap(text string, width int) string {
	return view.WordWrap(text, width)
}

// renderScrollBar generates ASCII scroll bar for visual position feedback
func renderScrollBar(currentPos, totalItems, viewportHeight int) []string {
	return view.RenderScrollBar(currentPos, totalItems, viewportHeight)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// highlightSearchTerms highlights search query matches in text using lipgloss styling
func highlightSearchTerms(text, query string) string {
	return view.HighlightSearchTerms(text, query)
}

// isColorConflictWithHighlight checks if the given color conflicts with highlight background
// Returns true if the colors have poor contrast with bright yellow background causing poor readability
func isColorConflictWithHighlight(textColor string) bool {
	return view.IsColorConflictWithHighlight(textColor)
}

// highlightSearchTermsWithColor highlights search query matches while preserving specified text color
func highlightSearchTermsWithColor(text, query, textColor string) string {
	return view.HighlightSearchTermsWithColor(text, query, textColor)
}

// stripANSI removes ANSI escape sequences from text for accurate length calculation
func stripANSI(text string) string {
	return view.StripANSI(text)
}

// truncatePreservingANSI truncates text while preserving ANSI escape sequences
func truncatePreservingANSI(text string, maxWidth int) string {
	return view.TruncatePreservingANSI(text, maxWidth)
}
