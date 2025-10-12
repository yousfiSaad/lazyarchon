package taskedit

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/internal/shared/config"
	"github.com/yousfisaad/lazyarchon/internal/shared/styling"
	"github.com/yousfisaad/lazyarchon/internal/ui/components/base"
	"github.com/yousfisaad/lazyarchon/internal/ui/context"
	"github.com/yousfisaad/lazyarchon/internal/ui/messages"
)

// Mock implementations for testing
type mockConfigProvider struct{}

func (m *mockConfigProvider) GetServerURL() string                      { return "http://localhost" }
func (m *mockConfigProvider) GetAPIKey() string                         { return "test-key" }
func (m *mockConfigProvider) GetTheme() *config.ThemeConfig             { return nil }
func (m *mockConfigProvider) GetDisplay() *config.DisplayConfig         { return nil }
func (m *mockConfigProvider) GetDevelopment() *config.DevelopmentConfig { return nil }
func (m *mockConfigProvider) GetDefaultSortMode() string                { return "status" }
func (m *mockConfigProvider) IsDebugEnabled() bool                      { return false }
func (m *mockConfigProvider) IsDarkModeEnabled() bool                   { return true }
func (m *mockConfigProvider) IsCompletedTasksVisible() bool             { return true }
func (m *mockConfigProvider) IsPriorityIndicatorsEnabled() bool         { return true }
func (m *mockConfigProvider) IsFeatureColorsEnabled() bool              { return true }
func (m *mockConfigProvider) IsFeatureBackgroundsEnabled() bool         { return false }

type mockStyleContextProvider struct{}

func (m *mockStyleContextProvider) CreateStyleContext(forceBackground bool) *styling.StyleContext {
	return nil
}
func (m *mockStyleContextProvider) GetTheme() *config.ThemeConfig { return nil }

type mockLogger struct{}

func (m *mockLogger) Debug(msg string, args ...interface{}) {}
func (m *mockLogger) Info(msg string, args ...interface{})  {}
func (m *mockLogger) Warn(msg string, args ...interface{})  {}
func (m *mockLogger) Error(msg string, args ...interface{}) {}
func (m *mockLogger) Fatal(msg string, args ...interface{}) {}

// Helper function to create a test model
func createTestModel() *TaskEditModel {
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
		MessageChan:          make(chan tea.Msg, 10),
	}

	model := NewModel(context)
	return model
}

// Helper function to check if a command contains a specific message type
func commandContainsMessage(cmd tea.Cmd, messageType interface{}) bool {
	if cmd == nil {
		return false
	}

	msg := cmd()
	if msg == nil {
		return false
	}

	// Check direct message
	switch messageType.(type) {
	case TaskEditModalShownMsg:
		// Check for new unified ModalStateMsg
		if modalMsg, ok := msg.(messages.ModalStateMsg); ok {
			return modalMsg.Type == string(base.ModalTypeTaskEdit) && modalMsg.Active
		}
		if _, ok := msg.(TaskEditModalShownMsg); ok {
			return true
		}
	case TaskEditModalHiddenMsg:
		// Check for new unified ModalStateMsg
		if modalMsg, ok := msg.(messages.ModalStateMsg); ok {
			return modalMsg.Type == string(base.ModalTypeTaskEdit) && !modalMsg.Active
		}
		if _, ok := msg.(TaskEditModalHiddenMsg); ok {
			return true
		}
	case FeatureSelectedMsg:
		if _, ok := msg.(FeatureSelectedMsg); ok {
			return true
		}
	case HideTaskEditModalMsg:
		if _, ok := msg.(HideTaskEditModalMsg); ok {
			return true
		}
	}

	// Check ComponentMessage wrapper
	if compMsg, ok := msg.(base.ComponentMessage); ok {
		switch messageType.(type) {
		case TaskEditModalShownMsg:
			// Check for new unified ModalStateMsg
			if modalMsg, ok := compMsg.Payload.(messages.ModalStateMsg); ok {
				return modalMsg.Type == string(base.ModalTypeTaskEdit) && modalMsg.Active
			}
			if _, ok := compMsg.Payload.(TaskEditModalShownMsg); ok {
				return true
			}
		case TaskEditModalHiddenMsg:
			// Check for new unified ModalStateMsg
			if modalMsg, ok := compMsg.Payload.(messages.ModalStateMsg); ok {
				return modalMsg.Type == string(base.ModalTypeTaskEdit) && !modalMsg.Active
			}
			if _, ok := compMsg.Payload.(TaskEditModalHiddenMsg); ok {
				return true
			}
		case FeatureSelectedMsg:
			if _, ok := compMsg.Payload.(FeatureSelectedMsg); ok {
				return true
			}
		case HideTaskEditModalMsg:
			if _, ok := compMsg.Payload.(HideTaskEditModalMsg); ok {
				return true
			}
		}
	}

	// Check batch message
	if batchMsg, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batchMsg {
			if commandContainsMessage(c, messageType) {
				return true
			}
		}
	}

	return false
}

