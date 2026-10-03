package mcpserver

import (
	"context"
	"sort"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/yousfisaad/lazyarchon/v2/internal/plugin"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/deps"
)

func (s *Server) registerDependencyTools() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name: "get_execution_plan",
		Description: "Compute dependency waves of parallelizable work. Loads every task " +
			"(paginated), collapses done blockers, and groups open (non-done, non-archived) " +
			"tasks into waves: wave 0 tasks have no unfinished blockers and can start now; " +
			"wave N opens once earlier waves finish. waves[i] is level i — an empty wave " +
			"means that level has no tasks in scope. project_id filters wave entries only; " +
			"blockers in other projects still shape the levels. Also suggests where to " +
			"start. Only the local backend records dependencies; other backends return one " +
			"flat wave.",
	}, s.getExecutionPlan)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name: "get_task_chain",
		Description: "Show one task's transitive dependency chain: unfinished blockers ahead " +
			"(with BFS depth; done blockers are omitted since they no longer gate anything), " +
			"all dependents downstream regardless of status, whether the task is ready now " +
			"(no unfinished direct blocker), its wave level, and the length of the longest " +
			"unfinished blocker path including the task itself.",
	}, s.getTaskChain)
}

// ---------- get_execution_plan ----------

type executionPlanIn struct {
	ProjectID string `json:"project_id,omitempty" jsonschema:"Only include wave entries from this project; blockers in other projects still shape the waves"`
}

type planTask struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	Priority  int    `json:"priority"`
	ProjectID string `json:"project_id,omitempty"`
}

type executionPlanOut struct {
	Waves          [][]planTask `json:"waves"`                     // index = level; gap waves stay empty
	TaskCount      int          `json:"task_count"`                // entries across waves
	SuggestedStart *planTask    `json:"suggested_start,omitempty"` // omitted when no level-0 open task is in scope
	Cycles         []string     `json:"cycles,omitempty"`
}

func (s *Server) getExecutionPlan(ctx context.Context, _ *mcp.CallToolRequest, in executionPlanIn) (*mcp.CallToolResult, executionPlanOut, error) {
	projectID := optionalString(in.ProjectID)

	all, err := s.loadAllTasks(ctx)
	if err != nil {
		return nil, executionPlanOut{}, err
	}

	graph := deps.Build(all)
	levels := graph.Levels()

	// Wave entries are open tasks (done and archived tasks never appear,
	// though an archived-undone blocker still shaped the levels).
	open := make([]plugin.Task, 0, len(all))
	for _, task := range all {
		if task.Archived || task.Status == plugin.StatusDone {
			continue
		}
		if projectID != nil && task.ProjectID != *projectID {
			continue
		}
		open = append(open, task)
	}

	// Sort first, then bucket: each wave inherits the shared determinism
	// order (level, priority, status lane, created DESC, id).
	sort.Slice(open, func(i, j int) bool { //nolint:varnamelen // i, j are idiomatic for sort functions
		return deps.Less(open[i], open[j], levels)
	})

	maxLevel := -1
	for _, task := range open {
		maxLevel = max(maxLevel, levels[task.ID])
	}

	// Every slot is pre-initialized so gap waves marshal as [], not null.
	waves := make([][]planTask, maxLevel+1)
	for i := range waves {
		waves[i] = make([]planTask, 0)
	}
	for _, task := range open {
		waves[levels[task.ID]] = append(waves[levels[task.ID]], toPlanTask(task))
	}

	out := executionPlanOut{
		Waves:     waves,
		TaskCount: len(open),
		Cycles:    graph.Cycles(),
	}
	if suggested := pickSuggestedStart(open, levels); suggested != nil {
		out.SuggestedStart = suggested
	}

	return nil, out, nil
}

// pickSuggestedStart chooses the deterministic "work on this next" task:
// among level-0 open tasks prefer todo, then doing (resume), then review;
// within a lane priority asc, created DESC, id ASC. Lane-first on purpose —
// unlike deps.Less, which is a display order, this ranks what to act on.
func pickSuggestedStart(open []plugin.Task, levels map[string]int) *planTask {
	var best *plugin.Task
	for i, task := range open {
		if levels[task.ID] != 0 {
			continue
		}
		if best == nil || suggestedLess(task, *best) {
			best = &open[i]
		}
	}

	if best == nil {
		return nil
	}
	suggested := toPlanTask(*best)
	return &suggested
}

// suggestedLess is the suggested-start ranking: lane before priority.
//
//nolint:varnamelen // a, b are idiomatic comparator operands
func suggestedLess(a, b plugin.Task) bool {
	if laneA, laneB := startLane(a.Status), startLane(b.Status); laneA != laneB {
		return laneA < laneB
	}
	if a.Priority != b.Priority {
		return a.Priority < b.Priority
	}
	if !a.CreatedAt.Equal(b.CreatedAt) {
		return a.CreatedAt.After(b.CreatedAt)
	}
	return a.ID < b.ID
}

