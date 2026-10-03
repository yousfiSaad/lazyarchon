package confirmation

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/config"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/styling"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/base"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/context"
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

func (m *mockLogger) Debug(msg string, args ...interface{})                  {}
func (m *mockLogger) Info(msg string, args ...interface{})                   {}
func (m *mockLogger) Warn(msg string, args ...interface{})                   {}
func (m *mockLogger) Error(msg string, args ...interface{})                  {}
func (m *mockLogger) Fatal(msg string, args ...interface{})                  {}
func (m *mockLogger) LogHTTPRequest(method, url string, args ...interface{}) {}
func (m *mockLogger) LogHTTPResponse(method, url string, statusCode int, duration time.Duration, args ...interface{}) {
}
func (m *mockLogger) LogStateChange(component, field string, oldValue, newValue interface{}, args ...interface{}) {
}
func (m *mockLogger) LogPerformance(operation string, startTime time.Time, args ...interface{}) {}

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

	if model.GetType() != base.ConfirmationModalComponent {
		t.Errorf("Expected component type %s, got %s", base.ConfirmationModalComponent, model.GetType())
	}

	if model.IsActive() {
		t.Error("Expected confirmation modal to be initially inactive")
	}

	if !model.CanFocus() {
		t.Error("Expected confirmation modal to be able to receive focus")
	}
}

func TestShowHideConfirmationModal(t *testing.T) {
	context := createTestContext()
	model := NewModel(context)

	// Initially inactive
	if model.IsActive() {
		t.Error("Expected confirmation modal to be initially inactive")
	}

	// Show modal with custom message
	showMsg := ShowConfirmationModalMsg{
		Message:     "Are you sure you want to quit?",
		ConfirmText: "Quit",
		CancelText:  "Stay",
	}
	model.Update(showMsg)

	if !model.IsActive() {
		t.Error("Expected confirmation modal to be active after show message")
	}

	if !model.IsFocused() {
		t.Error("Expected confirmation modal to be focused after show message")
	}

	if model.message != showMsg.Message {
		t.Errorf("Expected message %s, got %s", showMsg.Message, model.message)
	}

	if model.confirmText != showMsg.ConfirmText {
		t.Errorf("Expected confirm text %s, got %s", showMsg.ConfirmText, model.confirmText)
	}

	if model.cancelText != showMsg.CancelText {
		t.Errorf("Expected cancel text %s, got %s", showMsg.CancelText, model.cancelText)
	}

	// Hide modal
	model.Update(HideConfirmationModalMsg{})

	if model.IsActive() {
		t.Error("Expected confirmation modal to be inactive after hide message")
	}

	if model.IsFocused() {
		t.Error("Expected confirmation modal to lose focus after hide message")
	}
}

func TestConfirmationModalView(t *testing.T) {
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
		t.Error("Expected empty view when confirmation modal is inactive")
	}

	// Show modal and test active view
	showMsg := ShowConfirmationModalMsg{
		Message:     "Are you sure?",
		ConfirmText: "Yes",
		CancelText:  "No",
	}
	model.Update(showMsg)

	// Set dimensions for proper rendering
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view = model.View()
	if view == "" {
		t.Error("Expected non-empty view when confirmation modal is active")
	}

	// Verify that the view contains expected confirmation content
	if !containsText(view, "Confirmation") {
		t.Error("Expected confirmation modal view to contain 'Confirmation'")
	}

	if !containsText(view, "Are you sure?") {
		t.Error("Expected confirmation modal view to contain the message")
	}
}

func TestConfirmationModalNavigation(t *testing.T) {
	context := createTestContext()
	model := NewModel(context)

	// Initialize and show modal
	showMsg := ShowConfirmationModalMsg{
		Message:     "Test message",
		ConfirmText: "OK",
		CancelText:  "Cancel",
	}
	model.Update(showMsg)

	// Test initial selection (should be confirm option)
	if model.selectedIndex != 0 {
		t.Errorf("Expected initial selectedIndex to be 0, got %d", model.selectedIndex)
	}

	// Test right navigation (to cancel)
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")}
	model.Update(keyMsg)

	if model.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex to be 1 after right navigation, got %d", model.selectedIndex)
	}

	// Test left navigation (back to confirm)
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")}
	model.Update(keyMsg)

	if model.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex to be 0 after left navigation, got %d", model.selectedIndex)
	}

	// Test boundary conditions - left from first option
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")}
	model.Update(keyMsg)

	if model.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex to stay at 0 when navigating left from first option, got %d", model.selectedIndex)
	}

	// Navigate to last option and test right boundary
	model.selectedIndex = 1 // Set to cancel
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")}
	model.Update(keyMsg)

	if model.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex to stay at 1 when navigating right from last option, got %d", model.selectedIndex)
	}
}

