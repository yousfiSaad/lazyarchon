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

// Logger interface for optional logging in Client
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	LogHTTPRequest(method, url string, args ...interface{})
	LogHTTPResponse(method, url string, statusCode int, duration time.Duration, args ...interface{})
	LogWebSocketEvent(event string, args ...interface{})
	LogStateChange(component, field string, oldValue, newValue interface{}, args ...interface{})
	LogPerformance(operation string, startTime time.Time, args ...interface{})
}

// Client represents an Archon API client
type Client struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
	logger     Logger // Optional logger for debug mode
}

// NewClient creates a new Archon API client
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey: apiKey,
		logger: nil, // No logger by default
	}
}

// SetLogger sets the optional logger for the client
func (c *Client) SetLogger(logger Logger) {
	c.logger = logger
}

// makeRequest makes an HTTP request to the Archon API
func (c *Client) makeRequest(method, path string, body interface{}) (*http.Response, error) {
	startTime := time.Now()
	fullURL := c.baseURL + path

	var reqBody io.Reader
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			if c.logger != nil {
				c.logger.Error("Failed to marshal request body", "error", err, "method", method, "url", fullURL)
			}
			return nil, fmt.Errorf("error marshaling request body: %w", err)
		}
		reqBody = bytes.NewBuffer(bodyBytes)
	}

	// Log the outgoing request
	if c.logger != nil {
		logArgs := []interface{}{}
		if len(bodyBytes) > 0 && len(bodyBytes) < 1000 { // Only log body if reasonable size
			logArgs = append(logArgs, "body", string(bodyBytes))
		} else if len(bodyBytes) >= 1000 {
			logArgs = append(logArgs, "body_size", len(bodyBytes))
		}
		c.logger.LogHTTPRequest(method, fullURL, logArgs...)
	}

	req, err := http.NewRequest(method, fullURL, reqBody)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("Failed to create HTTP request", "error", err, "method", method, "url", fullURL)
		}
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		if c.logger != nil {
			c.logger.Error("HTTP request failed", "error", err, "method", method, "url", fullURL, "duration_ms", duration.Milliseconds())
		}
		return nil, fmt.Errorf("error making request: %w", err)
	}

	// Log the response
	if c.logger != nil {
		c.logger.LogHTTPResponse(method, fullURL, resp.StatusCode, duration)
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

// DeleteTask deletes/archives a task
func (c *Client) DeleteTask(taskID string) error {
	path := "/api/tasks/" + taskID

	resp, err := c.makeRequest("DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrTaskNotFound
	}

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete task: status %d", resp.StatusCode)
	}

	return nil
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