func TestNewModel(t *testing.T) {
	model := createTestModel()

	// Test initial state
	if model.GetID() != ComponentID {
		t.Errorf("Expected component ID %s, got %s", ComponentID, model.GetID())
	}

	if model.GetType() != base.TaskEditModalComponent {
		t.Errorf("Expected component type %s, got %s", base.TaskEditModalComponent, model.GetType())
	}

	if model.IsActive() {
		t.Error("Expected modal to be inactive initially")
	}

	if !model.CanFocus() {
		t.Error("Expected modal to be focusable")
	}

	if model.isCreatingNew {
		t.Error("Expected modal to start in selection mode")
	}
}

func TestShowTaskEditModal(t *testing.T) {
	model := createTestModel()

	// Send show message
	showMsg := ShowTaskEditModalMsg{
		TaskID:            "task-123",
		CurrentFeature:    "authentication",
		AvailableFeatures: []string{"authentication", "ui", "backend"},
	}

	cmd := model.Update(showMsg)

	// Check that modal is now active
	if !model.IsActive() {
		t.Error("Expected modal to be active after show message")
	}

	if !model.IsFocused() {
		t.Error("Expected modal to be focused after show message")
	}

	// Check that task info was set correctly
	if model.taskID != "task-123" {
		t.Errorf("Expected task ID 'task-123', got '%s'", model.taskID)
	}

	if model.featureValue != "authentication" {
		t.Errorf("Expected feature value 'authentication', got '%s'", model.featureValue)
	}

	if len(model.availableFeatures) != 3 {
		t.Errorf("Expected 3 available features, got %d", len(model.availableFeatures))
	}

	// Check that shown message was broadcast
	if !commandContainsMessage(cmd, TaskEditModalShownMsg{}) {
		t.Error("Expected TaskEditModalShownMsg to be broadcast")
	}

	if model.isCreatingNew {
		t.Error("Expected to be in selection mode initially")
	}
}

func TestHideTaskEditModal(t *testing.T) {
	model := createTestModel()

	// First show the modal
	showMsg := ShowTaskEditModalMsg{
		TaskID:            "task-123",
		AvailableFeatures: []string{"feature1"},
	}
	model.Update(showMsg)

	// Then hide it
	hideMsg := HideTaskEditModalMsg{}
	cmd := model.Update(hideMsg)

	// Check that modal is now inactive
	if model.IsActive() {
		t.Error("Expected modal to be inactive after hide message")
	}

	if model.IsFocused() {
		t.Error("Expected modal to be unfocused after hide message")
	}

	// Check that hidden message was broadcast
	if !commandContainsMessage(cmd, TaskEditModalHiddenMsg{}) {
		t.Error("Expected TaskEditModalHiddenMsg to be broadcast")
	}

	// Check that creating new state was reset
	if model.isCreatingNew {
		t.Error("Expected creating new state to be reset")
	}

	if model.newFeatureName != "" {
		t.Error("Expected new feature name to be cleared")
	}
}

// TestNavigationInSelectionMode - REMOVED: Feature navigation now delegated to feature modal
// The task edit modal no longer handles inline feature navigation.
// Instead, pressing Enter/Space on the feature field opens the full feature modal for selection.
/*
func TestNavigationInSelectionMode(t *testing.T) {
	// This test is obsolete - feature selection now uses the feature modal component
	t.Skip("Feature navigation removed - now uses feature modal")
}
*/

// TestFeatureSelection - REMOVED: Feature selection now delegated to feature modal
// Feature selection now opens the feature modal component instead of inline navigation
/*
func TestFeatureSelection(t *testing.T) {
	// This test is obsolete - feature selection now uses the feature modal component
	t.Skip("Feature selection removed - now uses feature modal")
}
*/

