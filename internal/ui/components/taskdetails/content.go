package taskdetails

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yousfisaad/lazyarchon/v2/internal/archon"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/styling"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/utils/view"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/base"
)

// TaskContentGenerator handles pure content generation for task details
// Separated from UI concerns for clean architecture
type TaskContentGenerator struct {
	task         *archon.Task
	searchQuery  string
	searchActive bool
	contentWidth int

	// Component context for accessing dependencies
	context *base.ComponentContext
}

// NewTaskContentGenerator creates a new task content generator
func NewTaskContentGenerator(contentWidth int, context *base.ComponentContext) TaskContentGenerator {
	return TaskContentGenerator{
		contentWidth: contentWidth,
		context:      context,
	}
}

// SetTask updates the task being displayed
func (c *TaskContentGenerator) SetTask(task *archon.Task) {
	c.task = task
}

// UpdateDimensions updates the content width
// Providers are set once in constructor and don't change during resize
func (c *TaskContentGenerator) UpdateDimensions(contentWidth int) {
	c.contentWidth = contentWidth
}

// SetSearch updates search parameters
func (c *TaskContentGenerator) SetSearch(query string, active bool) {
	c.searchQuery = query
	c.searchActive = active
}

// GenerateLines produces all content lines for the task
// This replaces the scattered render methods with a single clean interface
func (c *TaskContentGenerator) GenerateLines() []string {
	if c.task == nil {
		return []string{}
	}

	// Create style factory
	factory := c.createStyleFactory()

	// Build all content by calling focused generation functions
	var allContent []string
	allContent = append(allContent, c.generateTaskHeader(c.task, factory)...)
	allContent = append(allContent, c.generateTaskMetadata(c.task, factory)...)
	allContent = append(allContent, c.generateTaskTags(c.task, factory)...)
	allContent = append(allContent, c.generateTaskDescription(c.task, factory)...)
	allContent = append(allContent, c.generateTaskTimestamps(c.task, factory)...)
	allContent = append(allContent, c.generateTaskSources(c.task, factory)...)
	allContent = append(allContent, c.generateTaskCodeExamples(c.task, factory)...)

	return allContent
}

// createStyleFactory creates a style factory for task rendering with search state
func (c *TaskContentGenerator) createStyleFactory() *styling.StyleFactory {
	styleContext := c.CreateStyleContext(false).
		WithSearch(c.searchQuery, c.searchActive)
	return styleContext.Factory()
}

// CreateStyleContext creates a StyleContext for UI components with fallback support
func (c *TaskContentGenerator) CreateStyleContext(selected bool) *styling.StyleContext {
	if c.context != nil && c.context.StyleContextProvider != nil {
		return c.context.StyleContextProvider.CreateStyleContext(selected)
	}
	// Fallback to a basic style context with minimal theme
	theme := &styling.ThemeAdapter{
		TodoColor:   "yellow",
		DoingColor:  "blue",
		ReviewColor: "orange",
		DoneColor:   "green",
		HeaderColor: "cyan",
		MutedColor:  "gray",
		Name:        "fallback",
	}
	// Create a minimal style provider for the fallback
	styleProvider := &contentFallbackStyleProvider{}
	return styling.NewStyleContext(theme, styleProvider)
}

// generateTaskHeader generates the task header and title with search highlighting
func (c *TaskContentGenerator) generateTaskHeader(task *archon.Task, factory *styling.StyleFactory) []string {
	content := make([]string, 0, 8) // Preallocate for header, title, spacing

	// Task Details header
	taskDetailsHeader := factory.Header().Render("Task Details")
	content = append(content, styling.RenderLine(taskDetailsHeader, c.contentWidth))
	content = append(content, styling.RenderLine("", c.contentWidth))

	// Title with proper styling and search highlighting using status color
	titleHeader := factory.Header().Render("Title:")
	content = append(content, styling.RenderLine(titleHeader, c.contentWidth))

	statusColor := styling.GetThemeStatusColor(task.Status)

	// Apply search highlighting via StyleFactory (same as task list)
	title := factory.ApplySearchHighlighting(task.Title, statusColor)

	// Use lipgloss Width() for proper word wrapping with full style preservation
	// Width() wraps at word boundaries AND preserves styling on all lines (including highlights)
	styledTitle := factory.Text(statusColor).Width(c.contentWidth - 2).Render(title)
	titleLines := strings.Split(styledTitle, "\n")

	for _, line := range titleLines {
		content = append(content, styling.RenderLine(line, c.contentWidth))
	}
	content = append(content, styling.RenderLine("", c.contentWidth))

	return content
}

