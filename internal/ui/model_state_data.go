package ui

import (
	"strings"
	"time"

	"github.com/yousfisaad/lazyarchon/internal/archon"
)

// SetError sets the error state and clears loading
func (m *Model) SetError(err string) {
	m.Data.loading = false
	m.Data.loadingMessage = ""
	m.Data.error = err
	m.Data.lastRetryError = err // Store for retry functionality
}

// ClearError clears the error state
func (m *Model) ClearError() {
	m.Data.error = ""
	m.Data.lastRetryError = ""
}

// SetLoading sets the loading state with optional context message
func (m *Model) SetLoading(loading bool) {
	m.Data.loading = loading
	if loading {
		m.ClearError()
	} else {
		m.Data.loadingMessage = ""
	}
}

// SetLoadingWithMessage sets loading state with specific context message
func (m *Model) SetLoadingWithMessage(loading bool, message string) {
	m.Data.loading = loading
	m.Data.loadingMessage = message
	if loading {
		m.ClearError()
	} else {
		m.Data.loadingMessage = ""
	}
}

// GetLoadingSpinner returns the current spinner character
func (m *Model) GetLoadingSpinner() string {
	spinnerChars := []string{"|", "/", "-", "\\"}
	return spinnerChars[m.Data.spinnerIndex%len(spinnerChars)]
}

// AdvanceSpinner advances the spinner animation to the next frame
func (m *Model) AdvanceSpinner() {
	m.Data.spinnerIndex++
}

// FormatUserFriendlyError converts technical errors to user-friendly messages
func (m *Model) FormatUserFriendlyError(err string) string {
	if strings.Contains(err, "connection refused") || strings.Contains(err, "no such host") {
		m.Data.connected = false
		return "Unable to connect to Archon server. Check if it's running on localhost:8181"
	}
	if strings.Contains(err, "timeout") {
		m.Data.connected = false
		return "Connection timeout. The server may be slow or unreachable"
	}
	if strings.Contains(err, "status 401") || strings.Contains(err, "status 403") {
		return "Authentication failed. Check your API key configuration"
	}
	if strings.Contains(err, "status 404") {
		return "Resource not found. The task or project may have been deleted"
	}
	if strings.Contains(err, "status 500") {
		return "Server error. Please try again or contact support"
	}

	// Return the original error if no specific pattern matches
	return err
}

// SetConnectionStatus sets the connection status
func (m *Model) SetConnectionStatus(connected bool) {
	m.Data.connected = connected
}

// GetConnectionStatusText returns a text indicator for connection status
func (m *Model) GetConnectionStatusText() string {
	if m.Data.connected {
		return "●" // Connected
	}
	return "○" // Disconnected
}

// UpdateTasks updates the task list and adjusts selection bounds
func (m *Model) UpdateTasks(tasks []archon.Task) {
	m.Data.loading = false
	m.Data.tasks = tasks
	m.Data.connected = true // Mark as connected on successful data load
	m.ClearError()

	// Apply sorting and adjust selectedIndex if needed
	sortedTasks := m.GetSortedTasks()
	if len(sortedTasks) > 0 && m.Navigation.selectedIndex >= len(sortedTasks) {
		m.setSelectedTask(len(sortedTasks) - 1)
	} else if len(sortedTasks) == 0 {
		m.setSelectedTask(0)
	} else {
		// Tasks updated but selection still valid - update viewport content
		m.updateTaskDetailsViewport()
	}

	// Update search matches after task data changes
	m.updateSearchMatches()
}

// UpdateProjects updates the project list and validates current selection
func (m *Model) UpdateProjects(projects []archon.Project) {
	m.Data.projects = projects
	m.Data.connected = true // Mark as connected on successful data load

	// Reset project selection if selected project no longer exists
	if m.Data.selectedProjectID != nil {
		projectExists := false
		for _, project := range m.Data.projects {
			if project.ID == *m.Data.selectedProjectID {
				projectExists = true
				break
			}
		}
		if !projectExists {
			m.Data.selectedProjectID = nil // Reset to "All tasks"
		}
	}
}

// setTemporaryStatusMessage sets a temporary status message with timestamp
func (m Model) setTemporaryStatusMessage(message string) Model {
	// Set status message (this integrates with existing status display system)
	m.Data.statusMessage = message
	m.Data.statusMessageTime = time.Now()
	return m
}