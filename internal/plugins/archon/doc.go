// Package archon provides the client library for interacting with the Archon task management API.
//
// This package implements a comprehensive HTTP client for the Archon task management system.
//
// # Core Components
//
// The package is organized around several key components:
//
//   - Client: HTTP client for Archon API operations
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
// # Error Handling
//
// The package provides typed errors for common scenarios:
//
//	response, err := client.GetTask("non-existent-id")
//	if errors.Is(err, archon.ErrTaskNotFound) {
//		// Handle task not found specifically
//	}
//
// # Thread Safety
//
// The client implementation is thread-safe and can be used concurrently
// from multiple goroutines.
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
