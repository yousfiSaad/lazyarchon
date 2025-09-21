package ui

import (
	"testing"

	"github.com/yousfisaad/lazyarchon/internal/archon"
)

// TestBasicKeyNavigation tests basic navigation keys
func TestBasicKeyNavigation(t *testing.T) {
	model := NewModel(createTestConfig())
	model.Window.width = 120
	model.Window.height = 40

	// Setup test tasks
	model.Data.tasks = []archon.Task{
		{ID: "1", Title: "Task 1", Status: "todo"},
		{ID: "2", Title: "Task 2", Status: "doing"},
		{ID: "3", Title: "Task 3", Status: "done"},
	}

	// Test down navigation (j key)
	initialIndex := model.Navigation.selectedIndex
	newModel, _ := model.HandleKeyPress("j")

	if newModel.Navigation.selectedIndex != initialIndex+1 {
		t.Errorf("Expected selectedIndex to be %d, got %d", initialIndex+1, newModel.Navigation.selectedIndex)
	}

	// Test up navigation (k key)
	newModel, _ = newModel.HandleKeyPress("k")
	if newModel.Navigation.selectedIndex != initialIndex {
		t.Errorf("Expected selectedIndex to return to %d, got %d", initialIndex, newModel.Navigation.selectedIndex)
	}

	// Test down arrow
	newModel, _ = newModel.HandleKeyPress("down")
	if newModel.Navigation.selectedIndex != initialIndex+1 {
		t.Error("Expected down arrow to move selection down")
	}

	// Test up arrow
	newModel, _ = newModel.HandleKeyPress("up")
	if newModel.Navigation.selectedIndex != initialIndex {
		t.Error("Expected up arrow to move selection up")
	}
}

// TestJumpNavigation tests jump to first/last navigation
func TestJumpNavigation(t *testing.T) {
	model := NewModel(createTestConfig())
	model.Window.width = 120
	model.Window.height = 40

	// Setup test tasks
	model.Data.tasks = []archon.Task{
		{ID: "1", Title: "Task 1", Status: "todo"},
		{ID: "2", Title: "Task 2", Status: "doing"},
		{ID: "3", Title: "Task 3", Status: "done"},
		{ID: "4", Title: "Task 4", Status: "review"},
		{ID: "5", Title: "Task 5", Status: "todo"},
	}

	// Start from middle
	model.Navigation.selectedIndex = 2

	// Test jump to first (G key)
	newModel, _ := model.HandleKeyPress("G")
	sortedTasks := newModel.GetSortedTasks()
	expectedLastIndex := len(sortedTasks) - 1
	if newModel.Navigation.selectedIndex != expectedLastIndex {
		t.Errorf("Expected G to jump to last task (%d), got %d", expectedLastIndex, newModel.Navigation.selectedIndex)
	}

	// Test jump to first (gg sequence handled by multi-key system)
	newModel, _ = newModel.HandleKeyPress("gg")
	if newModel.Navigation.selectedIndex != 0 {
		t.Error("Expected gg to jump to first task")
	}

	// Test home key
	model.Navigation.selectedIndex = 3
	newModel, _ = model.HandleKeyPress("home")
	if newModel.Navigation.selectedIndex != 0 {
		t.Error("Expected home key to jump to first task")
	}

	// Test end key
	newModel, _ = newModel.HandleKeyPress("end")
	if newModel.Navigation.selectedIndex != expectedLastIndex {
		t.Errorf("Expected end key to jump to last task (%d), got %d", expectedLastIndex, newModel.Navigation.selectedIndex)
	}
}

// TestFastScrolling tests fast scrolling with J/K keys
func TestFastScrolling(t *testing.T) {
	model := NewModel(createTestConfig())
	model.Window.width = 120
	model.Window.height = 40

	// Setup many test tasks to allow fast scrolling
	tasks := make([]archon.Task, 20)
	for i := 0; i < 20; i++ {
		tasks[i] = archon.Task{
			ID:     string(rune('A' + i)),
			Title:  "Task " + string(rune('A'+i)),
			Status: "todo",
		}
	}
	model.Data.tasks = tasks

	// Test fast scroll down (J key) - should move 4 positions
	initialIndex := model.Navigation.selectedIndex
	newModel, _ := model.HandleKeyPress("J")

	expectedIndex := initialIndex + 4
	if newModel.Navigation.selectedIndex != expectedIndex {
		t.Errorf("Expected J to move selection by 4 positions to %d, got %d", expectedIndex, newModel.Navigation.selectedIndex)
	}

	// Test fast scroll up (K key) - should move back 4 positions
	newModel, _ = newModel.HandleKeyPress("K")
	if newModel.Navigation.selectedIndex != initialIndex {
		t.Errorf("Expected K to move selection back to %d, got %d", initialIndex, newModel.Navigation.selectedIndex)
	}
}

