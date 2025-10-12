package statusfilter

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/internal/archon"
	"github.com/yousfisaad/lazyarchon/internal/shared/config"
	"github.com/yousfisaad/lazyarchon/internal/shared/styling"
	"github.com/yousfisaad/lazyarchon/internal/ui/components/base"
	"github.com/yousfisaad/lazyarchon/internal/ui/context"
	"github.com/yousfisaad/lazyarchon/internal/ui/messages"
)

// Mock dependencies for testing
type mockLogger struct{}

func (m *mockLogger) Debug(msg string, args ...interface{}) {}
func (m *mockLogger) Info(msg string, args ...interface{})  {}
func (m *mockLogger) Warn(msg string, args ...interface{})  {}
func (m *mockLogger) Error(msg string, args ...interface{}) {}
func (m *mockLogger) Fatal(msg string, args ...interface{}) {}

type mockConfigProvider struct{}

func (m *mockConfigProvider) GetServerURL() string                      { return "http://localhost:8080" }
func (m *mockConfigProvider) GetAPIKey() string                         { return "test-key" }
func (m *mockConfigProvider) IsDebugEnabled() bool                      { return false }
func (m *mockConfigProvider) GetTheme() *config.ThemeConfig             { return nil }
func (m *mockConfigProvider) GetDisplay() *config.DisplayConfig         { return nil }
func (m *mockConfigProvider) GetDevelopment() *config.DevelopmentConfig { return nil }
func (m *mockConfigProvider) GetDefaultSortMode() string                { return "status+priority" }
func (m *mockConfigProvider) IsDarkModeEnabled() bool                   { return false }
func (m *mockConfigProvider) IsCompletedTasksVisible() bool             { return true }
func (m *mockConfigProvider) IsPriorityIndicatorsEnabled() bool         { return true }
func (m *mockConfigProvider) IsFeatureColorsEnabled() bool              { return true }
func (m *mockConfigProvider) IsFeatureBackgroundsEnabled() bool         { return false }

type mockStyleContextProvider struct{}

func (m *mockStyleContextProvider) CreateStyleContext(forceBackground bool) *styling.StyleContext {
	return nil
}
func (m *mockStyleContextProvider) GetTheme() *config.ThemeConfig { return nil }

func createTestStatusFilterModal() *Model {
	// Create a mock ProgramContext with screen dimensions
	mockProgramContext := &context.ProgramContext{
		ScreenWidth:  80,
		ScreenHeight: 24,
	}

	context := &base.ComponentContext{
		ProgramContext:       mockProgramContext,
		ConfigProvider:       &mockConfigProvider{},
		StyleContextProvider: &mockStyleContextProvider{},
		Logger:               &mockLogger{},
		MessageChan:          make(chan tea.Msg, 100),
	}

	modal := NewModel(context)
	return modal
}

func TestStatusFilterModalCreation(t *testing.T) {
	modal := createTestStatusFilterModal()

	if modal == nil {
		t.Fatal("Expected modal to be created")
	}

	if modal.IsActive() {
		t.Error("Expected modal to start inactive")
	}

	if !modal.CanFocus() {
		t.Error("Expected modal to be focusable")
	}
}

func TestStatusFilterModalShow(t *testing.T) {
	modal := createTestStatusFilterModal()

	// Test showing modal
	currentStatuses := map[string]bool{
		archon.TaskStatusTodo:  true,
		archon.TaskStatusDoing: false,
		archon.TaskStatusDone:  true,
	}

	showMsg := ShowStatusFilterModalMsg{
		CurrentStatuses: currentStatuses,
	}

	cmd := modal.Update(showMsg)

	if !modal.IsActive() {
		t.Error("Expected modal to be active after show message")
	}

	// Verify status state was initialized
	if !modal.selectedStatuses[archon.TaskStatusTodo] {
		t.Error("Expected todo status to be selected")
	}
	if modal.selectedStatuses[archon.TaskStatusDoing] {
		t.Error("Expected doing status to be unselected")
	}
	if !modal.selectedStatuses[archon.TaskStatusDone] {
		t.Error("Expected done status to be selected")
	}

	// Verify backup was created
	if !modal.backupStatuses[archon.TaskStatusTodo] {
		t.Error("Expected backup to contain todo status")
	}

	if cmd == nil {
		t.Error("Expected command to be returned for show message")
	}

	// Execute command and verify message type - check for unified ModalStateMsg
	msg := cmd()
	if modalMsg, ok := msg.(messages.ModalStateMsg); ok {
		if modalMsg.Type != string(base.ModalTypeStatusFilter) || !modalMsg.Active {
			t.Error("Expected ModalStateMsg with StatusFilter type and Active=true")
		}
	} else if _, ok := msg.(StatusFilterModalShownMsg); !ok {
		t.Error("Expected StatusFilterModalShownMsg or ModalStateMsg")
	}
}

