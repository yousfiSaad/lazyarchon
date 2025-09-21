package ui

import (
	"fmt"
	"strings"

	"github.com/yousfisaad/lazyarchon/internal/archon"
	"github.com/yousfisaad/lazyarchon/internal/ui/styling"
)

// calculateTaskDetailsWidth calculates the effective content width for task details
func (m *Model) calculateTaskDetailsWidth() int {
	panelWidth := m.GetRightPanelWidth()
	// Calculate effective content width: panel width minus borders and padding
	// BorderWidth=2 + 2*PanelPadding=2 = 4 total reduction
	return panelWidth - 4
}

// createTaskStyleContext creates a style context for task rendering
func (m *Model) createTaskStyleContext() (*styling.StyleFactory, int) {
	styleContext := m.CreateStyleContext(false)
	factory := styleContext.Factory()
	contentWidth := m.calculateTaskDetailsWidth()
	return factory, contentWidth
}

// validateTaskSelection validates task selection and returns the selected task
func (m *Model) validateTaskSelection() (*archon.Task, bool) {
	sortedTasks := m.GetSortedTasks()
	if len(sortedTasks) == 0 || m.Navigation.selectedIndex >= len(sortedTasks) {
		return nil, false
	}
	return &sortedTasks[m.Navigation.selectedIndex], true
}

// renderTaskHeader renders the task header and title with search highlighting
func (m *Model) renderTaskHeader(task *archon.Task, factory *styling.StyleFactory, contentWidth int) []string {
	var content []string

	// Task Details header
	taskDetailsHeader := factory.Header().Render("Task Details")
	content = append(content, styling.RenderLine(taskDetailsHeader, contentWidth))
	content = append(content, styling.RenderLine("", contentWidth))

	// Title with proper styling and search highlighting using status color
	titleHeader := factory.Header().Render("Title:")
	content = append(content, styling.RenderLine(titleHeader, contentWidth))

	title := task.Title
	statusColor := styling.GetThemeStatusColor(task.Status)
	if m.Data.searchActive && m.Data.searchQuery != "" {
		// Use status color for both highlighted and non-highlighted text in content panel
		title = highlightSearchTermsWithColor(task.Title, m.Data.searchQuery, statusColor)
	} else {
		// Apply foreground-only styling to plain title when no search is active
		title = factory.Text(statusColor).Render(task.Title)
	}

	titleLines := strings.Split(wordWrap(title, contentWidth-2), "\n")
	for _, line := range titleLines {
		content = append(content, styling.RenderLine(line, contentWidth))
	}
	content = append(content, styling.RenderLine("", contentWidth))

	return content
}

// renderTaskMetadata renders status, assignee, and priority information
func (m *Model) renderTaskMetadata(task *archon.Task, factory *styling.StyleFactory, contentWidth int) []string {
	var content []string

	// Status and assignee with colors - build complete styled strings first
	statusLabel := factory.Text(styling.CurrentTheme.MutedColor).Render("Status:")
	statusSymbol := factory.Text(styling.GetThemeStatusColor(task.Status)).Render(task.GetStatusSymbol())
	statusText := factory.Text(styling.GetThemeStatusColor(task.Status)).Render(strings.ToUpper(task.Status))
	statusLine := fmt.Sprintf("%s %s %s", statusLabel, statusSymbol, statusText)
	content = append(content, styling.RenderLine(statusLine, contentWidth))

	assigneeLabel := factory.Text(styling.CurrentTheme.MutedColor).Render("Assignee:")
	assigneeName := factory.Text(styling.CurrentTheme.HeaderColor).Render(task.Assignee)
	assigneeLine := fmt.Sprintf("%s %s", assigneeLabel, assigneeName)
	content = append(content, styling.RenderLine(assigneeLine, contentWidth))

	// Priority information with color and symbol (if enabled)
	if m.config.IsPriorityIndicatorsEnabled() {
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
		priorityLine := fmt.Sprintf("%s %s %s %s", priorityLabel, styledSymbol, styledText, orderText)
		content = append(content, styling.RenderLine(priorityLine, contentWidth))
	} else {
		// Just show the raw task order when priority indicators are disabled
		taskOrderLabel := factory.Text(styling.CurrentTheme.MutedColor).Render("Task Order:")
		taskOrderValue := factory.Text(styling.CurrentTheme.MutedColor).Render(fmt.Sprintf("%d", task.TaskOrder))
		taskOrderLine := fmt.Sprintf("%s %s", taskOrderLabel, taskOrderValue)
		content = append(content, styling.RenderLine(taskOrderLine, contentWidth))
	}

	return content
}

