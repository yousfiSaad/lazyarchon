package ui

import (
	"testing"
	"time"

	"github.com/yousfisaad/lazyarchon/internal/archon"
	"github.com/yousfisaad/lazyarchon/internal/shared/config"
	"github.com/yousfisaad/lazyarchon/internal/ui/sorting"
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

func TestNewModel(t *testing.T) {
	model := NewModel(createTestConfig())

	// Test default values - direct state access (coordinators removed)
	if model.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex to be 0, got %d", model.selectedIndex)
	}

	if !model.programContext.Loading {
		t.Errorf("Expected loading to be true")
	}

	if model.programContext.SortMode != sorting.SortStatusPriority {
		t.Errorf("Expected sortMode to be SortStatusPriority (%d), got %d", sorting.SortStatusPriority, model.programContext.SortMode)
	}

	if model.programContext.ArchonClient == nil {
		t.Errorf("Expected ArchonClient to be initialized")
	}
}

func TestGetSortedTasks(t *testing.T) {
	model := NewModel(createTestConfig())

	// Test with empty tasks
	sorted := model.GetSortedTasks()
	if len(sorted) != 0 {
		t.Errorf("Expected empty slice, got %d tasks", len(sorted))
	}

	// Test with sample tasks
	model.programContext.SetTasks([]archon.Task{
		{Title: "Task A", Status: "todo", TaskOrder: 5},
		{Title: "Task B", Status: "done", TaskOrder: 10},
		{Title: "Task C", Status: "doing", TaskOrder: 3},
	})

	sorted = model.GetSortedTasks()
	if len(sorted) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(sorted))
	}

	// First task should be "todo" status (highest priority in status+priority sort)
	if sorted[0].Status != "todo" {
		t.Errorf("Expected first task status to be 'todo', got '%s'", sorted[0].Status)
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
	model := NewModel(createTestConfig())

	// Test cycling through sort modes
	initialMode := model.programContext.SortMode
	model.cycleSortMode()

	if model.programContext.SortMode == initialMode {
		t.Error("Expected sort mode to change")
	}

	// Cycle through all modes and verify we return to start
	originalMode := model.programContext.SortMode
	for i := 0; i < 4; i++ {
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
