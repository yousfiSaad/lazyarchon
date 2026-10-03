package worktree

import (
	"maps"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/layout"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/utils/keys"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/utils/view"
	sharedviewport "github.com/yousfisaad/lazyarchon/v2/internal/shared/viewport"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/base"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/messages"
)

const ComponentID = "worktree-modal"

// Navigation constants
const (
	fastScrollDistance   = 4  // Number of items to skip for J/K fast scroll
	halfPageSize         = 5  // Number of items for ctrl+u/d half-page navigation
	searchInputMaxLength = 50 // Maximum characters in search input
)

// WorktreeModel represents the worktree selection modal component
// Architecture: Follows four-tier state pattern (clone of the feature modal)
// - No source data caching (receives worktree list via ShowWorktreeModalMsg, not from ProgramContext)
// - No display parameters (manages its own filtering and rendering)
// - Owned state: All fields are component-local concerns for modal interaction
// - No transient feedback (modal lifecycle managed by MainModel)
// - Modal lifecycle managed by BaseModal (active/visible state)
type WorktreeModel struct {
	base.BaseModal

	// ===================================================================
	// OWNED STATE - Component manages these directly
	// ===================================================================

	// Core worktree state (passed via message, managed locally during modal session)
	allWorktrees      []string        // All available worktrees (from ShowWorktreeModalMsg)
	selectedWorktrees map[string]bool // Currently selected worktrees
	backupWorktrees   map[string]bool // Backup for cancel functionality

	// Navigation state
	selectedIndex     int      // Currently highlighted worktree
	filteredWorktrees []string // Worktrees after search filtering

	// Search state (modal-local search, independent of global search)
	searchMode        bool   // Whether actively typing search
	searchInput       string // Current search input
	searchQuery       string // Committed search query
	matchingIndices   []int  // Indices of worktrees matching search
	currentMatchIndex int    // Current position in match list for n/N navigation

	// UI components
	viewport viewport.Model // Viewport for smooth scrolling
}

// NewModel creates a new worktree modal component
func NewModel(context *base.ComponentContext) *WorktreeModel {
	baseModal := base.NewBaseModal(
		ComponentID,
		base.WorktreeModalComponent,
		context,
	)

	// Initialize viewport for worktree list scrolling with reasonable defaults
	// These will be updated properly when the modal is shown and screen size is known
	vp := viewport.New(60, 15) //nolint:varnamelen // vp is idiomatic for viewport
	vp.SetContent("")          // Start with empty content

	model := &WorktreeModel{
		BaseModal:         baseModal,
		selectedWorktrees: make(map[string]bool),
		backupWorktrees:   make(map[string]bool),
		filteredWorktrees: []string{},
		matchingIndices:   []int{},
		selectedIndex:     0,
		searchMode:        false,
		searchInput:       "",
		searchQuery:       "",
		currentMatchIndex: 0,
		viewport:          vp,
	}
	// Set dimensions using base component
	model.SetDimensions(80, 20) // Wide enough for worktree list, height for worktree list + search
	return model
}

// CanFocus overrides the base implementation to allow focus
func (m *WorktreeModel) CanFocus() bool {
	return true
}

