package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/internal/ui/input"
)

// handleHelpModeInput handles keyboard input when help modal is open
func (m Model) handleHelpModeInput(key string) (Model, tea.Cmd) {
	if !input.IsHelpModalKey(key) {
		return m, nil // Ignore keys not handled by help modal
	}

	action := input.GetHelpModalAction(key)
	switch action {
	case "toggle", "close":
		m.SetHelpMode(false)
		return m, nil
	case "down1":
		m.helpModalViewport.LineDown(1)
		return m, nil
	case "up1":
		m.helpModalViewport.LineUp(1)
		return m, nil
	case "down4":
		m.helpModalViewport.LineDown(4)
		return m, nil
	case "up4":
		m.helpModalViewport.LineUp(4)
		return m, nil
	case "halfup":
		m.helpModalViewport.HalfViewUp()
		return m, nil
	case "halfdown":
		m.helpModalViewport.HalfViewDown()
		return m, nil
	case "top":
		m.helpModalViewport.GotoTop()
		return m, nil
	case "bottom":
		m.helpModalViewport.GotoBottom()
		return m, nil
	case "quit":
		return m, tea.Quit
	default:
		return m, nil
	}
}