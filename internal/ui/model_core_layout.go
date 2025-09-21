package ui

import (
	"github.com/yousfisaad/lazyarchon/internal/ui/styling"
)

// GetContentHeight returns the available height for content panels
func (m Model) GetContentHeight() int {
	return m.Window.height - styling.HeaderHeight - styling.StatusBarHeight - 1 // -1 for spacing
}

// GetLeftPanelWidth returns the width of the left panel
func (m Model) GetLeftPanelWidth() int {
	return m.Window.width / 2
}

// GetRightPanelWidth returns the width of the right panel
func (m Model) GetRightPanelWidth() int {
	return m.Window.width - m.GetLeftPanelWidth()
}