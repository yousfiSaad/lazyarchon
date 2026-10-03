package help

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/layout"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/styling"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/utils/keys"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/utils/view"
	sharedviewport "github.com/yousfisaad/lazyarchon/v2/internal/shared/viewport"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/base"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/messages"
)

const ComponentID = "help_modal"

// HelpModel represents the help modal component
// Architecture: Follows four-tier state pattern
// - No source data caching (reads from ProgramContext via GetContext() as needed)
// - No display parameters (self-contained modal)
// - Owned state only (viewport, contentWidth)
// - No transient feedback (modal lifecycle managed by MainModel)
// - Modal lifecycle managed by BaseModal (active/visible state)
type HelpModel struct {
	base.BaseModal

	// ===================================================================
	// OWNED STATE - Component manages these directly
	// ===================================================================
	viewport     viewport.Model // Viewport for scrolling help content
	contentWidth int            // Calculated content width for rendering
}

// NewModel creates a new help modal component
func NewModel(context *base.ComponentContext) *HelpModel {
	baseModal := base.NewBaseModal(ComponentID, base.HelpModalComponent, context)

	// Calculate initial dimensions using dimension calculator
	defaultWidth := 70
	defaultHeight := 25
	// Always reserve scrollbar space to prevent content overflow when scrollbar appears
	calc := layout.NewCalculator(defaultWidth, defaultHeight, layout.ModalComponent).
		WithScrollbar(). // Reserve space for scrollbar (4 chars)
		WithPadding(1)   // Modal padding

	dims := calc.Calculate()

	m := &HelpModel{ //nolint:varnamelen // m is idiomatic for model in Bubble Tea
		BaseModal:    baseModal,
		contentWidth: dims.Content, // Use Content width (accounts for scrollbar)
	}
	// Set dimensions using base component
	m.SetDimensions(defaultWidth, defaultHeight)

	// Initialize viewport with calculated dimensions - use Content width (accounts for scrollbar)
	m.viewport = viewport.New(dims.Content, dims.ViewportHeight)

	return m
}

// Init implements the Component interface
func (m *HelpModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the component state
func (m *HelpModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case ShowHelpModalMsg:
		m.SetActive(true)
		m.SetFocus(true)
		m.updateContent()
		m.viewport.GotoTop()
		return m.BroadcastMessage(messages.ModalStateMsg{
			Type:   string(base.ModalTypeHelp),
			Active: true,
		})

	case HideHelpModalMsg:
		m.SetActive(false)
		m.SetFocus(false)
		return m.BroadcastMessage(messages.ModalStateMsg{
			Type:   string(base.ModalTypeHelp),
			Active: false,
		})

	case tea.WindowSizeMsg:
		m.updateDimensions(msg.Width, msg.Height)
		if m.IsActive() {
			m.updateContent()
		}
		return nil

	case tea.KeyMsg:
		if !m.IsActive() || !m.IsFocused() {
			return nil
		}
		return m.handleKeyPress(msg)

	case HelpModalScrollMsg:
		if !m.IsActive() {
			return nil
		}
		return m.handleScroll(msg)

	default:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return cmd
	}
}

// View implements the Component interface
func (m *HelpModel) View() string {
	if !m.IsActive() {
		return ""
	}

	// Get viewport content
	viewportContent := m.viewport.View()

	// Add scrollbar if content is scrollable
	totalLines := m.viewport.TotalLineCount()
	viewportHeight := m.viewport.Height
	if totalLines > viewportHeight {
		// Generate scrollbar
		scrollbar := view.RenderScrollBarExact(m.viewport.YOffset, totalLines, viewportHeight)

		// Compose content with scrollbar
		// Modal width includes border (2) and padding (2), so content area is width - 4
		// The viewport content fits in this area, we need to account for modal structure
		contentWidth := m.GetWidth() - 4 // Border (2) + Padding (2)
		viewportContent = sharedviewport.ComposeWithScrollbar(viewportContent, scrollbar, contentWidth+2, 0)
	}

	// Use stored modal dimensions (calculated in updateDimensions)
	// Create help modal with border
	// Parent handles positioning - modal just returns its content
	helpModal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("51")). // Bright cyan like active panels
		Width(m.GetWidth()).
		Height(m.GetHeight()).
		Padding(1).
		Render(viewportContent)

	return helpModal
}

// CanFocus returns true as the help modal can receive focus
func (m *HelpModel) CanFocus() bool {
	return true
}

