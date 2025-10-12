package statusbar

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yousfisaad/lazyarchon/internal/archon"
	"github.com/yousfisaad/lazyarchon/internal/shared/styling"
	"github.com/yousfisaad/lazyarchon/internal/ui/components/base"
	"github.com/yousfisaad/lazyarchon/internal/ui/context"
	"github.com/yousfisaad/lazyarchon/internal/ui/messages"
)

const ComponentID = "statusbar_component"

// tickMsg is sent periodically to animate the loading spinner
type tickMsg time.Time

// StatusType represents the different states of the status bar
type StatusType int

const (
	StatusReady   StatusType = iota // Normal operation state
	StatusLoading                   // Loading/processing state
	StatusError                     // Error state
	StatusInfo                      // Informational/temporary message state
)

// String returns the string representation of StatusType
func (s StatusType) String() string {
	switch s {
	case StatusReady:
		return "ready"
	case StatusLoading:
		return "loading"
	case StatusError:
		return "error"
	case StatusInfo:
		return "info"
	default:
		return "ready" // Safe fallback
	}
}

// StatusBarModel represents the status bar component
// Architecture: Follows four-tier state pattern
// - Tier 1: Source data (Loading, Error, Connected) - read from ProgramContext via ctx()
// - Tier 2: UI Presentation State (ProjectView, SearchMode) - read from UIState
// - Tier 3: Owned state (spinner animation) - managed locally
// - Tier 4: Transient feedback (status messages) - temporary user feedback
//
// Computed data (task counts, sort mode, etc.) is accessed via context method calls
// instead of caching from messages. This eliminates the Display Parameters tier.
type StatusBarModel struct {
	base.BaseComponent

	// ===================================================================
	// OWNED STATE - Component manages these directly
	// ===================================================================
	spinnerIndex    int
	spinnerFrames   []string
	lastSpinnerTime time.Time

	// ===================================================================
	// TRANSIENT FEEDBACK - Temporary messages (not in ProgramContext)
	// ===================================================================
	statusMessage     string
	statusMessageTime time.Time
}

// NewModel creates a new status bar component
func NewModel(context *base.ComponentContext) *StatusBarModel {
	baseComponent := base.NewBaseComponent(ComponentID, base.StatusBarComponent, context)

	model := &StatusBarModel{
		BaseComponent:   baseComponent,
		spinnerFrames:   []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		spinnerIndex:    0,
		lastSpinnerTime: time.Now(),
	}
	// Set default dimensions - will be overridden by WindowSizeMsg
	model.SetDimensions(80, 1)
	return model
}

// ctx returns the program context for easy access to global state
func (m *StatusBarModel) ctx() *context.ProgramContext {
	return m.GetContext().ProgramContext
}

// tick sends a tickMsg after a delay for spinner animation
func tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Init initializes the status bar component
func (m *StatusBarModel) Init() tea.Cmd {
	return tick() // Start spinner animation
}

// Update handles messages for the status bar component
func (m *StatusBarModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tickMsg:
		// Advance spinner animation if loading (read directly from context)
		if m.ctx().Loading {
			m.advanceSpinner()
		}
		// Continue ticking (recursive pattern)
		return tick()

	case tea.WindowSizeMsg:
		m.HandleWindowResize(msg)

	// Transient feedback (not in ProgramContext)
	case messages.StatusFeedbackMsg:
		m.statusMessage = msg.Message
		m.statusMessageTime = time.Now()
	}

	return nil
}

// View renders the status bar component
func (m *StatusBarModel) View() string {
	// All states handled by buildSpecialStateStatus in 3-tier priority
	// Tier 1: Blocking (help, loading, error)
	// Tier 2: Transient feedback (temporary messages)
	// Tier 3: Mode/Context (project mode, task mode - always has fallback)
	statusText, statusType := m.buildSpecialStateStatus()
	return m.renderWithStatus(statusText, statusType)
}

// buildSpecialStateStatus handles status in 3-tier priority order
// Tier 1: Blocking states (highest) - block all interaction
// Tier 2: Transient feedback (medium) - brief user feedback, shows even in project mode
// Tier 3: Mode/Context (lowest) - fallback context status
func (m *StatusBarModel) buildSpecialStateStatus() (string, StatusType) {
	// Tier 1: Blocking states (highest priority - block all interaction)
	if blockingStatus, statusType := m.buildBlockingStatus(); blockingStatus != "" {
		return blockingStatus, statusType
	}

	// Tier 2: Transient feedback (medium priority - brief user feedback)
	if feedbackStatus := m.buildTransientFeedbackStatus(); feedbackStatus != "" {
		return feedbackStatus, StatusInfo
	}

	// Tier 3: Mode/Context (lowest priority - fallback context)
	return m.buildModeContextStatus()
}

