package di

import (
	"testing"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/internal/interfaces"
)

func TestNewContainer(t *testing.T) {
	container, err := NewContainer()
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}

	if container == nil {
		t.Fatal("Container should not be nil")
	}

	// Test that we can retrieve a UI model
	var uiModel tea.Model
	err = container.Invoke(func(model tea.Model) {
		uiModel = model
	})

	if err != nil {
		t.Fatalf("Failed to create UI model: %v", err)
	}

	if uiModel == nil {
		t.Fatal("UI model should not be nil")
	}
}

func TestNewTestContainer(t *testing.T) {
	container, err := NewTestContainer()
	if err != nil {
		t.Fatalf("Failed to create test container: %v", err)
	}

	if container == nil {
		t.Fatal("Test container should not be nil")
	}

	// Test that we can retrieve mock dependencies
	var configProvider interfaces.ConfigProvider
	err = container.Invoke(func(config interfaces.ConfigProvider) {
		configProvider = config
	})

	if err != nil {
		t.Fatalf("Failed to get config provider: %v", err)
	}

	// Verify it's the mock implementation
	if configProvider.GetServerURL() != "http://test-server:8080" {
		t.Errorf("Expected mock server URL, got: %s", configProvider.GetServerURL())
	}

	if configProvider.GetAPIKey() != "test-api-key" {
		t.Errorf("Expected mock API key, got: %s", configProvider.GetAPIKey())
	}
}

func TestMockArchonClient(t *testing.T) {
	client := NewMockArchonClient().(*MockArchonClient)

	// Test successful operation
	tasks, err := client.ListTasks(nil, nil, false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(tasks.Tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(tasks.Tasks))
	}

	// Test error injection
	testErr := &TestError{message: "test error"}
	client.SetError("ListTasks", testErr)

	_, err = client.ListTasks(nil, nil, false)
	if err != testErr {
		t.Errorf("Expected injected error, got: %v", err)
	}
}

func TestMockLogger(t *testing.T) {
	logger := NewMockLogger().(*MockLogger)

	// Test logging
	logger.Info("Test message", "arg1", "arg2")
	logger.Error("Error message")

	logs := logger.GetLogs()
	if len(logs) != 2 {
		t.Errorf("Expected 2 log entries, got %d", len(logs))
	}

	if logs[0].Level != "INFO" {
		t.Errorf("Expected INFO level, got %s", logs[0].Level)
	}

	if logs[0].Message != "Test message" {
		t.Errorf("Expected 'Test message', got %s", logs[0].Message)
	}

	if logs[1].Level != "ERROR" {
		t.Errorf("Expected ERROR level, got %s", logs[1].Level)
	}
}

func TestMockHealthChecker(t *testing.T) {
	checker := NewMockHealthChecker().(*MockHealthChecker)

	// Test default healthy state
	healthy, issues := checker.GetHealthStatus()
	if !healthy {
		t.Error("Expected healthy status by default")
	}

	if len(issues) != 0 {
		t.Errorf("Expected no issues by default, got %d", len(issues))
	}

	// Test setting unhealthy state
	testIssues := []string{"API connection failed", "Config invalid"}
	checker.SetHealthy(false, testIssues)

	healthy, issues = checker.GetHealthStatus()
	if healthy {
		t.Error("Expected unhealthy status after setting")
	}

	if len(issues) != 2 {
		t.Errorf("Expected 2 issues, got %d", len(issues))
	}
}

// TestError is a simple error type for testing
type TestError struct {
	message string
}

func (e *TestError) Error() string {
	return e.message
}

// Integration test showing how to use DI in tests
func TestDIIntegration(t *testing.T) {
	// Create test container
	container, err := NewTestContainer()
	if err != nil {
		t.Fatalf("Failed to create test container: %v", err)
	}

	// Test dependency injection with all mocks
	var (
		configProvider interfaces.ConfigProvider
		archonClient   interfaces.ArchonClient
		logger         interfaces.Logger
	)

	err = container.Invoke(func(
		config interfaces.ConfigProvider,
		client interfaces.ArchonClient,
		log interfaces.Logger,
	) {
		configProvider = config
		archonClient = client
		logger = log
	})

	if err != nil {
		t.Fatalf("Failed to invoke dependencies: %v", err)
	}

	// Verify all dependencies are injected and working
	if configProvider == nil {
		t.Fatal("Config provider should not be nil")
	}

	if archonClient == nil {
		t.Fatal("Archon client should not be nil")
	}

	if logger == nil {
		t.Fatal("Logger should not be nil")
	}

	// Test interactions
	logger.Info("Starting integration test")

	tasks, err := archonClient.ListTasks(nil, nil, false)
	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}

	if len(tasks.Tasks) == 0 {
		t.Error("Expected some mock tasks")
	}

	// Verify logger captured the message
	mockLogger := logger.(*MockLogger)
	logs := mockLogger.GetLogs()
	if len(logs) == 0 {
		t.Error("Expected at least one log entry")
	}
}