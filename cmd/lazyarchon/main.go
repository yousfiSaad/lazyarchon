package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yousfisaad/lazyarchon/internal/ui"
)

// Build-time variables (injected via ldflags)
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

func main() {
	// Handle version flag
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("LazyArchon %s\n", Version)
		fmt.Printf("Commit: %s\n", Commit)
		fmt.Printf("Built: %s\n", BuildTime)
		return
	}

	// Handle help flag
	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		fmt.Printf("LazyArchon %s - Terminal UI for Archon project management\n\n", Version)
		fmt.Printf("Usage:\n")
		fmt.Printf("  lazyarchon [flags]\n\n")
		fmt.Printf("Flags:\n")
		fmt.Printf("  -h, --help     Show this help message\n")
		fmt.Printf("  -v, --version  Show version information\n\n")
		fmt.Printf("Visit https://github.com/yousfisaad/lazyarchon for more information.\n")
		return
	}

	// Initialize the Bubble Tea application
	model := ui.NewModel()

	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
		os.Exit(1)
	}
}
