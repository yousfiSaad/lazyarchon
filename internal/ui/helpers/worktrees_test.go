package helpers

import (
	"testing"

	"github.com/yousfisaad/lazyarchon/v2/internal/shared/interfaces"
)

func TestGetUniqueWorktrees(t *testing.T) {
	tasks := []interfaces.Task{
		{ID: "t1", Title: "unstamped"},
		{ID: "t2", Title: "in b", Worktree: "wt-b"},
		{ID: "t3", Title: "in a", Worktree: "wt-a"},
		{ID: "t4", Title: "in b again", Worktree: "wt-b"},
	}

	worktrees := GetUniqueWorktrees(tasks)

	if len(worktrees) != 2 || worktrees[0] != "wt-a" || worktrees[1] != "wt-b" {
		t.Errorf("GetUniqueWorktrees = %v, want sorted [wt-a wt-b] without empties", worktrees)
	}
}

func TestGetUniqueWorktrees_EmptyWhenNoStamps(t *testing.T) {
	tasks := []interfaces.Task{{ID: "t1", Title: "unstamped"}}

	if worktrees := GetUniqueWorktrees(tasks); len(worktrees) != 0 {
		t.Errorf("GetUniqueWorktrees = %v, want empty", worktrees)
	}
}

func TestGetWorktreeFilterSummary(t *testing.T) {
	available := []string{"wt-a", "wt-b"}

	cases := []struct {
		name    string
		enabled map[string]bool
		want    string
	}{
		{"nil shows all", nil, "All worktrees"},
		{"empty shows none", map[string]bool{}, "No worktrees"},
		{"single selection names it", map[string]bool{"wt-a": true}, "@wt-a only"},
		{"all selected shows all", map[string]bool{"wt-a": true, "wt-b": true}, "All worktrees"},
	}

	// The n/m form needs 2+ selected out of 3+ (1-of-2 hits the "@x only" branch).
	three := []string{"wt-a", "wt-b", "wt-c"}
	partial := map[string]bool{"wt-a": true, "wt-b": true}
	if got := GetWorktreeFilterSummary(three, partial); got != "2/3 worktrees" {
		t.Errorf("partial summary = %q, want %q", got, "2/3 worktrees")
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetWorktreeFilterSummary(available, tt.enabled); got != tt.want {
				t.Errorf("GetWorktreeFilterSummary = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetWorktreeFilterSummary_NoWorktreesAvailable(t *testing.T) {
	if got := GetWorktreeFilterSummary(nil, nil); got != "No worktrees" {
		t.Errorf("summary with no worktrees = %q, want %q", got, "No worktrees")
	}
}

func TestGetWorktreeTaskCount(t *testing.T) {
	tasks := []interfaces.Task{
		{ID: "t1", Worktree: "wt-a"},
		{ID: "t2", Worktree: "wt-b"},
		{ID: "t3", Worktree: "wt-a"},
		{ID: "t4"},
	}

	if got := GetWorktreeTaskCount(tasks, "wt-a"); got != 2 {
		t.Errorf("GetWorktreeTaskCount(wt-a) = %d, want 2", got)
	}
	if got := GetWorktreeTaskCount(tasks, ""); got != 1 {
		t.Errorf("GetWorktreeTaskCount(\"\") = %d, want 1 (unstamped)", got)
	}
}