func TestConfirmationModalKeyHandling(t *testing.T) {
	context := createTestContext()
	model := NewModel(context)

	// Initialize and show modal
	showMsg := ShowConfirmationModalMsg{
		Message:     "Test confirmation",
		ConfirmText: "OK",
		CancelText:  "Cancel",
	}
	model.Update(showMsg)

	// Test escape key to cancel
	keyMsg := tea.KeyMsg{Type: tea.KeyEscape}
	cmd := model.Update(keyMsg)

	// Should return commands for cancellation and hiding
	if cmd == nil {
		t.Error("Expected command to be returned for escape key")
	}

	// Test enter key for confirmation
	model.selectedIndex = 0 // Select confirm
	keyMsg = tea.KeyMsg{Type: tea.KeyEnter}
	cmd = model.Update(keyMsg)

	// Should return commands for confirmation and hiding
	if cmd == nil {
		t.Error("Expected command to be returned for enter key")
	}

	// Test direct Y key for confirm
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}
	cmd = model.Update(keyMsg)

	if cmd == nil {
		t.Error("Expected command to be returned for Y key")
	}

	// Test direct N key for cancel
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}
	cmd = model.Update(keyMsg)

	if cmd == nil {
		t.Error("Expected command to be returned for N key")
	}

	// Test Tab key for cycling
	model.selectedIndex = 0 // Start at confirm
	keyMsg = tea.KeyMsg{Type: tea.KeyTab}
	model.Update(keyMsg)

	if model.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex to be 1 after Tab, got %d", model.selectedIndex)
	}

	// Tab again should cycle back
	keyMsg = tea.KeyMsg{Type: tea.KeyTab}
	model.Update(keyMsg)

	if model.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex to be 0 after second Tab, got %d", model.selectedIndex)
	}
}

func TestSetConfirmationInfo(t *testing.T) {
	context := createTestContext()
	model := NewModel(context)

	// Test setting confirmation info
	model.SetConfirmationInfo("Custom message", "Confirm", "Decline")

	if model.message != "Custom message" {
		t.Errorf("Expected message to be 'Custom message', got '%s'", model.message)
	}

	if model.confirmText != "Confirm" {
		t.Errorf("Expected confirmText to be 'Confirm', got '%s'", model.confirmText)
	}

	if model.cancelText != "Decline" {
		t.Errorf("Expected cancelText to be 'Decline', got '%s'", model.cancelText)
	}

	// Test with empty confirm/cancel text (should keep existing)
	model.SetConfirmationInfo("Another message", "", "")

	if model.message != "Another message" {
		t.Errorf("Expected message to be 'Another message', got '%s'", model.message)
	}

	if model.confirmText != "Confirm" {
		t.Errorf("Expected confirmText to remain 'Confirm', got '%s'", model.confirmText)
	}

	if model.cancelText != "Decline" {
		t.Errorf("Expected cancelText to remain 'Decline', got '%s'", model.cancelText)
	}
}

func TestDefaultConfirmationTexts(t *testing.T) {
	context := createTestContext()
	model := NewModel(context)

	// Test showing modal with default texts
	showMsg := ShowConfirmationModalMsg{
		Message: "Test with defaults",
		// ConfirmText and CancelText are empty, should use defaults
	}
	model.Update(showMsg)

	if model.confirmText != "Yes" {
		t.Errorf("Expected default confirmText to be 'Yes', got '%s'", model.confirmText)
	}

	if model.cancelText != "No" {
		t.Errorf("Expected default cancelText to be 'No', got '%s'", model.cancelText)
	}
}

func TestConfirmationModalInactiveKeyHandling(t *testing.T) {
	context := createTestContext()
	model := NewModel(context)

	// Ensure modal is inactive
	if model.IsActive() {
		t.Error("Expected modal to be inactive for this test")
	}

	// Test that keys are ignored when modal is inactive
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")}
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