// Init initializes the worktree modal component
func (m *WorktreeModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the worktree modal component
// High complexity (20) due to comprehensive message routing for modal state management
//
//nolint:gocyclo // Modal requires routing to 12+ message types for complete worktree selection UX
func (m *WorktreeModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case ShowWorktreeModalMsg:
		m.SetActive(true)
		m.SetFocus(true)
		m.allWorktrees = msg.AllWorktrees

		// Create backup of current selection for cancel functionality
		m.backupWorktrees = make(map[string]bool)
		maps.Copy(m.backupWorktrees, msg.SelectedWorktrees)

		// Set current selection
		m.selectedWorktrees = make(map[string]bool)
		maps.Copy(m.selectedWorktrees, msg.SelectedWorktrees)

		// Reset state
		m.selectedIndex = 0
		m.searchMode = false
		m.searchInput = ""
		m.searchQuery = ""
		m.updateFilteredWorktrees()

		return m.BroadcastMessage(messages.ModalStateMsg{
			Type:   string(base.ModalTypeWorktree),
			Active: true,
		})

	case HideWorktreeModalMsg:
		m.SetActive(false)
		m.SetFocus(false)
		m.searchMode = false
		m.searchInput = ""
		return m.BroadcastMessage(messages.ModalStateMsg{
			Type:   string(base.ModalTypeWorktree),
			Active: false,
		})

	case tea.WindowSizeMsg:
		m.updateDimensions(msg.Width, msg.Height)
		if m.IsActive() {
			m.updateViewport()
		}
		return nil

	case tea.KeyMsg:
		if !m.IsActive() || !m.IsFocused() {
			return nil
		}
		return m.handleKeyPress(msg)

	case WorktreeModalSearchMsg:
		if !m.IsActive() {
			return nil
		}
		m.searchQuery = msg.Query
		m.updateFilteredWorktrees()
		return nil

	case WorktreeModalScrollMsg:
		if !m.IsActive() {
			return nil
		}
		return m.handleScroll(msg)

	case WorktreeModalToggleMsg:
		if !m.IsActive() {
			return nil
		}
		m.toggleWorktree(msg.Worktree)
		return nil

	case WorktreeModalClearSearchMsg:
		if !m.IsActive() {
			return nil
		}
		m.clearSearch()
		return nil

	case WorktreeModalSelectAllMsg:
		if !m.IsActive() {
			return nil
		}
		m.selectAllVisible()
		return nil

	case WorktreeModalDeselectAllMsg:
		if !m.IsActive() {
			return nil
		}
		m.deselectAll()
		return nil

	default:
		return nil
	}
}

// View renders the worktree modal
func (m *WorktreeModel) View() string {
	if !m.IsActive() {
		return ""
	}

	return m.renderModal()
}

// handleKeyPress processes keyboard input for the worktree modal
func (m *WorktreeModel) handleKeyPress(key tea.KeyMsg) tea.Cmd {
	keyString := key.String()

	// Handle search mode
	if m.searchMode {
		return m.handleSearchMode(keyString)
	}

	// Handle selection mode
	return m.handleSelectionMode(keyString)
}

// handleSearchMode handles input when in search mode
func (m *WorktreeModel) handleSearchMode(keyString string) tea.Cmd {
	switch keyString {
	case keys.KeyEscape:
		// Exit search mode
		m.searchMode = false
		return nil

	case keys.KeyEnter:
		// Commit search and exit search mode
		m.searchQuery = m.searchInput
		m.searchMode = false
		m.updateFilteredWorktrees()
		return nil

	case keys.KeyBackspace:
		// Remove last character
		if len(m.searchInput) > 0 {
			m.searchInput = m.searchInput[:len(m.searchInput)-1]
			m.updateLiveSearch()
		}
		return nil

	case keys.KeyCtrlU:
		// Clear entire search input
		m.searchInput = ""
		m.updateLiveSearch()
		return nil

	case keys.KeyCtrlC:
		return tea.Quit

	default:
		// Add character to search input
		if len(keyString) == 1 && len(m.searchInput) < searchInputMaxLength {
			m.searchInput += keyString
			m.updateLiveSearch()
		}
		return nil
	}
}

// handleSelectionMode handles input when in selection mode
func (m *WorktreeModel) handleSelectionMode(keyString string) tea.Cmd {
	// Try modal control keys first
	if cmd := m.handleModalKeys(keyString); cmd != nil {
		return cmd
	}

	// Then navigation keys
	if cmd := m.handleNavigationKeys(keyString); cmd != nil {
		return cmd
	}

	// Then search keys
	if cmd := m.handleSearchKeys(keyString); cmd != nil {
		return cmd
	}

	// Finally selection keys
	if cmd := m.handleSelectionKeys(keyString); cmd != nil {
		return cmd
	}

	return nil
}

// handleModalKeys handles modal control keys (escape, enter, quit)
func (m *WorktreeModel) handleModalKeys(keyString string) tea.Cmd {
	switch keyString {
	case keys.KeyEscape, keys.KeyQ:
		// Cancel and restore backup
		m.selectedWorktrees = make(map[string]bool)
		maps.Copy(m.selectedWorktrees, m.backupWorktrees)
		return m.BroadcastMessage(HideWorktreeModalMsg{})

	case keys.KeyEnter:
		// Apply selection and close modal
		return tea.Batch(
			m.BroadcastMessage(WorktreeSelectionAppliedMsg{
				SelectedWorktrees: m.copySelectedWorktrees(),
			}),
			m.BroadcastMessage(HideWorktreeModalMsg{}),
		)

	case keys.KeyCtrlC:
		return tea.Quit
	}
	return nil
}

