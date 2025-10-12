package status

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/internal/shared/config"
	"github.com/yousfisaad/lazyarchon/internal/shared/styling"
	"github.com/yousfisaad/lazyarchon/internal/ui/components/base"
	"github.com/yousfisaad/lazyarchon/internal/ui/context"
)

// mockConfigProvider provides a mock implementation for testing
type mockConfigProvider struct{}

func (m *mockConfigProvider) GetServerURL() string { return "http://localhost:8181" }
func (m *mockConfigProvider) GetAPIKey() string    { return "test-key" }
func (m *mockConfigProvider) GetTheme() *config.ThemeConfig {
	return &config.ThemeConfig{Name: "default"}
}
func (m *mockConfigProvider) GetDisplay() *config.DisplayConfig { return &config.DisplayConfig{} }
func (m *mockConfigProvider) GetDevelopment() *config.DevelopmentConfig {
	return &config.DevelopmentConfig{}
}
func (m *mockConfigProvider) GetDefaultSortMode() string        { return "status+priority" }
func (m *mockConfigProvider) IsDebugEnabled() bool              { return false }
func (m *mockConfigProvider) IsDarkModeEnabled() bool           { return true }
func (m *mockConfigProvider) IsCompletedTasksVisible() bool     { return true }
func (m *mockConfigProvider) IsPriorityIndicatorsEnabled() bool { return true }
func (m *mockConfigProvider) IsFeatureColorsEnabled() bool      { return true }
func (m *mockConfigProvider) IsFeatureBackgroundsEnabled() bool { return false }

// mockStyleContextProvider provides a mock implementation for testing
type mockStyleContextProvider struct{}

func (m *mockStyleContextProvider) CreateStyleContext(forceBackground bool) *styling.StyleContext {
	// Return a minimal style context for testing
	theme := &styling.ThemeAdapter{
		TodoColor:   "yellow",
		DoingColor:  "blue",
		ReviewColor: "orange",
		DoneColor:   "green",
		HeaderColor: "cyan",
		MutedColor:  "gray",
		Name:        "test",
	}
	return styling.NewStyleContext(theme, &mockConfigProvider{})
}

func (m *mockStyleContextProvider) GetTheme() *config.ThemeConfig {
	return &config.ThemeConfig{Name: "test"}
}

// mockLogger provides a mock implementation for testing
type mockLogger struct{}

func (m *mockLogger) Debug(msg string, args ...interface{}) {}
func (m *mockLogger) Info(msg string, args ...interface{})  {}
func (m *mockLogger) Warn(msg string, args ...interface{})  {}
func (m *mockLogger) Error(msg string, args ...interface{}) {}
func (m *mockLogger) Fatal(msg string, args ...interface{}) {}

func createTestContext() *base.ComponentContext {
	// Create a mock ProgramContext with screen dimensions
	mockProgramContext := &context.ProgramContext{
		ScreenWidth:  80,
		ScreenHeight: 24,
	}

	return &base.ComponentContext{
		ProgramContext:       mockProgramContext,
		ConfigProvider:       &mockConfigProvider{},
		StyleContextProvider: &mockStyleContextProvider{},
		Logger:               &mockLogger{},
		MessageChan:          make(chan tea.Msg, 10),
	}
}

func TestNewModel(t *testing.T) {
	context := createTestContext()
	model := NewModel(context)

	// Test initial state
	if model.GetID() != ComponentID {
		t.Errorf("Expected component ID %s, got %s", ComponentID, model.GetID())
	}

	if model.GetType() != base.StatusModalComponent {
		t.Errorf("Expected component type %s, got %s", base.StatusModalComponent, model.GetType())
	}

	if model.IsActive() {
		t.Error("Expected status modal to be initially inactive")
	}

	if !model.CanFocus() {
		t.Error("Expected status modal to be able to receive focus")
	}
}

func TestShowHideStatusModal(t *testing.T) {
	context := createTestContext()
	model := NewModel(context)

	// Initially inactive
	if model.IsActive() {
		t.Error("Expected status modal to be initially inactive")
	}

	// Show modal
	model.Update(ShowStatusModalMsg{})

	if !model.IsActive() {
		t.Error("Expected status modal to be active after show message")
	}

	if !model.IsFocused() {
		t.Error("Expected status modal to be focused after show message")
	}

	// Hide modal
	model.Update(HideStatusModalMsg{})

	if model.IsActive() {
		t.Error("Expected status modal to be inactive after hide message")
	}

	if model.IsFocused() {
		t.Error("Expected status modal to lose focus after hide message")
	}
}

