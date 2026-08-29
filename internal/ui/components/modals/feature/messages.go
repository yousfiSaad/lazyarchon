package feature

import tea "github.com/charmbracelet/bubbletea"

// ShowFeatureModalMsg is sent to show the feature selection modal
type ShowFeatureModalMsg struct {
	AllFeatures          []string        // All available features
	SelectedFeatures     map[string]bool // Currently selected features
	FeatureColorsEnabled bool            // Whether to show feature colors
}

// HideFeatureModalMsg is sent to hide the feature selection modal
type HideFeatureModalMsg struct{}

// FeatureModalShownMsg is sent when the feature modal has been shown
type FeatureModalShownMsg struct{}

// FeatureModalHiddenMsg is sent when the feature modal has been hidden
type FeatureModalHiddenMsg struct{}

// FeatureSelectionAppliedMsg is sent when feature selection is applied
type FeatureSelectionAppliedMsg struct {
	SelectedFeatures map[string]bool // Final selected features
}

// FeatureModalSearchMsg is sent when search query changes
type FeatureModalSearchMsg struct {
	Query string // Search query
}

// FeatureModalScrollMsg is sent for scrolling within the modal
type FeatureModalScrollMsg struct {
	Direction int // -1 for up, 1 for down
}

// FeatureModalToggleMsg is sent to toggle a specific feature
type FeatureModalToggleMsg struct {
	Feature string // Feature to toggle
}

// FeatureModalClearSearchMsg is sent to clear the search
type FeatureModalClearSearchMsg struct{}

// FeatureModalSelectAllMsg is sent to select all visible features
type FeatureModalSelectAllMsg struct{}

// FeatureModalDeselectAllMsg is sent to deselect all features
type FeatureModalDeselectAllMsg struct{}

// Ensure all message types implement tea.Msg
var (
	_ tea.Msg = ShowFeatureModalMsg{}
	_ tea.Msg = HideFeatureModalMsg{}
	_ tea.Msg = FeatureModalShownMsg{}
	_ tea.Msg = FeatureModalHiddenMsg{}
	_ tea.Msg = FeatureSelectionAppliedMsg{}
	_ tea.Msg = FeatureModalSearchMsg{}
	_ tea.Msg = FeatureModalScrollMsg{}
	_ tea.Msg = FeatureModalToggleMsg{}
	_ tea.Msg = FeatureModalClearSearchMsg{}
	_ tea.Msg = FeatureModalSelectAllMsg{}
	_ tea.Msg = FeatureModalDeselectAllMsg{}
)
