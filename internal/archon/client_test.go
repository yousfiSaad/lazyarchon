package archon

import (
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	baseURL := "http://localhost:8181"
	apiKey := "test-key"

	client := NewClient(baseURL, apiKey)

	if client.baseURL != baseURL {
		t.Errorf("Expected baseURL %s, got %s", baseURL, client.baseURL)
	}

	if client.apiKey != apiKey {
		t.Errorf("Expected apiKey %s, got %s", apiKey, client.apiKey)
	}

	if client.httpClient == nil {
		t.Error("Expected httpClient to be initialized")
	}

	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", client.httpClient.Timeout)
	}
}

func TestClient_ListTasks(t *testing.T) {
	server := SetupMockServerWithData()
	defer server.Close()

	client := NewClient(server.URL, "test-key")

	tests := []struct {
		name          string
		projectID     *string
		status        *string
		includeClosed bool
		expectError   bool
		expectedCount int
	}{
		{
			name:          "list all tasks",
			projectID:     nil,
			status:        nil,
			includeClosed: true,
			expectError:   false,
			expectedCount: 9, // All sample tasks
		},
		{
			name:          "list tasks without closed",
			projectID:     nil,
			status:        nil,
			includeClosed: false,
			expectError:   false,
			expectedCount: 8, // All except done tasks
		},
		{
			name:          "filter by status",
			projectID:     nil,
			status:        stringPtr("todo"),
			includeClosed: true,
			expectError:   false,
			expectedCount: 5, // Todo tasks in sample data (including HighPriorityTask, AuthenticationTask, etc.)
		},
		{
			name:          "filter by project",
			projectID:     stringPtr("test-project-1"),
			status:        nil,
			includeClosed: true,
			expectError:   false,
			expectedCount: 9, // All sample tasks have this project ID by default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.ListTasks(tt.projectID, tt.status, tt.includeClosed)

			if tt.expectError {
				AssertError(t, err)
				return
			}

			AssertNoError(t, err)

			if resp == nil {
				t.Fatal("Expected response, got nil")
			}

			if resp.Count != tt.expectedCount {
				t.Errorf("Expected count %d, got %d", tt.expectedCount, resp.Count)
			}

			if len(resp.Tasks) != tt.expectedCount {
				t.Errorf("Expected %d tasks, got %d", tt.expectedCount, len(resp.Tasks))
			}

			// Validate filter was applied
			if tt.status != nil {
				for _, task := range resp.Tasks {
					if task.Status != *tt.status {
						t.Errorf("Expected all tasks to have status %s, found %s", *tt.status, task.Status)
					}
				}
			}

			if tt.projectID != nil {
				for _, task := range resp.Tasks {
					if task.ProjectID != *tt.projectID {
						t.Errorf("Expected all tasks to have project ID %s, found %s", *tt.projectID, task.ProjectID)
					}
				}
			}

			if !tt.includeClosed {
				for _, task := range resp.Tasks {
					if task.Status == "done" {
						t.Error("Expected no done tasks when includeClosed=false")
					}
				}
			}
		})
	}
}

func TestClient_ListTasks_Error(t *testing.T) {
	server := NewMockServer()
	defer server.Close()

	// Configure server to return error
	server.SetSimulatedError("tasks", SimulateNetworkError())

	client := NewClient(server.URL, "test-key")

	_, err := client.ListTasks(nil, nil, true)
	AssertError(t, err)
	AssertErrorContains(t, err, "network error")
}

func TestClient_GetTask(t *testing.T) {
	server := SetupMockServerWithData()
	defer server.Close()

	client := NewClient(server.URL, "test-key")

	// Add a known task
	expectedTask := NewTaskBuilder().
		WithID("known-task").
		WithTitle("Known Task").
		Build()
	server.AddTask(expectedTask)

	t.Run("get existing task", func(t *testing.T) {
		resp, err := client.GetTask("known-task")

		AssertNoError(t, err)

		if resp == nil {
			t.Fatal("Expected response, got nil")
		}

		AssertTaskEqual(t, expectedTask, resp.Task)
	})

	t.Run("get non-existent task", func(t *testing.T) {
		_, err := client.GetTask("non-existent")

		AssertError(t, err)
		AssertErrorContains(t, err, "404")
	})
}

