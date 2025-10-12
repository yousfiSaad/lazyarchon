package view

// Core utility functions for the view system
// These are basic mathematical and calculation utilities used throughout the UI

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

// CalculateScrollWindow calculates the start and end indices for center-focus scrolling
func CalculateScrollWindow(totalItems, selectedIndex, maxItems int) (int, int) {
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

// max is a helper function for internal use
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
