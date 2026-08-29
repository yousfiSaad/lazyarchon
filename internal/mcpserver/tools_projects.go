package mcpserver

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/yousfisaad/lazyarchon/v2/internal/plugin"
)

func (s *Server) registerProjectTools() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "list_projects",
		Description: "List all projects as compact summaries (id, title, pinned). Use get_project for full details.",
	}, s.listProjects)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_project",
		Description: "Get one project by ID, including description, color and extra fields.",
	}, s.getProject)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "create_project",
		Description: "Create a new project. Only title is required.",
	}, s.createProject)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "update_project",
		Description: "Update a project. Omitted fields are left unchanged; pass a field to change it.",
	}, s.updateProject)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name: "delete_project",
		Description: "Delete a project. Refuses while the project still has tasks; pass " +
			"delete_tasks=true to delete its tasks along with it.",
	}, s.deleteProject)
}

// ---------- list_projects ----------

type listProjectsIn struct{}

type listProjectsOut struct {
	Projects []projectSummary `json:"projects"`
	Total    int              `json:"total"`
}

func (s *Server) listProjects(ctx context.Context, _ *mcp.CallToolRequest, _ listProjectsIn) (*mcp.CallToolResult, listProjectsOut, error) {
	projects, err := s.client.ListProjects(ctx)
	if err != nil {
		return nil, listProjectsOut{}, toolError(s.backend, err)
	}

	out := listProjectsOut{Projects: make([]projectSummary, 0, len(projects)), Total: len(projects)}
	for _, project := range projects {
		out.Projects = append(out.Projects, toProjectSummary(project))
	}

	return nil, out, nil
}

// ---------- get_project ----------

type getProjectIn struct {
	ProjectID string `json:"project_id" jsonschema:"ID of the project"`
}

type getProjectOut struct {
	Project projectDetail `json:"project"`
}

func (s *Server) getProject(ctx context.Context, _ *mcp.CallToolRequest, in getProjectIn) (*mcp.CallToolResult, getProjectOut, error) {
	project, err := s.client.GetProject(ctx, in.ProjectID)
	if err != nil {
		return nil, getProjectOut{}, toolError(s.backend, err)
	}

	return nil, getProjectOut{Project: toProjectDetail(*project)}, nil
}

// ---------- create_project ----------

type createProjectIn struct {
	Title       string `json:"title" jsonschema:"Name of the new project"`
	Description string `json:"description,omitempty" jsonschema:"Optional description"`
	Color       string `json:"color,omitempty" jsonschema:"Optional color or theme"`
	Pinned      bool   `json:"pinned,omitempty" jsonschema:"Pin the project to the top of lists"`
}

type createProjectOut struct {
	Project projectDetail `json:"project"`
}

func (s *Server) createProject(ctx context.Context, _ *mcp.CallToolRequest, in createProjectIn) (*mcp.CallToolResult, createProjectOut, error) {
	project, err := s.client.CreateProject(ctx, plugin.CreateProjectRequest{
		Title:       in.Title,
		Description: in.Description,
		Color:       in.Color,
		Pinned:      in.Pinned,
	})
	if err != nil {
		return nil, createProjectOut{}, toolError(s.backend, err)
	}

	return nil, createProjectOut{Project: toProjectDetail(*project)}, nil
}

// ---------- update_project ----------

type updateProjectIn struct {
	ProjectID   string  `json:"project_id" jsonschema:"ID of the project to update"`
	Title       *string `json:"title,omitempty" jsonschema:"New title"`
	Description *string `json:"description,omitempty" jsonschema:"New description"`
	Color       *string `json:"color,omitempty" jsonschema:"New color"`
	Pinned      *bool   `json:"pinned,omitempty" jsonschema:"New pinned state"`
}

type updateProjectOut struct {
	Project projectDetail `json:"project"`
}

func (s *Server) updateProject(ctx context.Context, _ *mcp.CallToolRequest, in updateProjectIn) (*mcp.CallToolResult, updateProjectOut, error) {
	project, err := s.client.UpdateProject(ctx, in.ProjectID, plugin.UpdateProjectRequest{
		Title:       in.Title,
		Description: in.Description,
		Color:       in.Color,
		Pinned:      in.Pinned,
	})
	if err != nil {
		return nil, updateProjectOut{}, toolError(s.backend, err)
	}

	return nil, updateProjectOut{Project: toProjectDetail(*project)}, nil
}

// ---------- delete_project ----------

type deleteProjectIn struct {
	ProjectID   string `json:"project_id" jsonschema:"ID of the project to delete"`
	DeleteTasks bool   `json:"delete_tasks,omitempty" jsonschema:"Also delete the project's tasks (required when it still has any)"`
}

type deleteProjectOut struct {
	Deleted      bool `json:"deleted"`
	TasksDeleted int  `json:"tasks_deleted"`
}

func (s *Server) deleteProject(ctx context.Context, _ *mcp.CallToolRequest, in deleteProjectIn) (*mcp.CallToolResult, deleteProjectOut, error) {
	// Count tasks (including archived) up front so we can guard the delete
	// and report how many went away with the project.
	result, err := s.client.ListTasks(ctx, plugin.TaskFilters{
		ProjectID:       &in.ProjectID,
		IncludeArchived: true,
		Limit:           1,
	})
	if err != nil {
		return nil, deleteProjectOut{}, toolError(s.backend, err)
	}

	if result.TotalCount > 0 && !in.DeleteTasks {
		return nil, deleteProjectOut{}, fmt.Errorf(
			"project %s still has %d task(s); pass delete_tasks=true to delete them along with the project",
			in.ProjectID, result.TotalCount,
		)
	}

	if err := s.client.DeleteProject(ctx, in.ProjectID); err != nil {
		return nil, deleteProjectOut{}, toolError(s.backend, err)
	}

	return nil, deleteProjectOut{Deleted: true, TasksDeleted: result.TotalCount}, nil
}
