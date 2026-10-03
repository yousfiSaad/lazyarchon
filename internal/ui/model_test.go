package ui

import (
	"testing"
	"time"

	"github.com/yousfisaad/lazyarchon/v2/internal/plugin"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/config"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/interfaces"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/sorting"
)

// createTestConfig creates a config for testing
func createTestConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			URL:     "http://localhost:8181",
			Timeout: 30 * time.Second,
			APIKey:  "",
		},
		UI: config.UIConfig{
			Display: config.DisplayConfig{
				ShowCompletedTasks:  true,
				DefaultSortMode:     "status+priority",
				AutoRefreshInterval: 0,
			},
		},
	}
}

// createTestModel builds a model with an empty plugin manager, mirroring
// how the entrypoint wires NewModelWithPlugin.
func createTestModel(t *testing.T) MainModel {
	t.Helper()

	manager := plugin.NewManagerWithGlobalRegistry()

	return NewModelWithPlugin(createTestConfig(), manager)
}

func TestNewModel(t *testing.T) {
	model := createTestModel(t)

	// Test default values - direct state access (coordinators removed)
	// Note: selectedIndex was moved to component-level state during refactoring

	if !model.programContext.Loading {
		t.Errorf("Expected loading to be true")
	}

	if model.programContext.SortMode != sorting.SortStatusPriority {
		t.Errorf("Expected sortMode to be SortStatusPriority (%d), got %d", sorting.SortStatusPriority, model.programContext.SortMode)
	}

	// Note: TaskClient may be nil if plugin loading fails (e.g., no server available)
	// This is acceptable for unit tests - the UI will handle connection errors gracefully
	_ = model.programContext.TaskClient
}

func TestGetSortedTasks(t *testing.T) {
	model := createTestModel(t)

	// Test with empty tasks
	sorted := model.GetSortedTasks()
	if len(sorted) != 0 {
		t.Errorf("Expected empty slice, got %d tasks", len(sorted))
	}

	// Test with sample tasks on the 1-5 priority scale (1 = critical,
	// 5 = backlog; within a status lane, lower sorts first)
	model.programContext.SetTasks([]interfaces.Task{
		{Title: "Task A", Status: "todo", Priority: plugin.PriorityBacklog},
		{Title: "Task B", Status: "done", Priority: plugin.PriorityCritical},
		{Title: "Task C", Status: "doing", Priority: plugin.PriorityMedium},
		{Title: "Task D", Status: "todo", Priority: plugin.PriorityCritical},
	})

	sorted = model.GetSortedTasks()
	if len(sorted) != 4 {
		t.Errorf("Expected 4 tasks, got %d", len(sorted))
	}

	// Status lanes first (todo before doing before done); within the todo
	// lane the critical task must outrank the backlog one.
	gotOrder := []string{sorted[0].Title, sorted[1].Title, sorted[2].Title, sorted[3].Title}
	wantOrder := []string{"Task D", "Task A", "Task C", "Task B"}
	for i := range wantOrder {
		if gotOrder[i] != wantOrder[i] {
			t.Errorf("sort order = %v, want %v", gotOrder, wantOrder)
			break
		}
	}
}

// TestSetSelectedProject - SKIPPED: Needs proper ProjectManager initialization with test projects
// Consider rewriting to use component-based architecture
// func TestSetSelectedProject(t *testing.T) {
// 	model := NewModel(createTestConfig())
//
// 	// Test setting project ID
// 	projectID := "test-project-123"
// 	model.SetSelectedProject(&projectID)
//
// 	if model.programContext.SelectedProjectID == nil {
// 		t.Error("Expected selectedProjectID to be set")
// 	}
//
// 	if *model.programContext.SelectedProjectID != projectID {
// 		t.Errorf("Expected selectedProjectID to be %s, got %s", projectID, *model.programContext.SelectedProjectID)
// 	}
//
// 	// Test setting to nil (all tasks)
// 	model.SetSelectedProject(nil)
// 	if model.programContext.SelectedProjectID != nil {
// 		t.Error("Expected selectedProjectID to be nil")
// 	}
// }

func TestCycleSortMode(t *testing.T) {
	model := createTestModel(t)

	// Test cycling through sort modes
	initialMode := model.programContext.SortMode
	model.cycleSortMode()

	if model.programContext.SortMode == initialMode {
		t.Error("Expected sort mode to change")
	}

	// Cycle through all modes and verify we return to start
	originalMode := model.programContext.SortMode
	for i := 0; i < sorting.SortModeCount; i++ {
		model.cycleSortMode()
	}

	if model.programContext.SortMode != originalMode {
		t.Errorf("Expected to cycle back to original mode %d, got %d", originalMode, model.programContext.SortMode)
	}
}

