package archon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketClient handles real-time communication with Archon's Supabase backend
type WebSocketClient struct {
	conn           *websocket.Conn
	url            string
	apiKey         string
	ctx            context.Context
	cancel         context.CancelFunc
	mu             sync.RWMutex
	connected      bool
	reconnectDelay time.Duration
	maxReconnects  int
	reconnectCount int
	logger         Logger // Optional logger for debug mode

	// Internal channels
	sendCh   chan []byte
	closeCh  chan struct{}
	reconnCh chan struct{}

	// Bubble Tea message channel for real-time events
	eventCh chan interface{}
}

// Real-time event types matching Supabase realtime format
type RealtimeEvent struct {
	Event   string          `json:"event"`
	Topic   string          `json:"topic"`
	Payload RealtimePayload `json:"payload"`
	Ref     string          `json:"ref"`
}

type RealtimePayload struct {
	Schema string                 `json:"schema"`
	Table  string                 `json:"table"`
	Type   string                 `json:"type"`
	Record map[string]interface{} `json:"record"`
	Old    map[string]interface{} `json:"old_record,omitempty"`
}

// Typed event structures for application use
type TaskUpdateEvent struct {
	TaskID string `json:"task_id"`
	Task   Task   `json:"task"`
	Old    *Task  `json:"old,omitempty"`
}

type TaskCreateEvent struct {
	Task Task `json:"task"`
}

type TaskDeleteEvent struct {
	TaskID string `json:"task_id"`
	Task   Task   `json:"task"`
}

type ProjectUpdateEvent struct {
	ProjectID string   `json:"project_id"`
	Project   Project  `json:"project"`
	Old       *Project `json:"old,omitempty"`
}

// Bubble Tea message types for WebSocket events
// These mirror the types in internal/ui/commands.go to avoid circular imports
type RealtimeTaskCreateMsg struct {
	Task Task `json:"task"`
}

type RealtimeTaskUpdateMsg struct {
	TaskID string `json:"task_id"`
	Task   Task   `json:"task"`
	Old    *Task  `json:"old,omitempty"`
}

type RealtimeTaskDeleteMsg struct {
	TaskID string `json:"task_id"`
	Task   Task   `json:"task"`
}

type RealtimeProjectUpdateMsg struct {
	ProjectID string   `json:"project_id"`
	Project   Project  `json:"project"`
	Old       *Project `json:"old,omitempty"`
}

type RealtimeConnectedMsg struct{}

type RealtimeDisconnectedMsg struct {
	Error error
}

// NewWebSocketClient creates a new WebSocket client for real-time updates
func NewWebSocketClient(baseURL, apiKey string) *WebSocketClient {
	ctx, cancel := context.WithCancel(context.Background())

	// Convert HTTP URL to WebSocket URL
	wsURL := convertToWebSocketURL(baseURL)

	return &WebSocketClient{
		url:            wsURL,
		apiKey:         apiKey,
		ctx:            ctx,
		cancel:         cancel,
		reconnectDelay: 5 * time.Second,
		maxReconnects:  10,
		sendCh:         make(chan []byte, 100),
		closeCh:        make(chan struct{}),
		reconnCh:       make(chan struct{}, 1),
		eventCh:        make(chan interface{}, 100),
		logger:         nil, // No logger by default
	}
}

// SetLogger sets the optional logger for the WebSocket client
func (ws *WebSocketClient) SetLogger(logger Logger) {
	ws.logger = logger
}

// convertToWebSocketURL converts HTTP/HTTPS URL to WebSocket URL
func convertToWebSocketURL(httpURL string) string {
	// Parse the HTTP URL
	u, err := url.Parse(httpURL)
	if err != nil {
		return "ws://localhost:8000/realtime/v1/websocket"
	}

	// Convert scheme
	if u.Scheme == "https" {
		u.Scheme = "wss"
	} else {
		u.Scheme = "ws"
	}

	// Set the realtime path
	u.Path = "/realtime/v1/websocket"

	return u.String()
}

// Connect establishes the WebSocket connection
func (ws *WebSocketClient) Connect() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.connected {
		return nil
	}

	// Add query parameters for Supabase realtime
	u, _ := url.Parse(ws.url)
	q := u.Query()
	q.Set("apikey", ws.apiKey)
	q.Set("vsn", "1.0.0")
	u.RawQuery = q.Encode()

	// Establish WebSocket connection
	if ws.logger != nil {
		ws.logger.LogWebSocketEvent("connecting", "url", ws.url)
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		if ws.logger != nil {
			ws.logger.Error("Failed to connect to WebSocket", "error", err, "url", ws.url)
		}
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	ws.conn = conn
	ws.connected = true
	ws.reconnectCount = 0

	if ws.logger != nil {
		ws.logger.LogWebSocketEvent("connected", "url", ws.url)
	}

	// Start goroutines for handling messages
	go ws.readPump()
	go ws.writePump()
	go ws.pingPump()

	// Send initial join message for tasks channel
	if err := ws.joinChannel("realtime:public:tasks"); err != nil {
		if ws.logger != nil {
			ws.logger.Error("Failed to join tasks channel", "error", err)
		} else {
			log.Printf("Failed to join tasks channel: %v", err)
		}
	}

	// Send initial join message for projects channel
	if err := ws.joinChannel("realtime:public:projects"); err != nil {
		if ws.logger != nil {
			ws.logger.Error("Failed to join projects channel", "error", err)
		} else {
			log.Printf("Failed to join projects channel: %v", err)
		}
	}

	// Notify connection established via channel only

	// Send connection event to channel
	select {
	case ws.eventCh <- RealtimeConnectedMsg{}:
	default:
		if ws.logger != nil {
			ws.logger.Error("Event channel full, dropping connection event")
		} else {
			log.Printf("Event channel full, dropping connection event")
		}
	}

	return nil
}

