package helpers

import (
	"testing"

	"github.com/yousfisaad/lazyarchon/v2/internal/shared/interfaces"
)

func TestGetNextMatch_SingleMatch(t *testing.T) {
	matchingIndices := []int{5}
	currentIndex := 5

	// With single match, should stay on that match (wrap to itself)
	result := GetNextMatch(matchingIndices, currentIndex)
	if result != 5 {
		t.Errorf("Expected 5, got %d", result)
	}
}

func TestGetNextMatch_MultipleMatches_OnMatch(t *testing.T) {
	matchingIndices := []int{3, 7, 12}

	// At index 7 (second match), should go to 12 (third match)
	result := GetNextMatch(matchingIndices, 7)
	if result != 12 {
		t.Errorf("Expected 12, got %d", result)
	}

	// At index 12 (last match), should wrap to 3 (first match)
	result = GetNextMatch(matchingIndices, 12)
	if result != 3 {
		t.Errorf("Expected 3 (wrap), got %d", result)
	}

	// At index 3 (first match), should go to 7 (second match)
	result = GetNextMatch(matchingIndices, 3)
	if result != 7 {
		t.Errorf("Expected 7, got %d", result)
	}
}

func TestGetNextMatch_NotOnMatch(t *testing.T) {
	matchingIndices := []int{3, 7, 12}

	// At index 5 (between matches), should go to next match (7)
	result := GetNextMatch(matchingIndices, 5)
	if result != 7 {
		t.Errorf("Expected 7, got %d", result)
	}

	// At index 0 (before all matches), should go to first match (3)
	result = GetNextMatch(matchingIndices, 0)
	if result != 3 {
		t.Errorf("Expected 3, got %d", result)
	}

	// At index 15 (after all matches), should wrap to first match (3)
	result = GetNextMatch(matchingIndices, 15)
	if result != 3 {
		t.Errorf("Expected 3 (wrap), got %d", result)
	}
}

func TestGetPreviousMatch_SingleMatch(t *testing.T) {
	matchingIndices := []int{5}
	currentIndex := 5

	// With single match, should stay on that match (wrap to itself)
	result := GetPreviousMatch(matchingIndices, currentIndex)
	if result != 5 {
		t.Errorf("Expected 5, got %d", result)
	}
}

func TestGetPreviousMatch_MultipleMatches_OnMatch(t *testing.T) {
	matchingIndices := []int{3, 7, 12}

	// At index 7 (second match), should go to 3 (first match)
	result := GetPreviousMatch(matchingIndices, 7)
	if result != 3 {
		t.Errorf("Expected 3, got %d", result)
	}

	// At index 3 (first match), should wrap to 12 (last match)
	result = GetPreviousMatch(matchingIndices, 3)
	if result != 12 {
		t.Errorf("Expected 12 (wrap), got %d", result)
	}

	// At index 12 (last match), should go to 7 (second match)
	result = GetPreviousMatch(matchingIndices, 12)
	if result != 7 {
		t.Errorf("Expected 7, got %d", result)
	}
}

func TestGetPreviousMatch_NotOnMatch(t *testing.T) {
	matchingIndices := []int{3, 7, 12}

	// At index 10 (between matches), should go to previous match (7)
	result := GetPreviousMatch(matchingIndices, 10)
	if result != 7 {
		t.Errorf("Expected 7, got %d", result)
	}

	// At index 0 (before all matches), should wrap to last match (12)
	result = GetPreviousMatch(matchingIndices, 0)
	if result != 12 {
		t.Errorf("Expected 12 (wrap), got %d", result)
	}

	// At index 15 (after all matches), should go to last match (12)
	result = GetPreviousMatch(matchingIndices, 15)
	if result != 12 {
		t.Errorf("Expected 12, got %d", result)
	}
}

func TestGetNextMatch_EmptyMatches(t *testing.T) {
	matchingIndices := []int{}
	currentIndex := 5

	// With no matches, should stay at current position
	result := GetNextMatch(matchingIndices, currentIndex)
	if result != 5 {
		t.Errorf("Expected 5, got %d", result)
	}
}

func TestGetPreviousMatch_EmptyMatches(t *testing.T) {
	matchingIndices := []int{}
	currentIndex := 5

	// With no matches, should stay at current position
	result := GetPreviousMatch(matchingIndices, currentIndex)
	if result != 5 {
		t.Errorf("Expected 5, got %d", result)
	}
}

func TestSearchTasks_TitleMatch(t *testing.T) {
	tasks := []interfaces.Task{
		{ID: "t1", Title: "Fix login bug", Status: "todo"},
		{ID: "t2", Title: "Write documentation", Status: "todo"},
	}

	indices, total := SearchTasks(tasks, "login")

	if total != 1 || len(indices) != 1 || indices[0] != 0 {
		t.Errorf("Expected 1 title match at index 0, got indices=%v total=%d", indices, total)
	}
}

func TestSearchTasks_MatchesTagStatusAndID(t *testing.T) {
	// The search predicate must match the same fields the task list
	// highlights: title, status, tags, and ID.
	tasks := []interfaces.Task{
		{ID: "t1", Title: "Unrelated title", Status: "todo"},
		{ID: "t2", Title: "Also unrelated", Status: "todo", Tags: []string{"urgent"}},
		{ID: "t3", Title: "No match here", Status: "doing"},
		{ID: "t4", Title: "Nothing", Status: "todo", Tags: []string{"chore"}},
	}

	indices, total := SearchTasks(tasks, "urgent")
	if total != 1 || len(indices) != 1 || indices[0] != 1 {
		t.Errorf("tag match: expected only index 1, got indices=%v total=%d", indices, total)
	}

	indices, total = SearchTasks(tasks, "doing")
	if total != 1 || len(indices) != 1 || indices[0] != 2 {
		t.Errorf("status match: expected only index 2, got indices=%v total=%d", indices, total)
	}

	indices, total = SearchTasks(tasks, "t4")
	if total != 1 || len(indices) != 1 || indices[0] != 3 {
		t.Errorf("ID match: expected only index 3, got indices=%v total=%d", indices, total)
	}
}

func TestSearchTasks_CaseInsensitiveAndEmpty(t *testing.T) {
	tasks := []interfaces.Task{
		{ID: "t1", Title: "Fix LOGIN Bug", Status: "todo"},
	}

	if indices, total := SearchTasks(tasks, "  login "); total != 1 || indices[0] != 0 {
		t.Errorf("Expected trimmed case-insensitive match, got indices=%v total=%d", indices, total)
	}

	if indices, total := SearchTasks(tasks, ""); total != 0 || indices != nil {
		t.Errorf("Empty query must match nothing, got indices=%v total=%d", indices, total)
	}
}

func TestTaskMatchesQuery(t *testing.T) {
	task := interfaces.Task{
		ID:     "task-42",
		Title:  "Refactor auth module",
		Status: "review",
		Tags:   []string{"backend", "security"},
	}

	cases := []struct {
		name  string
		query string
		want  bool
	}{
		{"title substring", "auth", true},
		{"title case-insensitive", "REFACTOR", true},
		{"status", "review", true},
		{"tag", "security", true},
		{"id", "task-42", true},
		{"query with whitespace", "  backend ", true},
		{"no match", "nonexistent", false},
		{"empty query", "", false},
	}

	for _, tc := range cases {
		if got := TaskMatchesQuery(task, tc.query); got != tc.want {
			t.Errorf("%s: TaskMatchesQuery(task, %q) = %v, want %v", tc.name, tc.query, got, tc.want)
		}
	}
}
