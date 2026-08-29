package view

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Highlighting module handles search highlighting and text styling
// This module provides functionality for highlighting search terms and text processing

// HighlightSearchTerms highlights search query matches with default styling
func HighlightSearchTerms(text, query string) string {
	return HighlightSearchTermsWithColor(text, query, "0") // Default black text
}

// IsColorConflictWithHighlight checks if the given color conflicts with highlight background
// Returns true if the colors have poor contrast with bright yellow background causing poor readability
func IsColorConflictWithHighlight(textColor string) bool {
	// Highlight background is bright yellow "11"
	// Colors that have poor contrast with bright yellow background
	lowContrastColors := map[string]bool{
		// Yellow family - similar hues
		"11":  true, // Bright yellow (exact match)
		"220": true, // Yellow (used in themes)
		"228": true, // Yellow (dracula theme)
		"214": true, // Orange-yellow
		"226": true, // Bright yellow
		"227": true, // Yellow

		// Light colors - insufficient contrast
		"15":  true, // White (too light)
		"45":  true, // Light blue (too light)
		"46":  true, // Light green (too light)
		"84":  true, // Light green (dracula theme)
		"117": true, // Light cyan (too light)
		"118": true, // Light green (monokai theme)
		"142": true, // Light green (gruvbox theme)

		// Other light colors that may appear in themes
		"51":  true, // Light cyan
		"75":  true, // Light blue
		"81":  true, // Light green
		"231": true, // Very light/white
	}
	return lowContrastColors[textColor]
}

// HighlightSearchTermsWithColor highlights search query matches while preserving specified text color
func HighlightSearchTermsWithColor(text, query, textColor string) string {
	if query == "" {
		return text
	}

	// Trim and normalize query
	query = strings.TrimSpace(query)
	if query == "" {
		return text
	}

	// Style for highlighted text - bright yellow foreground
	highlightStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")). // Bright yellow foreground
		Bold(true)

	// Style for non-highlighted text - foreground only (background handled by styling.RenderLine)
	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(textColor)) // Apply status color to non-highlighted parts
		// Background removed: handled by styling.RenderLine()

	// Split query into individual words for multi-word highlighting
	queryWords := strings.Fields(strings.ToLower(query))
	if len(queryWords) == 0 {
		return text
	}

	result := text

	// Highlight each word in the query
	for _, word := range queryWords {
		if word == "" {
			continue
		}

		// Find all occurrences of this word (case-insensitive)
		lowerText := strings.ToLower(result)
		lowerWord := strings.ToLower(word)

		var highlighted strings.Builder
		lastEnd := 0

		for {
			// Find next occurrence
			index := strings.Index(lowerText[lastEnd:], lowerWord)
			if index == -1 {
				// No more matches, append remaining text with status color
				remainingText := result[lastEnd:]
				if remainingText != "" {
					highlighted.WriteString(normalStyle.Render(remainingText))
				}
				break
			}

			// Adjust index to absolute position
			absoluteIndex := lastEnd + index

			// Add text before the match with status color
			beforeText := result[lastEnd:absoluteIndex]
			if beforeText != "" {
				highlighted.WriteString(normalStyle.Render(beforeText))
			}

			// Add highlighted match (preserve original case)
			matchText := result[absoluteIndex : absoluteIndex+len(word)]
			highlighted.WriteString(highlightStyle.Render(matchText))

			// Move past this match
			lastEnd = absoluteIndex + len(word)
		}

		result = highlighted.String()
		// Note: lowerText will be recalculated in next iteration if needed
	}

	return result
}

// StripANSI removes ANSI escape sequences from text for accurate length calculation
func StripANSI(text string) string {
	// This is a simplified ANSI stripping - lipgloss.Width() is more accurate
	// but this can be used when we need the plain text
	result := ""
	inEscape := false

	for idx, char := range text {
		if char == '\x1b' && idx+1 < len(text) && text[idx+1] == '[' {
			inEscape = true
			continue
		}
		if inEscape {
			if char >= 'A' && char <= 'Z' || char >= 'a' && char <= 'z' {
				inEscape = false
			}
			continue
		}
		result += string(char)
	}

	return result
}

// TruncatePreservingANSI truncates text while preserving ANSI escape sequences
func TruncatePreservingANSI(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	// If the visual width is already within limits, return as-is
	visualWidth := lipgloss.Width(text)
	if visualWidth <= maxWidth {
		return text
	}

	// Need to truncate - use single character ellipsis
	if maxWidth == 1 {
		return "…"
	}

	// For ANSI text, we need to be careful about truncation
	// This is a simplified approach - for more complex cases,
	// we'd need to parse ANSI sequences properly
	plainText := StripANSI(text)
	if len(plainText) <= maxWidth-1 {
		return text // ANSI codes don't add much, keep original
	}

	// Truncate the plain text and try to preserve some styling
	truncatedPlain := plainText[:maxWidth-1] + "…"

	// Try to extract any styling from the beginning of the original text
	// and apply it to the truncated text
	if strings.HasPrefix(text, "\x1b[") {
		// Find the end of the first ANSI sequence
		endIdx := strings.Index(text, "m")
		if endIdx != -1 && endIdx < 20 { // Reasonable limit for ANSI sequence
			ansiPrefix := text[:endIdx+1]
			return ansiPrefix + truncatedPlain + "\x1b[0m" // Reset at end
		}
	}

	return truncatedPlain
}