// Disconnect closes the WebSocket connection
func (ws *WebSocketClient) Disconnect() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if !ws.connected {
		return nil
	}

	if ws.logger != nil {
		ws.logger.LogWebSocketEvent("disconnecting", "url", ws.url)
	}

	ws.cancel()
	close(ws.closeCh)

	if ws.conn != nil {
		err := ws.conn.Close()
		ws.conn = nil
		ws.connected = false

		if ws.logger != nil {
			if err != nil {
				ws.logger.Error("Error during WebSocket disconnect", "error", err)
			} else {
				ws.logger.LogWebSocketEvent("disconnected", "url", ws.url)
			}
		}
		return err
	}

	return nil
}

// IsConnected returns the current connection status
func (ws *WebSocketClient) IsConnected() bool {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.connected
}

// GetEventChannel returns the channel for receiving real-time events
func (ws *WebSocketClient) GetEventChannel() <-chan interface{} {
	return ws.eventCh
}

// joinChannel sends a join message for a specific channel
func (ws *WebSocketClient) joinChannel(topic string) error {
	joinMsg := map[string]interface{}{
		"topic":   topic,
		"event":   "phx_join",
		"payload": map[string]interface{}{},
		"ref":     fmt.Sprintf("%d", time.Now().UnixNano()),
	}

	msgBytes, err := json.Marshal(joinMsg)
	if err != nil {
		return err
	}

	select {
	case ws.sendCh <- msgBytes:
		return nil
	case <-ws.ctx.Done():
		return ws.ctx.Err()
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout sending join message")
	}
}

// readPump handles incoming WebSocket messages
func (ws *WebSocketClient) readPump() {
	defer func() {
		ws.mu.Lock()
		ws.connected = false
		ws.mu.Unlock()

		disconnectError := fmt.Errorf("read pump closed")

		// Notify disconnection via channel only

		// Send disconnection event to channel
		select {
		case ws.eventCh <- RealtimeDisconnectedMsg{Error: disconnectError}:
		default:
			log.Printf("Event channel full, dropping disconnection event")
		}

		// Trigger reconnection
		select {
		case ws.reconnCh <- struct{}{}:
		default:
		}
	}()

	for {
		select {
		case <-ws.ctx.Done():
			return
		default:
		}

		ws.mu.RLock()
		conn := ws.conn
		ws.mu.RUnlock()

		if conn == nil {
			log.Printf("WebSocket connection is nil in readPump")
			return
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			return
		}

		ws.handleMessage(message)
	}
}

// writePump handles outgoing WebSocket messages
func (ws *WebSocketClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ws.ctx.Done():
			return
		case <-ws.closeCh:
			return
		case message := <-ws.sendCh:
			ws.mu.RLock()
			conn := ws.conn
			ws.mu.RUnlock()

			if conn == nil {
				log.Printf("WebSocket connection is nil, dropping message")
				return
			}

			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		case <-ticker.C:
			ws.mu.RLock()
			conn := ws.conn
			ws.mu.RUnlock()

			if conn == nil {
				continue // Skip ping if not connected
			}

			// Send ping
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("WebSocket ping error: %v", err)
				return
			}
		}
	}
}

// pingPump handles ping/pong for connection health
func (ws *WebSocketClient) pingPump() {
	for {
		select {
		case <-ws.ctx.Done():
			return
		case <-ws.reconnCh:
			if ws.shouldReconnect() {
				ws.reconnect()
			}
		}
	}
}

// shouldReconnect determines if we should attempt to reconnect
func (ws *WebSocketClient) shouldReconnect() bool {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	return !ws.connected && ws.reconnectCount < ws.maxReconnects
}

// reconnect attempts to reestablish the WebSocket connection
func (ws *WebSocketClient) reconnect() {
	ws.mu.Lock()
	ws.reconnectCount++
	delay := time.Duration(ws.reconnectCount) * ws.reconnectDelay
	ws.mu.Unlock()

	if ws.logger != nil {
		ws.logger.LogWebSocketEvent("reconnecting",
			"attempt", ws.reconnectCount,
			"max_attempts", ws.maxReconnects,
			"delay", delay)
	} else {
		log.Printf("Attempting to reconnect WebSocket (attempt %d/%d) in %v",
			ws.reconnectCount, ws.maxReconnects, delay)
	}

	time.Sleep(delay)

	if err := ws.Connect(); err != nil {
		if ws.logger != nil {
			ws.logger.Error("Reconnection failed", "error", err, "attempt", ws.reconnectCount)
		} else {
			log.Printf("Reconnection failed: %v", err)
		}

		// Trigger another reconnection attempt
		go func() {
			select {
			case ws.reconnCh <- struct{}{}:
			default:
			}
		}()
	}
}

