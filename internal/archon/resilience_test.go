package archon

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetryConfig_Defaults(t *testing.T) {
	config := DefaultResilienceConfig()

	if !config.Enabled {
		t.Error("Expected resilience to be enabled by default")
	}

	if config.Retry.MaxAttempts != 3 {
		t.Errorf("Expected 3 max attempts, got %d", config.Retry.MaxAttempts)
	}

	if config.Retry.BaseDelay != 100*time.Millisecond {
		t.Errorf("Expected 100ms base delay, got %v", config.Retry.BaseDelay)
	}

	if config.CircuitBreaker.FailureThreshold != 5 {
		t.Errorf("Expected 5 failure threshold, got %d", config.CircuitBreaker.FailureThreshold)
	}
}

func TestCircuitBreaker_States(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold: 2,
		SuccessThreshold: 2,
		Timeout:          100 * time.Millisecond,
	}

	cb := NewCircuitBreaker(config)

	// Initial state should be closed
	if cb.GetState() != CircuitClosed {
		t.Errorf("Expected initial state to be closed, got %s", cb.GetState())
	}

	// First failure - should remain closed
	err := cb.Execute(func() error {
		return errors.New("test failure")
	})
	if err == nil {
		t.Error("Expected error from failing function")
	}
	if cb.GetState() != CircuitClosed {
		t.Errorf("Expected state to remain closed after first failure, got %s", cb.GetState())
	}

	// Second failure - should open circuit
	err = cb.Execute(func() error {
		return errors.New("test failure")
	})
	if err == nil {
		t.Error("Expected error from failing function")
	}
	if cb.GetState() != CircuitOpen {
		t.Errorf("Expected state to be open after second failure, got %s", cb.GetState())
	}

	// Immediate retry should be blocked
	err = cb.Execute(func() error {
		return nil
	})
	if err == nil {
		t.Error("Expected circuit breaker to block request")
	}
	if _, ok := err.(*CircuitBreakerError); !ok {
		t.Errorf("Expected CircuitBreakerError, got %T", err)
	}

	// Wait for timeout and try again - should be half-open
	time.Sleep(150 * time.Millisecond)

	// First success in half-open - should remain half-open
	err = cb.Execute(func() error {
		return nil
	})
	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if cb.GetState() != CircuitHalfOpen {
		t.Errorf("Expected state to be half-open after first success, got %s", cb.GetState())
	}

	// Second success - should close circuit
	err = cb.Execute(func() error {
		return nil
	})
	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if cb.GetState() != CircuitClosed {
		t.Errorf("Expected state to be closed after second success, got %s", cb.GetState())
	}
}

func TestCircuitBreaker_HalfOpenFailure(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold: 1,
		SuccessThreshold: 2,
		Timeout:          50 * time.Millisecond,
	}

	cb := NewCircuitBreaker(config)

	// Force circuit open
	cb.Execute(func() error {
		return errors.New("failure")
	})

	if cb.GetState() != CircuitOpen {
		t.Errorf("Expected circuit to be open, got %s", cb.GetState())
	}

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Fail in half-open state - should go back to open
	err := cb.Execute(func() error {
		return errors.New("failure in half-open")
	})

	if err == nil {
		t.Error("Expected error from failing function")
	}

	if cb.GetState() != CircuitOpen {
		t.Errorf("Expected circuit to go back to open after half-open failure, got %s", cb.GetState())
	}
}

func TestRetryableExecutor_Success(t *testing.T) {
	config := RetryConfig{
		MaxAttempts:     3,
		BaseDelay:       10 * time.Millisecond,
		MaxDelay:        100 * time.Millisecond,
		Multiplier:      2.0,
		Jitter:          false,
		RetryableErrors: []string{"retryable"},
	}

	executor := NewRetryableExecutor(config)
	ctx := context.Background()

	// Test successful execution on first try
	attempts := 0
	err := executor.Execute(ctx, func() error {
		attempts++
		return nil
	})

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}

	if attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", attempts)
	}
}