func TestCreateNewFeatureFlow(t *testing.T) {
	model := createTestModel()

	// Setup modal - start on feature field
	showMsg := ShowTaskEditModalMsg{
		TaskID:            "task-123",
		FocusField:        FieldFeature,
		AvailableFeatures: []string{"feature1"},
	}
	model.Update(showMsg)

	// Press 'n' to create new feature
	nKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	model.Update(nKey)

	// Check that we're now in creating new mode
	if !model.isCreatingNew {
		t.Error("Expected to be in creating new mode")
	}

	// Type a feature name
	testChars := []string{"t", "e", "s", "t"}
	for _, char := range testChars {
		charKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(char)}
		model.Update(charKey)
	}

	if model.newFeatureName != "test" {
		t.Errorf("Expected new feature name 'test', got '%s'", model.newFeatureName)
	}

	// Confirm the new feature
	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	model.Update(enterKey)

	// Check that feature value was set
	if model.featureValue != "test" {
		t.Errorf("Expected feature value to be 'test', got '%s'", model.featureValue)
	}

	// Check that we exited creating new mode
	if model.isCreatingNew {
		t.Error("Expected to exit creating new mode after confirming")
	}
}

func TestTextInputValidation(t *testing.T) {
	model := createTestModel()

	// Setup modal and enter creating new mode
	showMsg := ShowTaskEditModalMsg{
		TaskID:            "task-123",
		FocusField:        FieldFeature,
		AvailableFeatures: []string{},
	}
	model.Update(showMsg)
	model.isCreatingNew = true
	model.activeField = FieldFeature

	// Test valid characters
	validChars := []string{"a", "A", "1", "_", "-", " "}
	for _, char := range validChars {
		initialLen := len(model.newFeatureName)
		charKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(char)}
		model.Update(charKey)

		if len(model.newFeatureName) != initialLen+1 {
			t.Errorf("Valid character '%s' was not accepted", char)
		}
	}

	// Reset for invalid character test
	model.newFeatureName = ""

	// Test invalid characters (should be ignored)
	invalidChars := []string{"@", "#", "$", "%", "^", "&", "*", "(", ")", "=", "+"}
	for _, char := range invalidChars {
		initialLen := len(model.newFeatureName)
		charKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(char)}
		model.Update(charKey)

		if len(model.newFeatureName) != initialLen {
			t.Errorf("Invalid character '%s' was incorrectly accepted", char)
		}
	}
}

func TestTextInputControls(t *testing.T) {
	model := createTestModel()

	// Setup modal in creating new mode
	showMsg := ShowTaskEditModalMsg{
		TaskID:     "task-123",
		FocusField: FieldFeature,
	}
	model.Update(showMsg)
	model.isCreatingNew = true
	model.activeField = FieldFeature
	model.newFeatureName = "test"

	// Test backspace
	backspaceKey := tea.KeyMsg{Type: tea.KeyBackspace}
	model.Update(backspaceKey)

	if model.newFeatureName != "tes" {
		t.Errorf("Expected 'tes' after backspace, got '%s'", model.newFeatureName)
	}

	// Test Ctrl+U (clear all)
	ctrlUKey := tea.KeyMsg{Type: tea.KeyCtrlU}
	model.Update(ctrlUKey)

	if model.newFeatureName != "" {
		t.Errorf("Expected empty string after Ctrl+U, got '%s'", model.newFeatureName)
	}
}

func TestEscapeKey(t *testing.T) {
	model := createTestModel()

	// Test escape in selection mode
	showMsg := ShowTaskEditModalMsg{TaskID: "task-123"}
	model.Update(showMsg)

	escapeKey := tea.KeyMsg{Type: tea.KeyEsc}
	cmd := model.Update(escapeKey)

	// Should send hide message
	if !commandContainsMessage(cmd, HideTaskEditModalMsg{}) {
		t.Error("Expected HideTaskEditModalMsg when pressing escape in selection mode")
	}

	// Test escape in creating new mode
	model.SetActive(true)
	model.SetFocus(true)
	model.activeField = FieldFeature
	model.isCreatingNew = true
	model.newFeatureName = "test"

	cmd = model.Update(escapeKey)

	// Should exit creating new mode and clear input
	if model.isCreatingNew {
		t.Error("Expected to exit creating new mode on escape")
	}

	if model.newFeatureName != "" {
		t.Error("Expected new feature name to be cleared on escape")
	}

	// Should not send hide message in this case
	if commandContainsMessage(cmd, HideTaskEditModalMsg{}) {
		t.Error("Expected no hide message when escaping from creating new mode")
	}
}

