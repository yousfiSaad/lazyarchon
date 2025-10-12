package taskitem

import (
	"testing"

	"github.com/yousfisaad/lazyarchon/internal/archon"
	"github.com/yousfisaad/lazyarchon/internal/ui/components/base"
)

func TestNewModel(t *testing.T) {
	task := archon.Task{
		ID:     "test-task-1",
		Title:  "Test Task",
		Status: "todo",
	}

	opts := Options{
		Task:       task,
		Index:      0,
		Width:      80,
		IsSelected: false,
		Context:    &base.ComponentContext{},
	}

	model := NewModel(opts)

	// Test basic properties
	if model.task.ID != task.ID {
		t.Errorf("Expected task ID %s, got %s", task.ID, model.task.ID)
	}

	if model.index != 0 {
		t.Errorf("Expected index 0, got %d", model.index)
	}

	if model.GetWidth() != 80 {
		t.Errorf("Expected width 80, got %d", model.GetWidth())
	}

	if model.GetHeight() != 1 {
		t.Errorf("Expected height 1, got %d", model.GetHeight())
	}

	if model.isSelected != false {
		t.Errorf("Expected isSelected false, got %t", model.isSelected)
	}
}

func TestTaskItemUpdate(t *testing.T) {
	task := archon.Task{
		ID:     "test-task-1",
		Title:  "Test Task",
		Status: "todo",
	}

	opts := Options{
		Task:       task,
		Index:      0,
		Width:      80,
		IsSelected: false,
		Context:    &base.ComponentContext{},
	}

	model := NewModel(opts)

	// Update with new task data
	updatedTask := archon.Task{
		ID:     "test-task-1",
		Title:  "Updated Test Task",
		Status: "doing",
	}

	updateMsg := TaskItemUpdateMsg{
		Index:         0,
		Task:          updatedTask,
		IsSelected:    true,
		IsHighlighted: true,
		SearchQuery:   "test",
	}

	updatedModel, _ := model.Update(updateMsg)
	model = *(updatedModel.(*Model))

	// Verify updates
	if model.task.Title != updatedTask.Title {
		t.Errorf("Expected task title %s, got %s", updatedTask.Title, model.task.Title)
	}

	if model.task.Status != updatedTask.Status {
		t.Errorf("Expected task status %s, got %s", updatedTask.Status, model.task.Status)
	}

	if !model.isSelected {
		t.Error("Expected task to be selected")
	}

	if !model.isHighlighted {
		t.Error("Expected task to be highlighted")
	}

	if model.searchQuery != "test" {
		t.Errorf("Expected search query 'test', got %s", model.searchQuery)
	}
}

func TestTaskItemUpdateWrongIndex(t *testing.T) {
	task := archon.Task{
		ID:     "test-task-1",
		Title:  "Test Task",
		Status: "todo",
	}

	opts := Options{
		Task:    task,
		Index:   0,
		Width:   80,
		Context: &base.ComponentContext{},
	}

	model := NewModel(opts)
	originalTitle := model.task.Title

	// Update with wrong index should not affect this model
	updateMsg := TaskItemUpdateMsg{
		Index: 1, // Different index
		Task: archon.Task{
			ID:     "test-task-1",
			Title:  "Updated Test Task",
			Status: "doing",
		},
		IsSelected: true,
	}

	updatedModel, _ := model.Update(updateMsg)
	model = *(updatedModel.(*Model))

	// Should not have changed
	if model.task.Title != originalTitle {
		t.Errorf("Task should not have been updated, but title changed from %s to %s", originalTitle, model.task.Title)
	}

	if model.isSelected {
		t.Error("Task should not be selected")
	}
}

