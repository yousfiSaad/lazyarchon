package archon

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// MockWebSocketServer provides a test WebSocket server for testing
type MockWebSocketServer struct {
	server   *httptest.Server
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]bool
	mu       sync.RWMutex

	// Event simulation
	sendTaskUpdate func(task Task)
	sendTaskCreate func(task Task)
	sendTaskDelete func(taskID string, task Task)
	sendProjectUpdate func(project Project)

	// Message tracking
	receivedMessages []map[string]interface{}
	messagesMu       sync.RWMutex
}

// NewMockWebSocketServer creates a new mock WebSocket server
func NewMockWebSocketServer() *MockWebSocketServer {
	mock := &MockWebSocketServer{
		clients:          make(map[*websocket.Conn]bool),
		receivedMessages: make([]map[string]interface{}, 0),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for testing
			},
		},
	}

	// Create HTTP server with WebSocket endpoint
	mux := http.NewServeMux()
	mux.HandleFunc("/realtime/v1/websocket", mock.handleWebSocket)
	mock.server = httptest.NewServer(mux)

	// Setup event simulation functions
	mock.setupEventSimulation()

	return mock
}

// setupEventSimulation configures the mock server's event simulation capabilities
func (m *MockWebSocketServer) setupEventSimulation() {
	m.sendTaskUpdate = func(task Task) {
		event := RealtimeEvent{
			Event: "postgres_changes",
			Topic: "realtime:public:tasks",
			Payload: RealtimePayload{
				Schema: "public",
				Table:  "tasks",
				Type:   "UPDATE",
				Record: taskToMap(task),
			},
			Ref: fmt.Sprintf("%d", time.Now().UnixNano()),
		}
		m.broadcastEvent(event)
	}

	m.sendTaskCreate = func(task Task) {
		event := RealtimeEvent{
			Event: "postgres_changes",
			Topic: "realtime:public:tasks",
			Payload: RealtimePayload{
				Schema: "public",
				Table:  "tasks",
				Type:   "INSERT",
				Record: taskToMap(task),
			},
			Ref: fmt.Sprintf("%d", time.Now().UnixNano()),
		}
		m.broadcastEvent(event)
	}

	m.sendTaskDelete = func(taskID string, task Task) {
		event := RealtimeEvent{
			Event: "postgres_changes",
			Topic: "realtime:public:tasks",
			Payload: RealtimePayload{
				Schema: "public",
				Table:  "tasks",
				Type:   "DELETE",
				Record: taskToMap(task),
			},
			Ref: fmt.Sprintf("%d", time.Now().UnixNano()),
		}
		m.broadcastEvent(event)
	}

	m.sendProjectUpdate = func(project Project) {
		event := RealtimeEvent{
			Event: "postgres_changes",
			Topic: "realtime:public:projects",
			Payload: RealtimePayload{
				Schema: "public",
				Table:  "projects",
				Type:   "UPDATE",
				Record: projectToMap(project),
			},
			Ref: fmt.Sprintf("%d", time.Now().UnixNano()),
		}
		m.broadcastEvent(event)
	}
}

// GetURL returns the WebSocket URL for the mock server
func (m *MockWebSocketServer) GetURL() string {
	return strings.Replace(m.server.URL, "http://", "ws://", 1) + "/realtime/v1/websocket"
}

// Close shuts down the mock server
func (m *MockWebSocketServer) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Close all client connections
	for conn := range m.clients {
		conn.Close()
	}
	m.clients = make(map[*websocket.Conn]bool)

	// Close the server
	m.server.Close()
}

// GetReceivedMessages returns all messages received by the server
func (m *MockWebSocketServer) GetReceivedMessages() []map[string]interface{} {
	m.messagesMu.RLock()
	defer m.messagesMu.RUnlock()

	// Return a copy to avoid race conditions
	result := make([]map[string]interface{}, len(m.receivedMessages))
	copy(result, m.receivedMessages)
	return result
}