// startLane ranks statuses for "what to pick up": new work first, then
// resuming, then verification.
func startLane(status string) int {
	switch status {
	case plugin.StatusTodo:
		return 0
	case plugin.StatusDoing:
		return 1
	case plugin.StatusReview:
		return 2
	default:
		return 3
	}
}

// ---------- get_task_chain ----------

type taskChainIn struct {
	TaskID string `json:"task_id" jsonschema:"ID of the task whose chain to inspect"`
}

type chainEntry struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Priority int    `json:"priority"`
	Depth    int    `json:"depth"`  // 1 = direct edge
	Direct   bool   `json:"direct"` // Depth == 1
}

type taskChainOut struct {
	Task                  taskSummary  `json:"task"`
	Ready                 bool         `json:"ready"`
	Level                 int          `json:"level"`
	Blockers              []chainEntry `json:"blockers"`                // unfinished transitive blockers, depth-sorted
	Dependents            []chainEntry `json:"dependents"`              // all statuses, depth-sorted
	LongestRemainingChain int          `json:"longest_remaining_chain"` // node count including the task
	Cycles                []string     `json:"cycles,omitempty"`        // only cycles touching this task's upstream
}

func (s *Server) getTaskChain(ctx context.Context, _ *mcp.CallToolRequest, in taskChainIn) (*mcp.CallToolResult, taskChainOut, error) {
	task, err := s.client.GetTask(ctx, in.TaskID)
	if err != nil {
		return nil, taskChainOut{}, toolError(s.backend, err)
	}

	all, err := s.loadAllTasks(ctx)
	if err != nil {
		return nil, taskChainOut{}, err
	}

	// Defensive: the graph must contain the queried task even if the list
	// snapshot missed it (a concurrent delete and recreate, say).
	listed := false
	for _, candidate := range all {
		if candidate.ID == task.ID {
			listed = true
			break
		}
	}
	if !listed {
		all = append(all, *task)
	}

	graph := deps.Build(all)
	levels := graph.Levels()

	blockers := graph.Upstream(task.ID)
	dependents := graph.Downstream(task.ID)

	out := taskChainOut{
		Task:                  toTaskSummary(*task),
		Ready:                 graph.Ready(task.ID),
		Level:                 levels[task.ID],
		Blockers:              make([]chainEntry, 0, len(blockers)),
		Dependents:            make([]chainEntry, 0, len(dependents)),
		LongestRemainingChain: graph.ChainLength(task.ID),
	}

	for _, node := range blockers {
		out.Blockers = append(out.Blockers, toChainEntry(node))
	}
	for _, node := range dependents {
		out.Dependents = append(out.Dependents, toChainEntry(node))
	}

	// Only cycles touching this task's upstream (or the task itself) are
	// relevant; unrelated cycles elsewhere must not pollute the response.
	inChain := make(map[string]bool, len(blockers)+1)
	inChain[task.ID] = true
	for _, node := range blockers {
		inChain[node.Task.ID] = true
	}
	for _, id := range graph.Cycles() {
		if inChain[id] {
			out.Cycles = append(out.Cycles, id)
		}
	}

	return nil, out, nil
}

// ---------- shared helpers ----------

func toPlanTask(task plugin.Task) planTask {
	return planTask{
		ID:        task.ID,
		Title:     task.Title,
		Status:    task.Status,
		Priority:  task.Priority,
		ProjectID: task.ProjectID,
	}
}

func toChainEntry(node deps.ChainNode) chainEntry {
	return chainEntry{
		ID:       node.Task.ID,
		Title:    node.Task.Title,
		Status:   node.Task.Status,
		Priority: node.Task.Priority,
		Depth:    node.Depth,
		Direct:   node.Direct,
	}
}

// loadAllTasks pages through ListTasks to completion so graph derivation
// sees every edge (the default page is 500). IncludeArchived matters twice:
// archived-undone tasks still block, and statuses must be readable for the
// done-collapse.
func (s *Server) loadAllTasks(ctx context.Context) ([]plugin.Task, error) {
	var all []plugin.Task
	offset := 0
	for {
		result, err := s.client.ListTasks(ctx, plugin.TaskFilters{
			IncludeArchived: true,
			Limit:           500,
			Offset:          offset,
		})
		if err != nil {
			return nil, toolError(s.backend, err)
		}

		all = append(all, result.Tasks...)
		if !result.HasMore || len(result.Tasks) == 0 {
			break // empty page despite has_more: defensive against a broken backend
		}
		offset = result.NextOffset
	}

	return all, nil
}