func TestStatusFilterModalHide(t *testing.T) {
	modal := createTestStatusFilterModal()
	modal.SetActive(true)

	hideMsg := HideStatusFilterModalMsg{}
	cmd := modal.Update(hideMsg)

	if modal.IsActive() {
		t.Error("Expected modal to be inactive after hide message")
	}

	if cmd == nil {
		t.Error("Expected command to be returned for hide message")
	}

	// Execute command and verify message type - check for unified ModalStateMsg
	msg := cmd()
	if modalMsg, ok := msg.(messages.ModalStateMsg); ok {
		if modalMsg.Type != string(base.ModalTypeStatusFilter) || modalMsg.Active {
			t.Error("Expected ModalStateMsg with StatusFilter type and Active=false")
		}
	} else if _, ok := msg.(StatusFilterModalHiddenMsg); !ok {
		t.Error("Expected StatusFilterModalHiddenMsg or ModalStateMsg")
	}
}

func TestStatusFilterModalToggleStatus(t *testing.T) {
	modal := createTestStatusFilterModal()
	modal.SetActive(true)

	// Initialize with some statuses
	modal.selectedStatuses = map[string]bool{
		archon.TaskStatusTodo:  true,
		archon.TaskStatusDoing: false,
	}
	modal.filteredStatuses = []string{archon.TaskStatusTodo, archon.TaskStatusDoing}
	modal.selectedIndex = 0 // First status (todo)

	// Test toggling selection
	selectMsg := tea.KeyMsg{Type: tea.KeySpace}
	 modal.Update(selectMsg)

	// Todo should now be unselected
	if modal.selectedStatuses[archon.TaskStatusTodo] {
		t.Error("Expected todo status to be toggled off")
	}

	// Toggle again
	 modal.Update(selectMsg)

	// Todo should now be selected again
	if !modal.selectedStatuses[archon.TaskStatusTodo] {
		t.Error("Expected todo status to be toggled back on")
	}
}

func TestStatusFilterModalNavigation(t *testing.T) {
	modal := createTestStatusFilterModal()
	modal.SetActive(true)
	modal.filteredStatuses = []string{archon.TaskStatusTodo, archon.TaskStatusDoing, archon.TaskStatusDone}
	modal.selectedIndex = 0

	// Test down navigation
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	 modal.Update(downMsg)

	if modal.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex to be 1, got %d", modal.selectedIndex)
	}

	// Test up navigation
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	 modal.Update(upMsg)

	if modal.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex to be 0, got %d", modal.selectedIndex)
	}

	// Test up navigation at boundary (should not go below 0)
	 modal.Update(upMsg)

	if modal.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex to stay at 0, got %d", modal.selectedIndex)
	}
}

func TestStatusFilterModalSearch(t *testing.T) {
	modal := createTestStatusFilterModal()
	modal.SetActive(true)

	// Enter search mode
	searchMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	 modal.Update(searchMsg)

	if !modal.searchMode {
		t.Error("Expected to enter search mode")
	}

	// Type search query
	typeMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t', 'o'}}
	 modal.Update(typeMsg)

	if modal.searchInput != "to" {
		t.Errorf("Expected search input to be 'to', got '%s'", modal.searchInput)
	}

	// Should filter to only "todo"
	expectedFiltered := []string{archon.TaskStatusTodo}
	if len(modal.filteredStatuses) != len(expectedFiltered) {
		t.Errorf("Expected %d filtered statuses, got %d", len(expectedFiltered), len(modal.filteredStatuses))
	}

	// Commit search
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	 modal.Update(enterMsg)

	if modal.searchMode {
		t.Error("Expected to exit search mode after commit")
	}

	if modal.searchQuery != "to" {
		t.Errorf("Expected search query to be 'to', got '%s'", modal.searchQuery)
	}
}

func TestStatusFilterModalSelectAll(t *testing.T) {
	modal := createTestStatusFilterModal()
	modal.SetActive(true)
	modal.filteredStatuses = []string{archon.TaskStatusTodo, archon.TaskStatusDoing}
	modal.selectedStatuses = map[string]bool{
		archon.TaskStatusTodo:  false,
		archon.TaskStatusDoing: false,
	}

	// Select all
	selectAllMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	 modal.Update(selectAllMsg)

	// All filtered statuses should be selected
	for _, status := range modal.filteredStatuses {
		if !modal.selectedStatuses[status] {
			t.Errorf("Expected status %s to be selected", status)
		}
	}
}

func TestStatusFilterModalSelectNone(t *testing.T) {
	modal := createTestStatusFilterModal()
	modal.SetActive(true)
	modal.filteredStatuses = []string{archon.TaskStatusTodo, archon.TaskStatusDoing}
	modal.selectedStatuses = map[string]bool{
		archon.TaskStatusTodo:  true,
		archon.TaskStatusDoing: true,
	}

	// Select none
	selectNoneMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	 modal.Update(selectNoneMsg)

	// All filtered statuses should be unselected
	for _, status := range modal.filteredStatuses {
		if modal.selectedStatuses[status] {
			t.Errorf("Expected status %s to be unselected", status)
		}
	}
}