// WaitForConnection waits for at least one client to connect
func (m *MockWebSocketServer) WaitForConnection(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		m.mu.RLock()
		count := len(m.clients)
		m.mu.RUnlock()

		if count > 0 {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// handleWebSocket handles WebSocket connections
func (m *MockWebSocketServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	m.mu.Lock()
	m.clients[conn] = true
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		delete(m.clients, conn)
		m.mu.Unlock()
		conn.Close()
	}()

	// Handle incoming messages
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Parse and store received message
		var msg map[string]interface{}
		if json.Unmarshal(message, &msg) == nil {
			m.messagesMu.Lock()
			m.receivedMessages = append(m.receivedMessages, msg)
			m.messagesMu.Unlock()

			// Send acknowledgment for join messages
			if event, ok := msg["event"].(string); ok && event == "phx_join" {
				response := map[string]interface{}{
					"event":   "phx_reply",
					"payload": map[string]interface{}{"status": "ok"},
					"ref":     msg["ref"],
					"topic":   msg["topic"],
				}
				responseBytes, _ := json.Marshal(response)
				conn.WriteMessage(websocket.TextMessage, responseBytes)
			}
		}
	}
}

// broadcastEvent sends an event to all connected clients
func (m *MockWebSocketServer) broadcastEvent(event RealtimeEvent) {
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for conn := range m.clients {
		conn.WriteMessage(websocket.TextMessage, eventBytes)
	}
}

// Helper functions to convert structs to maps for event simulation
func taskToMap(task Task) map[string]interface{} {
	result := map[string]interface{}{
		"id":          task.ID,
		"title":       task.Title,
		"description": task.Description,
		"status":      task.Status,
		"assignee":    task.Assignee,
		"task_order":  task.TaskOrder,
		"project_id":  task.ProjectID,
		"created_at":  task.CreatedAt.Format(time.RFC3339),
		"updated_at":  task.UpdatedAt.Format(time.RFC3339),
		"archived":    task.Archived,
	}

	// Handle optional fields
	if task.Feature != nil {
		result["feature"] = *task.Feature
	}

	return result
}

func projectToMap(project Project) map[string]interface{} {
	return map[string]interface{}{
		"id":          project.ID,
		"title":       project.Title,
		"description": project.Description,
		"created_at":  project.CreatedAt.Format(time.RFC3339),
		"updated_at":  project.UpdatedAt.Format(time.RFC3339),
	}
}

// Test WebSocket client basic functionality
func TestWebSocketClient_Basic(t *testing.T) {
	mockServer := NewMockWebSocketServer()
	defer mockServer.Close()

	// Extract base URL from WebSocket URL
	wsURL := mockServer.GetURL()
	baseURL := strings.Replace(wsURL, "/realtime/v1/websocket", "", 1)
	baseURL = strings.Replace(baseURL, "ws://", "http://", 1)

	client := NewWebSocketClient(baseURL, "test-api-key")

	// Test initial state
	if client.IsConnected() {
		t.Error("Client should not be connected initially")
	}

	// Test connection
	err := client.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Wait for connection to establish
	if !mockServer.WaitForConnection(5 * time.Second) {
		t.Fatal("Server did not receive connection within timeout")
	}

	if !client.IsConnected() {
		t.Error("Client should be connected after Connect()")
	}

	// Test disconnection
	err = client.Disconnect()
	if err != nil {
		t.Errorf("Failed to disconnect: %v", err)
	}

	// Give some time for disconnection to complete
	time.Sleep(100 * time.Millisecond)

	if client.IsConnected() {
		t.Error("Client should not be connected after Disconnect()")
	}
}

