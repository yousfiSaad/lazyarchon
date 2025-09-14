package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/internal/config"
	"github.com/yousfisaad/lazyarchon/internal/ui"
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
		configPath = flag.String("config", "", "Path to configuration file")
		version    = flag.Bool("version", false, "Show version information")
		help       = flag.Bool("help", false, "Show help message")
	)
	
	// Parse flags
	flag.Parse()

	// Handle version flag
	if *version {
		fmt.Printf("LazyArchon %s\n", Version)
		fmt.Printf("Commit: %s\n", Commit)
		fmt.Printf("Built: %s\n", BuildTime)
		return
	}

	// Handle help flag
	if *help {
		fmt.Printf("LazyArchon %s - Terminal UI for Archon project management\n\n", Version)
		fmt.Printf("Usage:\n")
		fmt.Printf("  lazyarchon [flags]\n\n")
		fmt.Printf("Flags:\n")
		fmt.Printf("  -config string    Path to configuration file\n")
		fmt.Printf("  -help            Show this help message\n")
		fmt.Printf("  -version         Show version information\n\n")
		fmt.Printf("Visit https://github.com/yousfisaad/lazyarchon for more information.\n")
		return
	}

	// Load configuration
	var cfg *config.Config
	var err error
	
	if *configPath != "" {
		cfg, err = config.LoadFromPath(*configPath)
	} else {
		cfg, err = config.Load()
	}
	if err != nil {
		fmt.Printf("Warning: Failed to load configuration, using defaults: %v\n", err)
	}

	// Initialize the Bubble Tea application with configuration
	model := ui.NewModel(cfg)

	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
		os.Exit(1)
	}
}
