// Package mcpserver exposes a plugin.TaskClient as a Model Context Protocol
// server over stdio, so LLM clients (Claude, etc.) can manage tasks in the
// same database the TUI uses.
package mcpserver

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/yousfisaad/lazyarchon/v2/internal/plugin"
)

// Server wraps an MCP server bound to one task backend client.
type Server struct {
	mcpServer *mcp.Server
	client    plugin.TaskClient
	backend   string
	worktree  string
}

// New creates an MCP server serving the given backend client and registers
// all task and project tools. The worktree is the git worktree the server
// process runs in ("" for the main checkout); it stamps created tasks and
// tasks moved to doing/review so parallel sessions can see who owns what.
func New(client plugin.TaskClient, backend, version, worktree string) *Server {
	server := &Server{
		mcpServer: mcp.NewServer(
			&mcp.Implementation{Name: "lazyarchon", Version: version},
			&mcp.ServerOptions{Instructions: instructions(backend)},
		),
		client:   client,
		backend:  backend,
		worktree: worktree,
	}

	server.registerProjectTools()
	server.registerTaskTools()
	server.registerDependencyTools()

	return server
}

// Run serves the MCP protocol over stdio until the context is canceled.
func (s *Server) Run(ctx context.Context) error {
	return s.mcpServer.Run(ctx, &mcp.StdioTransport{})
}

// MCP returns the underlying SDK server (used by tests to connect over
// in-memory transports).
func (s *Server) MCP() *mcp.Server {
	return s.mcpServer
}

func instructions(backend string) string {
	return fmt.Sprintf(`Task manager backed by the %q plugin.

Conventions:
- Status lifecycle: tasks start "todo", move to "doing" when you start work, "review" when the
  work needs verification, and "done" when finished.
- Priority: 1 = critical, 2 = high, 3 = medium, 4 = low, 5 = backlog (lower is more urgent).
- Tasks nest via parent_id; children survive parent deletion.
- Tasks can declare blockers via blocked_by; list_tasks with ready=true keeps only tasks whose
  blockers are all done. Blocking is advisory by default; when tasks.enforce_dependencies is
  enabled, moving a blocked task to doing/review is rejected with the blockers named.
- get_execution_plan groups open tasks into dependency waves (wave 0 = ready now) and suggests a
  start; get_task_chain shows one task's transitive blockers and dependents.
- Due dates accept "YYYY-MM-DD" or RFC3339. On update_task, an empty string clears the value.
- Tasks record the git worktree of the session that created them or moved them to doing/review;
  pass "worktree" explicitly to override, or "" on update_task to clear.
- List tools return compact summaries; get/create/update tools return full objects.

Typical session: list_projects to discover IDs, then list_tasks (paginate with next_offset).
create_task for new work, update_task with status "doing" once you start and "done" when
finished.`, backend)
}
