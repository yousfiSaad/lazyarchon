package projectdetails

import (
	"fmt"
	"strings"

	"github.com/yousfisaad/lazyarchon/v2/internal/shared/interfaces"

	"github.com/charmbracelet/lipgloss"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/styling"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/utils/view"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/base"
)

// ProjectContentGenerator handles pure content generation for project details
// Separated from UI concerns for clean architecture
type ProjectContentGenerator struct {
	project      *interfaces.Project
	contentWidth int

	// Component context for accessing dependencies
	context *base.ComponentContext
}

// NewProjectContentGenerator creates a new project content generator
func NewProjectContentGenerator(contentWidth int, context *base.ComponentContext) ProjectContentGenerator {
	return ProjectContentGenerator{
		contentWidth: contentWidth,
		context:      context,
	}
}

// SetProject updates the project being displayed
func (c *ProjectContentGenerator) SetProject(project *interfaces.Project) {
	c.project = project
}

// UpdateDimensions updates the content width
// Providers are set once in constructor and don't change during resize
func (c *ProjectContentGenerator) UpdateDimensions(contentWidth int) {
	c.contentWidth = contentWidth
}

// GenerateLines produces all content lines for the project
func (c *ProjectContentGenerator) GenerateLines() []string {
	if c.project == nil {
		return []string{}
	}

	// Create style factory
	factory := c.createStyleFactory()

	// Build all content by calling focused generation functions
	var allContent []string
	allContent = append(allContent, c.generateProjectHeader(c.project, factory)...)
	allContent = append(allContent, c.generateProjectInfo(c.project, factory)...)
	allContent = append(allContent, c.generateProjectMetadata(c.project, factory)...)

	return allContent
}

// createStyleFactory creates a style factory for project rendering
func (c *ProjectContentGenerator) createStyleFactory() *styling.StyleFactory {
	styleContext := c.CreateStyleContext(false)
	return styleContext.Factory()
}

