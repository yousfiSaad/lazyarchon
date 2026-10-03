package header

import (
	"strings"
	"testing"

	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/base"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/context"
)

// readyArchivedHeader builds a header over a program context with the two
// filter flags preset, plus the minimal wiring View() needs.
func readyArchivedHeader(readyOnly bool, showArchived bool) *HeaderModel {
	programContext := &context.ProgramContext{
		ReadyOnly:         readyOnly,
		ShowArchivedTasks: showArchived,
	}
	componentContext := &base.ComponentContext{
		ProgramContext: programContext,
		UIState:        context.NewUIState(),
		GetSortedTasks: func() []interface{} { return nil },
	}
	return NewModel(componentContext)
}

func TestHeaderViewFilterIndicators(t *testing.T) {
	tests := []struct {
		name         string
		readyOnly    bool
		showArchived bool
		wantReady    bool
		wantArchived bool
	}{
		{"no filters", false, false, false, false},
		{"ready only", true, false, true, false},
		{"archived shown", false, true, false, true},
		{"both filters together", true, true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := readyArchivedHeader(tt.readyOnly, tt.showArchived).View()

			hasReady := strings.Contains(view, "⊘ Ready only")
			hasArchived := strings.Contains(view, "☒ Archived shown")
			if hasReady != tt.wantReady {
				t.Errorf("header = %q, ready-only indicator presence = %v, want %v", view, hasReady, tt.wantReady)
			}
			if hasArchived != tt.wantArchived {
				t.Errorf("header = %q, archived indicator presence = %v, want %v", view, hasArchived, tt.wantArchived)
			}
		})
	}
}
