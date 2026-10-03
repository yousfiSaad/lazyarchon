package taskdetails

import (
	"fmt"
	"strings"

	"github.com/yousfisaad/lazyarchon/v2/internal/shared/deps"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/interfaces"

	"github.com/charmbracelet/lipgloss"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/styling"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/utils/view"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/base"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/context"
)

// TaskContentGenerator handles pure content generation for task details
// Separated from UI concerns for clean architecture
type TaskContentGenerator struct {
	task         *interfaces.Task
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
func (c *TaskContentGenerator) SetTask(task *interfaces.Task) {
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
	allContent := make([]string, 0, 48) // eight generators, each a handful of lines
	allContent = append(allContent, c.generateTaskHeader(c.task, factory)...)
	allContent = append(allContent, c.generateTaskMetadata(c.task, factory)...)
	allContent = append(allContent, c.generateTaskDependencies(c.task, factory)...)
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
func (c *TaskContentGenerator) generateTaskHeader(task *interfaces.Task, factory *styling.StyleFactory) []string {
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
func (c *TaskContentGenerator) generateTaskMetadata(task *interfaces.Task, factory *styling.StyleFactory) []string {
	content := make([]string, 0, 5) // Preallocate for status, assignee, worktree, priority

	// Status and assignee with colors - use lipgloss.JoinHorizontal
	statusLabel := factory.Text(styling.CurrentTheme.MutedColor).Render("Status:")
	statusSymbol := factory.Text(styling.GetThemeStatusColor(task.Status)).Render(interfaces.GetStatusSymbol(task.Status))
	statusText := factory.Text(styling.GetThemeStatusColor(task.Status)).Render(strings.ToUpper(task.Status))
	statusLine := lipgloss.JoinHorizontal(lipgloss.Left, statusLabel, " ", statusSymbol, " ", statusText)
	content = append(content, styling.RenderLine(statusLine, c.contentWidth))

	assigneeLabel := factory.Text(styling.CurrentTheme.MutedColor).Render("Assignee:")
	assigneeName := factory.Text(styling.CurrentTheme.HeaderColor).Render(task.Assignee)
	assigneeLine := lipgloss.JoinHorizontal(lipgloss.Left, assigneeLabel, " ", assigneeName)
	content = append(content, styling.RenderLine(assigneeLine, c.contentWidth))

	// Worktree attribution, only when a parallel session claimed the task
	if task.Worktree != "" {
		worktreeLabel := factory.Text(styling.CurrentTheme.MutedColor).Render("Worktree:")
		worktreeName := factory.Text(styling.CurrentTheme.HeaderColor).Render(task.Worktree)
		worktreeLine := lipgloss.JoinHorizontal(lipgloss.Left, worktreeLabel, " ", worktreeName)
		content = append(content, styling.RenderLine(worktreeLine, c.contentWidth))
	}

	// Priority information with color and symbol (if enabled)
	if c.context != nil && c.context.ConfigProvider != nil && c.context.ConfigProvider.IsPriorityIndicatorsEnabled() {
		priority := styling.GetTaskPriority(task.Priority)
		prioritySymbol := styling.GetPrioritySymbol(priority)
		priorityColor := styling.GetPriorityColor(priority)

		priorityLabel := factory.Text(styling.CurrentTheme.MutedColor).Render("Priority:")
		styledSymbol := factory.Text(priorityColor).Render(prioritySymbol)
		styledText := factory.Text(priorityColor).Render(styling.PriorityLabel(task.Priority))
		priorityLine := lipgloss.JoinHorizontal(lipgloss.Left, priorityLabel, " ", styledSymbol, " ", styledText)
		content = append(content, styling.RenderLine(priorityLine, c.contentWidth))
	} else {
		// Just show the raw task priority when priority indicators are disabled
		priorityLabel := factory.Text(styling.CurrentTheme.MutedColor).Render("Priority:")
		priorityValue := factory.Text(styling.CurrentTheme.MutedColor).Render(fmt.Sprintf("%d", task.Priority))
		priorityLine := lipgloss.JoinHorizontal(lipgloss.Left, priorityLabel, " ", priorityValue)
		content = append(content, styling.RenderLine(priorityLine, c.contentWidth))
	}

	return content
}

// generateTaskDependencies generates blocked-by/blocks lines, shown only
// when the task declares or blocks at least one other task.
func (c *TaskContentGenerator) generateTaskDependencies(task *interfaces.Task, factory *styling.StyleFactory) []string {
	blockedBy := c.dependencyLabels(task.BlockedBy)
	blocks := c.dependencyLabels(c.dependentsOf(task))

	if len(blockedBy) == 0 && len(blocks) == 0 {
		return nil
	}

	content := make([]string, 0, 2)

	if len(blockedBy) > 0 {
		blockedByLabel := factory.Text(styling.CurrentTheme.MutedColor).Render("Blocked by:")
		blockedByText := factory.Text(styling.CurrentTheme.HeaderColor).Render(strings.Join(blockedBy, ", "))
		blockedByLine := lipgloss.JoinHorizontal(lipgloss.Left, blockedByLabel, " ", blockedByText)
		content = append(content, styling.RenderLine(blockedByLine, c.contentWidth))
	}

	if len(blocks) > 0 {
		blocksLabel := factory.Text(styling.CurrentTheme.MutedColor).Render("Blocks:")
		blocksText := factory.Text(styling.CurrentTheme.HeaderColor).Render(strings.Join(blocks, ", "))
		blocksLine := lipgloss.JoinHorizontal(lipgloss.Left, blocksLabel, " ", blocksText)
		content = append(content, styling.RenderLine(blocksLine, c.contentWidth))
	}

	// How deep the unfinished chain runs ahead of this task. Ready tasks
	// (depth 1) and no-deps tasks (skipped above) show no line.
	if depth := c.chainDepth(task); depth > 1 {
		tasksAhead := depth - 1
		plural := "s"
		if tasksAhead == 1 {
			plural = ""
		}
		depthLabel := factory.Text(styling.CurrentTheme.MutedColor).Render("Depth:")
		depthText := factory.Text(styling.CurrentTheme.HeaderColor).Render(fmt.Sprintf(
			"%d (%d task%s ahead on the longest unfinished chain)", depth, tasksAhead, plural))
		depthLine := lipgloss.JoinHorizontal(lipgloss.Left, depthLabel, " ", depthText)
		content = append(content, styling.RenderLine(depthLine, c.contentWidth))
	}

	return content
}

// chainDepth returns the node count of the longest path of unfinished
// blockers ahead of the task, including the task itself (1 = ready), or 0
// when this generator runs without task data.
func (c *TaskContentGenerator) chainDepth(task *interfaces.Task) int {
	programContext := c.programContext()
	if programContext == nil {
		return 0
	}

	return deps.Build(programContext.Tasks).ChainLength(task.ID)
}

// dependencyLabels renders dependency IDs as "title (id)" entries, using
// the loaded tasks for titles. Unknown IDs fall back to the ID alone.
func (c *TaskContentGenerator) dependencyLabels(ids []string) []string {
	if len(ids) == 0 {
		return nil
	}

	programContext := c.programContext()

	labels := make([]string, 0, len(ids))
	for _, id := range ids {
		if programContext != nil {
			title := programContext.TaskTitleByID(id)
			labels = append(labels, fmt.Sprintf("%s (%s)", title, id))
			continue
		}

		// No task data to resolve titles — show the raw IDs.
		labels = append(labels, id)
	}

	return labels
}

// dependentsOf returns the IDs of tasks blocked by this task, derived from
// the loaded tasks' BlockedBy edges (the seam only populates Blocks on
// GetTask, which list views never see).
func (c *TaskContentGenerator) dependentsOf(task *interfaces.Task) []string {
	programContext := c.programContext()
	if programContext == nil {
		return nil
	}

	dependents := make([]string, 0)
	for _, candidate := range programContext.Tasks {
		for _, blockerID := range candidate.BlockedBy {
			if blockerID == task.ID {
				dependents = append(dependents, candidate.ID)
				break
			}
		}
	}

	return dependents
}

// programContext returns the shared program context, or nil when this
// generator runs without one (fallback rendering).
func (c *TaskContentGenerator) programContext() *context.ProgramContext {
	if c.context == nil || c.context.ProgramContext == nil {
		return nil
	}

	return c.context.ProgramContext
}

// generateTaskTags generates feature tags and metadata
func (c *TaskContentGenerator) generateTaskTags(task *interfaces.Task, factory *styling.StyleFactory) []string {
	content := make([]string, 0, 2) // Preallocate for tags + spacing

	if len(task.Tags) > 0 && task.Tags[0] != "" {
		tagsLabel := factory.Text(styling.CurrentTheme.MutedColor).Render("Tags:")
		featureTag := factory.Text(styling.GetFeatureColor(task.Tags[0])).Render(fmt.Sprintf("#%s", task.Tags[0]))
		tagsLine := lipgloss.JoinHorizontal(lipgloss.Left, tagsLabel, " ", featureTag)
		content = append(content, styling.RenderLine(tagsLine, c.contentWidth))
	}
	content = append(content, styling.RenderLine("", c.contentWidth))

	return content
}

// generateTaskDescription generates the task description with markdown
func (c *TaskContentGenerator) generateTaskDescription(task *interfaces.Task, factory *styling.StyleFactory) []string {
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

// generateTaskTimestamps generates created and updated timestamps, plus the
// archived marker when the task is archived
func (c *TaskContentGenerator) generateTaskTimestamps(task *interfaces.Task, factory *styling.StyleFactory) []string {
	content := make([]string, 0, 3) // Preallocate for created + updated (+ archived)

	createdText := factory.Text(styling.CurrentTheme.MutedColor).Render(fmt.Sprintf("Created: %s", task.CreatedAt.Format("2006-01-02 15:04")))
	content = append(content, styling.RenderLine(createdText, c.contentWidth))
	updatedText := factory.Text(styling.CurrentTheme.MutedColor).Render(fmt.Sprintf("Updated: %s", task.UpdatedAt.Format("2006-01-02 15:04")))
	content = append(content, styling.RenderLine(updatedText, c.contentWidth))

	if task.Archived {
		when := "yes"
		if task.ArchivedAt != nil {
			when = task.ArchivedAt.Format("2006-01-02 15:04")
		}
		archivedText := factory.Text(styling.CurrentTheme.MutedColor).Render(fmt.Sprintf("Archived: %s", when))
		content = append(content, styling.RenderLine(archivedText, c.contentWidth))
	}

	return content
}

// generateTaskSources generates the task sources list
func (c *TaskContentGenerator) generateTaskSources(task *interfaces.Task, factory *styling.StyleFactory) []string {
	// Sources are stored in Extra["sources"] for plugin compatibility
	// For now, return empty - can be enhanced later to type-assert from Extra map
	_ = task
	_ = factory
	return []string{}
}

// generateTaskCodeExamples generates the task code examples list
func (c *TaskContentGenerator) generateTaskCodeExamples(task *interfaces.Task, factory *styling.StyleFactory) []string {
	// CodeExamples are stored in Extra["code_examples"] for plugin compatibility
	// For now, return empty - can be enhanced later to type-assert from Extra map
	_ = task
	_ = factory
	return []string{}
}

// contentFallbackStyleProvider provides minimal styling configuration for content generation
type contentFallbackStyleProvider struct{}

func (f *contentFallbackStyleProvider) IsPriorityIndicatorsEnabled() bool { return false }
func (f *contentFallbackStyleProvider) IsFeatureColorsEnabled() bool      { return false }