func TestTaskItemResize(t *testing.T) {
	task := archon.Task{
		ID:     "test-task-1",
		Title:  "Test Task",
		Status: "todo",
	}

	opts := Options{
		Task:    task,
		Index:   0,
		Width:   80,
		Context: &base.ComponentContext{},
	}

	model := NewModel(opts)

	// Resize
	resizeMsg := TaskItemResizeMsg{
		Index: 0,
		Width: 120,
	}

	updatedModel, _ := model.Update(resizeMsg)
	model = *(updatedModel.(*Model))

	if model.GetWidth() != 120 {
		t.Errorf("Expected width 120, got %d", model.GetWidth())
	}
}

func TestGetters(t *testing.T) {
	task := archon.Task{
		ID:     "test-task-1",
		Title:  "Test Task",
		Status: "todo",
	}

	opts := Options{
		Task:          task,
		Index:         5,
		Width:         80,
		IsSelected:    true,
		IsHighlighted: true,
		Context:       &base.ComponentContext{},
	}

	model := NewModel(opts)

	// Test getters
	if model.GetTask().ID != task.ID {
		t.Errorf("GetTask() returned wrong task ID")
	}

	if model.GetIndex() != 5 {
		t.Errorf("GetIndex() returned %d, expected 5", model.GetIndex())
	}

	if !model.IsSelected() {
		t.Error("IsSelected() should return true")
	}

	if !model.IsHighlighted() {
		t.Error("IsHighlighted() should return true")
	}

	if model.GetHeight() != 1 {
		t.Errorf("GetHeight() returned %d, expected 1", model.GetHeight())
	}

	if model.GetWidth() != 80 {
		t.Errorf("GetWidth() returned %d, expected 80", model.GetWidth())
	}
}

func TestSetters(t *testing.T) {
	task := archon.Task{
		ID:     "test-task-1",
		Title:  "Test Task",
		Status: "todo",
	}

	opts := Options{
		Task:    task,
		Index:   0,
		Width:   80,
		Context: &base.ComponentContext{},
	}

	model := NewModel(opts)

	// Test setters
	model.SetSelected(true)
	if !model.IsSelected() {
		t.Error("SetSelected(true) did not set selection")
	}

	model.SetHighlighted(true, "search")
	if !model.IsHighlighted() {
		t.Error("SetHighlighted(true) did not set highlight")
	}

	if model.searchQuery != "search" {
		t.Errorf("SetHighlighted did not set search query, got %s", model.searchQuery)
	}

	// Update task
	newTask := archon.Task{
		ID:     "test-task-2",
		Title:  "New Task",
		Status: "doing",
	}

	model.UpdateTask(newTask)
	if model.GetTask().ID != newTask.ID {
		t.Error("UpdateTask did not update the task")
	}

	// Resize
	model.Resize(100)
	if model.GetWidth() != 100 {
		t.Errorf("Resize did not update width, got %d", model.GetWidth())
	}
}

func TestRenderFallback(t *testing.T) {
	task := archon.Task{
		ID:     "test-task-1",
		Title:  "Test Task",
		Status: "todo",
	}

	opts := Options{
		Task:       task,
		Index:      0,
		Width:      20,
		IsSelected: false,
		Context:    &base.ComponentContext{},
	}

	model := NewModel(opts)

	// Test fallback rendering (no dependencies)
	view := model.View()

	if len(view) != 20 {
		t.Errorf("Expected view length 20, got %d", len(view))
	}

	// Should contain status and title
	if !contains(view, "todo") {
		t.Error("View should contain task status")
	}

	if !contains(view, "Test Task") {
		t.Error("View should contain task title")
	}
}

func TestRenderFallbackSelected(t *testing.T) {
	task := archon.Task{
		ID:     "test-task-1",
		Title:  "Test Task",
		Status: "todo",
	}

	opts := Options{
		Task:       task,
		Index:      0,
		Width:      20,
		IsSelected: true,
		Context:    &base.ComponentContext{},
	}

	model := NewModel(opts)

	// Test fallback rendering with selection
	view := model.View()

	// Should show selection indicator
	if !contains(view, ">") {
		t.Error("Selected task should show '>' indicator")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
