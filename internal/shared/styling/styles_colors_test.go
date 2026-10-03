package styling

import (
	"sync"
	"testing"
	"time"
)

// TestGetFeatureColorConsistency tests that the same feature returns the same color
func TestGetFeatureColorConsistency(t *testing.T) {
	// Clear cache for testing
	cache.mu.Lock()
	cache.featureColor = make(map[string]string)
	cache.dimmedColor = make(map[string]string)
	cache.mu.Unlock()

	featureName := "authentication"

	// Get color multiple times
	color1 := GetFeatureColor(featureName)
	color2 := GetFeatureColor(featureName)
	color3 := GetFeatureColor(featureName)

	if color1 != color2 || color2 != color3 {
		t.Errorf("GetFeatureColor returned inconsistent colors: %s, %s, %s", color1, color2, color3)
	}
}

// TestGetFeatureColorCaching tests that caching works correctly
func TestGetFeatureColorCaching(t *testing.T) {
	// Clear cache for testing
	cache.mu.Lock()
	cache.featureColor = make(map[string]string)
	cache.dimmedColor = make(map[string]string)
	cache.mu.Unlock()

	featureName := "api-gateway"

	// First call should compute and cache
	start := time.Now()
	color1 := GetFeatureColor(featureName)
	firstCallDuration := time.Since(start)

	// Second call should use cache and be faster
	start = time.Now()
	color2 := GetFeatureColor(featureName)
	secondCallDuration := time.Since(start)

	if color1 != color2 {
		t.Errorf("Cached color mismatch: %s != %s", color1, color2)
	}

	// Verify cache is being used (second call should be significantly faster)
	// Note: This is a heuristic test - cache should be much faster than computation
	if secondCallDuration > firstCallDuration {
		t.Logf("Warning: Second call took longer than first. First: %v, Second: %v", firstCallDuration, secondCallDuration)
	}

	// Verify the value is actually cached
	cache.mu.RLock()
	cachedColor, exists := cache.featureColor[featureName]
	cache.mu.RUnlock()

	if !exists {
		t.Error("Feature color was not cached")
	}
	if cachedColor != color1 {
		t.Errorf("Cached color mismatch: expected %s, got %s", color1, cachedColor)
	}
}

// TestGetDimmedFeatureColorCaching tests dimmed color caching
func TestGetDimmedFeatureColorCaching(t *testing.T) {
	// Clear cache for testing
	cache.mu.Lock()
	cache.featureColor = make(map[string]string)
	cache.dimmedColor = make(map[string]string)
	cache.mu.Unlock()

	featureName := "database-layer"

	// First call should compute and cache
	color1 := GetDimmedFeatureColor(featureName)

	// Second call should use cache
	color2 := GetDimmedFeatureColor(featureName)

	if color1 != color2 {
		t.Errorf("Cached dimmed color mismatch: %s != %s", color1, color2)
	}

	// Verify the value is actually cached
	cache.mu.RLock()
	cachedColor, exists := cache.dimmedColor[featureName]
	cache.mu.RUnlock()

	if !exists {
		t.Error("Dimmed feature color was not cached")
	}
	if cachedColor != color1 {
		t.Errorf("Cached dimmed color mismatch: expected %s, got %s", color1, cachedColor)
	}
}

// TestGetFeatureColorEmpty tests empty feature name handling
func TestGetFeatureColorEmpty(t *testing.T) {
	color := GetFeatureColor("")
	if color != CurrentTheme.MutedColor {
		t.Errorf("Expected muted color for empty feature, got %s", color)
	}
}

// TestGetFeatureColorConcurrency tests thread safety of color caching
func TestGetFeatureColorConcurrency(t *testing.T) {
	// Clear cache for testing
	cache.mu.Lock()
	cache.featureColor = make(map[string]string)
	cache.dimmedColor = make(map[string]string)
	cache.mu.Unlock()

	const goroutines = 10
	const iterations = 100

	features := []string{"auth", "api", "database", "frontend", "backend"}

	var wg sync.WaitGroup //nolint:varnamelen // wg is idiomatic for WaitGroup
	results := make([]map[string]string, goroutines)

	// Start multiple goroutines accessing colors concurrently
	for i := 0; i < goroutines; i++ { //nolint:varnamelen // i is idiomatic for loop index
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			colors := make(map[string]string)

			for j := 0; j < iterations; j++ {
				for _, feature := range features {
					color := GetFeatureColor(feature)
					if prevColor, exists := colors[feature]; exists && prevColor != color {
						t.Errorf("Concurrent access returned different colors for %s: %s vs %s", feature, prevColor, color)
					}
					colors[feature] = color
				}
			}
			results[index] = colors
		}(i)
	}

	wg.Wait()

	// Verify all goroutines got the same colors for each feature
	if len(results) > 0 {
		baseline := results[0]
		for i := 1; i < len(results); i++ {
			for feature, color := range baseline {
				if results[i][feature] != color {
					t.Errorf("Goroutine %d got different color for %s: expected %s, got %s",
						i, feature, color, results[i][feature])
				}
			}
		}
	}
}

// TestGetMutedFeatureColor tests that muted colors are consistent
func TestGetMutedFeatureColor(t *testing.T) {
	// Test various feature names
	features := []string{"auth", "api", "", "very-long-feature-name-test"}

	for _, feature := range features {
		color := GetMutedFeatureColor(feature)
		if color != CurrentTheme.MutedColor {
			t.Errorf("GetMutedFeatureColor(%s) returned %s, expected %s",
				feature, color, CurrentTheme.MutedColor)
		}
	}
}

// BenchmarkGetFeatureColorCold tests performance of first call (cache miss)
func BenchmarkGetFeatureColorCold(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// Clear cache for each iteration to simulate cold cache
		cache.mu.Lock()
		cache.featureColor = make(map[string]string)
		cache.mu.Unlock()

		GetFeatureColor("benchmark-feature")
	}
}

// BenchmarkGetFeatureColorWarm tests performance of cached calls (cache hit)
func BenchmarkGetFeatureColorWarm(b *testing.B) {
	// Warm up the cache
	GetFeatureColor("benchmark-feature")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetFeatureColor("benchmark-feature")
	}
}

// BenchmarkGetDimmedFeatureColorWarm tests performance of cached dimmed color calls
func BenchmarkGetDimmedFeatureColorWarm(b *testing.B) {
	// Warm up the cache
	GetDimmedFeatureColor("benchmark-feature")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetDimmedFeatureColor("benchmark-feature")
	}
}
