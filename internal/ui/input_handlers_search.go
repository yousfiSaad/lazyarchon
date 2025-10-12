package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// =============================================================================
// SEARCH KEY HANDLERS
// =============================================================================
// This file contains all search-related keyboard handlers

// HandleActivateSearchKey handles '/' and 'ctrl+f' keys - activate inline search
func (m *MainModel) handleActivateSearchKey(key string) (tea.Cmd, bool) {
	if m.uiState.IsTaskView() && !m.uiState.SearchMode {
		m.activateInlineSearch()
		return nil, true
	}
	return nil, false
}

// HandleClearSearchKey handles 'ctrl+x' and 'ctrl+l' keys - clear current search
func (m *MainModel) handleClearSearchKey(key string) (tea.Cmd, bool) {
	// Direct state access (coordinators removed)
	if m.uiState.IsTaskView() && m.uiState.SearchActive {
		m.clearSearch()
		return nil, true
	}
	return nil, false
}

// HandleNextSearchMatchKey handles 'n' key - next search match
func (m *MainModel) handleNextSearchMatchKey(key string) (tea.Cmd, bool) {
	// Direct state access (coordinators removed)
	if m.uiState.IsTaskView() && m.uiState.SearchActive && m.uiState.TaskTotalMatches > 0 {
		cmd := m.nextSearchMatch()
		return cmd, true
	}
	return nil, false
}

// HandlePrevSearchMatchKey handles 'N' key - previous search match
func (m *MainModel) handlePrevSearchMatchKey(key string) (tea.Cmd, bool) {
	// Direct state access (coordinators removed)
	if m.uiState.IsTaskView() && m.uiState.SearchActive && m.uiState.TaskTotalMatches > 0 {
		cmd := m.previousSearchMatch()
		return cmd, true
	}
	return nil, false
}
