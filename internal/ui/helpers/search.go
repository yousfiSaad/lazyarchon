package helpers

import (
	"strings"

	"github.com/yousfisaad/lazyarchon/v2/internal/shared/interfaces"
)

// TaskMatchesQuery reports whether a task matches the search query.
// Matches title, status, any tag, or ID (case-insensitive substring).
// Shared by search navigation (SearchTasks) and the task list's highlight
// rendering so navigation, highlighting, and match counts always agree.
func TaskMatchesQuery(task interfaces.Task, query string) bool {
	if query == "" {
		return false
	}

	query = strings.ToLower(strings.TrimSpace(query))

	// Check if any tag matches
	tagMatch := false
	for _, tag := range task.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			tagMatch = true
			break
		}
	}

	return strings.Contains(strings.ToLower(task.Title), query) ||
		strings.Contains(strings.ToLower(task.Status), query) ||
		tagMatch ||
		strings.Contains(strings.ToLower(task.ID), query)
}

// SearchTasks finds tasks matching the search query
// Returns matching indices and total matches
func SearchTasks(tasks []interfaces.Task, searchQuery string) (matchingIndices []int, totalMatches int) {
	if searchQuery == "" {
		return nil, 0
	}

	// Find all tasks that match the search query (title, status, tags, or ID)
	for i, task := range tasks {
		if TaskMatchesQuery(task, searchQuery) {
			matchingIndices = append(matchingIndices, i)
		}
	}

	totalMatches = len(matchingIndices)
	return matchingIndices, totalMatches
}

// GetNextMatch returns the index of the next search match
func GetNextMatch(matchingIndices []int, currentIndex int) int {
	if len(matchingIndices) == 0 {
		return currentIndex
	}

	// Find current position in match sequence
	for i, idx := range matchingIndices {
		if idx == currentIndex {
			// We're on a match, go to next in sequence
			nextPos := (i + 1) % len(matchingIndices)
			return matchingIndices[nextPos]
		}
	}

	// Not on a match, find first match after current position
	for _, idx := range matchingIndices {
		if idx > currentIndex {
			return idx
		}
	}

	// Wrap to first match (safe: we already checked len > 0 at start)
	if len(matchingIndices) > 0 {
		return matchingIndices[0]
	}
	return currentIndex
}

// GetPreviousMatch returns the index of the previous search match
func GetPreviousMatch(matchingIndices []int, currentIndex int) int {
	if len(matchingIndices) == 0 {
		return currentIndex
	}

	// Find current position in match sequence
	for i, idx := range matchingIndices {
		if idx == currentIndex {
			// We're on a match, go to previous in sequence
			prevPos := i - 1
			if prevPos < 0 {
				prevPos = len(matchingIndices) - 1
			}
			return matchingIndices[prevPos]
		}
	}

	// Not on a match, find first match before current position
	for i := len(matchingIndices) - 1; i >= 0; i-- {
		if matchingIndices[i] < currentIndex {
			return matchingIndices[i]
		}
	}

	// Wrap to last match (safe: we already checked len > 0 at start)
	if len(matchingIndices) > 0 {
		return matchingIndices[len(matchingIndices)-1]
	}
	return currentIndex
}

// FindMatchPosition returns the position (0-based) of selectedIndex within matchingIndices
// Returns -1 if selectedIndex is not in the matches
func FindMatchPosition(matchingIndices []int, selectedIndex int) int {
	for i, matchIdx := range matchingIndices {
		if matchIdx == selectedIndex {
			return i
		}
	}
	return -1 // Not in current matches
}