// buildBlockingStatus handles Tier 1: Blocking states that take over UI
func (m *StatusBarModel) buildBlockingStatus() (string, StatusType) {
	ctx := m.ctx()

	// Help modal (fully blocks interaction)
	// TODO: Need to track active modal - consider adding to ProgramContext or passing via MainModel
	// For now, this will need to be handled differently

	// Loading state (blocks actions)
	if ctx.Loading {
		return m.buildLoadingStatus(), StatusLoading
	}

	// Error state (blocks actions)
	if ctx.Error != "" {
		return m.buildErrorStatus(), StatusError
	}

	return "", StatusReady
}

// buildTransientFeedbackStatus handles Tier 2: Transient user feedback
// These messages show even in project mode, then auto-expire after 3 seconds
func (m *StatusBarModel) buildTransientFeedbackStatus() string {
	if m.hasTemporaryMessage() {
		return m.buildTemporaryMessageStatus()
	}
	return ""
}

// buildModeContextStatus handles Tier 3: Mode/Context fallback status
func (m *StatusBarModel) buildModeContextStatus() (string, StatusType) {
	// Project mode context (read from UIState)
	if m.GetContext().UIState.IsProjectView() {
		return m.buildProjectModeStatus(), StatusReady
	}

	// Task mode context - use existing buildContextAwareStatus
	return m.buildContextAwareStatus(), StatusReady
}

// buildProjectModeStatus creates status text for project selection mode
func (m *StatusBarModel) buildProjectModeStatus() string {
	projectCount := len(m.ctx().Projects)
	if projectCount > 0 {
		return fmt.Sprintf("[Project] %d projects available | l: select | h: back | q: quit", projectCount)
	}
	return "Project Selection | ?: help | q: quit"
}

// buildFeatureModeStatus creates status text for feature selection mode
// TODO: Feature count needs to be calculated from tasks or passed via message
func (m *StatusBarModel) buildFeatureModeStatus() string {
	// Calculate feature count from context tasks
	featureCount := calculateFeatureCount(m.ctx().Tasks)
	if featureCount > 0 {
		return fmt.Sprintf("[Features] %d features | j/k/J/K/gg/G: navigate | Space: toggle | a: all | n: none | Enter: apply | q: cancel", featureCount)
	}
	return "Feature Selection | No features available | Enter: apply | q: cancel"
}

// calculateFeatureCount counts unique features from tasks
func calculateFeatureCount(tasks []archon.Task) int {
	featureSet := make(map[string]bool)
	for _, task := range tasks {
		if task.Feature != nil && *task.Feature != "" {
			featureSet[*task.Feature] = true
		}
	}
	return len(featureSet)
}

// buildLoadingStatus creates status text for loading state
func (m *StatusBarModel) buildLoadingStatus() string {
	message := "Loading..."
	if m.ctx().LoadingMessage != "" {
		message = m.ctx().LoadingMessage
	}
	return fmt.Sprintf("[Tasks] %s %s | q: quit", m.getLoadingSpinner(), message)
}

// buildErrorStatus creates status text for error state
func (m *StatusBarModel) buildErrorStatus() string {
	// Format error for display
	errorMsg := m.ctx().Error
	if len(errorMsg) > 80 {
		errorMsg = errorMsg[:77] + "..."
	}
	return fmt.Sprintf("[Tasks] Error: %s | r: retry | q: quit", errorMsg)
}

// hasTemporaryMessage checks if there's an active temporary status message
func (m *StatusBarModel) hasTemporaryMessage() bool {
	if m.statusMessage == "" {
		return false
	}
	return time.Since(m.statusMessageTime) <= 3*time.Second
}

// buildTemporaryMessageStatus creates status text for temporary messages
// Uses context-aware prefix based on current mode
func (m *StatusBarModel) buildTemporaryMessageStatus() string {
	// Use appropriate prefix based on current mode (read from UIState)
	prefix := "[Tasks]"
	if m.GetContext().UIState.IsProjectView() {
		prefix = "[Project]"
	}
	return fmt.Sprintf("%s %s | ?: help | q: quit", prefix, m.statusMessage)
}

// buildSearchInputStatus creates status text for search input mode
func (m *StatusBarModel) buildSearchInputStatus() string {
	cursor := "_" // Simple cursor indicator
	// Read search input from UIState
	searchInput := m.GetContext().UIState.SearchInput
	searchText := fmt.Sprintf("[Search] %s%s", searchInput, cursor)

	// Add match indicator if search has matches
	totalMatches := m.GetContext().UIState.TaskTotalMatches
	if totalMatches > 0 {
		searchText += fmt.Sprintf(" (%d matches)", totalMatches)
	}

	return searchText + " | Enter: apply | Esc: cancel | Ctrl+U: clear"
}

