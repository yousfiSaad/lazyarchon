package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yousfisaad/lazyarchon/v2/internal/bootstrap"
	"github.com/yousfisaad/lazyarchon/v2/internal/mcpserver"
	"github.com/yousfisaad/lazyarchon/v2/internal/plugin"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/config"
	"github.com/yousfisaad/lazyarchon/v2/internal/ui"
)

// Build-time variables (injected via ldflags)
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

func main() {
	args := os.Args[1:]

	// Subcommand dispatch: only treat the first argument as a command when
	// it is not a flag, so --version/--help still flow through stdlib flag.
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		switch args[0] {
		case "mcp":
			if err := runMCP(args[1:]); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}
			return
		case "version":
			printVersion()
			return
		case "help":
			printHelp()
			return
		default:
			fmt.Fprintf(os.Stderr, "Unknown command %q\n\n", args[0])
			printUsage()
			os.Exit(2)
		}
	}

	if code := runTUI(); code != 0 {
		os.Exit(code)
	}
}

// runTUI starts the interactive terminal application. It returns a process
// exit code instead of calling os.Exit itself, so deferred cleanup (plugin
// Close) runs before main terminates the process.
func runTUI() int {
	var (
		version  = flag.Bool("version", false, "Show version information")
		help     = flag.Bool("help", false, "Show help message")
		debug    = flag.Bool("debug", false, "Enable debug mode with verbose logging")
		logFile  = flag.String("log-file", "", "Path to log file (default: /tmp/lazyarchon.log)")
		logLevel = flag.String("log-level", "", "Log level: debug, info, warn, error (default: info, or debug if --debug)")
	)

	flag.Parse()

	if *version {
		printVersion()
		return 0
	}

	if *help {
		printHelp()
		return 0
	}

	cfg := bootstrap.LoadConfig()
	applyDebugFlags(cfg, *debug, *logFile, *logLevel)
	bootstrap.SetupLogging(cfg)

	pluginManager, err := bootstrap.InitPlugin(context.Background(), cfg)
	if err != nil {
		slog.Error("failed to initialize plugin", "error", err)
		fmt.Printf("Error: Failed to initialize plugin '%s': %v\n", cfg.GetPlugin(), err)
		return 1
	}
	defer pluginManager.Close()

	// Pass a pointer since Model.Update() uses a pointer receiver to
	// maintain component references.
	mainModel := ui.NewModelWithPlugin(cfg, pluginManager)
	bubbleteaProgram := tea.NewProgram(&mainModel, tea.WithAltScreen())

	if _, err := bubbleteaProgram.Run(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
		return 1
	}

	return 0
}

// runMCP serves the Model Context Protocol over stdio for LLM clients.
func runMCP(args []string) error {
	flags := flag.NewFlagSet("mcp", flag.ExitOnError)
	debug := flags.Bool("debug", false, "Enable debug mode with verbose logging")
	logFile := flags.String("log-file", "", "Path to log file (default: /tmp/lazyarchon.log)")
	logLevel := flags.String("log-level", "", "Log level: debug, info, warn, error")

	if err := flags.Parse(args); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := bootstrap.LoadConfig()
	applyDebugFlags(cfg, *debug, *logFile, *logLevel)
	bootstrap.SetupLogging(cfg)

	slog.Info("starting lazyarchon MCP server", "version", Version)

	manager, err := bootstrap.InitPlugin(ctx, cfg)
	if err != nil {
		return err
	}
	defer manager.Close()

	client, err := manager.GetActiveClient()
	if err != nil {
		return fmt.Errorf("failed to get backend client: %w", err)
	}

	server := mcpserver.New(client, cfg.GetPlugin(), Version)

	slog.Info("MCP server listening on stdio", "backend", cfg.GetPlugin())

	// A client closing its end of stdin is a normal shutdown, not a failure:
	// JSON-RPC peers simply terminate the process when done.
	if err := server.Run(ctx); err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}

func printVersion() {
	fmt.Printf("LazyArchon %s\n", Version)
	fmt.Printf("Commit: %s\n", Commit)
	fmt.Printf("Built: %s\n", BuildTime)
	fmt.Printf("\nAvailable plugins: %v\n", plugin.List())
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\n  lazyarchon [flags]      Terminal UI (default)\n  lazyarchon mcp [flags]  MCP server over stdio\n\n")
	fmt.Fprintf(os.Stderr, "Run 'lazyarchon help' for details.\n")
}

func printHelp() {
	fmt.Printf("LazyArchon %s - Terminal task manager (standalone or with an HTTP backend)\n\n", Version)
	fmt.Printf("Usage:\n")
	fmt.Printf("  lazyarchon [flags]       Start the terminal UI\n")
	fmt.Printf("  lazyarchon mcp [flags]   Run a Model Context Protocol server on stdio\n")
	fmt.Printf("  lazyarchon version       Show version information\n\n")
	fmt.Printf("Flags (TUI):\n")
	fmt.Printf("  -help            Show this help message\n")
	fmt.Printf("  -version         Show version information\n")
	fmt.Printf("  -debug           Enable debug mode with verbose logging\n")
	fmt.Printf("  -log-file PATH   Custom log file path (default: /tmp/lazyarchon.log)\n")
	fmt.Printf("  -log-level LEVEL Set log level: debug, info, warn, error (default: info)\n\n")
	fmt.Printf("Flags (MCP): --debug, --log-file, --log-level (same meanings)\n\n")
	fmt.Printf("Configuration:\n")
	fmt.Printf("  Config file: ~/.config/lazyarchon/config.yaml (or ./config.yaml)\n")
	fmt.Printf("  Backend: plugin: local (default) | archon | gitea | vikunja\n")
	fmt.Printf("  Local database: plugins.local.path or LAZYARCHON_DB_PATH\n\n")
	fmt.Printf("MCP setup for Claude:\n")
	fmt.Printf("  claude mcp add lazyarchon -- lazyarchon mcp\n\n")
	fmt.Printf("Examples:\n")
	fmt.Printf("  lazyarchon                             # TUI on the local SQLite database\n")
	fmt.Printf("  lazyarchon mcp                         # stdio MCP server (for Claude and friends)\n")
	fmt.Printf("  lazyarchon --debug                     # Debug mode\n\n")
	fmt.Printf("Visit https://github.com/yousfisaad/lazyarchon for more information.\n")
}

// applyDebugFlags overrides configuration with CLI debug flags
func applyDebugFlags(cfg *config.Config, debug bool, logFile string, logLevel string) {
	if debug {
		cfg.Development.Debug = true
		cfg.Development.EnableProfiling = true
		// If no log level specified with --debug, default to debug level
		if logLevel == "" {
			cfg.Development.LogLevel = "debug"
		}
	}

	if logFile != "" {
		// Store custom log file in environment variable for logger to pick up
		os.Setenv("LAZYARCHON_LOG_FILE", logFile)
	}

	if logLevel != "" {
		// Validate log level
		validLevels := map[string]bool{
			"debug": true,
			"info":  true,
			"warn":  true,
			"error": true,
		}
		if validLevels[logLevel] {
			cfg.Development.LogLevel = logLevel
		} else {
			fmt.Fprintf(os.Stderr, "Warning: Invalid log level '%s', using default\n", logLevel)
		}
	}
}