func TestStatusFilterModalCancel(t *testing.T) {
	modal := createTestStatusFilterModal()
	modal.SetActive(true)

	// Set up initial state
	originalStatuses := map[string]bool{
		archon.TaskStatusTodo:  true,
		archon.TaskStatusDoing: false,
	}
	modal.selectedStatuses = make(map[string]bool)
	for k, v := range originalStatuses {
		modal.selectedStatuses[k] = v
	}
	modal.backupStatuses = make(map[string]bool)
	for k, v := range originalStatuses {
		modal.backupStatuses[k] = v
	}

	// Modify current selection
	modal.selectedStatuses[archon.TaskStatusTodo] = false
	modal.selectedStatuses[archon.TaskStatusDoing] = true

	// Cancel
	cancelMsg := tea.KeyMsg{Type: tea.KeyEsc}
	 modal.Update(cancelMsg)

	// Should restore backup state
	if modal.selectedStatuses[archon.TaskStatusTodo] != originalStatuses[archon.TaskStatusTodo] {
		t.Error("Expected todo status to be restored from backup")
	}
	if modal.selectedStatuses[archon.TaskStatusDoing] != originalStatuses[archon.TaskStatusDoing] {
		t.Error("Expected doing status to be restored from backup")
	}

	if modal.IsActive() {
		t.Error("Expected modal to be hidden after cancel")
	}
}

func TestStatusFilterModalApply(t *testing.T) {
	modal := createTestStatusFilterModal()
	modal.SetActive(true)

	// Set up status selection
	modal.selectedStatuses = map[string]bool{
		archon.TaskStatusTodo:  true,
		archon.TaskStatusDoing: false,
		archon.TaskStatusDone:  true,
	}

	// Apply selection
	applyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	cmd := modal.Update(applyMsg)

	if modal.IsActive() {
		t.Error("Expected modal to be hidden after apply")
	}

	if cmd == nil {
		t.Error("Expected command to be returned for apply")
	}

	// Execute command and verify message type
	msg := cmd()

	// Should return a batch command, so we need to handle that
	if batchCmd, ok := msg.(tea.BatchMsg); ok {
		// Execute the batch and check the messages
		var appliedMsg *StatusFilterAppliedMsg
		var hiddenMsg *StatusFilterModalHiddenMsg

		var modalStateMsg *messages.ModalStateMsg
		for _, cmdFunc := range batchCmd {
			result := cmdFunc()
			switch typedMsg := result.(type) {
			case StatusFilterAppliedMsg:
				appliedMsg = &typedMsg
			case StatusFilterModalHiddenMsg:
				hiddenMsg = &typedMsg
			case messages.ModalStateMsg:
				modalStateMsg = &typedMsg
			}
		}

		if appliedMsg == nil {
			t.Error("Expected StatusFilterAppliedMsg in batch")
		} else {
			// Verify selected statuses
			if !appliedMsg.SelectedStatuses[archon.TaskStatusTodo] {
				t.Error("Expected todo status to be selected in applied message")
			}
			if appliedMsg.SelectedStatuses[archon.TaskStatusDoing] {
				t.Error("Expected doing status to be unselected in applied message")
			}
			if !appliedMsg.SelectedStatuses[archon.TaskStatusDone] {
				t.Error("Expected done status to be selected in applied message")
			}
		}

		// Check for either old or new hide message
		if hiddenMsg == nil && modalStateMsg == nil {
			t.Error("Expected StatusFilterModalHiddenMsg or ModalStateMsg in batch")
		}
		if modalStateMsg != nil {
			if modalStateMsg.Type != string(base.ModalTypeStatusFilter) || modalStateMsg.Active {
				t.Error("Expected ModalStateMsg with StatusFilter type and Active=false")
			}
		}
	} else {
		t.Error("Expected batch command for apply action")
	}
}

func TestStatusFilterModalView(t *testing.T) {
	modal := createTestStatusFilterModal()

	// Test inactive view
	view := modal.View()
	if view != "" {
		t.Error("Expected empty view when modal is inactive")
	}

	// Test active view
	modal.SetActive(true)
	view = modal.View()
	if view == "" {
		t.Error("Expected non-empty view when modal is active")
	}

	// View should contain status filter elements
	expectedElements := []string{"Status Filter", "Search:", "Selected:"}
	for _, element := range expectedElements {
		if !contains(view, element) {
			t.Errorf("Expected view to contain '%s'", element)
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			contains(s[1:], substr) ||
			(len(s) > 0 && s[:len(substr)] == substr))
}