// renderTaskTags renders feature tags and metadata
func (m *Model) renderTaskTags(task *archon.Task, factory *styling.StyleFactory, contentWidth int) []string {
	var content []string

	if task.Feature != nil && *task.Feature != "" {
		tagsLabel := factory.Text(styling.CurrentTheme.MutedColor).Render("Tags:")
		featureTag := factory.Text(styling.GetFeatureColor(*task.Feature)).Render(fmt.Sprintf("#%s", *task.Feature))
		tagsLine := fmt.Sprintf("%s %s", tagsLabel, featureTag)
		content = append(content, styling.RenderLine(tagsLine, contentWidth))
	}
	content = append(content, styling.RenderLine("", contentWidth))

	return content
}

// renderTaskDescription renders the task description with markdown
func (m *Model) renderTaskDescription(task *archon.Task, factory *styling.StyleFactory, contentWidth int) []string {
	var content []string

	if task.Description != "" {
		descriptionHeader := factory.Header().Render("Description:")
		content = append(content, styling.RenderLine(descriptionHeader, contentWidth))
		descriptionContent := renderMarkdown(task.Description, contentWidth-2)
		descriptionLines := strings.Split(descriptionContent, "\n")

		// Pad each description line to full width (markdown provides foreground styling)
		for _, line := range descriptionLines {
			content = append(content, styling.RenderLine(line, contentWidth))
		}
		content = append(content, styling.RenderLine("", contentWidth))
	}

	return content
}

// renderTaskTimestamps renders created and updated timestamps
func (m *Model) renderTaskTimestamps(task *archon.Task, factory *styling.StyleFactory, contentWidth int) []string {
	var content []string

	createdText := factory.Text(styling.CurrentTheme.MutedColor).Render(fmt.Sprintf("Created: %s", task.CreatedAt.Format("2006-01-02 15:04")))
	content = append(content, styling.RenderLine(createdText, contentWidth))
	updatedText := factory.Text(styling.CurrentTheme.MutedColor).Render(fmt.Sprintf("Updated: %s", task.UpdatedAt.Format("2006-01-02 15:04")))
	content = append(content, styling.RenderLine(updatedText, contentWidth))

	return content
}

// renderTaskSources renders the task sources list
func (m *Model) renderTaskSources(task *archon.Task, factory *styling.StyleFactory, contentWidth int) []string {
	var content []string

	if len(task.Sources) > 0 {
		content = append(content, styling.RenderLine("", contentWidth))
		sourcesHeader := factory.Header().Render("Sources:")
		content = append(content, styling.RenderLine(sourcesHeader, contentWidth))
		for _, source := range task.Sources {
			sourceText := factory.Text(styling.CurrentTheme.MutedColor).Render(fmt.Sprintf("• %s (%s)", source.URL, source.Type))
			content = append(content, styling.RenderLine(sourceText, contentWidth))
		}
	}

	return content
}

// renderTaskCodeExamples renders the task code examples list
func (m *Model) renderTaskCodeExamples(task *archon.Task, factory *styling.StyleFactory, contentWidth int) []string {
	var content []string

	if len(task.CodeExamples) > 0 {
		content = append(content, styling.RenderLine("", contentWidth))
		examplesHeader := factory.Header().Render("Code Examples:")
		content = append(content, styling.RenderLine(examplesHeader, contentWidth))
		for _, example := range task.CodeExamples {
			exampleText := factory.Text(styling.CurrentTheme.MutedColor).Render(fmt.Sprintf("• %s - %s", example.File, example.Purpose))
			content = append(content, styling.RenderLine(exampleText, contentWidth))
		}
	}

	return content
}