// handleNavigationKeys handles navigation keys (j/k, J/K, gg/G, ctrl+u/d)
func (m *WorktreeModel) handleNavigationKeys(keyString string) tea.Cmd {
	switch keyString {
	case keys.KeyJ, keys.KeyArrowDown:
		m.navigateDown()
		return nil

	case keys.KeyK, keys.KeyArrowUp:
		m.navigateUp()
		return nil

	case keys.KeyJCap:
		// Fast scroll down (4 items)
		m.navigateFastDown()
		return nil

	case keys.KeyKCap:
		// Fast scroll up (4 items)
		m.navigateFastUp()
		return nil

	case keys.KeyCtrlU, keys.KeyPgUp:
		// Half-page up
		m.navigateHalfPageUp()
		return nil

	case keys.KeyCtrlD, keys.KeyPgDn:
		// Half-page down
		m.navigateHalfPageDown()
		return nil

	case keys.KeyGG, keys.KeyHome:
		// Jump to first item
		m.jumpToFirst()
		return nil

	case keys.KeyGCap, keys.KeyEnd:
		// Jump to last item
		m.jumpToLast()
		return nil
	}
	return nil
}

// handleSearchKeys handles search-related keys (/, n, N, ctrl+l/x)
func (m *WorktreeModel) handleSearchKeys(keyString string) tea.Cmd {
	switch keyString {
	case keys.KeySlash:
		// Activate search mode
		m.searchMode = true
		m.searchInput = ""
		return nil

	case keys.KeyN:
		// Next search match
		if len(m.matchingIndices) > 0 {
			m.nextSearchMatch()
		}
		return nil

	case keys.KeyNCap:
		// Previous search match
		if len(m.matchingIndices) > 0 {
			m.previousSearchMatch()
		}
		return nil

	case keys.KeyCtrlL, keys.KeyCtrlX:
		// Clear search
		m.clearSearch()
		return nil
	}
	return nil
}

// handleSelectionKeys handles worktree selection keys (space, a, A)
func (m *WorktreeModel) handleSelectionKeys(keyString string) tea.Cmd {
	switch keyString {
	case keys.KeySpace:
		// Toggle current worktree
		if m.selectedIndex < len(m.filteredWorktrees) {
			worktree := m.filteredWorktrees[m.selectedIndex]
			m.toggleWorktree(worktree)
		}
		return nil

	case keys.KeyA:
		// Smart toggle: if all visible worktrees are selected, deselect all; otherwise select all visible
		m.smartToggleAll()
		return nil

	case "A":
		// Shift+A: Always deselect all worktrees
		m.deselectAll()
		return nil
	}
	return nil
}

// handleScroll processes scroll messages
func (m *WorktreeModel) handleScroll(msg WorktreeModalScrollMsg) tea.Cmd {
	if msg.Direction > 0 {
		m.navigateDown()
	} else {
		m.navigateUp()
	}
	return nil
}

// Navigation methods - pure state updates, no rendering
func (m *WorktreeModel) navigateDown() {
	if len(m.filteredWorktrees) == 0 {
		return
	}
	m.selectedIndex++
	if m.selectedIndex >= len(m.filteredWorktrees) {
		m.selectedIndex = len(m.filteredWorktrees) - 1
	}
}

func (m *WorktreeModel) navigateUp() {
	m.selectedIndex--
	if m.selectedIndex < 0 {
		m.selectedIndex = 0
	}
}

// Advanced navigation methods
func (m *WorktreeModel) navigateFastDown() {
	if len(m.filteredWorktrees) == 0 {
		return
	}
	m.selectedIndex += fastScrollDistance
	if m.selectedIndex >= len(m.filteredWorktrees) {
		m.selectedIndex = len(m.filteredWorktrees) - 1
	}
}

func (m *WorktreeModel) navigateFastUp() {
	m.selectedIndex -= fastScrollDistance
	if m.selectedIndex < 0 {
		m.selectedIndex = 0
	}
}

