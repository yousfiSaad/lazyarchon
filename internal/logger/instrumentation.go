package logger

import (
	"time"
)

// Timer tracks the duration of an operation
type Timer struct {
	name      string
	startTime time.Time
	metrics   *Metrics
}

// StartTimer creates and starts a new timer for an operation
func (m *Metrics) StartTimer(operationName string) *Timer {
	return &Timer{
		name:      operationName,
		startTime: time.Now(),
		metrics:   m,
	}
}

// Stop stops the timer and records the operation metrics
func (t *Timer) Stop(success bool) time.Duration {
	duration := time.Since(t.startTime)
	if t.metrics != nil {
		t.metrics.TrackOperation(t.name, duration, success)
	}
	return duration
}

// StopWithError stops the timer and records an error
func (t *Timer) StopWithError(err error) time.Duration {
	return t.Stop(err == nil)
}

// Instrument is a helper function to instrument a function call
func (m *Metrics) Instrument(operationName string, fn func() error) error {
	timer := m.StartTimer(operationName)
	err := fn()
	timer.StopWithError(err)
	return err
}

// InstrumentWithResult instruments a function that returns a value and error
func (m *Metrics) InstrumentWithResult(operationName string, fn func() (interface{}, error)) (interface{}, error) {
	timer := m.StartTimer(operationName)
	result, err := fn()
	timer.StopWithError(err)
	return result, err
}

// InstrumentAsync instruments an async operation
func (m *Metrics) InstrumentAsync(operationName string, fn func() error) {
	go func() {
		timer := m.StartTimer(operationName)
		err := fn()
		timer.StopWithError(err)
	}()
}