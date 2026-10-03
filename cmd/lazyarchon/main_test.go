package main

import "testing"

// buildMetadata reads the package-level ldflags variables; save and restore
// them so tests stay independent of each other.
func withBuildVars(t *testing.T, version, commit, buildTime string) {
	t.Helper()

	origVersion, origCommit, origBuildTime := Version, Commit, BuildTime
	Version, Commit, BuildTime = version, commit, buildTime
	t.Cleanup(func() {
		Version, Commit, BuildTime = origVersion, origCommit, origBuildTime
	})
}

func TestBuildMetadataInjectedValuesWin(t *testing.T) {
	withBuildVars(t, "9.9.9", "abc1234", "2026-01-01T00:00:00Z")

	version, commit, buildTime := buildMetadata()
	if version != "9.9.9" || commit != "abc1234" || buildTime != "2026-01-01T00:00:00Z" {
		t.Errorf("buildMetadata() = %q, %q, %q; want injected values returned verbatim",
			version, commit, buildTime)
	}
}

func TestBuildMetadataNeverReturnsEmptyPlaceholders(t *testing.T) {
	withBuildVars(t, "dev", "unknown", "unknown")

	version, commit, buildTime := buildMetadata()
	if version == "" || commit == "" || buildTime == "" {
		t.Errorf("buildMetadata() = %q, %q, %q; want no empty values", version, commit, buildTime)
	}
}

func TestShortRevision(t *testing.T) {
	tests := []struct {
		name     string
		revision string
		want     string
	}{
		{"full hash", "6bcfa98c0d3e5a7b9d0f1e2a3b4c5d6e7f8a9b0c", "6bcfa98"},
		{"already short", "6bcfa98", "6bcfa98"},
		{"shorter than seven", "6bcfa9", "6bcfa9"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shortRevision(tt.revision); got != tt.want {
				t.Errorf("shortRevision(%q) = %q, want %q", tt.revision, got, tt.want)
			}
		})
	}
}