func (m *WorktreeModel) navigateHalfPageDown() {
	if len(m.filteredWorktrees) == 0 {
		return
	}
	m.selectedIndex += halfPageSize
	if m.selectedIndex >= len(m.filteredWorktrees) {
		m.selectedIndex = len(m.filteredWorktrees) - 1
	}
}

func (m *WorktreeModel) navigateHalfPageUp() {
	m.selectedIndex -= halfPageSize
	if m.selectedIndex < 0 {
		m.selectedIndex = 0
	}
}

func (m *WorktreeModel) jumpToFirst() {
	m.selectedIndex = 0
}

func (m *WorktreeModel) jumpToLast() {
	if len(m.filteredWorktrees) > 0 {
		m.selectedIndex = len(m.filteredWorktrees) - 1
	}
}

// Worktree manipulation methods
func (m *WorktreeModel) toggleWorktree(worktree string) {
	if m.selectedWorktrees[worktree] {
		delete(m.selectedWorktrees, worktree)
	} else {
		m.selectedWorktrees[worktree] = true
	}
}

func (m *WorktreeModel) selectAllVisible() {
	for _, worktree := range m.filteredWorktrees {
		m.selectedWorktrees[worktree] = true
	}
}

func (m *WorktreeModel) deselectAll() {
	// Only deselect the currently filtered/visible worktrees
	for _, worktree := range m.filteredWorktrees {
		delete(m.selectedWorktrees, worktree)
	}
}

func (m *WorktreeModel) areAllVisibleSelected() bool {
	if len(m.filteredWorktrees) == 0 {
		return false
	}
	for _, worktree := range m.filteredWorktrees {
		if !m.selectedWorktrees[worktree] {
			return false
		}
	}
	return true
}

func (m *WorktreeModel) smartToggleAll() {
	if m.areAllVisibleSelected() {
		// All visible worktrees are selected, so deselect all
		m.deselectAll()
	} else {
		// Not all visible worktrees are selected, so select all visible
		m.selectAllVisible()
	}
}

func (m *WorktreeModel) copySelectedWorktrees() map[string]bool {
	result := make(map[string]bool)
	maps.Copy(result, m.selectedWorktrees)
	return result
}

// Search methods
func (m *WorktreeModel) updateLiveSearch() {
	// Update filtered worktrees based on current input
	m.searchQuery = m.searchInput
	m.updateFilteredWorktrees()
}

func (m *WorktreeModel) clearSearch() {
	m.searchInput = ""
	m.searchQuery = ""
	m.searchMode = false
	m.updateFilteredWorktrees()
}

func (m *WorktreeModel) updateFilteredWorktrees() {
	if m.searchQuery == "" {
		// No search - show all worktrees
		m.filteredWorktrees = make([]string, len(m.allWorktrees))
		copy(m.filteredWorktrees, m.allWorktrees)
	} else {
		// Filter worktrees based on search query
		m.filteredWorktrees = []string{}
		query := strings.ToLower(m.searchQuery)
		for _, worktree := range m.allWorktrees {
			if strings.Contains(strings.ToLower(worktree), query) {
				m.filteredWorktrees = append(m.filteredWorktrees, worktree)
			}
		}
	}

	// Sort filtered worktrees for consistent display
	sort.Strings(m.filteredWorktrees)

	// Update matching indices for n/N navigation
	m.updateMatchingIndices()

	// Ensure selected index is valid
	if m.selectedIndex >= len(m.filteredWorktrees) {
		m.selectedIndex = 0
		if len(m.filteredWorktrees) == 0 {
			m.selectedIndex = -1
		}
	}
}

func (m *WorktreeModel) updateMatchingIndices() {
	m.matchingIndices = []int{}
	if m.searchQuery == "" {
		return
	}

	for i := range m.filteredWorktrees {
		m.matchingIndices = append(m.matchingIndices, i)
	}
	m.currentMatchIndex = 0
}

func (m *WorktreeModel) nextSearchMatch() {
	if len(m.matchingIndices) == 0 {
		return
	}
	m.currentMatchIndex++
	if m.currentMatchIndex >= len(m.matchingIndices) {
		m.currentMatchIndex = 0
	}
	m.selectedIndex = m.matchingIndices[m.currentMatchIndex]
}

