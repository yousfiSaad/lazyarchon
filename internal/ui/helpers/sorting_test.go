package helpers

import (
	"testing"

	"github.com/yousfisaad/lazyarchon/v2/internal/shared/interfaces"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/sorting"
)

// featureFilterTasks returns a fixed task set covering the feature filter cases:
// untagged, single-tag (alpha/beta), and multi-tag tasks.
func featureFilterTasks() []interfaces.Task {
	return []interfaces.Task{
		{ID: "t-untagged", Title: "Untagged task", Status: "todo"},
		{ID: "t-alpha", Title: "Alpha task", Status: "todo", Tags: []string{"alpha"}},
		{ID: "t-beta", Title: "Beta task", Status: "todo", Tags: []string{"beta"}},
		{ID: "t-both", Title: "Both task", Status: "todo", Tags: []string{"alpha", "beta"}},
	}
}

func taskIDSet(tasks []interfaces.Task) map[string]bool {
	ids := make(map[string]bool, len(tasks))
	for _, task := range tasks {
		ids[task.ID] = true
	}
	return ids
}

func TestFilterAndSortTasks_NilFeatureFilters_ShowsAll(t *testing.T) {
	filters := TaskFilters{FeatureFilters: nil, ShowCompletedTasks: true}

	result := FilterAndSortTasks(featureFilterTasks(), sorting.SortStatusPriority, filters)

	if got := len(result); got != 4 {
		t.Errorf("nil feature filter must show all tasks, got %d", got)
	}
}

func TestFilterAndSortTasks_EmptyFeatureFilters_ShowsOnlyUntagged(t *testing.T) {
	// Empty non-nil map = filter active with nothing selected:
	// only untagged tasks remain visible.
	filters := TaskFilters{FeatureFilters: map[string]bool{}, ShowCompletedTasks: true}

	result := FilterAndSortTasks(featureFilterTasks(), sorting.SortStatusPriority, filters)

	ids := taskIDSet(result)
	if len(result) != 1 || !ids["t-untagged"] {
		t.Errorf("empty feature filter must show only untagged tasks, got %v", ids)
	}
}

func TestFilterAndSortTasks_SelectedFeature_ExcludesUnselectedTagTasks(t *testing.T) {
	// Regression test: tasks carrying NONE of the selected features
	// must not appear in the filtered list.
	filters := TaskFilters{FeatureFilters: map[string]bool{"alpha": true}, ShowCompletedTasks: true}

	result := FilterAndSortTasks(featureFilterTasks(), sorting.SortStatusPriority, filters)

	ids := taskIDSet(result)
	for _, want := range []string{"t-untagged", "t-alpha", "t-both"} {
		if !ids[want] {
			t.Errorf("expected task %s in filtered list, got %v", want, ids)
		}
	}
	if ids["t-beta"] {
		t.Errorf("task tagged only with unselected feature must be excluded, got %v", ids)
	}
}

func TestFilterAndSortTasks_StatusFilterHidesDone(t *testing.T) {
	tasks := []interfaces.Task{
		{ID: "t-open", Title: "Open task", Status: "todo"},
		{ID: "t-done", Title: "Done task", Status: "done"},
	}
	filters := TaskFilters{ShowCompletedTasks: false}

	result := FilterAndSortTasks(tasks, sorting.SortStatusPriority, filters)

	ids := taskIDSet(result)
	if len(result) != 1 || !ids["t-open"] {
		t.Errorf("done tasks must be hidden when ShowCompletedTasks=false, got %v", ids)
	}
}

// worktreeFilterTasks returns a fixed task set covering the worktree filter
// cases: unstamped, and tasks stamped with one of two worktrees.
func worktreeFilterTasks() []interfaces.Task {
	return []interfaces.Task{
		{ID: "t-unstamped", Title: "Unstamped task", Status: "todo"},
		{ID: "t-wt-a", Title: "Worktree A task", Status: "todo", Worktree: "wt-a"},
		{ID: "t-wt-b", Title: "Worktree B task", Status: "todo", Worktree: "wt-b"},
	}
}

func TestFilterAndSortTasks_NilWorktreeFilters_ShowsAll(t *testing.T) {
	filters := TaskFilters{WorktreeFilters: nil, ShowCompletedTasks: true}

	result := FilterAndSortTasks(worktreeFilterTasks(), sorting.SortStatusPriority, filters)

	if got := len(result); got != 3 {
		t.Errorf("nil worktree filter must show all tasks, got %d", got)
	}
}

