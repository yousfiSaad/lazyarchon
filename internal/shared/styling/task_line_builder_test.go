package styling

import (
	"regexp"
	"testing"

	"github.com/yousfisaad/lazyarchon/v2/internal/plugin"
)

// testStyleProvider enables priority indicators and disables feature colors,
// the common default configuration.
type testStyleProvider struct{}

func (testStyleProvider) IsPriorityIndicatorsEnabled() bool { return true }
func (testStyleProvider) IsFeatureColorsEnabled() bool      { return false }

func newTestStyleContext() *StyleContext {
	return NewStyleContext(&ThemeAdapter{
		TodoColor:   "yellow",
		DoingColor:  "blue",
		ReviewColor: "orange",
		DoneColor:   "green",
		HeaderColor: "cyan",
		MutedColor:  "gray",
		Name:        "test",
	}, testStyleProvider{})
}

// ansiPattern matches the SGR escape sequences lipgloss embeds in output.
var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// plainText strips styling so content assertions see the raw characters.
func plainText(styled string) string {
	return ansiPattern.ReplaceAllString(styled, "")
}

func TestTaskLineBuilderWorktreeBadge(t *testing.T) {
	stamped := plugin.Task{Title: "claimed", Status: plugin.StatusTodo, Worktree: "wt-x"}

	line := plainText(NewTaskLineBuilder(80, newTestStyleContext()).
		AddPriorityIndicator(stamped).
		AddStatusIndicator(stamped).
		AddTitle(stamped, "", false).
		AddFeatureTag(stamped).
		AddWorktreeBadge(stamped).
		Build("", false))

	if !containsPlain(line, "@wt-x") {
		t.Errorf("stamped task line = %q, want it to carry the @wt-x badge", line)
	}
}

func TestTaskLineBuilderWorktreeBadgeOmittedWhenUnstamped(t *testing.T) {
	unstamped := plugin.Task{Title: "free", Status: plugin.StatusTodo}

	line := plainText(NewTaskLineBuilder(80, newTestStyleContext()).
		AddPriorityIndicator(unstamped).
		AddStatusIndicator(unstamped).
		AddTitle(unstamped, "", false).
		AddFeatureTag(unstamped).
		AddWorktreeBadge(unstamped).
		Build("", false))

	if containsPlain(line, "@") {
		t.Errorf("unstamped task line = %q, want no badge", line)
	}
}

func TestTaskLineBuilderWorktreeBadgeDropsBeforeTagAndTitle(t *testing.T) {
	stamped := plugin.Task{
		Title:    "T",
		Status:   plugin.StatusTodo,
		Tags:     []string{"tag"},
		Worktree: "wt-x",
	}

	// Width 14 fits the fixed indicators (8 bytes), the 1-char title and the
	// 5-char feature tag, but not the additional 5-char badge: the badge
	// (priority 40) must be dropped while the tag (priority 50) survives.
	line := plainText(NewTaskLineBuilder(14, newTestStyleContext()).
		AddPriorityIndicator(stamped).
		AddStatusIndicator(stamped).
		AddTitle(stamped, "", false).
		AddFeatureTag(stamped).
		AddWorktreeBadge(stamped).
		Build("", false))

	if containsPlain(line, "@wt-x") {
		t.Errorf("narrow line = %q, want the badge dropped under width pressure", line)
	}
	if !containsPlain(line, "#tag") {
		t.Errorf("narrow line = %q, want the feature tag to survive when the badge drops", line)
	}
}

func TestTaskLineBuilderBlockedIndicator(t *testing.T) {
	blocked := plugin.Task{Title: "waiting", Status: plugin.StatusTodo}

	line := plainText(NewTaskLineBuilder(80, newTestStyleContext()).
		AddPriorityIndicator(blocked).
		AddStatusIndicator(blocked).
		AddBlockedIndicator(true).
		AddTitle(blocked, "", false).
		AddFeatureTag(blocked).
		AddWorktreeBadge(blocked).
		Build("", false))

	if !containsPlain(line, "⊘") {
		t.Errorf("blocked task line = %q, want it to carry the ⊘ indicator", line)
	}
}

func TestTaskLineBuilderBlockedIndicatorOmittedWhenReady(t *testing.T) {
	ready := plugin.Task{Title: "go ahead", Status: plugin.StatusTodo}

	line := plainText(NewTaskLineBuilder(80, newTestStyleContext()).
		AddPriorityIndicator(ready).
		AddStatusIndicator(ready).
		AddBlockedIndicator(false).
		AddTitle(ready, "", false).
		AddFeatureTag(ready).
		AddWorktreeBadge(ready).
		Build("", false))

	if containsPlain(line, "⊘") {
		t.Errorf("ready task line = %q, want no blocked indicator", line)
	}
}