func TestClient_UpdateTask(t *testing.T) {
	server := SetupMockServerWithData()
	defer server.Close()

	client := NewClient(server.URL, "test-key")

	// Add a task to update
	originalTask := NewTaskBuilder().
		WithID("update-task").
		WithTitle("Original Title").
		WithStatus("todo").
		WithFeature("original").
		Build()
	server.AddTask(originalTask)

	tests := []struct {
		name           string
		taskID         string
		updates        UpdateTaskRequest
		expectError    bool
		expectedStatus string
		expectedTitle  string
	}{
		{
			name:   "update status",
			taskID: "update-task",
			updates: UpdateTaskRequest{
				Status: stringPtr("doing"),
			},
			expectError:    false,
			expectedStatus: "doing",
			expectedTitle:  "Original Title", // Should remain unchanged
		},
		{
			name:   "update feature",
			taskID: "update-task",
			updates: UpdateTaskRequest{
				Feature: stringPtr("updated"),
			},
			expectError:    false,
			expectedStatus: "doing", // From previous test
			expectedTitle:  "Original Title",
		},
		{
			name:   "update title",
			taskID: "update-task",
			updates: UpdateTaskRequest{
				Title: stringPtr("Updated Title"),
			},
			expectError:    false,
			expectedStatus: "doing",
			expectedTitle:  "Updated Title",
		},
		{
			name:   "update non-existent task",
			taskID: "non-existent",
			updates: UpdateTaskRequest{
				Status: stringPtr("doing"),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.UpdateTask(tt.taskID, tt.updates)

			if tt.expectError {
				AssertError(t, err)
				return
			}

			AssertNoError(t, err)

			if resp == nil {
				t.Fatal("Expected response, got nil")
			}

			if resp.Task.Status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, resp.Task.Status)
			}

			if resp.Task.Title != tt.expectedTitle {
				t.Errorf("Expected title %s, got %s", tt.expectedTitle, resp.Task.Title)
			}

			// Verify the update was applied on the server
			getResp, err := client.GetTask(tt.taskID)
			AssertNoError(t, err)

			if getResp.Task.Status != tt.expectedStatus {
				t.Errorf("Server state: expected status %s, got %s", tt.expectedStatus, getResp.Task.Status)
			}
		})
	}
}

func TestClient_ListProjects(t *testing.T) {
	server := SetupMockServerWithData()
	defer server.Close()

	client := NewClient(server.URL, "test-key")

	resp, err := client.ListProjects()

	AssertNoError(t, err)

	if resp == nil {
		t.Fatal("Expected response, got nil")
	}

	expectedProjects := SampleProjects()
	if resp.Count != len(expectedProjects) {
		t.Errorf("Expected count %d, got %d", len(expectedProjects), resp.Count)
	}

	if len(resp.Projects) != len(expectedProjects) {
		t.Errorf("Expected %d projects, got %d", len(expectedProjects), len(resp.Projects))
	}

	// Verify we got actual project data
	if len(resp.Projects) > 0 {
		project := resp.Projects[0]
		if project.ID == "" {
			t.Error("Expected project to have non-empty ID")
		}
		if project.Title == "" {
			t.Error("Expected project to have non-empty title")
		}
	}
}

func TestClient_ListProjects_Error(t *testing.T) {
	server := NewMockServer()
	defer server.Close()

	// Configure server to return error
	server.SetSimulatedError("projects", SimulateAPIError(500, "Internal server error"))

	client := NewClient(server.URL, "test-key")

	_, err := client.ListProjects()
	AssertError(t, err)
	AssertErrorContains(t, err, "500")
}

func TestClient_GetProject(t *testing.T) {
	server := SetupMockServerWithData()
	defer server.Close()

	client := NewClient(server.URL, "test-key")

	// Add a known project
	expectedProject := NewProjectBuilder().
		WithID("known-project").
		WithTitle("Known Project").
		Build()
	server.AddProject(expectedProject)

	t.Run("get existing project", func(t *testing.T) {
		resp, err := client.GetProject("known-project")

		AssertNoError(t, err)

		if resp == nil {
			t.Fatal("Expected response, got nil")
		}

		AssertProjectEqual(t, expectedProject, resp.Project)
	})

	t.Run("get non-existent project", func(t *testing.T) {
		_, err := client.GetProject("non-existent")

		AssertError(t, err)
		AssertErrorContains(t, err, "404")
	})
}

