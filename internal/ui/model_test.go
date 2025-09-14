package ui

import (
	"testing"
	"time"

	"github.com/yousfisaad/lazyarchon/internal/archon"
	"github.com/yousfisaad/lazyarchon/internal/config"
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

	// Test default values
	if model.Navigation.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex to be 0, got %d", model.Navigation.selectedIndex)
	}

	if !model.Data.loading {
		t.Errorf("Expected loading to be true")
	}

	if model.Data.sortMode != SortStatusPriority {
		t.Errorf("Expected sortMode to be SortStatusPriority (%d), got %d", SortStatusPriority, model.Data.sortMode)
	}

	if model.client == nil {
		t.Errorf("Expected client to be initialized")
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
	model.Data.tasks = []archon.Task{
		{Title: "Task A", Status: "todo", TaskOrder: 5},
		{Title: "Task B", Status: "done", TaskOrder: 10},
		{Title: "Task C", Status: "doing", TaskOrder: 3},
	}

	sorted = model.GetSortedTasks()
	if len(sorted) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(sorted))
	}

	// First task should be "todo" status (highest priority in status+priority sort)
	if sorted[0].Status != "todo" {
		t.Errorf("Expected first task status to be 'todo', got '%s'", sorted[0].Status)
	}
}

func TestSetSelectedProject(t *testing.T) {
	model := NewModel(createTestConfig())

	// Test setting project ID
	projectID := "test-project-123"
	model.SetSelectedProject(&projectID)

	if model.Data.selectedProjectID == nil {
		t.Error("Expected selectedProjectID to be set")
	}

	if *model.Data.selectedProjectID != projectID {
		t.Errorf("Expected selectedProjectID to be %s, got %s", projectID, *model.Data.selectedProjectID)
	}

	// Test setting to nil (all tasks)
	model.SetSelectedProject(nil)
	if model.Data.selectedProjectID != nil {
		t.Error("Expected selectedProjectID to be nil")
	}
}

func TestCycleSortMode(t *testing.T) {
	model := NewModel(createTestConfig())

	// Test cycling through sort modes
	initialMode := model.Data.sortMode
	model.CycleSortMode()

	if model.Data.sortMode == initialMode {
		t.Error("Expected sort mode to change")
	}

	// Cycle through all modes and verify we return to start
	originalMode := model.Data.sortMode
	for i := 0; i < 4; i++ {
		model.CycleSortMode()
	}

	if model.Data.sortMode != originalMode {
		t.Errorf("Expected to cycle back to original mode %d, got %d", originalMode, model.Data.sortMode)
	}
}
