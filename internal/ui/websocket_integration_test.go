package ui

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/yousfisaad/lazyarchon/internal/archon"
	"github.com/yousfisaad/lazyarchon/internal/interfaces"
)

// Ensure MockWebSocketClient implements interfaces.RealtimeClient
var _ interfaces.RealtimeClient = (*MockWebSocketClient)(nil)

// MockWebSocketClient implements a mock WebSocket client for testing
type MockWebSocketClient struct {
	connected bool
	eventCh   chan interface{}
	mu        sync.RWMutex

	// Event simulation functions
	SimulateTaskUpdate func(task archon.Task)
	SimulateTaskCreate func(task archon.Task)
	SimulateTaskDelete func(taskID string, task archon.Task)
	SimulateConnect    func()
	SimulateDisconnect func(err error)
}

// NewMockWebSocketClient creates a new mock WebSocket client
func NewMockWebSocketClient() *MockWebSocketClient {
	mock := &MockWebSocketClient{
		eventCh: make(chan interface{}, 100),
	}

	// Setup simulation functions
	mock.SimulateTaskUpdate = func(task archon.Task) {
		mock.eventCh <- archon.RealtimeTaskUpdateMsg{
			TaskID: task.ID,
			Task:   task,
			Old:    nil,
		}
	}

	mock.SimulateTaskCreate = func(task archon.Task) {
		mock.eventCh <- archon.RealtimeTaskCreateMsg{
			Task: task,
		}
	}

	mock.SimulateTaskDelete = func(taskID string, task archon.Task) {
		mock.eventCh <- archon.RealtimeTaskDeleteMsg{
			TaskID: taskID,
			Task:   task,
		}
	}

	mock.SimulateConnect = func() {
		mock.mu.Lock()
		mock.connected = true
		mock.mu.Unlock()
		mock.eventCh <- archon.RealtimeConnectedMsg{}
	}

	mock.SimulateDisconnect = func(err error) {
		mock.mu.Lock()
		mock.connected = false
		mock.mu.Unlock()
		mock.eventCh <- archon.RealtimeDisconnectedMsg{Error: err}
	}

	return mock
}

// Connect simulates connecting
func (m *MockWebSocketClient) Connect() error {
	m.mu.Lock()
	m.connected = true
	m.mu.Unlock()
	return nil
}

// Disconnect simulates disconnecting
func (m *MockWebSocketClient) Disconnect() error {
	m.mu.Lock()
	m.connected = false
	m.mu.Unlock()
	return nil
}

// IsConnected returns connection status
func (m *MockWebSocketClient) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connected
}

// SetEventHandlers is a no-op for the mock
func (m *MockWebSocketClient) SetEventHandlers(
	onTaskUpdate func(archon.TaskUpdateEvent),
	onTaskCreate func(archon.TaskCreateEvent),
	onTaskDelete func(archon.TaskDeleteEvent),
	onProjectUpdate func(archon.ProjectUpdateEvent),
	onConnect func(),
	onDisconnect func(error),
) {
	// No-op for mock
}

// GetEventChannel returns the event channel
func (m *MockWebSocketClient) GetEventChannel() <-chan interface{} {
	return m.eventCh
}