func TestClient_HealthCheck(t *testing.T) {
	server := NewMockServer()
	defer server.Close()

	client := NewClient(server.URL, "test-key")

	t.Run("healthy server", func(t *testing.T) {
		server.SetHealthStatus(200)

		err := client.HealthCheck()
		AssertNoError(t, err)
	})

	t.Run("unhealthy server", func(t *testing.T) {
		server.SetHealthStatus(500)

		err := client.HealthCheck()
		AssertError(t, err)
		AssertErrorContains(t, err, "500")
	})

	t.Run("server error", func(t *testing.T) {
		server.SetSimulatedError("health", SimulateTimeoutError())

		err := client.HealthCheck()
		AssertError(t, err)
		// The mock server returns a 500 status when error is simulated
		AssertErrorContains(t, err, "500")
	})
}

func TestClient_RequestAuthentication(t *testing.T) {
	server := NewMockServer()
	defer server.Close()

	t.Run("with API key", func(t *testing.T) {
		client := NewClient(server.URL, "test-api-key")

		// Make a request and verify the Authorization header
		_, _ = client.ListTasks(nil, nil, true)

		requests := server.GetRequestsForPath("/api/tasks")
		if len(requests) == 0 {
			t.Fatal("Expected at least one request")
		}

		req := requests[0]
		authHeader := req.Headers["Authorization"]
		expectedAuth := "Bearer test-api-key"

		if authHeader != expectedAuth {
			t.Errorf("Expected Authorization header %s, got %s", expectedAuth, authHeader)
		}
	})

	t.Run("without API key", func(t *testing.T) {
		client := NewClient(server.URL, "")

		// Reset server request history
		server.Reset()

		_, _ = client.ListTasks(nil, nil, true)

		requests := server.GetRequestsForPath("/api/tasks")
		if len(requests) == 0 {
			t.Fatal("Expected at least one request")
		}

		req := requests[0]
		authHeader := req.Headers["Authorization"]

		if authHeader != "" {
			t.Errorf("Expected no Authorization header, got %s", authHeader)
		}
	})
}

func TestClient_ContentType(t *testing.T) {
	server := NewMockServer()
	defer server.Close()

	client := NewClient(server.URL, "test-key")

	// Make a request with body
	updateReq := UpdateTaskRequest{Status: stringPtr("doing")}
	_, _ = client.UpdateTask("test-task", updateReq)

	requests := server.GetRequestsForPath("/api/tasks/test-task")
	if len(requests) == 0 {
		t.Fatal("Expected at least one request")
	}

	req := requests[0]
	contentType := req.Headers["Content-Type"]
	expectedContentType := "application/json"

	if contentType != expectedContentType {
		t.Errorf("Expected Content-Type %s, got %s", expectedContentType, contentType)
	}
}

// Benchmark tests

func BenchmarkClient_ListTasks(b *testing.B) {
	server := SetupMockServerWithData()
	defer server.Close()

	client := NewClient(server.URL, "test-key")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.ListTasks(nil, nil, true)
		if err != nil {
			b.Fatalf("ListTasks failed: %v", err)
		}
	}
}

func BenchmarkClient_GetTask(b *testing.B) {
	server := SetupMockServerWithData()
	defer server.Close()

	client := NewClient(server.URL, "test-key")

	// Add a task to get
	task := DefaultTask()
	server.AddTask(task)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GetTask(task.ID)
		if err != nil {
			b.Fatalf("GetTask failed: %v", err)
		}
	}
}

func BenchmarkClient_UpdateTask(b *testing.B) {
	server := SetupMockServerWithData()
	defer server.Close()

	client := NewClient(server.URL, "test-key")

	// Add a task to update
	task := DefaultTask()
	server.AddTask(task)

	updateReq := UpdateTaskRequest{Status: stringPtr("doing")}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.UpdateTask(task.ID, updateReq)
		if err != nil {
			b.Fatalf("UpdateTask failed: %v", err)
		}
	}
}

// Helper function is defined in test_fixtures.go
