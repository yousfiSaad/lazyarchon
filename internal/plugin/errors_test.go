package plugin

import (
	"errors"
	"testing"
)

func TestUnsupported(t *testing.T) {
	tests := []struct {
		name       string
		pluginName string
		operation  string
		hint       string
	}{
		{name: "with hint", pluginName: "gitea", operation: "CreateProject", hint: "creating repositories is out of scope"},
		{name: "without hint", pluginName: "archon", operation: "DeleteProject", hint: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Unsupported(tt.pluginName, tt.operation, tt.hint)

			var pe *PluginError
			if !errors.As(err, &pe) {
				t.Fatalf("expected *PluginError, got %T", err)
			}

			if pe.Plugin != tt.pluginName {
				t.Errorf("expected plugin %q, got %q", tt.pluginName, pe.Plugin)
			}

			if pe.Operation != tt.operation {
				t.Errorf("expected operation %q, got %q", tt.operation, pe.Operation)
			}

			if pe.Code != ErrorCodeUnsupported {
				t.Errorf("expected code %q, got %q", ErrorCodeUnsupported, pe.Code)
			}

			if !IsUnsupported(err) {
				t.Error("IsUnsupported(err) = false, want true")
			}

			if !errors.Is(err, ErrUnsupportedOperation) {
				t.Error("errors.Is(err, ErrUnsupportedOperation) = false, want true")
			}
		})
	}
}

func TestIsUnsupported(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "plain error", err: errors.New("boom"), want: false},
		{name: "wrapped unsupported", err: errors.Join(errors.New("ctx: "), Unsupported("gitea", "CreateProject", "")), want: true},
		{name: "other plugin error", err: &PluginError{Plugin: "archon", Code: ErrorCodeNotFound}, want: false},
		{name: "sentinel directly", err: ErrUnsupportedOperation, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUnsupported(tt.err); got != tt.want {
				t.Errorf("IsUnsupported(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestNotFound(t *testing.T) {
	err := NotFound("local", "GetTask", "task 123")

	var pe *PluginError
	if !errors.As(err, &pe) {
		t.Fatalf("expected *PluginError, got %T", err)
	}

	if pe.Code != ErrorCodeNotFound {
		t.Errorf("expected code %q, got %q", ErrorCodeNotFound, pe.Code)
	}

	if !IsNotFound(err) {
		t.Error("IsNotFound(err) = false, want true")
	}

	if IsUnsupported(err) {
		t.Error("IsUnsupported(err) = true, want false")
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "plain error", err: errors.New("boom"), want: false},
		{name: "not found", err: NotFound("local", "GetTask", "task x"), want: true},
		{name: "unsupported is not not-found", err: Unsupported("gitea", "CreateProject", ""), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFound(tt.err); got != tt.want {
				t.Errorf("IsNotFound(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