// TestWebSocketIntegration_ConnectionStatus tests connection status handling
func TestWebSocketIntegration_ConnectionStatus(t *testing.T) {
	model := NewModel(createTestConfig())
	mockWS := NewMockWebSocketClient()
	model.wsClient = mockWS

	// Test initial state
	if model.Data.connected {
		t.Error("Model should not be connected initially")
	}

	// Simulate connection
	mockWS.SimulateConnect()

	// Process connection event - need to convert from archon to UI type
	msg := <-mockWS.GetEventChannel()

	// Convert archon message to UI message (simulating what ListenForRealtimeEvents does)
	var uiMsg interface{}
	switch e := msg.(type) {
	case archon.RealtimeConnectedMsg:
		uiMsg = RealtimeConnectedMsg{}
	default:
		t.Fatalf("Unexpected message type: %T", e)
	}

	updatedModel, cmd := model.Update(uiMsg)
	model = updatedModel.(Model)

	if !model.Data.connected {
		t.Error("Model should be connected after connection event")
	}

	if model.GetConnectionStatusText() != "●" {
		t.Error("Expected connected status indicator")
	}

	// Verify command was returned (ListenForRealtimeEvents)
	if cmd == nil {
		t.Error("Expected command after connection event")
	}

	// Simulate disconnection
	mockWS.SimulateDisconnect(nil)

	// Process disconnection event
	msg = <-mockWS.GetEventChannel()

	// Convert archon message to UI message
	switch e := msg.(type) {
	case archon.RealtimeDisconnectedMsg:
		uiMsg = RealtimeDisconnectedMsg{Error: e.Error}
	default:
		t.Fatalf("Unexpected message type: %T", e)
	}

	updatedModel, _ = model.Update(uiMsg)
	model = updatedModel.(Model)

	if model.Data.connected {
		t.Error("Model should not be connected after disconnection event")
	}

	if model.GetConnectionStatusText() != "○" {
		t.Error("Expected disconnected status indicator")
	}
}

// TestWebSocketIntegration_TaskEvents tests real-time task event handling
func TestWebSocketIntegration_TaskEvents(t *testing.T) {
	model := NewModel(createTestConfig())
	model.Window.width = 120
	model.Window.height = 40
	mockWS := NewMockWebSocketClient()
	model.wsClient = mockWS

	// Setup initial tasks
	initialTasks := []archon.Task{
		{ID: "task1", Title: "Task 1", Status: "todo"},
		{ID: "task2", Title: "Task 2", Status: "doing"},
	}
	model.UpdateTasks(initialTasks)

	if len(model.Data.tasks) != 2 {
		t.Fatalf("Expected 2 initial tasks, got %d", len(model.Data.tasks))
	}

	// Test task update event
	updatedTask := archon.Task{
		ID:     "task1",
		Title:  "Updated Task 1",
		Status: "done",
	}

	mockWS.SimulateTaskUpdate(updatedTask)

	// Process the event - convert from archon to UI type
	msg := <-mockWS.GetEventChannel()
	var uiMsg interface{}
	switch e := msg.(type) {
	case archon.RealtimeTaskUpdateMsg:
		uiMsg = RealtimeTaskUpdateMsg{
			TaskID: e.TaskID,
			Task:   e.Task,
			Old:    e.Old,
		}
	default:
		t.Fatalf("Unexpected message type: %T", e)
	}

	updatedModel, cmd := model.Update(uiMsg)
	model = updatedModel.(Model)

	// Verify that a refresh command was returned
	if cmd == nil {
		t.Error("Expected refresh command after task update event")
	}

	// Test task creation event
	newTask := archon.Task{
		ID:     "task3",
		Title:  "New Task 3",
		Status: "todo",
	}

	mockWS.SimulateTaskCreate(newTask)

	// Process the event - convert from archon to UI type
	msg = <-mockWS.GetEventChannel()
	switch e := msg.(type) {
	case archon.RealtimeTaskCreateMsg:
		uiMsg = RealtimeTaskCreateMsg{Task: e.Task}
	default:
		t.Fatalf("Unexpected message type: %T", e)
	}

	updatedModel, cmd = model.Update(uiMsg)
	model = updatedModel.(Model)

	// Verify that a refresh command was returned
	if cmd == nil {
		t.Error("Expected refresh command after task create event")
	}

	// Test task deletion event
	mockWS.SimulateTaskDelete("task2", initialTasks[1])

	// Process the event - convert from archon to UI type
	msg = <-mockWS.GetEventChannel()
	switch e := msg.(type) {
	case archon.RealtimeTaskDeleteMsg:
		uiMsg = RealtimeTaskDeleteMsg{
			TaskID: e.TaskID,
			Task:   e.Task,
		}
	default:
		t.Fatalf("Unexpected message type: %T", e)
	}

	updatedModel, cmd = model.Update(uiMsg)
	model = updatedModel.(Model)

	// Verify that a refresh command was returned
	if cmd == nil {
		t.Error("Expected refresh command after task delete event")
	}
}