// buildContextAwareStatus creates status text based on the active panel context
func (m *StatusBarModel) buildContextAwareStatus() string {
	// Get active view name from UIState
	activeViewName := m.GetContext().UIState.GetActiveViewName()

	switch activeViewName {
	case "Task List":
		return m.buildTasksContextStatus()
	case "Task Details":
		return m.buildDetailsContextStatus()
	default:
		return fmt.Sprintf("[%s] Ready | ?: help | q: quit", activeViewName)
	}
}

// buildTasksContextStatus creates status text for the tasks panel context
func (m *StatusBarModel) buildTasksContextStatus() string {
	// Call context methods to get computed data instead of reading cached fields
	todo, doing, review, done := m.ctx().GetTaskStatusCounts()
	totalTasks := todo + doing + review + done

	// Connection status indicator (read from context)
	connectionStatus := "●" // Connected
	if !m.ctx().Connected {
		connectionStatus = "○" // Disconnected
	}

	if totalTasks == 0 {
		return fmt.Sprintf("[Tasks] %s No tasks found | r: refresh | q: quit", connectionStatus)
	}

	// Build task information (pass computed values directly)
	sortModeName := m.ctx().GetCurrentSortModeName()
	statusInfo := m.buildTaskStatusInfo(todo, doing, review, done, totalTasks, sortModeName)

	// Build shortcuts
	shortcutText := m.buildTaskShortcuts()

	return fmt.Sprintf("[Tasks] %s %s | %s", connectionStatus, statusInfo, shortcutText)
}

// buildTaskStatusInfo creates the task status information part of the status bar
func (m *StatusBarModel) buildTaskStatusInfo(todo, doing, review, done, totalTasks int, sortMode string) string {
	var statusParts []string
	statusParts = append(statusParts, fmt.Sprintf("%d items", totalTasks))

	// Add status distribution if there are active tasks
	if doing > 0 || review > 0 {
		if doing > 0 {
			statusParts = append(statusParts, fmt.Sprintf("%d doing", doing))
		}
		if review > 0 {
			statusParts = append(statusParts, fmt.Sprintf("%d review", review))
		}
	}

	// Add todo count if significant
	if todo > 0 {
		statusParts = append(statusParts, fmt.Sprintf("%d todo", todo))
	}

	// Add sort mode
	statusParts = append(statusParts, fmt.Sprintf("Sort: %s", sortMode))

	// Add search match information if search is active (call context method)
	// Need to get selectedIndex from UIState to compute current match
	selectedIndex := m.GetContext().UIState.GetSelectedTaskIndex()
	searchActive, searchQuery, _, currentMatch, totalMatches := m.GetContext().UIState.GetTaskSearchState(selectedIndex)
	if searchActive && searchQuery != "" {
		if totalMatches > 0 {
			matchInfo := fmt.Sprintf("Match %d/%d", currentMatch+1, totalMatches)
			statusParts = append(statusParts, matchInfo)
		} else {
			statusParts = append(statusParts, "No matches")
		}
	}

	// Join parts with bullet separator using lipgloss
	if len(statusParts) == 0 {
		return ""
	}
	if len(statusParts) == 1 {
		return statusParts[0]
	}

	// Build parts with separators for lipgloss.JoinHorizontal
	var parts []string
	for i, part := range statusParts {
		parts = append(parts, part)
		if i < len(statusParts)-1 {
			parts = append(parts, " • ")
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, parts...)
}

// buildTaskShortcuts creates the shortcuts part of the tasks status bar
func (m *StatusBarModel) buildTaskShortcuts() string {
	var shortcuts []string
	// Get feature count from context
	featureCount := len(m.ctx().GetUniqueFeatures())
	if featureCount > 0 {
		shortcuts = append(shortcuts, "f: features")
	}
	shortcuts = append(shortcuts, "/: search")
	// Check search state from UIState
	searchActive := m.GetContext().UIState.SearchActive
	if searchActive {
		totalMatches := m.GetContext().UIState.TaskTotalMatches
		if totalMatches > 0 {
			shortcuts = append(shortcuts, "n/N: next/prev match")
		}
		shortcuts = append(shortcuts, "Ctrl+L: clear search")
	}
	shortcuts = append(shortcuts, "?: help")

	// Join shortcuts with pipe separator using lipgloss
	if len(shortcuts) == 0 {
		return ""
	}
	if len(shortcuts) == 1 {
		return shortcuts[0]
	}

	var parts []string
	for i, shortcut := range shortcuts {
		parts = append(parts, shortcut)
		if i < len(shortcuts)-1 {
			parts = append(parts, " | ")
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, parts...)
}

// buildDetailsContextStatus creates status text for the details panel context
func (m *StatusBarModel) buildDetailsContextStatus() string {
	position := m.getCurrentPosition()

	// Connection status indicator (read from context)
	connectionStatus := "●" // Connected
	if !m.ctx().Connected {
		connectionStatus = "○" // Disconnected
	}

	return fmt.Sprintf("[Details] %s %s | ?: help", connectionStatus, position)
}

// renderWithStatus renders the final status bar with styling and truncation
func (m *StatusBarModel) renderWithStatus(statusText string, statusType StatusType) string {
	availableWidth := m.GetWidth() - 2 // Calculate from base component width
	truncatedText := m.truncateStatusText(statusText, availableWidth)
	styleContext := m.createStyleContext(false)
	return styleContext.Factory().StatusBar(statusType.String()).Width(m.GetWidth()).Render(truncatedText)
}

// ===================================================================
// HELPER METHODS
// ===================================================================

// getCurrentPosition returns position info for the current selection
func (m *StatusBarModel) getCurrentPosition() string {
	// Get sorted tasks from context callback (complex filtering logic in MainModel)
	sortedTasks := m.GetContext().GetSortedTasks()
	sortedTaskCount := len(sortedTasks)

	if sortedTaskCount == 0 {
		return "No tasks"
	}

	// Get selected index from UIState
	selectedIndex := m.GetContext().UIState.GetSelectedTaskIndex()
	if selectedIndex >= sortedTaskCount {
		return "No selection"
	}

	return fmt.Sprintf("Task %d of %d", selectedIndex+1, sortedTaskCount)
}

// truncateStatusText intelligently truncates status text to fit the available width
func (m *StatusBarModel) truncateStatusText(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}

	// Try to preserve the core status part by truncating from the shortcuts section
	parts := strings.Split(text, " | ")
	if len(parts) <= 1 {
		// No pipe separator, just truncate with ellipsis
		if maxWidth > 3 {
			return text[:maxWidth-3] + "..."
		}
		return text[:maxWidth]
	}

	// Try removing shortcuts from right to left while preserving status info
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		testResult := result + " | " + parts[i]
		if len(testResult) <= maxWidth {
			result = testResult
		} else {
			// Can't fit more, return what we have
			if len(result) > maxWidth {
				if maxWidth > 3 {
					return result[:maxWidth-3] + "..."
				}
				return result[:maxWidth]
			}
			return result
		}
	}

	return result
}

