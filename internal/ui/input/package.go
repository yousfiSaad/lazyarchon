package input

// InputHandler interface removed to avoid import cycle
// The input package provides key classification and validation functions
// but does not define interfaces that reference other packages

// Key classification helpers
func isNavigationKey(key string) bool {
	switch key {
	case "up", "k", "down", "j", "J", "K", "gg", "G", "home", "end", "ctrl+u", "ctrl+d", "pgup", "pgdown", "h", "l":
		return true
	default:
		return false
	}
}

func isSearchKey(key string) bool {
	switch key {
	case "/", "ctrl+f", "ctrl+x", "ctrl+l", "n", "N":
		return true
	default:
		return false
	}
}

func isTaskOperationKey(key string) bool {
	switch key {
	case "t", "e", "y", "Y", "f":
		return true
	default:
		return false
	}
}

func isApplicationKey(key string) bool {
	switch key {
	case "q", "ctrl+c", "r", "F5", "p", "a", "esc", "enter":
		return true
	default:
		return false
	}
}

func isModalKey(key string) bool {
	switch key {
	case "?":
		return true
	default:
		return false
	}
}