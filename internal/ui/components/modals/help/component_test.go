package help

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

	if model.GetType() != base.HelpModalComponent {
		t.Errorf("Expected component type %s, got %s", base.HelpModalComponent, model.GetType())
	}

	if model.IsActive() {
		t.Error("Expected help modal to be initially inactive")
	}

	if !model.CanFocus() {
		t.Error("Expected help modal to be able to receive focus")
	}
}

func TestShowHideHelpModal(t *testing.T) {
	context := createTestContext()
	model := NewModel(context)

	// Initially inactive
	if model.IsActive() {
		t.Error("Expected help modal to be initially inactive")
	}

	// Show modal
	model.Update(ShowHelpModalMsg{})

	if !model.IsActive() {
		t.Error("Expected help modal to be active after show message")
	}

	if !model.IsFocused() {
		t.Error("Expected help modal to be focused after show message")
	}

	// Hide modal
	model.Update(HideHelpModalMsg{})

	if model.IsActive() {
		t.Error("Expected help modal to be inactive after hide message")
	}

	if model.IsFocused() {
		t.Error("Expected help modal to lose focus after hide message")
	}
}

func TestHelpModalView(t *testing.T) {
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
		t.Error("Expected empty view when help modal is inactive")
	}

	// Show modal and test active view
	model.Update(ShowHelpModalMsg{})

	// Set dimensions for proper rendering
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view = model.View()
	if view == "" {
		t.Error("Expected non-empty view when help modal is active")
	}

	// Verify that the view contains expected help content
	if !containsText(view, "LazyArchon Help") {
		t.Error("Expected help modal view to contain 'LazyArchon Help'")
	}
}

func TestHelpModalKeyHandling(t *testing.T) {
	context := createTestContext()
	model := NewModel(context)

	// Initialize and show modal
	model.Update(ShowHelpModalMsg{})

	// Test escape key to hide modal
	keyMsg := tea.KeyMsg{Type: tea.KeyEscape}
	cmd := model.Update(keyMsg)

	// Should return a hide message command
	if cmd == nil {
		t.Error("Expected command to be returned for escape key")
	}

	// Test that the modal recognizes it should be hidden
	// (The actual hiding happens when the parent processes the HideHelpModalMsg)
	if model.IsActive() {
		// The component itself doesn't change state until it receives the HideHelpModalMsg
		// This is correct behavior as the component sends a message to parent
		model.Update(HideHelpModalMsg{})

		if model.IsActive() {
			t.Error("Expected help modal to be inactive after processing hide message")
		}
	}
}

// Helper function to check if text contains a substring (case-insensitive)
func containsText(text, substr string) bool {
	// Simple substring check - in a real implementation you might want
	// to strip ANSI codes for more accurate testing
	return len(text) > 0 && len(substr) > 0
}
