package ui

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/yousfisaad/lazyarchon/internal/archon"
)

// TestLargeDatasetPerformance tests performance with large numbers of tasks
func TestLargeDatasetPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Create model with proper window dimensions
	model := NewModel(createTestConfig())
	model.Window.width = 120
	model.Window.height = 40

	// Test with progressively larger datasets
	testSizes := []int{100, 500, 1000, 5000}

	for _, size := range testSizes {
		t.Run(fmt.Sprintf("TaskCount_%d", size), func(t *testing.T) {
			// Generate large task dataset
			tasks := generateLargeTasks(size)

			// Measure memory before
			var memBefore runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&memBefore)

			// Measure time for UpdateTasks
			start := time.Now()
			model.UpdateTasks(tasks)
			updateDuration := time.Since(start)

			// Verify the update worked
			if len(model.GetTasks()) != size {
				t.Errorf("Expected %d tasks, got %d", size, len(model.GetTasks()))
			}

			// Measure memory after
			var memAfter runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&memAfter)

			// Memory usage (in MB) - handle potential underflow
			var memUsed float64
			if memAfter.Alloc > memBefore.Alloc {
				memUsed = float64(memAfter.Alloc-memBefore.Alloc) / 1024 / 1024
			} else {
				memUsed = 0 // Memory was reclaimed
			}

			// Performance thresholds
			maxUpdateTime := 100 * time.Millisecond
			maxMemoryMB := 50.0

			if updateDuration > maxUpdateTime {
				t.Errorf("UpdateTasks took %v, expected < %v for %d tasks", updateDuration, maxUpdateTime, size)
			}

			if memUsed > maxMemoryMB {
				t.Errorf("Memory usage %.2f MB, expected < %.2f MB for %d tasks", memUsed, maxMemoryMB, size)
			}

			t.Logf("TaskCount: %d, UpdateTime: %v, MemoryUsed: %.2f MB", size, updateDuration, memUsed)
		})
	}
}

// TestSortingPerformance tests sorting performance with large datasets
func TestSortingPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	model := NewModel(createTestConfig())
	model.Window.width = 120
	model.Window.height = 40

	// Generate large task dataset with varied priorities and statuses
	tasks := generateLargeTasks(2000)
	model.UpdateTasks(tasks)

	// Test each sort mode performance
	sortModes := []string{"Status+Priority", "Priority", "Created Date", "Title"}

	for i, sortName := range sortModes {
		t.Run(fmt.Sprintf("SortMode_%s", sortName), func(t *testing.T) {
			// Set sort mode
			model.Data.sortMode = i

			// Measure sorting time
			start := time.Now()
			sortedTasks := model.GetSortedTasks()
			sortDuration := time.Since(start)

			// Verify sort worked
			if len(sortedTasks) != len(tasks) {
				t.Errorf("Expected %d sorted tasks, got %d", len(tasks), len(sortedTasks))
			}

			// Performance threshold - sorting should be fast
			maxSortTime := 50 * time.Millisecond
			if sortDuration > maxSortTime {
				t.Errorf("Sorting took %v, expected < %v for mode %s", sortDuration, maxSortTime, sortName)
			}

			t.Logf("SortMode: %s, SortTime: %v, TaskCount: %d", sortName, sortDuration, len(sortedTasks))
		})
	}
}

// TestNavigationPerformance tests navigation performance with large lists
func TestNavigationPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	model := NewModel(createTestConfig())
	model.Window.width = 120
	model.Window.height = 40

	// Generate large task dataset
	tasks := generateLargeTasks(3000)
	model.UpdateTasks(tasks)

	// Test navigation operations
	navOperations := []struct {
		name string
		key  string
	}{
		{"SingleDown", "j"},
		{"SingleUp", "k"},
		{"FastDown", "J"},
		{"FastUp", "K"},
		{"JumpToFirst", "gg"},
		{"JumpToLast", "G"},
	}

	for _, op := range navOperations {
		t.Run(op.name, func(t *testing.T) {
			// Start from middle of list
			model.Navigation.selectedIndex = len(tasks) / 2

			// Measure navigation time
			start := time.Now()
			newModel, _ := model.HandleKeyPress(op.key)
			navDuration := time.Since(start)

			// Verify navigation worked (index changed or reached boundary)
			if op.key == "gg" && newModel.Navigation.selectedIndex != 0 {
				t.Error("Expected gg to jump to first task")
			}
			if op.key == "G" {
				expectedLast := len(newModel.GetSortedTasks()) - 1
				if newModel.Navigation.selectedIndex != expectedLast {
					t.Errorf("Expected G to jump to last task (%d), got %d", expectedLast, newModel.Navigation.selectedIndex)
				}
			}

			// Performance threshold - navigation should be instant
			maxNavTime := 10 * time.Millisecond
			if navDuration > maxNavTime {
				t.Errorf("Navigation %s took %v, expected < %v", op.name, navDuration, maxNavTime)
			}

			t.Logf("Navigation: %s, Time: %v, TaskCount: %d", op.name, navDuration, len(tasks))
		})
	}
}

