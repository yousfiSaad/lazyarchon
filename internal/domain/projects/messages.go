package projects

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/internal/archon"
)

// =============================================================================
// PROJECT DOMAIN MESSAGES
// =============================================================================
// Messages representing project-related domain events and operations

// ProjectsLoadedMsg is sent when projects are loaded from the API
type ProjectsLoadedMsg struct {
	Projects []archon.Project
	Error    error
}

// Ensure all message types implement tea.Msg
var (
	_ tea.Msg = ProjectsLoadedMsg{}
)
