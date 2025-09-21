package ui

// ToggleFeature toggles a feature on/off in the selection
func (m *Model) ToggleFeature(featureName string) {
	if m.Modals.featureMode.selectedFeatures == nil {
		m.Modals.featureMode.selectedFeatures = make(map[string]bool)
	}

	// Toggle the feature
	currentState, exists := m.Modals.featureMode.selectedFeatures[featureName]
	if !exists {
		// If feature doesn't exist in map, it was implicitly enabled, so disable it
		m.Modals.featureMode.selectedFeatures[featureName] = false
	} else {
		m.Modals.featureMode.selectedFeatures[featureName] = !currentState
	}

	// Reset task selection since filtering changed
	m.setSelectedTask(0)
}

// SelectAllFeatures enables all available features
func (m *Model) SelectAllFeatures() {
	availableFeatures := m.GetUniqueFeatures()
	if m.Modals.featureMode.selectedFeatures == nil {
		m.Modals.featureMode.selectedFeatures = make(map[string]bool)
	}

	for _, feature := range availableFeatures {
		m.Modals.featureMode.selectedFeatures[feature] = true
	}

	// Reset task selection since filtering changed
	m.setSelectedTask(0)
}

// SelectNoFeatures disables all features
func (m *Model) SelectNoFeatures() {
	availableFeatures := m.GetUniqueFeatures()
	if m.Modals.featureMode.selectedFeatures == nil {
		m.Modals.featureMode.selectedFeatures = make(map[string]bool)
	}

	for _, feature := range availableFeatures {
		m.Modals.featureMode.selectedFeatures[feature] = false
	}

	// Reset task selection since filtering changed
	m.setSelectedTask(0)
}

// SmartToggleAllFeatures toggles between "select all" and "select none" intelligently
// When search is active, it only operates on filtered features
func (m *Model) SmartToggleAllFeatures() {
	// Check if search is active - if so, only operate on filtered features
	if m.Modals.featureMode.searchQuery != "" {
		m.smartToggleFilteredFeatures()
		return
	}

	// Default behavior: operate on all features
	availableFeatures := m.GetUniqueFeatures()
	if m.Modals.featureMode.selectedFeatures == nil {
		m.Modals.featureMode.selectedFeatures = make(map[string]bool)
	}

	// Check if all features are currently selected
	allSelected := true
	for _, feature := range availableFeatures {
		if enabled, exists := m.Modals.featureMode.selectedFeatures[feature]; !exists || !enabled {
			allSelected = false
			break
		}
	}

	if allSelected {
		// All are selected, so unselect all
		m.SelectNoFeatures()
	} else {
		// Some or none are selected, so select all
		m.SelectAllFeatures()
	}
}

// smartToggleFilteredFeatures toggles between "select all" and "select none" for filtered features only
func (m *Model) smartToggleFilteredFeatures() {
	filteredFeatures := m.GetFilteredFeatures()
	if m.Modals.featureMode.selectedFeatures == nil {
		m.Modals.featureMode.selectedFeatures = make(map[string]bool)
	}

	// Check if all filtered features are currently selected
	allFilteredSelected := true
	for _, feature := range filteredFeatures {
		if enabled, exists := m.Modals.featureMode.selectedFeatures[feature]; !exists || !enabled {
			allFilteredSelected = false
			break
		}
	}

	if allFilteredSelected {
		// All filtered features are selected, so unselect them
		m.selectNoFilteredFeatures()
	} else {
		// Some or none of the filtered features are selected, so select all of them
		m.selectAllFilteredFeatures()
	}
}

// selectAllFilteredFeatures enables only the currently filtered features
func (m *Model) selectAllFilteredFeatures() {
	filteredFeatures := m.GetFilteredFeatures()
	if m.Modals.featureMode.selectedFeatures == nil {
		m.Modals.featureMode.selectedFeatures = make(map[string]bool)
	}

	for _, feature := range filteredFeatures {
		m.Modals.featureMode.selectedFeatures[feature] = true
	}

	// Reset task selection since filtering changed
	m.setSelectedTask(0)
}

// selectNoFilteredFeatures disables only the currently filtered features
func (m *Model) selectNoFilteredFeatures() {
	filteredFeatures := m.GetFilteredFeatures()
	if m.Modals.featureMode.selectedFeatures == nil {
		m.Modals.featureMode.selectedFeatures = make(map[string]bool)
	}

	for _, feature := range filteredFeatures {
		m.Modals.featureMode.selectedFeatures[feature] = false
	}

	// Reset task selection since filtering changed
	m.setSelectedTask(0)
}

// backupFeatureState saves the current feature selection state for cancel functionality
func (m *Model) backupFeatureState() {
	if m.Modals.featureMode.selectedFeatures == nil {
		m.Modals.featureMode.backupFeatures = nil
		return
	}

	// Deep copy the current state
	m.Modals.featureMode.backupFeatures = make(map[string]bool, len(m.Modals.featureMode.selectedFeatures))
	for feature, enabled := range m.Modals.featureMode.selectedFeatures {
		m.Modals.featureMode.backupFeatures[feature] = enabled
	}
}

// restoreFeatureState restores the backup feature selection state (for cancel)
func (m *Model) restoreFeatureState() {
	if m.Modals.featureMode.backupFeatures == nil {
		m.Modals.featureMode.selectedFeatures = nil
	} else {
		// Deep copy the backup state back
		m.Modals.featureMode.selectedFeatures = make(map[string]bool, len(m.Modals.featureMode.backupFeatures))
		for feature, enabled := range m.Modals.featureMode.backupFeatures {
			m.Modals.featureMode.selectedFeatures[feature] = enabled
		}
	}

	// Reset task selection since filtering changed
	m.setSelectedTask(0)
}