package context

import (
	"testing"

	"github.com/yousfisaad/lazyarchon/v2/internal/shared/interfaces"
)

func featureSummaryContext() *ProgramContext {
	return &ProgramContext{
		Tasks: []interfaces.Task{
			{ID: "t1", Title: "Task 1", Status: "todo", Tags: []string{"alpha"}},
			{ID: "t2", Title: "Task 2", Status: "todo", Tags: []string{"beta"}},
		},
	}
}

func TestGetFeatureFilterSummary_NilMap_ShowsAll(t *testing.T) {
	ctx := featureSummaryContext()
	ctx.FeatureFilters = nil

	if got := ctx.GetFeatureFilterSummary(); got != "All features" {
		t.Errorf("nil filter: expected \"All features\", got %q", got)
	}
}

func TestGetFeatureFilterSummary_EmptyMap_ShowsNone(t *testing.T) {
	// Empty non-nil map = filter active with nothing selected.
	ctx := featureSummaryContext()
	ctx.FeatureFilters = map[string]bool{}

	if got := ctx.GetFeatureFilterSummary(); got != "No features" {
		t.Errorf("empty filter: expected \"No features\", got %q", got)
	}
}

func TestGetFeatureFilterSummary_SingleSelection_ShowsName(t *testing.T) {
	ctx := featureSummaryContext()
	ctx.FeatureFilters = map[string]bool{"alpha": true}

	if got := ctx.GetFeatureFilterSummary(); got != "alpha" {
		t.Errorf("single selection: expected \"alpha\", got %q", got)
	}
}

func TestUpdateFeatureFilterActiveState(t *testing.T) {
	ctx := featureSummaryContext()

	ctx.FeatureFilters = nil
	ctx.updateFeatureFilterActiveState()
	if ctx.FeatureFilterActive {
		t.Error("nil filter must not be active")
	}

	ctx.FeatureFilters = map[string]bool{}
	ctx.updateFeatureFilterActiveState()
	if !ctx.FeatureFilterActive {
		t.Error("empty non-nil filter is active (nothing selected) but FeatureFilterActive=false")
	}

	ctx.FeatureFilters = map[string]bool{"alpha": true}
	ctx.updateFeatureFilterActiveState()
	if !ctx.FeatureFilterActive {
		t.Error("populated filter must be active")
	}
}

func worktreeSummaryContext() *ProgramContext {
	return &ProgramContext{
		Tasks: []interfaces.Task{
			{ID: "t1", Title: "Task 1", Status: "todo"},
			{ID: "t2", Title: "Task 2", Status: "todo", Worktree: "wt-a"},
			{ID: "t3", Title: "Task 3", Status: "todo", Worktree: "wt-b"},
		},
	}
}

func TestGetUniqueWorktrees_SortedWithoutEmpty(t *testing.T) {
	ctx := worktreeSummaryContext()

	worktrees := ctx.GetUniqueWorktrees()
	if len(worktrees) != 2 || worktrees[0] != "wt-a" || worktrees[1] != "wt-b" {
		t.Errorf("GetUniqueWorktrees = %v, want [wt-a wt-b]", worktrees)
	}
}

func TestGetWorktreeFilterSummary_States(t *testing.T) {
	ctx := worktreeSummaryContext()

	ctx.WorktreeFilters = nil
	if got := ctx.GetWorktreeFilterSummary(); got != "All worktrees" {
		t.Errorf("nil filter: expected \"All worktrees\", got %q", got)
	}

	ctx.WorktreeFilters = map[string]bool{}
	if got := ctx.GetWorktreeFilterSummary(); got != "No worktrees" {
		t.Errorf("empty filter: expected \"No worktrees\", got %q", got)
	}

	ctx.WorktreeFilters = map[string]bool{"wt-a": true}
	if got := ctx.GetWorktreeFilterSummary(); got != "@wt-a" {
		t.Errorf("single selection: expected \"@wt-a\", got %q", got)
	}
}