func TestFilterAndSortTasks_EmptyWorktreeFilters_ShowsOnlyUnstamped(t *testing.T) {
	// Empty non-nil map = filter active with nothing selected:
	// only unstamped tasks remain visible.
	filters := TaskFilters{WorktreeFilters: map[string]bool{}, ShowCompletedTasks: true}

	result := FilterAndSortTasks(worktreeFilterTasks(), sorting.SortStatusPriority, filters)

	ids := taskIDSet(result)
	if len(result) != 1 || !ids["t-unstamped"] {
		t.Errorf("empty worktree filter must show only unstamped tasks, got %v", ids)
	}
}

func TestFilterAndSortTasks_SelectedWorktree_ExcludesOtherWorktreeTasks(t *testing.T) {
	// Unstamped tasks always pass; stamped tasks pass only when enabled.
	filters := TaskFilters{WorktreeFilters: map[string]bool{"wt-a": true}, ShowCompletedTasks: true}

	result := FilterAndSortTasks(worktreeFilterTasks(), sorting.SortStatusPriority, filters)

	ids := taskIDSet(result)
	for _, want := range []string{"t-unstamped", "t-wt-a"} {
		if !ids[want] {
			t.Errorf("expected task %s in filtered list, got %v", want, ids)
		}
	}
	if ids["t-wt-b"] {
		t.Errorf("task stamped with unselected worktree must be excluded, got %v", ids)
	}
}

func TestFilterAndSortTasks_WorktreeAndFeatureFiltersCombine(t *testing.T) {
	// Both filters apply: a task must pass both to be visible. Untagged tasks
	// pass the feature filter and unstamped tasks pass the worktree filter,
	// so only t-wt-b is excluded (stamped with an unselected worktree).
	tasks := append(worktreeFilterTasks(), interfaces.Task{
		ID: "t-wt-a-tagged", Title: "Tagged in worktree A", Status: "todo",
		Worktree: "wt-a", Tags: []string{"alpha"},
	})
	filters := TaskFilters{
		FeatureFilters:     map[string]bool{"alpha": true},
		WorktreeFilters:    map[string]bool{"wt-a": true},
		ShowCompletedTasks: true,
	}

	result := FilterAndSortTasks(tasks, sorting.SortStatusPriority, filters)

	ids := taskIDSet(result)
	for _, want := range []string{"t-unstamped", "t-wt-a", "t-wt-a-tagged"} {
		if !ids[want] {
			t.Errorf("expected task %s in combined filtered list, got %v", want, ids)
		}
	}
	if ids["t-wt-b"] {
		t.Errorf("task in unselected worktree must be excluded, got %v", ids)
	}
}

// readyFilterTasks returns a fixed task set covering the ready filter cases:
// a blocker (open/done), a dependent, and an unconnected task.
func readyFilterTasks() []interfaces.Task {
	return []interfaces.Task{
		{ID: "t-blocker", Title: "Fix auth", Status: "todo"},
		{ID: "t-dependent", Title: "Ship release", Status: "todo", BlockedBy: []string{"t-blocker"}},
		{ID: "t-free", Title: "Unrelated", Status: "todo"},
	}
}

func TestFilterAndSortTasks_ReadyFilterHidesBlockedTasks(t *testing.T) {
	filters := TaskFilters{Ready: true, ShowCompletedTasks: true}

	result := FilterAndSortTasks(readyFilterTasks(), sorting.SortStatusPriority, filters)

	ids := taskIDSet(result)
	for _, want := range []string{"t-blocker", "t-free"} {
		if !ids[want] {
			t.Errorf("expected task %s in ready list, got %v", want, ids)
		}
	}
	if ids["t-dependent"] {
		t.Errorf("blocked task must be hidden by the ready filter, got %v", ids)
	}
}

func TestFilterAndSortTasks_ReadyFilterOffShowsAll(t *testing.T) {
	filters := TaskFilters{Ready: false, ShowCompletedTasks: true}

	result := FilterAndSortTasks(readyFilterTasks(), sorting.SortStatusPriority, filters)

	if got := len(result); got != 3 {
		t.Errorf("ready filter off must show all tasks, got %d", got)
	}
}

