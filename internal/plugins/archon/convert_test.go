package archon

import (
	"context"
	"testing"
	"time"

	"github.com/yousfisaad/lazyarchon/v2/internal/plugin"
)

// newAdapterOverMock wires an ArchonClientAdapter to a fresh mock server.
func newAdapterOverMock(t *testing.T) (*ArchonClientAdapter, *MockServer) {
	t.Helper()

	server := SetupMockServerWithData()
	t.Cleanup(server.Close)

	adapter := &ArchonClientAdapter{
		client: NewResilientClient(server.URL, "test-key", 30*time.Second, DefaultResilienceConfig()),
	}

	return adapter, server
}

func intPtr(i int) *int { return &i }

// TestPriorityToTaskOrder pins the generic-to-legacy scale mapping: the
// generic scale counts down (1 = critical), Archon's task_order counts up
// (higher = more urgent, >=80 high / >=50 medium).
func TestPriorityToTaskOrder(t *testing.T) {
	tests := []struct {
		priority int
		want     int
	}{
		{priority: plugin.PriorityCritical, want: 95},
		{priority: plugin.PriorityHigh, want: 85},
		{priority: plugin.PriorityMedium, want: 60},
		{priority: plugin.PriorityLow, want: 40},
		{priority: plugin.PriorityBacklog, want: 10},
	}

	for _, tt := range tests { //nolint:varnamelen // tt is idiomatic for table-driven tests
		if got := priorityToTaskOrder(tt.priority); got != tt.want {
			t.Errorf("priorityToTaskOrder(%d) = %d, want %d", tt.priority, got, tt.want)
		}
	}
}

// TestTaskOrderToPriority pins the lossy inverse: Archon only distinguishes
// three tiers, so legacy orders collapse onto high/medium/low.
func TestTaskOrderToPriority(t *testing.T) {
	tests := []struct {
		taskOrder int
		want      int
	}{
		{taskOrder: 100, want: plugin.PriorityHigh},
		{taskOrder: 95, want: plugin.PriorityHigh},
		{taskOrder: 80, want: plugin.PriorityHigh},
		{taskOrder: 79, want: plugin.PriorityMedium},
		{taskOrder: 60, want: plugin.PriorityMedium},
		{taskOrder: 50, want: plugin.PriorityMedium},
		{taskOrder: 49, want: plugin.PriorityLow},
		{taskOrder: 10, want: plugin.PriorityLow},
		{taskOrder: 0, want: plugin.PriorityLow},
	}

	for _, tt := range tests { //nolint:varnamelen // tt is idiomatic for table-driven tests
		if got := taskOrderToPriority(tt.taskOrder); got != tt.want {
			t.Errorf("taskOrderToPriority(%d) = %d, want %d", tt.taskOrder, got, tt.want)
		}
	}
}

// TestAdapterCreateTaskConvertsPriority guards the inverted-scale bug: the
// adapter used to pass priority through unchanged, so a critical (1) task was
// stored with Archon task_order 1 — the lowest urgency on the legacy scale.
func TestAdapterCreateTaskConvertsPriority(t *testing.T) {
	adapter, server := newAdapterOverMock(t)
	ctx := context.Background()

	created, err := adapter.CreateTask(ctx, plugin.CreateTaskRequest{
		ProjectID: "test-project-1", Title: "urgent thing", Priority: plugin.PriorityCritical,
	})
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Wire check: the stored Archon task must carry the legacy-scale order.
	raw := NewClient(server.URL, "test-key", 30*time.Second)
	resp, err := raw.GetTask(created.ID)
	if err != nil {
		t.Fatalf("raw GetTask() error = %v", err)
	}

	if resp.Task.TaskOrder != 95 {
		t.Errorf("stored task_order = %d, want 95 (priority 1 must map to legacy high)", resp.Task.TaskOrder)
	}

	// Read-back is lossy: Archon's high tier converts back to priority 2.
	if created.Priority != plugin.PriorityHigh {
		t.Errorf("created.Priority = %d, want %d (lossy high tier)", created.Priority, plugin.PriorityHigh)
	}
}

// TestAdapterUpdateTaskConvertsPriority guards the update-side passthrough.
func TestAdapterUpdateTaskConvertsPriority(t *testing.T) {
	adapter, server := newAdapterOverMock(t)
	ctx := context.Background()

	updated, err := adapter.UpdateTask(ctx, "todo-task-1", plugin.UpdateTaskRequest{
		Priority: intPtr(plugin.PriorityCritical),
	})
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}

	raw := NewClient(server.URL, "test-key", 30*time.Second)
	resp, err := raw.GetTask("todo-task-1")
	if err != nil {
		t.Fatalf("raw GetTask() error = %v", err)
	}

	if resp.Task.TaskOrder != 95 {
		t.Errorf("stored task_order = %d, want 95", resp.Task.TaskOrder)
	}

	if updated.Priority != plugin.PriorityHigh {
		t.Errorf("updated.Priority = %d, want %d (lossy high tier)", updated.Priority, plugin.PriorityHigh)
	}
}

// TestAdapterConvertsTaskOrderOnRead guards convertTask: legacy orders must
// arrive on the generic 1-5 scale, never as raw 0-100 values.
func TestAdapterConvertsTaskOrderOnRead(t *testing.T) {
	adapter, _ := newAdapterOverMock(t)

	result, err := adapter.ListTasks(context.Background(), plugin.TaskFilters{})
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}

	priorities := make(map[string]int, len(result.Tasks))
	for _, task := range result.Tasks {
		priorities[task.ID] = task.Priority
	}

	// HighPriorityTask carries task_order 100; LowPriorityTask carries 1 and
	// the remaining fixtures default to 10 — all land in the low band.
	for id, want := range map[string]int{
		"high-priority-task": plugin.PriorityHigh,
		"low-priority-task":  plugin.PriorityLow,
		"todo-task-1":        plugin.PriorityLow,
	} {
		if got := priorities[id]; got != want {
			t.Errorf("task %s priority = %d, want %d", id, got, want)
		}
	}

	for _, task := range result.Tasks {
		if task.Priority < plugin.PriorityCritical || task.Priority > plugin.PriorityBacklog {
			t.Errorf("task %s priority = %d, outside the 1-5 scale (raw task_order leaked?)", task.ID, task.Priority)
		}
	}
}
