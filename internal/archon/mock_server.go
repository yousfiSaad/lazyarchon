package archon

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
)

// MockServer provides a test HTTP server that simulates the Archon API
type MockServer struct {
	*httptest.Server
	mu sync.RWMutex

	// Data stores
	tasks    map[string]Task
	projects map[string]Project

	// Request recording
	requests []RecordedRequest

	// Behavior configuration
	simulateErrors map[string]error // endpoint -> error mapping
	responseDelays map[string]int   // endpoint -> delay in milliseconds
	healthStatus   int              // HTTP status for health endpoint
	nextTaskID     int
	nextProjectID  int
}

// RecordedRequest captures details about requests made to the mock server
type RecordedRequest struct {
	Method   string
	Path     string
	Headers  map[string]string
	Body     string
	FormData map[string]string
}

// NewMockServer creates and starts a new mock Archon API server
func NewMockServer() *MockServer {
	server := &MockServer{
		tasks:          make(map[string]Task),
		projects:       make(map[string]Project),
		requests:       make([]RecordedRequest, 0),
		simulateErrors: make(map[string]error),
		responseDelays: make(map[string]int),
		healthStatus:   http.StatusOK,
		nextTaskID:     1,
		nextProjectID:  1,
	}

	// Create the HTTP test server
	mux := http.NewServeMux()
	mux.HandleFunc("/health", server.handleHealth)
	mux.HandleFunc("/api/tasks", server.handleTasks)
	mux.HandleFunc("/api/tasks/", server.handleTaskByID)
	mux.HandleFunc("/api/projects", server.handleProjects)
	mux.HandleFunc("/api/projects/", server.handleProjectByID)

	server.Server = httptest.NewServer(mux)
	return server
}