// updateDimensions updates the component's dimensions
func (m *HelpModel) updateDimensions(width, height int) {
	// Calculate final modal dimensions (with margins and limits)
	modalWidth := min(width-4, 70)   // Maximum 70 chars wide, with margins
	modalHeight := min(height-4, 25) // Maximum 25 lines high, with margins
	m.SetDimensions(modalWidth, modalHeight)

	// Calculate viewport dimensions using dimension calculator
	// Always reserve scrollbar space to prevent content overflow when scrollbar appears
	calc := layout.NewCalculator(modalWidth, modalHeight, layout.ModalComponent).
		WithScrollbar(). // Reserve space for scrollbar (4 chars)
		WithPadding(1)   // Modal padding

	dims := calc.Calculate()

	// Update stored dimensions - use Content width (accounts for scrollbar)
	m.contentWidth = dims.Content

	// Update viewport size - use Content width (accounts for scrollbar)
	m.viewport.Width = dims.Content
	m.viewport.Height = dims.ViewportHeight
}

// handleKeyPress handles key presses for the help modal
func (m *HelpModel) handleKeyPress(key tea.KeyMsg) tea.Cmd {
	keyString := key.String()

	switch keyString {
	case keys.KeyQuestion, keys.KeyEscape, keys.KeyQ:
		return m.BroadcastMessage(HideHelpModalMsg{})

	case keys.KeyJ, keys.KeyArrowDown:
		m.viewport.ScrollDown(1)
		return nil

	case keys.KeyK, keys.KeyArrowUp:
		m.viewport.ScrollUp(1)
		return nil

	case keys.KeyJCap:
		m.viewport.ScrollDown(4)
		return nil

	case keys.KeyKCap:
		m.viewport.ScrollUp(4)
		return nil

	case keys.KeyCtrlU, keys.KeyPgUp:
		m.viewport.HalfPageUp()
		return nil

	case keys.KeyCtrlD, keys.KeyPgDn:
		m.viewport.HalfPageDown()
		return nil

	case keys.KeyGG, keys.KeyHome:
		m.viewport.GotoTop()
		return nil

	case keys.KeyGCap, keys.KeyEnd:
		m.viewport.GotoBottom()
		return nil

	case keys.KeyCtrlC:
		return tea.Quit

	default:
		return nil
	}
}

// handleScroll handles programmatic scrolling
func (m *HelpModel) handleScroll(msg HelpModalScrollMsg) tea.Cmd {
	switch msg.Direction {
	case ScrollUp:
		m.viewport.ScrollUp(msg.Amount)
	case ScrollDown:
		m.viewport.ScrollDown(msg.Amount)
	case ScrollToTop:
		m.viewport.GotoTop()
	case ScrollToBottom:
		m.viewport.GotoBottom()
	case ScrollHalfUp:
		m.viewport.HalfPageUp()
	case ScrollHalfDown:
		m.viewport.HalfPageDown()
	}
	return nil
}

// updateContent updates the help modal content using dynamic key registry
func (m *HelpModel) updateContent() {
	// Create style context for help content styling
	styleContext := m.GetContext().StyleContextProvider.CreateStyleContext(false)
	factory := styleContext.Factory()

	// Get key registry (using defaults for now - custom keybindings will be passed from config in future)
	registry := keys.NewKeyRegistry(nil)

	// Get organized help sections from registry
	sections := registry.GetHelpSections()

	// Estimate capacity: title + sections (each with header + bindings + spacing) + status symbols + footer
	estimatedLines := 10 + len(sections)*15
	help := make([]string, 0, estimatedLines) // Preallocate for help lines

	// Title
	help = append(help, factory.Header().Render("LazyArchon Help"))
	help = append(help, "")

	for _, section := range sections {
		// Section header
		help = append(help, factory.Header().Render(section.Title+":"))

		// Section bindings
		for _, binding := range section.Bindings {
			// Format: "  key          description"
			if binding.Key != "" {
				line := "  " + binding.Key
				// Pad key to consistent width for alignment
				for len(line) < 17 {
					line += " "
				}
				line += binding.Description
				help = append(help, line)
			} else {
				// For visual indicators and status symbols (no key)
				help = append(help, "  "+binding.Description)
			}
		}
		help = append(help, "")
	}

	// Add task status symbols with actual styling symbols
	help = append(help, factory.Header().Render("Task Status Symbols:"))
	help = append(help, "  "+styling.StatusSymbolTodo+"  Todo       Not started")
	help = append(help, "  "+styling.StatusSymbolDoing+"  Doing      In progress")
	help = append(help, "  "+styling.StatusSymbolReview+"  Review     Under review")
	help = append(help, "  "+styling.StatusSymbolDone+"  Done       Completed")
	help = append(help, "")

	// Footer
	help = append(help, factory.Italic(styling.CurrentTheme.MutedColor).Render("Press ? or ESC to close this help"))

	// Set the content in the viewport using lipgloss.JoinVertical
	content := lipgloss.JoinVertical(lipgloss.Left, help...)
	m.viewport.SetContent(content)
}