// TestPanelSwitching tests switching between left and right panels
func TestPanelSwitching(t *testing.T) {
	model := NewModel(createTestConfig())
	model.Window.width = 120
	model.Window.height = 40

	// Test initial state - should be left panel
	if model.Window.activeView != LeftPanel {
		t.Error("Expected initial active view to be LeftPanel")
	}

	// Test switching to right panel (l key)
	newModel, _ := model.HandleKeyPress("l")
	if newModel.Window.activeView != RightPanel {
		t.Error("Expected l key to switch to RightPanel")
	}

	// Test switching back to left panel (h key)
	newModel, _ = newModel.HandleKeyPress("h")
	if newModel.Window.activeView != LeftPanel {
		t.Error("Expected h key to switch to LeftPanel")
	}
}

// TestProjectMode tests project selection functionality
func TestProjectMode(t *testing.T) {
	model := NewModel(createTestConfig())
	model.Window.width = 120
	model.Window.height = 40

	// Setup test projects
	model.Data.projects = []archon.Project{
		{ID: "proj1", Title: "Project 1"},
		{ID: "proj2", Title: "Project 2"},
	}

	// Test entering project mode (p key)
	newModel, _ := model.HandleKeyPress("p")
	if !newModel.Modals.projectMode.active {
		t.Error("Expected p key to activate project mode")
	}
	if newModel.Modals.projectMode.selectedIndex != 0 {
		t.Error("Expected project mode to start at index 0")
	}

	// Test navigation in project mode
	newModel, _ = newModel.HandleKeyPress("j")
	if newModel.Modals.projectMode.selectedIndex != 1 {
		t.Error("Expected j key to move down in project mode")
	}

	newModel, _ = newModel.HandleKeyPress("k")
	if newModel.Modals.projectMode.selectedIndex != 0 {
		t.Error("Expected k key to move up in project mode")
	}

	// Test selecting project with enter
	newModel, _ = newModel.HandleKeyPress("enter")
	if newModel.Modals.projectMode.active {
		t.Error("Expected enter to close project mode")
	}
	if newModel.Data.selectedProjectID == nil || *newModel.Data.selectedProjectID != "proj1" {
		t.Error("Expected enter to select the first project")
	}

	// Test "Show all tasks" (a key)
	newModel, _ = newModel.HandleKeyPress("a")
	if newModel.Data.selectedProjectID != nil {
		t.Error("Expected a key to clear project selection")
	}

	// Test escaping from project mode
	newModel, _ = newModel.HandleKeyPress("p") // Enter project mode
	newModel, _ = newModel.HandleKeyPress("esc")
	if newModel.Modals.projectMode.active {
		t.Error("Expected esc to close project mode")
	}
}

// TestModalHandling tests various modal interactions
func TestModalHandling(t *testing.T) {
	model := NewModel(createTestConfig())
	model.Window.width = 120
	model.Window.height = 40

	// Setup test tasks
	model.Data.tasks = []archon.Task{
		{ID: "1", Title: "Task 1", Status: "todo"},
	}

	// Test help modal (? key)
	newModel, _ := model.HandleKeyPress("?")
	if !newModel.IsHelpMode() {
		t.Error("Expected ? key to open help modal")
	}

	// Test closing help modal with ?
	newModel, _ = newModel.HandleKeyPress("?")
	if newModel.IsHelpMode() {
		t.Error("Expected ? key to close help modal")
	}

	// Test status change modal (t key)
	newModel, _ = newModel.HandleKeyPress("t")
	if !newModel.IsStatusChangeMode() {
		t.Error("Expected t key to open status change modal")
	}

	// Test task edit modal (e key)
	model.Modals.statusChange.active = false // Close status modal first
	newModel, _ = model.HandleKeyPress("e")
	if !newModel.IsTaskEditModeActive() {
		t.Error("Expected e key to open task edit modal")
	}
}

