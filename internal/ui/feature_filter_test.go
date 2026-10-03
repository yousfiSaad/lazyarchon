package ui

import (
	"testing"

	"github.com/yousfisaad/lazyarchon/v2/internal/shared/interfaces"
	"github.com/yousfisaad/lazyarchon/v2/internal/shared/utils/keys"
	feature "github.com/yousfisaad/lazyarchon/v2/internal/ui/components/modals/feature"
)

// featureFilterTestTasks returns tasks with a known tag distribution:
// one untagged, one tagged "alpha", one tagged "beta", one tagged both.
func featureFilterTestTasks() []interfaces.Task {
	return []interfaces.Task{
		{ID: "t-untagged", Title: "Untagged", Status: "todo", Priority: 3},
		{ID: "t-alpha", Title: "Alpha task", Status: "todo", Priority: 3, Tags: []string{"alpha"}},
		{ID: "t-beta", Title: "Beta task", Status: "todo", Priority: 3, Tags: []string{"beta"}},
		{ID: "t-both", Title: "Both task", Status: "todo", Priority: 3, Tags: []string{"alpha", "beta"}},
	}
}

// openFeatureModal drives the 'f' key handler and returns the modal message it produces.
func openFeatureModal(t *testing.T, model *MainModel) feature.ShowFeatureModalMsg {
	t.Helper()

	cmd, handled := model.handleFeatureSelectionKey(keys.KeyF)
	if !handled {
		t.Fatal("expected 'f' key to be handled")
	}
	if cmd == nil {
		t.Fatal("expected handler to return a command")
	}

	msg := cmd()
	showMsg, ok := msg.(feature.ShowFeatureModalMsg)
	if !ok {
		t.Fatalf("expected feature.ShowFeatureModalMsg, got %T", msg)
	}
	return showMsg
}

// Regression: after applying an empty selection ("deselect all" + Enter),
// reopening the modal must show NO features checked. Previously the handler
// conflated nil (no filter) with empty (nothing selected) and pre-checked
// every feature, so pressing Enter re-applied all features and tasks with
// unselected tags reappeared.
func TestFeatureModalReopen_EmptySelection_StaysEmpty(t *testing.T) {
	model := createTestModel(t)
	model.programContext.SetTasks(featureFilterTestTasks())

	// Empty non-nil map = filter active, nothing selected
	model.programContext.FeatureFilters = map[string]bool{}

	showMsg := openFeatureModal(t, &model)

	if len(showMsg.SelectedFeatures) != 0 {
		t.Errorf("empty selection must reopen with 0 features checked, got %d: %v",
			len(showMsg.SelectedFeatures), showMsg.SelectedFeatures)
	}
}

// nil = no filter active: modal opens with every feature checked (current,
// intended behavior for a fresh filter).
func TestFeatureModalReopen_NilSelection_PrechecksAll(t *testing.T) {
	model := createTestModel(t)
	model.programContext.SetTasks(featureFilterTestTasks())

	model.programContext.FeatureFilters = nil

	showMsg := openFeatureModal(t, &model)

	all := model.GetFeaturesForProjectSelection()
	if len(all) != 2 {
		t.Fatalf("expected 2 unique features (alpha, beta), got %v", all)
	}
	for _, f := range all {
		if !showMsg.SelectedFeatures[f] {
			t.Errorf("nil selection should precheck feature %q", f)
		}
	}
}

// Populated selection passes through to the modal unchanged.
func TestFeatureModalReopen_PreservedSelection_PassesThrough(t *testing.T) {
	model := createTestModel(t)
	model.programContext.SetTasks(featureFilterTestTasks())

	model.programContext.FeatureFilters = map[string]bool{"alpha": true}

	showMsg := openFeatureModal(t, &model)

	if len(showMsg.SelectedFeatures) != 1 || !showMsg.SelectedFeatures["alpha"] {
		t.Errorf("expected selection {alpha: true} to pass through, got %v", showMsg.SelectedFeatures)
	}
}
