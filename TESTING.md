# LazyArchon Testing Guide

This document provides comprehensive information about testing in the LazyArchon project.

## 🏗️ Test Infrastructure

LazyArchon uses a comprehensive testing framework with the following components:

### Core Testing Components

- **Client Interface** (`internal/archon/client_interface.go`) - Enables dependency injection and mocking
- **Mock Client** (`internal/archon/mock_client.go`) - Full-featured mock with call recording
- **Mock Server** (`internal/archon/mock_server.go`) - HTTP test server simulating Archon API
- **Test Fixtures** (`internal/archon/test_fixtures.go`) - Fluent builders for test data
- **Test Utilities** (`internal/testutil/`) - Assertion helpers and utilities

## 🧪 Running Tests

### Basic Commands

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for specific package
go test ./internal/archon/...
go test ./internal/ui/...
go test ./internal/config/...

# Run specific test function
go test -run TestNewClient ./internal/archon/...

# Run tests matching a pattern
go test -run "TestClient_.*" ./internal/archon/...
```

### Test Coverage

```bash
# Run tests with coverage
go test -cover ./...

# Generate detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# View coverage in terminal
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Coverage for specific package
go test -cover ./internal/archon/...
```

### Performance Testing

```bash
# Run all benchmarks
go test -bench=. ./...

# Run benchmarks for API client
go test -bench=. ./internal/archon/...

# Run benchmarks with memory allocation stats
go test -bench=. -benchmem ./...

# Run specific benchmark
go test -bench=BenchmarkClient_ListTasks ./internal/archon/...

# Benchmark comparison (save baseline first)
go test -bench=. ./internal/archon/... > old.txt
# After changes:
go test -bench=. ./internal/archon/... > new.txt
# Compare with benchcmp tool
```

### Advanced Testing Options

```bash
# Run tests with race detection
go test -race ./...

# Run tests in parallel
go test -parallel 4 ./...

# Set timeout for tests
go test -timeout 30s ./...

# Short mode (skip long-running tests)
go test -short ./...

# Verbose output with test names
go test -v -run . ./...

