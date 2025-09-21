package logger

import (
	"errors"
	"time"

	"github.com/yousfisaad/lazyarchon/internal/config"
)

// ExampleUsage demonstrates how to use the structured logging system
func ExampleUsage() {
	// 1. Create a config (normally loaded from file)
	cfg := &config.Config{
		Development: config.DevelopmentConfig{
			Debug:           true,
			LogLevel:        "debug",
			EnableProfiling: true,
		},
		Profile: "development",
	}

	// 2. Create structured logger
	logger := New(cfg)

	// 3. Create metrics tracker
	metrics := NewMetrics(logger)

	// 4. Create debug mode
	debugMode := NewDebugMode(logger, cfg.Development.Debug)

	// 5. Create telemetry tracker
	telemetry := NewTelemetry(logger, "session-123")

	// Example: Basic logging
	logger.Info("Application started", "version", "1.0.0")
	logger.Debug("main", "Loading configuration", "config_path", "/etc/app/config.yaml")

	// Example: Performance tracking
	timer := metrics.StartTimer("api_call")
	time.Sleep(50 * time.Millisecond) // Simulate work
	timer.Stop(true)

	// Example: Error tracking with context
	err := errors.New("connection failed")
	logger.Error("Database connection error", err, "host", "localhost", "port", 5432)

	// Example: Debug mode with function tracking
	defer debugMode.LogFunctionCall("ExampleUsage")()

	// Example: State change tracking
	debugMode.LogStateChange("ui", "activeModal", "none", "help")

	// Example: API call debugging
	debugMode.LogAPICall("GET", "/api/tasks", 200, 45*time.Millisecond, map[string]interface{}{"count": 10})

	// Example: Memory usage tracking
	debugMode.LogMemoryUsage("startup")

	// Example: Telemetry tracking
	telemetry.TrackKeyboardInput("ctrl+r", "main")
	telemetry.TrackModalAction("help", "open", map[string]interface{}{"trigger": "keyboard"})
	telemetry.TrackAPICall("/api/tasks", "GET", 200, 45*time.Millisecond)

	// Example: User journey tracking
	telemetry.TrackUserJourney("app_startup", map[string]interface{}{
		"config_source": "file",
		"debug_mode":    true,
	})

	// Example: Performance instrumentation
	err = metrics.Instrument("load_tasks", func() error {
		time.Sleep(30 * time.Millisecond) // Simulate work
		return nil
	})

	if err != nil {
		logger.Error("Task loading failed", err)
	}

	// Example: Get usage statistics
	stats := telemetry.GetUsageStatistics()
	logger.Info("Session statistics", "stats", stats)

	// Example: Log performance summary
	metrics.LogSummary()

	// Example: Log session summary
	telemetry.LogSessionSummary()
}