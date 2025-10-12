package view

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Scrolling module handles content scrolling and scroll bar rendering
// This module provides functionality for handling scrollable content with visual feedback

// ApplyScrolling applies scrolling to content and adds scroll indicators
func ApplyScrolling(content []string, scrollOffset, maxLines int, scrollIndicatorStyle lipgloss.Style) []string {
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
				visibleContent[len(visibleContent)-1] = scrollIndicatorStyle.Render(scrollInfo)
			} else {
				visibleContent = append(visibleContent, scrollIndicatorStyle.Render(scrollInfo))
			}
		}
	}

	return visibleContent
}

// ApplyScrollingWithScrollBar applies scrolling to content with visual scroll bar and enhanced feedback
func ApplyScrollingWithScrollBar(content []string, scrollOffset, maxLines int, scrollIndicatorStyle lipgloss.Style) ([]string, []string) {
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
	scrollBar := RenderDetailScrollBar(start, totalLines, maxLines)

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
			visibleContent[len(visibleContent)-1] = scrollIndicatorStyle.Render(scrollInfo)
		} else {
			visibleContent = append(visibleContent, scrollIndicatorStyle.Render(scrollInfo))
		}
	}

	return visibleContent, scrollBar
}

// RenderDetailScrollBar generates ASCII scroll bar for task details panel
func RenderDetailScrollBar(currentPos, totalLines, viewportHeight int) []string {
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

// RenderScrollBar generates ASCII scroll bar for visual position feedback
func RenderScrollBar(currentPos, totalItems, viewportHeight int) []string {
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

// RenderScrollBarExact generates ASCII scroll bar using exact height without reduction
// This is used when the calling component has already calculated the exact usable height
func RenderScrollBarExact(currentPos, totalItems, exactHeight int) []string {
	if totalItems <= exactHeight {
		return nil // No scroll bar needed when all items fit
	}

	// Use exact height without any reduction
	trackHeight := exactHeight
	if trackHeight < 3 {
		trackHeight = 3 // Minimum usable height
	}

	// Calculate thumb position and size
	thumbSize := max(1, (exactHeight*trackHeight)/totalItems) // Proportional thumb size
	maxThumbPos := trackHeight - thumbSize
	thumbPos := (currentPos * maxThumbPos) / max(1, totalItems-exactHeight)

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