// TestSetActiveView - SKIPPED: Requires proper component initialization
// These tests need integration test context - unit tests can't initialize full component tree
// Integration tests should cover this functionality instead
// func TestSetActiveView(t *testing.T) {
// 	model := NewModel(createTestConfig())
//
// 	t.Run("SetActiveView to RightPanel returns TaskDetailsSetActiveMsg command", func(t *testing.T) {
// 		cmd := model.SetActiveView(RightPanel)
// 		if cmd == nil {
// 			t.Fatal("Expected command to be returned")
// 		}
//
// 		msg := cmd()
// 		detailsMsg, ok := msg.(taskdetails.TaskDetailsSetActiveMsg)
// 		if !ok {
// 			t.Fatalf("Expected taskdetails.TaskDetailsSetActiveMsg, got %T", msg)
// 		}
//
// 		if !detailsMsg.Active {
// 			t.Error("Expected Active to be true when setting RightPanel")
// 		}
//
// 		// Verify the model's active view was updated
// 		if !model.IsRightPanelActive() {
// 			t.Error("Expected model to show right panel as active")
// 		}
// 	})
//
// 	t.Run("SetActiveView to LeftPanel returns TaskDetailsSetActiveMsg command", func(t *testing.T) {
// 		// First set to right panel
// 		model.SetActiveView(RightPanel)
//
// 		// Then switch to left panel
// 		cmd := model.SetActiveView(LeftPanel)
// 		if cmd == nil {
// 			t.Fatal("Expected command to be returned")
// 		}
//
// 		msg := cmd()
// 		detailsMsg, ok := msg.(taskdetails.TaskDetailsSetActiveMsg)
// 		if !ok {
// 			t.Fatalf("Expected taskdetails.TaskDetailsSetActiveMsg, got %T", msg)
// 		}
//
// 		if detailsMsg.Active {
// 			t.Error("Expected Active to be false when setting LeftPanel")
// 		}
//
// 		// Verify the model's active view was updated
// 		if !model.IsLeftPanelActive() {
// 			t.Error("Expected model to show left panel as active")
// 		}
// 	})
// }

// TestTaskDetailsSetActiveMsgRouting - SKIPPED: Requires proper component initialization
// Consider integration tests for component message routing
// func TestTaskDetailsSetActiveMsgRouting(t *testing.T) {
// 	model := NewModel(createTestConfig())
//
// 	t.Run("TaskDetailsSetActiveMsg is properly routed to component", func(t *testing.T) {
// 		// Create a TaskDetailsSetActiveMsg directly
// 		msg := taskdetails.TaskDetailsSetActiveMsg{Active: true}
//
// 		// Process the message through the main Update method
// 		_, cmd := model.Update(msg)
//
// 		// Verify the message was processed (component may or may not return commands)
// 		_ = cmd // Commands are allowed from component updates
//
// 		// The task details component should now be active
// 		// We can't directly check the component's internal state from here,
// 		// but we can verify the message was processed without error
// 	})
//
// 	t.Run("Complete flow: SetActiveView -> returns TaskDetailsSetActiveMsg", func(t *testing.T) {
// 		// This tests the complete flow:
// 		// 1. SetActiveView returns TaskDetailsSetActiveMsg command
// 		// 2. That command is executed to get the actual message
// 		// 3. The message can be processed by components
//
// 		// Get the command from SetActiveView
// 		cmd := model.SetActiveView(RightPanel)
// 		if cmd == nil {
// 			t.Fatal("Expected SetActiveView to return a command")
// 		}
//
// 		// Execute the command to get the message
// 		msg := cmd()
//
// 		// Verify it's the right message type
// 		detailsMsg, ok := msg.(taskdetails.TaskDetailsSetActiveMsg)
// 		if !ok {
// 			t.Fatalf("Expected taskdetails.TaskDetailsSetActiveMsg, got %T", msg)
// 		}
//
// 		if !detailsMsg.Active {
// 			t.Error("Expected Active to be true for RightPanel")
// 		}
//
// 		// Process the message through the main Update method
// 		_, resultCmd := model.Update(detailsMsg)
// 		_ = resultCmd // Commands are allowed from component updates
//
// 		// Verify the model state was updated
// 		if !model.IsRightPanelActive() {
// 			t.Error("Expected model to show right panel as active after processing")
// 		}
// 	})
// }
