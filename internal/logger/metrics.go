package logger

import (
	"log/slog"
	"sync"
	"time"
)

// Metrics tracks application performance and usage metrics
type Metrics struct {
	mu             sync.RWMutex
	operations     map[string]*OperationMetric
	totalRequests  int64
	errorCount     int64
	startTime      time.Time
	logger         *Logger
}

// OperationMetric tracks metrics for a specific operation
type OperationMetric struct {
	Name        string
	Count       int64
	TotalTime   time.Duration
	MinTime     time.Duration
	MaxTime     time.Duration
	ErrorCount  int64
	LastExecuted time.Time
}

// NewMetrics creates a new metrics tracker
func NewMetrics(logger *Logger) *Metrics {
	return &Metrics{
		operations: make(map[string]*OperationMetric),
		startTime:  time.Now(),
		logger:     logger,
	}
}

// TrackOperation tracks the execution of an operation
func (m *Metrics) TrackOperation(name string, duration time.Duration, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	metric, exists := m.operations[name]
	if !exists {
		metric = &OperationMetric{
			Name:     name,
			MinTime:  duration,
			MaxTime:  duration,
		}
		m.operations[name] = metric
	}

	metric.Count++
	metric.TotalTime += duration
	metric.LastExecuted = time.Now()

	if duration < metric.MinTime {
		metric.MinTime = duration
	}
	if duration > metric.MaxTime {
		metric.MaxTime = duration
	}

	if !success {
		metric.ErrorCount++
		m.errorCount++
	}

	m.totalRequests++

	// Log performance if enabled
	if m.logger != nil {
		m.logger.Performance(name, duration,
			slog.String("status", func() string {
				if success {
					return "success"
				}
				return "error"
			}()),
			slog.Int64("total_count", metric.Count),
			slog.Int64("error_count", metric.ErrorCount),
		)
	}
}

// GetOperationMetrics returns metrics for a specific operation
func (m *Metrics) GetOperationMetrics(name string) *OperationMetric {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if metric, exists := m.operations[name]; exists {
		// Return a copy to avoid race conditions
		return &OperationMetric{
			Name:        metric.Name,
			Count:       metric.Count,
			TotalTime:   metric.TotalTime,
			MinTime:     metric.MinTime,
			MaxTime:     metric.MaxTime,
			ErrorCount:  metric.ErrorCount,
			LastExecuted: metric.LastExecuted,
		}
	}
	return nil
}

// GetAllMetrics returns all operation metrics
func (m *Metrics) GetAllMetrics() map[string]*OperationMetric {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*OperationMetric)
	for name, metric := range m.operations {
		result[name] = &OperationMetric{
			Name:        metric.Name,
			Count:       metric.Count,
			TotalTime:   metric.TotalTime,
			MinTime:     metric.MinTime,
			MaxTime:     metric.MaxTime,
			ErrorCount:  metric.ErrorCount,
			LastExecuted: metric.LastExecuted,
		}
	}
	return result
}

// GetSummary returns overall application metrics
func (m *Metrics) GetSummary() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	uptime := time.Since(m.startTime)

	return map[string]interface{}{
		"uptime":           uptime.String(),
		"total_requests":   m.totalRequests,
		"total_errors":     m.errorCount,
		"error_rate":       func() float64 {
			if m.totalRequests == 0 {
				return 0
			}
			return float64(m.errorCount) / float64(m.totalRequests)
		}(),
		"operations_count": len(m.operations),
		"start_time":       m.startTime,
	}
}

// LogSummary logs a summary of all metrics
func (m *Metrics) LogSummary() {
	if m.logger == nil {
		return
	}

	summary := m.GetSummary()
	m.logger.Info("metrics_summary",
		"uptime", summary["uptime"],
		"total_requests", summary["total_requests"],
		"total_errors", summary["total_errors"],
		"error_rate", summary["error_rate"],
		"operations_count", summary["operations_count"],
	)

	// Log top operations by count
	allMetrics := m.GetAllMetrics()
	for name, metric := range allMetrics {
		if metric.Count > 0 {
			avgTime := metric.TotalTime / time.Duration(metric.Count)
			m.logger.Info("operation_summary",
				"operation", name,
				"count", metric.Count,
				"avg_duration", avgTime.String(),
				"min_duration", metric.MinTime.String(),
				"max_duration", metric.MaxTime.String(),
				"error_count", metric.ErrorCount,
				"last_executed", metric.LastExecuted,
			)
		}
	}
}