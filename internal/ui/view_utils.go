package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
)

// Min returns the minimum of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximum of two integers
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// calculateScrollWindow calculates the start and end indices for center-focus scrolling
func calculateScrollWindow(totalItems, selectedIndex, maxItems int) (int, int) {
	if totalItems <= maxItems {
		return 0, totalItems // All items fit, no scrolling needed
	}

	// Try to center the selected item for better UX
	halfView := maxItems / 2
	start := selectedIndex - halfView

	// Handle edge cases where centering isn't possible
	if start < 0 {
		start = 0 // At top edge, align to top
	} else if start+maxItems > totalItems {
		start = totalItems - maxItems // At bottom edge, align to bottom
	}

	end := start + maxItems
	return start, end
}

// applyScrolling applies scrolling to content and adds scroll indicators
func (m Model) applyScrolling(content []string, scrollOffset, maxLines int) []string {
	totalLines := len(content)

	// Calculate scroll window
	start := scrollOffset
	end := start + maxLines

	// Clamp scroll position
	if totalLines <= maxLines {
		// All content fits, no scrolling needed
		return content
	} else {
		// Scrolling needed
		maxScroll := totalLines - maxLines
		if start > maxScroll {
			start = maxScroll
		}
		if start < 0 {
			start = 0
		}
		end = start + maxLines
		if end > totalLines {
			end = totalLines
		}
	}

	// Get visible content
	visibleContent := content[start:end]

	// Add scroll indicator if needed
	if totalLines > maxLines {
		scrollInfo := fmt.Sprintf("[Lines %d-%d of %d]", start+1, end, totalLines)
		if len(visibleContent) > 0 {
			// Replace last line with scroll info or add if there's space
			if len(visibleContent) == maxLines {
				visibleContent[len(visibleContent)-1] = ScrollIndicatorStyle.Render(scrollInfo)
			} else {
				visibleContent = append(visibleContent, ScrollIndicatorStyle.Render(scrollInfo))
			}
		}
	}

	return visibleContent
}

// applyScrollingWithScrollBar applies scrolling to content with visual scroll bar and enhanced feedback
func (m Model) applyScrollingWithScrollBar(content []string, scrollOffset, maxLines int) ([]string, []string) {
	totalLines := len(content)

	// If all content fits, no scrolling needed
	if totalLines <= maxLines {
		return content, nil
	}

	// Use scroll offset as-is since model layer now handles bounds checking
	start := scrollOffset

	// Only apply minimal safety bounds to prevent array out of bounds errors
	if start < 0 {
		start = 0
	}

	// Allow the model to determine the maximum - only clamp if absolutely necessary
	maxScroll := totalLines - maxLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	if start > maxScroll {
		start = maxScroll
	}

	end := start + maxLines
	if end > totalLines {
		end = totalLines
	}

	// Get visible content
	visibleContent := content[start:end]

	// Generate scroll bar
	scrollBar := renderDetailScrollBar(start, totalLines, maxLines)

	// Add enhanced position indicator
	if len(visibleContent) > 0 && totalLines > maxLines {
		percentage := ((end * 100) / totalLines)
		if percentage > 100 {
			percentage = 100
		}

		scrollInfo := fmt.Sprintf("[Lines %d-%d of %d (%d%%)]",
			start+1, end, totalLines, percentage)

		// Replace last line with scroll info
		if len(visibleContent) == maxLines {
			visibleContent[len(visibleContent)-1] = ScrollIndicatorStyle.Render(scrollInfo)
		} else {
			visibleContent = append(visibleContent, ScrollIndicatorStyle.Render(scrollInfo))
		}
	}

	return visibleContent, scrollBar
}

// renderDetailScrollBar generates ASCII scroll bar for task details panel
func renderDetailScrollBar(currentPos, totalLines, viewportHeight int) []string {
	if totalLines <= viewportHeight {
		return nil // No scroll bar needed
	}

	// Calculate scroll bar dimensions
	trackHeight := viewportHeight - 1 // Account for scroll indicator line
	if trackHeight < 3 {
		trackHeight = 3
	}

	// Calculate thumb position and size
	thumbSize := max(1, (viewportHeight*trackHeight)/totalLines)
	maxThumbPos := trackHeight - thumbSize
	thumbPos := (currentPos * maxThumbPos) / max(1, totalLines-viewportHeight)

	// Generate scroll bar lines
	var scrollBar []string
	for i := 0; i < trackHeight; i++ {
		if i >= thumbPos && i < thumbPos+thumbSize {
			scrollBar = append(scrollBar, "▓") // Thumb
		} else {
			scrollBar = append(scrollBar, "░") // Track
		}
	}

	return scrollBar
}

// renderMarkdown renders markdown text using glamour for terminal display
func renderMarkdown(text string, width int) string {
	// Create a glamour renderer with appropriate styling
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),     // Auto-detect dark/light terminal
		glamour.WithWordWrap(width), // Set word wrap to available width
	)
	if err != nil {
		// Fallback to simple word wrap if glamour fails
		return wordWrap(text, width)
	}

	// Render the markdown
	rendered, err := renderer.Render(text)
	if err != nil {
		// Fallback to simple word wrap if rendering fails
		return wordWrap(text, width)
	}

	// Remove trailing newline that glamour adds
	return strings.TrimSuffix(rendered, "\n")
}

// wordWrap wraps text to fit within the specified width (fallback function)
func wordWrap(text string, width int) string {
	if len(text) <= width {
		return text
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	currentLine := words[0]

	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	lines = append(lines, currentLine)

	return strings.Join(lines, "\n")
}

// renderScrollBar generates ASCII scroll bar for visual position feedback
func renderScrollBar(currentPos, totalItems, viewportHeight int) []string {
	if totalItems <= viewportHeight {
		return nil // No scroll bar needed when all items fit
	}

	// Calculate scroll bar dimensions
	trackHeight := viewportHeight - 2 // Account for "Tasks:" header and spacing
	if trackHeight < 3 {
		trackHeight = 3 // Minimum usable height
	}

	// Calculate thumb position and size
	thumbSize := max(1, (viewportHeight*trackHeight)/totalItems) // Proportional thumb size
	maxThumbPos := trackHeight - thumbSize
	thumbPos := (currentPos * maxThumbPos) / max(1, totalItems-viewportHeight)

	// Generate scroll bar lines
	var scrollBar []string
	for i := 0; i < trackHeight; i++ {
		if i >= thumbPos && i < thumbPos+thumbSize {
			scrollBar = append(scrollBar, "▓") // Thumb (filled)
		} else {
			scrollBar = append(scrollBar, "░") // Track (light)
		}
	}

	return scrollBar
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
