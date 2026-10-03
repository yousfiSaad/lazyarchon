package viewport

// ScrollDirection represents different scroll operations
type ScrollDirection int

const (
	ScrollUp           ScrollDirection = iota // Scroll up 1 line
	ScrollDown                                // Scroll down 1 line
	ScrollToTop                               // Jump to top
	ScrollToBottom                            // Jump to bottom
	ScrollFastUp                              // Fast scroll up (4 lines)
	ScrollFastDown                            // Fast scroll down (4 lines)
	ScrollHalfPageUp                          // Half page up
	ScrollHalfPageDown                        // Half page down
)

// ScrollbarOptions configures scrollbar rendering
type ScrollbarOptions struct {
	Enabled   bool   // Whether to render scrollbar
	Width     int    // Total scrollbar column width (default 4: gap + bar + padding)
	GapChar   string // Character for gap before scrollbar (default " ")
	ThumbChar string // Character for scrollbar thumb (default "▓")
	TrackChar string // Character for scrollbar track (default "░")
}

// DefaultScrollbarOptions returns the default scrollbar configuration
func DefaultScrollbarOptions() ScrollbarOptions {
	return ScrollbarOptions{
		Enabled:   true,
		Width:     4,
		GapChar:   " ",
		ThumbChar: "▓",
		TrackChar: "░",
	}
}