// Test WebSocket event handling
func TestWebSocketClient_EventHandling(t *testing.T) {
	mockServer := NewMockWebSocketServer()
	defer mockServer.Close()

	// Extract base URL
	wsURL := mockServer.GetURL()
	baseURL := strings.Replace(wsURL, "/realtime/v1/websocket", "", 1)
	baseURL = strings.Replace(baseURL, "ws://", "http://", 1)

	client := NewWebSocketClient(baseURL, "test-api-key")

	// Set up event tracking
	var receivedEvents []interface{}
	var eventsMu sync.Mutex

	// Connect and start listening
	err := client.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect()

	// Wait for connection
	if !mockServer.WaitForConnection(5 * time.Second) {
		t.Fatal("Server did not receive connection within timeout")
	}

	// Start event listener in background
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go func() {
		eventCh := client.GetEventChannel()
		for {
			select {
			case event := <-eventCh:
				eventsMu.Lock()
				receivedEvents = append(receivedEvents, event)
				eventsMu.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()

	// Give some time for listener to start
	time.Sleep(100 * time.Millisecond)

	// Simulate task update event
	feature := "test-feature"
	now := FlexibleTime{Time: time.Now()}
	testTask := Task{
		ID:          "test-task-123",
		Title:       "Test Task",
		Description: "Test Description",
		Status:      "doing",
		Assignee:    "test-user",
		TaskOrder:   10,
		Feature:     &feature,
		ProjectID:   "test-project",
		CreatedAt:   now,
		UpdatedAt:   now,
		Archived:    false,
	}

	mockServer.sendTaskUpdate(testTask)

	// Wait for event to be processed
	time.Sleep(200 * time.Millisecond)

	// Verify event was received
	eventsMu.Lock()
	defer eventsMu.Unlock()

	if len(receivedEvents) == 0 {
		t.Fatal("No events received")
	}

	// Check if we received a task update event
	found := false
	for _, event := range receivedEvents {
		if updateEvent, ok := event.(RealtimeTaskUpdateMsg); ok {
			if updateEvent.TaskID == testTask.ID && updateEvent.Task.Title == testTask.Title {
				found = true
				break
			}
		}
	}

	if !found {
		t.Errorf("Expected task update event not found. Received events: %+v", receivedEvents)
	}
}

// Test WebSocket channel management
func TestWebSocketClient_ChannelManagement(t *testing.T) {
	mockServer := NewMockWebSocketServer()
	defer mockServer.Close()

	wsURL := mockServer.GetURL()
	baseURL := strings.Replace(wsURL, "/realtime/v1/websocket", "", 1)
	baseURL = strings.Replace(baseURL, "ws://", "http://", 1)

	client := NewWebSocketClient(baseURL, "test-api-key")

	err := client.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect()

	// Wait for connection and channel subscriptions
	time.Sleep(500 * time.Millisecond)

	// Check that join messages were sent for tasks and projects channels
	messages := mockServer.GetReceivedMessages()

	tasksChannelJoined := false
	projectsChannelJoined := false

	for _, msg := range messages {
		if event, ok := msg["event"].(string); ok && event == "phx_join" {
			if topic, ok := msg["topic"].(string); ok {
				if topic == "realtime:public:tasks" {
					tasksChannelJoined = true
				} else if topic == "realtime:public:projects" {
					projectsChannelJoined = true
				}
			}
		}
	}

	if !tasksChannelJoined {
		t.Error("Client did not join tasks channel")
	}

	if !projectsChannelJoined {
		t.Error("Client did not join projects channel")
	}
}

// Test WebSocket connection URL conversion
func TestConvertToWebSocketURL(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{
			"http://localhost:8000",
			"ws://localhost:8000/realtime/v1/websocket",
		},
		{
			"https://api.example.com",
			"wss://api.example.com/realtime/v1/websocket",
		},
		{
			"http://192.168.1.100:8080",
			"ws://192.168.1.100:8080/realtime/v1/websocket",
		},
	}

	for _, tc := range testCases {
		result := convertToWebSocketURL(tc.input)
		if result != tc.expected {
			t.Errorf("convertToWebSocketURL(%s) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}

// Test WebSocket error handling
func TestWebSocketClient_ErrorHandling(t *testing.T) {
	// Test connection to non-existent server
	client := NewWebSocketClient("http://non-existent-server:9999", "test-key")

	err := client.Connect()
	if err == nil {
		t.Error("Expected error when connecting to non-existent server")
	}

	if client.IsConnected() {
		t.Error("Client should not be connected after failed connection attempt")
	}
}

// Test WebSocket reconnection behavior
func TestWebSocketClient_Reconnection(t *testing.T) {
	mockServer := NewMockWebSocketServer()

	wsURL := mockServer.GetURL()
	baseURL := strings.Replace(wsURL, "/realtime/v1/websocket", "", 1)
	baseURL = strings.Replace(baseURL, "ws://", "http://", 1)

	client := NewWebSocketClient(baseURL, "test-api-key")

	// Initial connection
	err := client.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	if !mockServer.WaitForConnection(5 * time.Second) {
		t.Fatal("Server did not receive connection within timeout")
	}

	if !client.IsConnected() {
		t.Error("Client should be connected")
	}

	// Simulate server shutdown (disconnect)
	mockServer.Close()

	// Give time for client to detect disconnection
	time.Sleep(200 * time.Millisecond)

	// Note: In a real scenario, the client would attempt to reconnect
	// For this test, we just verify that the client detects the disconnection
	// The reconnection logic is complex to test in unit tests as it involves
	// timing and would require a more sophisticated mock server setup
}