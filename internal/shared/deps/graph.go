// Package deps derives deterministic dependency views — leveled waves,
// transitive chains and cycle detection — from a snapshot of tasks. It is
// pure: no I/O and no state beyond the snapshot, so the MCP server and the
// TUI share one set of semantics. Everything is computed over "unfinished"
// edges (blocker status != done; unknown blocker IDs count as unfinished),
// matching the readiness convention used by the local backend's SQL, the
// TUI ready filter and context.UnfinishedBlockers.
package deps

import (
	"slices"
	"sort"

	"github.com/yousfisaad/lazyarchon/v2/internal/plugin"
)

// ChainNode is one task in a transitive closure, with its BFS distance from
// the queried task.
type ChainNode struct {
	Task   plugin.Task
	Depth  int  // 1 = direct edge
	Direct bool // Depth == 1, kept as a convenience flag for output shapes
}

// Graph is a leveled view over one snapshot of tasks. Build computes the
// leveling eagerly; closure methods run BFS on demand and ChainLength
// memoizes lazily, so a Graph is not safe for concurrent use.
type Graph struct {
	byID       map[string]plugin.Task
	uBlockers  map[string][]string // known unfinished blocker IDs (status != done)
	hasUnknown map[string]bool     // has an edge to an ID absent from the snapshot
	depUnfin   map[string][]string // reverse of uBlockers
	depAll     map[string][]string // reverse of every blocked_by edge, any status
	levels     map[string]int
	cycles     []string
	chainMemo  map[string]int
}

// Build constructs the graph. Duplicate IDs: last one wins. Done and
// archived tasks participate (their statuses matter); they simply never
// become wave entries — that filtering belongs to callers.
func Build(tasks []plugin.Task) *Graph {
	graph := &Graph{
		byID:       make(map[string]plugin.Task, len(tasks)),
		uBlockers:  make(map[string][]string),
		hasUnknown: make(map[string]bool),
		depUnfin:   make(map[string][]string),
		depAll:     make(map[string][]string),
		levels:     make(map[string]int, len(tasks)),
		chainMemo:  make(map[string]int),
	}

	for _, task := range tasks {
		graph.byID[task.ID] = task
	}

	for _, task := range tasks {
		for _, blockerID := range task.BlockedBy {
			graph.depAll[blockerID] = append(graph.depAll[blockerID], task.ID)

			blocker, known := graph.byID[blockerID]
			if !known {
				graph.hasUnknown[task.ID] = true
				continue
			}
			if blocker.Status != plugin.StatusDone {
				graph.uBlockers[task.ID] = append(graph.uBlockers[task.ID], blockerID)
				graph.depUnfin[blockerID] = append(graph.depUnfin[blockerID], task.ID)
			}
		}
	}

	graph.levelGraph()
	graph.computeCycles()

	return graph
}

