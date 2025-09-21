package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

// model_state.go contains core Model utility functions and UI helper methods.
// This file has been refactored to focus on:
// - Sort mode management (CycleSortMode, CycleSortModePrevious)
// - Scroll handling and viewport management
// - Text processing utilities (wordWrap, truncation)
// - Copy-to-clipboard functionality
//
// Related files:
// - model_state_navigation.go: Navigation and task selection
// - model_state_modal.go: Modal feature management
// - model_state_data.go: Data state management and updates
// - model_core.go: Search functionality (existing)

// CycleSortMode cycles to the next sort mode
func (m *Model) CycleSortMode() {
	// Remember currently selected task
	var selectedTaskID string
	sortedTasks := m.GetSortedTasks()
	if len(sortedTasks) > 0 && m.Navigation.selectedIndex < len(sortedTasks) {
		selectedTaskID = sortedTasks[m.Navigation.selectedIndex].ID
	}

	// Change sort mode
	m.Data.sortMode = (m.Data.sortMode + 1) % 4

	// Find the same task in new sort order and select it
	m.findAndSelectTask(selectedTaskID)

	// Update search matches for new sort order
	m.updateSearchMatches()
}

// CycleSortModePrevious cycles to the previous sort mode
func (m *Model) CycleSortModePrevious() {
	// Remember currently selected task
	var selectedTaskID string
	sortedTasks := m.GetSortedTasks()
	if len(sortedTasks) > 0 && m.Navigation.selectedIndex < len(sortedTasks) {
		selectedTaskID = sortedTasks[m.Navigation.selectedIndex].ID
	}

	// Change sort mode
	m.Data.sortMode = (m.Data.sortMode + 3) % 4 // +3 for wrap-around

	// Find the same task in new sort order and select it
	m.findAndSelectTask(selectedTaskID)

	// Update search matches for new sort order
	m.updateSearchMatches()
}

// SetError sets the error state

// HandleGenericScroll handles scrolling for different UI contexts
// NOTE: Help modal and task details now use viewport, only list scrolling remains
func (m *Model) HandleGenericScroll(context ScrollContext, direction int, scrollType string) {
	switch context {
	case ListContext:
		m.handleListContextScroll(direction, scrollType)
		// HelpContext and DetailsContext are now handled by viewport, no longer needed
	}
}

// handleListContextScroll handles scrolling for task list
func (m *Model) handleListContextScroll(direction int, scrollType string) {
	sortedTasks := m.GetSortedTasks()
	if len(sortedTasks) == 0 {
		return
	}

	switch scrollType {
	case "line":
		newIndex := m.Navigation.selectedIndex + direction
		m.setSelectedTask(newIndex)
	case "fast":
		fastScrollLines := 4
		newIndex := m.Navigation.selectedIndex + (direction * fastScrollLines)
		m.setSelectedTask(newIndex)
	case "halfpage":
		halfPage := m.getTaskListHalfPageSize()
		newIndex := m.Navigation.selectedIndex + (direction * halfPage)
		m.setSelectedTask(newIndex)
	case "top":
		m.setSelectedTask(0)
	case "bottom":
		m.setSelectedTask(len(sortedTasks) - 1)
	}
}

// Note: getTaskListHalfPageSize is defined in input_navigation.go to avoid duplication

// getHalfPageSize calculates half-page size for detail panel scrolling
func (m *Model) getHalfPageSize() int {
	contentHeight := m.GetContentHeight()
	halfPage := (contentHeight - 4) / 2 // Account for border and padding
	if halfPage < 1 {
		halfPage = 1
	}
	return halfPage
}

// validateScrollOffset ensures scroll offset is within valid bounds (0 to maxScroll)
func (m *Model) validateScrollOffset(offset, contentLines, viewportLines int) int {
	if offset < 0 {
		return 0
	}

	// Calculate maximum scroll: total lines - viewport lines
	// If content fits entirely, max scroll is 0
	if contentLines <= viewportLines {
		return 0
	}

	maxScroll := contentLines - viewportLines
	if offset > maxScroll {
		return maxScroll
	}

	return offset
}

