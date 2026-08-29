package factories

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/base"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/layout/header"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/layout/maincontent"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/layout/statusbar"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/modals/confirmation"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/modals/feature"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/modals/help"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/modals/status"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/modals/taskcreate"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/modals/taskedit"
)

// ModalComponents contains all modal components
type ModalComponents struct {
	HelpModel         *help.HelpModel
	StatusModel       *status.StatusModel
	ConfirmationModel *confirmation.ConfirmationModel
	TaskEditModel     *taskedit.TaskEditModel
	TaskCreateModel   *taskcreate.TaskCreateModel
	FeatureModel      *feature.FeatureModel
}

// Update broadcasts messages to all modal components (hierarchical pattern)
// Commands from all modals are batched together for concurrent execution
func (mc *ModalComponents) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	// Each modal handles the message if relevant, ignores otherwise
	if mc.HelpModel != nil {
		cmds = append(cmds, mc.HelpModel.Update(msg))
	}
	if mc.StatusModel != nil {
		cmds = append(cmds, mc.StatusModel.Update(msg))
	}
	if mc.ConfirmationModel != nil {
		cmds = append(cmds, mc.ConfirmationModel.Update(msg))
	}
	if mc.TaskEditModel != nil {
		cmds = append(cmds, mc.TaskEditModel.Update(msg))
	}
	if mc.TaskCreateModel != nil {
		cmds = append(cmds, mc.TaskCreateModel.Update(msg))
	}
	if mc.FeatureModel != nil {
		cmds = append(cmds, mc.FeatureModel.Update(msg))
	}

	return tea.Batch(cmds...)
}

// HasActiveModal returns true if any modal is currently active
// This method ensures all modals are checked without MainModel needing to know modal details
func (mc *ModalComponents) HasActiveModal() bool {
	if mc.HelpModel != nil && mc.HelpModel.IsActive() {
		return true
	}
	if mc.StatusModel != nil && mc.StatusModel.IsActive() {
		return true
	}
	if mc.ConfirmationModel != nil && mc.ConfirmationModel.IsActive() {
		return true
	}
	if mc.TaskEditModel != nil && mc.TaskEditModel.IsActive() {
		return true
	}
	if mc.TaskCreateModel != nil && mc.TaskCreateModel.IsActive() {
		return true
	}
	if mc.FeatureModel != nil && mc.FeatureModel.IsActive() {
		return true
	}
	return false
}

// LayoutComponents contains all layout components
type LayoutComponents struct {
	Header      *header.HeaderModel
	StatusBar   *statusbar.StatusBarModel
	MainContent *maincontent.MainContentModel
}

// Update broadcasts messages to all layout components (hierarchical pattern)
func (lc *LayoutComponents) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	// Each layout component handles the message if relevant
	if lc.Header != nil {
		cmds = append(cmds, lc.Header.Update(msg))
	}
	if lc.StatusBar != nil {
		cmds = append(cmds, lc.StatusBar.Update(msg))
	}
	if lc.MainContent != nil {
		cmds = append(cmds, lc.MainContent.Update(msg))
	}

	return tea.Batch(cmds...)
}

// UIComponentSet contains all UI components organized by category
type UIComponentSet struct {
	Modals ModalComponents
	Layout LayoutComponents
}

// Update broadcasts messages to all component categories (hierarchical pattern)
func (cs *UIComponentSet) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	// Each component category handles its children
	cmds = append(cmds, cs.Modals.Update(msg))
	cmds = append(cmds, cs.Layout.Update(msg))

	return tea.Batch(cmds...)
}

// ComponentConfig contains configuration for creating UI components
type ComponentConfig struct {
	ComponentContext *base.ComponentContext
}

// CreateComponents creates all UI components with proper initialization
// Note: Some components require additional dependencies that must be passed after creation
// (e.g., header, statusbar, maincontent need coordinators/managers)
func CreateComponents(config ComponentConfig) *UIComponentSet {
	// Create modal components first (they have minimal dependencies)
	helpModal := help.NewModel(config.ComponentContext)
	statusModal := status.NewModel(config.ComponentContext)
	confirmationModal := confirmation.NewModel(config.ComponentContext)
	taskEditModal := taskedit.NewModel(config.ComponentContext)
	taskCreateModal := taskcreate.NewModel(config.ComponentContext)
	featureModal := feature.NewModel(config.ComponentContext)

	return &UIComponentSet{
		Modals: ModalComponents{
			HelpModel:         helpModal,
			StatusModel:       statusModal,
			ConfirmationModel: confirmationModal,
			TaskEditModel:     taskEditModal,
			TaskCreateModel:   taskCreateModal,
			FeatureModel:      featureModal,
		},
		Layout: LayoutComponents{
			// Header, StatusBar, and MainContent are initialized separately
			// after model creation since they need additional dependencies
			Header:      nil,
			StatusBar:   nil,
			MainContent: nil,
		},
	}
}
