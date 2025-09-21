package di

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/internal/archon"
	"github.com/yousfisaad/lazyarchon/internal/config"
	"github.com/yousfisaad/lazyarchon/internal/interfaces"
	"github.com/yousfisaad/lazyarchon/internal/logger"
	"github.com/yousfisaad/lazyarchon/internal/ui"
)

// NewConfigProvider creates a config provider instance
func NewConfigProvider() (interfaces.ConfigProvider, error) {
	cfg, err := config.Load()
	if err != nil {
		// Create a basic default config if loading fails
		cfg = &config.Config{
			Server: config.ServerConfig{
				URL:    "http://localhost:8080",
				APIKey: "",
			},
			UI: config.UIConfig{
				Theme: config.ThemeConfig{
					Name: "default",
				},
				Display: config.DisplayConfig{
					ShowCompletedTasks:  true,
					DefaultSortMode:     "priority",
					StatusColorScheme:   "blue",
					FeatureColors:       true,
					FeatureBackgrounds:  false,
					PriorityIndicators:  true,
				},
			},
			Development: config.DevelopmentConfig{
				Debug: false,
			},
		}
	}
	return cfg, nil
}

// NewArchonClient creates an Archon client instance
func NewArchonClient(config interfaces.ConfigProvider) (interfaces.ArchonClient, error) {
	baseClient := archon.NewClient(config.GetServerURL(), config.GetAPIKey())

	// Wrap with resilient client for production reliability
	resilientClient := archon.NewResilientClientFromBase(baseClient)

	return resilientClient, nil
}

// NewWebSocketClient creates a WebSocket client instance
func NewWebSocketClient(config interfaces.ConfigProvider) (interfaces.RealtimeClient, error) {
	// Create WebSocket client using the same config as HTTP client
	wsClient := archon.NewWebSocketClient(config.GetServerURL(), config.GetAPIKey())
	return wsClient, nil
}

// NewViewportFactory creates a viewport factory instance
func NewViewportFactory() interfaces.ViewportFactory {
	return &ViewportFactoryImpl{}
}

// ViewportFactoryImpl implements the ViewportFactory interface
type ViewportFactoryImpl struct{}

func (f *ViewportFactoryImpl) CreateTaskDetailsViewport(width, height int) viewport.Model {
	vp := viewport.New(width, height)
	vp.SetContent("")
	return vp
}

func (f *ViewportFactoryImpl) CreateHelpModalViewport(width, height int) viewport.Model {
	vp := viewport.New(width, height)
	vp.SetContent("")
	return vp
}

// NewStyleContextProvider creates a style context provider
func NewStyleContextProvider(config interfaces.ConfigProvider) interfaces.StyleContextProvider {
	return &StyleContextProviderImpl{config: config}
}

// StyleContextProviderImpl implements the StyleContextProvider interface
type StyleContextProviderImpl struct {
	config interfaces.ConfigProvider
}

func (s *StyleContextProviderImpl) CreateStyleContext(forceBackground bool) interface{} {
	// For now, return a simple placeholder
	// The actual styling system integration will be completed in the next phase
	return struct{}{}
}

func (s *StyleContextProviderImpl) GetTheme() interface{} {
	return s.config.GetTheme()
}

// NewCommandExecutor creates a command executor instance
func NewCommandExecutor() interfaces.CommandExecutor {
	return &CommandExecutorImpl{}
}

// CommandExecutorImpl implements the CommandExecutor interface
type CommandExecutorImpl struct{}

func (c *CommandExecutorImpl) LoadTasks(client interfaces.ArchonClient, projectID *string) interface{} {
	return func() tea.Msg {
		response, err := client.ListTasks(projectID, nil, false)
		if err != nil {
			return ui.TasksLoadedMsg{Error: err}
		}
		return ui.TasksLoadedMsg{Tasks: response.Tasks}
	}
}

func (c *CommandExecutorImpl) LoadProjects(client interfaces.ArchonClient) interface{} {
	return func() tea.Msg {
		response, err := client.ListProjects()
		if err != nil {
			return ui.ProjectsLoadedMsg{Error: err}
		}
		return ui.ProjectsLoadedMsg{Projects: response.Projects}
	}
}

func (c *CommandExecutorImpl) UpdateTaskStatus(client interfaces.ArchonClient, taskID string, newStatus string) interface{} {
	return func() tea.Msg {
		updates := archon.UpdateTaskRequest{Status: &newStatus}
		response, err := client.UpdateTask(taskID, updates)
		if err != nil {
			return ui.TaskUpdateMsg{Error: err}
		}
		return ui.TaskUpdateMsg{Task: &response.Task}
	}
}

func (c *CommandExecutorImpl) RefreshData(client interfaces.ArchonClient, selectedProjectID *string) interface{} {
	return tea.Batch(
		c.LoadTasks(client, selectedProjectID).(func() tea.Msg),
		c.LoadProjects(client).(func() tea.Msg),
	)
}

// NewLogger creates a structured logger instance
func NewLogger(configProvider interfaces.ConfigProvider) interfaces.Logger {
	// Convert config provider to our config type to create structured logger
	// Since NewConfigProvider returns *config.Config, we can safely cast it
	if cfg, ok := configProvider.(*config.Config); ok {
		structuredLogger := logger.New(cfg)
		logger.SetDefault(structuredLogger)
		return &StructuredLoggerAdapter{logger: structuredLogger}
	}

	// Fallback to basic logger if config conversion fails
	return &LoggerImpl{
		debugEnabled: configProvider.IsDebugEnabled(),
	}
}

// StructuredLoggerAdapter adapts our structured logger to the interface
type StructuredLoggerAdapter struct {
	logger *logger.Logger
}