// generateTaskMetadata generates status, assignee, and priority information
func (c *TaskContentGenerator) generateTaskMetadata(task *archon.Task, factory *styling.StyleFactory) []string {
	content := make([]string, 0, 4) // Preallocate for status, assignee, priority/order

	// Status and assignee with colors - use lipgloss.JoinHorizontal
	statusLabel := factory.Text(styling.CurrentTheme.MutedColor).Render("Status:")
	statusSymbol := factory.Text(styling.GetThemeStatusColor(task.Status)).Render(task.GetStatusSymbol())
	statusText := factory.Text(styling.GetThemeStatusColor(task.Status)).Render(strings.ToUpper(task.Status))
	statusLine := lipgloss.JoinHorizontal(lipgloss.Left, statusLabel, " ", statusSymbol, " ", statusText)
	content = append(content, styling.RenderLine(statusLine, c.contentWidth))

	assigneeLabel := factory.Text(styling.CurrentTheme.MutedColor).Render("Assignee:")
	assigneeName := factory.Text(styling.CurrentTheme.HeaderColor).Render(task.Assignee)
	assigneeLine := lipgloss.JoinHorizontal(lipgloss.Left, assigneeLabel, " ", assigneeName)
	content = append(content, styling.RenderLine(assigneeLine, c.contentWidth))

	// Priority information with color and symbol (if enabled)
	if c.context != nil && c.context.ConfigProvider != nil && c.context.ConfigProvider.IsPriorityIndicatorsEnabled() {
		priority := styling.GetTaskPriority(task.TaskOrder, nil)
		prioritySymbol := styling.GetPrioritySymbol(priority)
		priorityColor := styling.GetPriorityColor(priority)

		var priorityText string
		switch priority {
		case styling.PriorityHigh:
			priorityText = "High"
		case styling.PriorityMedium:
			priorityText = "Medium"
		case styling.PriorityLow:
			priorityText = "Low"
		default:
			priorityText = "Unknown"
		}

		priorityLabel := factory.Text(styling.CurrentTheme.MutedColor).Render("Priority:")
		styledSymbol := factory.Text(priorityColor).Render(prioritySymbol)
		styledText := factory.Text(priorityColor).Render(priorityText)
		orderText := factory.Text(styling.CurrentTheme.MutedColor).Render(fmt.Sprintf("(order: %d)", task.TaskOrder))
		priorityLine := lipgloss.JoinHorizontal(lipgloss.Left, priorityLabel, " ", styledSymbol, " ", styledText, " ", orderText)
		content = append(content, styling.RenderLine(priorityLine, c.contentWidth))
	} else {
		// Just show the raw task order when priority indicators are disabled
		taskOrderLabel := factory.Text(styling.CurrentTheme.MutedColor).Render("Task Order:")
		taskOrderValue := factory.Text(styling.CurrentTheme.MutedColor).Render(fmt.Sprintf("%d", task.TaskOrder))
		taskOrderLine := lipgloss.JoinHorizontal(lipgloss.Left, taskOrderLabel, " ", taskOrderValue)
		content = append(content, styling.RenderLine(taskOrderLine, c.contentWidth))
	}

	return content
}

