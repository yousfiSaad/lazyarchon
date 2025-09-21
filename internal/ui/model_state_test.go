package ui

import (
	"testing"
	"time"

	"github.com/yousfisaad/lazyarchon/internal/archon"
)

// TestModelStateTransitions tests model state management functions
func TestModelStateTransitions(t *testing.T) {
	model := NewModel(createTestConfig())

	// Test initial state
	if model.Data.loading != true {
		t.Error("Expected initial loading state to be true")
	}
	if model.Data.error != "" {
		t.Error("Expected initial error state to be empty")
	}
	if model.Data.connected != false {
		t.Error("Expected initial connected state to be false")
	}
}

// TestSetSelectedProjectAdvanced tests project selection state changes with detailed validation
func TestSetSelectedProjectAdvanced(t *testing.T) {
	model := NewModel(createTestConfig())

	// Setup test data
	model.Data.tasks = []archon.Task{
		{ID: "task1", Title: "Task 1", Status: "todo"},
		{ID: "task2", Title: "Task 2", Status: "doing"},
	}
	model.Navigation.selectedIndex = 1

	// Test setting project ID
	projectID := "test-project-123"
	model.SetSelectedProject(&projectID)

	if model.Data.selectedProjectID == nil {
		t.Fatal("Expected selectedProjectID to be set")
	}
	if *model.Data.selectedProjectID != projectID {
		t.Errorf("Expected selectedProjectID to be %s, got %s", projectID, *model.Data.selectedProjectID)
	}
	if model.Navigation.selectedIndex != 0 {
		t.Error("Expected selectedIndex to reset to 0 when changing projects")
	}
	if model.Modals.featureMode.selectedFeatures != nil && len(model.Modals.featureMode.selectedFeatures) > 0 {
		t.Error("Expected feature filters to be reset when changing projects")
	}

	// Test setting to nil (all tasks)
	model.SetSelectedProject(nil)
	if model.Data.selectedProjectID != nil {
		t.Error("Expected selectedProjectID to be nil")
	}
}

// TestSetError tests error state management
func TestSetError(t *testing.T) {
	model := NewModel(createTestConfig())
	model.Data.loading = true
	model.Data.loadingMessage = "Loading..."

	testError := "Connection failed"
	model.SetError(testError)

	if model.Data.error != testError {
		t.Errorf("Expected error to be %s, got %s", testError, model.Data.error)
	}
	if model.Data.loading != false {
		t.Error("Expected loading to be false after error")
	}
	if model.Data.loadingMessage != "" {
		t.Error("Expected loadingMessage to be cleared after error")
	}
	if model.Data.lastRetryError != testError {
		t.Error("Expected lastRetryError to be set")
	}
}

// TestClearError tests error clearing
func TestClearError(t *testing.T) {
	model := NewModel(createTestConfig())
	model.Data.error = "Test error"
	model.Data.lastRetryError = "Retry error"

	model.ClearError()

	if model.Data.error != "" {
		t.Error("Expected error to be cleared")
	}
	if model.Data.lastRetryError != "" {
		t.Error("Expected lastRetryError to be cleared")
	}
}

// TestSetLoading tests loading state management
func TestSetLoading(t *testing.T) {
	model := NewModel(createTestConfig())
	model.Data.error = "Previous error"

	// Test setting loading to true
	model.SetLoading(true)
	if model.Data.loading != true {
		t.Error("Expected loading to be true")
	}
	if model.Data.error != "" {
		t.Error("Expected error to be cleared when setting loading")
	}

	// Test setting loading to false
	model.Data.loadingMessage = "Test message"
	model.SetLoading(false)
	if model.Data.loading != false {
		t.Error("Expected loading to be false")
	}
	if model.Data.loadingMessage != "" {
		t.Error("Expected loadingMessage to be cleared when loading finishes")
	}
}

// TestSetLoadingWithMessage tests loading with custom message
func TestSetLoadingWithMessage(t *testing.T) {
	model := NewModel(createTestConfig())
	testMessage := "Updating task status..."

	model.SetLoadingWithMessage(true, testMessage)
	if model.Data.loading != true {
		t.Error("Expected loading to be true")
	}
	if model.Data.loadingMessage != testMessage {
		t.Errorf("Expected loadingMessage to be %s, got %s", testMessage, model.Data.loadingMessage)
	}

	model.SetLoadingWithMessage(false, "")
	if model.Data.loading != false {
		t.Error("Expected loading to be false")
	}
	if model.Data.loadingMessage != "" {
		t.Error("Expected loadingMessage to be cleared")
	}
}

