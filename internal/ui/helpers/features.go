package helpers

import (
	"fmt"
	"sort"

	"github.com/yousfisaad/lazyarchon/internal/archon"
)

// GetUniqueFeatures returns a sorted list of unique features from tasks
func GetUniqueFeatures(tasks []archon.Task) []string {
	featureSet := make(map[string]bool)

	// Collect unique features
	for _, task := range tasks {
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

// GetFeatureTaskCount returns the count of tasks for a specific feature
func GetFeatureTaskCount(tasks []archon.Task, feature string) int {
	count := 0
	for _, task := range tasks {
		if task.Feature != nil && *task.Feature == feature {
			count++
		}
	}
	return count
}

// GetFeatureFilterSummary returns a summary of active feature filters
// Three-state logic:
// - nil map: No filter active (show all)
// - empty map {}: Filter active, nothing selected (show none)
// - populated map: Filter active with selections
func GetFeatureFilterSummary(availableFeatures []string, enabledFeatures map[string]bool) string {
	if len(availableFeatures) == 0 {
		return "No features"
	}

	// nil = no filter active, show all
	if enabledFeatures == nil {
		return "All features"
	}

	enabledCount := 0
	var enabledList []string

	for _, feature := range availableFeatures {
		if enabled, exists := enabledFeatures[feature]; exists && enabled {
			enabledCount++
			enabledList = append(enabledList, feature)
		}
	}

	totalFeatures := len(availableFeatures)

	switch enabledCount {
	case 0:
		return "No features" // Empty map = explicitly deselected all
	case totalFeatures:
		return "All features"
	case 1:
		return fmt.Sprintf("#%s only", enabledList[0])
	default:
		return fmt.Sprintf("%d/%d features", enabledCount, totalFeatures)
	}
}
