package view

import (
	"strings"
	"github.com/charmbracelet/glamour"
)

// Markdown module handles markdown rendering and text processing
// This module provides functionality for rendering markdown content and text wrapping

// RenderMarkdown renders markdown text using glamour for terminal display
func RenderMarkdown(text string, width int) string {
	// Create a glamour renderer with minimal styling to avoid conflicts
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStylePath("notty"),  // Use minimal styling without background colors
		glamour.WithWordWrap(width),     // Set word wrap to available width
	)
	if err != nil {
		// Fallback to simple word wrap if glamour fails
		return WordWrap(text, width)
	}

	// Render the markdown
	rendered, err := renderer.Render(text)
	if err != nil {
		// Fallback to simple word wrap if rendering fails
		return WordWrap(text, width)
	}

	// Remove trailing newline that glamour adds
	return strings.TrimSuffix(rendered, "\n")
}

// WordWrap wraps text to fit within the specified width (fallback function)
func WordWrap(text string, width int) string {
	if len(text) <= width {
		return text
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	currentLine := words[0]

	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	lines = append(lines, currentLine)

	return strings.Join(lines, "\n")
}