// generateTaskTags generates feature tags and metadata
func (c *TaskContentGenerator) generateTaskTags(task *archon.Task, factory *styling.StyleFactory) []string {
	content := make([]string, 0, 2) // Preallocate for tags + spacing

	if task.Feature != nil && *task.Feature != "" {
		tagsLabel := factory.Text(styling.CurrentTheme.MutedColor).Render("Tags:")
		featureTag := factory.Text(styling.GetFeatureColor(*task.Feature)).Render(fmt.Sprintf("#%s", *task.Feature))
		tagsLine := lipgloss.JoinHorizontal(lipgloss.Left, tagsLabel, " ", featureTag)
		content = append(content, styling.RenderLine(tagsLine, c.contentWidth))
	}
	content = append(content, styling.RenderLine("", c.contentWidth))

	return content
}

// generateTaskDescription generates the task description with markdown
func (c *TaskContentGenerator) generateTaskDescription(task *archon.Task, factory *styling.StyleFactory) []string {
	content := make([]string, 0, 8) // Preallocate for description header + lines

	if task.Description != "" {
		descriptionHeader := factory.Header().Render("Description:")
		content = append(content, styling.RenderLine(descriptionHeader, c.contentWidth))
		descriptionContent := view.RenderMarkdown(task.Description, c.contentWidth-2)
		descriptionLines := strings.Split(descriptionContent, "\n")

		// Pad each description line to full width (markdown provides foreground styling)
		for _, line := range descriptionLines {
			content = append(content, styling.RenderLine(line, c.contentWidth))
		}
		content = append(content, styling.RenderLine("", c.contentWidth))
	}

	return content
}

// generateTaskTimestamps generates created and updated timestamps
func (c *TaskContentGenerator) generateTaskTimestamps(task *archon.Task, factory *styling.StyleFactory) []string {
	content := make([]string, 0, 2) // Preallocate for created + updated

	createdText := factory.Text(styling.CurrentTheme.MutedColor).Render(fmt.Sprintf("Created: %s", task.CreatedAt.Format("2006-01-02 15:04")))
	content = append(content, styling.RenderLine(createdText, c.contentWidth))
	updatedText := factory.Text(styling.CurrentTheme.MutedColor).Render(fmt.Sprintf("Updated: %s", task.UpdatedAt.Format("2006-01-02 15:04")))
	content = append(content, styling.RenderLine(updatedText, c.contentWidth))

	return content
}

// generateTaskSources generates the task sources list
func (c *TaskContentGenerator) generateTaskSources(task *archon.Task, factory *styling.StyleFactory) []string {
	content := make([]string, 0, len(task.Sources)+2) // Preallocate for header + sources + spacing

	if len(task.Sources) > 0 {
		content = append(content, styling.RenderLine("", c.contentWidth))
		sourcesHeader := factory.Header().Render("Sources:")
		content = append(content, styling.RenderLine(sourcesHeader, c.contentWidth))
		for _, source := range task.Sources {
			sourceText := factory.Text(styling.CurrentTheme.MutedColor).Render(fmt.Sprintf("• %s (%s)", source.URL, source.Type))
			content = append(content, styling.RenderLine(sourceText, c.contentWidth))
		}
	}

	return content
}

// generateTaskCodeExamples generates the task code examples list
func (c *TaskContentGenerator) generateTaskCodeExamples(task *archon.Task, factory *styling.StyleFactory) []string {
	content := make([]string, 0, len(task.CodeExamples)+2) // Preallocate for header + examples + spacing

	if len(task.CodeExamples) > 0 {
		content = append(content, styling.RenderLine("", c.contentWidth))
		examplesHeader := factory.Header().Render("Code Examples:")
		content = append(content, styling.RenderLine(examplesHeader, c.contentWidth))
		for _, example := range task.CodeExamples {
			exampleText := factory.Text(styling.CurrentTheme.MutedColor).Render(fmt.Sprintf("• %s - %s", example.File, example.Purpose))
			content = append(content, styling.RenderLine(exampleText, c.contentWidth))
		}
	}

	return content
}

// contentFallbackStyleProvider provides minimal styling configuration for content generation
type contentFallbackStyleProvider struct{}

func (f *contentFallbackStyleProvider) IsPriorityIndicatorsEnabled() bool { return false }
func (f *contentFallbackStyleProvider) IsFeatureColorsEnabled() bool      { return false }