func TestWorktreeFilterMutators(t *testing.T) {
	ctx := worktreeSummaryContext()

	// Toggle materializes the map and activates the filter.
	ctx.ToggleWorktreeFilter("wt-a")
	if !ctx.WorktreeFilterActive || !ctx.IsWorktreeVisible("wt-a") {
		t.Error("after toggle: filter must be active with wt-a visible")
	}
	if ctx.IsWorktreeVisible("wt-b") {
		t.Error("wt-b must be hidden once the filter is active")
	}

	// Set overrides a single entry.
	ctx.SetWorktreeFilter("wt-b", true)
	if !ctx.IsWorktreeVisible("wt-b") {
		t.Error("wt-b must be visible after SetWorktreeFilter(true)")
	}

	// Reset clears back to the nil (show all) state.
	ctx.ResetWorktreeFilters()
	if ctx.WorktreeFilters != nil || ctx.WorktreeFilterActive {
		t.Error("ResetWorktreeFilters must restore nil map and inactive state")
	}
	if !ctx.IsWorktreeVisible("wt-a") {
		t.Error("nil filter must show everything")
	}
}

func TestUpdateWorktreeFilterActiveState(t *testing.T) {
	ctx := worktreeSummaryContext()

	ctx.WorktreeFilters = nil
	ctx.updateWorktreeFilterActiveState()
	if ctx.WorktreeFilterActive {
		t.Error("nil filter must not be active")
	}

	ctx.WorktreeFilters = map[string]bool{}
	ctx.updateWorktreeFilterActiveState()
	if !ctx.WorktreeFilterActive {
		t.Error("empty non-nil filter is active (nothing selected) but WorktreeFilterActive=false")
	}

	ctx.WorktreeFilters = map[string]bool{"wt-a": true}
	ctx.updateWorktreeFilterActiveState()
	if !ctx.WorktreeFilterActive {
		t.Error("populated filter must be active")
	}
}

func TestGetCurrentSortModeName_AllModes(t *testing.T) {
	tests := []struct {
		mode int
		want string
	}{
		{0, "Status"},
		{1, "Priority"},
		{2, "Created"},
		{3, "Alpha"},
		{4, "Plan"},
		{99, "Unknown"},
	}

	for _, test := range tests {
		ctx := &ProgramContext{SortMode: test.mode}
		if got := ctx.GetCurrentSortModeName(); got != test.want {
			t.Errorf("GetCurrentSortModeName(%d) = %q, want %q", test.mode, got, test.want)
		}
	}
}

// archivedCountsContext returns a context whose task set mixes live and
// archived tasks across statuses and projects.
func archivedCountsContext() *ProgramContext {
	return &ProgramContext{
		Tasks: []interfaces.Task{
			{ID: "t-todo", Title: "Live todo", Status: interfaces.StatusTodo, ProjectID: "p-1"},
			{ID: "t-doing", Title: "Live doing", Status: interfaces.StatusDoing, ProjectID: "p-1"},
			{ID: "t-arch-todo", Title: "Archived todo", Status: interfaces.StatusTodo, ProjectID: "p-1", Archived: true},
			{ID: "t-arch-review", Title: "Archived review", Status: interfaces.StatusReview, ProjectID: "p-2", Archived: true},
			{ID: "t-p2", Title: "Live in p-2", Status: interfaces.StatusDone, ProjectID: "p-2"},
		},
	}
}

func TestGetTaskStatusCounts_ExcludesArchived(t *testing.T) {
	ctx := archivedCountsContext()

	todo, doing, review, done := ctx.GetTaskStatusCounts()

	if todo != 1 || doing != 1 || review != 0 || done != 1 {
		t.Errorf("archived tasks must not be counted, got todo=%d doing=%d review=%d done=%d", todo, doing, review, done)
	}
}

func TestGetTaskCountForProject_ExcludesArchived(t *testing.T) {
	ctx := archivedCountsContext()

	if got := ctx.GetTaskCountForProject("p-1"); got != 2 {
		t.Errorf("p-1 count must exclude archived tasks, got %d", got)
	}
	if got := ctx.GetTaskCountForProject("p-2"); got != 1 {
		t.Errorf("p-2 count must exclude archived tasks, got %d", got)
	}
}

func TestGetTotalTaskCount_ExcludesArchived(t *testing.T) {
	ctx := archivedCountsContext()

	if got := ctx.GetTotalTaskCount(); got != 3 {
		t.Errorf("total count must exclude archived tasks, got %d", got)
	}
}

func TestToggleShowArchivedTasks(t *testing.T) {
	ctx := &ProgramContext{}

	ctx.SetShowArchivedTasks(true)
	if !ctx.ShowArchivedTasks {
		t.Error("SetShowArchivedTasks(true) must enable the preference")
	}

	ctx.ToggleShowArchivedTasks()
	if ctx.ShowArchivedTasks {
		t.Error("ToggleShowArchivedTasks must disable the preference")
	}
}
