package header

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yousfisaad/lazyarchon/internal/shared/styling"
	"github.com/yousfisaad/lazyarchon/internal/ui/components/base"
	"github.com/yousfisaad/lazyarchon/internal/ui/context"
)

const ComponentID = "header_component"

// HeaderModel represents the header component
// Architecture: Follows four-tier state pattern (Display Parameters eliminated)
// - Source data: Read from ProgramContext (projects, feature filters)
// - UI Presentation State: Read from UIState (view mode, search state)
// - Owned state: None (component is stateless)
// - Transient feedback: None
//
// Components compute display data on-demand by calling ProgramContext/UIState methods.
type HeaderModel struct {
	base.BaseComponent

	// ===================================================================
	// OWNED STATE - Component manages these directly
	// ===================================================================
	// None - this component is now completely stateless
	//
	// All display data is computed on-demand in View() by calling context methods:
	// - displayProjectName → ctx.ProgramContext.GetCurrentProjectName()
	// - displayFeatureSummary → ctx.ProgramContext.GetFeatureFilterSummary()
	// - displayTaskCount → len(ctx.GetSortedTasks())
}

// ctx returns the program context for easy access to global state
func (m *HeaderModel) ctx() *context.ProgramContext {
	return m.GetContext().ProgramContext
}

// NewModel creates a new header component with pure message-based communication
func NewModel(context *base.ComponentContext) *HeaderModel {
	baseComponent := base.NewBaseComponent(ComponentID, base.HeaderComponent, context)

	return &HeaderModel{
		BaseComponent: baseComponent,
	}
}

// Init initializes the header component
func (m *HeaderModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the header component
func (m *HeaderModel) Update(msg tea.Msg) tea.Cmd {
	// This component is now completely stateless - all display data computed on-demand in View()
	// No owned state to update, no messages to handle
	//
	// NOTE: All display parameter messages removed - compute on-demand from context:
	// - ProjectDisplayMsg → call ctx.ProgramContext.GetCurrentProjectName()
	// - FeatureDisplayMsg → call ctx.ProgramContext.GetFeatureFilterSummary()
	// - SelectionPositionMsg → call len(ctx.GetSortedTasks())
	//
	// NOTE: Modal state tracking removed - header is covered by modal overlay anyway
	// - ModalStateMsg handler removed (dead code)
	// - featureModalActive field removed (had no effect)

	return nil
}

func (m *HeaderModel) View() string {
	if m.GetContext().UIState.IsProjectView() {
		return styling.HeaderStyle.Render("LazyArchon - Select Project")
	}

	parts := m.collectHeaderParts()
	content := m.joinHeaderParts(parts)
	taskCount := len(m.GetContext().GetSortedTasks())
	headerText := fmt.Sprintf("LazyArchon - %s (%d)", content, taskCount)

	return styling.HeaderStyle.Render(headerText)
}

// collectHeaderParts gathers all visible header parts in display order
func (m *HeaderModel) collectHeaderParts() []string {
	parts := []string{}

	// Project name (always shown)
	parts = append(parts, m.ctx().GetCurrentProjectName())

	// Active search query (if searching)
	if searchQuery := m.getSearchIndicator(); searchQuery != "" {
		parts = append(parts, searchQuery)
	}

	// Feature filters (if non-default)
	if featureFilter := m.getFeatureFilterDisplay(); featureFilter != "" {
		parts = append(parts, featureFilter)
	}

	return parts
}

// getSearchIndicator returns the search indicator if search is active
func (m *HeaderModel) getSearchIndicator() string {
	uiState := m.GetContext().UIState
	if uiState.SearchActive && uiState.SearchQuery != "" {
		return fmt.Sprintf("🔍 \"%s\"", uiState.SearchQuery)
	}
	return ""
}

// getFeatureFilterDisplay returns feature filter text if filters are active
func (m *HeaderModel) getFeatureFilterDisplay() string {
	summary := m.ctx().GetFeatureFilterSummary()
	if summary == "All features" || summary == "No features" {
		return ""
	}
	return summary
}

// joinHeaderParts joins header parts with bullet separators
func (m *HeaderModel) joinHeaderParts(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}

	// Build interleaved parts with bullet separators
	joined := []string{}
	for i, part := range parts {
		joined = append(joined, part)
		if i < len(parts)-1 {
			joined = append(joined, " • ")
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, joined...)
}
