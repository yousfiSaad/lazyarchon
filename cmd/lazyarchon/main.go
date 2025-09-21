package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/internal/di"
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
		version = flag.Bool("version", false, "Show version information")
		help    = flag.Bool("help", false, "Show help message")
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
		fmt.Printf("  -help            Show this help message\n")
		fmt.Printf("  -version         Show version information\n\n")
		fmt.Printf("Visit https://github.com/yousfisaad/lazyarchon for more information.\n")
		return
	}

	// Create DI container
	container, err := di.NewContainer()
	if err != nil {
		fmt.Printf("Failed to create dependency injection container: %v\n", err)
		os.Exit(1)
	}

	// Create the UI model through DI
	var uiModel tea.Model
	err = container.Invoke(func(model tea.Model) {
		uiModel = model
	})
	if err != nil {
		fmt.Printf("Failed to create UI model: %v\n", err)
		os.Exit(1)
	}

	// Initialize the Bubble Tea application
	p := tea.NewProgram(uiModel, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
		os.Exit(1)
	}
}
