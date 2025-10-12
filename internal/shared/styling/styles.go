package styling

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/yousfisaad/lazyarchon/v2/internal/archon"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/config"
)

// styles.go - Main entry point for LazyArchon UI styling
// This file provides a compatibility layer for the modular styling system.
//
// The styling system is now organized into focused modules:
// - styles_theme.go: Theme management and configuration
// - styles_factory.go: Style creation functions
// - styles_colors.go: Color management and status colors
// - styles_utils.go: Utility functions for rendering

// Global style variables that need to be accessible from other modules
var (
	// Header style for section headers
	HeaderStyle lipgloss.Style
	// Status bar style for bottom status line
	StatusBarStyle lipgloss.Style
	// Base panel style for containers
	BasePanelStyle lipgloss.Style
	// Detail header style for task detail sections
	DetailHeaderStyle lipgloss.Style
	// Scroll indicator style
	ScrollIndicatorStyle lipgloss.Style
	// Tag style for feature tags
	TagStyle lipgloss.Style
)

// Status symbols - Single Source of Truth
const (
	StatusSymbolTodo   = "○" // Empty circle - clear starting state
	StatusSymbolDoing  = "◐" // Half-filled circle - work in progress
	StatusSymbolReview = "◈" // Diamond with center dot - under review
	StatusSymbolDone   = "✓" // Checkmark - completed (universally understood)
)

// Task status constants
const (
	StatusTodo   = "todo"
	StatusDoing  = "doing"
	StatusReview = "review"
	StatusDone   = "done"
)

// Priority level constants
const (
	PriorityLowStr    = "low"
	PriorityMediumStr = "medium"
	PriorityHighStr   = "high"
)

// UI layout constants
const (
	HeaderHeight       = 1
	StatusBarHeight    = 1
	BorderWidth        = 2
	PanelPadding       = 1
	SelectionIndicator = "→ " // Arrow indicator for better visibility
	NoSelection        = "  "
	MaxTasksPerPage    = 100
)

// Priority levels for task ordering
type PriorityLevel int

const (
	PriorityLow PriorityLevel = iota
	PriorityMedium
	PriorityHigh
)

// DebugLog is available from debug_log.go

// Theme management compatibility functions
func InitializeTheme(cfg *config.Config) {
	InitializeThemeNew(cfg)
}

// Style factory compatibility functions
func CreatePanelStyle(width, height int) lipgloss.Style {
	factory := NewStyleFactory()
	return factory.Panel(width, height, false)
}

func CreateTaskItemStyle(selected bool, statusColor string) lipgloss.Style {
	factory := NewStyleFactory()
	return factory.Text(statusColor)
}

func CreateProjectItemStyle(selected bool, isAllTasks bool) lipgloss.Style {
	factory := NewStyleFactory()
	return factory.ProjectItem(selected, isAllTasks)
}

func CreateScrollBarStyle(width, height int) lipgloss.Style {
	factory := NewStyleFactory()
	return factory.Panel(width, height, false)
}

func CreateActivePanelStyle(width, height int, isActive bool) lipgloss.Style {
	factory := NewStyleFactory()
	return factory.Panel(width, height, isActive)
}

func CreateStatusBarStyle(state string) lipgloss.Style {
	factory := NewStyleFactory()
	return factory.StatusBar(state)
}

// Priority utility functions
func GetPrioritySymbol(priority PriorityLevel) string {
	switch priority {
	case PriorityHigh:
		return "▲" // Upward triangle - urgent/ascending
	case PriorityMedium:
		return "△" // Empty triangle - moderate attention
	case PriorityLow:
		return "▽" // Downward triangle - low priority
	default:
		return " " // Single space for no priority
	}
}

func GetTaskPriority(taskOrder int, allTasks []archon.Task) PriorityLevel {
	// Simplified priority calculation for compatibility
	if taskOrder >= 80 {
		return PriorityHigh
	} else if taskOrder >= 50 {
		return PriorityMedium
	}
	return PriorityLow
}

// Re-export CurrentTheme from styles_theme.go for backward compatibility
var CurrentTheme = &ActiveTheme
