package ui

import (
	"fmt"
	"sort"
	"strings"
	"github.com/yousfisaad/lazyarchon/internal/archon"
)

// GetUniqueFeatures returns a sorted list of unique features from current tasks
// Only considers tasks that match the current project filter (if any)
func (m Model) GetUniqueFeatures() []string {
	featureSet := make(map[string]bool)

	// Get tasks after applying project filter but before feature filter
	tasksToAnalyze := m.Data.tasks
	if m.Data.selectedProjectID != nil {
		tasksToAnalyze = []archon.Task{}
		for _, task := range m.Data.tasks {
			if task.ProjectID == *m.Data.selectedProjectID {
				tasksToAnalyze = append(tasksToAnalyze, task)
			}
		}
	}

	// Collect unique features
	for _, task := range tasksToAnalyze {
		if task.Feature != nil && *task.Feature != "" {
			featureSet[*task.Feature] = true
		}
	}

	// Convert to sorted slice
	features := make([]string, 0, len(featureSet))
	for feature := range featureSet {
		features = append(features, feature)
	}
	sort.Strings(features)

	return features
}

// InitFeatureModal initializes the feature modal with all features enabled
func (m *Model) InitFeatureModal() {
	availableFeatures := m.GetUniqueFeatures()
	m.Modals.featureMode.selectedFeatures = make(map[string]bool, len(availableFeatures))

	// Enable all features by default
	for _, feature := range availableFeatures {
		m.Modals.featureMode.selectedFeatures[feature] = true
	}

	// Reset selection index
	m.Modals.featureMode.selectedIndex = 0
}

// GetFeatureFilterSummary returns a summary of active feature filters
func (m Model) GetFeatureFilterSummary() string {
	availableFeatures := m.GetUniqueFeatures()
	if len(availableFeatures) == 0 {
		return "No features"
	}

	enabledCount := 0
	var enabledFeatures []string

	for _, feature := range availableFeatures {
		if enabled, exists := m.Modals.featureMode.selectedFeatures[feature]; exists && enabled {
			enabledCount++
			enabledFeatures = append(enabledFeatures, feature)
		}
	}

	// If no features are explicitly enabled, assume all are enabled (default state)
	if len(m.Modals.featureMode.selectedFeatures) == 0 {
		return "All features"
	}

	totalFeatures := len(availableFeatures)

	if enabledCount == 0 {
		return "No features"
	} else if enabledCount == totalFeatures {
		return "All features"
	} else if enabledCount == 1 {
		return fmt.Sprintf("#%s only", enabledFeatures[0])
	} else {
		return fmt.Sprintf("%d/%d features", enabledCount, totalFeatures)
	}
}

// GetFeatureTaskCount returns the count of tasks matching current filters
func (m Model) GetFeatureTaskCount(feature string) int {
	count := 0

	// Get tasks after applying project filter but before feature filter
	tasksToAnalyze := m.Data.tasks
	if m.Data.selectedProjectID != nil {
		tasksToAnalyze = []archon.Task{}
		for _, task := range m.Data.tasks {
			if task.ProjectID == *m.Data.selectedProjectID {
				tasksToAnalyze = append(tasksToAnalyze, task)
			}
		}
	}

	// Count tasks with this feature
	for _, task := range tasksToAnalyze {
		if task.Feature != nil && *task.Feature == feature {
			count++
		}
	}

	return count
}

// GetFilteredFeatures returns features matching the current search query
func (m Model) GetFilteredFeatures() []string {
	allFeatures := m.GetUniqueFeatures()

	// If no search is active, return all features
	if !m.Modals.featureMode.searchMode && m.Modals.featureMode.searchQuery == "" {
		return allFeatures
	}

	// Use search input for real-time filtering, or committed query
	searchQuery := m.Modals.featureMode.searchInput
	if !m.Modals.featureMode.searchMode {
		searchQuery = m.Modals.featureMode.searchQuery
	}

	if searchQuery == "" {
		return allFeatures
	}

	// Filter features based on search query (case-insensitive)
	var filteredFeatures []string
	searchQuery = strings.ToLower(strings.TrimSpace(searchQuery))

	for _, feature := range allFeatures {
		if strings.Contains(strings.ToLower(feature), searchQuery) {
			filteredFeatures = append(filteredFeatures, feature)
		}
	}

	return filteredFeatures
}