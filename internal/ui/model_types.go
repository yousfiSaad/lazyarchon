package ui

import (
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/yousfisaad/lazyarchon/internal/archon"
	"time"
)

// ActiveView represents which panel is currently active for user input
type ActiveView int

const (
	LeftPanel  ActiveView = 0 // Task list panel
	RightPanel ActiveView = 1 // Task details panel
)

// WindowState manages UI dimensions and view state
type WindowState struct {
	width      int
	height     int
	ready      bool
	activeView ActiveView
}

// KeySequenceState tracks multi-key commands like 'gg'
type KeySequenceState struct {
	lastKeyPressed string
	lastKeyTime    time.Time
}

// HelpModalState manages help modal state
type HelpModalState struct {
	active bool
}

// StatusChangeModalState manages status change modal state
type StatusChangeModalState struct {
	active        bool
	selectedIndex int // Selected status option (0-3)
}

// ProjectModeState manages project selection state
type ProjectModeState struct {
	active        bool
	selectedIndex int
}

// ConfirmationModalState manages confirmation modal state
type ConfirmationModalState struct {
	active         bool
	message        string
	confirmText    string
	cancelText     string
	selectedOption int // 0 = confirm, 1 = cancel
}

// FeatureModeState manages feature selection modal state
type FeatureModeState struct {
	active           bool            // Whether feature modal is open
	selectedIndex    int             // Currently highlighted feature in list
	selectedFeatures map[string]bool // Which features are enabled/disabled
	backupFeatures   map[string]bool // Backup of features before modal opened (for cancel)
}

// TaskEditModalState manages task editing modal state
type TaskEditModalState struct {
	active         bool   // Whether edit modal is open
	selectedIndex  int    // Currently highlighted option in list
	newFeatureName string // Text input for new feature creation
	isCreatingNew  bool   // Whether user is in "create new feature" mode
	currentField   string // Which field is being edited (extensible for future)
}

// ModalState groups all modal-related state
type ModalState struct {
	help         HelpModalState
	statusChange StatusChangeModalState
	projectMode  ProjectModeState
	confirmation ConfirmationModalState
	featureMode  FeatureModeState
	taskEdit     TaskEditModalState
}

// NavigationState manages movement and scrolling
type NavigationState struct {
	selectedIndex int
	keySequence   KeySequenceState
}

// DataState manages API data and loading state
type DataState struct {
	tasks             []archon.Task
	projects          []archon.Project
	selectedProjectID *string // nil = "All tasks", otherwise project UUID
	loading           bool
	error             string
	sortMode          int
}

// Model represents the state of the application using composition
type Model struct {
	// Core infrastructure
	client *archon.Client

	// Feature-focused state groups
	Window     WindowState
	Modals     ModalState
	Navigation NavigationState
	Data       DataState

	// UI Components
	taskDetailsViewport viewport.Model
	helpModalViewport   viewport.Model
}

// ScrollContext defines which part of the UI is being scrolled
type ScrollContext int

const (
	HelpContext    ScrollContext = 0 // Help modal
	DetailsContext ScrollContext = 1 // Task details panel
	ListContext    ScrollContext = 2 // Task list panel
)
