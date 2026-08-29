package helpers

import "testing"

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
