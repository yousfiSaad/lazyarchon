package deps

import (
	"reflect"
	"testing"
	"time"

	"github.com/yousfisaad/lazyarchon/v2/internal/plugin"
)

// newTask builds a fixture with explicit CreatedAt — the determinism chain
// ends there, so tests must never rely on the zero value.
func newTask(id string, opts ...func(*plugin.Task)) plugin.Task {
	task := plugin.Task{
		ID:        id,
		Title:     id,
		Status:    plugin.StatusTodo,
		Priority:  3,
		CreatedAt: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	for _, opt := range opts {
		opt(&task)
	}
	return task
}

func withStatus(status string) func(*plugin.Task) {
	return func(task *plugin.Task) { task.Status = status }
}

func withPriority(priority int) func(*plugin.Task) {
	return func(task *plugin.Task) { task.Priority = priority }
}

func withBlockedBy(ids ...string) func(*plugin.Task) {
	return func(task *plugin.Task) { task.BlockedBy = ids }
}

func withProject(projectID string) func(*plugin.Task) {
	return func(task *plugin.Task) { task.ProjectID = projectID }
}

func withArchived() func(*plugin.Task) {
	return func(task *plugin.Task) { task.Archived = true }
}

func withCreatedAt(createdAt time.Time) func(*plugin.Task) {
	return func(task *plugin.Task) { task.CreatedAt = createdAt }
}

func assertLevels(t *testing.T, graph *Graph, want map[string]int) {
	t.Helper()
	got := graph.Levels()
	for id, level := range want {
		if got[id] != level {
			t.Errorf("level(%s) = %d, want %d (all levels: %v)", id, got[id], level, got)
		}
	}
}

func upstreamIDs(nodes []ChainNode) []string {
	ids := make([]string, 0, len(nodes))
	for _, node := range nodes {
		ids = append(ids, node.Task.ID)
	}
	return ids
}

func TestChainLevels(t *testing.T) {
	// C <- B <- A: A blocked by B, B blocked by C, all todo.
	graph := Build([]plugin.Task{
		newTask("a", withBlockedBy("b")),
		newTask("b", withBlockedBy("c")),
		newTask("c"),
	})

	assertLevels(t, graph, map[string]int{"a": 2, "b": 1, "c": 0})

	if graph.Ready("a") {
		t.Error("Ready(a) = true, want false")
	}
	if graph.Ready("b") {
		t.Error("Ready(b) = true, want false")
	}
	if !graph.Ready("c") {
		t.Error("Ready(c) = false, want true")
	}

	if got := graph.ChainLength("a"); got != 3 {
		t.Errorf("ChainLength(a) = %d, want 3", got)
	}
	if got := graph.ChainLength("b"); got != 2 {
		t.Errorf("ChainLength(b) = %d, want 2", got)
	}
	if got := graph.ChainLength("c"); got != 1 {
		t.Errorf("ChainLength(c) = %d, want 1", got)
	}
	if got := graph.ChainLength("ghost"); got != 0 {
		t.Errorf("ChainLength(ghost) = %d, want 0", got)
	}

	upstream := graph.Upstream("a")
	if got := upstreamIDs(upstream); !reflect.DeepEqual(got, []string{"b", "c"}) {
		t.Errorf("Upstream(a) ids = %v, want [b c]", got)
	}
	if !upstream[0].Direct || upstream[0].Depth != 1 {
		t.Errorf("Upstream(a)[0] = {Depth: %d, Direct: %v}, want {1, true}", upstream[0].Depth, upstream[0].Direct)
	}
	if upstream[1].Direct || upstream[1].Depth != 2 {
		t.Errorf("Upstream(a)[1] = {Depth: %d, Direct: %v}, want {2, false}", upstream[1].Depth, upstream[1].Direct)
	}

	downstream := graph.Downstream("c")
	if got := upstreamIDs(downstream); !reflect.DeepEqual(got, []string{"b", "a"}) {
		t.Errorf("Downstream(c) ids = %v, want [b a]", got)
	}
}

func TestDiamondLevels(t *testing.T) {
	// A blocked by B and C, both blocked by D.
	graph := Build([]plugin.Task{
		newTask("a", withBlockedBy("b", "c")),
		newTask("b", withBlockedBy("d")),
		newTask("c", withBlockedBy("d")),
		newTask("d"),
	})

	assertLevels(t, graph, map[string]int{"a": 2, "b": 1, "c": 1, "d": 0})

	upstream := graph.Upstream("a")
	depths := make(map[string]int, len(upstream))
	for _, node := range upstream {
		depths[node.Task.ID] = node.Depth
	}
	want := map[string]int{"b": 1, "c": 1, "d": 2}
	if !reflect.DeepEqual(depths, want) {
		t.Errorf("Upstream(a) depths = %v, want %v", depths, want)
	}
}

func TestDoneCollapse(t *testing.T) {
	// A blocked by done-B, which is blocked by todo-C. The finished blocker
	// collapses: A is ready and its chain ignores C entirely. B itself keeps
	// level 1 (its own blocker is unfinished) — harmless, done tasks never
	// become wave entries.
	graph := Build([]plugin.Task{
		newTask("a", withBlockedBy("b")),
		newTask("b", withStatus(plugin.StatusDone), withBlockedBy("c")),
		newTask("c"),
	})

	assertLevels(t, graph, map[string]int{"a": 0, "b": 1, "c": 0})

	if !graph.Ready("a") {
		t.Error("Ready(a) = false, want true")
	}
	if got := graph.ChainLength("a"); got != 1 {
		t.Errorf("ChainLength(a) = %d, want 1", got)
	}
	if nodes := graph.Upstream("a"); len(nodes) != 0 {
		t.Errorf("Upstream(a) = %v, want empty", upstreamIDs(nodes))
	}
}

func TestArchivedUndoneBlockerStillBlocks(t *testing.T) {
	// An archived-undone blocker holds its dependent back (callers decide
	// whether the archived task itself becomes a wave entry).
	graph := Build([]plugin.Task{
		newTask("a", withBlockedBy("b")),
		newTask("b", withArchived()),
	})

	assertLevels(t, graph, map[string]int{"a": 1, "b": 0})
	if graph.Ready("a") {
		t.Error("Ready(a) = true, want false")
	}
}

func TestCrossProjectEdgeShapedGlobally(t *testing.T) {
	graph := Build([]plugin.Task{
		newTask("x", withProject("p2"), withBlockedBy("y")),
		newTask("y", withProject("p1")),
	})

	assertLevels(t, graph, map[string]int{"x": 1, "y": 0})
}

func TestUnknownBlockerCountsAsUnfinished(t *testing.T) {
	graph := Build([]plugin.Task{
		newTask("a", withBlockedBy("ghost")),
	})

	assertLevels(t, graph, map[string]int{"a": 1})
	if graph.Ready("a") {
		t.Error("Ready(a) = true, want false")
	}
	if got := graph.ChainLength("a"); got != 2 {
		t.Errorf("ChainLength(a) = %d, want 2 (the task plus one virtual node)", got)
	}
	if nodes := graph.Upstream("a"); len(nodes) != 0 {
		t.Errorf("Upstream(a) = %v, want empty (unknown ids cannot be represented)", upstreamIDs(nodes))
	}
}

func TestCycleSafety(t *testing.T) {
	// A <-> B plus C blocked by A. The write path guards against cycles, so
	// this only guards corrupt data — the assertions double as a hang test.
	graph := Build([]plugin.Task{
		newTask("a", withBlockedBy("b")),
		newTask("b", withBlockedBy("a")),
		newTask("c", withBlockedBy("a")),
	})

	if got := graph.Cycles(); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Errorf("Cycles() = %v, want [a b]", got)
	}

	// Everyone is downstream of the cycle: all stuck in the phantom wave.
	for _, id := range []string{"a", "b", "c"} {
		if level := graph.Levels()[id]; level < 1 {
			t.Errorf("level(%s) = %d, want >= 1 (stuck tasks must not look ready)", id, level)
		}
		if graph.Ready(id) {
			t.Errorf("Ready(%s) = true, want false", id)
		}
	}
}