// levelGraph assigns wave levels via Kahn's algorithm turned into a
// longest-path computation: a task enters the queue only once every known
// unfinished blocker is leveled, so its own level is one past the deepest.
func (g *Graph) levelGraph() {
	pending := make(map[string]int, len(g.uBlockers))
	queue := make([]string, 0, len(g.byID))
	for id, blockers := range g.uBlockers {
		pending[id] = len(blockers)
	}
	for id := range g.byID {
		if pending[id] == 0 {
			queue = append(queue, id)
		}
	}

	maxLevel := -1
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]

		level := 0
		if g.hasUnknown[id] {
			// An unknown blocker counts as an unfinished virtual leaf at
			// level 0, pushing this task to at least level 1.
			level = 1
		}
		for _, blocker := range g.uBlockers[id] {
			level = max(level, g.levels[blocker]+1)
		}

		g.levels[id] = level
		maxLevel = max(maxLevel, level)

		for _, dependent := range g.depUnfin[id] {
			pending[dependent]--
			if pending[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	// Never-popped nodes are cycle members plus everything downstream of a
	// cycle. They land in a single phantom final wave (never level 0, so
	// they cannot masquerade as ready); Cycles names the culprits.
	stuckLevel := max(maxLevel+1, 1)
	for id := range g.byID {
		if _, leveled := g.levels[id]; !leveled {
			g.levels[id] = stuckLevel
		}
	}
}

// computeCycles finds the tasks sitting on a dependency cycle in the
// unfinished subgraph: strongly connected components of size > 1, or a
// self-loop. Empty in normal operation — writes are cycle-guarded; this is
// defense in depth over possibly corrupt data.
//
//nolint:gocyclo // Tarjan's algorithm is inherently branch-heavy
func (g *Graph) computeCycles() {
	index := 0
	indices := make(map[string]int, len(g.byID))
	lowlink := make(map[string]int, len(g.byID))
	onStack := make(map[string]bool, len(g.byID))
	stack := make([]string, 0)

	ids := make([]string, 0, len(g.byID))
	for id := range g.byID {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	var strongconnect func(current string)
	strongconnect = func(current string) {
		indices[current] = index
		lowlink[current] = index
		index++
		stack = append(stack, current)
		onStack[current] = true

		for _, next := range g.uBlockers[current] {
			if _, seen := indices[next]; !seen {
				strongconnect(next)
				if lowlink[next] < lowlink[current] {
					lowlink[current] = lowlink[next]
				}
			} else if onStack[next] && indices[next] < lowlink[current] {
				lowlink[current] = indices[next]
			}
		}

		if lowlink[current] != indices[current] {
			return
		}

		component := make([]string, 0, 2)
		for {
			member := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			onStack[member] = false
			component = append(component, member)
			if member == current {
				break
			}
		}

		if len(component) > 1 {
			g.cycles = append(g.cycles, component...)
			return
		}
		if slices.Contains(g.uBlockers[current], current) {
			g.cycles = append(g.cycles, current) // self-loop
		}
	}

	for _, id := range ids {
		if _, seen := indices[id]; !seen {
			strongconnect(id)
		}
	}

	sort.Strings(g.cycles)
}

// Cycles returns the sorted IDs of all tasks on a dependency cycle in the
// unfinished subgraph. Empty in normal operation — writes are cycle-guarded;
// this is defense in depth.
func (g *Graph) Cycles() []string {
	return g.cycles
}

// Levels returns the wave level of every task in the snapshot.
//
//	level(t) = 0                                    if t has no unfinished direct blocker
//	level(t) = 1 + max(level of unfinished blockers) otherwise
//
// Done blockers collapse entirely, so level 0 means the task is ready (no
// unfinished direct blocker) — the same predicate as the ready filter.
func (g *Graph) Levels() map[string]int {
	return g.levels
}

// Ready reports whether id has no unfinished direct blocker. IDs absent
// from the snapshot are conservatively not ready.
func (g *Graph) Ready(id string) bool {
	level, known := g.levels[id]
	return known && level == 0
}

// ChainLength returns the node count of the longest path of unfinished
// blockers ahead of id, INCLUDING id itself (a task with one direct
// unfinished blocker returns 2; a ready task returns 1). Unknown id returns
// 0. Inside a cycle the value is best-effort (back-edges contribute 0) but
// always terminates.
func (g *Graph) ChainLength(id string) int {
	if _, known := g.byID[id]; !known {
		return 0
	}
	return g.chainLength(id, make(map[string]bool))
}

// chainLength is the recursive core of ChainLength; onStack breaks the
// recursion on back-edges (cycle members) by contributing 0.
func (g *Graph) chainLength(id string, onStack map[string]bool) int {
	if cached, memoized := g.chainMemo[id]; memoized {
		return cached
	}
	if onStack[id] {
		return 0
	}

	onStack[id] = true
	length := 1
	if g.hasUnknown[id] {
		length = 2 // one virtual unfinished node ahead
	}
	for _, blocker := range g.uBlockers[id] {
		length = max(length, g.chainLength(blocker, onStack)+1)
	}
	onStack[id] = false

	g.chainMemo[id] = length
	return length
}

// Upstream returns the transitive blockers of id over unfinished edges only
// (done blockers are omitted — they no longer gate anything). Depth is the
// BFS distance; entries are sorted by (Depth asc, then the Less chain).
// Unknown blocker IDs cannot be represented and are skipped silently.
func (g *Graph) Upstream(id string) []ChainNode {
	return g.closure(id, g.uBlockers)
}

// Downstream returns every task transitively blocked by id, regardless of
// statuses (mirrors the raw Blocks semantics of GetTask and the TUI details
// panel). Same ordering rule as Upstream.
func (g *Graph) Downstream(id string) []ChainNode {
	return g.closure(id, g.depAll)
}

// closure runs a plain BFS from start over the given edge map (cycle-safe
// by construction via the visited set) and sorts the result.
func (g *Graph) closure(start string, edges map[string][]string) []ChainNode {
	if _, known := g.byID[start]; !known {
		return nil
	}

	depths := map[string]int{start: 0}
	visited := map[string]bool{start: true}
	queue := []string{start}
	nodes := make([]ChainNode, 0)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, next := range edges[current] {
			if visited[next] {
				continue
			}
			visited[next] = true
			depths[next] = depths[current] + 1
			queue = append(queue, next)
			nodes = append(nodes, ChainNode{
				Task:   g.byID[next],
				Depth:  depths[next],
				Direct: depths[next] == 1,
			})
		}
	}

	sort.Slice(nodes, func(i, j int) bool { //nolint:varnamelen // i, j are idiomatic for sort functions
		if nodes[i].Depth != nodes[j].Depth {
			return nodes[i].Depth < nodes[j].Depth
		}
		return Less(nodes[i].Task, nodes[j].Task, g.levels)
	})

	return nodes
}

// Less is the shared determinism comparator for dependency views: level
// asc, then priority asc, then status lane (todo < doing < review < done <
// unknown), then created_at DESC, then id ASC. The trailing id comparison
// makes the order total (IDs are unique), so sort.Slice under Less is
// deterministic. A missing level entry counts as 0; a nil map means every
// task is level 0.
//
//nolint:varnamelen // a, b are idiomatic comparator operands
func Less(a, b plugin.Task, levels map[string]int) bool {
	if levels != nil {
		if levelA, levelB := levels[a.ID], levels[b.ID]; levelA != levelB {
			return levelA < levelB
		}
	}

	if a.Priority != b.Priority {
		return a.Priority < b.Priority
	}

	if laneA, laneB := statusLane(a.Status), statusLane(b.Status); laneA != laneB {
		return laneA < laneB
	}

	if !a.CreatedAt.Equal(b.CreatedAt) {
		return a.CreatedAt.After(b.CreatedAt) // newest first
	}

	return a.ID < b.ID
}

// statusLane maps a status onto the display order used across the TUI
// (todo first — it needs action — through done last).
func statusLane(status string) int {
	switch status {
	case plugin.StatusTodo:
		return 0
	case plugin.StatusDoing:
		return 1
	case plugin.StatusReview:
		return 2
	case plugin.StatusDone:
		return 3
	default:
		return 4
	}
}
