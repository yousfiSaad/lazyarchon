package mcpserver

import (
	"errors"
	"fmt"

	"github.com/yousfisaad/lazyarchon/v2/internal/plugin"
)

// toolError converts a backend error into a message suitable for an LLM
// caller: short, actionable, and including the backend name for unsupported
// operations (so the model knows to stop retrying).
func toolError(backend string, err error) error {
	var pluginErr *plugin.PluginError
	if errors.As(err, &pluginErr) {
		switch pluginErr.Code {
		case plugin.ErrorCodeUnsupported:
			return fmt.Errorf("%s backend: %s is not supported: %s", backend, pluginErr.Operation, pluginErr.Message)
		case plugin.ErrorCodeNotFound:
			return fmt.Errorf("%s", pluginErr.Message)
		default:
			return fmt.Errorf("%s backend: %s", backend, pluginErr.Message)
		}
	}

	return err
}
