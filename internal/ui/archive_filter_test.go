package ui

import (
	"testing"

	"github.com/yousfisaad/lazyarchon/v2/internal/shared/interfaces"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/utils/keys"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/context"
)

// archivedFilterTestTasks returns a fixed task set: one live task and one
// archived task, both open so only the archived flag distinguishes them.
func archivedFilterTestTasks() []interfaces.Task {
	return []interfaces.Task{
		{ID: "t-live", Title: "Live task", Status: "todo", Priority: 3},
		{ID: "t-archived", Title: "Archived task", Status: "todo", Priority: 3, Archived: true},
	}
}

func TestHandleArchivedFilterKey_TogglesVisibility(t *testing.T) {
	model := createTestModel(t)
	model.programContext.SetTasks(archivedFilterTestTasks())

	// Hidden by default
	if ids := sortedTaskIDs(&model); ids["t-archived"] {
		t.Errorf("archived task must be hidden before 'x', got %v", ids)
	}

	// 'x' reveals it
	if _, handled := model.handleTaskKey(keys.KeyX); !handled {
		t.Fatal("expected 'x' key to be handled")
	}
	if !model.programContext.ShowArchivedTasks {
		t.Error("expected ShowArchivedTasks to be enabled after 'x'")
	}
	if ids := sortedTaskIDs(&model); !ids["t-archived"] {
		t.Errorf("archived task must be visible after 'x', got %v", ids)
	}

	// 'x' again hides it
	if _, handled := model.handleTaskKey(keys.KeyX); !handled {
		t.Fatal("expected second 'x' key to be handled")
	}
	if model.programContext.ShowArchivedTasks {
		t.Error("expected ShowArchivedTasks to be disabled after second 'x'")
	}
	if ids := sortedTaskIDs(&model); ids["t-archived"] {
		t.Errorf("archived task must be hidden again, got %v", ids)
	}
}

func TestHandleArchivedFilterKey_InertInProjectView(t *testing.T) {
	model := createTestModel(t)
	model.programContext.SetTasks(archivedFilterTestTasks())
	model.uiState.CurrentViewMode = context.ProjectViewMode

	if _, handled := model.handleArchivedFilterKey(keys.KeyX); handled {
		t.Error("'x' must not be handled in project view")
	}
	if _, handled := model.handleTaskArchiveKey(keys.KeyXCap); handled {
		t.Error("'X' must not be handled in project view")
	}
}

// The 'X' handler returns the UpdateTaskWithRequest command without touching
// pending state (archiving is reversible, no confirmation flow). The command
// is NOT executed here: TaskClient may be nil under createTestModel.
func TestHandleTaskArchiveKey_ReturnsUpdateCommand(t *testing.T) {
	model := createTestModel(t)
	model.programContext.SetTasks(archivedFilterTestTasks())

	cmd, handled := model.handleTaskKey(keys.KeyXCap)
	if !handled {
		t.Fatal("expected 'X' key to be handled")
	}
	if cmd == nil {
		t.Fatal("expected 'X' handler to return a command")
	}
	if model.pendingDeleteTaskID != "" {
		t.Error("'X' must not set pendingDeleteTaskID (no confirmation flow)")
	}
}

func TestHandleTaskArchiveKey_NoTasksNoOp(t *testing.T) {
	model := createTestModel(t)

	if _, handled := model.handleTaskArchiveKey(keys.KeyXCap); handled {
		t.Error("'X' must not be handled with no tasks loaded")
	}
}

// sortedTaskIDs returns the ID set of the model's current sorted task list.
func sortedTaskIDs(model *MainModel) map[string]bool {
	tasks := model.GetSortedTasks()
	ids := make(map[string]bool, len(tasks))
	for _, task := range tasks {
		ids[task.ID] = true
	}
	return ids
}