// TestWebSocketIntegration_EventListener tests the event listener command
func TestWebSocketIntegration_EventListener(t *testing.T) {
	mockWS := NewMockWebSocketClient()

	// Simulate a task update event
	testTask := archon.Task{
		ID:     "test-task",
		Title:  "Test Task",
		Status: "doing",
	}
	mockWS.SimulateTaskUpdate(testTask)

	// Create the event listener command
	cmd := ListenForRealtimeEvents(mockWS)

	// Execute the command
	result := cmd()

	// Verify we got the expected message type
	if updateMsg, ok := result.(RealtimeTaskUpdateMsg); ok {
		if updateMsg.TaskID != testTask.ID {
			t.Errorf("Expected task ID %s, got %s", testTask.ID, updateMsg.TaskID)
		}
		if updateMsg.Task.Title != testTask.Title {
			t.Errorf("Expected task title %s, got %s", testTask.Title, updateMsg.Task.Title)
		}
	} else {
		t.Errorf("Expected RealtimeTaskUpdateMsg, got %T", result)
	}
}

// TestWebSocketIntegration_InitializeRealtime tests WebSocket initialization
func TestWebSocketIntegration_InitializeRealtime(t *testing.T) {
	mockWS := NewMockWebSocketClient()

	// Create the initialization command
	cmd := InitializeRealtimeCmd(mockWS)

	// Execute the command
	result := cmd()

	// Should return a RealtimeConnectedMsg since mock Connect() succeeds
	if _, ok := result.(RealtimeConnectedMsg); !ok {
		t.Errorf("Expected RealtimeConnectedMsg, got %T", result)
	}

	// Verify that the mock client is connected
	if !mockWS.IsConnected() {
		t.Error("Expected mock client to be connected after initialization")
	}
}

// TestWebSocketIntegration_ConcurrentEvents tests handling multiple concurrent events
func TestWebSocketIntegration_ConcurrentEvents(t *testing.T) {
	model := NewModel(createTestConfig())
	model.Window.width = 120
	model.Window.height = 40
	mockWS := NewMockWebSocketClient()
	model.wsClient = mockWS

	// Setup test context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Track processed events
	var processedEvents []interface{}
	var eventsMu sync.Mutex

	// Process events in background with proper type conversion
	go func() {
		for {
			select {
			case event := <-mockWS.GetEventChannel():
				// Convert archon event to UI event (like ListenForRealtimeEvents does)
				var uiEvent interface{}
				switch e := event.(type) {
				case archon.RealtimeTaskCreateMsg:
					uiEvent = RealtimeTaskCreateMsg{Task: e.Task}
				case archon.RealtimeTaskUpdateMsg:
					uiEvent = RealtimeTaskUpdateMsg{
						TaskID: e.TaskID,
						Task:   e.Task,
						Old:    e.Old,
					}
				case archon.RealtimeTaskDeleteMsg:
					uiEvent = RealtimeTaskDeleteMsg{
						TaskID: e.TaskID,
						Task:   e.Task,
					}
				case archon.RealtimeConnectedMsg:
					uiEvent = RealtimeConnectedMsg{}
				case archon.RealtimeDisconnectedMsg:
					uiEvent = RealtimeDisconnectedMsg{Error: e.Error}
				default:
					continue // Skip unknown event types
				}

				// Process event through model
				updatedModel, _ := model.Update(uiEvent)
				model = updatedModel.(Model)

				// Track processed event (use UI event type for counting)
				eventsMu.Lock()
				processedEvents = append(processedEvents, uiEvent)
				eventsMu.Unlock()

			case <-ctx.Done():
				return
			}
		}
	}()

	// Simulate multiple concurrent events
	tasks := []archon.Task{
		{ID: "task1", Title: "Task 1", Status: "todo"},
		{ID: "task2", Title: "Task 2", Status: "doing"},
		{ID: "task3", Title: "Task 3", Status: "done"},
	}

	// Send events rapidly
	for i, task := range tasks {
		switch i % 3 {
		case 0:
			mockWS.SimulateTaskCreate(task)
		case 1:
			mockWS.SimulateTaskUpdate(task)
		case 2:
			mockWS.SimulateTaskDelete(task.ID, task)
		}
	}

	// Add connection events
	mockWS.SimulateConnect()
	mockWS.SimulateDisconnect(nil)

	// Wait for all events to be processed
	time.Sleep(200 * time.Millisecond)

	// Verify all events were processed
	eventsMu.Lock()
	defer eventsMu.Unlock()

	expectedEventCount := len(tasks) + 2 // tasks + connect + disconnect
	if len(processedEvents) != expectedEventCount {
		t.Errorf("Expected %d events, processed %d", expectedEventCount, len(processedEvents))
	}

	// Verify event types
	var taskCreateCount, taskUpdateCount, taskDeleteCount, connectCount, disconnectCount int
	for _, event := range processedEvents {
		switch event.(type) {
		case RealtimeTaskCreateMsg:
			taskCreateCount++
		case RealtimeTaskUpdateMsg:
			taskUpdateCount++
		case RealtimeTaskDeleteMsg:
			taskDeleteCount++
		case RealtimeConnectedMsg:
			connectCount++
		case RealtimeDisconnectedMsg:
			disconnectCount++
		}
	}

	if taskCreateCount != 1 || taskUpdateCount != 1 || taskDeleteCount != 1 {
		t.Errorf("Expected 1 of each task event type, got create:%d update:%d delete:%d",
			taskCreateCount, taskUpdateCount, taskDeleteCount)
	}

	if connectCount != 1 || disconnectCount != 1 {
		t.Errorf("Expected 1 connect and 1 disconnect event, got connect:%d disconnect:%d",
			connectCount, disconnectCount)
	}
}

