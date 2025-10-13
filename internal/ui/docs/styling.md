# Styling System

## Overview

LazyArchon uses a **modular, context-aware styling system** built on [lipgloss](https://github.com/charmbracelet/lipgloss). The system supports multiple themes, configurable status colors, and intelligent style composition.

## Architecture

### Styling Modules

The styling system is organized into focused modules:

**Location**: `internal/shared/styling/`

```
styling/
├── styles.go              # Main entry point, constants
├── styles_theme.go        # Theme management
├── styles_factory.go      # Style creation functions
├── styles_colors.go       # Color management, status colors
├── styles_context.go      # Style context for state management
├── styles_state.go        # Selection and search state
├── styles_utils.go        # Utility rendering functions
└── task_line_builder.go   # Task line rendering
```

### Core Components

1. **Theme Management** (`styles_theme.go`)
   - Theme definitions (default, monokai, gruvbox, dracula)
   - Theme loading and initialization
   - Status color schemes (blue, gray, warm_gray, cool_gray)

2. **Style Factory** (`styles_factory.go`)
   - Creates lipgloss styles based on context
   - Handles panels, text, status bars
   - Contextual style variations

3. **Style Context** (`styles_context.go`)
   - Centralized styling state container
   - Selection and search state tracking
   - Immutable context transformations

4. **Color Management** (`styles_colors.go`)
   - Status color mappings
   - Feature tag colors
   - Color schemes for different preferences

## Themes

### Available Themes

LazyArchon supports 4 built-in themes:

```go
const (
    ThemeDefault = "default"
    ThemeMonokai = "monokai"
    ThemeGruvbox = "gruvbox"
    ThemeDracula = "dracula"
)
```

### Theme Structure

```go
type Theme struct {
    Name string

    // Core colors
    Background    string
    Foreground    string
    AccentColor   string
    HeaderColor   string
    StatusColor   string
    MutedColor    string

    // Status colors (default: blue scheme)
    TodoColor     string  // Task status: todo
    DoingColor    string  // Task status: doing
    ReviewColor   string  // Task status: review
    DoneColor     string  // Task status: done

    // UI element colors
    BorderColor        string
    PanelBorderColor   string
    SelectionColor     string
    ErrorColor         string
    WarningColor       string
    SuccessColor       string

    // Feature tag colors
    FeatureColors []string
}
```

### Theme Configuration

Set theme in `config.yaml`:

```yaml
ui:
  theme:
    name: "default"  # or monokai, gruvbox, dracula
  display:
    status_color_scheme: "gray"  # or blue, warm_gray, cool_gray
```

### Initializing Themes

```go
// Initialize theme from config
func InitializeThemeNew(cfg *config.Config) {
    themeName := cfg.GetThemeName()

    switch themeName {
    case ThemeMonokai:
        ActiveTheme = MonokaiTheme
    case ThemeGruvbox:
        ActiveTheme = GruvboxTheme
    case ThemeDracula:
        ActiveTheme = DraculaTheme
    default:
        ActiveTheme = DefaultTheme
    }

    // Apply status color scheme
    colorScheme := cfg.GetStatusColorScheme()
    ApplyStatusColorScheme(&ActiveTheme, colorScheme)
}
```

## Status Color Schemes

### Available Schemes

LazyArchon supports 4 status color schemes to match different preferences and work environments:

```go
const (
    StatusColorSchemeBlue     = "blue"       // Vibrant (default)
    StatusColorSchemeGray     = "gray"       // Neutral, productivity-focused
    StatusColorSchemeWarmGray = "warm_gray"  // Gentle warmth
    StatusColorSchemeCoolGray = "cool_gray"  // Modern, professional
)
```

### Color Scheme Details

**Blue Scheme (Default)**:
```go
TodoColor:   "#3498db"  // Bright blue
DoingColor:  "#9b59b6"  // Purple
ReviewColor: "#e67e22"  // Orange
DoneColor:   "#2ecc71"  // Green
```

**Gray Scheme**:
```go
TodoColor:   "#95a5a6"  // Light gray - lowest attention
DoingColor:  "#7f8c8d"  // Medium gray - moderate attention
ReviewColor: "#34495e"  // Dark gray - high attention
DoneColor:   "#ecf0f1"  // Very light gray - minimal attention
```

**Warm Gray Scheme**:
```go
TodoColor:   "#b8a898"  // Warm light gray
DoingColor:  "#a89478"  // Warm medium gray
ReviewColor: "#806040"  // Warm dark gray
DoneColor:   "#d8ccc0"  // Warm very light gray
```

**Cool Gray Scheme**:
```go
TodoColor:   "#98a8b8"  // Cool light gray
DoingColor:  "#7888a0"  // Cool medium gray
ReviewColor: "#405060"  // Cool dark gray
DoneColor:   "#c0ccd8"  // Cool very light gray
```

### Color Hierarchy

All color schemes maintain the same attention hierarchy:

```
Review > Doing > Todo > Done
(highest attention → lowest attention)
```

**Reasoning**:
- **Review**: Needs immediate attention/approval
- **Doing**: Active work in progress
- **Todo**: Can wait to be started
- **Done**: Completed, minimal attention needed

### Applying Color Schemes

```go
func ApplyStatusColorScheme(theme *Theme, scheme string) {
    switch scheme {
    case StatusColorSchemeGray:
        theme.ReviewColor = "#34495e"  // Darkest - highest priority
        theme.DoingColor = "#7f8c8d"   // Medium
        theme.TodoColor = "#95a5a6"    // Light
        theme.DoneColor = "#ecf0f1"    // Lightest - lowest priority

    case StatusColorSchemeWarmGray:
        theme.ReviewColor = "#806040"
        theme.DoingColor = "#a89478"
        theme.TodoColor = "#b8a898"
        theme.DoneColor = "#d8ccc0"

    case StatusColorSchemeCoolGray:
        theme.ReviewColor = "#405060"
        theme.DoingColor = "#7888a0"
        theme.TodoColor = "#98a8b8"
        theme.DoneColor = "#c0ccd8"

    case StatusColorSchemeBlue:
        // Keep default theme colors
        // No changes needed
    }
}
```

## Status Symbols

### Symbol Definitions

```go
const (
    StatusSymbolTodo   = "○"  // Empty circle - clear starting state
    StatusSymbolDoing  = "◐"  // Half-filled circle - work in progress
    StatusSymbolReview = "◈"  // Diamond with center dot - under review
    StatusSymbolDone   = "✓"  // Checkmark - completed
)
```

### Symbol Selection

```go
func GetStatusSymbol(status string) string {
    switch status {
    case archon.TaskStatusTodo:
        return StatusSymbolTodo
    case archon.TaskStatusDoing:
        return StatusSymbolDoing
    case archon.TaskStatusReview:
        return StatusSymbolReview
    case archon.TaskStatusDone:
        return StatusSymbolDone
    default:
        return "?"
    }
}
```

## Style Context

### Context Structure

The `StyleContext` provides centralized styling state:

```go
type StyleContext struct {
    theme          *ThemeAdapter
    selectionState SelectionState
    searchState    SearchState
    styleProvider  StyleProvider
    factory        *StyleFactory
}
```

### Selection State

```go
type SelectionState struct {
    IsSelected bool
    IsFocused  bool
}

func NewSelectionState() SelectionState {
    return SelectionState{
        IsSelected: false,
        IsFocused:  false,
    }
}

func (s SelectionState) WithSelected(selected bool) SelectionState {
    s.IsSelected = selected
    return s
}
```

### Search State

```go
type SearchState struct {
    Query    string
    IsActive bool
}

func NewSearchState() SearchState {
    return SearchState{
        Query:    "",
        IsActive: false,
    }
}

func (s SearchState) WithQuery(query string) SearchState {
    s.Query = query
    s.IsActive = query != ""
    return s
}
```

### Context Transformations

Style contexts are immutable - transformations create new contexts:

```go
// Create base context
ctx := styling.NewStyleContext(theme, styleProvider)

// Transform for selected item
selectedCtx := ctx.WithSelection(true)

// Transform for search
searchCtx := ctx.WithSearch("auth", true)

// Chain transformations
ctx := baseCtx.
    WithSelection(true).
    WithSearch(query, true)
```

## Style Factory

### Factory Structure

```go
type StyleFactory struct {
    context *StyleContext
}

func NewStyleFactory() *StyleFactory {
    return &StyleFactory{}
}
```

### Factory Methods

**Panel Styles**:
```go
func (f *StyleFactory) Panel(width, height int, isActive bool) lipgloss.Style {
    borderColor := f.context.theme.PanelBorderColor
    if isActive {
        borderColor = f.context.theme.AccentColor
    }

    return lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color(borderColor)).
        Width(width).
        Height(height)
}
```

**Text Styles**:
```go
func (f *StyleFactory) Text(color string) lipgloss.Style {
    style := lipgloss.NewStyle().
        Foreground(lipgloss.Color(color))

    if f.context.selectionState.IsSelected {
        style = style.Bold(true)
    }

    return style
}
```

**Status Bar Styles**:
```go
func (f *StyleFactory) StatusBar(state string) lipgloss.Style {
    backgroundColor := f.context.theme.StatusColor

    if state == "error" {
        backgroundColor = f.context.theme.ErrorColor
    } else if state == "loading" {
        backgroundColor = f.context.theme.AccentColor
    }

    return lipgloss.NewStyle().
        Background(lipgloss.Color(backgroundColor)).
        Foreground(lipgloss.Color(f.context.theme.Foreground)).
        Padding(0, 1)
}
```

## Task Line Rendering

### Task Line Builder

The `TaskLineBuilder` constructs formatted task lines:

```go
type TaskLineBuilder struct {
    styleContext *styling.StyleContext
    config       interfaces.ConfigProvider
}

func (b *TaskLineBuilder) BuildTaskLine(task archon.Task, allTasks []archon.Task) string {
    var parts []string

    // Status symbol with color
    statusPart := b.buildStatusPart(task.Status)
    parts = append(parts, statusPart)

    // Priority indicator (if enabled)
    if b.config.IsPriorityIndicatorsEnabled() {
        priorityPart := b.buildPriorityPart(task.TaskOrder, allTasks)
        parts = append(parts, priorityPart)
    }

    // Task title
    titlePart := b.buildTitlePart(task.Title)
    parts = append(parts, titlePart)

    // Feature tag (if enabled and present)
    if b.config.IsFeatureColorsEnabled() && task.Feature != "" {
        featurePart := b.buildFeaturePart(task.Feature)
        parts = append(parts, featurePart)
    }

    return strings.Join(parts, " ")
}
```

### Status Part Rendering

```go
func (b *TaskLineBuilder) buildStatusPart(status string) string {
    symbol := styling.GetStatusSymbol(status)
    color := styling.GetStatusColor(status)

    style := lipgloss.NewStyle().
        Foreground(lipgloss.Color(color))

    return style.Render(symbol)
}
```

### Priority Part Rendering

```go
func (b *TaskLineBuilder) buildPriorityPart(taskOrder int, allTasks []archon.Task) string {
    priority := styling.GetTaskPriority(taskOrder, allTasks)
    symbol := styling.GetPrioritySymbol(priority)

    var color string
    switch priority {
    case styling.PriorityHigh:
        color = "#e74c3c"  // Red
    case styling.PriorityMedium:
        color = "#f39c12"  // Orange
    case styling.PriorityLow:
        color = "#95a5a6"  // Gray
    }

    style := lipgloss.NewStyle().
        Foreground(lipgloss.Color(color))

    return style.Render(symbol)
}
```

### Feature Tag Rendering

```go
func (b *TaskLineBuilder) buildFeaturePart(feature string) string {
    // Get feature color from theme
    colorIndex := hashString(feature) % len(b.styleContext.Theme().FeatureColors)
    color := b.styleContext.Theme().FeatureColors[colorIndex]

    style := lipgloss.NewStyle().
        Foreground(lipgloss.Color(color)).
        Background(lipgloss.Color("#2c3e50")).
        Padding(0, 1)

    return style.Render(feature)
}
```

## Layout Constants

### UI Dimensions

```go
const (
    HeaderHeight       = 1
    StatusBarHeight    = 1
    BorderWidth        = 2
    PanelPadding       = 1
    SelectionIndicator = "→ "
    NoSelection        = "  "
    MaxTasksPerPage    = 100
)
```

### Panel Layout

```go
func CalculatePanelDimensions(screenWidth, screenHeight int) (int, int) {
    availableHeight := screenHeight - HeaderHeight - StatusBarHeight - BorderWidth

    // Split width for dual-panel layout
    panelWidth := (screenWidth - BorderWidth) / 2

    return panelWidth, availableHeight
}
```

## Styling Utilities

### Truncate Text

```go
func TruncateWithEllipsis(text string, maxWidth int) string {
    if lipgloss.Width(text) <= maxWidth {
        return text
    }

    ellipsis := "..."
    ellipsisWidth := lipgloss.Width(ellipsis)

    // Calculate how much text we can show
    availableWidth := maxWidth - ellipsisWidth

    // Truncate text
    truncated := text
    for lipgloss.Width(truncated) > availableWidth {
        runes := []rune(truncated)
        truncated = string(runes[:len(runes)-1])
    }

    return truncated + ellipsis
}
```

### Center Text

```go
func CenterText(text string, width int) string {
    textWidth := lipgloss.Width(text)
    if textWidth >= width {
        return text
    }

    padding := (width - textWidth) / 2
    return strings.Repeat(" ", padding) + text
}
```

### Align Text

```go
func AlignRight(text string, width int) string {
    textWidth := lipgloss.Width(text)
    if textWidth >= width {
        return text
    }

    padding := width - textWidth
    return strings.Repeat(" ", padding) + text
}
```

## Component Styling

### Header Styling

```go
func RenderHeader(title string, width int, theme *Theme) string {
    style := lipgloss.NewStyle().
        Background(lipgloss.Color(theme.HeaderColor)).
        Foreground(lipgloss.Color(theme.Foreground)).
        Bold(true).
        Width(width).
        Padding(0, 1)

    return style.Render(title)
}
```

### Status Bar Styling

```go
func RenderStatusBar(segments []string, width int, theme *Theme, hasError bool) string {
    backgroundColor := theme.StatusColor
    if hasError {
        backgroundColor = theme.ErrorColor
    }

    content := strings.Join(segments, " | ")

    style := lipgloss.NewStyle().
        Background(lipgloss.Color(backgroundColor)).
        Foreground(lipgloss.Color(theme.Foreground)).
        Width(width).
        Padding(0, 1)

    return style.Render(content)
}
```

### Panel Styling

```go
func RenderPanel(content string, width, height int, title string, isActive bool, theme *Theme) string {
    borderColor := theme.PanelBorderColor
    if isActive {
        borderColor = theme.AccentColor
    }

    panelStyle := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color(borderColor)).
        Width(width).
        Height(height).
        Padding(1)

    // Add title to border if provided
    if title != "" {
        panelStyle = panelStyle.BorderTop(true).
            BorderTopForeground(lipgloss.Color(borderColor))
    }

    return panelStyle.Render(content)
}
```

## Modal Styling

### Modal Overlay

```go
func RenderModal(content string, width, height int, screenWidth, screenHeight int, theme *Theme) string {
    // Create modal box
    modalStyle := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color(theme.AccentColor)).
        Background(lipgloss.Color(theme.Background)).
        Padding(1).
        Width(width).
        Height(height)

    modalBox := modalStyle.Render(content)

    // Center on screen
    return lipgloss.Place(
        screenWidth,
        screenHeight,
        lipgloss.Center,
        lipgloss.Center,
        modalBox,
    )
}
```

### Modal Title

```go
func RenderModalTitle(title string, width int, theme *Theme) string {
    style := lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color(theme.AccentColor)).
        Width(width).
        Align(lipgloss.Center).
        Padding(0, 0, 1, 0)

    return style.Render(title)
}
```

## Scroll Indicators

### Vertical Scroll Indicator

```go
func RenderScrollIndicator(yOffset, maxOffset, height int, theme *Theme) string {
    if maxOffset <= 0 {
        return ""  // No scrolling needed
    }

    // Calculate scroll position
    percentage := float64(yOffset) / float64(maxOffset)
    position := int(percentage * float64(height-1))

    var lines []string
    for i := 0; i < height; i++ {
        if i == position {
            lines = append(lines, "█")  // Solid block at position
        } else {
            lines = append(lines, "│")  // Light line
        }
    }

    style := lipgloss.NewStyle().
        Foreground(lipgloss.Color(theme.MutedColor))

    return style.Render(strings.Join(lines, "\n"))
}
```

### Scroll Position Indicator

```go
func RenderScrollPosition(current, total int, theme *Theme) string {
    text := fmt.Sprintf("%d/%d", current+1, total)

    style := lipgloss.NewStyle().
        Foreground(lipgloss.Color(theme.MutedColor)).
        Padding(0, 1)

    return style.Render(text)
}
```

## Configuration Integration

### Style Provider Interface

```go
type StyleProvider interface {
    IsPriorityIndicatorsEnabled() bool
    IsFeatureColorsEnabled() bool
}
```

### Config Provider

```go
type ConfigProvider interface {
    GetThemeName() string
    GetStatusColorScheme() string
    IsPriorityIndicatorsEnabled() bool
    IsFeatureColorsEnabled() bool
}
```

### Loading Configuration

```go
func LoadStylingFromConfig(cfg *config.Config) *StyleContext {
    // Initialize theme
    InitializeThemeNew(cfg)

    // Create theme adapter
    themeAdapter := &ThemeAdapter{
        TodoColor:     ActiveTheme.TodoColor,
        DoingColor:    ActiveTheme.DoingColor,
        ReviewColor:   ActiveTheme.ReviewColor,
        DoneColor:     ActiveTheme.DoneColor,
        HeaderColor:   ActiveTheme.HeaderColor,
        MutedColor:    ActiveTheme.MutedColor,
        AccentColor:   ActiveTheme.AccentColor,
        StatusColor:   ActiveTheme.StatusColor,
        FeatureColors: ActiveTheme.FeatureColors,
        Name:          ActiveTheme.Name,
    }

    // Create style context
    return NewStyleContext(themeAdapter, cfg)
}
```

## Best Practices

### 1. Use Style Context

```go
// ✅ GOOD: Context-aware styling
ctx := styling.NewStyleContext(theme, styleProvider)
style := ctx.Factory().Text(color)

// ❌ BAD: Direct style creation
style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
```

### 2. Use Immutable Transformations

```go
// ✅ GOOD: Immutable context
baseCtx := styling.NewStyleContext(theme, styleProvider)
selectedCtx := baseCtx.WithSelection(true)
// baseCtx unchanged, selectedCtx is new

// ❌ BAD: Mutable state
ctx.selectionState.IsSelected = true
```

### 3. Leverage Theme Colors

```go
// ✅ GOOD: Use theme colors
color := theme.AccentColor

// ❌ BAD: Hardcoded colors
color := "#3498db"
```

### 4. Respect Configuration

```go
// ✅ GOOD: Check config
if config.IsPriorityIndicatorsEnabled() {
    renderPriority()
}

// ❌ BAD: Always render
renderPriority()
```

### 5. Use Constants

```go
// ✅ GOOD: Use constants
symbol := styling.StatusSymbolDone

// ❌ BAD: Magic strings
symbol := "✓"
```

## Testing Styles

### Test Theme Loading

```go
func TestLoadTheme(t *testing.T) {
    cfg := &config.Config{
        UI: config.UIConfig{
            Theme: config.ThemeConfig{
                Name: "monokai",
            },
        },
    }

    InitializeThemeNew(cfg)

    assert.Equal(t, "monokai", ActiveTheme.Name)
    assert.NotEmpty(t, ActiveTheme.TodoColor)
}
```

### Test Color Scheme Application

```go
func TestApplyColorScheme(t *testing.T) {
    theme := DefaultTheme

    ApplyStatusColorScheme(&theme, StatusColorSchemeGray)

    assert.Equal(t, "#34495e", theme.ReviewColor)
    assert.Equal(t, "#7f8c8d", theme.DoingColor)
}
```

### Test Style Context

```go
func TestStyleContext(t *testing.T) {
    ctx := NewStyleContext(theme, styleProvider)

    selectedCtx := ctx.WithSelection(true)

    assert.True(t, selectedCtx.IsSelected())
    assert.False(t, ctx.IsSelected())  // Original unchanged
}
```

## References

- [Lipgloss Documentation](https://github.com/charmbracelet/lipgloss)
- [Component System](./components.md)
- [Configuration](../../config/)
- [Theme Definitions](../../shared/styling/styles_theme.go)