// calculateDetailContentLength estimates the content length for task details
// Uses conservative estimates to account for markdown rendering and formatting differences
func (m *Model) calculateDetailContentLength() int {
	sortedTasks := m.GetSortedTasks()
	if len(sortedTasks) == 0 || m.Navigation.selectedIndex >= len(sortedTasks) {
		return 20 // Default minimum for empty content
	}

	task := sortedTasks[m.Navigation.selectedIndex]
	width := m.GetRightPanelWidth() // Use same width as view for word wrapping

	// Conservative estimate with buffers for formatting differences
	var estimatedLines int

	// Fixed elements (headers, metadata, timestamps)
	estimatedLines += 8 // Task Details header, title header, status/assignee, timestamps, spacing

	// Title (word wrapped)
	titleLines := len(strings.Split(m.wordWrap(task.Title, width-6), "\n"))
	estimatedLines += titleLines

	// Feature tag if present
	if task.Feature != nil && *task.Feature != "" {
		estimatedLines += 1
	}

	// Description (with generous buffer for markdown)
	if task.Description != "" {
		descriptionLength := len(task.Description)
		// Rough estimate: chars per line with buffer for markdown formatting
		charsPerLine := width - 10 // Conservative line width
		if charsPerLine < 40 {
			charsPerLine = 40 // Minimum reasonable line width
		}

		basicLines := (descriptionLength / charsPerLine) + 1
		// Add 50% buffer for markdown formatting, code blocks, lists, etc.
		estimatedLines += int(float64(basicLines)*1.5) + 2 // +2 for header and spacing
	}

	// Sources
	if len(task.Sources) > 0 {
		estimatedLines += len(task.Sources) + 2 // +2 for header and spacing
	}

	// Code Examples
	if len(task.CodeExamples) > 0 {
		estimatedLines += len(task.CodeExamples) + 2 // +2 for header and spacing
	}

	// Add conservative buffer for any missing elements
	estimatedLines = int(float64(estimatedLines) * 1.1) // 10% additional buffer

	// Ensure minimum reasonable content length
	if estimatedLines < 10 {
		estimatedLines = 10
	}

	return estimatedLines
}

// wordWrap wraps text to fit within specified width (helper for content calculation)
func (m *Model) wordWrap(text string, width int) string {
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

// debugLogScroll logs scroll operations for debugging overshoot issues with enhanced content info
func (m *Model) debugLogScroll(context, operation string, before, after, maxScroll int, prevented bool) {
	logFile, err := os.OpenFile("/tmp/lazyarchon_scroll.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return // Fail silently if can't log
	}
	defer logFile.Close()

	timestamp := time.Now().Format("15:04:05.000")
	preventedStr := ""
	if prevented {
		preventedStr = " [PREVENTED]"
	}

	// Add enhanced content length information for debugging mismatches
	var contentInfo string
	if context == "DETAILS" {
		contentLength := m.calculateDetailContentLength()
		viewportLines := m.GetContentHeight() - 4
		actualMaxScroll := contentLength - viewportLines
		if actualMaxScroll < 0 {
			actualMaxScroll = 0
		}

		// Flag potential content length mismatches
		mismatchFlag := ""
		if actualMaxScroll != maxScroll {
			mismatchFlag = " [CONTENT_MISMATCH]"
		}

		contentInfo = fmt.Sprintf(" | content:%d viewport:%d actualMax:%d%s",
			contentLength, viewportLines, actualMaxScroll, mismatchFlag)
	}

	logMsg := fmt.Sprintf("[%s] %s %s: %d → %d (max: %d)%s%s\n",
		timestamp, context, operation, before, after, maxScroll, preventedStr, contentInfo)
	logFile.WriteString(logMsg)
}

// handleTaskIDCopy copies the current task ID to clipboard
func (m Model) handleTaskIDCopy() (Model, tea.Cmd) {
	sortedTasks := m.GetSortedTasks()
	if m.Navigation.selectedIndex >= len(sortedTasks) || len(sortedTasks) == 0 {
		return m.setTemporaryStatusMessage("No task selected"), nil
	}

	selectedTask := sortedTasks[m.Navigation.selectedIndex]
	err := clipboard.WriteAll(selectedTask.ID)
	if err != nil {
		return m.setTemporaryStatusMessage("Failed to copy task ID"), nil
	}

	return m.setTemporaryStatusMessage("Copied task ID to clipboard"), nil
}

// handleTaskTitleCopy copies the current task title to clipboard
func (m Model) handleTaskTitleCopy() (Model, tea.Cmd) {
	sortedTasks := m.GetSortedTasks()
	if m.Navigation.selectedIndex >= len(sortedTasks) || len(sortedTasks) == 0 {
		return m.setTemporaryStatusMessage("No task selected"), nil
	}

	selectedTask := sortedTasks[m.Navigation.selectedIndex]
	err := clipboard.WriteAll(selectedTask.Title)
	if err != nil {
		return m.setTemporaryStatusMessage("Failed to copy task title"), nil
	}

	return m.setTemporaryStatusMessage("Copied task title to clipboard"), nil
}
