package keys

import (
	"testing"
)

func TestKeyRegistry_NewKeyRegistry(t *testing.T) {
	registry := NewKeyRegistry(nil)

	if registry == nil {
		t.Fatal("NewKeyRegistry() returned nil")
	}

	if registry.contextBindings == nil {
		t.Error("contextBindings map is nil")
	}

	if registry.keyToAction == nil {
		t.Error("keyToAction map is nil")
	}

	if registry.actionToKey == nil {
		t.Error("actionToKey map is nil")
	}
}

func TestKeyRegistry_GetContextBindings(t *testing.T) {
	registry := NewKeyRegistry(nil)

	// Test main context bindings
	mainBindings := registry.GetContextBindings("main")
	if len(mainBindings) == 0 {
		t.Error("Expected main context to have bindings")
	}

	// Test help modal context bindings
	helpBindings := registry.GetContextBindings("help_modal")
	if len(helpBindings) == 0 {
		t.Error("Expected help_modal context to have bindings")
	}

	// Test non-existent context
	nonExistentBindings := registry.GetContextBindings("non_existent")
	if len(nonExistentBindings) != 0 {
		t.Error("Expected non-existent context to return empty bindings")
	}
}

func TestKeyRegistry_GetActionForKey(t *testing.T) {
	registry := NewKeyRegistry(nil)

	// Test some known key mappings
	tests := []struct {
		key            string
		expectedAction string
	}{
		{KeyQ, ActionQuit},
		{KeyP, ActionProjectMode},
		{KeyA, ActionShowAllTasks},
	}

	for _, test := range tests {
		action := registry.GetActionForKey(test.key)
		if action != test.expectedAction {
			t.Errorf("GetActionForKey(%s) = %s, expected %s", test.key, action, test.expectedAction)
		}
	}
}

func TestKeyRegistry_GetKeyForAction(t *testing.T) {
	registry := NewKeyRegistry(nil)

	// Test some known action mappings
	tests := []struct {
		action      string
		expectedKey string
	}{
		{ActionQuit, KeyQ},
		{ActionProjectMode, KeyP},
		{ActionShowAllTasks, KeyA},
	}

	for _, test := range tests {
		key := registry.GetKeyForAction(test.action)
		if key != test.expectedKey {
			t.Errorf("GetKeyForAction(%s) = %s, expected %s", test.action, key, test.expectedKey)
		}
	}
}

func TestKeyRegistry_GetBindingsByCategory(t *testing.T) {
	registry := NewKeyRegistry(nil)

	// Test getting navigation bindings
	navBindings := registry.GetBindingsByCategory(CategoryNavigation)
	if len(navBindings) == 0 {
		t.Error("Expected navigation category to have bindings")
	}

	// Test getting application bindings
	appBindings := registry.GetBindingsByCategory(CategoryApplication)
	if len(appBindings) == 0 {
		t.Error("Expected application category to have bindings")
	}

	// Test getting task bindings
	taskBindings := registry.GetBindingsByCategory(CategoryTask)
	if len(taskBindings) == 0 {
		t.Error("Expected task category to have bindings")
	}
}

func TestKeyRegistry_GetHelpSections(t *testing.T) {
	registry := NewKeyRegistry(nil)

	sections := registry.GetHelpSections()
	if len(sections) == 0 {
		t.Error("Expected help sections to be generated")
	}

	// Verify that sections have required fields
	for i, section := range sections { //nolint:varnamelen // i is idiomatic for loop index
		if section.Title == "" {
			t.Errorf("Section %d has empty title", i)
		}

		if section.Priority == 0 {
			t.Errorf("Section %d has zero priority", i)
		}

		// Most sections should have bindings (except visual indicators)
		if len(section.Bindings) == 0 && section.Title != "Visual Indicators" && section.Title != "Task Status Symbols" {
			t.Errorf("Section %s has no bindings", section.Title)
		}
	}
}

func TestKeyBinding_Structure(t *testing.T) {
	registry := NewKeyRegistry(nil)
	mainBindings := registry.GetContextBindings("main")

	if len(mainBindings) == 0 {
		t.Fatal("No main bindings found")
	}

	// Test that bindings have required fields
	for i, binding := range mainBindings { //nolint:varnamelen // i is idiomatic for loop index
		if binding.Key == "" {
			t.Errorf("Binding %d has empty key", i)
		}

		if binding.Action == "" {
			t.Errorf("Binding %d has empty action", i)
		}

		if binding.Category == "" {
			t.Errorf("Binding %d has empty category", i)
		}

		if binding.Description == "" {
			t.Errorf("Binding %d has empty description", i)
		}

		// Context should be set
		if binding.Context == "" {
			t.Errorf("Binding %d has empty context", i)
		}
	}
}

func TestHelpSection_Structure(t *testing.T) {
	registry := NewKeyRegistry(nil)
	sections := registry.GetHelpSections()

	if len(sections) == 0 {
		t.Fatal("No help sections found")
	}

	// Test help section structure
	for i, section := range sections { //nolint:varnamelen // i is idiomatic for loop index
		if section.Title == "" {
			t.Errorf("Section %d has empty title", i)
		}

		// Bindings should have descriptions
		for j, binding := range section.Bindings {
			if binding.Description == "" {
				t.Errorf("Section %d, binding %d has empty description", i, j)
			}
		}
	}
}