// createStyleContext creates a StyleContext for UI components with current model state
func (m *StatusBarModel) createStyleContext(isSelected bool) *styling.StyleContext {
	themeAdapter := &styling.ThemeAdapter{
		TodoColor:     styling.CurrentTheme.TodoColor,
		DoingColor:    styling.CurrentTheme.DoingColor,
		ReviewColor:   styling.CurrentTheme.ReviewColor,
		DoneColor:     styling.CurrentTheme.DoneColor,
		HeaderColor:   styling.CurrentTheme.HeaderColor,
		MutedColor:    styling.CurrentTheme.MutedColor,
		AccentColor:   styling.CurrentTheme.AccentColor,
		StatusColor:   styling.CurrentTheme.StatusColor,
		FeatureColors: styling.CurrentTheme.FeatureColors,
		Name:          styling.CurrentTheme.Name,
	}

	// Get search state from UIState
	return styling.NewStyleContext(themeAdapter, m.GetContext().ConfigProvider).
		WithSelection(isSelected).
		WithSearch(m.GetContext().UIState.SearchQuery, m.GetContext().UIState.SearchActive)
}

// ===================================================================
// SPINNER ANIMATION METHODS
// ===================================================================

// getLoadingSpinner returns the current spinner character
func (m *StatusBarModel) getLoadingSpinner() string {
	if len(m.spinnerFrames) == 0 {
		return "•"
	}
	return m.spinnerFrames[m.spinnerIndex]
}

// advanceSpinner advances the spinner animation to the next frame
func (m *StatusBarModel) advanceSpinner() {
	now := time.Now()
	// Only advance if enough time has passed (for smooth animation)
	if now.Sub(m.lastSpinnerTime) >= 100*time.Millisecond {
		m.spinnerIndex = (m.spinnerIndex + 1) % len(m.spinnerFrames)
		m.lastSpinnerTime = now
	}
}