func TestFilterAndSortTasks_ReadyFilterDoneBlockerUnblocks(t *testing.T) {
	// Only done unblocks: finishing the blocker makes the dependent ready.
	tasks := []interfaces.Task{
		{ID: "t-blocker", Title: "Fix auth", Status: "done"},
		{ID: "t-dependent", Title: "Ship release", Status: "todo", BlockedBy: []string{"t-blocker"}},
	}
	filters := TaskFilters{Ready: true, ShowCompletedTasks: true}

	result := FilterAndSortTasks(tasks, sorting.SortStatusPriority, filters)

	ids := taskIDSet(result)
	if !ids["t-dependent"] {
		t.Errorf("dependent with done blocker must be ready, got %v", ids)
	}
}

func TestFilterAndSortTasks_ReadyFilterMissingBlockerStillBlocks(t *testing.T) {
	// A blocker ID absent from the loaded tasks counts as unfinished —
	// conservative, matching the local backend where archived-but-undone
	// blockers still block.
	tasks := []interfaces.Task{
		{ID: "t-dependent", Title: "Ship release", Status: "todo", BlockedBy: []string{"t-gone"}},
	}
	filters := TaskFilters{Ready: true, ShowCompletedTasks: true}

	result := FilterAndSortTasks(tasks, sorting.SortStatusPriority, filters)

	if got := len(result); got != 0 {
		t.Errorf("dependent with unknown blocker must be hidden, got %d tasks", got)
	}
}

func TestFilterAndSortTasks_ReadyFilterResolvesDoneBlockerHiddenByStatusFilter(t *testing.T) {
	// The ready filter runs before the completed-tasks filter, so a done
	// blocker hidden by ShowCompletedTasks=false still unblocks.
	tasks := readyFilterTasks()
	tasks[0].Status = "done"
	filters := TaskFilters{Ready: true, ShowCompletedTasks: false}

	result := FilterAndSortTasks(tasks, sorting.SortStatusPriority, filters)

	ids := taskIDSet(result)
	if !ids["t-dependent"] {
		t.Errorf("dependent must be ready via the hidden done blocker, got %v", ids)
	}
}

// archivedFilterTasks returns a fixed task set covering the archived filter
// cases: a live task, an archived-but-open task, and an archived done task.
func archivedFilterTasks() []interfaces.Task {
	return []interfaces.Task{
		{ID: "t-live", Title: "Live task", Status: "todo"},
		{ID: "t-archived", Title: "Archived task", Status: "todo", Archived: true},
		{ID: "t-archived-done", Title: "Archived done task", Status: "done", Archived: true},
	}
}

func TestFilterAndSortTasks_ArchivedHiddenByDefault(t *testing.T) {
	// ShowArchived zero value = archived tasks hidden, regardless of status.
	filters := TaskFilters{ShowCompletedTasks: true}

	result := FilterAndSortTasks(archivedFilterTasks(), sorting.SortStatusPriority, filters)

	ids := taskIDSet(result)
	if len(result) != 1 || !ids["t-live"] {
		t.Errorf("archived tasks must be hidden by default, got %v", ids)
	}
}

func TestFilterAndSortTasks_ShowArchivedIncludesArchived(t *testing.T) {
	filters := TaskFilters{ShowCompletedTasks: true, ShowArchived: true}

	result := FilterAndSortTasks(archivedFilterTasks(), sorting.SortStatusPriority, filters)

	if got := len(result); got != 3 {
		t.Errorf("ShowArchived must include archived tasks, got %d", got)
	}
}

func TestFilterAndSortTasks_ArchivedButUndoneBlockerStillHoldsReady(t *testing.T) {
	// The ready filter runs on the unfiltered slice, so an archived-but-undone
	// blocker still keeps its dependent out of the ready list — hiding the
	// blocker must not silently unblock the work behind it.
	tasks := []interfaces.Task{
		{ID: "t-blocker", Title: "Fix auth", Status: "todo", Archived: true},
		{ID: "t-dependent", Title: "Ship release", Status: "todo", BlockedBy: []string{"t-blocker"}},
	}

	for _, showArchived := range []bool{false, true} {
		filters := TaskFilters{Ready: true, ShowCompletedTasks: true, ShowArchived: showArchived}

		result := FilterAndSortTasks(tasks, sorting.SortStatusPriority, filters)

		ids := taskIDSet(result)
		if ids["t-dependent"] {
			t.Errorf("archived-but-undone blocker must still block (ShowArchived=%v), got %v", showArchived, ids)
		}
	}
}

