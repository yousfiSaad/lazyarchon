package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
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
	// Define CLI flags
	var (
		version  = flag.Bool("version", false, "Show version information")
		help     = flag.Bool("help", false, "Show help message")
		debug    = flag.Bool("debug", false, "Enable debug mode with verbose logging")
		logFile  = flag.String("log-file", "", "Path to log file (default: /tmp/lazyarchon.log)")
		logLevel = flag.String("log-level", "", "Log level: debug, info, warn, error (default: info, or debug if --debug)")
	)

	// Parse flags
	flag.Parse()

	// Handle version flag
	if *version {
		printVersion()
		return
	}

	// Handle help flag
	if *help {
		printHelp()
		return
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		// TODO: use slog instead of print
		fmt.Println("error while loading configs -> using default configs")
	}

	// Override config with CLI flags
	applyDebugFlags(cfg, *debug, *logFile, *logLevel)

	// Create the UI model with simple constructor
	mainModel := ui.NewModel(cfg)

	// Initialize the Bubble Tea application
	// Pass pointer since Model.Update() uses pointer receiver to maintain component references
	bubbleteaProgram := tea.NewProgram(&mainModel, tea.WithAltScreen())

	if _, err := bubbleteaProgram.Run(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
		os.Exit(1)
	}
}

func printVersion() {
	fmt.Printf("LazyArchon %s\n", Version)
	fmt.Printf("Commit: %s\n", Commit)
	fmt.Printf("Built: %s\n", BuildTime)
}

func printHelp() {
	fmt.Printf("LazyArchon %s - Terminal UI for Archon project management\n\n", Version)
	fmt.Printf("Usage:\n")
	fmt.Printf("  lazyarchon [flags]\n\n")
	fmt.Printf("Flags:\n")
	fmt.Printf("  -help            Show this help message\n")
	fmt.Printf("  -version         Show version information\n")
	fmt.Printf("  -debug           Enable debug mode with verbose logging\n")
	fmt.Printf("  -log-file PATH   Custom log file path (default: /tmp/lazyarchon.log)\n")
	fmt.Printf("  -log-level LEVEL Set log level: debug, info, warn, error (default: info)\n\n")
	fmt.Printf("Examples:\n")
	fmt.Printf("  lazyarchon --debug                    # Enable debug mode\n")
	fmt.Printf("  lazyarchon --log-level warn           # Show warnings and errors only\n")
	fmt.Printf("  lazyarchon --debug --log-file ~/app.log  # Debug with custom log file\n\n")
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
