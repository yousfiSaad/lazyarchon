package logger

import (
	"log/slog"
	"sync"
	"time"
)

// UsageEvent represents a user interaction or system event
type UsageEvent struct {
	Timestamp  time.Time              `json:"timestamp"`
	EventType  string                 `json:"event_type"`
	Component  string                 `json:"component"`
	Action     string                 `json:"action"`
	Properties map[string]interface{} `json:"properties"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	SessionID  string                 `json:"session_id,omitempty"`
}

// Telemetry tracks usage patterns and application analytics
type Telemetry struct {
	mu        sync.RWMutex
	logger    *Logger
	events    []UsageEvent
	sessionID string
	startTime time.Time

	// Aggregated statistics
	keyboardEvents map[string]int64
	apiCalls       map[string]int64
	modalActions   map[string]int64
	errors         map[string]int64
}

// NewTelemetry creates a new telemetry tracker
func NewTelemetry(logger *Logger, sessionID string) *Telemetry {
	return &Telemetry{
		logger:         logger,
		sessionID:      sessionID,
		startTime:      time.Now(),
		events:         make([]UsageEvent, 0),
		keyboardEvents: make(map[string]int64),
		apiCalls:       make(map[string]int64),
		modalActions:   make(map[string]int64),
		errors:         make(map[string]int64),
	}
}

// TrackKeyboardInput tracks keyboard input events
func (t *Telemetry) TrackKeyboardInput(key, mode string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.keyboardEvents[key]++

	event := UsageEvent{
		Timestamp: time.Now(),
		EventType: "keyboard",
		Component: "input",
		Action:    key,
		Properties: map[string]interface{}{
			"mode": mode,
			"key":  key,
		},
		SessionID: t.sessionID,
	}

	t.events = append(t.events, event)
	t.logger.Usage("keyboard_input",
		slog.String("key", key),
		slog.String("mode", mode),
		slog.String("session_id", t.sessionID),
	)
}

// TrackModalAction tracks modal interactions
func (t *Telemetry) TrackModalAction(modalType, action string, properties map[string]interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	actionKey := modalType + ":" + action
	t.modalActions[actionKey]++

	if properties == nil {
		properties = make(map[string]interface{})
	}
	properties["modal_type"] = modalType

	event := UsageEvent{
		Timestamp:  time.Now(),
		EventType:  "modal",
		Component:  modalType,
		Action:     action,
		Properties: properties,
		SessionID:  t.sessionID,
	}

	t.events = append(t.events, event)
	t.logger.Usage("modal_action",
		slog.String("modal_type", modalType),
		slog.String("action", action),
		slog.Any("properties", properties),
		slog.String("session_id", t.sessionID),
	)
}

// TrackAPICall tracks API interactions
func (t *Telemetry) TrackAPICall(endpoint, method string, statusCode int, duration time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	callKey := method + " " + endpoint
	t.apiCalls[callKey]++

	event := UsageEvent{
		Timestamp: time.Now(),
		EventType: "api",
		Component: "http_client",
		Action:    callKey,
		Properties: map[string]interface{}{
			"endpoint":    endpoint,
			"method":      method,
			"status_code": statusCode,
			"duration_ms": duration.Milliseconds(),
			"success":     statusCode < 400,
		},
		SessionID: t.sessionID,
	}

	t.events = append(t.events, event)
	t.logger.Usage("api_call",
		slog.String("endpoint", endpoint),
		slog.String("method", method),
		slog.Int("status_code", statusCode),
		slog.Duration("duration", duration),
		slog.String("session_id", t.sessionID),
	)
}

// TrackError tracks error occurrences
func (t *Telemetry) TrackError(errorType, component string, err error, properties map[string]interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.errors[errorType]++

	if properties == nil {
		properties = make(map[string]interface{})
	}
	properties["error_type"] = errorType
	properties["error_message"] = err.Error()

	event := UsageEvent{
		Timestamp:  time.Now(),
		EventType:  "error",
		Component:  component,
		Action:     errorType,
		Properties: properties,
		SessionID:  t.sessionID,
	}

	t.events = append(t.events, event)
	t.logger.Usage("error",
		slog.String("error_type", errorType),
		slog.String("component", component),
		slog.String("error", err.Error()),
		slog.Any("properties", properties),
		slog.String("session_id", t.sessionID),
	)
}

// TrackUserJourney tracks high-level user workflow patterns
func (t *Telemetry) TrackUserJourney(step string, properties map[string]interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if properties == nil {
		properties = make(map[string]interface{})
	}

	event := UsageEvent{
		Timestamp:  time.Now(),
		EventType:  "journey",
		Component:  "workflow",
		Action:     step,
		Properties: properties,
		SessionID:  t.sessionID,
	}

	t.events = append(t.events, event)
	t.logger.Usage("user_journey",
		slog.String("step", step),
		slog.Any("properties", properties),
		slog.String("session_id", t.sessionID),
	)
}

// GetUsageStatistics returns aggregated usage statistics
func (t *Telemetry) GetUsageStatistics() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	sessionDuration := time.Since(t.startTime)

	return map[string]interface{}{
		"session_id":        t.sessionID,
		"session_duration":  sessionDuration.String(),
		"total_events":      len(t.events),
		"keyboard_events":   len(t.keyboardEvents),
		"api_calls":         len(t.apiCalls),
		"modal_actions":     len(t.modalActions),
		"errors":            len(t.errors),
		"start_time":        t.startTime,
		"most_used_keys":    t.getTopItems(t.keyboardEvents, 5),
		"most_used_apis":    t.getTopItems(t.apiCalls, 5),
		"most_used_modals":  t.getTopItems(t.modalActions, 5),
		"common_errors":     t.getTopItems(t.errors, 5),
	}
}

// getTopItems returns the top N items from a frequency map
func (t *Telemetry) getTopItems(items map[string]int64, n int) []map[string]interface{} {
	type item struct {
		key   string
		count int64
	}

	// Convert map to slice for sorting
	itemSlice := make([]item, 0, len(items))
	for k, v := range items {
		itemSlice = append(itemSlice, item{key: k, count: v})
	}

	// Simple bubble sort for top N (good enough for small datasets)
	for i := 0; i < len(itemSlice)-1; i++ {
		for j := 0; j < len(itemSlice)-i-1; j++ {
			if itemSlice[j].count < itemSlice[j+1].count {
				itemSlice[j], itemSlice[j+1] = itemSlice[j+1], itemSlice[j]
			}
		}
	}

	// Take top N
	if n > len(itemSlice) {
		n = len(itemSlice)
	}

	result := make([]map[string]interface{}, n)
	for i := 0; i < n; i++ {
		result[i] = map[string]interface{}{
			"item":  itemSlice[i].key,
			"count": itemSlice[i].count,
		}
	}

	return result
}

// LogSessionSummary logs a comprehensive session summary
func (t *Telemetry) LogSessionSummary() {
	stats := t.GetUsageStatistics()
	t.logger.Info("session_summary",
		"session_id", stats["session_id"],
		"session_duration", stats["session_duration"],
		"total_events", stats["total_events"],
		"keyboard_events", stats["keyboard_events"],
		"api_calls", stats["api_calls"],
		"modal_actions", stats["modal_actions"],
		"errors", stats["errors"],
		"most_used_keys", stats["most_used_keys"],
		"most_used_apis", stats["most_used_apis"],
	)
}

// GetEvents returns all tracked events (for export or analysis)
func (t *Telemetry) GetEvents() []UsageEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Return a copy to prevent race conditions
	events := make([]UsageEvent, len(t.events))
	copy(events, t.events)
	return events
}

// ClearEvents clears all tracked events (useful for memory management)
func (t *Telemetry) ClearEvents() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.events = make([]UsageEvent, 0)
}