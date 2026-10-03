package sorting

import (
	"testing"

	"github.com/yousfisaad/lazyarchon/v2/internal/shared/interfaces"
)

func TestGetSortModeName(t *testing.T) {
	tests := []struct {
		mode int
		want string
	}{
		{SortStatusPriority, "status+priority"},
		{SortPriorityOnly, "priority"},
		{SortTimeCreated, "time"},
		{SortAlphabetical, "alphabetical"},
		{SortPlan, "plan"},
		{SortModeCount, "unknown"},
		{-1, "unknown"},
	}

	for _, test := range tests {
		if got := GetSortModeName(test.mode); got != test.want {
			t.Errorf("GetSortModeName(%d) = %q, want %q", test.mode, got, test.want)
		}
	}
}

func TestSortModeCountMatchesRegisteredNames(t *testing.T) {
	// The `s` key cycles modulo SortModeCount; a name registered without
	// bumping the count (or vice versa) would skip or wrap a mode.
	if got := len(sortModeNames); got != SortModeCount {
		t.Errorf("len(sortModeNames) = %d, want SortModeCount = %d", got, SortModeCount)
	}
}

func TestSortTasksWithLevels_PlanUsesLevels(t *testing.T) {
	// "deep" has the better priority but sits at level 2; plan order puts
	// wave before priority.
	tasks := []interfaces.Task{
		{ID: "t-deep", Title: "Deep", Status: "todo", Priority: 1},
		{ID: "t-top", Title: "Top", Status: "todo", Priority: 4},
	}
	levels := map[string]int{"t-deep": 2, "t-top": 0}

	result := SortTasksWithLevels(tasks, SortPlan, levels)

	if result[0].ID != "t-top" || result[1].ID != "t-deep" {
		t.Errorf("plan order = [%s %s], want [t-top t-deep]", result[0].ID, result[1].ID)
	}
}

func TestSortTasksWithLevels_PlanNilLevelsFallsBackToTiebreaks(t *testing.T) {
	// Without levels every task is wave 0, so the tiebreak chain decides —
	// here priority alone.
	tasks := []interfaces.Task{
		{ID: "t-low", Title: "Low", Status: "todo", Priority: 4},
		{ID: "t-high", Title: "High", Status: "todo", Priority: 1},
	}

	result := SortTasksWithLevels(tasks, SortPlan, nil)

	if result[0].ID != "t-high" || result[1].ID != "t-low" {
		t.Errorf("plan order without levels = [%s %s], want [t-high t-low]", result[0].ID, result[1].ID)
	}
}

func TestSortTasksWithLevels_NonPlanModesIgnoreLevels(t *testing.T) {
	// Levels that would flip the plan order must not leak into other modes.
	tasks := []interfaces.Task{
		{ID: "t-deep", Title: "Deep", Status: "todo", Priority: 1},
		{ID: "t-top", Title: "Top", Status: "todo", Priority: 4},
	}
	levels := map[string]int{"t-deep": 2, "t-top": 0}

	if result := SortTasksWithLevels(tasks, SortPriorityOnly, levels); result[0].ID != "t-deep" {
		t.Errorf("priority order with levels present = [%s %s], want priority to decide", result[0].ID, result[1].ID)
	}
}

func TestSortTasks_OriginalSliceUntouched(t *testing.T) {
	tasks := []interfaces.Task{
		{ID: "t-a", Title: "A", Status: "todo", Priority: 1},
		{ID: "t-b", Title: "B", Status: "todo", Priority: 4},
	}
	levels := map[string]int{"t-a": 1, "t-b": 0}

	result := SortTasksWithLevels(tasks, SortPlan, levels)

	if tasks[0].ID != "t-a" || tasks[1].ID != "t-b" {
		t.Errorf("input slice was modified: [%s %s]", tasks[0].ID, tasks[1].ID)
	}
	if result[0].ID != "t-b" || result[1].ID != "t-a" {
		t.Errorf("plan order = [%s %s], want [t-b t-a] (level before priority)", result[0].ID, result[1].ID)
	}
}