// handleMessage processes incoming WebSocket messages
func (ws *WebSocketClient) handleMessage(message []byte) {
	var event RealtimeEvent
	if err := json.Unmarshal(message, &event); err != nil {
		log.Printf("Failed to unmarshal WebSocket message: %v", err)
		return
	}

	// Handle different event types
	switch event.Event {
	case "postgres_changes":
		ws.handlePostgresChange(event)
	case "phx_reply":
		// Handle join confirmations, etc.
		log.Printf("Channel reply: %s", string(message))
	case "heartbeat":
		// Connection health check
	default:
		log.Printf("Unknown event type: %s", event.Event)
	}
}

// handlePostgresChange processes database change events
func (ws *WebSocketClient) handlePostgresChange(event RealtimeEvent) {
	payload := event.Payload

	switch payload.Table {
	case "tasks":
		ws.handleTaskChange(payload)
	case "projects":
		ws.handleProjectChange(payload)
	default:
		log.Printf("Unhandled table change: %s", payload.Table)
	}
}

// handleTaskChange processes task-related database changes
func (ws *WebSocketClient) handleTaskChange(payload RealtimePayload) {
	switch payload.Type {
	case "INSERT":
		task := ws.mapToTask(payload.Record)

		if ws.logger != nil {
			ws.logger.Debug("Task created via realtime", "task_id", task.ID, "title", task.Title)
		}

		// Send to event channel for Bubble Tea integration
		select {
		case ws.eventCh <- RealtimeTaskCreateMsg{Task: task}:
		default:
			// Channel full, log and drop event
			if ws.logger != nil {
				ws.logger.Error("Event channel full, dropping task create event", "task_id", task.ID)
			} else {
				log.Printf("Event channel full, dropping task create event for task %s", task.ID)
			}
		}

	case "UPDATE":
		task := ws.mapToTask(payload.Record)
		var oldTask *Task
		if payload.Old != nil {
			old := ws.mapToTask(payload.Old)
			oldTask = &old
		}

		if ws.logger != nil {
			ws.logger.Debug("Task updated via realtime", "task_id", task.ID, "status", task.Status)
		}

		// Send to event channel for Bubble Tea integration
		select {
		case ws.eventCh <- RealtimeTaskUpdateMsg{
			TaskID: task.ID,
			Task:   task,
			Old:    oldTask,
		}:
		default:
			if ws.logger != nil {
				ws.logger.Error("Event channel full, dropping task update event", "task_id", task.ID)
			} else {
				log.Printf("Event channel full, dropping task update event for task %s", task.ID)
			}
		}

	case "DELETE":
		task := ws.mapToTask(payload.Record)

		if ws.logger != nil {
			ws.logger.Debug("Task deleted via realtime", "task_id", task.ID)
		}

		// Send to event channel for Bubble Tea integration
		select {
		case ws.eventCh <- RealtimeTaskDeleteMsg{
			TaskID: task.ID,
			Task:   task,
		}:
		default:
			if ws.logger != nil {
				ws.logger.Error("Event channel full, dropping task delete event", "task_id", task.ID)
			} else {
				log.Printf("Event channel full, dropping task delete event for task %s", task.ID)
			}
		}
	}
}

// handleProjectChange processes project-related database changes
func (ws *WebSocketClient) handleProjectChange(payload RealtimePayload) {
	switch payload.Type {
	case "UPDATE":
		project := ws.mapToProject(payload.Record)
		var oldProject *Project
		if payload.Old != nil {
			old := ws.mapToProject(payload.Old)
			oldProject = &old
		}

		// Project update handled via channel only

		// Send to event channel for Bubble Tea integration
		select {
		case ws.eventCh <- RealtimeProjectUpdateMsg{
			ProjectID: project.ID,
			Project:   project,
			Old:       oldProject,
		}:
		default:
			log.Printf("Event channel full, dropping project update event for project %s", project.ID)
		}
	}
}

// mapToTask converts a map to a Task struct
func (ws *WebSocketClient) mapToTask(record map[string]interface{}) Task {
	// Convert the map to JSON and then unmarshal to Task struct
	// This handles type conversions and nested structures
	jsonBytes, _ := json.Marshal(record)
	var task Task
	json.Unmarshal(jsonBytes, &task)
	return task
}

// mapToProject converts a map to a Project struct
func (ws *WebSocketClient) mapToProject(record map[string]interface{}) Project {
	jsonBytes, _ := json.Marshal(record)
	var project Project
	json.Unmarshal(jsonBytes, &project)
	return project
}