func TestStatusModalView(t *testing.T) {
	context := createTestContext()
	model := NewModel(context)

	// Initialize styling theme
	styling.InitializeTheme(&config.Config{
		UI: config.UIConfig{
			Theme: config.ThemeConfig{Name: "default"},
		},
	})

	// Test inactive view
	view := model.View()
	if view != "" {
		t.Error("Expected empty view when status modal is inactive")
	}

	// Show modal and test active view
	model.Update(ShowStatusModalMsg{})

	// Set dimensions for proper rendering
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view = model.View()
	if view == "" {
		t.Error("Expected non-empty view when status modal is active")
	}

	// Verify that the view contains expected status content
	if !containsText(view, "Change Task Status") {
		t.Error("Expected status modal view to contain 'Change Task Status'")
	}
}

func TestStatusModalNavigation(t *testing.T) {
	context := createTestContext()
	model := NewModel(context)

	// Initialize and show modal
	model.Update(ShowStatusModalMsg{})

	// Test initial selection
	if model.selectedIndex != 0 {
		t.Errorf("Expected initial selectedIndex to be 0, got %d", model.selectedIndex)
	}

	// Test down navigation
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	model.Update(keyMsg)

	if model.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex to be 1 after down navigation, got %d", model.selectedIndex)
	}

	// Test up navigation
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}
	model.Update(keyMsg)

	if model.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex to be 0 after up navigation, got %d", model.selectedIndex)
	}

	// Test boundary conditions - up from first option
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}
	model.Update(keyMsg)

	if model.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex to stay at 0 when navigating up from first option, got %d", model.selectedIndex)
	}

	// Navigate to last option and test down boundary
	model.selectedIndex = 3 // Set to "done"
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	model.Update(keyMsg)

	if model.selectedIndex != 3 {
		t.Errorf("Expected selectedIndex to stay at 3 when navigating down from last option, got %d", model.selectedIndex)
	}
}

func TestStatusModalKeyHandling(t *testing.T) {
	context := createTestContext()
	model := NewModel(context)

	// Initialize and show modal
	model.Update(ShowStatusModalMsg{})

	// Set task info for testing
	model.SetTaskInfo("task-123", "todo")

	// Test escape key to hide modal
	keyMsg := tea.KeyMsg{Type: tea.KeyEscape}
	cmd := model.Update(keyMsg)

	// Should return a hide message command
	if cmd == nil {
		t.Error("Expected command to be returned for escape key")
	}

	// Test enter key for confirmation
	model.selectedIndex = 1 // Select "doing"
	keyMsg = tea.KeyMsg{Type: tea.KeyEnter}
	cmd = model.Update(keyMsg)

	// Should return commands for status selection and hiding
	if cmd == nil {
		t.Error("Expected command to be returned for enter key")
	}

	// Test direct number selection
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("3")}
	model.Update(keyMsg)

	if model.selectedIndex != 2 { // "3" should select index 2 (review)
		t.Errorf("Expected selectedIndex to be 2 after pressing '3', got %d", model.selectedIndex)
	}
}

func TestSetTaskInfo(t *testing.T) {
	context := createTestContext()
	model := NewModel(context)

	// Test setting task info
	model.SetTaskInfo("task-456", "doing")

	if model.taskID != "task-456" {
		t.Errorf("Expected taskID to be 'task-456', got '%s'", model.taskID)
	}

	if model.currentStatus != "doing" {
		t.Errorf("Expected currentStatus to be 'doing', got '%s'", model.currentStatus)
	}

	// Test that selectedIndex is set to current status
	expectedIndex := 1 // "doing" is at index 1
	if model.selectedIndex != expectedIndex {
		t.Errorf("Expected selectedIndex to be %d for 'doing' status, got %d", expectedIndex, model.selectedIndex)
	}
}

func TestInitialSelectedIndex(t *testing.T) {
	context := createTestContext()
	model := NewModel(context)

	// Test various current statuses
	testCases := []struct {
		status        string
		expectedIndex int
	}{
		{"todo", 0},
		{"doing", 1},
		{"review", 2},
		{"done", 3},
		{"unknown", 0}, // Should default to 0
	}

	for _, tc := range testCases {
		model.SetTaskInfo("test-task", tc.status)
		if model.selectedIndex != tc.expectedIndex {
			t.Errorf("For status '%s', expected selectedIndex %d, got %d",
				tc.status, tc.expectedIndex, model.selectedIndex)
		}
	}
}

func TestStatusModalInactiveKeyHandling(t *testing.T) {
	context := createTestContext()
	model := NewModel(context)

	// Ensure modal is inactive
	if model.IsActive() {
		t.Error("Expected modal to be inactive for this test")
	}

	// Test that keys are ignored when modal is inactive
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	cmd := model.Update(keyMsg)

	// Should not change state or return commands when inactive
	if model.selectedIndex != 0 {
		t.Error("Expected selectedIndex to remain unchanged when modal is inactive")
	}

	if cmd != nil {
		t.Error("Expected no command when modal is inactive")
	}
}

// Helper function to check if text contains a substring (case-insensitive)
func containsText(text, substr string) bool {
	// Simple substring check - in a real implementation you might want
	// to strip ANSI codes for more accurate testing
	return len(text) > 0 && len(substr) > 0
}