func TestAllCyclicGraph(t *testing.T) {
	graph := Build([]plugin.Task{
		newTask("a", withBlockedBy("b")),
		newTask("b", withBlockedBy("a")),
	})

	if got := graph.Cycles(); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Errorf("Cycles() = %v, want [a b]", got)
	}
	if level := graph.Levels()["a"]; level != 1 {
		t.Errorf("level(a) = %d, want 1 (nothing processed; phantom wave floors at 1)", level)
	}
}

func TestSelfLoopCountsAsCycle(t *testing.T) {
	// Unreachable through the guarded write path; pins the Tarjan
	// self-loop branch regardless.
	graph := Build([]plugin.Task{
		newTask("a", withBlockedBy("a")),
	})

	if got := graph.Cycles(); !reflect.DeepEqual(got, []string{"a"}) {
		t.Errorf("Cycles() = %v, want [a]", got)
	}
	if graph.Ready("a") {
		t.Error("Ready(a) = true, want false")
	}
}

func TestEmptyGraph(t *testing.T) {
	graph := Build(nil)

	if levels := graph.Levels(); len(levels) != 0 {
		t.Errorf("Levels() = %v, want empty", levels)
	}
	if cycles := graph.Cycles(); len(cycles) != 0 {
		t.Errorf("Cycles() = %v, want empty", cycles)
	}
	if got := graph.ChainLength("a"); got != 0 {
		t.Errorf("ChainLength(a) = %d, want 0", got)
	}
	if nodes := graph.Upstream("a"); nodes != nil {
		t.Errorf("Upstream(a) = %v, want nil", upstreamIDs(nodes))
	}
	if graph.Ready("ghost") {
		t.Error("Ready(ghost) = true, want false")
	}
}

