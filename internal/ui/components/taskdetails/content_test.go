package taskdetails

import (
	"strings"
	"testing"
	"time"

	"github.com/yousfisaad/lazyarchon/v2/internal/shared/interfaces"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/base"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/context"
)

// dependencyDetailTasks returns a blocker/dependent pair plus an
// unconnected task, wired through a program context.
func dependencyDetailTasks() (*context.ProgramContext, []interfaces.Task) {
	tasks := []interfaces.Task{
		{ID: "t-blocker", Title: "Fix auth", Status: "todo"},
		{ID: "t-dependent", Title: "Ship release", Status: "todo", BlockedBy: []string{"t-blocker"}},
		{ID: "t-free", Title: "Unrelated", Status: "todo"},
	}

	return &context.ProgramContext{Tasks: tasks}, tasks
}

func dependencyGenerator(programContext *context.ProgramContext, task *interfaces.Task) *TaskContentGenerator {
	generator := NewTaskContentGenerator(80, &base.ComponentContext{ProgramContext: programContext})
	generator.SetTask(task)
	return &generator
}

func TestGenerateLinesDependentShowsBlockedBy(t *testing.T) {
	programContext, tasks := dependencyDetailTasks()

	lines := dependencyGenerator(programContext, &tasks[1]).GenerateLines()
	joined := strings.Join(lines, "\n")

	if !strings.Contains(joined, "Blocked by:") || !strings.Contains(joined, "Fix auth (t-blocker)") {
		t.Errorf("dependent details = %q, want a Blocked by line naming the blocker", joined)
	}
	if strings.Contains(joined, "Blocks:") {
		t.Errorf("dependent details = %q, want no Blocks line for a task nothing depends on", joined)
	}
}

func TestGenerateLinesBlockerShowsBlocks(t *testing.T) {
	programContext, tasks := dependencyDetailTasks()

	lines := dependencyGenerator(programContext, &tasks[0]).GenerateLines()
	joined := strings.Join(lines, "\n")

	if !strings.Contains(joined, "Blocks:") || !strings.Contains(joined, "Ship release (t-dependent)") {
		t.Errorf("blocker details = %q, want a Blocks line naming the dependent", joined)
	}
	if strings.Contains(joined, "Blocked by:") {
		t.Errorf("blocker details = %q, want no Blocked by line for an unblocked task", joined)
	}
}

func TestGenerateLinesUnconnectedTaskShowsNoDependencyLines(t *testing.T) {
	programContext, tasks := dependencyDetailTasks()

	lines := dependencyGenerator(programContext, &tasks[2]).GenerateLines()
	joined := strings.Join(lines, "\n")

	if strings.Contains(joined, "Blocked by:") || strings.Contains(joined, "Blocks:") {
		t.Errorf("unconnected task details = %q, want no dependency lines", joined)
	}
}

// chainDetailTasks returns a three-task chain C <- B <- A wired through a
// program context.
func chainDetailTasks() (*context.ProgramContext, []interfaces.Task) {
	tasks := []interfaces.Task{
		{ID: "t-c", Title: "Fix auth", Status: "todo"},
		{ID: "t-b", Title: "Write migration", Status: "todo", BlockedBy: []string{"t-c"}},
		{ID: "t-a", Title: "Ship release", Status: "todo", BlockedBy: []string{"t-b"}},
	}

	return &context.ProgramContext{Tasks: tasks}, tasks
}

func TestGenerateLinesDepthCountsTransitiveChain(t *testing.T) {
	programContext, tasks := chainDetailTasks()

	lines := dependencyGenerator(programContext, &tasks[2]).GenerateLines()
	joined := strings.Join(lines, "\n")

	if !strings.Contains(joined, "Depth: 3 (2 tasks ahead on the longest unfinished chain)") {
		t.Errorf("deepest task details = %q, want a Depth line counting both blockers ahead", joined)
	}
}

func TestGenerateLinesDepthSingularForDirectBlocker(t *testing.T) {
	programContext, tasks := chainDetailTasks()

	lines := dependencyGenerator(programContext, &tasks[1]).GenerateLines()
	joined := strings.Join(lines, "\n")

	if !strings.Contains(joined, "Depth: 2 (1 task ahead on the longest unfinished chain)") {
		t.Errorf("mid-chain task details = %q, want a singular Depth line", joined)
	}
}

func TestGenerateLinesReadyAndUnconnectedTasksShowNoDepth(t *testing.T) {
	programContext, tasks := chainDetailTasks()

	// The ready chain head has depth 1: nothing unfinished ahead, no line.
	ready := dependencyGenerator(programContext, &tasks[0]).GenerateLines()
	if joined := strings.Join(ready, "\n"); strings.Contains(joined, "Depth:") {
		t.Errorf("ready task details = %q, want no Depth line", joined)
	}

	// Without program context there is no graph to walk: no line, even for
	// a task declaring blockers.
	_, pair := dependencyDetailTasks()
	contextless := dependencyGenerator(nil, &pair[1]).GenerateLines()
	if joined := strings.Join(contextless, "\n"); strings.Contains(joined, "Depth:") {
		t.Errorf("contextless task details = %q, want no Depth line", joined)
	}
}

func TestGenerateLinesArchivedTaskShowsArchivedAt(t *testing.T) {
	archivedAt := time.Date(2026, 10, 3, 9, 30, 0, 0, time.UTC)
	programContext, tasks := dependencyDetailTasks()
	tasks[0].Archived = true
	tasks[0].ArchivedAt = &archivedAt

	lines := dependencyGenerator(programContext, &tasks[0]).GenerateLines()
	joined := strings.Join(lines, "\n")

	if !strings.Contains(joined, "Archived: 2026-10-03 09:30") {
		t.Errorf("archived details = %q, want an Archived line with the timestamp", joined)
	}
}

func TestGenerateLinesArchivedWithoutTimestampShowsYes(t *testing.T) {
	programContext, tasks := dependencyDetailTasks()
	tasks[0].Archived = true

	lines := dependencyGenerator(programContext, &tasks[0]).GenerateLines()
	joined := strings.Join(lines, "\n")

	if !strings.Contains(joined, "Archived: yes") {
		t.Errorf("archived details = %q, want an Archived: yes fallback line", joined)
	}
}

func TestGenerateLinesLiveTaskHasNoArchivedLine(t *testing.T) {
	programContext, tasks := dependencyDetailTasks()

	lines := dependencyGenerator(programContext, &tasks[0]).GenerateLines()
	joined := strings.Join(lines, "\n")

	if strings.Contains(joined, "Archived:") {
		t.Errorf("live details = %q, want no Archived line", joined)
	}
}
