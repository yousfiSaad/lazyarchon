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
}

// New creates an MCP server serving the given backend client and registers
// all task and project tools.
func New(client plugin.TaskClient, backend, version string) *Server {
	server := &Server{
		mcpServer: mcp.NewServer(
			&mcp.Implementation{Name: "lazyarchon", Version: version},
			&mcp.ServerOptions{Instructions: instructions(backend)},
		),
		client:  client,
		backend: backend,
	}

	server.registerProjectTools()
	server.registerTaskTools()

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
- Statuses: "todo", "doing", "review", "done".
- Priority: 1 = critical, 2 = high, 3 = medium, 4 = low (lower is more urgent).
- Tasks nest via parent_id; children survive parent deletion.
- Due dates accept "YYYY-MM-DD" or RFC3339. On update_task, an empty string clears the value.
- List tools return compact summaries; get/create/update tools return full objects.

Start with list_projects to discover project IDs.`, backend)
}
