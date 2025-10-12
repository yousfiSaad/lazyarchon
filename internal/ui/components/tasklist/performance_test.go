package tasklist

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/yousfisaad/lazyarchon/v2/internal/archon"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/config"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/styling"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui/components/base"
)

// BenchmarkTaskListRendering tests TaskList component rendering performance
func BenchmarkTaskListRendering(b *testing.B) {
	taskCounts := []int{100, 500, 1000, 2000, 5000}

	for _, count := range taskCounts {
		b.Run(fmt.Sprintf("TaskCount_%d", count), func(b *testing.B) {
			// Create test data
			tasks := generateTestTasks(count)

			// Create component
			model := createBenchmarkModel(tasks, 40, 20)

			// Benchmark rendering
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = model.View()
			}
		})
	}
}

// BenchmarkTaskListScrolling tests scrolling performance with large datasets
func BenchmarkTaskListScrolling(b *testing.B) {
	tasks := generateTestTasks(2000)
	model := createBenchmarkModel(tasks, 40, 20)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate scrolling through the entire list
		for index := 0; index < len(tasks); index += 10 {
			model.selectedIndex = index
			_ = model.View()
		}
	}
}

// BenchmarkTaskListFiltering tests filtering performance
func BenchmarkTaskListFiltering(b *testing.B) {
	tasks := generateTestTasks(1000)
	model := createBenchmarkModel(tasks, 40, 20)

	testQueries := []string{"test", "important", "bug", "feature", "urgent"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, query := range testQueries {
			model.searchQuery = query
			model.searchActive = true
			// Note: No updateSortedTasks() call - tasks are queried from parent on-demand
			_ = model.View()
		}
	}
}

// BenchmarkTaskListMemoryUsage measures memory consumption with large datasets
func BenchmarkTaskListMemoryUsage(b *testing.B) {
	taskCounts := []int{1000, 5000, 10000}

	for _, count := range taskCounts {
		b.Run(fmt.Sprintf("TaskCount_%d", count), func(b *testing.B) {
			var m1, m2 runtime.MemStats //nolint:varnamelen // m1, m2 are idiomatic for memory stats comparison

			// Measure before
			runtime.GC()
			runtime.ReadMemStats(&m1)

			// Create component with large dataset
			tasks := generateTestTasks(count)
			model := createBenchmarkModel(tasks, 40, 20)

			// Render multiple times to simulate real usage
			for i := 0; i < 100; i++ {
				_ = model.View()
			}

			// Measure after
			runtime.GC()
			runtime.ReadMemStats(&m2)

			// Report memory usage
			memUsed := float64(m2.Alloc-m1.Alloc) / 1024 / 1024 // MB
			b.ReportMetric(memUsed, "MB")

			b.Logf("TaskCount: %d, MemoryUsed: %.2f MB", count, memUsed)
		})
	}
}

// BenchmarkViewportCalculation tests scroll window calculation performance
func BenchmarkViewportCalculation(b *testing.B) {
	taskCounts := []int{1000, 5000, 10000}

	for _, count := range taskCounts {
		b.Run(fmt.Sprintf("TaskCount_%d", count), func(b *testing.B) {
			tasks := generateTestTasks(count)
			model := createBenchmarkModel(tasks, 40, 20)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				// Test various scroll positions
				for selectedIndex := 0; selectedIndex < count; selectedIndex += 100 {
					model.selectedIndex = selectedIndex
					// This triggers viewport calculation in View()
					_ = model.View()
				}
			}
		})
	}
}

// BenchmarkSearchHighlighting tests search highlighting performance
func BenchmarkSearchHighlighting(b *testing.B) {
	tasks := generateTestTasks(1000)
	model := createBenchmarkModel(tasks, 40, 20)
	model.searchQuery = "test"
	model.searchActive = true

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = model.View()
	}
}

// Helper functions

