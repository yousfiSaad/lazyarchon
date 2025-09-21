package ui

// ModelFeatureHelpers implements the modals.FeatureHelpers interface
type ModelFeatureHelpers struct {
	model *Model
}

// NewModelFeatureHelpers creates a new ModelFeatureHelpers adapter
func (m Model) NewModelFeatureHelpers() *ModelFeatureHelpers {
	return &ModelFeatureHelpers{model: &m}
}

// GetFeatureTaskCount returns the number of tasks for a given feature
func (h *ModelFeatureHelpers) GetFeatureTaskCount(feature string) int {
	return h.model.GetFeatureTaskCount(feature)
}

// GetFeatureColor returns the color for a given feature
func (h *ModelFeatureHelpers) GetFeatureColor(feature string) string {
	return GetFeatureColor(feature)
}

// GetMutedFeatureColor returns the muted color for a given feature
func (h *ModelFeatureHelpers) GetMutedFeatureColor(feature string) string {
	return GetMutedFeatureColor(feature)
}

// HighlightSearchTermsWithColor highlights search terms with color
func (h *ModelFeatureHelpers) HighlightSearchTermsWithColor(text, query, textColor string) string {
	return highlightSearchTermsWithColor(text, query, textColor)
}

// ModelTaskEditHelpers implements the modals.TaskEditHelpers interface
type ModelTaskEditHelpers struct {
	model *Model
}

// NewModelTaskEditHelpers creates a new ModelTaskEditHelpers adapter
func (m Model) NewModelTaskEditHelpers() *ModelTaskEditHelpers {
	return &ModelTaskEditHelpers{model: &m}
}

// GetUniqueFeatures returns unique features from tasks
func (h *ModelTaskEditHelpers) GetUniqueFeatures() []string {
	return h.model.GetUniqueFeatures()
}

// GetFeatureTaskCount returns the number of tasks for a given feature
func (h *ModelTaskEditHelpers) GetFeatureTaskCount(feature string) int {
	return h.model.GetFeatureTaskCount(feature)
}