// TestWebSocketIntegration_MessageTypeConversion tests type conversion between archon and ui packages
func TestWebSocketIntegration_MessageTypeConversion(t *testing.T) {
	mockWS := NewMockWebSocketClient()

	// Test task update conversion
	testTask := archon.Task{
		ID:     "test-task",
		Title:  "Test Task",
		Status: "doing",
	}
	mockWS.SimulateTaskUpdate(testTask)

	cmd := ListenForRealtimeEvents(mockWS)
	result := cmd()

	if updateMsg, ok := result.(RealtimeTaskUpdateMsg); ok {
		// Verify field conversion
		if updateMsg.TaskID != testTask.ID {
			t.Errorf("TaskID conversion failed: expected %s, got %s", testTask.ID, updateMsg.TaskID)
		}
		if updateMsg.Task.ID != testTask.ID {
			t.Errorf("Task.ID conversion failed: expected %s, got %s", testTask.ID, updateMsg.Task.ID)
		}
		if updateMsg.Task.Title != testTask.Title {
			t.Errorf("Task.Title conversion failed: expected %s, got %s", testTask.Title, updateMsg.Task.Title)
		}
		if updateMsg.Task.Status != testTask.Status {
			t.Errorf("Task.Status conversion failed: expected %s, got %s", testTask.Status, updateMsg.Task.Status)
		}
	} else {
		t.Errorf("Expected RealtimeTaskUpdateMsg, got %T", result)
	}

	// Test connection status conversion
	mockWS.SimulateConnect()
	cmd = ListenForRealtimeEvents(mockWS)
	result = cmd()

	if _, ok := result.(RealtimeConnectedMsg); !ok {
		t.Errorf("Expected RealtimeConnectedMsg, got %T", result)
	}
}

// BenchmarkWebSocketIntegration_EventProcessing benchmarks event processing performance
func BenchmarkWebSocketIntegration_EventProcessing(b *testing.B) {
	model := NewModel(createTestConfig())
	model.Window.width = 120
	model.Window.height = 40
	mockWS := NewMockWebSocketClient()
	model.wsClient = mockWS

	testTask := archon.Task{
		ID:     "bench-task",
		Title:  "Benchmark Task",
		Status: "doing",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Simulate event
			mockWS.SimulateTaskUpdate(testTask)

			// Process event
			msg := <-mockWS.GetEventChannel()
			_, _ = model.Update(msg)
		}
	})
}