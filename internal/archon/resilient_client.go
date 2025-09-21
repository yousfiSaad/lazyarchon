package archon

import (
	"context"
	"time"
)

// Ensure ResilientClient implements ClientInterface
var _ ClientInterface = (*ResilientClient)(nil)

// executeWithResilience wraps any operation with retry and circuit breaker logic
func (rc *ResilientClient) executeWithResilience(ctx context.Context, operation func() error) error {
	if !rc.config.Enabled {
		// If resilience is disabled, execute directly
		return operation()
	}

	rc.metrics.RecordRequest()

	// Wrap operation with circuit breaker
	circuitBreakerOperation := func() error {
		return rc.circuitBreaker.Execute(operation)
	}

	// Execute with retry logic
	err := rc.retryExecutor.Execute(ctx, circuitBreakerOperation)

	// Record metrics based on result
	if err != nil {
		rc.metrics.RecordFailure()

		// Check if it's a circuit breaker error
		if _, ok := err.(*CircuitBreakerError); ok {
			rc.metrics.RecordCircuitBreakerTrip()
		}

		// Check if it's a retry exhausted error
		if _, ok := err.(*RetryExhaustedError); ok {
			rc.metrics.RecordRetry()
		}
	} else {
		rc.metrics.RecordSuccess()
	}

	return err
}

// ListTasks retrieves all tasks with resilience
func (rc *ResilientClient) ListTasks(projectID *string, status *string, includeClosed bool) (*TasksResponse, error) {
	var result *TasksResponse
	var lastErr error

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	operation := func() error {
		var err error
		result, err = rc.client.ListTasks(projectID, status, includeClosed)
		lastErr = err
		return err
	}

	err := rc.executeWithResilience(ctx, operation)
	if err != nil {
		// Return the last error from the actual operation if available
		if lastErr != nil {
			return nil, lastErr
		}
		return nil, err
	}

	return result, nil
}

// GetTask retrieves a specific task with resilience
func (rc *ResilientClient) GetTask(taskID string) (*TaskResponse, error) {
	var result *TaskResponse
	var lastErr error

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	operation := func() error {
		var err error
		result, err = rc.client.GetTask(taskID)
		lastErr = err
		return err
	}

	err := rc.executeWithResilience(ctx, operation)
	if err != nil {
		if lastErr != nil {
			return nil, lastErr
		}
		return nil, err
	}

	return result, nil
}

// UpdateTask updates an existing task with resilience
func (rc *ResilientClient) UpdateTask(taskID string, updates UpdateTaskRequest) (*TaskResponse, error) {
	var result *TaskResponse
	var lastErr error

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	operation := func() error {
		var err error
		result, err = rc.client.UpdateTask(taskID, updates)
		lastErr = err
		return err
	}

	err := rc.executeWithResilience(ctx, operation)
	if err != nil {
		if lastErr != nil {
			return nil, lastErr
		}
		return nil, err
	}

	return result, nil
}

// ListProjects retrieves all projects with resilience
func (rc *ResilientClient) ListProjects() (*ProjectsResponse, error) {
	var result *ProjectsResponse
	var lastErr error

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	operation := func() error {
		var err error
		result, err = rc.client.ListProjects()
		lastErr = err
		return err
	}

	err := rc.executeWithResilience(ctx, operation)
	if err != nil {
		if lastErr != nil {
			return nil, lastErr
		}
		return nil, err
	}

	return result, nil
}

// GetProject retrieves a specific project with resilience
func (rc *ResilientClient) GetProject(projectID string) (*ProjectResponse, error) {
	var result *ProjectResponse
	var lastErr error

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	operation := func() error {
		var err error
		result, err = rc.client.GetProject(projectID)
		lastErr = err
		return err
	}

	err := rc.executeWithResilience(ctx, operation)
	if err != nil {
		if lastErr != nil {
			return nil, lastErr
		}
		return nil, err
	}

	return result, nil
}

// HealthCheck checks API accessibility with resilience
func (rc *ResilientClient) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	operation := func() error {
		return rc.client.HealthCheck()
	}

	return rc.executeWithResilience(ctx, operation)
}

// GetMetrics returns current resilience metrics
func (rc *ResilientClient) GetMetrics() ResilienceMetrics {
	return rc.metrics.GetSnapshot()
}

// GetCircuitBreakerMetrics returns circuit breaker specific metrics
func (rc *ResilientClient) GetCircuitBreakerMetrics() CircuitBreakerMetrics {
	return rc.circuitBreaker.GetMetrics()
}

// GetCircuitBreakerState returns current circuit breaker state
func (rc *ResilientClient) GetCircuitBreakerState() CircuitState {
	return rc.circuitBreaker.GetState()
}

// IsHealthy returns true if the client is healthy (circuit breaker is closed)
func (rc *ResilientClient) IsHealthy() bool {
	return rc.circuitBreaker.GetState() == CircuitClosed
}

// Reset resets the circuit breaker and metrics (useful for testing)
func (rc *ResilientClient) Reset() {
	rc.circuitBreaker = NewCircuitBreaker(rc.config.CircuitBreaker)
	rc.metrics = NewResilienceMetrics()
}

// Close gracefully shuts down the resilient client
func (rc *ResilientClient) Close() error {
	// Currently nothing to close, but useful for future enhancements
	// like connection pooling or background health checks
	return nil
}

// NewResilientClientWithDefaults creates a resilient client with default configuration
func NewResilientClientWithDefaults(baseURL, apiKey string) *ResilientClient {
	return NewResilientClient(baseURL, apiKey, DefaultResilienceConfig())
}

// ClientFactory provides a way to create clients with different configurations
type ClientFactory struct {
	baseURL           string
	apiKey            string
	resilienceConfig  ResilienceConfig
	enableResilience  bool
}

// NewClientFactory creates a new client factory
func NewClientFactory(baseURL, apiKey string) *ClientFactory {
	return &ClientFactory{
		baseURL:          baseURL,
		apiKey:           apiKey,
		resilienceConfig: DefaultResilienceConfig(),
		enableResilience: true,
	}
}

// WithResilienceConfig sets a custom resilience configuration
func (cf *ClientFactory) WithResilienceConfig(config ResilienceConfig) *ClientFactory {
	cf.resilienceConfig = config
	return cf
}

// WithoutResilience disables resilience features
func (cf *ClientFactory) WithoutResilience() *ClientFactory {
	cf.enableResilience = false
	return cf
}

// CreateClient creates a client based on the factory configuration
func (cf *ClientFactory) CreateClient() ClientInterface {
	if cf.enableResilience {
		return NewResilientClient(cf.baseURL, cf.apiKey, cf.resilienceConfig)
	}
	return NewClient(cf.baseURL, cf.apiKey)
}

// CreateResilientClient creates a resilient client (even if resilience is disabled in factory)
func (cf *ClientFactory) CreateResilientClient() *ResilientClient {
	return NewResilientClient(cf.baseURL, cf.apiKey, cf.resilienceConfig)
}

// CreateBasicClient creates a basic client without resilience
func (cf *ClientFactory) CreateBasicClient() *Client {
	return NewClient(cf.baseURL, cf.apiKey)
}