// TestSearchFunctionality tests search activation and handling
func TestSearchFunctionality(t *testing.T) {
	model := NewModel(createTestConfig())
	model.Window.width = 120
	model.Window.height = 40

	// Setup test tasks
	model.Data.tasks = []archon.Task{
		{ID: "1", Title: "Task One", Status: "todo"},
		{ID: "2", Title: "Another Task", Status: "doing"},
		{ID: "3", Title: "Final Task", Status: "done"},
	}

	// Test activating search (/ key)
	newModel, _ := model.HandleKeyPress("/")
	if !newModel.Data.searchMode {
		t.Error("Expected / key to activate search mode")
	}

	// Test ctrl+f for search
	model.Data.searchMode = false
	newModel, _ = model.HandleKeyPress("ctrl+f")
	if !newModel.Data.searchMode {
		t.Error("Expected ctrl+f to activate search mode")
	}

	// Test search input handling
	newModel, _ = newModel.HandleKeyPress("T") // Type 'T'
	if newModel.Data.searchInput != "T" {
		t.Errorf("Expected search input to be 'T', got '%s'", newModel.Data.searchInput)
	}

	newModel, _ = newModel.HandleKeyPress("a") // Type 'a'
	if newModel.Data.searchInput != "Ta" {
		t.Errorf("Expected search input to be 'Ta', got '%s'", newModel.Data.searchInput)
	}

	// Test backspace in search
	newModel, _ = newModel.HandleKeyPress("backspace")
	if newModel.Data.searchInput != "T" {
		t.Errorf("Expected search input to be 'T' after backspace, got '%s'", newModel.Data.searchInput)
	}

	// Test ctrl+u to clear search input
	newModel, _ = newModel.HandleKeyPress("ctrl+u")
	if newModel.Data.searchInput != "" {
		t.Error("Expected ctrl+u to clear search input")
	}

	// Test committing search with enter
	newModel.Data.searchInput = "Task"
	newModel, _ = newModel.HandleKeyPress("enter")
	if newModel.Data.searchMode {
		t.Error("Expected enter to exit search mode")
	}
	if newModel.Data.searchQuery != "Task" {
		t.Error("Expected search query to be committed")
	}

	// Test canceling search with esc
	newModel, _ = newModel.HandleKeyPress("/") // Activate search again
	newModel.Data.searchInput = "temp"
	newModel, _ = newModel.HandleKeyPress("esc")
	if newModel.Data.searchMode {
		t.Error("Expected esc to cancel search mode")
	}
	if newModel.Data.searchInput != "" {
		t.Error("Expected search input to be cleared on cancel")
	}
}

// TestSortingControls tests sort mode cycling
func TestSortingControls(t *testing.T) {
	model := NewModel(createTestConfig())
	model.Window.width = 120
	model.Window.height = 40

	// Setup test tasks
	model.Data.tasks = []archon.Task{
		{ID: "1", Title: "Task 1", Status: "todo", TaskOrder: 10},
		{ID: "2", Title: "Task 2", Status: "doing", TaskOrder: 5},
	}

	initialSortMode := model.Data.sortMode

	// Test forward sort cycling (s key)
	newModel, _ := model.HandleKeyPress("s")
	if newModel.Data.sortMode == initialSortMode {
		t.Error("Expected s key to change sort mode")
	}

	// Test backward sort cycling (S key)
	newModel, _ = newModel.HandleKeyPress("S")
	if newModel.Data.sortMode != initialSortMode {
		t.Error("Expected S key to cycle back to initial sort mode")
	}

	// Verify sort cycling doesn't work in project mode
	newModel.Modals.projectMode.active = true
	sortModeBeforeProjectMode := newModel.Data.sortMode
	newModel, _ = newModel.HandleKeyPress("s")
	if newModel.Data.sortMode != sortModeBeforeProjectMode {
		t.Error("Expected sort keys to be ignored in project mode")
	}
}

// TestCopyFunctionality tests yank (copy) operations
func TestCopyFunctionality(t *testing.T) {
	model := NewModel(createTestConfig())
	model.Window.width = 120
	model.Window.height = 40

	// Setup test tasks
	model.Data.tasks = []archon.Task{
		{ID: "task-123", Title: "Test Task Title", Status: "todo"},
	}

	// Test copying task ID (y key)
	newModel, _ := model.HandleKeyPress("y")
	// We can't easily test clipboard content, but we can verify the status message
	if newModel.Data.statusMessage == "" {
		t.Error("Expected status message after copying task ID")
	}

	// Test copying task title (Y key)
	newModel, _ = newModel.HandleKeyPress("Y")
	if newModel.Data.statusMessage == "" {
		t.Error("Expected status message after copying task title")
	}

	// Test that copy doesn't work with no tasks
	model.Data.tasks = []archon.Task{}
	newModel, _ = model.HandleKeyPress("y")
	// Should not crash and should not set a copy status message
}

