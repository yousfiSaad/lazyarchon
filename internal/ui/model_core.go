package ui

import (
	"fmt"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/yousfisaad/lazyarchon/internal/archon"
	"sort"
)

// NewModel creates a new application model
func NewModel() Model {
	// Connect to local Archon server
	client := archon.NewClient("http://localhost:8181", "")

	// Create viewport for task details with reasonable defaults
	// Will be resized when window size is available
	vp := viewport.New(80, 20)
	vp.SetContent("") // Empty content initially

	// Create viewport for help modal with reasonable defaults
	// Will be resized when modal is opened
	helpVp := viewport.New(60, 15)
	helpVp.SetContent("") // Empty content initially

	return Model{
		client: client,
		Window: WindowState{
			activeView: LeftPanel, // Default to task list panel active
		},
		Navigation: NavigationState{
			selectedIndex: 0,
		},
		Data: DataState{
			loading:  true,
			sortMode: SortStatusPriority, // Default to status+priority sorting
		},
		Modals: ModalState{
			featureMode: FeatureModeState{
				selectedFeatures: make(map[string]bool), // Initialize empty feature selection
			},
		},
		taskDetailsViewport: vp,
		helpModalViewport:   helpVp,
	}
}

// GetSortedTasks returns the tasks sorted according to the current sort mode
// This method applies both project and feature filtering before sorting
func (m Model) GetSortedTasks() []archon.Task {
	filteredTasks := m.Data.tasks

	// Apply project filter first (if any)
	if m.Data.selectedProjectID != nil {
		var projectFilteredTasks []archon.Task
		for _, task := range filteredTasks {
			if task.ProjectID == *m.Data.selectedProjectID {
				projectFilteredTasks = append(projectFilteredTasks, task)
			}
		}
		filteredTasks = projectFilteredTasks
	}

	// Apply feature filter (if any features are explicitly disabled)
	if len(m.Modals.featureMode.selectedFeatures) > 0 {
		var featureFilteredTasks []archon.Task
		for _, task := range filteredTasks {
			// Include task if:
			// 1. It has no feature (null/empty), OR
			// 2. Its feature is enabled in selectedFeatures
			if task.Feature == nil || *task.Feature == "" {
				// Tasks without features are always shown
				featureFilteredTasks = append(featureFilteredTasks, task)
			} else if enabled, exists := m.Modals.featureMode.selectedFeatures[*task.Feature]; exists && enabled {
				featureFilteredTasks = append(featureFilteredTasks, task)
			}
		}
		filteredTasks = featureFilteredTasks
	}

	return SortTasks(filteredTasks, m.Data.sortMode)
}

// IsProjectSelected returns true if a specific project is currently selected
func (m Model) IsProjectSelected() bool {
	return m.Data.selectedProjectID != nil
}

// GetSelectedProject returns the currently selected project, if any
func (m Model) GetSelectedProject() *archon.Project {
	if !m.IsProjectSelected() {
		return nil
	}

	for _, project := range m.Data.projects {
		if project.ID == *m.Data.selectedProjectID {
			return &project
		}
	}
	return nil
}

// GetCurrentProjectName returns the name of the current project or "All Tasks"
func (m Model) GetCurrentProjectName() string {
	if selectedProject := m.GetSelectedProject(); selectedProject != nil {
		return selectedProject.Title
	}
	return "All Tasks"
}

// GetContentHeight returns the available height for content panels
func (m Model) GetContentHeight() int {
	return m.Window.height - HeaderHeight - StatusBarHeight - 1 // -1 for spacing
}

// GetLeftPanelWidth returns the width of the left panel
func (m Model) GetLeftPanelWidth() int {
	return m.Window.width / 2
}

// GetRightPanelWidth returns the width of the right panel
func (m Model) GetRightPanelWidth() int {
	return m.Window.width - m.GetLeftPanelWidth()
}

// GetTaskStatusCounts returns counts of tasks by status
func (m Model) GetTaskStatusCounts() (int, int, int, int) {
	var todo, doing, review, done int

	for _, task := range m.Data.tasks {
		switch task.Status {
		case "todo":
			todo++
		case "doing":
			doing++
		case "review":
			review++
		case "done":
			done++
		}
	}

	return todo, doing, review, done
}

// GetCurrentPosition returns position info for the current selection
func (m Model) GetCurrentPosition() string {
	sortedTasks := m.GetSortedTasks()
	if len(sortedTasks) == 0 {
		return "No tasks"
	}

	if m.Navigation.selectedIndex >= len(sortedTasks) {
		return "No selection"
	}

	return fmt.Sprintf("Task %d of %d", m.Navigation.selectedIndex+1, len(sortedTasks))
}

// GetCurrentSortModeName returns the human-readable name of the current sort mode
func (m Model) GetCurrentSortModeName() string {
	switch m.Data.sortMode {
	case SortStatusPriority:
		return "Status"
	case SortPriorityOnly:
		return "Priority"
	case SortTimeCreated:
		return "Created"
	case SortAlphabetical:
		return "Alpha"
	default:
		return "Unknown"
	}
}

// GetScrollPosition returns scroll position as percentage for right panel
func (m Model) GetScrollPosition() string {
	// Use viewport scroll information
	if m.taskDetailsViewport.AtTop() {
		return "Top"
	} else if m.taskDetailsViewport.AtBottom() {
		return "Bottom"
	}
	return "Scrolled"
}

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
