package projectdetails

import (
	"github.com/yousfisaad/lazyarchon/v2/internal/archon"
	sharedviewport "github.com/yousfisaad/lazyarchon/v2/internal/shared/viewport"
)

// ProjectDetailsUpdateMsg updates the project being displayed
type ProjectDetailsUpdateMsg struct {
	SelectedProject *archon.Project
}

// ProjectDetailsScrollMsg handles scrolling in the project details panel
type ProjectDetailsScrollMsg struct {
	Direction sharedviewport.ScrollDirection
}

// NOTE: ProjectDetailsSetActiveMsg removed - components read active state from UIState directly

// ProjectDetailsResizeMsg resizes the project details panel
type ProjectDetailsResizeMsg struct {
	Width  int
	Height int
}

// ProjectDetailsScrollPositionChangedMsg is broadcast when scroll position changes
type ProjectDetailsScrollPositionChangedMsg struct {
	Position string // Use detailspanel.ScrollPosition* constants
}
