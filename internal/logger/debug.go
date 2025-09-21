package logger

import (
	"fmt"
	"log/slog"
	"runtime"
	"time"
)

// DebugMode enables comprehensive debug logging with stack traces and detailed context
type DebugMode struct {
	logger   *Logger
	enabled  bool
	startTime time.Time
}

// NewDebugMode creates a new debug mode instance
func NewDebugMode(logger *Logger, enabled bool) *DebugMode {
	return &DebugMode{
		logger:    logger,
		enabled:   enabled,
		startTime: time.Now(),
	}
}

// IsEnabled returns whether debug mode is active
func (d *DebugMode) IsEnabled() bool {
	return d.enabled
}

// LogFunctionCall logs function entry and exit with timing
func (d *DebugMode) LogFunctionCall(functionName string, args ...interface{}) func() {
	if !d.enabled {
		return func() {}
	}

	startTime := time.Now()
	d.logger.Debug("function_entry", "entering function",
		slog.String("function", functionName),
		slog.Any("args", args),
		slog.String("caller", d.getCaller(2)),
	)

	return func() {
		duration := time.Since(startTime)
		d.logger.Debug("function_exit", "exiting function",
			slog.String("function", functionName),
			slog.Duration("duration", duration),
		)
	}
}

// LogStateChange logs application state changes with before/after values
func (d *DebugMode) LogStateChange(component, field string, before, after interface{}) {
	if !d.enabled {
		return
	}

	d.logger.Debug("state_change", "application state changed",
		slog.String("component", component),
		slog.String("field", field),
		slog.Any("before", before),
		slog.Any("after", after),
		slog.String("caller", d.getCaller(2)),
	)
}

// LogKeyboardInput logs keyboard input events
func (d *DebugMode) LogKeyboardInput(key string, mode string) {
	if !d.enabled {
		return
	}

	d.logger.Debug("keyboard_input", "keyboard input received",
		slog.String("key", key),
		slog.String("mode", mode),
		slog.Time("timestamp", time.Now()),
	)
}

// LogAPICall logs API calls with request/response details
func (d *DebugMode) LogAPICall(method, endpoint string, statusCode int, duration time.Duration, body interface{}) {
	if !d.enabled {
		return
	}

	level := slog.LevelDebug
	if statusCode >= 400 {
		level = slog.LevelWarn
	}

	d.logger.Logger.Log(nil, level, "api_call",
		slog.String("method", method),
		slog.String("endpoint", endpoint),
		slog.Int("status_code", statusCode),
		slog.Duration("duration", duration),
		slog.Any("response_body", body),
	)
}

// LogMemoryUsage logs current memory usage statistics
func (d *DebugMode) LogMemoryUsage(context string) {
	if !d.enabled {
		return
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	d.logger.Debug("memory_usage", "memory usage snapshot",
		slog.String("context", context),
		slog.Uint64("alloc_mb", bToMb(m.Alloc)),
		slog.Uint64("total_alloc_mb", bToMb(m.TotalAlloc)),
		slog.Uint64("sys_mb", bToMb(m.Sys)),
		slog.Uint64("num_gc", uint64(m.NumGC)),
		slog.Uint64("goroutines", uint64(runtime.NumGoroutine())),
	)
}

// LogPanic logs panic recovery with stack trace
func (d *DebugMode) LogPanic(recovered interface{}, stackTrace []byte) {
	d.logger.Logger.Error("panic_recovered",
		slog.Any("panic_value", recovered),
		slog.String("stack_trace", string(stackTrace)),
		slog.String("caller", d.getCaller(2)),
	)
}

// LogUptime logs application uptime and basic statistics
func (d *DebugMode) LogUptime() {
	if !d.enabled {
		return
	}

	uptime := time.Since(d.startTime)
	d.logger.Debug("uptime", "application uptime",
		slog.Duration("uptime", uptime),
		slog.String("uptime_string", uptime.String()),
		slog.Time("start_time", d.startTime),
		slog.Uint64("goroutines", uint64(runtime.NumGoroutine())),
	)
}

// Trace logs detailed execution trace information
func (d *DebugMode) Trace(component, operation string, details map[string]interface{}) {
	if !d.enabled {
		return
	}

	attrs := []slog.Attr{
		slog.String("component", component),
		slog.String("operation", operation),
		slog.String("caller", d.getCaller(2)),
	}

	for key, value := range details {
		attrs = append(attrs, slog.Any(key, value))
	}

	d.logger.Logger.LogAttrs(nil, slog.LevelDebug, "trace", attrs...)
}

// getCaller returns the caller information
func (d *DebugMode) getCaller(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	return fmt.Sprintf("%s:%d", file, line)
}

// bToMb converts bytes to megabytes
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// WrapWithDebug wraps a function with debug logging
func (d *DebugMode) WrapWithDebug(functionName string, fn func() error) func() error {
	return func() error {
		defer d.LogFunctionCall(functionName)()
		return fn()
	}
}

// WrapWithDebugAndResult wraps a function that returns a value with debug logging
func (d *DebugMode) WrapWithDebugAndResult(functionName string, fn func() (interface{}, error)) func() (interface{}, error) {
	return func() (interface{}, error) {
		defer d.LogFunctionCall(functionName)()
		return fn()
	}
}