// TestFormatUserFriendlyError tests error message formatting
func TestFormatUserFriendlyError(t *testing.T) {
	model := NewModel(createTestConfig())

	testCases := []struct {
		input    string
		expected string
		shouldSetDisconnected bool
	}{
		{
			"connection refused",
			"Unable to connect to Archon server. Check if it's running on localhost:8181",
			true,
		},
		{
			"no such host",
			"Unable to connect to Archon server. Check if it's running on localhost:8181",
			true,
		},
		{
			"timeout occurred",
			"Connection timeout. The server may be slow or unreachable",
			true,
		},
		{
			"status 401",
			"Authentication failed. Check your API key configuration",
			false,
		},
		{
			"status 404",
			"Resource not found. The task or project may have been deleted",
			false,
		},
		{
			"status 500",
			"Server error. Please try again or contact support",
			false,
		},
		{
			"unknown error",
			"unknown error",
			false,
		},
	}

	for _, tc := range testCases {
		model.Data.connected = true // Reset connection status
		result := model.FormatUserFriendlyError(tc.input)

		if result != tc.expected {
			t.Errorf("For input %s, expected %s, got %s", tc.input, tc.expected, result)
		}

		if tc.shouldSetDisconnected && model.Data.connected {
			t.Errorf("Expected connection status to be false for error: %s", tc.input)
		}
	}
}

// TestSetConnectionStatus tests connection status management
func TestSetConnectionStatus(t *testing.T) {
	model := NewModel(createTestConfig())

	model.SetConnectionStatus(true)
	if model.Data.connected != true {
		t.Error("Expected connected to be true")
	}

	model.SetConnectionStatus(false)
	if model.Data.connected != false {
		t.Error("Expected connected to be false")
	}
}

// TestGetConnectionStatusText tests connection status display
func TestGetConnectionStatusText(t *testing.T) {
	model := NewModel(createTestConfig())

	model.Data.connected = true
	if model.GetConnectionStatusText() != "●" {
		t.Error("Expected connected indicator to be ●")
	}

	model.Data.connected = false
	if model.GetConnectionStatusText() != "○" {
		t.Error("Expected disconnected indicator to be ○")
	}
}

// TestUpdateTasks tests task list updates and index adjustments
func TestUpdateTasks(t *testing.T) {
	model := NewModel(createTestConfig())
	// Set proper window dimensions to avoid styling.RenderLine issues
	model.Window.width = 120
	model.Window.height = 40
	model.Navigation.selectedIndex = 5 // Set out of bounds initially

	tasks := []archon.Task{
		{ID: "1", Title: "Task 1", Status: "todo"},
		{ID: "2", Title: "Task 2", Status: "doing"},
		{ID: "3", Title: "Task 3", Status: "done"},
	}

	model.UpdateTasks(tasks)

	if model.Data.loading != false {
		t.Error("Expected loading to be false after update")
	}
	if len(model.Data.tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(model.Data.tasks))
	}
	if model.Data.connected != true {
		t.Error("Expected connected to be true after successful update")
	}
	if model.Data.error != "" {
		t.Error("Expected error to be cleared after successful update")
	}
	if model.Navigation.selectedIndex != 2 {
		t.Errorf("Expected selectedIndex to be adjusted to 2, got %d", model.Navigation.selectedIndex)
	}

	// Test with empty tasks
	model.UpdateTasks([]archon.Task{})
	if model.Navigation.selectedIndex != 0 {
		t.Error("Expected selectedIndex to be 0 with empty tasks")
	}
}

// TestUpdateProjects tests project list updates and validation
func TestUpdateProjects(t *testing.T) {
	model := NewModel(createTestConfig())

	// Set a selected project that won't exist in the new list
	nonExistentID := "non-existent-project"
	model.Data.selectedProjectID = &nonExistentID

	projects := []archon.Project{
		{ID: "proj1", Title: "Project 1"},
		{ID: "proj2", Title: "Project 2"},
	}

	model.UpdateProjects(projects)

	if len(model.Data.projects) != 2 {
		t.Errorf("Expected 2 projects, got %d", len(model.Data.projects))
	}
	if model.Data.connected != true {
		t.Error("Expected connected to be true after successful update")
	}
	if model.Data.selectedProjectID != nil {
		t.Error("Expected selectedProjectID to be reset when project no longer exists")
	}

	// Test with existing project
	existingID := "proj1"
	model.Data.selectedProjectID = &existingID
	model.UpdateProjects(projects)
	if model.Data.selectedProjectID == nil || *model.Data.selectedProjectID != existingID {
		t.Error("Expected selectedProjectID to be preserved when project exists")
	}
}

// TestSpinnerAnimation tests spinner state management
func TestSpinnerAnimation(t *testing.T) {
	model := NewModel(createTestConfig())

	// Test initial spinner
	initialSpinner := model.GetLoadingSpinner()
	if initialSpinner == "" {
		t.Error("Expected spinner to return a character")
	}

	// Test spinner advancement
	model.AdvanceSpinner()
	model.GetLoadingSpinner() // Just verify it doesn't crash

	// The spinner has 4 characters, so after 4 advances it should cycle back
	for i := 0; i < 5; i++ {
		model.AdvanceSpinner()
	}
	// Should have cycled back around
	cycledSpinner := model.GetLoadingSpinner()
	if cycledSpinner == "" {
		t.Error("Expected spinner to return a character after cycling")
	}
}

