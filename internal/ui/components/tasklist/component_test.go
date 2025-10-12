package tasklist

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/internal/archon"
	"github.com/yousfisaad/lazyarchon/internal/ui/components/base"
)

func TestTaskListComponent(t *testing.T) {
	t.Run("NewModel creates component with defaults", func(t *testing.T) {
		ctx := &base.ComponentContext{}
		model := NewModel(Options{Context: ctx})

		if model.GetWidth() != 40 {
			t.Errorf("Expected default width 40, got %d", model.GetWidth())
		}
		if model.GetHeight() != 20 {
			t.Errorf("Expected default height 20, got %d", model.GetHeight())
		}
		if model.maxLines != 14 {
			t.Errorf("Expected max lines 14, got %d", model.maxLines)
		}
		if model.selectedIndex != 0 {
			t.Errorf("Expected selected index 0, got %d", model.selectedIndex)
		}
	})

	t.Run("NewModel respects provided options", func(t *testing.T) {
		tasks := []archon.Task{
			{ID: "1", Title: "Task 1"},
			{ID: "2", Title: "Task 2"},
		}

		ctx := &base.ComponentContext{}
		opts := Options{
			Width:         60,
			Height:        30,
			Tasks:         tasks,
			SelectedIndex: 1,
			SearchQuery:   "task", // Use a query that won't filter out tasks
			SearchActive:  false,  // Don't enable search filtering
			Context:       ctx,
		}

		model := NewModel(opts)

		if model.GetWidth() != 60 {
			t.Errorf("Expected width 60, got %d", model.GetWidth())
		}
		if model.GetHeight() != 30 {
			t.Errorf("Expected height 30, got %d", model.GetHeight())
		}
		// Note: Tasks are no longer cached in component - they're queried from parent on-demand
		if model.selectedIndex != 1 {
			t.Errorf("Expected selected index 1, got %d", model.selectedIndex)
		}
		if model.searchQuery != "task" {
			t.Errorf("Expected search query 'task', got %s", model.searchQuery)
		}
		if model.searchActive {
			t.Error("Expected search active to be false")
		}
		// Note: Active state is no longer cached - it's read from parent via callback
	})

	t.Run("Init returns nil", func(t *testing.T) {
		ctx := &base.ComponentContext{}
		model := NewModel(Options{Context: ctx})

		cmd := model.Init()
		if cmd != nil {
			t.Error("Expected Init to return nil")
		}
	})

	t.Run("GetSelectedTask returns correct task", func(t *testing.T) {
		tasks := []archon.Task{
			{ID: "1", Title: "Task 1"},
			{ID: "2", Title: "Task 2"},
			{ID: "3", Title: "Task 3"},
		}

		ctx := &base.ComponentContext{}
		model := NewModel(Options{
			Tasks:         tasks,
			SelectedIndex: 1,
			Context:       ctx,
		})

		selectedTask := model.GetSelectedTask()
		if selectedTask == nil {
			t.Fatal("Expected selected task to not be nil")
		}
		if selectedTask.ID != "2" {
			t.Errorf("Expected selected task ID '2', got %s", selectedTask.ID)
		}
		if selectedTask.Title != "Task 2" {
			t.Errorf("Expected selected task title 'Task 2', got %s", selectedTask.Title)
		}
	})

	t.Run("GetSelectedTask returns nil for invalid index", func(t *testing.T) {
		tasks := []archon.Task{
			{ID: "1", Title: "Task 1"},
		}

		ctx := &base.ComponentContext{}
		model := NewModel(Options{
			Tasks:         tasks,
			SelectedIndex: 5, // Invalid index - parent model should validate this
			Context:       ctx,
		})

		// Component should not modify selectedIndex, so it remains invalid
		selectedTask := model.GetSelectedTask()
		if selectedTask != nil {
			t.Error("Expected selected task to be nil for invalid index")
		}
		if model.selectedIndex != 5 {
			t.Errorf("Expected selected index to remain 5, got %d", model.selectedIndex)
		}
	})

	t.Run("GetSelectedIndex returns correct index", func(t *testing.T) {
		tasks := []archon.Task{
			{ID: "1", Title: "Task 1"},
			{ID: "2", Title: "Task 2"},
			{ID: "3", Title: "Task 3"},
			{ID: "4", Title: "Task 4"},
		}

		ctx := &base.ComponentContext{}
		model := NewModel(Options{
			Tasks:         tasks,
			SelectedIndex: 3,
			Context:       ctx,
		})

		index := model.GetSelectedIndex()
		if index != 3 {
			t.Errorf("Expected selected index 3, got %d", index)
		}
	})

	t.Run("GetTaskCount returns correct count", func(t *testing.T) {
		tasks := []archon.Task{
			{ID: "1", Title: "Task 1"},
			{ID: "2", Title: "Task 2"},
			{ID: "3", Title: "Task 3"},
		}

		ctx := &base.ComponentContext{}
		model := NewModel(Options{
			Tasks:   tasks,
			Context: ctx,
		})

		count := model.GetTaskCount()
		if count != 3 {
			t.Errorf("Expected task count 3, got %d", count)
		}
	})

	t.Run("search highlighting preserves all tasks", func(t *testing.T) {
		// Note: This test verifies search state management only
		// Tasks are no longer cached - they're queried from parent on-demand
		ctx := &base.ComponentContext{}
		model := NewModel(Options{
			SearchQuery:  "auth",
			SearchActive: true,
			Context:      ctx,
		})

		// Search highlighting should be handled during rendering, not filtering
		if !model.searchActive || model.searchQuery != "auth" {
			t.Error("Search state should be preserved for rendering")
		}
	})

	t.Run("preserves selectedIndex from options", func(t *testing.T) {
		// Note: Tasks are no longer cached - they're queried from parent on-demand
		ctx := &base.ComponentContext{}
		model := NewModel(Options{
			SelectedIndex: 5, // Component preserves whatever index is provided
			Context:       ctx,
		})

		// Component should preserve the selectedIndex provided in options
		if model.selectedIndex != 5 {
			t.Errorf("Expected selected index to remain 5, got %d", model.selectedIndex)
		}
	})

	// Note: View special states test removed - loading/error/empty states are now read from
	// ProgramContext (single source of truth) instead of cached fields. These states should be
	// tested at the integration level where ProgramContext is properly set up.

	t.Run("Window resize updates dimensions", func(t *testing.T) {
		ctx := &base.ComponentContext{}
		model := NewModel(Options{Context: ctx})

		resizeMsg := tea.WindowSizeMsg{
			Width:  48, // Left panel width (half of MainContentWidth = 96/2 = 48)
			Height: 42, // MainContentHeight (50 - 8 = 42)
		}

		cmd := model.Update(resizeMsg)
		// Window resize returns a batch command which might be nil if no scrollbar update is needed
		_ = cmd

		if model.GetWidth() != 48 { // Half width for left panel (MainContentWidth/2 = 96/2 = 48)
			t.Errorf("Expected width to be updated to 48, got %d", model.GetWidth())
		}
		if model.GetHeight() != 42 { // Height with proper layout calculation (height - 8 = 50 - 8 = 42)
			t.Errorf("Expected height to be updated to 42, got %d", model.GetHeight())
		}
	})
}

func TestTaskListHelperFunctions(t *testing.T) {
	t.Run("max function works correctly", func(t *testing.T) {
		if max(5, 3) != 5 {
			t.Error("max(5, 3) should return 5")
		}
		if max(2, 8) != 8 {
			t.Error("max(2, 8) should return 8")
		}
		if max(4, 4) != 4 {
			t.Error("max(4, 4) should return 4")
		}
	})

	t.Run("min function works correctly", func(t *testing.T) {
		if min(5, 3) != 3 {
			t.Error("min(5, 3) should return 3")
		}
		if min(2, 8) != 2 {
			t.Error("min(2, 8) should return 2")
		}
		if min(4, 4) != 4 {
			t.Error("min(4, 4) should return 4")
		}
	})
}