// TestRefreshAndRetry tests refresh and retry functionality
func TestRefreshAndRetry(t *testing.T) {
	model := NewModel(createTestConfig())
	model.Window.width = 120
	model.Window.height = 40

	// Test normal refresh (r key)
	newModel, cmd := model.HandleKeyPress("r")
	if !newModel.Data.loading {
		t.Error("Expected refresh to set loading state")
	}
	if newModel.Data.loadingMessage == "" {
		t.Error("Expected refresh to set loading message")
	}
	if cmd == nil {
		t.Error("Expected refresh to return a command")
	}

	// Test F5 refresh
	model.Data.loading = false
	newModel, cmd = model.HandleKeyPress("F5")
	if !newModel.Data.loading {
		t.Error("Expected F5 to set loading state")
	}
	if cmd == nil {
		t.Error("Expected F5 to return a command")
	}

	// Test retry when there's an error
	model.Data.error = "Connection failed"
	model.Data.loading = false
	newModel, cmd = model.HandleKeyPress("r")
	if newModel.Data.error != "" {
		t.Error("Expected retry to clear error")
	}
	if !newModel.Data.loading {
		t.Error("Expected retry to set loading state")
	}
}

// TestQuitBehavior tests quit and emergency quit functionality
func TestQuitBehavior(t *testing.T) {
	model := NewModel(createTestConfig())
	model.Window.width = 120
	model.Window.height = 40

	// Test normal quit without modals (should show confirmation)
	newModel, cmd := model.HandleKeyPress("q")
	if !newModel.IsConfirmationMode() {
		t.Error("Expected q key to show quit confirmation when no modals active")
	}
	if cmd != nil {
		// Check if it's a quit command by inspecting the command
		// Since we can't directly compare to tea.Quit, we test behavior
		t.Error("Expected q key to not quit immediately, should show confirmation")
	}

	// Test quit with modal active (should close modal, not quit)
	model.SetHelpMode(true)
	newModel, cmd = model.HandleKeyPress("q")
	if newModel.IsHelpMode() {
		t.Error("Expected q key to close help modal")
	}
	if cmd != nil {
		// Should not return a quit command when closing modal
		t.Error("Expected q key to close modal, not return a command")
	}

	// Test emergency quit (ctrl+c)
	newModel, cmd = model.HandleKeyPress("ctrl+c")
	if cmd == nil {
		t.Error("Expected ctrl+c to return a quit command")
	}
}

// TestBoundaryConditions tests edge cases in navigation
func TestBoundaryConditions(t *testing.T) {
	model := NewModel(createTestConfig())
	model.Window.width = 120
	model.Window.height = 40

	// Test navigation with no tasks
	model.Data.tasks = []archon.Task{}

	// Should not crash with empty task list
	newModel, _ := model.HandleKeyPress("j")
	if newModel.Navigation.selectedIndex != 0 {
		t.Error("Expected selectedIndex to remain 0 with empty task list")
	}

	newModel, _ = newModel.HandleKeyPress("k")
	if newModel.Navigation.selectedIndex != 0 {
		t.Error("Expected selectedIndex to remain 0 with empty task list")
	}

	// Test with single task
	model.Data.tasks = []archon.Task{
		{ID: "1", Title: "Only Task", Status: "todo"},
	}

	// Try to go up from first task
	model.Navigation.selectedIndex = 0
	newModel, _ = model.HandleKeyPress("k")
	if newModel.Navigation.selectedIndex != 0 {
		t.Error("Expected selectedIndex to remain 0 when at first task")
	}

	// Try to go down from last (only) task
	newModel, _ = newModel.HandleKeyPress("j")
	if newModel.Navigation.selectedIndex != 0 {
		t.Error("Expected selectedIndex to remain 0 when at last task")
	}
}

// TestInvalidKeys tests handling of unrecognized keys
func TestInvalidKeys(t *testing.T) {
	model := NewModel(createTestConfig())
	model.Window.width = 120
	model.Window.height = 40

	originalState := model

	// Test that invalid keys don't change state
	newModel, cmd := model.HandleKeyPress("invalid_key_xyz")

	if cmd != nil {
		t.Error("Expected invalid key to return nil command")
	}

	// State should be unchanged (checking a few key fields)
	if newModel.Navigation.selectedIndex != originalState.Navigation.selectedIndex {
		t.Error("Expected invalid key to not change navigation state")
	}
	if newModel.Window.activeView != originalState.Window.activeView {
		t.Error("Expected invalid key to not change active view")
	}
}

// TestMultiKeySequences tests handling of multi-key sequences like "gg"
func TestMultiKeySequences(t *testing.T) {
	model := NewModel(createTestConfig())
	model.Window.width = 120
	model.Window.height = 40

	// Setup test tasks
	model.Data.tasks = []archon.Task{
		{ID: "1", Title: "Task 1", Status: "todo"},
		{ID: "2", Title: "Task 2", Status: "doing"},
		{ID: "3", Title: "Task 3", Status: "done"},
	}

	// Start from last task
	model.Navigation.selectedIndex = 2

	// Test "gg" sequence to jump to first
	newModel, _ := model.HandleKeyPress("gg")
	if newModel.Navigation.selectedIndex != 0 {
		t.Error("Expected 'gg' sequence to jump to first task")
	}
}