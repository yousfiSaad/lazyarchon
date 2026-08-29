package gitea

import (
	"context"
	"testing"

	"github.com/yousfisaad/lazyarchon/v2/internal/plugin"
)

// TestProjectWriteOperationsUnsupported verifies the Gitea adapter reports
// project write operations as unsupported rather than silently failing.
func TestProjectWriteOperationsUnsupported(t *testing.T) {
	adapter := &GiteaClientAdapter{}
	ctx := context.Background()

	tests := []struct {
		name string
		call func() error
	}{
		{
			name: "CreateProject",
			call: func() error {
				_, err := adapter.CreateProject(ctx, plugin.CreateProjectRequest{Title: "test"})
				return err
			},
		},
		{
			name: "UpdateProject",
			call: func() error {
				_, err := adapter.UpdateProject(ctx, "owner/repo", plugin.UpdateProjectRequest{})
				return err
			},
		},
		{
			name: "DeleteProject",
			call: func() error {
				return adapter.DeleteProject(ctx, "owner/repo")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call()

			if err == nil {
				t.Fatalf("%s: expected an error, got nil", tt.name)
			}

			if !plugin.IsUnsupported(err) {
				t.Errorf("%s: expected unsupported error, got: %v", tt.name, err)
			}
		})
	}
}