func TestRetryableExecutor_RetrySuccess(t *testing.T) {
	config := RetryConfig{
		MaxAttempts:     3,
		BaseDelay:       10 * time.Millisecond,
		MaxDelay:        100 * time.Millisecond,
		Multiplier:      2.0,
		Jitter:          false,
		RetryableErrors: []string{"retryable"},
	}

	executor := NewRetryableExecutor(config)
	ctx := context.Background()

	// Test retry on retryable error, success on third attempt
	attempts := 0
	err := executor.Execute(ctx, func() error {
		attempts++
		if attempts < 3 {
			return errors.New("retryable error")
		}
		return nil
	})

	if err != nil {
		t.Errorf("Expected success after retries, got error: %v", err)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestRetryableExecutor_NonRetryableError(t *testing.T) {
	config := RetryConfig{
		MaxAttempts:     3,
		BaseDelay:       10 * time.Millisecond,
		MaxDelay:        100 * time.Millisecond,
		Multiplier:      2.0,
		Jitter:          false,
		RetryableErrors: []string{"retryable"},
	}

	executor := NewRetryableExecutor(config)
	ctx := context.Background()

	// Test non-retryable error (use error message that doesn't contain "retryable")
	attempts := 0
	err := executor.Execute(ctx, func() error {
		attempts++
		return errors.New("bad request error")
	})

	if err == nil {
		t.Error("Expected error to be returned")
	}

	if attempts != 1 {
		t.Errorf("Expected 1 attempt for non-retryable error, got %d", attempts)
	}

	if err.Error() != "bad request error" {
		t.Errorf("Expected original error, got: %v", err)
	}
}

func TestRetryableExecutor_ExhaustedRetries(t *testing.T) {
	config := RetryConfig{
		MaxAttempts:     2,
		BaseDelay:       10 * time.Millisecond,
		MaxDelay:        100 * time.Millisecond,
		Multiplier:      2.0,
		Jitter:          false,
		RetryableErrors: []string{"retryable"},
	}

	executor := NewRetryableExecutor(config)
	ctx := context.Background()

	// Test exhausted retries
	attempts := 0
	err := executor.Execute(ctx, func() error {
		attempts++
		return errors.New("retryable error")
	})

	if err == nil {
		t.Error("Expected error after exhausted retries")
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}

	if _, ok := err.(*RetryExhaustedError); !ok {
		t.Errorf("Expected RetryExhaustedError, got %T", err)
	}
}

func TestRetryableExecutor_ContextCancellation(t *testing.T) {
	config := RetryConfig{
		MaxAttempts:     5,
		BaseDelay:       100 * time.Millisecond,
		MaxDelay:        1 * time.Second,
		Multiplier:      2.0,
		Jitter:          false,
		RetryableErrors: []string{"retryable"},
	}

	executor := NewRetryableExecutor(config)

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	attempts := 0
	err := executor.Execute(ctx, func() error {
		attempts++
		return errors.New("retryable error")
	})

	if err == nil {
		t.Error("Expected context cancellation error")
	}

	if err != context.DeadlineExceeded {
		t.Errorf("Expected context deadline exceeded, got: %v", err)
	}

	// Should have attempted at least once but not completed all retries
	if attempts == 0 {
		t.Error("Expected at least one attempt")
	}
	if attempts >= 5 {
		t.Errorf("Expected fewer than 5 attempts due to context cancellation, got %d", attempts)
	}
}

func TestRetryableExecutor_DelayCalculation(t *testing.T) {
	config := RetryConfig{
		MaxAttempts:     4,
		BaseDelay:       100 * time.Millisecond,
		MaxDelay:        500 * time.Millisecond,
		Multiplier:      2.0,
		Jitter:          false,
		RetryableErrors: []string{"retryable"},
	}

	executor := NewRetryableExecutor(config)

	tests := []struct {
		attempt       int
		expectedDelay time.Duration
	}{
		{0, 100 * time.Millisecond}, // 100 * 2^0 = 100ms
		{1, 200 * time.Millisecond}, // 100 * 2^1 = 200ms
		{2, 400 * time.Millisecond}, // 100 * 2^2 = 400ms
		{3, 500 * time.Millisecond}, // 100 * 2^3 = 800ms, capped at 500ms
	}

	for _, test := range tests {
		delay := executor.calculateDelay(test.attempt)
		if delay != test.expectedDelay {
			t.Errorf("For attempt %d, expected delay %v, got %v", test.attempt, test.expectedDelay, delay)
		}
	}
}

func TestRetryableExecutor_NetworkErrorRetryable(t *testing.T) {
	config := RetryConfig{
		MaxAttempts:     3,
		BaseDelay:       10 * time.Millisecond,
		MaxDelay:        100 * time.Millisecond,
		Multiplier:      2.0,
		Jitter:          false,
		RetryableErrors: []string{},
	}

	executor := NewRetryableExecutor(config)
	ctx := context.Background()

	// Test that NetworkError is retryable even without being in RetryableErrors list
	attempts := 0
	err := executor.Execute(ctx, func() error {
		attempts++
		if attempts < 3 {
			return SimulateNetworkError()
		}
		return nil
	})

	if err != nil {
		t.Errorf("Expected success after retries, got error: %v", err)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestResilientClient_Integration(t *testing.T) {
	// Set up mock server
	server := SetupMockServerWithData()
	defer server.Close()

	// Configure for fast testing
	config := ResilienceConfig{
		Enabled: true,
		Retry: RetryConfig{
			MaxAttempts:     2,
			BaseDelay:       10 * time.Millisecond,
			MaxDelay:        50 * time.Millisecond,
			Multiplier:      2.0,
			Jitter:          false,
			RetryableErrors: []string{"connection", "timeout"},
		},
		CircuitBreaker: CircuitBreakerConfig{
			FailureThreshold: 2,
			SuccessThreshold: 1,
			Timeout:          100 * time.Millisecond,
		},
	}

	client := NewResilientClient(server.URL, "test-key", config)

	// Test successful operation
	tasks, err := client.ListTasks(nil, nil, true)
	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if tasks == nil {
		t.Error("Expected tasks response")
	}

	// Verify metrics
	metrics := client.GetMetrics()
	if metrics.TotalRequests != 1 {
		t.Errorf("Expected 1 total request, got %d", metrics.TotalRequests)
	}
	if metrics.SuccessfulRequests != 1 {
		t.Errorf("Expected 1 successful request, got %d", metrics.SuccessfulRequests)
	}

	// Verify circuit breaker is healthy
	if !client.IsHealthy() {
		t.Error("Expected client to be healthy")
	}
}

func TestResilientClient_CircuitBreakerTrip(t *testing.T) {
	server := NewMockServer()
	defer server.Close()

	// Configure server to always return errors
	server.SetSimulatedError("tasks", errors.New("server error"))

	config := ResilienceConfig{
		Enabled: true,
		Retry: RetryConfig{
			MaxAttempts:     1, // No retries for faster testing
			BaseDelay:       10 * time.Millisecond,
			MaxDelay:        50 * time.Millisecond,
			Multiplier:      2.0,
			Jitter:          false,
			RetryableErrors: []string{},
		},
		CircuitBreaker: CircuitBreakerConfig{
			FailureThreshold: 2,
			SuccessThreshold: 1,
			Timeout:          100 * time.Millisecond,
		},
	}

	client := NewResilientClient(server.URL, "test-key", config)

	// First failure
	_, err := client.ListTasks(nil, nil, true)
	if err == nil {
		t.Error("Expected error from first request")
	}

	// Second failure - should trip circuit breaker
	_, err = client.ListTasks(nil, nil, true)
	if err == nil {
		t.Error("Expected error from second request")
	}

	// Third request should be blocked by circuit breaker
	_, err = client.ListTasks(nil, nil, true)
	if err == nil {
		t.Error("Expected circuit breaker to block request")
	}

	// Check if it's a circuit breaker error
	if _, ok := err.(*CircuitBreakerError); !ok {
		t.Errorf("Expected CircuitBreakerError, got %T: %v", err, err)
	}

	// Verify circuit breaker state
	if client.GetCircuitBreakerState() != CircuitOpen {
		t.Errorf("Expected circuit breaker to be open, got %s", client.GetCircuitBreakerState())
	}

	// Verify client is not healthy
	if client.IsHealthy() {
		t.Error("Expected client to be unhealthy when circuit breaker is open")
	}
}

func TestClientFactory(t *testing.T) {
	factory := NewClientFactory("http://localhost:8181", "test-key")

	// Test creating resilient client (it returns ClientInterface)
	client := factory.CreateClient()
	if _, ok := client.(*ResilientClient); !ok {
		t.Errorf("Expected ResilientClient, got %T", client)
	}

	// Test creating basic client
	basicClient := factory.WithoutResilience().CreateClient()
	if _, ok := basicClient.(*Client); !ok {
		t.Errorf("Expected basic Client, got %T", basicClient)
	}

	// Test creating resilient client even when factory is set to basic
	resilientClient := factory.CreateResilientClient()
	if resilientClient == nil {
		t.Error("Expected ResilientClient, got nil")
	}
}

func TestResilienceMetrics(t *testing.T) {
	metrics := NewResilienceMetrics()

	// Record some operations
	metrics.RecordRequest()
	metrics.RecordRequest()
	metrics.RecordSuccess()
	metrics.RecordFailure()
	metrics.RecordRetry()
	metrics.RecordCircuitBreakerTrip()

	snapshot := metrics.GetSnapshot()

	if snapshot.TotalRequests != 2 {
		t.Errorf("Expected 2 total requests, got %d", snapshot.TotalRequests)
	}
	if snapshot.SuccessfulRequests != 1 {
		t.Errorf("Expected 1 successful request, got %d", snapshot.SuccessfulRequests)
	}
	if snapshot.FailedRequests != 1 {
		t.Errorf("Expected 1 failed request, got %d", snapshot.FailedRequests)
	}
	if snapshot.RetriedRequests != 1 {
		t.Errorf("Expected 1 retried request, got %d", snapshot.RetriedRequests)
	}
	if snapshot.CircuitBreakerTrips != 1 {
		t.Errorf("Expected 1 circuit breaker trip, got %d", snapshot.CircuitBreakerTrips)
	}
}

// Benchmark tests for resilience overhead

func BenchmarkResilientClient_Success(b *testing.B) {
	server := SetupMockServerWithData()
	defer server.Close()

	client := NewResilientClientWithDefaults(server.URL, "test-key")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.ListTasks(nil, nil, true)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkBasicClient_Success(b *testing.B) {
	server := SetupMockServerWithData()
	defer server.Close()

	client := NewClient(server.URL, "test-key")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.ListTasks(nil, nil, true)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}
