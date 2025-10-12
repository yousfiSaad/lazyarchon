package logging

import (
	"log/slog"
	"os"
	"runtime"
	"time"
)

const (
	// DefaultLogFile is the default log file path for TUI applications
	DefaultLogFile = "/tmp/lazyarchon.log"
)

// SlogLogger provides structured logging using slog without dependency injection
type SlogLogger struct {
	logger         *slog.Logger
	debugEnabled   bool
	profilingEnabled bool
	logFile        *os.File
}

// NewSlogLogger creates a new slog logger with file output
func NewSlogLogger(debugEnabled bool) *SlogLogger {
	// Check for custom log file from environment
	logFilePath := os.Getenv("LAZYARCHON_LOG_FILE")
	if logFilePath == "" {
		logFilePath = DefaultLogFile
	}

	// For TUI applications, write logs to a file instead of stderr to avoid interfering with the UI
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		// Fallback to discard if we can't open the log file
		logFile = nil
	}

	var handler slog.Handler
	if logFile != nil {
		handler = slog.NewTextHandler(logFile, &slog.HandlerOptions{
			Level: func() slog.Level {
				if debugEnabled {
					return slog.LevelDebug
				}
				return slog.LevelInfo
			}(),
			AddSource: debugEnabled, // Add source info only in debug mode
		})
	} else {
		// If we can't open log file, create a no-op handler
		handler = slog.NewTextHandler(os.NewFile(0, os.DevNull), &slog.HandlerOptions{
			Level: slog.LevelError, // Only log errors if we can't write to file
		})
	}

	return &SlogLogger{
		logger:         slog.New(handler),
		debugEnabled:   debugEnabled,
		profilingEnabled: debugEnabled, // Enable profiling when debug is on
		logFile:        logFile,
	}
}

// Debug logs a debug message with key-value pairs
func (l *SlogLogger) Debug(msg string, args ...interface{}) {
	if l.debugEnabled {
		attrs := l.convertArgsToAttrs(args...)
		l.logger.LogAttrs(nil, slog.LevelDebug, msg, attrs...)
	}
}

// Info logs an info message with key-value pairs
func (l *SlogLogger) Info(msg string, args ...interface{}) {
	attrs := l.convertArgsToAttrs(args...)
	l.logger.LogAttrs(nil, slog.LevelInfo, msg, attrs...)
}

// Warn logs a warning message with key-value pairs
func (l *SlogLogger) Warn(msg string, args ...interface{}) {
	attrs := l.convertArgsToAttrs(args...)
	l.logger.LogAttrs(nil, slog.LevelWarn, msg, attrs...)
}

// Error logs an error message with key-value pairs
func (l *SlogLogger) Error(msg string, args ...interface{}) {
	attrs := l.convertArgsToAttrs(args...)
	l.logger.LogAttrs(nil, slog.LevelError, msg, attrs...)
}

// Fatal logs a fatal message and exits the application
func (l *SlogLogger) Fatal(msg string, args ...interface{}) {
	attrs := l.convertArgsToAttrs(args...)
	l.logger.LogAttrs(nil, slog.LevelError, msg, attrs...)
	os.Exit(1)
}

// convertArgsToAttrs converts interface{} args to slog.Attr
func (l *SlogLogger) convertArgsToAttrs(args ...interface{}) []slog.Attr {
	var attrs []slog.Attr

	// Convert pairs of args to key-value attributes
	for i := 0; i < len(args)-1; i += 2 {
		if key, ok := args[i].(string); ok {
			attrs = append(attrs, slog.Any(key, args[i+1]))
		}
	}

	// If there's an odd number of args, add the last one as "extra"
	if len(args)%2 == 1 {
		attrs = append(attrs, slog.Any("extra", args[len(args)-1]))
	}

	return attrs
}

// =============================================================================
// PERFORMANCE PROFILING METHODS
// =============================================================================