func (m *WorktreeModel) previousSearchMatch() {
	if len(m.matchingIndices) == 0 {
		return
	}
	m.currentMatchIndex--
	if m.currentMatchIndex < 0 {
		m.currentMatchIndex = len(m.matchingIndices) - 1
	}
	m.selectedIndex = m.matchingIndices[m.currentMatchIndex]
}

// UI methods - parent-child architecture compliant
func (m *WorktreeModel) updateDimensions(width, height int) {
	m.SetDimensions(width, height)
}

func (m *WorktreeModel) updateViewport() {
	// Calculate modal dimensions
	modalWidth := min(m.GetWidth()-4, 80)   // Maximum 80 chars wide, with margins
	modalHeight := min(m.GetHeight()-4, 40) // Maximum 40 lines high, with margins

	// Calculate viewport dimensions using dimension calculator
	// Always reserve scrollbar space to prevent content overflow when scrollbar appears
	// Modal has Padding(1, 2) = vertical 1, horizontal 2 -> total horizontal padding = 4
	calc := layout.NewCalculator(modalWidth, modalHeight, layout.ModalComponent).
		WithScrollbar().      // Reserve space for scrollbar (4 chars)
		WithPadding(2).       // Horizontal padding (left + right)
		WithReservedLines(12) // Title (3) + search (2) + help (3) + spacing (4)

	dims := calc.Calculate()

	// Apply minimum size constraints - use Content width (accounts for scrollbar)
	viewportHeight := max(5, dims.ViewportHeight)
	viewportWidth := max(30, dims.Content)

	m.viewport.Width = viewportWidth
	m.viewport.Height = viewportHeight
}

// renderModal renders the complete worktree modal
func (m *WorktreeModel) renderModal() string {
	// Create the content
	content := m.renderContent()

	// Calculate modal dimensions
	modalWidth := min(m.GetWidth()-4, 80)   // Maximum 80 chars wide, with margins
	modalHeight := min(m.GetHeight()-4, 40) // Maximum 40 lines high, with margins

	// Create the modal with border
	modal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("51")). // Bright cyan like other modals
		Width(modalWidth).
		Height(modalHeight).
		Padding(1, 2).
		Align(lipgloss.Left, lipgloss.Top). // Top align for list content
		Render(content)

	return modal
}

// renderContent renders the modal content
func (m *WorktreeModel) renderContent() string {
	var content strings.Builder

	// Title with better spacing
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("51")).
		Align(lipgloss.Center).
		MarginBottom(1)
	title := titleStyle.Render("Select Worktrees")
	content.WriteString(title)
	content.WriteString("\n")

	// Search section
	content.WriteString(m.renderSearchSection())
	content.WriteString("\n\n")

	// Worktree list
	content.WriteString(m.renderWorktreeList())

	// Instructions (with extra spacing for better visual separation)
	content.WriteString("\n\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Align(lipgloss.Center)
	if m.searchMode {
		instructions := helpStyle.Render("Type to search • Enter to confirm • Esc to cancel")
		content.WriteString(instructions)
	} else {
		// Multi-line help for better readability
		line1 := helpStyle.Render("j/k: navigate • J/K: fast scroll • gg/G: first/last • ctrl+u/d: half-page")
		line2 := helpStyle.Render("Space: toggle • a: smart select • A: deselect visible • /: search • Enter: apply • Esc: cancel")
		content.WriteString(line1 + "\n" + line2)
	}

	return content.String()
}

// renderSearchSection renders the search input and status
func (m *WorktreeModel) renderSearchSection() string {
	var content strings.Builder

	// Search input
	searchStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	switch {
	case m.searchMode:
		// Active search input with cursor
		inputStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("15")).
			Padding(0, 1).
			Width(30)

		searchText := m.searchInput + "▊" // Add cursor
		searchField := inputStyle.Render(searchText)
		content.WriteString(searchStyle.Render("Search: ") + searchField)
	case m.searchQuery != "":
		// Show committed search query
		content.WriteString(searchStyle.Render("Search: \"" + m.searchQuery + "\""))
	default:
		// Show search prompt
		promptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		content.WriteString(promptStyle.Render("Press / to search"))
	}

	// Search status
	if m.searchQuery != "" {
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		matches := len(m.filteredWorktrees)
		total := len(m.allWorktrees)
		status := statusStyle.Render(" (" + strconv.Itoa(matches) + "/" + strconv.Itoa(total) + " worktrees)")
		content.WriteString(status)
	}

	return content.String()
}