// CreateStyleContext creates a StyleContext for UI components with fallback support
func (c *ProjectContentGenerator) CreateStyleContext(selected bool) *styling.StyleContext {
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

// generateProjectHeader generates the project header and title
func (c *ProjectContentGenerator) generateProjectHeader(project *interfaces.Project, factory *styling.StyleFactory) []string {
	content := make([]string, 0, 8) // Preallocate for header, title, spacing

	// Project Details header
	projectDetailsHeader := factory.Header().Render("Project Details")
	content = append(content, styling.RenderLine(projectDetailsHeader, c.contentWidth))
	content = append(content, styling.RenderLine("", c.contentWidth))

	// Title with proper styling
	titleHeader := factory.Header().Render("Title:")
	content = append(content, styling.RenderLine(titleHeader, c.contentWidth))

	title := factory.Text(styling.CurrentTheme.HeaderColor).Render(project.Title)
	titleLines := strings.Split(view.WordWrap(title, c.contentWidth-2), "\n")
	for _, line := range titleLines {
		content = append(content, styling.RenderLine(line, c.contentWidth))
	}
	content = append(content, styling.RenderLine("", c.contentWidth))

	return content
}

// generateProjectInfo generates description and GitHub repo information
func (c *ProjectContentGenerator) generateProjectInfo(project *interfaces.Project, factory *styling.StyleFactory) []string {
	content := make([]string, 0, 8) // Preallocate for description, repo, spacing

	// Description (if present)
	if project.Description != "" {
		descHeader := factory.Header().Render("Description:")
		content = append(content, styling.RenderLine(descHeader, c.contentWidth))

		// Word wrap description text
		descLines := strings.Split(view.WordWrap(project.Description, c.contentWidth-2), "\n")
		for _, line := range descLines {
			content = append(content, styling.RenderLine(line, c.contentWidth))
		}
		content = append(content, styling.RenderLine("", c.contentWidth))
	}

	// GitHub repo (if present) - stored in Extra["github_repo"] for plugin compatibility
	if githubRepo, ok := project.Extra["github_repo"].(*string); ok && githubRepo != nil && *githubRepo != "" {
		repoLabel := factory.Text(styling.CurrentTheme.MutedColor).Render("GitHub:")
		repoURL := factory.Text(styling.CurrentTheme.HeaderColor).Render(*githubRepo)
		repoLine := lipgloss.JoinHorizontal(lipgloss.Left, repoLabel, " ", repoURL)
		content = append(content, styling.RenderLine(repoLine, c.contentWidth))
		content = append(content, styling.RenderLine("", c.contentWidth))
	}

	return content
}

// getExtraMapLength safely gets the length of a map stored in Extra
func getExtraMapLength(extra map[string]interface{}, key string) int {
	if val, ok := extra[key]; ok {
		if m, ok := val.(map[string]interface{}); ok {
			return len(m)
		}
	}
	return 0
}

// getExtraSliceLength safely gets the length of a slice stored in Extra
func getExtraSliceLength(extra map[string]interface{}, key string) int {
	if val, ok := extra[key]; ok {
		if s, ok := val.([]interface{}); ok {
			return len(s)
		}
	}
	return 0
}

// generateProjectMetadata generates features, docs count, and timestamps
func (c *ProjectContentGenerator) generateProjectMetadata(project *interfaces.Project, factory *styling.StyleFactory) []string {
	content := make([]string, 0, 6) // Preallocate for features, docs, timestamps, spacing

	// Features count (if present) - stored in Extra["features"]
	featuresCount := getExtraMapLength(project.Extra, "features")
	if featuresCount > 0 {
		featuresLabel := factory.Text(styling.CurrentTheme.MutedColor).Render("Features:")
		featuresCountStr := factory.Text(styling.CurrentTheme.HeaderColor).Render(fmt.Sprintf("%d", featuresCount))
		featuresLine := lipgloss.JoinHorizontal(lipgloss.Left, featuresLabel, " ", featuresCountStr)
		content = append(content, styling.RenderLine(featuresLine, c.contentWidth))
	}

	// Docs count (if present) - stored in Extra["docs"]
	docsCount := getExtraSliceLength(project.Extra, "docs")
	if docsCount > 0 {
		docsLabel := factory.Text(styling.CurrentTheme.MutedColor).Render("Documents:")
		docsCountStr := factory.Text(styling.CurrentTheme.HeaderColor).Render(fmt.Sprintf("%d", docsCount))
		docsLine := lipgloss.JoinHorizontal(lipgloss.Left, docsLabel, " ", docsCountStr)
		content = append(content, styling.RenderLine(docsLine, c.contentWidth))
	}

	// Timestamps
	content = append(content, styling.RenderLine("", c.contentWidth))
	createdLabel := factory.Text(styling.CurrentTheme.MutedColor).Render("Created:")
	createdDate := factory.Text(styling.CurrentTheme.HeaderColor).Render(project.CreatedAt.Format("2006-01-02 15:04"))
	createdLine := lipgloss.JoinHorizontal(lipgloss.Left, createdLabel, " ", createdDate)
	content = append(content, styling.RenderLine(createdLine, c.contentWidth))

	updatedLabel := factory.Text(styling.CurrentTheme.MutedColor).Render("Updated:")
	updatedDate := factory.Text(styling.CurrentTheme.HeaderColor).Render(project.UpdatedAt.Format("2006-01-02 15:04"))
	updatedLine := lipgloss.JoinHorizontal(lipgloss.Left, updatedLabel, " ", updatedDate)
	content = append(content, styling.RenderLine(updatedLine, c.contentWidth))

	return content
}

// contentFallbackStyleProvider provides minimal styling configuration for tests
type contentFallbackStyleProvider struct{}

func (f *contentFallbackStyleProvider) IsPriorityIndicatorsEnabled() bool { return false }
func (f *contentFallbackStyleProvider) IsFeatureColorsEnabled() bool      { return false }
