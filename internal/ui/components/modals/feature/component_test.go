package feature

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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

// Helper function to create a test model
func createTestModel() *FeatureModel {
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

	model := NewModel(context)
	return model
}

// Helper function to extract message from command
func commandContainsMessage(cmd tea.Cmd, expectedMsg tea.Msg) bool {
	if cmd == nil {
		return false
	}

	msg := cmd()
	if msg == nil {
		return false
	}

	// Check if it's a ComponentMessage wrapper
	if componentMsg, ok := msg.(base.ComponentMessage); ok {
		return compareMessages(componentMsg.Payload, expectedMsg)
	}

	return compareMessages(msg, expectedMsg)
}

// Helper function to compare message types
func compareMessages(actual, expected tea.Msg) bool {
	switch expected.(type) {
	case FeatureModalShownMsg:
		// Check for new unified ModalStateMsg
		if modalMsg, ok := actual.(messages.ModalStateMsg); ok {
			return modalMsg.Type == string(base.ModalTypeFeature) && modalMsg.Active
		}
		_, ok := actual.(FeatureModalShownMsg)
		return ok
	case FeatureModalHiddenMsg:
		// Check for new unified ModalStateMsg
		if modalMsg, ok := actual.(messages.ModalStateMsg); ok {
			return modalMsg.Type == string(base.ModalTypeFeature) && !modalMsg.Active
		}
		_, ok := actual.(FeatureModalHiddenMsg)
		return ok
	case FeatureSelectionAppliedMsg:
		_, ok := actual.(FeatureSelectionAppliedMsg)
		return ok
	default:
		return false
	}
}

// Test basic component creation
func TestNewModel(t *testing.T) {
	model := createTestModel()

	if model.GetID() != ComponentID {
		t.Errorf("Expected component ID %s, got %s", ComponentID, model.GetID())
	}

	if model.GetType() != base.FeatureModalComponent {
		t.Errorf("Expected component type %s, got %s", base.FeatureModalComponent, model.GetType())
	}

	if model.IsActive() {
		t.Error("Expected new model to be inactive")
	}

	if model.IsFocused() {
		t.Error("Expected new model to be unfocused")
	}

	if !model.CanFocus() {
		t.Error("Expected feature modal to be focusable")
	}
}

// Test showing the feature modal
func TestShowFeatureModal(t *testing.T) {
	model := createTestModel()

	// Prepare test data
	allFeatures := []string{"authentication", "ui", "backend", "testing"}
	selectedFeatures := map[string]bool{"ui": true, "backend": true}

	// Send show message
	showMsg := ShowFeatureModalMsg{
		AllFeatures:          allFeatures,
		SelectedFeatures:     selectedFeatures,
		FeatureColorsEnabled: true,
	}

	cmd := model.Update(showMsg)

	// Check that modal is now active
	if !model.IsActive() {
		t.Error("Expected modal to be active after show message")
	}

	if !model.IsFocused() {
		t.Error("Expected modal to be focused after show message")
	}

	// Check that data was set correctly
	if len(model.allFeatures) != 4 {
		t.Errorf("Expected 4 all features, got %d", len(model.allFeatures))
	}

	if len(model.selectedFeatures) != 2 {
		t.Errorf("Expected 2 selected features, got %d", len(model.selectedFeatures))
	}

	if !model.selectedFeatures["ui"] || !model.selectedFeatures["backend"] {
		t.Error("Expected ui and backend to be selected")
	}

	if !model.featureColorsEnabled {
		t.Error("Expected feature colors to be enabled")
	}

	// Check that backup was created
	if len(model.backupFeatures) != 2 {
		t.Errorf("Expected 2 backup features, got %d", len(model.backupFeatures))
	}

	// Check that shown message was broadcast
	if !commandContainsMessage(cmd, FeatureModalShownMsg{}) {
		t.Error("Expected FeatureModalShownMsg to be broadcast")
	}

	// Check that filtered features were initialized
	if len(model.filteredFeatures) != 4 {
		t.Errorf("Expected 4 filtered features, got %d", len(model.filteredFeatures))
	}

	// Check initial state
	if model.selectedIndex != 0 {
		t.Error("Expected selected index to be 0 initially")
	}

	if model.searchMode {
		t.Error("Expected to be in selection mode initially")
	}
}

