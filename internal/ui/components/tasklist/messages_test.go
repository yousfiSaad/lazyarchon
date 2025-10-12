package tasklist

import (
	"testing"

	"github.com/yousfisaad/lazyarchon/internal/archon"
	"github.com/yousfisaad/lazyarchon/internal/ui/components/base"
)

func TestTaskListMessages(t *testing.T) {
	t.Run("UpdateTaskList creates correct message", func(t *testing.T) {
		tasks := []archon.Task{
			{ID: "1", Title: "Test Task"},
		}
		cmd := UpdateTaskList(tasks, true, "test error")
		msg := cmd()

		updateMsg, ok := msg.(TaskListUpdateMsg)
		if !ok {
			t.Fatalf("Expected TaskListUpdateMsg, got %T", msg)
		}

		if len(updateMsg.Tasks) != 1 {
			t.Errorf("Expected 1 task, got %d", len(updateMsg.Tasks))
		}
		if updateMsg.Tasks[0].Title != "Test Task" {
			t.Errorf("Expected task title 'Test Task', got %s", updateMsg.Tasks[0].Title)
		}
		if !updateMsg.Loading {
			t.Error("Expected loading to be true")
		}
		if updateMsg.Error != "test error" {
			t.Errorf("Expected error 'test error', got %s", updateMsg.Error)
		}
	})

	t.Run("SelectTask creates correct message", func(t *testing.T) {
		cmd := SelectTask(5)
		msg := cmd()

		selectMsg, ok := msg.(TaskListSelectMsg)
		if !ok {
			t.Fatalf("Expected TaskListSelectMsg, got %T", msg)
		}

		if selectMsg.Index != 5 {
			t.Errorf("Expected index 5, got %d", selectMsg.Index)
		}
	})

	t.Run("UpdateSearch creates correct message", func(t *testing.T) {
		cmd := UpdateSearch("test query", true)
		msg := cmd()

		searchMsg, ok := msg.(TaskListSearchMsg)
		if !ok {
			t.Fatalf("Expected TaskListSearchMsg, got %T", msg)
		}

		if searchMsg.Query != "test query" {
			t.Errorf("Expected query 'test query', got %s", searchMsg.Query)
		}
		if !searchMsg.Active {
			t.Error("Expected active to be true")
		}
	})

	t.Run("UpdateFilter creates correct message", func(t *testing.T) {
		cmd := UpdateFilter("auth", "todo")
		msg := cmd()

		filterMsg, ok := msg.(TaskListFilterMsg)
		if !ok {
			t.Fatalf("Expected TaskListFilterMsg, got %T", msg)
		}

		if filterMsg.Feature != "auth" {
			t.Errorf("Expected feature 'auth', got %s", filterMsg.Feature)
		}
		if filterMsg.Status != "todo" {
			t.Errorf("Expected status 'todo', got %s", filterMsg.Status)
		}
	})

	t.Run("ScrollTaskList creates correct message", func(t *testing.T) {
		cmd := ScrollTaskList(ScrollDown)
		msg := cmd()

		scrollMsg, ok := msg.(TaskListScrollMsg)
		if !ok {
			t.Fatalf("Expected TaskListScrollMsg, got %T", msg)
		}

		if scrollMsg.Direction != ScrollDown {
			t.Errorf("Expected direction ScrollDown, got %v", scrollMsg.Direction)
		}
	})

	// NOTE: SetTaskListActive test removed - message type and helper deleted
}