// containsPlain reports whether the (already-stripped) line contains the
// marker. The leading space of chip content keeps "@wt-x" from matching a
// title that merely contains the same substring.
func containsPlain(line, marker string) bool {
	for _, field := range splitFields(line) {
		if field == marker || field == "#"+marker {
			return true
		}
	}
	return false
}

func splitFields(line string) []string {
	var fields []string
	start := -1
	for i, r := range line {
		if r == ' ' {
			if start >= 0 {
				fields = append(fields, line[start:i])
				start = -1
			}
			continue
		}
		if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		fields = append(fields, line[start:])
	}
	return fields
}

func TestTaskLineBuilderArchivedIndicator(t *testing.T) {
	archived := plugin.Task{Title: "boxed away", Status: plugin.StatusTodo, Archived: true}

	line := plainText(NewTaskLineBuilder(80, newTestStyleContext()).
		AddPriorityIndicator(archived).
		AddStatusIndicator(archived).
		AddArchivedIndicator(archived).
		AddTitle(archived, "", false).
		AddFeatureTag(archived).
		AddWorktreeBadge(archived).
		Build("", false))

	if !containsPlain(line, "☒") {
		t.Errorf("archived task line = %q, want it to carry the archived chip", line)
	}
}

func TestTaskLineBuilderArchivedIndicatorOmittedWhenLive(t *testing.T) {
	live := plugin.Task{Title: "active", Status: plugin.StatusTodo}

	line := plainText(NewTaskLineBuilder(80, newTestStyleContext()).
		AddPriorityIndicator(live).
		AddStatusIndicator(live).
		AddArchivedIndicator(live).
		AddTitle(live, "", false).
		AddFeatureTag(live).
		AddWorktreeBadge(live).
		Build("", false))

	if containsPlain(line, "☒") {
		t.Errorf("live task line = %q, want no archived chip", line)
	}
}

func TestTaskLineBuilderArchivedChipSurvivesOverFeatureTag(t *testing.T) {
	archived := plugin.Task{
		Title:    "T",
		Status:   plugin.StatusTodo,
		Tags:     []string{"tag"},
		Archived: true,
	}

	// Width 13 fits the three fixed indicators (4 bytes each: priority, status,
	// archived ☒) and the 1-char title, but leaves nothing for the 5-byte
	// feature tag: the chip is fixed like ⊘ and survives by construction,
	// while the tag (flexible, priority 50) is dropped.
	line := plainText(NewTaskLineBuilder(13, newTestStyleContext()).
		AddPriorityIndicator(archived).
		AddStatusIndicator(archived).
		AddArchivedIndicator(archived).
		AddTitle(archived, "", false).
		AddFeatureTag(archived).
		AddWorktreeBadge(archived).
		Build("", false))

	if !containsPlain(line, "☒") {
		t.Errorf("narrow line = %q, want the archived chip to survive", line)
	}
	if containsPlain(line, "#tag") {
		t.Errorf("narrow line = %q, want the feature tag dropped before the chip", line)
	}
}

func TestTaskLineBuilderArchivedChipSurvivesAlongsideBlockedIndicator(t *testing.T) {
	archived := plugin.Task{
		Title:    "A rather long archived and blocked title",
		Status:   plugin.StatusTodo,
		Archived: true,
	}

	// The regression: a blocked+archived row under width pressure kept the
	// fixed ⊘ but dropped the old flexible "archived " chip. Both markers are
	// fixed 2-cell glyphs now — they must co-display even when the title is
	// squeezed to nothing. Width 16 = 16 bytes of fixed indicators exactly
	// (4 × priority/status/blocked/archived), leaving the title zero budget.
	// (One byte narrower would trip Build()'s "..." fallback instead.)
	line := plainText(NewTaskLineBuilder(16, newTestStyleContext()).
		AddPriorityIndicator(archived).
		AddStatusIndicator(archived).
		AddBlockedIndicator(true).
		AddArchivedIndicator(archived).
		AddTitle(archived, "", false).
		AddFeatureTag(archived).
		AddWorktreeBadge(archived).
		Build("", false))

	if !containsPlain(line, "⊘") {
		t.Errorf("narrow line = %q, want the blocked indicator", line)
	}
	if !containsPlain(line, "☒") {
		t.Errorf("narrow line = %q, want the archived chip alongside the blocked indicator", line)
	}
}