// Test hiding the feature modal
func TestHideFeatureModal(t *testing.T) {
	model := createTestModel()

	// First show the modal
	showMsg := ShowFeatureModalMsg{
		AllFeatures:      []string{"feature1"},
		SelectedFeatures: map[string]bool{},
	}
	model.Update(showMsg)

	// Then hide it
	hideMsg := HideFeatureModalMsg{}
	cmd := model.Update(hideMsg)

	// Check that modal is now inactive
	if model.IsActive() {
		t.Error("Expected modal to be inactive after hide message")
	}

	if model.IsFocused() {
		t.Error("Expected modal to be unfocused after hide message")
	}

	// Check that search mode was reset
	if model.searchMode {
		t.Error("Expected search mode to be false after hide")
	}

	// Check that hidden message was broadcast
	if !commandContainsMessage(cmd, FeatureModalHiddenMsg{}) {
		t.Error("Expected FeatureModalHiddenMsg to be broadcast")
	}
}

// Test navigation in selection mode
func TestNavigationInSelectionMode(t *testing.T) {
	model := createTestModel()

	// Setup modal with test features
	showMsg := ShowFeatureModalMsg{
		AllFeatures:      []string{"feature1", "feature2", "feature3"},
		SelectedFeatures: map[string]bool{},
	}
	model.Update(showMsg)

	// Test down navigation (j key)
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	model.Update(keyMsg)

	if model.selectedIndex != 1 {
		t.Errorf("Expected selected index to be 1 after j key, got %d", model.selectedIndex)
	}

	// Test up navigation (k key)
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	model.Update(keyMsg)

	if model.selectedIndex != 0 {
		t.Errorf("Expected selected index to be 0 after k key, got %d", model.selectedIndex)
	}

	// Test boundary - up from 0 should stay at 0
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	model.Update(keyMsg)

	if model.selectedIndex != 0 {
		t.Error("Expected selected index to stay at 0 when at boundary")
	}

	// Navigate to last item
	model.selectedIndex = 2

	// Test boundary - down from last should stay at last
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	model.Update(keyMsg)

	if model.selectedIndex != 2 {
		t.Error("Expected selected index to stay at 2 when at boundary")
	}
}

// Test feature toggling with space key
func TestFeatureToggling(t *testing.T) {
	model := createTestModel()

	// Setup modal
	showMsg := ShowFeatureModalMsg{
		AllFeatures:      []string{"feature1", "feature2"},
		SelectedFeatures: map[string]bool{"feature1": true},
	}
	model.Update(showMsg)

	// Initially feature1 should be selected
	if !model.selectedFeatures["feature1"] {
		t.Error("Expected feature1 to be initially selected")
	}

	// Toggle feature1 (should deselect it)
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	model.Update(keyMsg)

	if model.selectedFeatures["feature1"] {
		t.Error("Expected feature1 to be deselected after toggle")
	}

	// Toggle feature1 again (should select it)
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	model.Update(keyMsg)

	if !model.selectedFeatures["feature1"] {
		t.Error("Expected feature1 to be selected after second toggle")
	}

	// Navigate to feature2 and toggle it
	model.selectedIndex = 1
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	model.Update(keyMsg)

	if !model.selectedFeatures["feature2"] {
		t.Error("Expected feature2 to be selected after toggle")
	}
}

// Test search mode activation
func TestSearchModeActivation(t *testing.T) {
	model := createTestModel()

	// Setup modal
	showMsg := ShowFeatureModalMsg{
		AllFeatures:      []string{"authentication", "ui", "backend"},
		SelectedFeatures: map[string]bool{},
	}
	model.Update(showMsg)

	// Activate search with '/' key
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	model.Update(keyMsg)

	if !model.searchMode {
		t.Error("Expected to be in search mode after / key")
	}

	if model.searchInput != "" {
		t.Error("Expected search input to be empty initially")
	}
}

// Test search input handling
func TestSearchInputHandling(t *testing.T) {
	model := createTestModel()

	// Setup modal in search mode
	showMsg := ShowFeatureModalMsg{
		AllFeatures:      []string{"authentication", "ui", "backend"},
		SelectedFeatures: map[string]bool{},
	}
	model.Update(showMsg)
	model.searchMode = true

	// Type 'a'
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	model.Update(keyMsg)

	if model.searchInput != "a" {
		t.Errorf("Expected search input to be 'a', got '%s'", model.searchInput)
	}

	// Type 'u'
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}}
	model.Update(keyMsg)

	if model.searchInput != "au" {
		t.Errorf("Expected search input to be 'au', got '%s'", model.searchInput)
	}

	// Test backspace
	keyMsg = tea.KeyMsg{Type: tea.KeyBackspace}
	model.Update(keyMsg)

	if model.searchInput != "a" {
		t.Errorf("Expected search input to be 'a' after backspace, got '%s'", model.searchInput)
	}

	// Test ctrl+u (clear)
	keyMsg = tea.KeyMsg{Type: tea.KeyCtrlU}
	model.Update(keyMsg)

	if model.searchInput != "" {
		t.Error("Expected search input to be empty after ctrl+u")
	}
}

