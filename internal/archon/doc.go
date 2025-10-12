// Package archon provides the client library for interacting with the Archon task management API.
//
// This package implements a comprehensive HTTP client with real-time WebSocket capabilities
// for the Archon task management system. It provides both basic and resilient client
// implementations with features like automatic retry, circuit breaking, and real-time updates.
//
// # Core Components
//
// The package is organized around several key components:
//
//   - Client: Basic HTTP client for Archon API operations
//   - ResilientClient: Enhanced client with retry logic and circuit breaker
//   - WebSocketClient: Real-time client using Supabase WebSocket protocol
//   - Models: Data structures representing tasks, projects, and API responses
//
// # Basic Usage
//
// Creating a basic client and fetching tasks:
//
//	client := archon.NewClient("https://api.archon.com", "your-api-key")
//	response, err := client.ListTasks(nil, nil, false)
//	if err != nil {
//		log.Fatal(err)
//	}
//	for _, task := range response.Tasks {
//		fmt.Printf("Task: %s - %s\n", task.Title, task.Status)
//	}
//
// # Resilient Client
//
// For production environments, use the resilient client which provides
// automatic retry and circuit breaker capabilities:
//
//	client := archon.NewResilientClient("https://api.archon.com", "your-api-key")
//	response, err := client.ListTasks(nil, nil, false)
//	// Automatically retries on transient failures
//
// # Real-time Updates
//
// The WebSocket client provides real-time updates for tasks and projects:
//
//	wsClient := archon.NewWebSocketClient("https://api.archon.com", "your-api-key")
//	wsClient.SetEventHandlers(
//		func(event archon.TaskUpdateEvent) {
//			fmt.Printf("Task updated: %s\n", event.Task.Title)
//		},
//		func(event archon.TaskCreateEvent) {
//			fmt.Printf("Task created: %s\n", event.Task.Title)
//		},
//		// ... other handlers
//	)
//	if err := wsClient.Connect(); err != nil {
//		log.Fatal(err)
//	}
//
// # Error Handling
//
// The package provides typed errors for common scenarios:
//
//	response, err := client.GetTask("non-existent-id")
//	if errors.Is(err, archon.ErrTaskNotFound) {
//		// Handle task not found specifically
//	}
//
// # Configuration and Resilience
//
// The resilient client can be configured with custom retry and circuit breaker settings:
//
//	config := archon.ResilienceConfig{
//		MaxRetries:    5,
//		InitialDelay:  time.Second,
//		MaxDelay:      30 * time.Second,
//		// ... other settings
//	}
//	client := archon.NewResilientClientWithConfig("https://api.archon.com", "your-api-key", config)
//
// # Thread Safety
//
// All client implementations are thread-safe and can be used concurrently
// from multiple goroutines. The WebSocket client handles connection state
// with appropriate synchronization.
//
// # Testing Support
//
// The package provides mock implementations and test utilities:
//
//	mockClient := archon.NewMockClient()
//	mockClient.SetListTasksResponse(&archon.TasksResponse{...}, nil)
//	// Use mockClient in tests
//
// For more detailed examples and API documentation, see the individual type
// and method documentation below.
package archon
