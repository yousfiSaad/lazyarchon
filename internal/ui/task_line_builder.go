package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yousfisaad/lazyarchon/internal/archon"
	"github.com/yousfisaad/lazyarchon/internal/ui/styling"
)

// LineComponent represents a styled component of a task line
type LineComponent struct {
	content   string         // Plain text content
	style     lipgloss.Style // Styling to apply
	priority  int            // Priority for truncation (higher = more important)
	isFixed   bool           // Cannot be truncated (indicators, status)
	minWidth  int            // Minimum width when truncated
}

// TaskLineBuilder builds task display lines with intelligent space management
type TaskLineBuilder struct {
	availableWidth int
	components     []LineComponent
	styleContext   *styling.StyleContext
	statusColor    string // Store status color for search highlighting
}

// NewTaskLineBuilder creates a new builder for the given available width with styling context
func NewTaskLineBuilder(availableWidth int, styleContext *styling.StyleContext) *TaskLineBuilder {
	return &TaskLineBuilder{
		availableWidth: availableWidth,
		components:     make([]LineComponent, 0),
		styleContext:   styleContext,
	}
}

// AddPriorityIndicator adds the priority indicator if enabled
func (b *TaskLineBuilder) AddPriorityIndicator(task archon.Task) *TaskLineBuilder {
	if !b.styleContext.Config().IsPriorityIndicatorsEnabled() {
		return b
	}

	priority := GetTaskPriority(task.TaskOrder, nil)
	symbol := GetPrioritySymbol(priority)
	priorityLevel := priorityLevelToString(priority)
	style := b.styleContext.Factory().Priority(priorityLevel)

	b.components = append(b.components, LineComponent{
		content:  symbol + " ",
		style:    style,
		priority: 100, // High priority - always show
		isFixed:  true,
		minWidth: 1,
	})

	return b
}

// AddStatusIndicator adds the status symbol
func (b *TaskLineBuilder) AddStatusIndicator(task archon.Task) *TaskLineBuilder {
	symbol := task.GetStatusSymbol()
	style := b.styleContext.Factory().Status(task.Status)

	b.components = append(b.components, LineComponent{
		content:  symbol + " ",
		style:    style,
		priority: 100, // High priority - always show
		isFixed:  true,
		minWidth: 2,
	})

	return b
}

// AddTitle adds the task title with search highlighting support
func (b *TaskLineBuilder) AddTitle(task archon.Task, searchQuery string, searchActive bool) *TaskLineBuilder {
	var content string
	var style lipgloss.Style

	// Store the task status color for search highlighting (if needed)
	b.statusColor = GetThemeStatusColor(task.Status)

	content = task.Title

	if searchActive && searchQuery != "" {
		// For search highlighting, we'll use empty style and apply highlighting in build phase
		style = lipgloss.NewStyle()
	} else {
		// Use factory to get properly styled text with selection support
		style = b.styleContext.Factory().Status(task.Status)
	}

	b.components = append(b.components, LineComponent{
		content:  content,
		style:    style,
		priority: 90, // High priority but can be truncated
		isFixed:  false,
		minWidth: 10, // Minimum readable title width
	})

	return b
}

// AddFeatureTag adds the feature tag if present and space permits
func (b *TaskLineBuilder) AddFeatureTag(task archon.Task) *TaskLineBuilder {
	if task.Feature == nil || *task.Feature == "" {
		return b
	}

	content := fmt.Sprintf(" #%s", *task.Feature)
	var style lipgloss.Style

	if b.styleContext.Config().IsFeatureColorsEnabled() {
		style = b.styleContext.Factory().Feature(*task.Feature)
	} else {
		style = b.styleContext.Factory().Muted()
	}

	b.components = append(b.components, LineComponent{
		content:  content,
		style:    style,
		priority: 50, // Lower priority - can be dropped if needed
		isFixed:  false,
		minWidth: 0, // Can be completely removed
	})

	return b
}