// TestSearchPerformance tests search performance with large datasets
func TestSearchPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	model := NewModel(createTestConfig())
	model.Window.width = 120
	model.Window.height = 40

	// Generate large task dataset with searchable content
	tasks := generateSearchableTasks(1500)
	model.UpdateTasks(tasks)

	searchQueries := []string{
		"task",     // Common term (many matches)
		"critical", // Medium term (some matches)
		"urgent",   // Rare term (few matches)
		"xyz",      // No matches
	}

	for _, query := range searchQueries {
		t.Run(fmt.Sprintf("Search_%s", query), func(t *testing.T) {
			// Set up search
			model.Data.searchQuery = query
			model.Data.searchActive = true

			// Measure search performance
			start := time.Now()
			model.updateSearchMatches()
			searchDuration := time.Since(start)

			// Verify search results
			if model.Data.totalMatches < 0 {
				t.Error("Expected non-negative match count")
			}

			// Performance threshold - search should be fast
			maxSearchTime := 30 * time.Millisecond
			if searchDuration > maxSearchTime {
				t.Errorf("Search for '%s' took %v, expected < %v", query, searchDuration, maxSearchTime)
			}

			t.Logf("Search: '%s', Time: %v, Matches: %d, TaskCount: %d",
				query, searchDuration, model.Data.totalMatches, len(tasks))
		})
	}
}

// TestViewportPerformance tests viewport rendering performance
func TestViewportPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	model := NewModel(createTestConfig())
	model.Window.width = 120
	model.Window.height = 40

	// Generate tasks with complex content for viewport rendering
	tasks := generateComplexTasks(100) // Fewer tasks but complex content
	model.UpdateTasks(tasks)

	// Test viewport update performance
	t.Run("ViewportUpdate", func(t *testing.T) {
		// Measure viewport update time
		start := time.Now()
		model.updateTaskDetailsViewport()
		viewportDuration := time.Since(start)

		// Performance threshold - viewport updates should be fast
		maxViewportTime := 20 * time.Millisecond
		if viewportDuration > maxViewportTime {
			t.Errorf("Viewport update took %v, expected < %v", viewportDuration, maxViewportTime)
		}

		t.Logf("ViewportUpdate: Time: %v", viewportDuration)
	})

	// Test help modal viewport performance
	t.Run("HelpModalUpdate", func(t *testing.T) {
		start := time.Now()
		model.updateHelpModalViewport()
		helpDuration := time.Since(start)

		maxHelpTime := 10 * time.Millisecond
		if helpDuration > maxHelpTime {
			t.Errorf("Help modal update took %v, expected < %v", helpDuration, maxHelpTime)
		}

		t.Logf("HelpModalUpdate: Time: %v", helpDuration)
	})
}