// writeJSONResponse writes a JSON response with error handling
func (s *MockServer) writeJSONResponse(w http.ResponseWriter, data interface{}) {
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// recordRequest captures request details for test verification
func (s *MockServer) recordRequest(r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	headers := make(map[string]string)
	for key, values := range r.Header {
		headers[key] = strings.Join(values, ", ")
	}

	// Read body if present
	body := ""
	if r.Body != nil {
		if bodyBytes, err := json.Marshal(r.Body); err == nil {
			body = string(bodyBytes)
		}
	}

	s.requests = append(s.requests, RecordedRequest{
		Method:  r.Method,
		Path:    r.URL.Path,
		Headers: headers,
		Body:    body,
	})
}

// Health endpoint handler
func (s *MockServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.recordRequest(r)

	if err := s.checkSimulatedError("health"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(s.healthStatus)
	s.writeJSONResponse(w, map[string]string{"status": "healthy"})
}

// Tasks endpoint handler
func (s *MockServer) handleTasks(w http.ResponseWriter, r *http.Request) {
	s.recordRequest(r)

	if err := s.checkSimulatedError("tasks"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleListTasks(w, r)
	case http.MethodPost:
		s.handleCreateTask(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Task by ID endpoint handler
func (s *MockServer) handleTaskByID(w http.ResponseWriter, r *http.Request) {
	s.recordRequest(r)

	// Extract task ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	taskID := strings.Split(path, "/")[0]

	if err := s.checkSimulatedError("task_by_id"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetTask(w, r, taskID)
	case http.MethodPut:
		s.handleUpdateTask(w, r, taskID)
	case http.MethodDelete:
		s.handleDeleteTask(w, r, taskID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Projects endpoint handler
func (s *MockServer) handleProjects(w http.ResponseWriter, r *http.Request) {
	s.recordRequest(r)

	if err := s.checkSimulatedError("projects"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleListProjects(w, r)
	case http.MethodPost:
		s.handleCreateProject(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Project by ID endpoint handler
func (s *MockServer) handleProjectByID(w http.ResponseWriter, r *http.Request) {
	s.recordRequest(r)

	// Extract project ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/projects/")
	projectID := strings.Split(path, "/")[0]

	if err := s.checkSimulatedError("project_by_id"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetProject(w, r, projectID)
	case http.MethodPut:
		s.handleUpdateProject(w, r, projectID)
	case http.MethodDelete:
		s.handleDeleteProject(w, r, projectID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// List tasks implementation
func (s *MockServer) handleListTasks(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	tasks := make([]Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		// Apply filters based on query parameters
		projectID := r.URL.Query().Get("project_id")
		status := r.URL.Query().Get("status")
		includeClosed := r.URL.Query().Get("include_closed") == "true"

		// Filter by project ID
		if projectID != "" && task.ProjectID != projectID {
			continue
		}

		// Filter by status
		if status != "" && task.Status != status {
			continue
		}

		// Filter out closed tasks if not included
		if !includeClosed && task.Status == "done" {
			continue
		}

		tasks = append(tasks, task)
	}
	s.mu.RUnlock()

	response := TasksResponse{
		Tasks: tasks,
		Count: len(tasks),
	}

	w.Header().Set("Content-Type", "application/json")
	s.writeJSONResponse(w, response)
}

// Get task implementation
func (s *MockServer) handleGetTask(w http.ResponseWriter, r *http.Request, taskID string) {
	s.mu.RLock()
	task, exists := s.tasks[taskID]
	s.mu.RUnlock()

	if !exists {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	response := TaskResponse{Task: task}
	w.Header().Set("Content-Type", "application/json")
	s.writeJSONResponse(w, response)
}

// Update task implementation
func (s *MockServer) handleUpdateTask(w http.ResponseWriter, r *http.Request, taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	var updateReq UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Apply updates
	if updateReq.Title != nil {
		task.Title = *updateReq.Title
	}
	if updateReq.Status != nil {
		task.Status = *updateReq.Status
	}
	if updateReq.Feature != nil {
		task.Feature = updateReq.Feature
	}

	s.tasks[taskID] = task

	response := TaskResponse{Task: task}
	w.Header().Set("Content-Type", "application/json")
	s.writeJSONResponse(w, response)
}

// Create task implementation
func (s *MockServer) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Assign ID
	task.ID = fmt.Sprintf("task-%d", s.nextTaskID)
	s.nextTaskID++

	s.tasks[task.ID] = task

	response := TaskResponse{Task: task}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	s.writeJSONResponse(w, response)
}

// Delete task implementation
func (s *MockServer) handleDeleteTask(w http.ResponseWriter, r *http.Request, taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[taskID]; !exists {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	delete(s.tasks, taskID)
	w.WriteHeader(http.StatusNoContent)
}

// List projects implementation
func (s *MockServer) handleListProjects(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	projects := make([]Project, 0, len(s.projects))
	for _, project := range s.projects {
		projects = append(projects, project)
	}
	s.mu.RUnlock()

	response := ProjectsResponse{
		Projects: projects,
		Count:    len(projects),
	}

	w.Header().Set("Content-Type", "application/json")
	s.writeJSONResponse(w, response)
}

// Get project implementation
func (s *MockServer) handleGetProject(w http.ResponseWriter, r *http.Request, projectID string) {
	s.mu.RLock()
	project, exists := s.projects[projectID]
	s.mu.RUnlock()

	if !exists {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	response := ProjectResponse{Project: project}
	w.Header().Set("Content-Type", "application/json")
	s.writeJSONResponse(w, response)
}

// Create project implementation
func (s *MockServer) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var project Project
	if err := json.NewDecoder(r.Body).Decode(&project); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Assign ID
	project.ID = fmt.Sprintf("project-%d", s.nextProjectID)
	s.nextProjectID++

	s.projects[project.ID] = project

	response := ProjectResponse{Project: project}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	s.writeJSONResponse(w, response)
}

// Update project implementation
func (s *MockServer) handleUpdateProject(w http.ResponseWriter, r *http.Request, projectID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	project, exists := s.projects[projectID]
	if !exists {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	var updateData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Apply updates (simplified for testing)
	if title, ok := updateData["title"].(string); ok {
		project.Title = title
	}

	s.projects[projectID] = project

	response := ProjectResponse{Project: project}
	w.Header().Set("Content-Type", "application/json")
	s.writeJSONResponse(w, response)
}

// Delete project implementation
func (s *MockServer) handleDeleteProject(w http.ResponseWriter, r *http.Request, projectID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.projects[projectID]; !exists {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	delete(s.projects, projectID)
	w.WriteHeader(http.StatusNoContent)
}

// Helper methods

// checkSimulatedError checks if an error should be simulated for the given endpoint
func (s *MockServer) checkSimulatedError(endpoint string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.simulateErrors[endpoint]
}

// SetSimulatedError configures the server to return an error for a specific endpoint
func (s *MockServer) SetSimulatedError(endpoint string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.simulateErrors[endpoint] = err
}

// ClearSimulatedErrors removes all configured error simulations
func (s *MockServer) ClearSimulatedErrors() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.simulateErrors = make(map[string]error)
}

// SetHealthStatus configures the HTTP status returned by the health endpoint
func (s *MockServer) SetHealthStatus(status int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.healthStatus = status
}

// AddTask adds a task to the mock server's data store
func (s *MockServer) AddTask(task Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if task.ID == "" {
		task.ID = fmt.Sprintf("task-%d", s.nextTaskID)
		s.nextTaskID++
	}
	s.tasks[task.ID] = task
}

// AddProject adds a project to the mock server's data store
func (s *MockServer) AddProject(project Project) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if project.ID == "" {
		project.ID = fmt.Sprintf("project-%d", s.nextProjectID)
		s.nextProjectID++
	}
	s.projects[project.ID] = project
}

// GetRequestCount returns the total number of requests made to the server
func (s *MockServer) GetRequestCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.requests)
}

// GetRequestsForPath returns all requests made to a specific path
func (s *MockServer) GetRequestsForPath(path string) []RecordedRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var matches []RecordedRequest
	for _, req := range s.requests {
		if req.Path == path {
			matches = append(matches, req)
		}
	}
	return matches
}

// Reset clears all data and recorded requests
func (s *MockServer) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tasks = make(map[string]Task)
	s.projects = make(map[string]Project)
	s.requests = make([]RecordedRequest, 0)
	s.simulateErrors = make(map[string]error)
	s.responseDelays = make(map[string]int)
	s.healthStatus = http.StatusOK
	s.nextTaskID = 1
	s.nextProjectID = 1
}