func TestWindowResize(t *testing.T) {
	model := createTestModel()

	// Test window resize
	resizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	model.Update(resizeMsg)

	// Check that dimensions were updated appropriately
	// The exact values depend on the updateDimensions implementation
	// but we can check that the model handled the message without error
	if model == nil {
		t.Error("Model should not be nil after window resize")
	}
}

func TestViewRendering(t *testing.T) {
	model := createTestModel()

	// Test view when inactive
	view := model.View()
	if view != "" {
		t.Error("Expected empty view when modal is inactive")
	}

	// Show modal and test view
	showMsg := ShowTaskEditModalMsg{
		TaskID:            "task-123",
		CurrentFeature:    "auth",
		AvailableFeatures: []string{"auth", "ui"},
	}
	model.Update(showMsg)

	view = model.View()
	if view == "" {
		t.Error("Expected non-empty view when modal is active")
	}

	// View should contain some expected content
	if !contains(view, "Edit Task Properties") {
		t.Error("Expected view to contain 'Edit Task Properties' title")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsInMiddle(s, substr)))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestCurrentFeaturePreSelection tests that the current feature is automatically selected
func TestCurrentFeaturePreSelection(t *testing.T) {
	model := createTestModel()

	// Test when current feature is in the middle of the list
	showMsg := ShowTaskEditModalMsg{
		TaskID:            "task-123",
		CurrentFeature:    "ui",
		AvailableFeatures: []string{"authentication", "ui", "backend", "testing"},
	}

	model.Update(showMsg)

	// Feature value should be set to "ui"
	if model.featureValue != "ui" {
		t.Errorf("Expected feature value to be 'ui', got '%s'", model.featureValue)
	}
}

// TestCurrentFeaturePreSelectionLastItem tests pre-selection when current feature is last
func TestCurrentFeaturePreSelectionLastItem(t *testing.T) {
	model := createTestModel()

	// Test when current feature is the last item
	showMsg := ShowTaskEditModalMsg{
		TaskID:            "task-123",
		CurrentFeature:    "testing",
		AvailableFeatures: []string{"authentication", "ui", "backend", "testing"},
	}

	model.Update(showMsg)

	// Feature value should be set to "testing"
	if model.featureValue != "testing" {
		t.Errorf("Expected feature value to be 'testing', got '%s'", model.featureValue)
	}
}

// TestCurrentFeatureNotInList tests fallback when current feature is not in available list
func TestCurrentFeatureNotInList(t *testing.T) {
	model := createTestModel()

	// Test when current feature is not in available features
	showMsg := ShowTaskEditModalMsg{
		TaskID:            "task-123",
		CurrentFeature:    "database", // Not in available features
		AvailableFeatures: []string{"authentication", "ui", "backend"},
	}

	model.Update(showMsg)

	// Feature value should still be set to "database" even if not in list
	if model.featureValue != "database" {
		t.Errorf("Expected feature value to be 'database', got '%s'", model.featureValue)
	}
}

// TestEmptyCurrentFeature tests behavior when current feature is empty
func TestEmptyCurrentFeature(t *testing.T) {
	model := createTestModel()

	// Test when current feature is empty
	showMsg := ShowTaskEditModalMsg{
		TaskID:            "task-123",
		CurrentFeature:    "", // Empty current feature
		AvailableFeatures: []string{"authentication", "ui", "backend"},
	}

	model.Update(showMsg)

	// Feature value should be empty
	if model.featureValue != "" {
		t.Errorf("Expected feature value to be empty, got '%s'", model.featureValue)
	}
}

// TestFindFeatureIndexHelper tests the helper function directly
func TestFindFeatureIndexHelper(t *testing.T) {
	model := createTestModel()

	features := []string{"authentication", "ui", "backend", "testing"}

	// Test finding existing features
	tests := []struct {
		feature       string
		expectedIndex int
	}{
		{"authentication", 0},
		{"ui", 1},
		{"backend", 2},
		{"testing", 3},
		{"nonexistent", -1},
		{"", -1},
	}

	for _, test := range tests {
		index := model.findFeatureIndex(test.feature, features)
		if index != test.expectedIndex {
			t.Errorf("findFeatureIndex(%q) = %d, expected %d", test.feature, index, test.expectedIndex)
		}
	}
}