// TestModalStateTransitions tests modal state management
func TestModalStateTransitions(t *testing.T) {
	model := NewModel(createTestConfig())

	// Test help modal
	if model.Modals.help.active {
		t.Error("Expected help modal to be inactive initially")
	}

	// Test status change modal
	if model.Modals.statusChange.active {
		t.Error("Expected status change modal to be inactive initially")
	}
	if model.Modals.statusChange.selectedIndex != 0 {
		t.Error("Expected status change modal selectedIndex to be 0 initially")
	}

	// Test project mode
	if model.Modals.projectMode.active {
		t.Error("Expected project mode to be inactive initially")
	}

	// Test feature mode initialization
	if model.Modals.featureMode.active {
		t.Error("Expected feature mode to be inactive initially")
	}
	if model.Modals.featureMode.selectedFeatures == nil {
		t.Error("Expected selectedFeatures to be initialized")
	}
}

// TestWindowStateManagement tests window state transitions
func TestWindowStateManagement(t *testing.T) {
	model := NewModel(createTestConfig())

	// Test initial state
	if model.Window.activeView != LeftPanel {
		t.Error("Expected activeView to be LeftPanel initially")
	}
	if model.Window.ready {
		t.Error("Expected window to not be ready initially")
	}
	if model.Window.width != 0 || model.Window.height != 0 {
		t.Error("Expected initial window dimensions to be 0")
	}
}

// TestNavigationStateManagement tests navigation state
func TestNavigationStateManagement(t *testing.T) {
	model := NewModel(createTestConfig())

	// Test initial navigation state
	if model.Navigation.selectedIndex != 0 {
		t.Error("Expected selectedIndex to be 0 initially")
	}

	// Test key sequence state initialization
	if model.Navigation.keySequence.lastKeyPressed != "" {
		t.Error("Expected lastKeyPressed to be empty initially")
	}
}

// TestCycleSortModeAdvanced tests sort mode cycling and task selection preservation
func TestCycleSortModeAdvanced(t *testing.T) {
	model := NewModel(createTestConfig())
	model.Window.width = 120
	model.Window.height = 40

	// Add test tasks
	model.Data.tasks = []archon.Task{
		{ID: "task1", Title: "Task 1", Status: "todo", TaskOrder: 10},
		{ID: "task2", Title: "Task 2", Status: "doing", TaskOrder: 5},
		{ID: "task3", Title: "Task 3", Status: "done", TaskOrder: 15},
	}

	initialMode := model.Data.sortMode
	selectedTaskIndex := 1
	model.Navigation.selectedIndex = selectedTaskIndex

	sortedTasks := model.GetSortedTasks()
	if len(sortedTasks) < 2 {
		t.Fatal("Need at least 2 tasks for this test")
	}
	selectedTaskID := sortedTasks[selectedTaskIndex].ID

	// Cycle sort mode
	model.CycleSortMode()

	if model.Data.sortMode == initialMode {
		t.Error("Expected sort mode to change")
	}

	// Verify task selection is preserved by ID (may have different index in new sort)
	newSortedTasks := model.GetSortedTasks()
	foundTask := false
	for i, task := range newSortedTasks {
		if task.ID == selectedTaskID {
			foundTask = true
			if model.Navigation.selectedIndex != i {
				t.Errorf("Expected selectedIndex to be updated to %d to track same task, got %d", i, model.Navigation.selectedIndex)
			}
			break
		}
	}
	if !foundTask {
		t.Error("Expected to find the originally selected task in new sort order")
	}

	// Test cycling through all modes
	originalMode := model.Data.sortMode
	for i := 0; i < 4; i++ {
		model.CycleSortMode()
	}
	if model.Data.sortMode != originalMode {
		t.Errorf("Expected to cycle back to original mode %d, got %d", originalMode, model.Data.sortMode)
	}
}

// TestCycleSortModePrevious tests reverse sort mode cycling
func TestCycleSortModePrevious(t *testing.T) {
	model := NewModel(createTestConfig())
	model.Data.tasks = []archon.Task{
		{ID: "task1", Title: "Task 1", Status: "todo"},
	}

	initialMode := model.Data.sortMode

	// Cycle previous once, then forward once - should return to initial
	model.CycleSortModePrevious()
	previousMode := model.Data.sortMode

	model.CycleSortMode()
	if model.Data.sortMode != initialMode {
		t.Errorf("Expected to return to initial mode %d after previous->forward cycle, got %d", initialMode, model.Data.sortMode)
	}

	// Verify previous mode was actually different
	if previousMode == initialMode {
		t.Error("Expected previous mode to be different from initial")
	}
}

// TestTemporaryStatusMessage tests status message timing
func TestTemporaryStatusMessage(t *testing.T) {
	model := NewModel(createTestConfig())

	testMessage := "Test status message"
	beforeTime := time.Now()

	updatedModel := model.setTemporaryStatusMessage(testMessage)

	afterTime := time.Now()

	if updatedModel.Data.statusMessage != testMessage {
		t.Errorf("Expected status message to be %s, got %s", testMessage, updatedModel.Data.statusMessage)
	}

	if updatedModel.Data.statusMessageTime.Before(beforeTime) || updatedModel.Data.statusMessageTime.After(afterTime) {
		t.Error("Expected status message time to be set to current time")
	}
}

// Note: TestDependencyInjectionModel moved to test/integration_test.go to avoid import cycles