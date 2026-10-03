package statusfilter

// ShowStatusFilterModalMsg is sent to show the status filter modal
type ShowStatusFilterModalMsg struct {
	CurrentStatuses map[string]bool // Current status filter state
}

// HideStatusFilterModalMsg is sent to hide the status filter modal
type HideStatusFilterModalMsg struct{}

// StatusFilterModalShownMsg is sent when the status filter modal is shown
type StatusFilterModalShownMsg struct{}

// StatusFilterModalHiddenMsg is sent when the status filter modal is hidden
type StatusFilterModalHiddenMsg struct{}

// StatusFilterAppliedMsg is sent when the user applies the status filter selection
type StatusFilterAppliedMsg struct {
	SelectedStatuses map[string]bool // The selected status filters
}
