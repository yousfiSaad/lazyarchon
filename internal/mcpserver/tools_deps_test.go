package mcpserver

import (
	"reflect"
	"strings"
	"testing"
)

// Dependency-view tests over the local backend. Cycles are deliberately
// absent: the write path rejects them, so cycle coverage lives in the deps
// package tests against hand-built graphs.

// planWaveIDs extracts the task IDs of one wave for assertions.
func planWaveIDs(wave []planTask) []string {
	ids := make([]string, 0, len(wave))
	for _, task := range wave {
		ids = append(ids, task.ID)
	}
	return ids
}

// chainEntryIDs extracts the task IDs of a blockers/dependents slice.
func chainEntryIDs(entries []chainEntry) []string {
	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, entry.ID)
	}
	return ids
}

func TestGetExecutionPlanWaves(t *testing.T) {
	session := newTestSession(t)

	// A(todo, P2) <- B <- C, plus an unconnected D(todo, P4). Priorities
	// differ so wave order never depends on creation timestamps.
	a := callTool[createTaskOut](t, session, "create_task", map[string]interface{}{
		"title":    "Fix auth",
		"priority": 2,
	})
	b := callTool[createTaskOut](t, session, "create_task", map[string]interface{}{
		"title":      "Write migration",
		"priority":   1,
		"blocked_by": []string{a.Task.ID},
	})
	c := callTool[createTaskOut](t, session, "create_task", map[string]interface{}{
		"title":      "Ship release",
		"priority":   1,
		"blocked_by": []string{b.Task.ID},
	})
	d := callTool[createTaskOut](t, session, "create_task", map[string]interface{}{
		"title":    "Side quest",
		"priority": 4,
	})

	plan := callTool[executionPlanOut](t, session, "get_execution_plan", map[string]interface{}{})

	if len(plan.Waves) != 3 {
		t.Fatalf("waves = %v, want 3", plan.Waves)
	}
	if got := planWaveIDs(plan.Waves[0]); !reflect.DeepEqual(got, []string{a.Task.ID, d.Task.ID}) {
		t.Errorf("wave 0 = %v, want [A D] (priority order)", got)
	}
	if got := planWaveIDs(plan.Waves[1]); !reflect.DeepEqual(got, []string{b.Task.ID}) {
		t.Errorf("wave 1 = %v, want [B]", got)
	}
	if got := planWaveIDs(plan.Waves[2]); !reflect.DeepEqual(got, []string{c.Task.ID}) {
		t.Errorf("wave 2 = %v, want [C]", got)
	}

	if plan.TaskCount != 4 {
		t.Errorf("task_count = %d, want 4", plan.TaskCount)
	}
	if plan.SuggestedStart == nil || plan.SuggestedStart.ID != a.Task.ID {
		t.Errorf("suggested_start = %+v, want A (level-0 todo, best priority)", plan.SuggestedStart)
	}
	if plan.Cycles != nil {
		t.Errorf("cycles = %v, want none", plan.Cycles)
	}
}

func TestGetExecutionPlanDoneCollapse(t *testing.T) {
	session := newTestSession(t)

	a := callTool[createTaskOut](t, session, "create_task", map[string]interface{}{
		"title":    "Fix auth",
		"priority": 2,
	})
	b := callTool[createTaskOut](t, session, "create_task", map[string]interface{}{
		"title":      "Write migration",
		"blocked_by": []string{a.Task.ID},
	})
	c := callTool[createTaskOut](t, session, "create_task", map[string]interface{}{
		"title":      "Ship release",
		"blocked_by": []string{b.Task.ID},
	})

	callTool[updateTaskOut](t, session, "update_task", map[string]interface{}{
		"task_id": a.Task.ID,
		"status":  "done",
	})

	plan := callTool[executionPlanOut](t, session, "get_execution_plan", map[string]interface{}{})

	if len(plan.Waves) != 2 {
		t.Fatalf("waves = %v, want 2 after the blocker finished", plan.Waves)
	}
	if got := planWaveIDs(plan.Waves[0]); !reflect.DeepEqual(got, []string{b.Task.ID}) {
		t.Errorf("wave 0 = %v, want [B] (blocker done, B promoted)", got)
	}
	if got := planWaveIDs(plan.Waves[1]); !reflect.DeepEqual(got, []string{c.Task.ID}) {
		t.Errorf("wave 1 = %v, want [C]", got)
	}
	if plan.TaskCount != 2 {
		t.Errorf("task_count = %d, want 2 (done A never appears)", plan.TaskCount)
	}
	if plan.SuggestedStart == nil || plan.SuggestedStart.ID != b.Task.ID {
		t.Errorf("suggested_start = %+v, want B", plan.SuggestedStart)
	}
}

func TestGetExecutionPlanSingleWaveAndEmpty(t *testing.T) {
	session := newTestSession(t)

	empty := callTool[executionPlanOut](t, session, "get_execution_plan", map[string]interface{}{})
	if len(empty.Waves) != 0 || empty.TaskCount != 0 || empty.SuggestedStart != nil {
		t.Errorf("empty plan = %+v, want no waves and no suggestion", empty)
	}

	solo := callTool[createTaskOut](t, session, "create_task", map[string]interface{}{
		"title": "Lone task",
	})

	plan := callTool[executionPlanOut](t, session, "get_execution_plan", map[string]interface{}{})
	if len(plan.Waves) != 1 || len(plan.Waves[0]) != 1 || plan.Waves[0][0].ID != solo.Task.ID {
		t.Errorf("plan waves = %v, want one wave holding the lone task", plan.Waves)
	}
	if plan.SuggestedStart == nil || plan.SuggestedStart.ID != solo.Task.ID {
		t.Errorf("suggested_start = %+v, want the lone task", plan.SuggestedStart)
	}
}

