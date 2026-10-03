package worktree

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/config"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/styling"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/base"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/context"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/messages"
)

// Mock dependencies for testing
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

// createTestModel builds a modal over mock dependencies.
func createTestModel() *WorktreeModel {
	mockProgramContext := &context.ProgramContext{
		ScreenWidth:  80,
		ScreenHeight: 24,
	}

	componentContext := &base.ComponentContext{
		ProgramContext:       mockProgramContext,
		ConfigProvider:       &mockConfigProvider{},
		StyleContextProvider: &mockStyleContextProvider{},
		Logger:               &mockLogger{},
		MessageChan:          make(chan tea.Msg, 100),
	}

	return NewModel(componentContext)
}

// showModal activates the modal with the given selection, mirroring how
// handleWorktreeSelectionKey drives it.
func showModal(model *WorktreeModel, all []string, selected map[string]bool) {
	model.Update(ShowWorktreeModalMsg{AllWorktrees: all, SelectedWorktrees: selected})
}

func TestShowModalActivatesAndRendersOptions(t *testing.T) {
	model := createTestModel()
	showModal(model, []string{"wt-a", "wt-b"}, map[string]bool{"wt-a": true})

	if !model.IsActive() {
		t.Fatal("modal must be active after ShowWorktreeModalMsg")
	}

	view := model.View()
	if !strings.Contains(view, "Select Worktrees") {
		t.Errorf("view missing title, got:\n%s", view)
	}
	if !strings.Contains(view, "@wt-a") || !strings.Contains(view, "@wt-b") {
		t.Errorf("view must list @-prefixed worktrees, got:\n%s", view)
	}
}

func TestSpaceTogglesCurrentWorktree(t *testing.T) {
	model := createTestModel()
	showModal(model, []string{"wt-a"}, map[string]bool{"wt-a": true})

	// Space toggles the highlighted worktree off.
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

	if model.selectedWorktrees["wt-a"] {
		t.Error("space must toggle the current worktree off")
	}

	// Space toggles it back on.
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

	if !model.selectedWorktrees["wt-a"] {
		t.Error("space must toggle the current worktree back on")
	}
}

func TestEnterAppliesSelection(t *testing.T) {
	model := createTestModel()
	showModal(model, []string{"wt-a", "wt-b"}, map[string]bool{"wt-a": true})

	cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if commandContainsMessage(cmd) {
		return
	}
	t.Error("enter must broadcast WorktreeSelectionAppliedMsg")
}

func TestEscapeRestoresBackupSelection(t *testing.T) {
	model := createTestModel()
	showModal(model, []string{"wt-a", "wt-b"}, map[string]bool{"wt-a": true})

	// Deselect everything, then cancel: the backup selection must survive.
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})
	if model.selectedWorktrees["wt-a"] {
		t.Fatal("deselect-all must clear the selection before cancel")
	}

	model.Update(tea.KeyMsg{Type: tea.KeyEscape})

	if !model.selectedWorktrees["wt-a"] {
		t.Errorf("escape must restore the backup selection, got selected=%v", model.selectedWorktrees)
	}
}

func TestSmartToggleSelectsAllVisible(t *testing.T) {
	model := createTestModel()
	showModal(model, []string{"wt-a", "wt-b"}, nil)

	// 'a' selects all visible worktrees.
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	if !model.selectedWorktrees["wt-a"] || !model.selectedWorktrees["wt-b"] {
		t.Errorf("smart toggle must select all visible worktrees, got %v", model.selectedWorktrees)
	}
}

func TestEmptyStateRendersPlaceholder(t *testing.T) {
	model := createTestModel()
	showModal(model, nil, nil)

	if !strings.Contains(model.View(), "No worktrees found") {
		t.Errorf("empty worktree list must show the placeholder, got:\n%s", model.View())
	}
}

// commandContainsMessage reports whether cmd broadcasts an applied-selection or
// modal-state message, unwrapping the ComponentMessage envelope and tea.BatchMsg.
func commandContainsMessage(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}

	msg := cmd()
	if msg == nil {
		return false
	}

	if componentMsg, ok := msg.(base.ComponentMessage); ok {
		msg = componentMsg.Payload
	}

	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, subCmd := range batch {
			if commandContainsMessage(subCmd) {
				return true
			}
		}
		return false
	}

	if _, ok := msg.(WorktreeSelectionAppliedMsg); ok {
		return true
	}
	if modalMsg, ok := msg.(messages.ModalStateMsg); ok {
		return modalMsg.Type == string(base.ModalTypeWorktree)
	}

	return false
}