// renderWorktreeList renders the list of worktrees with selection status using viewport
// This is a pure render function that always rebuilds from current model state
func (m *WorktreeModel) renderWorktreeList() string {
	// Validate viewport is properly initialized
	if m.viewport.Width <= 0 || m.viewport.Height <= 0 {
		return "Initializing..."
	}

	// Always rebuild content from current state
	m.buildViewportContent()

	// Always position viewport for current selectedIndex
	m.updateViewportPosition()

	// Get the viewport's rendered view
	viewportContent := m.viewport.View()

	// Add scrollbar if content is scrollable
	totalLines := m.viewport.TotalLineCount()
	viewportHeight := m.viewport.Height
	if totalLines > viewportHeight {
		// Generate scrollbar
		scrollbar := view.RenderScrollBarExact(m.viewport.YOffset, totalLines, viewportHeight)

		// Compose content with scrollbar
		// Calculate width with scrollbar offset
		// The +2 accounts for visual alignment with scrollbar positioning
		// which includes 1 char gap + scrollbar width in the composition
		viewportContent = sharedviewport.ComposeWithScrollbar(viewportContent, scrollbar, m.viewport.Width+2, 0)
	}

	return viewportContent
}

// renderWorktreeOption renders a single worktree option with checkbox
func (m *WorktreeModel) renderWorktreeOption(index int, worktree string) string {
	isSelected := index == m.selectedIndex
	isChecked := m.selectedWorktrees[worktree]

	// Checkbox with improved visibility
	var checkbox string
	if isChecked {
		// Green filled square for selected worktrees
		checkboxStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46")) // Bright green
		checkbox = checkboxStyle.Render("■")
	} else {
		// Empty square for unselected worktrees
		checkboxStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244")) // Light gray
		checkbox = checkboxStyle.Render("□")
	}

	// Worktrees render muted with the @ sigil matching the task row chip
	worktreeText := lipgloss.NewStyle().Foreground(lipgloss.Color("215")).Render("@" + worktree) // Orange

	// Build the core line content first with individual element styling
	// Format: "checkbox @worktree-name"
	line := checkbox + " " + worktreeText

	// Apply selection styling and indicators
	if isSelected {
		// Add selection indicators and apply bold styling to entire line
		line = "► " + line + " ◄"
		headerColor := "62" // Bright purple/blue for headers
		styledLine := lipgloss.NewStyle().Foreground(lipgloss.Color(headerColor)).Bold(true).Render(line)
		return styledLine
	}
	// Add spacing to align with selected items
	line = "  " + line
	return line
}

// buildViewportContent builds and sets the viewport content from current model state
// This is a helper method called only from renderWorktreeList() during View rendering
// Selection indicators (► ◄) are baked into strings, requiring rebuild on every render
func (m *WorktreeModel) buildViewportContent() {
	if len(m.filteredWorktrees) == 0 {
		m.viewport.SetContent("No worktrees found")
		return
	}

	listItems := make([]string, 0, len(m.filteredWorktrees)) // Preallocate for all worktrees
	for i, worktree := range m.filteredWorktrees {
		option := m.renderWorktreeOption(i, worktree)
		listItems = append(listItems, option)
	}

	// Join all worktrees with newlines using lipgloss.JoinVertical
	content := lipgloss.JoinVertical(lipgloss.Left, listItems...)
	m.viewport.SetContent(content)
}

// updateViewportPosition positions the viewport to center the selected item
// This is a helper method called only from renderWorktreeList() during View rendering
func (m *WorktreeModel) updateViewportPosition() {
	if len(m.filteredWorktrees) == 0 || m.selectedIndex < 0 {
		return
	}

	if m.selectedIndex >= len(m.filteredWorktrees) {
		return
	}

	totalLines := len(m.filteredWorktrees)
	viewportHeight := m.viewport.Height

	// Center the selected item in viewport for consistent behavior
	targetLine := m.selectedIndex - viewportHeight/2
	if targetLine < 0 {
		targetLine = 0 // Can't scroll above top
	}
	if targetLine+viewportHeight > totalLines {
		targetLine = totalLines - viewportHeight // Can't scroll below bottom
		if targetLine < 0 {
			targetLine = 0
		}
	}
	m.viewport.SetYOffset(targetLine)
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
