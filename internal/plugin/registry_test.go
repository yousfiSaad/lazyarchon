package plugin

import (
	"reflect"
	"testing"
)

// namedPlugin is a minimal Plugin stub that only answers Name(); the
// embedded interface keeps the test decoupled from the full method set.
type namedPlugin struct {
	Plugin
	name string
}

func (p namedPlugin) Name() string { return p.name }

func TestRegistryListReturnsSortedNames(t *testing.T) {
	registry := NewRegistry()

	// Registered deliberately out of order: List must not depend on map
	// iteration order.
	for _, name := range []string{"vikunja", "archon", "local", "gitea"} {
		if err := registry.Register(namedPlugin{name: name}); err != nil {
			t.Fatalf("Register(%q) failed: %v", name, err)
		}
	}

	want := []string{"archon", "gitea", "local", "vikunja"}
	if got := registry.List(); !reflect.DeepEqual(got, want) {
		t.Errorf("List() = %v, want %v", got, want)
	}
}

func TestFactoryRegistryListReturnsSortedNames(t *testing.T) {
	registry := NewFactoryRegistry()

	for _, name := range []string{"vikunja", "archon", "local", "gitea"} {
		factoryName := name
		if err := registry.RegisterFactory(factoryName, func() (Plugin, error) {
			return namedPlugin{name: factoryName}, nil
		}); err != nil {
			t.Fatalf("RegisterFactory(%q) failed: %v", name, err)
		}
	}

	want := []string{"archon", "gitea", "local", "vikunja"}
	if got := registry.List(); !reflect.DeepEqual(got, want) {
		t.Errorf("List() = %v, want %v", got, want)
	}
}
