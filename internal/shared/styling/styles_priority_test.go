package styling

import "testing"

// TestGetTaskPriority pins the pure 1-5 mapping: no legacy task_order
// fallback may return (a 60 must read as backlog, not medium).
func TestGetTaskPriority(t *testing.T) {
	tests := []struct {
		priority int
		want     PriorityLevel
	}{
		{priority: 1, want: PriorityHigh},
		{priority: 2, want: PriorityHigh},
		{priority: 3, want: PriorityMedium},
		{priority: 4, want: PriorityLow},
		{priority: 5, want: PriorityBacklog},
		// Out-of-range values clamp: legacy 0-100 leftovers collapse onto
		// backlog, unset (0) onto low.
		{priority: 60, want: PriorityBacklog},
		{priority: 100, want: PriorityBacklog},
		{priority: 0, want: PriorityLow},
	}

	for _, tt := range tests { //nolint:varnamelen // tt is idiomatic for table-driven tests
		if got := GetTaskPriority(tt.priority); got != tt.want {
			t.Errorf("GetTaskPriority(%d) = %v, want %v", tt.priority, got, tt.want)
		}
	}
}

// TestPriorityLabel pins the number + tier label form for all five tiers.
func TestPriorityLabel(t *testing.T) {
	tests := []struct {
		priority int
		want     string
	}{
		{priority: 1, want: "P1 · Critical"},
		{priority: 2, want: "P2 · High"},
		{priority: 3, want: "P3 · Medium"},
		{priority: 4, want: "P4 · Low"},
		{priority: 5, want: "P5 · Backlog"},
		{priority: 9, want: "P? · Unknown"},
	}

	for _, tt := range tests { //nolint:varnamelen // tt is idiomatic for table-driven tests
		if got := PriorityLabel(tt.priority); got != tt.want {
			t.Errorf("PriorityLabel(%d) = %q, want %q", tt.priority, got, tt.want)
		}
	}
}

// TestGetPrioritySymbol covers every display level; each glyph must stay
// single-width so task lines keep their alignment.
func TestGetPrioritySymbol(t *testing.T) {
	tests := []struct {
		level PriorityLevel
		want  string
	}{
		{level: PriorityHigh, want: "▲"},
		{level: PriorityMedium, want: "△"},
		{level: PriorityLow, want: "▽"},
		{level: PriorityBacklog, want: "▼"},
		{level: PriorityLevel(99), want: " "},
	}

	for _, tt := range tests { //nolint:varnamelen // tt is idiomatic for table-driven tests
		if got := GetPrioritySymbol(tt.level); got != tt.want {
			t.Errorf("GetPrioritySymbol(%v) = %q, want %q", tt.level, got, tt.want)
		}
	}
}
