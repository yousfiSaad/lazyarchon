# LazyArchon Architecture

## Overview

LazyArchon follows a simplified, concrete service architecture optimized for TUI (Terminal User Interface) applications. This architecture prioritizes clarity, maintainability, and follows Go idioms common in CLI tools.

## Architecture Principles

1. **Concrete Services** - Direct service implementations without interface abstraction
2. **Feature-Based Organization** - Code organized by business features, not technical layers
3. **Component-Based UI** - Bubble Tea components manage their own state
4. **Simple Coordination** - Coordinators handle complex multi-service workflows
5. **Configuration-Driven** - Behavior controlled by config, not interface swapping

## Directory Structure

```
internal/
├── features/                    # Feature-based organization
│   ├── tasks/
│   │   ├── services/           # Task business logic
│   │   ├── models/             # Task data models
│   │   ├── commands/           # Task commands
│   │   └── ui/                 # Task UI components
│   ├── projects/
│   │   ├── services/           # Project business logic
│   │   └── ui/                 # Project UI components
│   ├── search/
│   └── navigation/
├── app/
│   └── registry/               # Simple concrete service registry
├── shared/
│   ├── interfaces/             # Core interfaces (ArchonClient, Logger, etc.)
│   ├── config/                 # Configuration management
│   ├── styling/                # UI styling utilities
│   └── utils/                  # Shared utilities
└── ui/                         # Main UI coordination
    ├── components/             # Reusable UI components
    ├── coordinators/           # Multi-service workflow coordination
    └── model*.go               # Main Bubble Tea model
```

## Core Components

### 1. Feature Services

Each feature has a concrete service that handles business logic:

```go
// Direct, simple service usage
taskService := registry.TaskService()
tasks, err := taskService.ListTasks(ctx, projectID)
```

**Key Services:**
- `TaskService` - Task operations and business logic
- `ProjectService` - Project management
- `SearchService` - Search functionality
- `NavigationService` - Navigation history and state

### 2. Feature Registry

The `FeatureRegistry` provides centralized access to all services:

```go
type FeatureRegistry struct {
    taskService       *taskServices.TaskService
    projectService    *projectServices.ProjectService
    searchService     *searchServices.SearchService
    navigationService *navigationServices.NavigationService
    // ... commands
}

// Simple concrete access
func (r *FeatureRegistry) TaskService() *taskServices.TaskService {
    return r.taskService
}
```

**Benefits:**
- Thread-safe service access
- Centralized service lifecycle
- Simple dependency injection
- No interface overhead

### 3. UI Components

Bubble Tea components handle their own state and delegate to services:

```go
// Component uses services directly
type TaskListComponent struct {
    taskService *taskServices.TaskService
    tasks       []models.Task
}

func (c *TaskListComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case RefreshTasksMsg:
        tasks, err := c.taskService.ListTasks(ctx, msg.ProjectID)
        // ... handle response
    }
}
```

### 4. Coordinators

Coordinators handle complex workflows involving multiple services:

```go
type TaskCoordinator struct {
    taskService    *taskServices.TaskService
    projectService *projectServices.ProjectService
    logger         interfaces.Logger
}

func (c *TaskCoordinator) HandleTaskStatusChange(taskID, newStatus string) error {
    // Complex workflow involving multiple services
    task, err := c.taskService.GetTask(taskID)
    // ... validation, updates, notifications
}
```

## Data Flow

1. **User Input** → UI Components
2. **UI Components** → Services (via Registry)
3. **Services** → Archon API Client
4. **API Response** → Services → UI Components
5. **UI Components** → Update Display

## Service Patterns

### Direct Service Access
```go
// Get service from registry
taskService := registry.TaskService()

// Use service directly
tasks, err := taskService.ListTasks(ctx, projectID)
if err != nil {
    return handleError(err)
}

// Process results
return ProcessTasks(tasks)
```

### Multi-Service Coordination
```go
// Coordinator handles complex workflows
type ProjectTaskCoordinator struct {
    taskService    *taskServices.TaskService
    projectService *projectServices.ProjectService
}

func (c *ProjectTaskCoordinator) SwitchProject(projectID string) error {
    // Validate project exists
    project, err := c.projectService.GetProject(projectID)

    // Load project tasks
    tasks, err := c.taskService.ListTasks(ctx, projectID)

    // Coordinate UI updates
    return c.updateUI(project, tasks)
}
```

## Key Benefits

### 1. **Simplicity**
- One clear way to access services
- No adapter layer complexity
- Direct, obvious code paths

### 2. **Performance**
- No interface call overhead
- Direct method calls
- Minimal abstraction layers

### 3. **Maintainability**
- Easy to understand and debug
- Clear service responsibilities
- Simple dependency management

### 4. **Go Idiomatic**
- Follows standard Go CLI patterns
- Concrete types over interfaces
- Configuration over abstraction

### 5. **TUI Optimized**
- Optimized for single-implementation services
- Minimal memory overhead
- Fast startup and execution

## Service Implementation Example

```go
// Task Service - Simple and Direct
type TaskService struct {
    archonClient interfaces.ArchonClient
    logger       interfaces.Logger
}

func NewTaskService(client interfaces.ArchonClient, logger interfaces.Logger) *TaskService {
    return &TaskService{
        archonClient: client,
        logger:       logger,
    }
}

func (s *TaskService) ListTasks(ctx context.Context, projectID string) ([]models.Task, error) {
    s.logger.Debug("Fetching tasks", "project_id", projectID)

    // Direct API call
    response, err := s.archonClient.ListTasks(&projectID, nil, false)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch tasks: %w", err)
    }

    // Convert and return
    return s.convertTasks(response.Tasks), nil
}
```

## Testing Strategy

### Service Testing
```go
func TestTaskService_ListTasks(t *testing.T) {
    // Mock client
    mockClient := &mockArchonClient{}
    logger := &mockLogger{}

    // Create service
    service := NewTaskService(mockClient, logger)

    // Test behavior
    tasks, err := service.ListTasks(ctx, "project-123")

    // Assertions
    assert.NoError(t, err)
    assert.Len(t, tasks, 3)
}
```

### Component Testing
```go
func TestTaskListComponent(t *testing.T) {
    // Create component with mock service
    mockService := &mockTaskService{}
    component := NewTaskListComponent(mockService)

    // Test message handling
    _, cmd := component.Update(RefreshTasksMsg{ProjectID: "123"})

    // Verify behavior
    assert.NotNil(t, cmd)
}
```

## Migration Benefits

From the previous interface-based approach, this simplified architecture provides:

- **50% less code** - Removed entire adapter layer (~500 lines)
- **Faster development** - One clear pattern to follow
- **Better debugging** - Direct call paths with no adapter indirection
- **Go idiomatic** - Follows patterns from successful CLI tools (gh-dash, lazygit)
- **Reduced complexity** - No interface/adapter confusion

## Future Extensions

The architecture supports future enhancements:

1. **Plugin System** - Add concrete plugin services to registry
2. **Caching Layer** - Add caching services without interface changes
3. **Multiple Backends** - Configuration-driven backend selection
4. **Enhanced Coordination** - Add more sophisticated coordinators
5. **Performance Monitoring** - Add instrumentation to concrete services

This architecture provides a solid foundation for TUI applications while maintaining simplicity and Go idioms.