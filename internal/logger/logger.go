package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/yousfisaad/lazyarchon/internal/config"
)

var (
	defaultLogger *slog.Logger
)

// Logger wraps slog.Logger with application-specific functionality
type Logger struct {
	*slog.Logger
	config *config.Config
}

// New creates a new structured logger with the given configuration
func New(cfg *config.Config) *Logger {
	var handler slog.Handler
	var output io.Writer = os.Stdout

	// Configure output destination
	if cfg.Development.Debug {
		// In debug mode, log to both stdout and file
		logDir := filepath.Join(os.TempDir(), "lazyarchon")
		os.MkdirAll(logDir, 0755)

		logFile := filepath.Join(logDir, "debug.log")
		if file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			output = io.MultiWriter(os.Stdout, file)
		}
	}

	// Configure handler based on environment
	opts := &slog.HandlerOptions{
		Level: parseLogLevel(cfg.Development.LogLevel),
		AddSource: cfg.Development.Debug,
	}

	if cfg.IsDevelopmentProfile() {
		// Use text handler for development (more readable)
		handler = slog.NewTextHandler(output, opts)
	} else {
		// Use JSON handler for production (structured)
		handler = slog.NewJSONHandler(output, opts)
	}

	logger := slog.New(handler)

	return &Logger{
		Logger: logger,
		config: cfg,
	}
}

// parseLogLevel converts string log level to slog.Level
func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// SetDefault sets the global default logger
func SetDefault(logger *Logger) {
	defaultLogger = logger.Logger
	slog.SetDefault(defaultLogger)
}

// Default returns the default logger
func Default() *slog.Logger {
	if defaultLogger == nil {
		// Fallback to basic logger if not initialized
		return slog.Default()
	}
	return defaultLogger
}

// WithContext adds common application context to log entries
func (l *Logger) WithContext(ctx context.Context, component string) *slog.Logger {
	return l.With(
		"component", component,
		"timestamp", time.Now().UTC(),
	)
}

// Performance logs performance metrics
func (l *Logger) Performance(operation string, duration time.Duration, attrs ...slog.Attr) {
	if l.config.Development.EnableProfiling {
		args := []any{
			"operation", operation,
			"duration_ms", duration.Milliseconds(),
			"duration", duration.String(),
		}

		for _, attr := range attrs {
			args = append(args, attr.Key, attr.Value)
		}

		l.Info("performance", args...)
	}
}

// Usage logs usage patterns for telemetry
func (l *Logger) Usage(action string, attrs ...slog.Attr) {
	args := []any{
		"action", action,
		"timestamp", time.Now().UTC(),
	}

	for _, attr := range attrs {
		args = append(args, attr.Key, attr.Value)
	}

	l.Info("usage", args...)
}

// Error logs errors with context
func (l *Logger) Error(msg string, err error, attrs ...slog.Attr) {
	args := []any{
		"error", err.Error(),
	}

	for _, attr := range attrs {
		args = append(args, attr.Key, attr.Value)
	}

	l.Logger.Error(msg, args...)
}

// Debug logs debug information with component context
func (l *Logger) Debug(component, msg string, attrs ...slog.Attr) {
	if l.config.Development.Debug {
		args := []any{
			"component", component,
		}

		for _, attr := range attrs {
			args = append(args, attr.Key, attr.Value)
		}

		l.Logger.Debug(msg, args...)
	}
}

// Info logs informational messages
func (l *Logger) Info(msg string, attrs ...any) {
	l.Logger.Info(msg, attrs...)
}

// Warn logs warning messages
func (l *Logger) Warn(msg string, attrs ...any) {
	l.Logger.Warn(msg, attrs...)
}