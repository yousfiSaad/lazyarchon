package archon

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Common errors
var (
	ErrTaskNotFound    = errors.New("task not found")
	ErrProjectNotFound = errors.New("project not found")
)

// Client represents an Archon API client
type Client struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
}

// NewClient creates a new Archon API client
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey: apiKey,
	}
}

// makeRequest makes an HTTP request to the Archon API
func (c *Client) makeRequest(method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}

	return resp, nil
}

// parseResponse parses the HTTP response into the given structure
func (c *Client) parseResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("error unmarshaling response: %w", err)
	}

	return nil
}

// ListTasks retrieves all tasks from the API
func (c *Client) ListTasks(projectID *string, status *string, includeClosed bool) (*TasksResponse, error) {
	path := "/api/tasks"

	// Add query parameters for filtering
	params := url.Values{}
	if projectID != nil {
		params.Add("project_id", *projectID)
	}
	if status != nil {
		params.Add("status", *status)
	}
	if includeClosed {
		params.Add("include_closed", "true")
	}
	params.Add("per_page", "100")

	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	// Parse the API response which contains tasks in a "tasks" field
	var tasksResp TasksResponse
	if err := c.parseResponse(resp, &tasksResp); err != nil {
		return nil, err
	}

	return &tasksResp, nil
}

// GetTask retrieves a specific task by ID
func (c *Client) GetTask(taskID string) (*TaskResponse, error) {
	path := "/api/tasks/" + taskID

	resp, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var taskResp TaskResponse
	if err := c.parseResponse(resp, &taskResp); err != nil {
		return nil, err
	}

	return &taskResp, nil
}

// UpdateTask updates an existing task
func (c *Client) UpdateTask(taskID string, updates UpdateTaskRequest) (*TaskResponse, error) {
	path := "/api/tasks/" + taskID

	resp, err := c.makeRequest("PUT", path, updates)
	if err != nil {
		return nil, err
	}

	var taskResp TaskResponse
	if err := c.parseResponse(resp, &taskResp); err != nil {
		return nil, err
	}

	return &taskResp, nil
}

// ListProjects retrieves all projects from the API
func (c *Client) ListProjects() (*ProjectsResponse, error) {
	path := "/api/projects"

	resp, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var projectsResp ProjectsResponse
	if err := c.parseResponse(resp, &projectsResp); err != nil {
		return nil, err
	}

	return &projectsResp, nil
}

// GetProject retrieves a specific project by ID
func (c *Client) GetProject(projectID string) (*ProjectResponse, error) {
	path := "/api/projects/" + projectID

	resp, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var projectResp ProjectResponse
	if err := c.parseResponse(resp, &projectResp); err != nil {
		return nil, err
	}

	return &projectResp, nil
}

// HealthCheck checks if the API is accessible
func (c *Client) HealthCheck() error {
	resp, err := c.makeRequest("GET", "/health", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	return nil
}
