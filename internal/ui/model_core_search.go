package ui

import (
	"strings"
)

// updateSearchMatches updates the search matches for task search
func (m *Model) updateSearchMatches() {
	// Clear previous matches
	m.Data.matchingTaskIndices = nil
	m.Data.currentMatchIndex = 0
	m.Data.totalMatches = 0

	if !m.Data.searchActive || m.Data.searchQuery == "" {
		return
	}

	sortedTasks := m.GetSortedTasks()
	searchQuery := strings.ToLower(strings.TrimSpace(m.Data.searchQuery))

	// Find all tasks that match the search query (title only)
	for i, task := range sortedTasks {
		titleMatch := strings.Contains(strings.ToLower(task.Title), searchQuery)

		if titleMatch {
			m.Data.matchingTaskIndices = append(m.Data.matchingTaskIndices, i)
		}
	}

	m.Data.totalMatches = len(m.Data.matchingTaskIndices)

	// Update current match index based on current task selection
	if m.Data.totalMatches > 0 {
		// Find current selection in match list
		for i, matchIndex := range m.Data.matchingTaskIndices {
			if matchIndex == m.Navigation.selectedIndex {
				m.Data.currentMatchIndex = i
				return
			}
		}
		// If current selection is not a match, reset to first match
		m.Data.currentMatchIndex = 0
	}
}

// nextSearchMatch navigates to the next search match (n command)
func (m *Model) nextSearchMatch() {
	if m.Data.totalMatches == 0 {
		return
	}

	currentIndex := m.Navigation.selectedIndex

	// Find next match after current position
	for _, matchIndex := range m.Data.matchingTaskIndices {
		if matchIndex > currentIndex {
			m.setSelectedTask(matchIndex)
			return
		}
	}

	// No match found after current position, wrap to first match
	m.setSelectedTask(m.Data.matchingTaskIndices[0])
}

// previousSearchMatch navigates to the previous search match (N command)
func (m *Model) previousSearchMatch() {
	if m.Data.totalMatches == 0 {
		return
	}

	currentIndex := m.Navigation.selectedIndex

	// Find previous match before current position (reverse search)
	for i := len(m.Data.matchingTaskIndices) - 1; i >= 0; i-- {
		matchIndex := m.Data.matchingTaskIndices[i]
		if matchIndex < currentIndex {
			m.setSelectedTask(matchIndex)
			return
		}
	}

	// No match found before current position, wrap to last match
	lastIndex := len(m.Data.matchingTaskIndices) - 1
	m.setSelectedTask(m.Data.matchingTaskIndices[lastIndex])
}

// updateFeatureSearchMatches updates the filtered features list and match indices
func (m *Model) updateFeatureSearchMatches() {
	// Update filtered features cache
	m.Modals.featureMode.filteredFeatures = m.GetFilteredFeatures()

	// Clear match tracking
	m.Modals.featureMode.matchingIndices = nil
	m.Modals.featureMode.currentMatchIndex = 0

	// If no search is active, all features are "matches"
	if !m.Modals.featureMode.searchMode && m.Modals.featureMode.searchQuery == "" {
		for i := range m.Modals.featureMode.filteredFeatures {
			m.Modals.featureMode.matchingIndices = append(m.Modals.featureMode.matchingIndices, i)
		}
		return
	}

	// Get search query
	searchQuery := m.Modals.featureMode.searchInput
	if !m.Modals.featureMode.searchMode {
		searchQuery = m.Modals.featureMode.searchQuery
	}

	if searchQuery == "" {
		for i := range m.Modals.featureMode.filteredFeatures {
			m.Modals.featureMode.matchingIndices = append(m.Modals.featureMode.matchingIndices, i)
		}
		return
	}

	// All filtered features are matches by definition
	for i := range m.Modals.featureMode.filteredFeatures {
		m.Modals.featureMode.matchingIndices = append(m.Modals.featureMode.matchingIndices, i)
	}

	// Update current match index based on current selection
	if len(m.Modals.featureMode.matchingIndices) > 0 {
		// Find current selection in match list
		for i, matchIndex := range m.Modals.featureMode.matchingIndices {
			if matchIndex == m.Modals.featureMode.selectedIndex {
				m.Modals.featureMode.currentMatchIndex = i
				return
			}
		}
		// If current selection is not a match, reset to first match
		m.Modals.featureMode.currentMatchIndex = 0
	}
}

// activateFeatureSearch enters search mode in the feature modal
func (m *Model) activateFeatureSearch() {
	m.Modals.featureMode.searchMode = true
	m.Modals.featureMode.searchInput = m.Modals.featureMode.searchQuery // Start with current search
}

// cancelFeatureSearch exits search mode without applying changes
func (m *Model) cancelFeatureSearch() {
	m.Modals.featureMode.searchMode = false
	m.Modals.featureMode.searchInput = ""
	// Restore to previous search state
	m.updateFeatureSearchMatches()
}

// commitFeatureSearch applies the current search input and exits search mode
func (m *Model) commitFeatureSearch() {
	m.setFeatureSearchQuery(m.Modals.featureMode.searchInput)
	m.Modals.featureMode.searchMode = false
	m.Modals.featureMode.searchInput = ""
}

// setFeatureSearchQuery sets the current feature search query and updates search state
func (m *Model) setFeatureSearchQuery(query string) {
	// Trim whitespace
	query = strings.TrimSpace(query)

	// Update search state
	m.Modals.featureMode.searchQuery = query

	// Update search matches
	m.updateFeatureSearchMatches()

	// Reset selection to first match if available
	if len(m.Modals.featureMode.filteredFeatures) > 0 {
		m.Modals.featureMode.selectedIndex = 0
	}
}

// clearFeatureSearch clears the current feature search query
func (m *Model) clearFeatureSearch() {
	m.Modals.featureMode.searchQuery = ""
	m.updateFeatureSearchMatches()
	// Reset selection to first feature
	if len(m.Modals.featureMode.filteredFeatures) > 0 {
		m.Modals.featureMode.selectedIndex = 0
	}
}

// nextFeatureMatch navigates to the next search match in the feature modal
func (m *Model) nextFeatureMatch() {
	if len(m.Modals.featureMode.matchingIndices) == 0 {
		return
	}

	currentIndex := m.Modals.featureMode.selectedIndex

	// Find next match after current position
	for _, matchIndex := range m.Modals.featureMode.matchingIndices {
		if matchIndex > currentIndex {
			m.Modals.featureMode.selectedIndex = matchIndex
			return
		}
	}

	// No match found after current position, wrap to first match
	m.Modals.featureMode.selectedIndex = m.Modals.featureMode.matchingIndices[0]
}

// previousFeatureMatch navigates to the previous search match in the feature modal
func (m *Model) previousFeatureMatch() {
	if len(m.Modals.featureMode.matchingIndices) == 0 {
		return
	}

	currentIndex := m.Modals.featureMode.selectedIndex

	// Find previous match before current position (reverse search)
	for i := len(m.Modals.featureMode.matchingIndices) - 1; i >= 0; i-- {
		matchIndex := m.Modals.featureMode.matchingIndices[i]
		if matchIndex < currentIndex {
			m.Modals.featureMode.selectedIndex = matchIndex
			return
		}
	}

	// No match found before current position, wrap to last match
	lastIndex := len(m.Modals.featureMode.matchingIndices) - 1
	m.Modals.featureMode.selectedIndex = m.Modals.featureMode.matchingIndices[lastIndex]
}