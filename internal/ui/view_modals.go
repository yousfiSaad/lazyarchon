package ui

import (
	"github.com/yousfisaad/lazyarchon/internal/ui/view/modals"
	"github.com/yousfisaad/lazyarchon/internal/ui/styling"
)

// renderHelpModal renders the help modal overlay on top of the base UI
func (m Model) renderHelpModal(baseUI string) string {
	// The viewport content is managed by the model and updated when help is opened
	// We just need to render the viewport view
	viewportContent := m.helpModalViewport.View()

	return modals.RenderHelpModal(viewportContent, m.Window.width, m.Window.height)
}

// renderStatusChangeModal renders the status change modal overlay on top of the base UI
func (m Model) renderStatusChangeModal(baseUI string) string {
	// Configure status options
	statusOptions := []modals.StatusOption{
		{"Todo", styling.StatusSymbolTodo, "240"},     // gray
		{"Doing", styling.StatusSymbolDoing, "33"},   // yellow
		{"Review", styling.StatusSymbolReview, "34"}, // blue
		{"Done", styling.StatusSymbolDone, "32"},     // green
	}

	config := modals.StatusChangeConfig{
		SelectedIndex: m.Modals.statusChange.selectedIndex,
		StatusOptions: statusOptions,
	}

	// Create style context for modal styling
	styleContext := m.CreateStyleContext(false)
	factory := styleContext.Factory()

	return modals.RenderStatusChangeModal(config, factory, styling.CurrentTheme.MutedColor, m.Window.width, m.Window.height)
}


// renderConfirmationModal renders the confirmation modal overlay on top of the base UI
func (m Model) renderConfirmationModal(baseUI string) string {
	config := modals.ConfirmationConfig{
		Message:        m.Modals.confirmation.message,
		ConfirmText:    m.Modals.confirmation.confirmText,
		CancelText:     m.Modals.confirmation.cancelText,
		SelectedOption: m.Modals.confirmation.selectedOption,
	}

	// Create style context for modal styling
	styleContext := m.CreateStyleContext(false)
	factory := styleContext.Factory()

	return modals.RenderConfirmationModal(config, factory, styling.CurrentTheme.MutedColor, m.Window.width, m.Window.height)
}


// renderFeatureModal renders the feature selection modal overlay on top of the base UI
func (m Model) renderFeatureModal(baseUI string) string {
	config := modals.FeatureConfig{
		SearchQuery:          m.Modals.featureMode.searchQuery,
		SearchMode:           m.Modals.featureMode.searchMode,
		SearchInput:          m.Modals.featureMode.searchInput,
		SelectedIndex:        m.Modals.featureMode.selectedIndex,
		SelectedFeatures:     m.Modals.featureMode.selectedFeatures,
		FilteredFeatures:     m.GetFilteredFeatures(),
		AllFeatures:          m.GetUniqueFeatures(),
		FeatureColorsEnabled: m.config.IsFeatureColorsEnabled(),
	}

	// Create style context for modal styling
	styleContext := m.CreateStyleContext(false)
	factory := styleContext.Factory()

	// Create feature helpers adapter
	helpers := m.NewModelFeatureHelpers()

	return modals.RenderFeatureModal(config, factory, helpers, styling.CurrentTheme.HeaderColor, styling.CurrentTheme.MutedColor, m.Window.width, m.Window.height)
}


// renderTaskEditModal renders the task edit modal overlay on top of the base UI
func (m Model) renderTaskEditModal(baseUI string) string {
	config := modals.TaskEditConfig{
		IsCreatingNew:     m.Modals.taskEdit.isCreatingNew,
		NewFeatureName:    m.Modals.taskEdit.newFeatureName,
		SelectedIndex:     m.Modals.taskEdit.selectedIndex,
		AvailableFeatures: m.GetUniqueFeatures(),
	}

	// Create style context for modal styling
	styleContext := m.CreateStyleContext(false)
	factory := styleContext.Factory()

	// Create task edit helpers adapter
	helpers := m.NewModelTaskEditHelpers()

	return modals.RenderTaskEditModal(config, factory, helpers, styling.CurrentTheme.HeaderColor, styling.CurrentTheme.MutedColor, m.Window.width, m.Window.height)
}