// LogPerformance logs the duration of an operation with optional memory stats
func (l *SlogLogger) LogPerformance(operation string, startTime time.Time, args ...interface{}) {
	if !l.profilingEnabled {
		return
	}

	duration := time.Since(startTime)
	attrs := l.convertArgsToAttrs(args...)
	attrs = append(attrs, slog.Duration("duration", duration))
	attrs = append(attrs, slog.String("operation", operation))

	l.logger.LogAttrs(nil, slog.LevelDebug, "Performance", attrs...)
}

// LogPerformanceWithMemory logs performance with memory stats
func (l *SlogLogger) LogPerformanceWithMemory(operation string, startTime time.Time, args ...interface{}) {
	if !l.profilingEnabled {
		return
	}

	duration := time.Since(startTime)
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	attrs := l.convertArgsToAttrs(args...)
	attrs = append(attrs,
		slog.Duration("duration", duration),
		slog.String("operation", operation),
		slog.Uint64("alloc_mb", m.Alloc/1024/1024),
		slog.Uint64("sys_mb", m.Sys/1024/1024),
		slog.Uint64("num_gc", uint64(m.NumGC)),
	)

	l.logger.LogAttrs(nil, slog.LevelDebug, "Performance (with memory)", attrs...)
}

// StartOperation returns a function to log the operation duration when called
func (l *SlogLogger) StartOperation(operation string) func(...interface{}) {
	if !l.profilingEnabled {
		return func(...interface{}) {} // No-op if profiling disabled
	}

	startTime := time.Now()
	return func(args ...interface{}) {
		l.LogPerformance(operation, startTime, args...)
	}
}

// =============================================================================
// STRUCTURED LOGGING HELPERS
// =============================================================================

// LogHTTPRequest logs an HTTP request with details
func (l *SlogLogger) LogHTTPRequest(method, url string, args ...interface{}) {
	if !l.debugEnabled {
		return
	}

	attrs := l.convertArgsToAttrs(args...)
	attrs = append(attrs,
		slog.String("method", method),
		slog.String("url", url),
	)

	l.logger.LogAttrs(nil, slog.LevelDebug, "HTTP Request", attrs...)
}

// LogHTTPResponse logs an HTTP response with details
func (l *SlogLogger) LogHTTPResponse(method, url string, statusCode int, duration time.Duration, args ...interface{}) {
	attrs := l.convertArgsToAttrs(args...)
	attrs = append(attrs,
		slog.String("method", method),
		slog.String("url", url),
		slog.Int("status", statusCode),
		slog.Duration("duration", duration),
	)

	level := slog.LevelInfo
	if statusCode >= 400 {
		level = slog.LevelWarn
	}
	if statusCode >= 500 {
		level = slog.LevelError
	}

	l.logger.LogAttrs(nil, level, "HTTP Response", attrs...)
}

// LogStateChange logs a state transition
func (l *SlogLogger) LogStateChange(component, field string, oldValue, newValue interface{}, args ...interface{}) {
	if !l.debugEnabled {
		return
	}

	attrs := l.convertArgsToAttrs(args...)
	attrs = append(attrs,
		slog.String("component", component),
		slog.String("field", field),
		slog.Any("old_value", oldValue),
		slog.Any("new_value", newValue),
	)

	l.logger.LogAttrs(nil, slog.LevelDebug, "State Change", attrs...)
}

// LogWebSocketEvent logs WebSocket connection events
func (l *SlogLogger) LogWebSocketEvent(event string, args ...interface{}) {
	attrs := l.convertArgsToAttrs(args...)
	attrs = append(attrs, slog.String("event", event))

	l.logger.LogAttrs(nil, slog.LevelInfo, "WebSocket Event", attrs...)
}

// LogComponentUpdate logs component lifecycle events
func (l *SlogLogger) LogComponentUpdate(component, action string, args ...interface{}) {
	if !l.debugEnabled {
		return
	}

	attrs := l.convertArgsToAttrs(args...)
	attrs = append(attrs,
		slog.String("component", component),
		slog.String("action", action),
	)

	l.logger.LogAttrs(nil, slog.LevelDebug, "Component Update", attrs...)
}

// Close closes the log file if it's open
func (l *SlogLogger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}