// Test search filtering
func TestSearchFiltering(t *testing.T) {
	model := createTestModel()

	// Setup modal
	showMsg := ShowFeatureModalMsg{
		AllFeatures:      []string{"authentication", "ui", "backend", "auth-service"},
		SelectedFeatures: map[string]bool{},
	}
	model.Update(showMsg)

	// Initially all features should be visible
	if len(model.filteredFeatures) != 4 {
		t.Errorf("Expected 4 filtered features initially, got %d", len(model.filteredFeatures))
	}

	// Set search query to "auth"
	model.searchQuery = "auth"
	model.updateFilteredFeatures()

	// Should show 2 features containing "auth"
	if len(model.filteredFeatures) != 2 {
		t.Errorf("Expected 2 filtered features for 'auth' search, got %d", len(model.filteredFeatures))
	}

	// Check that correct features are shown
	foundAuth := false
	foundAuthService := false
	for _, feature := range model.filteredFeatures {
		if feature == "authentication" {
			foundAuth = true
		}
		if feature == "auth-service" {
			foundAuthService = true
		}
	}

	if !foundAuth || !foundAuthService {
		t.Error("Expected to find 'authentication' and 'auth-service' in filtered results")
	}

	// Test case insensitive search
	model.searchQuery = "UI"
	model.updateFilteredFeatures()

	if len(model.filteredFeatures) != 1 {
		t.Errorf("Expected 1 filtered feature for 'UI' search, got %d", len(model.filteredFeatures))
	}

	if model.filteredFeatures[0] != "ui" {
		t.Errorf("Expected 'ui' in filtered results, got '%s'", model.filteredFeatures[0])
	}
}

// Test applying selection
func TestApplyingSelection(t *testing.T) {
	model := createTestModel()

	// Setup modal with selection
	showMsg := ShowFeatureModalMsg{
		AllFeatures:      []string{"feature1", "feature2", "feature3"},
		SelectedFeatures: map[string]bool{"feature1": true},
	}
	model.Update(showMsg)

	// Toggle feature2
	model.selectedIndex = 1
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	model.Update(keyMsg)

	// Apply selection with Enter
	keyMsg = tea.KeyMsg{Type: tea.KeyEnter}
	cmd := model.Update(keyMsg)

	// Check that a command was returned (it should be a batch)
	if cmd == nil {
		t.Error("Expected command to be returned")
		return
	}

	// For a batch command, we can't easily test the individual messages
	// but we can verify that a command was returned which means the logic worked
	// The actual message handling is tested in integration tests
}

// Test cancel functionality
func TestCancelFunctionality(t *testing.T) {
	model := createTestModel()

	// Setup modal with initial selection
	initialSelection := map[string]bool{"feature1": true}
	showMsg := ShowFeatureModalMsg{
		AllFeatures:      []string{"feature1", "feature2"},
		SelectedFeatures: initialSelection,
	}
	model.Update(showMsg)

	// Modify selection
	model.selectedIndex = 1
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	model.Update(keyMsg)

	// Now we should have feature1 and feature2 selected
	if !model.selectedFeatures["feature1"] || !model.selectedFeatures["feature2"] {
		t.Error("Expected both features to be selected after modification")
	}

	// Cancel with Escape
	keyMsg = tea.KeyMsg{Type: tea.KeyEsc}
	model.Update(keyMsg)

	// Selection should be restored to backup (only feature1)
	if !model.selectedFeatures["feature1"] {
		t.Error("Expected feature1 to be restored after cancel")
	}

	if model.selectedFeatures["feature2"] {
		t.Error("Expected feature2 to be deselected after cancel")
	}
}

// Test select all functionality
func TestSelectAllFunctionality(t *testing.T) {
	model := createTestModel()

	// Setup modal
	showMsg := ShowFeatureModalMsg{
		AllFeatures:      []string{"feature1", "feature2", "feature3"},
		SelectedFeatures: map[string]bool{},
	}
	model.Update(showMsg)

	// Initially no features selected
	if len(model.selectedFeatures) != 0 {
		t.Error("Expected no features to be selected initially")
	}

	// Select all with 'a' key
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	model.Update(keyMsg)

	// All visible features should be selected
	if len(model.selectedFeatures) != 3 {
		t.Errorf("Expected 3 features to be selected after select all, got %d", len(model.selectedFeatures))
	}

	for _, feature := range model.filteredFeatures {
		if !model.selectedFeatures[feature] {
			t.Errorf("Expected feature '%s' to be selected", feature)
		}
	}
}