func generateTestTasks(count int) []archon.Task {
	tasks := make([]archon.Task, count)
	statuses := []string{"todo", "doing", "review", "done"}
	features := []string{"auth", "api", "ui", "testing", "deployment"}

	for i := 0; i < count; i++ { //nolint:varnamelen // i is idiomatic for loop index
		tasks[i] = archon.Task{
			ID:          fmt.Sprintf("task-%d", i),
			Title:       fmt.Sprintf("Test Task %d - Important work item", i),
			Description: fmt.Sprintf("This is a detailed description for test task %d with various keywords like urgent, important, bug, feature, and test", i),
			Status:      statuses[i%len(statuses)],
			Feature:     &features[i%len(features)],
			TaskOrder:   i % 10,
			CreatedAt:   archon.FlexibleTime{Time: time.Now().Add(-time.Duration(i) * time.Hour)},
		}
	}

	return tasks
}

//nolint:unparam // width always 40 in current tests - parameter kept for future flexibility
func createBenchmarkModel(tasks []archon.Task, width, height int) TaskListModel {
	// Create ComponentContext with providers and mock callback
	context := &base.ComponentContext{
		ConfigProvider:       &fallbackConfigProvider{},
		StyleContextProvider: &benchmarkStyleProvider{},
	}

	// Provide mock callback that returns tasks for benchmarking
	context.GetSortedTasks = func() []interface{} {
		result := make([]interface{}, len(tasks))
		for i := range tasks {
			result[i] = tasks[i]
		}
		return result
	}

	opts := Options{
		Width:         width,
		Height:        height,
		SelectedIndex: 0,
		SearchQuery:   "",
		SearchActive:  false,
		Context:       context,
	}

	return NewModel(opts)
}

// fallbackConfigProvider provides minimal configuration for benchmarks
type fallbackConfigProvider struct{}

func (f *fallbackConfigProvider) GetServerURL() string { return "http://localhost:8181" }
func (f *fallbackConfigProvider) GetAPIKey() string    { return "" }
func (f *fallbackConfigProvider) GetTheme() *config.ThemeConfig {
	return &config.ThemeConfig{Name: "default"}
}
func (f *fallbackConfigProvider) GetDisplay() *config.DisplayConfig { return &config.DisplayConfig{} }
func (f *fallbackConfigProvider) GetDevelopment() *config.DevelopmentConfig {
	return &config.DevelopmentConfig{}
}
func (f *fallbackConfigProvider) GetDefaultSortMode() string        { return "status+priority" }
func (f *fallbackConfigProvider) IsDebugEnabled() bool              { return false }
func (f *fallbackConfigProvider) IsDarkModeEnabled() bool           { return false }
func (f *fallbackConfigProvider) IsCompletedTasksVisible() bool     { return true }
func (f *fallbackConfigProvider) IsPriorityIndicatorsEnabled() bool { return false }
func (f *fallbackConfigProvider) IsFeatureColorsEnabled() bool      { return false }
func (f *fallbackConfigProvider) IsFeatureBackgroundsEnabled() bool { return false }

// benchmarkStyleProvider provides minimal style context for benchmarks
type benchmarkStyleProvider struct{}

func (b *benchmarkStyleProvider) CreateStyleContext(isDarkMode bool) *styling.StyleContext {
	// Create a minimal theme adapter for benchmarks
	theme := &styling.ThemeAdapter{
		TodoColor:   "240",
		DoingColor:  "33",
		ReviewColor: "214",
		DoneColor:   "82",
		HeaderColor: "39",
		MutedColor:  "240",
		AccentColor: "51",
		StatusColor: "205",
		Name:        "default",
	}

	return styling.NewStyleContext(theme, b)
}

func (b *benchmarkStyleProvider) IsPriorityIndicatorsEnabled() bool { return false }
func (b *benchmarkStyleProvider) IsFeatureColorsEnabled() bool      { return false }
func (b *benchmarkStyleProvider) GetTheme() *config.ThemeConfig {
	return &config.ThemeConfig{Name: "default"}
}
