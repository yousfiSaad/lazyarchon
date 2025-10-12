package viewport

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ComposeWithScrollbar combines content with a scrollbar column
// This is a utility function that can be used by any component that needs scrollbar rendering
//
// Parameters:
//   - content: The content to display (multi-line string)
//   - scrollbar: The scrollbar characters (one per line, or nil for no scrollbar)
//   - panelWidth: The total panel width (including borders)
//   - targetHeight: Target height for padding (0 = no padding, content height used as-is)
//
// Returns: Content with scrollbar column appended, ready to be wrapped in a panel
//
// Note: Always uses default scrollbar options (gap char, width=4)
func ComposeWithScrollbar(content string, scrollbar []string, panelWidth int, targetHeight int) string {
	// Use default options
	opts := DefaultScrollbarOptions()

	// Split content into lines
	lines := strings.Split(content, "\n")

	// Pad lines to target height if specified
	if targetHeight > 0 {
		for len(lines) < targetHeight {
			lines = append(lines, "")
		}
	}

	lineCount := len(lines)
	panelContentWidth := panelWidth - 2 // Account for panel borders
	scrollbarColWidth := opts.Width
	targetContentWidth := panelContentWidth - scrollbarColWidth

	// Build each line: padded content + scrollbar column
	combined := make([]string, lineCount)
	for i := range lineCount { //nolint:varnamelen // i is idiomatic for loop index
		// Pad content to target width
		lineContent := lines[i]
		contentWidth := lipgloss.Width(lineContent)
		if contentWidth < targetContentWidth {
			lineContent += strings.Repeat(" ", targetContentWidth-contentWidth)
		}

		// Append scrollbar column
		var scrollbarCol string
		if scrollbar != nil && i < len(scrollbar) {
			// Build scrollbar column: gap + thumb/track + padding
			scrollbarCol = opts.GapChar +
				scrollbar[i] +
				strings.Repeat(" ", opts.Width-2)
		} else {
			// Empty scrollbar area
			scrollbarCol = strings.Repeat(" ", opts.Width)
		}

		combined[i] = lineContent + scrollbarCol
	}

	return strings.Join(combined, "\n")
}
