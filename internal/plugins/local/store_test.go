package local

import (
	"path/filepath"
	"testing"
	"time"
)

func TestOpenStoreCreatesAndSeeds(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")

	st, err := openStore(path)
	if err != nil {
		t.Fatalf("openStore() error = %v", err)
	}
	defer st.Close()

	// Fresh database must be at schema version 1 with a seeded Inbox project.
	var version int
	if err := st.db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		t.Fatalf("reading user_version: %v", err)
	}

	if version != schemaVersion() {
		t.Errorf("user_version = %d, want %d", version, schemaVersion())
	}

	var count int
	if err := st.db.QueryRow("SELECT COUNT(*) FROM projects").Scan(&count); err != nil {
		t.Fatalf("counting projects: %v", err)
	}

	if count != 1 {
		t.Fatalf("seeded project count = %d, want 1", count)
	}

	var title string
	if err := st.db.QueryRow("SELECT title FROM projects").Scan(&title); err != nil {
		t.Fatalf("reading seeded project: %v", err)
	}

	if title != defaultProjectTitle {
		t.Errorf("seeded project title = %q, want %q", title, defaultProjectTitle)
	}
}

func TestOpenStoreReopenDoesNotReseed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")

	st, err := openStore(path)
	if err != nil {
		t.Fatalf("first openStore() error = %v", err)
	}
	st.Close()

	st, err = openStore(path)
	if err != nil {
		t.Fatalf("second openStore() error = %v", err)
	}
	defer st.Close()

	var count int
	if err := st.db.QueryRow("SELECT COUNT(*) FROM projects").Scan(&count); err != nil {
		t.Fatalf("counting projects: %v", err)
	}

	if count != 1 {
		t.Errorf("project count after reopen = %d, want 1 (no reseed)", count)
	}
}

func TestTimestampFormatIsFixedWidth(t *testing.T) {
	reference := time.Date(2026, 8, 28, 14, 3, 9, 123456789, time.UTC)

	formatted := formatTime(reference)

	// Fixed width guarantees lexicographic == chronological ordering.
	if len(formatted) != len(timeStampLayout) {
		t.Errorf("formatTime length = %d, want %d (%q)", len(formatted), len(timeStampLayout), formatted)
	}

	parsed, err := parseTime(formatted)
	if err != nil {
		t.Fatalf("parseTime(%q) error = %v", formatted, err)
	}

	// Round-trip loses nothing beyond microsecond precision.
	want := reference.Truncate(time.Microsecond)
	if !parsed.Equal(want) {
		t.Errorf("round-trip = %v, want %v", parsed, want)
	}

	// Lexicographic ordering must match chronological ordering.
	earlier := formatTime(reference.Add(-time.Second))
	if earlier >= formatted {
		t.Errorf("earlier timestamp %q not lexically before %q", earlier, formatted)
	}
}

func TestSplitExtraSeparatesReservedKeys(t *testing.T) {
	extra := map[string]interface{}{
		"sources":       []interface{}{map[string]interface{}{"url": "https://example.com"}},
		"code_examples": []interface{}{},
		"github_repo":   "owner/repo",
	}

	sources, codeExamples, rest := splitExtra(extra)

	if sources == "" || codeExamples == "" || rest == "" {
		t.Fatalf("splitExtra returned empty part: %q %q %q", sources, codeExamples, rest)
	}

	if sources == "null" {
		t.Errorf("sources = %q, want JSON array", sources)
	}

	if codeExamples != "[]" {
		t.Errorf("codeExamples = %q, want []", codeExamples)
	}

	decodedRest, err := decodeMap(rest)
	if err != nil {
		t.Fatalf("decodeMap(rest) error = %v", err)
	}

	if _, ok := decodedRest["sources"]; ok {
		t.Error("rest should not contain sources")
	}

	if decodedRest["github_repo"] != "owner/repo" {
		t.Errorf("rest[github_repo] = %v, want owner/repo", decodedRest["github_repo"])
	}
}