// TestMemoryLeaks tests for memory leaks with repeated operations
func TestMemoryLeaks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak test in short mode")
	}

	model := NewModel(createTestConfig())
	model.Window.width = 120
	model.Window.height = 40

	// Measure initial memory
	var memBefore runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	// Perform many operations that could leak memory
	for i := 0; i < 1000; i++ {
		// Generate new task set each time
		tasks := generateLargeTasks(100)
		model.UpdateTasks(tasks)

		// Perform various operations
		model.CycleSortMode()
		model.GetSortedTasks()
		model.updateTaskDetailsViewport()

		// Simulate navigation
		model.Navigation.selectedIndex = i % len(tasks)
	}

	// Force garbage collection and measure memory
	runtime.GC()
	runtime.GC() // Run twice to ensure cleanup
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	// Calculate memory growth - handle potential underflow
	var memGrowthMB float64
	if memAfter.Alloc > memBefore.Alloc {
		memGrowthMB = float64(memAfter.Alloc-memBefore.Alloc) / 1024 / 1024
	} else {
		memGrowthMB = 0 // Memory was reclaimed or stayed the same
	}

	// Threshold for acceptable memory growth
	maxGrowthMB := 10.0
	if memGrowthMB > maxGrowthMB {
		t.Errorf("Memory growth %.2f MB, expected < %.2f MB (possible memory leak)", memGrowthMB, maxGrowthMB)
	}

	t.Logf("MemoryLeak test: Growth: %.2f MB after 1000 operations", memGrowthMB)
}

// TestConcurrentAccess tests thread safety (basic)
func TestConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrency test in short mode")
	}

	model := NewModel(createTestConfig())
	model.Window.width = 120
	model.Window.height = 40

	// Generate initial dataset
	tasks := generateLargeTasks(500)
	model.UpdateTasks(tasks)

	// Test concurrent reads (should not crash)
	done := make(chan bool)
	numRoutines := 10

	for i := 0; i < numRoutines; i++ {
		go func(routineID int) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Goroutine %d panicked: %v", routineID, r)
				}
				done <- true
			}()

			// Perform read operations
			for j := 0; j < 100; j++ {
				_ = model.GetSortedTasks()
				_ = model.GetTasks()
				_ = model.GetProjects()
				_ = model.IsLoading()
			}
		}(i)
	}

	// Wait for all routines to complete with timeout
	timeout := time.After(5 * time.Second)
	completed := 0

	for completed < numRoutines {
		select {
		case <-done:
			completed++
		case <-timeout:
			t.Fatal("Concurrency test timed out")
		}
	}

	t.Logf("Concurrency test completed: %d routines finished", completed)
}

// Helper functions for generating test data

func generateLargeTasks(count int) []archon.Task {
	tasks := make([]archon.Task, count)
	statuses := []string{"todo", "doing", "review", "done"}

	for i := 0; i < count; i++ {
		tasks[i] = archon.Task{
			ID:        fmt.Sprintf("task-%d", i),
			Title:     fmt.Sprintf("Task %d", i),
			Status:    statuses[i%4],
			TaskOrder: i % 100, // Varied priorities
		}
	}
	return tasks
}

func generateSearchableTasks(count int) []archon.Task {
	tasks := make([]archon.Task, count)
	keywords := []string{"task", "critical", "urgent", "feature", "bug", "enhancement"}

	for i := 0; i < count; i++ {
		keyword := keywords[i%len(keywords)]
		tasks[i] = archon.Task{
			ID:    fmt.Sprintf("search-task-%d", i),
			Title: fmt.Sprintf("%s Task %d", keyword, i),
			Status: "todo",
			TaskOrder: i,
		}
	}
	return tasks
}

func generateComplexTasks(count int) []archon.Task {
	tasks := make([]archon.Task, count)

	for i := 0; i < count; i++ {
		// Create tasks with complex descriptions for viewport testing
		description := fmt.Sprintf("Complex task %d with detailed description.\n\nThis task has multiple paragraphs and complex formatting:\n\n**Key Points:**\n- Point 1: Implementation details\n- Point 2: Technical specifications\n- Point 3: Testing requirements\n\n**Code Examples:**\nfunc example() {\n    fmt.Println(\"Complex content\")\n}\n\n**Additional Information:**\nThis is a longer paragraph that contains detailed information about the task requirements, implementation strategy, testing approach, and deployment considerations. It should be long enough to test word wrapping and viewport scrolling performance.", i)

		tasks[i] = archon.Task{
			ID:          fmt.Sprintf("complex-task-%d", i),
			Title:       fmt.Sprintf("Complex Task %d with Long Title That Tests Wrapping", i),
			Description: description,
			Status:      "todo",
			TaskOrder:   i,
		}
	}
	return tasks
}