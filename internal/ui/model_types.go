package ui

import (
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/yousfisaad/lazyarchon/internal/archon"
	"github.com/yousfisaad/lazyarchon/internal/config"
	"github.com/yousfisaad/lazyarchon/internal/interfaces"
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

	// Search functionality
	searchMode       bool     // Whether user is actively typing search
	searchInput      string   // Current search input (while typing)
	searchQuery      string   // Committed search query
	filteredFeatures []string // Cached filtered results
	matchingIndices  []int    // Indices of features that match search (in filtered list)
	currentMatchIndex int     // Current position in match list for n/N navigation
}

// TaskEditModalState manages task editing modal state
type TaskEditModalState struct {
	active         bool   // Whether edit modal is open
	selectedIndex  int    // Currently highlighted option in list
	newFeatureName string // Text input for new feature creation
	isCreatingNew  bool   // Whether user is in "create new feature" mode
	currentField   string // Which field is being edited (extensible for future)
}

// StatusFilterModalState manages status filter modal state
type StatusFilterModalState struct {
	active           bool            // Whether modal is open
	selectedIndex    int             // Currently highlighted status
	selectedStatuses map[string]bool // Which statuses are enabled
	backupStatuses   map[string]bool // Backup for cancel functionality
}

// ModalState groups all modal-related state
type ModalState struct {
	help         HelpModalState
	statusChange StatusChangeModalState
	projectMode  ProjectModeState
	confirmation ConfirmationModalState
	featureMode  FeatureModeState
	taskEdit     TaskEditModalState
	statusFilter StatusFilterModalState
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
	loadingMessage    string // Context-specific loading message
	error             string
	sortMode          int
	spinnerIndex      int    // Current spinner animation frame
	lastRetryError    string // Last error for retry functionality
	connected         bool   // Connection status to Archon server

	// Search functionality
	searchQuery       string   // Current search query
	searchHistory     []string // Recent search queries
	searchActive      bool     // Whether search is currently active
	searchMode        bool     // Whether user is actively typing in status bar search
	searchInput       string   // Current real-time search input (while typing)

	// Match tracking for n/N navigation
	matchingTaskIndices []int // Indices of tasks that match the search (in sorted order)
	currentMatchIndex   int   // Current position in the match list (0-based)
	totalMatches        int   // Total number of matching tasks

	// Status messages
	statusMessage     string    // Temporary status message (for copy confirmations, etc.)
	statusMessageTime time.Time // When the status message was set

	// Status filtering
	statusFilters     map[string]bool // Status visibility (todo, doing, review, done)
	statusFilterActive bool           // Whether custom status filtering is active
}

// Model represents the state of the application using composition
type Model struct {
	// Core infrastructure (keeping concrete types for compatibility)
	client   *archon.Client
	wsClient interfaces.RealtimeClient
	config   *config.Config

	// Feature-focused state groups
	Window     WindowState
	Modals     ModalState
	Navigation NavigationState
	Data       DataState

	// UI Components
	taskDetailsViewport viewport.Model
	helpModalViewport   viewport.Model

	// Dependencies (for gradual migration to interfaces)
	deps *ModelDependencies
}

// ModelDependencies holds interface-based dependencies
type ModelDependencies struct {
	ArchonClient        interfaces.ArchonClient
	ConfigProvider      interfaces.ConfigProvider
	ViewportFactory     interfaces.ViewportFactory
	StyleContextProvider interfaces.StyleContextProvider
	CommandExecutor     interfaces.CommandExecutor
	Logger              interfaces.Logger
	HealthChecker       interfaces.HealthChecker
}

// ScrollContext defines which part of the UI is being scrolled
type ScrollContext int

const (
	HelpContext    ScrollContext = 0 // Help modal
	DetailsContext ScrollContext = 1 // Task details panel
	ListContext    ScrollContext = 2 // Task list panel
)