func TestGetExecutionPlanProjectFilterKeepsCrossProjectBlockers(t *testing.T) {
	session := newTestSession(t)

	// Y first, while only the default project exists (create_task demands
	// an explicit project_id once there are several).
	y := callTool[createTaskOut](t, session, "create_task", map[string]interface{}{
		"title": "Upstream",
	})
	other := callTool[createProjectOut](t, session, "create_project", map[string]interface{}{
		"title": "Other",
	})
	x := callTool[createTaskOut](t, session, "create_task", map[string]interface{}{
		"title":      "Downstream",
		"project_id": other.Project.ID,
		"blocked_by": []string{y.Task.ID},
	})

	plan := callTool[executionPlanOut](t, session, "get_execution_plan", map[string]interface{}{
		"project_id": other.Project.ID,
	})

	// Y is out of scope but still shapes X's level: wave 0 is empty, X sits
	// alone at level 1 — waves[i] is always level i.
	if len(plan.Waves) != 2 {
		t.Fatalf("waves = %v, want 2 (gap wave 0)", plan.Waves)
	}
	if got := planWaveIDs(plan.Waves[0]); len(got) != 0 {
		t.Errorf("wave 0 = %v, want empty (Y filtered out)", got)
	}
	if got := planWaveIDs(plan.Waves[1]); !reflect.DeepEqual(got, []string{x.Task.ID}) {
		t.Errorf("wave 1 = %v, want [X]", got)
	}
	if plan.TaskCount != 1 {
		t.Errorf("task_count = %d, want 1", plan.TaskCount)
	}
	if plan.SuggestedStart != nil {
		t.Errorf("suggested_start = %+v, want none (X is not ready)", plan.SuggestedStart)
	}
}

func TestGetTaskChain(t *testing.T) {
	session := newTestSession(t)

	// C <- B <- A, plus D blocked by B. A gets a better priority than D so
	// the dependents order is timestamp-independent.
	c := callTool[createTaskOut](t, session, "create_task", map[string]interface{}{
		"title": "Fix auth",
	})
	b := callTool[createTaskOut](t, session, "create_task", map[string]interface{}{
		"title":      "Write migration",
		"blocked_by": []string{c.Task.ID},
	})
	a := callTool[createTaskOut](t, session, "create_task", map[string]interface{}{
		"title":      "Ship release",
		"priority":   2,
		"blocked_by": []string{b.Task.ID},
	})
	d := callTool[createTaskOut](t, session, "create_task", map[string]interface{}{
		"title":      "Notify users",
		"priority":   4,
		"blocked_by": []string{b.Task.ID},
	})

	chain := callTool[taskChainOut](t, session, "get_task_chain", map[string]interface{}{
		"task_id": b.Task.ID,
	})

	if chain.Task.ID != b.Task.ID {
		t.Errorf("chain task = %s, want B", chain.Task.ID)
	}
	if chain.Ready {
		t.Error("ready = true, want false (C is unfinished)")
	}
	if chain.Level != 1 {
		t.Errorf("level = %d, want 1", chain.Level)
	}
	if chain.LongestRemainingChain != 2 {
		t.Errorf("longest_remaining_chain = %d, want 2 (C plus B)", chain.LongestRemainingChain)
	}
	if got := chainEntryIDs(chain.Blockers); !reflect.DeepEqual(got, []string{c.Task.ID}) {
		t.Errorf("blockers = %v, want [C]", got)
	}
	if len(chain.Blockers) != 1 || !chain.Blockers[0].Direct || chain.Blockers[0].Depth != 1 {
		t.Errorf("blockers[0] = %+v, want direct depth-1 entry", chain.Blockers[0])
	}
	if got := chainEntryIDs(chain.Dependents); !reflect.DeepEqual(got, []string{a.Task.ID, d.Task.ID}) {
		t.Errorf("dependents = %v, want [A D] (priority order)", got)
	}

	// Finishing C makes B ready without touching its dependents.
	callTool[updateTaskOut](t, session, "update_task", map[string]interface{}{
		"task_id": c.Task.ID,
		"status":  "done",
	})

	unblocked := callTool[taskChainOut](t, session, "get_task_chain", map[string]interface{}{
		"task_id": b.Task.ID,
	})

	if !unblocked.Ready {
		t.Error("ready after C done = false, want true")
	}
	if unblocked.Level != 0 {
		t.Errorf("level after C done = %d, want 0", unblocked.Level)
	}
	if len(unblocked.Blockers) != 0 {
		t.Errorf("blockers after C done = %v, want empty", chainEntryIDs(unblocked.Blockers))
	}
	if unblocked.LongestRemainingChain != 1 {
		t.Errorf("longest_remaining_chain after C done = %d, want 1", unblocked.LongestRemainingChain)
	}
	if got := chainEntryIDs(unblocked.Dependents); !reflect.DeepEqual(got, []string{a.Task.ID, d.Task.ID}) {
		t.Errorf("dependents after C done = %v, want unchanged", got)
	}
}

func TestGetTaskChainUnknownTask(t *testing.T) {
	session := newTestSession(t)

	message := callToolError(t, session, "get_task_chain", map[string]interface{}{
		"task_id": "no-such-task",
	})

	if !strings.Contains(message, "not found") {
		t.Errorf("error = %q, want a not-found message", message)
	}
}