// updateTaskDetailsViewport updates the task details viewport with rendered content
func (m *Model) updateTaskDetailsViewport() {
	// Validate task selection
	task, valid := m.validateTaskSelection()
	if !valid {
		m.taskDetailsViewport.SetContent("No task selected")
		return
	}

	// Create style context
	factory, contentWidth := m.createTaskStyleContext()

	// Build all content by calling focused rendering functions
	var allContent []string
	allContent = append(allContent, m.renderTaskHeader(task, factory, contentWidth)...)
	allContent = append(allContent, m.renderTaskMetadata(task, factory, contentWidth)...)
	allContent = append(allContent, m.renderTaskTags(task, factory, contentWidth)...)
	allContent = append(allContent, m.renderTaskDescription(task, factory, contentWidth)...)
	allContent = append(allContent, m.renderTaskTimestamps(task, factory, contentWidth)...)
	allContent = append(allContent, m.renderTaskSources(task, factory, contentWidth)...)
	allContent = append(allContent, m.renderTaskCodeExamples(task, factory, contentWidth)...)

	// Set the content in the viewport
	content := strings.Join(allContent, "\n")
	m.taskDetailsViewport.SetContent(content)
}

// updateHelpModalViewport updates the help modal viewport with current help content
// This builds the same rich content as the original getHelpContent with full styling
func (m *Model) updateHelpModalViewport() {
	// Create style context for help content styling
	styleContext := m.CreateStyleContext(false)
	factory := styleContext.Factory()

	// Build all content with original styling - exactly like getHelpContent
	var help []string

	// Title
	help = append(help, factory.Header().Render("LazyArchon Help"))
	help = append(help, "")

	// Panel Navigation section
	help = append(help, factory.Header().Render("Panel Navigation:"))
	help = append(help, "  h/l          Switch between panels")
	help = append(help, "  ↑/↓ or j/k   Navigate/scroll (1 line)")
	help = append(help, "  J/K          Fast scroll (4 lines)")
	help = append(help, "  Ctrl+u/d     Half-page scroll")
	help = append(help, "  gg/G         Jump to top/bottom")
	help = append(help, "  Home/End     Jump to start/end")
	help = append(help, "")

	// Project Management section
	help = append(help, factory.Header().Render("Project Management:"))
	help = append(help, "  p            Project selection mode")
	help = append(help, "  a            Show all tasks")
	help = append(help, "  Enter        Select project")
	help = append(help, "  Esc          Exit project mode")
	help = append(help, "")

	// Task List section (when left panel is active)
	help = append(help, factory.Header().Render("Task List (when Tasks panel active):"))
	help = append(help, "  s/S          Sort tasks by different criteria")
	help = append(help, "")

	// Task Management section
	help = append(help, factory.Header().Render("Task Management:"))
	help = append(help, "  t            Change task status (Todo/Doing/Review/Done)")
	help = append(help, "  y            Copy task ID to clipboard (yank)")
	help = append(help, "  Y            Copy task title to clipboard (yank)")
	help = append(help, "")

	// Application Controls section
	help = append(help, factory.Header().Render("Application Controls:"))
	help = append(help, "  r or F5      Refresh data from API")
	help = append(help, "  q            Quit application")
	help = append(help, "  ?            Toggle this help")
	help = append(help, "")

	// Visual Indicators section
	help = append(help, factory.Header().Render("Visual Indicators:"))
	help = append(help, "  Bright cyan border    Active panel")
	help = append(help, "  Dim gray border       Inactive panel")
	help = append(help, "  [Tasks]/[Details]     Active panel in status bar")
	help = append(help, "  ▓░ scroll bar         Position indicator")
	help = append(help, "")

	// Task Status Symbols section
	help = append(help, factory.Header().Render("Task Status Symbols:"))
	help = append(help, "  "+styling.StatusSymbolTodo+"  Todo       Not started")
	help = append(help, "  "+styling.StatusSymbolDoing+"  Doing      In progress")
	help = append(help, "  "+styling.StatusSymbolReview+"  Review     Under review")
	help = append(help, "  "+styling.StatusSymbolDone+"  Done       Completed")
	help = append(help, "")

	// Help Navigation section
	help = append(help, factory.Header().Render("Help Navigation:"))
	help = append(help, "  j/k          Scroll help (1 line)")
	help = append(help, "  J/K          Fast scroll help (4 lines)")
	help = append(help, "  Ctrl+u/d     Half-page scroll help")
	help = append(help, "  gg/G         Jump to help top/bottom")
	help = append(help, "")

	// Footer
	help = append(help, factory.Italic(styling.CurrentTheme.MutedColor).Render("Press ? or ESC to close this help"))

	// Set the content in the viewport
	content := strings.Join(help, "\n")
	m.helpModalViewport.SetContent(content)
}