// Build assembles the line with intelligent truncation
func (b *TaskLineBuilder) Build(searchQuery string, searchActive bool) string {
	if len(b.components) == 0 {
		return ""
	}

	// Calculate total width needed without truncation
	totalWidth := 0
	for _, comp := range b.components {
		totalWidth += len(comp.content) // Use plain text length for calculation
	}

	// If everything fits, build the line normally
	if totalWidth <= b.availableWidth {
		return b.buildFullLine(searchQuery, searchActive)
	}

	// Need to truncate - start with minimum widths for fixed components
	usedWidth := 0
	for _, comp := range b.components {
		if comp.isFixed {
			usedWidth += comp.minWidth
		}
	}

	// Calculate available width for flexible components
	remainingWidth := b.availableWidth - usedWidth

	// Allocate width to flexible components based on priority
	flexComponents := make([]int, 0) // indices of flexible components
	for i, comp := range b.components {
		if !comp.isFixed {
			flexComponents = append(flexComponents, i)
		}
	}

	// Sort flexible components by priority (higher first)
	for i := 0; i < len(flexComponents)-1; i++ {
		for j := i + 1; j < len(flexComponents); j++ {
			if b.components[flexComponents[i]].priority < b.components[flexComponents[j]].priority {
				flexComponents[i], flexComponents[j] = flexComponents[j], flexComponents[i]
			}
		}
	}

	// Allocate width to flexible components
	componentWidths := make([]int, len(b.components))
	for i, comp := range b.components {
		if comp.isFixed {
			// Fixed components ALWAYS get at least their minWidth, never 0
			componentWidths[i] = max(comp.minWidth, len(comp.content))
		}
	}

	// Distribute remaining width to flexible components in priority order
	for _, idx := range flexComponents {
		comp := b.components[idx]
		contentLen := len(comp.content)

		if remainingWidth >= contentLen {
			// Full content fits
			componentWidths[idx] = contentLen
			remainingWidth -= contentLen
		} else if remainingWidth >= comp.minWidth {
			// Partial content fits
			componentWidths[idx] = remainingWidth
			remainingWidth = 0
		} else {
			// No space for this component
			componentWidths[idx] = 0
		}
	}

	return b.buildTruncatedLine(componentWidths, searchQuery, searchActive)
}

// buildFullLine builds the line when all components fit
func (b *TaskLineBuilder) buildFullLine(searchQuery string, searchActive bool) string {
	var parts []string

	for i, comp := range b.components {
		var styledContent string

		// Special handling for title with search highlighting
		if i == b.getTitleIndex() && searchActive && searchQuery != "" {
			styledContent = b.styleContext.Factory().ApplySearchHighlighting(comp.content, b.statusColor)
		} else {
			styledContent = comp.style.Render(comp.content)
		}

		parts = append(parts, styledContent)
	}

	return strings.Join(parts, "")
}

// buildTruncatedLine builds the line with truncated components
func (b *TaskLineBuilder) buildTruncatedLine(widths []int, searchQuery string, searchActive bool) string {
	var parts []string

	for i, comp := range b.components {
		allocatedWidth := widths[i]

		if allocatedWidth == 0 {
			// Safety check: Never skip fixed components (priority/status indicators)
			if comp.isFixed {
				allocatedWidth = comp.minWidth // Force minimum width for fixed components
			} else {
				continue // Skip flexible components with 0 width
			}
		}

		content := comp.content
		if len(content) > allocatedWidth {
			if allocatedWidth <= 1 {
				content = "…"
			} else {
				content = content[:allocatedWidth-1] + "…"
			}
		}

		var styledContent string

		// Special handling for title with search highlighting
		if i == b.getTitleIndex() && searchActive && searchQuery != "" {
			styledContent = b.styleContext.Factory().ApplySearchHighlighting(content, b.statusColor)
		} else {
			styledContent = comp.style.Render(content)
		}

		parts = append(parts, styledContent)
	}

	return strings.Join(parts, "")
}

// getTitleIndex finds the index of the title component
func (b *TaskLineBuilder) getTitleIndex() int {
	for i, comp := range b.components {
		if !comp.isFixed && comp.priority == 90 { // Title has priority 90
			return i
		}
	}
	return -1 // Title not found
}

// Helper method to get actual status color from task
func (b *TaskLineBuilder) getStatusColor(task archon.Task) string {
	return GetThemeStatusColor(task.Status)
}

// priorityLevelToString converts PriorityLevel enum to string
func priorityLevelToString(priority PriorityLevel) string {
	switch priority {
	case PriorityLow:
		return "low"
	case PriorityMedium:
		return "medium"
	case PriorityHigh:
		return "high"
	default:
		return "low"
	}
}