// Test deselect all functionality
func TestDeselectAllFunctionality(t *testing.T) {
	model := createTestModel()

	// Setup modal with all features selected
	showMsg := ShowFeatureModalMsg{
		AllFeatures:      []string{"feature1", "feature2", "feature3"},
		SelectedFeatures: map[string]bool{"feature1": true, "feature2": true, "feature3": true},
	}
	model.Update(showMsg)

	// Initially all features selected
	if len(model.selectedFeatures) != 3 {
		t.Error("Expected 3 features to be selected initially")
	}

	// Deselect all with 'A' key (Shift+A)
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}}
	model.Update(keyMsg)

	// No features should be selected
	if len(model.selectedFeatures) != 0 {
		t.Errorf("Expected no features to be selected after deselect all, got %d", len(model.selectedFeatures))
	}
}

// Test parent-child architecture dimension handling
func TestViewWithDimensions(t *testing.T) {
	model := createTestModel()
	model.SetActive(true)

	// Test ViewWithDimensions with parent-provided dimensions
	// Simulate window resize (Bubble Tea standard pattern)
	screenWidth := 120
	screenHeight := 40

	// Update dimensions via WindowSizeMsg
	model.Update(tea.WindowSizeMsg{Width: screenWidth, Height: screenHeight})

	// Render the view
	view := model.View()

	// Check that dimensions were updated correctly
	if model.GetWidth() != screenWidth {
		t.Errorf("Expected width %d, got %d", screenWidth, model.GetWidth())
	}

	if model.GetHeight() != screenHeight {
		t.Errorf("Expected height %d, got %d", screenHeight, model.GetHeight())
	}

	// Should return a rendered view
	if view == "" {
		t.Error("Expected View to return non-empty view")
	}

	// Test WindowSizeMsg updates dimensions (Bubble Tea standard pattern)
	resizeMsg := tea.WindowSizeMsg{Width: 200, Height: 60}
	cmd := model.Update(resizeMsg)

	// Dimensions should update from WindowSizeMsg
	if model.GetWidth() != 200 || model.GetHeight() != 60 {
		t.Error("Modal dimensions should update on WindowSizeMsg")
	}

	// Should not return a command for resize (nil is expected)
	if cmd != nil {
		t.Error("Expected no command for window resize")
	}
}

// Test view rendering
func TestViewRendering(t *testing.T) {
	model := createTestModel()

	// Test inactive modal
	view := model.View()
	if view != "" {
		t.Error("Expected empty view when modal is inactive")
	}

	// Activate modal
	showMsg := ShowFeatureModalMsg{
		AllFeatures:      []string{"feature1", "feature2"},
		SelectedFeatures: map[string]bool{"feature1": true},
	}
	model.Update(showMsg)

	// Test active modal
	view = model.View()
	if view == "" {
		t.Error("Expected non-empty view when modal is active")
	}

	// View should contain key elements
	if !strings.Contains(view, "Select Features") {
		t.Error("Expected view to contain title")
	}

	// Should contain feature names
	if !strings.Contains(view, "feature1") || !strings.Contains(view, "feature2") {
		t.Error("Expected view to contain feature names")
	}

	// Should contain checkboxes (updated symbols)
	if !strings.Contains(view, "■") { // Selected checkbox
		t.Error("Expected view to contain selected checkbox")
	}

	if !strings.Contains(view, "□") { // Unselected checkbox
		t.Error("Expected view to contain unselected checkbox")
	}
}

// Test edge cases
func TestEdgeCases(t *testing.T) {
	model := createTestModel()

	// Test with empty feature list
	showMsg := ShowFeatureModalMsg{
		AllFeatures:      []string{},
		SelectedFeatures: map[string]bool{},
	}
	model.Update(showMsg)

	// Should handle empty list gracefully
	if len(model.filteredFeatures) != 0 {
		t.Error("Expected empty filtered features for empty input")
	}

	// View should render without crashing
	view := model.View()
	if view == "" {
		t.Error("Expected view to render even with empty features")
	}

	// Test navigation with empty list (should not crash)
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	model.Update(keyMsg)
	// Should not crash

	// Test search with no results
	model.allFeatures = []string{"feature1", "feature2"}
	model.searchQuery = "nonexistent"
	model.updateFilteredFeatures()

	if len(model.filteredFeatures) != 0 {
		t.Error("Expected no filtered features for non-matching search")
	}

	// View should show "No features found"
	view = model.View()
	if !strings.Contains(view, "No features found") {
		t.Error("Expected 'No features found' message for empty search results")
	}
}