func (l *StructuredLoggerAdapter) Debug(msg string, args ...interface{}) {
	l.logger.Debug("app", msg, convertToSlogArgs(args...)...)
}

func (l *StructuredLoggerAdapter) Info(msg string, args ...interface{}) {
	l.logger.Info(msg, convertToSlogArgs(args...)...)
}

func (l *StructuredLoggerAdapter) Warn(msg string, args ...interface{}) {
	l.logger.Warn(msg, convertToSlogArgs(args...)...)
}

func (l *StructuredLoggerAdapter) Error(msg string, args ...interface{}) {
	if len(args) > 0 {
		if err, ok := args[0].(error); ok {
			l.logger.Error(msg, err, convertToSlogArgs(args[1:]...)...)
			return
		}
	}
	l.logger.Logger.Error(msg, convertToSlogArgs(args...)...)
}

func (l *StructuredLoggerAdapter) Fatal(msg string, args ...interface{}) {
	l.logger.Logger.Error(msg, convertToSlogArgs(args...)...)
	os.Exit(1)
}

// LoggerImpl implements the Logger interface (fallback)
type LoggerImpl struct {
	debugEnabled bool
}

func (l *LoggerImpl) Debug(msg string, args ...interface{}) {
	if l.debugEnabled {
		log.Printf("[DEBUG] "+msg, args...)
	}
}

func (l *LoggerImpl) Info(msg string, args ...interface{}) {
	log.Printf("[INFO] "+msg, args...)
}

func (l *LoggerImpl) Warn(msg string, args ...interface{}) {
	log.Printf("[WARN] "+msg, args...)
}

func (l *LoggerImpl) Error(msg string, args ...interface{}) {
	log.Printf("[ERROR] "+msg, args...)
}

func (l *LoggerImpl) Fatal(msg string, args ...interface{}) {
	log.Printf("[FATAL] "+msg, args...)
	os.Exit(1)
}

// NewHealthChecker creates a health checker instance
func NewHealthChecker(logger interfaces.Logger) interfaces.HealthChecker {
	return &HealthCheckerImpl{
		logger: logger,
	}
}

// HealthCheckerImpl implements the HealthChecker interface
type HealthCheckerImpl struct {
	logger interfaces.Logger
}

func (h *HealthCheckerImpl) CheckAPIConnection(client interfaces.ArchonClient) error {
	h.logger.Debug("Checking API connection...")
	if err := client.HealthCheck(); err != nil {
		h.logger.Error("API connection check failed: %v", err)
		return fmt.Errorf("API connection failed: %w", err)
	}
	h.logger.Debug("API connection check passed")
	return nil
}

func (h *HealthCheckerImpl) CheckConfiguration(config interfaces.ConfigProvider) error {
	h.logger.Debug("Checking configuration...")

	if config.GetServerURL() == "" {
		err := fmt.Errorf("server URL not configured")
		h.logger.Error("Configuration check failed: %v", err)
		return err
	}

	if config.GetAPIKey() == "" {
		err := fmt.Errorf("API key not configured")
		h.logger.Error("Configuration check failed: %v", err)
		return err
	}

	h.logger.Debug("Configuration check passed")
	return nil
}

func (h *HealthCheckerImpl) GetHealthStatus() (bool, []string) {
	var issues []string

	// This would be populated by the actual health checks
	// For now, return healthy status
	return len(issues) == 0, issues
}

// NewUIModel creates the main UI model with all dependencies injected
func NewUIModel(
	client interfaces.ArchonClient,
	wsClient interfaces.RealtimeClient,
	config interfaces.ConfigProvider,
	viewportFactory interfaces.ViewportFactory,
	styleContextProvider interfaces.StyleContextProvider,
	commandExecutor interfaces.CommandExecutor,
	logger interfaces.Logger,
	healthChecker interfaces.HealthChecker,
) (tea.Model, error) {

	logger.Info("Creating UI model with dependency injection")

	// Create the concrete UI model
	model := ui.NewModelWithDependencies(
		client,
		wsClient,
		config,
		viewportFactory,
		styleContextProvider,
		commandExecutor,
		logger,
		healthChecker,
	)

	// Return the model wrapped to implement tea.Model
	return &UIModelWrapper{model: model}, nil
}

// UIModelWrapper wraps ui.Model to implement interfaces.UIModel
type UIModelWrapper struct {
	model ui.Model
}

func (w *UIModelWrapper) Init() tea.Cmd {
	return w.model.Init()
}

func (w *UIModelWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	newModel, cmd := w.model.Update(msg)
	// Update the wrapped model
	if newModel, ok := newModel.(ui.Model); ok {
		w.model = newModel
	}
	return w, cmd
}

func (w *UIModelWrapper) View() string {
	return w.model.View()
}

func (w *UIModelWrapper) GetTasks() []archon.Task {
	return w.model.GetTasks()
}

func (w *UIModelWrapper) GetProjects() []archon.Project {
	return w.model.GetProjects()
}

func (w *UIModelWrapper) GetSelectedProjectID() *string {
	return w.model.GetSelectedProjectID()
}

func (w *UIModelWrapper) IsLoading() bool {
	return w.model.IsLoading()
}

func (w *UIModelWrapper) GetError() string {
	return w.model.GetError()
}

// Helper functions for converting args to slog format
func convertToSlogArgs(args ...interface{}) []any {
	result := make([]any, len(args))
	copy(result, args)
	return result
}

func convertToSlogAttrs(args ...interface{}) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(args)/2)
	for i := 0; i < len(args)-1; i += 2 {
		if key, ok := args[i].(string); ok {
			attrs = append(attrs, slog.Any(key, args[i+1]))
		}
	}
	return attrs
}