# Generate test binary without running
go test -c ./internal/archon/...
```

## 📋 Test Categories

### 1. Unit Tests

**API Client Tests** (`internal/archon/client_test.go`)
- Client initialization
- HTTP request handling
- Authentication
- Error scenarios
- Response parsing

**UI Model Tests** (`internal/ui/model_test.go`)
- Model initialization
- State management
- Task sorting
- Navigation

**Configuration Tests** (`internal/config/config_test.go`)
- Config loading
- Environment overrides
- Default values

### 2. Integration Tests

**Mock Server Tests**
- Full HTTP request/response cycle
- API endpoint simulation
- Error condition testing

### 3. Performance Tests

**Benchmarks**
- API client operations
- Task sorting algorithms
- UI rendering performance

## 🛠️ Writing Tests

### Test Structure

```go
func TestFeatureName(t *testing.T) {
    // Arrange - Set up test data and dependencies
    client := NewClient("http://localhost", "test-key")

    // Act - Execute the code under test
    result, err := client.ListTasks(nil, nil, true)

    // Assert - Verify the results
    if err != nil {
        t.Fatalf("Expected no error, got: %v", err)
    }
    if len(result.Tasks) != expectedCount {
        t.Errorf("Expected %d tasks, got %d", expectedCount, len(result.Tasks))
    }
}
```

### Table-Driven Tests

```go
func TestClient_ListTasks(t *testing.T) {
    tests := []struct {
        name          string
        projectID     *string
        status        *string
        expectError   bool
        expectedCount int
    }{
        {
            name:          "list all tasks",
            projectID:     nil,
            status:        nil,
            expectError:   false,
            expectedCount: 9,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Using Test Fixtures

```go
func TestTaskOperations(t *testing.T) {
    // Create test data using builders
    task := NewTaskBuilder().
        WithID("test-task").
        WithTitle("Test Task").
        WithStatus("todo").
        Build()

    // Use in tests
    server := SetupMockServerWithData()
    server.AddTask(task)
    defer server.Close()
}
```

### Benchmark Tests

```go
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
```

## 🎯 Test Best Practices

### General Guidelines

1. **Test Names**: Use descriptive names that explain what is being tested
2. **Arrange-Act-Assert**: Structure tests with clear setup, execution, and verification
3. **Independence**: Each test should be independent and not rely on other tests
4. **Deterministic**: Tests should produce consistent results across runs
5. **Fast**: Unit tests should run quickly (< 100ms each)

### Error Testing

```go
func TestClient_ErrorHandling(t *testing.T) {
    server := NewMockServer()
    defer server.Close()

    // Configure server to return error
    server.SetSimulatedError("tasks", SimulateNetworkError())

    client := NewClient(server.URL, "test-key")
    _, err := client.ListTasks(nil, nil, true)

    // Verify error is handled correctly
    AssertError(t, err)
    AssertErrorContains(t, err, "network error")
}
```

### Mock Usage

```go
func TestWithMockClient(t *testing.T) {
    mockClient := NewMockClient()

    // Set up expected response
    expectedTasks := []Task{DefaultTask()}
    mockClient.SetListTasksResponse(
        TasksResponseFixture(expectedTasks),
        nil,
    )

    // Test code that uses the client
    // ...

    // Verify mock was called correctly
    if mockClient.GetListTasksCallCount() != 1 {
        t.Error("Expected ListTasks to be called once")
    }
}
```

## 📊 Current Test Coverage

As of the latest implementation:

### API Client Package (`internal/archon`)
- **Coverage**: ~95%
- **Tests**: 11 test functions
- **Benchmarks**: 3 performance tests
- **Features Covered**:
  - All HTTP client methods
  - Error handling scenarios
  - Authentication
  - Request filtering
  - Mock server functionality

### UI Package (`internal/ui`)
- **Coverage**: ~15% (basic tests)
- **Tests**: 4 test functions
- **Areas to Expand**:
  - State management
  - Keyboard handling
  - View rendering
  - Modal interactions

### Configuration Package (`internal/config`)
- **Coverage**: ~60%
- **Tests**: Basic configuration loading

## 🚀 Test Automation

### Continuous Integration

Create `.github/workflows/test.yml`:

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: '1.24'

    - name: Run tests
      run: go test -v -race -coverprofile=coverage.out ./...

    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
```

### Pre-commit Hooks

Add to `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: local
    hooks:
      - id: go-test
        name: go test
        entry: go test ./...
        language: system
        pass_filenames: false

      - id: go-test-coverage
        name: go test coverage
        entry: bash -c 'go test -cover ./... | grep -E "coverage: [0-9]+\.[0-9]+%" | awk "{if(\$3+0 < 80) exit 1}"'
        language: system
        pass_filenames: false
```

### Local Development

Create `Makefile`:

```makefile
.PHONY: test test-verbose test-coverage test-bench

test:
	go test ./...

test-verbose:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-bench:
	go test -bench=. ./...

test-race:
	go test -race ./...

test-all: test-race test-coverage test-bench
	@echo "All tests completed successfully"
```

## 🔧 Troubleshooting

### Common Issues

**Import Cycles**
```bash
# Error: import cycle not allowed in test
# Solution: Move shared test utilities to separate package
```

**Race Conditions**
```bash
# Run with race detector
go test -race ./...
# Fix by adding proper synchronization (mutexes, channels)
```

**Slow Tests**
```bash
# Identify slow tests
go test -v ./... | grep -E "PASS|FAIL" | sort -k3 -nr
# Optimize by reducing test data size or using mocks
```

**Flaky Tests**
```bash
# Run multiple times to identify
go test -count=10 ./internal/package/...
# Fix by removing timing dependencies and external factors
```

### Debug Mode

```bash
# Run tests with additional logging
go test -v -args -debug ./...

# Run specific test with debugging
go test -run TestSpecificFunction -v ./internal/package/...
```

## 📚 Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Go Benchmark Guide](https://pkg.go.dev/testing#hdr-Benchmarks)
- [Table-Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Testify Library](https://github.com/stretchr/testify) (alternative assertion library)

## 🎯 Testing Goals

### Short-term (Current Phase)
- ✅ API client: 95%+ coverage
- 🔄 UI components: 80%+ coverage
- 🔄 Configuration: 90%+ coverage
- 🔄 Integration tests for critical workflows

### Long-term
- 🎯 Overall coverage: 85%+
- 🎯 Performance regression tests
- 🎯 End-to-end testing with real Archon server
- 🎯 Property-based testing for complex algorithms

---

*This testing guide is maintained as part of the LazyArchon architecture improvements initiative.*