func TestAllDone(t *testing.T) {
	graph := Build([]plugin.Task{
		newTask("a", withStatus(plugin.StatusDone), withBlockedBy("b")),
		newTask("b", withStatus(plugin.StatusDone)),
	})

	assertLevels(t, graph, map[string]int{"a": 0, "b": 0})
	if got := graph.ChainLength("a"); got != 1 {
		t.Errorf("ChainLength(a) = %d, want 1", got)
	}
}

func TestUpstreamDownstreamAsymmetry(t *testing.T) {
	// Blockers collapse when done; dependents never do — finishing work
	// must still show its impact, and the raw Blocks field keeps edges to
	// done tasks too.
	graph := Build([]plugin.Task{
		newTask("a", withBlockedBy("b")),
		newTask("b", withStatus(plugin.StatusDone)),
		newTask("c", withStatus(plugin.StatusDone), withBlockedBy("a")),
	})

	if nodes := graph.Upstream("a"); len(nodes) != 0 {
		t.Errorf("Upstream(a) = %v, want empty (done blocker omitted)", upstreamIDs(nodes))
	}
	if got := upstreamIDs(graph.Downstream("a")); !reflect.DeepEqual(got, []string{"c"}) {
		t.Errorf("Downstream(a) = %v, want [c] (done dependent kept)", got)
	}
}

func TestDuplicateIDsLastWins(t *testing.T) {
	graph := Build([]plugin.Task{
		newTask("a", withPriority(1)),
		newTask("a", withPriority(5)),
	})

	if got := graph.Levels()["a"]; got != 0 {
		t.Errorf("level(a) = %d, want 0", got)
	}
}

func TestLessTiebreaks(t *testing.T) {
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	newer := base.Add(time.Hour)

	levels := map[string]int{"deep": 2, "top": 0}

	tests := []struct {
		name   string
		a, b   plugin.Task
		levels map[string]int // nil exercises the level-skip path
		want   bool
	}{
		{"level decides", newTask("top"), newTask("deep"), levels, true},
		{"priority before lane", newTask("z", withPriority(1)), newTask("a", withPriority(2)), nil, true},
		{"lane breaks priority tie", newTask("z", withStatus(plugin.StatusTodo)), newTask("a", withStatus(plugin.StatusDoing)), nil, true},
		{"unknown lane sorts last", newTask("z", withStatus("weird")), newTask("a", withStatus(plugin.StatusDone)), nil, false},
		{"newer first on full tie", newTask("a", withCreatedAt(newer)), newTask("b", withCreatedAt(base)), nil, true},
		{"id breaks created tie", newTask("a", withCreatedAt(base)), newTask("b", withCreatedAt(base)), nil, true},
		{"nil levels skips level", newTask("deep"), newTask("top"), nil, true}, // priority/lane/created equal, id decides
	}

	for _, test := range tests {
		if got := Less(test.a, test.b, test.levels); got != test.want {
			t.Errorf("%s: Less(a, b) = %v, want %v", test.name, got, test.want)
		}
	}
}