func TestFilterAndSortTasks_ArchivedFilterComposesWithStatusFilter(t *testing.T) {
	// An archived done task stays hidden by the completed-tasks filter even
	// when archived rows are shown; the filters compose, not override.
	filters := TaskFilters{ShowCompletedTasks: false, ShowArchived: true}

	result := FilterAndSortTasks(archivedFilterTasks(), sorting.SortStatusPriority, filters)

	ids := taskIDSet(result)
	if len(result) != 2 || !ids["t-live"] || !ids["t-archived"] {
		t.Errorf("archived done task must stay hidden by the status filter, got %v", ids)
	}
}

func TestFilterAndSortTasks_PlanSortOrdersByWave(t *testing.T) {
	// B has the best priority but sits a wave behind A; the plan order is
	// the execution order, so A and the unconnected C come first.
	tasks := []interfaces.Task{
		{ID: "t-a", Title: "Fix auth", Status: "todo", Priority: 2},
		{ID: "t-b", Title: "Ship release", Status: "todo", Priority: 1, BlockedBy: []string{"t-a"}},
		{ID: "t-c", Title: "Side quest", Status: "todo", Priority: 4},
	}
	filters := TaskFilters{ShowCompletedTasks: true}

	result := FilterAndSortTasks(tasks, sorting.SortPlan, filters)

	want := []string{"t-a", "t-c", "t-b"}
	for i, id := range want {
		if result[i].ID != id {
			t.Fatalf("plan order[%d] = %s, want %s (full: %v)", i, result[i].ID, id, taskIDSet(result))
		}
	}
}

func TestFilterAndSortTasks_PlanSortHiddenTodoBlockerStillLevels(t *testing.T) {
	// A status filter hides the todo blocker from the visible list, but it
	// must still hold its doing dependent a wave behind the equal-priority
	// unblocked peer — levels come from the unfiltered slice.
	tasks := []interfaces.Task{
		{ID: "t-blocker", Title: "Fix auth", Status: "todo"},
		{ID: "t-dependent", Title: "Ship release", Status: "doing", Priority: 3, BlockedBy: []string{"t-blocker"}},
		{ID: "t-peer", Title: "Other work", Status: "doing", Priority: 3},
	}
	filters := TaskFilters{
		StatusFilters:      map[string]bool{"doing": true},
		StatusFilterActive: true,
	}

	result := FilterAndSortTasks(tasks, sorting.SortPlan, filters)

	if len(result) != 2 {
		t.Fatalf("plan sort must show only the doing tasks, got %v", taskIDSet(result))
	}
	if result[0].ID != "t-peer" || result[1].ID != "t-dependent" {
		t.Errorf("hidden todo blocker must sink its dependent below the peer, got [%s %s]",
			result[0].ID, result[1].ID)
	}
}

func TestFilterAndSortTasks_PlanSortDoneBlockerCollapsesDespiteHidden(t *testing.T) {
	// The regression guard for pre-filter leveling: a done blocker hidden by
	// ShowCompletedTasks=false must NOT hold its dependent back. Computing
	// levels after filtering would turn the absent blocker into an unknown
	// (unfinished) one and wrongly sink the dependent below the peer.
	tasks := []interfaces.Task{
		{ID: "t-blocker", Title: "Fix auth", Status: "done"},
		{ID: "t-dependent", Title: "Ship release", Status: "todo", Priority: 2, BlockedBy: []string{"t-blocker"}},
		{ID: "t-peer", Title: "Other work", Status: "todo", Priority: 4},
	}
	filters := TaskFilters{ShowCompletedTasks: false}

	result := FilterAndSortTasks(tasks, sorting.SortPlan, filters)

	if len(result) != 2 {
		t.Fatalf("plan sort must show only the open tasks, got %v", taskIDSet(result))
	}
	if result[0].ID != "t-dependent" || result[1].ID != "t-peer" {
		t.Errorf("done blocker must collapse: dependent is wave 0 and outranks the peer, got [%s %s]",
			result[0].ID, result[1].ID)
	}
}