func TestTaskListComponentMessages(t *testing.T) {
	ctx := &base.ComponentContext{}
	model := NewModel(Options{Context: ctx})

	t.Run("TaskListUpdateMsg triggers viewport update", func(t *testing.T) {
		// Note: TaskList no longer caches tasks - it queries parent on-demand
		// TaskListUpdateMsg is now just a notification to refresh viewport
		msg := TaskListUpdateMsg{
			Tasks:   []archon.Task{}, // Not used - legacy field
			Loading: false,
			Error:   "",
		}

		cmd := model.Update(msg)
		if cmd != nil {
			t.Error("Expected no command from task list update")
		}

		// No fields to check - update only refreshes viewport from parent-provided callback
	})

	t.Run("TaskListSelectMsg updates selection", func(t *testing.T) {
		// Note: setSelectedIndex now queries parent for task count via callback
		// For this test, provide a mock callback that returns 2 tasks
		ctx.GetSortedTasks = func() []interface{} {
			return []interface{}{
				archon.Task{ID: "1", Title: "Task 1"},
				archon.Task{ID: "2", Title: "Task 2"},
			}
		}

		msg := TaskListSelectMsg{Index: 1}

		cmd := model.Update(msg)
		if cmd == nil {
			t.Error("Expected command from task list select")
		}

		if model.selectedIndex != 1 {
			t.Errorf("Expected selected index 1, got %d", model.selectedIndex)
		}

		// Check that the command returns selection changed message
		returnedMsg := cmd()
		if _, ok := returnedMsg.(TaskListSelectionChangedMsg); !ok {
			t.Errorf("Expected TaskListSelectionChangedMsg, got %T", returnedMsg)
		}
	})

	t.Run("TaskListSearchMsg updates search state", func(t *testing.T) {
		msg := TaskListSearchMsg{
			Query:  "test search",
			Active: true,
		}

		cmd := model.Update(msg)
		if cmd != nil {
			t.Error("Expected no command from search update")
		}

		if model.searchQuery != "test search" {
			t.Errorf("Expected search query 'test search', got %s", model.searchQuery)
		}
		if !model.searchActive {
			t.Error("Expected search active to be true")
		}
	})

	t.Run("TaskListFilterMsg updates filter state", func(t *testing.T) {
		msg := TaskListFilterMsg{
			Feature: "authentication",
			Status:  "todo",
		}

		cmd := model.Update(msg)
		if cmd != nil {
			t.Error("Expected no command from filter update")
		}

		if model.filterFeature != "authentication" {
			t.Errorf("Expected filter feature 'authentication', got %s", model.filterFeature)
		}
		if model.filterStatus != "todo" {
			t.Errorf("Expected filter status 'todo', got %s", model.filterStatus)
		}
	})

	t.Run("TaskListScrollMsg updates selection", func(t *testing.T) {
		// Provide mock callback that returns 3 tasks
		ctx.GetSortedTasks = func() []interface{} {
			return []interface{}{
				archon.Task{ID: "1", Title: "Task 1"},
				archon.Task{ID: "2", Title: "Task 2"},
				archon.Task{ID: "3", Title: "Task 3"},
			}
		}
		model.selectedIndex = 1

		// Test scroll down
		msg := TaskListScrollMsg{Direction: ScrollDown}
		cmd := model.Update(msg)
		if cmd == nil {
			t.Error("Expected command from scroll message")
		}

		if model.selectedIndex != 2 {
			t.Errorf("Expected selected index 2, got %d", model.selectedIndex)
		}

		// Test scroll up
		msg = TaskListScrollMsg{Direction: ScrollUp}
		cmd = model.Update(msg)
		if cmd == nil {
			t.Error("Expected command from scroll message")
		}

		if model.selectedIndex != 1 {
			t.Errorf("Expected selected index 1, got %d", model.selectedIndex)
		}

		// Test scroll to top
		msg = TaskListScrollMsg{Direction: ScrollToTop}
		cmd = model.Update(msg)
		if model.selectedIndex != 0 {
			t.Errorf("Expected selected index 0, got %d", model.selectedIndex)
		}

		// Test scroll to bottom
		msg = TaskListScrollMsg{Direction: ScrollToBottom}
		cmd = model.Update(msg)
		if model.selectedIndex != 2 {
			t.Errorf("Expected selected index 2, got %d", model.selectedIndex)
		}
	})

	// NOTE: TaskListSetActiveMsg test removed - message type and handler deleted

	t.Run("TaskListScrollMsg respects bounds", func(t *testing.T) {
		// Provide mock callback that returns 2 tasks
		ctx.GetSortedTasks = func() []interface{} {
			return []interface{}{
				archon.Task{ID: "1", Title: "Task 1"},
				archon.Task{ID: "2", Title: "Task 2"},
			}
		}
		model.selectedIndex = 0

		// Test scroll up at top (should stay at 0)
		msg := TaskListScrollMsg{Direction: ScrollUp}
		_ = model.Update(msg)

		if model.selectedIndex != 0 {
			t.Errorf("Expected selected index to remain 0, got %d", model.selectedIndex)
		}

		// Set to bottom
		model.selectedIndex = 1

		// Test scroll down at bottom (should stay at last index)
		msg = TaskListScrollMsg{Direction: ScrollDown}
		_ = model.Update(msg)

		if model.selectedIndex != 1 {
			t.Errorf("Expected selected index to remain 1, got %d", model.selectedIndex)
		}
	})
}
