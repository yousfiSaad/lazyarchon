package batch

import (
	"sync"
	"sync/atomic"
	"time"
)

// UpdateFunc represents a function that performs an expensive update operation
type UpdateFunc func()

// Manager provides generic update batching for smooth performance
// This system batches rapid update requests to prevent performance degradation
type Manager struct {
	mu sync.Mutex

	// Batching configuration
	delay    time.Duration // How long to wait before executing batched updates
	maxBatch int           // Maximum number of updates to batch together

	// Current batch state
	pendingFunc UpdateFunc  // The function to execute when batch triggers
	timer       *time.Timer // Timer for batch delay
	batchCount  int         // Number of updates in current batch

	// Performance metrics
	schedules  int64     // Total number of Schedule() calls
	executions int64     // Total number of actual executions
	batches    int64     // Total number of batches processed
	lastExec   time.Time // Last execution time
}

// NewManager creates a new batch update manager
// delay: how long to wait after the last schedule before executing
// maxBatch: maximum number of updates to batch (0 = unlimited)
func NewManager(delay time.Duration, maxBatch int) *Manager {
	return &Manager{
		delay:    delay,
		maxBatch: maxBatch,
	}
}

// Schedule requests that an update function be executed
// Multiple rapid calls will be batched together for performance
func (bm *Manager) Schedule(updateFunc UpdateFunc) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	atomic.AddInt64(&bm.schedules, 1)

	// Store the update function (latest one wins for simplicity)
	bm.pendingFunc = updateFunc
	bm.batchCount++

	// Check if we should execute immediately due to batch limit
	if bm.maxBatch > 0 && bm.batchCount >= bm.maxBatch {
		bm.executeNow()
		return
	}

	// Reset or start the timer for delayed execution
	if bm.timer != nil {
		bm.timer.Stop()
	}
	bm.timer = time.AfterFunc(bm.delay, func() {
		bm.executeBatch()
	})
}

// ScheduleImmediate bypasses batching and executes immediately
// Use this for critical updates that cannot be delayed
func (bm *Manager) ScheduleImmediate(updateFunc UpdateFunc) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	atomic.AddInt64(&bm.schedules, 1)

	// Cancel any pending batch
	if bm.timer != nil {
		bm.timer.Stop()
		bm.timer = nil
	}

	// Execute immediately
	bm.pendingFunc = updateFunc
	bm.batchCount++
	bm.executeNow()
}

// executeNow executes the pending function immediately (must hold lock)
func (bm *Manager) executeNow() {
	if bm.pendingFunc != nil {
		bm.pendingFunc()
		atomic.AddInt64(&bm.executions, 1)
		atomic.AddInt64(&bm.batches, 1)
		bm.lastExec = time.Now()

		// Reset batch state
		bm.pendingFunc = nil
		bm.batchCount = 0
	}

	// Clean up timer
	if bm.timer != nil {
		bm.timer.Stop()
		bm.timer = nil
	}
}

// executeBatch is called by the timer to execute batched updates
func (bm *Manager) executeBatch() {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.executeNow()
}

// Flush executes any pending updates immediately
// Use this when you need to ensure all updates are processed
func (bm *Manager) Flush() {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if bm.timer != nil {
		bm.timer.Stop()
	}
	bm.executeNow()
}

// Cancel cancels any pending updates without executing them
func (bm *Manager) Cancel() {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if bm.timer != nil {
		bm.timer.Stop()
		bm.timer = nil
	}

	bm.pendingFunc = nil
	bm.batchCount = 0
}

// GetStats returns batch manager performance statistics
func (bm *Manager) GetStats() BatchStats {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	return BatchStats{
		Schedules:    atomic.LoadInt64(&bm.schedules),
		Executions:   atomic.LoadInt64(&bm.executions),
		Batches:      atomic.LoadInt64(&bm.batches),
		Delay:        bm.delay,
		MaxBatch:     bm.maxBatch,
		PendingCount: bm.batchCount,
		LastExec:     bm.lastExec,
		HasPending:   bm.pendingFunc != nil,
	}
}

// SetDelay updates the batch delay duration
func (bm *Manager) SetDelay(delay time.Duration) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.delay = delay
}

// SetMaxBatch updates the maximum batch size
func (bm *Manager) SetMaxBatch(maxBatch int) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.maxBatch = maxBatch
}

// BatchStats provides batch manager performance metrics
type BatchStats struct {
	Schedules    int64         // Total number of Schedule() calls
	Executions   int64         // Total number of actual executions
	Batches      int64         // Total number of batches processed
	Delay        time.Duration // Current batch delay
	MaxBatch     int           // Current maximum batch size
	PendingCount int           // Number of updates in current batch
	LastExec     time.Time     // Last execution time
	HasPending   bool          // Whether there are pending updates
}

// GetBatchEfficiency returns the batching efficiency as a ratio
// Higher values indicate more effective batching (fewer executions per schedule)
func (bs *BatchStats) GetBatchEfficiency() float64 {
	if bs.Executions == 0 {
		return 0.0
	}
	return float64(bs.Schedules) / float64(bs.Executions)
}

// GetReductionRate returns the percentage of updates that were batched away
// Higher values indicate better performance improvement from batching
func (bs *BatchStats) GetReductionRate() float64 {
	if bs.Schedules == 0 {
		return 0.0
	}
	saved := bs.Schedules - bs.Executions
	return float64(saved) / float64(bs.Schedules) * 100.0
}

// GetAverageBatchSize returns the average number of updates per batch
func (bs *BatchStats) GetAverageBatchSize() float64 {
	if bs.Batches == 0 {
		return 0.0
	}
	return float64(bs.Schedules) / float64(bs.Batches)
}
